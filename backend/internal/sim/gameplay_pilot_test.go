package sim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGameplayDevelopmentPilot(t *testing.T) {
	for _, pop := range []int{3, 4} {
		j := Job{Index: 0, Population: pop, Block: "development-fixture", Rotation: Rotation{ID: "r0"}, Games: 1}
		for i := 0; i < pop; i++ {
			policy := "economic"
			if i%2 == 1 {
				policy = "pressure"
			}
			j.Policies = append(j.Policies, Policy{Policy: policy, Version: "v1"})
		}
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		r := RunGameplay(ctx, GameplayOptions{Job: j, Root: "14", Pairing: "development-fixture", MaxSteps: 10000, BotBudget: 100000})
		cancel()
		runElapsed := time.Since(start)
		marshalStart := time.Now()
		b, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		marshalElapsed := time.Since(marshalStart)
		if dir := os.Getenv("CGMS_PILOT_CAPTURE"); dir != "" {
			writeStart := time.Now()
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("p%d-gameplay.json", pop)), b, 0600); err != nil {
				t.Fatal(err)
			}
			t.Logf("population=%d marshal=%s write=%s", pop, marshalElapsed, time.Since(writeStart))
		}
		t.Logf("population=%d finalized=%d steps=%d code=%d reason=%s elapsed=%s bytes=%d", pop, len(r.Match.Game.Ledger.Completed), len(r.Trace), r.ExitCode, r.Reason, runElapsed, len(b))
		t.Logf("development transcript digest=%s", hash(r))
		replayStart := time.Now()
		if e := ReplayGameplay(r); e != nil {
			t.Fatal(e)
		}
		t.Logf("population=%d replay=%s", pop, time.Since(replayStart))
		if r.ExitCode != 0 {
			t.Fatalf("autonomous development pilot incomplete: %s", r.Reason)
		}
	}
}

// BenchmarkGameplayDevelopmentCampaign is an explicit opt-in development pilot:
// go test -run '^$' -bench BenchmarkGameplayDevelopmentCampaign -benchtime=1x.
func BenchmarkGameplayDevelopmentCampaign(b *testing.B) {
	dir := os.Getenv("CGMS_DEV3_OUT")
	if dir == "" {
		b.Fatal("CGMS_DEV3_OUT required for immutable private artifacts")
	}
	if e := os.Mkdir(dir, 0700); e != nil {
		b.Fatal(e)
	}
	jobs := []Job{}
	for _, pop := range []int{3, 4} {
		j := Job{Index: pop - 3, Population: pop, Block: "development-fixture", Rotation: Rotation{ID: "r0"}, Games: 1}
		if pop == 3 {
			j.Games = 2
		}
		for i := 0; i < pop; i++ {
			policy := "economic"
			if i%2 == 1 {
				policy = "pressure"
			}
			j.Policies = append(j.Policies, Policy{Policy: policy, Version: "v1"})
		}
		jobs = append(jobs, j)
	}
	prior := []GameplayRun{}
	failed := false
	for _, workers := range []int{1, 4} {
		results := make([]GameplayRun, len(jobs))
		elapsed := make([]time.Duration, len(jobs))
		queue := make(chan int)
		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := range queue {
					start := time.Now()
					ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
					results[i] = RunGameplay(ctx, GameplayOptions{Job: jobs[i], Root: "14", Pairing: "development-fixture", MaxSteps: 10000, BotBudget: 100000})
					cancel()
					elapsed[i] = time.Since(start)
				}
			}()
		}
		for i := range jobs {
			queue <- i
		}
		close(queue)
		wg.Wait()
		for i, r := range results {
			raw, e := json.Marshal(r)
			if e != nil {
				b.Fatal(e)
			}
			path := filepath.Join(dir, fmt.Sprintf("p%d-workers%d-gameplay.json", r.Job.Population, workers))
			if e = os.WriteFile(path, raw, 0600); e != nil {
				b.Fatal(e)
			}
			replayStart := time.Now()
			e = ReplayGameplay(r)
			b.Logf("population=%d workers=%d configured=%d finalized=%d transitions=%d exit=%d run=%s replay=%s bytes=%d", r.Job.Population, workers, r.Job.Games, len(r.Match.Game.Ledger.Completed), len(r.Trace), r.ExitCode, elapsed[i], time.Since(replayStart), len(raw))
			if e != nil {
				b.Errorf("replay failed: %v", e)
			}
			if r.ExitCode != 0 {
				failed = true
			}
			if workers == 4 && hash(r) != hash(prior[i]) {
				b.Error("worker count changed complete transcript")
			}
		}
		if workers == 1 {
			prior = results
		}
	}
	if failed {
		b.Fatal("development campaign incomplete; private outcomes retained")
	}
}

func TestGameplayExtendedTranscriptRoundTripAndResume(t *testing.T) {
	j := Job{Population: 3, Games: 1, Policies: []Policy{{Policy: "heuristic", Version: "v1"}, {Policy: "heuristic", Version: "v1"}, {Policy: "heuristic", Version: "v1"}}}
	padded := func(p string, o game.Observation, r *randomstream.Stream, b int) (bots.Decision, error) {
		d, e := restrictedEndTurn(p, o, r, b)
		d.Reasons = []string{strings.Repeat("synthetic inspection rationale; ", 40000)}
		return d, e
	}
	options := GameplayOptions{Job: j, Root: "3", Pairing: "fixture", MaxSteps: 24, BotBudget: 100000, Policy: padded, PolicyLabel: "scripted-end-turn-test-only", MaxOutputBytes: 32 << 20}
	old := RunGameplay(context.Background(), options)
	if old.ExitCode != 5 || !strings.Contains(old.Reason, "output") {
		t.Fatal("fixture does not reproduce original 4MiB cap", old.Reason)
	}
	options.MaxTranscriptBytes = 32 << 20
	larger := RunGameplay(context.Background(), options)
	if len(larger.Trace) <= len(old.Trace) || !equalGameplayPrefix(larger.Trace, old.Trace) {
		t.Fatal("explicit storage budget changed committed prefix")
	}
	raw := mustCanonical(&larger)
	if len(raw) <= canonical.MaxBytes || len(raw) > 32<<20 {
		t.Fatal("fixture must cross old byte boundary", len(raw))
	}
	if _, err := canonical.Marshal(larger); err == nil {
		t.Fatal("generic canonical limit silently widened")
	}
	var loaded GameplayRun
	if err := decodeRequired(raw, &loaded); err != nil {
		t.Fatal(err)
	}
	if err := ReplayGameplay(loaded); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, mustCanonical(&loaded)) {
		t.Fatal("extended artifact bytes changed")
	}
	loaded.ArtifactProfile = ""
	if err := decodeRequired(mustCanonical(&larger), &loaded); err != nil {
		t.Fatal(err)
	}
	tampered := bytes.Replace(raw, []byte(GameplayLargeArtifactProfile), []byte("unsupported-profile"), 1)
	if decodeRequired(tampered, &loaded) == nil {
		t.Fatal("unsupported profile accepted")
	}
	var small GameplayRun
	if canonical.Decode(raw, &small) == nil {
		t.Fatal("default decoder accepted extended artifact")
	}
	resumed := ResumeGameplay(context.Background(), larger, 2, 100000, padded)
	if resumed.ExitCode != 5 || len(resumed.Trace) != len(larger.Trace)+2 || !equalGameplayPrefix(resumed.Trace, larger.Trace) {
		t.Fatal("extended resume lost prefix/budget", resumed.Reason)
	}
	if err := ReplayGameplay(resumed); err != nil {
		t.Fatal(err)
	}
	sc := Scenario{Schema: "cgms-scenario-v1", ID: "large-checkpoint", Kind: "gameplay-checkpoint", Purpose: "synthetic boundary regression", Players: 3, RuleRefs: []string{"T03"}, Parent: "synthetic-parent", Gameplay: &larger}
	var decoded Scenario
	if err := decodeRequired(mustCanonical(sc), &decoded); err != nil {
		t.Fatal("extended checkpoint envelope", err)
	}
}
