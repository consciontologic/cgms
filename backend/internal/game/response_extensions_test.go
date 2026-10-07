package game

import "testing"

func TestCompensationResponseCountsOnlyLaterAttackLosses(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-spades-02"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-hearts-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "a", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "comp", Kind: "compensation-response", Actor: 2, WindowID: "a", AceID: "deck-1-hearts-01"})
	if e != nil {
		t.Fatal(e)
	}
	if !n.Players[1].CompensationUsed || n.Players[1].AceRound != 1 {
		t.Fatal("allowances not consumed")
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "pass", Kind: "pass", Actor: 3, WindowID: "a"})
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Effects) != 1 || n.Effects[0].Losses != 1 {
		t.Fatal("Compensation did not count later loss")
	}
}
