package sim

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPrivateFinanceExactSnapshotDetached(t *testing.T) {
	l := game.NewFinancialLedger("g", 3)
	amount, _ := game.IntAmount(13).Quo(game.IntAmount(7))
	var err error
	l, err = l.Charge("private", 0, 1, amount)
	if err != nil {
		t.Fatal(err)
	}
	d := newGameplayPrivateFinance()
	d.capture("round", 1, 2, l)
	if d.Privacy != "restricted" || d.Observer != "omniscient" || d.Snapshots[0].UnpaidTotal.Cmp(amount) != 0 {
		t.Fatal("private exact labeled snapshot")
	}
	l.Debts[0].ID = "mutated"
	if d.Snapshots[0].Ledger.Debts[0].ID == "mutated" {
		t.Fatal("snapshot aliases ledger")
	}
	b, _ := json.Marshal(d)
	var restored GameplayPrivateFinance
	if err := json.Unmarshal(b, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Snapshots[0].UnpaidTotal.String() != "13/7" {
		t.Fatal("exact fraction missing")
	}
}
func TestPrivateFinanceCollectedAlongsidePublicProjection(t *testing.T) {
	j := Job{Population: 3, Games: 1}
	for i := 0; i < 3; i++ {
		j.Policies = append(j.Policies, Policy{Policy: "economic", Version: "v1"})
	}
	r := RunGameplay(context.Background(), GameplayOptions{Job: j, Root: "19", Pairing: "dev-private-economy", MaxSteps: 5, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"})
	publicOnly, err := GameplayStats(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	a := fullGameplayResult(context.Background(), r)
	b := fullGameplayResult(context.Background(), r)
	if a.PrivateFinance == nil || len(a.PrivateFinance.Snapshots) == 0 || hash(a.PrivateFinance) != hash(b.PrivateFinance) {
		t.Fatal("deterministic private snapshot missing")
	}
	if hash(publicOnly) != hash(*a.PublicStats) {
		t.Fatal("enabling omniscient collector changed public projection")
	}
	raw, _ := json.Marshal(a.PublicStats)
	if strings.Contains(string(raw), "unpaid_total") || strings.Contains(RenderGameplayPublicReport(*a.PublicStats), "13/7") {
		t.Fatal("private diagnostics escaped")
	}
}

func TestPrivateFinanceArtifactPublishedAndReplayVerified(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "run")
	code, msg := invoke(t, "run", "--config", "../../../config/rules/game-rules.json", "--seed", "42", "--players", "3", "--bots", "economic@v1,pressure@v1,economic@v1", "--max-steps", "2", "--out", dir)
	if code != 5 {
		t.Fatalf("expected bounded retained run: %d %s", code, msg)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "games/000000/finance-private.json"))
	if err != nil {
		t.Fatal(err)
	}
	var private GameplayPrivateFinance
	if err = json.Unmarshal(raw, &private); err != nil {
		t.Fatal(err)
	}
	if private.Observer != "omniscient" || private.Privacy != "restricted" || len(private.Snapshots) == 0 {
		t.Fatal("missing private artifact labels or snapshots")
	}
	public, err := os.ReadFile(filepath.Join(dir, "games/000000/gameplay-public.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(public), "unpaid_total") || strings.Contains(string(public), "ledger") {
		t.Fatal("private artifact merged into public")
	}
	code, msg = invoke(t, "replay", "--manifest", filepath.Join(dir, "manifest.json"), "--out", filepath.Join(t.TempDir(), "replay"))
	if code != 5 {
		t.Fatalf("partial replay lost status: %d %s", code, msg)
	}
}

func TestPrivateFinanceOversizeHistoryRetainsBoundedPrefix(t *testing.T) {
	ledger := game.NewFinancialLedger("large-history", 3)
	for i := 0; i < 600; i++ {
		ledger.Operations = append(ledger.Operations, fmt.Sprintf("operation-%04d-%s", i, strings.Repeat("x", 1000)))
	}
	before := game.Digest(ledger)
	diagnostic := newGameplayPrivateFinance()
	for round := 1; round <= 13; round++ {
		diagnostic.capture("begin-round", 1, round, ledger)
	}
	raw, err := canonical.Marshal(diagnostic)
	if err != nil {
		t.Fatal("diagnostic accumulation exceeds canonical artifact bound", err)
	}
	var shape map[string]any
	if err = json.Unmarshal(raw, &shape); err != nil {
		t.Fatal(err)
	}
	if shape["status"] != "incomplete" || shape["reason"] != "diagnostic-byte-budget" {
		t.Fatal("missing explicit partial diagnostic status")
	}
	if len(diagnostic.Snapshots) == 0 || len(diagnostic.Snapshots) >= 13 {
		t.Fatal("bounded prefix not retained")
	}
	if shape["omitted_snapshots"] != float64(13-len(diagnostic.Snapshots)) {
		t.Fatal("omitted diagnostic denominator")
	}
	for i, snapshot := range diagnostic.Snapshots {
		if snapshot.Round != i+1 {
			t.Fatal("diagnostic dropped an interior boundary")
		}
	}
	if game.Digest(ledger) != before {
		t.Fatal("diagnostic cap changed authoritative obligations/checkpoint")
	}
	again := newGameplayPrivateFinance()
	for round := 1; round <= 13; round++ {
		again.capture("begin-round", 1, round, ledger)
	}
	if hash(again) != hash(diagnostic) {
		t.Fatal("diagnostic cap not replay deterministic")
	}
}

func TestPrivateFinanceCapIncludesTerminalMetadata(t *testing.T) {
	ledger := game.NewFinancialLedger("exact-boundary", 3)
	ledger.Operations = []string{"x"}
	sizing := newGameplayPrivateFinance()
	sizing.capture("round", 1, 1, ledger)
	raw, err := canonical.Marshal(sizing)
	if err != nil {
		t.Fatal(err)
	}
	ledger.Operations[0] = strings.Repeat("x", 1+privateFinanceMaxBytes-len(raw))
	diagnostic := newGameplayPrivateFinance()
	diagnostic.capture("round", 1, 1, ledger)
	diagnostic.capture("round", 1, 2, ledger)
	raw, err = canonical.Marshal(diagnostic)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) > privateFinanceMaxBytes {
		t.Fatal("terminal metadata exceeded declared diagnostic byte limit")
	}
	if diagnostic.Status != "incomplete" || diagnostic.OmittedSnapshots < 1 {
		t.Fatal("missing bounded diagnostic completion status")
	}
}
