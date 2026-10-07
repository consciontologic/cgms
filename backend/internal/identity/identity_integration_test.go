//go:build integration

package identity

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	return testStoreWithMigration(t, Up)
}

func testStoreWithMigration(t *testing.T, migrate func(context.Context, *pgxpool.Pool) error) *Store {
	t.Helper()
	url := os.Getenv("CGMS_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("CGMS_TEST_DATABASE_URL required")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("identity_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
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
	if err = migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	return New(pool)
}

func TestBotIdentityMigrationPreservesGuestsAndRejectsAuthentication(t *testing.T) {
	ctx := context.Background()
	s := testStoreWithMigration(t, func(ctx context.Context, p *pgxpool.Pool) error {
		if _, err := p.Exec(ctx, "CREATE TABLE identity_schema(checksum bytea PRIMARY KEY);"+schema); err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(schema))
		_, err := p.Exec(ctx, "INSERT INTO identity_schema VALUES($1)", sum[:])
		return err
	})
	guest, err := s.CreateGuest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = Up(ctx, s.pool); err != nil {
			t.Fatal(err)
		}
	}
	if account, err := s.Authenticate(ctx, guest.Token); err != nil || account.ID != guest.Account.ID {
		t.Fatal("migration invalidated guest", err)
	}
	if _, err = s.pool.Exec(ctx, "INSERT INTO identity_accounts(account_id,kind) VALUES('server-bot','bot')"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.pool.Exec(ctx, "UPDATE identity_accounts SET username='bot',password_hash=$1 WHERE account_id='server-bot'", []byte("not-credentials")); err == nil {
		t.Fatal("bot credentials accepted")
	}
	// Even a corrupted session row cannot turn a server opponent into a login.
	token, err := opaque()
	if err != nil {
		t.Fatal(err)
	}
	hash, err := tokenHash(token)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.pool.Exec(ctx, "INSERT INTO identity_sessions VALUES($1,'server-bot',$2)", hash, time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, token); !errors.Is(err, ErrCredentials) {
		t.Fatal("bot authenticated", err)
	}
	if _, err = s.Upgrade(ctx, token, "bot_account", "not a valid bot login 123"); !errors.Is(err, ErrConflict) {
		t.Fatal("bot upgraded", err)
	}
	if _, err = s.pool.Exec(ctx, "UPDATE identity_migrations SET checksum=$1 WHERE name='bots-v1'", []byte("tampered")); err != nil {
		t.Fatal(err)
	}
	if err = Up(ctx, s.pool); err == nil {
		t.Fatal("changed migration accepted")
	}
}
func TestGuestUpgradeExpiryRevocationDeletion(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	guest, err := s.CreateGuest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	a, err := s.Authenticate(ctx, guest.Token)
	if err != nil || a.ID != guest.Account.ID || a.Kind != "guest" {
		t.Fatal("guest auth", err)
	}
	member, err := s.Upgrade(ctx, guest.Token, "Player_One", "a secure passphrase 123")
	if err != nil || member.Account.ID != a.ID || member.Account.Kind != "member" {
		t.Fatal("upgrade changed identity", err)
	}
	if _, err = s.Authenticate(ctx, guest.Token); !errors.Is(err, ErrCredentials) {
		t.Fatal("old guest token survived", err)
	}
	logged, err := s.Login(ctx, "player_one", "a secure passphrase 123")
	if err != nil || logged.Account.ID != a.ID {
		t.Fatal("login", err)
	}
	if _, err = s.Login(ctx, "player_one", "not the password 123"); !errors.Is(err, ErrCredentials) {
		t.Fatal("wrong password accepted")
	}
	if _, err = s.Register(ctx, "PLAYER_ONE", "a second passphrase 456"); !errors.Is(err, ErrConflict) {
		t.Fatal("duplicate identity", err)
	}
	if err = s.Logout(ctx, logged.Token); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, logged.Token); !errors.Is(err, ErrCredentials) {
		t.Fatal("logout ineffective")
	}
	if _, err = s.pool.Exec(ctx, `UPDATE identity_sessions SET expires_at=now()-interval '1 second'`); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, member.Token); !errors.Is(err, ErrCredentials) {
		t.Fatal("expired token accepted")
	}
	logged, err = s.Login(ctx, "player_one", "a secure passphrase 123")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Delete(ctx, logged.Token); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Authenticate(ctx, logged.Token); !errors.Is(err, ErrCredentials) {
		t.Fatal("deleted session accepted")
	}
	if _, err = s.Login(ctx, "player_one", "a secure passphrase 123"); !errors.Is(err, ErrCredentials) {
		t.Fatal("deleted credentials accepted")
	}
	var kind string
	var username *string
	var password []byte
	if err = s.pool.QueryRow(ctx, `SELECT kind,username,password_hash FROM identity_accounts WHERE account_id=$1`, a.ID).Scan(&kind, &username, &password); err != nil || kind != "deleted" || username != nil || password != nil {
		t.Fatal("deletion retained credentials", err)
	}
}

func TestConcurrentGuestUpgradeNeverMergesAccounts(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	a, err := s.CreateGuest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.CreateGuest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		session Session
		err     error
	}
	ch := make(chan outcome, 2)
	for _, guest := range []Session{a, b} {
		go func(guest Session) {
			out, e := s.Upgrade(ctx, guest.Token, "same_name", "a secure long password 123")
			ch <- outcome{out, e}
		}(guest)
	}
	success := 0
	for range 2 {
		r := <-ch
		if r.err == nil {
			success++
			if r.session.Account.ID != a.Account.ID && r.session.Account.ID != b.Account.ID {
				t.Fatal("upgrade replaced actor")
			}
		} else if !errors.Is(r.err, ErrConflict) {
			t.Fatal(r.err)
		}
	}
	if success != 1 {
		t.Fatal("credential collision admitted", success)
	}
	var guests, members int
	if err = s.pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE kind='guest'),count(*) FILTER(WHERE kind='member') FROM identity_accounts`).Scan(&guests, &members); err != nil || guests != 1 || members != 1 {
		t.Fatal("merged or partial account", guests, members, err)
	}
}
