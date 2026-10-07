package economy

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// CanStartTx is a read-only eligibility hint within a caller-owned snapshot.
// ReserveStartTx still revalidates and charges atomically when the game starts.
// Callers must authorize their roster and must not publish another account's
// entitlement or remaining free starts.
func CanStartTx(ctx context.Context, tx pgx.Tx, actor string, now time.Time) (bool, error) {
	if !validID(actor) || now.IsZero() {
		return false, ErrInvalid
	}
	var bot bool
	if err := tx.QueryRow(ctx, "SELECT kind='bot' FROM identity_accounts WHERE account_id=$1", actor).Scan(&bot); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return false, err
	}
	if bot {
		return true, nil
	}
	benefits, err := readBenefits(ctx, tx, actor, now)
	if errors.Is(err, ErrIneligible) {
		return false, nil
	}
	if err != nil || benefits.Unlimited {
		return benefits.Unlimited, err
	}
	var count int
	err = tx.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE account_id=$1 AND utc_day=$2::date AND charged_free", actor, Day(now)).Scan(&count)
	return count < FreeStarts, err
}
