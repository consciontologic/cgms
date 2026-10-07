package game

import (
	"errors"
	"slices"
)

// declarePonzi covers acceptance and the response/cancellation boundary. Surviving
// financial resolution is explicitly unsupported by the card-only Apply API.
func declarePonzi(s State, c Command) (State, error) {
	target := c.Value
	if s.Pending != nil || !eligible(s, c.Actor) || !eligible(s, target) || target == c.Actor || len(s.CurrentSeries(c.Actor)) < 2 || !slices.Contains(s.CurrentSeries(c.Actor), Clubs) || len(s.Players[target-1].History) < 2 || s.Players[c.Actor-1].Quotas["ponzi"] == turnCycle(s, c.Actor) {
		return s, errors.New("illegal Ponzi")
	}
	registered := false
	valid := false
	for _, r := range s.Formations {
		if r.Controller == c.Actor && r.Spec.Kind == "ponzi" {
			registered = true
			if (c.FormationID == "" || c.FormationID == r.ID) && FormationFunctioning(s, r.ID) {
				valid = true
			}
		}
	}
	if !registered {
		counts := map[Suit]int{}
		for _, card := range s.Cards {
			if card.Controller == c.Actor && card.Zone == Formation && card.Allocation == "ponzi" && card.Card.Rank == 11 && card.AvailableFromRound <= s.Round {
				counts[card.Card.Suit]++
			}
		}
		valid = counts[Clubs] == 2 && counts[Spades] == 2 && (c.FormationID == "" || c.FormationID == "ponzi")
	}
	if !valid {
		return s, errors.New("Ponzi requires supported four physical jacks")
	}
	n := s.Clone()
	n.Players[c.Actor-1].Quotas["ponzi"] = turnCycle(s, c.Actor)
	n.Pending = &PendingAction{ID: c.ID, Kind: "ponzi", Actor: c.Actor, Defender: target, Responders: responseSeats(n, c.Actor)}
	return n, nil
}

// A turn-cycle begins at this player's scheduled turn, not every opponent turn.
func turnCycle(s State, actor int) int {
	order := slices.Clone(s.Order)
	if len(order) == 0 {
		for _, p := range s.Players {
			order = append(order, p.Seat)
		}
	}
	cycle := s.Round
	if slices.Index(order, actor) > slices.Index(order, s.Active) {
		cycle--
	}
	return cycle + 1
}
