package economy

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Promotion is an internal premium grant contract, not a native store payment
// channel. Administrative issuance and public redemption are intentionally not
// exposed until store applicability and channel authorization are settled.
type Promotion struct {
	Days, TotalLimit, PerAccountLimit int
	StartsAt, ExpiresAt               time.Time
	MembersOnly                       bool
}

// Generate once per administrative operation. Retain the same code securely for
// uncertain-outcome recovery; the database stores its hash and cannot recover it.
func NewPromotionCode() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func promotionHash(code string) ([]byte, error) {
	if len(code) != 43 {
		return nil, ErrInvalid
	}
	b, e := base64.RawURLEncoding.DecodeString(code)
	if e != nil || len(b) != 32 || base64.RawURLEncoding.EncodeToString(b) != code {
		return nil, ErrInvalid
	}
	h := sha256.Sum256(b)
	return h[:], nil
}

func CreatePromotionTx(ctx context.Context, tx pgx.Tx, code string, p Promotion) error {
	h, e := promotionHash(code)
	if e != nil {
		return e
	}
	p.StartsAt = p.StartsAt.UTC().Truncate(time.Microsecond)
	p.ExpiresAt = p.ExpiresAt.UTC().Truncate(time.Microsecond)
	if p.Days < 1 || p.Days > 366 || p.TotalLimit < 1 || p.TotalLimit > 1_000_000 || p.PerAccountLimit < 1 || p.PerAccountLimit > p.TotalLimit || p.StartsAt.IsZero() || !p.ExpiresAt.After(p.StartsAt) || p.ExpiresAt.Year() > 9998 {
		return ErrInvalid
	}
	return savepointOperation(ctx, tx, func(tx pgx.Tx) error {
		tag, e := tx.Exec(ctx, "INSERT INTO economy_promotions(code_hash,days,total_limit,per_account_limit,starts_at,expires_at,members_only) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING", h, p.Days, p.TotalLimit, p.PerAccountLimit, p.StartsAt, p.ExpiresAt, p.MembersOnly)
		if e != nil || tag.RowsAffected() != 0 {
			return e
		}
		var old Promotion
		if e = tx.QueryRow(ctx, "SELECT days,total_limit,per_account_limit,starts_at,expires_at,members_only FROM economy_promotions WHERE code_hash=$1", h).Scan(&old.Days, &old.TotalLimit, &old.PerAccountLimit, &old.StartsAt, &old.ExpiresAt, &old.MembersOnly); e != nil {
			return e
		}
		if old.Days != p.Days || old.TotalLimit != p.TotalLimit || old.PerAccountLimit != p.PerAccountLimit || !old.StartsAt.Equal(p.StartsAt) || !old.ExpiresAt.Equal(p.ExpiresAt) || old.MembersOnly != p.MembersOnly {
			return ErrConflict
		}
		return nil
	})
}

// RedeemPromotionTx authenticates through its caller and grants premium once.
// Lock order is one promotion then one account. A recorded operation recovers
// even after the code expires or its limits are exhausted; it never extends twice.
func RedeemPromotionTx(ctx context.Context, tx pgx.Tx, actor, operation, code string, now time.Time) (time.Time, error) {
	h, e := promotionHash(code)
	if e != nil || !validID(actor) || !validID(operation) || now.IsZero() || now.Year() > 9998 {
		return time.Time{}, ErrInvalid
	}
	var expiry time.Time
	e = savepointOperation(ctx, tx, func(tx pgx.Tx) error {
		var p Promotion
		var spent int
		e := tx.QueryRow(ctx, "SELECT days,total_limit,per_account_limit,starts_at,expires_at,members_only,redeemed FROM economy_promotions WHERE code_hash=$1 FOR UPDATE", h).Scan(&p.Days, &p.TotalLimit, &p.PerAccountLimit, &p.StartsAt, &p.ExpiresAt, &p.MembersOnly, &spent)
		if errors.Is(e, pgx.ErrNoRows) {
			return ErrIneligible
		}
		if e != nil {
			return e
		}
		a, e := lockAccount(ctx, tx, actor)
		if e != nil {
			return e
		}
		var oldHash []byte
		e = tx.QueryRow(ctx, "SELECT code_hash,expires_at FROM economy_promotion_operations WHERE account_id=$1 AND operation_id=$2", actor, operation).Scan(&oldHash, &expiry)
		if e == nil {
			if subtle.ConstantTimeCompare(oldHash, h) != 1 {
				return ErrConflict
			}
			expiry = expiry.UTC()
			return nil
		}
		if !errors.Is(e, pgx.ErrNoRows) {
			return e
		}
		if now.Before(p.StartsAt) || !now.Before(p.ExpiresAt) || spent >= p.TotalLimit {
			return ErrIneligible
		}
		if p.MembersOnly {
			var kind string
			if e = tx.QueryRow(ctx, "SELECT kind FROM identity_accounts WHERE account_id=$1", actor).Scan(&kind); e != nil {
				return e
			}
			if kind != "member" {
				return ErrIneligible
			}
		}
		var count int
		if e = tx.QueryRow(ctx, "SELECT count(*) FROM economy_promotion_operations WHERE code_hash=$1 AND account_id=$2", h, actor).Scan(&count); e != nil {
			return e
		}
		if count >= p.PerAccountLimit {
			return ErrIneligible
		}
		expiry = now.UTC()
		if a.premium.After(expiry) {
			expiry = a.premium.UTC()
		}
		if expiry.Year() > 9998 {
			return ErrInvalid
		}
		expiry = expiry.Add(time.Duration(p.Days) * 24 * time.Hour).Truncate(time.Microsecond)
		if expiry.Year() > 9999 {
			return ErrInvalid
		}
		if _, e = tx.Exec(ctx, "UPDATE economy_accounts SET premium_until=$2 WHERE account_id=$1", actor, expiry); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, "UPDATE economy_promotions SET redeemed=redeemed+1 WHERE code_hash=$1", h); e != nil {
			return e
		}
		_, e = tx.Exec(ctx, "INSERT INTO economy_promotion_operations(account_id,operation_id,code_hash,expires_at) VALUES($1,$2,$3,$4)", actor, operation, h, expiry)
		return e
	})
	if e != nil {
		return time.Time{}, e
	}
	return expiry, nil
}
