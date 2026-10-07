package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func developmentMenuState(players int) State {
	s, _ := NewState(players, "menu-development")
	for seat := 1; seat <= players; seat++ {
		ids := []string{}
		for _, su := range []Suit{Hearts, Clubs, Spades, Diamonds} {
			ids = append(ids, fmt.Sprintf("deck-%d-%s-%02d", 1+(seat-1)/2, su, 2+(seat-1)%2))
		}
		s, _ = s.Move(ids, seat, Series, false)
		s.Players[seat-1].History = []Suit{Hearts, Clubs, Spades, Diamonds}
	}
	ids := []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12", "deck-1-hearts-13", "deck-2-hearts-13", "deck-1-clubs-13", "deck-1-spades-13", "deck-1-diamonds-13", "deck-1-spades-04", "deck-1-spades-05", "deck-1-clubs-04", "deck-1-clubs-05"}
	s, _ = s.Move(ids, 1, Hand, false)
	for i := 0; i < 200; i++ {
		s.Commands[fmt.Sprintf("prior/%d", i)] = Digest(i)
	}
	return s
}
func BenchmarkGameplayMenuDevelopment(b *testing.B) {
	s := developmentMenuState(4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e := GameplayObservation(s, 1, 100000); e != nil {
			b.Fatal(e)
		}
	}
}

// Optional private development capture verifies byte-exact menus without exposing
// hidden-card fixtures in tool output. Paths must be deliberately supplied.
func TestGameplayMenuDevelopmentCapture(t *testing.T) {
	dir := os.Getenv("CGMS_MENU_CAPTURE")
	if dir == "" {
		return
	}
	for _, players := range []int{3, 4} {
		s := developmentMenuState(players)
		o, e := GameplayObservation(s, 1, 100000)
		if e != nil {
			t.Fatal(e)
		}
		data, e := json.Marshal(o)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dir, fmt.Sprintf("menu-%dp.json", players)), data, 0600); e != nil {
			t.Fatal(e)
		}
	}
}

// These fixtures contain no concealed Aces or attached attack modifiers. The v2
// expansion must preserve their entire v1 menu and change only its version tag.
func TestGameplayMenuOptimizationGolden(t *testing.T) {
	for players, want := range map[int]string{3: "b5ab5004b21236dee9a27933cfb8be41d80509759e7d151d35531a62507e84bf", 4: "5aeb67fc809e4ec236d24cf21c7c36450e87ef9bc0a979124dffb6d88ece4bbf"} {
		o, e := GameplayObservation(developmentMenuState(players), 1, 100000)
		if e != nil {
			t.Fatal(e)
		}
		data, e := json.Marshal(o)
		if e != nil {
			t.Fatal(e)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != want {
			t.Fatal("menu differs from reviewed v2 development observation")
		}
		o.MenuCoverage = "sampled-all-categories-v1"
		prior, err := json.Marshal(o)
		if err != nil {
			t.Fatal(err)
		}
		v1 := map[int]string{3: "9463fc17372ca355e9961055ca0008466562c7f4660313935001a2946e7a71e0", 4: "faf3958c2185b10fc18d3c9dd4eb60ae4ec3048889b155dd12ec239b0e33ed39"}
		if fmt.Sprintf("%x", sha256.Sum256(prior)) != v1[players] {
			t.Fatal("v2 changed legacy fixture beyond reviewed version tag")
		}
	}
}

func TestObserveWithoutMenuEquivalence(t *testing.T) {
	for _, players := range []int{3, 4} {
		state := developmentMenuState(players)
		for seat := 0; seat <= players+1; seat++ {
			before := Digest(state)
			full, oldErr := Observe(state, seat)
			projected, err := ObserveWithoutMenu(state, seat)
			full.Legal = nil
			if (oldErr == nil) != (err == nil) || Digest(full) != Digest(projected) {
				t.Fatalf("projection differs players=%d seat=%d", players, seat)
			}
			if before != Digest(state) {
				t.Fatal("projection mutated state")
			}
		}
	}
}

func TestGameplayDenseOrdinaryMenuFitsLocalBudget(t *testing.T) {
	s, err := NewState(4, "dense-local-menu")
	if err != nil {
		t.Fatal(err)
	}
	move := func(ids []string, seat int, zone Zone) {
		t.Helper()
		s, err = s.Move(ids, seat, zone, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	for rank := 2; rank <= 10; rank++ {
		move([]string{fmt.Sprintf("deck-1-clubs-%02d", rank)}, 1, Series)
		move([]string{fmt.Sprintf("deck-1-hearts-%02d", rank)}, 2, Series)
	}
	move([]string{"deck-1-spades-02"}, 1, Series)
	move([]string{"deck-2-clubs-02"}, 2, Series)
	for _, suit := range []Suit{Clubs, Hearts, Spades, Diamonds} {
		move([]string{fmt.Sprintf("deck-1-%s-01", suit)}, 1, ConcealedAce)
	}
	s.Players[0].History = []Suit{Clubs, Spades}
	s.Players[1].History = []Suit{Hearts, Clubs}
	before := Digest(s)
	complete, err := GameplayObservation(s, 1, 500000)
	if err != nil {
		t.Fatal(err)
	}
	local, err := GameplayObservation(s, 1, 50000)
	if err != nil {
		t.Fatalf("ordinary board exceeded fixed local budget despite only %d legal candidates: %v", len(complete.Legal), err)
	}
	if len(local.Legal) == 0 || Digest(local) != Digest(complete) || Digest(s) != before {
		t.Fatal("menu pruning dropped a legal candidate or mutated the board")
	}
}
