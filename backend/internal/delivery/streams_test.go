package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/coder/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamReplayAndShutdown(t *testing.T) {
	streams := NewStreams(Options{PollInterval: time.Millisecond, WriteTimeout: time.Second})
	done := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		done <- streams.Serve(w, r, context.Background(), "a", 0, func(context.Context) error { return nil }, func(ctx context.Context, after int64) ([]Event, error) {
			if after == 0 {
				return []Event{{Cursor: 1, Data: json.RawMessage(`{"safe":true}`)}}, nil
			}
			return nil, nil
		})
	}))
	defer server.Close()
	conn, _, e := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if e != nil {
		t.Fatal(e)
	}
	defer conn.CloseNow()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, data, e := conn.Read(ctx)
	if e != nil || !strings.Contains(string(data), `"cursor":1`) {
		t.Fatal(string(data), e)
	}
	if e = streams.Close(ctx); e != nil {
		t.Fatal(e)
	}
	select {
	case <-done:
	case <-ctx.Done():
		t.Fatal("socket worker leaked")
	}
	if streams.Active() != 0 {
		t.Fatal("connection leaked")
	}
}
func TestStreamRejectsAuthBeforeUpgrade(t *testing.T) {
	s := NewStreams(Options{})
	defer s.Close(context.Background())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost/stream", nil)
	e := s.Serve(w, r, context.Background(), "a", 0, func(context.Context) error { return errors.New("private auth detail") }, func(context.Context, int64) ([]Event, error) { t.Fatal("queried unauthenticated"); return nil, nil })
	if e == nil || w.Code != 401 || strings.Contains(w.Body.String(), "private") {
		t.Fatal(w.Code, w.Body.String(), e)
	}
}

func TestSlowReaderAndConnectionCap(t *testing.T) {
	streams := NewStreams(Options{MaxConnections: 1, PerAccount: 1, PollInterval: time.Nanosecond, WriteTimeout: 20 * time.Millisecond})
	defer streams.Close(context.Background())
	done := make(chan error, 1)
	payload := json.RawMessage(`"` + strings.Repeat("x", 65536) + `"`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := streams.Serve(w, r, context.Background(), "a", 0, func(context.Context) error { return nil }, func(ctx context.Context, after int64) ([]Event, error) {
			return []Event{{Cursor: after + 1, Data: payload}}, nil
		})
		done <- err
	}))
	defer server.Close()
	conn, _, e := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if e != nil {
		t.Fatal(e)
	}
	defer conn.CloseNow()
	// A non-reader must not retain its connection permit indefinitely.
	select {
	case e = <-done:
		if e == nil {
			t.Fatal("slow stream did not fail")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("slow-reader deadline ignored")
	}
	if streams.Active() != 0 {
		t.Fatal("slow reader leaked worker")
	}
}
func TestStreamConnectionLimitAndCursorGap(t *testing.T) {
	s := NewStreams(Options{MaxConnections: 1})
	defer s.Close(context.Background())
	if e := s.admit("a"); e != nil {
		t.Fatal(e)
	}
	w := httptest.NewRecorder()
	e := s.Serve(w, httptest.NewRequest("GET", "/", nil), context.Background(), "b", 0, func(context.Context) error { return nil }, func(context.Context, int64) ([]Event, error) { return nil, nil })
	if !errors.Is(e, ErrOverloaded) || w.Code != 429 {
		t.Fatal(e, w.Code)
	}
	s.release("a")
	done := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		done <- s.Serve(w, r, context.Background(), "a", 0, func(context.Context) error { return nil }, func(context.Context, int64) ([]Event, error) {
			return []Event{{Cursor: 2, Data: json.RawMessage(`{}`)}}, nil
		})
	}))
	defer server.Close()
	c, _, e := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if e != nil {
		t.Fatal(e)
	}
	defer c.CloseNow()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, payload, readErr := c.Read(ctx)
	if readErr != nil || !strings.Contains(string(payload), "RESNAPSHOT_REQUIRED") {
		t.Fatal("missing explicit cursor recovery", readErr)
	}
	select {
	case e = <-done:
		if !errors.Is(e, ErrCursor) {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("gap not rejected")
	}
}

func TestQuietStreamSendsHeartbeat(t *testing.T) {
	streams := NewStreams(Options{HeartbeatInterval: 10 * time.Millisecond, PollInterval: time.Millisecond})
	defer streams.Close(context.Background())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = streams.Serve(w, r, context.Background(), "a", 0, func(context.Context) error { return nil }, func(context.Context, int64) ([]Event, error) { return nil, nil })
	}))
	defer server.Close()
	ping := make(chan struct{}, 1)
	conn, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), &websocket.DialOptions{OnPingReceived: func(context.Context, []byte) bool {
		select {
		case ping <- struct{}{}:
		default:
		}
		return true
	}})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); _, _, _ = conn.Read(ctx) }()
	select {
	case <-ping:
	case <-ctx.Done():
		t.Fatal("quiet stream sent no heartbeat")
	}
	cancel()
	<-done
}
