//go:build integration

package matchstore

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
)

func TestEconomyPinnedAtomicFinalizationAndNextGame(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	if err := identity.Up(ctx, s.pool); err != nil {
		t.Fatal(err)
	}
	if err := economy.Up(ctx, s.pool); err != nil {
		t.Fatal(err)
	}
	var err error
	s, err = NewWithEconomy(s.pool, "rules-test", economy.Policy{Reward: 7, PassDayPrice: 21})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 30, 23, 59, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	actors := []string{}
	for range 3 {
		guest, err := identity.New(s.pool).CreateGuest(ctx)
		if err != nil {
			t.Fatal(err)
		}
		actors = append(actors, guest.Account.ID)
	}
	makeEnvelope := func(id string, settlement bool) Envelope {
		b, err := game.NewState(3, id)
		if err != nil {
			t.Fatal(err)
		}
		m, err := game.NewMatchLifecycle(b, 3)
		if err != nil {
			t.Fatal(err)
		}
		if settlement {
			m.Game.Board, err = game.CloseCardBoard(m.Game.Board)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Ledger, err = m.Game.Ledger.CloseOrdinary(nil)
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Promises, err = m.Game.Promises.CloseBoard()
			if err != nil {
				t.Fatal(err)
			}
			m.Game.Ending = "ordinary"
		}
		e, err := NewEnvelope(m, "rules-test")
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	if err = s.Create(ctx, "economic", makeEnvelope("economic-g1", true), actors); err != nil {
		t.Fatal(err)
	}
	marked, _, err := s.Restore(ctx, "economic")
	if err != nil || marked.Schema != "cgms-match-envelope-economy-v1" {
		t.Fatal("old readers would silently ignore economy", marked.Schema, err)
	}
	for i, actor := range actors {
		id := fmt.Sprintf("finish-%d", i)
		op := game.LifecycleOperation{ID: id, GameID: "economic-g1", Actor: i + 1, Kind: "finish-settlement"}
		if _, err = s.Submit(ctx, "economic", actor, Intent{ID: id, GameID: op.GameID, ExpectedVersion: int64(i), Request: Request{Operation: &op}}); err != nil {
			t.Fatal(err)
		}
	}
	final := ServerIntent{ID: "final", GameID: "economic-g1", ExpectedVersion: 3, Request: ServerRequest{ID: "final", GameID: "economic-g1", Kind: "operation", Operation: &game.LifecycleOperation{ID: "final", GameID: "economic-g1", Kind: "finalize", Actor: 1}}}
	// A changed server default must not reprice an already started match.
	s, err = NewWithEconomy(s.pool, "rules-test", economy.Policy{Reward: 99, PassDayPrice: 99})
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return now }
	s.fault = func(stage string) error {
		if stage == "receipt" {
			return errors.New("fixture rollback")
		}
		return nil
	}
	if _, err = s.Continue(ctx, "economic", final); err == nil {
		t.Fatal("missing fault")
	}
	var rewards int
	if err = s.pool.QueryRow(ctx, "SELECT count(*) FROM economy_rewards").Scan(&rewards); err != nil || rewards != 0 {
		t.Fatal("reward escaped rollback", rewards, err)
	}
	s.fault = func(stage string) error {
		if stage == "acknowledgement" {
			return errors.New("fixture lost acknowledgement")
		}
		return nil
	}
	if _, err = s.Continue(ctx, "economic", final); !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatal("lost acknowledgement", err)
	}
	// Restart without a default policy must honor the policy already pinned to this match.
	s = New(s.pool, "rules-test")
	s.now = func() time.Time { return now }
	if _, err = s.Continue(ctx, "economic", final); err != nil {
		t.Fatal("recover finalized receipt", err)
	}
	for _, actor := range actors {
		var dirt int64
		if err = s.pool.QueryRow(ctx, "SELECT dirt FROM economy_accounts WHERE account_id=$1", actor).Scan(&dirt); err != nil || dirt != 7 {
			t.Fatal("once-only pinned reward", dirt, err)
		}
	}
	// Consume this day's remaining two starts using another explicitly enabled store.
	enabled, err := NewWithEconomy(s.pool, "rules-test", economy.Policy{Reward: 99, PassDayPrice: 99})
	if err != nil {
		t.Fatal(err)
	}
	enabled.now = s.now
	for i := 0; i < 2; i++ {
		id := fmt.Sprintf("extra-%d", i)
		if err = enabled.Create(ctx, id, makeEnvelope(id, false), actors); err != nil {
			t.Fatal(err)
		}
	}
	if err = enabled.Create(ctx, "denied", makeEnvelope("denied", false), actors); !errors.Is(err, economy.ErrAllowance) {
		t.Fatal("exhausted initial game", err)
	}
	var deniedRows int
	if err = s.pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM matches WHERE match_id='denied')+(SELECT count(*) FROM economy_starts WHERE game_id='denied')").Scan(&deniedRows); err != nil || deniedRows != 0 {
		t.Fatal("partial room/game admission", deniedRows, err)
	}
	next := ServerIntent{ID: "next", GameID: "economic-g1", ExpectedVersion: 4, Request: ServerRequest{ID: "next", GameID: "economic-g1", Kind: "next-game", NextGameID: "economic-g2"}}
	if _, err = s.Continue(ctx, "economic", next); !errors.Is(err, economy.ErrAllowance) {
		t.Fatal("exhausted next game", err)
	}
	for _, actor := range actors {
		view, err := s.Snapshot(ctx, "economic", actor)
		if err != nil || !view.AdmissionWaiting {
			t.Fatal("finalized snapshot must explain the allowance wait without identifying a seat", err)
		}
	}
	if _, err := s.Snapshot(ctx, "economic", "outsider"); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("admission status disclosed outside membership", err)
	}
	restored, version, err := s.Restore(ctx, "economic")
	if err != nil || version != 4 || restored.Match.Game.Board.GameID != "economic-g1" {
		t.Fatal("rejected start altered game", version, err)
	}
	now = now.Add(2 * time.Minute)
	view, err := s.Snapshot(ctx, "economic", actors[0])
	if err != nil || view.AdmissionWaiting {
		t.Fatal("UTC reset left stale allowance wait", err)
	}
	if _, err = s.Continue(ctx, "economic", next); err != nil {
		t.Fatal("next UTC day", err)
	}
	if _, err = s.Continue(ctx, "economic", next); err != nil {
		t.Fatal("next-game retry", err)
	}
	var charges int
	if err = s.pool.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE game_id='economic-g2'").Scan(&charges); err != nil || charges != 3 {
		t.Fatal("next-game charges", charges, err)
	}
	if _, err = s.pool.Exec(ctx, "UPDATE matches SET config=convert_to((convert_from(config,'UTF8')::jsonb-'economy')::text,'UTF8') WHERE match_id='economic'"); err != nil {
		t.Fatal(err)
	}
	if _, _, err = s.Restore(ctx, "economic"); !errors.Is(err, ErrCompatibility) {
		t.Fatal("missing pinned policy silently disabled economy", err)
	}
}

func TestBotIntentFaultsRecoverOnceWithoutAnsweringForHuman(t *testing.T) {
	for _, stage := range []string{"snapshot", "ledger", "events", "receipt", "acknowledgement"} {
		t.Run(stage, func(t *testing.T) {
			ctx := context.Background()
			s := testStore(t)
			if err := identity.Up(ctx, s.pool); err != nil {
				t.Fatal(err)
			}
			human, err := identity.New(s.pool).CreateGuest(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = s.pool.Exec(ctx, `INSERT INTO identity_accounts(account_id,kind) VALUES('bot-2','bot'),('bot-3','bot')`); err != nil {
				t.Fatal(err)
			}
			env := envelopeFixture(t)
			env.RulesHash = "rules-test"
			env.BotDifficulty = "beginner"
			env.Match.Game.RoundClosed = true
			env.Match.Game.Board.Active = 3
			env.Match.Game.TurnIndex = 2
			if err = s.Create(ctx, "bots", env, []string{human.Account.ID, "bot-2", "bot-3"}); err != nil {
				t.Fatal(err)
			}
			op := game.LifecycleOperation{ID: "bot/0/2", GameID: env.Match.Game.Board.GameID, Actor: 2, Kind: "departure-choice"}
			intent := Intent{ID: op.ID, GameID: op.GameID, ExpectedVersion: 0, Request: Request{Operation: &op}}
			s.fault = func(at string) error {
				if at == stage {
					return errors.New("injected bot persistence failure")
				}
				return nil
			}
			_, err = s.Submit(ctx, "bots", "bot-2", intent)
			if err == nil || stage == "acknowledgement" && !errors.Is(err, ErrOutcomeUnknown) {
				t.Fatal("fault contract", err)
			}
			// A new repository instance has no transient actor/receipt cache.
			s = New(s.pool, "rules-test")
			_, before, err := s.Restore(ctx, "bots")
			want := int64(0)
			if stage == "acknowledgement" {
				want = 1
			}
			if err != nil || before != want {
				t.Fatal("partial bot commit", before, want, err)
			}
			for i := 0; i < 2; i++ {
				result, err := s.Submit(ctx, "bots", "bot-2", intent)
				if err != nil || result.Version != 1 {
					t.Fatal("bot receipt recovery", result.Version, err)
				}
			}
			restored, version, err := s.Restore(ctx, "bots")
			if err != nil || version != 1 || len(restored.Match.Game.DepartureChoices) != 1 || restored.Match.Game.DepartureChoices[0].Seat != 2 || restored.Match.Game.DeparturesResolved {
				t.Fatal("bot replay or human answer fabricated", version, err)
			}
			var events, seatEvents, receipts int
			if err = s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM match_events), (SELECT count(*) FROM seat_events), (SELECT count(*) FROM command_results)`).Scan(&events, &seatEvents, &receipts); err != nil || events != 1 || seatEvents != 3 || receipts != 1 {
				t.Fatal("non-atomic bot event fanout", events, seatEvents, receipts, err)
			}
			if err = s.VerifyJournal(ctx, "bots"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPreUpgradeRoundSnapshotAndJournalRemainCompatible(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	env := envelopeFixture(t)
	env.RulesHash = "rules-test"
	env.Match.Game.RoundClosed = true
	env.Match.Game.Board.Active = 3
	env.Match.Game.TurnIndex = 2
	if err := s.Create(ctx, "legacy-round", env, []string{"human", "other", "third"}); err != nil {
		t.Fatal(err)
	}
	op := game.LifecycleOperation{ID: "stay-other", GameID: env.Match.Game.Board.GameID, Actor: 2, Kind: "departure-choice"}
	in := Intent{ID: op.ID, GameID: op.GameID, Request: Request{Operation: &op}}
	if _, err := s.Submit(ctx, "legacy-round", "other", in); err != nil {
		t.Fatal(err)
	}
	// Simulate the exact pre-upgrade public format. The private checkpoint,
	// command identity and every other byte of the receipt remain unchanged.
	legacy := func(b []byte) []byte {
		for _, field := range []string{"round_closed", "departure_pending"} {
			for _, value := range []string{"true", "false"} {
				b = bytes.ReplaceAll(b, []byte(`"`+field+`":`+value+`,`), nil)
			}
		}
		return b
	}
	var cached, receipt []byte
	if err := s.pool.QueryRow(ctx, `SELECT projection FROM seat_projections WHERE match_id='legacy-round' AND seat=1`).Scan(&cached); err != nil {
		t.Fatal(err)
	}
	if err := s.pool.QueryRow(ctx, `SELECT result FROM command_results WHERE match_id='legacy-round'`).Scan(&receipt); err != nil {
		t.Fatal(err)
	}
	oldReceipt := legacy(receipt)
	if _, err := s.pool.Exec(ctx, `UPDATE seat_projections SET projection=$1 WHERE match_id='legacy-round' AND seat=1`, legacy(cached)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.pool.Exec(ctx, `UPDATE command_results SET result=$1 WHERE match_id='legacy-round'`, oldReceipt); err != nil {
		t.Fatal(err)
	}
	s = New(s.pool, "rules-test")
	view, err := s.Snapshot(ctx, "legacy-round", "human")
	if err != nil {
		t.Fatal(err)
	}
	var projection Projection
	if err = json.Unmarshal(view.Projection, &projection); err != nil || view.Version != 1 || !projection.Online.RoundClosed || !projection.Online.DeparturePending {
		t.Fatal("pre-upgrade human wait has no prompt", view.Version, err)
	}
	if err = s.VerifyJournal(ctx, "legacy-round"); err != nil {
		t.Fatal("valid historical receipt rejected", err)
	}
	if err = s.pool.QueryRow(ctx, `SELECT result FROM command_results WHERE match_id='legacy-round'`).Scan(&receipt); err != nil || !bytes.Equal(receipt, oldReceipt) {
		t.Fatal("historical receipt rewritten", err)
	}
	// Compatibility must not excuse a changed historical game field or an
	// explicitly incorrect new field; only absent round metadata is allowed.
	for _, modified := range [][]byte{
		bytes.Replace(oldReceipt, []byte(`"games_per_match":3`), []byte(`"games_per_match":99`), 1),
		bytes.Replace(oldReceipt, []byte(`"online":{`), []byte(`"online":{"round_closed":false,`), 1),
	} {
		if _, err = s.pool.Exec(ctx, `UPDATE command_results SET result=$1 WHERE match_id='legacy-round'`, modified); err != nil {
			t.Fatal(err)
		}
		if err = s.VerifyJournal(ctx, "legacy-round"); !errors.Is(err, ErrJournal) {
			t.Fatal("compatibility hid altered historical fields", err)
		}
	}
}

func TestPreUpgradeServerReadinessSnapshotAndJournalRemainCompatible(t *testing.T) {
	for _, beforeRoundMetadata := range []bool{false, true} {
		t.Run(fmt.Sprint("before-round-", beforeRoundMetadata), func(t *testing.T) {
			ctx := context.Background()
			s := testStore(t)
			env := envelopeFixture(t)
			env.RulesHash = "rules-test"
			env.Match.Game.Board.Order = nil
			if err := s.Create(ctx, "legacy-start", env, []string{"human", "other", "third"}); err != nil {
				t.Fatal(err)
			}
			initial, ok := NextServer(env, 0)
			if !ok || initial.Request.Kind != "deal" {
				t.Fatal("missing initial automatic deal")
			}
			if _, err := s.Continue(ctx, "legacy-start", initial); err != nil {
				t.Fatal(err)
			}
			legacy := func(b []byte) []byte {
				fields := []string{"server_pending"}
				if beforeRoundMetadata {
					fields = append(fields, "round_closed", "departure_pending")
				}
				for _, field := range fields {
					for _, value := range []string{"true", "false"} {
						b = bytes.ReplaceAll(b, []byte(`"`+field+`":`+value+`,`), nil)
					}
				}
				return b
			}
			// First prove an already persisted provisional deal is enriched before
			// initialization resumes; no browser refresh may call it a ready turn.
			var provisional []byte
			if err := s.pool.QueryRow(ctx, `SELECT projection FROM seat_projections WHERE match_id='legacy-start' AND seat=1`).Scan(&provisional); err != nil {
				t.Fatal(err)
			}
			if _, err := s.pool.Exec(ctx, `UPDATE seat_projections SET projection=$1 WHERE match_id='legacy-start' AND seat=1`, legacy(provisional)); err != nil {
				t.Fatal(err)
			}
			first, err := s.Snapshot(ctx, "legacy-start", "human")
			var firstProjection Projection
			if err != nil || json.Unmarshal(first.Projection, &firstProjection) != nil || first.Version != 1 || !firstProjection.Online.ServerPending {
				t.Fatal("persisted initial deal exposed as actionable", first.Version, err)
			}
			// Server receipts deliberately contain no public projection. Reach a
			// genuine turn, then record a human EndTurn receipt whose next turn is
			// pending, to exercise historical projected receipt verification too.
			var version int64
			for steps := 0; steps < 32; steps++ {
				env, version, err = s.Restore(ctx, "legacy-start")
				if err != nil {
					t.Fatal(err)
				}
				next, pending := NextServer(env, version)
				if !pending {
					break
				}
				if _, err = s.Continue(ctx, "legacy-start", next); err != nil {
					t.Fatal(err)
				}
			}
			if !env.Match.Game.TurnStarted {
				t.Fatal("bounded setup did not reach a real turn")
			}
			active := env.Match.Game.Board.Active
			op := game.LifecycleOperation{ID: "human-end", GameID: env.Match.Game.Board.GameID, Actor: active, Kind: "end-turn"}
			intent := Intent{ID: op.ID, GameID: op.GameID, ExpectedVersion: version, Request: Request{Operation: &op}}
			if _, err = s.Submit(ctx, "legacy-start", []string{"human", "other", "third"}[active-1], intent); err != nil {
				t.Fatal(err)
			}
			version++
			var cached, receipt []byte
			if err := s.pool.QueryRow(ctx, `SELECT projection FROM seat_projections WHERE match_id='legacy-start' AND seat=1`).Scan(&cached); err != nil {
				t.Fatal(err)
			}
			if err := s.pool.QueryRow(ctx, `SELECT result FROM command_results WHERE match_id='legacy-start'`).Scan(&receipt); err != nil {
				t.Fatal(err)
			}
			oldCached, oldReceipt := legacy(cached), legacy(receipt)
			if _, err := s.pool.Exec(ctx, `UPDATE seat_projections SET projection=$1 WHERE match_id='legacy-start' AND seat=1`, oldCached); err != nil {
				t.Fatal(err)
			}
			if _, err := s.pool.Exec(ctx, `UPDATE command_results SET result=$1 WHERE match_id='legacy-start'`, oldReceipt); err != nil {
				t.Fatal(err)
			}
			s = New(s.pool, "rules-test")
			view, err := s.Snapshot(ctx, "legacy-start", "human")
			if err != nil {
				t.Fatal(err)
			}
			var projection Projection
			if err = json.Unmarshal(view.Projection, &projection); err != nil || view.Version != version || !projection.Online.ServerPending || projection.Online.AutomaticPending {
				t.Fatal("old provisional active seat exposed as ready", view.Version, err)
			}
			if err = s.VerifyJournal(ctx, "legacy-start"); err != nil {
				t.Fatal("valid historical initialization receipt rejected", err)
			}
			if err = s.pool.QueryRow(ctx, `SELECT result FROM command_results WHERE match_id='legacy-start'`).Scan(&receipt); err != nil || !bytes.Equal(receipt, oldReceipt) {
				t.Fatal("immutable historical receipt rewritten", err)
			}
			if err = s.pool.QueryRow(ctx, `SELECT projection FROM seat_projections WHERE match_id='legacy-start' AND seat=1`).Scan(&cached); err != nil || !bytes.Equal(cached, oldCached) {
				t.Fatal("cached historical projection rewritten", err)
			}
			for _, modified := range [][]byte{
				bytes.Replace(oldReceipt, []byte(`"automatic_pending":false,`), []byte(`"automatic_pending":false,"server_pending":false,`), 1),
				bytes.Replace(oldReceipt, []byte(`"games_per_match":3`), []byte(`"games_per_match":99`), 1),
			} {
				if _, err = s.pool.Exec(ctx, `UPDATE command_results SET result=$1 WHERE match_id='legacy-start'`, modified); err != nil {
					t.Fatal(err)
				}
				if err = s.VerifyJournal(ctx, "legacy-start"); !errors.Is(err, ErrJournal) {
					t.Fatal("compatibility accepted incorrect persisted metadata", err)
				}
			}
		})
	}
}
