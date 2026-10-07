//go:build integration

package matchstore

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/metaphy6/cgms/backend/internal/game"
)

func createBoardStore(t *testing.T, board game.State) *Store {
	t.Helper()
	s := testStore(t)
	m, err := game.NewMatchLifecycle(board, 3)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	e, err := NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(context.Background(), "m", e, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	return s
}
func cardIntent(id string, v int64, c game.Command) Intent {
	c.ID = id
	return Intent{ID: id, GameID: c.GameID, ExpectedVersion: v, Request: Request{Command: &c}}
}

func TestDurableCoupCurrentVersionAndOneEnding(t *testing.T) {
	ctx := context.Background()
	for _, operation := range []bool{false, true} {
		t.Run(fmt.Sprint(operation), func(t *testing.T) {
			b, err := game.NewState(3, "g1")
			if err != nil {
				t.Fatal(err)
			}
			for i, suits := range [][]string{{"diamonds", "hearts"}, {"clubs", "spades"}} {
				var ids []string
				for _, suit := range suits {
					for copy := 1; copy <= 2; copy++ {
						ids = append(ids, fmt.Sprintf("deck-%d-%s-12", copy, suit))
					}
				}
				b, err = b.Move(ids, i+1, game.Hand, true)
				if err != nil {
					t.Fatal(err)
				}
				b.Players[i].History = []game.Suit{game.Clubs}
			}
			s := createBoardStore(t, b)
			unrelated := nullify("unrelated", 3)
			if _, err = s.Submit(ctx, "m", "c", unrelated); err != nil {
				t.Fatal(err)
			}
			var wg sync.WaitGroup
			out := make(chan error, 2)
			for seat, actor := range []string{"a", "b"} {
				wg.Add(1)
				go func(seat int, actor string) {
					defer wg.Done()
					id := "coup-" + actor
					in := cardIntent(id, 0, game.Command{GameID: "g1", Actor: seat + 1, Kind: "coup"})
					if operation {
						in.Request = Request{Operation: &game.LifecycleOperation{ID: id, GameID: "g1", Actor: seat + 1, Kind: "coup"}}
					}
					_, e := s.Submit(ctx, "m", actor, in)
					out <- e
				}(seat, actor)
			}
			wg.Wait()
			close(out)
			good := 0
			for e := range out {
				if e == nil {
					good++
				} else if !errors.Is(e, ErrInvalid) {
					t.Fatal("wrong competing ending result", e)
				}
			}
			if good != 1 {
				t.Fatal("endings", good)
			}
			e, v, err := s.Restore(ctx, "m")
			if err != nil || v != 2 || e.Match.Game.Ending != "coup" {
				t.Fatal("ending not atomic", v, err)
			}
		})
	}
}

func TestDurableProposalAcceptanceRevisionWithdrawalRace(t *testing.T) {
	ctx := context.Background()
	b, _ := game.NewState(3, "g1")
	var err error
	b, err = b.Move([]string{"deck-1-spades-05"}, 1, game.Hand, true)
	if err != nil {
		t.Fatal(err)
	}
	terms := &game.ProposalTerms{To: 2, Give: []string{"deck-1-spades-05"}}
	b, _, err = game.Apply(b, game.Command{GameID: "g1", ID: "offer", Actor: 1, Kind: "offer", OfferID: "offer", Revision: 1, Terms: terms})
	if err != nil {
		t.Fatal(err)
	}
	s := createBoardStore(t, b)
	commands := []game.Command{
		{GameID: "g1", Actor: 2, Kind: "accept-offer", OfferID: "offer", Revision: 1},
		{GameID: "g1", Actor: 1, Kind: "offer", OfferID: "offer", Revision: 2, Terms: terms},
		{GameID: "g1", Actor: 1, Kind: "withdraw-offer", OfferID: "offer", Revision: 1},
	}
	var wg sync.WaitGroup
	out := make(chan error, 3)
	for i, c := range commands {
		wg.Add(1)
		go func(i int, c game.Command) {
			defer wg.Done()
			actor := "a"
			if c.Actor == 2 {
				actor = "b"
			}
			_, e := s.Submit(ctx, "m", actor, cardIntent(fmt.Sprintf("race-%d", i), 0, c))
			out <- e
		}(i, c)
	}
	wg.Wait()
	close(out)
	good := 0
	for e := range out {
		if e == nil {
			good++
		} else if !errors.Is(e, ErrStale) {
			t.Fatal(e)
		}
	}
	if good != 1 {
		t.Fatal("multiple proposal outcomes", good)
	}
	e, v, err := s.Restore(ctx, "m")
	if err != nil || v != 1 {
		t.Fatal(v, err)
	}
	card, _ := e.Match.Game.Board.Card("deck-1-spades-05")
	if card.Controller != 1 || e.Match.Game.Board.Players[0].Ally != 0 {
		t.Fatal("early transfer or alliance")
	}
	var status string
	if err = s.pool.QueryRow(ctx, `SELECT status FROM proposals`).Scan(&status); err != nil || status != e.Match.Game.Board.Proposals[0].Status {
		t.Fatal("proposal projection drift", err)
	}
}

func TestDurablePurchaseRetriesDoNotRepeatPaymentOrDraw(t *testing.T) {
	ctx := context.Background()
	b, _ := game.NewState(3, "g1")
	payment := "deck-1-spades-10"
	var err error
	b, err = b.Move([]string{payment}, 1, game.Hand, false)
	if err != nil {
		t.Fatal(err)
	}
	s := createBoardStore(t, b)
	buy := cardIntent("buy", 0, game.Command{GameID: "g1", Kind: "purchase", Actor: 1, Value: 1, Cards: []string{payment}})
	if _, err = s.Submit(ctx, "m", "a", buy); err != nil {
		t.Fatal(err)
	}
	for i, actor := range []string{"b", "c"} {
		e, v, err := s.Restore(ctx, "m")
		if err != nil {
			t.Fatal(err)
		}
		pass := cardIntent("pass-"+actor, v, game.Command{GameID: "g1", Kind: "pass", Actor: i + 2, WindowID: e.Match.Game.Board.Pending.ID})
		if _, err = s.Submit(ctx, "m", actor, pass); err != nil {
			t.Fatal(err)
		}
	}
	chance := ServerIntent{ID: "purchase-chance", GameID: "g1", ExpectedVersion: 3, Request: ServerRequest{GameID: "g1", Kind: "pending-chance"}}
	if _, err = s.Continue(ctx, "m", chance); err != nil {
		t.Fatal(err)
	}
	before, v, err := s.Restore(ctx, "m")
	if err != nil || v != 4 {
		t.Fatal(v, err)
	}
	if _, err = s.Submit(ctx, "m", "a", buy); err != nil {
		t.Fatal("purchase retry", err)
	}
	if _, err = s.Continue(ctx, "m", chance); err != nil {
		t.Fatal("chance retry", err)
	}
	after, v, err := s.Restore(ctx, "m")
	if err != nil || v != 4 || game.Digest(before) != game.Digest(after) {
		t.Fatal("duplicate payment/draw", v, err)
	}
	count := 0
	for _, c := range after.Match.Game.Board.Cards {
		if c.Controller == 1 {
			count++
		}
	}
	if count != 1 || after.Match.Game.Board.Pending != nil || len(after.Chance) != 1 {
		t.Fatal("purchase quantity/random outcome", count, len(after.Chance))
	}
}

func TestDurableLargeDenominatorSettlementProgress(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	b, _ := game.NewState(3, "g1")
	m, err := game.NewMatchLifecycle(b, 3)
	if err != nil {
		t.Fatal(err)
	}
	l := m.Game.Ledger
	l, err = l.Charge("ab", 0, 1, game.IntAmount(1))
	if err != nil {
		t.Fatal(err)
	}
	l, err = l.Charge("ba", 1, 0, game.IntAmount(1))
	if err != nil {
		t.Fatal(err)
	}
	denominator := "100000000000000000000000000000000000000000000000003"
	receipt, err := game.NewAmount("1", denominator)
	if err != nil {
		t.Fatal(err)
	}
	cursor, err := game.BeginSettlement(l, []game.FinancialReceipt{{Seat: 0, Amount: receipt}})
	if err != nil {
		t.Fatal(err)
	}
	m.Game.Ledger = l
	m.Game.Automatic = &cursor
	m.Game.AutomaticPurpose = "transfer"
	env, err := NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, "m", env, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	first := ServerIntent{ID: "batch-one", GameID: "g1", Request: ServerRequest{GameID: "g1", Kind: "operation", Operation: &game.LifecycleOperation{GameID: "g1", Kind: "resume-automatic", Budget: 1}}}
	if _, err = s.Continue(ctx, "m", first); err != nil {
		t.Fatal(err)
	}
	s = New(s.pool, "rules-test") // discard the adapter; resume only from persisted state.
	e, v, err := s.Restore(ctx, "m")
	if err != nil || v != 1 || e.Match.Game.Automatic == nil {
		t.Fatal(v, err)
	}
	c := e.Match.Game.Automatic
	if c.Operations != 1 || len(c.Audit) != 1 || c.Audit[0].Repeats != denominator || c.Ledger.Debts[0].Remaining.Sign() != 0 || e.Match.Game.Ledger.Debts[0].Remaining.String() != "1" {
		t.Fatal("no exact bounded private progress")
	}
	var private []byte
	if err = s.pool.QueryRow(ctx, `SELECT payload FROM settlement_workspaces`).Scan(&private); err != nil || len(private) == 0 {
		t.Fatal("missing durable workspace", err)
	}
	var remaining string
	if err = s.pool.QueryRow(ctx, `SELECT remaining_num FROM financial_debts WHERE debt_id='ab'`).Scan(&remaining); err != nil || remaining != "1" {
		t.Fatal("partial financial publication", remaining, err)
	}
	if _, err = s.Continue(ctx, "m", first); err != nil {
		t.Fatal("batch retry", err)
	}
	e, _, err = s.Restore(ctx, "m")
	if err != nil || e.Match.Game.Automatic.Operations != 1 {
		t.Fatal("repeated batch", err)
	}
	overtake := nullify("overtake", 1)
	overtake.ExpectedVersion = 1
	if _, err = s.Submit(ctx, "m", "a", overtake); !errors.Is(err, ErrInvalid) {
		t.Fatal("new operation overtook pending work")
	}
	second := first
	second.ID = "batch-two"
	second.ExpectedVersion = 1
	if _, err = s.Continue(ctx, "m", second); err != nil {
		t.Fatal(err)
	}
	e, v, err = s.Restore(ctx, "m")
	if err != nil || v != 2 || e.Match.Game.Automatic != nil {
		t.Fatal("settlement did not finish", v, err)
	}
	for _, d := range e.Match.Game.Ledger.Debts {
		if d.Remaining.Sign() != 0 {
			t.Fatal("unpaid exact remainder")
		}
	}
	var count int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM settlement_workspaces`).Scan(&count); err != nil || count != 0 {
		t.Fatal("published work retained active cursor", err)
	}
}

func TestDurableFinalizationCorrectionLineageAndTiedRanks(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	prior := game.NewFinancialLedger("prior", 3)
	var err error
	prior, err = prior.Charge("carried", 0, 1, game.IntAmount(10))
	if err != nil {
		t.Fatal(err)
	}
	prior, err = prior.CloseOrdinary(nil)
	if err != nil {
		t.Fatal(err)
	}
	for seat := 0; seat < 3; seat++ {
		prior, err = prior.FinishSettlement(seat)
		if err != nil {
			t.Fatal(err)
		}
	}
	prior, err = prior.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	completed := game.Digest(prior.Completed)
	ledger, err := prior.NextGame("g1")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := game.NewState(3, "g1")
	b.Order = []int{1, 2, 3}
	b.Round = 13
	b.Active = 3
	life, err := game.NewLifecycle(b, ledger)
	if err != nil {
		t.Fatal(err)
	}
	life.TurnStarted = true
	m := game.MatchLifecycle{GameLimit: 2, Game: life, Instances: []string{"prior", "g1"}, StartLedger: ledger.Clone()}
	env, err := NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, "m", env, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	forgive := Intent{ID: "forgive", GameID: "g1", Request: Request{Operation: &game.LifecycleOperation{ID: "forgive", GameID: "g1", Kind: "forgive", Actor: 2, DebtID: "carried", Amount: game.IntAmount(10)}}}
	if _, err = s.Submit(ctx, "m", "b", forgive); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Submit(ctx, "m", "c", Intent{ID: "end", GameID: "g1", ExpectedVersion: 1, Request: Request{Operation: &game.LifecycleOperation{ID: "end", GameID: "g1", Kind: "end-turn", Actor: 3}}}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		e, v, err := s.Restore(ctx, "m")
		if err != nil {
			t.Fatal(err)
		}
		if e.Match.Game.Automatic == nil {
			break
		}
		id := fmt.Sprintf("resume-%d", i)
		if _, err = s.Continue(ctx, "m", ServerIntent{ID: id, GameID: "g1", ExpectedVersion: v, Request: ServerRequest{GameID: "g1", Kind: "operation", Operation: &game.LifecycleOperation{GameID: "g1", Kind: "resume-automatic", Budget: 1}}}); err != nil {
			t.Fatal(err)
		}
	}
	e, v, err := s.Restore(ctx, "m")
	if err != nil || e.Match.Game.Ledger.Phase != "settlement" {
		t.Fatal("round closure", err)
	}
	next := ServerIntent{ID: "too-early", GameID: "g1", ExpectedVersion: v, Request: ServerRequest{GameID: "g1", Kind: "next-game", NextGameID: "g2"}}
	if _, err = s.Continue(ctx, "m", next); !errors.Is(err, ErrInvalid) {
		t.Fatal("early next game", err)
	}
	for seat, actor := range []string{"a", "b", "c"} {
		id := "finish-" + actor
		in := Intent{ID: id, GameID: "g1", ExpectedVersion: v, Request: Request{Operation: &game.LifecycleOperation{ID: id, GameID: "g1", Kind: "finish-settlement", Actor: seat + 1}}}
		if _, err = s.Submit(ctx, "m", actor, in); err != nil {
			t.Fatal(err)
		}
		v++
	}
	finalize := ServerIntent{ID: "finalize", GameID: "g1", ExpectedVersion: v, Request: ServerRequest{GameID: "g1", Kind: "operation", Operation: &game.LifecycleOperation{GameID: "g1", Kind: "finalize"}}}
	if _, err = s.Continue(ctx, "m", finalize); err != nil {
		t.Fatal(err)
	}
	e, v, err = s.Restore(ctx, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Match.Game.Ledger.Completed) != 2 || game.Digest(e.Match.Game.Ledger.Completed[:1]) != completed || len(e.Match.Game.Ledger.Corrections) != 1 || !e.Match.Game.Ledger.Corrections[0].MatchLevel {
		t.Fatal("completed result or correction lineage changed")
	}
	for _, cash := range e.Match.Game.Ledger.Cash {
		if cash.Sign() != 0 {
			t.Fatal("final cash did not expire")
		}
	}
	for _, rank := range e.Match.Ranks() {
		if rank != 1 {
			t.Fatal("equal cumulative scores not tied", e.Match.Ranks())
		}
	}
	var original, creating string
	var matchLevel bool
	if err = s.pool.QueryRow(ctx, `SELECT c.game_id,x.game_id,x.match_level FROM financial_corrections x JOIN financial_charges c USING(match_id,charge_id)`).Scan(&original, &creating, &matchLevel); err != nil || original != "prior" || creating != "g1" || !matchLevel {
		t.Fatal("durable N06 provenance", err)
	}
	next.ID = "over-fixed-horizon"
	next.ExpectedVersion = v
	if _, err = s.Continue(ctx, "m", next); !errors.Is(err, ErrInvalid) {
		t.Fatal("fixed game count ignored", err)
	}
	if _, err = s.Continue(ctx, "m", finalize); err != nil {
		t.Fatal("finalization replay", err)
	}
}

func TestDurableOrdinaryVictoryUsesCurrentEligibility(t *testing.T) {
	for _, eligible := range []bool{false, true} {
		t.Run(fmt.Sprint(eligible), func(t *testing.T) {
			b, _ := game.NewState(3, "g1")
			var ids []string
			for copy := 1; copy <= 2; copy++ {
				for rank := 2; rank <= 10; rank++ {
					ids = append(ids, fmt.Sprintf("deck-%d-diamonds-%02d", copy, rank))
				}
			}
			count := 12
			if eligible {
				count = 13
			}
			var err error
			b, err = b.Move(ids[:count], 1, game.Hand, false)
			if err != nil {
				t.Fatal(err)
			}
			b.Players[0].History = []game.Suit{game.Clubs, game.Hearts, game.Spades, game.Diamonds}
			s := createBoardStore(t, b)
			ctx := context.Background()
			if _, err = s.Submit(ctx, "m", "c", nullify("unrelated", 3)); err != nil {
				t.Fatal(err)
			}
			request := Intent{ID: "declare", GameID: "g1", ExpectedVersion: 0, Request: Request{Operation: &game.LifecycleOperation{ID: "declare", GameID: "g1", Kind: "declare-ordinary", Actor: 1, Threshold: "diamonds"}}}
			_, err = s.Submit(ctx, "m", "a", request)
			if eligible && err != nil || !eligible && !errors.Is(err, ErrInvalid) {
				t.Fatal("wrong current-state admission", err)
			}
			e, v, err := s.Restore(ctx, "m")
			if err != nil {
				t.Fatal(err)
			}
			if !eligible && (v != 1 || e.Match.Game.Ending != "") {
				t.Fatal("invalid declaration changed game")
			}
			if eligible && (v != 2 || e.Match.Game.Ending != "diamonds") {
				t.Fatal("eligible same-game declaration not committed")
			}
		})
	}
}
