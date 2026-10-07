package game

import (
	"slices"
	"testing"
)

func TestPublicHistoryRetainsIdentityWithoutHiddenCustody(t *testing.T) {
	s := position(t)
	id := cid(Spades, 7)
	s, _ = s.Move([]string{id}, 2, Series, true)
	n, e := s.Move([]string{id}, 1, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	if !slices.Contains(n.PublicHistory, id) {
		t.Fatal("publicly captured identity forgotten")
	}
	o, _ := Observe(n, 3)
	if !slices.Contains(o.PublicHistory, id) {
		t.Fatal("observer lost public history")
	}
	for _, c := range o.Cards {
		if c.Card.ID == id {
			t.Fatal("historical identity disclosed hidden custody")
		}
	}
	o.PublicHistory[0] = "mutated"
	if slices.Contains(n.PublicHistory, "mutated") {
		t.Fatal("history alias")
	}
}
func TestHiddenReturnAndShuffleDoNotRevealIdentities(t *testing.T) {
	s := position(t)
	hidden := cid(Spades, 7)
	s, _ = s.Move([]string{hidden}, 2, Hand, true)
	n, e := s.Move([]string{hidden}, 0, Draw, true)
	if e != nil {
		t.Fatal(e)
	}
	if slices.Contains(n.PublicHistory, hidden) {
		t.Fatal("private return leaked face")
	}
	known := cid(Hearts, 8)
	n, _ = n.Move([]string{known}, 2, Series, true)
	n, _ = n.Move([]string{known}, 0, Draw, true)
	n, e = ShuffleSupply(n, n.DrawOrder)
	if e != nil {
		t.Fatal(e)
	}
	before, _ := Observe(n, 1)
	x, _ := n.Move([]string{hidden}, 2, Hand, true)
	y, _ := n.Move([]string{known}, 2, Hand, true)
	a, _ := Observe(x, 1)
	b, _ := Observe(y, 1)
	if Digest(a) != Digest(b) {
		t.Fatal("history tracks card through unseen draw")
	}
	if !slices.Contains(before.PublicHistory, known) {
		t.Fatal("shuffle erased historical knowledge")
	}
}
