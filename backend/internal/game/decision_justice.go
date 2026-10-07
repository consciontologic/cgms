package game

import (
	"errors"
	"slices"
)

// Justice's random selection is an explicit adapter-supplied outcome, recorded
// once in the private continuation. Replay must additionally verify its RNG word.
func justice(s State, c Command) (State, error) {
	if s.Pending != nil || s.Active != c.Actor || !eligible(s, c.Actor) || len(s.CurrentSeries(c.Actor)) < 2 || !slices.Contains(s.CurrentSeries(c.Actor), Clubs) || s.Players[c.Actor-1].Quotas["justice"] == turnCycle(s, c.Actor) {
		return s, errors.New("illegal justice")
	}
	if !justiceSource(s, c.Actor, c.FormationID, c.AceID) {
		return s, errors.New("justice unsupported formation")
	}
	n := s.Clone()
	n.Players[c.Actor-1].Quotas["justice"] = turnCycle(s, c.Actor)
	n.Pending = &PendingAction{Kind: "justice", ID: c.ID, Actor: c.Actor, JusticeQueen: c.AceID, Responders: responseSeats(s, c.Actor)}
	return n, nil
}
func beginJustice(s State) (State, []Event, error) {
	n := s.Clone()
	p := n.Pending
	if !eligible(n, p.Actor) || len(n.CurrentSeries(p.Actor)) < 2 || !justiceSource(n, p.Actor, "", p.JusticeQueen) {
		n.Pending = nil
		return n, []Event{{Kind: "justice-invalidated", Actor: p.Actor}}, nil
	}
	opponents := []int{}
	for _, seat := range responseSeats(n, p.Actor) {
		if len(n.Players[seat-1].History) >= 2 {
			opponents = append(opponents, seat)
		}
	}
	p.Decision = &EffectDecision{Kind: "justice", ID: p.ID + "/select/0", EffectID: p.ID, Actor: p.Actor, Stage: "selection", Opponents: opponents, QueenID: p.JusticeQueen}
	n, e := justiceNext(n)
	return n, nil, e
}
func justiceNext(s State) (State, error) {
	n := s.Clone()
	d := n.Pending.Decision
	used := map[int]bool{}
	for _, id := range d.Reserved {
		card, _ := n.Card(id)
		used[card.Card.Rank] = true
	}
	for d.Cursor < len(d.Opponents) {
		d.Choices = nil
		for _, card := range n.Cards {
			if card.Controller == d.Opponents[d.Cursor] && card.Zone != ActiveAce && !(card.Card.Rank == 1 && card.Zone != ConcealedAce) && !used[card.Card.Rank] {
				d.Choices = append(d.Choices, []string{card.Card.ID})
			}
		}
		if len(d.Choices) > 0 {
			return n, nil
		}
		d.Cursor++
	}
	recordInvoluntaryLoss(&n, d.Reserved)
	var e error
	for _, id := range d.Reserved {
		card, _ := n.Card(id)
		zone := Hand
		if card.Card.Rank == 1 {
			zone = ConcealedAce
		} else if card.Card.Suit == Diamonds && card.Card.Rank <= 10 {
			zone = Series
		}
		n, e = n.Move([]string{id}, d.Actor, zone, true)
		if e != nil {
			return s, e
		}
	}
	n, e = n.Move([]string{d.QueenID}, 0, Draw, false)
	if e != nil {
		return s, e
	}
	n.NeedsShuffle = true
	n.Pending = nil
	return n, nil
}
func justiceAnswer(s State, c Command) (State, []Event, error) {
	n := s.Clone()
	d := n.Pending.Decision
	d.Reserved = append(d.Reserved, c.Cards...)
	d.Cursor++
	d.ID = d.EffectID + "/select/" + string(rune('0'+d.Cursor))
	n, e := justiceNext(n)
	return n, nil, e
}
func coup(s State, c Command) (State, []Event, error) {
	if s.Players[c.Actor-1].Departed || s.Players[c.Actor-1].Ally != 0 || !slices.Contains(s.Players[c.Actor-1].History, Clubs) {
		return s, nil, errors.New("illegal coup")
	}
	q := map[Suit]int{}
	for _, card := range s.Cards {
		if card.Controller == c.Actor && card.Card.Rank == 12 && (card.Card.Suit == Diamonds || card.Card.Suit == Hearts) && card.AvailableFromRound <= s.Round {
			q[card.Card.Suit]++
		}
	}
	if q[Hearts] != 2 || q[Diamonds] != 2 {
		return s, nil, errors.New("missing coup queens")
	}
	n := s.Clone()
	if n.Pending != nil {
		for _, id := range n.Pending.Aces {
			var e error
			n, e = n.Move([]string{id}, 0, Discard, true)
			if e != nil {
				return s, nil, e
			}
		}
	}
	var e error
	n, e = CloseCardBoard(n)
	if e != nil {
		return s, nil, e
	}
	return n, []Event{{Kind: "coup", Actor: c.Actor, Amount: IntAmount(50)}}, nil
}

func justiceSource(s State, actor int, formation, cost string) bool {
	queen, ok := s.Card(cost)
	if !ok || queen.Card.Rank != 12 || queen.Controller != actor || queen.Zone != Formation || queen.AvailableFromRound > s.Round {
		return false
	}
	registered := false
	for _, r := range s.Formations {
		if r.ID == queen.Allocation {
			registered = true
			if (formation == "" || formation == r.ID) && r.Controller == actor && r.Spec.Kind == "justice" && slices.Contains(r.Spec.Cards, cost) && FormationFunctioning(s, r.ID) {
				return true
			}
		}
	}
	if registered || formation != "" && formation != "justice" || queen.Allocation != "justice" {
		return false
	}
	suits := map[Suit]bool{}
	for _, c := range s.Cards {
		if c.Controller == actor && c.Zone == Formation && c.Allocation == "justice" && c.Card.Rank == 12 && c.AvailableFromRound <= s.Round {
			suits[c.Card.Suit] = true
		}
	}
	return len(suits) == 4 && slices.Contains(s.CurrentSeries(actor), Clubs)
}
