package sim

import (
	"context"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"testing"
)

func TestCuratedNarrowBotScenarios(t *testing.T) {
	for _, population := range []int{3, 4} {
		t.Run(fmt.Sprint(population), func(t *testing.T) {
			name := "narrow-combat-bots.json"
			if population == 4 {
				name = "narrow-combat-bots-4p.json"
			}
			raw, e := os.ReadFile("../../../sims/scenarios/" + name)
			if e != nil {
				t.Fatal(e)
			}
			sc, e := loadScenario(raw, population)
			if e != nil {
				t.Fatal(e)
			}
			if !sc.RunBots || sc.State == nil || len(sc.State.CurrentSeries(1)) != 2 || len(sc.State.CurrentSeries(2)) != 2 {
				t.Fatal("fixture omitted physical support")
			}
			for _, id := range []string{"deck-1-clubs-08", "deck-1-hearts-08"} {
				found := false
				for _, known := range sc.State.PublicHistory {
					if known == id {
						found = true
					}
				}
				if !found {
					t.Fatal("public fixture identity missing history")
				}
			}
			j := Job{Index: 0, Population: population, Block: "fixture-block", Rotation: Rotation{ID: "r0"}, Treatment: "baseline", Games: 1}
			for seat := 0; seat < population; seat++ {
				id := fmt.Sprintf("p%d", seat)
				j.Rotation.Seats = append(j.Rotation.Seats, id)
				j.Policies = append(j.Policies, Policy{ParticipantID: id, Policy: "legal-random", Version: "v1", Information: "seat-projection"})
			}
			r := runScenario(context.Background(), j, &sc, "42", "fixture", 100)
			if r.Outcome.ExitCode != 3 || r.Outcome.GameComplete || r.Outcome.MatchComplete || len(r.Trace) != population {
				t.Fatalf("dishonest or incomplete fixture outcome: %#v trace=%d", r.Outcome, len(r.Trace))
			}
			final, ok := r.Checkpoint.(game.State)
			if !ok {
				t.Fatal("missing card checkpoint")
			}
			heart, _ := final.Card("deck-1-hearts-08")
			club, _ := final.Card("deck-1-clubs-08")
			if heart.Zone != game.Draw || club.Zone != game.Series || club.UsedTurn != final.Turn || !final.NeedsShuffle {
				t.Fatal("accepted exact attack did not preserve expected boundary")
			}
			for _, trace := range r.Trace {
				if trace.BotDecision == nil || trace.BotRandomBefore == nil || trace.BotRandomAfter == nil {
					t.Fatal("missing private bot replay provenance")
				}
			}
		})
	}
}
