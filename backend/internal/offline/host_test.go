package offline

import (
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"testing"
)

func TestHostTutorialSurvivesReopenAndOnlyAcceptsHumanCommands(t *testing.T) {
	dir := t.TempDir()
	store, e := OpenStore(dir, "rules")
	if e != nil {
		t.Fatal(e)
	}
	host := NewHost(store)
	view, e := host.Handle(Request{Kind: "tutorial", Slot: "lesson", Population: 3, Games: 1, Difficulty: "beginner"})
	if e != nil {
		t.Fatal(e)
	}
	if view.Instruction == "" {
		t.Fatal("missing beginner instruction")
	}
	for ticks := 0; ticks < 100 && view.TutorialStep < 3; ticks++ {
		if view.TutorialStep == 0 || view.TutorialStep == 2 {
			var chosen matchstore.Request
			for _, a := range view.Actions {
				if view.TutorialStep == 0 && a.Command != nil && a.Command.Kind == "open-series" || view.TutorialStep == 2 && a.Operation != nil && a.Operation.Kind == "end-turn" {
					chosen = a
					break
				}
			}
			if chosen.Command == nil && chosen.Operation == nil {
				t.Fatalf("lesson has no required legal action at %d", view.TutorialStep)
			}
			view, e = host.Handle(Request{Kind: "act", Slot: "lesson", Version: view.Version, Action: &chosen})
		} else {
			view, e = host.Handle(Request{Kind: "step", Slot: "lesson", Version: view.Version})
		}
		if e != nil {
			t.Fatal(e)
		}
		store.Close()
		store, e = OpenStore(dir, "rules")
		if e != nil {
			t.Fatal(e)
		}
		host = NewHost(store)
		resumed, err := host.Handle(Request{Kind: "load", Slot: "lesson"})
		if err != nil {
			t.Fatal(err)
		}
		if resumed.Version != view.Version || resumed.TutorialStep != view.TutorialStep {
			t.Fatal("lost tutorial progress")
		}
		view = resumed
	}
	defer store.Close()
	if view.TutorialStep != 3 {
		t.Fatal("tutorial did not complete")
	}
	if _, e = host.Handle(Request{Kind: "step", Slot: "lesson", Version: view.Version - 1}); e == nil {
		t.Fatal("stale version accepted")
	}
}
