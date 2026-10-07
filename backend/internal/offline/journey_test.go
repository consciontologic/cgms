package offline

import (
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestFullLocalBotGamesAndDurableRestart(t *testing.T) {
	for _, population := range []int{3, 4} {
		t.Run(fmt.Sprint(population), func(t *testing.T) {
			s, err := NewSession(population, 2, 1, "beginner", "rules")
			if err != nil {
				t.Fatal(err)
			}
			store, err := OpenStore(t.TempDir(), "rules")
			if err != nil {
				t.Fatal(err)
			}
			defer store.Close()
			steps := 0
			for ; steps < 2500 && len(s.Engine.Match.Game.Ledger.Completed) == 0; steps++ {
				n, advanced, e := s.Advance()
				if e != nil {
					t.Fatal(e)
				}
				if !advanced {
					chosen := false
					for offset := 0; offset < population; offset++ {
						seat := (s.Engine.Match.Game.Board.Active-1+offset)%population + 1
						request, e := s.Suggest(seat)
						if errors.Is(e, ErrNoDecision) {
							continue
						}
						if e != nil {
							t.Fatalf("step %d seat %d: %v", steps, seat, e)
						}
						n, e = s.Apply(seat, request)
						if e != nil {
							t.Fatalf("illegal bot step %d: %v", steps, e)
						}
						chosen = true
						break
					}
					if !chosen {
						t.Fatalf("local game stalled step %d round %d", steps, s.Engine.Match.Game.Board.Round)
					}
				}
				s = n
				if steps%25 == 0 {
					if e = store.Save("journey", s); e != nil {
						t.Fatal(e)
					}
					restored, e := store.Load("journey")
					if e != nil {
						t.Fatal(e)
					}
					if game.Digest(restored) != game.Digest(s) {
						t.Fatal("restart divergence")
					}
					s = restored
				}
			}
			if len(s.Engine.Match.Game.Ledger.Completed) != 1 {
				t.Fatalf("incomplete after %d actions", steps)
			}
			oldID := s.Engine.Match.Game.Board.GameID
			next, advanced, err := s.Advance()
			if err != nil || !advanced || next.Engine.Match.Game.Board.GameID == oldID || len(next.Engine.Match.Game.Ledger.Completed) != 1 {
				t.Fatalf("next game progression: %v", err)
			}
			if err = store.Save("next", next); err != nil {
				t.Fatal(err)
			}
			restored, err := store.Load("next")
			if err != nil || game.Digest(next) != game.Digest(restored) {
				t.Fatal("next game restore")
			}
			t.Logf("%d-player full local engine completed in %d transitions (%s)", population, steps, s.Engine.Match.Game.Ending)
		})
	}
}
