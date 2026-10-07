package offline

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDurableSaveRejectsCorruptionAndKeepsPrior(t *testing.T) {
	dir := t.TempDir()
	store, err := OpenStore(dir, "test-rules")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	session, err := NewSession(3, 1, 1, "beginner", "test-rules")
	if err != nil {
		t.Fatal(err)
	}
	if err = store.Save("game-1", session); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "game-1.json")
	prior, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load("game-1")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != session.Version || loaded.Engine.Match.Game.Board.GameID != session.Engine.Match.Game.Board.GameID {
		t.Fatal("changed restored identity")
	}
	bad := *session
	bad.Schema = "future"
	if store.Save("game-1", &bad) == nil {
		t.Fatal("bad schema accepted")
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(prior, after) {
		t.Fatal("destroyed prior save")
	}
	for _, slot := range []string{"../escape", "/absolute", "a/b", "", ".hidden"} {
		if store.Save(slot, session) == nil {
			t.Fatalf("accepted path %q", slot)
		}
	}
	if err = os.WriteFile(path, bytes.Replace(prior, []byte("beginner"), []byte("advanced"), 1), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.Load("game-1"); err == nil {
		t.Fatal("checksum corruption accepted")
	}
	if err = os.WriteFile(path, prior, 0600); err != nil {
		t.Fatal(err)
	}
	store.Close()
	other, err := OpenStore(dir, "different-rules")
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if _, err = other.Load("game-1"); err == nil {
		t.Fatal("incompatible rules accepted")
	}
}

func TestSaveFailureBeforeRenamePreservesPriorAndBoundsReads(t *testing.T) {
	dir := t.TempDir()
	store, e := OpenStore(dir, "rules")
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	s, _ := NewSession(3, 1, 1, "beginner", "rules")
	if e = store.Save("one", s); e != nil {
		t.Fatal(e)
	}
	before, _ := os.ReadFile(filepath.Join(dir, "one.json"))
	store.rename = func(string, string) error { return os.ErrPermission }
	s.Version++
	if store.Save("one", s) == nil {
		t.Fatal("injected failure accepted")
	}
	after, _ := os.ReadFile(filepath.Join(dir, "one.json"))
	if !bytes.Equal(before, after) {
		t.Fatal("prior save replaced on failed write")
	}
	f, e := os.Create(filepath.Join(dir, "large.json"))
	if e != nil {
		t.Fatal(e)
	}
	e = f.Truncate(MaxSaveBytes + 1)
	f.Close()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = store.Load("large"); e == nil {
		t.Fatal("oversize accepted")
	}
	outside := filepath.Join(t.TempDir(), "foreign.json")
	if e = os.WriteFile(outside, before, 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.Symlink(outside, filepath.Join(dir, "linked.json")); e != nil {
		t.Fatal(e)
	}
	if _, e = store.Load("linked"); e == nil {
		t.Fatal("symlink escape accepted")
	}
	s.Engine.Schema = "cgms-match-envelope-economy-v1"
	if s.Validate("rules") == nil {
		t.Fatal("online economy envelope imported")
	}
}

func TestPostRenameSyncFailureIsExplicitlyUncertain(t *testing.T) {
	store, e := OpenStore(t.TempDir(), "rules")
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	s, _ := NewSession(3, 1, 1, "beginner", "rules")
	if e = store.Save("one", s); e != nil {
		t.Fatal(e)
	}
	store.syncDirectory = func() error { return os.ErrPermission }
	s.Version++
	if e = store.Save("one", s); e != ErrDurabilityUnknown {
		t.Fatalf("expected uncertain durable outcome, got %v", e)
	}
	restored, e := store.Load("one")
	if e != nil || restored.Version != s.Version {
		t.Fatal("must reconcile committed rename")
	}
}

func TestStoreRejectsConcurrentOwner(t *testing.T) {
	dir := t.TempDir()
	first, e := OpenStore(dir, "rules")
	if e != nil {
		t.Fatal(e)
	}
	second, e := OpenStore(dir, "rules")
	if e == nil {
		second.Close()
		first.Close()
		t.Fatal("concurrent owner accepted")
	}
	first.Close()
	third, e := OpenStore(dir, "rules")
	if e != nil {
		t.Fatal(e)
	}
	third.Close()
}
