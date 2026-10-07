package sim

import (
	"context"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCommandSpecificFlagsRejected(t *testing.T) {
	a := runArgs(filepath.Join(t.TempDir(), "out"), fixture(t))
	a = append(a, "--manifest", "ignored")
	if c, _ := invoke(t, a...); c != 2 {
		t.Fatalf("irrelevant flag accepted: %d", c)
	}
}
func TestReplayRejectsRehashedOutcomeAndInitial(t *testing.T) {
	d := filepath.Join(t.TempDir(), "out")
	if c, m := invoke(t, runArgs(d, fixture(t))...); c != 0 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(d, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var out []Outcome
	if e = decodeRequired(f["outcomes.json"], &out); e != nil {
		t.Fatal(e)
	}
	out[0].Scores[0] = game.IntAmount(777)
	f["outcomes.json"] = mustCanonical(out)
	if e = verifyPrefix(m, f); e == nil {
		t.Fatal("forged outcome accepted despite valid trace")
	}
}
func TestReplayBindsScenarioInitial(t *testing.T) {
	d := filepath.Join(t.TempDir(), "out")
	if c, m := invoke(t, runArgs(d, fixture(t))...); c != 0 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(d, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var sc Scenario
	if e = decodeRequired(f["scenario.json"], &sc); e != nil {
		t.Fatal(e)
	}
	sc.Ledger.GameID = "forged-origin"
	f["scenario.json"] = mustCanonical(sc)
	if e = verifyPrefix(m, f); e == nil {
		t.Fatal("unbound scenario accepted")
	}
}
func TestOutputIOExit(t *testing.T) {
	p := filepath.Join(t.TempDir(), "file")
	if e := os.WriteFile(p, []byte("x"), 0600); e != nil {
		t.Fatal(e)
	}
	a := runArgs(filepath.Join(p, "out"), fixture(t))
	if c, _ := invoke(t, a...); c != 6 {
		t.Fatal(c)
	}
}

func TestCLIBotTranscriptAndReplay(t *testing.T) {
	s, _ := game.NewState(3, "bot-fixture")
	for _, x := range []struct {
		id   string
		seat int
	}{{"deck-1-clubs-08", 1}, {"deck-1-diamonds-02", 1}, {"deck-1-hearts-08", 2}} {
		s, _ = s.Move([]string{x.id}, x.seat, game.Series, true)
	}
	s.Players[1].History = []game.Suit{game.Hearts, game.Diamonds}
	sc := map[string]any{"schema_version": "cgms-scenario-v1", "id": "bot-slice", "kind": "transition-sequence", "purpose": "fair narrow menu", "players": 3, "rule_refs": []string{"N01"}, "state": s, "run_bots": true}
	p := filepath.Join(t.TempDir(), "bot.json")
	if e := os.WriteFile(p, mustCanonical(sc), 0600); e != nil {
		t.Fatal(e)
	}
	out := filepath.Join(t.TempDir(), "out")
	if c, m := invoke(t, runArgs(out, p)...); c != 3 {
		t.Fatalf("bot slice expected explicit unsupported stop: %d %s", c, m)
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	ts, e := decodeTrace(f["games/000000/trace.jsonl"])
	if e != nil || len(ts) < 2 {
		t.Fatal("missing bot transcript", e, len(ts))
	}
	if e = verifyPrefix(m, f); e != nil {
		t.Fatal(e)
	}
}

func TestBotRandomTranscriptTamperRejected(t *testing.T) {
	// Deterministic seed snapshots must be tied to the logical private stream.
	s, _ := game.NewState(3, "bot-rng")
	for _, x := range []struct {
		id   string
		seat int
	}{{"deck-1-clubs-08", 1}, {"deck-1-diamonds-02", 1}, {"deck-1-hearts-08", 2}} {
		s, _ = s.Move([]string{x.id}, x.seat, game.Series, true)
	}
	s.Players[1].History = []game.Suit{game.Hearts, game.Diamonds}
	sc := map[string]any{"schema_version": "cgms-scenario-v1", "id": "bot-rng", "kind": "transition-sequence", "purpose": "random provenance", "players": 3, "rule_refs": []string{"N01"}, "state": s, "run_bots": true}
	p := filepath.Join(t.TempDir(), "bot.json")
	os.WriteFile(p, mustCanonical(sc), 0600)
	out := filepath.Join(t.TempDir(), "out")
	if c, _ := invoke(t, runArgs(out, p)...); c != 3 {
		t.Fatal(c)
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	ts, e := decodeTrace(f["games/000000/trace.jsonl"])
	if e != nil {
		t.Fatal(e)
	}
	ts[0].BotRandomAfter.Counter = "999"
	f["games/000000/trace.jsonl"], _ = traceBytes(ts)
	if e = verifyPrefix(m, f); e == nil {
		t.Fatal("forged bot RNG accepted")
	}
}

func TestTournamentReportsUnstartedSlots(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	a := []string{"tournament", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--experiment", "../../../sims/experiments/baseline-3p.json", "--games", "3", "--max-steps", "2", "--out", out}
	if c, m := invoke(t, a...); c != 5 {
		t.Fatal(c, m)
	}
	b, e := os.ReadFile(filepath.Join(out, "report.md"))
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(b), "Planned game slots: 72") || !strings.Contains(string(b), "Later slots not started: 48") {
		t.Fatal("missing fixed match slot denominators", string(b))
	}
}

func TestExperimentSettlementBudgetApplied(t *testing.T) {
	b, e := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	if e != nil {
		t.Fatal(e)
	}
	x, e := LoadExperiment(b)
	if e != nil {
		t.Fatal(e)
	}
	x.Budgets.MaxSettlementSteps = 1
	p := filepath.Join(t.TempDir(), "experiment.json")
	os.WriteFile(p, mustCanonical(x), 0600)
	out := filepath.Join(t.TempDir(), "out")
	a := []string{"tournament", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--experiment", p, "--scenario", fixture(t), "--out", out}
	if c, _ := invoke(t, a...); c != 5 {
		t.Fatalf("settlement budget ignored: %d", c)
	}
}

func TestRuntimeProvenanceAndMetrics(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	if c, m := invoke(t, runArgs(out, fixture(t))...); c != 0 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	b := string(mustCanonical(m.Provenance))
	if !strings.Contains(b, "go-chacha8-v1") || !strings.Contains(b, "cgms-seed-v1") {
		t.Fatal("missing RNG compatibility")
	}
	if len(f["metrics.json"]) == 0 {
		t.Fatal("no measured runtime metrics")
	}
}

func TestOutputBudgetStopsBeforeMutation(t *testing.T) {
	sc, e := loadScenario(mustRead(t, fixture(t)), 3)
	if e != nil {
		t.Fatal(e)
	}
	j := Job{Index: 0, Population: 3, Games: 1}
	r := runScenarioLimited(context.Background(), j, &sc, "42", "standalone", Budgets{MaxTransitions: 100, MaxSettlementSteps: 100, MaxBotOperations: 1000, MaxOutputBytes: 1})
	if r.Outcome.ExitCode != 5 || len(r.Trace) != 0 {
		t.Fatal("output limit ignored", r.Outcome.ExitCode, len(r.Trace))
	}
}
func mustRead(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

func TestReplayRejectsMalformedSchedule(t *testing.T) {
	d := filepath.Join(t.TempDir(), "out")
	if c, _ := invoke(t, runArgs(d, fixture(t))...); c != 0 {
		t.Fatal(c)
	}
	m, f, e := readManifest(filepath.Join(d, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var jobs []Job
	if e = decodeRequired(f["schedule.json"], &jobs); e != nil {
		t.Fatal(e)
	}
	jobs[0].Policies = nil
	f["schedule.json"] = mustCanonical(jobs)
	if e = verifyPrefix(m, f); e == nil {
		t.Fatal("malformed schedule accepted")
	}
}

func TestSingleDashAndMixedDuplicateRejected(t *testing.T) {
	for _, flags := range [][]string{{"-games", "3"}, {"-seed", "43"}} {
		a := append(runArgs(filepath.Join(t.TempDir(), "out"), fixture(t)), flags...)
		if c, _ := invoke(t, a...); c != 2 {
			t.Fatal("single dash bypass", c)
		}
	}
}
func TestCheckpointRejectsOversizedExpectation(t *testing.T) {
	l := game.NewFinancialLedger("x", 3)
	c, e := game.BeginSettlement(l, nil)
	if e != nil {
		t.Fatal(e)
	}
	sc := Scenario{Schema: "cgms-scenario-v1", ID: "x", Purpose: "x", Kind: "settlement-checkpoint", Players: 3, RuleRefs: []string{"Q10"}, Checkpoint: &c, Parent: "parent", ExpectedScores: make([]game.Amount, 4)}
	if _, e = loadScenario(mustCanonical(sc), 3); e == nil {
		t.Fatal("oversized expected scores")
	}
}

func TestExperimentMatchIndexChangesEnvironmentStream(t *testing.T) {
	b := mustRead(t, "../../../sims/experiments/baseline-3p.json")
	x, e := LoadExperiment(b)
	if e != nil {
		t.Fatal(e)
	}
	jobs, e := x.Schedule()
	if e != nil {
		t.Fatal(e)
	}
	a := runScenario(context.Background(), jobs[0], nil, "42", x.PairingID, 2)
	x.MatchIndex = 9
	jobs, e = x.Schedule()
	if e != nil {
		t.Fatal(e)
	}
	z := runScenario(context.Background(), jobs[0], nil, "42", x.PairingID, 2)
	if a.Gameplay == nil || z.Gameplay == nil || hash(a.Gameplay.Trace[0].Shuffle) == hash(z.Gameplay.Trace[0].Shuffle) {
		t.Fatal("match index omitted from stream identity")
	}
}

func TestReplayBindsEffectiveExperimentSchedule(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	args := []string{"tournament", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--experiment", "../../../sims/experiments/baseline-3p.json", "--max-steps", "2", "--out", out}
	if c, m := invoke(t, args...); c != 5 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	x, e := LoadExperiment(f["effective-experiment.json"])
	if e != nil {
		t.Fatal(e)
	}
	x.MatchIndex = 99
	f["effective-experiment.json"] = mustCanonical(x)
	if e = verifyPrefix(m, f); e == nil {
		t.Fatal("unbound effective experiment")
	}
}
func TestReplayBindsFixedScenarioCommands(t *testing.T) {
	sc, e := loadScenario(mustRead(t, "../../../sims/scenarios/narrow-combat-bots.json"), 3)
	if e != nil {
		t.Fatal(e)
	}
	j := Job{Population: 3, Games: 1, Block: "b", Rotation: Rotation{ID: "r0", Seats: []string{"p0", "p1", "p2"}}, Policies: []Policy{{ParticipantID: "p0", Policy: "legal-random", Version: "v1", Information: "seat-projection"}, {ParticipantID: "p1", Policy: "legal-random", Version: "v1", Information: "seat-projection"}, {ParticipantID: "p2", Policy: "legal-random", Version: "v1", Information: "seat-projection"}}}
	r := runScenario(context.Background(), j, &sc, "42", "standalone", 100)
	sc.RunBots = false
	sc.Commands = []game.Command{*r.Trace[0].Command}
	sc.ExpectedDigests = []string{r.Trace[0].Post}
	p := filepath.Join(t.TempDir(), "fixed.json")
	os.WriteFile(p, mustCanonical(sc), 0600)
	out := filepath.Join(t.TempDir(), "out")
	if c, m := invoke(t, runArgs(out, p)...); c != 0 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	sc.ExpectedDigests[0] = "forged"
	f["scenario.json"] = mustCanonical(sc)
	if e = verifyPrefix(m, f); e == nil {
		t.Fatal("unbound expected transition")
	}
}

func TestReplayRejectsForgedDenominators(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	if c, m := invoke(t, runArgs(out, fixture(t))...); c != 0 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	m.Finalized = 1
	if e = verifyPrefix(m, f); e == nil {
		t.Fatal("fabricated game denominator accepted")
	}
}

func TestPublicationFailureManifest(t *testing.T) {
	d := t.TempDir()
	if e := os.WriteFile(filepath.Join(d, "blocked"), nil, 0600); e != nil {
		t.Fatal(e)
	}
	m := manifestBase(nil, "run", "", "", 1, 1, nil)
	if e := publish(d, m, map[string][]byte{"a.txt": []byte("kept"), "blocked/fail.txt": []byte("x")}); e == nil {
		t.Fatal("expected I/O failure")
	}
	b, e := os.ReadFile(filepath.Join(d, "manifest.json"))
	if e != nil {
		t.Fatal("missing publishable failure manifest", e)
	}
	if e = decodeRequired(b, &m); e != nil || m.ExitCode != 6 || m.Status != "io-failure" || len(m.Artifacts) != 1 {
		t.Fatal("lost completed artifact inventory", m, e)
	}
}

func TestCompareRejectsDifferentScenario(t *testing.T) {
	root := t.TempDir()
	sc := fixture(t)
	a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
	if c, m := invoke(t, runArgs(a, sc)...); c != 0 {
		t.Fatal(c, m)
	}
	var s Scenario
	if e := decodeRequired(mustRead(t, sc), &s); e != nil {
		t.Fatal(e)
	}
	s.Purpose = "different scenario lineage"
	os.WriteFile(sc, mustCanonical(s), 0600)
	if c, m := invoke(t, runArgs(b, sc)...); c != 0 {
		t.Fatal(c, m)
	}
	if c, _ := invoke(t, "compare", "--baseline", filepath.Join(a, "manifest.json"), "--candidate", filepath.Join(b, "manifest.json"), "--out", filepath.Join(root, "compare")); c != 2 {
		t.Fatal("unmatched scenario labeled paired", c)
	}
}

func TestPairingIncludesMatchAndBlockIdentity(t *testing.T) {
	b := mustRead(t, "../../../sims/experiments/baseline-3p.json")
	x, e := LoadExperiment(b)
	if e != nil {
		t.Fatal(e)
	}
	jobs, e := x.Schedule()
	if e != nil {
		t.Fatal(e)
	}
	a := map[string][]byte{"effective-experiment.json": mustCanonical(x), "schedule.json": mustCanonical(jobs)}
	z := map[string][]byte{"effective-experiment.json": mustCanonical(x), "schedule.json": mustCanonical(jobs)}
	if e = compatiblePairing(a, z); e != nil {
		t.Fatal(e)
	}
	jobs[0].Match++
	z["schedule.json"] = mustCanonical(jobs)
	if e = compatiblePairing(a, z); e == nil {
		t.Fatal("match identity ignored")
	}
	jobs[0].Match--
	z["schedule.json"] = mustCanonical(jobs)
	x.PairingID = "different-pair"
	z["effective-experiment.json"] = mustCanonical(x)
	if e = compatiblePairing(a, z); e == nil {
		t.Fatal("pairing id ignored")
	}
}

func TestContinuationRequiresVerifiedParent(t *testing.T) {
	l := game.NewFinancialLedger("x", 3)
	c, e := game.BeginSettlement(l, nil)
	if e != nil {
		t.Fatal(e)
	}
	sc := Scenario{Schema: "cgms-scenario-v1", ID: "x", Kind: "settlement-checkpoint", Purpose: "resume", Players: 3, RuleRefs: []string{"Q10"}, Checkpoint: &c, Parent: filepath.Join(t.TempDir(), "missing.json")}
	p := filepath.Join(t.TempDir(), "sc.json")
	os.WriteFile(p, mustCanonical(sc), 0600)
	if code, _ := invoke(t, runArgs(filepath.Join(t.TempDir(), "out"), p)...); code != 6 {
		t.Fatal("unverified parent accepted", code)
	}
}
func TestReviewHoldRecordedAtPublication(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out")
	a := append(runArgs(out, fixture(t)), "--review-hold")
	if c, m := invoke(t, a...); c != 0 {
		t.Fatal(c, m)
	}
	m, _, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil || !m.ReviewHold {
		t.Fatal("hold not recorded", e)
	}
}

func TestVerifiedSettlementContinuationAndReplay(t *testing.T) {
	base := filepath.Join(t.TempDir(), "partial")
	args := append(runArgs(base, fixture(t)), "--max-steps", "1")
	if c, m := invoke(t, args...); c != 5 {
		t.Fatal(c, m)
	}
	var cursor game.SettlementCursor
	if e := decodeRequired(mustRead(t, filepath.Join(base, "games/000000/checkpoint.json")), &cursor); e != nil {
		t.Fatal(e)
	}
	sc := Scenario{Schema: "cgms-scenario-v1", ID: "continued", Kind: "settlement-checkpoint", Purpose: "continue original exact FIFO cursor", Players: 3, RuleRefs: []string{"Q10"}, Checkpoint: &cursor, Parent: filepath.Join(base, "manifest.json")}
	p := filepath.Join(t.TempDir(), "resume.json")
	os.WriteFile(p, mustCanonical(sc), 0600)
	out := filepath.Join(t.TempDir(), "continued")
	if c, m := invoke(t, runArgs(out, p)...); c != 0 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	if m.Parent != sc.Parent {
		t.Fatal("parent metadata missing")
	}
	if e = verifyPrefix(m, f); e != nil {
		t.Fatal(e)
	}
}
