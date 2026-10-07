package game

import "testing"

func TestDoubleDeckPhysicalIdentity(t *testing.T) {
	cards := Deck()
	if len(cards) != 104 {
		t.Fatalf("physical inventory = %d, want 104", len(cards))
	}
	ids := map[string]bool{}
	types := map[[2]int]int{}
	suits := map[Suit]int{Hearts: 0, Clubs: 1, Spades: 2, Diamonds: 3}
	for _, c := range cards {
		if ids[c.ID] || c.ID == "" {
			t.Fatalf("duplicate/empty physical ID %q", c.ID)
		}
		ids[c.ID] = true
		if c.Deck < 1 || c.Deck > 2 || c.Rank < 1 || c.Rank > 13 {
			t.Fatalf("invalid card: %+v", c)
		}
		s, ok := suits[c.Suit]
		if !ok {
			t.Fatal("unknown suit")
		}
		types[[2]int{s, c.Rank}]++
	}
	if len(types) != 52 {
		t.Fatal("missing rank/suit")
	}
	for k, n := range types {
		if n != 2 {
			t.Fatalf("%v has %d copies", k, n)
		}
	}
}
