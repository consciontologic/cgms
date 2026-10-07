package game

import (
	"fmt"
	"testing"
)

func TestRegisteredBindingLoanSupportAndReunion(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-1-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	spec := FormationSpec{Kind: "people", Cards: ids, Substitute: &FormationSubstitution{Kings: ids[1:], Slot: FormationSlot{12, Hearts}}, Protection: &FormationProtection{Series: Clubs}}
	var err error
	s, err = OpenFormation(s, 1, "registered-defense", spec)
	if err != nil {
		t.Fatal(err)
	}
	s, _ = s.Move([]string{"deck-2-hearts-02", "deck-2-diamonds-02"}, 2, Series, false)
	n, err := s.LoanFormation(ids, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if n.BindingStatus("formation/registered-defense") != "active" {
		t.Fatal("registered defensive loan with Hearts and two series must be active")
	}
	n, _ = n.Move([]string{"deck-2-hearts-02"}, 0, Draw, false)
	n.Players[1].Underground = true
	if n.BindingStatus("formation/registered-defense") != "dormant-together" {
		t.Fatal("Underground must not replace Hearts support")
	}
	n, _ = n.Move(ids[1:2], 3, Hand, false)
	if n.BindingStatus("formation/registered-defense") != "dormant-separated" {
		t.Fatal("split binding active")
	}
	n, _ = n.Move(ids[1:2], 2, Hand, false)
	if n.BindingStatus("formation/registered-defense") != "dormant-separated" {
		t.Fatal("same controller in different zones must remain dormant separated")
	}
	n, _ = n.Move(ids[2:], 2, Hand, false)
	if n.BindingStatus("formation/registered-defense") != "dormant-together" {
		t.Fatal("reunion must not automatically activate")
	}
}

func TestRegisteredOffensiveBindingUndergroundRequiresBothSupports(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	spec := FormationSpec{Kind: "justice", Cards: ids, Substitute: &FormationSubstitution{Kings: ids[3:], Slot: FormationSlot{12, Diamonds}}}
	var err error
	s, err = OpenFormation(s, 1, "registered-offense", spec)
	if err != nil {
		t.Fatal(err)
	}
	s, _ = s.Move([]string{"deck-2-hearts-02", "deck-2-diamonds-02"}, 2, Series, false)
	s.Players[1].Underground = true
	n, err := s.LoanFormation(ids, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if n.BindingStatus("formation/registered-offense") != "dormant-together" {
		t.Fatal("Underground cannot replace offensive Clubs support")
	}
	n, _ = n.Move([]string{"deck-2-clubs-02"}, 2, Series, false)
	if n.BindingStatus("formation/registered-offense") != "active" {
		t.Fatal("support restored to intact allocated binding")
	}
	n, _ = n.Move([]string{"deck-2-hearts-02"}, 0, Draw, false)
	if n.BindingStatus("formation/registered-offense") != "dormant-together" {
		t.Fatal("Doppelganger still requires Hearts")
	}
}

func TestPeopleProtectionRejectsChainsAndSelf(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-1-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	spec := FormationSpec{Kind: "people", Cards: ids, Substitute: &FormationSubstitution{Kings: ids[1:], Slot: FormationSlot{12, Hearts}}, Protection: &FormationProtection{Series: Clubs}}
	var err error
	s, err = OpenFormation(s, 1, "protected-people", spec)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"people", "great-people"} {
		candidate := FormationSpec{Kind: kind, Protection: &FormationProtection{Formation: "protected-people"}}
		if protectionValid(s, 1, "new-protector", candidate) {
			t.Fatal("People protection chain admitted", kind)
		}
		candidate.Protection.Formation = "new-protector"
		if protectionValid(s, 1, "new-protector", candidate) {
			t.Fatal("self protection admitted", kind)
		}
		candidate.Protection = &FormationProtection{Series: Hearts}
		if protectionValid(s, 1, "new-protector", candidate) {
			t.Fatal("Hearts protection admitted", kind)
		}
	}
}

func TestFourSeatDisjointAlliancesAndThirdPartnerRejection(t *testing.T) {
	s, _ := NewState(4, "alliance-pairs")
	first := []string{"deck-1-clubs-11", "deck-2-clubs-11"}
	second := []string{"deck-1-hearts-12", "deck-2-hearts-12"}
	for _, p := range []struct {
		seat     int
		ids      []string
		suit     Suit
		kind, id string
	}{{1, first, Clubs, "kidnapper", "first"}, {3, second, Hearts, "great-people", "second"}} {
		s.Active = p.seat
		s, _ = s.Move([]string{fmt.Sprintf("deck-%d-%s-02", 1+(p.seat-1)/2, p.suit), fmt.Sprintf("deck-%d-diamonds-02", 1+(p.seat-1)/2)}, p.seat, Series, false)
		s, _ = s.Move(p.ids, p.seat, Hand, false)
		spec := FormationSpec{Kind: p.kind, Cards: p.ids}
		if p.kind == "great-people" {
			spec.Protection = &FormationProtection{Series: Diamonds}
		}
		var err error
		s, err = OpenFormation(s, p.seat, p.id, spec)
		if err != nil {
			t.Fatal(err)
		}
	}
	loan := func(from, to int, ids []string, id string) {
		t.Helper()
		s.Active = from
		var err error
		s, _, err = Apply(s, Command{GameID: s.GameID, ID: id + "-offer", Kind: "offer", Actor: from, OfferID: id, Revision: 1, Terms: &ProposalTerms{To: to, Give: ids, Loan: true}})
		if err != nil {
			t.Fatal(err)
		}
		if s.Players[from-1].Ally != 0 {
			t.Fatal("offer created alliance")
		}
		s, _, err = Apply(s, Command{GameID: s.GameID, ID: id + "-accept", Kind: "accept-offer", Actor: to, OfferID: id, Revision: 1})
		if err != nil {
			t.Fatal(err)
		}
		if s.Players[from-1].Ally != 0 {
			t.Fatal("acceptance created early alliance")
		}
		for s.Pending != nil {
			actor := s.Pending.Responders[s.Pending.Cursor]
			s, _, err = Apply(s, Command{GameID: s.GameID, ID: fmt.Sprintf("%s-pass-%d", id, actor), Kind: "pass", Actor: actor, WindowID: id + "-accept"})
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	loan(1, 2, first, "pair-a")
	loan(3, 4, second, "pair-b")
	for i, want := range []int{2, 1, 4, 3} {
		if s.Players[i].Ally != want {
			t.Fatal("not two disjoint pairs")
		}
	}
	s.Active = 2
	before := Digest(s)
	bad, _, err := Apply(s, Command{GameID: s.GameID, ID: "third-partner", Kind: "offer", Actor: 2, OfferID: "third", Revision: 1, Terms: &ProposalTerms{To: 3, Give: first, Loan: true}})
	if err == nil || Digest(bad) != before {
		t.Fatal("third partner loan altered state or accepted")
	}
}
