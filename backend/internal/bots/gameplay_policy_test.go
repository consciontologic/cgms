package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestGameplayPolicyImmediateWinAndExplicitBudget(t *testing.T) {
	o := game.Observation{GameID: "g", Seat: 1, MenuCoverage: game.GameplayMenuVersion, Legal: []game.Command{{GameID: "g", Actor: 1, Kind: "end-turn"}, {GameID: "g", Actor: 1, Kind: "ordinary-victory", Value: 13}}}
	for _, policy := range []string{"economic@v1", "pressure@v1"} {
		d, e := ChooseGameplay(policy, o, nil, 20)
		if e != nil || d.Command.Kind != "ordinary-victory" {
			t.Fatal("missed immediate valid ending", e)
		}
		if _, e = ChooseGameplay(policy, o, nil, 1); e != ErrBudget {
			t.Fatal("budget not explicit")
		}
	}
}
func TestGameplayPoliciesDistinctAndLegalDefense(t *testing.T) {
	o := game.Observation{GameID: "g", Seat: 1, Legal: []game.Command{{GameID: "g", Actor: 1, Kind: "ringleader"}, {GameID: "g", Actor: 1, Kind: "attack", Targets: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}}}}
	a, _ := ChooseGameplay("economic@v1", o, nil, 20)
	b, _ := ChooseGameplay("pressure@v1", o, nil, 20)
	if a.Command.Kind == b.Command.Kind {
		t.Fatal("strategies did not differ")
	}
	o.Legal = []game.Command{{GameID: "g", Actor: 1, Kind: "pass"}, {GameID: "g", Actor: 1, Kind: "confinement"}}
	a, e := ChooseGameplay("economic@v1", o, nil, 20)
	if e != nil || a.Command.Kind != "confinement" {
		t.Fatal("ignored legal defense")
	}
}

func TestGameplayPolicyTieBreakIgnoresAdministrativeIdentity(t *testing.T) {
	choice := ""
	for _, id := range []string{"g0", "g1", "g2", "g3", "g4", "g5", "g6", "g7", "g8", "g9"} {
		o := game.Observation{GameID: id, Seat: 1, Legal: []game.Command{{GameID: id, ID: id + "/a", Actor: 1, Kind: "attach", Suit: game.Hearts, Cards: []string{"king-a"}}, {GameID: id, ID: id + "/b", Actor: 1, Kind: "attach", Suit: game.Clubs, Cards: []string{"king-b"}}}}
		d, e := ChooseGameplay("economic@v1", o, nil, 20)
		if e != nil {
			t.Fatal(e)
		}
		if choice == "" {
			choice = d.Command.Cards[0]
		}
		if d.Command.Cards[0] != choice {
			t.Fatal("same semantic tie changed with game identity")
		}
	}
}
