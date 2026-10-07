package sim

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"io"
	"os"
	"path/filepath"
)

func verifyPrefix(m Manifest, files map[string][]byte) error {
	return verifyPrefixContext(context.Background(), m, files)
}
func verifyPrefixContext(ctx context.Context, m Manifest, files map[string][]byte) error {
	var rules game.Rules
	if e := canonical.Decode(files["effective-rules.json"], &rules); e != nil {
		return e
	}
	if hash(rules) != m.ConfigHash {
		return fmt.Errorf("configuration digest")
	}
	if _, e := game.LoadRules(files["source-rules.json"], files["effective-rules.json"]); e != nil {
		return e
	}
	if checksum(files["rule-source.md"]) != m.Provenance.RuleSourceSHA256 {
		return fmt.Errorf("rule source digest")
	}
	if e := verifyParentSnapshot(m, files); e != nil {
		return e
	}
	var outcomes []Outcome
	if e := decodeRequired(files["outcomes.json"], &outcomes); e != nil {
		return e
	}
	if e := verifySummary(m, outcomes); e != nil {
		return e
	}
	var jobs []Job
	if e := decodeRequired(files["schedule.json"], &jobs); e != nil {
		return e
	}
	pairing := "standalone"
	var resume *GameplayRun
	if raw := files["scenario.json"]; raw != nil {
		var sc Scenario
		if err := decodeRequired(raw, &sc); err != nil {
			return err
		}
		if sc.Kind == "gameplay-checkpoint" {
			resume = sc.Gameplay
			if resume == nil || len(jobs) != 1 || hash(jobs[0]) != hash(resume.Job) {
				return fmt.Errorf("continuation frozen schedule mismatch")
			}
			pairing = resume.Pairing
		}
	}
	if b := files["effective-experiment.json"]; b != nil {
		exp, e := LoadExperiment(b)
		if e != nil {
			return e
		}
		expected, e := exp.Schedule()
		if e != nil {
			return e
		}
		if hash(expected) != hash(jobs) {
			return fmt.Errorf("effective experiment schedule provenance")
		}
		pairing = exp.PairingID
	}
	if len(jobs) != len(outcomes) || len(jobs) != m.Planned {
		return fmt.Errorf("schedule/outcome denominator")
	}
	for i, j := range jobs {
		if e := ctx.Err(); e != nil {
			return e
		}
		if e := validateRecordedJob(j); e != nil {
			return e
		}
		if j.Index != i && resume == nil {
			return fmt.Errorf("schedule sequence")
		}
		prefix := fmt.Sprintf("games/%06d/", j.Index)
		if raw := files[prefix+"gameplay.json"]; raw != nil {
			if resume != nil {
				var continued GameplayRun
				if err := decodeRequired(raw, &continued); err != nil {
					return err
				}
				if len(continued.Trace) < len(resume.Trace) || !equalGameplayPrefix(continued.Trace, resume.Trace) || continued.PolicyOverride != resume.PolicyOverride {
					return fmt.Errorf("continuation committed prefix mismatch")
				}
			}
			if e := verifyGameplayUnit(ctx, m, j, outcomes[i], files, prefix, pairing); e != nil {
				return e
			}
			continue
		}
		if e := verifyUnitInputs(j, outcomes[i], files, prefix); e != nil {
			return e
		}
		b, ok := files[prefix+"trace.jsonl"]
		if !ok {
			if outcomes[i].Status != "not-started" {
				return fmt.Errorf("started unit missing trace")
			}
			continue
		}
		ts, e := decodeTrace(b)
		if e != nil {
			return e
		}
		if e := verifyScenarioTranscript(files["scenario.json"], j.Population, ts); e != nil {
			return e
		}
		if outcomes[i].Steps != len(ts) {
			return fmt.Errorf("outcome transition count")
		}
		initial := files[prefix+"initial.json"]
		if len(ts) == 0 {
			if !bytes.Equal(initial, files[prefix+"checkpoint.json"]) {
				return fmt.Errorf("empty trace checkpoint mismatch")
			}
			var state game.State
			if e := decodeRequired(initial, &state); e == nil {
				if e = state.Validate(); e != nil {
					return e
				}
			} else {
				var c game.SettlementCursor
				if e = decodeRequired(initial, &c); e != nil {
					return e
				}
				if _, _, e = game.ResumeSettlement(c, 0, true); e != nil {
					return e
				}
			}
			continue
		}
		if ts[0].Kind == "settlement" {
			var c game.SettlementCursor
			if e = decodeRequired(initial, &c); e != nil {
				return e
			}
			for i, t := range ts {
				if e := ctx.Err(); e != nil {
					return e
				}
				if t.Kind != "settlement" || t.GameID != c.Ledger.GameID || t.Command != nil || len(t.Shuffle) > 0 || t.RandomBefore != "" || t.RandomAfter != "" {
					return fmt.Errorf("settlement trace metadata")
				}
				if hash(c) != t.Pre {
					return fmt.Errorf("first divergence game %d transition %d pre", j.Index, i)
				}
				n, _, e := game.ResumeSettlement(c, 1, true)
				if e != nil {
					return e
				}
				if hash(n) != t.Post || hash(n.Audit[len(c.Audit):]) != t.Events {
					return fmt.Errorf("first divergence game %d transition %d state/events", j.Index, i)
				}
				c = n
			}
			if !bytes.Equal(mustCanonical(c), files[prefix+"checkpoint.json"]) {
				return fmt.Errorf("final checkpoint mismatch")
			}
			continue
		}
		var s game.State
		if e = decodeRequired(initial, &s); e != nil {
			return e
		}
		if e = s.Validate(); e != nil {
			return e
		}
		seed, e := randomstream.Derive(randomstream.Identity{Match: j.Match, Root: m.RootSeed, Population: j.Population, Pairing: pairing, Block: j.Block, Rotation: j.Rotation.ID, Kind: "deal"})
		if e != nil {
			return e
		}
		rng := randomstream.New(seed)
		botStreams := map[int]*randomstream.Stream{}
		for i, t := range ts {
			if e := ctx.Err(); e != nil {
				return e
			}
			if t.GameID != s.GameID {
				return fmt.Errorf("trace game identity")
			}
			if hash(s) != t.Pre {
				return fmt.Errorf("first divergence game %d transition %d pre", j.Index, i)
			}
			var n game.State
			var ev any
			switch t.Kind {
			case "deal":
				order := append([]string(nil), s.DrawOrder...)
				if t.RandomBefore != fmt.Sprint(rng.Counter) {
					return fmt.Errorf("random before counter")
				}
				rng.Shuffle(len(order), func(i, k int) { order[i], order[k] = order[k], order[i] })
				if hash(order) != hash(t.Shuffle) || t.RandomAfter != fmt.Sprint(rng.Counter) {
					return fmt.Errorf("random outcome divergence")
				}
				var redo bool
				n, redo, e = game.DealAttempt(s, t.Shuffle)
				ev = redo
			case "command":
				if e = verifyBotRandom(m, j, pairing, botStreams, t); e != nil {
					return e
				}
				if t.Command == nil {
					return fmt.Errorf("missing command")
				}
				var events []game.Event
				n, events, e = game.Apply(s, *t.Command)
				ev = events
			default:
				return fmt.Errorf("unknown trace kind")
			}
			if e != nil {
				return e
			}
			if hash(n) != t.Post || hash(ev) != t.Events {
				return fmt.Errorf("first divergence game %d transition %d state/events", j.Index, i)
			}
			s = n
		}
		if !bytes.Equal(mustCanonical(s), files[prefix+"checkpoint.json"]) {
			return fmt.Errorf("final checkpoint mismatch")
		}
	}
	return nil
}
func mustCanonical(v any) []byte {
	limit := canonical.MaxBytes
	switch x := v.(type) {
	case *GameplayRun:
		limit = gameplayTranscriptLimit(*x)
	case GameplayRun:
		limit = gameplayTranscriptLimit(x)
	case *Scenario:
		if x.Gameplay != nil {
			limit = gameplayScenarioLimit(*x.Gameplay)
		}
	case Scenario:
		if x.Gameplay != nil {
			limit = gameplayScenarioLimit(*x.Gameplay)
		}
	}
	b, e := canonical.MarshalLimit(v, limit)
	if e != nil {
		panic(e)
	}
	return b
}
func replayRun(ctx context.Context, path, out string, args []string, stdout, stderr io.Writer, hold ...bool) int {
	m, files, e := readManifestContext(ctx, path)
	if e != nil {
		fmt.Fprintln(stderr, "manifest integrity or compatibility failure")
		if errors.Is(e, ErrCompatibility) {
			return 2
		}
		if errors.Is(e, context.Canceled) {
			return 130
		}
		if errors.Is(e, ErrResourceBudget) {
			return 5
		}
		return artifactExit(e)
	}
	if e = reserveOutput(out); e != nil {
		if !os.IsExist(e) {
			return 6
		}
		return 2
	}
	code := m.ExitCode
	diagnosis := "recorded decisions and random outcomes verified; bots not rerun"
	if ctx.Err() != nil {
		code = 130
		diagnosis = "canceled before replay"
	} else if e = verifyPrefixContext(ctx, m, files); e != nil {
		code = 4
		if errors.Is(e, context.Canceled) {
			code = 130
		}
		diagnosis = e.Error()
	}
	replay := m
	if len(hold) > 0 && hold[0] {
		replay.ReviewHold = true
	}
	replay.Argv = args
	replay.Command = "replay"
	replay.Parent = path
	replay.Status = status(code)
	replay.ExitCode = code
	replay.Artifacts = nil
	files["replay-verifier.json"] = mustCanonical(provenance(files["rule-source.md"]))
	files["replay-diagnostic.txt"] = []byte(diagnosis)
	files["report.md"] = []byte("# Replay\n\nObserver: public.\n\n" + status(code) + ". Original incomplete status is retained. Restricted diagnostic contains divergence details.\n")
	if e = publish(out, replay, files); e != nil {
		return 6
	}
	fmt.Fprintln(stdout, filepath.Join(out, "manifest.json"))
	return code
}
func compareRuns(base, candidate, out string, args []string, stdout, stderr io.Writer) int {
	return compareRunsContext(context.Background(), base, candidate, out, args, stdout, stderr)
}
func compareRunsContext(ctx context.Context, base, candidate, out string, args []string, stdout, stderr io.Writer, hold ...bool) int {
	bm, bf, e := readManifestContext(ctx, base)
	if e != nil {
		return artifactExit(e)
	}
	cm, cf, e := readManifestContext(ctx, candidate)
	if e != nil {
		return artifactExit(e)
	}
	if bm.Provenance.Engine != cm.Provenance.Engine || bm.ConfigHash != cm.ConfigHash || bm.Interpretation != cm.Interpretation || bm.Provenance.RuleSourceSHA256 != cm.Provenance.RuleSourceSHA256 {
		return 2
	}
	if e = verifyPrefixContext(ctx, bm, bf); e != nil {
		if errors.Is(e, context.Canceled) {
			return 130
		}
		return 4
	}
	if e = verifyPrefixContext(ctx, cm, cf); e != nil {
		if errors.Is(e, context.Canceled) {
			return 130
		}
		return 4
	}
	if e := compatiblePairing(bf, cf); e != nil {
		return 2
	}
	var bo, co []Outcome
	if canonical.Decode(bf["outcomes.json"], &bo) != nil || canonical.Decode(cf["outcomes.json"], &co) != nil {
		return 4
	}
	if hash(string(bf["execution-limits.json"])) != hash(string(cf["execution-limits.json"])) {
		return 2
	}
	if len(bo) != len(co) || bm.RootSeed != cm.RootSeed {
		return 2
	}
	for i := range bo {
		a, b := bo[i], co[i]
		if a.Population != b.Population || a.Block != b.Block || a.Rotation != b.Rotation || hash(a.Seats) != hash(b.Seats) {
			return 2
		}
	}
	if e = reserveOutput(out); e != nil {
		if !os.IsExist(e) {
			return 6
		}
		return 2
	}
	m := manifestBase(args, "compare", "", bm.ConfigHash, 1, len(bo), bf["rule-source.md"])
	m.Status = "complete"
	if len(hold) > 0 {
		m.ReviewHold = hold[0]
	}
	m.Parent = base + " | " + candidate
	m.Interpretation = bm.Interpretation
	text := "# Paired comparison\n\nObserver: public. Analysis artifact complete; input games retain their original status.\n\n" + completedComparison(bo, co)
	for _, pop := range []int{3, 4} {
		planned, bc, cc := 0, 0, 0
		for i, a := range bo {
			if a.Population != pop {
				continue
			}
			planned++
			bc += boolInt(a.GameComplete)
			cc += boolInt(co[i].GameComplete)
		}
		text += fmt.Sprintf("%d-player planned paired slots: %d; baseline finalized: %d; candidate finalized: %d.\n\n", pop, planned, bc, cc)
	}
	var bj, cj []Job
	if decodeRequired(bf["schedule.json"], &bj) != nil || decodeRequired(cf["schedule.json"], &cj) != nil {
		return 4
	}
	text += fmt.Sprintf("Baseline engine binary: %s; candidate engine binary: %s. Rules hash: %s. Equal recorded operation limits.\n\n", bm.Provenance.BinarySHA256, cm.Provenance.BinarySHA256, bm.ConfigHash)
	text += "Baseline:\n\n" + report(bm.Command, bo, nil) + policySummary(bj) + "\nCandidate:\n\n" + report(cm.Command, co, nil) + policySummary(cj)
	files := map[string][]byte{"report.md": []byte(text), "baseline-outcomes.json": bf["outcomes.json"], "candidate-outcomes.json": cf["outcomes.json"]}
	if e = publish(out, m, files); e != nil {
		return 6
	}
	fmt.Fprintln(stdout, filepath.Join(out, "manifest.json"))
	return 0
}

func artifactExit(e error) int {
	if errors.Is(e, ErrCompatibility) {
		return 2
	}
	if errors.Is(e, context.Canceled) {
		return 130
	}
	if errors.Is(e, ErrResourceBudget) || errors.Is(e, context.DeadlineExceeded) {
		return 5
	}
	var path *os.PathError
	if errors.As(e, &path) {
		return 6
	}
	return 4
}

func equalGameplayPrefix(a, b []GameplayStep) bool {
	if len(a) < len(b) {
		return false
	}
	for i := range b {
		if hash(a[i]) != hash(b[i]) {
			return false
		}
	}
	return true
}
