package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"slices"
	"testing"
)

func orderV2Fixture(id string) game.Observation {
	existing := game.FormationRecord{ID: id + "-people", Spec: game.FormationSpec{Kind: "people", Cards: []string{"q1", "q2"}, Protection: &game.FormationProtection{Formation: id + "-target"}}}
	target := game.FormationRecord{ID: id + "-target", Spec: game.FormationSpec{Kind: "kidnapper", Cards: []string{"j1", "j2"}}}
	o := game.Observation{GameID: id, Seat: 1, Formations: []game.FormationRecord{existing, target}}
	for _, suit := range []game.Suit{game.Clubs, game.Spades} {
		o.Legal = append(o.Legal, game.Command{GameID: id, ID: id + string(suit), Actor: 1, Kind: "open-formation", FormationID: existing.ID, Formation: &game.FormationSpec{Kind: "people", Cards: []string{"q1", "q2"}, Protection: &game.FormationProtection{Series: suit}}})
	}
	return o
}
func TestV2ReconfigurationKeyPreservesRequestedSpec(t *testing.T) {
	o := orderV2Fixture("a")
	if gameplayActionKey(o, o.Legal[0]) != gameplayActionKey(o, o.Legal[1]) {
		t.Fatal("historical v1 collision behavior unexpectedly changed")
	}
	if gameplayActionKeyV2(o, o.Legal[0]) == gameplayActionKeyV2(o, o.Legal[1]) {
		t.Fatal("distinct requested protection collapsed")
	}
}
func TestV2FormationOrderIgnoresMenuAndAdministrativeIDs(t *testing.T) {
	for _, policy := range []string{"economic@v2", "pressure@v2", "opportunity@v2", "opportunity-no-retain@v2", "opportunity-no-finance@v2", "bargaining@v2"} {
		var selected game.Suit
		for _, id := range []string{"a", "different-game"} {
			for _, reverse := range []bool{false, true} {
				o := orderV2Fixture(id)
				if reverse {
					slices.Reverse(o.Legal)
					slices.Reverse(o.Formations)
				}
				before := game.Digest(o)
				d, e := ChooseGameplay(policy, o, nil, 100)
				if e != nil || d.Policy != policy {
					t.Fatal(policy, e)
				}
				if selected == "" {
					selected = d.Command.Formation.Protection.Series
				}
				if selected != d.Command.Formation.Protection.Series || game.Digest(o) != before {
					t.Fatal("menu/admin identity changed choice or mutated observation", policy)
				}
			}
		}
	}
	a := orderV2Fixture("a")
	b := orderV2Fixture("b")
	a.Legal[0].Formation.Protection = &game.FormationProtection{Formation: "a-target"}
	b.Legal[0].Formation.Protection = &game.FormationProtection{Formation: "b-target"}
	if gameplayActionKeyV2(a, a.Legal[0]) != gameplayActionKeyV2(b, b.Legal[0]) {
		t.Fatal("requested protection admin reference not normalized")
	}
}
func TestV2FinancialAliasesRetainPolicyLabelsAndBehavior(t *testing.T) {
	l := triggeredFinance(t, 10, 10)
	o, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"economic", "pressure", "opportunity", "opportunity-no-retain", "opportunity-no-finance", "bargaining"} {
		a, e := ChooseGameplayFinancial(name+"@v1", o, nil, 1000)
		if e != nil {
			t.Fatal(e)
		}
		b, e := ChooseGameplayFinancial(name+"@v2", o, nil, 1000)
		if e != nil || b.Policy != name+"@v2" || game.Digest(a.Operation) != game.Digest(b.Operation) || a.Operations != b.Operations {
			t.Fatal("financial alias changed policy", name, e)
		}
	}
}

func TestV2AuthorizedHiddenSwapSameChoice(t *testing.T) {
	s, _ := game.NewState(3, "v2-private")
	s, _ = s.Move([]string{"deck-1-clubs-02", "deck-1-hearts-02"}, 1, game.Series, false)
	s.Players[0].History = []game.Suit{game.Clubs, game.Hearts}
	s, _ = s.Move([]string{"deck-1-clubs-11"}, 1, game.Hand, false)
	s, _ = s.Move([]string{"deck-2-hearts-12"}, 2, game.Hand, false)
	s, _ = s.Move([]string{"deck-2-spades-03"}, 3, game.Hand, false)
	n := s.Clone()
	for i := range n.Cards {
		switch n.Cards[i].Card.ID {
		case "deck-2-hearts-12":
			n.Cards[i].Controller = 3
		case "deck-2-spades-03":
			n.Cards[i].Controller = 2
		}
	}
	a, e := game.GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	b, e := game.GameplayObservation(n, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	if game.Digest(a) != game.Digest(b) {
		t.Fatal("hidden observation or menu leak")
	}
	for _, p := range []string{"economic@v2", "pressure@v2", "opportunity@v2", "opportunity-no-retain@v2", "opportunity-no-finance@v2", "bargaining@v2"} {
		x, e := ChooseGameplay(p, a, nil, 100000)
		if e != nil {
			t.Fatal(e)
		}
		y, e := ChooseGameplay(p, b, nil, 100000)
		if e != nil || game.Digest(x) != game.Digest(y) {
			t.Fatal("hidden state affected v2", p, e)
		}
	}
}

func TestV2ActualMenuReconfigurationCollision(t *testing.T) {
	s, _ := game.NewState(3, "v2-real-menu")
	s, _ = s.Move([]string{"deck-1-clubs-02", "deck-1-hearts-02", "deck-1-spades-02", "deck-1-diamonds-02"}, 1, game.Series, false)
	s.Players[0].History = []game.Suit{game.Clubs, game.Hearts, game.Spades, game.Diamonds}
	s, _ = s.Move([]string{"deck-1-hearts-12", "deck-2-hearts-12"}, 1, game.Hand, false)
	var e error
	s, e = game.OpenFormation(s, 1, "public-protector", game.FormationSpec{Kind: "great-people", Cards: []string{"deck-1-hearts-12", "deck-2-hearts-12"}, Protection: &game.FormationProtection{Series: game.Diamonds}})
	if e != nil {
		t.Fatal(e)
	}
	s.Turn++
	s.Players[0].OpeningUsed = false
	o, e := game.GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	var commands []game.Command
	for _, c := range o.Legal {
		if c.Kind == "open-formation" && c.FormationID == "public-protector" && c.Formation != nil && c.Formation.Protection != nil && (c.Formation.Protection.Series == game.Clubs || c.Formation.Protection.Series == game.Spades) {
			commands = append(commands, c)
		}
	}
	if len(commands) != 2 {
		t.Fatalf("expected actual legal competing reconfigurations, got %d", len(commands))
	}
	if gameplayActionKey(o, commands[0]) != gameplayActionKey(o, commands[1]) || gameplayActionKeyV2(o, commands[0]) == gameplayActionKeyV2(o, commands[1]) {
		t.Fatal("real menu collision not preserved historically/repaired prospectively")
	}
	for _, c := range commands {
		if _, _, e = game.Apply(s, c); e != nil {
			t.Fatal("fixture menu command not legal", e)
		}
	}
}
