package game

import (
	"errors"
	"slices"
)

// A voluntary custody return uses the parties' loan action boundary (§2.3),
// keeps the returning controller as actor, and requires no new alliance/consent.
func declareLoanReturn(s State, c Command) (State, error) {
	if s.Pending != nil || !eligible(s, c.Actor) || !eligible(s, c.TargetSeat) || (s.Active != c.Actor && s.Active != c.TargetSeat) {
		return s, errors.New("loan return boundary")
	}
	if _, e := s.ReturnLoan(c.Cards, c.Actor, c.TargetSeat); e != nil {
		return s, e
	}
	n := s.Clone()
	n.Pending = &PendingAction{ID: c.ID, Kind: "loan-return", Actor: c.Actor, Defender: c.TargetSeat, Targets: slices.Clone(c.Cards), Responders: responseSeats(s, c.Actor)}
	return n, nil
}
func finishLoanReturn(s State) (State, []Event, error) {
	p := s.Pending
	if p == nil || p.Kind != "loan-return" {
		return s, nil, errors.New("missing loan return")
	}
	n := s.Clone()
	n.Pending = nil
	if !eligible(n, p.Actor) || !eligible(n, p.Defender) {
		return n, []Event{{Kind: "loan-return-invalidated", Actor: p.Actor}}, nil
	}
	result, e := n.ReturnLoan(p.Targets, p.Actor, p.Defender)
	if e != nil {
		return n, []Event{{Kind: "loan-return-invalidated", Actor: p.Actor}}, nil
	}
	return result, []Event{{Kind: "loan-return-completed", Actor: p.Actor}}, nil
}
