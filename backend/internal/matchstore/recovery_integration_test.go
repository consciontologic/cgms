//go:build integration

package matchstore

import (
	"context"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

type recoveryCase struct {
	name     string
	envelope Envelope
	command  *game.Command
	server   *ServerRequest
}

func recoveryCases(t *testing.T) []recoveryCase {
	t.Helper()
	var cases []recoveryCase
	base := func() game.State {
		s, err := game.NewState(3, "recovery-game")
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	move := func(s game.State, seat int, z game.Zone, ids ...string) game.State {
		n, err := s.Move(ids, seat, z, true)
		if err != nil {
			t.Fatal(err)
		}
		return n
	}
	apply := func(s game.State, c game.Command) game.State {
		c.GameID = s.GameID
		n, _, err := game.Apply(s, c)
		if err != nil {
			t.Fatalf("fixture %s: %v", c.Kind, err)
		}
		return n
	}
	env := func(s game.State) Envelope {
		m, err := game.NewMatchLifecycle(s, 3)
		if err != nil {
			t.Fatal(err)
		}
		m.Game.TurnStarted = true
		e, err := NewEnvelope(m, "rules-test")
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	add := func(name string, s game.State, c game.Command) {
		c.GameID = s.GameID
		cases = append(cases, recoveryCase{name: name, envelope: env(s), command: &c})
	}
	combat := func() game.State {
		s := base()
		s = move(s, 1, game.Series, "deck-1-clubs-08", "deck-1-diamonds-02")
		s = move(s, 2, game.Series, "deck-1-hearts-08")
		s.Players[1].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
		return s
	}
	attack := func(s game.State) game.State {
		return apply(s, game.Command{ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	}
	s := attack(combat())
	add("response-first", s, game.Command{ID: "response", Kind: "pass", Actor: 2, WindowID: "attack"})
	s = apply(s, game.Command{ID: "first", Kind: "pass", Actor: 2, WindowID: "attack"})
	add("response-last", s, game.Command{ID: "response", Kind: "pass", Actor: 3, WindowID: "attack"})
	s = combat()
	s = move(s, 2, game.Series, "deck-1-spades-06", "deck-1-spades-09")
	s = move(s, 2, game.Attachment, "deck-1-spades-11")
	s = move(s, 1, game.Series, "deck-1-clubs-07")
	s = attack(s)
	s = apply(s, game.Command{ID: "negotiate", Kind: "negotiate", Actor: 2, WindowID: "attack"})
	c := game.Command{ID: "clubs", Kind: "decision", Actor: 1, WindowID: "attack", DecisionID: "negotiate/clubs", Cards: []string{"deck-1-clubs-08", "deck-1-clubs-07"}}
	add("negotiator-clubs", s, c)
	s = apply(s, c)
	add("negotiator-payment", s, game.Command{ID: "payment", Kind: "decision", Actor: 2, WindowID: "attack", DecisionID: "negotiate/payment", Cards: []string{"deck-1-spades-06", "deck-1-spades-09"}})
	s = combat()
	s = move(s, 1, game.Attachment, "deck-1-diamonds-11")
	s = move(s, 2, game.Series, "deck-1-spades-02")
	s = move(s, 2, game.ConcealedAce, "deck-1-diamonds-01")
	s = attack(s)
	s = apply(s, game.Command{ID: "confine", Kind: "confinement", Actor: 2, WindowID: "attack", AceID: "deck-1-diamonds-01"})
	add("dexter-confinement", s, game.Command{ID: "prevent", Kind: "decision", Actor: 1, WindowID: "attack", DecisionID: "confine/dexter", Cards: []string{"deck-1-diamonds-02"}})
	s = combat()
	queens := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}
	s = move(s, 1, game.Formation, queens...)
	for i := range s.Cards {
		if s.Cards[i].Controller == 1 && s.Cards[i].Zone == game.Formation {
			s.Cards[i].Allocation = "justice"
		}
	}
	s = apply(s, game.Command{ID: "justice", Kind: "justice", Actor: 1, AceID: queens[0]})
	for _, seat := range []int{2, 3} {
		s = apply(s, game.Command{ID: fmt.Sprintf("pass/%d", seat), Kind: "pass", Actor: seat, WindowID: "justice"})
	}
	add("justice-selection", s, game.Command{ID: "select", Kind: "decision", Actor: 1, WindowID: "justice", DecisionID: s.Pending.Decision.ID, Cards: []string{"deck-1-hearts-08"}})
	s = base()
	s = move(s, 1, game.Hand, "deck-1-spades-05")
	s = move(s, 2, game.Hand, "deck-1-clubs-05")
	s = apply(s, game.Command{ID: "offer", Kind: "offer", Actor: 1, OfferID: "offer", Revision: 1, Terms: &game.ProposalTerms{To: 2, Give: []string{"deck-1-spades-05"}, Receive: []string{"deck-1-clubs-05"}}})
	s = apply(s, game.Command{ID: "accept", Kind: "accept-offer", Actor: 2, OfferID: "offer", Revision: 1})
	add("accepted-transfer", s, game.Command{ID: "pass", Kind: "pass", Actor: 3, WindowID: "accept"})
	s = apply(s, game.Command{ID: "pass-first", Kind: "pass", Actor: 3, WindowID: "accept"})
	add("accepted-transfer-final", s, game.Command{ID: "finish", Kind: "pass", Actor: 1, WindowID: "accept"})
	s = base()
	s = move(s, 1, game.Hand, "deck-1-spades-10")
	s = apply(s, game.Command{ID: "buy", Kind: "purchase", Actor: 1, Value: 2, Cards: []string{"deck-1-spades-10"}})
	add("purchase-response", s, game.Command{ID: "pass", Kind: "pass", Actor: 2, WindowID: "buy"})
	for _, seat := range []int{2, 3} {
		s = apply(s, game.Command{ID: fmt.Sprintf("pass/%d", seat), Kind: "pass", Actor: seat, WindowID: "buy"})
	}
	cases = append(cases, recoveryCase{name: "purchase-chance", envelope: env(s), server: &ServerRequest{ID: "chance", GameID: s.GameID, Kind: "pending-chance"}})
	s = base()
	s = move(s, 2, game.ConcealedAce, "deck-1-hearts-01")
	var err error
	s, err = s.ActivateEffect(game.AceEffect{ID: "compensation", CardID: "deck-1-hearts-01", Kind: "compensation", Source: 2, Target: 2, Custodian: 2, Expiry: "target-departure-or-closure", Losses: 5})
	if err != nil {
		t.Fatal(err)
	}
	s, err = game.QueueCompensation(s)
	if err != nil {
		t.Fatal(err)
	}
	cases = append(cases, recoveryCase{name: "compensation-queue", envelope: env(s), server: &ServerRequest{ID: "continue", GameID: s.GameID, Kind: "after-action"}})
	for _, promise := range []bool{false, true} {
		e := env(base())
		e.Match.Game.Ledger, err = game.SettleReceipts(e.Match.Game.Ledger, []game.FinancialReceipt{{Seat: 0, Amount: game.IntAmount(10)}})
		if err != nil {
			t.Fatal(err)
		}
		e.Match.Game.Ledger, err = e.Match.Game.Ledger.Charge("debt", 1, 2, game.IntAmount(6))
		if err != nil {
			t.Fatal(err)
		}
		name := "automatic-transfer"
		op := game.LifecycleOperation{ID: "transfer", GameID: e.Match.Game.Board.GameID, Actor: 1, Kind: "voluntary-transfer", Recipient: 2, Amount: game.IntAmount(10)}
		if promise {
			name = "promise-payment"
			b := e.Match.Game.Promises
			b, err = b.Offer(game.Promise{ID: "promise", Payer: 0, Recipient: 1, AwardID: "award", Mode: "fixed", Amount: game.IntAmount(10)})
			if err != nil {
				t.Fatal(err)
			}
			b, err = b.Accept("promise", 1)
			if err != nil {
				t.Fatal(err)
			}
			b, err = b.RecordAward(game.PromiseAward{ID: "award", Payer: 0, Amount: game.IntAmount(10)})
			if err != nil {
				t.Fatal(err)
			}
			e.Match.Game.Promises = b
			op.Kind = "promise-pay"
			op.PromiseID = "promise"
		}
		e.Match.Game, err = e.Match.Game.ApplyOperation(op)
		if err != nil {
			t.Fatal(err)
		}
		cases = append(cases, recoveryCase{name: name, envelope: e, server: &ServerRequest{ID: "resume", GameID: e.Match.Game.Board.GameID, Kind: "operation", Operation: &game.LifecycleOperation{ID: "resume", GameID: e.Match.Game.Board.GameID, Kind: "resume-automatic", Budget: 1}}})
	}
	return cases
}

func TestEveryPendingStageRestoresAndAdvances(t *testing.T) {
	for _, tc := range recoveryCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			var expected Envelope
			var err error
			var input Intent
			var serverInput ServerIntent
			var actor string
			if tc.command != nil {
				c := *tc.command
				actor = []string{"a", "b", "c"}[c.Actor-1]
				input = Intent{ID: c.ID, GameID: c.GameID, Request: Request{Command: &c}}
				scoped := c
				scoped.ID = scopedID("actor:"+actor, c.ID)
				expected, err = applyRequest(tc.envelope, c.Actor, Request{Command: &scoped})
			} else {
				r := *tc.server
				serverInput = ServerIntent{ID: r.ID, GameID: r.GameID, Request: r}
				r.ID = scopedID("scheduler:", r.ID)
				if r.Operation != nil {
					op := *r.Operation
					op.ID = r.ID
					r.Operation = &op
				}
				expected, err = applyServerRequest(tc.envelope, r)
				// Synthetic fixtures can include preselected private outcomes. Persist them
				// normally through Create, then require restored execution to reuse them.
				if err == nil && len(expected.Chance) > len(tc.envelope.Chance) {
					tc.envelope.Chance = expected.Clone().Chance
					expected, err = applyServerRequest(tc.envelope, r)
				}
			}
			if err != nil {
				t.Fatal("fixture successor", err)
			}
			s := testStore(t)
			ctx := context.Background()
			if err = s.Create(ctx, "m", tc.envelope, []string{"a", "b", "c"}); err != nil {
				t.Fatal(err)
			}
			restored, v, err := s.Restore(ctx, "m")
			if err != nil || v != 0 || game.Digest(restored) != game.Digest(tc.envelope) {
				t.Fatal("checkpoint did not survive durable read", err)
			}
			var result Result
			if tc.command != nil {
				result, err = s.Submit(ctx, "m", actor, input)
			} else {
				result, err = s.Continue(ctx, "m", serverInput)
			}
			if err != nil || result.Version != 1 {
				t.Fatal("durable successor", err)
			}

			var retry Result
			if tc.command != nil {
				retry, err = s.Submit(ctx, "m", actor, input)
			} else {
				retry, err = s.Continue(ctx, "m", serverInput)
			}
			if err != nil || game.Digest(retry) != game.Digest(result) {
				t.Fatal("duplicate repeated or changed result", err)
			}
			if tc.command != nil {
				bad := *input.Request.Command
				bad.ID += "/stale"
				bad.WindowID += "/stale"
				if bad.DecisionID != "" {
					bad.DecisionID += "/stale"
				}
				_, err = s.Submit(ctx, "m", actor, Intent{ID: bad.ID, GameID: bad.GameID, ExpectedVersion: 1, Request: Request{Command: &bad}})
			} else {
				bad := serverInput
				bad.ID += "/stale"
				_, err = s.Continue(ctx, "m", bad)
			}
			if err == nil {
				t.Fatal("stale continuation accepted")
			}
			got, v, err := s.Restore(ctx, "m")
			if err != nil || v != 1 || game.Digest(got) != game.Digest(expected) {
				t.Fatal("restored successor differs from pure transition", err)
			}
			if err = s.VerifyJournal(ctx, "m"); err != nil {
				t.Fatal("restored successor journal", err)
			}
		})
	}
}

func TestLawfulCoupInterruptsRestoredNegotiator(t *testing.T) {
	var env Envelope
	for _, tc := range recoveryCases(t) {
		if tc.name == "negotiator-clubs" {
			env = tc.envelope
			break
		}
	}
	s := env.Match.Game.Board
	s.Players[0].History = []game.Suit{game.Clubs}
	s.Players[0].Confined = true
	var err error
	s, err = s.Move([]string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12", "deck-2-diamonds-12"}, 1, game.Hand, true)
	if err != nil {
		t.Fatal(err)
	}
	env.Match.Game.Board = s
	store := testStore(t)
	ctx := context.Background()
	if err = store.Create(ctx, "m", env, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	restored, _, err := store.Restore(ctx, "m")
	if err != nil || restored.Match.Game.Board.Pending.Decision.Stage != "clubs" {
		t.Fatal("lost negotiated decision", err)
	}
	cmd := game.Command{ID: "coup", GameID: s.GameID, Actor: 1, Kind: "coup"}
	input := Intent{ID: cmd.ID, GameID: cmd.GameID, ExpectedVersion: 0, Request: Request{Command: &cmd}}
	result, err := store.Submit(ctx, "m", "a", input)
	if err != nil {
		t.Fatal(err)
	}
	got, v, err := store.Restore(ctx, "m")
	if err != nil || v != 1 || got.Match.Game.Board.Pending != nil || got.Match.Game.Ending != "coup" {
		t.Fatal("Coup did not interrupt pending decision", err)
	}
	club, _ := got.Match.Game.Board.Card("deck-1-clubs-08")
	if club.UsedTurn != 1 || got.Match.Game.Board.Players[1].Quotas["negotiator"] != 1 {
		t.Fatal("Coup reversed committed costs")
	}
	retry, err := store.Submit(ctx, "m", "a", input)
	if err != nil || game.Digest(retry) != game.Digest(result) {
		t.Fatal("Coup repeated", err)
	}
}

func TestCompletedGameAdvancesWithinMatchAndKeepsReceipts(t *testing.T) {
	ctx := context.Background()
	store := testStore(t)
	b, err := game.NewState(3, "completed-one")
	if err != nil {
		t.Fatal(err)
	}
	b.Order = []int{1, 2, 3}
	b.Round = 13
	b.Active = 3
	m, err := game.NewMatchLifecycle(b, 3)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	m.Game.Ledger, err = m.Game.Ledger.Charge("old-charge", 0, 1, game.IntAmount(7))
	if err != nil {
		t.Fatal(err)
	}
	m.Game.Ledger, err = m.Game.Ledger.Forgive("partial-forgiveness", "old-charge", game.IntAmount(2), []int{1})
	if err != nil {
		t.Fatal(err)
	}
	env, err := NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Create(ctx, "m", env, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	end := Intent{ID: "end", GameID: b.GameID, Request: Request{Operation: &game.LifecycleOperation{ID: "end", GameID: b.GameID, Actor: 3, Kind: "end-turn"}}}
	priorReceipt, err := store.Submit(ctx, "m", "c", end)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		e, v, err := store.Restore(ctx, "m")
		if err != nil {
			t.Fatal(err)
		}
		if e.Match.Game.Automatic == nil {
			break
		}
		if i == 19 {
			t.Fatal("settlement did not finish")
		}
		id := fmt.Sprintf("resume/%d", i)
		_, err = store.Continue(ctx, "m", ServerIntent{ID: id, GameID: b.GameID, ExpectedVersion: v, Request: ServerRequest{GameID: b.GameID, Kind: "operation", Operation: &game.LifecycleOperation{GameID: b.GameID, Kind: "resume-automatic", Budget: 1}}})
		if err != nil {
			t.Fatal(err)
		}
	}
	e, v, err := store.Restore(ctx, "m")
	if err != nil || e.Match.Game.Ledger.Phase != "settlement" {
		t.Fatal("not settlement", err)
	}
	for seat, actor := range []string{"a", "b", "c"} {
		id := "finish/" + actor
		_, err = store.Submit(ctx, "m", actor, Intent{ID: id, GameID: b.GameID, ExpectedVersion: v, Request: Request{Operation: &game.LifecycleOperation{ID: id, GameID: b.GameID, Actor: seat + 1, Kind: "finish-settlement"}}})
		if err != nil {
			t.Fatal(err)
		}
		v++
	}
	_, err = store.Continue(ctx, "m", ServerIntent{ID: "finalize", GameID: b.GameID, ExpectedVersion: v, Request: ServerRequest{GameID: b.GameID, Kind: "operation", Operation: &game.LifecycleOperation{GameID: b.GameID, Kind: "finalize"}}})
	if err != nil {
		t.Fatal(err)
	}
	v++
	before, _, err := store.Restore(ctx, "m")
	if err != nil || len(before.Match.Game.Ledger.Completed) != 1 {
		t.Fatal("first result missing", err)
	}
	_, err = store.Continue(ctx, "m", ServerIntent{ID: "next", GameID: b.GameID, ExpectedVersion: v, Request: ServerRequest{GameID: b.GameID, Kind: "next-game", NextGameID: "completed-two"}})
	if err != nil {
		t.Fatal(err)
	}
	v++
	after, gotVersion, err := store.Restore(ctx, "m")
	if err != nil || gotVersion != v || after.Match.Game.Board.GameID != "completed-two" || len(after.Match.Game.Board.DrawOrder) != 104 {
		t.Fatal("fresh game not persisted", err)
	}
	if game.Digest(before.Match.Game.Ledger.Completed) != game.Digest(after.Match.Game.Ledger.Completed) || game.Digest(before.Match.Game.Ledger.Corrections) != game.Digest(after.Match.Game.Ledger.Corrections) || game.Digest(before.Match.Game.Ledger.Debts) != game.Digest(after.Match.Game.Ledger.Debts) {
		t.Fatal("historical financial lineage changed")
	}
	if len(after.Match.Game.Ledger.Debts) != 1 || after.Match.Game.Ledger.Debts[0].Remaining.Cmp(game.IntAmount(5)) != 0 {
		t.Fatal("carried debt wrong")
	}
	for _, cash := range after.Match.Game.Ledger.Cash {
		if cash.Sign() != 0 {
			t.Fatal("cash carried")
		}
	}
	retry, err := store.Submit(ctx, "m", "c", end)
	if err != nil || game.Digest(retry) != game.Digest(priorReceipt) {
		t.Fatal("old-game receipt lost", err)
	}
	unchanged, version, err := store.Restore(ctx, "m")
	if err != nil || version != v || game.Digest(unchanged) != game.Digest(after) {
		t.Fatal("historical retry mutated new game", err)
	}
}
