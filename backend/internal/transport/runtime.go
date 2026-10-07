package transport

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"log/slog"
	"sync"
	"time"
)

// PersistentKey is internal credential material. Back up this table with the
// authoritative database; rotation invalidates outstanding card capabilities.
func PersistentKey(ctx context.Context, p *pgxpool.Pool) ([]byte, error) {
	tx, e := p.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(c)
	}()
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(734982619)"); e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS online_settings(name text PRIMARY KEY,value bytea NOT NULL CHECK(octet_length(value)=32))"); e != nil {
		return nil, e
	}
	key := make([]byte, 32)
	if _, e = rand.Read(key); e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO online_settings(name,value) VALUES('handle-key-v1',$1) ON CONFLICT DO NOTHING", key); e != nil {
		return nil, e
	}
	if e = tx.QueryRow(ctx, "SELECT value FROM online_settings WHERE name='handle-key-v1'").Scan(&key); e != nil {
		return nil, e
	}
	return key, tx.Commit(ctx)
}

// Authority prevents accidental concurrent local server startup. It is not a
// distributed fencing lease: supported operation is one process, no hot standby.
type Authority struct {
	conn *pgxpool.Conn
	once sync.Once
}

func AcquireAuthority(ctx context.Context, p *pgxpool.Pool) (*Authority, error) {
	c, e := p.Acquire(ctx)
	if e != nil {
		return nil, e
	}
	var ok bool
	e = c.QueryRow(ctx, "SELECT pg_try_advisory_lock(734982620)").Scan(&ok)
	if e != nil || !ok {
		c.Release()
		if e == nil {
			e = errors.New("online authority already running")
		}
		return nil, e
	}
	return &Authority{conn: c}, nil
}
func (a *Authority) Ping(ctx context.Context) error { return a.conn.Conn().Ping(ctx) }
func (a *Authority) Close() {
	a.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, _ = a.conn.Exec(ctx, "SELECT pg_advisory_unlock(734982620)")
		_ = a.conn.Conn().Close(ctx)
		a.conn.Release()
	})
}

// automaticWaits belongs only to the scheduler goroutine. Expected allowance
// denials preserve the finalized game and avoid retrying a write every scan.
// The cache is an optimization: eviction/restart merely rechecks admission.
type automaticWaits map[string]time.Time

func (w automaticWaits) pending(id string, now time.Time) bool {
	until, ok := w[id]
	if ok && !now.Before(until) {
		delete(w, id)
		return false
	}
	return ok
}

func (w automaticWaits) block(id string, now time.Time) {
	if len(w) >= 1024 {
		oldest := ""
		for key, until := range w {
			if oldest == "" || until.Before(w[oldest]) {
				oldest = key
			}
		}
		delete(w, oldest)
	}
	until := now.Add(30 * time.Second)
	utc := now.UTC()
	reset := time.Date(utc.Year(), utc.Month(), utc.Day()+1, 0, 0, 0, 0, time.UTC)
	if reset.Before(until) {
		until = reset
	}
	w[id] = until
}

// automaticIdle is owned only by the scheduler goroutine. A successful no-op
// remains idle until durable state changes. Its bounded FIFO is an optimization:
// eviction and process restart only cause an extra check, never lost work.
type automaticIdle struct {
	versions map[string]int64
	order    [1024]string
	next     int
}

func (i *automaticIdle) unchanged(id string, version int64) bool {
	checked, ok := i.versions[id]
	return ok && checked == version
}

func (i *automaticIdle) remember(id string, version int64) {
	if i.versions == nil {
		i.versions = make(map[string]int64)
	}
	if _, ok := i.versions[id]; !ok {
		delete(i.versions, i.order[i.next])
		i.order[i.next] = id
		i.next = (i.next + 1) % len(i.order)
	}
	i.versions[id] = version
}

// Entitlement handlers signal a prompt rescan after a committed redemption.
// No account identity or entitlement data enters this process-owned queue.
func (s *Server) wakeAutomatic() {
	select {
	case s.automaticWake <- struct{}{}:
	default:
	}
}

// Start owns one fair, paginated automatic-transition loop. It never substitutes
// for human actions and is restartable entirely from persisted match state.
func (s *Server) Start() {
	s.lifecycleMu.Lock()
	defer s.lifecycleMu.Unlock()
	if s.closed {
		return
	}
	s.startOnce.Do(func() {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()
			after := ""
			waits := automaticWaits{}
			idle := automaticIdle{}
			for {
				select {
				case <-s.ctx.Done():
					return
				case <-ticker.C:
				case <-s.automaticWake:
					clear(waits)
					idle = automaticIdle{}
					after = ""
				}
				ctx, cancel := context.WithTimeout(s.ctx, 2*time.Second)
				rows, e := s.cfg.Pool.Query(ctx, "SELECT match_id,state_version FROM matches WHERE match_id>$1 ORDER BY match_id LIMIT 32", after)
				type matchVersion struct {
					id      string
					version int64
				}
				matches := []matchVersion{}
				if e == nil {
					for rows.Next() {
						var match matchVersion
						if e = rows.Scan(&match.id, &match.version); e != nil {
							break
						}
						matches = append(matches, match)
					}
					if e == nil {
						e = rows.Err()
					}
					rows.Close()
				}
				cancel()
				if e != nil {
					if s.ctx.Err() == nil {
						slog.Warn("automatic scan unavailable")
					}
					continue
				}
				if len(matches) < 32 {
					after = ""
				} else {
					after = matches[len(matches)-1].id
				}
				for _, match := range matches {
					id := match.id
					if idle.unchanged(id, match.version) || waits.pending(id, time.Now()) {
						continue
					}
					ctx, cancel := context.WithTimeout(s.ctx, 2*time.Second)
					result, e := s.dispatcher.Submit(ctx, id, 100, func(ctx context.Context) (any, error) {
						return s.automaticStep(ctx, id)
					})
					cancel()
					if e == nil && result == nil {
						// Cache only the scanned version: a command racing this check
						// leaves a newer durable version eligible on the next scan.
						idle.remember(id, match.version)
					}
					if errors.Is(e, economy.ErrAllowance) {
						waits.block(id, time.Now())
					} else if e != nil && s.ctx.Err() == nil {
						slog.Warn("automatic transition deferred", "reason", automaticFailureReason(e))
					}
					if s.ctx.Err() != nil {
						return
					}
				}
			}
		}()
	})
}

// Report a bounded category, never an error string that may contain private
// commands, database details or account identifiers.
func automaticFailureReason(err error) string {
	if err == nil {
		return "none"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	return "transition"
}
