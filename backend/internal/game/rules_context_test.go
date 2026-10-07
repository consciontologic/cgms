package game

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

// Decode the public contract independently so the regression first fails for
// missing behavior, rather than for a not-yet-declared Go field.
type pendingContextView struct {
	Type          string   `json:"type"`
	Actor         int      `json:"actor"`
	TargetSeat    int      `json:"target_seat"`
	ResponseTypes []string `json:"response_types"`
}

func pendingView(t *testing.T, s State, seat int) *pendingContextView {
	t.Helper()
	c, err := ProjectRulesContext(s, seat)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	var view struct {
		Pending *pendingContextView `json:"pending"`
	}
	if err = json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	return view.Pending
}

func responseContextFixture(t *testing.T) State {
	t.Helper()
	s := combatFixture(t)
	var err error
	s, err = s.Move([]string{"deck-1-spades-05"}, 2, Series, true)
	if err != nil {
		t.Fatal(err)
	}
	s.Pending = &PendingAction{ID: "window", Actor: 1, Defender: 2, Responders: []int{2, 3}, Clubs: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}}
	return s
}

func TestObserveKidnapperPublicOutcomeBoundary(t *testing.T) {
	s := responseContextFixture(t)
	// Public choices have no hidden-card dependency until response traversal has
	// finished and the engine has reached its exposed-choice outcome boundary.
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-1-hearts-12" || s.Cards[i].Card.ID == "deck-1-clubs-13" {
			s.Cards[i].Controller, s.Cards[i].Zone = 2, Unassigned
		}
	}
	s.Pending = &PendingAction{ID: "kidnap", Kind: "ability", Actor: 1, Responders: []int{2, 3}, Cursor: 2, Ability: &AbilityIntent{Kind: "kidnapper", TargetSeat: 2}}
	for _, seat := range []int{1, 2, 3} {
		o, err := ObserveWithoutMenu(s, seat)
		if err != nil {
			t.Fatal(err)
		}
		if o.RequiredActor != 1 || o.DecisionKind != "kidnapper-outcome" || o.DecisionID != "" {
			t.Fatalf("seat %d missing outcome boundary: %+v", seat, o)
		}
		if seat == 1 {
			if len(o.Choices) != 2 {
				t.Fatalf("missing public alternatives: %+v", o.Choices)
			}
			for _, choice := range o.Choices {
				c, ok := s.Card(choice[0])
				if !ok || !publicZone(c.Zone) || c.Controller != 2 {
					t.Fatal("private or unrelated choice")
				}
			}
		} else if len(o.Choices) != 0 {
			t.Fatal("other seat received acting seat choices")
		}
	}
	for _, change := range []func(*State){
		func(s *State) { s.Pending.Cursor = 0 },
		func(s *State) { s.Pending.Decision = &EffectDecision{ID: "other", Kind: "dexter", Actor: 2} },
		func(s *State) {
			for i := range s.Cards {
				if s.Cards[i].Card.ID == "deck-1-spades-12" {
					s.Cards[i].Controller, s.Cards[i].Zone = 2, Hand
				}
			}
		},
	} {
		n := s.Clone()
		change(&n)
		o, err := ObserveWithoutMenu(n, 1)
		if err != nil || o.DecisionKind == "kidnapper-outcome" || len(o.Choices) != 0 {
			t.Fatal("exposed choices leaked before outcome boundary", err)
		}
	}
	s.Pending.Ability.Targets = []string{"deck-1-hearts-12"}
	o, err := ObserveWithoutMenu(s, 1)
	if err != nil || !reflect.DeepEqual(o.Choices, [][]string{{"deck-1-hearts-12"}}) {
		t.Fatal("declared target must not be replaceable", err)
	}
	o.Choices[0][0] = "changed"
	if s.Pending.Ability.Targets[0] != "deck-1-hearts-12" {
		t.Fatal("outcome choices alias state")
	}
}

func TestRulesContextPendingResponseCategories(t *testing.T) {
	confinement := []string{"confinement"}
	attack := []string{"confinement", "compensation-response", "inflation"}
	numbers := append(append([]string{}, attack...), "numerical-defense", "advancement-defense")
	for _, tc := range []struct {
		name, kind, ability, opening, typ string
		bomb                              bool
		responses                         []string
		target                            int
	}{
		{name: "number Hearts", typ: "attack", responses: numbers, target: 2},
		{name: "Bomb", typ: "bomb-attack", bomb: true, responses: attack, target: 2},
		{name: "opening", kind: "opening", opening: "open-series", typ: "open-series", responses: confinement},
		{name: "purchase", kind: "purchase", typ: "purchase", responses: confinement},
		{name: "transfer", kind: "transfer", typ: "transfer", responses: confinement},
		{name: "return", kind: "loan-return", typ: "loan-return", responses: confinement},
		{name: "Ponzi", kind: "ponzi", typ: "ponzi", responses: attack, target: 2},
		{name: "Justice", kind: "justice", typ: "justice", responses: attack},
		{name: "Fate", kind: "ability", ability: "fate", typ: "fate", responses: attack, target: 2},
		{name: "Kidnapper", kind: "ability", ability: "kidnapper", typ: "kidnapper", responses: attack, target: 2},
		{name: "Inflation", kind: "ability", ability: "main-inflation", typ: "main-inflation", responses: attack, target: 2},
		{name: "Code", kind: "ability", ability: "code", typ: "code", responses: confinement},
		{name: "Compensation", kind: "ability", ability: "compensation", typ: "compensation", responses: confinement},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := responseContextFixture(t)
			s.Pending.Kind = tc.kind
			s.Pending.Bomb = tc.bomb
			if tc.kind != "" {
				s.Pending.Clubs = nil
				s.Pending.Targets = nil
			}
			if tc.ability != "" {
				s.Pending.Ability = &AbilityIntent{Kind: tc.ability, TargetSeat: 2}
			}
			if tc.opening != "" {
				s.Pending.Opening = &OpeningIntent{Kind: tc.opening}
			}
			v := pendingView(t, s, 2)
			if v == nil || v.Type != tc.typ || v.Actor != 1 || v.TargetSeat != tc.target || !reflect.DeepEqual(v.ResponseTypes, tc.responses) {
				t.Fatalf("wrong public response context: %+v", v)
			}
			for _, seat := range []int{1, 3} {
				v := pendingView(t, s, seat)
				if v == nil || v.Type != tc.typ || len(v.ResponseTypes) != 0 {
					t.Fatalf("seat %d received response categories out of order: %+v", seat, v)
				}
			}
		})
	}
}

func TestRulesContextPendingSupportTimingAndPublicTargets(t *testing.T) {
	s := responseContextFixture(t)
	if pendingView(t, combatFixture(t), 2) != nil {
		t.Fatal("idle pending descriptor")
	}
	for _, change := range []func(*State){
		func(s *State) { s.Players[1].AceRound = s.Round },
		func(s *State) { s.Players[1].Confined = true },
		func(s *State) { s.Players[1].Departed = true },
		func(s *State) { s.Pending.Cursor = 1 },
		func(s *State) { s.Pending.Cursor = -1 },
		func(s *State) { s.Pending.Decision = &EffectDecision{} },
		func(s *State) { s.Phase = "finished" },
		func(s *State) {
			for i := range s.Cards {
				if s.Cards[i].Card.ID == "deck-1-spades-05" {
					s.Cards[i].Zone = Hand
				}
			}
		},
	} {
		n := s.Clone()
		change(&n)
		v := pendingView(t, n, 2)
		if v == nil || len(v.ResponseTypes) != 0 {
			t.Fatalf("inactive/unsupported categories: %+v", v)
		}
	}
	s.Players[1].CompensationUsed = true
	s.Effects = []AceEffect{{Kind: "inflation", Target: 1}}
	v := pendingView(t, s, 2)
	if v == nil || !reflect.DeepEqual(v.ResponseTypes, []string{"confinement", "numerical-defense", "advancement-defense"}) {
		t.Fatalf("used effects offered: %+v", v)
	}
	s.Pending.Kind = "ability"
	s.Pending.Clubs = nil
	s.Pending.Ability = &AbilityIntent{Kind: "dexter-assassination", Targets: []string{"deck-1-spades-05"}}
	v = pendingView(t, s, 2)
	if v == nil || v.TargetSeat != 2 {
		t.Fatalf("public Dexter target missing: %+v", v)
	}
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-1-spades-05" {
			s.Cards[i].Zone = Hand
		}
	}
	v = pendingView(t, s, 1)
	if v == nil || v.TargetSeat != 0 {
		t.Fatalf("private Dexter target leaked: %+v", v)
	}
}

func TestRulesContextPendingNegotiatorAndNoPrivateDetails(t *testing.T) {
	s := responseContextFixture(t)
	var err error
	s, err = s.Move([]string{"deck-1-spades-11"}, 2, Attachment, true)
	if err != nil {
		t.Fatal(err)
	}
	s.Players[1].AceRound = s.Round
	v := pendingView(t, s, 2)
	if v == nil || !reflect.DeepEqual(v.ResponseTypes, []string{"negotiate"}) {
		t.Fatalf("Negotiator must be independent of Ace quota: %+v", v)
	}
	s.Players[1].Quotas["negotiator"] = s.Round
	if len(pendingView(t, s, 2).ResponseTypes) != 0 {
		t.Fatal("used Negotiator")
	}
	s.Pending.Kind = "purchase"
	s.Pending.Clubs = nil
	s.Pending.Targets = nil
	a := pendingView(t, s, 2)
	n := s.Clone()
	n.Pending.Quantity = 99
	n.Pending.ProposalID = "private-offer"
	n.Pending.JusticeQueen = "private-queen"
	n.Pending.Aces = []string{"deck-1-hearts-01"}
	n.Pending.Ability = &AbilityIntent{Kind: "private-ability", Cards: []string{"deck-1-hearts-02"}, AceID: "secret", PurchasePrice: 99}
	for i := range n.Cards {
		if n.Cards[i].Card.ID == "deck-1-hearts-02" {
			n.Cards[i].Controller, n.Cards[i].Zone = 3, Hand
		}
	}
	for _, seat := range []int{1, 2, 3} {
		before, after := pendingView(t, s, seat), pendingView(t, n, seat)
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("seat %d context depends on private terms/cards", seat)
		}
		context, _ := ProjectRulesContext(n, seat)
		raw, _ := json.Marshal(context)
		if bytes.Contains(raw, []byte("deck-")) || bytes.Contains(raw, []byte("private")) || bytes.Contains(raw, []byte("secret")) {
			t.Fatalf("private pending data: %s", raw)
		}
	}
	if a == nil || a.TargetSeat != 0 {
		t.Fatalf("purchase must not reuse an unrelated defender field: %+v", a)
	}
}

func TestRulesContextInterpretsOwnQuotaPeriods(t *testing.T) {
	s := combatFixture(t)
	s.Round = 4
	s.Order = []int{1, 2, 3}
	s.Active = 1
	s.Players[1].Quotas = map[string]int{"ponzi": 4, "baron": 3, "ringleader": 1}
	c, _ := ProjectRulesContext(s, 2)
	if !c.QuotaUsed["ponzi"] || c.QuotaUsed["baron"] || !c.QuotaUsed["ringleader"] || c.QuotaUsed["justice"] {
		t.Fatal("own cycle/round/lifetime quota status", c.QuotaUsed)
	}
	s.Active = 2
	c, _ = ProjectRulesContext(s, 2)
	if c.QuotaUsed["ponzi"] || !c.QuotaUsed["ringleader"] {
		t.Fatal("own turn begins fresh cycle, lifetime remains consumed", c.QuotaUsed)
	}
	other, _ := ProjectRulesContext(s, 1)
	for _, used := range other.QuotaUsed {
		if used {
			t.Fatal("opponent quota exposed")
		}
	}
	s.Players[1].AceRound = s.Round
	c, _ = ProjectRulesContext(s, 2)
	if !c.QuotaUsed["compensation"] || !c.QuotaUsed["main-inflation"] {
		t.Fatal("shared Ace quota omitted")
	}
	s.Round++
	s.Players[1].CompensationUsed = true
	c, _ = ProjectRulesContext(s, 2)
	if !c.QuotaUsed["compensation"] || c.QuotaUsed["main-inflation"] {
		t.Fatal("Compensation lifetime and Ace round quotas conflated")
	}
}

func TestRulesContextOwnQuotasAndNoHiddenCardDependence(t *testing.T) {
	s := combatFixture(t)
	s.Players[0].OpeningUsed = true
	s.Players[0].AceRound = 1
	s.Players[0].Quotas["baron"] = 1
	a, e := ProjectRulesContext(s, 1)
	if e != nil {
		t.Fatal(e)
	}
	if !a.OpeningUsed || !a.AceUsedThisRound || a.Quotas["baron"] != 1 || a.PurchaseUnitPrice != 5 {
		t.Fatal("own rule context")
	}
	b, e := ProjectRulesContext(s, 2)
	if e != nil {
		t.Fatal(e)
	}
	if b.OpeningUsed || b.AceUsedThisRound || len(b.Quotas) != 0 {
		t.Fatal("other seat quota leaked")
	}
	changed := s.Clone()
	changed, _ = changed.Move([]string{"deck-1-spades-03"}, 2, Hand, true)
	c, _ := ProjectRulesContext(changed, 1)
	if Digest(a) != Digest(c) {
		t.Fatal("opponent hidden holdings affect context")
	}
	a.Quotas["baron"] = 99
	if s.Players[0].Quotas["baron"] != 1 {
		t.Fatal("alias")
	}
	if _, e := ProjectRulesContext(s, 0); e == nil {
		t.Fatal("invalid observer")
	}
}

func TestRulesContextPendingCombatUsesEngineAmounts(t *testing.T) {
	s := combatFixture(t)
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if e != nil {
		t.Fatal(e)
	}
	for _, seat := range []int{1, 2, 3} {
		c, e := ProjectRulesContext(n, seat)
		if e != nil {
			t.Fatal(e)
		}
		if c.Combat == nil || c.Combat.Attack.Cmp(IntAmount(8)) != 0 || c.Combat.Defense.Cmp(IntAmount(8)) != 0 || c.Combat.Comparison != "equal" {
			t.Fatal("combat calculation")
		}
	}
}

func TestRulesContextPendingCombatExposesOnlyFrozenPublicOperands(t *testing.T) {
	s := combatFixture(t)
	s, _, err := Apply(s, Command{GameID: s.GameID, ID: "public-attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, seat := range []int{1, 2, 3} {
		context, err := ProjectRulesContext(s, seat)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(context.Combat)
		if err != nil {
			t.Fatal(err)
		}
		var combat map[string]any
		if err = json.Unmarshal(raw, &combat); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(combat["attack_cards"], []any{"deck-1-clubs-08"}) || !reflect.DeepEqual(combat["target_cards"], []any{"deck-1-hearts-08"}) {
			t.Fatalf("seat %d lacks frozen public attack/target identities: %s", seat, raw)
		}
		changed := s.Clone()
		changed, err = changed.Move([]string{"deck-1-spades-03"}, 2, Hand, true)
		if err != nil {
			t.Fatal(err)
		}
		changed.Pending.Aces = []string{"deck-1-spades-03"}
		changed.Pending.Decision = &EffectDecision{Reserved: []string{"deck-1-spades-03"}, Choices: [][]string{{"deck-1-spades-03"}}}
		other, err := ProjectRulesContext(changed, seat)
		if err != nil || Digest(context.Combat) != Digest(other.Combat) {
			t.Fatal("hidden holdings, reserved cards or choices changed public combat", err)
		}
		context.Combat.AttackCards[0] = "changed-view"
		context.Combat.TargetCards[0] = "changed-view"
		if s.Pending.Clubs[0] != "deck-1-clubs-08" || s.Pending.Targets[0] != "deck-1-hearts-08" {
			t.Fatal("projected operand slices alias the authoritative pending action")
		}
	}
	for _, hidden := range []string{"deck-1-clubs-08", "deck-1-hearts-08"} {
		changed := s.Clone()
		for i := range changed.Cards {
			if changed.Cards[i].Card.ID == hidden {
				changed.Cards[i].Zone = Hand
			}
		}
		for _, seat := range []int{1, 2, 3} {
			context, err := ProjectRulesContext(changed, seat)
			if err != nil || context.Combat != nil {
				t.Fatalf("seat %d received a concealed combat operand %s: %v", seat, hidden, err)
			}
		}
	}
}

func TestRulesContextComparisonAndHiddenOperand(t *testing.T) {
	s := combatFixture(t)
	s.Pending = &PendingAction{Actor: 1, Defender: 2, Clubs: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}, Baron: true, AttackModifier: 3, DefenseModifier: 2}
	c, _ := ProjectRulesContext(s, 3)
	if c.Combat.Comparison != "greater" || c.Combat.Attack.Cmp(IntAmount(11)) != 0 || c.Combat.Defense.Cmp(IntAmount(10)) != 0 {
		t.Fatal("Baron uses strict superior strength")
	}
	s.Pending.Bomb = true
	c, _ = ProjectRulesContext(s, 3)
	if c.Combat.Comparison != "at_least" || c.Combat.Defense.Cmp(IntAmount(5)) != 0 {
		t.Fatal("Bomb threshold")
	}
	for i := range s.Cards {
		if s.Cards[i].Card.ID == "deck-1-hearts-08" {
			s.Cards[i].Zone = Hand
		}
	}
	c, _ = ProjectRulesContext(s, 3)
	if c.Combat != nil {
		t.Fatal("concealed operand disclosed")
	}
	s.Code = &CodeSettings{Declarer: 2, PurchasePrice: 7}
	s.Effects = []AceEffect{{Kind: "inflation", Target: 1}}
	c, _ = ProjectRulesContext(s, 1)
	if c.PurchaseUnitPrice != 7 || !c.Inflation {
		t.Fatal("Code/Inflation economics")
	}
	c, _ = ProjectRulesContext(s, 2)
	if c.PurchaseUnitPrice != 5 || c.Inflation {
		t.Fatal("Code declarer exempt from own price")
	}
}
