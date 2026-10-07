package game

import (
	"fmt"
	"testing"
)

func completeF07Loan(t *testing.T, s State, ids []string) State {
	t.Helper()
	apply := func(c Command) {
		t.Helper()
		c.GameID = s.GameID
		var e error
		s, _, e = Apply(s, c)
		if e != nil {
			t.Fatal(e)
		}
	}
	apply(Command{ID: "f07offer", Actor: 1, Kind: "offer", OfferID: "f07", Revision: 1, Terms: &ProposalTerms{To: 2, Give: ids, Loan: true}})
	apply(Command{ID: "f07accept", Actor: 2, Kind: "accept-offer", OfferID: "f07", Revision: 1})
	for s.Pending != nil {
		seat := s.Pending.Responders[s.Pending.Cursor]
		apply(Command{ID: fmt.Sprintf("f07pass%d", seat), Actor: seat, Kind: "pass", WindowID: "f07accept"})
	}
	return s
}
func f07Round(t *testing.T, s State) Lifecycle {
	t.Helper()
	s.Order = []int{}
	for _, p := range s.Players {
		s.Order = append(s.Order, p.Seat)
	}
	s.Active = len(s.Players)
	l, e := NewLifecycle(s, NewFinancialLedger(s.GameID, len(s.Players)))
	if e != nil {
		t.Fatal(e)
	}
	l.TurnIndex = len(s.Players) - 1
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
	return l
}

func TestUndergroundLoanBreakReturnKeepsPrivilegeAndAllocatedIncome(t *testing.T) {
	s, _ := NewState(3, "f07-underground")
	kings := []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-clubs-13", "deck-2-clubs-13", "deck-1-spades-13"}
	var e error
	s, e = s.Move(kings, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	s, e = OpenFormation(s, 1, "ug", FormationSpec{Kind: "underground", Cards: kings})
	if e != nil {
		t.Fatal(e)
	}
	// Extra hidden/unassigned kings do not count toward the five allocated kings.
	s, e = s.Move([]string{"deck-2-spades-13", "deck-1-diamonds-13"}, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	s = completeF07Loan(t, s, kings)
	if !s.Players[0].Underground || !s.Players[1].Underground {
		t.Fatal("earner or borrower privilege missing")
	}
	receipts, e := roundReceipts(s)
	if e != nil {
		t.Fatal(e)
	}
	if receipts[0].Amount.Sign() != 0 || receipts[1].Amount.Cmp(IntAmount(10)) != 0 {
		t.Fatal("income followed privilege instead of allocated functioning kings")
	}
	returned, e := s.ReturnLoan(kings, 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	receipts, e = roundReceipts(returned)
	if e != nil {
		t.Fatal(e)
	}
	if receipts[0].Amount.Cmp(IntAmount(10)) != 0 || receipts[1].Amount.Sign() != 0 || !returned.Players[1].Underground {
		t.Fatal("intact return lost lasting borrower privilege or moved income incorrectly")
	}
	broken, e := s.Move(kings[:1], 0, Discard, true)
	if e != nil {
		t.Fatal(e)
	}
	receipts, e = roundReceipts(broken)
	if e != nil {
		t.Fatal(e)
	}
	if receipts[1].Amount.Sign() != 0 || !broken.Players[1].Underground || !broken.Players[0].Underground {
		t.Fatal("broken formation still paid or revoked privilege")
	}
	broken, e = broken.ReturnLoan(kings[1:], 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	receipts, e = roundReceipts(broken)
	if e != nil {
		t.Fatal(e)
	}
	if receipts[0].Amount.Sign() != 0 || !broken.Players[0].Underground || !broken.Players[1].Underground {
		t.Fatal("partial surviving return auto-reformed income or revoked privilege")
	}
}

func codeF07Fixture(t *testing.T) (State, []string) {
	t.Helper()
	s, _ := NewState(4, "f07-code")
	var e error
	for seat := 1; seat <= 4; seat++ {
		deck := 1
		if seat > 2 {
			deck = 2
		}
		rank := seat + 2
		ids := []string{fmt.Sprintf("deck-%d-clubs-%02d", deck, rank), fmt.Sprintf("deck-%d-diamonds-%02d", deck, rank)}
		s, e = s.Move(ids, seat, Series, false)
		if e != nil {
			t.Fatal(e)
		}
		s.Players[seat-1].History = []Suit{Clubs, Diamonds}
	}
	queens := []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-clubs-12", "deck-2-clubs-12", "deck-1-spades-12"}
	s, e = s.Move(queens, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	s, e = OpenFormation(s, 1, "code", FormationSpec{Kind: "code", Cards: queens})
	if e != nil {
		t.Fatal(e)
	}
	n := readyAbility(t, s, Command{ID: "code-initial", Actor: 1, Kind: "code", FormationID: "code", Value: 1, Price: 8})
	s, _, e = ResolveAbility(n, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	return s, queens
}

func TestCodeLoanBreakReturnRedeclarationPreservesOrReplacesDeclarer(t *testing.T) {
	s, queens := codeF07Fixture(t)
	loan := completeF07Loan(t, s, queens)
	after := f07Round(t, loan)
	if len(after.Ledger.Debts) != 0 || loan.Code.Declarer != 1 || loan.Code.DiamondBonus != 1 || loan.Code.PurchasePrice != 8 {
		t.Fatal("loan moved governing exemption or kept old recurring penalty")
	}
	awards := ordinaryReceipts(loan, 0)
	if awards[0].Amount.Cmp(IntAmount(8)) != 0 || awards[1].Amount.Cmp(IntAmount(15)) != 0 {
		t.Fatal("Code exemption must remain with declarer; borrower owns Spade queen10 plus Diamond4+1")
	}
	returned, e := loan.ReturnLoan(queens, 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	after = f07Round(t, returned)
	if len(after.Ledger.Debts) != 3 {
		t.Fatal("intact return did not restore declarer penalty")
	}
	broken, e := returned.Move(queens[:1], 0, Discard, true)
	if e != nil {
		t.Fatal(e)
	}
	after = f07Round(t, broken)
	if len(after.Ledger.Debts) != 0 || broken.Code.Declarer != 1 {
		t.Fatal("broken Code still penalized or forgot governing settings")
	}
	loan.Active = 2
	pending := readyAbility(t, loan, Command{ID: "code-replace", Actor: 2, Kind: "code", FormationID: "code", Value: 10, Price: 5})
	replaced, _, e := ResolveAbility(pending, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	if replaced.Code.Declarer != 2 || replaced.Code.DiamondBonus != 10 || replaced.Code.PurchasePrice != 5 {
		t.Fatal("legal redeclaration failed to replace governing owner/settings")
	}
	after = f07Round(t, replaced)
	if len(after.Ledger.Debts) != 3 || after.Ledger.Scores[1].Sign() != 0 {
		t.Fatal("new declarer penalty exemption missing")
	}
}

func TestCodePenaltyMixedProtectionConfinementAndUndergroundSupport(t *testing.T) {
	s, _ := codeF07Fixture(t)
	s.Players[1].History = []Suit{Clubs}
	s.Players[2].Confined = true
	got := f07Round(t, s)
	if len(got.Ledger.Debts) != 1 || got.Ledger.Debts[0].Debtor != 3 || got.Ledger.Scores[3].Cmp(IntAmount(-5)) != 0 {
		t.Fatal("protected/confined/declarer penalty exclusions")
	}
	s.Players[0].Underground = true
	// Earned Underground waives declaration support, never hostile penalty's two series.
	var e error
	s, e = s.Move([]string{"deck-1-diamonds-03"}, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	got = f07Round(t, s)
	if len(got.Ledger.Debts) != 0 {
		t.Fatal("Underground incorrectly waived penalty two-series prerequisite")
	}
}
