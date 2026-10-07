package sim

import (
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
)

// GameplayPrivateFinance is forensic output. Never give it to policies or render
// it in ordinary reports; party records use canonical zero-based ledger seats.
const privateFinanceMaxBytes = canonical.MaxBytes / 2

type GameplayPrivateFinance struct {
	Status           string                           `json:"status"`
	Reason           string                           `json:"reason"`
	OmittedSnapshots int                              `json:"omitted_snapshots"`
	MaxBytes         int                              `json:"max_bytes"`
	Schema           string                           `json:"schema"`
	Privacy          string                           `json:"privacy"`
	Observer         string                           `json:"observer"`
	Snapshots        []GameplayPrivateFinanceSnapshot `json:"snapshots"`
}
type GameplayPrivateFinanceSnapshot struct {
	Boundary       string               `json:"boundary"`
	Slot           int                  `json:"slot"`
	Round          int                  `json:"round"`
	RecordedTotal  game.Amount          `json:"recorded_total"`
	AvailableTotal game.Amount          `json:"available_total"`
	UnpaidTotal    game.Amount          `json:"unpaid_total"`
	Ledger         game.FinancialLedger `json:"ledger"`
}

func newGameplayPrivateFinance() *GameplayPrivateFinance {
	return &GameplayPrivateFinance{Schema: "cgms-gameplay-finance-private-v2", Status: "complete", MaxBytes: privateFinanceMaxBytes, Privacy: "restricted", Observer: "omniscient", Snapshots: []GameplayPrivateFinanceSnapshot{}}
}
func (d *GameplayPrivateFinance) capture(boundary string, slot, round int, l game.FinancialLedger) {
	if d == nil {
		return
	}
	if d.Status == "incomplete" {
		d.OmittedSnapshots++
		return
	}
	s := GameplayPrivateFinanceSnapshot{Boundary: boundary, Slot: slot, Round: round, Ledger: l.Clone()}
	for _, a := range l.Scores {
		s.RecordedTotal = s.RecordedTotal.Add(a)
	}
	for _, a := range l.Cash {
		s.AvailableTotal = s.AvailableTotal.Add(a)
	}
	for _, a := range l.Debts {
		s.UnpaidTotal = s.UnpaidTotal.Add(a.Remaining)
	}
	d.Snapshots = append(d.Snapshots, s)
	// Reserve the largest terminal metadata before retaining a snapshot, including
	// every canonical-safe decimal digit of the omitted counter.
	terminal := *d
	terminal.Status = "incomplete"
	terminal.Reason = "diagnostic-byte-budget"
	terminal.OmittedSnapshots = 9007199254740991 // canonical maximum exact JSON integer
	raw, err := canonical.Marshal(terminal)
	if err != nil || len(raw) > privateFinanceMaxBytes {
		d.Snapshots = d.Snapshots[:len(d.Snapshots)-1]
		d.Status = "incomplete"
		d.Reason = "diagnostic-byte-budget"
		d.OmittedSnapshots++
	}
}
