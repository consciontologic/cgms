package matchstore

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/economy"
)

// NewWithEconomy explicitly enables economy for newly created matches. The
// caller must migrate identity/economy first. There are no default prices.
// Each match pins its policy; reopening it with New cannot disable its charges
// or change its rewards. Server activation remains a separate product gate.
func NewWithEconomy(pool *pgxpool.Pool, rulesHash string, policy economy.Policy) (*Store, error) {
	if policy.Validate() != nil {
		return nil, economy.ErrInvalid
	}
	s := New(pool, rulesHash)
	s.economyPolicy = &policy
	return s, nil
}

func (s *Store) syncEconomy(ctx context.Context, tx pgx.Tx, id, previousGame string, next Envelope) error {
	if previousGame == next.Match.Game.Board.GameID && next.Match.Game.Ledger.Phase != "finalized" {
		return nil
	}
	var raw []byte
	if err := tx.QueryRow(ctx, "SELECT config FROM matches WHERE match_id=$1", id).Scan(&raw); err != nil {
		return err
	}
	var cfg struct {
		Economy *economy.Policy `json:"economy"`
	}
	if json.Unmarshal(raw, &cfg) != nil {
		return ErrCompatibility
	}
	if cfg.Economy == nil {
		return nil
	}
	if cfg.Economy.Validate() != nil {
		return ErrCompatibility
	}
	if previousGame != next.Match.Game.Board.GameID {
		rows, err := tx.Query(ctx, "SELECT actor_id FROM members WHERE match_id=$1 ORDER BY actor_id", id)
		if err != nil {
			return err
		}
		actors, err := pgx.CollectRows(rows, pgx.RowTo[string])
		if err != nil {
			return err
		}
		if err = economy.ReserveStartTx(ctx, tx, next.Match.Game.Board.GameID, actors, s.now()); err != nil {
			return err
		}
	}
	if next.Match.Game.Ledger.Phase == "finalized" {
		return economy.RewardFinalizedTx(ctx, tx, next.Match.Game.Ledger, *cfg.Economy)
	}
	return nil
}

// admissionWaiting reveals only that this roster cannot yet start another game.
// The status is read-only and separate from the immutable projection/event cursor.
func (s *Store) admissionWaiting(ctx context.Context, id, actor, gameID string) (bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return false, err
	}
	defer rollback(tx)
	var raw []byte
	var currentGame string
	if err = tx.QueryRow(ctx, "SELECT config,current_game_id FROM matches JOIN members USING(match_id) WHERE match_id=$1 AND actor_id=$2", id, actor).Scan(&raw, &currentGame); err != nil {
		return false, err
	}
	if currentGame != gameID {
		return false, nil // The observed predecessor has already advanced.
	}
	var cfg struct {
		Economy *economy.Policy `json:"economy"`
	}
	if json.Unmarshal(raw, &cfg) != nil || cfg.Economy != nil && cfg.Economy.Validate() != nil {
		return false, ErrCompatibility
	}
	if cfg.Economy == nil {
		return false, nil
	}
	rows, err := tx.Query(ctx, "SELECT actor_id FROM members WHERE match_id=$1 ORDER BY actor_id", id)
	if err != nil {
		return false, err
	}
	actors, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return false, err
	}
	now := s.now()
	for _, member := range actors {
		eligible, err := economy.CanStartTx(ctx, tx, member, now)
		if err != nil || !eligible {
			return !eligible, err
		}
	}
	return false, nil
}
