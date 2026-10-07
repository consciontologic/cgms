package game

import (
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"os"
	"reflect"
	"testing"
)

func goldenPosition(t *testing.T) State {
	t.Helper()
	s := combatFixture(t)
	var e error
	s, e = s.Move([]string{cid(Spades, 7)}, 2, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Move([]string{cid(Diamonds, 1)}, 2, ConcealedAce, true)
	if e != nil {
		t.Fatal(e)
	}
	kings := []string{cid(Hearts, 13), cid(Clubs, 13)}
	s, e = s.Move(kings, 2, Formation, true)
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Bind(Binding{ID: "synthetic-binding", Kings: kings, Slot: "synthetic-slot", Formation: "people", Supported: false})
	if e != nil {
		t.Fatal(e)
	}
	s, e = s.Move(kings[1:], 3, Hand, true)
	if e != nil {
		t.Fatal(e)
	}
	// Inactive formation avoids claiming People protection, but retains the public
	// king and remote dormant binding. Its allocation is not a functioning combo.
	for i := range s.Cards {
		if s.Cards[i].Card.ID == kings[0] {
			s.Cards[i].Allocation = ""
		}
		if s.Cards[i].Card.ID == cid(Spades, 7) {
			s.Cards[i].AvailableFromRound = 3
		}
	}
	return s
}
func TestCanonicalStateEventGolden(t *testing.T) {
	raw, e := os.ReadFile("../../testdata/replay/synthetic-combat.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Provenance string   `json:"provenance"`
		Engine     string   `json:"engine"`
		Canonical  string   `json:"canonical"`
		States     []string `json:"state_hashes"`
		Events     []string `json:"event_hashes"`
	}
	if e = canonical.Decode(raw, &v); e != nil {
		t.Fatal(e)
	}
	if v.Engine != "cgms-engine-v1" || v.Canonical != canonical.Version {
		t.Fatal("golden compatibility")
	}
	s := goldenPosition(t)
	states := []string{Digest(s)}
	events := []string{}
	commands := []Command{{GameID: s.GameID, ID: "golden-attack", Kind: "attack", Actor: 1, Cards: []string{cid(Clubs, 8)}, Targets: []string{cid(Hearts, 8)}}, {GameID: s.GameID, ID: "golden-pass2", Kind: "pass", Actor: 2, WindowID: "golden-attack"}, {GameID: s.GameID, ID: "golden-pass3", Kind: "pass", Actor: 3, WindowID: "golden-attack"}}
	for _, c := range commands {
		raw, e := canonical.Marshal(s)
		if e != nil {
			t.Fatal(e)
		}
		var restored State
		if e = canonical.Decode(raw, &restored); e != nil {
			t.Fatal(e)
		}
		n, ev, e := Apply(s, c)
		if e != nil {
			t.Fatal(e)
		}
		rn, rev, e := Apply(restored, c)
		if e != nil || Digest(n) != Digest(rn) || Digest(ev) != Digest(rev) {
			t.Fatal("canonical continuation divergence")
		}
		if e = n.Validate(); e != nil {
			t.Fatal(e)
		}
		s = n
		states = append(states, Digest(s))
		events = append(events, Digest(ev))
	}
	target, _ := s.Card(cid(Hearts, 8))
	club, _ := s.Card(cid(Clubs, 8))
	if target.Zone != Draw || club.UsedTurn != 1 || club.Zone != Series || s.Pending != nil {
		t.Fatal("golden must satisfy independently asserted combat contract")
	}
	if !reflect.DeepEqual(states, v.States) || !reflect.DeepEqual(events, v.Events) {
		t.Fatalf("golden mismatch states=%q events=%q", states, events)
	}
}
func TestCanonicalObservationGolden(t *testing.T) {
	raw, e := os.ReadFile("../../testdata/observations/synthetic-combat.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Provenance string   `json:"provenance"`
		Canonical  string   `json:"canonical"`
		Hashes     []string `json:"seat_hashes"`
	}
	if e = canonical.Decode(raw, &v); e != nil {
		t.Fatal(e)
	}
	if v.Canonical != canonical.Version {
		t.Fatal("golden compatibility")
	}
	s := goldenPosition(t)
	hashes := []string{}
	for seat := 1; seat <= 3; seat++ {
		o, e := Observe(s, seat)
		if e != nil {
			t.Fatal(e)
		}
		hashes = append(hashes, Digest(o))
		if o.Players[1].HandCount != 1 || o.Players[1].AceCount != 1 || o.Players[2].HandCount != 1 {
			t.Fatal("public counts")
		}
		for _, c := range o.Cards {
			if c.Controller != seat && (c.Zone == Hand || c.Zone == ConcealedAce) {
				t.Fatal("concealed identity leaked")
			}
		}
	}
	// Change only hidden marker and a remote binding identity: seat 1 must receive
	// the identical authorized observation and legal menu.
	altered := s.Clone()
	for i := range altered.Cards {
		if altered.Cards[i].Card.ID == cid(Spades, 7) {
			altered.Cards[i].AvailableFromRound = 9
		}
	}
	altered.Bindings[0].ID = "different-private-link"
	before, _ := Observe(s, 1)
	after, _ := Observe(altered, 1)
	if Digest(before) != Digest(after) {
		t.Fatal("hidden marker/binding leak")
	}
	if !reflect.DeepEqual(hashes, v.Hashes) {
		t.Fatalf("golden mismatch seat_hashes=%q", hashes)
	}
}
