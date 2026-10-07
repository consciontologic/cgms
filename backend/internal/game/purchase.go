package game

import (
	"errors"
	"slices"
)

// CodeSettings belongs to the recorded declarer, independently of custody.
type CodeSettings struct {
	Declarer      int    `json:"declarer"`
	Formation     string `json:"formation"`
	DiamondBonus  int    `json:"diamond_bonus"`
	PurchasePrice int    `json:"purchase_price"`
}

func purchasePrice(s State, seat int) int {
	if s.Code != nil && s.Code.Declarer != seat {
		return s.Code.PurchasePrice
	}
	return 5
}
func validatePurchase(s State, seat, q int, ids []string) error {
	if q < 1 || q > 104 || len(ids) == 0 || !eligible(s, seat) || s.Active != seat {
		return errors.New("invalid purchase quantity or actor")
	}
	seen := map[string]bool{}
	value := IntAmount(0)
	supply := 0
	for _, c := range s.Cards {
		if c.Zone == Draw || c.Zone == Discard {
			supply++
		}
	}
	for _, id := range ids {
		c, ok := s.Card(id)
		if !ok || seen[id] || c.Controller != seat || (c.Zone != Hand && c.Zone != Series) || c.Card.Suit != Spades || c.Card.Rank < 2 || c.Card.Rank > 10 || c.AvailableFromRound > s.Round {
			return errors.New("invalid purchase payment")
		}
		seen[id] = true
		value = value.Add(physicalValue(s, c))
		supply++
	}
	if supply < q || value.Cmp(IntAmount(int64(q*purchasePrice(s, seat)))) < 0 {
		return errors.New("purchase supply or affordability")
	}
	return nil
}
func declarePurchase(s State, c Command) (State, error) {
	if s.Pending != nil {
		return s, errors.New("action pending")
	}
	if e := validatePurchase(s, c.Actor, c.Value, c.Cards); e != nil {
		return s, e
	}
	n := s.Clone()
	n.Pending = &PendingAction{ID: c.ID, Kind: "purchase", Actor: c.Actor, Targets: slices.Clone(c.Cards), Quantity: c.Value, Responders: responseSeats(s, c.Actor)}
	return n, nil
}

// resolvePurchase consumes a recorded permutation only after all responses.
// The adapter verifies the permutation against its pinned chance transcript.
func resolvePurchase(s State, c Command) (State, []Event, error) {
	p := s.Pending
	if p == nil || p.Kind != "purchase" || p.ID != c.WindowID || p.Actor != c.Actor || p.Decision != nil || p.Cursor != len(p.Responders) {
		return s, nil, errors.New("purchase outcome boundary")
	}
	if e := validatePurchase(s, p.Actor, p.Quantity, p.Targets); e != nil {
		n := s.Clone()
		n.Pending = nil
		return n, []Event{{Kind: "purchase-invalidated", Actor: p.Actor}}, nil
	}
	printed := 0
	for _, id := range p.Targets {
		v, _ := s.Card(id)
		printed += v.Card.Rank
	}
	n, e := s.Move(p.Targets, 0, Draw, false)
	if e != nil {
		return s, nil, e
	}
	for _, id := range p.Targets {
		rememberPublic(&n, id)
	}
	// Only returned payment joins the existing draw pile; recycling remains separate.
	n, e = ShuffleSupply(n, c.Cards)
	if e != nil {
		return s, nil, e
	}
	n, got, e := DrawCards(n, p.Actor, p.Quantity, false, c.Targets)
	if e != nil {
		return s, nil, e
	}
	if len(got) != p.Quantity {
		return s, nil, errors.New("purchase exact supply invariant")
	}
	if printed >= 20 {
		for i, v := range n.Effects {
			if v.Kind == "inflation" && v.Target == p.Actor {
				for j := range n.Cards {
					if n.Cards[j].Card.ID == v.CardID {
						n.Cards[j].Zone = Discard
						n.Cards[j].Controller = 0
					}
				}
				n.Effects = append(n.Effects[:i], n.Effects[i+1:]...)
				break
			}
		}
	}
	n.Pending = nil
	return n, []Event{{Kind: "purchase-completed", Actor: p.Actor}}, nil
}
