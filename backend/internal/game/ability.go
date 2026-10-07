package game

import (
	"errors"
	"slices"
)

type AbilityIntent struct {
	PreventChecked bool     `json:"prevent_checked,omitempty"`
	Kind           string   `json:"kind"`
	FormationID    string   `json:"formation_id,omitempty"`
	TargetSeat     int      `json:"target_seat,omitempty"`
	Cards          []string `json:"cards"`
	Targets        []string `json:"targets"`
	AceID          string   `json:"ace_id,omitempty"`
	DiamondBonus   int      `json:"diamond_bonus,omitempty"`
	PurchasePrice  int      `json:"purchase_price,omitempty"`
}

func cloneAbilityIntent(a *AbilityIntent) *AbilityIntent {
	if a == nil {
		return nil
	}
	n := *a
	n.Cards = append([]string(nil), a.Cards...)
	n.Targets = append([]string(nil), a.Targets...)
	return &n
}

func abilityQuota(s State, a AbilityIntent) (string, int) {
	switch a.Kind {
	case "ringleader":
		return "ringleader", 1
	case "richer-sacrifice":
		return "richer", s.Round
	case "dexter-assassination":
		return "dexter", s.Round
	case "barricade-sacrifice":
		return "barricade", turnCycle(s, s.Active)
	default:
		return a.Kind, turnCycle(s, s.Active)
	}
}
func abilityRoyal(s State, seat int, id string, rank int, suit Suit) (PlacedCard, bool) {
	c, ok := s.Card(id)
	if !ok || c.Controller != seat || c.Zone != Attachment || c.Card.Rank != rank || suit != "" && c.Card.Suit != suit || c.AvailableFromRound > s.Round || !slices.Contains(s.CurrentSeries(seat), Suit(c.Allocation)) {
		return c, false
	}
	return c, true
}
func abilityFormation(s State, seat int, id, kind string) bool {
	for _, r := range s.Formations {
		if r.ID == id && r.Controller == seat && r.Spec.Kind == kind {
			return FormationFunctioning(s, id)
		}
	}
	return false
}
func abilityAttackTarget(s State, seat, target int) bool {
	return target != seat && eligible(s, target) && len(s.Players[target-1].History) >= 2 && len(s.CurrentSeries(seat)) >= 2
}
func validateAbility(s State, seat int, a AbilityIntent) error {
	if !eligible(s, seat) {
		return errors.New("ineligible ability actor")
	}
	bad := errors.New("illegal ability support or selection")
	switch a.Kind {
	case "kidnapper":
		return validateKidnapper(s, seat, a)
	case "compensation", "main-inflation":
		ace, ok := s.Card(a.AceID)
		if !ok || ace.Zone != ConcealedAce || ace.Controller != seat || ace.AvailableFromRound > s.Round || len(s.CurrentSeries(seat)) < 2 {
			return bad
		}
		if a.Kind == "compensation" {
			if ace.Card.Suit != Hearts {
				return bad
			}
		} else {
			if ace.Card.Suit != Spades || !abilityAttackTarget(s, seat, a.TargetSeat) {
				return bad
			}
			for _, e := range s.Effects {
				if e.Kind == "inflation" && e.Target == a.TargetSeat {
					return bad
				}
			}
		}
	case "ringleader", "richer-sacrifice", "barricade-sacrifice":
		if len(a.Cards) == 0 {
			return bad
		}
		rank := 13
		suit := Suit("")
		if a.Kind == "ringleader" {
			suit = Diamonds
		}
		if a.Kind == "barricade-sacrifice" {
			rank = 12
			suit = Hearts
		}
		c, ok := abilityRoyal(s, seat, a.Cards[0], rank, suit)
		if !ok {
			return bad
		}
		if a.Kind == "richer-sacrifice" {
			if c.Card.Suit == Diamonds {
				return bad
			}
			count := 0
			for _, v := range s.Cards {
				if v.Controller == seat && v.Zone == Attachment && v.Allocation == c.Allocation && v.Card.Rank == 13 {
					count++
				}
			}
			if count != 1 {
				return bad
			}
		}
		if a.Kind == "barricade-sacrifice" {
			if c.Allocation != string(Hearts) || len(a.Cards) != 6 {
				return bad
			}
			seen := map[string]bool{}
			for _, id := range a.Cards {
				v, ok := s.Card(id)
				if !ok || seen[id] || v.Controller != seat || v.Zone == ActiveAce || v.AvailableFromRound > s.Round {
					return bad
				}
				seen[id] = true
			}
		} else if len(a.Cards) != 1 {
			return bad
		}
	case "dexter-assassination":
		if len(a.Cards) != 1 || len(a.Targets) != 1 || attached(s, seat, Diamonds) == "" || exposedDiamondTotal(s, seat).Cmp(IntAmount(20)) < 0 {
			return bad
		}
		c, ok := s.Card(a.Cards[0])
		if !ok || c.Controller != seat || c.Zone != Series || c.Card.Suit != Diamonds || c.AvailableFromRound > s.Round {
			return bad
		}
		target, ok := s.Card(a.Targets[0])
		if !ok || target.Card.Rank < 11 || (target.Zone != Formation && target.Zone != Attachment && target.Zone != Unassigned) || !abilityAttackTarget(s, seat, target.Controller) || RoyalProtected(s, target.Card.ID) {
			return bad
		}
	case "fate":
		if !abilityFormation(s, seat, a.FormationID, "fate") || !abilityAttackTarget(s, seat, a.TargetSeat) || len(a.Cards) != 1 {
			return bad
		}
		c, ok := s.Card(a.Cards[0])
		if !ok || c.Card.Rank != 13 || c.Allocation != a.FormationID || c.Zone != Formation || c.AvailableFromRound > s.Round {
			return bad
		}
	case "code":
		if !abilityFormation(s, seat, a.FormationID, "code") || a.DiamondBonus < 1 || a.DiamondBonus > 10 || a.PurchasePrice < 5 || a.PurchasePrice > 10 {
			return bad
		}
	default:
		return ErrUnsupported
	}
	return nil
}

// DeclareAbility records an accepted action and quota, never resolution-only costs.
func DeclareAbility(s State, c Command) (State, error) {
	if !formationBoundary(s, c.Actor) {
		return s, errors.New("ability boundary")
	}
	a := AbilityIntent{Kind: c.Kind, FormationID: c.FormationID, TargetSeat: c.TargetSeat, Cards: slices.Clone(c.Cards), Targets: slices.Clone(c.Targets), AceID: c.AceID, DiamondBonus: c.Value, PurchasePrice: c.Price}
	if a.Kind == "code" && a.PurchasePrice == 0 {
		a.PurchasePrice = 5
	}
	if err := validateAbility(s, c.Actor, a); err != nil {
		return s, err
	}
	if a.Kind == "compensation" || a.Kind == "main-inflation" {
		if s.Players[c.Actor-1].AceRound == s.Round || a.Kind == "compensation" && s.Players[c.Actor-1].CompensationUsed {
			return s, errors.New("ace quota consumed")
		}
	}
	key, value := abilityQuota(s, a)
	if s.Players[c.Actor-1].Quotas[key] == value {
		return s, errors.New("ability quota consumed")
	}
	n := s.Clone()
	n.Players[c.Actor-1].Quotas[key] = value
	if a.Kind == "compensation" || a.Kind == "main-inflation" {
		n.Players[c.Actor-1].AceRound = s.Round
		if a.Kind == "compensation" {
			n.Players[c.Actor-1].CompensationUsed = true
		}
	}
	n.Pending = &PendingAction{ID: c.ID, Kind: "ability", Actor: c.Actor, Responders: responseSeats(s, c.Actor), Ability: &a}
	return n, nil
}

// ResolveAbility commits a surviving action after all responses. Random adapters
// supply recorded permutations; invalid outcomes preserve the pending state.
func ResolveAbility(s State, order, recycle []string) (State, []Event, error) {
	p := s.Pending
	if p == nil || p.Kind != "ability" || p.Ability == nil || p.Decision != nil || p.Cursor != len(p.Responders) {
		return s, nil, errors.New("ability resolution boundary")
	}
	a := *p.Ability
	if err := validateAbility(s, p.Actor, a); err != nil {
		n := s.Clone()
		n.Pending = nil
		return n, []Event{{Kind: "ability-invalidated", Actor: p.Actor, Amount: IntAmount(0)}}, nil
	}
	n := s.Clone()
	if a.Kind == "main-inflation" && !a.PreventChecked {
		n.Pending.Ability.PreventChecked = true
		if beginDexter(&n, a.TargetSeat, p.ID, "ability-inflation", a.AceID, p.Actor) {
			return n, nil, nil
		}
	}
	n.Pending = nil
	var err error
	events := []Event{}
	draw := 0
	restricted := false
	switch a.Kind {
	case "compensation":
		n, err = n.ActivateEffect(AceEffect{ID: p.ID + "/effect", CardID: a.AceID, Kind: "compensation", Source: p.Actor, Target: p.Actor, Custodian: p.Actor, Expiry: "game-end"})
	case "main-inflation":
		n, err = n.ActivateEffect(AceEffect{ID: p.ID + "/effect", CardID: a.AceID, Kind: "inflation", Source: p.Actor, Target: a.TargetSeat, Custodian: a.TargetSeat, Expiry: "qualifying-spend"})
	case "ringleader":
		n, err = n.Move(a.Cards, 0, Discard, false)
		events = append(events, Event{Kind: "ringleader", Actor: p.Actor, Amount: IntAmount(10)})
	case "richer-sacrifice":
		n, err = n.Move(a.Cards, 0, Discard, false)
		draw = 2
	case "barricade-sacrifice":
		n, err = n.Move(a.Cards[:1], 0, Discard, false)
		if err == nil {
			n, err = n.Move(a.Cards[1:], 0, Draw, false)
		}
		draw = 5
		restricted = true
	case "dexter-assassination":
		recordInvoluntaryLoss(&n, a.Targets)
		n, err = n.Move(a.Cards, 0, Draw, false)
		if err == nil {
			target, _ := n.Card(a.Targets[0])
			if target.Card.Rank == 13 {
				n, err = n.AssassinateBinding(a.Targets[0])
			} else {
				n, err = n.Move(a.Targets, 0, Discard, true)
			}
		}
	case "fate":
		n, err = n.Move(a.Cards, 0, Draw, false)
		if err == nil {
			n, draw, err = n.FateReturn(a.TargetSeat)
		}
	case "code":
		n.Code = &CodeSettings{Declarer: p.Actor, Formation: a.FormationID, DiamondBonus: a.DiamondBonus, PurchasePrice: a.PurchasePrice}
	default:
		return s, nil, ErrUnsupported
	}
	if err != nil {
		return s, nil, err
	}
	if draw > 0 {
		if n.NeedsShuffle {
			n, err = ShuffleSupply(n, order)
			if err != nil {
				return s, nil, err
			}
		} else if len(order) > 0 {
			return s, nil, errors.New("unexpected shuffle outcome")
		}
		recipient := p.Actor
		if a.Kind == "fate" {
			recipient = a.TargetSeat
		}
		var got []string
		n, got, err = DrawCards(n, recipient, draw, a.Kind == "fate", recycle)
		if err != nil {
			return s, nil, err
		}
		if restricted {
			for i := range n.Cards {
				if slices.Contains(got, n.Cards[i].Card.ID) {
					n.Cards[i].AvailableFromRound = n.Round + 1
				}
			}
		}
	}
	if err = n.Validate(); err != nil {
		return s, nil, err
	}
	return n, events, nil
}

func exposedDiamondTotal(s State, seat int) Amount {
	sum := IntAmount(0)
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Series && c.Card.Suit == Diamonds {
			sum = sum.Add(IntAmount(int64(c.Card.Rank)))
		}
	}
	return sum
}
