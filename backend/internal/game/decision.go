package game

import (
	"errors"
	"slices"
	"strconv"
)

func attached(s State, seat int, suit Suit) string {
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Attachment && c.Card.Suit == suit && c.Card.Rank == 11 && c.AvailableFromRound <= s.Round {
			return c.Card.ID
		}
	}
	return ""
}
func availableNumbers(s State, seat int, suit Suit) []string {
	out := []string{}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Series && c.Card.Suit == suit && c.AvailableFromRound <= s.Round {
			out = append(out, c.Card.ID)
		}
	}
	return out
}
func subsets(ids []string) [][]string {
	out := [][]string{}
	for _, id := range ids {
		more := [][]string{{id}}
		for _, v := range out {
			more = append(more, append(slices.Clone(v), id))
		}
		out = append(out, more...)
	}
	return out
}
func negotiate(s State, c Command) (State, error) {
	p := s.Pending
	if p.Kind != "" {
		return s, errors.New("not a club attack")
	}
	if p.Defender != c.Actor || attached(s, c.Actor, Spades) == "" || s.Players[c.Actor-1].Quotas["negotiator"] == s.Round {
		return s, errors.New("illegal negotiator")
	}
	spades := availableNumbers(s, c.Actor, Spades)
	if len(spades) == 0 {
		return s, errors.New("no payment")
	}
	a := sumCards(s, p.Clubs).Add(IntAmount(int64(p.AttackModifier)))
	choices := [][]string{}
	for _, v := range subsets(spades) {
		if sumCards(s, v).Cmp(a) == 0 {
			choices = append(choices, v)
		}
	}
	stage := "payment"
	actor := c.Actor
	if len(choices) == 0 && sumCards(s, spades).Cmp(a) < 0 {
		choices = append(choices, spades)
	}
	if len(choices) == 0 {
		clubs := []string{}
		for _, id := range availableNumbers(s, p.Actor, Clubs) {
			card, _ := s.Card(id)
			if card.UsedTurn != s.Turn || slices.Contains(p.Clubs, id) {
				clubs = append(clubs, id)
			}
		}
		best := IntAmount(0)
		totals := map[string]bool{}
		for _, v := range subsets(spades) {
			totals[sumCards(s, v).String()] = true
		}
		for _, v := range subsets(clubs) {
			total := sumCards(s, v).Add(IntAmount(int64(p.AttackModifier)))
			if totals[total.String()] && total.Sign() > 0 {
				if total.Cmp(best) > 0 {
					best = total
					choices = nil
				}
				if total.Cmp(best) == 0 {
					choices = append(choices, v)
				}
			}
		}
		stage = "clubs"
		actor = p.Actor
	}
	if len(choices) == 0 {
		return s, errors.New("no positive common total")
	}
	n := s.Clone()
	n.Players[c.Actor-1].Quotas["negotiator"] = s.Round
	n.Pending.Decision = &EffectDecision{ID: c.ID + "/" + stage, EffectID: c.ID, Actor: actor, Kind: "negotiator", Stage: stage, Choices: choices}
	return n, nil
}
func containsSelection(options [][]string, v []string) bool {
	for _, o := range options {
		if len(o) != len(v) {
			continue
		}
		seen := map[string]bool{}
		for _, id := range v {
			seen[id] = true
		}
		if len(seen) != len(o) {
			continue
		}
		ok := true
		for _, id := range o {
			if !seen[id] {
				ok = false
			}
		}
		if ok {
			return true
		}
	}
	return false
}
func aceResponse(s State, c Command) (State, []Event, error) {
	p := s.Pending
	if (p.Kind != "" || p.Bomb) && (c.Kind == "numerical-defense" || c.Kind == "advancement-defense") {
		return s, nil, errors.New("response requires club attack")
	}
	ace, ok := s.Card(c.AceID)
	if !ok || ace.Controller != c.Actor || ace.Zone != ConcealedAce || ace.AvailableFromRound > s.Round || s.Players[c.Actor-1].AceRound == s.Round || len(s.CurrentSeries(c.Actor)) < 2 {
		return s, nil, errors.New("illegal ace")
	}
	switch c.Kind {
	case "numerical-defense":
		if c.Actor != p.Defender || c.Value < 2 || c.Value > 10 {
			return s, nil, errors.New("illegal defense")
		}
	case "advancement-defense":
		if len(p.Targets) == 0 {
			return s, nil, errors.New("no combat target")
		}
		target, _ := s.Card(p.Targets[0])
		if c.Actor != p.Defender || target.Card.Suit != Hearts || ace.Card.Suit != Clubs {
			return s, nil, errors.New("illegal advancement")
		}
	case "confinement":
		if ace.Card.Suit != Diamonds {
			return s, nil, errors.New("wrong ace")
		}
	case "compensation-response":
		if !attackTargetsSeat(s, p, c.Actor) || ace.Card.Suit != Hearts || s.Players[c.Actor-1].CompensationUsed {
			return s, nil, errors.New("illegal compensation")
		}
	case "inflation":
		if !attackTargetsSeat(s, p, c.Actor) || ace.Card.Suit != Spades {
			return s, nil, errors.New("illegal inflation")
		}
		for _, e := range s.Effects {
			if e.Kind == "inflation" && e.Target == p.Actor {
				return s, nil, errors.New("duplicate inflation")
			}
		}
	}
	n := s.Clone()
	n.Players[c.Actor-1].AceRound = s.Round
	if c.Kind == "compensation-response" {
		n.Players[c.Actor-1].CompensationUsed = true
		var err error
		n, err = n.ActivateEffect(AceEffect{ID: c.ID + "/effect", CardID: c.AceID, Kind: "compensation", Source: c.Actor, Target: c.Actor, Custodian: c.Actor, Expiry: "game-end"})
		if err != nil {
			return s, nil, err
		}
		return advanceResponse(n)
	}
	if c.Kind == "numerical-defense" || c.Kind == "advancement-defense" {
		v := c.Value
		if c.Kind == "advancement-defense" {
			v = 10
		}
		n.Pending.DefenseModifier += v
		n.Pending.Aces = append(n.Pending.Aces, c.AceID)
		if beginDexter(&n, p.Actor, c.ID, c.Kind, c.AceID, c.Actor) {
			return n, nil, nil
		}
		return advanceResponse(n)
	}
	if beginDexter(&n, p.Actor, c.ID, c.Kind, c.AceID, c.Actor) {
		return n, nil, nil
	}
	return resolveHostile(n, c.Kind, c.AceID, c.Actor, false)
}

// beginDexter creates a persisted ace-resolution decision, never a nested
// response window. Merely offering prevent/pass does not consume Dexter's quota.
func beginDexter(n *State, target int, effectID, stage, aceID string, caster int) bool {
	if !eligible(*n, target) {
		return false
	}
	diamonds := availableNumbers(*n, target, Diamonds)
	total := sumCards(*n, targetNumbers(*n, target, Diamonds))
	if attached(*n, target, Diamonds) == "" || total.Sign() <= 0 || total.Cmp(IntAmount(10)) > 0 || n.Players[target-1].Quotas["dexter"] == n.Round || len(diamonds) == 0 {
		return false
	}
	choices := [][]string{{}}
	for _, id := range diamonds {
		choices = append(choices, []string{id})
	}
	n.Pending.Decision = &EffectDecision{ID: effectID + "/dexter", EffectID: effectID, Actor: target, Kind: "dexter", Stage: stage, Choices: choices, AceID: aceID, AceActor: caster}
	return true
}
func resolveHostile(s State, kind, id string, actor int, prevent bool) (State, []Event, error) {
	n := s.Clone()
	p := n.Pending
	var e error
	if kind == "ability-inflation" {
		if prevent {
			n, e = n.Move([]string{id}, 0, Discard, true)
			if e != nil {
				return s, nil, e
			}
			n.Pending = nil
			return n, nil, nil
		}
		return ResolveAbility(n, nil, nil)
	}
	if kind == "attack-modifier" || kind == "numerical-defense" || kind == "advancement-defense" {
		if prevent {
			n, e = n.Move([]string{id}, 0, Discard, true)
			if e != nil {
				return s, nil, e
			}
			n.Pending.Aces = slices.DeleteFunc(n.Pending.Aces, func(ace string) bool { return ace == id })
			if kind == "attack-modifier" {
				n.Pending.AttackModifier = 0
			} else {
				n.Pending.DefenseModifier = 0
			}
		}
		// An initial attack modifier resolves before the original response window;
		// the defender still retains their normal opportunity in that window.
		if kind == "attack-modifier" {
			return n, nil, nil
		}
		return advanceResponse(n)
	}
	if kind == "inflation" && !prevent {
		n, e = n.ActivateEffect(AceEffect{ID: id + "/effect", CardID: id, Kind: "inflation", Source: actor, Target: p.Actor, Custodian: p.Actor, Expiry: "qualifying-spend"})
		if e != nil {
			return s, nil, e
		}
		return advanceResponse(n)
	}
	n, e = n.Move([]string{id}, 0, Discard, true)
	if e != nil {
		return s, nil, e
	}
	if prevent {
		return advanceResponse(n)
	}
	n.Players[p.Actor-1].Confined = true
	for _, ace := range p.Aces {
		n, e = n.Move([]string{ace}, 0, Discard, true)
		if e != nil {
			return s, nil, e
		}
	}
	if p.Kind == "transfer" {
		idx := offerIndex(n, p.ProposalID)
		if idx >= 0 {
			n.Proposals[idx].Status = "canceled"
		}
	}
	n.Pending = nil
	if p.Actor == n.Active {
		order := slices.Clone(n.Order)
		if len(order) == 0 {
			for _, pl := range n.Players {
				order = append(order, pl.Seat)
			}
		}
		pos := slices.Index(order, p.Actor)
		for j := 1; j <= len(order); j++ {
			next := (pos + j) % len(order)
			seat := order[next]
			if !n.Players[seat-1].Departed {
				n.Active = seat
				n.Players[seat-1].Confined = false
				n.Turn++
				n.Pending = &PendingAction{Kind: "scheduled-turn-draw", ID: p.ID + "/next-turn", Actor: seat}
				if next <= pos {
					n.Round++
				}
				break
			}
		}
	}
	return n, []Event{{Kind: "confined", Actor: p.Actor}}, nil
}
func answerDecision(s State, c Command) (State, []Event, error) {
	if e := validateRecordedChoices(s); e != nil {
		return s, nil, e
	}
	d := s.Pending.Decision
	if d != nil && d.Kind == "justice" {
		if len(c.RandomWords) > 0 {
			if len(c.Cards) != 0 {
				return s, nil, errors.New("sample cannot select identity")
			}
			pool := [][]string{}
			for _, v := range d.Choices {
				card, _ := s.Card(v[0])
				if card.Zone == Hand || card.Zone == ConcealedAce {
					pool = append(pool, v)
				}
			}
			if len(pool) == 0 {
				return s, nil, errors.New("empty hidden sample")
			}
			count := uint64(len(pool))
			threshold := -count % count
			for i, word := range c.RandomWords {
				x, e := strconv.ParseUint(word, 10, 64)
				if e != nil || strconv.FormatUint(x, 10) != word {
					return s, nil, errors.New("invalid random word")
				}
				if x >= threshold {
					if i != len(c.RandomWords)-1 {
						return s, nil, errors.New("extra random words")
					}
					c.Cards = slices.Clone(pool[x%count])
					break
				}
			}
			if len(c.Cards) == 0 {
				return s, nil, errors.New("incomplete random sample")
			}
		} else {
			for _, id := range c.Cards {
				card, _ := s.Card(id)
				if card.Zone == Hand || card.Zone == ConcealedAce {
					return s, nil, errors.New("hidden Justice requires random outcome")
				}
			}
		}
	}
	if d == nil || c.DecisionID != d.ID || c.Actor != d.Actor || !containsSelection(d.Choices, c.Cards) {
		return s, nil, errors.New("invalid decision")
	}
	n := s.Clone()
	p := n.Pending
	n.Decisions = append(n.Decisions, DecisionReceipt{d.ID, c.ID})
	if d.Kind == "justice" {
		return justiceAnswer(n, c)
	}
	if d.Kind == "dexter" {
		if len(c.Cards) > 0 {
			var e error
			n, e = n.Move(c.Cards, 0, Draw, false)
			if e != nil {
				return s, nil, e
			}
			n.NeedsShuffle = true
			n.Players[c.Actor-1].Quotas["dexter"] = s.Round
		}
		n.Pending.Decision = nil
		return resolveHostile(n, d.Stage, d.AceID, d.AceActor, len(c.Cards) > 0)
	}
	if d.Kind == "negotiator" {
		if d.Stage == "clubs" {
			p.Decision.Selected = slices.Clone(c.Cards)
			p.Decision.Stage = "payment"
			p.Decision.ID = d.EffectID + "/payment"
			p.Decision.Actor = p.Defender
			p.Decision.Choices = nil
			total := sumCards(n, c.Cards).Add(IntAmount(int64(p.AttackModifier)))
			for _, v := range subsets(availableNumbers(n, p.Defender, Spades)) {
				if sumCards(n, v).Cmp(total) == 0 {
					p.Decision.Choices = append(p.Decision.Choices, v)
				}
			}
			return n, nil, nil
		}
		for i := range n.Cards {
			if slices.Contains(d.Selected, n.Cards[i].Card.ID) {
				n.Cards[i].UsedTurn = s.Turn
			}
		}
		var e error
		n, e = n.Move(c.Cards, p.Actor, Hand, false)
		if e != nil {
			return s, nil, e
		}
		n, e = n.Move([]string{attached(n, p.Defender, Spades)}, 0, Draw, false)
		if e != nil {
			return s, nil, e
		}
		n.NeedsShuffle = true
		for _, ace := range p.Aces {
			n, e = n.Move([]string{ace}, 0, Discard, true)
			if e != nil {
				return s, nil, e
			}
		}
		n.Pending = nil
		return n, []Event{{Kind: "negotiated", Actor: p.Defender}}, nil
	}
	return s, nil, ErrUnsupported
}

// attackTargetsSeat uses declared targets, never hidden-card eligibility.
func attackTargetsSeat(s State, p *PendingAction, seat int) bool {
	if p == nil || p.Actor == seat {
		return false
	}
	switch p.Kind {
	case "", "ponzi":
		return p.Defender == seat
	case "justice":
		return eligible(s, seat) && len(s.Players[seat-1].History) >= 2
	case "ability":
		if p.Ability == nil {
			return false
		}
		switch p.Ability.Kind {
		case "fate", "kidnapper", "main-inflation":
			return p.Ability.TargetSeat == seat
		case "dexter-assassination":
			for _, id := range p.Ability.Targets {
				if c, ok := s.Card(id); ok && c.Controller == seat {
					return true
				}
			}
		}
	}
	return false
}
