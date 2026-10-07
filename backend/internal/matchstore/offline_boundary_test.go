package matchstore

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestLocalCapabilityUsesEngineGuardsWithoutStore(t *testing.T) {
	b, _ := game.NewState(3, "local-test")
	m, _ := game.NewMatchLifecycle(b, 1)
	e, _ := NewEnvelope(m, "rules-test")
	before := game.Digest(e)
	bad := Request{Command: &game.Command{ID: "bad", GameID: b.GameID, Actor: 2, Kind: "open-series", Suit: game.Clubs}}
	if _, err := ApplyLocalRequest(e, 1, bad); err == nil {
		t.Fatal("actor spoof accepted")
	}
	next, ok := NextServer(e, 0)
	if !ok {
		t.Fatal("missing local deal")
	}
	next.Request.ID = next.ID
	after, err := ApplyLocalAutomatic(e, next.Request)
	if err != nil {
		t.Fatal(err)
	}
	if game.Digest(e) != before {
		t.Fatal("mutated input")
	}
	if len(after.Chance) == 0 {
		t.Fatal("missing retained local chance")
	}
	if _, err = SeatProjection(after, 1); err != nil {
		t.Fatal(err)
	}
}
