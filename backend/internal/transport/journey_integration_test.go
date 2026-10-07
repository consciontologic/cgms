//go:build integration

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/rooms"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func journeyServer(t *testing.T) *Server {
	t.Helper()
	url := os.Getenv("CGMS_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("CGMS_TEST_DATABASE_URL required")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("http_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	p, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, up := range []func(context.Context, *pgxpool.Pool) error{matchstore.Up, identity.Up, rooms.Up} {
		if err = up(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
	s, err := New(Config{Pool: p, RulesHash: "rules-test", Origin: "https://cgms.test", HandleKey: bytes.Repeat([]byte{8}, 32)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Close(ctx); err != nil {
			t.Error(err)
		}
		p.Close()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	return s
}
func journeyRequest(t *testing.T, s *Server, seat int, token, method, path string, body any, status int) []byte {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(method, path, bytes.NewReader(b))
	r.RemoteAddr = fmt.Sprintf("192.0.2.%d:1234", seat+1)
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: status %d wanted%d: %s", method, path, w.Code, status, w.Body.String())
	}
	return w.Body.Bytes()
}
func journeyGuests(t *testing.T, s *Server, n int) []identity.Session {
	t.Helper()
	out := make([]identity.Session, n)
	for i := range out {
		b := journeyRequest(t, s, i, "", "POST", "/v1/guests", struct{}{}, 200)
		if err := json.Unmarshal(b, &out[i]); err != nil || out[i].Token == "" {
			t.Fatal("guest", err)
		}
	}
	return out
}
func TestHTTPThreeAndFourClientRooms(t *testing.T) {
	for _, n := range []int{3, 4} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			s := journeyServer(t)
			clients := journeyGuests(t, s, n)
			body := map[string]any{"command_id": "room", "capacity": n, "games_per_match": 3}
			data := journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms", body, 200)
			var room rooms.Room
			_ = json.Unmarshal(data, &room)
			retry := journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms", body, 200)
			if !bytes.Equal(data, retry) || room.Games != 3 {
				t.Fatal("room retry/horizon")
			}
			data = journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms/"+room.ID+"/invitations", struct{}{}, 200)
			var invitation map[string]string
			_ = json.Unmarshal(data, &invitation)
			for i := 1; i < n; i++ {
				journeyRequest(t, s, i, clients[i].Token, "POST", "/v1/rooms/"+room.ID+"/join", invitation, 200)
			}
			data = journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms/"+room.ID+"/matches", map[string]string{"command_id": "start"}, 200)
			var started map[string]string
			_ = json.Unmarshal(data, &started)
			id := started["match_id"]
			if id == "" {
				t.Fatal("no match")
			}
			for i := range clients {
				snap := journeyRequest(t, s, i, clients[i].Token, "GET", "/v1/matches/"+id, nil, 200)
				if bytes.Contains(snap, []byte("deck-1")) || bytes.Contains(snap, []byte("random_state")) {
					t.Fatal("physical or random identity leak")
				}
			}
			journeyRequest(t, s, 99, "invalid", "GET", "/v1/matches/"+id, nil, 401)
		})
	}
}
func TestHTTPScriptedTradeAndRecovery(t *testing.T) {
	for _, n := range []int{3, 4} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			s := journeyServer(t)
			clients := journeyGuests(t, s, n)
			board, _ := game.NewState(n, "scripted-game")
			board, _ = board.Move([]string{"deck-1-spades-05"}, 1, game.Hand, true)
			m, err := game.NewMatchLifecycle(board, 3)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.TurnStarted = true
			env, err := matchstore.NewEnvelope(m, "rules-test")
			if err != nil {
				t.Fatal(err)
			}
			actors := []string{}
			for _, client := range clients {
				actors = append(actors, client.Account.ID)
			}
			if err = s.matches.Create(context.Background(), "scripted", env, actors); err != nil {
				t.Fatal(err)
			}
			snapshot := func(seat int) (matchstore.View, matchstore.Projection) {
				var v matchstore.View
				var p matchstore.Projection
				data := journeyRequest(t, s, seat, clients[seat].Token, "GET", "/v1/matches/scripted", nil, 200)
				if err := json.Unmarshal(data, &v); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(v.Projection, &p); err != nil {
					t.Fatal(err)
				}
				return v, p
			}
			command := func(seat int, id, kind string, payload any) ([]byte, map[string]any) {
				v, _ := snapshot(seat)
				b := map[string]any{"command_id": id, "game_id": "scripted-game", "expected_version": v.Version, "type": kind, "payload": payload}
				return journeyRequest(t, s, seat, clients[seat].Token, "POST", "/v1/matches/scripted/commands", b, 200), b
			}
			_, p := snapshot(0)
			handle := p.Board.Cards[0].Card.ID
			// Authentication derives actor; forged payload fields and physical IDs fail.
			v, _ := snapshot(1)
			journeyRequest(t, s, 1, clients[1].Token, "POST", "/v1/matches/scripted/commands", map[string]any{"command_id": "bad", "game_id": "scripted-game", "expected_version": v.Version, "type": "offer", "payload": map[string]any{"actor": 1}}, 400)
			offered, offerBody := command(0, "offer", "offer", map[string]any{"offer_id": "gift", "revision": 1, "terms": map[string]any{"to": 2, "give": []string{handle}}})
			_, party := snapshot(1)
			_, outsider := snapshot(2)
			if len(party.Board.Proposals) != 1 || len(outsider.Board.Proposals) != 0 {
				t.Fatal("proposal terms privacy")
			}
			command(1, "accept", "accept-offer", map[string]any{"offer_id": "gift", "revision": 1})
			for _, seat := range append(func() []int {
				x := []int{}
				for i := 2; i < n; i++ {
					x = append(x, i)
				}
				return x
			}(), 0) {
				_, p := snapshot(seat)
				command(seat, fmt.Sprintf("pass-%d", seat), "pass", map[string]any{"pending_action_id": p.Board.WindowID})
			}
			state, _, err := s.matches.Restore(context.Background(), "scripted")
			if err != nil {
				t.Fatal(err)
			}
			card, _ := state.Match.Game.Board.Card("deck-1-spades-05")
			if card.Controller != 2 || state.Match.Game.Board.Proposals[0].Status != "completed" {
				t.Fatal("HTTP transfer differs from rules")
			}
			retry := journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/matches/scripted/commands", offerBody, 200)
			if !bytes.Equal(offered, retry) {
				t.Fatal("old committed offer retry changed")
			}
			reconciled := journeyRequest(t, s, 0, clients[0].Token, "GET", "/v1/matches/scripted/commands/offer", nil, 200)
			if !bytes.Equal(offered, reconciled) {
				t.Fatal("reconciliation differs")
			}
			journeyRequest(t, s, 2, clients[2].Token, "GET", "/v1/matches/scripted/events?after=0", nil, 200)
			if err = s.matches.VerifyJournal(context.Background(), "scripted"); err != nil {
				t.Fatal("HTTP engine trace diverged", err)
			}
		})
	}
}

func TestHTTPSessionUpgradeCookieAndWebsocketReplay(t *testing.T) {
	s := journeyServer(t)
	guests := journeyGuests(t, s, 3)
	credentials := map[string]string{"username": "journey_user", "password": "correct-horse-cgms-123"}
	data := journeyRequest(t, s, 0, guests[0].Token, "POST", "/v1/upgrade", credentials, 200)
	var upgraded identity.Session
	if err := json.Unmarshal(data, &upgraded); err != nil || upgraded.Account.ID != guests[0].Account.ID {
		t.Fatal("guest identity migration", err)
	}
	data = journeyRequest(t, s, 0, "", "POST", "/v1/login", credentials, 200)
	var logged identity.Session
	_ = json.Unmarshal(data, &logged)
	if logged.Account.ID != upgraded.Account.ID {
		t.Fatal("login identity changed")
	}
	b, _ := game.NewState(3, "ws-game")
	m, _ := game.NewMatchLifecycle(b, 3)
	e, _ := matchstore.NewEnvelope(m, "rules-test")
	if err := s.matches.Create(context.Background(), "ws", e, []string{logged.Account.ID, guests[1].Account.ID, guests[2].Account.ID}); err != nil {
		t.Fatal(err)
	}
	body := map[string]any{"command_id": "consent", "game_id": "ws-game", "expected_version": 0, "type": "nullify", "payload": map[string]any{}}
	result := journeyRequest(t, s, 0, logged.Token, "POST", "/v1/matches/ws/commands", body, 200)
	var committed matchstore.Result
	_ = json.Unmarshal(result, &committed)
	tcp := httptest.NewServer(s)
	defer tcp.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(tcp.URL, "http")+"/v1/matches/ws/stream?after=0", &websocket.DialOptions{HTTPHeader: http.Header{"Authorization": []string{"Bearer " + logged.Token}}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	_, event, err := conn.Read(ctx)
	if err != nil || !bytes.Contains(event, []byte(`"cursor":1`)) || bytes.Contains(event, []byte("deck-1")) {
		t.Fatal("websocket replay", err)
	}
	// Browser credentials are HttpOnly+Secure and state mutation requires CSRF.
	raw, _ := json.Marshal(credentials)
	r := httptest.NewRequest("POST", "/v1/login", bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", "https://cgms.test")
	r.RemoteAddr = "192.0.2.90:1234"
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatal("cookie login", w.Code)
	}
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || !cookies[0].Secure {
		t.Fatal("insecure cookie")
	}
	var browser map[string]json.RawMessage
	_ = json.Unmarshal(w.Body.Bytes(), &browser)
	if _, ok := browser["token"]; ok {
		t.Fatal("browser token exposed")
	}
	var csrf string
	_ = json.Unmarshal(browser["csrf_token"], &csrf)
	for _, valid := range []bool{false, true} {
		req := httptest.NewRequest("POST", "/v1/logout", bytes.NewReader([]byte("{}")))
		req.RemoteAddr = "192.0.2.90:1234"
		req.Header.Set("Origin", "https://cgms.test")
		req.AddCookie(cookies[0])
		if valid {
			req.Header.Set("X-CSRF-Token", csrf)
		}
		res := httptest.NewRecorder()
		s.ServeHTTP(res, req)
		want := 403
		if valid {
			want = 200
		}
		if res.Code != want {
			t.Fatal("CSRF", res.Code, want)
		}
	}
}

func TestHTTPPostBoardPromiseSettlement(t *testing.T) {
	for _, n := range []int{3, 4} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			s := journeyServer(t)
			clients := journeyGuests(t, s, n)
			board, _ := game.NewState(n, "settlement-game")
			m, _ := game.NewMatchLifecycle(board, 1)
			var err error
			m.Game.Promises, err = m.Game.Promises.Offer(game.Promise{ID: "promise", Payer: 0, Recipient: 1, AwardID: "award", Mode: "fixed", Amount: game.IntAmount(100)})
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Promises, err = m.Game.Promises.Accept("promise", 1)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Promises, err = m.Game.Promises.RecordAward(game.PromiseAward{ID: "award", Payer: 0, Amount: game.IntAmount(100)})
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Ledger, err = game.SettleReceipts(m.Game.Ledger, []game.FinancialReceipt{{Seat: 0, Amount: game.IntAmount(100)}})
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Board, err = game.CloseCardBoard(m.Game.Board)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Promises, err = m.Game.Promises.CloseBoard()
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Ledger.Phase = "settlement"
			m.Game.Ending = "ordinary"
			env, err := matchstore.NewEnvelope(m, "rules-test")
			if err != nil {
				t.Fatal(err)
			}
			actors := []string{}
			for _, client := range clients {
				actors = append(actors, client.Account.ID)
			}
			if err = s.matches.Create(context.Background(), "settle", env, actors); err != nil {
				t.Fatal(err)
			}
			s.Start()
			cmd := func(seat int, id, kind string, payload any) {
				t.Helper()
				_, v, err := s.matches.Restore(context.Background(), "settle")
				if err != nil {
					t.Fatal(err)
				}
				journeyRequest(t, s, seat, clients[seat].Token, "POST", "/v1/matches/settle/commands", map[string]any{"command_id": id, "game_id": "settlement-game", "expected_version": v, "type": kind, "payload": payload}, 200)
			}
			cmd(0, "reduce", "promise-final-offer", map[string]any{"promise_id": "promise", "amount": map[string]string{"numerator": "40", "denominator": "1"}})
			cmd(1, "accept", "promise-final-answer", map[string]any{"promise_id": "promise", "accept": true})
			pending, _, err := s.matches.Restore(context.Background(), "settle")
			if err != nil || pending.Match.Game.Promises.Promises[0].Paid.Sign() != 0 {
				t.Fatal("acceptance treated as payment", err)
			}
			cmd(0, "pay", "promise-pay", map[string]any{"promise_id": "promise"})
			deadline := time.Now().Add(3 * time.Second)
			for {
				state, _, err := s.matches.Restore(context.Background(), "settle")
				if err != nil {
					t.Fatal(err)
				}
				if state.Match.Game.Automatic == nil {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("supervisor failed automatic settlement")
				}
				time.Sleep(10 * time.Millisecond)
			}
			for seat := range clients {
				cmd(seat, fmt.Sprintf("finish-%d", seat), "finish-settlement", map[string]any{})
			}
			for {
				state, _, err := s.matches.Restore(context.Background(), "settle")
				if err != nil {
					t.Fatal(err)
				}
				if state.Match.Game.Promises.Phase == "finalized" {
					if state.Match.Game.Promises.Promises[0].Paid.Cmp(game.IntAmount(40)) != 0 {
						t.Fatal("wrong paid amount")
					}
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("supervisor failed finalization")
				}
				time.Sleep(10 * time.Millisecond)
			}
			var outcome string
			if err = s.cfg.Pool.QueryRow(context.Background(), "SELECT outcome FROM reputation_outcomes WHERE match_id='settle'").Scan(&outcome); err != nil || outcome != "anyhoo" {
				t.Fatal("Q12 outcome", outcome, err)
			}
			if err = s.matches.VerifyJournal(context.Background(), "settle"); err != nil {
				t.Fatal("settlement engine trace", err)
			}
		})
	}
}

func TestOnlineSchemaRoutesAndRequiredEnvelopeFields(t *testing.T) {
	b, err := os.ReadFile("../../../docs/api/online.openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		OpenAPI    string                                `json:"openapi"`
		Paths      map[string]map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if err = json.Unmarshal(b, &doc); err != nil || doc.OpenAPI != "3.1.0" {
		t.Fatal("OpenAPI", err)
	}
	for path, method := range map[string]string{"/v1/guests": "post", "/v1/register": "post", "/v1/login": "post", "/v1/upgrade": "post", "/v1/logout": "post", "/v1/account": "delete", "/v1/rooms": "post", "/v1/rooms/{room}": "get", "/v1/rooms/{room}/invitations": "post", "/v1/rooms/{room}/join": "post", "/v1/rooms/{room}/matches": "post", "/v1/matches/{match}": "get", "/v1/matches/{match}/commands": "post", "/v1/matches/{match}/commands/{command}": "get", "/v1/matches/{match}/events": "get", "/v1/matches/{match}/stream": "get"} {
		if _, ok := doc.Paths[path][method]; !ok {
			t.Fatal("route missing from contract", path)
		}
	}
	var command struct {
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
		Additional bool                       `json:"additionalProperties"`
	}
	if err = json.Unmarshal(doc.Components.Schemas["Command"], &command); err != nil || command.Additional || len(command.Required) != 5 {
		t.Fatal("command schema", err)
	}
	for _, name := range []string{"command_id", "game_id", "expected_version", "type", "payload"} {
		if _, ok := command.Properties[name]; !ok {
			t.Fatal("wire field", name)
		}
	}
}

func TestHTTPExclusiveLoanOnlyAfterResolution(t *testing.T) {
	s := journeyServer(t)
	clients := journeyGuests(t, s, 4)
	b, _ := game.NewState(4, "loan-game")
	ids := []string{"deck-1-hearts-11", "deck-2-hearts-11"}
	b, _ = b.Move(ids, 1, game.Formation, true)
	for i := range b.Cards {
		if b.Cards[i].Controller == 1 {
			b.Cards[i].Allocation = "kidnapper"
		}
	}
	b.Formations = []game.FormationRecord{{ID: "kidnapper", Controller: 1, Spec: game.FormationSpec{Kind: "kidnapper", Cards: ids}}}
	m, err := game.NewMatchLifecycle(b, 3)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	env, _ := matchstore.NewEnvelope(m, "rules-test")
	actors := []string{}
	for _, c := range clients {
		actors = append(actors, c.Account.ID)
	}
	if err = s.matches.Create(context.Background(), "loan", env, actors); err != nil {
		t.Fatal(err)
	}
	snapshot := func(seat int) (matchstore.View, matchstore.Projection) {
		t.Helper()
		data := journeyRequest(t, s, seat, clients[seat].Token, "GET", "/v1/matches/loan", nil, 200)
		var v matchstore.View
		var p matchstore.Projection
		_ = json.Unmarshal(data, &v)
		_ = json.Unmarshal(v.Projection, &p)
		return v, p
	}
	submit := func(seat int, id, kind string, p any) {
		v, _ := snapshot(seat)
		journeyRequest(t, s, seat, clients[seat].Token, "POST", "/v1/matches/loan/commands", map[string]any{"command_id": id, "game_id": "loan-game", "expected_version": v.Version, "type": kind, "payload": p}, 200)
	}
	_, p := snapshot(0)
	handles := []string{}
	for _, c := range p.Board.Cards {
		if c.Controller == 1 {
			handles = append(handles, c.Card.ID)
		}
	}
	submit(0, "offer", "offer", map[string]any{"offer_id": "loan", "revision": 1, "terms": map[string]any{"to": 2, "give": handles, "loan": true}})
	submit(1, "accept", "accept-offer", map[string]any{"offer_id": "loan", "revision": 1})
	pending, _, _ := s.matches.Restore(context.Background(), "loan")
	if pending.Match.Game.Board.Players[0].Ally != 0 {
		t.Fatal("premature alliance")
	}
	for _, seat := range []int{2, 3, 0} {
		_, p := snapshot(seat)
		submit(seat, fmt.Sprintf("pass%d", seat), "pass", map[string]any{"pending_action_id": p.Board.WindowID})
	}
	done, _, err := s.matches.Restore(context.Background(), "loan")
	if err != nil || done.Match.Game.Board.Players[0].Ally != 2 || done.Match.Game.Board.Players[1].Ally != 1 || done.Match.Game.Board.Formations[0].Controller != 2 {
		t.Fatal("successful loan missing exclusive alliance", err)
	}
	if err = s.matches.VerifyJournal(context.Background(), "loan"); err != nil {
		t.Fatal(err)
	}
}
