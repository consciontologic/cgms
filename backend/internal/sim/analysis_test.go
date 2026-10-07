package sim

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"testing"
)

func TestExactPairedBlocks(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	x, e := LoadExperiment(b)
	if e != nil {
		t.Fatal(e)
	}
	x.Blocks = []string{"a", "b", "c"}
	x.Analysis.Resamples = 100
	rows := []PairOutcome{}
	for _, block := range x.Blocks {
		for _, r := range x.Rotations {
			rows = append(rows, PairOutcome{block, r.ID, "baseline", "finalized", game.IntAmount(1)})
			v := int64(3)
			if block == "b" {
				v = 5
			}
			status := "finalized"
			if block == "c" && r.ID == "r2" {
				status = "unsupported"
			}
			rows = append(rows, PairOutcome{block, r.ID, "candidate", status, game.IntAmount(v)})
		}
	}
	s, e := Analyze(x, rows)
	if e != nil {
		t.Fatal(e)
	}
	if s.CompleteBlocks != 2 || s.Mean == nil || s.Mean.String() != "3" {
		t.Fatalf("%+v", s)
	}
}
func TestNoCompleteAndUnmatched(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-4p.json")
	x, _ := LoadExperiment(b)
	s, e := Analyze(x, nil)
	if e != nil || s.Mean != nil || s.CompleteBlocks != 0 || len(s.Missing) != len(x.Blocks)*len(x.Rotations)*2 {
		t.Fatalf("%+v %v", s, e)
	}
	if _, e = Analyze(x, []PairOutcome{{Block: "alien"}}); e == nil {
		t.Fatal("unmatched accepted")
	}
}
func TestBootstrapExactQuantiles(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	x, _ := LoadExperiment(b)
	x.Blocks = []string{"a", "b"}
	x.Analysis.Resamples = 10000
	rows := []PairOutcome{}
	for _, block := range x.Blocks {
		for _, r := range x.Rotations {
			v := int64(2)
			if block == "b" {
				v = 4
			}
			rows = append(rows, PairOutcome{block, r.ID, "baseline", "finalized", game.IntAmount(0)}, PairOutcome{block, r.ID, "candidate", "finalized", game.IntAmount(v)})
		}
	}
	s, e := Analyze(x, rows)
	if e != nil || s.Lower.String() != "2" || s.Upper.String() != "4" {
		t.Fatalf("%+v %v", s, e)
	}
}
