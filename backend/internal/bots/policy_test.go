package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"testing"
)

func fixture(t *testing.T) game.State {
	s, _ := game.NewState(3, "bot")
	for _, x := range []struct {
		id   string
		seat int
	}{{"deck-1-clubs-08", 1}, {"deck-1-diamonds-02", 1}, {"deck-1-hearts-08", 2}} {
		var e error
		s, e = s.Move([]string{x.id}, x.seat, game.Series, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	s.Players[1].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
	return s
}
func TestPoliciesFairLegal(t *testing.T) {
	s := fixture(t)
	o, e := game.Observe(s, 1)
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range []string{"legal-random@v1", "heuristic@v1"} {
		d, e := Choose(p, o, randomstream.New([32]byte{}), 1000)
		if e != nil {
			t.Fatal(e)
		}
		if _, _, e = game.Apply(s, d.Command); e != nil {
			t.Fatal(e)
		}
	}
}
func TestEmptyBudget(t *testing.T) {
	o := game.Observation{}
	if _, e := Choose("legal-random@v1", o, randomstream.New([32]byte{}), 100); e == nil {
		t.Fatal("empty fabricated action")
	}
}
func TestHiddenPermutationAndMutation(t *testing.T) {
	s := fixture(t)
	s, _ = s.Move([]string{"deck-1-spades-02"}, 2, game.Hand, true)
	n := s.Clone()
	n, _ = n.Move([]string{"deck-1-spades-02"}, 0, game.Draw, true)
	n, _ = n.Move([]string{"deck-1-spades-03"}, 2, game.Hand, true)
	n, _ = game.ShuffleSupply(n, n.DrawOrder)
	a, _ := game.Observe(s, 1)
	b, _ := game.Observe(n, 1)
	for _, p := range []string{"legal-random@v1", "heuristic@v1"} {
		x, e := Choose(p, a, randomstream.New([32]byte{}), 1000)
		if e != nil {
			t.Fatal(e)
		}
		y, e := Choose(p, b, randomstream.New([32]byte{}), 1000)
		if e != nil || game.Digest(x) != game.Digest(y) {
			t.Fatal("hidden information influenced decision")
		}
	}
	before := game.Digest(s)
	a.Cards[0].Controller = 99
	a.Legal[0].Cards[0] = "malicious"
	if game.Digest(s) != before {
		t.Fatal("observation alias")
	}
}
func TestAdversarialCommandsRejected(t *testing.T) {
	s := fixture(t)
	o, _ := game.Observe(s, 1)
	d, e := Choose("legal-random@v1", o, randomstream.New([32]byte{}), 1000)
	if e != nil {
		t.Fatal(e)
	}
	for _, field := range []string{"actor", "game", "card", "window"} {
		c := d.Command
		c.Cards = append([]string(nil), c.Cards...)
		switch field {
		case "actor":
			c.Actor = 2
		case "game":
			c.GameID = "old"
		case "card":
			c.Cards[0] = "hidden-invalid"
		case "window":
			c.Kind = "pass"
			c.WindowID = "stale"
		}
		before := game.Digest(s)
		next, _, e := game.Apply(s, c)
		if e == nil || game.Digest(s) != before || game.Digest(next) != before {
			t.Fatalf("accepted malicious %s", field)
		}
	}
	if _, e = Choose("heuristic@v1", o, randomstream.New([32]byte{}), 1); e == nil {
		t.Fatal("budget ignored")
	}
}
func TestSamplingBudget(t *testing.T) {
	calls := 0
	_, used, e := boundedSample(func() uint64 { calls++; return 0 }, 9223372036854775809, 4)
	if e != ErrBudget || calls > 4 || used > 4 {
		t.Fatal("unbounded rejection")
	}
	v, used, e := boundedSample(func() uint64 { return 9 }, 3, 1)
	if e != nil || v != 0 || used != 1 {
		t.Fatalf("%d %d %v", v, used, e)
	}
}
func TestMenuIdentityMatchesProjection(t *testing.T) {
	s := fixture(t)
	o, e := game.Observe(s, 1)
	if e != nil {
		t.Fatal(e)
	}
	d, e := Choose("legal-random@v1", o, randomstream.New([32]byte{}), 1000)
	if e != nil {
		t.Fatal(e)
	}
	if d.MenuVersion != o.MenuCoverage {
		t.Fatalf("decision menu %s differs from projection %s", d.MenuVersion, o.MenuCoverage)
	}
}
