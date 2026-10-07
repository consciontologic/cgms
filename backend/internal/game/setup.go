package game

import (
	"errors"
	"slices"
	"sort"
)

// ShuffleSupply applies an explicit random outcome. The adapter records and
// independently verifies its pinned shuffle stream; the engine checks permutation.
func ShuffleSupply(s State, order []string) (State, error) {
	want := map[string]bool{}
	for _, c := range s.Cards {
		if c.Zone == Draw {
			want[c.Card.ID] = true
		}
	}
	if len(order) != len(want) {
		return s, errors.New("shuffle length")
	}
	for _, id := range order {
		if !want[id] {
			return s, errors.New("shuffle not permutation")
		}
		delete(want, id)
	}
	n := s.Clone()
	n.DrawOrder = slices.Clone(order)
	n.NeedsShuffle = false
	return n, nil
}
func DrawCards(s State, seat, q int, exposeDiamonds bool, recycle []string) (State, []string, error) {
	if seat < 1 || seat > len(s.Players) || q < 0 || s.Phase != "playing" || s.NeedsShuffle {
		return s, nil, errors.New("invalid draw boundary")
	}
	n := s.Clone()
	got := []string{}
	for len(got) < q {
		if len(n.DrawOrder) == 0 {
			discarded := []string{}
			for _, c := range n.Cards {
				if c.Zone == Discard {
					discarded = append(discarded, c.Card.ID)
				}
			}
			if len(discarded) == 0 {
				break
			}
			var e error
			n, e = n.Move(discarded, 0, Draw, true)
			if e != nil {
				return s, nil, e
			}
			n, e = ShuffleSupply(n, recycle)
			if e != nil {
				return s, nil, e
			}
			recycle = nil
		}
		id := n.DrawOrder[0]
		c, _ := n.Card(id)
		z := Hand
		if c.Card.Rank == 1 {
			z = ConcealedAce
		} else if exposeDiamonds && c.Card.Suit == Diamonds && c.Card.Rank <= 10 {
			z = Series
		}
		var e error
		n, e = n.Move([]string{id}, seat, z, true)
		if e != nil {
			return s, nil, e
		}
		got = append(got, id)
	}
	return n, got, nil
}

// DealAttempt is one atomic attempt; a failed guarantee leaves the original
// complete supply for the next independently recorded shuffle, not a partial deal.
func DealAttempt(s State, order []string) (State, bool, error) {
	if len(s.DrawOrder) != 104 {
		return s, false, errors.New("deal requires full supply")
	}
	n, e := ShuffleSupply(s, order)
	if e != nil {
		return s, false, e
	}
	for seat := 1; seat <= len(s.Players); seat++ {
		n, _, e = DrawCards(n, seat, 20, true, nil)
		if e != nil {
			return s, false, e
		}
	}
	for seat := 1; seat <= len(s.Players); seat++ {
		if !slices.Contains(n.CurrentSeries(seat), Diamonds) {
			return s, true, nil
		}
		n.Players[seat-1].History = []Suit{Diamonds}
	}
	return n, false, nil
}
func OpenSeries(s State, seat int, suit Suit) (State, error) {
	if seat < 1 || seat > len(s.Players) || s.Active != seat || s.Phase != "playing" || s.Pending != nil || postActionWorkPending(s) || s.Players[seat-1].Confined || s.Players[seat-1].Departed {
		return s, errors.New("invalid opening boundary")
	}
	p := s.Players[seat-1]
	old := slices.Contains(p.History, suit)
	if !old && p.OpeningUsed {
		return s, errors.New("opening allowance used")
	}
	ids := []string{}
	for _, c := range s.Cards {
		if c.Controller == seat && c.Zone == Hand && c.Card.Suit == suit && c.Card.Rank >= 2 && c.Card.Rank <= 10 {
			if c.AvailableFromRound > s.Round {
				return s, errors.New("restricted opening card")
			}
			ids = append(ids, c.Card.ID)
		}
	}
	if len(ids) == 0 {
		return s, errors.New("no number cards to open")
	}
	n, e := s.Move(ids, seat, Series, false)
	if e != nil {
		return s, e
	}
	if !old {
		n.Players[seat-1].History = append(n.Players[seat-1].History, suit)
		n.Players[seat-1].OpeningUsed = true
	}
	return n, nil
}
func ResolveInitiative(s State, draws []string, returnOrder []string) (State, error) {
	return resolveInitiative(s, draws, returnOrder, false)
}
func resolveInitiative(s State, draws []string, returnOrder []string, plan bool) (State, error) {
	if s.Phase != "playing" || s.Pending != nil {
		return s, errors.New("initiative boundary")
	}
	n := s.Clone()
	scores := map[int]int{}
	seats := []int{}
	for _, p := range n.Players {
		if !p.Departed {
			seats = append(seats, p.Seat)
		}
	}
	for _, c := range n.Cards {
		if c.Zone == Series && c.Card.Suit == Diamonds {
			scores[c.Controller] += c.Card.Rank
		}
	}
	sort.SliceStable(seats, func(i, j int) bool { return scores[seats[i]] > scores[seats[j]] })
	supply := slices.Clone(n.DrawOrder)
	discards := []string{}
	for _, c := range n.Cards {
		if c.Zone == Discard {
			discards = append(discards, c.Card.ID)
		}
	}
	if len(draws) > 0 {
		if len(draws) != len(supply)+len(discards) || !slices.Equal(draws[:len(supply)], supply) {
			return s, errors.New("initiative supply order")
		}
		tail := slices.Clone(draws[len(supply):])
		sort.Strings(tail)
		sort.Strings(discards)
		if !slices.Equal(tail, discards) {
			return s, errors.New("initiative discard permutation")
		}
		supply = slices.Clone(draws)
	} else if len(discards) > 0 {
		return s, errors.New("initiative recycle outcome required")
	}
	used := []string{}
	cursor := 0
	priority := func(group []int) []int {
		out := slices.Clone(group)
		start := (s.Round-1)%len(s.Players) + 1
		sort.Slice(out, func(i, j int) bool {
			return (out[i]-start+len(s.Players))%len(s.Players) < (out[j]-start+len(s.Players))%len(s.Players)
		})
		return out
	}
	var resolve func([]int) []int
	resolve = func(group []int) []int {
		if len(group) < 2 {
			return group
		}
		values := map[int]int{}
		for _, seat := range group {
			for cursor < len(supply) {
				id := supply[cursor]
				cursor++
				used = append(used, id)
				c, _ := n.Card(id)
				if c.Card.Rank >= 2 && c.Card.Rank <= 10 {
					values[seat] = c.Card.Rank
					break
				}
			}
			if values[seat] == 0 {
				return priority(group)
			}
		}
		sort.SliceStable(group, func(i, j int) bool { return values[group[i]] > values[group[j]] })
		out := []int{}
		for i := 0; i < len(group); {
			j := i + 1
			for j < len(group) && values[group[j]] == values[group[i]] {
				j++
			}
			out = append(out, resolve(group[i:j])...)
			i = j
		}
		return out
	}
	order := []int{}
	for i := 0; i < len(seats); {
		j := i + 1
		for j < len(seats) && scores[seats[j]] == scores[seats[i]] {
			j++
		}
		order = append(order, resolve(slices.Clone(seats[i:j]))...)
		i = j
	}
	// All draws, even royals, remain isolated until every rank is resolved.
	if len(used) > 0 {
		var e error
		n, e = n.Move(used, 0, Draw, true)
		if e != nil {
			return s, e
		}
		if plan {
			return n, nil
		}
		n, e = ShuffleSupply(n, returnOrder)
		if e != nil {
			return s, e
		}
	}
	n.Order = order
	if len(order) > 0 {
		n.Active = order[0]
	}
	return n, nil
}

// InitiativeReturnSupply plans the exact draw supply after isolated draw-offs.
// It consumes no randomness and leaves the input untouched.
func InitiativeReturnSupply(s State, draws []string) ([]string, error) {
	n, e := resolveInitiative(s, draws, nil, true)
	if e != nil {
		return nil, e
	}
	if !n.NeedsShuffle {
		return nil, nil
	}
	return slices.Clone(n.DrawOrder), nil
}
