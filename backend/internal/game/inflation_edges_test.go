package game

import "testing"

func inflationFixture(t *testing.T) State {
	t.Helper()
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-spades-09", "deck-1-spades-10", "deck-2-spades-10"}, 1, Series, true)
	s, _ = s.Move([]string{"deck-2-spades-01"}, 3, ConcealedAce, true)
	var e error
	s, e = s.ActivateEffect(AceEffect{ID: "inflation", CardID: "deck-2-spades-01", Kind: "inflation", Source: 3, Target: 1, Custodian: 1, Expiry: "qualifying-spend"})
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestInflationNineteenTwentyAndNoCumulativeRemoval(t *testing.T) {
	s := inflationFixture(t)
	n, value, e := CompleteSpadeSpend(s, 1, []string{"deck-1-spades-09", "deck-1-spades-10"}, true, true)
	if e != nil || value.String() != "19/2" || len(n.Effects) != 1 {
		t.Fatal("nineteen must remain exact and active", value, e)
	}
	n, value, e = CompleteSpadeSpend(n, 1, []string{"deck-1-spades-09"}, true, true)
	if e != nil || value.String() != "9/2" {
		t.Fatal("odd value rounded", value, e)
	}
	n, value, e = CompleteSpadeSpend(n, 1, []string{"deck-1-spades-10"}, true, true)
	if e != nil || value.String() != "5" || len(n.Effects) != 1 {
		t.Fatal("smaller transactions accumulated")
	}
	n, value, e = CompleteSpadeSpend(n, 1, []string{"deck-1-spades-10", "deck-2-spades-10"}, true, true)
	if e != nil || value.String() != "10" || len(n.Effects) != 0 {
		t.Fatal("twenty not priced before removal")
	}
	_, value, e = CompleteSpadeSpend(n, 1, []string{"deck-1-spades-10"}, true, true)
	if e != nil || value.String() != "10" {
		t.Fatal("subsequent spend not full value")
	}
}
func TestDuplicateInflationResponseSpendsNothing(t *testing.T) {
	s := inflationFixture(t)
	s, _ = s.Move([]string{"deck-1-spades-02"}, 2, Series, true)
	s, _ = s.Move([]string{"deck-1-spades-01"}, 2, ConcealedAce, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(n)
	same, _, e := Apply(n, Command{GameID: s.GameID, ID: "duplicate", Kind: "inflation", Actor: 2, WindowID: "attack", AceID: "deck-1-spades-01"})
	if e == nil || Digest(same) != before || same.Players[1].AceRound == same.Round {
		t.Fatal("duplicate Inflation consumed an allowance")
	}
}
