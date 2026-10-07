package economy

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"errors"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/game"
)

//go:embed schema.sql
var schema string

// The original schema is immutable: deployed databases pin its checksum.
//
//go:embed access_products.sql
var accessProductsMigration string

//go:embed ad_free_products.sql
var adFreeProductsMigration string

//go:embed fixed_term_products.sql
var fixedTermProductsMigration string

func Up(ctx context.Context, p *pgxpool.Pool) error {
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer rollback(tx)
	if _, e = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(734982622); CREATE TABLE IF NOT EXISTS economy_schema(checksum bytea PRIMARY KEY)"); e != nil {
		return e
	}
	sum := sha256.Sum256([]byte(schema))
	accessProductsSum := sha256.Sum256([]byte(schema + "\n" + accessProductsMigration))
	adFreeProductsSum := sha256.Sum256([]byte(schema + "\n" + accessProductsMigration + "\n" + adFreeProductsMigration))
	currentSum := sha256.Sum256([]byte(schema + "\n" + accessProductsMigration + "\n" + adFreeProductsMigration + "\n" + fixedTermProductsMigration))
	var old []byte
	e = tx.QueryRow(ctx, "SELECT checksum FROM economy_schema").Scan(&old)
	if errors.Is(e, pgx.ErrNoRows) {
		if _, e = tx.Exec(ctx, schema); e != nil {
			return e
		}
		_, e = tx.Exec(ctx, "INSERT INTO economy_schema VALUES($1)", sum[:])
	} else if e == nil && !bytes.Equal(old, sum[:]) && !bytes.Equal(old, accessProductsSum[:]) && !bytes.Equal(old, adFreeProductsSum[:]) && !bytes.Equal(old, currentSum[:]) {
		return errors.New("unsupported economy schema")
	}
	if e != nil {
		return e
	}
	if _, e = tx.Exec(ctx, "CREATE TABLE IF NOT EXISTS economy_migrations(name text PRIMARY KEY, checksum bytea NOT NULL)"); e != nil {
		return e
	}
	migrationSum := sha256.Sum256([]byte(accessProductsMigration))
	e = tx.QueryRow(ctx, "SELECT checksum FROM economy_migrations WHERE name='access-products-v1'").Scan(&old)
	if errors.Is(e, pgx.ErrNoRows) {
		if _, e = tx.Exec(ctx, accessProductsMigration); e != nil {
			return e
		}
		_, e = tx.Exec(ctx, "INSERT INTO economy_migrations VALUES('access-products-v1',$1)", migrationSum[:])
	} else if e == nil && !bytes.Equal(old, migrationSum[:]) {
		return errors.New("unsupported access products migration")
	}
	if e != nil {
		return e
	}
	migrationSum = sha256.Sum256([]byte(adFreeProductsMigration))
	e = tx.QueryRow(ctx, "SELECT checksum FROM economy_migrations WHERE name='ad-free-products-v1'").Scan(&old)
	if errors.Is(e, pgx.ErrNoRows) {
		if _, e = tx.Exec(ctx, adFreeProductsMigration); e != nil {
			return e
		}
		_, e = tx.Exec(ctx, "INSERT INTO economy_migrations VALUES('ad-free-products-v1',$1)", migrationSum[:])
	} else if e == nil && !bytes.Equal(old, migrationSum[:]) {
		return errors.New("unsupported ad-free products migration")
	}
	if e != nil {
		return e
	}
	migrationSum = sha256.Sum256([]byte(fixedTermProductsMigration))
	e = tx.QueryRow(ctx, "SELECT checksum FROM economy_migrations WHERE name='fixed-term-products-v1'").Scan(&old)
	if errors.Is(e, pgx.ErrNoRows) {
		if _, e = tx.Exec(ctx, fixedTermProductsMigration); e != nil {
			return e
		}
		_, e = tx.Exec(ctx, "INSERT INTO economy_migrations VALUES('fixed-term-products-v1',$1)", migrationSum[:])
	} else if e == nil && !bytes.Equal(old, migrationSum[:]) {
		return errors.New("unsupported fixed-term products migration")
	}
	if e != nil {
		return e
	}
	// Older binaries must not reopen this database with the retired product policy.
	if _, e = tx.Exec(ctx, "UPDATE economy_schema SET checksum=$1", currentSum[:]); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

func rollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}

// Each operation uses a savepoint: a rejected call leaves no partial economy
// mutation even if its caller commits unrelated work. Outer commit ownership
// stays with the game transaction; an unknown COMMIT outcome must be reconciled
// using the same game/operation identity, never a fresh identity.
func savepointOperation(ctx context.Context, tx pgx.Tx, fn func(pgx.Tx) error) error {
	sp, e := tx.Begin(ctx)
	if e != nil {
		return e
	}
	defer rollback(sp)
	if e = fn(sp); e != nil {
		return e
	}
	return sp.Commit(ctx)
}
func validID(id string) bool { return len(id) > 0 && len(id) <= 256 }

type account struct {
	dirt          int64
	pass, premium time.Time
}

func lockAccount(ctx context.Context, tx pgx.Tx, id string) (account, error) {
	var a account
	// Match admission also locks identities first. Lock the two tables explicitly:
	// a joint FOR UPDATE leaves acquisition order to the database query plan.
	var actor string
	e := tx.QueryRow(ctx, "SELECT account_id FROM identity_accounts WHERE account_id=$1 AND kind IN ('guest','member') FOR UPDATE", id).Scan(&actor)
	if errors.Is(e, pgx.ErrNoRows) {
		return a, ErrIneligible
	}
	if e != nil {
		return a, e
	}
	if _, e := tx.Exec(ctx, "INSERT INTO economy_accounts(account_id) VALUES($1) ON CONFLICT DO NOTHING", actor); e != nil {
		return a, e
	}
	e = tx.QueryRow(ctx, "SELECT dirt,pass_until,premium_until FROM economy_accounts WHERE account_id=$1 FOR UPDATE", actor).Scan(&a.dirt, &a.pass, &a.premium)
	if errors.Is(e, pgx.ErrNoRows) {
		e = ErrIneligible
	}
	return a, e
}

// ReserveStartTx is an internal hook for the very same transaction that creates
// an online game. Its caller authenticates the roster; it is not a public API.
func ReserveStartTx(ctx context.Context, tx pgx.Tx, gameID string, actors []string, now time.Time) error {
	if !validID(gameID) || (len(actors) != 3 && len(actors) != 4) || now.IsZero() {
		return ErrInvalid
	}
	roster := slices.Clone(actors)
	slices.Sort(roster)
	for i, a := range roster {
		if !validID(a) || (i > 0 && a == roster[i-1]) {
			return ErrInvalid
		}
	}
	return savepointOperation(ctx, tx, func(tx pgx.Tx) error {
		tag, e := tx.Exec(ctx, "INSERT INTO economy_starts(game_id,actors,started_at) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", gameID, roster, now.UTC())
		if e != nil {
			return e
		}
		if tag.RowsAffected() == 0 {
			var prior []string
			if e = tx.QueryRow(ctx, "SELECT actors FROM economy_starts WHERE game_id=$1", gameID).Scan(&prior); e != nil {
				return e
			}
			if !slices.Equal(prior, roster) {
				return ErrConflict
			}
			return nil
		}
		for _, actor := range roster {
			// Bot identities have no account benefits, allowance or spendable
			// rewards. Preserve the complete roster while charging humans normally.
			var kind string
			if e := tx.QueryRow(ctx, "SELECT kind FROM identity_accounts WHERE account_id=$1 FOR UPDATE", actor).Scan(&kind); e != nil {
				return e
			}
			if kind == "bot" {
				continue
			}
			a, e := lockAccount(ctx, tx, actor)
			if e != nil {
				return e
			}
			free := !a.pass.After(now) && !a.premium.After(now)
			if free {
				var premium bool
				if e = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM economy_provider_purchases WHERE account_id=$1 AND kind IN ('premium','daily','weekly','monthly','yearly') AND environment='production' AND status='purchased' AND expires_at>$2 AND coalesce((grant_state->>'ConfirmedAt')::timestamptz,'1970-01-01'::timestamptz)<=$2)", actor, now).Scan(&premium); e != nil {
					return e
				}
				free = !premium
			}
			if free {
				var count int
				if e = tx.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE account_id=$1 AND utc_day=$2::date AND charged_free", actor, Day(now)).Scan(&count); e != nil {
					return e
				}
				if count >= FreeStarts {
					return ErrAllowance
				}
			}
			if _, e = tx.Exec(ctx, "INSERT INTO economy_charges(game_id,account_id,utc_day,charged_free) VALUES($1,$2,$3::date,$4)", gameID, actor, Day(now), free); e != nil {
				return e
			}
		}
		return nil
	})
}

// RewardFinalizedTx accepts only a valid finalized ledger for an online-start
// identity. Call inside the authority's finalization transaction. It never
// accepts a client-reported result or changes the game ledger.
func RewardFinalizedTx(ctx context.Context, tx pgx.Tx, l game.FinancialLedger, p Policy) error {
	if p.Validate() != nil {
		return ErrInvalid
	}
	if l.Phase != "finalized" || l.Validate() != nil {
		return ErrIneligible
	}
	return savepointOperation(ctx, tx, func(tx pgx.Tx) error {
		var roster []string
		e := tx.QueryRow(ctx, "SELECT actors FROM economy_starts WHERE game_id=$1 FOR UPDATE", l.GameID).Scan(&roster)
		if errors.Is(e, pgx.ErrNoRows) {
			return ErrIneligible
		}
		if e != nil {
			return e
		}
		if len(roster) != len(l.Scores) {
			return ErrConflict
		}
		for _, actor := range roster {
			// Deleted accounts do not receive new spendable rewards; shared outcomes remain.
			a, e := lockAccount(ctx, tx, actor)
			if errors.Is(e, ErrIneligible) {
				continue
			}
			if e != nil {
				return e
			}
			tag, e := tx.Exec(ctx, "INSERT INTO economy_rewards(game_id,account_id,amount) VALUES($1,$2,$3) ON CONFLICT DO NOTHING", l.GameID, actor, p.Reward)
			if e != nil {
				return e
			}
			if tag.RowsAffected() == 0 {
				continue
			}
			if a.dirt > MaxBalance-p.Reward {
				return ErrInvalid
			}
			if _, e = tx.Exec(ctx, "UPDATE economy_accounts SET dirt=dirt+$2 WHERE account_id=$1", actor, p.Reward); e != nil {
				return e
			}
		}
		return nil
	})
}

// BuyPassTx consumes dirt once and extends the expiry. The resulting pass is
// unlimited; advertisements require a separate entitlement. Retried IDs retain expiry even
// after clock/pricing changes; changing the chosen duration is a conflict.
func BuyPassTx(ctx context.Context, tx pgx.Tx, actor, operation string, days int, now time.Time, p Policy) (time.Time, error) {
	if !validID(actor) || !validID(operation) || days < 1 || days > 7 || now.IsZero() || p.Validate() != nil {
		return time.Time{}, ErrInvalid
	}
	var expiry time.Time
	e := savepointOperation(ctx, tx, func(tx pgx.Tx) error {
		a, e := lockAccount(ctx, tx, actor)
		if e != nil {
			return e
		}
		var oldDays int
		e = tx.QueryRow(ctx, "SELECT days,expires_at FROM economy_pass_operations WHERE account_id=$1 AND operation_id=$2", actor, operation).Scan(&oldDays, &expiry)
		if e == nil {
			expiry = expiry.UTC()
			if oldDays != days {
				return ErrConflict
			}
			return nil
		}
		if !errors.Is(e, pgx.ErrNoRows) {
			return e
		}
		// Retired durations remain replayable above, but cannot be bought again.
		if days != 1 && days != 7 {
			return ErrProductUnavailable
		}
		cost := int64(days) * p.PassDayPrice
		if a.dirt < cost {
			return ErrFunds
		}
		expiry, e = ExtendPass(now, a.pass, days)
		if e != nil {
			return e
		}
		// PostgreSQL timestamptz stores microseconds. Return the same exact value
		// on the initial request and every later recovery of this operation.
		expiry = expiry.UTC().Truncate(time.Microsecond)
		if _, e = tx.Exec(ctx, "UPDATE economy_accounts SET dirt=dirt-$2,pass_until=$3 WHERE account_id=$1", actor, cost, expiry); e != nil {
			return e
		}
		_, e = tx.Exec(ctx, "INSERT INTO economy_pass_operations(account_id,operation_id,days,cost,expires_at) VALUES($1,$2,$3,$4,$5)", actor, operation, days, cost, expiry)
		return e
	})
	if e != nil {
		return time.Time{}, e
	}
	return expiry, nil
}
