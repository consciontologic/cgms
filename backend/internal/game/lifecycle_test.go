package game

import (
	"fmt"
	"testing"
)

func lifecycleFixture(t *testing.T) Lifecycle {
	t.Helper()
	s, _ := NewState(3, "lifecycle-game")
	s.Order = []int{1, 2, 3}
	l, e := NewLifecycle(s, NewFinancialLedger(s.GameID, 3))
	if e != nil {
		t.Fatal(e)
	}
	return l
}
func TestLifecycleScheduledDrawAndRoundBoundary(t *testing.T) {
	l := lifecycleFixture(t)
	l.Board.Players[0].Confined = true
	before := Digest(l)
	n, got, e := l.BeginTurn(nil)
	if e != nil {
		t.Fatal(e)
	}
	if len(got) != 1 || n.Board.Players[0].Confined || !n.TurnStarted || Digest(l) != before {
		t.Fatal("scheduled turn must expire confinement, draw once, and preserve input")
	}
	if _, _, e = n.BeginTurn(nil); e == nil {
		t.Fatal("duplicate start")
	}
	for seat := 1; seat <= 3; seat++ {
		if seat > 1 {
			n, _, e = n.BeginTurn(nil)
			if e != nil {
				t.Fatal(e)
			}
		}
		n, e = n.EndTurn()
		if e != nil {
			t.Fatal(e)
		}
	}
	if !n.RoundClosed || n.Board.Round != 1 || n.TurnStarted {
		t.Fatal("round must await round settlement and fresh initiative")
	}
}
func TestLifecycleOrdinaryPossessionScoringAndDeclaration(t *testing.T) {
	l := lifecycleFixture(t)
	ids := []string{}
	for _, c := range Deck() {
		if c.Suit == Diamonds && c.Rank >= 2 && c.Rank <= 10 && len(ids) < 13 {
			ids = append(ids, c.ID)
		}
	}
	var e error
	l.Board, e = l.Board.Move(ids, 1, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	l.Board.Players[0].History = []Suit{Hearts, Clubs, Spades, Diamonds}
	n, e := l.DeclareOrdinary(1, "diamonds")
	if e != nil {
		t.Fatal(e)
	}
	if n.Board.Phase != "settlement" || n.Ending != "diamonds" || n.Automatic == nil {
		t.Fatal("close board and retain automatic settlement")
	}
	for n.Automatic != nil {
		n, e = n.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	// Thirteen physical diamonds: first deck 2..10 ->54; second 2..5 ->14;
	// 13 default bonuses ->65; one declaration bonus ->50. Total 183.
	if n.Ledger.Scores[0].Cmp(IntAmount(183)) != 0 {
		t.Fatalf("rule-derived award: %v", n.Ledger.Scores[0])
	}
	if _, e = n.DeclareOrdinary(1, "diamonds"); e == nil {
		t.Fatal("duplicate declaration")
	}
}
func TestLifecycleRejectsIllegalDeclarationWithoutPublishingOrMutation(t *testing.T) {
	l := lifecycleFixture(t)
	before := Digest(l)
	if _, e := l.DeclareOrdinary(1, "diamonds"); e == nil {
		t.Fatal("invalid claim")
	}
	if Digest(l) != before {
		t.Fatal("rejected claim mutated input")
	}
}

func TestLifecycleCodePersistentEndSettings(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	l.Board, e = l.Board.Move([]string{"deck-1-diamonds-02", "deck-1-spades-13"}, 2, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	l.Board.Code = &CodeSettings{Declarer: 1, Formation: "gone", DiamondBonus: 1, PurchasePrice: 10}
	a := ordinaryReceipts(l.Board, 0)
	if a[1].Amount.Cmp(IntAmount(13)) != 0 {
		t.Fatalf("persisted Code settings and spade royal: got %s want 13", a[1].Amount)
	}
}
func TestLifecycleRound13ClosesOnlyAfterIncome(t *testing.T) {
	l := lifecycleFixture(t)
	l.Board.Round = 13
	l.TurnIndex = 2
	l.Board.Active = 3
	l.TurnStarted = true
	var e error
	l, e = l.EndTurn()
	if e != nil {
		t.Fatal(e)
	}
	if l.Ending != "" || l.Automatic == nil {
		t.Fatal("round income is explicit before closure")
	}
	for l.Automatic != nil {
		l, e = l.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if l.Ending != "round-limit" || l.Ledger.Phase != "settlement" {
		t.Fatal("round thirteen did not close")
	}
}

func TestLifecycleMatchCarryFreshIdentityAndTiedRanks(t *testing.T) {
	s, _ := NewState(3, "match/one")
	s.Order = []int{1, 2, 3}
	m, e := NewMatchLifecycle(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.Ledger, e = m.Game.Ledger.Charge("old", 0, -1, IntAmount(7))
	if e != nil {
		t.Fatal(e)
	}
	m.Game, e = m.Game.closeOrdinary("round-limit", 0)
	if e != nil {
		t.Fatal(e)
	}
	for m.Game.Automatic != nil {
		m.Game, e = m.Game.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	for seat := 1; seat <= 3; seat++ {
		m.Game, e = m.Game.FinishSettlement(seat)
		if e != nil {
			t.Fatal(e)
		}
	}
	m.Game, e = m.Game.Finalize()
	if e != nil {
		t.Fatal(e)
	}
	if got := m.Ranks(); len(got) != 3 || got[0] != 3 || got[1] != 1 || got[2] != 1 {
		t.Fatalf("tied competition ranks %v", got)
	}
	n, e := m.NextGame("match/two")
	if e != nil {
		t.Fatal(e)
	}
	if n.Game.Board.GameID != "match/two" || len(n.Game.Board.DrawOrder) != 104 || n.Game.Ledger.Debts[0].Remaining.Cmp(IntAmount(7)) != 0 || len(n.Game.Ledger.Completed) != 1 {
		t.Fatal("next game lost persistent debt/results or failed fresh cards")
	}
	if _, e = m.NextGame("match/one"); e == nil {
		t.Fatal("reused identity")
	}
	for _, cash := range n.Game.Ledger.Cash {
		if cash.Sign() != 0 {
			t.Fatal("cash leaked into next game")
		}
	}
}

func TestLifecycleCoupAndDepartureAndNullification(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	l.Board, e = l.Board.Move([]string{"deck-1-diamonds-12", "deck-2-diamonds-12", "deck-1-hearts-12", "deck-2-hearts-12"}, 1, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	l.Board.Players[0].History = []Suit{Clubs}
	l.Board.Players[0].Confined = true
	l.Ledger.Scores[1] = IntAmount(20)
	l.Ledger.Cash[1] = IntAmount(20)
	n, e := l.DeclareCoup(Command{GameID: l.Board.GameID, ID: "coup", Kind: "coup", Actor: 1})
	if e != nil {
		t.Fatal(e)
	}
	for n.Automatic != nil {
		n, e = n.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Ledger.Scores[0].Cmp(IntAmount(50)) != 0 || n.Ledger.Scores[1].Cmp(IntAmount(-50)) != 0 || n.Ledger.Cash[1].Sign() != 0 {
		t.Fatal("Coup replacement")
	}
	if n.Ledger.Debts[0].Remaining.Cmp(IntAmount(50)) != 0 {
		t.Fatal("old cash incorrectly reduced new debt")
	}
	l = lifecycleFixture(t)
	l.RoundClosed = true
	l.TurnIndex = 2
	l.Board.Active = 3
	l.Ledger.Scores[1] = IntAmount(-3)
	n, e = l.DepartSeats([]int{1, 2})
	if e != nil {
		t.Fatal(e)
	}
	if n.Ending != "departures" || !n.Board.Players[0].Departed || n.Ledger.Scores[1].Cmp(IntAmount(-3)) != 0 {
		t.Fatal("simultaneous departure result")
	}
	m, e := NewMatchLifecycle(l.Board, 3)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.Ledger, e = m.Game.Ledger.Charge("void-debt", 0, -1, IntAmount(5))
	if e != nil {
		t.Fatal(e)
	}
	m, e = m.Nullify([]int{1, 2, 3})
	if e != nil {
		t.Fatal(e)
	}
	if len(m.Game.Ledger.Debts) != 0 || m.Game.Ledger.Phase != "void" {
		t.Fatal("nullification did not restore snapshot")
	}
	m, e = m.NextGame("replacement")
	if e != nil {
		t.Fatal(e)
	}
	if len(m.Game.Ledger.Completed) != 0 || len(m.Instances) != 2 {
		t.Fatal("nullification consumed match slot")
	}
}

func TestLifecycleCyclicOrdinarySettlementResumes(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	l.Ledger, e = l.Ledger.Charge("ab", 0, 1, IntAmount(1))
	if e != nil {
		t.Fatal(e)
	}
	l.Ledger, e = l.Ledger.Charge("ba", 1, 0, IntAmount(1))
	if e != nil {
		t.Fatal(e)
	}
	l.Board, e = l.Board.Move([]string{"deck-1-diamonds-02"}, 1, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	l, e = l.closeOrdinary("round-limit", 0)
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(l)
	n, e := l.ResumeAutomatic(1)
	if e != nil {
		t.Fatal(e)
	}
	if Digest(l) != before {
		t.Fatal("resume mutated checkpoint")
	}
	for n.Automatic != nil {
		n, e = n.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Ledger.Cash[0].Cmp(IntAmount(7)) != 0 || n.Ledger.Debts[0].Remaining.Sign() != 0 || n.Ledger.Debts[1].Remaining.Sign() != 0 {
		t.Fatal("cycle obligations lost")
	}
}

func TestLifecycleRicherCountsAllAttachedKings(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	l.Board, e = l.Board.Move([]string{"deck-1-hearts-02"}, 1, Series, true)
	if e != nil {
		t.Fatal(e)
	}
	l.Board, e = l.Board.Move([]string{"deck-1-spades-13", "deck-1-diamonds-13"}, 1, Attachment, true)
	if e != nil {
		t.Fatal(e)
	}
	for i := range l.Board.Cards {
		if l.Board.Cards[i].Zone == Attachment {
			l.Board.Cards[i].Allocation = string(Hearts)
		}
	}
	a, e := roundReceipts(l.Board)
	if e != nil {
		t.Fatal(e)
	}
	if a[0].Amount.Sign() != 0 {
		t.Fatal("two kings in series disables Richer, even when one is diamond")
	}
}
func TestLifecycleDeclarationPublicProof(t *testing.T) {
	l := lifecycleFixture(t)
	ids := []string{}
	for _, c := range Deck() {
		if c.Suit == Diamonds && c.Rank >= 2 && c.Rank <= 10 {
			ids = append(ids, c.ID)
		}
	}
	var e error
	l.Board, e = l.Board.Move(ids, 1, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	l.Board.Players[0].History = []Suit{Clubs, Hearts, Diamonds, Spades}
	n, e := l.DeclareOrdinary(1, "diamonds")
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Board.PublicHistory) != 13 {
		t.Fatalf("sufficient proof only: %d", len(n.Board.PublicHistory))
	}
}

func TestLifecycleExplicitFinancialFinalization(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	l, e = l.closeOrdinary("round-limit", 0)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = l.FinishSettlement(1); e == nil {
		t.Fatal("automatic settlement precedes discretionary decisions")
	}
	for l.Automatic != nil {
		l, e = l.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if _, e = l.Finalize(); e == nil {
		t.Fatal("silence is not finish")
	}
	for seat := 1; seat <= 3; seat++ {
		l, e = l.FinishSettlement(seat)
		if e != nil {
			t.Fatal(e)
		}
	}
	l, e = l.Finalize()
	if e != nil {
		t.Fatal(e)
	}
	if len(l.Ledger.Completed) != 1 || l.Ledger.Phase != "finalized" {
		t.Fatal("missing immutable final result")
	}
}

func TestLifecycleOperationRetryDoesNotDuplicateDraw(t *testing.T) {
	l := lifecycleFixture(t)
	op := LifecycleOperation{ID: "draw-one", GameID: l.Board.GameID, Kind: "begin-turn", Actor: 1}
	n, e := l.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(n)
	twice, e := n.ApplyOperation(op)
	if e != nil {
		t.Fatal(e)
	}
	if Digest(twice) != before {
		t.Fatal("retry duplicated draw or mutated state")
	}
	op.Kind = "end-turn"
	if _, e = n.ApplyOperation(op); e == nil {
		t.Fatal("conflicting receipt accepted")
	}
}

func TestLifecycleConfinementCannotSkipRoundThirteenSettlement(t *testing.T) {
	s := combatFixture(t)
	s.Order = []int{3, 2, 1}
	s.Round = 13
	var e error
	s, e = s.Move([]string{"deck-1-spades-03"}, 2, Series, true)
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce, true)
	if e != nil {
		t.Fatal(e)
	}
	l, e := NewLifecycle(s, NewFinancialLedger(s.GameID, 3))
	if e != nil {
		t.Fatal(e)
	}
	l.TurnIndex = 2
	l.TurnStarted = true
	l, _, e = l.ApplyBoardCommand(Command{GameID: s.GameID, ID: "attack-last", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	l, _, e = l.ApplyBoardCommand(Command{GameID: s.GameID, ID: "confine-last", Kind: "confinement", Actor: 2, WindowID: "attack-last", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	if l.Board.Round != 13 || !l.RoundClosed || l.Automatic == nil || l.Board.Pending != nil {
		t.Fatal("Confinement skipped round closure")
	}
	for l.Automatic != nil {
		l, e = l.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if l.Ending != "round-limit" {
		t.Fatal("Confinement created round fourteen")
	}
}

func TestLifecycleCoupPreservesEarnedReceiptsBeforeAbandoningDraws(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	l.Board, e = l.Board.Move([]string{"deck-1-diamonds-12", "deck-2-diamonds-12", "deck-1-hearts-12", "deck-2-hearts-12"}, 1, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	l.Board.Players[0].History = []Suit{Clubs}
	l.Ledger, e = l.Ledger.Charge("earned-before-coup", 0, 1, IntAmount(1))
	if e != nil {
		t.Fatal(e)
	}
	l.PendingReceipts = []FinancialReceipt{{Seat: 0, Amount: IntAmount(3)}}
	l.Board.DrawQueue = []DrawEntitlement{{Seat: 1, Remaining: 2}}
	n, e := l.DeclareCoup(Command{GameID: l.Board.GameID, ID: "immediate", Kind: "coup", Actor: 1})
	if e != nil {
		t.Fatal(e)
	}
	if n.Board.Phase != "settlement" || len(n.Board.DrawQueue) != 0 {
		t.Fatal("Coup must close board before accounting continuation")
	}
	for n.Automatic != nil {
		n, e = n.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Ledger.Cash[0].Cmp(IntAmount(50)) != 0 || n.Ledger.Debts[0].Remaining.Sign() != 0 {
		t.Fatal("completed award's repayment was discarded by Coup")
	}
}

func TestLifecyclePonziExactAlliedShares(t *testing.T) {
	l := lifecycleFixture(t)
	l.TurnStarted = true
	var e error
	for _, v := range []struct {
		ids  []string
		seat int
		zone Zone
	}{{[]string{"deck-1-clubs-02", "deck-1-diamonds-02"}, 1, Series}, {[]string{"deck-1-clubs-11", "deck-2-clubs-11", "deck-1-spades-11", "deck-2-spades-11"}, 1, Formation}} {
		l.Board, e = l.Board.Move(v.ids, v.seat, v.zone, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	for i := range l.Board.Cards {
		if l.Board.Cards[i].Zone == Formation {
			l.Board.Cards[i].Allocation = "ponzi"
		}
	}
	l.Board.Players[0].Ally = 3
	l.Board.Players[2].Ally = 1
	l.Board.Players[1].History = []Suit{Diamonds, Hearts}
	l.Ledger.Scores[1] = IntAmount(11)
	l.Ledger.Cash[1] = IntAmount(11)
	l, _, e = l.ApplyBoardCommand(Command{GameID: l.Board.GameID, ID: "steal", Kind: "ponzi", Actor: 1, Value: 2})
	if e != nil {
		t.Fatal(e)
	}
	for _, seat := range []int{2, 3} {
		l, _, e = l.ApplyBoardCommand(Command{GameID: l.Board.GameID, ID: fmt.Sprintf("pass-%d", seat), Kind: "pass", Actor: seat, WindowID: "steal"})
		if e != nil {
			t.Fatal(e)
		}
	}
	l, e = l.ContinueAfterAction(l.Board.DrawOrder, nil)
	if e != nil {
		t.Fatal(e)
	}
	for l.Automatic != nil {
		l, e = l.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	half, _ := NewAmount("11", "2")
	if l.Ledger.Cash[0].Cmp(half) != 0 || l.Ledger.Cash[2].Cmp(half) != 0 || l.Ledger.Cash[1].Sign() != 0 {
		t.Fatal("Ponzi shares not direct exact halves")
	}
}

func TestLifecycleRejectsForgedPhaseAndRoundBoundary(t *testing.T) {
	l := lifecycleFixture(t)
	bad := l.Clone()
	bad.RoundClosed = true
	if bad.Validate() == nil {
		t.Fatal("unfinished round labeled closed")
	}
	bad = l.Clone()
	bad.Ledger.Phase = "settlement"
	if bad.Validate() == nil {
		t.Fatal("card and finance phases disagree")
	}
	bad = l.Clone()
	bad.CodeCursor = -1
	if bad.Validate() == nil {
		t.Fatal("negative Code cursor")
	}
}

func TestLifecycleRoundIncomePrecedesCodeChargeAndResets(t *testing.T) {
	l := lifecycleFixture(t)
	var e error
	for _, v := range []struct {
		ids  []string
		seat int
		zone Zone
	}{{[]string{"deck-1-clubs-02", "deck-1-diamonds-02"}, 1, Series}, {[]string{"deck-1-hearts-03", "deck-1-diamonds-03"}, 2, Series}, {[]string{"deck-1-spades-13"}, 2, Attachment}} {
		l.Board, e = l.Board.Move(v.ids, v.seat, v.zone, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	queens := []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-clubs-12", "deck-2-clubs-12", "deck-1-spades-12"}
	l.Board, e = l.Board.Move(queens, 1, Formation, true)
	if e != nil {
		t.Fatal(e)
	}
	for i := range l.Board.Cards {
		if l.Board.Cards[i].Zone == Formation {
			l.Board.Cards[i].Allocation = "governing"
		}
		if l.Board.Cards[i].Zone == Attachment {
			l.Board.Cards[i].Allocation = string(Diamonds)
		}
	}
	l.Board.Formations = []FormationRecord{{ID: "governing", Controller: 1, Spec: FormationSpec{Kind: "code", Cards: queens}}}
	l.Board.Code = &CodeSettings{Declarer: 1, Formation: "governing", DiamondBonus: 1, PurchasePrice: 5}
	l.Board.Players[1].History = []Suit{Hearts, Diamonds}
	l.TurnIndex = 2
	l.Board.Active = 3
	l.TurnStarted = true
	l, e = l.EndTurn()
	if e != nil {
		t.Fatal(e)
	}
	for l.Automatic != nil {
		l, e = l.ResumeAutomatic(1)
		if e != nil {
			t.Fatal(e)
		}
	}
	if l.Ledger.Scores[1].Cmp(IntAmount(-3)) != 0 || len(l.Ledger.Debts) != 1 || l.Ledger.Debts[0].Remaining.Cmp(IntAmount(3)) != 0 {
		t.Fatal("Richer 2 then Code5 should charge3 unpaid exactly once")
	}
	l, e = l.BeginRound(nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	if l.CodeCursor != 0 || l.Board.Round != 2 {
		t.Fatal("round scoped Code cursor not reset")
	}
}
