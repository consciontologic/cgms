package sim

import (
	"bytes"
	"context"
	"github.com/metaphy6/cgms/backend/internal/game"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

type boundaryCancel struct {
	context.Context
	left int
	done chan struct{}
	once sync.Once
}

func (c *boundaryCancel) Done() <-chan struct{} { return c.done }
func (c *boundaryCancel) Err() error {
	c.left--
	if c.left <= 0 {
		c.once.Do(func() { close(c.done) })
		return context.Canceled
	}
	return nil
}
func TestSettlementCancellationRetainsCommittedPrefix(t *testing.T) {
	sc, e := loadScenario(mustRead(t, fixture(t)), 3)
	if e != nil {
		t.Fatal(e)
	}
	ctx := &boundaryCancel{Context: context.Background(), left: 3, done: make(chan struct{})}
	r := runScenario(ctx, Job{Games: 1, Population: 3}, &sc, "42", "standalone", 100)
	if r.Outcome.ExitCode != 130 || len(r.Trace) != 1 {
		t.Fatal(r.Outcome.ExitCode, len(r.Trace))
	}
	c := r.Checkpoint.(game.SettlementCursor)
	if hash(c) != r.Trace[0].Post || hash(c) == hash(r.Initial) {
		t.Fatal("committed prefix lost")
	}
	for !c.Done {
		c, _, e = game.ResumeSettlement(c, 1, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	if c.Ledger.Scores[0].String() != "1/3" {
		t.Fatal("cancellation lost obligations")
	}
}
func TestBatchCancellationPublishesBoundedPartial(t *testing.T) {
	sc := fixture(t)
	out := filepath.Join(t.TempDir(), "out")
	a := runArgs(out, sc)
	a[0] = "batch"
	a = append(a, "--games", "500", "--workers", "2")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	timer := time.AfterFunc(20*time.Millisecond, cancel)
	defer timer.Stop()
	var b bytes.Buffer
	started := time.Now()
	code := Execute(ctx, a, &b, &b)
	if code != 130 {
		t.Fatal(code, b.String())
	}
	if time.Since(started) > 5*time.Second {
		t.Fatal("unbounded cancel")
	}
	m, f, e := readManifest(filepath.Join(out, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	if m.Started+m.NotStarted != 500 || m.ExitCode != 130 {
		t.Fatal("missing denominator")
	}
	if e = verifyPrefix(m, f); e != nil {
		t.Fatal("partial prefix cannot replay", e)
	}
}
func TestReplayCancellationAtBoundary(t *testing.T) {
	d := filepath.Join(t.TempDir(), "out")
	if c, m := invoke(t, runArgs(d, fixture(t))...); c != 0 {
		t.Fatal(c, m)
	}
	m, f, e := readManifest(filepath.Join(d, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	ctx := &boundaryCancel{Context: context.Background(), left: 2, done: make(chan struct{})}
	if e = verifyPrefixContext(ctx, m, f); e != context.Canceled {
		t.Fatal("replay ignored boundary cancellation", e)
	}
}
