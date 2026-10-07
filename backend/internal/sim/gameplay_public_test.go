package sim

import (
	"context"
	"encoding/json"
	"github.com/metaphy6/cgms/backend/internal/game"
	"strings"
	"testing"
)

func TestGameplayPublicReasonNeverLeaksPrivateDiagnostics(t *testing.T) {
	r := GameplayRun{ExitCode: 3, Reason: "private-seed-canary hidden-deck-1-hearts-12"}
	got := publicGameplayStop(r)
	if strings.Contains(got, "canary") || strings.Contains(got, "hearts") || got != "unsupported-or-waiting" {
		t.Fatal("public reason leaked raw diagnostic", got)
	}
}

func TestGameplayPublicReportStripsPrivateInputs(t *testing.T) {
	j := Job{Population: 3, Block: "private-block-canary", Rotation: Rotation{ID: "private-rotation-canary"}, Games: 1}
	for i := 0; i < 3; i++ {
		j.Policies = append(j.Policies, Policy{Policy: "economic", Version: "v1"})
	}
	r := RunGameplay(context.Background(), GameplayOptions{Job: j, Root: "91827364509182736450", Pairing: "private-pairing-canary", MaxSteps: 5, BotBudget: 1000, Policy: restrictedEndTurn, PolicyLabel: "scripted-end-turn-test-only"})
	r.Reason = "private-diagnostic-canary"
	report, e := GameplayPublicReport(context.Background(), r)
	if e != nil {
		t.Fatal(e)
	}
	for _, forbidden := range []string{"canary", "91827364509182736450", "deck-1", "deck-2", "random_words", "private-pairing"} {
		if strings.Contains(report, forbidden) {
			t.Fatalf("report leaked %s", forbidden)
		}
	}
	stats, e := GameplayStats(context.Background(), r)
	if e != nil {
		t.Fatal(e)
	}
	if stats.StartedGames != 1 || stats.FinanciallyFinalizedGames != 0 || stats.FinalizedMatches != 0 || stats.IncompleteReason != "resource-budget-exhausted" {
		t.Fatal("partial denominators mislabeled", stats.StartedGames, stats.FinanciallyFinalizedGames)
	}
}

func TestGameplayPublicFinancialObservationalEquivalence(t *testing.T) {
	// Two otherwise identical public views differ only in a private charge between
	// other seats. Neither JSON nor Markdown may disclose its exact total.
	a := GameplayPublicStats{Schema: "cgms-gameplay-public-v2", Population: 3, Economy: []GameplayPublicEconomy{{Slot: 1, Round: 2, RecordedTotal: game.IntAmount(0), AvailableTotal: game.IntAmount(0), UnpaidTotal: game.IntAmount(0)}}}
	b := a
	b.Economy = []GameplayPublicEconomy{{Slot: 1, Round: 2, RecordedTotal: game.IntAmount(-13), AvailableTotal: game.IntAmount(0), UnpaidTotal: game.IntAmount(13)}}
	aj, err := json.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	bj, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	if string(aj) != string(bj) {
		t.Fatal("private financial state changed public JSON")
	}
	if RenderGameplayPublicReport(a) != RenderGameplayPublicReport(b) {
		t.Fatal("private financial state changed public Markdown")
	}
	if strings.Contains(string(aj), "unpaid_total") || strings.Contains(string(aj), "available_total") {
		t.Fatal("private financial fields remain public")
	}
}

func TestGameplayOutcomePrivateChargeCannotChangePartialPublicScores(t *testing.T) {
	s, _ := game.NewState(3, "private-outcome")
	m, _ := game.NewMatchLifecycle(s, 1)
	run := GameplayRun{Job: Job{Population: 3, Games: 1}, Match: m, ExitCode: 5}
	before := gameplayOutcome(run)
	ledger, err := m.Game.Ledger.Charge("private-party-charge", 0, 1, game.IntAmount(13))
	if err != nil {
		t.Fatal(err)
	}
	run.Match.Game.Ledger = ledger
	after := gameplayOutcome(run)
	a, _ := json.Marshal(before)
	b, _ := json.Marshal(after)
	if string(a) != string(b) {
		t.Fatal("unfinalized private charge changed public outcome")
	}
	if len(after.Scores) != 0 || len(after.MatchRanks) != 0 {
		t.Fatal("partial current private scores or ranks exposed")
	}
}
