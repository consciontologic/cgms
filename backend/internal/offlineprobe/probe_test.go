package offlineprobe

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
)

func position(t *testing.T, population int) game.State {
	t.Helper()
	s, err := game.NewState(population, "synthetic-offline-probe")
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range []struct {
		id   string
		seat int
		zone game.Zone
	}{
		{"deck-1-clubs-08", 1, game.Series}, {"deck-1-diamonds-02", 1, game.Series},
		{"deck-1-hearts-08", 2, game.Series}, {"deck-1-spades-07", 2, game.Hand},
		{"deck-1-diamonds-01", 2, game.ConcealedAce},
	} {
		s, err = s.Move([]string{x.id}, x.seat, x.zone, true)
		if err != nil {
			t.Fatal(err)
		}
	}
	s.Players[1].History = []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds}
	return s
}
func request(t *testing.T, population int) Request {
	s := position(t, population)
	cs := []game.Command{{GameID: s.GameID, ID: "attack", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-hearts-08"}}}
	for seat := 2; seat <= population; seat++ {
		cs = append(cs, game.Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "attack"})
	}
	return Request{Schema: Version, State: s, Commands: cs}
}
func invoke(t *testing.T, r Request) Response {
	t.Helper()
	b, e := canonical.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	var out Response
	if e = canonical.Decode(Execute(b), &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func TestTransitionsAndPrivateViews(t *testing.T) {
	for _, population := range []int{3, 4} {
		r := request(t, population)
		before := game.Digest(r)
		out := invoke(t, r)
		if out.Error != "" || len(out.Frames) != len(r.Commands)+1 {
			t.Fatalf("probe failed: %s", out.Error)
		}
		if before != game.Digest(r) {
			t.Fatal("mutated request")
		}
		last := out.Frames[len(out.Frames)-1].State
		target, _ := last.Card("deck-1-hearts-08")
		club, _ := last.Card("deck-1-clubs-08")
		if target.Zone != game.Draw || club.UsedTurn != 1 || last.Pending != nil {
			t.Fatal("combat contract")
		}
		for _, f := range out.Frames {
			if len(f.Observations) != population {
				t.Fatal("missing seats")
			}
			for _, o := range f.Observations {
				if o.Players[1].HandCount != 1 || o.Players[1].AceCount != 1 {
					t.Fatal("public counts")
				}
				for _, c := range o.Cards {
					if c.Controller != o.Seat && (c.Zone == game.Hand || c.Zone == game.ConcealedAce || c.Zone == game.Draw) {
						t.Fatal("private card leaked")
					}
				}
			}
		}
		if out.Frames[0].Decision == nil || out.Frames[0].Decision.Command.Kind != "attack" {
			t.Fatal("missing bounded baseline bot decision")
		}
		// A serialized pending action resumes with exactly the same successor.
		resume := r
		resume.State = out.Frames[1].State
		resume.Commands = r.Commands[1:]
		restored := invoke(t, resume)
		if restored.Error != "" || game.Digest(restored.Frames[len(restored.Frames)-1].State) != game.Digest(last) {
			t.Fatal("resume diverged")
		}
	}
}
func TestHiddenNoninterference(t *testing.T) {
	r := request(t, 3)
	r.Commands = nil
	before := invoke(t, r)
	for i := range r.State.Cards {
		if r.State.Cards[i].Card.ID == "deck-1-spades-07" {
			r.State.Cards[i].AvailableFromRound = 9
		}
	}
	after := invoke(t, r)
	if game.Digest(before.Frames[0].Observations[0]) != game.Digest(after.Frames[0].Observations[0]) {
		t.Fatal("private restriction leaked")
	}
}
func TestRejectsMalformedAndIncompatibleRequests(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte(`{}`), []byte(`{"schema":"bad"}`), []byte(`{"schema":"x","schema":"y"}`), []byte(`{"unknown":1}`), bytes.Repeat([]byte{' '}, canonical.MaxBytes+1)} {
		var out Response
		if e := canonical.Decode(Execute(raw), &out); e != nil || out.Error == "" || len(out.Frames) != 0 {
			t.Fatal("invalid input accepted")
		}
	}
	r := request(t, 3)
	r.Commands[0].GameID = "another-game"
	out := invoke(t, r)
	if out.Error != "command_rejected" || out.FailedStep != 0 || len(out.Frames) != 0 {
		t.Fatal("wrong-game input accepted or partial output returned")
	}
	r = request(t, 3)
	r.Commands = make([]game.Command, MaxCommands+1)
	if invoke(t, r).Error != "invalid_request" {
		t.Fatal("unbounded work accepted")
	}
	r = request(t, 3)
	r.State.Cards = r.State.Cards[:103]
	if invoke(t, r).Error != "invalid_state" {
		t.Fatal("invalid partition accepted")
	}
}

// Export only synthetic inputs/results to the runner's private ignored directory.
// This also runs normally without an output path; no tests are skipped.
func TestExportParityFixtures(t *testing.T) {
	for _, population := range []int{3, 4} {
		r := request(t, population)
		raw, e := canonical.Marshal(r)
		if e != nil {
			t.Fatal(e)
		}
		expected := Execute(raw)
		var out Response
		if e = canonical.Decode(expected, &out); e != nil || out.Error != "" {
			t.Fatal("bad export fixture")
		}
		dir := os.Getenv("CGMS_PROBE_FIXTURES")
		if dir == "" {
			continue
		}
		name := string(rune('0' + population))
		for suffix, data := range map[string][]byte{"request": raw, "expected": expected} {
			if e = os.WriteFile(filepath.Join(dir, name+"."+suffix+".json"), data, 0600); e != nil {
				t.Fatal(e)
			}
		}
	}
}
