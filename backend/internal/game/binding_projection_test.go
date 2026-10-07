package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

func projectedOwnBindings(t *testing.T, s State, seat int) []FormationSubstitution {
	t.Helper()
	o, err := ObserveWithoutMenu(s, seat)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	var view struct {
		Bindings []FormationSubstitution `json:"own_bindings"`
	}
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	return view.Bindings
}

func TestOwnBindingProjectionSurvivesTakeBack(t *testing.T) {
	s := formationFixture(t)
	ids := []string{"deck-1-hearts-12", "deck-1-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, false)
	spec := FormationSpec{Kind: "people", Cards: ids, Substitute: &FormationSubstitution{Kings: ids[1:], Slot: FormationSlot{12, Hearts}}, Protection: &FormationProtection{Series: Clubs}}
	var err error
	s, err = OpenFormation(s, 1, "people-1", spec)
	if err != nil {
		t.Fatal(err)
	}
	s, err = TakeBackFormation(s, 1, "people-1")
	if err != nil {
		t.Fatal(err)
	}
	// Existing authority rejects the very attachment the old client suggested.
	if _, err = AttachRoyal(s, 1, ids[1], Hearts); err == nil {
		t.Fatal("dormant bound king attached")
	}
	want := []FormationSubstitution{{Kings: ids[1:], Slot: FormationSlot{12, Hearts}}}
	if got := projectedOwnBindings(t, s, 1); !reflect.DeepEqual(got, want) {
		t.Fatalf("own dormant binding missing: got %#v, want %#v", got, want)
	}
	if got := projectedOwnBindings(t, s, 2); len(got) != 0 {
		t.Fatalf("opponent's concealed binding disclosed: %#v", got)
	}
	// A separated pair reveals only the physical king currently controlled by
	// this viewer, never the counterpart's concealed identity or location.
	for i := range s.Cards {
		if s.Cards[i].Card.ID == ids[2] {
			s.Cards[i].Controller = 2
		}
	}
	for seat, king := range map[int]string{1: ids[1], 2: ids[2]} {
		got := projectedOwnBindings(t, s, seat)
		if len(got) != 1 || !reflect.DeepEqual(got[0].Kings, []string{king}) || got[0].Slot != want[0].Slot {
			t.Fatalf("seat %d saw a concealed counterpart: %#v", seat, got)
		}
	}
	if got := projectedOwnBindings(t, s, 3); len(got) != 0 {
		t.Fatalf("uninvolved viewer received bindings: %#v", got)
	}
}
