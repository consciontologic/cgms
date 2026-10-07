package game

import "testing"

func TestAcceptedAbilityCancellationKeepsCostsAndQuota(t *testing.T) {
	for _, kind := range []string{"ringleader", "richer-sacrifice", "barricade-sacrifice", "compensation", "main-inflation", "fate", "code", "kidnapper", "dexter-assassination"} {
		t.Run(kind, func(t *testing.T) {
			s, c := abilityTimingFixture(t, kind)
			before := s.Clone()
			n, _, e := Apply(s, c)
			if e != nil {
				t.Fatal(e)
			}
			quota := Digest(n.Players[0].Quotas)
			aceRound := n.Players[0].AceRound
			compUsed := n.Players[0].CompensationUsed
			n, _, e = Apply(n, Command{ID: "cancel", GameID: s.GameID, Kind: "confinement", Actor: 2, WindowID: c.ID, AceID: "deck-1-diamonds-01"})
			if e != nil {
				t.Fatal(e)
			}
			if !n.Players[0].Confined || Digest(n.Players[0].Quotas) != quota || n.Players[0].AceRound != aceRound || n.Players[0].CompensationUsed != compUsed {
				t.Fatal("cancellation refunded accepted allowance")
			}
			for _, old := range before.Cards {
				if old.Controller == 1 || old.Card.ID == "deck-2-spades-13" {
					now, _ := n.Card(old.Card.ID)
					if Digest(now) != Digest(old) {
						t.Fatalf("cancellation charged or moved %s", old.Card.ID)
					}
				}
			}
			if len(n.Effects) != 0 || n.Code != nil {
				t.Fatal("canceled effect applied")
			}
			if e = n.Validate(); e != nil {
				t.Fatal(e)
			}
		})
	}
}

func TestSubstitutedKidnapperReturnsEveryPhysicalCostAndPreservesOwnBinding(t *testing.T) {
	s := formationFixture(t)
	var e error
	s, e = s.Move([]string{"deck-2-hearts-02", "deck-2-clubs-02"}, 2, Series, false)
	if e != nil {
		t.Fatal(e)
	}
	s.Players[1].History = []Suit{Hearts, Clubs}
	own := []string{"deck-1-clubs-11", "deck-1-clubs-13", "deck-1-spades-13"}
	s, e = s.Move(own, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	s, e = OpenFormation(s, 1, "sub-kid", FormationSpec{Kind: "kidnapper", Cards: own, Substitute: &FormationSubstitution{Kings: own[1:], Slot: FormationSlot{11, Clubs}}})
	if e != nil {
		t.Fatal(e)
	}
	victim := []string{"deck-1-hearts-12", "deck-2-clubs-13", "deck-2-spades-13"}
	s.Active = 2
	s, e = s.Move(victim, 2, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	s, e = OpenFormation(s, 2, "victim-people", FormationSpec{Kind: "people", Cards: victim, Substitute: &FormationSubstitution{Kings: victim[1:], Slot: FormationSlot{12, Hearts}}, Protection: &FormationProtection{Series: Clubs}})
	if e != nil {
		t.Fatal(e)
	}
	s.Active = 1
	s.Order = []int{1, 2, 3}
	l, e := NewLifecycle(s, NewFinancialLedger(s.GameID, 3))
	if e != nil {
		t.Fatal(e)
	}
	l.TurnStarted = true
	apply := func(c Command) {
		t.Helper()
		c.GameID = s.GameID
		var err error
		l, _, err = l.ApplyBoardCommand(c)
		if err != nil {
			t.Fatal(err)
		}
	}
	apply(Command{ID: "special", Actor: 1, Kind: "kidnapper", FormationID: "sub-kid", TargetSeat: 2, Targets: victim[1:2]})
	apply(Command{ID: "pass2", Actor: 2, Kind: "pass", WindowID: "special"})
	apply(Command{ID: "pass3", Actor: 3, Kind: "pass", WindowID: "special"})
	apply(Command{ID: "outcome", Actor: 1, Kind: "kidnapper-outcome", WindowID: "special", Cards: victim[1:2]})
	for _, id := range own {
		c, _ := l.Board.Card(id)
		if c.Zone != Draw || c.Controller != 0 {
			t.Fatalf("unreturned physical cost %s", id)
		}
	}
	for _, id := range victim[1:] {
		c, _ := l.Board.Card(id)
		if c.Controller != 1 || c.Zone != Unassigned {
			t.Fatal("special theft did not expose both kings")
		}
	}
	if len(l.Board.Bindings) != 1 || l.Board.Bindings[0].Formation != "sub-kid" {
		t.Fatal("own returned binding lost or stolen binding retained")
	}
	if len(l.Ledger.Debts) != 1 || l.Ledger.Debts[0].Debtor != 1 || l.Ledger.Debts[0].Creditor != 0 || l.Ledger.Debts[0].Remaining.Cmp(IntAmount(10)) != 0 || l.Ledger.Scores[1].Cmp(IntAmount(-10)) != 0 {
		t.Fatal("special theft ten-point obligation missing")
	}
	if e = l.Board.Validate(); e != nil {
		t.Fatal("conservation/binding", e)
	}
}

func TestAcceptedTransferCancellationKeepsAceUseAndNoAlliance(t *testing.T) {
	for _, loan := range []bool{false, true} {
		name := "ace-trade"
		if loan {
			name = "formation-loan"
		}
		t.Run(name, func(t *testing.T) {
			s := formationFixture(t)
			var e error
			ids := []string{"deck-1-hearts-01"}
			zone := ConcealedAce
			if loan {
				ids = []string{"deck-1-clubs-11", "deck-2-clubs-11"}
				zone = Hand
			}
			s, e = s.Move(ids, 1, zone, false)
			if e != nil {
				t.Fatal(e)
			}
			if loan {
				s, e = OpenFormation(s, 1, "loan", FormationSpec{Kind: "kidnapper", Cards: ids})
				if e != nil {
					t.Fatal(e)
				}
			}
			s, e = s.Move([]string{"deck-2-hearts-03", "deck-2-clubs-03"}, 3, Series, false)
			if e != nil {
				t.Fatal(e)
			}
			s, e = s.Move([]string{"deck-1-diamonds-01"}, 3, ConcealedAce, false)
			if e != nil {
				t.Fatal(e)
			}
			apply := func(c Command) {
				t.Helper()
				c.GameID = s.GameID
				var err error
				s, _, err = Apply(s, c)
				if err != nil {
					t.Fatal(err)
				}
			}
			apply(Command{ID: "offer", Actor: 1, Kind: "offer", OfferID: "terms", Revision: 1, Terms: &ProposalTerms{To: 2, Give: ids, Loan: loan}})
			apply(Command{ID: "accept", Actor: 2, Kind: "accept-offer", OfferID: "terms", Revision: 1})
			before := s.Clone()
			apply(Command{ID: "cancel", Actor: 3, Kind: "confinement", WindowID: "accept", AceID: "deck-1-diamonds-01"})
			for _, id := range ids {
				old, _ := before.Card(id)
				now, _ := s.Card(id)
				if Digest(old) != Digest(now) {
					t.Fatal("canceled transfer moved card")
				}
			}
			if s.Players[0].Ally != 0 || s.Players[1].Ally != 0 || len(s.Loans) != 0 || s.Proposals[0].Status != "canceled" {
				t.Fatal("cancellation created alliance/loan")
			}
			if !loan && s.Players[0].AceRound != s.Round {
				t.Fatal("Ace supplying player's committed quota refunded")
			}
		})
	}
}

func TestOpeningCancellationKeepsOpeningAllowance(t *testing.T) {
	s := formationFixture(t)
	s.Players[0].OpeningUsed = false
	var e error
	s, e = s.Move([]string{"deck-1-spades-05"}, 1, Hand, false)
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Move([]string{"deck-2-hearts-03", "deck-2-clubs-03"}, 2, Series, false)
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce, false)
	if e != nil {
		t.Fatal(e)
	}
	s, _, e = Apply(s, Command{ID: "open", GameID: s.GameID, Actor: 1, Kind: "open-series", Suit: Spades})
	if e != nil {
		t.Fatal(e)
	}
	s, _, e = Apply(s, Command{ID: "cancel", GameID: s.GameID, Actor: 2, Kind: "confinement", WindowID: "open", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	c, _ := s.Card("deck-1-spades-05")
	if !s.Players[0].OpeningUsed || c.Zone != Hand {
		t.Fatal("opening cancel refunded allowance or exposed card")
	}
}
