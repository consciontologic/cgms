package game

import (
	"reflect"
	"slices"
	"testing"
)

func TestDealRedealAndInitialDiamonds(t *testing.T) {
	s := position(t)
	n, redo, e := DealAttempt(s, s.DrawOrder)
	if e != nil {
		t.Fatal(e)
	}
	if !redo || !reflect.DeepEqual(s, n) {
		t.Fatal("diamondless deal did not return all cards")
	}
	order := slices.Clone(s.DrawOrder)
	for seat := 0; seat < 3; seat++ {
		id := cid(Diamonds, seat+2)
		i := slices.Index(order, id)
		order[seat*20], order[i] = order[i], order[seat*20]
	}
	n, redo, e = DealAttempt(s, order)
	if e != nil || redo {
		t.Fatalf("valid deal %v %v", redo, e)
	}
	for seat := 1; seat <= 3; seat++ {
		count := 0
		diamonds := 0
		for _, c := range n.Cards {
			if c.Controller == seat {
				count++
				if c.Card.Suit == Diamonds && c.Card.Rank >= 2 && c.Card.Rank <= 10 {
					if c.Zone != Series {
						t.Fatal("initial diamond hidden")
					}
					diamonds++
				}
				if c.Card.Rank == 1 && c.Zone != ConcealedAce {
					t.Fatal("ace not separate")
				}
			}
		}
		if count != 20 || diamonds == 0 {
			t.Fatal("deal holdings")
		}
	}
	if e = n.Validate(); e != nil {
		t.Fatal(e)
	}
}
func TestOpeningHistoryAndAllowance(t *testing.T) {
	s := position(t)
	s, _ = s.Move([]string{cid(Diamonds, 2)}, 1, Series, true)
	s.Players[0].History = []Suit{Diamonds}
	s, _ = s.Move([]string{cid(Hearts, 2), cid(Hearts, 3), cid(Clubs, 2)}, 1, Hand, true)
	n, e := OpenSeries(s, 1, Hearts)
	if e != nil {
		t.Fatal(e)
	}
	if len(n.CurrentSeries(1)) != 2 || !n.Players[0].OpeningUsed {
		t.Fatal("opening did not expose entire suit")
	}
	if _, e = OpenSeries(n, 1, Clubs); e == nil {
		t.Fatal("second opening allowed")
	}
	n, _ = n.Move([]string{cid(Hearts, 2), cid(Hearts, 3)}, 0, Draw, true)
	n, _ = n.Move([]string{cid(Hearts, 4)}, 1, Hand, true)
	n, e = ShuffleSupply(n, n.DrawOrder)
	if e != nil {
		t.Fatal(e)
	}
	n, e = OpenSeries(n, 1, Hearts)
	if e != nil {
		t.Fatal("historical reopening used allowance", e)
	}
	if len(n.Players[0].History) != 2 {
		t.Fatal("history erased")
	}
}
func TestSupplyRecyclingAndExhaustion(t *testing.T) {
	s := position(t)
	ids := slices.Clone(s.DrawOrder)
	s, _ = s.Move(ids[1:], 0, Discard, true)
	id := ids[0]
	n, got, e := DrawCards(s, 1, 1, false, nil)
	if e != nil || len(got) != 1 || got[0] != id {
		t.Fatalf("draw before recycling %v %v", got, e)
	}
	if _, _, e = DrawCards(n, 1, 1, false, nil); e == nil {
		t.Fatal("unshuffled recycling accepted")
	}
	n, got, e = DrawCards(n, 1, 200, false, ids[1:])
	if e != nil || len(got) != 103 {
		t.Fatalf("mandatory exhausted draw %d %v", len(got), e)
	}
	_, got, e = DrawCards(n, 1, 1, false, nil)
	if e != nil || len(got) != 0 {
		t.Fatal("invented entitlement")
	}
}
func TestInitiativeExhaustionRotation(t *testing.T) {
	s := position(t)
	ids := slices.Clone(s.DrawOrder)
	s, _ = s.Move(ids, 1, Unassigned, true)
	s.Round = 2
	n, e := ResolveInitiative(s, nil, nil)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(n.Order, []int{2, 3, 1}) {
		t.Fatalf("rotation %v", n.Order)
	}
	n.Players[2].Departed = true
	n.Round = 3
	n, e = ResolveInitiative(n, nil, nil)
	if e != nil || !reflect.DeepEqual(n.Order, []int{1, 2}) {
		t.Fatalf("departed order %v %v", n.Order, e)
	}
}
func TestInitiativeNonNumbersRepeatedTieIsolationAndFrozenOrder(t *testing.T) {
	s := position(t)
	order := []string{cid(Hearts, 13), cid(Clubs, 5), cid(Spades, 5), cid(Hearts, 2), cid(Clubs, 7), cid(Spades, 7)}
	rest := []string{}
	for _, id := range s.DrawOrder {
		if !slices.Contains(order, id) {
			rest = append(rest, id)
		}
	}
	s, _ = s.Move(rest, 1, Unassigned, true)
	s, _ = ShuffleSupply(s, order)
	s.Round = 2
	before := s.Clone()
	if _, e := ResolveInitiative(s, order, nil); e == nil {
		t.Fatal("returns accepted without shuffle outcome")
	}
	if !reflect.DeepEqual(s, before) {
		t.Fatal("invalid initiative mutated")
	}
	n, e := ResolveInitiative(s, order, order)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(n.Order, []int{2, 1, 3}) {
		t.Fatalf("resolved rank lost or rotation wrong: %v", n.Order)
	}
	if len(n.DrawOrder) != len(order) {
		t.Fatal("nonnumber not returned or card reused")
	}
	n, _ = n.Move([]string{cid(Diamonds, 10)}, 1, Series, true)
	if !reflect.DeepEqual(n.Order, []int{2, 1, 3}) {
		t.Fatal("diamond change unfroze order")
	}
}

func TestInitiativeReturnSupplyPlansWithoutMutation(t *testing.T) {
	s, _ := NewState(3, "initiative-plan")
	before := Digest(s)
	order, e := InitiativeReturnSupply(s, nil)
	if e != nil {
		t.Fatal(e)
	}
	n, e := ResolveInitiative(s, nil, order)
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Order) != 3 || len(order) != 104 || Digest(s) != before {
		t.Fatal("bad exact supply planning")
	}
}
