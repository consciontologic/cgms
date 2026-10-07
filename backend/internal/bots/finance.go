package bots

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
)

// LedgerProjection contains only a seat's own finances and debts to which it is a
// party. Financial seats are zero-based, unlike board observation seats.
type LedgerProjection struct {
	GameID string               `json:"game_id"`
	Seat   int                  `json:"seat"`
	Phase  string               `json:"phase"`
	Score  game.Amount          `json:"score"`
	Cash   game.Amount          `json:"cash"`
	Debts  []game.FinancialDebt `json:"debts"`
	Legal  []FinanceCommand     `json:"legal"`
}
type FinanceCommand struct {
	GameID string `json:"game_id"`
	Seat   int    `json:"seat"`
	Kind   string `json:"kind"`
}

// ProjectLedger is an adapter boundary; policies never receive the ledger itself.
func ProjectLedger(l game.FinancialLedger, seat int) (LedgerProjection, error) {
	if e := l.Validate(); e != nil {
		return LedgerProjection{}, e
	}
	if seat < 0 || seat >= len(l.Scores) {
		return LedgerProjection{}, fmt.Errorf("invalid financial observer")
	}
	o := LedgerProjection{GameID: l.GameID, Seat: seat, Phase: l.Phase, Score: l.Scores[seat], Cash: l.Cash[seat], Debts: []game.FinancialDebt{}, Legal: []FinanceCommand{}}
	for _, d := range l.Debts {
		if d.Debtor == seat || d.Creditor == seat {
			o.Debts = append(o.Debts, d)
		}
	}
	if l.Phase == "settlement" && !l.Finished[seat] {
		if _, e := l.FinishSettlement(seat); e == nil {
			o.Legal = append(o.Legal, FinanceCommand{l.GameID, seat, "finish"})
		}
	}
	return o, nil
}

// ChooseFinance explicitly answers the supported finish menu. Missing offers or
// unsupported choices never turn into an inferred refusal or silent completion.
func ChooseFinance(policy string, o LedgerProjection, budget int) (FinanceCommand, error) {
	if policy != "legal-random@v1" && policy != "heuristic@v1" {
		return FinanceCommand{}, fmt.Errorf("unknown financial policy")
	}
	if len(o.Legal) == 0 {
		return FinanceCommand{}, ErrUnsupported
	}
	if budget < len(o.Legal) {
		return FinanceCommand{}, ErrBudget
	}
	for _, c := range o.Legal {
		if c.GameID != o.GameID || c.Seat != o.Seat || o.Phase != "settlement" || c.Kind != "finish" {
			return FinanceCommand{}, fmt.Errorf("unsupported financial menu")
		}
	}
	return o.Legal[0], nil
}
