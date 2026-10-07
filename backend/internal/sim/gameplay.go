package sim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"slices"
	"strconv"
)

const GameplayVersion = "cgms-gameplay-transcript-v1"

type GameplayPolicy func(string, game.Observation, *randomstream.Stream, int) (bots.Decision, error)
type GameplayOptions struct {
	financialDeclines   map[int]string
	Job                 Job
	Root, Pairing       string
	MaxSteps, BotBudget int
	MenuBudget          int
	MaxOutputBytes      int
	MaxTranscriptBytes  int
	Policy              GameplayPolicy
	PolicyLabel         string
	Resume              *GameplayRun
}
type GameplayRandom struct {
	Stream string                `json:"stream"`
	Kind   string                `json:"kind"`
	Input  []string              `json:"input"`
	Output []string              `json:"output"`
	Before randomstream.Snapshot `json:"before"`
	After  randomstream.Snapshot `json:"after"`
}
type GameplayStep struct {
	OpportunityKey  string                        `json:"opportunity_key,omitempty"`
	FinanceDecision *bots.GameplayFinanceDecision `json:"finance_decision,omitempty"`
	Index           int                           `json:"index"`
	Kind            string                        `json:"kind"`
	GameID          string                        `json:"game_id"`
	Pre             string                        `json:"pre"`
	Post            string                        `json:"post"`
	Events          string                        `json:"events"`
	Operation       *game.LifecycleOperation      `json:"operation,omitempty"`
	Command         *game.Command                 `json:"command,omitempty"`
	Shuffle         []string                      `json:"shuffle,omitempty"`
	Recycle         []string                      `json:"recycle,omitempty"`
	NextID          string                        `json:"next_id,omitempty"`
	Random          []GameplayRandom              `json:"random"`
	Decision        *bots.Decision                `json:"decision,omitempty"`
	BotBefore       *randomstream.Snapshot        `json:"bot_before,omitempty"`
	BotAfter        *randomstream.Snapshot        `json:"bot_after,omitempty"`
}
type GameplayRun struct {
	ArtifactProfile string                           `json:"artifact_profile,omitempty"`
	PolicyOverride  string                           `json:"policy_override,omitempty"`
	Schema          string                           `json:"schema"`
	Job             Job                              `json:"job"`
	Root            string                           `json:"root"`
	Pairing         string                           `json:"pairing"`
	Initial         game.MatchLifecycle              `json:"initial"`
	Match           game.MatchLifecycle              `json:"checkpoint"`
	Trace           []GameplayStep                   `json:"trace"`
	ExitCode        int                              `json:"exit_code"`
	Reason          string                           `json:"reason"`
	Redeals         int                              `json:"redeals"`
	Streams         map[string]randomstream.Snapshot `json:"streams"`
}

func gameplayStreams(o GameplayOptions, slot int) (map[string]*randomstream.Stream, error) {
	out := map[string]*randomstream.Stream{}
	for _, kind := range []string{"deal", "initiative", "effect"} {
		seed, e := randomstream.Derive(randomstream.Identity{Root: o.Root, Pairing: o.Pairing, Block: o.Job.Block, Rotation: o.Job.Rotation.ID, Population: o.Job.Population, Match: o.Job.Match, Slot: slot, Kind: kind})
		if e != nil {
			return nil, e
		}
		out[kind] = randomstream.New(seed)
	}
	for seat := 1; seat <= o.Job.Population; seat++ {
		seed, e := randomstream.Derive(randomstream.Identity{Root: o.Root, Pairing: o.Pairing, Block: o.Job.Block, Rotation: o.Job.Rotation.ID, Population: o.Job.Population, Match: o.Job.Match, Slot: slot, Kind: "bot", Seat: seat})
		if e != nil {
			return nil, e
		}
		out[fmt.Sprint("bot-", seat)] = randomstream.New(seed)
	}
	return out, nil
}
func gameIDs(s game.State, zone game.Zone) []string {
	out := []string{}
	for _, c := range s.Cards {
		if c.Zone == zone {
			out = append(out, c.Card.ID)
		}
	}
	return out
}
func recordShuffle(streams map[string]*randomstream.Stream, key string, input []string, step *GameplayStep) []string {
	out := slices.Clone(input)
	r := streams[key]
	before, _ := r.Snapshot()
	_ = r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	after, _ := r.Snapshot()
	step.Random = append(step.Random, GameplayRandom{key, "shuffle", slices.Clone(input), slices.Clone(out), before, after})
	return out
}
func postWork(s game.State) bool {
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
func snapshotStreams(streams map[string]*randomstream.Stream) map[string]randomstream.Snapshot {
	out := map[string]randomstream.Snapshot{}
	for k, v := range streams {
		out[k], _ = v.Snapshot()
	}
	return out
}

func RunGameplay(ctx context.Context, o GameplayOptions) GameplayRun {
	r := GameplayRun{PolicyOverride: o.PolicyLabel, Schema: GameplayVersion, Job: o.Job, Root: o.Root, Pairing: o.Pairing, ExitCode: 3}
	if o.Policy != nil && o.PolicyLabel == "" {
		r.ExitCode = 2
		r.Reason = "custom policy requires frozen label"
		return r
	}
	if o.MaxTranscriptBytes != 0 && o.MaxTranscriptBytes != canonical.MaxBytes && o.MaxTranscriptBytes != 32<<20 {
		r.ExitCode = 2
		r.Reason = "invalid gameplay transcript budget"
		return r
	}
	if o.MaxTranscriptBytes == 32<<20 {
		r.ArtifactProfile = GameplayLargeArtifactProfile
	}
	if o.MaxSteps < 1 || o.BotBudget < 1 || o.Job.Games < 1 || len(o.Job.Policies) != o.Job.Population {
		r.ExitCode = 2
		r.Reason = "invalid gameplay limits/policies"
		return r
	}
	s, e := game.NewState(o.Job.Population, fmt.Sprintf("match-%06d/game-0/instance-0", o.Job.Index))
	if e != nil {
		r.ExitCode = 2
		r.Reason = e.Error()
		return r
	}
	r.Match, e = game.NewMatchLifecycle(s, o.Job.Games)
	if e != nil {
		r.ExitCode = 2
		r.Reason = e.Error()
		return r
	}
	r.Initial = r.Match
	streams, e := gameplayStreams(o, 0)
	if e != nil {
		r.ExitCode = 2
		r.Reason = e.Error()
		return r
	}
	if o.Resume != nil {
		if e = ReplayGameplay(*o.Resume); e != nil {
			r.ExitCode = 4
			r.Reason = e.Error()
			return r
		}
		r = *o.Resume
		r.Trace = slices.Clone(r.Trace)
		r.ExitCode = 3
		r.Reason = "resuming"
		for k, v := range r.Streams {
			streams[k], e = randomstream.Restore(v)
			if e != nil {
				r.ExitCode = 4
				r.Reason = e.Error()
				return r
			}
		}
	}
	transcriptLimit := canonical.MaxBytes
	if o.MaxTranscriptBytes != 0 {
		transcriptLimit = o.MaxTranscriptBytes
	}
	if o.Resume != nil && gameplayTranscriptLimit(*o.Resume) != transcriptLimit {
		r.ExitCode = 2
		r.Reason = "continuation transcript budget drift"
		return r
	}
	maxBytes := o.MaxOutputBytes
	if maxBytes == 0 || maxBytes > transcriptLimit {
		maxBytes = transcriptLimit
	}
	initialBytes, _ := json.Marshal(r.Initial)
	traceBytes := 0
	for _, old := range r.Trace {
		b, _ := json.Marshal(old)
		traceBytes += len(b) + 1
	}
	choose := o.Policy
	if choose == nil {
		choose = bots.ChooseGameplay
	}
	for len(r.Trace) < o.MaxSteps {
		if ctx.Err() != nil {
			r.ExitCode = 130
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				r.ExitCode = 5
			}
			r.Reason = "canceled at gameplay boundary"
			break
		}
		if len(r.Match.Game.Ledger.Completed) == r.Match.GameLimit {
			r.ExitCode = 0
			r.Reason = "fixed match financially finalized"
			break
		}
		step := GameplayStep{Index: len(r.Trace), GameID: r.Match.Game.Board.GameID, Pre: hash(r.Match)}
		saved := snapshotStreams(streams)
		declined := []int{}
		for i := len(r.Trace) - 1; i >= 0; i-- {
			old := r.Trace[i]
			if old.Kind != "policy-wait" || old.Pre != step.Pre {
				break
			}
			if old.Command != nil {
				declined = append(declined, old.Command.Actor)
			}
			if old.Operation != nil {
				declined = append(declined, -old.Operation.Actor)
			}
		}
		o.financialDeclines = map[int]string{}
		for _, old := range r.Trace {
			if old.Kind == "policy-wait" && old.Operation != nil && old.OpportunityKey != "" {
				o.financialDeclines[old.Operation.Actor] = old.OpportunityKey
			}
		}
		e = planGameplayStep(r.Match, streams, choose, o, &step, declined)
		if e != nil {
			for k, v := range saved {
				streams[k], _ = randomstream.Restore(v)
			}
			r.ExitCode = 3
			if errors.Is(e, bots.ErrBudget) || errors.Is(e, game.ErrMenuBudget) {
				r.ExitCode = 5
			}
			r.Reason = e.Error()
			break
		}
		next, events, e := applyGameplayStep(r.Match, step)
		if e != nil {
			for k, v := range saved {
				streams[k], _ = randomstream.Restore(v)
			}
			r.ExitCode = 4
			if errors.Is(e, game.ErrUnsupported) {
				r.ExitCode = 3
			}
			if errors.Is(e, game.ErrStateBudget) {
				r.ExitCode = 5
			}
			r.Reason = "rejected gameplay transition: " + e.Error()
			break
		}
		step.Events = hash(events)
		step.Post = hash(next)
		stepBytes, _ := json.Marshal(step)
		checkpointBytes, _ := json.Marshal(next)
		if len(initialBytes)+traceBytes+len(stepBytes)+len(checkpointBytes)+65536 > maxBytes {
			for k, v := range saved {
				streams[k], _ = randomstream.Restore(v)
			}
			r.ExitCode = 5
			r.Reason = "gameplay output byte budget; retained committed prefix"
			break
		}
		traceBytes += len(stepBytes) + 1
		r.Trace = append(r.Trace, step)
		r.Match = next
		if step.Kind == "deal" && len(next.Game.Board.Order) == 0 && next.Game.Board.Players[0].History == nil {
			r.Redeals++
		}
		if step.Kind == "next-game" {
			streams, e = gameplayStreams(o, len(next.Instances)-1)
			if e != nil {
				r.ExitCode = 2
				r.Reason = e.Error()
				break
			}
		}
	}
	if len(r.Match.Game.Ledger.Completed) == r.Match.GameLimit {
		r.ExitCode = 0
		r.Reason = "fixed match financially finalized"
	}
	if len(r.Trace) >= o.MaxSteps && r.ExitCode != 0 {
		r.ExitCode = 5
		r.Reason = "gameplay transition budget; no ending inferred"
	}
	r.Streams = snapshotStreams(streams)
	return r
}

func planGameplayStep(m game.MatchLifecycle, streams map[string]*randomstream.Stream, choose GameplayPolicy, o GameplayOptions, st *GameplayStep, declined []int) error {
	originalChoose := choose
	choose = func(policy string, obs game.Observation, rng *randomstream.Stream, budget int) (bots.Decision, error) {
		if o.Policy == nil && (policy == "bargaining@v1" || policy == "opportunity@v1" || policy == "opportunity-no-retain@v1" || policy == "opportunity-no-finance@v1" || bots.GameplayOrderV2(policy)) {
			return bots.ChooseGameplayWithEconomy(policy, obs, projectPolicyEconomy(m, obs.Seat), rng, budget)
		}
		return originalChoose(policy, obs, rng, budget)
	}

	l := m.Game
	s := l.Board
	op := func(kind string, actor int) {
		st.Kind = "operation"
		st.Operation = &game.LifecycleOperation{ID: fmt.Sprintf("gameplay/%d", st.Index), GameID: s.GameID, Kind: kind, Actor: actor}
	}
	if l.Ledger.Phase == "finalized" || l.Ledger.Phase == "void" {
		st.Kind = "next-game"
		st.NextID = fmt.Sprintf("match-%06d/game-%d/instance-%d", o.Job.Index, len(l.Ledger.Completed), len(m.Instances))
		return nil
	}
	if len(s.Players[0].History) == 0 && len(s.Order) == 0 {
		st.Kind = "deal"
		st.Shuffle = recordShuffle(streams, "deal", gameIDs(s, game.Draw), st)
		return nil
	}
	if len(s.Order) == 0 {
		st.Kind = "initiative"
		return planInitiative(s, streams, st)
	}
	if l.Automatic != nil {
		op("resume-automatic", s.Active)
		st.Operation.Budget = 128
		return nil
	}
	if l.Ending != "" {
		unfinished := false
		for i, f := range l.Ledger.Finished {
			if f {
				continue
			}
			unfinished = true
			obs, e := bots.ProjectGameplayFinance(l, i+1)
			if e != nil {
				return e
			}
			if len(obs.Legal) == 0 {
				continue
			}
			policy := o.Job.Policies[i].Policy + "@" + o.Job.Policies[i].Version
			rng := streams[fmt.Sprint("bot-", i+1)]
			before, _ := rng.Snapshot()
			d, e := bots.ChooseGameplayFinancial(policy, obs, rng, o.BotBudget)
			if e != nil {
				return e
			}
			after, _ := rng.Snapshot()
			operation := d.Operation
			operation.ID = fmt.Sprintf("gameplay/%d", st.Index)
			st.Kind = "operation"
			st.Operation = &operation
			st.FinanceDecision = &d
			st.BotBefore = &before
			st.BotAfter = &after
			return nil
		}
		if unfinished {
			return errors.New("required financial decisions remain waiting")
		}
		op("finalize", s.Active)
		return nil
	}
	if handled, e := planOutOfTurn(l, streams, choose, o, st, declined); handled || e != nil {
		return e
	}
	if handled, e := planPlayingFinance(m, streams, o, st, declined); handled || e != nil {
		return e
	}
	if s.Pending == nil && (postWork(s) || l.EndTurnPending || len(l.PendingReceipts) > 0) {
		st.Kind = "after-action"
		if s.NeedsShuffle {
			st.Shuffle = recordShuffle(streams, "effect", gameIDs(s, game.Draw), st)
		}
		if len(s.DrawOrder) == 0 && !s.NeedsShuffle && len(gameIDs(s, game.Discard)) > 0 {
			st.Recycle = recordShuffle(streams, "effect", gameIDs(s, game.Discard), st)
		}
		return nil
	}
	if l.RoundClosed {
		if !l.DeparturesResolved {
			for _, p := range s.Players {
				answered := p.Departed
				for _, answer := range l.DepartureChoices {
					if answer.Seat == p.Seat {
						answered = true
					}
				}
				if answered {
					continue
				}
				if !p.Connected {
					return errors.New("departure choice waiting for disconnected participant")
				}
				obs, e := bots.ProjectGameplayFinance(l, p.Seat)
				if e != nil {
					return e
				}
				obs.Legal = nil
				for _, leave := range []bool{false, true} {
					candidate := game.LifecycleOperation{GameID: s.GameID, ID: fmt.Sprintf("gameplay/%d", st.Index), Actor: p.Seat, Kind: "departure-choice", Accept: leave}
					if _, e := m.ApplyOperation(candidate); e == nil {
						obs.Legal = append(obs.Legal, candidate)
					}
				}
				return recordFinanceChoice(obs, streams, o, st)
			}
			return errors.New("departure choices unresolved")
		}
		op("begin-round", s.Active)
		next := s.Clone()
		next.Round++
		return planInitiative(next, streams, st)
	}
	if !l.TurnStarted {
		op("begin-turn", s.Active)
		if len(s.DrawOrder) == 0 && len(gameIDs(s, game.Discard)) > 0 {
			st.Operation.Recycle = recordShuffle(streams, "effect", gameIDs(s, game.Discard), st)
		}
		return nil
	}
	if p := s.Pending; p != nil && p.Decision == nil && p.Cursor == len(p.Responders) {
		if !(p.Kind == "ability" && p.Ability != nil && p.Ability.Kind == "kidnapper" && len(game.KidnapperClosedCandidates(s, p.Ability.TargetSeat)) == 0 && len(game.KidnapperOpenCandidates(s, p.Ability.TargetSeat)) > 0) {
			return planPendingChance(s, streams, st)
		}
	}
	actor := s.Active
	if p := s.Pending; p != nil && p.Kind == "ability" && p.Ability != nil && p.Ability.Kind == "kidnapper" && p.Cursor == len(p.Responders) {
		actor = p.Actor
	}
	if p := s.Pending; p != nil {
		if p.Decision != nil {
			actor = p.Decision.Actor
		} else if p.Cursor < len(p.Responders) {
			actor = p.Responders[p.Cursor]
		}
	}
	if actor < 1 || actor > len(s.Players) || !s.Players[actor-1].Connected {
		return errors.New("disconnected required actor; wait retained")
	}
	menuBudget := o.MenuBudget
	if menuBudget == 0 {
		menuBudget = 100000
	}
	obs, e := game.GameplayObservation(s, actor, menuBudget)
	if e != nil {
		return e
	}
	rng := streams[fmt.Sprint("bot-", actor)]
	before, _ := rng.Snapshot()
	policy := o.Job.Policies[actor-1].Policy + "@" + o.Job.Policies[actor-1].Version
	d, e := choose(policy, obs, rng, o.BotBudget)
	if e != nil {
		return e
	}
	expectedPolicy := policy
	if o.PolicyLabel != "" {
		expectedPolicy = o.PolicyLabel
	}
	if d.Policy != expectedPolicy {
		return errors.New("policy version does not match frozen profile")
	}
	if d.Command.Actor != actor || d.Command.GameID != s.GameID {
		return errors.New("policy selected unauthorized actor/game")
	}
	after, _ := rng.Snapshot()
	st.BotBefore = &before
	st.BotAfter = &after
	st.Decision = &d
	c := d.Command
	c.ID = fmt.Sprintf("gameplay/%d", st.Index)
	c.GameID = s.GameID
	c.Actor = actor
	if c.Kind == "decision-closed" {
		c.Cards = nil
		pool := justiceHiddenPool(s)
		if len(pool) > 0 {
			c.RandomWords = recordWords(streams, "effect", pool, st)
		}
	}
	switch c.Kind {
	case "end-turn":
		op("end-turn", actor)
	case "ordinary-victory":
		op("declare-ordinary", actor)
		st.Operation.Threshold = string(c.Suit)
		if c.Value == 15 {
			st.Operation.Threshold = "royals"
		} else if st.Operation.Threshold == "" {
			st.Operation.Threshold = "diamonds"
		}
	default:
		st.Kind = "command"
		st.Command = &c
	}
	return nil
}
func applyGameplayStep(m game.MatchLifecycle, s GameplayStep) (game.MatchLifecycle, []game.Event, error) {
	n := m
	n.Game = m.Game.Clone()
	events := []game.Event{}
	var e error
	switch s.Kind {
	case "deal":
		n.Game.Board, _, e = game.DealAttempt(n.Game.Board, s.Shuffle)
	case "initiative":
		n.Game.Board, e = game.ResolveInitiative(n.Game.Board, s.Shuffle, s.Recycle)
	case "operation":
		if s.Operation == nil {
			return m, nil, errors.New("missing operation")
		}
		n, e = n.ApplyOperation(*s.Operation)
	case "command":
		if s.Command == nil {
			return m, nil, errors.New("missing command")
		}
		n.Game, events, e = n.Game.ApplyBoardCommand(*s.Command)
	case "after-action":
		n.Game, e = n.Game.ContinueAfterAction(s.Shuffle, s.Recycle)
	case "policy-wait":
		if s.Operation != nil && s.Operation.Kind == "decline-financial-opportunity" && s.Operation.Actor >= 1 && s.Operation.Actor <= len(m.Game.Board.Players) && s.Operation.GameID == m.Game.Board.GameID {
			return m, events, nil
		}
		if s.Command == nil || s.Command.Kind != "decline-out-of-turn" || s.Command.Actor < 1 || s.Command.Actor > len(m.Game.Board.Players) {
			return m, nil, errors.New("invalid optional policy decision")
		}
		return m, events, nil
	case "next-game":
		n, e = n.NextGame(s.NextID)
	default:
		e = errors.New("unsupported transcript kind")
	}
	if e != nil {
		return m, nil, e
	}
	return n, events, nil
}

func ReplayGameplay(r GameplayRun) error { return ReplayGameplayContext(context.Background(), r) }
func ReplayGameplayContext(ctx context.Context, r GameplayRun) error {
	if r.ExitCode == 0 && len(r.Match.Game.Ledger.Completed) != r.Match.GameLimit {
		return errors.New("fabricated complete gameplay outcome")
	}
	if r.Schema != GameplayVersion || gameplayTranscriptLimit(r) == 0 || r.Initial.Game.Validate() != nil {
		return errors.New("gameplay transcript version/initial state")
	}
	initialBoard, e := game.NewState(r.Job.Population, fmt.Sprintf("match-%06d/game-0/instance-0", r.Job.Index))
	if e != nil {
		return e
	}
	initial, e := game.NewMatchLifecycle(initialBoard, r.Job.Games)
	if e != nil || hash(initial) != hash(r.Initial) {
		return errors.New("initial gameplay provenance mismatch")
	}
	o := GameplayOptions{Job: r.Job, Root: r.Root, Pairing: r.Pairing}
	streams, e := gameplayStreams(o, 0)
	if e != nil {
		return e
	}
	m := r.Initial
	for i, st := range r.Trace {
		requiredDecision := st.Kind == "policy-wait" && st.Operation == nil || st.Kind == "command" && st.Command != nil && st.Command.Kind != "resolve-ready" && st.Command.Kind != "purchase-outcome" && st.Command.Kind != "ability-outcome" && st.Command.Kind != "kidnapper-outcome" || st.Kind == "operation" && st.Operation != nil && (st.Operation.Kind == "end-turn" || st.Operation.Kind == "declare-ordinary")
		if st.Operation != nil && (st.Operation.Kind == "departure-choice" || st.Operation.Kind == "nullify" || st.Operation.Kind == "decline-financial-opportunity" || st.Operation.Kind == "promise-offer" || st.Operation.Kind == "promise-accept" || st.Operation.Kind == "finish-settlement" || st.Operation.Kind == "promise-pay" || st.Operation.Kind == "promise-refuse" || st.Operation.Kind == "promise-final-offer" || st.Operation.Kind == "promise-final-answer" || st.Operation.Kind == "voluntary-transfer" || st.Operation.Kind == "forgive") && (st.FinanceDecision == nil || st.BotBefore == nil || st.BotAfter == nil) {
			return errors.New("missing financial policy decision")
		}
		if requiredDecision && (st.Decision == nil || st.BotBefore == nil || st.BotAfter == nil) {
			return errors.New("missing required recorded policy decision")
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if st.Index != i || st.GameID != m.Game.Board.GameID || st.Pre != hash(m) {
			return fmt.Errorf("gameplay pre digest %d", i)
		}
		if st.OpportunityKey != "" {
			if st.Operation == nil {
				return errors.New("financial opportunity missing operation")
			}
			obs, e := bots.ProjectGameplayFinance(m.Game, st.Operation.Actor)
			if e != nil {
				return e
			}
			if financeOpportunityKey(m, obs) != st.OpportunityKey {
				return errors.New("financial opportunity provenance")
			}
		}
		if e := verifyGameplayChance(m, st, streams); e != nil {
			return fmt.Errorf("chance binding %d: %w", i, e)
		}
		for _, v := range st.Random {
			rng := streams[v.Stream]
			if rng == nil {
				return errors.New("unknown chance stream")
			}
			before, _ := rng.Snapshot()
			if before != v.Before {
				return errors.New("chance before snapshot")
			}
			out := slices.Clone(v.Input)
			switch v.Kind {
			case "shuffle":
				_ = rng.Shuffle(len(out), func(a, b int) { out[a], out[b] = out[b], out[a] })
			case "sample":
				j, e := rng.Sample(uint64(len(out)))
				if e != nil {
					return e
				}
				out = []string{out[j]}
			case "words":
				words := []string{}
				_, e := randomstream.Sample(func() uint64 { x := rng.Uint64(); words = append(words, strconv.FormatUint(x, 10)); return x }, uint64(len(out)))
				if e != nil {
					return e
				}
				out = words
			default:
				return errors.New("unknown random outcome")
			}
			after, _ := rng.Snapshot()
			if after != v.After || !slices.Equal(out, v.Output) {
				return errors.New("recorded chance mismatch")
			}
		}
		if st.BotBefore != nil || st.BotAfter != nil {
			if st.BotBefore == nil || st.BotAfter == nil || (st.Decision == nil) == (st.FinanceDecision == nil) {
				return errors.New("partial bot record")
			}
			if e := verifyGameplayDecision(st); e != nil {
				return e
			}
			actor := m.Game.Board.Active
			if st.Command != nil {
				actor = st.Command.Actor
			} else if st.Operation != nil {
				actor = st.Operation.Actor
			}
			if actor < 1 || actor > len(r.Job.Policies) {
				return errors.New("invalid policy actor")
			}
			expectedPolicy := r.Job.Policies[actor-1].Policy + "@" + r.Job.Policies[actor-1].Version
			recordedPolicy := ""
			if st.Decision != nil {
				recordedPolicy = st.Decision.Policy
				if r.PolicyOverride != "" {
					expectedPolicy = r.PolicyOverride
				}
			} else {
				recordedPolicy = st.FinanceDecision.Policy
			}
			if recordedPolicy != expectedPolicy {
				return errors.New("recorded policy provenance mismatch")
			}
			rng := streams[fmt.Sprint("bot-", actor)]
			if rng == nil {
				return errors.New("invalid recorded bot actor")
			}
			before, _ := rng.Snapshot()
			if before != *st.BotBefore {
				return errors.New("bot before snapshot")
			}
			beforeCount, _ := strconv.ParseUint(st.BotBefore.Counter, 10, 64)
			afterCount, e := strconv.ParseUint(st.BotAfter.Counter, 10, 64)
			if e != nil || afterCount < beforeCount || afterCount-beforeCount > 1000000 {
				return errors.New("bot random consumption")
			}
			for j := beforeCount; j < afterCount; j++ {
				rng.Uint64()
			}
			after, _ := rng.Snapshot()
			if after != *st.BotAfter {
				return errors.New("bot after snapshot")
			}
		}
		n, events, e := applyGameplayStep(m, st)
		if e != nil {
			return fmt.Errorf("gameplay replay transition %d: %w", i, e)
		}
		if st.Post != hash(n) || st.Events != hash(events) {
			return fmt.Errorf("gameplay post digest %d", i)
		}
		m = n
		if st.Kind == "next-game" {
			streams, e = gameplayStreams(o, len(m.Instances)-1)
			if e != nil {
				return e
			}
		}
	}
	if hash(m) != hash(r.Match) || hash(snapshotStreams(streams)) != hash(r.Streams) {
		return errors.New("gameplay final checkpoint mismatch")
	}
	return nil
}

// ResumeGameplay retains the completed prefix and continues saved random streams.
func ResumeGameplay(ctx context.Context, prior GameplayRun, moreSteps, botBudget int, policy GameplayPolicy) GameplayRun {
	return RunGameplay(ctx, GameplayOptions{Job: prior.Job, Root: prior.Root, Pairing: prior.Pairing, MaxSteps: len(prior.Trace) + moreSteps, BotBudget: botBudget, MaxTranscriptBytes: gameplayTranscriptLimit(prior), Policy: policy, PolicyLabel: prior.PolicyOverride, Resume: &prior})
}
func verifyGameplayDecision(st GameplayStep) error {
	if st.FinanceDecision != nil {
		if st.Operation == nil {
			return errors.New("missing financial operation")
		}
		op := st.FinanceDecision.Operation
		op.ID = st.Operation.ID
		if hash(op) != hash(st.Operation) {
			return errors.New("financial decision detached")
		}
		return nil
	}
	d := st.Decision.Command
	if st.Command != nil {
		c := *st.Command
		if d.Actor != c.Actor || d.GameID != c.GameID {
			return errors.New("recorded decision actor/game mismatch")
		}
		d.ID = c.ID
		d.GameID = c.GameID
		d.Actor = c.Actor
		if d.Kind == "decision-closed" {
			d.Cards = nil
			d.RandomWords = slices.Clone(c.RandomWords)
		}
		if hash(d) != hash(c) {
			return errors.New("recorded decision detached from command")
		}
		return nil
	}
	if st.Operation == nil {
		return errors.New("decision has no action")
	}
	want := ""
	switch st.Operation.Kind {
	case "end-turn":
		want = "end-turn"
	case "declare-ordinary":
		want = "ordinary-victory"
	}
	if want == "" || d.Kind != want || d.Actor != st.Operation.Actor || d.GameID != st.Operation.GameID {
		return errors.New("recorded decision detached from operation")
	}
	return nil
}
