//go:build integration && rollback

package matchstore

import (
	"context"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestPinnedRollbackReaders(t *testing.T) {
	binary := os.Getenv("CGMS_ROLLBACK_READER")
	if binary == "" {
		t.Fatal("run xops/rollback_rehearsal.py to build archived reader")
	}
	s := testStore(t)
	fixture(t, s)
	ctx := context.Background()
	if _, err := s.Submit(ctx, "m", "a", nullify("rollback-receipt", 1)); err != nil {
		t.Fatal(err)
	}
	before, version, err := s.Restore(ctx, "m")
	if err != nil {
		t.Fatal(err)
	}
	var schema string
	if err = s.pool.QueryRow(ctx, "SELECT current_schema()").Scan(&schema); err != nil {
		t.Fatal(err)
	}
	address, err := url.Parse(os.Getenv("CGMS_TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	query := address.Query()
	query.Set("search_path", schema)
	address.RawQuery = query.Encode()
	run := func() error {
		bounded, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		cmd := exec.CommandContext(bounded, binary)
		cmd.Env = append(os.Environ(), "CGMS_ROLLBACK_DATABASE_URL="+address.String())
		cmd.Stdin = strings.NewReader(game.Digest(before))
		return cmd.Run()
	}
	if err = run(); err != nil {
		t.Fatal("archived reader cannot restore compatible active match", err)
	}
	if err = Up(ctx, s.pool); err != nil {
		t.Fatal("new reader migration replay", err)
	}
	after, v, err := s.Restore(ctx, "m")
	if err != nil || v != version || game.Digest(before) != game.Digest(after) {
		t.Fatal("new reader changed pinned state", err)
	}
	// Unknown migration identity must reject both binaries, without changing the
	// active checkpoint. Remove only the deliberately injected fixture identity.
	if _, err = s.pool.Exec(ctx, "INSERT INTO matchstore_migrations(version,checksum) VALUES(999,'')"); err != nil {
		t.Fatal(err)
	}
	if err = run(); err == nil {
		t.Fatal("archived reader reinterpreted unsupported migration")
	}
	if err = Up(ctx, s.pool); err == nil {
		t.Fatal("new reader reinterpreted unsupported migration")
	}
	if _, err = s.pool.Exec(ctx, "DELETE FROM matchstore_migrations WHERE version=999"); err != nil {
		t.Fatal(err)
	}
	if err = run(); err != nil {
		t.Fatal("archived reader failed after unsupported identity removed", err)
	}
	var original []byte
	if err = s.pool.QueryRow(ctx, "SELECT snapshot FROM matches WHERE match_id='m'").Scan(&original); err != nil {
		t.Fatal(err)
	}
	incompatible := before.Clone()
	incompatible.EngineVersion = "unsupported-next-engine"
	bad, err := encode(incompatible)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.pool.Exec(ctx, "UPDATE matches SET snapshot=$1 WHERE match_id='m'", bad); err != nil {
		t.Fatal(err)
	}
	if err = run(); err == nil {
		t.Fatal("archived reader reinterpreted unsupported engine")
	}
	if _, _, err = s.Restore(ctx, "m"); err == nil {
		t.Fatal("new reader reinterpreted unsupported engine")
	}
	if _, err = s.pool.Exec(ctx, "UPDATE matches SET snapshot=$1 WHERE match_id='m'", original); err != nil {
		t.Fatal(err)
	}
	if err = run(); err != nil {
		t.Fatal("original pinned reader unavailable after rollback", err)
	}
	t.Log("old/new readers preserve active pinned checkpoint and reject unsupported migration and engine; no schema downgrade or financial history deletion")
}

func TestEconomyRollbackReaders(t *testing.T) {
	binary := os.Getenv("CGMS_ROLLBACK_READER")
	if binary == "" {
		t.Fatal("run xops/rollback_rehearsal.py")
	}
	ctx := context.Background()
	base := testStore(t)
	if e := identity.Up(ctx, base.pool); e != nil {
		t.Fatal(e)
	}
	if e := economy.Up(ctx, base.pool); e != nil {
		t.Fatal(e)
	}
	s, e := NewWithEconomy(base.pool, "rules-test", economy.Policy{Reward: 7, PassDayPrice: 21})
	if e != nil {
		t.Fatal(e)
	}
	var actors []string
	for range 3 {
		v, e := identity.New(base.pool).CreateGuest(ctx)
		if e != nil {
			t.Fatal(e)
		}
		actors = append(actors, v.Account.ID)
	}
	b, e := game.NewState(3, "economic-game")
	if e != nil {
		t.Fatal(e)
	}
	m, e := game.NewMatchLifecycle(b, 3)
	if e != nil {
		t.Fatal(e)
	}
	env, e := NewEnvelope(m, "rules-test")
	if e != nil {
		t.Fatal(e)
	}
	if e = s.Create(ctx, "m", env, actors); e != nil {
		t.Fatal(e)
	}
	before, version, e := base.Restore(ctx, "m")
	if e != nil || before.Schema != EconomyEnvelopeSchema {
		t.Fatal("current reader", e)
	}
	var schema string
	if e = s.pool.QueryRow(ctx, "SELECT current_schema()").Scan(&schema); e != nil {
		t.Fatal(e)
	}
	address, e := url.Parse(os.Getenv("CGMS_TEST_DATABASE_URL"))
	if e != nil {
		t.Fatal(e)
	}
	query := address.Query()
	query.Set("search_path", schema)
	address.RawQuery = query.Encode()
	bounded, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(bounded, binary)
	cmd.Env = append(os.Environ(), "CGMS_ROLLBACK_DATABASE_URL="+address.String())
	cmd.Stdin = strings.NewReader(game.Digest(before))
	out, e := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 3 || !strings.Contains(string(out), "unsupported checkpoint") {
		t.Fatal("old reader must reject checkpoint rather than mismatch or bypass", e, string(out))
	}
	after, v, e := base.Restore(ctx, "m")
	if e != nil || v != version || game.Digest(after) != game.Digest(before) {
		t.Fatal("rejected rollback changed state", e)
	}
	t.Log("archived reader rejects economy checkpoint before reinterpretation; current reader preserves pinned state and policy")
}
