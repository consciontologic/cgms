// Package economy implements internal online economy contracts. It is separate
// from game score/debt and is not enabled until an approved price policy is wired.
package economy

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalid            = errors.New("invalid economy operation")
	ErrConflict           = errors.New("economy operation identity conflict")
	ErrAllowance          = errors.New("daily online allowance exhausted")
	ErrFunds              = errors.New("insufficient dirt")
	ErrIneligible         = errors.New("economy account or game ineligible")
	ErrProductUnavailable = fmt.Errorf("retired economy product: %w", ErrInvalid)
)

const FreeStarts = 3
const MaxBalance int64 = 1_000_000_000_000_000

// AdoptedPolicy is decision 0014's online economy. Configuration explicitly
// enables it; constructing a policy does not activate commerce or store adapters.
func AdoptedPolicy() Policy { return Policy{Reward: 10, PassDayPrice: 30} }

// Policy has no active defaults: reward and price require explicit configuration.
type Policy struct {
	Reward       int64 `json:"reward"`
	PassDayPrice int64 `json:"pass_day_price"`
}

func (p Policy) Validate() error {
	if p.Reward <= 0 || p.Reward > 1_000_000_000 || p.PassDayPrice <= 0 || p.PassDayPrice > 1_000_000_000 {
		return ErrInvalid
	}
	return nil
}
func Day(now time.Time) string { return now.UTC().Format("2006-01-02") }
func ExtendPass(now, expiry time.Time, days int) (time.Time, error) {
	if (days != 1 && days != 7) || now.IsZero() || now.Year() > 9998 || expiry.Year() > 9998 {
		return time.Time{}, ErrInvalid
	}
	start := now.UTC()
	if expiry.After(start) {
		start = expiry.UTC()
	}
	return start.Add(time.Duration(days) * 24 * time.Hour), nil
}
