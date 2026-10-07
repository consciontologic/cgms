package game

import "errors"

// answerClosedJustice records the recipient's explicit closed-selection choice.
// Empty eligible hidden custody yields no capture even if exposed cards exist;
// exposing hidden rank counts through the legal menu is unnecessary.
func answerClosedJustice(s State, c Command) (State, []Event, error) {
	if s.Pending == nil || s.Pending.Kind != "justice" || s.Pending.Decision == nil || c.WindowID != s.Pending.ID || c.DecisionID != s.Pending.Decision.ID || c.Actor != s.Pending.Decision.Actor || len(c.Cards) != 0 {
		return s, nil, errors.New("invalid closed Justice choice")
	}
	if e := validateRecordedChoices(s); e != nil {
		return s, nil, e
	}
	d := s.Pending.Decision
	hidden := 0
	for _, v := range d.Choices {
		card, _ := s.Card(v[0])
		if card.Zone == Hand || card.Zone == ConcealedAce {
			hidden++
		}
	}
	if hidden == 0 {
		if len(c.RandomWords) != 0 {
			return s, nil, errors.New("random words for empty hidden selection")
		}
		return justiceAnswer(s, c)
	}
	if len(c.RandomWords) == 0 {
		return s, nil, errors.New("closed Justice needs committed randomness")
	}
	c.Kind = "decision"
	return answerDecision(s, c)
}
