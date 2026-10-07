package sim

import (
	"bytes"
	"context"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	l := game.NewFinancialLedger("accounting-g1", 3)
	var e error
	l, e = l.Charge("a", 0, 1, game.IntAmount(1))
	if e != nil {
		t.Fatal(e)
	}
	l, e = l.Charge("b", 1, 0, game.IntAmount(1))
	if e != nil {
		t.Fatal(e)
	}
	third, _ := game.NewAmount("1", "3")
	v := map[string]any{"schema_version": "cgms-scenario-v1", "id": "q10-thirds", "kind": "accounting-only", "purpose": "Q10 exact reciprocal thirds; no game completion", "players": 3, "rule_refs": []string{"Q10", "GAME_RULES §4.2"}, "ledger": l, "receipts": []game.FinancialReceipt{{Seat: 0, Amount: third}}, "expected_scores": []game.Amount{third, game.IntAmount(0), game.IntAmount(0)}}
	b, e := canonical.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(t.TempDir(), "scenario.json")
	if e = os.WriteFile(p, b, 0600); e != nil {
		t.Fatal(e)
	}
	return p
}
func invoke(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var out, err bytes.Buffer
	code := Execute(context.Background(), args, &out, &err)
	return code, out.String() + err.String()
}
func runArgs(out, scenario string) []string {
	return []string{"run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "legal-random@v1,legal-random@v1,heuristic@v1", "--scenario", scenario, "--out", out}
}
func TestCLIRunReplayAndNoOverwrite(t *testing.T) {
	sc := fixture(t)
	dir := filepath.Join(t.TempDir(), "run")
	code, msg := invoke(t, runArgs(dir, sc)...)
	if code != 0 {
		t.Fatalf("run exit=%d %s", code, msg)
	}
	if !strings.Contains(msg, "manifest.json") {
		t.Fatal("manifest path missing")
	}
	b, e := os.ReadFile(filepath.Join(dir, "report.md"))
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(string(b), "scenario-complete") || !strings.Contains(string(b), "Finalized games: 0 / 1") {
		t.Fatal("dishonest report", string(b))
	}
	if strings.Contains(string(b), "--seed") {
		t.Fatal("private argv leaked")
	}
	code, _ = invoke(t, runArgs(dir, sc)...)
	if code != 2 {
		t.Fatal("overwrite accepted")
	}
	replay := filepath.Join(t.TempDir(), "replay")
	code, msg = invoke(t, "replay", "--manifest", filepath.Join(dir, "manifest.json"), "--out", replay)
	if code != 0 {
		t.Fatal(code, msg)
	}
}
func TestCLIInvalidAndCancellation(t *testing.T) {
	sc := fixture(t)
	args := runArgs(filepath.Join(t.TempDir(), "run"), sc)
	args = append(args, "--unknown", "1")
	if c, _ := invoke(t, args...); c != 2 {
		t.Fatal(c)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	c := Execute(ctx, runArgs(filepath.Join(t.TempDir(), "cancel"), sc), &out, &out)
	if c != 130 {
		t.Fatalf("cancel=%d", c)
	}
}
func TestCLIWorkerDeterminismPartialReplayAndCompare(t *testing.T) {
	sc := fixture(t)
	root := t.TempDir()
	dirs := []string{filepath.Join(root, "one"), filepath.Join(root, "four")}
	for i, d := range dirs {
		a := runArgs(d, sc)
		a[0] = "batch"
		a = append(a, "--games", "5", "--workers", []string{"1", "4"}[i])
		c, m := invoke(t, a...)
		if c != 3 {
			t.Fatal(c, m)
		}
	}
	a, e := os.ReadFile(filepath.Join(dirs[0], "outcomes.json"))
	if e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(dirs[1], "outcomes.json"))
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("worker count changed outcomes")
	}
	c, m := invoke(t, "compare", "--baseline", filepath.Join(dirs[0], "manifest.json"), "--candidate", filepath.Join(dirs[1], "manifest.json"), "--out", filepath.Join(root, "compare"))
	if c != 0 {
		t.Fatal(c, m)
	}
	c, m = invoke(t, "replay", "--manifest", filepath.Join(dirs[0], "manifest.json"), "--out", filepath.Join(root, "replay"))
	if c != 3 {
		t.Fatal("partial status not preserved", c, m)
	}
}
func TestCLIReplayTamper(t *testing.T) {
	d := filepath.Join(t.TempDir(), "run")
	if c, m := invoke(t, runArgs(d, fixture(t))...); c != 0 {
		t.Fatal(c, m)
	}
	p := filepath.Join(d, "games/000000/trace.jsonl")
	if e := os.WriteFile(p, []byte("{}\n"), 0600); e != nil {
		t.Fatal(e)
	}
	c, _ := invoke(t, "replay", "--manifest", filepath.Join(d, "manifest.json"), "--out", filepath.Join(t.TempDir(), "replay"))
	if c != 4 {
		t.Fatal("tamper accepted", c)
	}
}
func TestCLIBudgetAndUnsupported(t *testing.T) {
	sc := fixture(t)
	a := runArgs(filepath.Join(t.TempDir(), "budget"), sc)
	a = append(a, "--max-steps", "1")
	if c, m := invoke(t, a...); c != 5 {
		t.Fatal(c, m)
	}
	a = runArgs(filepath.Join(t.TempDir(), "unsupported"), sc)
	a = append(a[:len(a)-4], a[len(a)-2:]...)
	a = append(a, "--max-steps", "2")
	if c, m := invoke(t, a...); c != 5 {
		t.Fatal(c, m)
	}
}
func TestArtifactTraversalSymlinkAndPermissions(t *testing.T) {
	d := t.TempDir()
	if _, e := safePath(d, "../escape"); e == nil {
		t.Fatal("traversal")
	}
	if e := os.Symlink(t.TempDir(), filepath.Join(d, "link")); e != nil {
		t.Fatal(e)
	}
	if _, e := safePath(d, "link/output"); e == nil {
		t.Fatal("symlink")
	}
	out := filepath.Join(d, "run")
	if c, m := invoke(t, runArgs(out, fixture(t))...); c != 0 {
		t.Fatal(c, m)
	}
	info, e := os.Stat(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatal("private manifest permissions", info.Mode())
	}
}
