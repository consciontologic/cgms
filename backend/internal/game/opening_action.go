package game

import "errors"

type OpeningIntent struct {
	Kind        string         `json:"kind"`
	Suit        Suit           `json:"suit,omitempty"`
	FormationID string         `json:"formation_id,omitempty"`
	Formation   *FormationSpec `json:"formation,omitempty"`
	CardID      string         `json:"card_id,omitempty"`
	PriorUsed   bool           `json:"prior_used"`
}

func cloneOpening(v *OpeningIntent) *OpeningIntent {
	if v == nil {
		return nil
	}
	n := *v
	if v.Formation != nil {
		c := cloneFormations([]FormationRecord{{Spec: *v.Formation}})[0].Spec
		n.Formation = &c
	}
	return &n
}
func openingPrimitive(s State, seat int, v OpeningIntent) (State, error) {
	switch v.Kind {
	case "open-series":
		return OpenSeries(s, seat, v.Suit)
	case "open-formation":
		if v.Formation == nil {
			return s, errors.New("missing formation")
		}
		return OpenFormation(s, seat, v.FormationID, *v.Formation)
	case "take-back":
		return TakeBackFormation(s, seat, v.FormationID)
	case "attach":
		return AttachRoyal(s, seat, v.CardID, v.Suit)
	}
	return s, ErrUnsupported
}
func declareOpening(s State, c Command) (State, error) {
	v := OpeningIntent{Kind: c.Kind, Suit: c.Suit, FormationID: c.FormationID, Formation: c.Formation, PriorUsed: s.Players[c.Actor-1].OpeningUsed}
	if c.Kind == "attach" {
		if len(c.Cards) != 1 {
			return s, errors.New("attachment selection")
		}
		v.CardID = c.Cards[0]
	}
	candidate, e := openingPrimitive(s, c.Actor, v)
	if e != nil {
		return s, e
	}
	n := s.Clone()
	n.Players[c.Actor-1].OpeningUsed = candidate.Players[c.Actor-1].OpeningUsed
	n.Pending = &PendingAction{ID: c.ID, Kind: "opening", Actor: c.Actor, Responders: responseSeats(s, c.Actor), Opening: cloneOpening(&v)}
	return n, nil
}
func finishOpening(s State) (State, []Event, error) {
	p := s.Pending
	if p == nil || p.Opening == nil {
		return s, nil, errors.New("missing opening")
	}
	n := s.Clone()
	n.Pending = nil
	committed := n.Players[p.Actor-1].OpeningUsed
	n.Players[p.Actor-1].OpeningUsed = p.Opening.PriorUsed
	result, e := openingPrimitive(n, p.Actor, *p.Opening)
	if e != nil {
		n.Players[p.Actor-1].OpeningUsed = committed
		return n, []Event{{Kind: "opening-invalidated", Actor: p.Actor}}, nil
	}
	return result, []Event{{Kind: p.Opening.Kind, Actor: p.Actor}}, nil
}
