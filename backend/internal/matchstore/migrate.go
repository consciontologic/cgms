package matchstore

import (
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrations embed.FS

// Up installs the durable match schema atomically. Repeated calls verify the
// installed migration identity; concurrent callers serialize in PostgreSQL.
func Up(ctx context.Context, pool *pgxpool.Pool) error { return migrate(ctx, pool, true) }

// Down removes this schema and its data atomically. Callers must explicitly opt
// into this destructive rollback; normal application startup only invokes Up.
func Down(ctx context.Context, pool *pgxpool.Pool) error { return migrate(ctx, pool, false) }

func migrate(ctx context.Context, pool *pgxpool.Pool, up bool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(cleanup)
	}()
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(734982613)`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS matchstore_migrations(version integer PRIMARY KEY,checksum bytea NOT NULL)`); err != nil {
		return err
	}

	names := []string{"001_matches", "002_reputation"}
	bodies := make([][]byte, len(names))
	digests := make([][32]byte, len(names))
	installed := map[int]bool{}
	for i, name := range names {
		b, e := migrations.ReadFile("migrations/" + name + ".up.sql")
		if e != nil {
			return e
		}
		bodies[i] = b
		digests[i] = sha256.Sum256(b)
	}
	rows, err := tx.Query(ctx, `SELECT version,checksum FROM matchstore_migrations ORDER BY version`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var version int
		var sum []byte
		if err = rows.Scan(&version, &sum); err != nil {
			rows.Close()
			return err
		}
		if version < 1 || version > len(names) || string(sum) != string(digests[version-1][:]) {
			rows.Close()
			return fmt.Errorf("unsupported matchstore migration identity")
		}
		installed[version] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for i := range names {
		if installed[i+1] && i > 0 && !installed[i] {
			return fmt.Errorf("noncontiguous matchstore migrations")
		}
	}
	if up {
		for i := range names {
			if installed[i+1] {
				continue
			}
			if _, err = tx.Exec(ctx, string(bodies[i])); err != nil {
				return err
			}
			if _, err = tx.Exec(ctx, `INSERT INTO matchstore_migrations(version,checksum) VALUES($1,$2)`, i+1, digests[i][:]); err != nil {
				return err
			}
		}
	} else {
		for i := len(names) - 1; i >= 0; i-- {
			if !installed[i+1] {
				continue
			}
			b, e := migrations.ReadFile("migrations/" + names[i] + ".down.sql")
			if e != nil {
				return e
			}
			if _, err = tx.Exec(ctx, string(b)); err != nil {
				return err
			}
			if _, err = tx.Exec(ctx, `DELETE FROM matchstore_migrations WHERE version=$1`, i+1); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}
