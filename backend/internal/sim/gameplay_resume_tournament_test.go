package sim

import (
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"os"
	"path/filepath"
	"testing"
)

func TestCLIGameplayResumeTournamentParent(t *testing.T) {
	root := t.TempDir()
	initial := filepath.Join(root, "initial")
	code, msg := invoke(t, "tournament", "--config", "../../../config/rules/game-rules.json", "--experiment", "../../../sims/experiments/gameplay-dev-3p-v3.json", "--seed", "42", "--games", "2", "--max-steps", "2", "--out", initial)
	if code != 5 {
		t.Fatalf("initial %d %s", code, msg)
	}
	b, e := os.ReadFile(filepath.Join(initial, "games/000001/gameplay.json"))
	if e != nil {
		t.Fatal(e)
	}
	var g GameplayRun
	if e = canonical.Decode(b, &g); e != nil {
		t.Fatal(e)
	}
	sc := map[string]any{"schema_version": "cgms-scenario-v1", "id": "resume-tournament", "kind": "gameplay-checkpoint", "purpose": "retain frozen tournament job", "players": 3, "rule_refs": []string{"N05"}, "parent": filepath.Join(initial, "manifest.json"), "gameplay": g}
	path := filepath.Join(root, "resume.json")
	if e = os.WriteFile(path, mustCanonical(sc), 0600); e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(root, "continued")
	code, msg = invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--scenario", path, "--max-steps", "2", "--out", out)
	if code != 5 {
		t.Fatalf("resume %d %s", code, msg)
	}
	lb, le := os.ReadFile(filepath.Join(out, "execution-limits.json"))
	if le != nil {
		t.Fatal(le)
	}
	var limits Budgets
	if le = canonical.Decode(lb, &limits); le != nil {
		t.Fatal(le)
	}
	if limits.MaxBotOperations != 100000 {
		t.Fatalf("policy budget changed: %d", limits.MaxBotOperations)
	}
	b, e = os.ReadFile(filepath.Join(out, "games/000001/gameplay.json"))
	if e != nil {
		t.Fatal(e)
	}
	var next GameplayRun
	if e = canonical.Decode(b, &next); e != nil {
		t.Fatal(e)
	}
	if hash(next.Job) != hash(g.Job) || next.Root != g.Root || next.Pairing != g.Pairing || len(next.Trace) != 4 {
		t.Fatal("frozen identity or prefix changed")
	}
	code, msg = invoke(t, "replay", "--manifest", filepath.Join(out, "manifest.json"), "--out", filepath.Join(root, "replay"))
	if code != 5 {
		t.Fatalf("replay %d %s", code, msg)
	}
	code, _ = invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--scenario", path, "--seed", "43", "--max-steps", "2", "--out", filepath.Join(root, "conflict"))
	if code != 2 {
		t.Fatalf("override conflict %d", code)
	}
}

func TestCLIGameplayExtendedBudgetResumeTournamentParent(t *testing.T) {
	root := t.TempDir()
	raw, err := os.ReadFile("../../../sims/experiments/gameplay-dev-3p-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	exp, err := LoadExperiment(raw)
	if err != nil {
		t.Fatal(err)
	}
	exp.Budgets.MaxMenuOperations = 1000000
	exp.Budgets.MaxGameplayTranscriptBytes = 32 << 20
	experimentPath := filepath.Join(root, "expanded.json")
	if err = os.WriteFile(experimentPath, mustCanonical(exp), 0600); err != nil {
		t.Fatal(err)
	}
	initial := filepath.Join(root, "initial")
	code, msg := invoke(t, "tournament", "--config", "../../../config/rules/game-rules.json", "--experiment", experimentPath, "--seed", "42", "--games", "2", "--max-steps", "2", "--out", initial)
	if code != 5 {
		t.Fatalf("initial %d %s", code, msg)
	}
	b, e := os.ReadFile(filepath.Join(initial, "games/000001/gameplay.json"))
	if e != nil {
		t.Fatal(e)
	}
	var g GameplayRun
	if e = canonical.Decode(b, &g); e != nil {
		t.Fatal(e)
	}
	sc := map[string]any{"schema_version": "cgms-scenario-v1", "id": "resume-tournament", "kind": "gameplay-checkpoint", "purpose": "retain frozen tournament job", "players": 3, "rule_refs": []string{"N05"}, "parent": filepath.Join(initial, "manifest.json"), "gameplay": g}
	path := filepath.Join(root, "resume.json")
	if e = os.WriteFile(path, mustCanonical(sc), 0600); e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(root, "continued")
	code, msg = invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--scenario", path, "--max-steps", "2", "--out", out)
	if code != 5 {
		t.Fatalf("resume %d %s", code, msg)
	}
	lb, le := os.ReadFile(filepath.Join(out, "execution-limits.json"))
	if le != nil {
		t.Fatal(le)
	}
	var limits Budgets
	if le = canonical.Decode(lb, &limits); le != nil {
		t.Fatal(le)
	}
	if limits.MaxGameplayTranscriptBytes != 32<<20 || limits.MaxMenuOperations != 1000000 || limits.MaxBotOperations != 100000 {
		t.Fatalf("policy budget changed: %d", limits.MaxBotOperations)
	}
	b, e = os.ReadFile(filepath.Join(out, "games/000001/gameplay.json"))
	if e != nil {
		t.Fatal(e)
	}
	var next GameplayRun
	if e = canonical.Decode(b, &next); e != nil {
		t.Fatal(e)
	}
	if hash(next.Job) != hash(g.Job) || next.Root != g.Root || next.Pairing != g.Pairing || len(next.Trace) != 4 {
		t.Fatal("frozen identity or prefix changed")
	}
	code, msg = invoke(t, "replay", "--manifest", filepath.Join(out, "manifest.json"), "--out", filepath.Join(root, "replay"))
	if code != 5 {
		t.Fatalf("replay %d %s", code, msg)
	}
	code, _ = invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--scenario", path, "--seed", "43", "--max-steps", "2", "--out", filepath.Join(root, "conflict"))
	if code != 2 {
		t.Fatalf("override conflict %d", code)
	}
}
