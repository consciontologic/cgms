//go:build integration

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
	"io"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func automaticDeliveryCount(t *testing.T, recorder *telemetry.Recorder, outcome string) uint64 {
	t.Helper()
	w := httptest.NewRecorder()
	recorder.Metrics(w, nil)
	prefix := fmt.Sprintf("cgms_boundary_total{boundary=\"delivery\",outcome=%q} ", outcome)
	for _, line := range strings.Split(w.Body.String(), "\n") {
		if value, ok := strings.CutPrefix(line, prefix); ok {
			count, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return count
		}
	}
	t.Fatal("automatic delivery metric missing")
	return 0
}

func TestSupervisorSkipsUnchangedIdleMatches(t *testing.T) {
	s := journeyServer(t)
	ctx := context.Background()
	clients := journeyGuests(t, s, 3)
	board, err := game.NewState(3, "idle-game")
	if err != nil {
		t.Fatal(err)
	}
	board.Order = []int{1, 2, 3}
	board.Active = 1
	m, err := game.NewMatchLifecycle(board, 1)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.TurnStarted = true
	env, err := matchstore.NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if _, automatic := matchstore.NextServer(env, 0); automatic {
		t.Fatal("idle fixture must have no pending automatic transition")
	}
	if err = s.matches.Create(ctx, "idle", env, []string{clients[0].Account.ID, clients[1].Account.ID, clients[2].Account.ID}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a recoverable restore failure without changing the match version.
	// A failed attempt must remain eligible on subsequent scans.
	if _, err = s.cfg.Pool.Exec(ctx, "UPDATE matches SET snapshot=$2 WHERE match_id=$1", "idle", []byte("{}")); err != nil {
		t.Fatal(err)
	}
	recorder := telemetry.New(io.Discard, false)
	s.ctx = recorder.Context(s.ctx)
	s.Start()
	waitFor := func(recorder *telemetry.Recorder, outcome string, minimum uint64) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for automaticDeliveryCount(t, recorder, outcome) < minimum {
			if time.Now().After(deadline) {
				t.Fatalf("automatic %s deliveries did not reach %d", outcome, minimum)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	waitFor(recorder, "error", 2)
	if _, err = s.cfg.Pool.Exec(ctx, "UPDATE matches SET snapshot=$2 WHERE match_id=$1", "idle", snapshot); err != nil {
		t.Fatal(err)
	}
	waitFor(recorder, "ok", 1)
	time.Sleep(250 * time.Millisecond)
	if count := automaticDeliveryCount(t, recorder, "ok"); count != 1 {
		t.Fatalf("unchanged idle match was dispatched %d times; want one check", count)
	}
	// An explicit entitlement wake conservatively rechecks even an unchanged head.
	s.wakeAutomatic()
	waitFor(recorder, "ok", 2)
	time.Sleep(250 * time.Millisecond)
	if count := automaticDeliveryCount(t, recorder, "ok"); count != 2 {
		t.Fatalf("entitlement wake dispatched idle match %d times; want one additional check", count)
	}
	if err = s.Close(ctx); err != nil {
		t.Fatal(err)
	}
	// Restart discards only the optimization and discovers durable idle state anew.
	cfg := s.cfg
	cfg.Telemetry = telemetry.New(io.Discard, false)
	restarted, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := restarted.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	restarted.Start()
	waitFor(cfg.Telemetry, "ok", 1)
	time.Sleep(250 * time.Millisecond)
	if count := automaticDeliveryCount(t, cfg.Telemetry, "ok"); count != 1 {
		t.Fatalf("restarted idle match was dispatched %d times; want one check", count)
	}
}

func TestRuntimeKeyAndExclusiveAuthority(t *testing.T) {
	s := journeyServer(t)
	ctx := context.Background()
	a, e := PersistentKey(ctx, s.cfg.Pool)
	if e != nil {
		t.Fatal(e)
	}
	b, e := PersistentKey(ctx, s.cfg.Pool)
	if e != nil || len(a) != 32 || !bytes.Equal(a, b) {
		t.Fatal("key changed", e)
	}
	lease, e := AcquireAuthority(ctx, s.cfg.Pool)
	if e != nil {
		t.Fatal(e)
	}
	if x, e := AcquireAuthority(ctx, s.cfg.Pool); e == nil {
		x.Close()
		t.Fatal("second authority admitted")
	}
	lease.Close()
	lease, e = AcquireAuthority(ctx, s.cfg.Pool)
	if e != nil {
		t.Fatal(e)
	}
	lease.Close()
}
func TestSupervisorStartsButNeverPlaysForHumans(t *testing.T) {
	s := journeyServer(t)
	ctx := context.Background()
	actors := make([]identity.Session, 3)
	for i := range actors {
		var e error
		actors[i], e = s.identity.CreateGuest(ctx)
		if e != nil {
			t.Fatal(e)
		}
	}
	room, e := s.rooms.Create(ctx, actors[0].Account.ID, "create", 3, 3)
	if e != nil {
		t.Fatal(e)
	}
	invite, e := s.rooms.Invite(ctx, actors[0].Account.ID, room.ID)
	if e != nil {
		t.Fatal(e)
	}
	for _, a := range actors[1:] {
		if _, e = s.rooms.Join(ctx, a.Account.ID, room.ID, invite); e != nil {
			t.Fatal(e)
		}
	}
	id, e := s.rooms.Start(ctx, actors[0].Account.ID, room.ID, "start")
	if e != nil {
		t.Fatal(e)
	}
	s.Start()
	deadline := time.Now().Add(5 * time.Second)
	var version int64
	for {
		env, v, e := s.matches.Restore(ctx, id)
		if e != nil {
			t.Fatal(e)
		}
		if env.Match.Game.TurnStarted {
			version = v
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("automatic startup did not reach human turn")
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(300 * time.Millisecond)
	_, v, e := s.matches.Restore(ctx, id)
	if e != nil || v != version {
		t.Fatal("disconnected human was auto-played", v, version, e)
	}
	if e = s.matches.VerifyJournal(ctx, id); e != nil {
		t.Fatal(e)
	}
}
func TestConcurrentStartAndClose(t *testing.T) {
	s := journeyServer(t)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			s.Start()
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if e := s.Close(ctx); e != nil {
		t.Fatal(e)
	}
	<-done
	s.Start()
}
func TestDurableEventGapAndMembership(t *testing.T) {
	s := journeyServer(t)
	ctx := context.Background()
	clients := journeyGuests(t, s, 3)
	board, e := game.NewState(3, "events-game")
	if e != nil {
		t.Fatal(e)
	}
	m, e := game.NewMatchLifecycle(board, 3)
	if e != nil {
		t.Fatal(e)
	}
	env, e := matchstore.NewEnvelope(m, "rules-test")
	if e != nil {
		t.Fatal(e)
	}
	actors := []string{}
	for _, c := range clients {
		actors = append(actors, c.Account.ID)
	}
	if e = s.matches.Create(ctx, "events", env, actors); e != nil {
		t.Fatal(e)
	}
	next, ok := matchstore.NextServer(env, 0)
	if !ok {
		t.Fatal("no deal")
	}
	if _, e = s.matches.Continue(ctx, "events", next); e != nil {
		t.Fatal(e)
	}
	events, e := s.matches.Events(ctx, "events", actors[0], 0, 128)
	if e != nil || len(events) != 1 || events[0].Cursor != 1 {
		t.Fatal(events, e)
	}
	if _, e = s.matches.Events(ctx, "events", "outsider", 0, 128); !errors.Is(e, matchstore.ErrUnauthorized) {
		t.Fatal("outsider", e)
	}
	if _, e = s.matches.Events(ctx, "events", actors[0], 2, 128); !errors.Is(e, matchstore.ErrCursor) {
		t.Fatal("future cursor", e)
	}
	if _, e = s.cfg.Pool.Exec(ctx, "DELETE FROM seat_events WHERE match_id='events' AND seat=1"); e != nil {
		t.Fatal(e)
	}
	if _, e = s.matches.Events(ctx, "events", actors[0], 0, 128); !errors.Is(e, matchstore.ErrCursor) {
		t.Fatal("missing durable event", e)
	}
}

func TestSupervisorRetainsFinalizedGameAndWakesAfterAllowanceRedemption(t *testing.T) {
	s := journeyServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := economy.Up(ctx, s.cfg.Pool); err != nil {
		t.Fatal(err)
	}
	policy := economy.Policy{Reward: 10, PassDayPrice: 30}
	var err error
	s.matches, err = matchstore.NewWithEconomy(s.cfg.Pool, "rules-test", policy)
	if err != nil {
		t.Fatal(err)
	}
	service, err := economy.NewService(s.cfg.Pool, policy)
	if err != nil {
		t.Fatal(err)
	}
	clients := journeyGuests(t, s, 3)
	actors := []string{clients[0].Account.ID, clients[1].Account.ID, clients[2].Account.ID}
	board, err := game.NewState(3, "allowance-predecessor")
	if err != nil {
		t.Fatal(err)
	}
	m, err := game.NewMatchLifecycle(board, 2)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.Board, err = game.CloseCardBoard(m.Game.Board)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.Ledger, err = m.Game.Ledger.CloseOrdinary(nil)
	if err != nil {
		t.Fatal(err)
	}
	m.Game.Promises, err = m.Game.Promises.CloseBoard()
	if err != nil {
		t.Fatal(err)
	}
	m.Game.Ending = "ordinary"
	for seat := 1; seat <= 3; seat++ {
		m.Game, err = m.Game.FinishSettlement(seat)
		if err != nil {
			t.Fatal(err)
		}
	}
	m.Game, err = m.Game.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	env, err := matchstore.NewEnvelope(m, "rules-test")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.matches.Create(ctx, "allowance-match", env, actors); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		tx, err := s.cfg.Pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := economy.ReserveStartTx(ctx, tx, fmt.Sprintf("other-start-%d", i), actors, time.Now()); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	s.Start()
	time.Sleep(200 * time.Millisecond)
	view, err := s.matches.Snapshot(ctx, "allowance-match", actors[0])
	if err != nil || !view.AdmissionWaiting || view.Version != 0 {
		t.Fatal("allowance denial must retain the finalized predecessor", view.Version, view.AdmissionWaiting, err)
	}
	// This is explicit test funding, never a client-controlled balance route.
	if _, err := s.cfg.Pool.Exec(ctx, "UPDATE economy_accounts SET dirt=30"); err != nil {
		t.Fatal(err)
	}
	for _, actor := range actors {
		if _, err := service.BuyPass(ctx, actor, "pass", 1); err != nil {
			t.Fatal(err)
		}
	}
	s.wakeAutomatic()
	deadline := time.Now().Add(3 * time.Second)
	for {
		restored, _, err := s.matches.Restore(ctx, "allowance-match")
		if err != nil {
			t.Fatal(err)
		}
		if restored.Match.Game.Board.GameID != "allowance-predecessor" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("redemption did not wake the paused next game before its retry delay")
		}
		time.Sleep(10 * time.Millisecond)
	}
	var starts, charges int
	if err := s.cfg.Pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM economy_starts),(SELECT count(*) FROM economy_charges)").Scan(&starts, &charges); err != nil || starts != 4 || charges != 12 {
		t.Fatal("next game did not reserve its roster exactly once", starts, charges, err)
	}
}
