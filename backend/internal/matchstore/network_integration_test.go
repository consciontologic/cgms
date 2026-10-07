//go:build integration

package matchstore

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// commitProxy is a test-only, plaintext loopback wire proxy. It forwards startup,
// authentication and all queries unchanged; one selected COMMIT loses its actual
// network connection. No SQL, authentication payload or connection URL is logged.
func commitProxy(t *testing.T, upstream string, afterCommit bool) (string, *atomic.Bool) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	var dropped atomic.Bool
	var workers sync.WaitGroup
	var mu sync.Mutex
	peers := map[net.Conn]bool{}
	track := func(c net.Conn) { mu.Lock(); peers[c] = true; mu.Unlock() }
	untrack := func(c net.Conn) { c.Close(); mu.Lock(); delete(peers, c); mu.Unlock() }
	readFrame := func(c net.Conn, typed bool) ([]byte, error) {
		header := make([]byte, 4)
		if typed {
			header = make([]byte, 5)
		}
		if _, e := io.ReadFull(c, header); e != nil {
			return nil, e
		}
		offset := 0
		if typed {
			offset = 1
		}
		n := int(binary.BigEndian.Uint32(header[offset:]))
		if n < 4 || n > 70<<20 {
			return nil, errors.New("invalid PostgreSQL test frame")
		}
		frame := append(header, make([]byte, n-4)...)
		_, e := io.ReadFull(c, frame[len(header):])
		return frame, e
	}
	send := func(c net.Conn, frame []byte) error { _, e := io.Copy(c, bytes.NewReader(frame)); return e }
	workers.Add(1)
	go func() {
		defer workers.Done()
		for {
			client, e := listener.Accept()
			if e != nil {
				return
			}
			track(client)
			server, e := net.DialTimeout("tcp", upstream, 5*time.Second)
			if e != nil {
				untrack(client)
				continue
			}
			track(server)
			workers.Add(1)
			go func() {
				defer workers.Done()
				defer untrack(client)
				defer untrack(server)
				_ = client.SetDeadline(time.Now().Add(30 * time.Second))
				_ = server.SetDeadline(time.Now().Add(30 * time.Second))
				startup, e := readFrame(client, false)
				if e != nil || send(server, startup) != nil {
					return
				}
				done := make(chan struct{})
				go func() {
					defer close(done)
					defer client.Close()
					defer server.Close()
					for {
						frame, e := readFrame(server, true)
						if e != nil {
							return
						}
						isCommit := frame[0] == 'C' && string(frame[5:]) == "COMMIT\x00"
						if afterCommit && isCommit && dropped.CompareAndSwap(false, true) {
							return
						}
						if send(client, frame) != nil {
							return
						}
					}
				}()
				for {
					frame, e := readFrame(client, true)
					if e != nil {
						break
					}
					isCommit := frame[0] == 'Q' && strings.EqualFold(strings.TrimSuffix(string(frame[5:]), "\x00"), "commit")
					if !afterCommit && isCommit && dropped.CompareAndSwap(false, true) {
						break
					}
					if send(server, frame) != nil {
						break
					}
				}
				client.Close()
				server.Close()
				<-done
			}()
		}
	}()
	t.Cleanup(func() {
		listener.Close()
		mu.Lock()
		for c := range peers {
			c.Close()
		}
		mu.Unlock()
		workers.Wait()
	})
	return listener.Addr().String(), &dropped
}

func TestRealConnectionLossAroundCommit(t *testing.T) {
	for _, after := range []bool{false, true} {
		name := "before-commit"
		if after {
			name = "committed-before-ack"
		}
		t.Run(name, func(t *testing.T) {
			s := testStore(t)
			fixture(t, s)
			cfg := s.pool.Config().Copy()
			addr, dropped := commitProxy(t, net.JoinHostPort(cfg.ConnConfig.Host, strconv.Itoa(int(cfg.ConnConfig.Port))), after)
			host, port, e := net.SplitHostPort(addr)
			if e != nil {
				t.Fatal(e)
			}
			number, e := strconv.Atoi(port)
			if e != nil {
				t.Fatal(e)
			}
			cfg.ConnConfig.Host = host
			cfg.ConnConfig.Port = uint16(number)
			cfg.ConnConfig.TLSConfig = nil
			cfg.ConnConfig.Fallbacks = nil
			cfg.MaxConns = 1
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			pool, e := pgxpool.NewWithConfig(ctx, cfg)
			if e != nil {
				t.Fatal(e)
			}
			t.Cleanup(pool.Close)
			proxied := New(pool, s.rulesHash)
			intent := nullify("wire-disconnect", 1)
			_, e = proxied.Submit(ctx, "m", "a", intent)
			if !errors.Is(e, ErrOutcomeUnknown) || !dropped.Load() {
				t.Fatalf("expected real interrupted commit, got %v (cut=%v)", e, dropped.Load())
			}
			recovered, e := s.Reconcile(ctx, "m", "a", intent)
			if after {
				if e != nil || recovered.Version != 1 {
					t.Fatalf("committed receipt unavailable: %v", e)
				}
			} else if !errors.Is(e, ErrOutcomeUnknown) {
				t.Fatalf("uncommitted receipt unexpectedly exists: %v", e)
			}
			env, v, e := s.Restore(ctx, "m")
			if e != nil {
				t.Fatal(e)
			}
			want := int64(0)
			if after {
				want = 1
			}
			if v != want || len(env.Match.NullificationConsents) != int(want) {
				t.Fatal("wrong durable effect count", v)
			}
			retry, e := proxied.Submit(ctx, "m", "a", intent)
			if e != nil || retry.Version != 1 {
				t.Fatalf("same-ID retry: %v", e)
			}
			again, e := proxied.Submit(ctx, "m", "a", intent)
			if e != nil || !bytes.Equal(retry.Projection, again.Projection) || again.Version != 1 {
				t.Fatal("duplicate changed receipt", e)
			}
			var count int
			if e = s.pool.QueryRow(ctx, "SELECT count(*) FROM command_results WHERE match_id='m'").Scan(&count); e != nil || count != 1 {
				t.Fatal("duplicate effects", count, e)
			}
		})
	}
}
