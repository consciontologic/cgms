//go:build integration

package economy

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/metaphy6/cgms/backend/internal/game"
)

func TestBotRosterNeverOwnsAllowanceRewardsOrBenefits(t *testing.T) {
	p, actors := economyDB(t)
	ctx := context.Background()
	now := time.Now().UTC()
	// This fixture retains a normal human; the two server actors have no sessions.
	for _, actor := range actors[1:] {
		if _, err := p.Exec(ctx, "DELETE FROM identity_sessions WHERE account_id=$1", actor); err != nil {
			t.Fatal(err)
		}
		if _, err := p.Exec(ctx, "UPDATE identity_accounts SET kind='bot' WHERE account_id=$1", actor); err != nil {
			t.Fatal(err)
		}
	}
	for start := 0; start < 4; start++ {
		err := transact(p, func(tx pgx.Tx) error { return ReserveStartTx(ctx, tx, fmt.Sprintf("bot-game-%d", start), actors, now) })
		if start < 3 && err != nil {
			t.Fatal(err)
		}
		if start == 3 && !errors.Is(err, ErrAllowance) {
			t.Fatal("bot mode bypassed human quota", err)
		}
	}
	if err := transact(p, func(tx pgx.Tx) error {
		for i, actor := range actors {
			eligible, err := CanStartTx(ctx, tx, actor, now)
			if err != nil || eligible != (i > 0) {
				t.Fatalf("admission human/bot distinction %d %v %v", i, eligible, err)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	ledger := game.NewFinancialLedger("bot-game-0", 3)
	ledger, err := ledger.CloseOrdinary(nil)
	if err != nil {
		t.Fatal(err)
	}
	for seat := 0; seat < 3; seat++ {
		ledger, err = ledger.FinishSettlement(seat)
		if err != nil {
			t.Fatal(err)
		}
	}
	ledger, err = ledger.Finalize()
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err = transact(p, func(tx pgx.Tx) error { return RewardFinalizedTx(ctx, tx, ledger, AdoptedPolicy()) }); err != nil {
			t.Fatal(err)
		}
	}
	var charges, rewards, accounts int
	if err = p.QueryRow(ctx, "SELECT (SELECT count(*) FROM economy_charges),(SELECT count(*) FROM economy_rewards),(SELECT count(*) FROM economy_accounts)").Scan(&charges, &rewards, &accounts); err != nil || charges != 3 || rewards != 1 || accounts != 1 {
		t.Fatal("bots acquired economy ownership", charges, rewards, accounts, err)
	}
	for _, actor := range actors[1:] {
		if _, err = readBenefits(ctx, p, actor, now); !errors.Is(err, ErrIneligible) {
			t.Fatal("bot benefits readable", err)
		}
		if err = transact(p, func(tx pgx.Tx) error { _, e := BuyPassTx(ctx, tx, actor, "no-pass", 1, now, AdoptedPolicy()); return e }); !errors.Is(err, ErrIneligible) {
			t.Fatal("bot bought a pass", err)
		}
	}
}
