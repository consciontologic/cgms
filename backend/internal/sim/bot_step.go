package sim

import (
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"strings"
)

type JobPolicySettings struct {
	Budgets Budgets
	Weights []bots.Weights
}

// ResolveJob resolves operation budgets for both paired treatments, and applies
// bot feature sweeps only to the focal candidate. Other opponents remain fixed.
func ResolveJob(j Job, b Budgets) (JobPolicySettings, error) {
	out := JobPolicySettings{Budgets: b}
	for _, p := range j.Policies {
		if e := p.Validate(); e != nil {
			return out, e
		}
		w := bots.DefaultWeights()
		for i, f := range p.Features {
			if e := setWeight(&w, f, p.Weights[i]); e != nil {
				return out, e
			}
		}
		out.Weights = append(out.Weights, w)
	}
	for k, v := range j.Parameters {
		switch k {
		case "budget.transitions":
			if v < 1 || v > 1000000 {
				return out, fmt.Errorf("transition budget")
			}
			out.Budgets.MaxTransitions = v
		case "budget.bot-operations":
			if v < 1 || v > 1000000 {
				return out, fmt.Errorf("bot budget")
			}
			out.Budgets.MaxBotOperations = v
		default:
			if !strings.HasPrefix(k, "bot.") || v < -10000 || v > 10000 || j.FocalParticipantID == "" {
				return out, fmt.Errorf("invalid grid parameter")
			}
			check := bots.DefaultWeights()
			if e := setWeight(&check, strings.TrimPrefix(k, "bot."), v); e != nil {
				return out, e
			}
			found := false
			for i, p := range j.Policies {
				if p.ParticipantID == j.FocalParticipantID {
					found = true
					if j.Treatment == "candidate" {
						if e := setWeight(&out.Weights[i], strings.TrimPrefix(k, "bot."), v); e != nil {
							return out, e
						}
					}
				}
			}
			if !found {
				return out, fmt.Errorf("focal participant absent")
			}
		}
	}
	if out.Budgets.MaxTransitions < 1 || out.Budgets.MaxBotOperations < 1 {
		return out, fmt.Errorf("nonpositive job budget")
	}
	return out, nil
}
func setWeight(w *bots.Weights, key string, v int) error {
	switch key {
	case "immediate-score":
		w.ImmediateScore = v
	case "exposed-holdings":
		w.ExposedHoldings = v
	case "protection":
		w.Protection = v
	case "known-debt":
		w.KnownDebt = v
	case "legal-declaration":
		w.LegalDeclaration = v
	default:
		return fmt.Errorf("unknown heuristic feature")
	}
	return nil
}

var ErrDisconnectedWait = errors.New("disconnected required actor; decision remains pending")

type BotStep struct {
	State        game.State
	Events       []game.Event
	Decision     bots.Decision
	Actor        int
	RandomBefore randomstream.Snapshot
	RandomAfter  randomstream.Snapshot
}

// RunBotStep runs one supported, independently adjudicated command. It does not
// infer EndTurn, refusal or victory from an empty menu. RNG ownership stays with
// the caller; exact replay consumes Decision.Command, never reruns this policy.
func RunBotStep(s game.State, j Job, streams map[int]*randomstream.Stream, budget int) (BotStep, error) {
	out := BotStep{State: s}
	actor := s.Active
	if s.Pending != nil {
		p := s.Pending
		if p.Decision != nil {
			actor = p.Decision.Actor
		} else if p.Cursor < len(p.Responders) {
			actor = p.Responders[p.Cursor]
		}
	}
	if actor < 1 || actor > len(s.Players) || actor > len(j.Policies) {
		return out, fmt.Errorf("invalid bot actor")
	}
	if !s.Players[actor-1].Connected {
		return out, ErrDisconnectedWait
	}
	o, e := game.Observe(s, actor)
	if e != nil {
		return out, e
	}
	cfg, e := ResolveJob(j, Budgets{MaxTransitions: 1, MaxBotOperations: budget})
	if e != nil {
		return out, e
	}
	rng := streams[actor]
	if rng == nil {
		return out, fmt.Errorf("missing private bot stream")
	}
	out.RandomBefore, e = rng.Snapshot()
	if e != nil {
		return out, e
	}
	p := j.Policies[actor-1]
	d, e := bots.ChooseWeighted(p.Policy+"@"+p.Version, o, rng, cfg.Budgets.MaxBotOperations, cfg.Weights[actor-1])
	out.RandomAfter, _ = rng.Snapshot()
	out.Actor = actor
	if e != nil {
		return out, e
	}
	n, events, e := game.Apply(s, d.Command)
	if e != nil {
		return out, e
	}
	out.State = n
	out.Events = events
	out.Decision = d
	return out, nil
}
