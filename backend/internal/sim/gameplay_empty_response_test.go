package sim

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestGameplayOpeningWithNoEligibleRespondersResolvesWithoutPass(t *testing.T) {
	s, _ := game.NewState(3, "empty-response-fixture")
	s.Order = []int{1, 2, 3}
	s.Players[0].History = []game.Suit{game.Diamonds}
	s.Players[1].Confined = true
	s.Players[2].Confined = true
	var e error
	s, e = s.Move([]string{"deck-1-clubs-02"}, 1, game.Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	m, e := game.NewMatchLifecycle(s, 1)
	if e != nil {
		t.Fatal(e)
	}
	m.Game.TurnStarted = true
	m.Game, _, e = m.Game.ApplyBoardCommand(game.Command{GameID: s.GameID, ID: "opening", Actor: 1, Kind: "open-series", Suit: game.Clubs})
	if e != nil {
		t.Fatal(e)
	}
	if m.Game.Board.Pending == nil || len(m.Game.Board.Pending.Responders) != 0 {
		t.Fatal("fixture response boundary")
	}
	j := Job{Population: 3, Games: 1, Policies: []Policy{{Policy: "economic", Version: "v1"}, {Policy: "economic", Version: "v1"}, {Policy: "economic", Version: "v1"}}}
	o := GameplayOptions{Job: j, Root: "2", Pairing: "fixture", BotBudget: 1000}
	streams, _ := gameplayStreams(o, 0)
	st := GameplayStep{}
	if e = planGameplayStep(m, streams, nil, o, &st, nil); e != nil {
		t.Fatal(e)
	}
	if st.Decision != nil || st.FinanceDecision != nil || st.Command == nil || st.Command.Kind == "pass" {
		t.Fatal("fabricated policy response")
	}
	if e = verifyGameplayChance(m, st, streams); e != nil {
		t.Fatal("recorded deterministic continuation replay", e)
	}
	next, _, e := applyGameplayStep(m, st)
	if e != nil {
		t.Fatal(e)
	}
	card, _ := next.Game.Board.Card("deck-1-clubs-02")
	if next.Game.Board.Pending != nil || card.Zone != game.Series {
		t.Fatal("opening not resolved")
	}
	if m.Game.Board.Pending == nil {
		t.Fatal("prior state mutated")
	}
}
