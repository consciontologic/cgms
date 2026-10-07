package game

import "errors"

type DrawEntitlement struct {
	Seat      int `json:"seat"`
	Remaining int `json:"remaining"`
}

// QueueCompensation begins only after the original action and returned-card
// shuffle. It consumes each earned entitlement even if later supply is empty.
func QueueCompensation(s State) (State, error) {
	if s.Phase != "playing" || s.Pending != nil || s.NeedsShuffle || len(s.DrawQueue) > 0 {
		return s, errors.New("draw queue boundary")
	}
	n := s.Clone()
	for seat := 1; seat <= len(n.Players); seat++ {
		for i, e := range n.Effects {
			if e.Kind == "compensation" && e.Target == seat {
				q := e.Losses / 5
				n.Effects[i].Losses = e.Losses % 5
				if q > 0 {
					n.DrawQueue = append(n.DrawQueue, DrawEntitlement{seat, q})
				}
			}
		}
	}
	return n, nil
}
func StepCompensation(s State, recycle []string) (State, []string, error) {
	if validateDrawQueue(s) != nil || len(s.DrawQueue) == 0 || s.Phase != "playing" || s.Pending != nil {
		return s, nil, errors.New("no queued draw")
	}
	head := s.DrawQueue[0]
	n, got, e := DrawCards(s, head.Seat, 1, false, recycle)
	if e != nil {
		return s, nil, e
	}
	n.DrawQueue[0].Remaining--
	if n.DrawQueue[0].Remaining == 0 {
		n.DrawQueue = n.DrawQueue[1:]
	}
	return n, got, nil
}

// CloseCardBoard is the card-side atomic closure primitive. The authorized
// ending and ledger replacement are checked by the caller before committing it.
func CloseCardBoard(s State) (State, error) {
	n := s.Clone()
	n.Phase = "settlement"
	n.Pending = nil
	n.DrawQueue = nil
	n.Effects = nil
	for i := range n.Cards {
		n.Cards[i].AvailableFromRound = 0
		if n.Cards[i].Zone == ActiveAce {
			n.Cards[i].Zone = Discard
			n.Cards[i].Controller = 0
		}
	}
	if e := n.Validate(); e != nil {
		return s, e
	}
	return n, nil
}

// validateDrawQueue rejects malformed resume progress before any committed draw.
func validateDrawQueue(s State) error {
	previous := 0
	for _, e := range s.DrawQueue {
		if e.Seat <= previous || e.Seat > len(s.Players) || e.Remaining <= 0 {
			return errors.New("invalid deferred draw progress")
		}
		previous = e.Seat
	}
	return nil
}

// postActionWorkPending guards the interval before entitlement queue creation,
// as well as the queue itself. Responses already in flight must still resolve.
func postActionWorkPending(s State) bool {
	if s.NeedsShuffle || len(s.DrawQueue) > 0 {
		return true
	}
	for _, e := range s.Effects {
		if e.Kind == "compensation" && e.Losses >= 5 {
			return true
		}
	}
	return false
}
