package sim

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestRetentionPreviewProtectsEvidence(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	makeRun := func(name string, expiry time.Time, hold bool, parent string) string {
		t.Helper()
		dir := filepath.Join(root, name)
		if e := os.Mkdir(dir, 0700); e != nil {
			t.Fatal(e)
		}
		m := manifestBase(nil, "run", "seed", "config", 1, 1, nil)
		m.Expires = expiry.Format(time.RFC3339)
		m.ReviewHold = hold
		m.Parent = parent
		if name == "compare" {
			m.Command = "compare"
		}
		if e := publish(dir, m, nil); e != nil {
			t.Fatal(e)
		}
		return filepath.Join(dir, "manifest.json")
	}
	old := now.Add(-time.Hour)
	a := makeRun("a", old, false, "")
	b := makeRun("b", old, false, "")
	makeRun("held", old, true, "")
	makeRun("future", now.Add(time.Hour), false, "")
	makeRun("equal", now, false, "")
	makeRun("compare", now.Add(time.Hour), false, a+" | "+b)
	expired := makeRun("expired", old, false, "")
	if e := os.Symlink(filepath.Dir(a), filepath.Join(root, "linked")); e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadFile(a)
	if e != nil {
		t.Fatal(e)
	}
	got, e := RetentionPreview(root, now)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(got, []string{filepath.Dir(expired)}) {
		t.Fatalf("preview=%v", got)
	}
	after, e := os.ReadFile(a)
	if e != nil || string(after) != string(before) {
		t.Fatal("manifest modified", e)
	}
}
func TestRetentionPreviewAmbiguousParentFailsClosed(t *testing.T) {
	root := t.TempDir()
	m := manifestBase(nil, "compare", "", "", 1, 1, nil)
	m.Parent = "a | b | c"
	m.Expires = time.Now().Add(-time.Hour).Format(time.RFC3339)
	if e := publish(root, m, nil); e != nil {
		t.Fatal(e)
	}
	if got, e := RetentionPreview(root, time.Now()); e == nil || len(got) != 0 {
		t.Fatalf("ambiguous reference produced preview %v %v", got, e)
	}
}
