package sim

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"os"
	"testing"
)

func TestGridResolvedWithoutMutation(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	x, e := LoadExperiment(b)
	if e != nil {
		t.Fatal(e)
	}
	jobs, e := x.Schedule()
	if e != nil {
		t.Fatal(e)
	}
	j := jobs[1]
	j.Parameters = map[string]int{"budget.transitions": 9, "budget.bot-operations": 17, "bot.protection": 7}
	cfg, e := ResolveJob(j, Budgets{MaxTransitions: 100, MaxBotOperations: 1000})
	if e != nil {
		t.Fatal(e)
	}
	if cfg.Budgets.MaxTransitions != 9 || cfg.Budgets.MaxBotOperations != 17 || cfg.Weights[0].Protection != 7 || j.Policies[0].Weights[2] != 2 {
		t.Fatalf("%+v", cfg)
	}
	base := jobs[0]
	base.Parameters = j.Parameters
	cfg, e = ResolveJob(base, Budgets{MaxTransitions: 100, MaxBotOperations: 1000})
	if e != nil || cfg.Weights[0].Protection != 2 {
		t.Fatal("candidate grid changed baseline")
	}
}

func TestBotStepReplayInputAndDisconnected(t *testing.T) {
	s, _ := game.NewState(3, "bot-step")
	for _, x := range []struct {
		id   string
		seat int
	}{{"deck-1-clubs-08", 1}, {"deck-1-diamonds-02", 1}, {"deck-1-hearts-08", 2}} {
		s, _ = s.Move([]string{x.id}, x.seat, game.Series, true)
	}
	s.Players[1].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	x, _ := LoadExperiment(b)
	jobs, _ := x.Schedule()
	streams := map[int]*randomstream.Stream{1: randomstream.New([32]byte{}), 2: randomstream.New([32]byte{1})}
	step, e := RunBotStep(s, jobs[0], streams, 1000)
	if e != nil {
		t.Fatal(e)
	}
	replayed, events, e := game.Apply(s, step.Decision.Command)
	if e != nil || game.Digest(replayed) != game.Digest(step.State) || game.Digest(events) != game.Digest(step.Events) || len(step.Decision.Reasons) == 0 {
		t.Fatal("bot decision not replayable")
	}
	waiting := step.State.Clone()
	waiting.Players[1].Connected = false
	before := game.Digest(waiting)
	next, e := RunBotStep(waiting, jobs[0], streams, 1000)
	if e != ErrDisconnectedWait || game.Digest(next.State) != before {
		t.Fatal("disconnected action fabricated")
	}
}
func TestUnknownBaselineGridRejected(t *testing.T) {
	b, _ := os.ReadFile("../../../sims/experiments/baseline-3p.json")
	x, _ := LoadExperiment(b)
	jobs, _ := x.Schedule()
	j := jobs[0]
	j.Parameters = map[string]int{"bot.unknown": 1}
	if _, e := ResolveJob(j, Budgets{MaxTransitions: 1, MaxBotOperations: 1}); e == nil {
		t.Fatal("unknown baseline grid accepted")
	}
}
