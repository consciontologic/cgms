package game

import (
	"slices"
	"testing"
)

func TestPurchaseResponseThenExactQuantity(t *testing.T) {
	s, _ := NewState(3, "purchase")
	s, _ = s.Move([]string{"deck-1-spades-10"}, 1, Hand, true)
	c := Command{GameID: s.GameID, ID: "buy", Actor: 1, Kind: "purchase", Value: 2, Cards: []string{"deck-1-spades-10"}}
	n, _, e := Apply(s, c)
	if e != nil {
		t.Fatal(e)
	}
	if n.Pending == nil || n.Pending.Kind != "purchase" {
		t.Fatal("missing response window")
	}
	if card, _ := n.Card(c.Cards[0]); card.Zone != Hand {
		t.Fatal("payment before responses")
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Actor: seat, Kind: "pass", WindowID: "buy"})
		if e != nil {
			t.Fatal(e)
		}
	}
	if n.Pending == nil || n.Pending.Cursor != len(n.Pending.Responders) {
		t.Fatal("random outcome boundary lost")
	}
	order := append(append([]string{}, n.DrawOrder...), c.Cards...)
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "chance", Actor: 1, Kind: "purchase-outcome", WindowID: "buy", Cards: order})
	if e != nil {
		t.Fatal(e)
	}
	if n.Pending != nil {
		t.Fatal("purchase unresolved")
	}
	count := 0
	for _, v := range n.Cards {
		if v.Controller == 1 {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("got %d cards", count)
	}
	if e = n.Validate(); e != nil {
		t.Fatal(e)
	}
}
func TestPurchaseSupplyIncludesPaymentAndRejectsWithoutMutation(t *testing.T) {
	s, _ := NewState(3, "purchase")
	// Exhaust both piles with legal physical zones.
	for _, card := range s.Cards {
		z := Hand
		if card.Card.Rank == 1 {
			z = ConcealedAce
		}
		s, _ = s.Move([]string{card.Card.ID}, 2, z, true)
	}
	s, _ = s.Move([]string{"deck-1-spades-10"}, 1, Hand, true)
	before := Digest(s)
	c := Command{GameID: s.GameID, ID: "buy", Actor: 1, Kind: "purchase", Value: 2, Cards: []string{"deck-1-spades-10"}}
	n, _, e := Apply(s, c)
	if e == nil || Digest(n) != before {
		t.Fatal("unavailable chosen quantity accepted or mutated")
	}
	c.Value = 1
	n, _, e = Apply(s, c)
	if e != nil {
		t.Fatalf("one-card purchase should be legal: %v", e)
	}
	if n.Players[0].OpeningUsed {
		t.Fatal("purchase consumes opening")
	}
}

func TestPurchaseCancellationRetainsPaymentAndQuantity(t *testing.T) {
	s, _ := NewState(3, "cancel-buy")
	for _, move := range []struct {
		id   string
		seat int
		zone Zone
	}{{"deck-1-spades-10", 1, Hand}, {"deck-1-diamonds-02", 2, Series}, {"deck-1-clubs-02", 2, Series}, {"deck-1-diamonds-01", 2, ConcealedAce}} {
		var e error
		s, e = s.Move([]string{move.id}, move.seat, move.zone, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "buy", Actor: 1, Kind: "purchase", Value: 2, Cards: []string{"deck-1-spades-10"}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "confine", Actor: 2, Kind: "confinement", WindowID: "buy", AceID: "deck-1-diamonds-01"})
	if e != nil {
		t.Fatal(e)
	}
	p, _ := n.Card("deck-1-spades-10")
	if p.Controller != 1 || p.Zone != Hand {
		t.Fatal("canceled payment moved")
	}
	if n.Pending != nil && n.Pending.Kind == "purchase" {
		t.Fatal("canceled purchase remains")
	}
}
func TestPurchaseInflationPriceAndPrivatePayment(t *testing.T) {
	s, _ := NewState(3, "price")
	s, _ = s.Move([]string{"deck-1-spades-10", "deck-2-spades-10"}, 1, Hand, true)
	s, _ = s.Move([]string{"deck-1-spades-01"}, 2, ConcealedAce, true)
	s, e := s.ActivateEffect(AceEffect{ID: "infl", CardID: "deck-1-spades-01", Kind: "inflation", Source: 2, Target: 1, Custodian: 1, Expiry: "transaction"})
	if e != nil {
		t.Fatal(e)
	}
	s.Code = &CodeSettings{Declarer: 2, Formation: "code", DiamondBonus: 3, PurchasePrice: 6}
	c := Command{GameID: s.GameID, ID: "buy", Actor: 1, Kind: "purchase", Value: 2, Cards: []string{"deck-1-spades-10", "deck-2-spades-10"}}
	if _, _, e = Apply(s, c); e == nil {
		t.Fatal("inflated payment 10 cannot cover 12")
	}
	c.Value = 1
	n, _, e := Apply(s, c)
	if e != nil {
		t.Fatal(e)
	}
	o, _ := Observe(n, 3)
	for _, card := range o.Cards {
		if card.Card.ID == c.Cards[0] || card.Card.ID == c.Cards[1] {
			t.Fatal("private payment leaked")
		}
	}
	for _, seat := range []int{2, 3} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('k' + seat)), Actor: seat, Kind: "pass", WindowID: "buy"})
		if e != nil {
			t.Fatal(e)
		}
	}
	order := append(append([]string{}, n.DrawOrder...), c.Cards...)
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "chance", Actor: 1, Kind: "purchase-outcome", WindowID: "buy", Cards: order})
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range c.Cards {
		if !slices.Contains(n.PublicHistory, id) {
			t.Fatal("successful payment absent from public history")
		}
	}
	if len(n.Effects) != 0 {
		t.Fatal("20 printed voluntary payment must clear inflation")
	}
}
