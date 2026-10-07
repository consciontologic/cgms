package game

import (
	"errors"
	"slices"
)

// KidnapperClosedCandidates is omniscient adjudication data. Never include this
// list in a policy observation: only the recorded random outcome is disclosed.
func KidnapperClosedCandidates(s State, target int) []string {
	out := []string{}
	for _, c := range s.Cards {
		if c.Controller == target && (c.Zone == Hand || c.Zone == ConcealedAce) && (c.Card.Rank == 1 || c.Card.Rank >= 11) {
			out = append(out, c.Card.ID)
		}
	}
	return out
}
func validateKidnapper(s State, seat int, a AbilityIntent) error {
	if !abilityFormation(s, seat, a.FormationID, "kidnapper") || !abilityAttackTarget(s, seat, a.TargetSeat) {
		return errors.New("illegal Kidnapper support or target")
	}
	if len(a.Targets) == 0 {
		return nil
	}
	closed := KidnapperClosedCandidates(s, a.TargetSeat)
	if len(closed) > 0 {
		if len(a.Targets) != 0 {
			return errors.New("Kidnapper must randomly select closed cards")
		}
		return nil
	}
	if len(a.Targets) != 1 {
		return errors.New("Kidnapper needs exposed choice")
	}
	c, ok := s.Card(a.Targets[0])
	if !ok || c.Controller != a.TargetSeat || c.Card.Rank < 11 || (c.Zone != Formation && c.Zone != Attachment && c.Zone != Unassigned) || RoyalProtected(s, c.Card.ID) {
		return errors.New("illegal exposed Kidnapper selection")
	}
	return nil
}

// ResolveKidnapper validates a sampled physical outcome. The replay adapter must
// verify uniform selection from KidnapperClosedCandidates using recorded words.
// With no closed candidates, selectedID must be the declared exposed target.
// The special-theft event is a 10-point obligation from Defender to Actor;
// the match adapter must apply it atomically with this board transition.
func ResolveKidnapper(s State, selectedID string) (State, []Event, error) {
	p := s.Pending
	if p == nil || p.Ability == nil || p.Ability.Kind != "kidnapper" || p.Kind != "ability" || p.Decision != nil || p.Cursor != len(p.Responders) {
		return s, nil, errors.New("Kidnapper resolution boundary")
	}
	a := *p.Ability
	if e := validateKidnapper(s, p.Actor, a); e != nil {
		n := s.Clone()
		n.Pending = nil
		return n, []Event{{Kind: "ability-invalidated", Actor: p.Actor, Amount: IntAmount(0)}}, nil
	}
	closed := KidnapperClosedCandidates(s, a.TargetSeat)
	if len(closed) > 0 {
		if !slices.Contains(closed, selectedID) {
			return s, nil, errors.New("invalid Kidnapper random outcome")
		}
	} else {
		if len(a.Targets) > 0 && selectedID != a.Targets[0] {
			return s, nil, errors.New("Kidnapper changed exposed selection")
		}
		choices := KidnapperOpenCandidates(s, a.TargetSeat)
		if len(choices) == 0 && selectedID == "" {
			n := s.Clone()
			n.Pending = nil
			return n, nil, nil
		}
		if !slices.Contains(choices, selectedID) {
			return s, nil, errors.New("invalid exposed Kidnapper choice")
		}
	}
	n := s.Clone()
	n.Pending = nil
	card, _ := s.Card(selectedID)
	if len(closed) == 0 {
		for bi, b := range s.Bindings {
			if !slices.Contains(b.Kings, selectedID) {
				continue
			}
			// An exposed bound king separated from its partner is an ordinary
			// royal, not an intact Doppelganger formation available for theft.
			intact := true
			for _, id := range b.Kings {
				v, _ := s.Card(id)
				if v.Controller != a.TargetSeat || v.Zone != Formation || v.Allocation != b.Formation {
					intact = false
				}
			}
			if !intact {
				continue
			}
			costs := []string{}
			for _, r := range s.Formations {
				if r.ID == a.FormationID {
					costs = slices.Clone(r.Spec.Cards)
				}
			}
			var e error
			n, e = n.Move(costs, 0, Draw, false)
			if e != nil {
				return s, nil, e
			}
			n.Bindings = append(n.Bindings[:bi], n.Bindings[bi+1:]...)
			n, e = n.Move(b.Kings, p.Actor, Unassigned, true)
			if e != nil {
				return s, nil, e
			}
			return n, []Event{{Kind: "kidnapper-obligation", Actor: p.Actor, Amount: IntAmount(10)}}, nil
		}
	}
	zone := Hand
	if card.Card.Rank == 1 {
		zone = ConcealedAce
	}
	recordInvoluntaryLoss(&n, []string{selectedID})
	n, e := n.Move([]string{selectedID}, p.Actor, zone, true)
	if e != nil {
		return s, nil, e
	}
	return n, nil, nil
}

func KidnapperOpenCandidates(s State, seat int) []string {
	out := []string{}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Card.Rank >= 11 && (c.Zone == Formation || c.Zone == Attachment || c.Zone == Unassigned) && !RoyalProtected(s, c.Card.ID) {
			out = append(out, c.Card.ID)
		}
	}
	return out
}
