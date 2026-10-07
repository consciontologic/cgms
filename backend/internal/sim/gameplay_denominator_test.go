package sim

import (
	"context"
	"testing"
)

func TestGameplayStartedDenominatorsRequireCommittedDeal(t *testing.T) {
	j := Job{Population: 3, Games: 2, Policies: []Policy{{Policy: "heuristic", Version: "v1"}, {Policy: "heuristic", Version: "v1"}, {Policy: "heuristic", Version: "v1"}}}
	opts := GameplayOptions{Job: j, Root: "1", Pairing: "fixture", MaxSteps: 2000, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"}
	check := func(t *testing.T, r GameplayRun, want int) {
		t.Helper()
		if e := ReplayGameplay(r); e != nil {
			t.Fatal(e)
		}
		stats, e := GameplayStats(context.Background(), r)
		if e != nil {
			t.Fatal(e)
		}
		out := gameplayOutcome(r)
		if out.StartedGames != want || out.StartedInstances != want || out.NotStartedGameSlots != 2-want {
			t.Fatalf("outcome counts allocated but undealt instance: started=%d instances=%d never=%d want started=%d", out.StartedGames, out.StartedInstances, out.NotStartedGameSlots, want)
		}
		if stats.StartedGames != out.StartedGames || stats.StartedInstances != out.StartedInstances || stats.NeverStartedGames != out.NotStartedGameSlots {
			t.Fatal("public/outcome denominators disagree")
		}
	}
	t.Run("canceled-before-first-deal", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		r := RunGameplay(ctx, opts)
		if r.ExitCode != 130 || len(r.Trace) != 0 {
			t.Fatal("unexpected cancellation boundary")
		}
		check(t, r, 0)
	})
	t.Run("canceled-after-next-game-before-deal", func(t *testing.T) {
		whole := RunGameplay(context.Background(), opts)
		if whole.ExitCode != 0 {
			t.Fatal(whole.Reason)
		}
		boundary := 0
		for i, step := range whole.Trace {
			if step.Kind == "next-game" {
				boundary = i + 1
				break
			}
		}
		if boundary == 0 {
			t.Fatal("missing next game")
		}
		opts.MaxSteps = boundary
		prefix := RunGameplay(context.Background(), opts)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		r := ResumeGameplay(ctx, prefix, 10, 1000, restrictedEndTurn)
		if r.ExitCode != 130 || len(r.Match.Game.Ledger.Completed) != 1 || len(r.Trace) != boundary {
			t.Fatal("completed game lost at cancellation")
		}
		check(t, r, 1)
	})
}

func TestGameplayCommittedStartsRedealsAndUndealtVoidReplacement(t *testing.T) {
	r := GameplayRun{}
	r.Match.Instances = []string{"voided", "replacement"}
	r.Match.Game.Board.GameID = "replacement"
	r.Trace = []GameplayStep{{Kind: "deal", GameID: "voided"}, {Kind: "deal", GameID: "voided"}}
	games, instances, voided := committedGameplayStarts(r)
	if games != 1 || instances != 1 || voided != 1 {
		t.Fatalf("redeal or undealt replacement inflated counts: %d %d %d", games, instances, voided)
	}
	r.Trace = append(r.Trace, GameplayStep{Kind: "deal", GameID: "replacement"})
	games, instances, voided = committedGameplayStarts(r)
	if games != 1 || instances != 2 || voided != 1 {
		t.Fatalf("replacement changed configured slot: %d %d %d", games, instances, voided)
	}
}
