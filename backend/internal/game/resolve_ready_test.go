package game

import "testing"

func TestResolveReadyRequiresExhaustedAuthenticatedWindow(t *testing.T) {
	s, _ := NewState(3, "ready-game")
	s.Players[0].History = []Suit{Diamonds}
	s.Players[1].Confined = true
	s.Players[2].Confined = true
	s, _ = s.Move([]string{"deck-1-clubs-02"}, 1, Hand, true)
	s, _, e := Apply(s, Command{GameID: s.GameID, ID: "open", Actor: 1, Kind: "open-series", Suit: Clubs})
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(s)
	c := Command{GameID: s.GameID, ID: "resolve", Actor: 1, Kind: "resolve-ready", WindowID: "open"}
	for _, mutate := range []func(*Command){func(c *Command) { c.Actor = 2 }, func(c *Command) { c.WindowID = "stale" }, func(c *Command) { c.GameID = "old" }} {
		bad := c
		mutate(&bad)
		n, _, err := Apply(s, bad)
		if err == nil || Digest(n) != before {
			t.Fatal("invalid continuation mutated", err)
		}
	}
	n, events, e := Apply(s, c)
	if e != nil || n.Pending != nil || len(events) != 1 || events[0].Kind != "open-series" {
		t.Fatal("completion", e)
	}
	again, events, e := Apply(n, c)
	if e != nil || Digest(again) != Digest(n) || len(events) != 0 {
		t.Fatal("retry", e)
	}
	waiting := s.Clone()
	waiting.Pending.Responders = []int{2}
	if _, _, e = Apply(waiting, c); e == nil {
		t.Fatal("skipped required response")
	}
	decision := s.Clone()
	decision.Pending.Decision = &EffectDecision{ID: "required"}
	if _, _, e = Apply(decision, c); e == nil {
		t.Fatal("skipped decision")
	}
}

func TestResolveReadyDoesNotTreatDisconnectedWaitAsPass(t *testing.T) {
	s := combatFixture(t)
	s.Players[1].Connected = false
	s.Players[2].Connected = false
	s, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(s)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "completion", Kind: "resolve-ready", Actor: 1, WindowID: "attack"})
	if e == nil || Digest(n) != before {
		t.Fatal("disconnected waiting actor was skipped")
	}
}
