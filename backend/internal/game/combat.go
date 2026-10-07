package game

import (
	"errors"
	"slices"
)

func eligible(s State, seat int) bool {
	return seat > 0 && seat <= len(s.Players) && !s.Players[seat-1].Departed && !s.Players[seat-1].Confined
}
func responseSeats(s State, actor int) []int {
	out := []int{}
	for j := 1; j < len(s.Players); j++ {
		seat := (actor-1+j)%len(s.Players) + 1
		if eligible(s, seat) {
			out = append(out, seat)
		}
	}
	return out
}
func physicalValue(s State, c PlacedCard) Amount {
	a := IntAmount(int64(c.Card.Rank))
	if c.Card.Suit == Spades {
		for _, e := range s.Effects {
			if e.Kind == "inflation" && e.Target == c.Controller {
				a, _ = a.Quo(IntAmount(2))
			}
		}
	}
	return a
}
func sumCards(s State, ids []string) Amount {
	v := IntAmount(0)
	for _, id := range ids {
		c, _ := s.Card(id)
		v = v.Add(physicalValue(s, c))
	}
	return v
}
func selection(s State, ids []string, actor int, suit Suit, unused bool) bool {
	seen := map[string]bool{}
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		c, ok := s.Card(id)
		if !ok || seen[id] || c.Controller != actor || c.Zone != Series || c.Card.Suit != suit || (unused && c.AvailableFromRound > s.Round) || (unused && c.UsedTurn == s.Turn) {
			return false
		}
		seen[id] = true
	}
	return true
}
func attack(s State, c Command) (State, error) {
	if s.Pending != nil || s.Active != c.Actor || !eligible(s, c.Actor) || len(s.CurrentSeries(c.Actor)) < 2 || !selection(s, c.Cards, c.Actor, Clubs, true) || len(c.Targets) == 0 {
		return s, errors.New("illegal attack")
	}
	target, ok := s.Card(c.Targets[0])
	baron := c.Kind == "baron-attack" || c.Kind == "baron-bomb-attack"
	bomb := c.Kind == "bomb-attack" || c.Kind == "baron-bomb-attack"
	if !ok || target.Controller == c.Actor || !eligible(s, target.Controller) || len(s.Players[target.Controller-1].History) < 2 {
		return s, errors.New("illegal target")
	}
	suit := target.Card.Suit
	if bomb {
		suit = Suit(target.Allocation)
		if len(c.Targets) != 1 || !functioningBomb(s, target) {
			return s, errors.New("illegal Bomb target")
		}
	} else if !selection(s, c.Targets, target.Controller, suit, false) {
		return s, errors.New("illegal series target")
	}
	if SeriesProtected(s, c.Actor, Clubs) || !bomb && SeriesProtected(s, target.Controller, suit) {
		return s, errors.New("People protection")
	}
	if s.Players[target.Controller-1].Quotas["bomb-immunity/"+string(suit)] == s.Round {
		return s, errors.New("Bomb round immunity")
	}
	if !bomb {
		for _, card := range s.Cards {
			if card.Controller == target.Controller && Suit(card.Allocation) == suit && functioningBomb(s, card) {
				return s, errors.New("Bomb must be attacked first")
			}
		}
	}
	series := s.CurrentSeries(target.Controller)
	bypass := c.Kind == "infiltrator-attack" || c.Infiltrator
	if bypass && (suit == Diamonds || attached(s, c.Actor, Clubs) == "" || s.Players[c.Actor-1].Quotas["infiltrator"] == turnCycle(s, c.Actor)) {
		return s, errors.New("illegal Infiltrator")
	}
	if len(series) == 0 || series[0] != suit && !bypass {
		return s, errors.New("attack order")
	}
	if baron {
		if !hasBaron(s, c.Actor) || s.Players[c.Actor-1].Quotas["baron"] == s.Round {
			return s, errors.New("illegal Baron")
		}
		if !bomb && !wholeSeries(s, target.Controller, suit, c.Targets) {
			return s, errors.New("Baron must target entire series")
		}
	}
	if c.Exile && (bomb || suit != Hearts || !hasBarricade(s, target.Controller) || !hasExile(s, c.Actor) || s.Players[c.Actor-1].Quotas["exile"] == s.Round) {
		return s, errors.New("illegal Exile")
	}
	modifier := 0
	if c.AceID != "" {
		ace, ok := s.Card(c.AceID)
		if !ok || ace.Controller != c.Actor || ace.Zone != ConcealedAce || ace.AvailableFromRound > s.Round || s.Players[c.Actor-1].AceRound == s.Round {
			return s, errors.New("illegal attack ace")
		}
		modifier = c.Value
		if c.Value == 0 && ace.Card.Suit == Clubs {
			modifier = 10
		} else if c.Value < 2 || c.Value > 10 {
			return s, errors.New("invalid numerical ace")
		}
	}
	if (!bomb && !CombatMatches(sumCards(s, c.Cards).Add(IntAmount(int64(modifier))), sumCards(s, c.Targets), baron)) || (bomb && sumCards(s, c.Cards).Add(IntAmount(int64(modifier))).Cmp(IntAmount(5)) < 0) {
		return s, errors.New("not exact")
	}
	n := s.Clone()
	if bypass {
		var e error
		n, e = n.Move([]string{attached(n, c.Actor, Clubs)}, 0, Draw, false)
		if e != nil {
			return s, e
		}
		n.NeedsShuffle = true
		n.Players[c.Actor-1].Quotas["infiltrator"] = turnCycle(s, c.Actor)
	}
	for i := range n.Cards {
		if slices.Contains(c.Cards, n.Cards[i].Card.ID) {
			n.Cards[i].UsedTurn = s.Turn
		}
	}
	if baron {
		n.Players[c.Actor-1].Quotas["baron"] = s.Round
	}
	if c.Exile {
		n.Players[c.Actor-1].Quotas["exile"] = s.Round
	}
	n.Pending = &PendingAction{Baron: baron, Bomb: bomb, Exile: c.Exile, Infiltrator: bypass, ID: c.ID, Actor: c.Actor, Defender: target.Controller, Clubs: slices.Clone(c.Cards), Targets: slices.Clone(c.Targets), Responders: responseSeats(s, c.Actor), AttackModifier: modifier}
	if bomb {
		n.Pending.TargetSuit = suit
	}
	if c.AceID != "" {
		n.Pending.Aces = []string{c.AceID}
		n.Players[c.Actor-1].AceRound = s.Round
		beginDexter(&n, target.Controller, c.ID, "attack-modifier", c.AceID, c.Actor)
	}
	return n, nil
}
func finishCombat(s State) (State, []Event, error) {
	n := s.Clone()
	p := n.Pending
	if p != nil && p.Kind == "opening" {
		return finishOpening(n)
	}
	if p != nil && p.Kind == "ability" {
		switch p.Ability.Kind {
		case "fate", "richer-sacrifice", "barricade-sacrifice", "kidnapper":
			return n, nil, nil
		}
		return ResolveAbility(n, nil, nil)
	}
	if p != nil && p.Kind == "loan-return" {
		return finishLoanReturn(n)
	}
	if p != nil && p.Kind == "transfer" {
		return finishTransfer(n)
	}
	if p != nil && p.Kind == "purchase" {
		return n, nil, nil
	}
	if p != nil && p.Kind == "ponzi" {
		return n, []Event{{Kind: "ponzi-ready", Actor: p.Actor}}, nil
	}
	if p != nil && p.Kind == "justice" {
		return beginJustice(n)
	}
	if p == nil {
		return n, nil, nil
	}
	events := []Event{}
	if p.Bomb {
		return finishBombCombat(s)
	}
	if combatStillLegal(n, p) && CombatMatches(sumCards(n, p.Clubs).Add(IntAmount(int64(p.AttackModifier))), sumCards(n, p.Targets).Add(IntAmount(int64(p.DefenseModifier))), p.Baron) {
		target, _ := n.Card(p.Targets[0])
		barricade := hasBarricade(n, p.Defender) && !hasBarricade(n, p.Actor) && !p.Exile
		zone := Hand
		seat := p.Actor
		if target.Card.Suit == Hearts || target.Card.Suit == Clubs {
			zone = Draw
			seat = 0
		}
		recordInvoluntaryLoss(&n, p.Targets)
		var e error
		n, e = n.Move(p.Targets, seat, zone, true)
		if e != nil {
			return s, nil, e
		}
		if target.Card.Suit == Diamonds {
			n, e = n.Move(p.Targets, p.Actor, Series, true)
			if e != nil {
				return s, nil, e
			}
		}
		if target.Card.Suit == Clubs || target.Card.Suit == Hearts && barricade {
			n, e = n.Move(p.Clubs, 0, Draw, true)
			if e != nil {
				return s, nil, e
			}
			if target.Card.Suit == Clubs && !p.Baron {
				events = append(events, Event{Kind: "combat-score", Actor: p.Actor, Amount: IntAmount(int64(3 * len(p.Clubs)))})
			}
		}
		if zone == Draw {
			n.NeedsShuffle = true
		}
		events = append(events, Event{Kind: "combat-resolved", Actor: p.Actor})
	} else {
		events = append(events, Event{Kind: "combat-no-match", Actor: p.Actor})
	}
	for _, id := range p.Aces {
		var e error
		n, e = n.Move([]string{id}, 0, Discard, true)
		if e != nil {
			return s, nil, e
		}
	}
	n.Pending = nil
	return n, events, nil
}
func advanceResponse(s State) (State, []Event, error) {
	n := s.Clone()
	n.Pending.Cursor++
	for n.Pending.Cursor < len(n.Pending.Responders) && !eligible(n, n.Pending.Responders[n.Pending.Cursor]) {
		n.Pending.Cursor++
	}
	if n.Pending.Cursor == len(n.Pending.Responders) {
		return finishCombat(n)
	}
	return n, nil, nil
}

func recordInvoluntaryLoss(s *State, ids []string) {
	for _, id := range ids {
		c, _ := s.Card(id)
		if c.Zone != Series && c.Zone != Attachment {
			continue
		}
		for i := range s.Effects {
			if s.Effects[i].Kind == "compensation" && s.Effects[i].Target == c.Controller {
				s.Effects[i].Losses++
			}
		}
	}
}

func hasBarricade(s State, seat int) bool {
	if !slices.Contains(s.CurrentSeries(seat), Hearts) {
		return false
	}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Attachment && c.Card.Rank == 12 && c.Card.Suit == Hearts && c.AvailableFromRound <= s.Round && (c.Allocation == "" || c.Allocation == string(Hearts)) {
			return true
		}
	}
	return false
}

// Responses may consume supporting cards. Revalidate eligibility and physical
// custody before applying any success-only movement or award (GAME_RULES 2.8.5).
func combatStillLegal(s State, p *PendingAction) bool {
	if !eligible(s, p.Actor) || !eligible(s, p.Defender) || s.Active != p.Actor || len(s.CurrentSeries(p.Actor)) < 2 || len(s.Players[p.Defender-1].History) < 2 || !selection(s, p.Clubs, p.Actor, Clubs, false) || len(p.Targets) == 0 {
		return false
	}
	target, ok := s.Card(p.Targets[0])
	if !ok || !selection(s, p.Targets, p.Defender, target.Card.Suit, false) || SeriesProtected(s, p.Actor, Clubs) || SeriesProtected(s, p.Defender, target.Card.Suit) {
		return false
	}
	if s.Players[p.Defender-1].Quotas["bomb-immunity/"+string(target.Card.Suit)] == s.Round {
		return false
	}
	if p.Baron && (!hasBaron(s, p.Actor) || !wholeSeries(s, p.Defender, target.Card.Suit, p.Targets)) {
		return false
	}
	for _, c := range s.Cards {
		if c.Controller == p.Defender && Suit(c.Allocation) == target.Card.Suit && functioningBomb(s, c) {
			return false
		}
	}
	return true
}

func wholeSeries(s State, seat int, suit Suit, ids []string) bool {
	count := 0
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Series && c.Card.Suit == suit {
			count++
		}
	}
	return len(ids) == count
}
func hasExile(s State, seat int) bool {
	if !slices.Contains(s.CurrentSeries(seat), Clubs) {
		return false
	}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Attachment && c.Card.Rank == 12 && c.Card.Suit == Clubs && (c.Allocation == "" || c.Allocation == string(Clubs)) && c.AvailableFromRound <= s.Round {
			return true
		}
	}
	return false
}
