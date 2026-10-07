package sim

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func sortedKeys[V any](m map[string]V) []string {
	a := make([]string, 0, len(m))
	for k := range m {
		a = append(a, k)
	}
	sort.Strings(a)
	return a
}
func status(code int) string {
	switch code {
	case 0:
		return "complete"
	case 2:
		return "invalid-input"
	case 3:
		return "unsupported"
	case 4:
		return "replay-invariant-failure"
	case 5:
		return "budget-exhausted"
	case 6:
		return "io-failure"
	case 130:
		return "canceled"
	}
	return "invalid-input"
}
func aggregate(a, b int) int {
	priority := map[int]int{0: 0, 5: 1, 3: 2, 130: 3, 4: 4, 6: 5, 2: 6}
	if priority[b] > priority[a] {
		return b
	}
	return a
}
func readInput(path string) ([]byte, error) {
	return readArtifact(context.Background(), path, canonical.MaxBytes)
}
func Execute(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return 2
	}
	command := args[0]
	if !contains([]string{"run", "batch", "tournament", "sweep", "replay", "compare"}, command) {
		return 2
	}
	f := flag.NewFlagSet(command, flag.ContinueOnError)
	f.SetOutput(io.Discard)
	config := f.String("config", "", "")
	seed := f.String("seed", "", "")
	players := f.Int("players", 0, "")
	policies := f.String("bots", "", "")
	out := f.String("out", "", "")
	scenario := f.String("scenario", "", "")
	games := f.Int("games", 1, "")
	workers := f.Int("workers", 1, "")
	experiment := f.String("experiment", "", "")
	interpret := f.String("interpretations", "accepted-game-rules", "")
	manifest := f.String("manifest", "", "")
	baseline := f.String("baseline", "", "")
	candidate := f.String("candidate", "", "")
	budget := f.Int("max-steps", 10000, "")
	hold := f.Bool("review-hold", false, "")
	// Flag permits duplicates; the simulator contract deliberately does not.
	seen := map[string]bool{}
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "-") && !strings.HasPrefix(a, "--") {
			return 2
		}
		if strings.HasPrefix(a, "--") {
			k := strings.SplitN(strings.TrimPrefix(a, "--"), "=", 2)[0]
			if seen[k] {
				return 2
			}
			seen[k] = true
		}
	}
	allowed := map[string][]string{"run": {"config", "seed", "players", "bots", "out", "scenario", "workers", "interpretations", "max-steps"}, "batch": {"config", "seed", "players", "bots", "out", "scenario", "workers", "interpretations", "max-steps", "games"}, "tournament": {"config", "seed", "players", "bots", "out", "scenario", "workers", "interpretations", "max-steps", "games", "experiment"}, "sweep": {"config", "seed", "players", "bots", "out", "scenario", "workers", "interpretations", "max-steps", "games", "experiment"}, "replay": {"manifest", "out"}, "compare": {"baseline", "candidate", "out"}}
	for k := range seen {
		if k != "review-hold" && !contains(allowed[command], k) {
			return 2
		}
	}
	if f.Parse(args[1:]) != nil || f.NArg() != 0 || *out == "" {
		fmt.Fprintln(stderr, "invalid input")
		return 2
	}
	if command == "replay" {
		if *manifest == "" {
			return 2
		}
		return replayRun(ctx, *manifest, *out, args, stdout, stderr, *hold)
	}
	if command == "compare" {
		if *baseline == "" || *candidate == "" {
			return 2
		}
		return compareRunsContext(ctx, *baseline, *candidate, *out, args, stdout, stderr, *hold)
	}
	var resume *GameplayRun
	if command == "run" && *scenario != "" {
		b, err := readArtifact(context.Background(), *scenario, (32<<20)+65536)
		if err != nil {
			return 6
		}
		var preliminary Scenario
		if err = decodeRequired(b, &preliminary); err != nil {
			return 2
		}
		if preliminary.Kind == "gameplay-checkpoint" {
			if preliminary.Gameplay == nil {
				return 2
			}
			resume = preliminary.Gameplay
			ps := []string{}
			for _, p := range resume.Job.Policies {
				ps = append(ps, p.Policy+"@"+p.Version)
			}
			frozenBots := strings.Join(ps, ",")
			if seen["seed"] && *seed != resume.Root || seen["players"] && *players != resume.Job.Population || seen["bots"] && *policies != frozenBots {
				return 2
			}
			*seed, *players, *policies = resume.Root, resume.Job.Population, frozenBots
		}
	}
	if *config == "" || *seed == "" || *workers < 1 || *workers > 256 || *games < 1 || *games > 100000 || *budget < 1 || *budget > 1000000 || *interpret != "accepted-game-rules" {
		return 2
	}
	raw, e := readInput(*config)
	if e != nil {
		return 6
	}
	defaultsPath, e := findDefaults(*config)
	if e != nil {
		return 2
	}
	defaults, e := readInput(defaultsPath)
	if e != nil {
		return 6
	}
	rules, e := game.LoadRules(defaults, raw)
	if e != nil {
		return 2
	}
	rulePath := filepath.Join(filepath.Dir(defaultsPath), "../../docs/project/GAME_RULES.md")
	ruleSource, e := readInput(rulePath)
	if e != nil {
		return 6
	}
	pairing := "standalone"
	jobs := []Job{}
	var expRaw []byte
	var exp Experiment
	if command == "tournament" || command == "sweep" {
		if *experiment == "" {
			return 2
		}
		expRaw, e = readInput(*experiment)
		if e != nil {
			return 6
		}
		exp, e = LoadExperiment(expRaw)
		if e != nil {
			return 2
		}
		exp.RootSeed = *seed
		if seen["players"] {
			exp.Population = *players
		}
		if seen["workers"] {
			exp.Budgets.Workers = *workers
		} else {
			*workers = exp.Budgets.Workers
		}
		if seen["games"] {
			exp.GamesPerMatch = *games
		}
		if seen["max-steps"] {
			exp.Budgets.MaxTransitions = *budget
		} else {
			*budget = exp.Budgets.MaxTransitions
		}
		if *policies != "" {
			ps := strings.Split(*policies, ",")
			if len(ps) != len(exp.Bots) {
				return 2
			}
			for i, p := range ps {
				bits := strings.Split(p, "@")
				if len(bits) != 2 {
					return 2
				}
				exp.Bots[i].Policy = bits[0]
				exp.Bots[i].Version = bits[1]
			}
		}
		if command == "tournament" && len(exp.Grid) != 0 {
			return 2
		}
		jobs, e = exp.Schedule()
		if e != nil {
			return 2
		}
		pairing = exp.PairingID
		*players = exp.Population
	} else {
		if *players != 3 && *players != 4 {
			return 2
		}
		ps := strings.Split(*policies, ",")
		if len(ps) != *players {
			return 2
		}
		for _, p := range ps {
			if !bots.GameplayOrderV2(p) && !contains([]string{"legal-random@v1", "heuristic@v1", "economic@v1", "pressure@v1", "gameplay-random@v1", "gameplay-random@v2", "opportunity@v1", "opportunity-no-retain@v1", "opportunity-no-finance@v1", "bargaining@v1"}, p) {
				return 2
			}
		}
		n := 1
		if command == "batch" {
			n = *games
		}
		for i := 0; i < n; i++ {
			seats := []string{}
			pp := []Policy{}
			for k, p := range ps {
				id := fmt.Sprintf("p%d", k)
				seats = append(seats, id)
				bits := strings.Split(p, "@")
				pp = append(pp, Policy{ParticipantID: id, Policy: bits[0], Version: bits[1], Information: "seat-projection"})
			}
			jobs = append(jobs, Job{Index: i, Population: *players, Block: fmt.Sprintf("block-%06d", i), Rotation: Rotation{ID: "r0", Seats: seats}, Treatment: "baseline", Games: 1, Policies: pp})
		}
	}
	if _, e = randomstream.AnalysisSeed(*seed, *players, pairing); e != nil {
		return 2
	}
	var sc *Scenario
	var scRaw []byte
	if *scenario != "" {
		scRaw, e = readArtifact(context.Background(), *scenario, (32<<20)+65536)
		if e != nil {
			return 6
		}
		digitLimit := 256
		if len(expRaw) > 0 {
			digitLimit = exp.Budgets.MaxRationalDigits
		}
		if e = checkRationalDigitsLimit(scRaw, digitLimit, (32<<20)+65536); e != nil {
			return 2
		}
		s, e := loadScenario(scRaw, *players)
		if e != nil {
			return 2
		}
		sc = &s
	}
	if sc != nil && sc.Kind == "gameplay-checkpoint" {
		if command != "run" || resume == nil {
			return 2
		}
		jobs = []Job{resume.Job}
		pairing = resume.Pairing
	}
	limits := Budgets{MaxTransitions: *budget, MaxSettlementSteps: *budget, MaxBotOperations: 100000, MaxOutputBytes: 1073741824, MaxInputBytes: canonical.MaxBytes, MaxRationalDigits: 256, WallSeconds: 3600}
	if command == "tournament" || command == "sweep" {
		limits = exp.Budgets
	}
	scenarioLimit := limits.MaxInputBytes
	if sc != nil && sc.Kind == "gameplay-checkpoint" && sc.Gameplay != nil && sc.Gameplay.ArtifactProfile == GameplayLargeArtifactProfile {
		scenarioLimit = (32 << 20) + 65536
	}
	if len(raw) > limits.MaxInputBytes || len(scRaw) > scenarioLimit || len(expRaw) > limits.MaxInputBytes {
		return 2
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(limits.WallSeconds)*time.Second)
	defer cancel()
	rules["match_games"] = json.Number("1")
	if resume != nil {
		rules["match_games"] = json.Number(fmt.Sprint(resume.Job.Games))
	}
	if command == "tournament" || command == "sweep" {
		rules["match_games"] = json.Number(fmt.Sprint(exp.GamesPerMatch))
	}
	effective, e := canonical.Marshal(rules)
	if e != nil {
		return 2
	}
	parentFiles := map[string][]byte{}
	if sc != nil && (sc.Kind == "settlement-checkpoint" || sc.Kind == "gameplay-checkpoint") {
		parentFiles, e = verifyContinuationParent(ctx, *sc, hash(rules))
		if e != nil {
			return artifactExit(e)
		}
	}
	if resume != nil {
		var frozen Budgets
		if err := decodeRequired(parentFiles["parent-execution-limits.json"], &frozen); err != nil {
			return 2
		}
		frozen.MaxTransitions = limits.MaxTransitions
		frozen.WallSeconds = limits.WallSeconds
		limits = frozen
	}
	m := manifestBase(args, command, *seed, hash(rules), *workers, len(jobs), ruleSource)
	if sc == nil || sc.Kind == "gameplay-checkpoint" {
		m.Provenance.Menu = game.GameplayMenuVersion
		m.Coverage = []string{"accepted-gameplay"}
		m.Exclusions = []string{"exhaustive-conformance-not-implied", "human-play-not-evaluated"}
	}
	m.ReviewHold = *hold
	if sc != nil {
		m.Parent = sc.Parent
	}
	if len(expRaw) > 0 {
		m.ReviewHold = m.ReviewHold || exp.Retention.ReviewHold
	}
	files := map[string][]byte{"effective-rules.json": effective, "source-rules.json": defaults, "selected-rules.json": raw, "rule-source.md": ruleSource}
	for k, v := range parentFiles {
		files[k] = v
	}
	if sc != nil {
		files["scenario.json"] = scRaw
	}
	if len(expRaw) > 0 {
		files["experiment-source.json"] = expRaw
		b, _ := canonical.Marshal(exp)
		files["effective-experiment.json"] = b
	}
	execution := limits
	execution.Workers = 0
	files["execution-limits.json"] = mustCanonical(execution)
	schedule, e := canonical.Marshal(jobs)
	if e != nil {
		return 2
	}
	files["schedule.json"] = schedule
	for _, name := range []string{"rules", "experiment", "scenario", "manifest", "outcome", "trace", "checkpoint", "observation", "metrics", "gameplay", "gameplay-public", "finance-private"} {
		b, e := readInput(filepath.Join(filepath.Dir(defaultsPath), "../schemas", name+".schema.json"))
		if e != nil {
			return 6
		}
		files["schemas/"+name+".schema.json"] = b
	}
	results := make([]runResult, len(jobs))
	for i, j := range jobs {
		results[i].Outcome = Outcome{PlannedGameSlots: j.Games, NotStartedGameSlots: j.Games, Schema: "cgms-outcome-v1", Index: j.Index, Population: j.Population, Block: j.Block, Rotation: j.Rotation.ID, Seats: j.Rotation.Seats, Treatment: j.Treatment, Status: "not-started", Reason: "not dispatched", ExitCode: 130, Observer: "public", Privacy: "public"}
	}
	fixedBytes := 65536
	for _, b := range files {
		fixedBytes += len(b)
	}
	fixedBytes += len(jobs) * 16384
	if limits.MaxOutputBytes <= fixedBytes {
		return 2
	}
	limits.MaxOutputBytes -= fixedBytes
	// Reserve two copies of each initial checkpoint before dispatch. Budgets too
	// small for mandatory recovery artifacts are invalid, rather than silently
	// discarding a completed unit after execution.
	for _, j := range jobs {
		var initial any
		if sc == nil {
			initial, _ = game.NewState(j.Population, fmt.Sprintf("match-%06d/game-0/instance-0", j.Index))
		} else if sc.Gameplay != nil {
			initial = sc.Gameplay.Initial
		} else if sc.Checkpoint != nil {
			initial = sc.Checkpoint
		} else if sc.Ledger != nil {
			initial, e = game.BeginSettlement(*sc.Ledger, sc.Receipts)
			if e != nil {
				return 2
			}
		} else {
			initial = sc.State
		}
		b, e := canonical.Marshal(initial)
		if e != nil || 2*len(b)+65536 > limits.MaxOutputBytes/len(jobs) {
			return 2
		}
	}
	if e = reserveOutput(*out); e != nil {
		if os.IsExist(e) {
			return 2
		}
		return 6
	}
	finishMetrics := startMetrics()
	queue := make(chan Job)
	returned := make(chan runResult, *workers)
	var wg sync.WaitGroup
	for w := 0; w < *workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range queue {
				jobLimits := limits
				jobLimits.MaxOutputBytes = limits.MaxOutputBytes / len(jobs)
				settings, e := ResolveJob(j, jobLimits)
				r := runScenarioLimited(ctx, j, sc, *seed, pairing, settings.Budgets)
				if e != nil {
					r.Outcome.ExitCode = 2
					r.Outcome.Status = "invalid-input"
				}
				returned <- r
			}
		}()
	}
	go func() {
		defer close(queue)
		for _, j := range jobs {
			select {
			case <-ctx.Done():
				return
			case queue <- j:
			}
		}
	}()
	go func() { wg.Wait(); close(returned) }()
	jobSlots := map[int]int{}
	for slot, job := range jobs {
		jobSlots[job.Index] = slot
	}
	for r := range returned {
		results[jobSlots[r.Outcome.Index]] = r
	}
	metrics := finishMetrics(results)
	files["metrics.json"] = mustCanonical(metrics)
	code := 0
	outcomes := []Outcome{}
	for i, r := range results {
		o := r.Outcome
		if o.Status == "not-started" || r.Initial == nil {
			m.NotStarted++
			o.Status = "not-started"
			o.NotStartedGameSlots = o.PlannedGameSlots
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				o.ExitCode = 5
				o.Reason = "wall watchdog before dispatch"
			}
		} else {
			m.Started++
			prefix := fmt.Sprintf("games/%06d/", jobs[i].Index)
			b, _ := canonical.Marshal(r.Initial)
			files[prefix+"initial.json"] = b
			b, _ = canonical.Marshal(r.Checkpoint)
			files[prefix+"checkpoint.json"] = b
			if r.Gameplay != nil {
				files[prefix+"gameplay.json"] = mustCanonical(r.Gameplay)
				if r.PublicStats != nil {
					files[prefix+"gameplay-public.json"] = mustCanonical(r.PublicStats)
				}
				if r.PrivateFinance != nil {
					files[prefix+"finance-private.json"] = mustCanonical(r.PrivateFinance)
				}
			}
			b, _ = traceBytes(r.Trace)
			files[prefix+"trace.jsonl"] = b
			files[prefix+"report.md"] = []byte(gameReport(o) + publicNarrative(r.Trace))
			if r.Gameplay != nil {
				if r.PublicStats != nil {
					files[prefix+"report.md"] = []byte(gameReport(o) + RenderGameplayPublicReport(*r.PublicStats))
				}
			}
			if r.Derived != "" {
				files[prefix+"derived-seed.txt"] = []byte(r.Derived)
			}
		}
		if o.ExitCode == 130 {
			m.Canceled++
		}
		m.Finalized += o.FinalizedGames
		code = aggregate(code, o.ExitCode)
		outcomes = append(outcomes, o)
	}
	if command != "run" {
		for _, o := range outcomes {
			if !o.GameComplete && o.ExitCode == 0 {
				code = aggregate(code, 3)
			}
		}
	}
	m.ExitCode = code
	m.Status = status(code)
	b, _ := canonical.Marshal(outcomes)
	files["outcomes.json"] = b
	extra := ""
	if len(expRaw) > 0 {
		ef, er, e := experimentReports(exp, jobs, outcomes)
		if e != nil {
			return 4
		}
		for k, v := range ef {
			files[k] = v
		}
		extra = er
	}
	files["report.md"] = []byte(report(command, outcomes, jobs) + policySummary(jobs) + metricsReport(metrics) + extra)
	if e = publish(*out, m, files); e != nil {
		fmt.Fprintln(stderr, "artifact I/O failure")
		return 6
	}
	fmt.Fprintln(stdout, filepath.Join(*out, "manifest.json"))
	return code
}
func contains(a []string, s string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}
func gameReport(o Outcome) string {
	return fmt.Sprintf("# Match/scenario %d\n\nObserver: public.\n\nStatus: %s; %s. %s\n\nFinalized games: %d / %d. Board-ended: %d; started: %d; finalized match: %t.\n\nExact cumulative scores: %v; tied ranks: %v; current available points: %v; debts: %v.\n\nRecorded steps: %d; redeals: %d. Targeted conformance, autonomous exercise, policy competence and balance are separate evidence categories.\n", o.Index, o.Status, o.ScenarioStatus, o.Reason, o.FinalizedGames, o.PlannedGameSlots, o.BoardEndedGames, o.StartedGames, o.MatchComplete, o.Scores, o.MatchRanks, o.Cash, o.Debts, o.Steps, o.Redeals)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func report(command string, out []Outcome, jobs []Job) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# CGMS %s report\n\nObserver: public. Private reproduction commands, seeds and traces remain in restricted artifacts. Accepted-rule results and targeted scenario fixtures retain distinct completion statuses. Complete-game execution alone does not establish full rule conformance, stronger play or balance.\n\n", command)
	final, slots, later := 0, 0, 0
	for _, o := range out {
		slots += o.PlannedGameSlots
		later += o.NotStartedGameSlots
		final += o.FinalizedGames
	}
	fmt.Fprintf(&b, "Planned game slots: %d; Later slots not started: %d.\n\nFinalized games: %d / %d. Completion-only estimates are conditional and potentially selection-biased.\n\n", slots, later, final, slots)
	for _, pop := range []int{3, 4} {
		planned, started, ended, done, matches := 0, 0, 0, 0, 0
		causes := map[string]int{}
		for _, o := range out {
			if o.Population != pop {
				continue
			}
			planned += o.PlannedGameSlots
			started += o.StartedGames
			ended += o.BoardEndedGames
			done += o.FinalizedGames
			matches += boolInt(o.MatchComplete)
			if o.Status != "not-started" && o.StartedGames == 0 {
				started++
			}
			if !o.MatchComplete {
				causes[o.Status+": "+o.Reason]++
			}
		}
		fmt.Fprintf(&b, "## %d-player population\n\nPlanned: %d; started: %d; board-ended: %d; finalized: %d; finalized matches: %d; incomplete or never started: %d.\n\n", pop, planned, started, ended, done, matches, planned-done)
		for _, k := range sortedKeys(causes) {
			fmt.Fprintf(&b, "- %s: %d match/scenario units\n", k, causes[k])
		}
	}
	b.WriteString("\n## Seat schedules and results\n\n")
	for _, o := range out {
		fmt.Fprintf(&b, "- [%d](games/%06d/report.md): %dp, block %s, rotation %s, seats %v, %s, %s, %s. Cumulative score %v; tied ranks %v.\n", o.Index, o.Index, o.Population, o.Block, o.Rotation, o.Seats, o.Treatment, o.Status, o.ScenarioStatus, o.Scores, o.MatchRanks)
	}
	b.WriteString("\nCyclic schedules balance seats only, not every opponent order. Declaration outcomes, game score leaders and cumulative match ranks are different metrics. Interpret policy or balance effects only under the frozen experiment and its reported uncertainty.\n")
	return b.String()
}

func findDefaults(selected string) (string, error) {
	cwd, e := os.Getwd()
	if e != nil {
		return "", e
	}
	abs, e := filepath.Abs(selected)
	if e != nil {
		return "", e
	}
	for _, start := range []string{cwd, filepath.Dir(abs)} {
		for p := start; ; p = filepath.Dir(p) {
			candidate := filepath.Join(p, "config/rules/game-rules.json")
			if info, e := os.Stat(candidate); e == nil && info.Mode().IsRegular() {
				return candidate, nil
			}
			if p == filepath.Dir(p) {
				break
			}
		}
	}
	return "", fmt.Errorf("canonical defaults unavailable")
}

func publicNarrative(ts []Trace) string {
	var b strings.Builder
	b.WriteString("\n## Public transition narrative\n\n")
	if len(ts) == 0 {
		b.WriteString("No transition committed.\n")
	}
	for _, t := range ts {
		switch t.Kind {
		case "settlement":
			fmt.Fprintf(&b, "- Step %d: exact FIFO settlement progress committed; private audit records individual routing.\n", t.Index)
		case "deal":
			fmt.Fprintf(&b, "- Step %d: seeded deal attempt committed; hidden cards and seed omitted.\n", t.Index)
		case "command":
			if t.Command != nil {
				fmt.Fprintf(&b, "- Step %d: seat %d declared %s.\n", t.Index, t.Command.Actor, t.Command.Kind)
				if t.BotDecision != nil {
					for _, reason := range t.BotDecision.Reasons {
						fmt.Fprintf(&b, "  Reason: %s.\n", reason)
					}
				}
			}
		}
	}
	return b.String()
}
