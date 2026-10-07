package sim

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"sort"
)

type PairOutcome struct {
	Block     string      `json:"block"`
	Rotation  string      `json:"rotation"`
	Treatment string      `json:"treatment"`
	Status    string      `json:"status"`
	Score     game.Amount `json:"score"`
}
type BlockDifference struct {
	Block      string      `json:"block"`
	Difference game.Amount `json:"difference"`
}
type MissingUnit struct {
	Block     string `json:"block"`
	Rotation  string `json:"rotation"`
	Treatment string `json:"treatment"`
	Status    string `json:"status"`
}
type Summary struct {
	Population      int               `json:"population"`
	PlannedBlocks   int               `json:"planned_blocks"`
	CompleteBlocks  int               `json:"complete_blocks"`
	Mean            *game.Amount      `json:"mean"`
	Lower           *game.Amount      `json:"lower_95"`
	Upper           *game.Amount      `json:"upper_95"`
	Differences     []BlockDifference `json:"differences"`
	Missing         []MissingUnit     `json:"missing"`
	StatusCounts    map[string]int    `json:"status_counts"`
	AnalysisVersion string            `json:"analysis_version"`
	Resamples       int               `json:"resamples"`
	Limitation      string            `json:"limitation"`
}

// Analyze estimates only financially finalized paired whole seed blocks. Scores in
// incomplete rows are never imputed or included, including scenario-complete rows.
func Analyze(x Experiment, rows []PairOutcome) (Summary, error) {
	out := Summary{Population: x.Population, PlannedBlocks: len(x.Blocks), StatusCounts: map[string]int{}, AnalysisVersion: "cgms-analysis-v1", Resamples: x.Analysis.Resamples, Limitation: "Conditional on complete paired blocks; selection bias possible. Partial pilots establish neither strength nor balance."}
	if e := x.Validate(); e != nil {
		return out, e
	}
	expected := map[string]bool{}
	key := func(b, r, t string) string { return b + "\x00" + r + "\x00" + t }
	for _, b := range x.Blocks {
		for _, r := range x.Rotations {
			for _, t := range []string{"baseline", "candidate"} {
				expected[key(b, r.ID, t)] = true
			}
		}
	}
	lookup := map[string]PairOutcome{}
	for _, r := range rows {
		k := key(r.Block, r.Rotation, r.Treatment)
		if !expected[k] {
			return out, fmt.Errorf("unmatched analysis row")
		}
		if _, ok := lookup[k]; ok {
			return out, fmt.Errorf("duplicate analysis row")
		}
		switch r.Status {
		case "finalized", "unsupported", "budget-exhausted", "canceled", "not-started", "replay-failure", "io-failure", "scenario-complete", "adjudication-required":
		default:
			return out, fmt.Errorf("unknown analysis status")
		}
		lookup[k] = r
	}
	for _, b := range x.Blocks {
		complete := true
		total := game.IntAmount(0)
		for _, r := range x.Rotations {
			pair := map[string]game.Amount{}
			for _, t := range []string{"baseline", "candidate"} {
				row, ok := lookup[key(b, r.ID, t)]
				if !ok {
					row = PairOutcome{Block: b, Rotation: r.ID, Treatment: t, Status: "not-started"}
				}
				out.StatusCounts[row.Status]++
				if row.Status != "finalized" {
					complete = false
					out.Missing = append(out.Missing, MissingUnit{b, r.ID, t, row.Status})
				}
				pair[t] = row.Score
			}
			total = total.Add(pair["candidate"].Sub(pair["baseline"]))
		}
		if complete {
			d, e := total.Quo(game.IntAmount(int64(len(x.Rotations))))
			if e != nil {
				return out, e
			}
			out.Differences = append(out.Differences, BlockDifference{b, d})
		}
	}
	out.CompleteBlocks = len(out.Differences)
	if out.CompleteBlocks == 0 {
		return out, nil
	}
	sum := game.IntAmount(0)
	for _, d := range out.Differences {
		sum = sum.Add(d.Difference)
	}
	mean, e := sum.Quo(game.IntAmount(int64(out.CompleteBlocks)))
	if e != nil {
		return out, e
	}
	out.Mean = &mean
	if out.CompleteBlocks < 2 {
		out.Limitation = "One complete seed block cannot estimate sampling uncertainty; descriptive mean only."
		return out, nil
	}
	seed, e := randomstream.AnalysisSeed(x.Analysis.RootSeed, x.Population, x.PairingID)
	if e != nil {
		return out, e
	}
	stream := randomstream.New(seed)
	means := make([]game.Amount, x.Analysis.Resamples)
	for i := range means {
		v := game.IntAmount(0)
		for j := 0; j < out.CompleteBlocks; j++ {
			index, e := stream.Sample(uint64(out.CompleteBlocks))
			if e != nil {
				return out, e
			}
			v = v.Add(out.Differences[index].Difference)
		}
		means[i], e = v.Quo(game.IntAmount(int64(out.CompleteBlocks)))
		if e != nil {
			return out, e
		}
	}
	sort.Slice(means, func(i, j int) bool { return means[i].Cmp(means[j]) < 0 })
	lower := means[(25*len(means)+999)/1000-1]
	upper := means[(975*len(means)+999)/1000-1]
	out.Lower = &lower
	out.Upper = &upper
	return out, nil
}
