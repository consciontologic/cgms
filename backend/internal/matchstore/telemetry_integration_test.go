//go:build integration

package matchstore

import (
	"bytes"
	"context"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
	"strings"
	"testing"
)

func TestBoundaryTelemetryUnknownCommitIsPrivate(t *testing.T) {
	s := testStore(t)
	fixture(t, s)
	var output bytes.Buffer
	rec := telemetry.New(&output, false)
	ctx, done := rec.Start(context.Background(), telemetry.Request)
	defer done(nil)
	s.fault = func(stage string) error {
		if stage == "acknowledgement" {
			return errors.New("secret-seed-token")
		}
		return nil
	}
	_, err := s.Submit(ctx, "m", "a", nullify("private-command", 1))
	if !errors.Is(err, ErrOutcomeUnknown) {
		t.Fatal("expected unknown commit")
	}
	for _, wanted := range []string{"lock_wait", "engine", "transaction", "unknown_commit"} {
		if !strings.Contains(output.String(), wanted) {
			t.Fatalf("missing boundary %s", wanted)
		}
	}
	for _, secret := range []string{"secret-seed-token", "private-command", "projection", "snapshot"} {
		if strings.Contains(output.String(), secret) {
			t.Fatal("private telemetry")
		}
	}
}
