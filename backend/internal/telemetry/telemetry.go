// Package telemetry exposes bounded, allowlisted operational measurements. It
// never accepts payloads, SQL, headers, caller IDs or arbitrary attributes.
package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Boundary string

const (
	Request     Boundary = "request"
	Transaction Boundary = "transaction"
	Engine      Boundary = "engine"
	Queue       Boundary = "queue_wait"
	Lock        Boundary = "lock_wait"
	Delivery    Boundary = "delivery"
	ReplayGap   Boundary = "replay_gap"
)

var UnknownCommit = errors.New("unknown commit")

type key struct{}
type measurement struct {
	Count   uint64
	Seconds float64
}
type Recorder struct {
	logger   *slog.Logger
	mu       sync.Mutex
	values   map[Boundary]map[string]measurement
	provider *sdk.TracerProvider
}
type scope struct {
	recorder *Recorder
	id       string
}

func New(w io.Writer, human bool) *Recorder {
	return NewWithLevel(w, human, slog.LevelInfo)
}

// NewWithLevel permits explicit diagnostic capture (for example load-test
// latency samples). Normal server startup uses New's quiet INFO threshold.
func NewWithLevel(w io.Writer, human bool, level slog.Level) *Recorder {
	opts := &slog.HandlerOptions{Level: level, ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey {
			return slog.Time(slog.TimeKey, a.Value.Time().UTC())
		}
		return a
	}}
	var handler slog.Handler = slog.NewJSONHandler(w, opts)
	if human {
		handler = slog.NewTextHandler(w, opts)
	}
	return &Recorder{logger: slog.New(handler), values: map[Boundary]map[string]measurement{}}
}

// Enable creates a bounded exporter only when explicitly called at startup.
// No URL, raw exporter error or payload is ever logged. Disabled mode creates
// no exporter and does not consult ambient OTEL environment variables.
func (r *Recorder) Enable(ctx context.Context, endpoint string) error {
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint), otlptracehttp.WithTimeout(time.Second), otlptracehttp.WithRetry(otlptracehttp.RetryConfig{Enabled: false}))
	if err != nil {
		return errors.New("telemetry initialization failed")
	}
	r.provider = sdk.NewTracerProvider(sdk.WithResource(resource.NewSchemaless(attribute.String("service.name", "cgms-backend"))), sdk.WithBatcher(&safeExporter{next: exporter, recorder: r}, sdk.WithMaxQueueSize(256), sdk.WithMaxExportBatchSize(64), sdk.WithBatchTimeout(time.Second), sdk.WithExportTimeout(time.Second)))
	return nil
}

type safeExporter struct {
	next     sdk.SpanExporter
	recorder *Recorder
	mu       sync.Mutex
	last     time.Time
}

func (s *safeExporter) ExportSpans(ctx context.Context, spans []sdk.ReadOnlySpan) error {
	// Circuit cooldown bounds repeated failed connections; gameplay never waits.
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.last.IsZero() && time.Since(s.last) < time.Minute {
		return nil
	}
	if err := s.next.ExportSpans(ctx, spans); err != nil {
		s.last = time.Now()
		s.recorder.logger.Warn("telemetry_export_unavailable")
		return nil
	}
	s.last = time.Time{}
	return nil
}
func (s *safeExporter) Shutdown(ctx context.Context) error { return s.next.Shutdown(ctx) }
func (r *Recorder) Close(ctx context.Context) error {
	if r.provider != nil {
		return r.provider.Shutdown(ctx)
	}
	return nil
}
func (r *Recorder) Start(ctx context.Context, b Boundary) (context.Context, func(error)) {
	switch b {
	case Request, Transaction, Engine, Queue, Lock, Delivery, ReplayGap:
	default:
		b = Request
	}
	sc, ok := ctx.Value(key{}).(scope)
	if !ok || sc.recorder != r || sc.id == "" {
		var id [12]byte
		_, _ = rand.Read(id[:])
		sc = scope{r, hex.EncodeToString(id[:])}
		ctx = context.WithValue(ctx, key{}, sc)
	}
	var span trace.Span
	if r.provider != nil {
		ctx, span = r.provider.Tracer("cgms.boundaries").Start(ctx, string(b))
	}
	start := time.Now()
	var once sync.Once
	return ctx, func(err error) {
		once.Do(func() {
			outcome := "ok"
			if err != nil {
				outcome = "error"
			}
			if errors.Is(err, UnknownCommit) {
				outcome = "unknown_commit"
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				outcome = "canceled"
			}
			seconds := time.Since(start).Seconds()
			r.mu.Lock()
			if r.values[b] == nil {
				r.values[b] = map[string]measurement{}
			}
			m := r.values[b][outcome]
			m.Count++
			m.Seconds += seconds
			r.values[b][outcome] = m
			r.mu.Unlock()
			// Routine boundaries are measured without flooding the default logs.
			// Expected cancellation is quiet; failures and deadlines stay visible.
			level := slog.LevelDebug
			if err != nil && !errors.Is(err, context.Canceled) {
				level = slog.LevelWarn
			}
			r.logger.Log(ctx, level, "boundary", "boundary", string(b), "outcome", outcome, "operation_id", sc.id, "duration_seconds", seconds)
			if span != nil {
				span.SetAttributes(attribute.String("outcome", outcome))
				span.End()
			}
		})
	}
}
func Start(ctx context.Context, b Boundary) (context.Context, func(error)) {
	if sc, ok := ctx.Value(key{}).(scope); ok {
		return sc.recorder.Start(ctx, b)
	}
	return ctx, func(error) {}
}
func (r *Recorder) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, key{}, scope{recorder: r})
}
func (r *Recorder) Metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	fmt.Fprintf(w, "cgms_heap_bytes %d\ncgms_goroutines %d\n", memory.HeapAlloc, runtime.NumGoroutine())
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range []Boundary{Request, Transaction, Engine, Queue, Lock, Delivery, ReplayGap} {
		for _, o := range []string{"ok", "error", "unknown_commit", "canceled"} {
			m := r.values[b][o]
			fmt.Fprintf(w, "cgms_boundary_total{boundary=%q,outcome=%q} %d\ncgms_boundary_duration_seconds_sum{boundary=%q,outcome=%q} %g\n", b, o, m.Count, b, o, m.Seconds)
		}
	}
}
