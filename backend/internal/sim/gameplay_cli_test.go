package sim

import (
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIGameplayCheckpointReplay(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "gameplay")
	code, msg := invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "legal-random@v1,legal-random@v1,heuristic@v1", "--max-steps", "2", "--out", dir)
	if code != 5 {
		t.Fatalf("budget must retain gameplay checkpoint: %d %s", code, msg)
	}
	if _, err := os.Stat(filepath.Join(dir, "games/000000/gameplay.json")); err != nil {
		t.Fatal(err)
	}
	replayed := filepath.Join(t.TempDir(), "replay")
	code, msg = invoke(t, "replay", "--manifest", filepath.Join(dir, "manifest.json"), "--out", replayed)
	if code != 5 {
		t.Fatalf("replay %d %s", code, msg)
	}
	b, err := os.ReadFile(filepath.Join(replayed, "replay-diagnostic.txt"))
	if err != nil || !strings.Contains(string(b), "recorded decisions and random outcomes verified") {
		t.Fatalf("verification failed: %s %v", b, err)
	}
}

func TestGameplayPolicyExperimentAdmission(t *testing.T) {
	for _, name := range []string{"economic", "pressure", "gameplay-random"} {
		p := Policy{ParticipantID: "p0", Policy: name, Version: "v1", Information: "seat-projection"}
		if err := p.Validate(); err != nil {
			t.Fatal(name, err)
		}
	}
}

func TestGameplayReportCountsEveryFinalizedGame(t *testing.T) {
	o := Outcome{Index: 0, Population: 3, PlannedGameSlots: 2, StartedGames: 2, BoardEndedGames: 2, FinalizedGames: 2, GameComplete: true, MatchComplete: true, Status: "complete", Observer: "public", Privacy: "public"}
	s := report("tournament", []Outcome{o}, nil)
	if !strings.Contains(s, "Finalized games: 2 / 2") || !strings.Contains(s, "finalized matches: 1") {
		t.Fatal(s)
	}
}

func TestGameplayAnalysisUsesFocalFinalMatchScore(t *testing.T) {
	raw, err := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	if err != nil {
		t.Fatal(err)
	}
	x, err := LoadExperiment(raw)
	if err != nil {
		t.Fatal(err)
	}
	x.Blocks = x.Blocks[:1]
	jobs, err := x.Schedule()
	if err != nil {
		t.Fatal(err)
	}
	outcomes := make([]Outcome, len(jobs))
	for i, j := range jobs {
		o := Outcome{Status: "complete", MatchComplete: true, Scores: []game.Amount{game.IntAmount(0), game.IntAmount(0), game.IntAmount(0)}}
		for seat, id := range j.Rotation.Seats {
			if id == j.FocalParticipantID && j.Treatment == "candidate" {
				o.Scores[seat] = game.IntAmount(7)
			}
		}
		outcomes[i] = o
	}
	files, _, err := experimentReports(x, jobs, outcomes)
	if err != nil {
		t.Fatal(err)
	}
	var s Summary
	for k, b := range files {
		if strings.HasPrefix(k, "analysis/") {
			if err := canonical.Decode(b, &s); err != nil {
				t.Fatal(err)
			}
		}
	}
	if s.Mean == nil || s.Mean.Cmp(game.IntAmount(7)) != 0 || s.CompleteBlocks != 1 {
		t.Fatalf("wrong focal paired effect: %+v", s)
	}
}

func TestCLIGameplayResumeParent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "initial")
	args := []string{"run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "legal-random@v1,legal-random@v1,heuristic@v1", "--max-steps", "2", "--out", dir}
	if code, _ := invoke(t, args...); code != 5 {
		t.Fatal(code)
	}
	b, err := os.ReadFile(filepath.Join(dir, "games/000000/gameplay.json"))
	if err != nil {
		t.Fatal(err)
	}
	var g any
	if err := canonical.Decode(b, &g); err != nil {
		t.Fatal(err)
	}
	sc := map[string]any{"schema_version": "cgms-scenario-v1", "id": "resume-gameplay", "kind": "gameplay-checkpoint", "purpose": "resume same committed game", "players": 3, "rule_refs": []string{"N05"}, "parent": filepath.Join(dir, "manifest.json"), "gameplay": g}
	path := filepath.Join(root, "resume.json")
	if err := os.WriteFile(path, mustCanonical(sc), 0600); err != nil {
		t.Fatal(err)
	}
	args[len(args)-1] = filepath.Join(root, "continued")
	args = append(args, "--scenario", path)
	if code, msg := invoke(t, args...); code != 5 {
		t.Fatalf("resume %d %s", code, msg)
	}
	b, err = os.ReadFile(filepath.Join(root, "continued/games/000000/gameplay.json"))
	if err != nil {
		t.Fatal(err)
	}
	var continued GameplayRun
	if err := canonical.Decode(b, &continued); err != nil {
		t.Fatal(err)
	}
	if len(continued.Trace) != 4 {
		t.Fatalf("prefix not continued: %d", len(continued.Trace))
	}
}

func TestGameplayDevelopmentProfiles(t *testing.T) {
	for _, pop := range []string{"3", "4"} {
		b, e := os.ReadFile("../../../sims/experiments/gameplay-dev-" + pop + "p-v3.json")
		if e != nil {
			t.Fatal(e)
		}
		x, e := LoadExperiment(b)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = x.Schedule(); e != nil {
			t.Fatal(e)
		}
	}
}

func TestGameplayPublicOutcomeHidesPrivateFinancialTerms(t *testing.T) {
	s, _ := game.NewState(3, "g")
	m, _ := game.NewMatchLifecycle(s, 1)
	l, e := m.Game.Ledger.Charge("private", 0, 1, game.IntAmount(13))
	if e != nil {
		t.Fatal(e)
	}
	m.Game.Ledger = l
	o := gameplayOutcome(GameplayRun{Job: Job{Population: 3, Games: 1}, Match: m, ExitCode: 5})
	if len(o.Debts) != 0 || len(o.Cash) != 0 {
		t.Fatal("private financial terms exposed in public outcome")
	}
}

func TestGameplayMetricsCountsGamesNotMatches(t *testing.T) {
	end := startMetrics()
	m := end([]runResult{{Outcome: Outcome{GameComplete: true, FinalizedGames: 3}}})
	if m.FinalizedGames != 3 {
		t.Fatal(m.FinalizedGames)
	}
}

func TestCompletedComparisonUsesObservedPairs(t *testing.T) {
	a := []Outcome{{Population: 3, Block: "b", MatchComplete: true, Scores: []game.Amount{game.IntAmount(4), game.IntAmount(0), game.IntAmount(1)}}}
	b := []Outcome{{Population: 3, Block: "b", MatchComplete: true, Scores: []game.Amount{game.IntAmount(6), game.IntAmount(0), game.IntAmount(1)}}}
	s := completedComparison(a, b)
	if !strings.Contains(s, "completed paired matches: 1") || !strings.Contains(s, "changed exact score vectors: 1") {
		t.Fatal(s)
	}
}

func TestOneSeedBlockHasNoUncertaintyInterval(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	x, _ := LoadExperiment(b)
	x.Blocks = []string{"one"}
	rows := []PairOutcome{}
	for _, r := range x.Rotations {
		rows = append(rows, PairOutcome{"one", r.ID, "baseline", "finalized", game.IntAmount(0)}, PairOutcome{"one", r.ID, "candidate", "finalized", game.IntAmount(7)})
	}
	s, e := Analyze(x, rows)
	if e != nil {
		t.Fatal(e)
	}
	if s.Mean == nil || s.Lower != nil || s.Upper != nil {
		t.Fatal("one independent block cannot estimate sampling uncertainty")
	}
}

func TestGameplayRandomV2Admission(t *testing.T) {
	p := Policy{ParticipantID: "p0", Policy: "gameplay-random", Version: "v2", Information: "seat-projection"}
	if e := p.Validate(); e != nil {
		t.Fatal(e)
	}
	p.Policy = "economic"
	if e := p.Validate(); e != nil {
		t.Fatal("implemented deterministic v2 rejected", e)
	}
	p.Version = "v3"
	if p.Validate() == nil {
		t.Fatal("unimplemented economic version accepted")
	}
	out := filepath.Join(t.TempDir(), "out")
	code, msg := invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "gameplay-random@v2,economic@v1,pressure@v1", "--max-steps", "2", "--out", out)
	if code != 5 {
		t.Fatalf("v2 input rejected or budget not preserved: %d %s", code, msg)
	}
}

func TestStandaloneGameplayBudgetMatchesFullMenu(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	code, msg := invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "economic@v1,pressure@v1,economic@v1", "--max-steps", "2", "--out", out)
	if code != 5 {
		t.Fatalf("%d %s", code, msg)
	}
	raw, e := os.ReadFile(filepath.Join(out, "execution-limits.json"))
	if e != nil {
		t.Fatal(e)
	}
	var b Budgets
	if e = canonical.Decode(raw, &b); e != nil {
		t.Fatal(e)
	}
	if b.MaxBotOperations != 100000 {
		t.Fatalf("standalone full menu receives narrow budget: %d", b.MaxBotOperations)
	}
}

func TestGameplayArtifactSchemasTravelWithRun(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	code, msg := invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "economic@v1,pressure@v1,economic@v1", "--max-steps", "2", "--out", out)
	if code != 5 {
		t.Fatalf("%d %s", code, msg)
	}
	for _, name := range []string{"gameplay", "gameplay-public", "finance-private"} {
		b, e := os.ReadFile(filepath.Join(out, "schemas", name+".schema.json"))
		if e != nil {
			t.Error("missing archived output contract", name, e)
			continue
		}
		if name == "gameplay-public" && strings.Contains(string(b), `"economy"`) {
			t.Fatal("public schema includes private trajectory")
		}
	}
}

func TestOpportunityPoliciesRegistered(t *testing.T) {
	for _, name := range []string{"opportunity", "opportunity-no-retain", "opportunity-no-finance", "bargaining"} {
		p := Policy{ParticipantID: "candidate", Policy: name, Version: "v1", Information: "seat-projection"}
		if e := p.Validate(); e != nil {
			t.Fatal(name, e)
		}
	}
	out := filepath.Join(t.TempDir(), "out")
	code, msg := invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "opportunity@v1,opportunity-no-retain@v1,bargaining@v1", "--max-steps", "2", "--out", out)
	if code != 5 {
		t.Fatalf("candidate registration/budget status %d %s", code, msg)
	}
}

func TestDeterministicV2PoliciesRegistered(t *testing.T) {
	for _, name := range []string{"economic", "pressure", "opportunity", "opportunity-no-retain", "opportunity-no-finance", "bargaining"} {
		p := Policy{ParticipantID: "candidate", Policy: name, Version: "v2", Information: "seat-projection"}
		if e := p.Validate(); e != nil {
			t.Fatal(name, e)
		}
		p.Version = "v3"
		if p.Validate() == nil {
			t.Fatal("unsupported version accepted")
		}
	}
	p := Policy{ParticipantID: "candidate", Policy: "heuristic", Version: "v2", Information: "seat-projection"}
	if p.Validate() == nil {
		t.Fatal("unversioned policy alias accepted")
	}
	out := filepath.Join(t.TempDir(), "out")
	code, msg := invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "economic@v2,pressure@v2,bargaining@v2", "--max-steps", "10", "--out", out)
	if code != 5 {
		t.Fatalf("v2 registration/decision label failed %d %s", code, msg)
	}
}
