package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestGameplayPoliciesPreferAvailableFormationToAttachment(t *testing.T) {
	s, _ := game.NewState(3, "formation-competence")
	s, _ = s.Move([]string{"deck-1-clubs-02", "deck-1-hearts-02"}, 1, game.Series, false)
	s.Players[0].History = []game.Suit{game.Clubs, game.Hearts}
	s, _ = s.Move([]string{"deck-1-clubs-11", "deck-2-clubs-11"}, 1, game.Hand, false)
	o, err := game.GameplayObservation(s, 1, 100000)
	if err != nil {
		t.Fatal(err)
	}
	hasFormation, hasAttachment := false, false
	for _, c := range o.Legal {
		hasFormation = hasFormation || c.Kind == "open-formation"
		hasAttachment = hasAttachment || c.Kind == "attach"
	}
	if !hasFormation || !hasAttachment {
		t.Fatal("fixture must exercise actual competing legal menu categories")
	}
	for _, policy := range []string{"economic@v1", "pressure@v1"} {
		decision, err := ChooseGameplay(policy, o, nil, 100000)
		if err != nil || decision.Command.Kind != "open-formation" {
			t.Fatal("policy ignored available combo", policy, err)
		}
	}
}

// This records the current myopic policy, not a claim that immediate attachment
// is optimal: an unpaired royal is not reserved for a possible future combo.
func TestGameplayPolicyDevelopmentUnpairedRoyalUsesImmediateAttachment(t *testing.T) {
	s, _ := game.NewState(3, "formation-reservation")
	s, _ = s.Move([]string{"deck-1-clubs-02", "deck-1-hearts-02"}, 1, game.Series, false)
	s.Players[0].History = []game.Suit{game.Clubs, game.Hearts}
	s, _ = s.Move([]string{"deck-1-clubs-11"}, 1, game.Hand, false)
	o, err := game.GameplayObservation(s, 1, 100000)
	if err != nil {
		t.Fatal(err)
	}
	for _, policy := range []string{"economic@v1", "pressure@v1"} {
		decision, err := ChooseGameplay(policy, o, nil, 100000)
		if err != nil || decision.Command.Kind != "attach" {
			t.Fatal("documented immediate-use policy changed", policy, err)
		}
	}
}
