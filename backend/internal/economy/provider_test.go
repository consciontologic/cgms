package economy

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func purchaseConfirmedAt(t *testing.T, purchase VerifiedPurchase, confirmed time.Time) VerifiedPurchase {
	t.Helper()
	encoded, err := json.Marshal(purchase)
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]json.RawMessage
	if err = json.Unmarshal(encoded, &data); err != nil {
		t.Fatal(err)
	}
	data["ConfirmedAt"], err = json.Marshal(confirmed)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err = json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(encoded, &purchase); err != nil {
		t.Fatal(err)
	}
	return purchase
}

func TestFixedTermProductsValidateConfirmedTimeAndNeverExtend(t *testing.T) {
	confirmed := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		kind     string
		duration time.Duration
	}{{"daily", 24 * time.Hour}, {"ad_free", 30 * 24 * time.Hour}} {
		t.Run(tc.kind, func(t *testing.T) {
			c := ProviderContract{Provider: "apple", Environment: "production", App: "fixture.app", Products: map[string]Product{tc.kind: {Kind: tc.kind}}}
			v := VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: tc.kind, Account: "member", Transaction: tc.kind, Revision: 1, Status: "purchased", ExpiresAt: confirmed.Add(tc.duration)}
			v = purchaseConfirmedAt(t, v, confirmed)
			g, err := ReconcilePurchase(c, v.Account, nil, v)
			if err != nil || g.Active(confirmed.Add(-time.Microsecond)) || !g.Active(confirmed) || g.Active(v.ExpiresAt) {
				t.Fatal("confirmed term boundaries", g, err)
			}
			for _, revision := range []int64{1, 2} {
				restore := v
				restore.Revision = revision
				got, err := ReconcilePurchase(c, v.Account, &g, restore)
				if err != nil || !got.ExpiresAt.Equal(v.ExpiresAt) {
					t.Fatal("restore extended term", got, err)
				}
			}
			changed := purchaseConfirmedAt(t, v, confirmed.Add(time.Hour))
			changed.ExpiresAt, changed.Revision = v.ExpiresAt.Add(time.Hour), 2
			if _, err := ReconcilePurchase(c, v.Account, &g, changed); !errors.Is(err, ErrConflict) {
				t.Fatal("same purchase reset its confirmed term", err)
			}
			revoke := purchaseConfirmedAt(t, v, time.Time{})
			revoke.ExpiresAt, revoke.Status, revoke.Revision = time.Time{}, "revoked", 2
			revoked, err := ReconcilePurchase(c, v.Account, &g, revoke)
			if err != nil || revoked.Active(confirmed) || !revoked.ExpiresAt.Equal(g.ExpiresAt) {
				t.Fatal("revocation must preserve metadata while removing benefit", revoked, err)
			}
			if got, err := ReconcilePurchase(c, v.Account, &revoked, v); err != nil || got != revoked {
				t.Fatal("late purchased event restored revoked term", got, err)
			}
		})
	}
}

func TestFixedTermProductsRejectMissingOrWrongTerms(t *testing.T) {
	confirmed := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	for _, kind := range []string{"daily", "ad_free"} {
		c := ProviderContract{Provider: "google", Environment: "production", App: "fixture.app", Products: map[string]Product{kind: {Kind: kind}}}
		v := VerifiedPurchase{Provider: c.Provider, Environment: c.Environment, App: c.App, Product: kind, Account: "member", Transaction: kind, Revision: 1, Status: "purchased", ExpiresAt: confirmed.Add(2 * time.Hour)}
		for _, anchor := range []time.Time{time.Time{}, confirmed, time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)} {
			bad := purchaseConfirmedAt(t, v, anchor)
			if _, err := ReconcilePurchase(c, v.Account, nil, bad); !errors.Is(err, ErrInvalid) {
				t.Errorf("%s accepted missing or invalid fixed term: %v", kind, err)
			}
		}
	}
}

func TestVerifiedProviderReconciliation(t *testing.T) {
	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	c := ProviderContract{Provider: "apple", Environment: "sandbox", App: "fixture.app", Products: map[string]Product{"premium": {Kind: "weekly"}, "deck": {Kind: "cosmetic", Cosmetic: "fixture-deck"}}}
	s := VerifiedPurchase{Provider: "apple", Environment: "sandbox", App: "fixture.app", Product: "premium", Account: "member", Transaction: "original-1", Revision: 1, Status: "purchased", ExpiresAt: now.Add(24 * time.Hour)}
	g, e := ReconcilePurchase(c, "member", nil, s)
	if e != nil || !g.Active(now) || g.Active(s.ExpiresAt) {
		t.Fatal("expiry", g, e)
	}
	// Restore and duplicate notification preserve an absolute expiry.
	dup, e := ReconcilePurchase(c, "member", &g, s)
	if e != nil || dup != g {
		t.Fatal("duplicate changed grant", dup, e)
	}
	s.Revision = 2
	s.Status = "revoked"
	revoked, e := ReconcilePurchase(c, "member", &g, s)
	if e != nil || revoked.Active(now) {
		t.Fatal("revocation", revoked, e)
	}
	s.Revision = 1
	s.Status = "purchased"
	late, e := ReconcilePurchase(c, "member", &revoked, s)
	if e != nil || late != revoked {
		t.Fatal("late notification resurrected", late, e)
	}
	s.Revision = 2
	if _, e = ReconcilePurchase(c, "member", &revoked, s); !errors.Is(e, ErrConflict) {
		t.Fatal("same revision changed body", e)
	}
	for _, status := range []string{"pending", "unknown"} {
		s.Status = status
		s.Revision = 3
		pending, e := ReconcilePurchase(c, "member", nil, s)
		if e != nil || pending.Active(now) {
			t.Fatal("unconfirmed granted", pending, e)
		}
		if _, e = ReconcilePurchase(c, "member", &g, s); !errors.Is(e, ErrPendingPurchase) {
			t.Fatal("ambiguous result overwrote known purchase", e)
		}
	}
	s.Status = "purchased"
	s.Product = "deck"
	s.ExpiresAt = time.Time{}
	if _, e := ReconcilePurchase(c, "member", nil, s); !errors.Is(e, ErrIneligible) {
		t.Fatal("new cosmetic unlock allowed", e)
	}
	// Historical confirmed ownership can be restored or revoked, never newly sold.
	deck := PurchaseGrant{VerifiedPurchase: s, Kind: "cosmetic", Cosmetic: "fixture-deck"}
	restored, e := ReconcilePurchase(c, "member", &deck, s)
	if e != nil || restored != deck || !restored.Active(now) {
		t.Fatal("historical cosmetic restore", restored, e)
	}
	s.Revision++
	s.Status = "revoked"
	revokedDeck, e := ReconcilePurchase(c, "member", &deck, s)
	if e != nil || revokedDeck.Active(now) {
		t.Fatal("historical cosmetic revocation", revokedDeck, e)
	}
	s.Revision++
	s.Status = "purchased"
	if _, e := ReconcilePurchase(c, "member", &revokedDeck, s); !errors.Is(e, ErrIneligible) {
		t.Fatal("retired cosmetic reactivated", e)
	}
	pendingDeck := deck
	pendingDeck.Status = "pending"
	if _, e := ReconcilePurchase(c, "member", &pendingDeck, s); !errors.Is(e, ErrIneligible) {
		t.Fatal("pending cosmetic unlocked", e)
	}
}

func TestProviderContractRejectsUnboundAndChangedPurchases(t *testing.T) {
	c := ProviderContract{Provider: "google", Environment: "production", App: "fixture.app", Products: map[string]Product{"premium": {Kind: "weekly"}}}
	s := VerifiedPurchase{Provider: "google", Environment: "production", App: "fixture.app", Product: "premium", Account: "member", Transaction: "transaction", Revision: 1, Status: "purchased", ExpiresAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)}
	g, e := ReconcilePurchase(c, "member", nil, s)
	if e != nil {
		t.Fatal(e)
	}
	for _, mutate := range []func(*VerifiedPurchase){
		func(v *VerifiedPurchase) { v.Provider = "apple" }, func(v *VerifiedPurchase) { v.Environment = "sandbox" },
		func(v *VerifiedPurchase) { v.App = "another.app" }, func(v *VerifiedPurchase) { v.Account = "other" },
		func(v *VerifiedPurchase) { v.Product = "unknown" }, func(v *VerifiedPurchase) { v.Revision = 0 },
		func(v *VerifiedPurchase) { v.Status = "client-success" }, func(v *VerifiedPurchase) { v.ExpiresAt = time.Time{} },
	} {
		bad := s
		mutate(&bad)
		if _, e := ReconcilePurchase(c, "member", nil, bad); e == nil {
			t.Fatal("accepted invalid binding", bad)
		}
	}
	if _, e := ReconcilePurchase(c, "other", &g, s); e == nil {
		t.Fatal("cross-account restore")
	}
	s.Transaction = "other-transaction"
	if _, e := ReconcilePurchase(c, "member", &g, s); !errors.Is(e, ErrConflict) {
		t.Fatal("changed transaction", e)
	}
}

func TestRetiredPremiumIsReconciledButNotSold(t *testing.T) {
	c := ProviderContract{Provider: "apple", Environment: "production", App: "fixture.app", Products: map[string]Product{"legacy": {Kind: "premium"}}}
	s := VerifiedPurchase{Provider: "apple", Environment: "production", App: "fixture.app", Product: "legacy", Account: "member", Transaction: "old-premium", Revision: 1, Status: "purchased", ExpiresAt: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)}
	if _, e := ReconcilePurchase(c, "member", nil, s); !errors.Is(e, ErrIneligible) {
		t.Fatal("retired premium sold", e)
	}
	old := PurchaseGrant{VerifiedPurchase: s, Kind: "premium"}
	got, e := ReconcilePurchase(c, "member", &old, s)
	if e != nil || got != old {
		t.Fatal("historic restore", got, e)
	}
	s.Revision++
	s.ExpiresAt = s.ExpiresAt.Add(24 * time.Hour)
	if _, e = ReconcilePurchase(c, "member", &old, s); !errors.Is(e, ErrIneligible) {
		t.Fatal("retired premium extended", e)
	}
	s.Status = "revoked"
	got, e = ReconcilePurchase(c, "member", &old, s)
	if e != nil || got.Status != "revoked" {
		t.Fatal("historic revocation", got, e)
	}
}

func TestStandaloneAdFreeUsesVerifiedExpiryAndReconciliation(t *testing.T) {
	now := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	for _, provider := range []string{"apple", "google"} {
		t.Run(provider, func(t *testing.T) {
			c := ProviderContract{Provider: provider, Environment: "production", App: "fixture.app", Products: map[string]Product{"ad-free": {Kind: "ad_free"}}}
			// A prior confirmed purchase has nine hours left in its approved term.
			s := VerifiedPurchase{Provider: provider, Environment: c.Environment, App: c.App, Product: "ad-free", Account: "member", Transaction: "ad-free-purchase", Revision: 1, Status: "purchased", ExpiresAt: now.Add(9 * time.Hour)}
			s = purchaseConfirmedAt(t, s, s.ExpiresAt.Add(-30*24*time.Hour))
			g, err := ReconcilePurchase(c, s.Account, nil, s)
			if err != nil || !g.Active(now) || g.Active(s.ExpiresAt) || g.Kind != "ad_free" {
				t.Fatal("fixed-term ad-free grant", g, err)
			}
			if restored, err := ReconcilePurchase(c, s.Account, &g, s); err != nil || restored != g {
				t.Fatal("restore changed absolute expiry", restored, err)
			}
			for _, status := range []string{"pending", "unknown"} {
				uncertain := s
				uncertain.Status, uncertain.Revision = status, 2
				if _, err := ReconcilePurchase(c, s.Account, &g, uncertain); !errors.Is(err, ErrPendingPurchase) {
					t.Fatal("uncertain lookup replaced confirmed grant", status, err)
				}
				if unconfirmed, err := ReconcilePurchase(c, s.Account, nil, uncertain); err != nil || unconfirmed.Active(now) {
					t.Fatal("unconfirmed ad-free grant active", unconfirmed, err)
				}
			}
			bad := s
			bad.ExpiresAt = time.Time{}
			if _, err := ReconcilePurchase(c, s.Account, nil, bad); !errors.Is(err, ErrInvalid) {
				t.Fatal("ad-free requires a verified finite expiry", err)
			}
			revoke := s
			revoke.Status, revoke.Revision = "revoked", 2
			revoked, err := ReconcilePurchase(c, s.Account, &g, revoke)
			if err != nil || revoked.Active(now) {
				t.Fatal("revoked ad-free remained active", revoked, err)
			}
			if late, err := ReconcilePurchase(c, s.Account, &revoked, s); err != nil || late != revoked {
				t.Fatal("late duplicate resurrected ad-free", late, err)
			}
		})
	}
}
