package game

import (
	"errors"
	"slices"
)

// BombBoundary is the narrow Q01/F02 physical-cost fixture boundary. resolved
// denotes an independently resolved response window; false preserves legal
// declaration commitments without charging successful-destruction costs.
func BombBoundary(s State, actor int, clubs []string, bombID string, baron, resolved bool) (State, error) {
	bomb, ok := s.Card(bombID)
	if !ok || s.Active != actor || !eligible(s, actor) || len(s.CurrentSeries(actor)) < 2 || !selection(s, clubs, actor, Clubs, true) || bomb.Controller == actor || !eligible(s, bomb.Controller) || len(s.Players[bomb.Controller-1].History) < 2 || bomb.Zone != Attachment || bomb.Card.Suit != Hearts || bomb.Card.Rank != 11 {
		return s, errors.New("illegal Bomb attack")
	}
	suit := Suit(bomb.Allocation)
	series := s.CurrentSeries(bomb.Controller)
	if !slices.Contains(series, suit) || len(series) == 0 || series[0] != suit || sumCards(s, clubs).Cmp(IntAmount(5)) < 0 {
		return s, errors.New("Bomb target or strength")
	}
	if baron && !hasBaron(s, actor) {
		return s, errors.New("no supported Baron")
	}
	if baron && s.Players[actor-1].Quotas["baron"] == s.Round {
		return s, errors.New("Baron used")
	}
	n := s.Clone()
	for i := range n.Cards {
		if slices.Contains(clubs, n.Cards[i].Card.ID) {
			n.Cards[i].UsedTurn = s.Turn
		}
	}
	if baron {
		n.Players[actor-1].Quotas["baron"] = s.Round
	}
	if !resolved {
		return n, nil
	}
	var e error
	if !baron {
		n, e = n.Move(clubs, 0, Draw, false)
		if e != nil {
			return s, e
		}
		n.NeedsShuffle = true
	}
	n, e = n.Move([]string{bombID}, 0, Discard, true)
	if e != nil {
		return s, e
	}
	n.Players[bomb.Controller-1].Quotas["bomb-immunity/"+string(suit)] = s.Round
	return n, nil
}
func hasBaron(s State, seat int) bool {
	for _, r := range s.Formations {
		if r.Controller == seat && r.Spec.Kind == "baron" && FormationFunctioning(s, r.ID) {
			return true
		}
	}
	if !slices.Contains(s.CurrentSeries(seat), Clubs) {
		return false
	}
	for _, r := range s.Formations {
		if r.ID == "baron" {
			return false
		}
	} // A modern record cannot fall back to a stale legacy allocation.
	counts := map[Suit]int{}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Formation && c.Allocation == "baron" && c.Card.Rank == 13 && c.AvailableFromRound <= s.Round {
			counts[c.Card.Suit]++
		}
	}
	if counts[Diamonds] == 0 {
		return false
	}
	for suit, count := range counts {
		if suit != Diamonds && count == 2 {
			return true
		}
	}
	return false
}

// CombatMatches is the N01 exact arithmetic boundary shared by ordinary and
// Baron scenario fixtures. Full Baron ability activation is separately scoped.
func CombatMatches(attack, defense Amount, baron bool) bool {
	if baron {
		return attack.Cmp(defense) > 0
	}
	return attack.Cmp(defense) == 0
}

func functioningBomb(s State, c PlacedCard) bool {
	return c.Zone == Attachment && c.Card.Rank == 11 && c.Card.Suit == Hearts && c.AvailableFromRound <= s.Round && slices.Contains(s.CurrentSeries(c.Controller), Suit(c.Allocation))
}
func finishBombCombat(s State) (State, []Event, error) {
	n := s.Clone()
	p := n.Pending
	target, ok := n.Card(p.Targets[0])
	valid := ok && functioningBomb(n, target) && target.Controller == p.Defender && eligible(n, p.Actor) && eligible(n, p.Defender) && n.Active == p.Actor && len(n.CurrentSeries(p.Actor)) >= 2 && len(n.Players[p.Defender-1].History) >= 2 && selection(n, p.Clubs, p.Actor, Clubs, false) && !SeriesProtected(n, p.Actor, Clubs) && sumCards(n, p.Clubs).Add(IntAmount(int64(p.AttackModifier))).Cmp(IntAmount(5)) >= 0
	if p.Baron && !hasBaron(n, p.Actor) {
		valid = false
	}
	if n.Players[p.Defender-1].Quotas["bomb-immunity/"+string(p.TargetSuit)] == n.Round {
		valid = false
	}
	events := []Event{}
	var err error
	if valid {
		recordInvoluntaryLoss(&n, p.Targets)
		if !p.Baron {
			n, err = n.Move(p.Clubs, 0, Draw, true)
			if err != nil {
				return s, nil, err
			}
		}
		n, err = n.Move(p.Targets, 0, Discard, true)
		if err != nil {
			return s, nil, err
		}
		n.Players[p.Defender-1].Quotas["bomb-immunity/"+string(p.TargetSuit)] = n.Round
		events = append(events, Event{Kind: "bomb-destroyed", Actor: p.Actor, Amount: IntAmount(0)})
	} else {
		events = append(events, Event{Kind: "combat-no-match", Actor: p.Actor, Amount: IntAmount(0)})
	}
	for _, id := range p.Aces {
		n, err = n.Move([]string{id}, 0, Discard, true)
		if err != nil {
			return s, nil, err
		}
	}
	n.Pending = nil
	return n, events, nil
}
