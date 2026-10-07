package delivery

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func waitFor(t *testing.T, f func() bool) {
	t.Helper()
	end := time.Now().Add(3 * time.Second)
	for !f() {
		if time.Now().After(end) {
			t.Fatal("bounded wait expired")
		}
		runtime.Gosched()
	}
}
func TestDispatcherReevaluatesPriorityAndBounds(t *testing.T) {
	d := New(4, 3)
	ctx := context.Background()
	defer d.Close(ctx)
	started := make(chan struct{})
	release := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = d.Submit(ctx, "m", 0, func(context.Context) (any, error) { close(started); <-release; return nil, nil })
	}()
	<-started
	var rank atomic.Int32
	rank.Store(2)
	var mu sync.Mutex
	var order []int
	enqueue := func(n int, r func(context.Context) (int, error)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, e := d.SubmitRanked(ctx, "m", r, func(context.Context) (any, error) { mu.Lock(); order = append(order, n); mu.Unlock(); return nil, nil })
			if e != nil {
				t.Error(e)
			}
		}()
	}
	enqueue(1, func(context.Context) (int, error) { return int(rank.Load()), nil })
	waitFor(t, func() bool { return d.Pending() == 2 })
	enqueue(2, func(context.Context) (int, error) { return 1, nil })
	waitFor(t, func() bool { return d.Pending() == 3 })
	if _, e := d.Submit(ctx, "m", 0, func(context.Context) (any, error) { t.Error("overload ran"); return nil, nil }); !errors.Is(e, ErrOverloaded) {
		t.Fatal(e)
	}
	rank.Store(-1)
	close(release)
	wg.Wait()
	if len(order) != 2 || order[0] != 1 {
		t.Fatal(order)
	}
}
func TestDispatcherShutdownCancelsAndJoins(t *testing.T) {
	d := New(2, 2)
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = d.Submit(context.Background(), "m", 0, func(ctx context.Context) (any, error) { close(started); <-ctx.Done(); return nil, ctx.Err() })
	}()
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if e := d.Close(ctx); e != nil {
		t.Fatal(e)
	}
	<-done
	if d.Pending() != 0 {
		t.Fatal("owned work leaked")
	}
	if _, e := d.Submit(ctx, "m", 0, nil); !errors.Is(e, ErrClosed) {
		t.Fatal(e)
	}
}

func TestVictoryDoesNotWaitForOrdinaryRankQueries(t *testing.T) {
	d := New(4, 4)
	defer d.Close(context.Background())
	started, release := make(chan struct{}), make(chan struct{})
	var wg sync.WaitGroup
	var victory atomic.Bool
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = d.Submit(context.Background(), "m", 0, func(context.Context) (any, error) { close(started); <-release; return nil, nil })
	}()
	<-started
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = d.SubmitRanked(context.Background(), "m", func(context.Context) (int, error) {
			if !victory.Load() {
				t.Error("ordinary rank query delayed queued victory")
			}
			return 0, nil
		}, func(context.Context) (any, error) { return nil, nil })
	}()
	waitFor(t, func() bool { return d.Pending() == 2 })
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = d.Submit(context.Background(), "m", -1, func(context.Context) (any, error) { victory.Store(true); return nil, nil })
	}()
	waitFor(t, func() bool { return d.Pending() == 3 })
	close(release)
	wg.Wait()
}

func TestVictoryArrivingDuringRankingPrecedesOrdinaryMutation(t *testing.T) {
	d := New(4, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	defer d.Close(ctx)
	ranking, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	var wg sync.WaitGroup
	order := make(chan string, 2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := d.SubmitRanked(ctx, "m", func(ctx context.Context) (int, error) {
			once.Do(func() { close(ranking) })
			select {
			case <-release:
				return 0, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		}, func(context.Context) (any, error) { order <- "ordinary"; return nil, nil })
		if err != nil {
			t.Error(err)
		}
	}()
	<-ranking
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := d.Submit(ctx, "m", -1, func(context.Context) (any, error) { order <- "victory"; return nil, nil })
		if err != nil {
			t.Error(err)
		}
	}()
	waitFor(t, func() bool { return d.Pending() == 2 })
	close(release)
	wg.Wait()
	if first := <-order; first != "victory" {
		t.Fatalf("ordinary mutation overtook a queued victory: first=%s", first)
	}
}

func TestActivePlayerArrivingDuringRankingPrecedesOtherOrdinaryMutation(t *testing.T) {
	d := New(4, 4)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	defer d.Close(ctx)
	ranking, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	var wg sync.WaitGroup
	order := make(chan string, 2)
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := d.SubmitRanked(ctx, "m", func(ctx context.Context) (int, error) {
			once.Do(func() { close(ranking) })
			select {
			case <-release:
				return 2, nil
			case <-ctx.Done():
				return 0, ctx.Err()
			}
		}, func(context.Context) (any, error) { order <- "other"; return nil, nil })
		if err != nil {
			t.Error(err)
		}
	}()
	<-ranking
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := d.SubmitRanked(ctx, "m", func(context.Context) (int, error) {
			return 0, nil
		}, func(context.Context) (any, error) { order <- "active"; return nil, nil })
		if err != nil {
			t.Error(err)
		}
	}()
	waitFor(t, func() bool { return d.Pending() == 2 })
	close(release)
	wg.Wait()
	select {
	case first := <-order:
		if first != "active" {
			t.Fatalf("lower-priority ordinary mutation overtook queued active player: first=%s", first)
		}
	case <-ctx.Done():
		t.Fatal("ordinary commands failed to dispatch before deadline")
	}
}
