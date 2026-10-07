//go:build integration

package matchstore

import (
	"bytes"
	"context"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestDurableProjectionNoninterference(t *testing.T) {
	ctx := context.Background()
	var before, after, event []byte
	var rejection error
	for i, id := range []string{"left", "right"} {
		s := testStore(t)
		board, err := game.NewState(3, "privacy-game")
		if err != nil {
			t.Fatal(err)
		}
		hidden := []string{"deck-1-hearts-09", "deck-1-clubs-09"}[i]
		board, err = board.Move([]string{hidden}, 2, game.Hand, true)
		if err != nil {
			t.Fatal(err)
		}
		m, err := game.NewMatchLifecycle(board, 3)
		if err != nil {
			t.Fatal(err)
		}
		m.Game.Ledger, err = m.Game.Ledger.Charge("private-debt", 1, 2, game.IntAmount(int64(5+i*7)))
		if err != nil {
			t.Fatal(err)
		}
		m.Game.Promises, err = m.Game.Promises.Offer(game.Promise{ID: "private-promise", Payer: 1, Recipient: 2, AwardID: "private-award", Mode: "fixed", Amount: game.IntAmount(int64(2 + i*3))})
		if err != nil {
			t.Fatal(err)
		}
		env, err := NewEnvelope(m, "rules-test")
		if err != nil {
			t.Fatal(err)
		}
		if err = s.Create(ctx, id, env, []string{"a", "b", "c"}); err != nil {
			t.Fatal(err)
		}
		view, err := s.Snapshot(ctx, id, "a")
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			before = view.Projection
		} else if !bytes.Equal(before, view.Projection) {
			t.Fatal("hidden custody or unrelated financial terms changed snapshot")
		}
		// Both hidden worlds must return the same generic rejection, not a rule
		// detail identifying whether an opponent privately holds the guessed card.
		bad := Intent{ID: "guess", GameID: board.GameID, Request: Request{Command: &game.Command{ID: "guess", GameID: board.GameID, Actor: 1, Kind: "attach", Cards: []string{"deck-1-hearts-09"}}}}
		_, err = s.Submit(ctx, id, "a", bad)
		if !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid private probe", err)
		}
		if i == 0 {
			rejection = err
		} else if err.Error() != rejection.Error() {
			t.Fatal("error leaks hidden state")
		}
		op := game.LifecycleOperation{ID: "consent", GameID: board.GameID, Kind: "nullify", Actor: 1}
		res, err := s.Submit(ctx, id, "a", Intent{ID: op.ID, GameID: op.GameID, Request: Request{Operation: &op}})
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			after = res.Projection
		} else if !bytes.Equal(after, res.Projection) {
			t.Fatal("command response leaks private state")
		}
		var ownEvent []byte
		if err = s.pool.QueryRow(ctx, "SELECT projection FROM seat_events WHERE match_id=$1 AND seat=1 AND seq=1", id).Scan(&ownEvent); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			event = ownEvent
		} else if !bytes.Equal(event, ownEvent) {
			t.Fatal("seat event leaks private state")
		}
		if bytes.Contains(ownEvent, []byte("private-debt")) || bytes.Contains(ownEvent, []byte("private-promise")) {
			t.Fatal("nonparty term identity exposed")
		}
		party, err := s.Snapshot(ctx, id, "b")
		if err != nil || !bytes.Contains(party.Projection, []byte("private-debt")) || !bytes.Contains(party.Projection, []byte("private-promise")) {
			t.Fatal("party lost authorized terms", err)
		}
		if _, err = s.Snapshot(ctx, id, "outsider"); !errors.Is(err, ErrUnauthorized) {
			t.Fatal("outsider read", err)
		}
	}
}
