//go:build integration

package matchstore

import (
	"context"
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestLedgerProjectionExactLineageAndRollback(t *testing.T) {
	ctx := context.Background()
	p := migrationPool(t)
	if err := Up(ctx, p); err != nil {
		t.Fatal(err)
	}
	e := envelopeFixture(t)
	amount, err := game.NewAmount("1", "99999999999999999999999999999999999999999999999999999999999999999999991")
	if err != nil {
		t.Fatal(err)
	}
	e.Match.Game.Ledger, err = e.Match.Game.Ledger.Charge("charge", 0, 1, amount)
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.Exec(ctx, `INSERT INTO matches VALUES('m','',0,$1,'','','')`, e.Match.Game.Board.GameID)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err = syncLedger(ctx, tx, "m", e, 0); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var num, den, charge, gameID string
	if err = p.QueryRow(ctx, `SELECT remaining_num,remaining_den,charge_id,game_id FROM financial_debts WHERE match_id='m'`).Scan(&num, &den, &charge, &gameID); err != nil {
		t.Fatal(err)
	}
	if num != "1" || den != "99999999999999999999999999999999999999999999999999999999999999999999991" || charge != "charge" || gameID != e.Match.Game.Board.GameID {
		t.Fatal("exact provenance lost")
	}
	for _, invalid := range []string{"0", "-1", "01"} {
		if _, err = p.Exec(ctx, `UPDATE financial_debts SET remaining_den=$1 WHERE match_id='m'`, invalid); err == nil {
			t.Fatal("invalid denominator admitted")
		}
	}
	// Fully repaid debts remain as lineage with an exact zero remainder.
	paid := e.Clone()
	paid.Match.Game.Ledger, err = game.SettleReceipts(paid.Match.Game.Ledger, []game.FinancialReceipt{{Seat: 0, Amount: amount}})
	if err != nil {
		t.Fatal(err)
	}
	tx, err = p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err = syncLedger(ctx, tx, "m", paid, 1); err != nil {
		t.Fatal("zero remaining debt is valid lineage", err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	// A nullification restores the start ledger; current projections must remove
	// rolled-back charge/debt rows, while a rolled-back SQL transaction retains them.
	restored := e.Clone()
	restored.Match.Game.Ledger = restored.Match.StartLedger.Clone()
	tx, err = p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err = syncLedger(ctx, tx, "m", restored, 1); err != nil {
		t.Fatal(err)
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	var count int
	if err = p.QueryRow(ctx, `SELECT count(*) FROM financial_debts`).Scan(&count); err != nil || count != 1 {
		t.Fatal("rollback lost ledger", count, err)
	}
	tx, err = p.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err = syncLedger(ctx, tx, "m", restored, 1); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err = p.QueryRow(ctx, `SELECT count(*) FROM financial_charges`).Scan(&count); err != nil || count != 0 {
		t.Fatal("restored ledger retained void charge", count, err)
	}
}
