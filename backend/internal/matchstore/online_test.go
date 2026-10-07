package matchstore

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestOnlineSchedulerNeverAnswersHumanWait(t *testing.T) {
	e := envelopeFixture(t)
	e.Match.Game.Board.Pending = &game.PendingAction{ID: "wait", Actor: 1, Responders: []int{2}, Cursor: 0}
	e.Match.Game.TurnStarted = true
	e.Match.Game.Board.Order = []int{1, 2, 3}
	e.Match.Game.Board.Players[0].History = []game.Suit{game.Clubs}
	if _, ok := NextServer(e, 1); ok {
		t.Fatal("scheduler answered human wait")
	}
}
