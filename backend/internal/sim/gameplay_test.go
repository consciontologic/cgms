package sim

import (
	"context"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"testing"
)

func restrictedEndTurn(policy string, o game.Observation, r *randomstream.Stream, budget int) (bots.Decision, error) {
	kind := "end-turn"
	for _, c := range o.Legal {
		if c.Kind == "decline-out-of-turn" {
			kind = c.Kind
		}
	}
	return bots.Decision{Command: game.Command{GameID: o.GameID, Actor: o.Seat, Kind: kind}, Policy: "scripted-end-turn-test-only", Reasons: []string{"restricted integration fixture: explicitly choose end turn"}}, nil
}
func TestGameplayNormalDealCompleteAndReplay(t *testing.T) {
	for _, pop := range []int{3, 4} {
		j := Job{Index: 0, Population: pop, Block: "fixture", Rotation: Rotation{ID: "r0"}, Games: 2}
		for i := 0; i < pop; i++ {
			j.Policies = append(j.Policies, Policy{Policy: "heuristic", Version: "v1"})
		}
		r := RunGameplay(context.Background(), GameplayOptions{Job: j, Root: "1", Pairing: "fixture", MaxSteps: 2000, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"})
		if r.ExitCode != 0 || len(r.Match.Game.Ledger.Completed) != 2 {
			t.Fatalf("pop%d complete %d reason %s", pop, len(r.Match.Game.Ledger.Completed), r.Reason)
		}
		if e := ReplayGameplay(r); e != nil {
			t.Fatal(e)
		}
	}
}

func TestGameplayCancellationBudgetAndTamper(t *testing.T) {
	j := Job{Index: 1, Population: 3, Block: "fixture", Rotation: Rotation{ID: "r0"}, Games: 1}
	for i := 0; i < 3; i++ {
		j.Policies = append(j.Policies, Policy{Policy: "heuristic", Version: "v1"})
	}
	o := GameplayOptions{Job: j, Root: "2", Pairing: "fixture", MaxSteps: 20, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"}
	r := RunGameplay(context.Background(), o)
	if r.ExitCode != 5 || len(r.Match.Game.Ledger.Completed) != 0 {
		t.Fatal("budget fabricated ending", r.ExitCode)
	}
	if e := ReplayGameplay(r); e != nil {
		t.Fatal(e)
	}
	r.Trace[0].Shuffle[0], r.Trace[0].Shuffle[1] = r.Trace[0].Shuffle[1], r.Trace[0].Shuffle[0]
	if ReplayGameplay(r) == nil {
		t.Fatal("detached permutation accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r = RunGameplay(ctx, o)
	if r.ExitCode != 130 || len(r.Trace) != 0 {
		t.Fatal("cancellation mutation")
	}
	if e := ReplayGameplay(r); e != nil {
		t.Fatal(e)
	}
}

func TestGameplayResumeRetainsPrefix(t *testing.T) {
	j := Job{Index: 0, Population: 3, Block: "resume", Rotation: Rotation{ID: "r0"}, Games: 1}
	for i := 0; i < 3; i++ {
		j.Policies = append(j.Policies, Policy{Policy: "heuristic", Version: "v1"})
	}
	o := GameplayOptions{Job: j, Root: "3", Pairing: "fixture", MaxSteps: 6, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"}
	prefix := RunGameplay(context.Background(), o)
	resumed := ResumeGameplay(context.Background(), prefix, 1000, 1000, restrictedEndTurn)
	if resumed.ExitCode != 0 {
		t.Fatal(resumed.Reason)
	}
	o.MaxSteps = 1006
	whole := RunGameplay(context.Background(), o)
	if hash(resumed.Match) != hash(whole.Match) || hash(resumed.Trace) != hash(whole.Trace) {
		t.Fatal("resume changed outcome or transcript")
	}
	if e := ReplayGameplay(resumed); e != nil {
		t.Fatal(e)
	}
}

func TestGameplayWorkerCountPrefixIndependent(t *testing.T) {
	opts := []GameplayOptions{}
	for idx := 0; idx < 2; idx++ {
		j := Job{Index: idx, Population: 3, Block: "workers", Rotation: Rotation{ID: "r0"}, Games: 1}
		for i := 0; i < 3; i++ {
			j.Policies = append(j.Policies, Policy{Policy: "heuristic", Version: "v1"})
		}
		opts = append(opts, GameplayOptions{Job: j, Root: "8", Pairing: "fixture", MaxSteps: 4, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"})
	}
	serial := []GameplayRun{RunGameplay(context.Background(), opts[0]), RunGameplay(context.Background(), opts[1])}
	ch := make(chan GameplayRun, 2)
	for _, o := range opts {
		go func(o GameplayOptions) { ch <- RunGameplay(context.Background(), o) }(o)
	}
	for range opts {
		r := <-ch
		if hash(r) != hash(serial[r.Job.Index]) {
			t.Fatal("worker schedule changed semantic run")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if ReplayGameplayContext(ctx, serial[0]) == nil {
		t.Fatal("replay ignored cancellation")
	}
}

func TestGameplayPairedJobsShareEnvironmentStreams(t *testing.T) {
	a := GameplayOptions{Job: Job{Index: 1, Match: 7, Population: 3, Block: "paired", Rotation: Rotation{ID: "r0"}}, Root: "91", Pairing: "paired"}
	b := a
	b.Job.Index = 2
	b.Job.Treatment = "candidate"
	x, e := gameplayStreams(a, 0)
	if e != nil {
		t.Fatal(e)
	}
	y, e := gameplayStreams(b, 0)
	if e != nil {
		t.Fatal(e)
	}
	if hash(snapshotStreams(x)) != hash(snapshotStreams(y)) {
		t.Fatal("paired jobs use different streams solely because treatment output index differs")
	}
}

func TestGameplayOutputBudgetRetainsReplayablePrefix(t *testing.T) {
	j := Job{Population: 3, Block: "bytes", Rotation: Rotation{ID: "r0"}, Games: 1}
	for i := 0; i < 3; i++ {
		j.Policies = append(j.Policies, Policy{Policy: "economic", Version: "v1"})
	}
	r := RunGameplay(context.Background(), GameplayOptions{Job: j, Root: "2", Pairing: "fixture", MaxSteps: 1000, BotBudget: 1000, MaxOutputBytes: 100000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"})
	if r.ExitCode != 5 || len(r.Match.Game.Ledger.Completed) != 0 {
		t.Fatal("byte exhaustion fabricated completion")
	}
	if e := ReplayGameplay(r); e != nil {
		t.Fatal(e)
	}
}

func TestGameplayReplayRequiresDecisionAndPolicyProvenance(t *testing.T) {
	j := Job{Population: 3, Block: "decision-provenance", Rotation: Rotation{ID: "r0"}, Games: 1}
	for i := 0; i < 3; i++ {
		j.Policies = append(j.Policies, Policy{Policy: "economic", Version: "v1"})
	}
	r := RunGameplay(context.Background(), GameplayOptions{Job: j, Root: "2", Pairing: "fixture", MaxSteps: 20, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"})
	found := false
	for i, s := range r.Trace {
		if s.Decision == nil || s.BotBefore == nil {
			continue
		}
		found = true
		saved := r.Trace[i]
		r.Trace[i].Decision = nil
		r.Trace[i].BotBefore = nil
		r.Trace[i].BotAfter = nil
		if ReplayGameplay(r) == nil {
			t.Fatal("missing decision accepted")
		}
		r.Trace[i] = saved
		d := *saved.Decision
		d.Policy = "forged@v1"
		r.Trace[i].Decision = &d
		if ReplayGameplay(r) == nil {
			t.Fatal("forged policy provenance accepted")
		}
		r.Trace[i] = saved
		break
	}
	if !found {
		t.Fatal("fixture did not exercise decision")
	}
	r.ExitCode = 0
	if ReplayGameplay(r) == nil {
		t.Fatal("partial checkpoint labeled complete")
	}
}
