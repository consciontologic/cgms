// Package delivery owns bounded, process-local online delivery and arbitration.
// PostgreSQL remains the authority; this dispatcher is not a multi-node leader.
package delivery

import (
	"context"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
	"sync"
)

var ErrOverloaded = errors.New("delivery overloaded")
var ErrClosed = errors.New("delivery closed")

type outcome struct {
	value any
	err   error
}
type job struct {
	victory bool
	ctx     context.Context
	cancel  context.CancelFunc
	stop    func() bool
	queued  func(error)
	rank    func(context.Context) (int, error)
	work    func(context.Context) (any, error)
	reply   chan outcome
}
type matchQueue struct {
	jobs  []*job
	count int
}
type Dispatcher struct {
	mu                           sync.Mutex
	queues                       map[string]*matchQueue
	total, maxTotal, maxPerMatch int
	closed                       bool
	ctx                          context.Context
	cancel                       context.CancelFunc
	wg                           sync.WaitGroup
	done                         chan struct{}
}

func New(maxTotal, maxPerMatch int) *Dispatcher {
	if maxTotal < 1 {
		maxTotal = 128
	}
	if maxPerMatch < 1 {
		maxPerMatch = 32
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Dispatcher{queues: map[string]*matchQueue{}, maxTotal: maxTotal, maxPerMatch: maxPerMatch, ctx: ctx, cancel: cancel, done: make(chan struct{})}
}
func (d *Dispatcher) Submit(ctx context.Context, match string, priority int, work func(context.Context) (any, error)) (any, error) {
	return d.submit(ctx, match, func(context.Context) (int, error) { return priority, nil }, work, priority == -1)
}

// SubmitRanked re-evaluates queued priorities between atomic work calls. Lower
// ranks run first; equal ranks retain server admission order. Rank and work must
// honor their context; no database callback runs under the queue mutex.
func (d *Dispatcher) SubmitRanked(ctx context.Context, match string, rank func(context.Context) (int, error), work func(context.Context) (any, error)) (any, error) {
	return d.submit(ctx, match, rank, work, false)
}
func (d *Dispatcher) submit(ctx context.Context, match string, rank func(context.Context) (int, error), work func(context.Context) (any, error), victory bool) (any, error) {
	d.mu.Lock()
	if d.closed {
		d.mu.Unlock()
		return nil, ErrClosed
	}
	if ctx.Err() != nil {
		d.mu.Unlock()
		return nil, ctx.Err()
	}
	if match == "" || rank == nil || work == nil {
		d.mu.Unlock()
		return nil, errors.New("invalid delivery job")
	}
	q := d.queues[match]
	if d.total >= d.maxTotal || (q != nil && q.count >= d.maxPerMatch) {
		d.mu.Unlock()
		return nil, ErrOverloaded
	}
	ctx, queued := telemetry.Start(ctx, telemetry.Queue)
	jc, cancel := context.WithCancel(ctx)
	j := &job{victory: victory, ctx: jc, cancel: cancel, rank: rank, work: work, reply: make(chan outcome, 1)}
	j.queued = queued
	j.stop = context.AfterFunc(d.ctx, cancel)
	fresh := q == nil
	if fresh {
		q = &matchQueue{}
		d.queues[match] = q
	}
	q.jobs = append(q.jobs, j)
	q.count++
	d.total++
	if fresh {
		d.wg.Add(1)
		go d.run(match, q)
	}
	d.mu.Unlock()
	select {
	case r := <-j.reply:
		return r.value, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-d.ctx.Done():
		return nil, ErrClosed
	}
}
func (d *Dispatcher) run(match string, q *matchQueue) {
	defer d.wg.Done()
	for {
		d.mu.Lock()
		if len(q.jobs) == 0 {
			delete(d.queues, match)
			d.mu.Unlock()
			return
		}
		candidates := append([]*job(nil), q.jobs...)
		d.mu.Unlock()
		pick := 0
		best := int(^uint(0) >> 1)
		var selectedErr error
		victory := false
		for i, j := range candidates {
			if j.victory && j.ctx.Err() == nil {
				pick = i
				victory = true
				break
			}
		}
		for i, j := range candidates {
			if victory {
				break
			}
			err := j.ctx.Err()
			rank := 0
			if err == nil {
				rank, err = j.rank(j.ctx)
			}
			if err != nil {
				pick = i
				selectedErr = err
				break
			}
			if rank < best {
				best = rank
				pick = i
			}
		}
		j := candidates[pick]
		d.mu.Lock()
		// Ranking may read the database without holding this mutex. A victory can
		// arrive during those reads; select against the live queue at the atomic
		// dispatch boundary so an ordinary mutation cannot overtake it.
		for _, candidate := range q.jobs {
			if candidate.victory && candidate.ctx.Err() == nil {
				j = candidate
				selectedErr = nil
				break
			}
		}
		// New ordinary intents also participate in active-player/turn priority.
		// Only this worker removes jobs, so growth means the ranked snapshot is
		// incomplete. Re-rank before dispatch; the per-match admission bound
		// bounds these restarts. A queued victory still bypasses rank queries.
		if !j.victory && selectedErr == nil && len(q.jobs) > len(candidates) {
			d.mu.Unlock()
			continue
		}
		for i, c := range q.jobs {
			if c == j {
				q.jobs = append(q.jobs[:i], q.jobs[i+1:]...)
				break
			}
		}
		d.mu.Unlock()
		var r outcome
		r.err = selectedErr
		if r.err == nil {
			r.err = j.ctx.Err()
		}
		j.queued(r.err)
		if r.err == nil {
			jobCtx, complete := telemetry.Start(j.ctx, telemetry.Delivery)
			r.value, r.err = j.work(jobCtx)
			complete(r.err)
		}
		j.stop()
		j.cancel()
		j.reply <- r
		d.mu.Lock()
		q.count--
		d.total--
		d.mu.Unlock()
	}
}
func (d *Dispatcher) Pending() int { d.mu.Lock(); defer d.mu.Unlock(); return d.total }
func (d *Dispatcher) Close(ctx context.Context) error {
	d.mu.Lock()
	if !d.closed {
		d.closed = true
		d.cancel()
		go func() { d.wg.Wait(); close(d.done) }()
	}
	d.mu.Unlock()
	select {
	case <-d.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
