package offlineprobe

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCardplayPolicyUsesOnlyOwnProjection(t *testing.T) {
	s, _ := game.NewState(3, "cardplay-test")
	s, _ = s.Move([]string{"deck-1-clubs-02"}, 1, game.Hand, true)
	o, _ := game.ObserveWithoutMenu(s, 1)
	d := CardplayChoice(o, false, true, false)
	if d == nil || d.Kind != "open-series" || d.Suit != game.Clubs {
		t.Fatal("expected legal own-card opening")
	}
	o.Seat = 2
	if CardplayChoice(o, false, true, false) != nil {
		t.Fatal("out of turn action")
	}
}

func TestCardplayCompleteMatchesAndExport(t *testing.T) {
	for _, population := range []int{3, 4} {
		b, _ := game.NewState(population, "cardplay-1")
		m, _ := game.NewMatchLifecycle(b, 3)
		request := CardplayRequest{Schema: CardplayVersion, Match: m}
		frames := []CardplayFrame{}
		f, _ := cardplayFrame(m, nil)
		frames = append(frames, f)
		rng := rand.New(rand.NewPCG(uint64(population), 41))
		openings, passes, ends := 0, 0, 0
		for index := 0; index < 1800; index++ {
			l := m.Game
			if l.Ledger.Phase == "finalized" && len(l.Ledger.Completed) == 3 {
				break
			}
			step := CardplayStep{GameID: l.Board.GameID}
			switch {
			case l.Ledger.Phase == "finalized":
				step.Kind = "next"
				step.NextGame = fmt.Sprintf("cardplay-%d", len(l.Ledger.Completed)+1)
			case l.Ending != "":
				step.Kind = "finalize"
				for i, done := range l.Ledger.Finished {
					if !done {
						step.Kind = "finish"
						step.Seat = i + 1
						break
					}
				}
			case len(l.Board.DrawOrder) == 104:
				step.Kind = "deal"
				for tries := 0; tries < 100; tries++ {
					order := slices.Clone(l.Board.DrawOrder)
					rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
					_, retry, e := game.DealAttempt(l.Board, order)
					if e != nil {
						t.Fatal(e)
					}
					if !retry {
						step.Order = order
						break
					}
				}
				if step.Order == nil {
					t.Fatal("deal guarantee budget")
				}
			case len(l.Board.Order) == 0 || l.RoundClosed:
				step.Kind = "initiative"
				board := l.Board.Clone()
				if l.RoundClosed {
					step.Kind = "begin-round"
					board.Round++
				}
				order, e := game.InitiativeReturnSupply(board, nil)
				if e != nil {
					t.Fatal(e)
				}
				rng.Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })
				step.Order = order
			case !l.TurnStarted:
				step.Kind = "begin-turn"
			default:
				frame, e := cardplayFrame(m, nil)
				if e != nil {
					t.Fatal(e)
				}
				var decision *CardplayDecision
				for _, d := range frame.Decisions {
					if d != nil {
						if decision != nil {
							t.Fatal("multiple decision actors")
						}
						decision = d
					}
				}
				if decision == nil {
					t.Fatal("policy stalled")
				}
				if decision.Kind == "end-turn" {
					step.Kind = "end-turn"
					ends++
				} else {
					step.Kind = "command"
					step.Command = &game.Command{GameID: l.Board.GameID, ID: fmt.Sprintf("cardplay-step-%d", index), Kind: decision.Kind, Actor: decision.Actor, Suit: decision.Suit, WindowID: decision.WindowID}
					if decision.Kind == "open-series" {
						openings++
					} else {
						passes++
					}
				}
			}
			var events []game.Event
			var e error
			m, events, e = cardplayStep(m, step)
			if e != nil {
				t.Fatalf("population %d step %d %s: %v", population, index, step.Kind, e)
			}
			request.Steps = append(request.Steps, step)
			frame, e := cardplayFrame(m, events)
			if e != nil {
				t.Fatal(e)
			}
			frames = append(frames, frame)
		}
		if len(m.Game.Ledger.Completed) != 3 || m.Game.Ledger.Phase != "finalized" || m.Game.Ending != "round-limit" || ends != 39*population || openings < 9 || passes != openings*(population-1) {
			t.Fatalf("not a complete naturally played match: games=%d turns=%d openings=%d passes=%d", len(m.Game.Ledger.Completed), ends, openings, passes)
		}
		raw, e := canonical.Marshal(request)
		if e != nil {
			t.Fatal(e)
		}
		expected := Execute(raw)
		var result CardplayResponse
		if e = canonical.Decode(expected, &result); e != nil || result.Error != "" || len(result.Digests) != len(frames) {
			t.Fatalf("trace failed: %v %s", e, result.Error)
		}
		for i, f := range frames {
			if result.Digests[i] != game.Digest(f) {
				t.Fatalf("frame %d differs", i)
			}
		}
		dir := os.Getenv("CGMS_PROBE_FIXTURES")
		if dir != "" {
			name := fmt.Sprintf("cardplay-%d", population)
			for suffix, data := range map[string][]byte{"request.json": raw, "expected.json": expected} {
				if e = os.WriteFile(filepath.Join(dir, name+"."+suffix), data, 0600); e != nil {
					t.Fatal(e)
				}
			}
			file, e := os.OpenFile(filepath.Join(dir, name+".frames.jsonl"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
			if e != nil {
				t.Fatal(e)
			}
			for _, f := range frames {
				raw, e := canonical.Marshal(f)
				if e != nil {
					t.Fatal(e)
				}
				if _, e = file.Write(append(raw, '\n')); e != nil {
					t.Fatal(e)
				}
			}
			if e = file.Close(); e != nil {
				t.Fatal(e)
			}
		}
		t.Logf("%d players: 3 complete games, %d turns, %d openings, %d passes, %d frames", population, ends, openings, passes, len(frames))
	}
}

func TestCardplayRejectsUnsupportedProfile(t *testing.T) {
	b, _ := game.NewState(3, "profile")
	b.Players[0].Confined = true
	m, _ := game.NewMatchLifecycle(b, 3)
	raw, _ := canonical.Marshal(CardplayRequest{Schema: CardplayVersion, Match: m})
	var result CardplayResponse
	if e := canonical.Decode(Execute(raw), &result); e != nil || result.Error != "invalid_state" || result.Final != nil {
		t.Fatal("unsupported state accepted")
	}
}
