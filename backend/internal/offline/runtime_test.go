package offline

import (
	"encoding/json"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"strings"
	"testing"
)

func TestRuntimeSaveResumePendingResponseAndPrivacy(t *testing.T) {
	s, err := NewSession(3, 1, 1, "beginner", "rules")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		next, advanced, e := s.Advance()
		if e != nil {
			t.Fatal(e)
		}
		s = next
		if !advanced {
			break
		}
	}
	s.HumanSeat = s.Engine.Match.Game.Board.Active
	choices, err := s.Actions(s.HumanSeat)
	if err != nil {
		t.Fatal(err)
	}
	var open matchstore.Request
	for _, r := range choices {
		if r.Command != nil && r.Command.Kind == "open-series" {
			open = r
			break
		}
	}
	if open.Command == nil {
		t.Fatal("no ordinary opening")
	}
	s, err = s.Apply(s.HumanSeat, open)
	if err != nil {
		t.Fatal(err)
	}
	if s.Engine.Match.Game.Board.Pending == nil {
		t.Fatal("response window missing")
	}
	store, err := OpenStore(t.TempDir(), "rules")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err = store.Save("pending", s); err != nil {
		t.Fatal(err)
	}
	restored, err := store.Load("pending")
	if err != nil {
		t.Fatal(err)
	}
	if game.Digest(s) != game.Digest(restored) {
		t.Fatal("restore changed complete session")
	}
	view, err := restored.View()
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(view)
	for _, secret := range []string{"\"chance\"", "\"start_ledger\"", "\"command_history\"", "\"envelope\""} {
		if strings.Contains(string(raw), secret) {
			t.Fatalf("private snapshot leaked: %s", secret)
		}
	}
	for _, c := range restored.Engine.Match.Game.Board.Cards {
		if c.Controller != s.HumanSeat && (c.Zone == game.Hand || c.Zone == game.ConcealedAce) && strings.Contains(string(raw), c.Card.ID) {
			t.Fatal("opponent face leaked")
		}
	}
	actor := restored.Engine.Match.Game.Board.Pending.Responders[0]
	choice, err := restored.Suggest(actor)
	if err != nil {
		t.Fatal(err)
	}
	next, err := restored.Apply(actor, choice)
	if err != nil {
		t.Fatal(err)
	}
	if next.Version != restored.Version+1 {
		t.Fatal("version not advanced")
	}
	wrong := *choice.Command
	wrong.Actor = s.HumanSeat
	before := game.Digest(restored)
	if _, err = restored.Apply(actor, matchstore.Request{Command: &wrong}); err == nil {
		t.Fatal("actor spoof")
	}
	if game.Digest(restored) != before {
		t.Fatal("failed request mutated state")
	}
}

func TestBotsDoNotChurnOptionalFinanceWhileHumanWaits(t *testing.T) {
	s, e := NewSession(3, 1, 1, "beginner", "rules")
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 20; i++ {
		next, ok, err := s.Advance()
		if err != nil {
			t.Fatal(err)
		}
		s = next
		if !ok {
			break
		}
	}
	actor := s.Engine.Match.Game.Board.Active%3 + 1
	if _, err := s.Suggest(actor); err != ErrNoDecision {
		t.Fatalf("idle bot should wait, got %v", err)
	}
}

func TestEveryTierTakesProvenVictoryAndBeginnerBuildsSeries(t *testing.T) {
	for _, difficulty := range []string{"beginner", "standard", "advanced"} {
		s, e := NewSession(3, 1, 1, difficulty, "rules")
		if e != nil {
			t.Fatal(e)
		}
		board := s.Engine.Match.Game.Board
		var diamonds []string
		for _, c := range board.Cards {
			if c.Card.Suit == game.Diamonds && c.Card.Rank >= 2 && c.Card.Rank <= 10 && len(diamonds) < 13 {
				diamonds = append(diamonds, c.Card.ID)
			}
		}
		board, e = board.Move(diamonds, 1, game.Hand, true)
		if e != nil {
			t.Fatal(e)
		}
		board.Order = []int{1, 2, 3}
		board.Players[0].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
		s.Engine.Match.Game.Board = board
		s.Engine.Match.Game.TurnStarted = true
		choice, e := s.Suggest(1)
		if e != nil {
			t.Fatal(e)
		}
		if choice.Operation == nil || choice.Operation.Kind != "declare-ordinary" || choice.Operation.Threshold != "diamonds" {
			t.Fatalf("%s missed proven win", difficulty)
		}
		next, e := s.Apply(1, choice)
		if e != nil || next.Engine.Match.Game.Ending != "diamonds" {
			t.Fatalf("%s chose illegal win: %v", difficulty, e)
		}
	}
	s, _ := NewSession(3, 1, 1, "beginner", "rules")
	b, e := s.Engine.Match.Game.Board.Move([]string{"deck-1-clubs-02"}, 1, game.Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	b.Order = []int{1, 2, 3}
	s.Engine.Match.Game.Board = b
	s.Engine.Match.Game.TurnStarted = true
	choice, e := s.Suggest(1)
	if e != nil || choice.Command == nil || choice.Command.Kind != "open-series" {
		t.Fatalf("beginner wasted productive turn: %v", e)
	}
}

func TestAdvancedReservesRoyalWhereBeginnerAttaches(t *testing.T) {
	s, _ := NewSession(3, 1, 1, "beginner", "rules")
	b, e := s.Engine.Match.Game.Board.Move([]string{"deck-1-hearts-02"}, 1, game.Series, true)
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.Move([]string{"deck-1-hearts-12"}, 1, game.Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	b.Order = []int{1, 2, 3}
	b.Players[0].History = []game.Suit{game.Hearts}
	s.Engine.Match.Game.Board = b
	s.Engine.Match.Game.TurnStarted = true
	simple, e := s.Suggest(1)
	if e != nil {
		t.Fatal(e)
	}
	s.Difficulty = "advanced"
	advanced, e := s.Suggest(1)
	if e != nil {
		t.Fatal(e)
	}
	if simple.Command == nil || simple.Command.Kind != "attach" || advanced.Operation == nil || advanced.Operation.Kind != "end-turn" {
		t.Fatal("tier royal-reservation behavior did not differ")
	}
}
