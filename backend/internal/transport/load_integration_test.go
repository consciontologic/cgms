//go:build integration

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
)

type loadMeasurements struct {
	mu     sync.Mutex
	values map[string][]float64
}

func (m *loadMeasurements) Write(data []byte) (int, error) {
	var event struct {
		Boundary string  `json:"boundary"`
		Seconds  float64 `json:"duration_seconds"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return 0, err
	}
	m.mu.Lock()
	m.values[event.Boundary] = append(m.values[event.Boundary], event.Seconds)
	m.mu.Unlock()
	return len(data), nil
}
func (m *loadMeasurements) report(t *testing.T) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, boundary := range []string{"engine", "transaction", "queue_wait", "lock_wait"} {
		values := m.values[boundary]
		if len(values) == 0 {
			t.Errorf("missing measured boundary %s", boundary)
			continue
		}
		sort.Float64s(values)
		t.Logf("boundary=%s samples=%d p50_ms=%.3f p95_ms=%.3f p99_ms=%.3f", boundary, len(values), values[(len(values)-1)*50/100]*1000, values[(len(values)-1)*95/100]*1000, values[(len(values)-1)*99/100]*1000)
	}
	if data, err := os.ReadFile("/proc/self/status"); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "VmRSS:") || strings.HasPrefix(line, "VmHWM:") {
				t.Log(line)
			}
		}
	}
}

// This bounded load fixture measures light nullification-consent transitions.
// It is not a full-game mix, WAN test, sustained release SLA or 400-socket claim.
func TestOperationalCapacityAndTelemetryLoad(t *testing.T) {
	for _, profile := range []struct {
		streams  int
		exporter bool
	}{{64, false}, {400, false}, {400, true}} {
		enabled := profile.exporter
		t.Run(fmt.Sprintf("streams=%d/exporter=%v", profile.streams, enabled), func(t *testing.T) {
			base := journeyServer(t)
			cfg := base.cfg
			cfg.ProxyClientIP = true
			cfg.MaxStreams = profile.streams
			s, err := New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(context.Background()); err != nil {
					t.Error(err)
				}
			})
			measurements := &loadMeasurements{values: map[string][]float64{}}
			recorder := telemetry.NewWithLevel(measurements, false, slog.LevelDebug)
			s.cfg.Telemetry = recorder
			sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.Copy(io.Discard, r.Body); w.WriteHeader(200) }))
			defer sink.Close()
			if enabled {
				if err := recorder.Enable(context.Background(), sink.URL); err != nil {
					t.Fatal(err)
				}
			}
			defer recorder.Close(context.Background())
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			actors := make([]identity.Session, 400)
			for i := range actors {
				var err error
				actors[i], err = s.identity.CreateGuest(ctx)
				if err != nil {
					t.Fatal(err)
				}
			}
			for m := 0; m < 100; m++ {
				board, err := game.NewState(4, fmt.Sprintf("g%d", m))
				if err != nil {
					t.Fatal(err)
				}
				match, err := game.NewMatchLifecycle(board, 3)
				if err != nil {
					t.Fatal(err)
				}
				env, err := matchstore.NewEnvelope(match, "rules-test")
				if err != nil {
					t.Fatal(err)
				}
				members := make([]string, 4)
				for seat := range members {
					members[seat] = actors[m*4+seat].Account.ID
				}
				if err = s.matches.Create(ctx, fmt.Sprintf("m%d", m), env, members); err != nil {
					t.Fatal(err)
				}
			}
			host := httptest.NewServer(s)
			defer host.Close()
			var sockets []*websocket.Conn
			defer func() {
				for _, c := range sockets {
					_ = c.CloseNow()
				}
			}()
			rejected := 0
			for i, actor := range actors {
				header := http.Header{"Authorization": []string{"Bearer " + actor.Token}, "Origin": []string{s.cfg.Origin}, "X-Real-IP": []string{fmt.Sprintf("192.0.%d.%d", i/250, i%250+1)}}
				c, response, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(host.URL, "http")+fmt.Sprintf("/v1/matches/m%d/stream?after=0", i/4), &websocket.DialOptions{HTTPHeader: header})
				if err != nil {
					if response == nil || response.StatusCode != 429 {
						t.Fatal("unexpected socket refusal", err)
					}
					rejected++
					continue
				}
				sockets = append(sockets, c)
				go func() {
					for {
						if _, _, err := c.Read(ctx); err != nil {
							return
						}
					}
				}()
			}
			if len(sockets) != profile.streams || rejected != 400-profile.streams {
				t.Fatalf("bounded stream admission changed: accepted=%d rejected=%d", len(sockets), rejected)
			}
			type sample struct {
				duration  time.Duration
				status    int
				err       error
				errorBody string
			}
			t.Logf("load bounds: streams=%d active_requests=%d pool_max=%d request_timeout=2s command_interval=50ms commands=300 matches=100 gomaxprocs=%d cpus=%d", s.cfg.MaxStreams, cap(s.active), s.cfg.Pool.Stat().MaxConns(), runtime.GOMAXPROCS(0), runtime.NumCPU())
			if data, err := os.ReadFile("/proc/loadavg"); err == nil {
				fields := strings.Fields(string(data))
				if len(fields) >= 3 {
					t.Logf("host load averages before command phase: %s", strings.Join(fields[:3], " "))
				}
			}
			results := make(chan sample, 300)
			var wg sync.WaitGroup
			before := s.cfg.Pool.Stat()
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()
			started := time.Now()
			for i := 0; i < 300; i++ {
				<-ticker.C
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					m := i % 100
					body, _ := json.Marshal(map[string]any{"command_id": fmt.Sprintf("load-%d", i), "game_id": fmt.Sprintf("g%d", m), "expected_version": i / 100, "type": "nullify", "payload": map[string]any{}})
					request, _ := http.NewRequestWithContext(ctx, "POST", host.URL+fmt.Sprintf("/v1/matches/m%d/commands", m), bytes.NewReader(body))
					request.Header.Set("Authorization", "Bearer "+actors[m*4].Token)
					request.Header.Set("Content-Type", "application/json")
					request.Header.Set("X-Real-IP", fmt.Sprintf("198.51.100.%d", m+1))
					begin := time.Now()
					response, err := host.Client().Do(request)
					status := 0
					errorBody := ""
					if err == nil {
						status = response.StatusCode
						if status != 200 {
							// Synthetic fixture error payload only; never log tokens or headers.
							body, readErr := io.ReadAll(io.LimitReader(response.Body, 1024))
							errorBody = string(body)
							if readErr != nil {
								err = readErr
							}
						}
						_, _ = io.Copy(io.Discard, response.Body)
						response.Body.Close()
					}
					results <- sample{time.Since(begin), status, err, errorBody}
				}(i)
			}
			wg.Wait()
			close(results)
			latencies := make([]time.Duration, 0, 300)
			for result := range results {
				if result.err != nil || result.status != 200 {
					t.Fatalf("command failed: status=%d err=%v duration=%s error_body=%q pool_acquired=%d pool_wait_delta=%d", result.status, result.err, result.duration, result.errorBody, s.cfg.Pool.Stat().AcquiredConns(), s.cfg.Pool.Stat().EmptyAcquireCount()-before.EmptyAcquireCount())
				}
				latencies = append(latencies, result.duration)
			}
			sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
			measurements.report(t)
			var memory runtime.MemStats
			runtime.ReadMemStats(&memory)
			after := s.cfg.Pool.Stat()
			t.Logf("matches=100 seats=400 sockets_admitted=%d sockets_rejected=%d commands=300 gomaxprocs=%d duration=%s response_p50=%s response_p95=%s response_p99=%s heap_bytes=%d goroutines=%d pool_max=%d pool_wait_delta=%d pool_acquire_duration_delta=%s exporter=%v runtime=%s cpus=%d", len(sockets), rejected, runtime.GOMAXPROCS(0), time.Since(started), latencies[149], latencies[284], latencies[296], memory.HeapAlloc, runtime.NumGoroutine(), after.MaxConns(), after.EmptyAcquireCount()-before.EmptyAcquireCount(), after.AcquireDuration()-before.AcquireDuration(), enabled, runtime.Version(), runtime.NumCPU())
		})
	}
}

// Scripted business-operation mix, not a bot-strength or full-game benchmark.
// All measured mutations cross HTTP authentication, wire decoding and durable
// authority. Synthetic fixture preparation is excluded from measured stages.
func TestOperationalGameplayMix(t *testing.T) {
	// Keep the default integration gate short; measured runs opt into 30s stages.
	seconds := 1
	if value := os.Getenv("CGMS_GAMEPLAY_STAGE_SECONDS"); value != "" {
		if _, err := fmt.Sscan(value, &seconds); err != nil || seconds < 1 || seconds > 90 {
			t.Fatal("stage seconds must be 1..90")
		}
	}
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprintf("exporter=%v", enabled), func(t *testing.T) {
			base := journeyServer(t)
			measurements := &loadMeasurements{values: map[string][]float64{}}
			recorder := telemetry.NewWithLevel(measurements, false, slog.LevelDebug)
			cfg := base.cfg
			cfg.Telemetry = recorder
			cfg.ProxyClientIP = true
			cfg.MaxStreams = 400
			s, err := New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := s.Close(context.Background()); err != nil {
					t.Error(err)
				}
			})
			sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.Copy(io.Discard, r.Body); w.WriteHeader(200) }))
			defer sink.Close()
			if enabled {
				if err := recorder.Enable(context.Background(), sink.URL); err != nil {
					t.Fatal(err)
				}
			}
			defer recorder.Close(context.Background())
			host := httptest.NewServer(s)
			defer host.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			const matches = 100
			workers := make([]*gameplayLoadClient, matches)
			for i := range workers {
				workers[i] = newGameplayLoadClient(t, ctx, s, host, i)
			}
			var streamWG sync.WaitGroup
			var eventCount atomic.Int64
			streamErrors := make(chan error, 400)
			var sockets []*websocket.Conn
			var closing atomic.Bool
			defer func() {
				closing.Store(true)
				for _, socket := range sockets {
					_ = socket.CloseNow()
				}
				streamWG.Wait()
			}()
			for _, c := range workers {
				for seat, actor := range c.actors {
					header := http.Header{"Authorization": []string{"Bearer " + actor.Token}, "Origin": []string{s.cfg.Origin}, "X-Real-IP": []string{fmt.Sprintf("198.18.%d.%d", c.index, seat+1)}}
					socket, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(host.URL, "http")+"/v1/matches/"+c.id+"/stream?after=0", &websocket.DialOptions{HTTPHeader: header})
					if err != nil {
						t.Fatal(err)
					}
					sockets = append(sockets, socket)
					streamWG.Add(1)
					go func() {
						defer streamWG.Done()
						var previous int64
						for {
							_, data, err := socket.Read(ctx)
							if err != nil {
								if !closing.Load() {
									streamErrors <- fmt.Errorf("live stream disconnected: %w", err)
								}
								return
							}
							var event struct {
								Cursor int64           `json:"cursor"`
								Data   json.RawMessage `json:"data"`
							}
							if json.Unmarshal(data, &event) != nil || event.Cursor <= previous || len(event.Data) == 0 || bytes.Contains(data, []byte("deck-1")) || bytes.Contains(data, []byte("deck-2")) {
								streamErrors <- fmt.Errorf("invalid ordered authorized event")
								return
							}
							previous = event.Cursor
							eventCount.Add(1)
						}
					}()
				}
			}
			if len(sockets) != 400 {
				t.Fatal("missing live subscriptions")
			}
			s.Start()
			runGameplayClients(t, workers, func(c *gameplayLoadClient) error {
				if err := c.command(0, "open-series", map[string]any{"suit": "diamonds"}); err != nil {
					return err
				}
				for _, seat := range []int{1, 2, 3} {
					if err := c.pass(seat); err != nil {
						return err
					}
				}
				return nil
			})
			for _, rate := range []int{20, 80, 200} {
				t.Run(fmt.Sprintf("commands_per_second=%d", rate), func(t *testing.T) {
					// One bounded worker per match; discard scheduling opportunities
					// rather than accumulate an unbounded client-side backlog.
					busy := make([]chan struct{}, matches)
					for i := range busy {
						busy[i] = make(chan struct{}, 1)
					}
					var wg sync.WaitGroup
					errors := make(chan error, matches)
					before := s.cfg.Pool.Stat()
					started := time.Now()
					for _, c := range workers {
						c.samples = nil
						c.attempts = 0
						c.rejections = 0
					}
					measurements.mu.Lock()
					measurements.values = map[string][]float64{}
					measurements.mu.Unlock()
					cycles := rate * seconds / 4
					ticker := time.NewTicker(4 * time.Second / time.Duration(rate))
					dropped, maxAcquired := 0, int32(0)
					peakHTTP := 0
					for n := 0; n < cycles; n++ {
						select {
						case <-ticker.C:
						case <-ctx.Done():
							t.Fatal(ctx.Err())
						}
						if v := s.cfg.Pool.Stat().AcquiredConns(); v > maxAcquired {
							maxAcquired = v
						}
						if v := len(s.active); v > peakHTTP {
							peakHTTP = v
						}
						i := n % matches
						select {
						case busy[i] <- struct{}{}:
						default:
							dropped++
							continue
						}
						wg.Add(1)
						go func(i int) {
							defer wg.Done()
							defer func() { <-busy[i] }()
							if err := workers[i].cycle(); err != nil {
								select {
								case errors <- err:
								default:
								}
							}
						}(i)
					}
					ticker.Stop()
					wg.Wait()
					close(errors)
					for err := range errors {
						t.Error(err)
					}
					var samples []time.Duration
					attempts, rejections := 0, 0
					for _, c := range workers {
						samples = append(samples, c.samples...)
						attempts += c.attempts
						rejections += c.rejections
					}
					if len(samples) == 0 {
						t.Fatal("no commands measured")
					}
					sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
					stat := s.cfg.Pool.Stat()
					t.Logf("target_commands_per_second=%d commands=%d skipped_cycles=%d elapsed=%s http_p50=%s http_p95=%s http_p99=%s pool_max=%d pool_peak_sampled=%d pool_empty_acquires=%d pool_acquire_duration=%s", rate, len(samples), dropped, time.Since(started), samples[(len(samples)-1)/2], samples[(len(samples)-1)*95/100], samples[(len(samples)-1)*99/100], stat.MaxConns(), maxAcquired, stat.EmptyAcquireCount()-before.EmptyAcquireCount(), stat.AcquireDuration()-before.AcquireDuration())
					t.Logf("HTTP_attempts_including_reads=%d admission_429=%d peak_active_HTTP_sampled=%d", attempts, rejections, peakHTTP)
					measurements.report(t)
				})
				if t.Failed() {
					return
				}
			}
			// Complete actual card transfers after sustained proposal and financial
			// traffic; check canonical journal replay and balance conservation.
			runGameplayClients(t, workers, func(c *gameplayLoadClient) error {
				if err := c.offer(); err != nil {
					return err
				}
				if err := c.command(1, "accept-offer", map[string]any{"offer_id": c.offerID, "revision": 1}); err != nil {
					return err
				}
				for _, seat := range []int{2, 3, 0} {
					if err := c.pass(seat); err != nil {
						return err
					}
				}
				if err := c.command(0, "end-turn", map[string]any{}); err != nil {
					return err
				}
				e, _, err := s.matches.Restore(ctx, c.id)
				if err != nil {
					return err
				}
				card, _ := e.Match.Game.Board.Card("deck-1-spades-05")
				if card.Controller != 2 || e.Match.Game.Ledger.Cash[0].Cmp(game.IntAmount(100)) != 0 || e.Match.Game.Ledger.Cash[1].Sign() != 0 {
					return fmt.Errorf("committed gameplay/accounting invariant failed")
				}
				return nil
			})
			// Replay one representative full trace: replaying all 100 is unrelated
			// to throughput measurement and needlessly multiplies test runtime.
			if err := s.matches.VerifyJournal(ctx, workers[0].id); err != nil {
				t.Fatal(err)
			}
			if len(streamErrors) > 0 {
				t.Fatal(<-streamErrors)
			}
			if eventCount.Load() < 400 {
				t.Fatal("streams did not deliver authorized mutations")
			}
			t.Logf("live_subscriptions=%d validated_events=%d", len(sockets), eventCount.Load())
			t.Log("mix validated: openings, ordered responses, offers/declines, voluntary transfers, accepted card transfers, turn completion; 100 matches")
		})
	}
}

type gameplayLoadClient struct {
	ctx        context.Context
	host       *httptest.Server
	id         string
	actors     []identity.Session
	index      int
	serial     int
	offerID    string
	samples    []time.Duration
	last       [4]time.Time
	attempts   int
	rejections int
}

func newGameplayLoadClient(t *testing.T, ctx context.Context, s *Server, host *httptest.Server, index int) *gameplayLoadClient {
	t.Helper()
	c := &gameplayLoadClient{ctx: ctx, host: host, id: fmt.Sprintf("mix-%d", index), index: index}
	members := []string{}
	for i := 0; i < 4; i++ {
		actor, err := s.identity.CreateGuest(ctx)
		if err != nil {
			t.Fatal(err)
		}
		c.actors = append(c.actors, actor)
		members = append(members, actor.Account.ID)
	}
	board, err := game.NewState(4, c.id)
	if err != nil {
		t.Fatal(err)
	}
	board, err = board.Move([]string{"deck-1-diamonds-05", "deck-1-spades-05"}, 1, game.Hand, true)
	if err != nil {
		t.Fatal(err)
	}
	board.Order = []int{1, 2, 3, 4}
	m, err := game.NewMatchLifecycle(board, 3)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	m.Game.Ledger, err = game.SettleReceipts(m.Game.Ledger, []game.FinancialReceipt{{Seat: 0, Amount: game.IntAmount(100)}})
	if err != nil {
		t.Fatal(err)
	}
	env, err := matchstore.NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.matches.Create(ctx, c.id, env, members); err != nil {
		t.Fatal(err)
	}
	return c
}
func (c *gameplayLoadClient) request(seat int, method, path string, body any) ([]byte, error) {
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}
	// 429 is an explicit non-commit admission result. Retain the exact command
	// identity/body and bound retries; never retry unknown or validation failures.
	for attempt := 0; attempt < 5; attempt++ {
		if delay := 250*time.Millisecond - time.Since(c.last[seat]); delay > 0 {
			select {
			case <-time.After(delay):
			case <-c.ctx.Done():
				return nil, c.ctx.Err()
			}
		}
		c.last[seat] = time.Now()
		r, err := http.NewRequestWithContext(c.ctx, method, c.host.URL+path, bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		r.Header.Set("Authorization", "Bearer "+c.actors[seat].Token)
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Real-IP", fmt.Sprintf("198.18.%d.%d", c.index, seat+1))
		begin := time.Now()
		response, err := c.host.Client().Do(r)
		c.attempts++
		if err != nil {
			return nil, err
		}
		out, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return nil, err
		}
		if response.StatusCode == 429 {
			c.rejections++
			continue
		}
		if response.StatusCode != 200 {
			return nil, fmt.Errorf("%s %s: HTTP %d", method, path, response.StatusCode)
		}
		if method == "POST" {
			c.samples = append(c.samples, time.Since(begin))
		}
		return out, nil
	}
	return nil, fmt.Errorf("HTTP admission did not recover within five attempts")
}
func (c *gameplayLoadClient) snapshot(seat int) (matchstore.View, matchstore.Projection, error) {
	var v matchstore.View
	var p matchstore.Projection
	data, err := c.request(seat, "GET", "/v1/matches/"+c.id, nil)
	if err == nil {
		err = json.Unmarshal(data, &v)
	}
	if err == nil {
		err = json.Unmarshal(v.Projection, &p)
	}
	return v, p, err
}
func (c *gameplayLoadClient) command(seat int, kind string, payload any) error {
	deadline := time.Now().Add(5 * time.Second)
	for {
		v, p, err := c.snapshot(seat)
		if err != nil {
			return err
		}
		if p.Online.AutomaticPending {
			if time.Now().After(deadline) {
				return fmt.Errorf("automatic continuation deadline")
			}
			time.Sleep(10 * time.Millisecond)
			continue
		}
		c.serial++
		_, err = c.request(seat, "POST", "/v1/matches/"+c.id+"/commands", map[string]any{"command_id": fmt.Sprintf("mix-%d", c.serial), "game_id": c.id, "expected_version": v.Version, "type": kind, "payload": payload})
		return err
	}
}
func (c *gameplayLoadClient) pass(seat int) error {
	_, p, err := c.snapshot(seat)
	if err != nil {
		return err
	}
	return c.command(seat, "pass", map[string]any{"pending_action_id": p.Board.WindowID})
}
func (c *gameplayLoadClient) offer() error {
	_, p, err := c.snapshot(0)
	if err != nil {
		return err
	}
	handle := ""
	for _, card := range p.Board.Cards {
		if card.Controller == 1 && card.Zone == game.Hand && card.Card.Suit == game.Spades && card.Card.Rank == 5 {
			handle = card.Card.ID
		}
	}
	if handle == "" {
		return fmt.Errorf("fixture card not authorized")
	}
	c.offerID = fmt.Sprintf("offer-%d", c.serial+1)
	return c.command(0, "offer", map[string]any{"offer_id": c.offerID, "revision": 1, "terms": map[string]any{"to": 2, "give": []string{handle}}})
}
func (c *gameplayLoadClient) cycle() error {
	if err := c.offer(); err != nil {
		return err
	}
	if err := c.command(1, "decline-offer", map[string]any{"offer_id": c.offerID, "revision": 1}); err != nil {
		return err
	}
	amount := map[string]string{"numerator": "1", "denominator": "1"}
	if err := c.command(0, "voluntary-transfer", map[string]any{"recipient": 2, "amount": amount}); err != nil {
		return err
	}
	return c.command(1, "voluntary-transfer", map[string]any{"recipient": 1, "amount": amount})
}

func runGameplayClients(t *testing.T, clients []*gameplayLoadClient, fn func(*gameplayLoadClient) error) {
	t.Helper()
	var wg sync.WaitGroup
	limit := make(chan struct{}, 16)
	errs := make(chan error, len(clients))
	for _, c := range clients {
		limit <- struct{}{}
		wg.Add(1)
		go func(c *gameplayLoadClient) {
			defer wg.Done()
			defer func() { <-limit }()
			if err := fn(c); err != nil {
				errs <- err
			}
		}(c)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if t.Failed() {
		t.FailNow()
	}
}

func TestGameplayLoadAdmissionRetries(t *testing.T) {
	for _, tc := range []struct {
		name      string
		statuses  []int
		attempts  int
		wantError bool
	}{
		{"bounded-recovery", []int{429, 429, 200}, 3, false},
		{"exhausted", []int{429}, 5, true},
		{"unavailable-not-retried", []int{503}, 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			var first []byte
			host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if calls == 0 {
					first = data
				} else if !bytes.Equal(first, data) {
					t.Error("retry changed immutable command body")
				}
				status := tc.statuses[min(calls, len(tc.statuses)-1)]
				calls++
				w.WriteHeader(status)
				_, _ = w.Write([]byte(`{}`))
			}))
			defer host.Close()
			c := &gameplayLoadClient{ctx: context.Background(), host: host, actors: []identity.Session{{Token: "synthetic"}}}
			_, err := c.request(0, "POST", "/commands", map[string]any{"command_id": "immutable", "expected_version": 7})
			if (err != nil) != tc.wantError || calls != tc.attempts {
				t.Fatalf("err=%v attempts=%d", err, calls)
			}
			wantSamples := 0
			if !tc.wantError {
				wantSamples = 1
			}
			if len(c.samples) != wantSamples {
				t.Fatal("rejected attempt included in successful response samples")
			}
		})
	}
}
