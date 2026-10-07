package matchstore

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

func syncReputation(ctx context.Context, tx pgx.Tx, id string, e Envelope) error {
	if e.Match.Game.Promises.Phase != "finalized" {
		return nil
	}
	outcomes, err := e.Match.Game.Promises.Outcomes()
	if err != nil {
		return err
	}
	for _, o := range outcomes {
		if o.Outcome == "none" {
			continue
		}
		if o.Outcome != "trust" && o.Outcome != "anyhoo" && o.Outcome != "scam" {
			return errors.New("unsettled finalized reputation")
		}
		var existing string
		err = tx.QueryRow(ctx, `INSERT INTO reputation_outcomes(match_id,game_id,payer,recipient,outcome) VALUES($1,$2,$3,$4,$5) ON CONFLICT(match_id,game_id,payer,recipient) DO UPDATE SET outcome=reputation_outcomes.outcome RETURNING outcome`, id, e.Match.Game.Board.GameID, o.Payer+1, o.Recipient+1, o.Outcome).Scan(&existing)
		if err != nil {
			return err
		}
		if existing != o.Outcome {
			return errors.New("reputation outcome conflict")
		}
	}
	return nil
}
