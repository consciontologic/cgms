package game

import (
	"errors"
	"slices"
)

const LegalMenuVersion = "narrow-single-pair-combat-v1"

type PlayerView struct {
	Seat      int    `json:"seat"`
	HandCount int    `json:"hand_count"`
	AceCount  int    `json:"ace_count"`
	Confined  bool   `json:"confined"`
	Departed  bool   `json:"departed"`
	History   []Suit `json:"history"`
}
type Observation struct {
	// OwnBindings contains only this viewer's currently controlled physical
	// kings. A separated partner is omitted, while the immutable role survives.
	OwnBindings   []FormationSubstitution `json:"own_bindings,omitempty"`
	Loans         []FormationLoan         `json:"loans,omitempty"`
	Formations    []FormationRecord       `json:"formations,omitempty"`
	Proposals     []Proposal              `json:"proposals,omitempty"`
	Code          *CodeSettings           `json:"code,omitempty"`
	PublicHistory []string                `json:"public_history"`
	MenuCoverage  string                  `json:"menu_coverage"`
	KnownReserved []Card                  `json:"known_reserved,omitempty"`
	GameID        string                  `json:"game_id"`
	Seat          int                     `json:"seat"`
	Round         int                     `json:"round"`
	Turn          int                     `json:"turn"`
	Active        int                     `json:"active"`
	Phase         string                  `json:"phase"`
	Version       int                     `json:"version"`
	Cards         []PlacedCard            `json:"cards"`
	Players       []PlayerView            `json:"players"`
	Effects       []AceEffect             `json:"effects"`
	WindowID      string                  `json:"window_id,omitempty"`
	RequiredActor int                     `json:"required_actor"`
	DecisionID    string                  `json:"decision_id,omitempty"`
	DecisionKind  string                  `json:"decision_kind,omitempty"`
	Choices       [][]string              `json:"choices,omitempty"`
	Legal         []Command               `json:"legal"`
}

func Observe(s State, seat int) (Observation, error) {
	o, e := observeBase(s, seat)
	if e == nil {
		o.Legal = LegalCommands(s, seat)
	}
	return o, e
}

// ObserveWithoutMenu projects the same authorized information as Observe, without
// generating its narrow combat menu. Callers providing their own legal menu avoid
// speculative transitions whose results would otherwise be discarded.
func ObserveWithoutMenu(s State, seat int) (Observation, error) {
	return observeBase(s, seat)
}

func observeBase(s State, seat int) (Observation, error) {
	if seat < 1 || seat > len(s.Players) {
		return Observation{}, errors.New("invalid observer")
	}
	o := Observation{PublicHistory: slices.Clone(s.PublicHistory), MenuCoverage: LegalMenuVersion, GameID: s.GameID, Seat: seat, Round: s.Round, Turn: s.Turn, Active: s.Active, Phase: s.Phase, Version: s.Version, Effects: slices.Clone(s.Effects)}
	for _, p := range s.Players {
		v := PlayerView{Seat: p.Seat, Confined: p.Confined, Departed: p.Departed, History: slices.Clone(p.History)}
		for _, c := range s.Cards {
			if c.Controller == p.Seat {
				if c.Zone == Hand {
					v.HandCount++
				}
				if c.Zone == ConcealedAce {
					v.AceCount++
				}
			}
		}
		o.Players = append(o.Players, v)
	}
	for _, c := range s.Cards {
		if publicZone(c.Zone) && !slices.Contains(o.PublicHistory, c.Card.ID) {
			o.PublicHistory = append(o.PublicHistory, c.Card.ID)
		}
		if c.Controller == seat || (c.Zone != Hand && c.Zone != ConcealedAce && c.Zone != Draw) {
			o.Cards = append(o.Cards, c)
		}
	}
	for _, binding := range s.Bindings {
		view := FormationSubstitution{}
		for _, id := range binding.Kings {
			card, ok := s.Card(id)
			if ok && card.Controller == seat && (card.Zone == Hand || card.Zone == Formation || card.Zone == Unassigned) {
				view.Kings = append(view.Kings, id)
			}
		}
		if len(view.Kings) == 0 {
			continue
		}
		for _, suit := range []Suit{Hearts, Diamonds, Clubs, Spades} {
			for rank := 11; rank <= 13; rank++ {
				slot := FormationSlot{Rank: rank, Suit: suit}
				if formationSlotKey(slot) == binding.Slot {
					view.Slot = slot
				}
			}
		}
		if view.Slot.Rank != 0 {
			o.OwnBindings = append(o.OwnBindings, view)
		}
	}
	if s.Pending != nil {
		p := s.Pending
		o.WindowID = p.ID
		if p.Kind == "scheduled-turn-draw" {
			o.RequiredActor = p.Actor
			o.DecisionKind = p.Kind
		}
		if p.Decision != nil {
			d := p.Decision
			o.RequiredActor = d.Actor
			o.DecisionKind = d.Kind
			if d.Actor == seat {
				o.DecisionID = d.ID
				for _, id := range d.Reserved {
					card, _ := s.Card(id)
					o.KnownReserved = append(o.KnownReserved, card.Card)
				}
				for _, v := range d.Choices {
					if d.Kind == "justice" {
						card, _ := s.Card(v[0])
						if card.Zone == Hand || card.Zone == ConcealedAce {
							continue
						}
					}
					o.Choices = append(o.Choices, slices.Clone(v))
				}
			}
		} else if p.Cursor < len(p.Responders) {
			o.RequiredActor = p.Responders[p.Cursor]
		} else if p.Cursor == len(p.Responders) && p.Kind == "ability" && p.Ability != nil && p.Ability.Kind == "kidnapper" && len(KidnapperClosedCandidates(s, p.Ability.TargetSeat)) == 0 {
			// The server owns the random closed-card outcome. Only after every
			// response, with no closed candidates, does the acting player choose
			// an exposed royal. This is an outcome command, not an EffectDecision.
			choices := KidnapperOpenCandidates(s, p.Ability.TargetSeat)
			if len(choices) > 0 {
				o.RequiredActor = p.Actor
				o.DecisionKind = "kidnapper-outcome"
				if seat == p.Actor {
					for _, id := range choices {
						if len(p.Ability.Targets) == 0 || p.Ability.Targets[0] == id {
							o.Choices = append(o.Choices, []string{id})
						}
					}
				}
			}
		}
	}
	for _, r := range s.Formations {
		public := true
		for _, id := range r.Spec.Cards {
			v, _ := s.Card(id)
			if !publicZone(v.Zone) {
				public = false
			}
		}
		if public {
			o.Formations = append(o.Formations, cloneFormations([]FormationRecord{r})[0])
		}
	}
	for _, loan := range s.Loans {
		if loan.From == seat || loan.To == seat {
			v := loan
			v.Cards = slices.Clone(loan.Cards)
			o.Loans = append(o.Loans, v)
		}
	}
	for _, p := range s.Proposals {
		if p.From == seat || p.Terms.To == seat {
			o.Proposals = append(o.Proposals, cloneProposals([]Proposal{p})[0])
		}
	}
	if s.Code != nil {
		v := *s.Code
		o.Code = &v
	}
	return o, nil
}

// LegalCommands returns an authorized finite menu for the implemented narrow slice.
// It is not a promise of full-rule coverage or a license to infer a pass on wait.
func LegalCommands(s State, seat int) []Command {
	out := []Command{}
	add := func(c Command) {
		c.GameID = s.GameID
		c.Actor = seat
		c.ID = "menu/" + Digest(c)
		if _, _, e := Apply(s, c); e == nil {
			out = append(out, c)
		}
	}
	if s.Pending != nil {
		p := s.Pending
		if p.Decision != nil {
			d := p.Decision
			if d.Actor == seat {
				for _, v := range d.Choices {
					if d.Kind == "justice" {
						card, _ := s.Card(v[0])
						if card.Zone == Hand || card.Zone == ConcealedAce {
							continue
						}
					}
					add(Command{Kind: "decision", WindowID: p.ID, DecisionID: d.ID, Cards: slices.Clone(v)})
				}
			}
			return out
		}
		if p.Cursor < len(p.Responders) && p.Responders[p.Cursor] == seat {
			add(Command{Kind: "pass", WindowID: p.ID})
			add(Command{Kind: "negotiate", WindowID: p.ID})
			for _, c := range s.Cards {
				if c.Controller == seat && c.Zone == ConcealedAce {
					for _, kind := range []string{"confinement", "inflation", "advancement-defense"} {
						add(Command{Kind: kind, WindowID: p.ID, AceID: c.Card.ID})
					}
					for v := 2; v <= 10; v++ {
						add(Command{Kind: "numerical-defense", WindowID: p.ID, AceID: c.Card.ID, Value: v})
					}
				}
			}
			return out
		}
	}
	if s.Active != seat {
		return out
	}
	clubs := availableNumbers(s, seat, Clubs)
	for _, cs := range menuSubsets(clubs) {
		for _, p := range s.Players {
			if p.Seat == seat {
				continue
			}
			for _, suit := range s.CurrentSeries(p.Seat) {
				for _, targets := range menuSubsets(targetNumbers(s, p.Seat, suit)) {
					add(Command{Kind: "attack", Cards: cs, Targets: targets})
				}
			}
		}
	}
	return out
}

// Menu deliberately restricts attacks to one or two physical cards per side.
func menuSubsets(ids []string) [][]string {
	out := [][]string{}
	for i, id := range ids {
		out = append(out, []string{id})
		for j := i + 1; j < len(ids); j++ {
			out = append(out, []string{id, ids[j]})
		}
	}
	return out
}
func targetNumbers(s State, seat int, suit Suit) []string {
	ids := []string{}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Series && c.Card.Suit == suit {
			ids = append(ids, c.Card.ID)
		}
	}
	return ids
}
