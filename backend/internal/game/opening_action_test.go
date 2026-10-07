package game

import "testing"

func TestOpeningCommandWaitsAndConsumesAllowance(t *testing.T) {
	s, _ := NewState(3, "opening")
	s, _ = s.Move([]string{"deck-1-clubs-04"}, 1, Hand, true)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "open", Actor: 1, Kind: "open-series", Suit: Clubs})
	if e != nil {
		t.Fatal(e)
	}
	v, _ := n.Card("deck-1-clubs-04")
	if v.Zone != Hand || !n.Players[0].OpeningUsed {
		t.Fatal("opening acceptance lifecycle")
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Actor: seat, Kind: "pass", WindowID: "open"})
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Pending != nil || len(n.CurrentSeries(1)) != 1 || !n.Players[0].OpeningUsed {
		t.Fatal("opening resolution")
	}
}
