package game

import (
	"errors"
	"fmt"
)

// resolveReady completes an accepted action whose eligible response traversal is
// already exhausted. It records no response, pass, policy intent or random outcome.
func resolveReady(s State, c Command) (State, []Event, error) {
	p := s.Pending
	if s.Phase != "playing" || p == nil || p.Actor != c.Actor || p.ID != c.WindowID || p.Cursor != len(p.Responders) || p.Decision != nil || len(s.DrawQueue) > 0 {
		return s, nil, errors.New("action is not ready for deterministic completion")
	}
	switch p.Kind {
	case "", "opening", "transfer", "loan-return", "justice", "ponzi":
	case "ability":
		if p.Ability == nil {
			return s, nil, errors.New("missing ability")
		}
		switch p.Ability.Kind {
		case "fate", "richer-sacrifice", "barricade-sacrifice", "kidnapper":
			return s, nil, errors.New("action requires recorded chance or selection")
		}
	default:
		return s, nil, fmt.Errorf("%w: ready completion %s", ErrUnsupported, p.Kind)
	}
	return finishCombat(s)
}
