package bots

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestFrozenHeuristicsRemoveVisibleImmediateDiamondOpportunity(t *testing.T) {
	s, _ := game.NewState(3, "visible-loss-competence")
	move := func(ids []string, seat int) {
		t.Helper()
		var err error
		s, err = s.Move(ids, seat, game.Series, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	attackCard := "deck-1-clubs-02"
	move([]string{attackCard, "deck-1-hearts-02"}, 1)
	diamonds := []string{}
	for rank := 2; rank <= 10; rank++ {
		diamonds = append(diamonds, fmt.Sprintf("deck-1-diamonds-%02d", rank))
	}
	for rank := 3; rank <= 6; rank++ {
		diamonds = append(diamonds, fmt.Sprintf("deck-2-diamonds-%02d", rank))
	}
	move(diamonds, 2)
	move([]string{"deck-1-spades-02"}, 3)
	s.Players[0].History = []game.Suit{game.Clubs, game.Hearts}
	for _, seat := range []int{2, 3} {
		s.Players[seat-1].History = []game.Suit{game.Clubs, game.Hearts, game.Spades, game.Diamonds}
	}
	count := func(state game.State, seat int) int {
		n := 0
		for _, c := range state.Cards {
			if c.Controller == seat && c.Card.Suit == game.Diamonds && c.Card.Rank >= 2 && c.Card.Rank <= 10 {
				n++
			}
		}
		return n
	}
	if count(s, 2) != 13 {
		t.Fatal("fixture lacks visible threshold")
	}
	o, err := game.GameplayObservation(s, 1, 100000)
	if err != nil {
		t.Fatal(err)
	}
	hasThreatAttack, hasOtherAttack, hasEnd := false, false, false
	for _, c := range o.Legal {
		hasEnd = hasEnd || c.Kind == "end-turn"
		if c.Kind == "attack" && len(c.Targets) == 1 {
			hasThreatAttack = hasThreatAttack || c.Targets[0] == "deck-1-diamonds-02"
			hasOtherAttack = hasOtherAttack || c.Targets[0] == "deck-1-spades-02"
		}
	}
	if !hasThreatAttack || !hasOtherAttack || !hasEnd {
		t.Fatal("actual menu lacks compared alternatives")
	}
	for _, policy := range []string{"economic@v1", "pressure@v1", "opportunity@v1", "opportunity-no-retain@v1", "opportunity-no-finance@v1", "bargaining@v1"} {
		t.Run(policy, func(t *testing.T) {
			decision, err := ChooseGameplay(policy, o, nil, 100000)
			if err != nil {
				t.Fatal(err)
			}
			if decision.Command.Kind != "attack" || len(decision.Command.Targets) != 1 || decision.Command.Targets[0] != "deck-1-diamonds-02" {
				t.Fatal("frozen policy misses public imminent Diamond opportunity")
			}
			before := game.Digest(s)
			n, _, err := game.Apply(s, decision.Command)
			if err != nil {
				t.Fatal("chosen costs not legal", err)
			}
			if game.Digest(s) != before {
				t.Fatal("choice mutated initial state")
			}
			c, _ := n.Card(attackCard)
			if c.UsedTurn != s.Turn {
				t.Fatal("Club cost not committed")
			}
			for n.Pending != nil {
				seat := n.Pending.Responders[n.Pending.Cursor]
				n, _, err = game.Apply(n, game.Command{GameID: n.GameID, ID: fmt.Sprintf("reply-%d", seat), Kind: "pass", Actor: seat, WindowID: decision.Command.ID})
				if err != nil {
					t.Fatal(err)
				}
			}
			if count(n, 2) != 12 || count(n, 1) != 1 {
				t.Fatal("resolved action failed to remove visible immediate threshold")
			}
		})
	}
}
