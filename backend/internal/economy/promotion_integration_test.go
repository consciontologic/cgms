//go:build integration

package economy

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestPromotionConcurrentLimitAndRecovery(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 30, 12, 0, 0, 123456789, time.UTC)
	code, err := NewPromotionCode()
	if err != nil {
		t.Fatal(err)
	}
	rules := Promotion{Days: 7, TotalLimit: 2, PerAccountLimit: 1, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)}
	issue := func() error {
		return transact(p, func(tx pgx.Tx) error { return CreatePromotionTx(ctx, tx, code, rules) })
	}
	if err = issue(); err != nil {
		t.Fatal(err)
	}
	if err = issue(); err != nil {
		t.Fatal("issuance retry", err)
	}
	var admitted atomic.Int32
	var wg sync.WaitGroup
	for _, actor := range actors {
		wg.Add(1)
		go func(actor string) {
			defer wg.Done()
			err := transact(p, func(tx pgx.Tx) error { _, e := RedeemPromotionTx(ctx, tx, actor, "same", code, now); return e })
			if err == nil {
				admitted.Add(1)
			} else if !errors.Is(err, ErrIneligible) {
				t.Error(err)
			}
		}(actor)
	}
	wg.Wait()
	if admitted.Load() != 2 {
		t.Fatal("global code limit", admitted.Load())
	}
	var actor string
	var original time.Time
	if err = p.QueryRow(ctx, "SELECT account_id,expires_at FROM economy_promotion_operations LIMIT 1").Scan(&actor, &original); err != nil {
		t.Fatal(err)
	}
	var recovered time.Time
	if err = transact(p, func(tx pgx.Tx) error {
		var e error
		recovered, e = RedeemPromotionTx(ctx, tx, actor, "same", code, now.Add(2*time.Hour))
		return e
	}); err != nil || !recovered.Equal(original) {
		t.Fatal("expired depleted code retry", recovered, original, err)
	}
	if err = transact(p, func(tx pgx.Tx) error { _, e := RedeemPromotionTx(ctx, tx, actor, "new", code, now); return e }); !errors.Is(err, ErrIneligible) {
		t.Fatal("per-account limit", err)
	}
	var total int
	if err = p.QueryRow(ctx, "SELECT count(*) FROM economy_promotion_operations").Scan(&total); err != nil || total != 2 {
		t.Fatal("duplicate grant", total, err)
	}
}

func TestPromotionEligibilityExpiryConflictAndRollback(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	code, err := NewPromotionCode()
	if err != nil {
		t.Fatal(err)
	}
	rules := Promotion{Days: 3, TotalLimit: 10, PerAccountLimit: 1, StartsAt: now, ExpiresAt: now.Add(time.Hour), MembersOnly: true}
	if err = transact(p, func(tx pgx.Tx) error { return CreatePromotionTx(ctx, tx, code, rules) }); err != nil {
		t.Fatal(err)
	}
	redeem := func(at time.Time) error {
		return transact(p, func(tx pgx.Tx) error { _, e := RedeemPromotionTx(ctx, tx, actors[0], "op", code, at); return e })
	}
	if err = redeem(now); !errors.Is(err, ErrIneligible) {
		t.Fatal("guest eligibility", err)
	}
	rules.MembersOnly = false
	if err = transact(p, func(tx pgx.Tx) error { return CreatePromotionTx(ctx, tx, code, rules) }); !errors.Is(err, ErrConflict) {
		t.Fatal("changed issuance", err)
	}
	// A different code is a distinct immutable promotion.
	second, err := NewPromotionCode()
	if err != nil {
		t.Fatal(err)
	}
	if err = transact(p, func(tx pgx.Tx) error { return CreatePromotionTx(ctx, tx, second, rules) }); err != nil {
		t.Fatal(err)
	}
	for _, at := range []time.Time{now.Add(-time.Nanosecond), rules.ExpiresAt} {
		if err = transact(p, func(tx pgx.Tx) error { _, e := RedeemPromotionTx(ctx, tx, actors[0], "op", second, at); return e }); !errors.Is(err, ErrIneligible) {
			t.Fatal("time bounds", err)
		}
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = RedeemPromotionTx(ctx, tx, actors[0], "op", second, now); err != nil {
		t.Fatal(err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = p.QueryRow(ctx, "SELECT count(*) FROM economy_promotion_operations").Scan(&count); err != nil || count != 0 {
		t.Fatal("grant escaped outer rollback", count, err)
	}
	if err = transact(p, func(tx pgx.Tx) error { _, e := RedeemPromotionTx(ctx, tx, actors[0], "op", second, now); return e }); err != nil {
		t.Fatal(err)
	}
	if err = redeem(now); !errors.Is(err, ErrConflict) {
		t.Fatal("operation changed code", err)
	}
	if err = transact(p, func(tx pgx.Tx) error {
		_, e := RedeemPromotionTx(ctx, tx, actors[0], "second-op", second, now)
		return e
	}); !errors.Is(err, ErrIneligible) {
		t.Fatal("per-account cap with unused global capacity", err)
	}
}
