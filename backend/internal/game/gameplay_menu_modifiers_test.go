package game

import "testing"

func TestGameplayMenuNamedAttackAceCombinations(t *testing.T) {
	for _, kind := range []string{"infiltrator", "baron", "bomb"} {
		t.Run(kind, func(t *testing.T) {
			s := combatFixture(t)
			s, _ = s.Move([]string{"deck-1-clubs-01"}, 1, ConcealedAce, false)
			c := Command{ID: "manual", GameID: s.GameID, Actor: 1, Kind: "attack", Cards: []string{"deck-1-clubs-08"}, AceID: "deck-1-clubs-01", Value: 2, Targets: []string{"deck-1-hearts-08"}}
			switch kind {
			case "infiltrator":
				s, _ = s.Move([]string{"deck-1-clubs-11"}, 1, Hand, false)
				s, _ = AttachRoyal(s, 1, "deck-1-clubs-11", Clubs)
				s, _ = s.Move([]string{"deck-1-spades-10"}, 2, Series, true)
				c.Targets = []string{"deck-1-spades-10"}
				c.Infiltrator = true
			case "baron":
				ids := []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-diamonds-13"}
				s, _ = s.Move(ids, 1, Hand, false)
				var e error
				s, e = OpenFormation(s, 1, "registered-baron", FormationSpec{Kind: "baron", Cards: ids})
				if e != nil {
					t.Fatal(e)
				}
				c.Kind = "baron-attack"
			case "bomb":
				s, _ = s.Move([]string{"deck-1-clubs-08"}, 0, Discard, false)
				s, _ = s.Move([]string{"deck-1-clubs-02"}, 1, Series, true)
				s, _ = s.Move([]string{"deck-1-hearts-11"}, 2, Attachment, true)
				for i := range s.Cards {
					if s.Cards[i].Card.ID == "deck-1-hearts-11" {
						s.Cards[i].Allocation = "hearts"
					}
				}
				c.Kind = "bomb-attack"
				c.Cards = []string{"deck-1-clubs-02"}
				c.Targets = []string{"deck-1-hearts-11"}
				c.Value = 3
			}
			if _, _, e := Apply(s, c); e != nil {
				t.Fatalf("manual accepted-rule fixture illegal: %v", e)
			}
			o, e := GameplayObservation(s, 1, 100000)
			if e != nil {
				t.Fatal(e)
			}
			for _, x := range o.Legal {
				if x.Kind == c.Kind && x.Infiltrator == c.Infiltrator && x.AceID == c.AceID && x.Value == c.Value && Digest(x.Cards) == Digest(c.Cards) && Digest(x.Targets) == Digest(c.Targets) {
					return
				}
			}
			t.Fatal("legal named attack and Ace combination absent")
		})
	}
}
