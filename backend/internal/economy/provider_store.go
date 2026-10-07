package economy

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyVerifiedPurchaseTx is reachable only from a trusted internal verifier.
// No HTTP handler accepts VerifiedPurchase. Identity locks serialize allowance
// starts with grants/revocations; the provider key prevents cross-account restore.
// A lost outer COMMIT acknowledgement is recovered by repeating the same status.
func ApplyVerifiedPurchaseTx(ctx context.Context, tx pgx.Tx, c ProviderContract, actor string, s VerifiedPurchase) (PurchaseGrant, error) {
	var g PurchaseGrant
	e := savepointOperation(ctx, tx, func(tx pgx.Tx) error {
		if _, e := lockAccount(ctx, tx, actor); e != nil {
			return e
		}
		var body []byte
		e := tx.QueryRow(ctx, `SELECT grant_state FROM economy_provider_purchases WHERE provider=$1 AND environment=$2 AND app=$3 AND transaction_id=$4 FOR UPDATE`, s.Provider, s.Environment, s.App, s.Transaction).Scan(&body)
		if errors.Is(e, pgx.ErrNoRows) {
			g, e = ReconcilePurchase(c, actor, nil, s)
			if e != nil {
				return e
			}
			body, e = json.Marshal(g)
			if e != nil {
				return e
			}
			_, e = tx.Exec(ctx, `INSERT INTO economy_provider_purchases(provider,environment,app,transaction_id,account_id,kind,cosmetic,status,expires_at,grant_state) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`, s.Provider, s.Environment, s.App, s.Transaction, actor, g.Kind, g.Cosmetic, g.Status, g.ExpiresAt, body)
			if e != nil {
				return e
			}
			e = tx.QueryRow(ctx, `SELECT grant_state FROM economy_provider_purchases WHERE provider=$1 AND environment=$2 AND app=$3 AND transaction_id=$4 FOR UPDATE`, s.Provider, s.Environment, s.App, s.Transaction).Scan(&body)
		}
		if e != nil {
			return e
		}
		var old PurchaseGrant
		if e = json.Unmarshal(body, &old); e != nil {
			return e
		}
		g, e = ReconcilePurchase(c, actor, &old, s)
		if e != nil {
			return e
		}
		body, e = json.Marshal(g)
		if e != nil {
			return e
		}
		_, e = tx.Exec(ctx, `UPDATE economy_provider_purchases SET status=$5,expires_at=$6,grant_state=$7 WHERE provider=$1 AND environment=$2 AND app=$3 AND transaction_id=$4`, s.Provider, s.Environment, s.App, s.Transaction, g.Status, g.ExpiresAt, body)
		return e
	})
	if e != nil {
		return PurchaseGrant{}, e
	}
	return g, nil
}

type Benefits struct {
	Dirt              int64     `json:"dirt"`
	PassUntil         time.Time `json:"pass_until"`
	PremiumUntil      time.Time `json:"premium_until"`
	Unlimited         bool      `json:"unlimited"`
	NoAds             bool      `json:"no_ads"`
	Cosmetics         []string  `json:"cosmetics"`
	LegacyAdFreeUntil time.Time `json:"legacy_ad_free_until"`
	AdFreeUntil       time.Time `json:"ad_free_until"`
}

// ReadBenefits projects only the authenticated caller's economic entitlements.
// It excludes receipt identities, provider payloads and promotional inventory.
// The caller supplies its authorized actor; no public account lookup is exposed.
func ReadBenefits(ctx context.Context, p *pgxpool.Pool, actor string, now time.Time) (Benefits, error) {
	return readBenefits(ctx, p, actor, now)
}

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func readBenefits(ctx context.Context, p rowQuerier, actor string, now time.Time) (Benefits, error) {
	var b Benefits
	var standaloneAdFreeUntil time.Time
	if !validID(actor) || now.IsZero() {
		return b, ErrInvalid
	}
	e := p.QueryRow(ctx, `SELECT coalesce(e.dirt,0),coalesce(e.pass_until,'1970-01-01'::timestamptz),greatest(coalesce(e.premium_until,'1970-01-01'::timestamptz),coalesce((SELECT max(expires_at) FROM economy_provider_purchases WHERE account_id=i.account_id AND kind IN ('premium','daily','weekly','monthly','yearly') AND environment='production' AND status='purchased' AND coalesce((grant_state->>'ConfirmedAt')::timestamptz,'1970-01-01'::timestamptz)<=$2),'1970-01-01'::timestamptz)),ARRAY(SELECT DISTINCT cosmetic FROM economy_provider_purchases WHERE account_id=i.account_id AND kind='cosmetic' AND environment='production' AND status='purchased' AND coalesce((grant_state->>'ConfirmedAt')::timestamptz,'1970-01-01'::timestamptz)<=$2 ORDER BY cosmetic),coalesce(e.legacy_ad_free_until,'1970-01-01'::timestamptz),coalesce((SELECT max(expires_at) FROM economy_provider_purchases WHERE account_id=i.account_id AND kind='ad_free' AND environment='production' AND status='purchased' AND coalesce((grant_state->>'ConfirmedAt')::timestamptz,'1970-01-01'::timestamptz)<=$2),'1970-01-01'::timestamptz) FROM identity_accounts i LEFT JOIN economy_accounts e USING(account_id) WHERE i.account_id=$1 AND i.kind IN ('guest','member')`, actor, now).Scan(&b.Dirt, &b.PassUntil, &b.PremiumUntil, &b.Cosmetics, &b.LegacyAdFreeUntil, &standaloneAdFreeUntil)
	if errors.Is(e, pgx.ErrNoRows) {
		e = ErrIneligible
	}
	if e != nil {
		return Benefits{}, e
	}
	b.PassUntil = b.PassUntil.UTC()
	b.PremiumUntil = b.PremiumUntil.UTC()
	b.Unlimited = b.PassUntil.After(now) || b.PremiumUntil.After(now)
	b.LegacyAdFreeUntil = b.LegacyAdFreeUntil.UTC()
	b.AdFreeUntil = b.PremiumUntil
	if b.LegacyAdFreeUntil.After(b.AdFreeUntil) {
		b.AdFreeUntil = b.LegacyAdFreeUntil
	}
	if standaloneAdFreeUntil.After(b.AdFreeUntil) {
		b.AdFreeUntil = standaloneAdFreeUntil.UTC()
	}
	b.NoAds = b.AdFreeUntil.After(now)
	return b, nil
}
