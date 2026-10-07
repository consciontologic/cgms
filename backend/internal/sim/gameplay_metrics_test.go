package sim

import (
	"encoding/json"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"strings"
	"testing"
)

func TestMetricsCountGameplayDecisionsWithoutPrivatePayload(t *testing.T) {
	finish := startMetrics()
	r := runResult{Gameplay: &GameplayRun{Root: "private-seed-canary", Trace: []GameplayStep{{Decision: &bots.Decision{Operations: 7, Reasons: []string{"private-card-canary"}}}, {FinanceDecision: &bots.GameplayFinanceDecision{Operations: 5, Reasons: []string{"private-finance-canary"}}}, {}}}, Trace: make([]Trace, 3)}
	m := finish([]runResult{r})
	if m.DecisionCount != 2 {
		t.Fatalf("gameplay decisions falsely counted %d", m.DecisionCount)
	}
	if m.PolicyOperations != 12 || m.TimedDecisionCount != 0 {
		t.Fatal("operation/timing denominators")
	}
	if m.DecisionP50Micros != nil || m.DecisionP95Micros != nil {
		t.Fatal("fabricated per-decision timing")
	}
	b, _ := json.Marshal(m)
	report := metricsReport(m)
	if strings.Contains(string(b)+report, "private-") {
		t.Fatal("private metric payload leak")
	}
	if strings.Contains(report, "no bot decisions") {
		t.Fatal("false absence claim")
	}
}
