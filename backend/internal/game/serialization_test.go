package game

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"strings"
	"testing"
)

func TestApplyRejectsUnserializableAccumulatedState(t *testing.T) {
	s := combatFixture(t)
	s.Commands["historical/"+strings.Repeat("x", canonical.MaxBytes-60000)] = strings.Repeat("a", 64)
	if _, e := canonical.Marshal(s); e != nil {
		t.Fatal("prestate must fit protocol", e)
	}
	c := Command{GameID: s.GameID, ID: "new/" + strings.Repeat("y", 40000), Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}}
	if _, e := canonical.Marshal(c); e != nil {
		t.Fatal("command must fit protocol", e)
	}
	before, e := canonical.Hash(s)
	if e != nil {
		t.Fatal(e)
	}
	n, events, e := Apply(s, c)
	if !errors.Is(e, ErrStateBudget) {
		t.Fatal("accepted transition beyond serialized state budget or wrong error", e)
	}
	if len(events) != 0 {
		t.Fatal("rejected transition published events")
	}
	after, err := canonical.Hash(n)
	if err != nil || after != before {
		t.Fatal("budget rejection did not retain committed state", err)
	}
	card, _ := n.Card("deck-1-clubs-08")
	if card.UsedTurn != 0 {
		t.Fatal("budget rejection consumed Club")
	}
}

func TestApplyRetainsStateAtCanonicalIntegerBoundary(t *testing.T) {
	s := combatFixture(t)
	s.Version = 9007199254740991
	before, e := canonical.Hash(s)
	if e != nil {
		t.Fatal(e)
	}
	n, events, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if !errors.Is(e, ErrStateBudget) || len(events) != 0 {
		t.Fatal("integer-domain overflow must stop before commitment", e)
	}
	after, e := canonical.Hash(n)
	if e != nil || after != before {
		t.Fatal("lost canonical integer checkpoint")
	}
}
