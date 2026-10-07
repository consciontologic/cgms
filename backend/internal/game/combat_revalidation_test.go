package game

import "testing"

func TestCombatRevalidatesTwoSeriesAfterDexterPayment(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-diamonds-11"}, 1, Attachment, true)
	s, _ = s.Move([]string{"deck-1-spades-02"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "confine", Kind: "confinement", Actor: 2, WindowID: "attack", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "prevent", Kind: "decision", Actor: 1, WindowID: "attack", DecisionID: "confine/dexter", Cards: []string{"deck-1-diamonds-02"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "pass", Kind: "pass", Actor: 3, WindowID: "attack"})
	if e != nil {
		t.Fatal(e)
	}
	target, _ := n.Card("deck-1-hearts-08")
	club, _ := n.Card("deck-1-clubs-08")
	payment, _ := n.Card("deck-1-diamonds-02")
	ace, _ := n.Card("deck-1-diamonds-01")
	if target.Zone != Series || target.Controller != 2 {
		t.Fatal("attack resolved after losing two-series eligibility")
	}
	if club.UsedTurn != s.Turn || club.Zone != Series || payment.Zone != Draw || ace.Zone != Discard || n.Players[0].Quotas["dexter"] != s.Round {
		t.Fatal("invalidation undid paid costs/usage")
	}
}
func TestJusticeRequiresClubSupportWithUnderground(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-08"}, 1, Hand, true)
	s, _ = s.Move([]string{"deck-1-hearts-02"}, 1, Series, true)
	queens := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}
	s, _ = s.Move(queens, 1, Formation, true)
	for i := range s.Cards {
		for _, id := range queens {
			if s.Cards[i].Card.ID == id {
				s.Cards[i].Allocation = "justice"
			}
		}
	}
	cmd := Command{GameID: s.GameID, ID: "justice", Kind: "justice", Actor: 1, AceID: queens[0]}
	before := Digest(s)
	n, _, e := Apply(s, cmd)
	if e == nil || Digest(n) != before {
		t.Fatal("unsupported offensive formation accepted without Clubs")
	}
	s.Players[0].Underground = true
	if _, _, e = Apply(s, cmd); e == nil {
		t.Fatal("Underground must not waive Clubs support under 1.6/2.1")
	}
	s, _ = s.Move([]string{"deck-1-clubs-08"}, 1, Series, true)
	if _, _, e = Apply(s, cmd); e != nil {
		t.Fatal("supported Justice rejected", e)
	}
}

func TestJusticeUndergroundCannotWaiveClubs(t *testing.T) {
	s, _ := NewState(3, "justice-clubs")
	s.Players[0].Underground = true
	for _, id := range []string{"deck-1-hearts-02", "deck-1-diamonds-02"} {
		s, _ = s.Move([]string{id}, 1, Series, true)
	}
	for _, id := range []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"} {
		s, _ = s.Move([]string{id}, 1, Formation, true)
		for i := range s.Cards {
			if s.Cards[i].Card.ID == id {
				s.Cards[i].Allocation = "justice"
			}
		}
	}
	_, _, e := Apply(s, Command{GameID: s.GameID, ID: "j", Actor: 1, Kind: "justice", AceID: "deck-1-hearts-12"})
	if e == nil {
		t.Fatal("Underground waived offensive Clubs support")
	}
}
