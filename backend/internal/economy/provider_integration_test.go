//go:build integration

package economy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestProviderDurableRestoreRevocationAndAllowance(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	c := ProviderContract{Provider: "google", Environment: "production", App: "fixture.app", Products: map[string]Product{"premium": {Kind: "weekly"}, "deck": {Kind: "cosmetic", Cosmetic: "fixture-deck"}}}
	s := VerifiedPurchase{Provider: "google", Environment: "production", App: "fixture.app", Product: "premium", Account: actors[0], Transaction: "purchase", Revision: 1, Status: "purchased", ExpiresAt: now.Add(24 * time.Hour)}
	apply := func(v VerifiedPurchase) error {
		return transact(p, func(tx pgx.Tx) error { _, e := ApplyVerifiedPurchaseTx(ctx, tx, c, v.Account, v); return e })
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if e := apply(s); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	check := func(premium bool, cosmetics int) {
		t.Helper()
		b, e := ReadBenefits(ctx, p, actors[0], now)
		if e != nil || b.Unlimited != premium || b.NoAds != premium || len(b.Cosmetics) != cosmetics {
			t.Fatal("benefits", b, e)
		}
	}
	check(true, 0)
	// Restore cannot bind the same provider identity to another account.
	stolen := s
	stolen.Account = actors[1]
	if e := apply(stolen); !errors.Is(e, ErrConflict) {
		t.Fatal("cross-account restore", e)
	}
	pending := s
	pending.Revision = 2
	pending.Status = "unknown"
	if e := apply(pending); !errors.Is(e, ErrPendingPurchase) {
		t.Fatal("ambiguous status", e)
	}
	check(true, 0)
	deck := s
	deck.Transaction = "deck-purchase"
	deck.Product = "deck"
	deck.ExpiresAt = time.Time{}
	if e := apply(deck); !errors.Is(e, ErrIneligible) {
		t.Fatal("new cosmetic purchase", e)
	}
	check(true, 0)
	// A historical record remains restorable, but is not a catalogue offering.
	legacyDeck := PurchaseGrant{VerifiedPurchase: deck, Kind: "cosmetic", Cosmetic: "fixture-deck"}
	body, e := json.Marshal(legacyDeck)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.Exec(ctx, `INSERT INTO economy_provider_purchases(provider,environment,app,transaction_id,account_id,kind,cosmetic,status,expires_at,grant_state) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, deck.Provider, deck.Environment, deck.App, deck.Transaction, deck.Account, legacyDeck.Kind, legacyDeck.Cosmetic, deck.Status, deck.ExpiresAt, body); e != nil {
		t.Fatal(e)
	}
	if e = apply(deck); e != nil {
		t.Fatal("legacy cosmetic restore", e)
	}
	check(true, 1)
	// An uncommitted revocation cannot erase the acknowledged entitlement.
	revoke := s
	revoke.Revision = 2
	revoke.Status = "revoked"
	tx, e := p.Begin(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = ApplyVerifiedPurchaseTx(ctx, tx, c, actors[0], revoke); e != nil {
		t.Fatal(e)
	}
	if e = tx.Rollback(ctx); e != nil {
		t.Fatal(e)
	}
	check(true, 1)
	if e = apply(revoke); e != nil {
		t.Fatal(e)
	}
	check(false, 1)
	if e = apply(s); e != nil {
		t.Fatal(e)
	}
	check(false, 1)
	// A distinct active subscription remains after revoking the first.
	other := s
	other.Transaction = "other-purchase"
	if e = apply(other); e != nil {
		t.Fatal(e)
	}
	check(true, 1)
	for i, id := range []string{"p1", "p2", "p3"} {
		if e = transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, id, actors, now) }); e != nil {
			t.Fatal(i, e)
		}
	}
	var charged int
	if e = p.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE account_id=$1 AND charged_free", actors[0]).Scan(&charged); e != nil || charged != 0 {
		t.Fatal("premium charged", charged, e)
	}
	var receipts int
	if e = p.QueryRow(ctx, "SELECT count(*) FROM economy_provider_purchases").Scan(&receipts); e != nil || receipts != 3 {
		t.Fatal("duplicate receipts", receipts, e)
	}
}

func TestSandboxPurchasesCannotUnlockRealEconomy(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	c := ProviderContract{Provider: "apple", Environment: "sandbox", App: "fixture.app", Products: map[string]Product{"premium": {Kind: "weekly"}}}
	s := VerifiedPurchase{Provider: "apple", Environment: "sandbox", App: "fixture.app", Product: "premium", Account: actors[0], Transaction: "sandbox-transaction", Revision: 1, Status: "purchased", ExpiresAt: now.Add(24 * time.Hour)}
	if e := transact(p, func(tx pgx.Tx) error { _, e := ApplyVerifiedPurchaseTx(ctx, tx, c, actors[0], s); return e }); e != nil {
		t.Fatal(e)
	}
	b, e := ReadBenefits(ctx, p, actors[0], now)
	if e != nil || b.Unlimited || b.NoAds {
		t.Fatal("sandbox unlocked real benefits", b, e)
	}
	if e = transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, "sandbox-cap", actors, now) }); e != nil {
		t.Fatal(e)
	}
	var charged bool
	if e = p.QueryRow(ctx, "SELECT charged_free FROM economy_charges WHERE game_id='sandbox-cap' AND account_id=$1", actors[0]).Scan(&charged); e != nil || !charged {
		t.Fatal("sandbox bypassed allowance", charged, e)
	}
}

func TestApprovedPaidAccessRetainsUnlimitedAndAdRemoval(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	for i, kind := range []string{"weekly", "monthly", "yearly"} {
		c := ProviderContract{Provider: "apple", Environment: "production", App: "fixture.app", Products: map[string]Product{kind: {Kind: kind}}}
		v := VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: kind, Account: actors[i], Transaction: kind, Revision: 1, Status: "purchased", ExpiresAt: now.Add(24 * time.Hour)}
		if err := transact(p, func(tx pgx.Tx) error { _, e := ApplyVerifiedPurchaseTx(ctx, tx, c, actors[i], v); return e }); err != nil {
			t.Fatal(kind, err)
		}
		b, err := ReadBenefits(ctx, p, actors[i], now)
		if err != nil || !b.Unlimited || !b.NoAds {
			t.Fatal(kind, b, err)
		}
	}
	if err := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, "paid-plans", actors, now) }); err != nil {
		t.Fatal(err)
	}
	var charged int
	if err := p.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE game_id='paid-plans' AND charged_free").Scan(&charged); err != nil || charged != 0 {
		t.Fatal("paid access allowance", charged, err)
	}
}

func TestStandaloneAdFreeKeepsAllowanceAndEarnedAccessSeparate(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	c := ProviderContract{Provider: "apple", Environment: "production", App: "fixture.app", Products: map[string]Product{"ad-free": {Kind: "ad_free"}}}
	v := VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: "ad-free", Account: actors[0], Transaction: "ad-free", Revision: 1, Status: "purchased", ExpiresAt: now.Add(9 * time.Hour)}
	v = purchaseConfirmedAt(t, v, v.ExpiresAt.Add(-30*24*time.Hour))
	apply := func(purchase VerifiedPurchase) error {
		return transact(p, func(tx pgx.Tx) error {
			_, err := ApplyVerifiedPurchaseTx(ctx, tx, c, purchase.Account, purchase)
			return err
		})
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := apply(v); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if t.Failed() {
		t.FailNow()
	}
	b, err := ReadBenefits(ctx, p, actors[0], now)
	if err != nil || b.Unlimited || !b.NoAds || b.Dirt != 0 {
		t.Fatal("standalone ad-free benefits", b, err)
	}
	assertAdFreeUntil(t, b, v.ExpiresAt)
	for _, id := range []string{"ad-free-1", "ad-free-2", "ad-free-3"} {
		if err = transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, id, actors, now) }); err != nil {
			t.Fatal(err)
		}
	}
	var charged int
	if err = p.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE account_id=$1 AND charged_free", actors[0]).Scan(&charged); err != nil || charged != 3 {
		t.Fatal("standalone ad-free owner must consume free starts", charged, err)
	}
	if err = transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, "ad-free-fourth", actors, now) }); !errors.Is(err, ErrAllowance) {
		t.Fatal("ad-free bypassed free allowance", err)
	}
	if err = transact(p, func(tx pgx.Tx) error {
		canStart, err := CanStartTx(ctx, tx, actors[0], now)
		if err == nil && canStart {
			t.Error("ad-free eligibility hint bypassed allowance")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	var receipts int
	if err = p.QueryRow(ctx, "SELECT count(*) FROM economy_provider_purchases").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatal("duplicate ad-free receipts", receipts, err)
	}
	stolen := v
	stolen.Account = actors[1]
	if err = apply(stolen); !errors.Is(err, ErrConflict) {
		t.Fatal("cross-account ad-free restore", err)
	}
	unknown := v
	unknown.Status, unknown.Revision = "unknown", 2
	if err = apply(unknown); !errors.Is(err, ErrPendingPurchase) {
		t.Fatal("uncertain ad-free reconciliation", err)
	}
	if _, err = p.Exec(ctx, "UPDATE economy_accounts SET dirt=30 WHERE account_id=$1", actors[0]); err != nil {
		t.Fatal(err)
	}
	service, err := NewService(p, AdoptedPolicy())
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time { return now }
	if _, err = service.BuyPass(ctx, actors[0], "earned-daily", 1); err != nil {
		t.Fatal(err)
	}
	b, err = ReadBenefits(ctx, p, actors[0], now)
	if err != nil || !b.Unlimited || !b.NoAds || b.Dirt != 0 {
		t.Fatal("combined earned and ad-free", b, err)
	}
	assertAdFreeUntil(t, b, v.ExpiresAt)
	b, err = ReadBenefits(ctx, p, actors[0], v.ExpiresAt)
	if err != nil || !b.Unlimited || b.NoAds {
		t.Fatal("earned time extended ad-free expiry", b, err)
	}
	revoke := v
	revoke.Status, revoke.Revision = "revoked", 2
	if err = apply(revoke); err != nil {
		t.Fatal(err)
	}
	if err = apply(v); err != nil {
		t.Fatal("late duplicate", err)
	}
	b, err = ReadBenefits(ctx, p, actors[0], now)
	if err != nil || !b.Unlimited || b.NoAds {
		t.Fatal("revocation changed earned access", b, err)
	}
	assertAdFreeUntil(t, b, time.Unix(0, 0).UTC())
	if err = transact(p, func(tx pgx.Tx) error {
		canStart, err := CanStartTx(ctx, tx, actors[0], now)
		if err == nil && !canStart {
			t.Error("earned access was not eligible after ad-free revocation")
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAdFreeUntilCombinesAllCurrentAdRemovalSources(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	c := ProviderContract{Provider: "google", Environment: "production", App: "fixture.app", Products: map[string]Product{"ad-free": {Kind: "ad_free"}, "weekly": {Kind: "weekly"}}}
	for _, product := range []string{"ad-free", "weekly"} {
		v := VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: product, Account: actors[0], Transaction: product, Revision: 1, Status: "purchased", ExpiresAt: now.Add(9 * time.Hour)}
		if product == "weekly" {
			v.ExpiresAt = now.Add(4 * time.Hour)
		} else {
			v = purchaseConfirmedAt(t, v, v.ExpiresAt.Add(-30*24*time.Hour))
		}
		if err := transact(p, func(tx pgx.Tx) error { _, err := ApplyVerifiedPurchaseTx(ctx, tx, c, actors[0], v); return err }); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct{ premium, legacy, want int }{{6, 3, 9}, {6, 12, 12}, {15, 3, 15}} {
		if _, err := p.Exec(ctx, "UPDATE economy_accounts SET premium_until=$2,legacy_ad_free_until=$3 WHERE account_id=$1", actors[0], now.Add(time.Duration(tc.premium)*time.Hour), now.Add(time.Duration(tc.legacy)*time.Hour)); err != nil {
			t.Fatal(err)
		}
		b, err := ReadBenefits(ctx, p, actors[0], now)
		if err != nil || !b.NoAds {
			t.Fatal(b, err)
		}
		until := now.Add(time.Duration(tc.want) * time.Hour)
		assertAdFreeUntil(t, b, until)
		b, err = ReadBenefits(ctx, p, actors[0], until)
		if err != nil || b.NoAds || b.Unlimited {
			t.Fatal("exact expiry boundary", b, err)
		}
	}
	if _, err := p.Exec(ctx, "UPDATE economy_accounts SET premium_until=$2,legacy_ad_free_until=$2 WHERE account_id=$1", actors[0], time.Unix(0, 0)); err != nil {
		t.Fatal(err)
	}
	revoked := VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: "ad-free", Account: actors[0], Transaction: "ad-free", Revision: 2, Status: "revoked", ExpiresAt: now.Add(9 * time.Hour)}
	if err := transact(p, func(tx pgx.Tx) error { _, err := ApplyVerifiedPurchaseTx(ctx, tx, c, actors[0], revoked); return err }); err != nil {
		t.Fatal(err)
	}
	sandboxContract := c
	sandboxContract.Environment = "sandbox"
	sandbox := revoked
	sandbox.Environment, sandbox.Transaction, sandbox.Status, sandbox.Revision = "sandbox", "sandbox-ad-free", "purchased", 1
	sandbox = purchaseConfirmedAt(t, sandbox, sandbox.ExpiresAt.Add(-30*24*time.Hour))
	if err := transact(p, func(tx pgx.Tx) error {
		_, err := ApplyVerifiedPurchaseTx(ctx, tx, sandboxContract, actors[0], sandbox)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	b, err := ReadBenefits(ctx, p, actors[0], now)
	if err != nil || !b.Unlimited || !b.NoAds {
		t.Fatal("paid weekly fallback", b, err)
	}
	assertAdFreeUntil(t, b, now.Add(4*time.Hour))
	b, err = ReadBenefits(ctx, p, actors[1], now)
	if err != nil || b.Unlimited || b.NoAds {
		t.Fatal("other account benefit privacy", b, err)
	}
	assertAdFreeUntil(t, b, time.Unix(0, 0).UTC())
}

func assertAdFreeUntil(t *testing.T, benefits Benefits, want time.Time) {
	t.Helper()
	encoded, err := json.Marshal(benefits)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	var got time.Time
	if err = json.Unmarshal(fields["ad_free_until"], &got); err != nil || !got.Equal(want) {
		t.Fatal("private ad_free_until must expose effective absolute expiry", got, want, err)
	}
}

func TestPaidAccessProductContractMatrix(t *testing.T) {
	for _, kind := range []string{"daily", "weekly", "monthly", "yearly"} {
		t.Run(kind, func(t *testing.T) {
			p, actors := economyDB(t)
			ctx := context.Background()
			now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
			c := ProviderContract{Provider: "apple", Environment: "production", App: "fixture.app", Products: map[string]Product{kind: {Kind: kind}}}
			// Only Daily defines a local duration. Other rows use an arbitrary verified provider expiry.
			v := VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: kind, Revision: 1, Status: "purchased", ExpiresAt: now.Add(24 * time.Hour)}
			if kind == "daily" {
				v = purchaseConfirmedAt(t, v, now)
			}
			apply := func(contract ProviderContract, purchase VerifiedPurchase) error {
				return transact(p, func(tx pgx.Tx) error {
					_, err := ApplyVerifiedPurchaseTx(ctx, tx, contract, purchase.Account, purchase)
					return err
				})
			}
			for _, actor := range actors {
				v.Account, v.Transaction = actor, kind+actor
				if err := apply(c, v); err != nil {
					t.Fatal("paid product grant", err)
				}
			}
			v.Account, v.Transaction = actors[0], kind+actors[0]
			for _, revision := range []int64{1, 2} {
				v.Revision = revision
				if err := apply(c, v); err != nil {
					t.Fatal("duplicate or restore", err)
				}
				b, err := ReadBenefits(ctx, p, actors[0], now)
				if err != nil || !b.Unlimited || !b.NoAds || !b.PremiumUntil.Equal(v.ExpiresAt) {
					t.Fatal("paid benefits or absolute restore expiry", b, err)
				}
				assertAdFreeUntil(t, b, v.ExpiresAt)
			}
			unknown := v
			unknown.Status, unknown.Revision = "unknown", 3
			if err := apply(c, unknown); !errors.Is(err, ErrPendingPurchase) {
				t.Fatal("uncertain provider result", err)
			}
			for i := 0; i < 5; i++ {
				if err := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, fmt.Sprintf("%s-%d", kind, i), actors, now) }); err != nil {
					t.Fatal("paid start beyond free cap", i, err)
				}
			}
			var receipts, charged int
			if err := p.QueryRow(ctx, "SELECT count(*),count(*) FILTER(WHERE charged_free) FROM economy_charges WHERE account_id=$1", actors[0]).Scan(&receipts, &charged); err != nil || receipts != 5 || charged != 0 {
				t.Fatal("paid start charged allowance", receipts, charged, err)
			}
			b, err := ReadBenefits(ctx, p, actors[0], v.ExpiresAt)
			if err != nil || b.Unlimited || b.NoAds {
				t.Fatal("expiry must remove both paid benefits", b, err)
			}
			revoke := purchaseConfirmedAt(t, v, time.Time{})
			revoke.Status, revoke.Revision, revoke.ExpiresAt = "revoked", 3, time.Time{}
			if err = apply(c, revoke); err != nil {
				t.Fatal("revoke without repeated term metadata", err)
			}
			if err = apply(c, v); err != nil {
				t.Fatal("late duplicate", err)
			}
			sandboxContract, sandbox := c, v
			sandboxContract.Environment, sandbox.Environment, sandbox.Transaction = "sandbox", "sandbox", "sandbox-"+kind
			if err = apply(sandboxContract, sandbox); err != nil {
				t.Fatal(err)
			}
			b, err = ReadBenefits(ctx, p, actors[0], now)
			if err != nil || b.Unlimited || b.NoAds {
				t.Fatal("sandbox or revoked product remained active", b, err)
			}
			if _, err = p.Exec(ctx, "UPDATE economy_accounts SET dirt=30 WHERE account_id=$1", actors[0]); err != nil {
				t.Fatal(err)
			}
			s, err := NewService(p, AdoptedPolicy())
			if err != nil {
				t.Fatal(err)
			}
			s.now = func() time.Time { return now }
			if _, err = s.BuyPass(ctx, actors[0], "earned-after-paid", 1); err != nil {
				t.Fatal(err)
			}
			b, err = ReadBenefits(ctx, p, actors[0], now)
			if err != nil || !b.Unlimited || b.NoAds {
				t.Fatal("earned access after paid revocation must keep ads", b, err)
			}
		})
	}
}

func TestFixedTermFutureConfirmationDoesNotGrantEarly(t *testing.T) {
	for _, kind := range []string{"daily", "ad_free"} {
		t.Run(kind, func(t *testing.T) {
			p, actors := economyDB(t)
			ctx := context.Background()
			now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
			confirmed := now.Add(time.Hour)
			term := 24 * time.Hour
			if kind == "ad_free" {
				term = 30 * 24 * time.Hour
			}
			c := ProviderContract{Provider: "google", Environment: "production", App: "fixture.app", Products: map[string]Product{kind: {Kind: kind}}}
			v := purchaseConfirmedAt(t, VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: kind, Account: actors[0], Transaction: kind, Revision: 1, Status: "purchased", ExpiresAt: confirmed.Add(term)}, confirmed)
			if err := transact(p, func(tx pgx.Tx) error { _, err := ApplyVerifiedPurchaseTx(ctx, tx, c, actors[0], v); return err }); err != nil {
				t.Fatal(err)
			}
			b, err := ReadBenefits(ctx, p, actors[0], now)
			if err != nil || b.NoAds || b.Unlimited {
				t.Fatal("future purchase granted before confirmation", b, err)
			}
			assertAdFreeUntil(t, b, time.Unix(0, 0).UTC())
			for i, at := range []time.Time{now, confirmed} {
				if err := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, fmt.Sprintf("future-%d", i), actors, at) }); err != nil {
					t.Fatal(err)
				}
			}
			var charged int
			want := 2
			if kind == "daily" {
				want = 1
			}
			if err = p.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE account_id=$1 AND charged_free", actors[0]).Scan(&charged); err != nil || charged != want {
				t.Fatal("admission disagrees with confirmed time", charged, want, err)
			}
			b, err = ReadBenefits(ctx, p, actors[0], confirmed)
			if err != nil || !b.NoAds || b.Unlimited != (kind == "daily") {
				t.Fatal("exact confirmation boundary", b, err)
			}
			b, err = ReadBenefits(ctx, p, actors[0], v.ExpiresAt)
			if err != nil || b.NoAds || b.Unlimited {
				t.Fatal("exact expiry boundary", b, err)
			}
		})
	}
}
