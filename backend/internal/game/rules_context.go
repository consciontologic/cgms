package game

import (
	"errors"
	"maps"
)

// RulesContext contains own quota markers and public calculations, never a
// speculative legal menu or hidden-card-dependent availability for other seats.
type RulesContext struct {
	OpeningUsed       bool               `json:"opening_used"`
	AceUsedThisRound  bool               `json:"ace_used_this_round"`
	CompensationUsed  bool               `json:"compensation_used"`
	Underground       bool               `json:"underground"`
	Quotas            map[string]int     `json:"quotas"`
	QuotaUsed         map[string]bool    `json:"quota_used"`
	PurchaseUnitPrice int                `json:"purchase_unit_price"`
	Inflation         bool               `json:"inflation"`
	Combat            *CombatCalculation `json:"combat,omitempty"`
	Pending           *PendingContext    `json:"pending,omitempty"`
}

// PendingContext describes the declared public action. ResponseTypes contains
// coarse categories for this observer's current response, not legal commands:
// possession of an available Ace and all selected payloads are checked by Apply.
// It deliberately carries no card IDs, private offer terms or effect choices.
type PendingContext struct {
	Type          string   `json:"type"`
	Actor         int      `json:"actor"`
	TargetSeat    int      `json:"target_seat,omitempty"`
	ResponseTypes []string `json:"response_types"`
}

type CombatCalculation struct {
	Actor           int      `json:"actor"`
	Defender        int      `json:"defender"`
	AttackCards     []string `json:"attack_cards"`
	TargetCards     []string `json:"target_cards"`
	PhysicalAttack  Amount   `json:"physical_attack"`
	AttackModifier  int      `json:"attack_modifier"`
	Attack          Amount   `json:"attack"`
	PhysicalDefense Amount   `json:"physical_defense"`
	DefenseModifier int      `json:"defense_modifier"`
	Defense         Amount   `json:"defense"`
	Comparison      string   `json:"comparison"`
}

func ProjectRulesContext(s State, seat int) (RulesContext, error) {
	if seat < 1 || seat > len(s.Players) {
		return RulesContext{}, errors.New("invalid observer")
	}
	p := s.Players[seat-1]
	out := RulesContext{OpeningUsed: p.OpeningUsed, AceUsedThisRound: p.AceRound == s.Round, CompensationUsed: p.CompensationUsed, Underground: p.Underground, Quotas: maps.Clone(p.Quotas), PurchaseUnitPrice: purchasePrice(s, seat)}
	// This interprets only the observer's counters, not support, timing or full
	// action legality. A cycle refreshes at that seat's own scheduled turn.
	out.QuotaUsed = map[string]bool{}
	for _, key := range []string{"richer", "dexter", "negotiator", "baron", "exile"} {
		out.QuotaUsed[key] = p.Quotas[key] == s.Round
	}
	for _, key := range []string{"barricade", "infiltrator", "justice", "ponzi", "code", "fate", "kidnapper", "compensation", "main-inflation"} {
		out.QuotaUsed[key] = p.Quotas[key] == turnCycle(s, seat)
	}
	out.QuotaUsed["ringleader"] = p.Quotas["ringleader"] == 1
	out.QuotaUsed["compensation"] = out.QuotaUsed["compensation"] || out.CompensationUsed || out.AceUsedThisRound
	out.QuotaUsed["main-inflation"] = out.QuotaUsed["main-inflation"] || out.AceUsedThisRound
	for _, e := range s.Effects {
		if e.Kind == "inflation" && e.Target == seat {
			out.Inflation = true
		}
	}
	if a := s.Pending; a != nil && len(a.Clubs) > 0 {
		// Only exposed physical operands are public. Never infer an operand from a
		// hidden card merely because an internal pending record references it.
		public := true
		for _, ids := range [][]string{a.Clubs, a.Targets} {
			for _, id := range ids {
				c, ok := s.Card(id)
				if !ok || !publicZone(c.Zone) {
					public = false
				}
			}
		}
		if public {
			c := &CombatCalculation{Actor: a.Actor, Defender: a.Defender, AttackCards: append([]string{}, a.Clubs...), TargetCards: append([]string{}, a.Targets...), PhysicalAttack: sumCards(s, a.Clubs), AttackModifier: a.AttackModifier, PhysicalDefense: sumCards(s, a.Targets), DefenseModifier: a.DefenseModifier, Comparison: "equal"}
			if a.Bomb {
				c.PhysicalDefense = IntAmount(5)
				c.DefenseModifier = 0
				c.Comparison = "at_least"
			} else if a.Baron {
				c.Comparison = "greater"
			}
			c.Attack = c.PhysicalAttack.Add(IntAmount(int64(c.AttackModifier)))
			c.Defense = c.PhysicalDefense.Add(IntAmount(int64(c.DefenseModifier)))
			out.Combat = c
		}
	}
	out.Pending = projectPendingContext(s, seat, out.Combat)
	return out, nil
}

func projectPendingContext(s State, seat int, combat *CombatCalculation) *PendingContext {
	a := s.Pending
	if a == nil {
		return nil
	}
	out := &PendingContext{Type: "action", Actor: a.Actor, ResponseTypes: []string{}}
	switch a.Kind {
	case "":
		out.Type, out.TargetSeat = "attack", a.Defender
		if a.Infiltrator {
			out.Type = "infiltrator-attack"
		}
		if a.Baron {
			out.Type = "baron-attack"
		}
		if a.Bomb {
			out.Type = "bomb-attack"
		}
		if a.Baron && a.Bomb {
			out.Type = "baron-bomb-attack"
		}
	case "ponzi":
		out.Type, out.TargetSeat = a.Kind, a.Defender
	case "justice", "purchase", "transfer", "loan-return":
		out.Type = a.Kind
	case "opening":
		if a.Opening != nil {
			switch a.Opening.Kind {
			case "open-series", "open-formation", "take-back", "attach":
				out.Type = a.Opening.Kind
			}
		}
	case "ability":
		if a.Ability != nil {
			switch a.Ability.Kind {
			case "fate", "kidnapper", "main-inflation":
				out.Type, out.TargetSeat = a.Ability.Kind, a.Ability.TargetSeat
			case "ringleader", "richer-sacrifice", "barricade-sacrifice", "code", "compensation":
				out.Type = a.Ability.Kind
			case "dexter-assassination":
				out.Type = a.Ability.Kind
				// Only a currently exposed target may identify its controller.
				// Do not turn malformed private references into a visibility oracle.
				for _, id := range a.Ability.Targets {
					if c, ok := s.Card(id); ok && publicZone(c.Zone) {
						out.TargetSeat = c.Controller
						break
					}
				}
			}
		}
	}
	if out.TargetSeat < 1 || out.TargetSeat > len(s.Players) {
		out.TargetSeat = 0
	}
	if s.Phase != "playing" || a.Decision != nil || a.Cursor < 0 || a.Cursor >= len(a.Responders) || a.Responders[a.Cursor] != seat || !eligible(s, seat) {
		return out
	}
	p := s.Players[seat-1]
	if p.AceRound != s.Round && len(s.CurrentSeries(seat)) >= 2 {
		out.ResponseTypes = append(out.ResponseTypes, "confinement")
		incoming := a.Actor != seat && out.TargetSeat == seat
		if a.Kind == "justice" {
			incoming = a.Actor != seat && len(p.History) >= 2
		}
		if incoming {
			if !p.CompensationUsed {
				out.ResponseTypes = append(out.ResponseTypes, "compensation-response")
			}
			inflated := false
			for _, effect := range s.Effects {
				if effect.Kind == "inflation" && effect.Target == a.Actor {
					inflated = true
				}
			}
			if !inflated {
				out.ResponseTypes = append(out.ResponseTypes, "inflation")
			}
		}
		if a.Kind == "" && !a.Bomb && a.Defender == seat && combat != nil {
			out.ResponseTypes = append(out.ResponseTypes, "numerical-defense")
			if len(a.Targets) > 0 {
				if target, ok := s.Card(a.Targets[0]); ok && target.Card.Suit == Hearts {
					out.ResponseTypes = append(out.ResponseTypes, "advancement-defense")
				}
			}
		}
	}
	if a.Kind == "" && a.Defender == seat && combat != nil && attached(s, seat, Spades) != "" && p.Quotas["negotiator"] != s.Round && len(availableNumbers(s, seat, Spades)) > 0 {
		out.ResponseTypes = append(out.ResponseTypes, "negotiate")
	}
	return out
}
