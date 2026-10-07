package sim

import (
	"context"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
)

// committedGameplayStarts counts a physical instance only after its first
// committed deal (redeals do not add starts). Voided replacements share the
// configured slot until a financially finalized game advances it.
func committedGameplayStarts(g GameplayRun) (games, instances, voided int) {
	dealt := map[string]bool{}
	for _, step := range g.Trace {
		if step.Kind == "deal" {
			dealt[step.GameID] = true
		}
	}
	finalized := map[string]bool{}
	for _, result := range g.Match.Game.Ledger.Completed {
		finalized[result.GameID] = true
	}
	slots := map[int]bool{}
	slot := 1
	for _, id := range g.Match.Instances {
		if dealt[id] {
			instances++
			slots[slot] = true
			if !finalized[id] && (id != g.Match.Game.Board.GameID || g.Match.Game.Ending == "void") {
				voided++
			}
		}
		if finalized[id] {
			slot++
		}
	}
	return len(slots), instances, voided
}

func gameplayOutcome(g GameplayRun) Outcome {
	j := g.Job
	completed := len(g.Match.Game.Ledger.Completed)
	started, instances, voided := committedGameplayStarts(g)
	boardEnded := completed
	if g.Match.Game.Board.Phase == "settlement" && g.Match.Game.Ledger.Phase != "finalized" && g.Match.Game.Ending != "void" {
		boardEnded++
	}
	// Reasons here are a public taxonomy. Detailed engine failures remain private.
	reason := map[int]string{0: "match financially finalized", 2: "invalid gameplay input", 3: "unsupported or waiting gameplay boundary", 4: "gameplay invariant failure", 5: "gameplay resource budget", 130: "canceled at committed boundary"}[g.ExitCode]
	remaining := j.Games - started
	if remaining < 0 {
		remaining = 0
	}
	var scores []game.Amount
	var ranks []int
	if g.Match.Game.Ledger.Phase == "finalized" {
		scores = g.Match.Game.Ledger.MatchScores()
		ranks = g.Match.Ranks()
	}
	return Outcome{StartedInstances: instances, VoidedGames: voided, Schema: "cgms-outcome-v1", Index: j.Index, Population: j.Population, Block: j.Block, Rotation: j.Rotation.ID, Seats: j.Rotation.Seats, Treatment: j.Treatment, GameID: g.Match.Game.Board.GameID, PlannedGameSlots: j.Games, NotStartedGameSlots: remaining, StartedGames: started, BoardEndedGames: boardEnded, FinalizedGames: completed, GameComplete: completed > 0, MatchComplete: completed == j.Games, Status: status(g.ExitCode), ExitCode: g.ExitCode, Reason: reason, Scores: scores, Steps: len(g.Trace), Redeals: g.Redeals, Observer: "public", Privacy: "public", MatchRanks: ranks, GameResults: g.Match.Game.Ledger.Completed}
}
func gameplayLegacyTrace(g GameplayRun) []Trace {
	ts := make([]Trace, len(g.Trace))
	for i, t := range g.Trace {
		ts[i] = Trace{Schema: "cgms-trace-v1", Index: i, GameID: t.GameID, Kind: "gameplay-v1", Pre: t.Pre, Post: t.Post, Events: t.Events}
	}
	return ts
}
func runFullGameplay(ctx context.Context, j Job, root, pairing string, limits Budgets) runResult {
	g := RunGameplay(ctx, GameplayOptions{Job: j, Root: root, Pairing: pairing, MaxSteps: limits.MaxTransitions, BotBudget: limits.MaxBotOperations, MenuBudget: limits.MenuOperations(), MaxOutputBytes: limits.MaxOutputBytes / 2, MaxTranscriptBytes: limits.MaxGameplayTranscriptBytes})
	if g.ExitCode == 130 && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		g.ExitCode = 5
	}
	return fullGameplayResult(ctx, g)
}
func verifyGameplayUnit(ctx context.Context, m Manifest, j Job, o Outcome, files map[string][]byte, prefix, pairing string) error {
	var g GameplayRun
	if err := decodeRequired(files[prefix+"gameplay.json"], &g); err != nil {
		return err
	}
	var limits Budgets
	if err := decodeRequired(files["execution-limits.json"], &limits); err != nil {
		return err
	}
	expected := limits.MaxGameplayTranscriptBytes
	if expected == 0 {
		expected = canonical.MaxBytes
	}
	if gameplayTranscriptLimit(g) != expected {
		return fmt.Errorf("gameplay transcript budget provenance")
	}
	if hash(j) != hash(g.Job) || g.Root != m.RootSeed || g.Pairing != pairing {
		return fmt.Errorf("gameplay provenance mismatch")
	}
	if err := ReplayGameplayContext(ctx, g); err != nil {
		return err
	}
	if raw := files[prefix+"gameplay-public.json"]; raw != nil {
		var recorded GameplayPublicStats
		if err := decodeRequired(raw, &recorded); err != nil {
			return err
		}
		expected, err := GameplayStats(ctx, g)
		if err != nil {
			return err
		}
		if hash(recorded) != hash(expected) {
			return fmt.Errorf("public gameplay projection divergence")
		}
	}
	if raw := files[prefix+"finance-private.json"]; raw != nil {
		var recorded GameplayPrivateFinance
		if err := decodeRequired(raw, &recorded); err != nil {
			return err
		}
		expected := newGameplayPrivateFinance()
		if _, err := gameplayStatsWithFinance(ctx, g, expected); err != nil {
			return err
		}
		if hash(recorded) != hash(expected) {
			return fmt.Errorf("private financial projection divergence")
		}
	}
	if hash(o) != hash(gameplayOutcome(g)) {
		return fmt.Errorf("gameplay outcome divergence")
	}
	var initial, checkpoint game.MatchLifecycle
	if err := decodeRequired(files[prefix+"initial.json"], &initial); err != nil {
		return err
	}
	if err := decodeRequired(files[prefix+"checkpoint.json"], &checkpoint); err != nil {
		return err
	}
	if hash(initial) != hash(g.Initial) || hash(checkpoint) != hash(g.Match) {
		return fmt.Errorf("gameplay checkpoint divergence")
	}
	ts, err := decodeTrace(files[prefix+"trace.jsonl"])
	if err != nil {
		return err
	}
	expectedTrace := gameplayLegacyTrace(g)
	if len(ts) != len(expectedTrace) {
		return fmt.Errorf("gameplay trace length divergence")
	}
	for i := range ts {
		if hash(ts[i]) != hash(expectedTrace[i]) {
			return fmt.Errorf("gameplay trace divergence")
		}
	}
	return nil
}

func continueFullGameplay(ctx context.Context, j Job, root, pairing string, limits Budgets, prior GameplayRun) runResult {
	if hash(j) != hash(prior.Job) || root != prior.Root || pairing != prior.Pairing {
		return runResult{Outcome: Outcome{Index: j.Index, ExitCode: 2, Status: "invalid-input", Reason: "continuation inputs differ"}}
	}
	g := RunGameplay(ctx, GameplayOptions{Job: j, Root: root, Pairing: pairing, MaxSteps: len(prior.Trace) + limits.MaxTransitions, BotBudget: limits.MaxBotOperations, MenuBudget: limits.MenuOperations(), MaxOutputBytes: limits.MaxOutputBytes / 2, MaxTranscriptBytes: limits.MaxGameplayTranscriptBytes, Resume: &prior})
	if g.ExitCode == 130 && errors.Is(ctx.Err(), context.DeadlineExceeded) {
		g.ExitCode = 5
	}
	return fullGameplayResult(ctx, g)
}

func fullGameplayResult(ctx context.Context, g GameplayRun) runResult {
	r := runResult{Gameplay: &g, Outcome: gameplayOutcome(g), Initial: g.Initial, Checkpoint: g.Match, Trace: gameplayLegacyTrace(g)}
	private := newGameplayPrivateFinance()
	if stats, err := gameplayStatsWithFinance(ctx, g, private); err == nil {
		r.PublicStats = &stats
		r.PrivateFinance = private
	}
	return r
}

func completedComparison(a, b []Outcome) string {
	text := ""
	for _, pop := range []int{3, 4} {
		planned, complete, changed := 0, 0, 0
		blocks := map[string]bool{}
		for i, x := range a {
			if x.Population != pop {
				continue
			}
			planned++
			if i >= len(b) || !x.MatchComplete || !b[i].MatchComplete {
				continue
			}
			complete++
			blocks[x.Block] = true
			if hash(x.Scores) != hash(b[i].Scores) {
				changed++
			}
		}
		text += fmt.Sprintf("%d-player planned paired matches: %d; completed paired matches: %d; represented seed blocks: %d; changed exact score vectors: %d.\n\n", pop, planned, complete, len(blocks), changed)
	}
	return text + "These paired diagnostics are descriptive. Seed-block estimates and clustered intervals require the frozen named experiment; repeated seats are not independent samples. Completion-only results can be selection-biased.\n\n"
}

// A named artifact profile changes the experimental storage budget, never rules.
const GameplayLargeArtifactProfile = "cgms-gameplay-artifact-32mib-v1"

func gameplayTranscriptLimit(g GameplayRun) int {
	switch g.ArtifactProfile {
	case "":
		return canonical.MaxBytes
	case GameplayLargeArtifactProfile:
		return 32 << 20
	default:
		return 0
	}
}
func validateExtendedArtifact(raw []byte, v any) error {
	switch x := v.(type) {
	case *GameplayRun:
		limit := gameplayTranscriptLimit(*x)
		if limit == 0 || len(raw) > limit {
			return fmt.Errorf("gameplay artifact profile/byte limit")
		}
	case *Scenario:
		if x.Gameplay == nil {
			if len(raw) > canonical.MaxBytes {
				return fmt.Errorf("scenario byte limit")
			}
			return nil
		}
		if x.Kind != "gameplay-checkpoint" {
			return fmt.Errorf("unexpected embedded gameplay")
		}
		limit := gameplayTranscriptLimit(*x.Gameplay)
		if limit == 0 || len(raw) > gameplayScenarioLimit(*x.Gameplay) {
			return fmt.Errorf("gameplay scenario byte limit")
		}
		if _, err := canonical.MarshalLimit(x.Gameplay, limit); err != nil {
			return err
		}
	}
	return nil
}

func gameplayScenarioLimit(g GameplayRun) int {
	limit := gameplayTranscriptLimit(g)
	if g.ArtifactProfile == GameplayLargeArtifactProfile {
		limit += 65536
	}
	return limit
}
