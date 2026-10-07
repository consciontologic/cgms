package game

import (
	"errors"
	"slices"
)

// exposeUnusedAce records the controller's explicit public disclosure, not a
// named Ace activation or authorized private inspection. The exposure itself
// wastes the card under §3.4; it neither consumes an ability quota nor responds
// to an action. A previously committed combat/effect Ace is not unused.
func exposeUnusedAce(s State, c Command) (State, []Event, error) {
	ace, ok := s.Card(c.AceID)
	if !ok || ace.Controller != c.Actor || ace.Card.Rank != 1 || ace.Zone != ConcealedAce || s.Players[c.Actor-1].Departed {
		return s, nil, errors.New("not an unused owned Ace")
	}
	if p := s.Pending; p != nil {
		if slices.Contains(p.Aces, c.AceID) || p.Ability != nil && p.Ability.AceID == c.AceID || p.Decision != nil && p.Decision.AceID == c.AceID {
			return s, nil, errors.New("Ace already committed")
		}
	}
	n, e := s.Move([]string{c.AceID}, 0, Discard, true)
	if e != nil {
		return s, nil, e
	}
	return n, []Event{{Kind: "unused-ace-wasted", Actor: c.Actor, Amount: IntAmount(0)}}, nil
}
