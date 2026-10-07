//go:build integration

package economy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
)

func economyDB(t *testing.T) (*pgxpool.Pool, []string) {
	return economyDBWithMigration(t, Up)
}
func economyDBWithMigration(t *testing.T, migrate func(context.Context, *pgxpool.Pool) error) (*pgxpool.Pool, []string) {
	t.Helper()
	ctx := context.Background()
	url := os.Getenv("CGMS_TEST_DATABASE_URL")
	if url == "" {
		t.Fatal("CGMS_TEST_DATABASE_URL required")
	}
	admin, e := pgxpool.New(ctx, url)
	if e != nil {
		t.Fatal(e)
	}
	schema := fmt.Sprintf("economy_%d", time.Now().UnixNano())
	if _, e = admin.Exec(ctx, "CREATE SCHEMA "+schema); e != nil {
		t.Fatal(e)
	}
	cfg, e := pgxpool.ParseConfig(url)
	if e != nil {
		t.Fatal(e)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	p, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		p.Close()
		if _, e := admin.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); e != nil {
			t.Error(e)
		}
		admin.Close()
	})
	if e = identity.Up(ctx, p); e != nil {
		t.Fatal(e)
	}
	if e = migrate(ctx, p); e != nil {
		t.Fatal(e)
	}
	if e = migrate(ctx, p); e != nil {
		t.Fatal("repeat migration", e)
	}
	actors := []string{}
	for range 3 {
		s, e := identity.New(p).CreateGuest(ctx)
		if e != nil {
			t.Fatal(e)
		}
		actors = append(actors, s.Account.ID)
	}
	return p, actors
}
func transact(p *pgxpool.Pool, fn func(pgx.Tx) error) error {
	ctx := context.Background()
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if e = fn(tx); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

func TestAtomicAllowanceRetryRolloverAndRollback(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 30, 23, 59, 0, 0, time.UTC)
	var admitted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			e := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, fmt.Sprintf("game-%d", i), actors, now) })
			if e == nil {
				admitted.Add(1)
			} else if !errors.Is(e, ErrAllowance) {
				t.Error(e)
			}
		}(i)
	}
	wg.Wait()
	if admitted.Load() != 3 {
		t.Fatal("overspent allowance", admitted.Load())
	}
	var id string
	if e := p.QueryRow(ctx, "SELECT game_id FROM economy_starts LIMIT 1").Scan(&id); e != nil {
		t.Fatal(e)
	}
	if e := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, id, actors, now) }); e != nil {
		t.Fatal("retry charged again", e)
	}
	if e := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, "tomorrow", actors, now.Add(time.Minute)) }); e != nil {
		t.Fatal("UTC reset", e)
	}
	sentinel := errors.New("outer game transaction failed")
	if e := transact(p, func(tx pgx.Tx) error {
		if e := ReserveStartTx(ctx, tx, "rollback", actors, now.Add(time.Minute)); e != nil {
			return e
		}
		return sentinel
	}); !errors.Is(e, sentinel) {
		t.Fatal(e)
	}
	var count int
	if e := p.QueryRow(ctx, "SELECT count(*) FROM economy_starts WHERE game_id='rollback'").Scan(&count); e != nil || count != 0 {
		t.Fatal("orphan allowance", e, count)
	}
}

func TestFinalRewardsAndPassOperationsAreIdempotent(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	policy := Policy{Reward: 300, PassDayPrice: 30}
	ledger := game.NewFinancialLedger("game", 3)
	if e := transact(p, func(tx pgx.Tx) error { return RewardFinalizedTx(ctx, tx, ledger, policy) }); !errors.Is(e, ErrIneligible) {
		t.Fatal("unfinished/offline reward", e)
	}
	if e := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, "game", actors, now) }); e != nil {
		t.Fatal(e)
	}
	var e error
	ledger, e = ledger.CloseOrdinary(nil)
	if e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 3; i++ {
		ledger, e = ledger.FinishSettlement(i)
		if e != nil {
			t.Fatal(e)
		}
	}
	ledger, e = ledger.Finalize()
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := transact(p, func(tx pgx.Tx) error { return RewardFinalizedTx(ctx, tx, ledger, policy) }); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	for _, actor := range actors {
		var balance int64
		if e = p.QueryRow(ctx, "SELECT dirt FROM economy_accounts WHERE account_id=$1", actor).Scan(&balance); e != nil || balance != 300 {
			t.Fatal("duplicate reward", balance, e)
		}
	}
	var firstExpiry time.Time
	buy := func(op string, days int) error {
		return transact(p, func(tx pgx.Tx) error {
			expiry, e := BuyPassTx(ctx, tx, actors[0], op, days, now.Add(123456789*time.Nanosecond), policy)
			if e == nil && op == "pass" {
				if !firstExpiry.IsZero() && !expiry.Equal(firstExpiry) {
					t.Error("retry changed exact expiry", firstExpiry, expiry)
				}
				firstExpiry = expiry
			}
			return e
		})
	}
	if e = buy("pass", 7); e != nil {
		t.Fatal(e)
	}
	if e = buy("pass", 7); e != nil {
		t.Fatal("duplicate pass", e)
	}
	if e = buy("pass", 1); !errors.Is(e, ErrConflict) {
		t.Fatal("operation body changed", e)
	}
	if e = buy("too-much", 7); !errors.Is(e, ErrFunds) {
		t.Fatal("overspent", e)
	}
	if e = buy("next", 1); e != nil {
		t.Fatal("stack", e)
	}
	var balance int64
	var expiry time.Time
	if e = p.QueryRow(ctx, "SELECT dirt,pass_until FROM economy_accounts WHERE account_id=$1", actors[0]).Scan(&balance, &expiry); e != nil || balance != 60 || !expiry.Equal(now.Add(8*24*time.Hour+123456*time.Microsecond)) {
		t.Fatal("pass stacking", balance, expiry, e)
	}
	// Premium/pass starts remain recorded, with no free-allowance charge.
	for i := 0; i < 3; i++ {
		if e = transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, fmt.Sprintf("after-%d", i), actors, now) }); i < 2 && e != nil {
			t.Fatal(e)
		}
		if i == 2 && !errors.Is(e, ErrAllowance) {
			t.Fatal("mixed roster must reject exhausted free accounts", e)
		}
	}
	var rejectedRows int
	if e = p.QueryRow(ctx, "SELECT (SELECT count(*) FROM economy_starts WHERE game_id='after-2')+(SELECT count(*) FROM economy_charges WHERE game_id='after-2')").Scan(&rejectedRows); e != nil || rejectedRows != 0 {
		t.Fatal("partial mixed-roster charge", rejectedRows, e)
	}
	var free int
	if e = p.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE account_id=$1 AND charged_free", actors[0]).Scan(&free); e != nil || free != 1 {
		t.Fatal("pass spent free starts", free, e)
	}
}

func TestLegacyPassMigrationPreservesReceiptsAndAdExpiry(t *testing.T) {
	legacy := func(ctx context.Context, p *pgxpool.Pool) error {
		var exists bool
		if e := p.QueryRow(ctx, "SELECT to_regclass('economy_schema') IS NOT NULL").Scan(&exists); e != nil {
			return e
		}
		if exists {
			return nil
		}
		sum := sha256.Sum256([]byte(schema))
		if _, e := p.Exec(ctx, "CREATE TABLE economy_schema(checksum bytea PRIMARY KEY);"+schema); e != nil {
			return e
		}
		_, e := p.Exec(ctx, "INSERT INTO economy_schema VALUES($1)", sum[:])
		return e
	}
	p, actors := economyDBWithMigration(t, legacy)
	ctx := context.Background()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	oldExpiry := now.Add(2 * 24 * time.Hour)
	if _, e := p.Exec(ctx, "INSERT INTO economy_accounts(account_id,dirt,pass_until) VALUES($1,300,$2)", actors[0], oldExpiry); e != nil {
		t.Fatal(e)
	}
	for days := 2; days <= 6; days++ {
		if _, e := p.Exec(ctx, "INSERT INTO economy_pass_operations(account_id,operation_id,days,cost,expires_at) VALUES($1,$2,$3,$4,$5)", actors[0], fmt.Sprint("legacy-", days), days, days*30, oldExpiry); e != nil {
			t.Fatal(e)
		}
	}
	if e := Up(ctx, p); e != nil {
		t.Fatal(e)
	}
	var installed []byte
	if e := p.QueryRow(ctx, "SELECT checksum FROM economy_schema").Scan(&installed); e != nil {
		t.Fatal(e)
	}
	oldSum := sha256.Sum256([]byte(schema))
	if bytes.Equal(installed, oldSum[:]) {
		t.Fatal("legacy binary could reopen changed economy policy")
	}
	s, e := NewService(p, AdoptedPolicy())
	if e != nil {
		t.Fatal(e)
	}
	s.now = func() time.Time { return now }
	for days := 2; days <= 6; days++ {
		op := fmt.Sprint("legacy-", days)
		receipt, e := s.LookupPass(ctx, actors[0], op)
		if e != nil || receipt.Days != days || receipt.Cost != int64(days*30) || !receipt.ExpiresAt.Equal(oldExpiry) {
			t.Fatal("legacy receipt", receipt, e)
		}
		replay, e := s.BuyPass(ctx, actors[0], op, days)
		if e != nil || replay != receipt {
			t.Fatal("legacy replay", replay, e)
		}
		if _, e = s.BuyPass(ctx, actors[0], fmt.Sprint("new-", days), days); !errors.Is(e, ErrInvalid) {
			t.Fatal("new retired purchase", days, e)
		}
	}
	a, e := s.Account(ctx, actors[0])
	if e != nil || a.Benefits.Dirt != 300 || !a.Benefits.Unlimited || !a.Benefits.NoAds {
		t.Fatal("legacy promise", a, e)
	}
	if _, e = s.BuyPass(ctx, actors[0], "new-weekly", 7); e != nil {
		t.Fatal(e)
	}
	if e = Up(ctx, p); e != nil {
		t.Fatal("repeat migration", e)
	}
	now = oldExpiry
	a, e = s.Account(ctx, actors[0])
	if e != nil || a.Benefits.Dirt != 90 || !a.Benefits.Unlimited || a.Benefits.NoAds {
		t.Fatal("new earned time extended old ad promise", a, e)
	}
}

func TestAdFreeMigrationPreservesPriorPolicyAndRejectsOldWriter(t *testing.T) {
	priorSum := sha256.Sum256([]byte(schema + "\n" + accessProductsMigration))
	prior := func(ctx context.Context, p *pgxpool.Pool) error {
		var exists bool
		if err := p.QueryRow(ctx, "SELECT to_regclass('economy_schema') IS NOT NULL").Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		migrationSum := sha256.Sum256([]byte(accessProductsMigration))
		if _, err := p.Exec(ctx, "CREATE TABLE economy_schema(checksum bytea PRIMARY KEY);"+schema+accessProductsMigration+"CREATE TABLE economy_migrations(name text PRIMARY KEY,checksum bytea NOT NULL);"); err != nil {
			return err
		}
		if _, err := p.Exec(ctx, "INSERT INTO economy_schema VALUES($1)", priorSum[:]); err != nil {
			return err
		}
		_, err := p.Exec(ctx, "INSERT INTO economy_migrations VALUES('access-products-v1',$1)", migrationSum[:])
		return err
	}
	p, actors := economyDBWithMigration(t, prior)
	ctx := context.Background()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	legacyExpiry, passExpiry := now.Add(time.Hour), now.Add(7*24*time.Hour)
	if _, err := p.Exec(ctx, "INSERT INTO economy_accounts(account_id,dirt,pass_until,legacy_ad_free_until) VALUES($1,300,$2,$3)", actors[0], passExpiry, legacyExpiry); err != nil {
		t.Fatal(err)
	}
	if _, err := p.Exec(ctx, "INSERT INTO economy_pass_operations(account_id,operation_id,days,cost,expires_at) VALUES($1,'historical-2',2,60,$2)", actors[0], passExpiry); err != nil {
		t.Fatal(err)
	}
	if err := Up(ctx, p); err != nil {
		t.Fatal(err)
	}
	var installed []byte
	if err := p.QueryRow(ctx, "SELECT checksum FROM economy_schema").Scan(&installed); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(installed, priorSum[:]) {
		t.Fatal("previous binary can reopen standalone ad-free policy")
	}
	if err := Up(ctx, p); err != nil {
		t.Fatal("idempotent migration", err)
	}
	b, err := ReadBenefits(ctx, p, actors[0], now)
	if err != nil || b.Dirt != 300 || !b.PassUntil.Equal(passExpiry) || !b.LegacyAdFreeUntil.Equal(legacyExpiry) {
		t.Fatal("migration rewrote prior benefits", b, err)
	}
	assertAdFreeUntil(t, b, legacyExpiry)
	s, err := NewService(p, AdoptedPolicy())
	if err != nil {
		t.Fatal(err)
	}
	s.now = func() time.Time { return now }
	if r, err := s.BuyPass(ctx, actors[0], "historical-2", 2); err != nil || r.Cost != 60 || !r.ExpiresAt.Equal(passExpiry) {
		t.Fatal("historical receipt replay", r, err)
	}
	if _, err := s.BuyPass(ctx, actors[0], "retired-2", 2); !errors.Is(err, ErrProductUnavailable) {
		t.Fatal("migration reopened retired pass", err)
	}
	if _, err := p.Exec(ctx, "UPDATE economy_migrations SET checksum=$1 WHERE name='ad-free-products-v1'", []byte("tampered")); err != nil {
		t.Fatal(err)
	}
	if err := Up(ctx, p); err == nil {
		t.Fatal("mismatched ad-free migration checksum accepted")
	}
}

func TestFixedTermMigrationPreservesHistoricalAdFreeAndRejectsOldWriter(t *testing.T) {
	priorSum := sha256.Sum256([]byte(schema + "\n" + accessProductsMigration + "\n" + adFreeProductsMigration))
	prior := func(ctx context.Context, p *pgxpool.Pool) error {
		var exists bool
		if err := p.QueryRow(ctx, "SELECT to_regclass('economy_schema') IS NOT NULL").Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		if _, err := p.Exec(ctx, "CREATE TABLE economy_schema(checksum bytea PRIMARY KEY);"+schema+accessProductsMigration+adFreeProductsMigration+"CREATE TABLE economy_migrations(name text PRIMARY KEY,checksum bytea NOT NULL);"); err != nil {
			return err
		}
		if _, err := p.Exec(ctx, "INSERT INTO economy_schema VALUES($1)", priorSum[:]); err != nil {
			return err
		}
		for name, migration := range map[string]string{"access-products-v1": accessProductsMigration, "ad-free-products-v1": adFreeProductsMigration} {
			sum := sha256.Sum256([]byte(migration))
			if _, err := p.Exec(ctx, "INSERT INTO economy_migrations VALUES($1,$2)", name, sum[:]); err != nil {
				return err
			}
		}
		return nil
	}
	p, actors := economyDBWithMigration(t, prior)
	ctx := context.Background()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	old := PurchaseGrant{VerifiedPurchase: VerifiedPurchase{Provider: "apple", Environment: "production", App: "fixture.app", Product: "ad-free", Account: actors[0], Transaction: "legacy-ad-free", Revision: 1, Status: "purchased", ExpiresAt: now.Add(9 * time.Hour)}, Kind: "ad_free"}
	body, err := json.Marshal(old)
	if err != nil {
		t.Fatal(err)
	}
	var historical map[string]json.RawMessage
	if err = json.Unmarshal(body, &historical); err != nil {
		t.Fatal(err)
	}
	delete(historical, "ConfirmedAt")
	body, err = json.Marshal(historical)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = p.Exec(ctx, "INSERT INTO economy_accounts(account_id,dirt,pass_until,legacy_ad_free_until) VALUES($1,30,$2,$3)", actors[0], now.Add(24*time.Hour), now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Exec(ctx, `INSERT INTO economy_provider_purchases(provider,environment,app,transaction_id,account_id,kind,cosmetic,status,expires_at,grant_state) VALUES($1,$2,$3,$4,$5,$6,'',$7,$8,$9)`, old.Provider, old.Environment, old.App, old.Transaction, old.Account, old.Kind, old.Status, old.ExpiresAt, body); err != nil {
		t.Fatal(err)
	}
	var before, after []byte
	if err = p.QueryRow(ctx, "SELECT grant_state FROM economy_provider_purchases").Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err = Up(ctx, p); err != nil {
		t.Fatal(err)
	}
	var installed []byte
	if err = p.QueryRow(ctx, "SELECT checksum FROM economy_schema").Scan(&installed); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(installed, priorSum[:]) {
		t.Fatal("previous binary could reopen confirmed-term policy")
	}
	if err = Up(ctx, p); err != nil {
		t.Fatal("idempotent migration", err)
	}
	if err = p.QueryRow(ctx, "SELECT grant_state FROM economy_provider_purchases").Scan(&after); err != nil || !bytes.Equal(before, after) {
		t.Fatal("migration rewrote original receipt", err)
	}
	c := ProviderContract{Provider: old.Provider, Environment: old.Environment, App: old.App, Products: map[string]Product{"ad-free": {Kind: "ad_free"}, "daily": {Kind: "daily"}}}
	apply := func(v VerifiedPurchase) error {
		return transact(p, func(tx pgx.Tx) error { _, err := ApplyVerifiedPurchaseTx(ctx, tx, c, old.Account, v); return err })
	}
	if err = apply(old.VerifiedPurchase); err != nil {
		t.Fatal("historical restore", err)
	}
	extended := old.VerifiedPurchase
	extended.Revision = 2
	extended.ExpiresAt = extended.ExpiresAt.Add(time.Hour)
	if err = apply(extended); !errors.Is(err, ErrConflict) {
		t.Fatal("historical grant extended without original confirmation", err)
	}
	fresh := old.VerifiedPurchase
	fresh.Transaction = "new-unanchored"
	if err = apply(fresh); !errors.Is(err, ErrInvalid) {
		t.Fatal("new unanchored grant", err)
	}
	revoked := old.VerifiedPurchase
	revoked.Revision = 2
	revoked.Status = "revoked"
	revoked.ExpiresAt = time.Time{}
	if err = apply(revoked); err != nil {
		t.Fatal("historical revoke without repeated expiry", err)
	}
	if err = apply(old.VerifiedPurchase); err != nil {
		t.Fatal("historical late event", err)
	}
	b, err := ReadBenefits(ctx, p, actors[0], now)
	if err != nil || b.Dirt != 30 || !b.PassUntil.Equal(now.Add(24*time.Hour)) || !b.LegacyAdFreeUntil.Equal(now.Add(time.Hour)) {
		t.Fatal("historical account altered", b, err)
	}
	assertAdFreeUntil(t, b, now.Add(time.Hour))
	daily := purchaseConfirmedAt(t, VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: "daily", Account: actors[0], Transaction: "daily", Revision: 1, Status: "purchased", ExpiresAt: now.Add(24 * time.Hour)}, now)
	if err = apply(daily); err != nil {
		t.Fatal("migrated database cannot accept approved Daily", err)
	}
	if _, err = p.Exec(ctx, "UPDATE economy_migrations SET checksum=$1 WHERE name='fixed-term-products-v1'", []byte("tampered")); err != nil {
		t.Fatal(err)
	}
	if err = Up(ctx, p); err == nil {
		t.Fatal("fixed-term migration checksum mismatch accepted")
	}
}
