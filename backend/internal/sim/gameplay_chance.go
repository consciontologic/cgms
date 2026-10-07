package sim

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"slices"
	"strconv"
)

func planInitiative(s game.State, streams map[string]*randomstream.Stream, st *GameplayStep) error {
	draws := slices.Clone(s.DrawOrder)
	discard := gameIDs(s, game.Discard)
	if len(discard) > 0 {
		draws = append(draws, recordShuffle(streams, "initiative", discard, st)...)
	}
	returned, e := game.InitiativeReturnSupply(s, draws)
	if e != nil {
		return e
	}
	var order []string
	if len(returned) > 0 {
		order = recordShuffle(streams, "initiative", returned, st)
	}
	if st.Operation != nil {
		st.Operation.InitiativeDraws = draws
		st.Operation.ReturnOrder = order
	} else {
		st.Shuffle = draws
		st.Recycle = order
	}
	return nil
}
func planPendingChance(s game.State, streams map[string]*randomstream.Stream, st *GameplayStep) error {
	p := s.Pending
	c := game.Command{ID: "chance/" + p.ID, GameID: s.GameID, Actor: p.Actor, WindowID: p.ID}
	var candidate game.State
	var e error
	var quantity int
	switch p.Kind {
	case "purchase":
		c.Kind = "purchase-outcome"
		candidate, e = s.Move(p.Targets, 0, game.Draw, false)
		quantity = p.Quantity
	case "ability":
		if p.Ability == nil {
			return errors.New("missing ability intent")
		}
		a := p.Ability
		c.Kind = "ability-outcome"
		switch a.Kind {
		case "kidnapper":
			c.Kind = "kidnapper-outcome"
			pool := game.KidnapperClosedCandidates(s, a.TargetSeat)
			if len(pool) > 0 {
				c.Cards = recordSample(streams, "effect", pool, st)
			} else if len(game.KidnapperOpenCandidates(s, a.TargetSeat)) > 0 {
				return errors.New("required exposed Kidnapper choice")
			}
			st.Kind = "command"
			st.Command = &c
			return nil
		case "richer-sacrifice":
			candidate, e = s.Move(a.Cards, 0, game.Discard, false)
			quantity = 2
		case "barricade-sacrifice":
			candidate, e = s.Move(a.Cards[:1], 0, game.Discard, false)
			if e == nil {
				candidate, e = candidate.Move(a.Cards[1:], 0, game.Draw, false)
			}
			quantity = 5
		case "fate":
			candidate, e = s.Move(a.Cards, 0, game.Draw, false)
			if e == nil {
				candidate, quantity, e = candidate.FateReturn(a.TargetSeat)
			}
		default:
			c.Kind = "resolve-ready"
			st.Kind = "command"
			st.Command = &c
			return nil
		}
	default:
		c.Kind = "resolve-ready"
		st.Kind = "command"
		st.Command = &c
		return nil
	}
	if e != nil {
		return e
	}
	if candidate.NeedsShuffle {
		c.Cards = recordShuffle(streams, "effect", gameIDs(candidate, game.Draw), st)
	}
	if quantity > len(candidate.DrawOrder) && len(gameIDs(candidate, game.Discard)) > 0 {
		c.Targets = recordShuffle(streams, "effect", gameIDs(candidate, game.Discard), st)
	}
	st.Kind = "command"
	st.Command = &c
	return nil
}

func verifyGameplayChance(m game.MatchLifecycle, st GameplayStep, streams map[string]*randomstream.Stream) error {
	copyStreams := map[string]*randomstream.Stream{}
	for k, r := range streams {
		p, _ := r.Snapshot()
		copyStreams[k], _ = randomstream.Restore(p)
	}
	expected := GameplayStep{Kind: st.Kind}
	s := m.Game.Board
	switch st.Kind {
	case "deal":
		expected.Shuffle = recordShuffle(copyStreams, "deal", gameIDs(s, game.Draw), &expected)
	case "initiative":
		if e := planInitiative(s, copyStreams, &expected); e != nil {
			return e
		}
	case "operation":
		if st.Operation == nil {
			return errors.New("missing operation")
		}
		op := *st.Operation
		expected.Operation = &op
		if op.Kind == "begin-round" {
			s.Round++
			if e := planInitiative(s, copyStreams, &expected); e != nil {
				return e
			}
			if !slices.Equal(op.InitiativeDraws, st.Operation.InitiativeDraws) || !slices.Equal(op.ReturnOrder, st.Operation.ReturnOrder) {
				return errors.New("initiative outcome detached")
			}
		}
		if op.Kind == "begin-turn" {
			op.Recycle = nil
			if len(s.DrawOrder) == 0 && len(gameIDs(s, game.Discard)) > 0 {
				op.Recycle = recordShuffle(copyStreams, "effect", gameIDs(s, game.Discard), &expected)
			}
			if !slices.Equal(op.Recycle, st.Operation.Recycle) {
				return errors.New("turn recycle detached")
			}
		}
	case "after-action":
		if s.NeedsShuffle {
			expected.Shuffle = recordShuffle(copyStreams, "effect", gameIDs(s, game.Draw), &expected)
		}
		if len(s.DrawOrder) == 0 && !s.NeedsShuffle && len(gameIDs(s, game.Discard)) > 0 {
			expected.Recycle = recordShuffle(copyStreams, "effect", gameIDs(s, game.Discard), &expected)
		}
	case "command":
		if st.Command == nil {
			return errors.New("missing command")
		}
		if (st.Command.Kind == "decision" || st.Command.Kind == "decision-closed") && len(st.Command.RandomWords) > 0 {
			words := recordWords(copyStreams, "effect", justiceHiddenPool(s), &expected)
			if !slices.Equal(words, st.Command.RandomWords) {
				return errors.New("hidden selection detached")
			}
		}
		if st.Command.Kind == "resolve-ready" || st.Command.Kind == "purchase-outcome" || st.Command.Kind == "ability-outcome" || st.Command.Kind == "kidnapper-outcome" && st.BotBefore == nil {
			if e := planPendingChance(s, copyStreams, &expected); e != nil {
				return e
			}
			if hash(expected.Command) != hash(st.Command) {
				return errors.New("command outcome detached")
			}
		}
	}
	if hash(expected.Random) != hash(st.Random) || !slices.Equal(expected.Shuffle, st.Shuffle) || !slices.Equal(expected.Recycle, st.Recycle) {
		return errors.New("random outcome not bound to transition")
	}
	return nil
}

func justiceHiddenPool(s game.State) []string {
	out := []string{}
	if s.Pending == nil || s.Pending.Decision == nil || s.Pending.Decision.Kind != "justice" {
		return out
	}
	for _, v := range s.Pending.Decision.Choices {
		if len(v) != 1 {
			continue
		}
		c, _ := s.Card(v[0])
		if c.Zone == game.Hand || c.Zone == game.ConcealedAce {
			out = append(out, v[0])
		}
	}
	return out
}
func recordWords(streams map[string]*randomstream.Stream, key string, input []string, st *GameplayStep) []string {
	r := streams[key]
	before, _ := r.Snapshot()
	words := []string{}
	_, _ = randomstream.Sample(func() uint64 { x := r.Uint64(); words = append(words, strconv.FormatUint(x, 10)); return x }, uint64(len(input)))
	after, _ := r.Snapshot()
	st.Random = append(st.Random, GameplayRandom{key, "words", slices.Clone(input), words, before, after})
	return words
}

func recordSample(streams map[string]*randomstream.Stream, key string, input []string, st *GameplayStep) []string {
	r := streams[key]
	before, _ := r.Snapshot()
	j, _ := r.Sample(uint64(len(input)))
	out := []string{input[j]}
	after, _ := r.Snapshot()
	st.Random = append(st.Random, GameplayRandom{key, "sample", slices.Clone(input), out, before, after})
	return out
}
