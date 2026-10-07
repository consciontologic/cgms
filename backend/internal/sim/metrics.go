package sim

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Metrics struct {
	PolicyOperations     int    `json:"policy_operations"`
	TimedDecisionCount   int    `json:"timed_decision_count"`
	Schema               string `json:"schema_version"`
	ElapsedMicros        int64  `json:"elapsed_microseconds"`
	DecisionCount        int    `json:"decision_count"`
	DecisionP50Micros    *int64 `json:"decision_p50_microseconds"`
	DecisionP95Micros    *int64 `json:"decision_p95_microseconds"`
	PeakSampledHeapBytes uint64 `json:"peak_sampled_heap_bytes"`
	FinalizedGames       int    `json:"finalized_games"`
	AttemptedScenarios   int    `json:"attempted_scenarios"`
	Actions              int    `json:"actions"`
	Redeals              int    `json:"redeals"`
	Platform             string `json:"platform"`
	LogicalCPUs          int    `json:"logical_cpus"`
	Limitation           string `json:"limitation"`
}

func startMetrics() func([]runResult) Metrics {
	started := time.Now()
	done := make(chan struct{})
	var wg sync.WaitGroup
	var peak uint64
	sample := func() {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.HeapAlloc > peak {
			peak = m.HeapAlloc
		}
	}
	sample()
	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-done:
				sample()
				return
			case <-tick.C:
				sample()
			}
		}
	}()
	return func(rs []runResult) Metrics {
		close(done)
		wg.Wait()
		m := Metrics{Schema: "cgms-metrics-v1", ElapsedMicros: time.Since(started).Microseconds(), PeakSampledHeapBytes: peak, Platform: runtime.GOOS + "/" + runtime.GOARCH, LogicalCPUs: runtime.NumCPU(), Limitation: "Process-wide sampled Go heap, not RSS; 10ms sampling may miss peaks. Timings are diagnostics excluded from semantic hashes. No performance release threshold has been established."}
		var times []int64
		for _, r := range rs {
			times = append(times, r.DecisionMicros...)
			if r.Gameplay != nil {
				for _, step := range r.Gameplay.Trace {
					if step.Decision != nil {
						m.DecisionCount++
						m.PolicyOperations += step.Decision.Operations
					}
					if step.FinanceDecision != nil {
						m.DecisionCount++
						m.PolicyOperations += step.FinanceDecision.Operations
					}
				}
			} else {
				m.DecisionCount += len(r.DecisionMicros)
				for _, step := range r.Trace {
					if step.BotDecision != nil {
						m.PolicyOperations += step.BotDecision.Operations
					}
				}
			}
			m.Actions += len(r.Trace)
			m.Redeals += r.Outcome.Redeals
			if r.Initial != nil {
				m.AttemptedScenarios++
			}
			m.FinalizedGames += r.Outcome.FinalizedGames
		}
		m.TimedDecisionCount = len(times)
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
		if len(times) > 0 {
			p50, p95 := times[(50*len(times)+99)/100-1], times[(95*len(times)+99)/100-1]
			m.DecisionP50Micros = &p50
			m.DecisionP95Micros = &p95
		}
		return m
	}
}
func metricsReport(m Metrics) string {
	decision := fmt.Sprintf("not measured (%d recorded decisions; %d deterministic policy operations)", m.DecisionCount, m.PolicyOperations)
	if m.DecisionP50Micros != nil {
		decision = fmt.Sprintf("p50 %d µs; p95 %d µs (%d decisions)", *m.DecisionP50Micros, *m.DecisionP95Micros, m.TimedDecisionCount)
	}
	return fmt.Sprintf("\n## Measured execution\n\n%d recorded actions, %d redeals, %d attempted scenarios and %d finalized games in %d µs. Bot latency: %s. Sampled peak Go heap: %d bytes on %s, %d logical CPUs.\n\n%s\n", m.Actions, m.Redeals, m.AttemptedScenarios, m.FinalizedGames, m.ElapsedMicros, decision, m.PeakSampledHeapBytes, m.Platform, m.LogicalCPUs, m.Limitation)
}
