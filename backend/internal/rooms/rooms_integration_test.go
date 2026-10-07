//go:build integration

package rooms

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
)

func fixture(t *testing.T) (*Store, *identity.Store, []identity.Session) {
	return fixtureWithMigration(t, Up)
}

func fixtureWithMigration(t *testing.T, migrate func(context.Context, *pgxpool.Pool) error) (*Store, *identity.Store, []identity.Session) {
	t.Helper()
	ctx := context.Background()
	url := os.Getenv("CGMS_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("CGMS_TEST_DATABASE_URL required")
	}
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("rooms_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	cfg.ConnConfig.RuntimeParams["application_name"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, err := admin.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		admin.Close()
		if err != nil {
			t.Error(err)
		}
	})
	if err = identity.Up(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err = matchstore.Up(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err = migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	ids := identity.New(pool)
	var actors []identity.Session
	for range 5 {
		a, err := ids.CreateGuest(ctx)
		if err != nil {
			t.Fatal(err)
		}
		actors = append(actors, a)
	}
	return New(pool, matchstore.New(pool, "rules-test"), "rules-test"), ids, actors
}

func TestBotRoomMigrationPreservesExistingHumanRoom(t *testing.T) {
	ctx := context.Background()
	s, _, actors := fixtureWithMigration(t, func(ctx context.Context, p *pgxpool.Pool) error {
		if _, err := p.Exec(ctx, "CREATE TABLE rooms_schema(checksum bytea PRIMARY KEY);"+schema); err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(schema))
		_, err := p.Exec(ctx, "INSERT INTO rooms_schema VALUES($1)", sum[:])
		return err
	})
	if _, err := s.pool.Exec(ctx, "INSERT INTO rooms VALUES('old-room',$1,3,3,'open',NULL);", actors[0].Account.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.pool.Exec(ctx, "INSERT INTO room_members VALUES('old-room',$1,1)", actors[0].Account.ID); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := Up(ctx, s.pool); err != nil {
			t.Fatal(err)
		}
	}
	r, err := s.Get(ctx, actors[0].Account.ID, "old-room")
	if err != nil || r.BotDifficulty != "" || len(r.Members) != 1 || r.Members[0].Bot {
		t.Fatal("human room changed", r, err)
	}
	if _, err = s.pool.Exec(ctx, "UPDATE rooms_migrations SET checksum=$1 WHERE name='bots-v1'", []byte("changed")); err != nil {
		t.Fatal(err)
	}
	if err = Up(ctx, s.pool); err == nil {
		t.Fatal("changed migration accepted")
	}
}

func TestBotRoomFailedCreationLeavesNoIdentitiesOrPartialRoster(t *testing.T) {
	s, _, actors := fixture(t)
	ctx := context.Background()
	owner := actors[0].Account.ID
	// Fail after the first bot identity and membership have been inserted.
	if _, err := s.pool.Exec(ctx, `ALTER TABLE room_members ADD CONSTRAINT fixture_reject_third_seat CHECK(seat<>3)`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateWithBots(ctx, owner, "retry", 4, 2, "advanced"); err == nil {
		t.Fatal("fixture failed to reject third seat")
	}
	var bots, rooms, receipts int
	if err := s.pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM identity_accounts WHERE kind='bot'),(SELECT count(*) FROM rooms),(SELECT count(*) FROM room_commands)`).Scan(&bots, &rooms, &receipts); err != nil || bots != 0 || rooms != 0 || receipts != 0 {
		t.Fatal("orphaned bot room state", bots, rooms, receipts, err)
	}
	if _, err := s.pool.Exec(ctx, `ALTER TABLE room_members DROP CONSTRAINT fixture_reject_third_seat`); err != nil {
		t.Fatal(err)
	}
	r, err := s.CreateWithBots(ctx, owner, "retry", 4, 2, "advanced")
	if err != nil || len(r.Members) != 4 {
		t.Fatal("original identity could not recover failed create", err)
	}
	if _, err = s.CreateWithBots(ctx, owner, "invalid", 3, 1, "omniscient"); !errors.Is(err, ErrInvalid) {
		t.Fatal("unapproved difficulty accepted", err)
	}
	if _, err = s.CreateWithBots(ctx, r.Members[1].ActorID, "bot-owner", 3, 1, "beginner"); !errors.Is(err, ErrForbidden) {
		t.Fatal("bot acquired human room authority", err)
	}
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM identity_accounts WHERE kind='bot'`).Scan(&bots); err != nil || bots != 3 {
		t.Fatal("unexpected bot identities", bots, err)
	}
}
func TestRoomIdempotencyCapacityAndAtomicStart(t *testing.T) {
	s, ids, actors := fixture(t)
	ctx := context.Background()
	owner := actors[0].Account.ID
	room, err := s.Create(ctx, owner, "create", 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	again, err := s.Create(ctx, owner, "create", 3, 0)
	if err != nil || again.ID != room.ID || room.Games != 3 {
		t.Fatal("create replay", err)
	}
	if _, err = s.Create(ctx, owner, "create", 4, 0); !errors.Is(err, ErrConflict) {
		t.Fatal("changed body accepted", err)
	}
	invite, err := s.Invite(ctx, owner, room.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Join(ctx, actors[1].Account.ID, room.ID, "invalid"); !errors.Is(err, ErrForbidden) {
		t.Fatal("invalid invite", err)
	}
	var wg sync.WaitGroup
	out := make(chan error, 4)
	for _, a := range actors[1:] {
		wg.Add(1)
		go func(id string) { defer wg.Done(); _, e := s.Join(ctx, id, room.ID, invite); out <- e }(a.Account.ID)
	}
	wg.Wait()
	close(out)
	good := 0
	for e := range out {
		if e == nil {
			good++
		} else if !errors.Is(e, ErrFull) {
			t.Fatal(e)
		}
	}
	if good != 2 {
		t.Fatal("room capacity race", good)
	}
	current, err := s.Get(ctx, owner, room.ID)
	if err != nil || len(current.Members) != 3 {
		t.Fatal("membership", err)
	}
	results := make(chan string, 2)
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() { defer wg.Done(); id, e := s.Start(ctx, owner, room.ID, "start"); results <- id; errs <- e }()
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	match := ""
	for id := range results {
		if match == "" {
			match = id
		} else if match != id {
			t.Fatal("duplicate match")
		}
	}
	env, _, err := s.matches.Restore(ctx, match)
	if err != nil || env.Match.GameLimit != 3 || len(env.Match.Game.Board.Players) != 3 {
		t.Fatal("match settings", err)
	}
	var count int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM matches`).Scan(&count); err != nil || count != 1 {
		t.Fatal("start not atomic", err)
	}
	for _, a := range actors {
		if a.Account.ID == current.Members[1].ActorID {
			if err = ids.Delete(ctx, a.Token); err != nil {
				t.Fatal(err)
			}
		}
	}
	if duplicate, err := s.Start(ctx, owner, room.ID, "start"); err != nil || duplicate != match {
		t.Fatal("committed start lost after participant deletion", err)
	}
	if _, err = s.Join(ctx, owner, room.ID, invite); !errors.Is(err, ErrClosed) {
		t.Fatal("late join", err)
	}
}
func TestRoomStartRollbackAndDeletedParticipant(t *testing.T) {
	s, ids, actors := fixture(t)
	ctx := context.Background()
	owner := actors[0].Account.ID
	r, err := s.Create(ctx, owner, "create", 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	invite, err := s.Invite(ctx, owner, r.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range actors[1:3] {
		if _, err = s.Join(ctx, a.Account.ID, r.ID, invite); err != nil {
			t.Fatal(err)
		}
	}
	// Abort after match creation via a database constraint, not a mocked store.
	if _, err = s.pool.Exec(ctx, `ALTER TABLE rooms ADD CONSTRAINT force_rollback CHECK(status<>'started')`); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Start(ctx, owner, r.ID, "start"); err == nil {
		t.Fatal("fault not injected")
	}
	var count int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FROM matches`).Scan(&count); err != nil || count != 0 {
		t.Fatal("orphan match after rollback", err)
	}
	if _, err = s.pool.Exec(ctx, `ALTER TABLE rooms DROP CONSTRAINT force_rollback`); err != nil {
		t.Fatal(err)
	}
	if err = ids.Delete(ctx, actors[1].Token); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Start(ctx, owner, r.ID, "start"); !errors.Is(err, ErrForbidden) {
		t.Fatal("deleted participant admitted", err)
	}
}

func TestConcurrentEconomyRoomStartsDoNotUpgradeSharedRosterLocks(t *testing.T) {
	s, _, actors := fixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := economy.Up(ctx, s.pool); err != nil {
		t.Fatal(err)
	}
	var err error
	s.matches, err = matchstore.NewWithEconomy(s.pool, "rules-test", economy.Policy{Reward: 7, PassDayPrice: 21})
	if err != nil {
		t.Fatal(err)
	}
	roomIDs := make([]string, 2)
	for i := range roomIDs {
		owner := actors[i].Account.ID
		r, err := s.Create(ctx, owner, "create", 3, 1)
		if err != nil {
			t.Fatal(err)
		}
		invite, err := s.Invite(ctx, owner, r.ID)
		if err != nil {
			t.Fatal(err)
		}
		for _, a := range actors[:3] {
			if _, err := s.Join(ctx, a.Account.ID, r.ID, invite); err != nil {
				t.Fatal(err)
			}
		}
		roomIDs[i] = r.ID
	}
	barrier, err := s.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer rollback(barrier)
	// Both starts encounter this lock before they can reserve an allowance. The
	// old SHARE-then-UPDATE path lets each retain locks that block the other's
	// upgrade; acquiring exclusive roster locks first serializes safely.
	if _, err := barrier.Exec(ctx, "SELECT account_id FROM identity_accounts ORDER BY account_id FOR SHARE"); err != nil {
		t.Fatal(err)
	}
	results := make(chan error, 2)
	for i, roomID := range roomIDs {
		go func(owner, id string) { _, err := s.Start(ctx, owner, id, "start"); results <- err }(actors[i].Account.ID, roomID)
	}
	for {
		var waiting int
		if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE application_name=current_setting('application_name') AND wait_event_type='Lock'").Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting == 2 {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if err := barrier.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	for range roomIDs {
		if err := <-results; err != nil {
			t.Errorf("overlapping enabled room starts must serialize, not fail: %v", err)
		}
	}
	var matches, starts, charges int
	if err := s.pool.QueryRow(ctx, "SELECT (SELECT count(*) FROM matches),(SELECT count(*) FROM economy_starts),(SELECT count(*) FROM economy_charges)").Scan(&matches, &starts, &charges); err != nil {
		t.Fatal(err)
	}
	if matches != 2 || starts != 2 || charges != 6 {
		t.Fatalf("starts not atomically admitted: matches=%d starts=%d charges=%d", matches, starts, charges)
	}
}
