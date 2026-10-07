package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
)

// seedEconomy exists only in the loopback synthetic QA command. Three prior-day
// valid finalized ledger fixtures earn the adopted rewards through the production
// transaction hooks; there is no direct balance UPDATE or client outcome API.
func seedEconomy(ctx context.Context, pool *pgxpool.Pool, id string, actors []string, policy economy.Policy, scenario string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		end, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(end)
	}()
	now := time.Now().UTC().Truncate(time.Microsecond)
	for i := 0; i < 3; i++ {
		gameID := fmt.Sprintf("%s-prior-day-%d", id, i)
		if err = economy.ReserveStartTx(ctx, tx, gameID, actors, now.AddDate(0, 0, -1)); err != nil {
			return err
		}
		ledger := game.NewFinancialLedger(gameID, len(actors))
		ledger, err = ledger.CloseOrdinary(nil)
		if err != nil {
			return err
		}
		for seat := range actors {
			ledger, err = ledger.FinishSettlement(seat)
			if err != nil {
				return err
			}
		}
		ledger, err = ledger.Finalize()
		if err != nil {
			return err
		}
		if err = economy.RewardFinalizedTx(ctx, tx, ledger, policy); err != nil {
			return err
		}
	}
	if contract, purchase, ok := fixturePurchase(id, actors[0], scenario, now); ok {
		if _, err = economy.ApplyVerifiedPurchaseTx(ctx, tx, contract, actors[0], purchase); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// fixturePurchase is trusted test data confined to this disposable loopback
// authority. The production label exercises benefits without contacting a store;
// no client supplies purchase identity, confirmation time, or entitlement fields.
func fixturePurchase(id, actor, scenario string, confirmed time.Time) (economy.ProviderContract, economy.VerifiedPurchase, bool) {
	var kind string
	var duration time.Duration
	switch scenario {
	case "ad-free":
		kind, duration = "ad_free", 30*24*time.Hour
	case "paid-daily":
		kind, duration = "daily", 24*time.Hour
	case "paid-weekly", "paid-monthly", "paid-yearly":
		kind = scenario[len("paid-"):]
		// Only a synthetic provider QA expiry; these offer durations are not
		// established by this fixture or derived from the product label.
		duration = 48 * time.Hour
	default:
		return economy.ProviderContract{}, economy.VerifiedPurchase{}, false
	}
	confirmed = confirmed.UTC().Truncate(time.Microsecond)
	product := "qa-" + scenario
	contract := economy.ProviderContract{
		Provider: "apple", Environment: "production", App: "cgms.webfixture",
		Products: map[string]economy.Product{product: {Kind: kind}},
	}
	purchase := economy.VerifiedPurchase{
		Provider: contract.Provider, Environment: contract.Environment, App: contract.App,
		Product: product, Account: actor, Transaction: id + "-" + product,
		Revision: 1, Status: "purchased", ConfirmedAt: confirmed, ExpiresAt: confirmed.Add(duration),
	}
	return contract, purchase, true
}
