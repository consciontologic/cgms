package game

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type ruleBoundaryDigest struct {
	Boundary string `json:"boundary"`
	State    string `json:"state"`
	Events   string `json:"events"`
}

func TestRuleBoundaryGoldens(t *testing.T) {
	traces := map[string][]ruleBoundaryDigest{}
	record := func(name, label string, s State, events []Event) {
		t.Helper()
		if err := s.Validate(); err != nil {
			t.Fatal(err)
		}
		traces[name] = append(traces[name], ruleBoundaryDigest{label, Digest(s), Digest(events)})
	}
	step := func(name string, s State, c Command) State {
		t.Helper()
		before := Digest(s)
		n, events, err := Apply(s, c)
		if err != nil {
			t.Fatal(c.Kind, err)
		}
		if Digest(s) != before {
			t.Fatal("mutated prior state")
		}
		record(name, c.Kind, n, events)
		return n
	}
	s, _ := NewState(3, "golden-purchase")
	payment := "deck-1-spades-10"
	s, _ = s.Move([]string{payment}, 1, Hand, false)
	record("purchase", "initial", s, nil)
	s = step("purchase", s, Command{GameID: s.GameID, ID: "buy", Kind: "purchase", Actor: 1, Value: 1, Cards: []string{payment}})
	if c, _ := s.Card(payment); c.Zone != Hand || s.Pending.Quantity != 1 {
		t.Fatal("purchase committed payment early or resized quantity")
	}
	for _, seat := range []int{2, 3} {
		s = step("purchase", s, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "buy"})
	}
	if s.Pending.Cursor != 2 {
		t.Fatal("purchase chance boundary")
	}
	order := append(append([]string{}, s.DrawOrder...), payment)
	s = step("purchase", s, Command{GameID: s.GameID, ID: "purchase-random", Kind: "purchase-outcome", Actor: 1, WindowID: "buy", Cards: order})
	count := 0
	for _, c := range s.Cards {
		if c.Controller == 1 {
			count++
		}
	}
	if count != 1 || s.Pending != nil {
		t.Fatal("chosen one-card quantity must not become maximum affordable two")
	}

	s = compensationPosition(t)
	record("compensation", "initial", s, nil)
	var err error
	s, err = QueueCompensation(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.DrawQueue) != 2 || s.DrawQueue[0].Seat != 2 || s.DrawQueue[1].Seat != 3 {
		t.Fatal("fixed seat queue order")
	}
	record("compensation", "queue", s, nil)
	var drawn []string
	s, drawn, err = StepCompensation(s, nil)
	if err != nil || len(drawn) != 1 {
		t.Fatal("first entitlement", err)
	}
	if c, _ := s.Card(cid(Clubs, 12)); c.Controller != 2 {
		t.Fatal("initiative order overrode fixed seat")
	}
	record("compensation", "seat-two-draw", s, nil)
	s, drawn, err = StepCompensation(s, nil)
	if err != nil || len(drawn) != 0 || len(s.DrawQueue) != 0 {
		t.Fatal("exhaustion fabricated cards or future entitlement")
	}
	for _, effect := range s.Effects {
		if effect.Losses != 0 {
			t.Fatal("unconsumed entitlement")
		}
	}
	record("compensation", "seat-three-exhausted", s, nil)

	s = formationFixture(t)
	s, _ = s.Move([]string{"deck-2-diamonds-02", "deck-2-hearts-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Diamonds, Hearts}
	kings := []string{"deck-1-hearts-13", "deck-1-clubs-13", "deck-1-spades-13", "deck-1-diamonds-13"}
	s, _ = s.Move(kings, 1, Hand, false)
	s, err = OpenFormation(s, 1, "fate-golden", FormationSpec{Kind: "fate", Cards: kings})
	if err != nil {
		t.Fatal(err)
	}
	s, _ = s.Move([]string{"deck-1-hearts-01"}, 2, ConcealedAce, false)
	s, err = s.ActivateEffect(AceEffect{ID: "comp", CardID: "deck-1-hearts-01", Kind: "compensation", Source: 2, Target: 2, Custodian: 2, Expiry: "game-end", Losses: 3})
	if err != nil {
		t.Fatal(err)
	}
	record("fate", "initial", s, nil)
	s = step("fate", s, Command{GameID: s.GameID, ID: "fate", Kind: "fate", Actor: 1, FormationID: "fate-golden", TargetSeat: 2, Cards: kings[:1]})
	for _, seat := range []int{2, 3} {
		s = step("fate", s, Command{GameID: s.GameID, ID: string(rune('f' + seat)), Kind: "pass", Actor: seat, WindowID: "fate"})
	}
	order = append(append([]string{}, s.DrawOrder...), kings[0], "deck-2-hearts-02", "deck-2-diamonds-02")
	s = step("fate", s, Command{GameID: s.GameID, ID: "fate-random", Kind: "ability-outcome", Actor: 1, WindowID: "fate", Cards: order})
	if c, _ := s.Card("deck-1-hearts-01"); c.Zone != ActiveAce || s.Effects[0].Losses != 3 {
		t.Fatal("Fate changed persistent compensation")
	}
	count = 0
	for _, c := range s.Cards {
		if c.Controller == 2 && c.Zone != ActiveAce {
			count++
		}
	}
	if count != 2 || FormationFunctioning(s, "fate-golden") {
		t.Fatal("Fate exact redraw or resolution king cost")
	}
	path := "../../testdata/rule-boundaries-v1.json"
	if os.Getenv("CGMS_UPDATE_RULE_BOUNDARIES") == "1" {
		data, err := json.MarshalIndent(traces, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var expected map[string][]ruleBoundaryDigest
	if err = json.Unmarshal(data, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, traces) {
		t.Fatal("rule boundary digests differ from reviewed semantic fixtures")
	}
}
