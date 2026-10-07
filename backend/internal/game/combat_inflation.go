package game

import "errors"

// CompleteSpadeSpend is the committed monetary transaction boundary. It values
// payment while Inflation is active, then clears only a single qualifying
// voluntary transaction. Forced payments never clear Inflation.
func CompleteSpadeSpend(s State, seat int, ids []string, voluntary, completed bool) (State, Amount, error) {
	if !selection(s, ids, seat, Spades, false) {
		return s, Amount{}, errors.New("invalid spend")
	}
	for _, id := range ids {
		c, _ := s.Card(id)
		if c.AvailableFromRound > s.Round {
			return s, Amount{}, errors.New("restricted payment")
		}
	}
	value := sumCards(s, ids)
	if !completed {
		return s.Clone(), value, nil
	}
	printed := 0
	for _, id := range ids {
		c, _ := s.Card(id)
		printed += c.Card.Rank
	}
	n := s.Clone()
	if voluntary && printed >= 20 {
		for i, e := range n.Effects {
			if e.Kind == "inflation" && e.Target == seat {
				for j := range n.Cards {
					if n.Cards[j].Card.ID == e.CardID {
						n.Cards[j].Zone = Discard
						n.Cards[j].Controller = 0
					}
				}
				n.Effects = append(n.Effects[:i], n.Effects[i+1:]...)
				break
			}
		}
	}
	return n, value, nil
}
