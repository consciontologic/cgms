// webfixture is a loopback-only browser QA authority. Never deploy this command:
// its synthetic accounts are deliberately selectable without credentials.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/rooms"
	"github.com/metaphy6/cgms/backend/internal/transport"
)

const fixtureRules = "browser-synthetic-fixture-v1"

var fixtureNames = []string{"turn-draw-ace", "turn-draw-diamond", "opening-large", "trade", "settlement", "opening", "purchase", "combat", "decision", "coup", "loan", "confinement", "justice", "attack", "final-settlement", "compensation", "card-style", "ad-free", "paid-daily", "paid-weekly", "paid-monthly", "paid-yearly"}

func fixture(name string, n int) (game.MatchLifecycle, error) {
	return fixtureWithID(name, n, fmt.Sprintf("fixture-%s-%d", name, n))
}
func fixtureWithID(name string, n int, id string) (game.MatchLifecycle, error) {
	b, e := game.NewState(n, id+"-game")
	if e != nil {
		return game.MatchLifecycle{}, e
	}
	for seat := 1; seat <= n; seat++ {
		b.Order = append(b.Order, seat)
	}
	move := func(ids []string, seat int, zone game.Zone) {
		if e == nil {
			b, e = b.Move(ids, seat, zone, true)
		}
	}
	switch name {
	case "turn-draw-ace", "turn-draw-diamond":
		// QA-only deterministic supply. A real End Turn followed by the server's
		// scheduled BeginTurn is still required; no client draw or synthetic event.
		top := "deck-1-hearts-01"
		if name == "turn-draw-diamond" {
			top = "deck-1-diamonds-04"
		}
		for i, id := range b.DrawOrder {
			if id == top {
				b.DrawOrder[0], b.DrawOrder[i] = b.DrawOrder[i], b.DrawOrder[0]
				break
			}
		}
	case "trade":
		move([]string{"deck-1-spades-05"}, 1, game.Hand)
	case "opening", "ad-free", "paid-daily", "paid-weekly", "paid-monthly", "paid-yearly":
		move([]string{"deck-1-clubs-04"}, 1, game.Hand)
	case "opening-large":
		ids := []string{}
		for _, c := range b.Cards {
			if c.Card.Suit == game.Clubs && c.Card.Rank >= 2 && c.Card.Rank <= 10 {
				ids = append(ids, c.Card.ID)
			}
		}
		move(ids, 1, game.Hand)
	case "purchase":
		move([]string{"deck-1-spades-10"}, 1, game.Hand)
	case "card-style":
		// Open Diamond capture changes its controller while preserving physical
		// identity. Each seat also has private hand/Ace examples for projection QA.
		move([]string{"deck-1-clubs-08", "deck-1-diamonds-02"}, 1, game.Series)
		move([]string{"deck-1-diamonds-08"}, 2, game.Series)
		move([]string{"deck-1-spades-04"}, 3, game.Series)
		if n == 4 {
			move([]string{"deck-1-hearts-04"}, 4, game.Series)
		}
		b.Players[1].History = []game.Suit{game.Clubs, game.Diamonds}
		for seat, suit := range []string{"hearts", "clubs", "spades", "diamonds"} {
			if seat >= n {
				break
			}
			move([]string{"deck-1-" + suit + "-12"}, seat+1, game.Hand)
			move([]string{"deck-1-" + suit + "-01"}, seat+1, game.ConcealedAce)
		}
	case "coup":
		move([]string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12", "deck-2-diamonds-12"}, 1, game.Hand)
		b.Players[0].History = []game.Suit{game.Clubs}
	case "loan":
		ids := []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-clubs-13", "deck-2-clubs-13", "deck-1-spades-13"}
		move(ids, 1, game.Hand)
		if e == nil {
			b, e = game.OpenFormation(b, 1, "fixture-underground", game.FormationSpec{Kind: "underground", Cards: ids})
		}
	case "justice":
		move([]string{"deck-1-clubs-08", "deck-1-diamonds-02"}, 1, game.Series)
		for seat := 2; seat <= n; seat++ {
			move([]string{fmt.Sprintf("deck-1-hearts-%02d", seat+2)}, seat, game.Series)
			b.Players[seat-1].History = []game.Suit{game.Hearts, game.Diamonds}
		}
		ids := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}
		move(ids, 1, game.Hand)
		if e == nil {
			b, e = game.OpenFormation(b, 1, "fixture-justice", game.FormationSpec{Kind: "justice", Cards: ids})
		}
		if e == nil {
			b, _, e = game.Apply(b, game.Command{GameID: b.GameID, ID: "fixture-justice-start", Kind: "justice", Actor: 1, FormationID: "fixture-justice", AceID: ids[0]})
		}
		for seat := 2; seat <= n && e == nil; seat++ {
			b, _, e = game.Apply(b, game.Command{GameID: b.GameID, ID: fmt.Sprintf("fixture-pass-%d", seat), Kind: "pass", Actor: seat, WindowID: b.Pending.ID})
		}
	case "combat", "decision", "confinement", "attack", "compensation":
		club := "deck-1-clubs-08"
		if name == "decision" {
			club = "deck-1-clubs-05"
		}
		move([]string{club, "deck-1-diamonds-02"}, 1, game.Series)
		move([]string{"deck-1-hearts-08", "deck-2-diamonds-02", "deck-2-diamonds-03"}, 2, game.Series)
		b.Players[1].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
		command := game.Command{GameID: b.GameID, ID: "fixture-attack", Kind: "attack", Actor: 1, Cards: []string{club}, Targets: []string{"deck-1-hearts-08"}}
		if name == "confinement" {
			move([]string{"deck-1-diamonds-01"}, 2, game.ConcealedAce)
		}
		if name == "decision" {
			move([]string{"deck-1-diamonds-11"}, 2, game.Attachment)
			move([]string{"deck-1-clubs-01"}, 1, game.ConcealedAce)
			command.AceID = "deck-1-clubs-01"
			command.Value = 3
		}
		if e == nil && name != "attack" {
			b, _, e = game.Apply(b, command)
		}
		if name == "compensation" {
			// Existing responses may finish while Compensation is earned. The
			// last response triggers return/shuffle, then ascending-seat draws.
			for seat := 2; seat <= 3; seat++ {
				ace := fmt.Sprintf("deck-%d-hearts-01", seat-1)
				move([]string{ace}, seat, game.ConcealedAce)
				if e == nil {
					b, e = b.ActivateEffect(game.AceEffect{ID: ace, CardID: ace, Kind: "compensation", Source: seat, Target: seat, Custodian: seat, Expiry: "target-departure-or-closure", Losses: seat + 2})
				}
			}
		}
	case "settlement", "final-settlement":
	default:
		return game.MatchLifecycle{}, errors.New("unknown fixture")
	}
	if e != nil {
		return game.MatchLifecycle{}, e
	}
	gameLimit := 3
	if name == "final-settlement" {
		gameLimit = 1
	}
	m, e := game.NewMatchLifecycle(b, gameLimit)
	if e != nil {
		return m, e
	}
	if name != "settlement" && name != "final-settlement" {
		m.Game.TurnStarted = true
		return m, m.Game.Validate()
	}
	m.Game.Promises, e = m.Game.Promises.Offer(game.Promise{ID: "promise", Payer: 0, Recipient: 1, AwardID: "award", Mode: "fixed", Amount: game.IntAmount(100)})
	if e != nil {
		return m, e
	}
	m.Game.Promises, e = m.Game.Promises.Accept("promise", 1)
	if e != nil {
		return m, e
	}
	m.Game.Promises, e = m.Game.Promises.RecordAward(game.PromiseAward{ID: "award", Payer: 0, Amount: game.IntAmount(100)})
	if e != nil {
		return m, e
	}
	m.Game.Ledger, e = game.SettleReceipts(m.Game.Ledger, []game.FinancialReceipt{{Seat: 0, Amount: game.IntAmount(100)}})
	if e != nil {
		return m, e
	}
	m.Game.Board, e = game.CloseCardBoard(m.Game.Board)
	if e != nil {
		return m, e
	}
	m.Game.Promises, e = m.Game.Promises.CloseBoard()
	if e != nil {
		return m, e
	}
	m.Game.Ledger.Phase = "settlement"
	m.Game.Ending = "ordinary"
	return m, m.Game.Validate()
}

func validListen(address string) bool {
	host, port, e := net.SplitHostPort(address)
	if e != nil {
		return false
	}
	ip := net.ParseIP(host)
	p, e := strconv.Atoi(port)
	return ip != nil && ip.IsLoopback() && e == nil && p >= 0 && p <= 65535
}

type harness struct {
	mu       sync.Mutex
	create   func(context.Context, string, string, int) ([]identity.Session, error)
	origin   string
	sessions map[string][]identity.Session
	api      http.Handler
	assets   http.Handler
}

var runPattern = regexp.MustCompile(`^[a-zA-Z0-9-]{0,32}$`)

func (h *harness) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Refuse DNS rebinding and cross-origin fixture mutations as well as API calls.
	if "http://"+r.Host != h.origin {
		http.Error(w, "invalid host", 403)
		return
	}
	if r.URL.Path == "/__fixture/session" {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "POST" || r.Header.Get("Origin") != h.origin || r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "forbidden", 403)
			return
		}
		var in struct {
			Scenario string `json:"scenario"`
			Players  int    `json:"players"`
			Seat     int    `json:"seat"`
			Run      string `json:"run,omitempty"`
		}
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		d.DisallowUnknownFields()
		if d.Decode(&in) != nil || d.Decode(new(any)) != io.EOF || !runPattern.MatchString(in.Run) || !slices.Contains(fixtureNames, in.Scenario) || (in.Players != 3 && in.Players != 4) || in.Seat < 1 || in.Seat > in.Players {
			http.Error(w, "invalid fixture", 400)
			return
		}
		id := fmt.Sprintf("fixture-%s-%d", in.Scenario, in.Players)
		if in.Run != "" {
			id += "-" + in.Run
		}
		h.mu.Lock()
		seats, ok := h.sessions[id]
		if !ok && h.create != nil && len(h.sessions) < 128 {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			var err error
			seats, err = h.create(ctx, id, in.Scenario, in.Players)
			cancel()
			if err == nil {
				h.sessions[id] = seats
				ok = true
			}
		}
		h.mu.Unlock()
		if !ok || in.Seat < 1 || in.Seat > len(seats) {
			http.Error(w, "invalid fixture", 400)
			return
		}
		s := seats[in.Seat-1]
		http.SetCookie(w, &http.Cookie{Name: "cgms_session", Value: s.Token, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, Expires: s.ExpiresAt})
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"match_id": id})
		return
	}
	if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/v1/" {
		h.api.ServeHTTP(w, r)
		return
	}
	h.assets.ServeHTTP(w, r)
}

func run(ctx context.Context, address, assets string) error {
	return runWithEconomy(ctx, address, assets, false)
}

func runWithEconomy(ctx context.Context, address, assets string, enabled bool) error {
	if !validListen(address) {
		return errors.New("fixture bind must be a literal loopback address")
	}
	if os.Getenv("CGMS_TEST_DATABASE_URL") == "" {
		return errors.New("CGMS_TEST_DATABASE_URL required")
	}
	if info, e := os.Stat(assets + "/index.html"); e != nil || info.IsDir() {
		return errors.New("built Flutter index.html required")
	}
	listener, e := net.Listen("tcp", address)
	if e != nil {
		return errors.New("fixture listener unavailable")
	}
	defer listener.Close()
	origin := "http://" + listener.Addr().String()
	admin, e := pgxpool.New(ctx, os.Getenv("CGMS_TEST_DATABASE_URL"))
	if e != nil {
		return errors.New("fixture database configuration invalid")
	}
	defer admin.Close()
	random := make([]byte, 12)
	if _, e = rand.Read(random); e != nil {
		return e
	}
	schema := "webfixture_" + hex.EncodeToString(random)
	if _, e = admin.Exec(ctx, "CREATE SCHEMA "+schema); e != nil {
		return errors.New("fixture schema creation failed")
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(c, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			log.Print("fixture schema cleanup failed")
		}
	}()
	cfg, e := pgxpool.ParseConfig(os.Getenv("CGMS_TEST_DATABASE_URL"))
	if e != nil {
		return errors.New("invalid fixture database")
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		return errors.New("fixture pool unavailable")
	}
	defer pool.Close()
	for _, up := range []func(context.Context, *pgxpool.Pool) error{matchstore.Up, identity.Up, rooms.Up, economy.Up} {
		if e = up(ctx, pool); e != nil {
			return errors.New("fixture migration failed")
		}
	}
	key, e := transport.PersistentKey(ctx, pool)
	if e != nil {
		return errors.New("fixture key unavailable")
	}
	var policy *economy.Policy
	if enabled {
		p := economy.AdoptedPolicy()
		policy = &p
	}
	api, e := transport.New(transport.Config{EconomyPolicy: policy, Pool: pool, RulesHash: fixtureRules, Origin: origin, HandleKey: key, InsecureLocal: true})
	if e != nil {
		return errors.New("fixture authority unavailable")
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := api.Close(c); err != nil {
			log.Print("fixture authority cleanup failed")
		}
	}()
	h := &harness{origin: origin, sessions: map[string][]identity.Session{}, api: api, assets: http.FileServer(http.Dir(assets))}
	ids, store := identity.New(pool), matchstore.New(pool, fixtureRules)
	if policy != nil {
		store, e = matchstore.NewWithEconomy(pool, fixtureRules, *policy)
		if e != nil {
			return e
		}
	}
	h.create = func(ctx context.Context, id, name string, n int) ([]identity.Session, error) {
		seats := []identity.Session{}
		actors := []string{}
		for i := 0; i < n; i++ {
			s, err := ids.CreateGuest(ctx)
			if err != nil {
				return nil, errors.New("fixture identity failed")
			}
			seats = append(seats, s)
			actors = append(actors, s.Account.ID)
		}
		m, err := fixtureWithID(name, n, id)
		if err != nil {
			return nil, err
		}
		env, err := matchstore.NewEnvelope(m, fixtureRules)
		if err != nil {
			return nil, err
		}
		if policy != nil {
			if err = seedEconomy(ctx, pool, id, actors, *policy, name); err != nil {
				return nil, err
			}
		}
		// Seed approved paid access before admission, so its owner consumes no
		// free start. Ad-free alone and all other seats still use their allowance.
		if err = store.Create(ctx, id, env, actors); err != nil {
			return nil, errors.New("fixture persistence failed")
		}
		return seats, nil
	}
	api.Start()
	server := &http.Server{Handler: h, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	log.Printf("synthetic browser fixture ready: %s", origin)
	select {
	case <-ctx.Done():
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(c)
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func main() {
	address := flag.String("listen", "127.0.0.1:8081", "literal loopback bind")
	assets := flag.String("assets", "client/build/web", "built Flutter assets")
	enabled := flag.Bool("economy", false, "enable approved economy with synthetic prior-day earned rewards")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if e := runWithEconomy(ctx, *address, *assets, *enabled); e != nil {
		log.Fatal(e)
	}
}
