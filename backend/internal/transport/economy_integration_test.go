//go:build integration

package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/rooms"
	"testing"
	"time"
)

func TestHTTPEconomyDisabledAndAuthentication(t *testing.T) {
	s := journeyServer(t)
	clients := journeyGuests(t, s, 1)
	journeyRequest(t, s, 1, "", "GET", "/v1/economy", nil, 401)
	body := journeyRequest(t, s, 0, clients[0].Token, "GET", "/v1/economy", nil, 200)
	var view map[string]any
	if err := json.Unmarshal(body, &view); err != nil || len(view) != 1 || view["enabled"] != false {
		t.Fatal("disabled economy exposes unapproved benefits", string(body), err)
	}
	journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/economy/passes", map[string]any{"operation_id": "disabled", "days": 1}, 409)
}

func economyJourneyServer(t *testing.T) *Server {
	t.Helper()
	base := journeyServer(t)
	if err := economy.Up(context.Background(), base.cfg.Pool); err != nil {
		t.Fatal(err)
	}
	cfg := base.cfg
	policy := economy.AdoptedPolicy()
	cfg.EconomyPolicy = &policy
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.Close(ctx); err != nil {
			t.Error(err)
		}
	})
	return s
}

func TestHTTPEconomyRoomStartsChargeOnceAndRejectExhaustion(t *testing.T) {
	s := economyJourneyServer(t)
	clients := journeyGuests(t, s, 3)
	for i := 0; i < 4; i++ {
		data := journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms", map[string]any{"command_id": fmt.Sprint("room-", i), "capacity": 3, "games_per_match": 3}, 200)
		var room rooms.Room
		if err := json.Unmarshal(data, &room); err != nil {
			t.Fatal(err)
		}
		data = journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms/"+room.ID+"/invitations", struct{}{}, 200)
		var invitation map[string]string
		if err := json.Unmarshal(data, &invitation); err != nil {
			t.Fatal(err)
		}
		for seat := 1; seat < 3; seat++ {
			journeyRequest(t, s, seat, clients[seat].Token, "POST", "/v1/rooms/"+room.ID+"/join", invitation, 200)
		}
		status := 200
		if i == 3 {
			status = 409
		}
		start := map[string]string{"command_id": fmt.Sprint("start-", i)}
		first := journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms/"+room.ID+"/matches", start, status)
		if i == 0 {
			retry := journeyRequest(t, s, 0, clients[0].Token, "POST", "/v1/rooms/"+room.ID+"/matches", start, 200)
			if !bytes.Equal(first, retry) {
				t.Fatal("start retry changed result")
			}
		}
		if i == 3 && !bytes.Contains(first, []byte("ALLOWANCE_EXHAUSTED")) {
			t.Fatal("opaque allowance error", string(first))
		}
	}
	var starts, charges int
	if err := s.cfg.Pool.QueryRow(context.Background(), "SELECT (SELECT count(*) FROM economy_starts),(SELECT count(*) FROM economy_charges)").Scan(&starts, &charges); err != nil || starts != 3 || charges != 9 {
		t.Fatal("partial or double starts", starts, charges, err)
	}
	data := journeyRequest(t, s, 1, clients[1].Token, "GET", "/v1/economy", nil, 200)
	var account economy.AccountView
	if err := json.Unmarshal(data, &account); err != nil || account.FreeStartsRemaining != 0 || account.Benefits.Dirt != 0 {
		t.Fatal("unfinished games granted rewards", string(data), err)
	}
}

func TestHTTPEconomyFinalizedGamesFundPass(t *testing.T) {
	for _, n := range []int{3, 4} {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			s := economyJourneyServer(t)
			clients := journeyGuests(t, s, n)
			actors := make([]string, n)
			for i, c := range clients {
				actors[i] = c.Account.ID
			}
			// Seed a legal settlement boundary, then execute all user finalization
			// confirmations through HTTP and the real automatic authority worker.
			for i := 0; i < 3; i++ {
				id := fmt.Sprint("reward-", i)
				b, err := game.NewState(n, id)
				if err != nil {
					t.Fatal(err)
				}
				m, err := game.NewMatchLifecycle(b, 1)
				if err != nil {
					t.Fatal(err)
				}
				m.Game.Board, err = game.CloseCardBoard(m.Game.Board)
				if err != nil {
					t.Fatal(err)
				}
				m.Game.Ledger, err = m.Game.Ledger.CloseOrdinary(nil)
				if err != nil {
					t.Fatal(err)
				}
				m.Game.Promises, err = m.Game.Promises.CloseBoard()
				if err != nil {
					t.Fatal(err)
				}
				m.Game.Ending = "ordinary"
				env, err := matchstore.NewEnvelope(m, "rules-test")
				if err != nil {
					t.Fatal(err)
				}
				if err = s.matches.Create(context.Background(), id, env, actors); err != nil {
					t.Fatal(err)
				}
				for seat, c := range clients {
					journeyRequest(t, s, seat, c.Token, "POST", "/v1/matches/"+id+"/commands", map[string]any{"command_id": fmt.Sprint("finish-", seat), "game_id": id, "expected_version": seat, "type": "finish-settlement", "payload": map[string]any{}}, 200)
				}
			}
			s.Start()
			deadline := time.Now().Add(5 * time.Second)
			for {
				var rewards int
				if err := s.cfg.Pool.QueryRow(context.Background(), "SELECT count(*) FROM economy_rewards").Scan(&rewards); err != nil {
					t.Fatal(err)
				}
				if rewards == 3*n {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("finalization failed to grant once-only rewards", rewards)
				}
				time.Sleep(20 * time.Millisecond)
			}
			for seat, c := range clients {
				data := journeyRequest(t, s, seat, c.Token, "GET", "/v1/economy", nil, 200)
				var account economy.AccountView
				if err := json.Unmarshal(data, &account); err != nil || account.Benefits.Dirt != 30 || account.FreeStartsRemaining != 0 {
					t.Fatal("completed reward", string(data), err)
				}
				journeyRequest(t, s, seat, c.Token, "POST", "/v1/economy/passes", map[string]any{"operation_id": "earned-pass", "days": 1}, 200)
				data = journeyRequest(t, s, seat, c.Token, "GET", "/v1/economy", nil, 200)
				if err := json.Unmarshal(data, &account); err != nil || account.Benefits.Dirt != 0 || !account.Benefits.Unlimited || account.Benefits.NoAds {
					t.Fatal("earned pass", string(data), err)
				}
			}
		})
	}
}

func TestHTTPEconomyAccountPassAndRecovery(t *testing.T) {
	base := journeyServer(t)
	if err := economy.Up(context.Background(), base.cfg.Pool); err != nil {
		t.Fatal(err)
	}
	cfg := base.cfg
	policy := economy.AdoptedPolicy()
	cfg.EconomyPolicy = &policy
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	clients := journeyGuests(t, s, 2)
	actor, token := clients[0].Account.ID, clients[0].Token
	if _, err = cfg.Pool.Exec(context.Background(), "INSERT INTO economy_accounts(account_id,dirt) VALUES($1,210)", actor); err != nil {
		t.Fatal(err)
	}
	before := journeyRequest(t, s, 0, token, "GET", "/v1/economy", nil, 200)
	var account economy.AccountView
	if err = json.Unmarshal(before, &account); err != nil || !account.Enabled || account.Benefits.Dirt != 210 || account.FreeStartsRemaining != 3 || account.Reward != 10 || account.PassDayPrice != 30 {
		t.Fatal("account", string(before), err)
	}
	for _, days := range []int{2, 3, 4, 5, 6} {
		rejected := journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", map[string]any{"operation_id": fmt.Sprint("retired-", days), "days": days}, 409)
		if !bytes.Contains(rejected, []byte("PRODUCT_UNAVAILABLE")) {
			t.Fatal("retired product recovery signal", string(rejected))
		}
	}
	for _, product := range []string{"ad_free", "monthly", "yearly", "cosmetic"} {
		journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", map[string]any{"operation_id": "invalid-product", "days": 1, "product": product}, 400)
	}
	body := map[string]any{"operation_id": "pass-recovery", "days": 7}
	first := journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", body, 200)
	duplicate := journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", body, 200)
	recovered := journeyRequest(t, s, 0, token, "GET", "/v1/economy/passes/pass-recovery", nil, 200)
	if !bytes.Equal(first, duplicate) || !bytes.Equal(first, recovered) {
		t.Fatal("pass recovery changed original receipt")
	}
	var receipt economy.PassReceipt
	if err = json.Unmarshal(first, &receipt); err != nil || receipt.Days != 7 || receipt.Cost != 210 || !receipt.ExpiresAt.After(time.Now().Add(167*time.Hour)) {
		t.Fatal("receipt", string(first), err)
	}
	journeyRequest(t, s, 1, clients[1].Token, "GET", "/v1/economy/passes/pass-recovery", nil, 404)
	journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", map[string]any{"operation_id": "pass-recovery", "days": 1}, 409)
	journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", map[string]any{"operation_id": "insufficient", "days": 1}, 409)
	journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", map[string]any{"operation_id": "forged", "days": 1, "actor": clients[1].Account.ID}, 400)
	after := journeyRequest(t, s, 0, token, "GET", "/v1/economy", nil, 200)
	if err = json.Unmarshal(after, &account); err != nil || account.Benefits.Dirt != 0 || !account.Benefits.Unlimited || account.Benefits.NoAds {
		t.Fatal("post-redemption", string(after), err)
	}
	for _, forbidden := range []string{"transaction", "code_hash", "grant_state", clients[1].Account.ID} {
		if bytes.Contains(after, []byte(forbidden)) {
			t.Fatal("account data leak", forbidden)
		}
	}
}

func TestHTTPEconomyAdFreeExpiryIsPrivateAndIndependentOfEarnedAccess(t *testing.T) {
	s := economyJourneyServer(t)
	ctx := context.Background()
	clients := journeyGuests(t, s, 2)
	actor, token := clients[0].Account.ID, clients[0].Token
	// A confirmed 30-day purchase with 48 hours remaining is trusted fixture
	// data, not a public purchase-verification route.
	expiry := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Microsecond)
	contract := economy.ProviderContract{
		Provider: "apple", Environment: "production", App: "cgms-contract-fixture",
		Products: map[string]economy.Product{"ad-free-fixture": {Kind: "ad_free"}},
	}
	purchase := economy.VerifiedPurchase{
		Provider: contract.Provider, Environment: contract.Environment, App: contract.App,
		Product: "ad-free-fixture", Account: actor, Transaction: "private-provider-fixture",
		Revision: 1, Status: "purchased", ConfirmedAt: expiry.Add(-30 * 24 * time.Hour), ExpiresAt: expiry,
	}
	apply := func() {
		t.Helper()
		tx, err := s.cfg.Pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if _, err = economy.ApplyVerifiedPurchaseTx(ctx, tx, contract, actor, purchase); err != nil {
			t.Fatal(err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	read := func(seat int) (economy.AccountView, time.Time) {
		t.Helper()
		// A caller-supplied account selector must not change the authenticated view.
		body := journeyRequest(t, s, seat, clients[seat].Token, "GET", "/v1/economy?account_id="+actor, nil, 200)
		var view economy.AccountView
		var raw struct {
			Benefits map[string]json.RawMessage `json:"benefits"`
		}
		if err := json.Unmarshal(body, &view); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			t.Fatal(err)
		}
		var until time.Time
		if err := json.Unmarshal(raw.Benefits["ad_free_until"], &until); err != nil {
			t.Fatal("missing or invalid ad-free expiry", err)
		}
		for _, forbidden := range []string{purchase.Transaction, purchase.Product, actor, "grant_state", "provider"} {
			if bytes.Contains(body, []byte(forbidden)) {
				t.Fatal("private provider identity leaked in account view")
			}
		}
		return view, until
	}
	apply()
	owner, until := read(0)
	if !owner.Benefits.NoAds || owner.Benefits.Unlimited || owner.FreeStartsRemaining != 3 || !until.Equal(expiry) {
		t.Fatal("standalone ad-free changed online allowance or expiry", owner, until)
	}
	other, otherUntil := read(1)
	if other.Benefits.NoAds || other.Benefits.Unlimited || !otherUntil.Equal(time.Unix(0, 0).UTC()) {
		t.Fatal("another caller received owner's entitlement", other, otherUntil)
	}
	if _, err := s.cfg.Pool.Exec(ctx, "UPDATE economy_accounts SET dirt=30 WHERE account_id=$1", actor); err != nil {
		t.Fatal(err)
	}
	journeyRequest(t, s, 0, token, "POST", "/v1/economy/passes", map[string]any{"operation_id": "separate-earned-access", "days": 1}, 200)
	combined, combinedUntil := read(0)
	if !combined.Benefits.NoAds || !combined.Benefits.Unlimited || combined.Benefits.Dirt != 0 || !combinedUntil.Equal(expiry) {
		t.Fatal("earned access extended ad-free benefit or lost independent benefit", combined, combinedUntil)
	}
	purchase.Revision = 2
	purchase.Status = "revoked"
	apply()
	after, afterUntil := read(0)
	if after.Benefits.NoAds || !after.Benefits.Unlimited || !after.Benefits.PassUntil.Equal(combined.Benefits.PassUntil) || !afterUntil.Equal(time.Unix(0, 0).UTC()) {
		t.Fatal("revoking ad-free changed earned access", after, afterUntil)
	}
}

func TestHTTPEconomyEveryPaidAccessTierRemovesAds(t *testing.T) {
	for _, kind := range []string{"daily", "weekly", "monthly", "yearly"} {
		t.Run(kind, func(t *testing.T) {
			s := economyJourneyServer(t)
			ctx := context.Background()
			clients := journeyGuests(t, s, 2)
			actor := clients[0].Account.ID
			confirmed := time.Now().UTC().Truncate(time.Microsecond)
			// Daily uses the adopted exact term. For the other paid kinds this
			// arbitrary remaining window exercises verified provider expiry;
			// it does not define their future store period/calendar mapping.
			expiry := confirmed.Add(24 * time.Hour)
			contract := economy.ProviderContract{
				Provider: "google", Environment: "production", App: "cgms-contract-fixture",
				Products: map[string]economy.Product{"paid-fixture": {Kind: kind}},
			}
			purchase := economy.VerifiedPurchase{
				Provider: contract.Provider, Environment: contract.Environment, App: contract.App,
				Product: "paid-fixture", Account: actor, Transaction: "private-paid-fixture",
				Revision: 1, Status: "purchased", ConfirmedAt: confirmed, ExpiresAt: expiry,
			}
			tx, err := s.cfg.Pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if _, err = economy.ApplyVerifiedPurchaseTx(ctx, tx, contract, actor, purchase); err != nil {
				t.Fatal(err)
			}
			if err = tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			for seat := range clients {
				body := journeyRequest(t, s, seat, clients[seat].Token, "GET", "/v1/economy?account_id="+actor, nil, 200)
				var view economy.AccountView
				if err = json.Unmarshal(body, &view); err != nil {
					t.Fatal(err)
				}
				if !view.Enabled || view.Benefits.NoAds != (seat == 0) || view.Benefits.Unlimited != (seat == 0) || view.FreeStartsRemaining != 3 {
					t.Fatal("paid benefit scope or free allowance changed", kind, seat, view)
				}
				if seat == 0 && (!view.Benefits.PremiumUntil.Equal(expiry) || !view.Benefits.AdFreeUntil.Equal(expiry)) {
					t.Fatal("paid unlimited and ad-free expiry must agree", kind, view)
				}
				for _, forbidden := range []string{purchase.Transaction, purchase.Product, actor, "ConfirmedAt", "confirmed_at", "grant_state", "provider"} {
					if bytes.Contains(body, []byte(forbidden)) {
						t.Fatal("provider metadata leaked through account response", kind)
					}
				}
			}
		})
	}
}
