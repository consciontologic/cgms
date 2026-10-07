// Package sim contains local research orchestration, separate from adjudication.
package sim

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

type Policy struct {
	ParticipantID string   `json:"participant_id"`
	Policy        string   `json:"policy"`
	Version       string   `json:"version"`
	Information   string   `json:"information"`
	Features      []string `json:"features"`
	Weights       []int    `json:"weights"`
}
type Rotation struct {
	ID    string   `json:"id"`
	Seats []string `json:"seats"`
}
type GridParameter struct {
	Parameter string `json:"parameter"`
	Values    []int  `json:"values"`
}
type Budgets struct {
	MaxGameplayTranscriptBytes int `json:"max_gameplay_transcript_bytes,omitempty"`
	MaxMenuOperations          int `json:"max_menu_operations_per_decision,omitempty"`
	Workers                    int `json:"workers"`
	MaxMatches                 int `json:"max_matches"`
	MaxTransitions             int `json:"max_transitions_per_match"`
	MaxSettlementSteps         int `json:"max_settlement_steps_per_match"`
	MaxBotOperations           int `json:"max_bot_operations_per_decision"`
	MaxOutputBytes             int `json:"max_output_bytes"`
	MaxInputBytes              int `json:"max_input_bytes"`
	MaxRationalDigits          int `json:"max_input_rational_digits"`
	WallSeconds                int `json:"wall_watchdog_seconds"`
}

// MenuOperations preserves the historical cap when the optional field is omitted or zero.
func (b Budgets) MenuOperations() int {
	if b.MaxMenuOperations == 0 {
		return 100000
	}
	return b.MaxMenuOperations
}

type AnalysisConfig struct {
	Version          string `json:"version"`
	RootSeed         string `json:"root_seed"`
	PrimaryMetric    string `json:"primary_metric"`
	Resamples        int    `json:"resamples"`
	Confidence       int    `json:"confidence_percent"`
	Cluster          string `json:"cluster"`
	IncompletePolicy string `json:"incomplete_policy"`
}
type OutputPolicy struct {
	Root          string `json:"root"`
	PrivateReplay bool   `json:"private_replay"`
	Observer      string `json:"report_observer"`
	Omniscient    bool   `json:"omniscient_comparison"`
}
type Retention struct {
	Days       int    `json:"days"`
	ReviewHold bool   `json:"review_hold"`
	Prune      string `json:"prune"`
}
type Experiment struct {
	SchemaVersion     string          `json:"schema_version"`
	ID                string          `json:"id"`
	Kind              string          `json:"kind"`
	Purpose           string          `json:"purpose"`
	Population        int             `json:"population"`
	Phase             string          `json:"phase"`
	ClaimScope        string          `json:"claim_scope"`
	RulesProfile      string          `json:"rules_profile"`
	Interpretations   []string        `json:"interpretation_replacements"`
	EngineVersion     string          `json:"engine_version"`
	RequiredCoverage  []string        `json:"required_coverage"`
	PairingID         string          `json:"pairing_id"`
	RootSeed          string          `json:"root_seed"`
	Blocks            []string        `json:"block_ids"`
	DevelopmentBlocks []string        `json:"development_block_ids,omitempty"`
	GamesPerMatch     int             `json:"games_per_match"`
	MatchIndex        int             `json:"match_index"`
	Bots              []Policy        `json:"bots"`
	Candidate         Policy          `json:"candidate"`
	RotationPolicy    string          `json:"rotation_policy"`
	Rotations         []Rotation      `json:"rotations"`
	Grid              []GridParameter `json:"grid"`
	Budgets           Budgets         `json:"budgets"`
	Analysis          AnalysisConfig  `json:"analysis"`
	Output            OutputPolicy    `json:"output"`
	Retention         Retention       `json:"retention"`
}

func LoadExperiment(b []byte) (Experiment, error) {
	var x Experiment
	if e := canonical.Decode(b, &x); e != nil {
		return x, e
	}
	var raw any
	if e := canonical.Decode(b, &raw); e != nil {
		return x, e
	}
	if e := requiredExperiment(raw, reflect.TypeOf(x)); e != nil {
		return x, e
	}
	return x, x.Validate()
}
func (p Policy) Validate() error {
	if !safeExperimentID.MatchString(p.ParticipantID) || p.Information != "seat-projection" || (p.Version != "v1" && !bots.GameplayOrderV2(p.Policy+"@"+p.Version) && !(p.Policy == "gameplay-random" && p.Version == "v2")) || !slices.Contains([]string{"legal-random", "heuristic", "economic", "pressure", "gameplay-random", "opportunity", "opportunity-no-retain", "opportunity-no-finance", "bargaining"}, p.Policy) || len(p.Features) != len(p.Weights) {
		return fmt.Errorf("unsupported fair policy")
	}
	allowed := []string{"immediate-score", "exposed-holdings", "protection", "known-debt", "legal-declaration"}
	for _, w := range p.Weights {
		if w < -10000 || w > 10000 {
			return fmt.Errorf("weight bounds")
		}
	}
	if len(p.Features) > 20 {
		return fmt.Errorf("feature bounds")
	}
	for _, f := range p.Features {
		if !slices.Contains(allowed, f) {
			return fmt.Errorf("unknown feature")
		}
	}
	return nil
}
func (x Experiment) Validate() error {
	bad := func(s string) error { return fmt.Errorf("invalid experiment: %s", s) }
	if x.SchemaVersion != "cgms-experiment-v1" || (x.EngineVersion != "cgms-engine-v1" && x.EngineVersion != EngineVersion) || x.Kind != "experiment" || !safeExperimentID.MatchString(x.ID) || x.Purpose == "" || len(x.Purpose) > 256 || !safeExperimentID.MatchString(x.PairingID) || x.RulesProfile != "accepted-game-rules" || len(x.Interpretations) > 0 {
		return bad("identity/profile")
	}
	if x.Population != 3 && x.Population != 4 {
		return bad("population")
	}
	if x.Phase != "development-pilot" && x.Phase != "held-out" {
		return bad("phase")
	}
	if x.ClaimScope != "partial-scenario-only" && !(x.ClaimScope == "accepted-gameplay" && x.EngineVersion == EngineVersion) {
		return bad("coverage claim")
	}
	if x.GamesPerMatch < 1 || x.GamesPerMatch > 100000 || x.MatchIndex < 0 {
		return bad("match bounds")
	}
	if _, e := randomstream.AnalysisSeed(x.RootSeed, x.Population, x.PairingID); e != nil {
		return e
	}
	if len(x.Bots) != x.Population {
		return bad("lineup size")
	}
	ids := []string{}
	for _, p := range x.Bots {
		if e := p.Validate(); e != nil {
			return e
		}
		if slices.Contains(ids, p.ParticipantID) {
			return bad("duplicate participant")
		}
		ids = append(ids, p.ParticipantID)
	}
	if e := x.Candidate.Validate(); e != nil {
		return e
	}
	if !slices.Contains(ids, x.Candidate.ParticipantID) {
		return bad("candidate absent")
	}
	seen := map[string]bool{}
	for _, b := range x.Blocks {
		if !safeExperimentID.MatchString(b) || seen[b] {
			return bad("duplicate block")
		}
		seen[b] = true
	}
	if len(seen) == 0 || len(seen) > 10000 {
		return bad("empty blocks")
	}
	for _, b := range x.DevelopmentBlocks {
		if seen[b] {
			return bad("development/held-out overlap")
		}
	}
	if len(x.RequiredCoverage) == 0 || len(x.RequiredCoverage) > 10 {
		return bad("coverage")
	}
	for _, c := range x.RequiredCoverage {
		if !slices.Contains([]string{"S04", "S05", "S06", "S07", "D01", "D02", "D03", "D04", "D08-domain", "D09-domain"}, c) {
			return bad("coverage")
		}
	}
	expected := map[string]bool{}
	switch x.RotationPolicy {
	case "complete-cyclic":
		for r := 0; r < x.Population; r++ {
			a := make([]string, x.Population)
			for i, id := range ids {
				a[(i+r)%x.Population] = id
			}
			expected[strings.Join(a, "\x00")] = true
		}
	case "full-permutation":
		var permute func([]string, int)
		permute = func(a []string, i int) {
			if i == len(a) {
				expected[strings.Join(a, "\x00")] = true
				return
			}
			for j := i; j < len(a); j++ {
				a[i], a[j] = a[j], a[i]
				permute(a, i+1)
				a[i], a[j] = a[j], a[i]
			}
		}
		permute(slices.Clone(ids), 0)
	default:
		return bad("rotation policy")
	}
	if len(x.Rotations) != len(expected) {
		return bad("incomplete rotations")
	}
	rotIDs := map[string]bool{}
	for _, r := range x.Rotations {
		key := strings.Join(r.Seats, "\x00")
		if !safeExperimentID.MatchString(r.ID) || rotIDs[r.ID] || !expected[key] {
			return bad("rotation identity/mapping")
		}
		rotIDs[r.ID] = true
		delete(expected, key)
	}
	b := x.Budgets
	if (b.MaxGameplayTranscriptBytes != 0 && b.MaxGameplayTranscriptBytes != canonical.MaxBytes && b.MaxGameplayTranscriptBytes != 32<<20) || b.MaxMenuOperations < 0 || b.MaxMenuOperations > 10000000 || b.Workers < 1 || b.Workers > 256 || b.MaxMatches < 1 || b.MaxMatches > 1000000 || b.MaxTransitions < 1 || b.MaxTransitions > 10000000 || b.MaxSettlementSteps < 1 || b.MaxSettlementSteps > 10000000 || b.MaxBotOperations < 1 || b.MaxBotOperations > 10000000 || b.MaxOutputBytes < 1024 || b.MaxOutputBytes > 1073741824 || b.MaxInputBytes < 1024 || b.MaxInputBytes > canonical.MaxBytes || b.MaxRationalDigits < 1 || b.MaxRationalDigits > 256 || b.WallSeconds < 1 || b.WallSeconds > 86400 {
		return bad("budgets")
	}
	product := int64(len(x.Blocks)) * int64(len(x.Rotations)) * 2
	parameters := map[string]bool{}
	for _, g := range x.Grid {
		if !slices.Contains([]string{"bot.immediate-score", "bot.exposed-holdings", "bot.protection", "bot.known-debt", "bot.legal-declaration", "budget.bot-operations", "budget.transitions"}, g.Parameter) || parameters[g.Parameter] || len(g.Values) == 0 {
			return bad("grid")
		}
		parameters[g.Parameter] = true
		for _, v := range g.Values {
			if (strings.HasPrefix(g.Parameter, "budget.") && (v < 1 || v > 1000000)) || (strings.HasPrefix(g.Parameter, "bot.") && (v < -10000 || v > 10000)) {
				return bad("grid value")
			}
		}
		if product > int64(b.MaxMatches)/int64(len(g.Values)) {
			return bad("grid work bound")
		}
		product *= int64(len(g.Values))
	}
	if product > int64(b.MaxMatches) || product*int64(x.GamesPerMatch) > 1000000 {
		return bad("schedule work bound")
	}
	a := x.Analysis
	if a.Version != "cgms-analysis-v1" || a.PrimaryMetric != "paired-final-match-score" || a.Resamples < 1 || a.Resamples > 100000 || a.Confidence != 95 || a.Cluster != "seed-block" || a.IncompletePolicy != "report-all-complete-pairs-only" {
		return bad("analysis")
	}
	if _, e := randomstream.AnalysisSeed(a.RootSeed, x.Population, x.PairingID); e != nil {
		return e
	}
	if !x.Output.PrivateReplay || x.Output.Observer != "public" || x.Output.Omniscient || x.Output.Root != "sims/artifacts" || x.Retention.Days != 30 || x.Retention.Prune != "preview-before-delete" {
		return bad("privacy/retention")
	}
	return nil
}

type Job struct {
	Match              int            `json:"match_index"`
	FocalParticipantID string         `json:"focal_participant_id"`
	Index              int            `json:"index"`
	Population         int            `json:"population"`
	Block              string         `json:"block"`
	Rotation           Rotation       `json:"rotation"`
	Treatment          string         `json:"treatment"`
	Parameters         map[string]int `json:"parameters"`
	Games              int            `json:"games"`
	Policies           []Policy       `json:"policies"`
}

func (x Experiment) Schedule() ([]Job, error) {
	if e := x.Validate(); e != nil {
		return nil, e
	}
	grids := []map[string]int{{}}
	for _, p := range x.Grid {
		next := []map[string]int{}
		for _, g := range grids {
			for _, v := range p.Values {
				m := map[string]int{}
				for k, w := range g {
					m[k] = w
				}
				m[p.Parameter] = v
				next = append(next, m)
			}
		}
		grids = next
	}
	jobs := []Job{}
	for _, g := range grids {
		for _, b := range x.Blocks {
			for _, r := range x.Rotations {
				for _, t := range []string{"baseline", "candidate"} {
					policies := []Policy{}
					for _, id := range r.Seats {
						for _, p := range x.Bots {
							if p.ParticipantID == id {
								if t == "candidate" && id == x.Candidate.ParticipantID {
									p = x.Candidate
								}
								p.Features = slices.Clone(p.Features)
								p.Weights = slices.Clone(p.Weights)
								policies = append(policies, p)
							}
						}
					}
					params := map[string]int{}
					for k, v := range g {
						params[k] = v
					}
					jobs = append(jobs, Job{Match: x.MatchIndex, FocalParticipantID: x.Candidate.ParticipantID, Index: len(jobs), Population: x.Population, Block: b, Rotation: Rotation{r.ID, slices.Clone(r.Seats)}, Treatment: t, Parameters: params, Games: x.GamesPerMatch, Policies: policies})
				}
			}
		}
	}
	return jobs, nil
}

var safeExperimentID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

func requiredExperiment(v any, t reflect.Type) error {
	if v == nil {
		return fmt.Errorf("null experiment field")
	}
	switch t.Kind() {
	case reflect.Struct:
		m, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("object required")
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			tag := strings.Split(f.Tag.Get("json"), ",")
			child, ok := m[tag[0]]
			if !ok {
				if len(tag) > 1 && tag[1] == "omitempty" {
					continue
				}
				return fmt.Errorf("missing experiment field %s", tag[0])
			}
			if e := requiredExperiment(child, f.Type); e != nil {
				return e
			}
		}
	case reflect.Slice:
		a, ok := v.([]any)
		if !ok {
			return fmt.Errorf("array required")
		}
		for _, child := range a {
			if e := requiredExperiment(child, t.Elem()); e != nil {
				return e
			}
		}
	}
	return nil
}
