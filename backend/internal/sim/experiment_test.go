package sim

import (
	"os"
	"strings"
	"testing"
)

func TestExperiment(t *testing.T) {
	for _, p := range []string{"3", "4"} {
		b, e := os.ReadFile("../../../sims/experiments/baseline-" + p + "p.json")
		if e != nil {
			t.Fatal(e)
		}
		x, e := LoadExperiment(b)
		if e != nil {
			t.Fatal(e)
		}
		if x.Population != 3 && x.Population != 4 {
			t.Fatal(x)
		}
		bad := strings.Replace(string(b), `"complete-cyclic"`, `"partial"`, 1)
		if _, e = LoadExperiment([]byte(bad)); e == nil {
			t.Fatal("accepted partial")
		}
	}
}
func TestScheduleAndInvalidRotations(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	x, _ := LoadExperiment(b)
	jobs, e := x.Schedule()
	if e != nil || len(jobs) != 24 {
		t.Fatalf("jobs %d %v", len(jobs), e)
	}
	for i := 0; i < len(jobs); i += 2 {
		if jobs[i].Block != jobs[i+1].Block || jobs[i].Rotation.ID != jobs[i+1].Rotation.ID {
			t.Fatal("not paired")
		}
	}
	x.Rotations[1] = x.Rotations[0]
	if x.Validate() == nil {
		t.Fatal("duplicate rotation accepted")
	}
	x, _ = LoadExperiment(b)
	x.DevelopmentBlocks = []string{x.Blocks[0]}
	if x.Validate() == nil {
		t.Fatal("overlap accepted")
	}
	x, _ = LoadExperiment(b)
	x.Grid = []GridParameter{{"rounds", []int{10}}}
	if x.Validate() == nil {
		t.Fatal("gameplay sweep accepted")
	}
}
func TestRequiredAndPolicyPrivacy(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	for _, bad := range []string{strings.Replace(string(b), `"review_hold": false,`, "", 1), strings.Replace(string(b), `"seat-projection"`, `"omniscient"`, 1), strings.Replace(string(b), `"workers": 2`, `"workers": 0`, 1), strings.Replace(string(b), `"max_matches": 24`, `"max_matches": 23`, 1)} {
		if _, e := LoadExperiment([]byte(bad)); e == nil {
			t.Fatal("invalid accepted")
		}
	}
}

func TestExperimentOptionalMenuBudget(t *testing.T) {
	raw, e := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	if e != nil {
		t.Fatal(e)
	}
	old, e := LoadExperiment(raw)
	if e != nil {
		t.Fatal(e)
	}
	if old.Budgets.MenuOperations() != 100000 {
		t.Fatal("legacy menu default changed")
	}
	for value, want := range map[string]int{"0": 100000, "1": 1, "1000000": 1000000, "10000000": 10000000} {
		changed := strings.Replace(string(raw), `"budgets": {`, `"budgets": {"max_menu_operations_per_decision":`+value+`,`, 1)
		x, e := LoadExperiment([]byte(changed))
		if e != nil || x.Budgets.MenuOperations() != want {
			t.Fatal(value, e)
		}
	}
	for _, value := range []string{"-1", "10000001", "null"} {
		changed := strings.Replace(string(raw), `"budgets": {`, `"budgets": {"max_menu_operations_per_decision":`+value+`,`, 1)
		if _, e := LoadExperiment([]byte(changed)); e == nil {
			t.Fatal("invalid menu budget", value)
		}
	}
}

func TestExperimentOptionalTranscriptBudget(t *testing.T) {
	raw, e := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	if e != nil {
		t.Fatal(e)
	}
	for _, value := range []string{"0", "4194304", "33554432"} {
		changed := strings.Replace(string(raw), `"budgets": {`, `"budgets": {"max_gameplay_transcript_bytes":`+value+`,`, 1)
		if _, e := LoadExperiment([]byte(changed)); e != nil {
			t.Fatal(value, e)
		}
	}
	for _, value := range []string{"-1", "1", "67108864", "null"} {
		changed := strings.Replace(string(raw), `"budgets": {`, `"budgets": {"max_gameplay_transcript_bytes":`+value+`,`, 1)
		if _, e := LoadExperiment([]byte(changed)); e == nil {
			t.Fatal("invalid transcript budget", value)
		}
	}
}
