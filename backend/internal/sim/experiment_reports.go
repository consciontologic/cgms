package sim

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"strings"
)

func experimentReports(x Experiment, jobs []Job, out []Outcome) (map[string][]byte, string, error) {
	files := map[string][]byte{}
	groups := map[string][]PairOutcome{}
	var report strings.Builder
	fmt.Fprintf(&report, "\n## Named experiment\n\n%s (%s): %s\n\nClaim scope: %s. Accepted rules; no interpretation replacements.\n\n", x.ID, x.Phase, x.Purpose, x.ClaimScope)
	for i, j := range jobs {
		key := hash(j.Parameters)
		o := out[i]
		status := o.Status
		if o.ScenarioStatus == "scenario-complete" {
			status = "scenario-complete"
		}
		if status == "replay-invariant-failure" {
			status = "replay-failure"
		}
		score := game.IntAmount(0)
		if o.MatchComplete {
			status = "finalized"
			found := false
			for seat, id := range j.Rotation.Seats {
				if id == j.FocalParticipantID && seat < len(o.Scores) {
					score = o.Scores[seat]
					found = true
				}
			}
			if !found {
				return nil, "", fmt.Errorf("missing focal finalized score")
			}
		}
		groups[key] = append(groups[key], PairOutcome{Block: j.Block, Rotation: j.Rotation.ID, Treatment: j.Treatment, Status: status, Score: score})
		b := mustCanonical(j)
		files[fmt.Sprintf("jobs/%06d/effective.json", i)] = b
	}
	for _, key := range sortedKeys(groups) {
		summary, e := Analyze(x, groups[key])
		if e != nil {
			return nil, "", e
		}
		files["analysis/"+key+".json"] = mustCanonical(summary)
		fmt.Fprintf(&report, "Parameter group %s: %d / %d complete paired seed blocks; %d missing or incomplete units. No imputation.\n\n", key[:12], summary.CompleteBlocks, summary.PlannedBlocks, len(summary.Missing))
		if summary.Mean != nil && summary.Lower != nil {
			fmt.Fprintf(&report, "Paired focal cumulative-match score effect: %s; seed-block bootstrap 95%% interval [%s, %s]. Conditional on complete blocks; not a causal action-usage estimate.\n\n", summary.Mean, summary.Lower, summary.Upper)
		}
	}
	if len(x.Blocks) < 2 {
		report.WriteString("Fewer than two independent planned seed blocks: uncertainty cannot be estimated; descriptive pilot only.\n\n")
	}
	return files, report.String(), nil
}
func policySummary(jobs []Job) string {
	seen := map[string]bool{}
	var b strings.Builder
	b.WriteString("\n## Policies and effective inputs\n\n")
	for _, j := range jobs {
		key := hash(j.Policies)
		if seen[key] {
			continue
		}
		seen[key] = true
		fmt.Fprintf(&b, "- %s: ", j.Treatment)
		for _, p := range j.Policies {
			fmt.Fprintf(&b, "%s=%s@%s ", p.ParticipantID, p.Policy, p.Version)
		}
		fmt.Fprintf(&b, "(policy configuration SHA-256 %s).\n", key)
	}
	return b.String()
}
