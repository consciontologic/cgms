package game

import "testing"

func TestNaturalFormationShapes(t *testing.T) {
	s, _ := NewState(3, "formation")
	for i := range s.Cards {
		s.Cards[i].Controller = 1
	}
	tests := []struct {
		kind string
		ids  []string
	}{
		{"fate", []string{"deck-1-hearts-13", "deck-1-clubs-13", "deck-1-spades-13", "deck-1-diamonds-13"}},
		{"baron", []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-diamonds-13"}},
		{"underground", []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-clubs-13", "deck-2-clubs-13", "deck-1-spades-13"}},
		{"great-people", []string{"deck-1-hearts-12", "deck-2-hearts-12"}},
		{"coup", []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12", "deck-2-diamonds-12"}},
		{"justice", []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}},
		{"code", []string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-clubs-12", "deck-2-clubs-12", "deck-1-spades-12"}},
		{"kidnapper", []string{"deck-1-clubs-11", "deck-2-clubs-11"}},
		{"ponzi", []string{"deck-1-clubs-11", "deck-2-clubs-11", "deck-1-spades-11", "deck-2-spades-11"}},
	}
	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			spec := FormationSpec{Kind: tt.kind, Cards: tt.ids}
			if err := ValidateFormationShape(s, 1, spec); err != nil {
				t.Fatal(err)
			}
			spec.Cards = append(append([]string{}, tt.ids...), tt.ids[0])
			if ValidateFormationShape(s, 1, spec) == nil {
				t.Fatal("duplicate physical member accepted")
			}
		})
	}
}

func formationFixture(t *testing.T) State {
	t.Helper()
	s, _ := NewState(3, "formations")
	var e error
	s, e = s.Move([]string{"deck-1-diamonds-02", "deck-1-hearts-02", "deck-1-clubs-02"}, 1, Series, false)
	if e != nil {
		t.Fatal(e)
	}
	s.Players[0].History = []Suit{Diamonds, Hearts, Clubs}
	return s
}
func TestFormationOpeningBindingAndIsolation(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-1-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	spec := FormationSpec{Kind: "people", Cards: ids, Substitute: &FormationSubstitution{Kings: ids[1:], Slot: FormationSlot{12, Hearts}}, Protection: &FormationProtection{Series: Clubs}}
	n, e := OpenFormation(s, 1, "people-1", spec)
	if e != nil {
		t.Fatal(e)
	}
	if !FormationFunctioning(n, "people-1") {
		t.Fatal("People not functioning")
	}
	if len(s.Bindings) != 0 || s.Players[0].OpeningUsed {
		t.Fatal("mutated input")
	}
	spec.Cards[0] = "tamper"
	spec.Substitute.Kings[0] = "tamper"
	spec.Protection.Series = Hearts
	if !FormationFunctioning(n, "people-1") {
		t.Fatal("spec aliases state")
	}
	n, e = TakeBackFormation(n, 1, "people-1")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = AttachRoyal(n, 1, "deck-1-clubs-13", Hearts); e == nil {
		t.Fatal("dormant bound king attached")
	}
	n.Players[0].OpeningUsed = false
	change := FormationSpec{Kind: "kidnapper", Cards: []string{"deck-1-clubs-11", "deck-1-clubs-13", "deck-1-spades-13"}, Substitute: &FormationSubstitution{Kings: []string{"deck-1-clubs-13", "deck-1-spades-13"}, Slot: FormationSlot{11, Clubs}}}
	n, _ = n.Move([]string{"deck-1-clubs-11"}, 1, Hand, false)
	if _, e = OpenFormation(n, 1, "kidnapper", change); e == nil {
		t.Fatal("fixed slot reassigned")
	}
}
func TestFormationUndergroundAndAttachments(t *testing.T) {
	s, _ := NewState(3, "underground")
	ids := []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-clubs-13", "deck-2-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	n, e := OpenFormation(s, 1, "ug", FormationSpec{Kind: "underground", Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	if !n.Players[0].Underground || !FormationFunctioning(n, "ug") {
		t.Fatal("Underground exemption missing")
	}
	n, e = TakeBackFormation(n, 1, "ug")
	if e != nil || !n.Players[0].Underground {
		t.Fatal("privilege lost")
	}
	n, _ = n.Move([]string{"deck-1-clubs-02"}, 1, Series, false)
	n, _ = n.Move([]string{"deck-1-hearts-12"}, 1, Hand, false)
	if _, e = AttachRoyal(n, 1, "deck-1-hearts-12", Clubs); e == nil {
		t.Fatal("wrong queen attachment")
	}
	n, _ = n.Move([]string{"deck-1-hearts-11"}, 1, Hand, false)
	n, e = AttachRoyal(n, 1, "deck-1-hearts-11", Clubs)
	if e != nil {
		t.Fatal(e)
	}
}
func TestFormationProtectionRestrictions(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-2-hearts-12"}
	s, _ = s.Move(ids, 1, Hand, false)
	for _, p := range []*FormationProtection{nil, {Series: Hearts}, {Series: Clubs, Formation: "x"}, {Formation: "gp"}} {
		if _, e := OpenFormation(s, 1, "gp", FormationSpec{Kind: "great-people", Cards: ids, Protection: p}); e == nil {
			t.Fatal("forbidden protection accepted")
		}
	}
	s.Players[0].OpeningUsed = true
	if _, e := OpenFormation(s, 1, "gp", FormationSpec{Kind: "great-people", Cards: ids, Protection: &FormationProtection{Series: Clubs}}); e == nil {
		t.Fatal("second opening accepted")
	}
}

func TestFormationEarnedUndergroundDoesNotConsumeSeriesOpening(t *testing.T) {
	s := formationFixture(t)
	s.Players[0].Underground = true
	ids := []string{"deck-1-clubs-11", "deck-2-clubs-11"}
	s, _ = s.Move(ids, 1, Hand, false)
	n, e := OpenFormation(s, 1, "kid", FormationSpec{Kind: "kidnapper", Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	if n.Players[0].OpeningUsed {
		t.Fatal("waived combo consumed number-series opening")
	}
}

func TestFormationExposureRemainsInPublicHistory(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-clubs-11", "deck-2-clubs-11"}
	s, _ = s.Move(ids, 1, Hand, false)
	n, e := OpenFormation(s, 1, "kid", FormationSpec{Kind: "kidnapper", Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	n, e = TakeBackFormation(n, 1, "kid")
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range ids {
		found := false
		for _, known := range n.PublicHistory {
			if known == id {
				found = true
			}
		}
		if !found {
			t.Fatal("opened formation face missing from history")
		}
	}
}

func TestRegisteredJusticeSubstitutionAndPonziActivation(t *testing.T) {
	for _, sub := range []bool{false, true} {
		s := formationFixture(t)
		ids := []string{"deck-1-hearts-12", "deck-1-clubs-12", "deck-1-spades-12", "deck-1-diamonds-12"}
		spec := FormationSpec{Kind: "justice", Cards: ids}
		if sub {
			spec.Cards = []string{ids[0], ids[1], ids[2], "deck-1-hearts-13", "deck-1-clubs-13"}
			spec.Substitute = &FormationSubstitution{Kings: spec.Cards[3:], Slot: FormationSlot{12, Diamonds}}
		}
		s, _ = s.Move(spec.Cards, 1, Hand, false)
		s, e := OpenFormation(s, 1, "arbitrary-justice-id", spec)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = justice(s, Command{ID: "justice", Actor: 1, Kind: "justice", FormationID: "arbitrary-justice-id", AceID: ids[0]}); e != nil {
			t.Fatal("registered Justice rejected", sub, e)
		}
	}
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-2-diamonds-02", "deck-2-hearts-02"}, 2, Series, false)
	s.Players[1].History = []Suit{Hearts, Diamonds}
	ids := []string{"deck-1-clubs-11", "deck-2-clubs-11", "deck-1-spades-11", "deck-2-spades-11"}
	s, _ = s.Move(ids, 1, Hand, false)
	s, e := OpenFormation(s, 1, "arbitrary-ponzi-id", FormationSpec{Kind: "ponzi", Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = declarePonzi(s, Command{ID: "ponzi", Actor: 1, Kind: "ponzi", Value: 2}); e != nil {
		t.Fatal("registered Ponzi rejected", e)
	}
}
