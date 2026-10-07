//go:build integration

package matchstore

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func migrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("CGMS_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("CGMS_TEST_DATABASE_URL is required for integration tests")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("migration_%d", time.Now().UnixNano())
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, e := admin.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE")
		admin.Close()
		if e != nil {
			t.Error(e)
		}
	})
	return pool
}

func TestMigrationsRoundTripAndAtomicConstraints(t *testing.T) {
	ctx := context.Background()
	p := migrationPool(t)
	if err := Up(ctx, p); err != nil {
		t.Fatal(err)
	}
	if err := Up(ctx, p); err != nil {
		t.Fatal("repeat up", err)
	}
	_, err := p.Exec(ctx, `INSERT INTO matches VALUES ('m',decode('00','hex'),0,'g',decode('01','hex'),decode('02','hex'),decode('03','hex'));
 INSERT INTO members VALUES ('m','alice',1),('m','bob',2);
 INSERT INTO command_results VALUES ('m','alice','c',decode('aa','hex'),decode('bb','hex'));`)
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO members VALUES ('m','charlie',1)`,
		`INSERT INTO members VALUES ('m','charlie',5)`,
		`INSERT INTO command_results VALUES ('m','alice','c',decode('ab','hex'),decode('cc','hex'))`,
		`INSERT INTO command_results VALUES ('m','outsider','c',decode('ab','hex'),decode('cc','hex'))`,
		`INSERT INTO matches VALUES ('m2',decode('00','hex'),0,'g',decode('01','hex'),decode('02','hex'),decode('03','hex'))`,
	} {
		if _, e := p.Exec(ctx, q); e == nil {
			t.Fatal("constraint accepted invalid mutation")
		}
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `UPDATE matches SET state_version=1 WHERE match_id='m'; INSERT INTO match_events VALUES('m',1,'g',decode('aa','hex')); INSERT INTO seat_events VALUES('m',1,1,decode('bb','hex'))`); err != nil {
		t.Fatal(err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var version, events int
	if err = p.QueryRow(ctx, `SELECT state_version FROM matches WHERE match_id='m'`).Scan(&version); err != nil || version != 0 {
		t.Fatal(version, err)
	}
	if err = p.QueryRow(ctx, `SELECT count(*) FROM match_events`).Scan(&events); err != nil || events != 0 {
		t.Fatal(events, err)
	}
	if err = Down(ctx, p); err != nil {
		t.Fatal(err)
	}
	var table *string
	if err = p.QueryRow(ctx, `SELECT to_regclass('matches')::text`).Scan(&table); err != nil || table != nil {
		t.Fatal(table, err)
	}
	if err = Down(ctx, p); err != nil {
		t.Fatal("repeat down", err)
	}
	if err = Up(ctx, p); err != nil {
		t.Fatal("up after down", err)
	}
}

func TestMigrationRejectsDriftAndRollsBackDDL(t *testing.T) {
	ctx := context.Background()
	p := migrationPool(t)
	if _, err := p.Exec(ctx, `CREATE TABLE members(collision integer)`); err != nil {
		t.Fatal(err)
	}
	if err := Up(ctx, p); err == nil {
		t.Fatal("conflicting schema accepted")
	}
	var table *string
	if err := p.QueryRow(ctx, `SELECT to_regclass('matches')::text`).Scan(&table); err != nil || table != nil {
		t.Fatal("failed migration retained partial DDL", table, err)
	}
	if _, err := p.Exec(ctx, `DROP TABLE members`); err != nil {
		t.Fatal(err)
	}
	if err := Up(ctx, p); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, `UPDATE matchstore_migrations SET checksum=decode('00','hex')`); err != nil {
		t.Fatal(err)
	}
	if err := Up(ctx, p); err == nil {
		t.Fatal("changed migration admitted")
	}
	if err := Down(ctx, p); err == nil {
		t.Fatal("unknown migration destroyed")
	}
	if err := p.QueryRow(ctx, `SELECT to_regclass('matches')::text`).Scan(&table); err != nil || table == nil {
		t.Fatal("failed rollback removed tables", err)
	}
}

func TestMigrationConcurrentUp(t *testing.T) {
	p := migrationPool(t)
	ctx := context.Background()
	out := make(chan error, 2)
	go func() { out <- Up(ctx, p) }()
	go func() { out <- Up(ctx, p) }()
	for range 2 {
		if err := <-out; err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := p.QueryRow(ctx, `SELECT count(*) FROM matchstore_migrations`).Scan(&count); err != nil || count != 2 {
		t.Fatal("migration applied more than once", count, err)
	}
}
