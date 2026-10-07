//go:build integration

package presence

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestRedisLossAndRecoveryRemainHints(t *testing.T) {
	raw, name := os.Getenv("CGMS_TEST_REDIS_URL"), os.Getenv("CGMS_TEST_REDIS_CONTAINER")
	if raw == "" || !strings.HasPrefix(name, "cgms-integration-") || !strings.HasSuffix(name, "-redis") {
		t.Fatal("bounded --redis integration harness required")
	}
	p, e := New(raw, Options{})
	if e != nil {
		t.Fatal("invalid test Redis configuration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if !p.Touch(ctx, "seat-a") {
		t.Fatal("Redis touch failed")
	}
	if v, source := p.Online(ctx, "seat-a"); !v || source != "redis" {
		t.Fatal(v, source)
	}
	if e = exec.CommandContext(ctx, "docker", "kill", "--signal=KILL", name).Run(); e != nil {
		t.Fatal("Redis crash failed")
	}
	if p.Touch(ctx, "seat-b") {
		t.Fatal("stopped Redis reported shared update")
	}
	if v, source := p.Online(ctx, "seat-b"); !v || source != "local" {
		t.Fatal(v, source)
	}
	if e = exec.CommandContext(ctx, "docker", "start", name).Run(); e != nil {
		t.Fatal("Redis restart failed")
	}
	for !p.Touch(ctx, "seat-b") {
		if ctx.Err() != nil {
			t.Fatal("Redis did not recover")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if v, source := p.Online(ctx, "seat-b"); !v || source != "redis" {
		t.Fatal(v, source)
	}
}
