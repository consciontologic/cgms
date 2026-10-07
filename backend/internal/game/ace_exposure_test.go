package game

import "testing"

func TestUnusedAcePublicExposureWastesWithoutPowerQuota(t *testing.T) {
	s := combatFixture(t)
	s, _ = s.Move([]string{"deck-1-hearts-01"}, 1, ConcealedAce, false)
	c := Command{GameID: s.GameID, ID: "public-disclosure", Kind: "expose-unused-ace", Actor: 1, AceID: "deck-1-hearts-01"}
	n, ev, e := Apply(s, c)
	if e != nil {
		t.Fatal(e)
	}
	ace, _ := n.Card(c.AceID)
	if ace.Zone != Discard || n.Players[0].AceRound != 0 || n.Players[0].CompensationUsed || len(n.Effects) != 0 || len(ev) != 1 {
		t.Fatal("wasted exposure incorrectly activated power")
	}
	if _, _, e := Apply(s, Command{GameID: s.GameID, ID: "other", Kind: c.Kind, Actor: 2, AceID: c.AceID}); e == nil {
		t.Fatal("opponent disclosure accepted")
	}
	old, _ := s.Card(c.AceID)
	if old.Zone != ConcealedAce {
		t.Fatal("mutated input")
	}
}
