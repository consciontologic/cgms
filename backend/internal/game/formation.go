package game

import (
	"errors"
	"fmt"
	"slices"
)

// FormationSlot names a rank-and-suit role, never an additional physical card.
type FormationSlot struct {
	Rank int  `json:"rank"`
	Suit Suit `json:"suit"`
}
type FormationSubstitution struct {
	Kings []string      `json:"kings"`
	Slot  FormationSlot `json:"slot"`
}
type FormationProtection struct {
	Series    Suit   `json:"series,omitempty"`
	Formation string `json:"formation,omitempty"`
}
type FormationSpec struct {
	Kind       string                 `json:"kind"`
	Cards      []string               `json:"cards"`
	Substitute *FormationSubstitution `json:"substitute,omitempty"`
	Protection *FormationProtection   `json:"protection,omitempty"`
}
type FormationRecord struct {
	ID         string        `json:"id"`
	Controller int           `json:"controller"`
	Spec       FormationSpec `json:"spec"`
}

// ValidateFormationShape checks physical identity and canonical rank/suit roles only.
// It deliberately does not authorize an action, a phase, or support.
func ValidateFormationShape(s State, seat int, spec FormationSpec) error {
	seen := map[string]bool{}
	slots := []FormationSlot{}
	sub := spec.Substitute
	if sub != nil && (len(sub.Kings) != 2 || sub.Kings[0] == sub.Kings[1] || spec.Kind == "coup" || spec.Kind == "ponzi" || spec.Kind == "great-people" || sub.Slot.Rank < 11 || sub.Slot.Rank > 13 || !slices.Contains([]Suit{Hearts, Clubs, Spades, Diamonds}, sub.Slot.Suit)) {
		return errors.New("illegal substitution")
	}
	kings := 0
	for _, id := range spec.Cards {
		c, ok := s.Card(id)
		if !ok || seen[id] || c.Controller != seat || c.Card.Rank < 11 {
			return errors.New("invalid physical member")
		}
		seen[id] = true
		if sub != nil && slices.Contains(sub.Kings, id) {
			if c.Card.Rank != 13 {
				return errors.New("substitute is not king")
			}
			kings++
			continue
		}
		for _, b := range s.Bindings {
			if slices.Contains(b.Kings, id) {
				return errors.New("bound king cannot serve printed role")
			}
		}
		slots = append(slots, FormationSlot{c.Card.Rank, c.Card.Suit})
	}
	if sub != nil {
		if kings != 2 {
			return errors.New("missing substitute king")
		}
		slots = append(slots, sub.Slot)
		for _, b := range s.Bindings {
			if slices.Contains(b.Kings, sub.Kings[0]) || slices.Contains(b.Kings, sub.Kings[1]) {
				if !slices.Contains(b.Kings, sub.Kings[0]) || !slices.Contains(b.Kings, sub.Kings[1]) || b.Slot != formationSlotKey(sub.Slot) {
					return errors.New("fixed Doppelganger role")
				}
			}
		}
	}
	counts := map[FormationSlot]int{}
	for _, v := range slots {
		counts[v]++
	}
	allRank := func(rank int) bool {
		for _, v := range slots {
			if v.Rank != rank {
				return false
			}
		}
		return true
	}
	different := func(rank int) bool { return len(slots) == 4 && len(counts) == 4 && allRank(rank) }
	valid := false
	switch spec.Kind {
	case "fate":
		valid = different(13)
	case "baron":
		if len(slots) == 3 && allRank(13) && counts[FormationSlot{13, Diamonds}] > 0 {
			for _, su := range []Suit{Hearts, Clubs, Spades, Diamonds} {
				need := 2
				if su == Diamonds {
					need = 3
				}
				if counts[FormationSlot{13, su}] >= need {
					valid = true
				}
			}
		}
	case "underground":
		valid = len(slots) >= 5 && allRank(13)
	case "great-people":
		valid = sub == nil && len(slots) == 2 && counts[FormationSlot{12, Hearts}] == 2
	case "people":
		valid = sub != nil && sub.Slot == (FormationSlot{12, Hearts}) && len(slots) == 2 && counts[FormationSlot{12, Hearts}] == 2
	case "coup":
		valid = len(slots) == 4 && counts[FormationSlot{12, Hearts}] == 2 && counts[FormationSlot{12, Diamonds}] == 2
	case "justice":
		valid = different(12)
	case "code":
		valid = len(slots) >= 5 && allRank(12)
	case "kidnapper":
		valid = len(slots) == 2 && allRank(11) && len(counts) == 1
	case "ponzi":
		valid = len(slots) == 4 && counts[FormationSlot{11, Clubs}] == 2 && counts[FormationSlot{11, Spades}] == 2
	}
	if !valid {
		return errors.New("noncanonical formation shape")
	}
	return nil
}
func formationSlotKey(v FormationSlot) string { return fmt.Sprintf("%s-%02d", v.Suit, v.Rank) }
func cloneFormations(in []FormationRecord) []FormationRecord {
	out := slices.Clone(in)
	for i := range out {
		out[i].Spec.Cards = slices.Clone(in[i].Spec.Cards)
		if v := in[i].Spec.Substitute; v != nil {
			c := *v
			c.Kings = slices.Clone(v.Kings)
			out[i].Spec.Substitute = &c
		}
		if v := in[i].Spec.Protection; v != nil {
			c := *v
			out[i].Spec.Protection = &c
		}
	}
	return out
}

func formationSupport(s State, seat int, spec FormationSpec) bool {
	if !eligible(s, seat) {
		return false
	}
	series := s.CurrentSeries(seat)
	if spec.Kind == "coup" {
		return slices.Contains(s.Players[seat-1].History, Clubs) && s.Players[seat-1].Ally == 0
	}
	if spec.Kind != "underground" && len(series) < 2 && !s.Players[seat-1].Underground {
		return false
	}
	if spec.Substitute != nil && !slices.Contains(series, Hearts) {
		return false
	}
	if spec.Kind == "underground" {
		return true
	}
	if spec.Kind == "people" || spec.Kind == "great-people" {
		return slices.Contains(series, Hearts)
	}
	return slices.Contains(series, Clubs)
}

// FormationFunctioning is a current physical/support check, not activation permission.
func FormationFunctioning(s State, id string) bool {
	for _, r := range s.Formations {
		if r.ID != id {
			continue
		}
		if ValidateFormationShape(s, r.Controller, r.Spec) != nil || !formationSupport(s, r.Controller, r.Spec) {
			return false
		}
		for _, cid := range r.Spec.Cards {
			c, _ := s.Card(cid)
			if c.Zone != Formation || c.Allocation != id || c.AvailableFromRound > s.Round {
				return false
			}
		}
		return true
	}
	return false
}
func formationBoundary(s State, seat int) bool {
	return s.Phase == "playing" && s.Active == seat && eligible(s, seat) && s.Pending == nil && len(s.DrawQueue) == 0 && !s.NeedsShuffle
}
func protectionValid(s State, seat int, id string, spec FormationSpec) bool {
	p := spec.Protection
	protective := spec.Kind == "people" || spec.Kind == "great-people"
	if !protective {
		return p == nil
	}
	if p == nil || (p.Series == "") == (p.Formation == "") {
		return false
	}
	if p.Series != "" {
		return slices.Contains([]Suit{Clubs, Spades, Diamonds}, p.Series) && slices.Contains(s.CurrentSeries(seat), p.Series)
	}
	if spec.Kind != "great-people" || p.Formation == id {
		return false
	}
	for _, r := range s.Formations {
		if r.ID == p.Formation && r.Controller == seat {
			if r.Spec.Kind == "people" || r.Spec.Kind == "great-people" {
				return false
			}
			// The protected combo need not currently have ability support.
			if ValidateFormationShape(s, seat, r.Spec) != nil {
				return false
			}
			for _, cid := range r.Spec.Cards {
				c, _ := s.Card(cid)
				if c.Zone != Formation || c.Allocation != r.ID {
					return false
				}
			}
			return true
		}
	}
	return false
}

// OpenFormation commits an already resolved opening/reconfiguration. Callers must
// create and resolve response windows separately before calling this primitive.
func OpenFormation(s State, seat int, id string, spec FormationSpec) (State, error) {
	if id == "" || !formationBoundary(s, seat) || ValidateFormationShape(s, seat, spec) != nil || !formationSupport(s, seat, spec) || !protectionValid(s, seat, id, spec) {
		return s, errors.New("illegal formation opening")
	}
	if s.Players[seat-1].OpeningUsed && !s.Players[seat-1].Underground {
		return s, errors.New("opening allowance used")
	}
	for _, r := range s.Formations {
		if r.ID == id && r.Controller != seat {
			return s, errors.New("formation identity owned by another seat")
		}
	}
	for _, cid := range spec.Cards {
		c, _ := s.Card(cid)
		if c.AvailableFromRound > s.Round || (c.Zone != Hand && c.Zone != Unassigned && (c.Zone != Formation || c.Allocation != id)) {
			return s, errors.New("unavailable formation member")
		}
	}
	n := s.Clone()
	for i := range n.Cards {
		c := &n.Cards[i]
		if c.Zone == Formation && c.Allocation == id {
			c.Zone = Unassigned
			c.Allocation = ""
		}
		if slices.Contains(spec.Cards, c.Card.ID) {
			c.Zone = Formation
			c.Allocation = id
			rememberPublic(&n, c.Card.ID)
		}
	}
	n.Formations = slices.DeleteFunc(n.Formations, func(r FormationRecord) bool { return r.ID == id })
	n.Formations = append(n.Formations, cloneFormations([]FormationRecord{{id, seat, spec}})[0])
	if sub := spec.Substitute; sub != nil {
		found := false
		for i := range n.Bindings {
			b := &n.Bindings[i]
			if slices.Contains(b.Kings, sub.Kings[0]) {
				b.Formation = id
				b.Supported = true
				found = true
			}
		}
		if !found {
			n.Bindings = append(n.Bindings, Binding{ID: "formation/" + id, Kings: slices.Clone(sub.Kings), Slot: formationSlotKey(sub.Slot), Formation: id, Supported: true})
		}
	}
	if !s.Players[seat-1].Underground {
		n.Players[seat-1].OpeningUsed = true
	}
	if spec.Kind == "underground" {
		n.Players[seat-1].Underground = true
	}
	if err := n.Validate(); err != nil {
		return s, err
	}
	return n, nil
}
func TakeBackFormation(s State, seat int, id string) (State, error) {
	if !formationBoundary(s, seat) {
		return s, errors.New("illegal take back boundary")
	}
	found := false
	for _, r := range s.Formations {
		if r.ID == id && r.Controller == seat {
			found = true
		}
	}
	if !found {
		return s, errors.New("unknown formation")
	}
	n := s.Clone()
	for i := range n.Cards {
		c := &n.Cards[i]
		if c.Controller == seat && c.Zone == Formation && c.Allocation == id {
			c.Zone = Hand
			c.Allocation = ""
		}
	}
	n.Formations = slices.DeleteFunc(n.Formations, func(r FormationRecord) bool { return r.ID == id })
	for i := range n.Bindings {
		if n.Bindings[i].Formation == id {
			n.Bindings[i].Supported = false
		}
	}
	return n, n.Validate()
}

// AttachRoyal attaches a legal non-combo royal without consuming an opening.
func AttachRoyal(s State, seat int, cardID string, suit Suit) (State, error) {
	c, ok := s.Card(cardID)
	if !formationBoundary(s, seat) || !ok || c.Controller != seat || (c.Zone != Hand && c.Zone != Unassigned) || c.AvailableFromRound > s.Round || !slices.Contains(s.CurrentSeries(seat), suit) {
		return s, errors.New("illegal attachment")
	}
	for _, b := range s.Bindings {
		if slices.Contains(b.Kings, cardID) {
			return s, errors.New("bound king cannot attach")
		}
	}
	legal := c.Card.Rank == 13 || c.Card.Rank == 12 && (c.Card.Suit == Clubs || c.Card.Suit == Hearts) && c.Card.Suit == suit || c.Card.Rank == 11 && (c.Card.Suit == Hearts || c.Card.Suit == suit)
	if !legal {
		return s, errors.New("royal has no attachment role")
	}
	n := s.Clone()
	for i := range n.Cards {
		if n.Cards[i].Card.ID == cardID {
			n.Cards[i].Zone = Attachment
			n.Cards[i].Allocation = string(suit)
			rememberPublic(&n, cardID)
		}
	}
	return n, n.Validate()
}

// SeriesProtected reports People/Great People protection against club attacks.
func SeriesProtected(s State, seat int, suit Suit) bool {
	for _, r := range s.Formations {
		if r.Controller == seat && r.Spec.Protection != nil && r.Spec.Protection.Series == suit && FormationFunctioning(s, r.ID) {
			return true
		}
	}
	return false
}

// RoyalProtected reports Great People protection against kidnapping/assassination.
func RoyalProtected(s State, cardID string) bool {
	c, ok := s.Card(cardID)
	if !ok || c.Zone != Formation {
		return false
	}
	for _, r := range s.Formations {
		if r.Controller == c.Controller && r.Spec.Kind == "great-people" && r.Spec.Protection != nil && r.Spec.Protection.Formation == c.Allocation && FormationFunctioning(s, r.ID) {
			return true
		}
	}
	return false
}
