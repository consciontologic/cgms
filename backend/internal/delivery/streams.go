package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/coder/websocket"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
	"net/http"
	"sync"
	"time"
)

var ErrCursor = errors.New("stream cursor requires snapshot")

type Event struct {
	Cursor int64           `json:"cursor"`
	Data   json.RawMessage `json:"data"`
}
type Options struct {
	MaxConnections, PerAccount, MaxReads, MaxEvents, MaxBytes  int
	HeartbeatInterval, PollInterval, WriteTimeout, ReadTimeout time.Duration
	OriginPatterns                                             []string
}
type Streams struct {
	mu       sync.Mutex
	options  Options
	accounts map[string]int
	active   int
	closed   bool
	ctx      context.Context
	cancel   context.CancelFunc
	reads    chan struct{}
	wg       sync.WaitGroup
	done     chan struct{}
}

func NewStreams(o Options) *Streams {
	if o.MaxConnections < 1 {
		o.MaxConnections = 64
	}
	if o.PerAccount < 1 {
		o.PerAccount = 2
	}
	if o.MaxReads < 1 {
		o.MaxReads = 8
	}
	if o.MaxEvents < 1 {
		o.MaxEvents = 128
	}
	if o.MaxBytes < 1 {
		o.MaxBytes = 1 << 20
	}
	if o.HeartbeatInterval <= 0 {
		o.HeartbeatInterval = 20 * time.Second
	}
	if o.PollInterval <= 0 {
		o.PollInterval = time.Second
	}
	if o.WriteTimeout <= 0 {
		o.WriteTimeout = 5 * time.Second
	}
	if o.ReadTimeout <= 0 {
		o.ReadTimeout = 2 * time.Second
	}
	o.OriginPatterns = append([]string(nil), o.OriginPatterns...)
	ctx, cancel := context.WithCancel(context.Background())
	return &Streams{options: o, accounts: map[string]int{}, ctx: ctx, cancel: cancel, reads: make(chan struct{}, o.MaxReads), done: make(chan struct{})}
}
func (s *Streams) Active() int { s.mu.Lock(); defer s.mu.Unlock(); return s.active }
func (s *Streams) admit(account string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrClosed
	}
	if s.active >= s.options.MaxConnections || s.accounts[account] >= s.options.PerAccount {
		return ErrOverloaded
	}
	s.active++
	s.accounts[account]++
	s.wg.Add(1)
	return nil
}
func (s *Streams) release(account string) {
	s.mu.Lock()
	s.active--
	s.accounts[account]--
	if s.accounts[account] == 0 {
		delete(s.accounts, account)
	}
	s.mu.Unlock()
	s.wg.Done()
}
func (s *Streams) read(ctx context.Context, auth func(context.Context) error, next func(context.Context, int64) ([]Event, error), after int64) ([]Event, error) {
	ctx, cancel := context.WithTimeout(ctx, s.options.ReadTimeout)
	defer cancel()
	select {
	case s.reads <- struct{}{}:
		defer func() { <-s.reads }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if err := auth(ctx); err != nil {
		return nil, err
	}
	if next == nil {
		return nil, nil
	}
	return next(ctx, after)
}

// Serve owns the upgraded connection until both its read and write loops stop.
// It writes sanitized pre-upgrade failures itself. Callers must not write another
// HTTP response after Serve. Next must read durable, authorized, bounded events;
// notification loss is harmless because this stream always catches up by cursor.
func (s *Streams) Serve(w http.ResponseWriter, r *http.Request, parent context.Context, account string, after int64, authorize func(context.Context) error, next func(context.Context, int64) ([]Event, error)) error {
	if account == "" || after < 0 || authorize == nil || next == nil {
		http.Error(w, "INVALID_STREAM", 400)
		return ErrCursor
	}
	if err := s.admit(account); err != nil {
		http.Error(w, "OVERLOADED", 429)
		return err
	}
	defer s.release(account)
	ctx, cancel := context.WithCancel(parent)
	stop := context.AfterFunc(s.ctx, cancel)
	defer stop()
	defer cancel()
	if _, err := s.read(ctx, authorize, nil, after); err != nil {
		http.Error(w, "UNAUTHORIZED", 401)
		return err
	}
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: s.options.OriginPatterns})
	if err != nil {
		return err
	}
	c.SetReadLimit(16 << 10)
	readDone := make(chan struct{})
	go func() { defer close(readDone); defer cancel(); _, _, _ = c.Read(ctx) }()
	defer func() { cancel(); c.CloseNow(); <-readDone }()
	terminal := func(reason error) error {
		code := "STREAM_UNAVAILABLE"
		if errors.Is(reason, ErrCursor) {
			_, done := telemetry.Start(ctx, telemetry.ReplayGap)
			done(reason)
			code = "RESNAPSHOT_REQUIRED"
		}
		if errors.Is(reason, ErrOverloaded) {
			code = "OVERLOADED"
		}
		body, _ := json.Marshal(struct {
			Type string `json:"type"`
			Code string `json:"code"`
		}{"error", code})
		writeCtx, stop := context.WithTimeout(ctx, s.options.WriteTimeout)
		defer stop()
		_ = c.Write(writeCtx, websocket.MessageText, body)
		return reason
	}

	heartbeat := time.NewTicker(s.options.HeartbeatInterval)
	defer heartbeat.Stop()
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		case <-heartbeat.C:
			pingCtx, stop := context.WithTimeout(ctx, s.options.WriteTimeout)
			err := c.Ping(pingCtx)
			stop()
			if err != nil {
				return err
			}
			continue
		}
		events, err := s.read(ctx, authorize, next, after)
		if err != nil {
			return terminal(err)
		}
		if len(events) > s.options.MaxEvents {
			return terminal(ErrOverloaded)
		}
		size := 0
		for _, e := range events {
			size += len(e.Data)
			if size > s.options.MaxBytes {
				return terminal(ErrOverloaded)
			}
			if e.Cursor != after+1 {
				return terminal(ErrCursor)
			}
			data, err := json.Marshal(e)
			if err != nil {
				return terminal(ErrCursor)
			}
			writeCtx, stopWrite := context.WithTimeout(ctx, s.options.WriteTimeout)
			err = c.Write(writeCtx, websocket.MessageText, data)
			stopWrite()
			if err != nil {
				return terminal(err)
			}
			after = e.Cursor
		}
		// Pulling, rather than broadcasting under match locks, leaves no unbounded
		// outbound queue. At most one bounded batch plus one frame is retained.
		timer.Reset(s.options.PollInterval)
	}
}
func (s *Streams) Close(ctx context.Context) error {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		s.cancel()
		go func() { s.wg.Wait(); close(s.done) }()
	}
	s.mu.Unlock()
	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
