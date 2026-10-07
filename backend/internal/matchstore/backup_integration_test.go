//go:build integration

package matchstore

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestLogicalBackupRestoresPinnedFinancialStateAndReceipts(t *testing.T) {
	ctx := context.Background()
	s := testStore(t)
	b, err := game.NewState(3, "g1")
	if err != nil {
		t.Fatal(err)
	}
	var cards []string
	for _, suit := range []string{"diamonds", "hearts"} {
		for copy := 1; copy <= 2; copy++ {
			cards = append(cards, fmt.Sprintf("deck-%d-%s-12", copy, suit))
		}
	}
	b, err = b.Move(cards, 1, game.Hand, true)
	if err != nil {
		t.Fatal(err)
	}
	b.Players[0].History = []game.Suit{game.Clubs}
	m, err := game.NewMatchLifecycle(b, 3)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	amount, err := game.NewAmount("1", "99999999999999999999999999999999999999999999999999999999999999999999991")
	if err != nil {
		t.Fatal(err)
	}
	m.Game.Ledger, err = m.Game.Ledger.Charge("backup-charge", 0, 1, amount)
	if err != nil {
		t.Fatal(err)
	}
	m.StartLedger = m.Game.Ledger.Clone()
	envelope, err := NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Create(ctx, "m", envelope, []string{"a", "b", "c"}); err != nil {
		t.Fatal(err)
	}
	intent := cardIntent("backup-coup", 0, game.Command{GameID: "g1", Actor: 1, Kind: "coup"})
	original, err := s.Submit(ctx, "m", "a", intent)
	if err != nil {
		t.Fatal(err)
	}
	restored := logicalRestoreFixture(t, s)
	before, version, err := s.Restore(ctx, "m")
	if err != nil {
		t.Fatal(err)
	}
	after, restoredVersion, err := restored.Restore(ctx, "m")
	if err != nil || version != restoredVersion || game.Digest(before) != game.Digest(after) {
		t.Fatal("snapshot or pinned configuration changed", err)
	}
	duplicate, err := restored.Submit(ctx, "m", "a", intent)
	if err != nil || game.Digest(original) != game.Digest(duplicate) {
		t.Fatal("receipt retry changed after restore", err)
	}
}

func logicalRestoreFixture(t *testing.T, source *Store) *Store {
	t.Helper()
	ctx := context.Background()
	container := os.Getenv("CGMS_TEST_POSTGRES_CONTAINER")
	if container == "" {
		t.Fatal("CGMS_TEST_POSTGRES_CONTAINER required")
	}
	var schema string
	if err := source.pool.QueryRow(ctx, "SELECT current_schema()").Scan(&schema); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	dump, err := exec.CommandContext(ctx, "docker", "exec", container, "pg_dump", "-U", "cgms", "-d", "cgms", "--format=custom", "--schema="+schema).Output()
	if err != nil {
		t.Fatal("pg_dump failed", err)
	}
	database := fmt.Sprintf("restore_%d", time.Now().UnixNano())
	if err = exec.CommandContext(ctx, "docker", "exec", container, "createdb", "-U", "cgms", database).Run(); err != nil {
		t.Fatal("createdb failed", err)
	}
	t.Cleanup(func() {
		if e := exec.Command("docker", "exec", container, "dropdb", "-U", "cgms", database).Run(); e != nil {
			t.Error("restore database cleanup failed", e)
		}
	})
	command := exec.CommandContext(ctx, "docker", "exec", "-i", container, "pg_restore", "-U", "cgms", "-d", database, "--exit-on-error", "--single-transaction", "--no-owner", "--no-privileges")
	command.Stdin = bytes.NewReader(dump)
	if err = command.Run(); err != nil {
		t.Fatal("pg_restore failed", err)
	}
	config := source.pool.Config().Copy()
	config.ConnConfig.Database = database
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	rows, err := source.pool.Query(ctx, "SELECT table_name FROM information_schema.tables WHERE table_schema=$1 ORDER BY table_name", schema)
	if err != nil {
		t.Fatal(err)
	}
	var tables []string
	for rows.Next() {
		var table string
		if err = rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, table)
	}
	rows.Close()
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}
	for _, table := range tables {
		query := "SELECT coalesce(jsonb_agg(t ORDER BY t::text),'[]')::text FROM (SELECT to_jsonb(r) AS t FROM " + pgx.Identifier{schema, table}.Sanitize() + " AS r) q"
		var original, restored string
		if err = source.pool.QueryRow(ctx, query).Scan(&original); err != nil {
			t.Fatal(err)
		}
		if err = pool.QueryRow(ctx, query).Scan(&restored); err != nil {
			t.Fatal(err)
		}
		if original != restored {
			t.Fatal("restored table differs", table)
		}
	}
	for _, table := range []string{"financial_debts", "command_results", "games", "seat_projections"} {
		var count int
		if err = pool.QueryRow(ctx, "SELECT count(*) FROM "+pgx.Identifier{table}.Sanitize()).Scan(&count); err != nil || count == 0 {
			t.Fatal("nonempty fixture required", table, err)
		}
	}
	t.Logf("logical backup+restore+all-table comparison: %s; archive_bytes=%d tables=%d", time.Since(start), len(dump), len(tables))
	return New(pool, "rules-test")
}
