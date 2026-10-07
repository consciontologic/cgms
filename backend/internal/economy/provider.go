package economy

import (
	"errors"
	"time"
)

var ErrPendingPurchase = errors.New("provider result requires reconciliation")

// These are internal adapter contracts, not provider API payloads. No production
// verifier or public purchase endpoint is installed. An adapter must verify the
// signature/API response and authoritative account binding before constructing
// VerifiedPurchase; accepting client JSON as this type is forbidden.
type ProviderContract struct {
	Provider, Environment, App string
	Products                   map[string]Product
}
type Product struct{ Kind, Cosmetic string }
type VerifiedPurchase struct {
	Provider, Environment, App, Product, Account, Transaction string
	// Revision is adapter-normalized authoritative ordering, not delivery order.
	// Each real adapter must prove this mapping; notifications trigger a fresh
	// authoritative lookup when they cannot supply such an ordering themselves.
	Revision  int64
	Status    string
	ExpiresAt time.Time
	// ConfirmedAt is the provider-verified purchase instant, never receipt or
	// reconciliation time. Fixed-term purchases persist it for exact restoration.
	ConfirmedAt time.Time
}
type PurchaseGrant struct {
	VerifiedPurchase
	Kind, Cosmetic string
}

func (g PurchaseGrant) Active(now time.Time) bool {
	return g.Status == "purchased" && !g.ConfirmedAt.After(now) && (g.Kind == "cosmetic" || g.ExpiresAt.After(now))
}

// ReconcilePurchase is pure so the persistence adapter can run it under the
// transaction-identity lock. Premium uses provider absolute expiry, never now+
// duration; restore and repeated notification therefore cannot mint extra time.
// Unknown/pending lookup results never erase an existing confirmed entitlement.
func ReconcilePurchase(c ProviderContract, actor string, old *PurchaseGrant, s VerifiedPurchase) (PurchaseGrant, error) {
	p, ok := c.Products[s.Product]
	if (c.Provider != "apple" && c.Provider != "google") || (c.Environment != "sandbox" && c.Environment != "production") || !validID(c.App) || !validID(actor) || !validID(s.Transaction) || s.Provider != c.Provider || s.Environment != c.Environment || s.App != c.App || s.Account != actor || s.Revision <= 0 || !ok || (!paidAccessKind(p.Kind) && p.Kind != "ad_free" && p.Kind != "premium" && p.Kind != "cosmetic") || (p.Kind == "cosmetic" && !validID(p.Cosmetic)) || (p.Kind != "cosmetic" && p.Cosmetic != "") {
		return PurchaseGrant{}, ErrInvalid
	}
	if s.Status != "purchased" && s.Status != "revoked" && s.Status != "pending" && s.Status != "unknown" {
		return PurchaseGrant{}, ErrInvalid
	}
	s.ExpiresAt = s.ExpiresAt.UTC().Truncate(time.Microsecond)
	s.ConfirmedAt = s.ConfirmedAt.UTC().Truncate(time.Microsecond)
	if old != nil {
		if old.Provider != s.Provider || old.Environment != s.Environment || old.App != s.App || old.Account != actor || old.Transaction != s.Transaction || old.Product != s.Product || old.Kind != p.Kind || old.Cosmetic != p.Cosmetic {
			return PurchaseGrant{}, ErrConflict
		}
		if s.Revision < old.Revision {
			return *old, nil
		}
		// Revocation removes a benefit even when the verifier does not repeat
		// purchase-time metadata. Preserve the confirmed receipt, never erase it.
		if s.Status == "revoked" {
			s.ConfirmedAt, s.ExpiresAt = old.ConfirmedAt, old.ExpiresAt
		}
		if s.Revision == old.Revision {
			if old.Status != s.Status || !old.ExpiresAt.Equal(s.ExpiresAt) || !old.ConfirmedAt.Equal(s.ConfirmedAt) {
				return PurchaseGrant{}, ErrConflict
			}
			return *old, nil
		}
		if s.Status == "unknown" || s.Status == "pending" {
			return PurchaseGrant{}, ErrPendingPurchase
		}
	}
	if (p.Kind != "cosmetic" && s.Status == "purchased" && (s.ExpiresAt.IsZero() || s.ExpiresAt.Year() > 9999)) || (p.Kind == "cosmetic" && !s.ExpiresAt.IsZero()) {
		return PurchaseGrant{}, ErrInvalid
	}
	if !s.ConfirmedAt.IsZero() && (s.ConfirmedAt.Year() < 1 || s.ConfirmedAt.Year() > 9999) {
		return PurchaseGrant{}, ErrInvalid
	}
	if term := fixedPurchaseTerm(p.Kind); term != 0 && s.Status == "purchased" {
		legacyAdFree := old != nil && p.Kind == "ad_free" && old.ConfirmedAt.IsZero() && !old.ExpiresAt.IsZero() && (old.Status == "purchased" || old.Status == "revoked")
		if legacyAdFree {
			// Do not invent a confirmation instant for prior absolute-expiry
			// promises. They may be restored unchanged, never extended or resold.
			if !s.ConfirmedAt.IsZero() || !s.ExpiresAt.Equal(old.ExpiresAt) {
				return PurchaseGrant{}, ErrConflict
			}
			if old.Status != "purchased" {
				return PurchaseGrant{}, ErrIneligible
			}
		} else {
			if old != nil && !old.ConfirmedAt.IsZero() && (old.Status == "purchased" || old.Status == "revoked") && (!s.ConfirmedAt.Equal(old.ConfirmedAt) || !s.ExpiresAt.Equal(old.ExpiresAt)) {
				return PurchaseGrant{}, ErrConflict
			}
			if s.ConfirmedAt.IsZero() || !s.ExpiresAt.Equal(s.ConfirmedAt.Add(term)) {
				return PurchaseGrant{}, ErrInvalid
			}
		}
	}
	if old == nil {
		if p.Kind == "cosmetic" || p.Kind == "premium" {
			return PurchaseGrant{}, ErrIneligible
		}
	} else if (p.Kind == "cosmetic" || p.Kind == "premium") && s.Status == "purchased" && (old.Status != "purchased" || !s.ExpiresAt.Equal(old.ExpiresAt)) {
		return PurchaseGrant{}, ErrIneligible
	}
	return PurchaseGrant{VerifiedPurchase: s, Kind: p.Kind, Cosmetic: p.Cosmetic}, nil
}

// These approved paid access products share unlimited starts and ad removal.
// Standalone ad-free deliberately does not belong to this set: it never grants
// unlimited starts. No provider verifier or public purchase route is installed.
func paidAccessKind(kind string) bool {
	return kind == "daily" || kind == "weekly" || kind == "monthly" || kind == "yearly"
}

func fixedPurchaseTerm(kind string) time.Duration {
	switch kind {
	case "daily":
		return 24 * time.Hour
	case "ad_free":
		return 30 * 24 * time.Hour
	default:
		return 0 // Other approved products keep their verified provider expiry.
	}
}
