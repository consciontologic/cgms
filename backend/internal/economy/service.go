package economy

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("economy operation not found")
	ErrOutcomeUnknown = errors.New("economy operation outcome unknown")
)

// Service owns account-operation transactions, separately from game transactions.
// Only authenticated transport supplies actor IDs. It installs no store verifier.
type Service struct {
	pool   *pgxpool.Pool
	policy Policy
	now    func() time.Time
}

func NewService(pool *pgxpool.Pool, policy Policy) (*Service, error) {
	if pool == nil || policy.Validate() != nil {
		return nil, ErrInvalid
	}
	return &Service{pool: pool, policy: policy, now: time.Now}, nil
}

type AccountView struct {
	Enabled             bool      `json:"enabled"`
	Benefits            Benefits  `json:"benefits"`
	FreeStartsRemaining int       `json:"free_starts_remaining"`
	ResetsAt            time.Time `json:"resets_at"`
	Reward              int64     `json:"reward"`
	PassDayPrice        int64     `json:"pass_day_price"`
}

type PassReceipt struct {
	OperationID string    `json:"operation_id"`
	Days        int       `json:"days"`
	Cost        int64     `json:"cost"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (s *Service) Account(ctx context.Context, actor string) (AccountView, error) {
	now := s.now().UTC()
	// One repeatable-read snapshot prevents mixing pre/post-redemption balances
	// and entitlement/allowance states in the displayed account view.
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return AccountView{}, err
	}
	defer rollback(tx)
	b, err := readBenefits(ctx, tx, actor, now)
	if err != nil {
		return AccountView{}, err
	}
	var count int
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM economy_charges WHERE account_id=$1 AND utc_day=$2::date AND charged_free", actor, Day(now)).Scan(&count); err != nil {
		return AccountView{}, err
	}
	v := AccountView{Enabled: true, Benefits: b, FreeStartsRemaining: max(0, FreeStarts-count), ResetsAt: time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC), Reward: s.policy.Reward, PassDayPrice: s.policy.PassDayPrice}
	if err = tx.Commit(ctx); err != nil {
		return AccountView{}, err
	}
	return v, nil
}

func readPass(ctx context.Context, q rowQuerier, actor, operation string) (PassReceipt, error) {
	if !validID(actor) || !validID(operation) {
		return PassReceipt{}, ErrInvalid
	}
	v := PassReceipt{OperationID: operation}
	err := q.QueryRow(ctx, "SELECT p.days,p.cost,p.expires_at FROM economy_pass_operations p JOIN identity_accounts i USING(account_id) WHERE p.account_id=$1 AND p.operation_id=$2 AND i.kind<>'deleted'", actor, operation).Scan(&v.Days, &v.Cost, &v.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return PassReceipt{}, ErrNotFound
	}
	v.ExpiresAt = v.ExpiresAt.UTC()
	return v, err
}

func (s *Service) LookupPass(ctx context.Context, actor, operation string) (PassReceipt, error) {
	return readPass(ctx, s.pool, actor, operation)
}

func (s *Service) BuyPass(ctx context.Context, actor, operation string, days int) (PassReceipt, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return PassReceipt{}, err
	}
	defer rollback(tx)
	if _, err = BuyPassTx(ctx, tx, actor, operation, days, s.now(), s.policy); err != nil {
		return PassReceipt{}, err
	}
	v, err := readPass(ctx, tx, actor, operation)
	if err != nil {
		return PassReceipt{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return PassReceipt{}, ErrOutcomeUnknown
	}
	return v, nil
}
