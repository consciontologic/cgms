package game

import "testing"

func TestJusticeExileInfiltratorConfinementCosts(t *testing.T) {
	for _, kind := range []string{"justice", "exile", "infiltrator"} {
		t.Run(kind, func(t *testing.T) {
			s := combatFixture(t)
			s, _ = s.Move([]string{"deck-1-spades-03"}, 2, Series, true)
			s, _ = s.Move([]string{"deck-1-diamonds-01"}, 2, ConcealedAce, false)
			c := Command{GameID: s.GameID, ID: "accepted", Actor: 1, Kind: "attack", Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}}
			var retained string
			switch kind {
			case "justice":
				ids := []string{"deck-1-hearts-12", "deck-1-spades-12", "deck-1-clubs-12", "deck-1-diamonds-12"}
				s, _ = s.Move(ids, 1, Hand, false)
				var e error
				s, e = OpenFormation(s, 1, "justice-real", FormationSpec{Kind: "justice", Cards: ids})
				if e != nil {
					t.Fatal(e)
				}
				c.Kind = "justice"
				c.Cards = nil
				c.Targets = nil
				c.FormationID = "justice-real"
				c.AceID = ids[0]
				retained = ids[0]
			case "exile":
				retained = "deck-1-clubs-12"
				s, _ = s.Move([]string{retained}, 1, Hand, false)
				s, _ = AttachRoyal(s, 1, retained, Clubs)
				s, _ = s.Move([]string{"deck-1-hearts-12"}, 2, Attachment, true)
				c.Exile = true
			case "infiltrator":
				s, _ = s.Move([]string{"deck-1-clubs-11"}, 1, Hand, false)
				s, _ = AttachRoyal(s, 1, "deck-1-clubs-11", Clubs)
				s, _ = s.Move([]string{"deck-1-spades-08"}, 2, Series, true)
				c.Infiltrator = true
				c.Targets = []string{"deck-1-spades-08"}
			}
			n, _, e := Apply(s, c)
			if e != nil {
				t.Fatal(e)
			}
			n = roundtripCombat(t, n)
			n, _, e = Apply(n, Command{GameID: s.GameID, ID: "cancel", Kind: "confinement", Actor: 2, WindowID: c.ID, AceID: "deck-1-diamonds-01"})
			if e != nil {
				t.Fatal(e)
			}
			if !n.Players[0].Confined || n.Players[1].AceRound != s.Round {
				t.Fatal("cancellation/ace quota missing")
			}
			if retained != "" {
				before, _ := s.Card(retained)
				after, _ := n.Card(retained)
				if after.Zone != before.Zone || after.Controller != 1 {
					t.Fatal("success-only cost incorrectly paid")
				}
			}
			switch kind {
			case "justice":
				if n.Players[0].Quotas["justice"] != turnCycle(s, 1) {
					t.Fatal("Justice quota refunded")
				}
			case "exile":
				if n.Players[0].Quotas["exile"] != s.Round {
					t.Fatal("Exile quota refunded")
				}
			case "infiltrator":
				jack, _ := n.Card("deck-1-clubs-11")
				if jack.Zone != Draw || n.Players[0].Quotas["infiltrator"] != turnCycle(s, 1) {
					t.Fatal("committed bypass return/quota refunded")
				}
			}
			if kind != "justice" {
				club, _ := n.Card("deck-1-clubs-08")
				if club.UsedTurn != s.Turn || club.Zone != Series {
					t.Fatal("committed physical Club or success-only return incorrect")
				}
			}
		})
	}
}
