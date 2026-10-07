package sim

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"time"
)

type Scenario struct {
	Gameplay        *GameplayRun            `json:"gameplay,omitempty"`
	RunBots         bool                    `json:"run_bots,omitempty"`
	Schema          string                  `json:"schema_version"`
	ID              string                  `json:"id"`
	Kind            string                  `json:"kind"`
	Purpose         string                  `json:"purpose"`
	Players         int                     `json:"players"`
	RuleRefs        []string                `json:"rule_refs"`
	Ledger          *game.FinancialLedger   `json:"ledger,omitempty"`
	Receipts        []game.FinancialReceipt `json:"receipts,omitempty"`
	ExpectedScores  []game.Amount           `json:"expected_scores,omitempty"`
	State           *game.State             `json:"state,omitempty"`
	Commands        []game.Command          `json:"commands,omitempty"`
	ExpectedDigests []string                `json:"expected_digests,omitempty"`
	Checkpoint      *game.SettlementCursor  `json:"checkpoint,omitempty"`
	Parent          string                  `json:"parent,omitempty"`
}

func loadScenario(b []byte, players int) (Scenario, error) {
	var s Scenario
	if e := decodeRequired(b, &s); e != nil {
		return s, e
	}
	if s.Schema != "cgms-scenario-v1" || s.ID == "" || s.Purpose == "" || s.Players != players || len(s.RuleRefs) == 0 {
		return s, fmt.Errorf("invalid scenario identity")
	}
	switch s.Kind {
	case "gameplay-checkpoint":
		if s.Gameplay == nil || s.Parent == "" || s.Gameplay.Job.Population != players || s.Ledger != nil || s.State != nil || s.Checkpoint != nil || len(s.Commands) > 0 {
			return s, fmt.Errorf("invalid gameplay checkpoint")
		}
	case "accounting-only":
		if s.Ledger == nil || s.State != nil || len(s.Commands) > 0 || len(s.ExpectedScores) != players {
			return s, fmt.Errorf("invalid accounting scenario")
		}
		if e := s.Ledger.Validate(); e != nil {
			return s, e
		}
		if len(s.Ledger.Scores) != players {
			return s, fmt.Errorf("scenario population")
		}
	case "transition-sequence":
		if s.State == nil || s.Ledger != nil || len(s.Commands) != len(s.ExpectedDigests) {
			return s, fmt.Errorf("invalid transition scenario")
		}
		if e := s.State.Validate(); e != nil {
			return s, e
		}
		if len(s.State.Players) != players {
			return s, fmt.Errorf("scenario population")
		}
	case "settlement-checkpoint":
		if s.Checkpoint == nil || s.Parent == "" || len(s.ExpectedScores) > 0 && len(s.ExpectedScores) != players || s.Checkpoint != nil && len(s.Checkpoint.Ledger.Scores) != players {
			return s, fmt.Errorf("checkpoint provenance")
		}
		if _, _, e := game.ResumeSettlement(*s.Checkpoint, 0, true); e != nil {
			return s, e
		}
	default:
		return s, fmt.Errorf("unsupported scenario kind")
	}
	return s, nil
}

type Trace struct {
	BotDecision     *bots.Decision         `json:"bot_decision,omitempty"`
	BotRandomBefore *randomstream.Snapshot `json:"bot_random_before,omitempty"`
	BotRandomAfter  *randomstream.Snapshot `json:"bot_random_after,omitempty"`
	Schema          string                 `json:"schema_version"`
	Index           int                    `json:"index"`
	GameID          string                 `json:"game_id"`
	Kind            string                 `json:"kind"`
	Command         *game.Command          `json:"command,omitempty"`
	Shuffle         []string               `json:"shuffle,omitempty"`
	RandomBefore    string                 `json:"random_before"`
	RandomAfter     string                 `json:"random_after"`
	Pre             string                 `json:"pre"`
	Post            string                 `json:"post"`
	Events          string                 `json:"events"`
}
type Outcome struct {
	StartedInstances    int                    `json:"started_instances,omitempty"`
	VoidedGames         int                    `json:"voided_games,omitempty"`
	StartedGames        int                    `json:"started_games,omitempty"`
	BoardEndedGames     int                    `json:"board_ended_games,omitempty"`
	FinalizedGames      int                    `json:"finalized_games,omitempty"`
	MatchRanks          []int                  `json:"match_ranks,omitempty"`
	GameResults         []game.FinancialResult `json:"game_results,omitempty"`
	PlannedGameSlots    int                    `json:"planned_game_slots"`
	NotStartedGameSlots int                    `json:"not_started_game_slots"`
	Schema              string                 `json:"schema_version"`
	Index               int                    `json:"index"`
	Population          int                    `json:"population"`
	Block               string                 `json:"block"`
	Rotation            string                 `json:"rotation"`
	Seats               []string               `json:"seats"`
	Treatment           string                 `json:"treatment"`
	GameID              string                 `json:"game_id"`
	ScenarioStatus      string                 `json:"scenario_status"`
	GameComplete        bool                   `json:"game_complete"`
	MatchComplete       bool                   `json:"match_complete"`
	Status              string                 `json:"status"`
	ExitCode            int                    `json:"exit_code"`
	Reason              string                 `json:"reason"`
	Scores              []game.Amount          `json:"scores"`
	Cash                []game.Amount          `json:"cash"`
	Debts               []game.FinancialDebt   `json:"debts"`
	Steps               int                    `json:"steps"`
	Redeals             int                    `json:"redeals"`
	Observer            string                 `json:"observer"`
	Privacy             string                 `json:"privacy"`
}
type runResult struct {
	PrivateFinance *GameplayPrivateFinance
	PublicStats    *GameplayPublicStats
	Gameplay       *GameplayRun
	DecisionMicros []int64
	Outcome        Outcome
	Trace          []Trace
	Checkpoint     any
	Initial        any
	Derived        string
}

func hash(v any) string {
	h, e := canonical.Hash(v)
	if e != nil {
		panic(e)
	}
	return h
}
func runScenario(ctx context.Context, j Job, sc *Scenario, root, pairing string, budget int) runResult {
	return runScenarioLimited(ctx, j, sc, root, pairing, Budgets{MaxTransitions: budget, MaxSettlementSteps: budget, MaxBotOperations: 1000})
}
func runScenarioLimited(ctx context.Context, j Job, sc *Scenario, root, pairing string, limits Budgets) runResult {
	if sc == nil {
		return runFullGameplay(ctx, j, root, pairing, limits)
	}
	if sc.Kind == "gameplay-checkpoint" {
		return continueFullGameplay(ctx, j, root, pairing, limits, *sc.Gameplay)
	}
	budget := limits.MaxTransitions
	if sc != nil && (sc.Kind == "accounting-only" || sc.Kind == "settlement-checkpoint") {
		budget = limits.MaxSettlementSteps
	}

	o := Outcome{PlannedGameSlots: j.Games, NotStartedGameSlots: j.Games - 1, Schema: "cgms-outcome-v1", Index: j.Index, Population: j.Population, Block: j.Block, Rotation: j.Rotation.ID, Seats: j.Rotation.Seats, Treatment: j.Treatment, GameID: fmt.Sprintf("match-%06d/game-0/instance-0", j.Index), Status: "unsupported", ExitCode: 3, Reason: "full-game-loop not implemented", Observer: "public", Privacy: "public"}
	r := runResult{Outcome: o}
	stop := func(code int, reason string) runResult {
		if code == 130 && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			code = 5
			reason = "wall watchdog exhausted"
		}
		r.Outcome.ExitCode = code
		r.Outcome.Status = status(code)
		r.Outcome.Reason = reason
		r.Outcome.Steps = len(r.Trace)
		return r
	}
	if ctx.Err() != nil {
		return stop(130, "canceled before start")
	}
	if sc != nil && (sc.Kind == "accounting-only" || sc.Kind == "settlement-checkpoint") {
		var c game.SettlementCursor
		var e error
		if sc.Checkpoint != nil {
			c = *sc.Checkpoint
		} else {
			c, e = game.BeginSettlement(*sc.Ledger, sc.Receipts)
		}
		if e != nil {
			return stop(4, "invalid settlement input")
		}
		r.Initial = c
		r.Outcome.GameID = c.Ledger.GameID
		for !c.Done {
			if ctx.Err() != nil {
				r.Checkpoint = c
				return stop(130, "canceled at settlement boundary")
			}
			if len(r.Trace) >= budget {
				r.Checkpoint = c
				return stop(5, "settlement work budget")
			}
			before := hash(c)
			next, _, e := game.ResumeSettlement(c, 1, true)
			if e != nil {
				r.Checkpoint = c
				return stop(4, "settlement invariant")
			}
			if !fitsProgress(r, next, limits.MaxOutputBytes) {
				r.Checkpoint = c
				return stop(5, "output/checkpoint byte budget")
			}
			r.Trace = append(r.Trace, Trace{Schema: "cgms-trace-v1", Index: len(r.Trace), GameID: c.Ledger.GameID, Kind: "settlement", Pre: before, Post: hash(next), Events: hash(next.Audit[len(c.Audit):])})
			c = next
		}
		if sc.ExpectedScores != nil {
			for i, a := range sc.ExpectedScores {
				if c.Ledger.Scores[i].Cmp(a) != 0 {
					r.Checkpoint = c
					return stop(4, "expected ledger mismatch")
				}
			}
		}
		r.Checkpoint = c
		r.Outcome.Scores = c.Ledger.Scores
		r.Outcome.Cash = c.Ledger.Cash
		r.Outcome.Debts = c.Ledger.Debts
		r.Outcome.ScenarioStatus = "scenario-complete"
		return stop(0, "accounting fixture complete; game not finalized")
	}
	var s game.State
	if sc != nil {
		s = sc.State.Clone()
		r.Initial = s
		r.Outcome.GameID = s.GameID
		for i, c := range sc.Commands {
			if ctx.Err() != nil {
				r.Checkpoint = s
				return stop(130, "canceled at transition boundary")
			}
			if len(r.Trace) >= budget {
				r.Checkpoint = s
				return stop(5, "transition work budget")
			}
			n, ev, e := game.Apply(s, c)
			if e != nil {
				r.Checkpoint = s
				if errors.Is(e, game.ErrStateBudget) {
					return stop(5, "canonical checkpoint representation budget")
				}
				if errors.Is(e, game.ErrUnsupported) {
					return stop(3, "unsupported scenario transition "+c.Kind)
				}
				return stop(4, "rejected scenario command at committed prefix")
			}
			if !fitsProgress(r, n, limits.MaxOutputBytes) {
				r.Checkpoint = s
				return stop(5, "output/checkpoint byte budget")
			}
			tr := Trace{Schema: "cgms-trace-v1", Index: i, GameID: s.GameID, Kind: "command", Command: &c, Pre: hash(s), Post: hash(n), Events: hash(ev)}
			if sc.ExpectedDigests[i] != tr.Post {
				r.Checkpoint = s
				return stop(4, "expected transition digest mismatch")
			}
			r.Trace = append(r.Trace, tr)
			s = n
		}
		r.Checkpoint = s
		if sc.RunBots {
			streams := map[int]*randomstream.Stream{}
			for seat := 1; seat <= j.Population; seat++ {
				seed, e := randomstream.Derive(randomstream.Identity{Match: j.Match, Root: root, Population: j.Population, Pairing: pairing, Block: j.Block, Rotation: j.Rotation.ID, Kind: "bot", Seat: seat})
				if e != nil {
					return stop(2, "invalid bot stream")
				}
				streams[seat] = randomstream.New(seed)
			}
			for {
				if ctx.Err() != nil {
					return stop(130, "canceled at bot boundary")
				}
				if len(r.Trace) >= budget {
					return stop(5, "transition work budget")
				}
				started := time.Now()
				step, e := RunBotStep(s, j, streams, limits.MaxBotOperations)
				r.DecisionMicros = append(r.DecisionMicros, time.Since(started).Microseconds())
				if e != nil {
					if errors.Is(e, game.ErrStateBudget) {
						return stop(5, "canonical checkpoint representation budget")
					}
					if errors.Is(e, bots.ErrBudget) {
						return stop(5, "bot operation budget")
					}
					if errors.Is(e, ErrDisconnectedWait) {
						return stop(3, "disconnected required actor; waiting retained")
					}
					return stop(3, "unsupported narrow legal menu; no pass or end inferred")
				}
				if !fitsProgress(r, step.State, limits.MaxOutputBytes) {
					return stop(5, "output/checkpoint byte budget")
				}
				c := step.Decision.Command
				r.Trace = append(r.Trace, Trace{Schema: "cgms-trace-v1", Index: len(r.Trace), GameID: s.GameID, Kind: "command", Command: &c, BotDecision: &step.Decision, BotRandomBefore: &step.RandomBefore, BotRandomAfter: &step.RandomAfter, Pre: hash(s), Post: hash(step.State), Events: hash(step.Events)})
				s = step.State
				r.Checkpoint = s
			}
		}
		r.Outcome.ScenarioStatus = "scenario-complete"
		return stop(0, "accepted transition fixture complete; financial finalization not requested")
	}
	s, _ = game.NewState(j.Population, o.GameID)
	r.Initial = s
	seed, e := randomstream.Derive(randomstream.Identity{Match: j.Match, Root: root, Population: j.Population, Pairing: pairing, Block: j.Block, Rotation: j.Rotation.ID, Kind: "deal"})
	if e != nil {
		return stop(2, "invalid seed")
	}
	r.Derived = fmt.Sprintf("%x", seed)
	rng := randomstream.New(seed)
	for {
		if ctx.Err() != nil {
			r.Checkpoint = s
			return stop(130, "canceled at deal boundary")
		}
		if len(r.Trace) >= budget {
			r.Checkpoint = s
			return stop(5, "deal work budget")
		}
		order := append([]string(nil), s.DrawOrder...)
		before := fmt.Sprint(rng.Counter)
		if e = rng.Shuffle(len(order), func(i, k int) { order[i], order[k] = order[k], order[i] }); e != nil {
			return stop(4, "shuffle failure")
		}
		n, redo, e := game.DealAttempt(s, order)
		if e != nil {
			return stop(4, "deal invariant")
		}
		if !fitsProgress(r, n, limits.MaxOutputBytes) {
			r.Checkpoint = s
			return stop(5, "output/checkpoint byte budget")
		}
		r.Trace = append(r.Trace, Trace{Schema: "cgms-trace-v1", Index: len(r.Trace), GameID: s.GameID, Kind: "deal", Shuffle: order, RandomBefore: before, RandomAfter: fmt.Sprint(rng.Counter), Pre: hash(s), Post: hash(n), Events: hash(redo)})
		s = n
		if !redo {
			break
		}
		r.Outcome.Redeals++
	}
	r.Checkpoint = s
	return stop(3, "unsupported full-game-loop after accepted deal; no victory/draw inferred")
}
func traceBytes(ts []Trace) ([]byte, error) {
	var out []byte
	for _, t := range ts {
		b, e := canonical.Marshal(t)
		if e != nil {
			return nil, e
		}
		out = append(out, b...)
		out = append(out, '\n')
	}
	return out, nil
}
func decodeTrace(b []byte) ([]Trace, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	var out []Trace
	for d.More() {
		var raw json.RawMessage
		if e := d.Decode(&raw); e != nil {
			return nil, e
		}
		var t Trace
		if e := decodeRequired(raw, &t); e != nil {
			return nil, e
		}
		if t.Schema != "cgms-trace-v1" || t.Index != len(out) {
			return nil, fmt.Errorf("trace sequence")
		}
		out = append(out, t)
	}
	return out, nil
}

func fitsProgress(r runResult, next any, limit int) bool {
	a, e := canonical.Marshal(next)
	if e != nil {
		return false
	}
	if limit <= 0 {
		return true
	}
	b, e := canonical.Marshal(r.Initial)
	if e != nil {
		return false
	}
	tr, e := traceBytes(r.Trace)
	if e != nil {
		return false
	}
	return serializationBudget(limit-65536, a, b, tr) == nil
}
