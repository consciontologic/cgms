//go:build integration

package matchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	url := os.Getenv("CGMS_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("CGMS_TEST_DATABASE_URL required")
	}
	ctx := context.Background()
	admin, e := pgxpool.New(ctx, url)
	if e != nil {
		t.Fatal(e)
	}
	schema := fmt.Sprintf("store_%d", time.Now().UnixNano())
	if _, e = admin.Exec(ctx, "CREATE SCHEMA "+schema); e != nil {
		t.Fatal(e)
	}
	cfg, e := pgxpool.ParseConfig(url)
	if e != nil {
		t.Fatal(e)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		pool.Close()
		admin.Reset() // Crash tests invalidate pre-restart idle connections.
		_, e := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		if e != nil {
			t.Error(e)
		}
		admin.Close()
	})
	if e = Up(ctx, pool); e != nil {
		t.Fatal(e)
	}
	return New(pool, "rules-test")
}

func TestSnapshotRefreshesDerivedPendingContextWithoutAdvancingMatch(t *testing.T) {
	for _, kind := range []string{"open-series", "attack", "kidnapper"} {
		t.Run(kind, func(t *testing.T) {
			ctx := context.Background()
			s := testStore(t)
			env := envelopeFixture(t)
			env.RulesHash = "rules-test"
			b := env.Match.Game.Board
			var err error
			for _, move := range []struct {
				ids  []string
				seat int
				zone game.Zone
			}{
				{[]string{"deck-1-hearts-02", "deck-1-diamonds-02", "deck-1-clubs-08"}, 1, game.Series},
				{[]string{"deck-1-hearts-08", "deck-1-diamonds-08"}, 2, game.Series},
				{[]string{"deck-1-clubs-11", "deck-2-clubs-11", "deck-1-hearts-03"}, 1, game.Hand},
				{[]string{"deck-1-hearts-12"}, 2, game.Unassigned},
			} {
				b, err = b.Move(move.ids, move.seat, move.zone, true)
				if err != nil {
					t.Fatal(err)
				}
			}
			b.Players[0].History = []game.Suit{game.Hearts, game.Diamonds, game.Clubs}
			b.Players[1].History = []game.Suit{game.Hearts, game.Diamonds}
			b, err = game.OpenFormation(b, 1, "kid", game.FormationSpec{Kind: "kidnapper", Cards: []string{"deck-1-clubs-11", "deck-2-clubs-11"}})
			if err != nil {
				t.Fatal(err)
			}
			command := game.Command{GameID: b.GameID, ID: "pending", Actor: 1, Kind: kind, Suit: game.Hearts, FormationID: "kid", TargetSeat: 2}
			if kind == "attack" {
				command.Cards = []string{"deck-1-clubs-08"}
				command.Targets = []string{"deck-1-hearts-08"}
			}
			b, _, err = game.Apply(b, command)
			if err != nil {
				t.Fatal(err)
			}
			if kind == "kidnapper" {
				b.Pending.Cursor = len(b.Pending.Responders)
			}
			env.Match.Game.Board = b
			if err = s.Create(ctx, "legacy-pending", env, []string{"human", "other", "third"}); err != nil {
				t.Fatal(err)
			}
			before, err := s.Snapshot(ctx, "legacy-pending", "human")
			if err != nil {
				t.Fatal(err)
			}
			var original Projection
			if err = json.Unmarshal(before.Projection, &original); err != nil {
				t.Fatal(err)
			}
			var old map[string]any
			if err = json.Unmarshal(before.Projection, &old); err != nil {
				t.Fatal(err)
			}
			rules := old["online"].(map[string]any)["rules_context"].(map[string]any)
			if kind == "open-series" {
				delete(rules, "pending")
			}
			if combat, ok := rules["combat"].(map[string]any); ok {
				delete(combat, "attack_cards")
				delete(combat, "target_cards")
			}
			if kind == "kidnapper" {
				board := old["board"].(map[string]any)
				board["required_actor"] = 0
				delete(board, "decision_kind")
				delete(board, "choices")
			}
			legacy, err := json.Marshal(old)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.pool.Exec(ctx, `UPDATE seat_projections SET projection=$1 WHERE match_id='legacy-pending' AND seat=1`, legacy); err != nil {
				t.Fatal(err)
			}
			after, err := s.Snapshot(ctx, "legacy-pending", "human")
			if err != nil {
				t.Fatal(err)
			}
			var got Projection
			if err = json.Unmarshal(after.Projection, &got); err != nil {
				t.Fatal(err)
			}
			if got.Online.RulesContext.Pending == nil || got.Online.RulesContext.Pending.Type != kind {
				t.Fatalf("old cached action context not refreshed: %+v", got.Online.RulesContext.Pending)
			}
			if kind == "attack" && (len(got.Online.RulesContext.Combat.AttackCards) != 1 || len(got.Online.RulesContext.Combat.TargetCards) != 1) {
				t.Fatal("old cached combat operands missing")
			}
			if kind == "kidnapper" && (got.Board.RequiredActor != 1 || got.Board.DecisionKind != "kidnapper-outcome" || len(got.Board.Choices) != 1 || got.Board.DecisionID != "") {
				t.Fatal("old cached outcome boundary missing")
			}
			if after.Version != before.Version || after.Cursor != before.Cursor || game.Digest(got.Board.Cards) != game.Digest(original.Board.Cards) {
				t.Fatal("snapshot refresh changed version/cursor/card bindings")
			}
			var cached []byte
			var events, seatEvents, receipts int
			if err = s.pool.QueryRow(ctx, `SELECT projection FROM seat_projections WHERE match_id='legacy-pending' AND seat=1`).Scan(&cached); err != nil || !bytes.Equal(cached, legacy) {
				t.Fatal("snapshot refresh rewrote historical cache", err)
			}
			if err = s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM match_events),(SELECT count(*) FROM seat_events),(SELECT count(*) FROM command_results)`).Scan(&events, &seatEvents, &receipts); err != nil || events != 0 || seatEvents != 0 || receipts != 0 {
				t.Fatal("snapshot refresh advanced event or receipt state", err)
			}
		})
	}
}

func fixture(t *testing.T, s *Store) {
	t.Helper()
	b, e := game.NewState(3, "g1")
	if e != nil {
		t.Fatal(e)
	}
	m, e := game.NewMatchLifecycle(b, 3)
	if e != nil {
		t.Fatal(e)
	}
	env, e := NewEnvelope(m, "rules-test")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Create(context.Background(), "m", env, []string{"a", "b", "c"}); e != nil {
		t.Fatal(e)
	}
}
func nullify(id string, seat int) Intent {
	return Intent{ID: id, GameID: "g1", Request: Request{Operation: &game.LifecycleOperation{ID: id, GameID: "g1", Kind: "nullify", Actor: seat}}}
}
func TestDurableReceiptAndConcurrentAdmission(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	r := nullify("same", 1)
	var wg sync.WaitGroup
	out := make(chan Result, 8)
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); v, e := s.Submit(ctx, "m", "a", r); out <- v; errs <- e }()
	}
	wg.Wait()
	close(out)
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	var first []byte
	for v := range out {
		b, _ := json.Marshal(v)
		if first == nil {
			first = b
		} else if string(first) != string(b) {
			t.Fatal("retry response changed")
		}
	}
	v, e := s.Snapshot(ctx, "m", "a")
	if e != nil || v.Version != 1 {
		t.Fatalf("snapshot %v %v", v, e)
	}
	r.ExpectedVersion = 2
	if _, e = s.Submit(ctx, "m", "a", r); !errors.Is(e, ErrConflict) {
		t.Fatalf("changed body: %v", e)
	}
	if _, e = s.Snapshot(ctx, "m", "intruder"); !errors.Is(e, ErrUnauthorized) {
		t.Fatalf("membership: %v", e)
	}
}

func TestAtomicFaultRollbackAndLostAcknowledgement(t *testing.T) {
	for _, stage := range []string{"snapshot", "ledger", "events", "receipt", "acknowledgement"} {
		t.Run(stage, func(t *testing.T) {
			s := testStore(t)
			fixture(t, s)
			ctx := context.Background()
			r := nullify("fault", 1)
			s.fault = func(at string) error {
				if at == stage {
					return errors.New("injected")
				}
				return nil
			}
			before := tableState(t, s)
			_, err := s.Submit(ctx, "m", "a", r)
			if err == nil {
				t.Fatal("fault missing")
			}
			if stage == "acknowledgement" && !errors.Is(err, ErrOutcomeUnknown) {
				t.Fatal("missing unknown outcome")
			}
			s.fault = nil
			if stage != "acknowledgement" && !bytes.Equal(before, tableState(t, s)) {
				t.Fatal("partial database write")
			}
			env, v, err := s.Restore(ctx, "m")
			if err != nil {
				t.Fatal(err)
			}
			want := int64(0)
			if stage == "acknowledgement" {
				want = 1
				if !errors.Is(err, ErrOutcomeUnknown) && err != nil {
					t.Fatal(err)
				}
			}
			if v != want || len(env.Match.NullificationConsents) != int(want) {
				t.Fatalf("partial state %d %#v", v, env.Match.NullificationConsents)
			}
			for _, table := range []string{"match_events", "command_results"} {
				var n int
				if e := s.pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n); e != nil || n != int(want) {
					t.Fatalf("%s count %d %v", table, n, e)
				}
			}
			recovered, e := s.Reconcile(ctx, "m", "a", r)
			if want == 1 {
				if e != nil || recovered.Version != 1 {
					t.Fatalf("reconcile %v %v", recovered, e)
				}
			} else if !errors.Is(e, ErrOutcomeUnknown) {
				t.Fatal(e)
			}
			got, e := s.Submit(ctx, "m", "a", r)
			if e != nil || got.Version != 1 {
				t.Fatalf("retry %v %v", got, e)
			}
		})
	}
}
func TestOldGameReceiptAndFreshReplacement(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	old := nullify("consent-a", 1)
	for i, a := range []string{"a", "b", "c"} {
		r := nullify("consent-"+a, i+1)
		r.ExpectedVersion = int64(i)
		if _, e := s.Submit(ctx, "m", a, r); e != nil {
			t.Fatal(e)
		}
	}
	next := ServerIntent{ID: "next", GameID: "g1", ExpectedVersion: 3, Request: ServerRequest{GameID: "g1", Kind: "next-game", NextGameID: "g2"}}
	if _, e := s.Continue(ctx, "m", next); e != nil {
		t.Fatal(e)
	}
	got, e := s.Submit(ctx, "m", "a", old)
	if e != nil || got.GameID != "g1" || got.Version != 1 {
		t.Fatalf("historical receipt %v %v", got, e)
	}
	delayed := nullify("never-committed", 1)
	if _, e = s.Submit(ctx, "m", "a", delayed); !errors.Is(e, ErrStale) {
		t.Fatalf("retarget: %v", e)
	}
	if _, e = s.Continue(ctx, "m", next); e != nil {
		t.Fatalf("server retry %v", e)
	}
	env, v, e := s.Restore(ctx, "m")
	if e != nil || v != 4 || env.Match.Game.Board.GameID != "g2" || len(env.Match.Instances) != 2 {
		t.Fatalf("replacement %v %d %v", env.Match.Instances, v, e)
	}
	for _, cash := range env.Match.Game.Ledger.Cash {
		if cash.Sign() != 0 {
			t.Fatal("new cash")
		}
	}
}
func TestCompetingVersionsAndPrivacy(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i, a := range []string{"a", "b"} {
		wg.Add(1)
		go func(i int, a string) { defer wg.Done(); _, e := s.Submit(ctx, "m", a, nullify("race", i+1)); errs <- e }(i, a)
	}
	wg.Wait()
	close(errs)
	good, stale := 0, 0
	for e := range errs {
		if e == nil {
			good++
		} else if errors.Is(e, ErrStale) {
			stale++
		} else {
			t.Fatal(e)
		}
	}
	if good != 1 || stale != 1 {
		t.Fatalf("race %d %d", good, stale)
	}
	if _, e := s.Reconcile(ctx, "m", "intruder", nullify("race", 1)); !errors.Is(e, ErrUnauthorized) {
		t.Fatal(e)
	}
	var counts int
	if e := s.pool.QueryRow(ctx, "SELECT count(*) FROM seat_events WHERE seq=1").Scan(&counts); e != nil || counts != 3 {
		t.Fatalf("seat envelopes %d %v", counts, e)
	}
}
func TestPriorityDistinguishesUnavailableMembershipFromUnauthorized(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	if _, err := s.Priority(ctx, "m", "outsider", false); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("nonmember priority must reject authorization: %v", err)
	}
	// Preserve the match and membership rows, but make the membership relation
	// unavailable to the query in this isolated test schema.
	if _, err := s.pool.Exec(ctx, "ALTER TABLE members RENAME TO unavailable_members"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Priority(ctx, "m", "a", false); err == nil || errors.Is(err, ErrUnauthorized) {
		t.Fatalf("membership dependency failure became an authorization decision: %v", err)
	}
}

func TestCompatibilityRefusesUnsupportedCheckpoint(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	env, _, e := s.Restore(ctx, "m")
	if e != nil {
		t.Fatal(e)
	}
	env.EngineVersion = "future"
	b, _ := json.Marshal(env)
	if _, e = s.pool.Exec(ctx, "UPDATE matches SET snapshot=$1", b); e != nil {
		t.Fatal(e)
	}
	if _, _, e = s.Restore(ctx, "m"); !errors.Is(e, ErrCompatibility) {
		t.Fatalf("unsupported %v", e)
	}
}

func tableState(t *testing.T, s *Store) []byte {
	t.Helper()
	all := map[string][]string{}
	for _, table := range []string{"matches", "members", "command_results", "server_results", "match_events", "seat_events", "seat_projections", "games", "financial_charges", "financial_debts", "financial_corrections", "game_balances", "settlement_workspaces", "promises", "proposals"} {
		rows, e := s.pool.Query(context.Background(), "SELECT row_to_json(x)::text FROM "+table+" x ORDER BY row_to_json(x)::text")
		if e != nil {
			t.Fatal(e)
		}
		for rows.Next() {
			var v string
			if e = rows.Scan(&v); e != nil {
				t.Fatal(e)
			}
			all[table] = append(all[table], v)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			t.Fatal(e)
		}
	}
	b, _ := json.Marshal(all)
	return b
}
func TestPostgresCrashPreservesCommittedRandomAndReceipts(t *testing.T) {
	container := os.Getenv("CGMS_TEST_POSTGRES_CONTAINER")
	if container == "" {
		t.Fatal("bounded Docker harness required")
	}
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	deal := ServerIntent{ID: "deal", GameID: "g1", Request: ServerRequest{GameID: "g1", Kind: "deal"}}
	first, e := s.Continue(ctx, "m", deal)
	if e != nil {
		t.Fatal(e)
	}
	before := tableState(t, s)
	for _, args := range [][]string{{"kill", "--signal=KILL", container}, {"start", container}} {
		if b, e := exec.Command("docker", args...).CombinedOutput(); e != nil {
			t.Fatalf("database crash/restart %v: %s", e, b)
		}
	}
	deadline := time.Now().Add(20 * time.Second)
	for {
		if e = s.pool.Ping(ctx); e == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("database did not recover")
		}
		time.Sleep(100 * time.Millisecond)
	}
	s = New(s.pool, "rules-test")
	if !bytes.Equal(before, tableState(t, s)) {
		t.Fatal("committed database state changed after crash")
	}
	again, e := s.Continue(ctx, "m", deal)
	if e != nil || again.Version != first.Version {
		t.Fatalf("recovered random receipt %v %v", again, e)
	}
	env, _, e := s.Restore(ctx, "m")
	if e != nil || len(env.Chance) == 0 {
		t.Fatalf("committed chance lost %v", e)
	}
}
func TestRoleDomainIDs(t *testing.T) {
	if scopedID("actor:server", "x") == scopedID("scheduler:", "x") {
		t.Fatal("role collision")
	}
}

func TestReceiptPrecedesCheckpointCompatibility(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	in := nullify("before-upgrade", 1)
	want, e := s.Submit(ctx, "m", "a", in)
	if e != nil {
		t.Fatal(e)
	}
	env, _, e := s.Restore(ctx, "m")
	if e != nil {
		t.Fatal(e)
	}
	env.EngineVersion = "unsupported-future"
	b, _ := json.Marshal(env)
	if _, e = s.pool.Exec(ctx, "UPDATE matches SET snapshot=$1", b); e != nil {
		t.Fatal(e)
	}
	got, e := s.Submit(ctx, "m", "a", in)
	if e != nil || got.Version != want.Version {
		t.Fatalf("durable receipt inaccessible %v", e)
	}
	if _, e = s.Reconcile(ctx, "m", "a", in); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Submit(ctx, "m", "stranger", in); !errors.Is(e, ErrUnauthorized) {
		t.Fatalf("private compatibility leak %v", e)
	}
}

func TestPrivateJournalRetainsReplayInput(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	in := nullify("audit", 1)
	if _, e := s.Submit(ctx, "m", "a", in); e != nil {
		t.Fatal(e)
	}
	var raw []byte
	if e := s.pool.QueryRow(ctx, "SELECT event FROM match_events WHERE match_id='m' AND seq=1").Scan(&raw); e != nil {
		t.Fatal(e)
	}
	var j struct {
		ActorID   string   `json:"actor_id"`
		CommandID string   `json:"command_id"`
		Input     Intent   `json:"input"`
		Snapshot  Envelope `json:"snapshot"`
		Version   int64    `json:"version"`
	}
	if e := json.Unmarshal(raw, &j); e != nil {
		t.Fatal(e)
	}
	if j.ActorID != "a" || j.CommandID != "audit" || j.Input.Request.Operation == nil || j.Input.Request.Operation.Kind != "nullify" || j.Version != 1 || j.Snapshot.Validate() != nil {
		t.Fatal("missing private replay input")
	}
}

func TestCreateOversizedConfigurationDoesNotWrite(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	b, e := game.NewState(3, "large")
	if e != nil {
		t.Fatal(e)
	}
	m, e := game.NewMatchLifecycle(b, 3)
	if e != nil {
		t.Fatal(e)
	}
	// The initial envelope fits, but duplicated pinned config identity crosses the
	// durable cap. This must be an encoding rejection, never a NULL database write.
	rules := string(bytes.Repeat([]byte{'r'}, 33<<20))
	env, e := NewEnvelope(m, rules)
	if e != nil {
		t.Fatal(e)
	}
	s.rulesHash = rules
	e = s.Create(ctx, "large", env, []string{"a", "b", "c"})
	if e == nil || e.Error() != "snapshot size limit" {
		t.Fatal("configuration encoding failure was not preserved")
	}
	var n int
	if e = s.pool.QueryRow(ctx, "SELECT count(*) FROM matches").Scan(&n); e != nil || n != 0 {
		t.Fatal("oversized create wrote state")
	}
}
