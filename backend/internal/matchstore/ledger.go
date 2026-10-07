package matchstore

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/metaphy6/cgms/backend/internal/game"
)

// syncLedger rebuilds current materialized financial projections under the
// caller's match-row transaction. It never modifies the immutable command/event
// journal, and preserves old game instance records, including void instances.
func syncLedger(ctx context.Context, tx pgx.Tx, matchID string, e Envelope, version int64) error {
	l := e.Match.Game.Ledger
	completed := map[string]game.FinancialResult{}
	for _, r := range l.Completed {
		completed[r.GameID] = r
	}
	slot := 0
	for _, id := range e.Match.Instances {
		phase := "void"
		var payload any = struct {
			GameID string `json:"game_id"`
		}{id}
		if r, ok := completed[id]; ok {
			phase = "finalized"
			payload = r
		}
		if id == l.GameID {
			phase = l.Phase
			payload = e.Match.Game
		}
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO games(match_id,game_id,slot,phase,payload) VALUES($1,$2,$3,$4,$5) ON CONFLICT(match_id,game_id) DO NOTHING`, matchID, id, slot, phase, raw); err != nil {
			return err
		}
		if id == l.GameID {
			if _, err = tx.Exec(ctx, `UPDATE games SET phase=$3,payload=$4 WHERE match_id=$1 AND game_id=$2`, matchID, id, phase, raw); err != nil {
				return err
			}
		}
		if _, ok := completed[id]; ok {
			slot++
		}
	}
	// Cascades are explicit here so restoring a start ledger cannot preserve a
	// charge/correction whose creation was nullified. Raw history stays in journal.
	for _, table := range []string{"financial_corrections", "financial_debts", "financial_charges", "game_balances", "settlement_workspaces"} {
		if _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE match_id=$1", matchID); err != nil {
			return err
		}
	}
	balance := func(id string, seat int, score, cash game.Amount) error {
		sn, sd := amountParts(score)
		cn, cd := amountParts(cash)
		_, err := tx.Exec(ctx, `INSERT INTO game_balances VALUES($1,$2,$3,$4,$5,$6,$7)`, matchID, id, seat, sn, sd, cn, cd)
		return err
	}
	for _, r := range l.Completed {
		for seat, score := range r.Scores {
			if err := balance(r.GameID, seat, score, game.Amount{}); err != nil {
				return err
			}
		}
	}
	// A finalized current game also appears in Completed; update its row once.
	for seat, score := range l.Scores {
		sn, sd := amountParts(score)
		cn, cd := amountParts(l.Cash[seat])
		if _, err := tx.Exec(ctx, `INSERT INTO game_balances VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(match_id,game_id,seat) DO UPDATE SET score_num=EXCLUDED.score_num,score_den=EXCLUDED.score_den,cash_num=EXCLUDED.cash_num,cash_den=EXCLUDED.cash_den`, matchID, l.GameID, seat, sn, sd, cn, cd); err != nil {
			return err
		}
	}
	for _, c := range l.Charges {
		n, d := amountParts(c.Amount)
		if _, err := tx.Exec(ctx, `INSERT INTO financial_charges VALUES($1,$2,$3,$4,$5,$6,$7)`, matchID, c.ID, c.GameID, c.Debtor, c.Creditor, n, d); err != nil {
			return err
		}
	}
	for _, d := range l.Debts {
		if d.Sequence > math.MaxInt64 {
			return fmt.Errorf("debt sequence exceeds durable range")
		}
		n, den := amountParts(d.Remaining)
		if _, err := tx.Exec(ctx, `INSERT INTO financial_debts VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, matchID, d.ID, d.ChargeID, d.GameID, d.Debtor, d.Creditor, n, den, int64(d.Sequence)); err != nil {
			return err
		}
	}
	for _, c := range l.Corrections {
		n, d := amountParts(c.Amount)
		if _, err := tx.Exec(ctx, `INSERT INTO financial_corrections VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, matchID, c.OperationID, c.DebtID, c.ChargeID, c.GameID, c.Debtor, n, d, c.MatchLevel); err != nil {
			return err
		}
	}
	// Retain previous game terms for private audit, replace this game's current
	// materialization atomically with its snapshot (never publish these raw rows).
	for _, table := range []string{"promises", "proposals"} {
		if _, err := tx.Exec(ctx, "DELETE FROM "+table+" WHERE match_id=$1 AND game_id=$2", matchID, l.GameID); err != nil {
			return err
		}
	}
	for _, p := range e.Match.Game.Promises.Promises {
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO promises VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, matchID, l.GameID, p.ID, p.Payer, p.Recipient, p.AwardID, p.Status, b); err != nil {
			return err
		}
	}
	for _, p := range e.Match.Game.Board.Proposals {
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO proposals VALUES($1,$2,$3,$4,$5,$6,$7)`, matchID, p.GameID, p.ID, p.Revision, p.Turn, p.Status, b); err != nil {
			return err
		}
	}
	if e.Match.Game.Automatic != nil || e.Match.Game.PromisePayment != nil {
		// Store the whole detached continuation, including committed ledger, private
		// workspace, payment prior, FIFO cursor and reproducible compressed audit.
		b, err := json.Marshal(e.Match.Game)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO settlement_workspaces VALUES($1,$2,$3,$4)`, matchID, l.GameID+"/automatic", version, b); err != nil {
			return err
		}
	}
	return nil
}

func amountParts(a game.Amount) (string, string) {
	n, d, ok := strings.Cut(a.String(), "/")
	if !ok {
		d = "1"
	}
	return n, d
}
