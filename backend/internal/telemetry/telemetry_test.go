package telemetry

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestExplicitDebugCaptureRetainsRoutineMeasurements(t *testing.T) {
	var log bytes.Buffer
	r := NewWithLevel(&log, false, slog.LevelDebug)
	_, done := r.Start(context.Background(), Engine)
	done(nil)
	if !strings.Contains(log.String(), `"level":"DEBUG"`) ||
		!strings.Contains(log.String(), `"boundary":"engine"`) ||
		!strings.Contains(log.String(), `"outcome":"ok"`) {
		t.Fatal(log.String())
	}
}

func TestRoutineBoundariesCountWithoutFloodingLogs(t *testing.T) {
	for _, human := range []bool{false, true} {
		var log bytes.Buffer
		r := New(&log, human)
		for _, boundary := range []Boundary{Request, Transaction, Engine, Queue, Lock, Delivery, ReplayGap} {
			for _, err := range []error{nil, context.Canceled} {
				_, done := r.Start(context.Background(), boundary)
				done(err)
				done(err) // Completion is still idempotent.
			}
		}
		if log.Len() != 0 {
			t.Fatalf("routine work flooded logs (human=%v): %s", human, log.String())
		}
		w := httptest.NewRecorder()
		r.Metrics(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		for _, boundary := range []Boundary{Request, Transaction, Engine, Queue, Lock, Delivery, ReplayGap} {
			for outcome, count := range map[string]string{"ok": "1", "canceled": "1"} {
				want := `cgms_boundary_total{boundary="` + string(boundary) + `",outcome="` + outcome + `"} ` + count + "\n"
				if !strings.Contains(w.Body.String(), want) {
					t.Fatalf("missing measurement %q", want)
				}
			}
		}
	}
}

func TestBoundaryFailuresRemainVisibleWarnings(t *testing.T) {
	var log bytes.Buffer
	r := New(&log, false)
	for _, err := range []error{errors.New("private failure detail"), UnknownCommit, context.DeadlineExceeded} {
		_, done := r.Start(context.Background(), Transaction)
		done(err)
	}
	if strings.Count(log.String(), `"level":"WARN"`) != 3 ||
		!strings.Contains(log.String(), `"outcome":"error"`) ||
		!strings.Contains(log.String(), `"outcome":"unknown_commit"`) ||
		!strings.Contains(log.String(), `"outcome":"canceled"`) ||
		strings.Contains(log.String(), "private failure detail") {
		t.Fatal(log.String())
	}
}

func TestSafeBoundaryIgnoresPrivateErrorsAndPath(t *testing.T) {
	var log bytes.Buffer
	r := New(&log, false)
	ctx, done := r.Start(context.Background(), Request)
	_, inner := Start(ctx, Engine)
	inner(errors.New("hand=secret seed=secret token=secret"))
	done(nil)
	for _, private := range []string{"hand=", "seed=", "token=", "secret"} {
		if strings.Contains(log.String(), private) {
			t.Fatal("private field escaped")
		}
	}
	if !strings.Contains(log.String(), "engine") || !strings.Contains(log.String(), "error") {
		t.Fatal(log.String())
	}
	w := httptest.NewRecorder()
	r.Metrics(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(w.Body.String(), `cgms_boundary_total{boundary="engine",outcome="error"} 1`) {
		t.Fatal(w.Body.String())
	}
}

func TestDisabledStillLogsAndCounts(t *testing.T) {
	var log bytes.Buffer
	r := New(&log, false)
	_, done := r.Start(context.Background(), Lock)
	done(UnknownCommit)
	if !strings.Contains(log.String(), "unknown_commit") {
		t.Fatal(log.String())
	}
}

func TestUntrustedBoundaryRejected(t *testing.T) {
	var log bytes.Buffer
	r := New(&log, false)
	_, done := r.Start(context.Background(), Boundary("private-card-token"))
	done(nil)
	if strings.Contains(log.String(), "private-card-token") {
		t.Fatal("unbounded label")
	}
}

func TestExporterEnabledAndDisabled(t *testing.T) {
	var count atomic.Int32
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(200)
	}))
	defer sink.Close()
	var log bytes.Buffer
	disabled := New(&log, false)
	_, done := disabled.Start(context.Background(), Request)
	done(nil)
	if e := disabled.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
	if count.Load() != 0 {
		t.Fatal("disabled export")
	}
	enabled := New(&log, false)
	if e := enabled.Enable(context.Background(), sink.URL); e != nil {
		t.Fatal(e)
	}
	_, done = enabled.Start(context.Background(), Request)
	done(errors.New("SECRET"))
	if e := enabled.Close(context.Background()); e != nil {
		t.Fatal(e)
	}
	if count.Load() != 1 {
		t.Fatal("enabled span not exported")
	}
	if strings.Contains(log.String(), "SECRET") {
		t.Fatal("private error")
	}
}

func TestFailedExporterCooldown(t *testing.T) {
	var calls atomic.Int32
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "private exporter detail", 503)
	}))
	defer sink.Close()
	var log bytes.Buffer
	rec := New(&log, false)
	if err := rec.Enable(context.Background(), sink.URL); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		_, done := rec.Start(context.Background(), Request)
		done(nil)
		if err := rec.provider.ForceFlush(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if err := rec.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 1 {
		t.Fatal("repeated failed export")
	}
	if strings.Count(log.String(), "telemetry_export_unavailable") != 1 || strings.Contains(log.String(), "private exporter detail") {
		t.Fatal("unsafe or repeated failure log")
	}
}

func BenchmarkBoundary(b *testing.B) {
	for _, human := range []bool{false, true} {
		name := "json"
		if human {
			name = "human"
		}
		b.Run(name, func(b *testing.B) {
			rec := New(io.Discard, human)
			b.ReportAllocs()
			for b.Loop() {
				_, done := rec.Start(context.Background(), Request)
				done(nil)
			}
		})
	}
}
