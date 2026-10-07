package sim

import (
	"bytes"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"strconv"
)

// Validate semantic linkage as well as the manifest's transport checksums.
func verifyUnitInputs(j Job, o Outcome, files map[string][]byte, prefix string) error {
	if o.PlannedGameSlots != j.Games || o.NotStartedGameSlots != j.Games-1 && o.Status != "not-started" {
		return fmt.Errorf("match slot denominator")
	}
	if o.Index != j.Index || o.Population != j.Population || o.Block != j.Block || o.Rotation != j.Rotation.ID || hash(o.Seats) != hash(j.Rotation.Seats) || o.Treatment != j.Treatment || o.Observer != "public" || o.Privacy != "public" {
		return fmt.Errorf("outcome schedule/projection mismatch")
	}
	if o.GameComplete || o.MatchComplete {
		return fmt.Errorf("unsupported finalized outcome")
	}
	if o.Status == "not-started" {
		if _, ok := files[prefix+"initial.json"]; ok {
			return fmt.Errorf("not-started unit has committed state")
		}
		return nil
	}
	initial := files[prefix+"initial.json"]
	if raw := files["scenario.json"]; raw != nil {
		sc, e := loadScenario(raw, j.Population)
		if e != nil {
			return e
		}
		if o.ExitCode == 0 && sc.Kind == "transition-sequence" {
			if sc.RunBots || o.Steps != len(sc.Commands) || o.ScenarioStatus != "scenario-complete" {
				return fmt.Errorf("premature scenario completion")
			}
		}
		var expected any
		switch sc.Kind {
		case "accounting-only":
			c, e := game.BeginSettlement(*sc.Ledger, sc.Receipts)
			if e != nil {
				return e
			}
			expected = c
		case "settlement-checkpoint":
			expected = sc.Checkpoint
		case "transition-sequence":
			expected = sc.State
		}
		if !bytes.Equal(mustCanonical(expected), initial) {
			return fmt.Errorf("scenario initial-state provenance mismatch")
		}
	} else {
		s, e := game.NewState(j.Population, fmt.Sprintf("match-%06d/game-0/instance-0", j.Index))
		if e != nil {
			return e
		}
		if o.ExitCode == 0 {
			return fmt.Errorf("unsupported complete-game claim")
		}
		if !bytes.Equal(mustCanonical(s), initial) {
			return fmt.Errorf("generated initial-state provenance mismatch")
		}
	}
	var c game.SettlementCursor
	if e := decodeRequired(files[prefix+"checkpoint.json"], &c); e == nil {
		if o.GameID != c.Ledger.GameID {
			return fmt.Errorf("outcome game identity")
		}
		if o.ExitCode == 0 {
			if !c.Done || o.ScenarioStatus != "scenario-complete" {
				return fmt.Errorf("premature accounting completion")
			}
			if hash(o.Scores) != hash(c.Ledger.Scores) || hash(o.Cash) != hash(c.Ledger.Cash) || hash(o.Debts) != hash(c.Ledger.Debts) {
				return fmt.Errorf("outcome ledger divergence")
			}
		}
	} else {
		var s game.State
		if e := decodeRequired(files[prefix+"checkpoint.json"], &s); e != nil {
			return e
		}
		if o.GameID != s.GameID {
			return fmt.Errorf("outcome game identity")
		}
		if len(o.Scores) > 0 || len(o.Cash) > 0 || len(o.Debts) > 0 {
			return fmt.Errorf("card-only trace has financial outcome")
		}
	}
	if o.Status != status(o.ExitCode) {
		return fmt.Errorf("outcome status/exit mismatch")
	}
	return nil
}

// Verify recorded RNG consumption without invoking a policy or regenerating its
// decision. Policy determinism is a separate test, not replay adjudication.
func verifyBotRandom(m Manifest, j Job, pairing string, streams map[int]*randomstream.Stream, t Trace) error {
	if t.BotDecision == nil {
		if t.BotRandomBefore != nil || t.BotRandomAfter != nil {
			return fmt.Errorf("orphan bot randomness")
		}
		return nil
	}
	if t.Command == nil || hash(t.BotDecision.Command) != hash(*t.Command) || t.BotDecision.MenuVersion != bots.MenuVersion || t.BotRandomBefore == nil || t.BotRandomAfter == nil {
		return fmt.Errorf("bot transcript mismatch")
	}
	actor := t.Command.Actor
	if actor < 1 || actor > j.Population {
		return fmt.Errorf("bot actor")
	}
	if t.BotDecision.Policy != j.Policies[actor-1].Policy+"@"+j.Policies[actor-1].Version {
		return fmt.Errorf("bot policy version")
	}
	r := streams[actor]
	if r == nil {
		seed, e := randomstream.Derive(randomstream.Identity{Match: j.Match, Root: m.RootSeed, Population: j.Population, Pairing: pairing, Block: j.Block, Rotation: j.Rotation.ID, Kind: "bot", Seat: actor})
		if e != nil {
			return e
		}
		r = randomstream.New(seed)
		streams[actor] = r
	}
	before, e := r.Snapshot()
	if e != nil {
		return e
	}
	if hash(before) != hash(t.BotRandomBefore) {
		return fmt.Errorf("bot random before")
	}
	after, e := strconv.ParseUint(t.BotRandomAfter.Counter, 10, 64)
	if e != nil || after < r.Counter || after-r.Counter > 1000000 {
		return fmt.Errorf("bot random consumption")
	}
	for r.Counter < after {
		r.Uint64()
	}
	snapshot, e := r.Snapshot()
	if e != nil {
		return e
	}
	if hash(snapshot) != hash(t.BotRandomAfter) {
		return fmt.Errorf("bot random after")
	}
	return nil
}

func validateRecordedJob(j Job) error {
	if j.Population != 3 && j.Population != 4 || len(j.Policies) != j.Population || len(j.Rotation.Seats) != j.Population || j.Games < 1 || j.Block == "" || j.Rotation.ID == "" {
		return fmt.Errorf("recorded job shape")
	}
	seen := map[string]bool{}
	for i, p := range j.Policies {
		if e := p.Validate(); e != nil {
			return e
		}
		if p.ParticipantID != j.Rotation.Seats[i] || seen[p.ParticipantID] {
			return fmt.Errorf("recorded lineup")
		}
		seen[p.ParticipantID] = true
	}
	_, e := ResolveJob(j, Budgets{MaxTransitions: 1, MaxBotOperations: 1000})
	return e
}

func verifyScenarioTranscript(raw []byte, pop int, ts []Trace) error {
	if raw == nil {
		for _, t := range ts {
			if t.Kind != "deal" {
				return fmt.Errorf("unscripted non-deal transition")
			}
		}
		return nil
	}
	sc, e := loadScenario(raw, pop)
	if e != nil {
		return e
	}
	if sc.Kind != "transition-sequence" {
		for _, t := range ts {
			if t.Kind != "settlement" || t.BotDecision != nil {
				return fmt.Errorf("accounting transcript kind")
			}
		}
		return nil
	}
	for i, t := range ts {
		if t.Kind != "command" || t.Command == nil {
			return fmt.Errorf("scenario transition kind")
		}
		if i < len(sc.Commands) {
			if hash(t.Command) != hash(sc.Commands[i]) || t.Post != sc.ExpectedDigests[i] || t.BotDecision != nil {
				return fmt.Errorf("fixed scenario command/random/expected digest provenance")
			}
		} else {
			if !sc.RunBots || t.BotDecision == nil || len(t.Command.RandomWords) > 0 {
				return fmt.Errorf("unrecorded scenario input")
			}
		}
	}
	return nil
}

func verifySummary(m Manifest, out []Outcome) error {
	started, notStarted, canceled, finalized, code := 0, 0, 0, 0, 0
	for _, o := range out {
		if o.Status == "not-started" {
			notStarted++
			if o.NotStartedGameSlots != o.PlannedGameSlots || o.Steps != 0 {
				return fmt.Errorf("not-started progress")
			}
		} else {
			started++
		}
		if o.ExitCode == 130 {
			canceled++
		}
		finalized += o.FinalizedGames
		code = aggregate(code, o.ExitCode)
		if m.Command != "run" && o.ExitCode == 0 && !o.GameComplete {
			code = aggregate(code, 3)
		}
	}
	if m.Started != started || m.NotStarted != notStarted || m.Canceled != canceled || m.Finalized != finalized || m.Status != status(m.ExitCode) {
		return fmt.Errorf("manifest outcome denominator/status divergence")
	}
	if m.Command != "replay" && m.ExitCode != code {
		return fmt.Errorf("manifest aggregate exit divergence")
	}
	return nil
}

func compatiblePairing(a, b map[string][]byte) error {
	if checksum(a["scenario.json"]) != checksum(b["scenario.json"]) {
		return fmt.Errorf("different scenario lineage")
	}
	var aj, bj []Job
	if e := decodeRequired(a["schedule.json"], &aj); e != nil {
		return e
	}
	if e := decodeRequired(b["schedule.json"], &bj); e != nil {
		return e
	}
	if len(aj) != len(bj) {
		return fmt.Errorf("different schedule length")
	}
	for i := range aj {
		if aj[i].Match != bj[i].Match || aj[i].Games != bj[i].Games || hash(aj[i].Parameters) != hash(bj[i].Parameters) {
			return fmt.Errorf("different logical match/grid")
		}
	}
	pair := func(f map[string][]byte) (string, error) {
		if f["effective-experiment.json"] == nil {
			return "standalone", nil
		}
		x, e := LoadExperiment(f["effective-experiment.json"])
		return x.PairingID, e
	}
	ap, e := pair(a)
	if e != nil {
		return e
	}
	bp, e := pair(b)
	if e != nil {
		return e
	}
	if ap != bp {
		return fmt.Errorf("different pairing identity")
	}
	return nil
}
