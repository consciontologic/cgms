package game

import (
	"fmt"
	"testing"
)

func reviewLoanFixture(t *testing.T) (State, []string) {
	t.Helper()
	s, _ := NewState(3, "return-game")
	ids := []string{"deck-1-hearts-11", "deck-2-hearts-11"}
	var e error
	s, e = s.Move(ids, 2, Formation, true)
	if e != nil {
		t.Fatal(e)
	}
	for i := range s.Cards {
		if s.Cards[i].Zone == Formation {
			s.Cards[i].Allocation = "loaned"
		}
	}
	s.Formations = []FormationRecord{{ID: "loaned", Controller: 2, Spec: FormationSpec{Kind: "kidnapper", Cards: ids}}}
	s.Loans = []FormationLoan{{Cards: ids, From: 1, To: 2, Allocation: "loaned"}}
	s.Players[0].Ally = 2
	s.Players[1].Ally = 1
	return s, ids
}
func TestLoanPartialReturnRetainsRemainingCustodyProvenance(t *testing.T) {
	s, ids := reviewLoanFixture(t)
	n, e := s.ReturnLoan(ids[:1], 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Loans) != 1 || len(n.Loans[0].Cards) != 1 || n.Loans[0].Cards[0] != ids[1] {
		t.Fatal("partial return erased remaining debt of custody")
	}
	n, e = n.ReturnLoan(ids[1:], 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Loans) != 0 {
		t.Fatal("fully returned custody remains")
	}
	for _, id := range ids {
		c, _ := n.Card(id)
		if c.Controller != 1 || c.Zone != Unassigned {
			t.Fatal("partial returns rebuilt a formation")
		}
	}
	if s.Players[0].Ally != 2 || n.Players[0].Ally != 2 {
		t.Fatal("return dissolved alliance")
	}
}
func TestLoanIntactReturnRestoresFormationController(t *testing.T) {
	s, ids := reviewLoanFixture(t)
	n, e := s.ReturnLoan(ids, 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	if n.Formations[0].Controller != 1 {
		t.Fatal("intact return left stale formation owner")
	}
}

func TestLoanReturnCommandHasResponseWindowAndRevalidation(t *testing.T) {
	s, ids := reviewLoanFixture(t)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "return", Kind: "return-loan", Actor: 2, TargetSeat: 1, Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card(ids[0])
	if c.Controller != 2 || n.Pending == nil {
		t.Fatal("return moved before response window")
	}
	for _, seat := range []int{3, 1} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: fmt.Sprintf("return-pass-%d", seat), Kind: "pass", Actor: seat, WindowID: "return"})
		if e != nil {
			t.Fatal(e)
		}
	}
	c, _ = n.Card(ids[0])
	if c.Controller != 1 || n.Pending != nil || len(n.Loans) != 0 {
		t.Fatal("resolved return failed")
	}
	s, ids = reviewLoanFixture(t)
	s.Active = 3
	if _, _, e = Apply(s, Command{GameID: s.GameID, ID: "offturn", Kind: "return-loan", Actor: 2, TargetSeat: 1, Cards: ids}); e == nil {
		t.Fatal("unrelated player's turn allowed return")
	}
}

func TestLoanReturnCancellationPreservesCustodyAndAlliance(t *testing.T) {
	s, ids := reviewLoanFixture(t)
	var e error
	s, e = s.Move([]string{"deck-1-clubs-03", "deck-1-diamonds-03"}, 3, Series, true)
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Move([]string{"deck-1-diamonds-01"}, 3, ConcealedAce, true)
	if e != nil {
		t.Fatal(e)
	}
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "return-cancel", Kind: "return-loan", Actor: 2, TargetSeat: 1, Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "cancel-return", Kind: "confinement", Actor: 3, WindowID: "return-cancel", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range ids {
		c, _ := n.Card(id)
		if c.Controller != 2 {
			t.Fatal("canceled return transferred custody")
		}
	}
	if len(n.Loans) != 1 || n.Players[0].Ally != 2 || n.Pending != nil || n.Players[2].AceRound != n.Round {
		t.Fatal("cancellation changed loan/alliance or refunded ace quota")
	}
}
