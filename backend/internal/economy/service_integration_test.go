//go:build integration

package economy

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestAccountProjectionUTCAndConcurrentPasses(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	s, err := NewService(p, AdoptedPolicy())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 1, 2, 59, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	s.now = func() time.Time { return now }
	for i := 0; i < 3; i++ {
		if err := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, fmt.Sprint("utc-", i), actors, now) }); err != nil {
			t.Fatal(err)
		}
	}
	a, err := s.Account(ctx, actors[0])
	if err != nil || a.FreeStartsRemaining != 0 || a.ResetsAt != time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatal("UTC quota", a, err)
	}
	now = now.Add(time.Minute)
	a, err = s.Account(ctx, actors[0])
	if err != nil || a.FreeStartsRemaining != 3 {
		t.Fatal("UTC rollover", a, err)
	}
	if _, err := p.Exec(ctx, "UPDATE economy_accounts SET dirt=180 WHERE account_id=$1", actors[0]); err != nil {
		t.Fatal(err)
	}
	var count atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.BuyPass(ctx, actors[0], fmt.Sprint("pass-", i), 1)
			if err == nil {
				count.Add(1)
			} else if !errors.Is(err, ErrFunds) {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	a, err = s.Account(ctx, actors[0])
	if err != nil || count.Load() != 6 || a.Benefits.Dirt != 0 || a.Benefits.NoAds || !a.Benefits.PassUntil.Equal(now.Add(6*24*time.Hour)) {
		t.Fatal("concurrent spend", count.Load(), a, err)
	}
	if _, err := s.LookupPass(ctx, actors[1], "pass-1"); !errors.Is(err, ErrNotFound) {
		t.Fatal("cross-account receipt", err)
	}
	if _, err := p.Exec(ctx, "UPDATE identity_accounts SET kind='deleted' WHERE account_id=$1", actors[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Account(ctx, actors[0]); !errors.Is(err, ErrIneligible) {
		t.Fatal("deleted account", err)
	}
}

func TestEarnedAccessKeepsAdsAndRejectsRetiredDurations(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	s, err := NewService(p, AdoptedPolicy())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	if err = transact(p, func(tx pgx.Tx) error { _, e := lockAccount(ctx, tx, actors[0]); return e }); err != nil {
		t.Fatal(err)
	}
	if _, err = p.Exec(ctx, "UPDATE economy_accounts SET dirt=300 WHERE account_id=$1", actors[0]); err != nil {
		t.Fatal(err)
	}
	for _, days := range []int{2, 3, 4, 5, 6, 30, 365} {
		if _, err = s.BuyPass(ctx, actors[0], fmt.Sprint("retired-", days), days); !errors.Is(err, ErrInvalid) {
			t.Fatalf("new %d-day purchase: %v", days, err)
		}
	}
	first, err := s.BuyPass(ctx, actors[0], "weekly", 7)
	if err != nil || first.Cost != 210 || !first.ExpiresAt.Equal(now.Add(7*24*time.Hour)) {
		t.Fatal(first, err)
	}
	duplicate, err := s.BuyPass(ctx, actors[0], "weekly", 7)
	if err != nil || duplicate != first {
		t.Fatal("duplicate", duplicate, err)
	}
	a, err := s.Account(ctx, actors[0])
	if err != nil || a.Benefits.Dirt != 90 || !a.Benefits.Unlimited || a.Benefits.NoAds {
		t.Fatal("earned access and ads must be separate", a, err)
	}
	if _, err = p.Exec(ctx, "UPDATE economy_accounts SET premium_until=$2 WHERE account_id=$1", actors[0], now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	a, err = s.Account(ctx, actors[0])
	if err != nil || !a.Benefits.Unlimited || !a.Benefits.NoAds {
		t.Fatal("independent existing premium", a, err)
	}
	now = now.Add(2 * time.Hour)
	a, err = s.Account(ctx, actors[0])
	if err != nil || !a.Benefits.Unlimited || a.Benefits.NoAds {
		t.Fatal("earned access must not extend ad removal", a, err)
	}
}
