package offlineprobe

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
)

const FinanceVersion = "cgms-offline-finance-probe-v1"

// FinanceRequest is an omniscient synthetic financial scenario, not a gameplay
// command API. Card eligibility and promises are deliberately outside this probe.
type FinanceRequest struct {
	Schema    string               `json:"schema"`
	Ledger    game.FinancialLedger `json:"ledger"`
	GameLimit int                  `json:"game_limit"`
	Commands  []FinanceCommand     `json:"commands"`
}
type FinanceCommand struct {
	Start    *game.FinancialLedger   `json:"start"`
	GameID   string                  `json:"game_id"`
	Kind     string                  `json:"kind"`
	ID       string                  `json:"id"`
	Seat     int                     `json:"seat"`
	Other    int                     `json:"other"`
	Partner  int                     `json:"partner"`
	Amount   game.Amount             `json:"amount"`
	Debt     string                  `json:"debt"`
	Consent  []int                   `json:"consent"`
	Receipts []game.FinancialReceipt `json:"receipts"`
	NextGame string                  `json:"next_game"`
}
type FinanceFrame struct {
	Ledger       game.FinancialLedger    `json:"ledger"`
	Totals       []game.Amount           `json:"totals"`
	Ranks        []int                   `json:"ranks"`
	Complete     bool                    `json:"complete"`
	Observations []bots.LedgerProjection `json:"observations"`
	Decisions    []*bots.FinanceCommand  `json:"decisions"`
}
type FinanceResponse struct {
	Schema     string         `json:"schema"`
	Error      string         `json:"error,omitempty"`
	FailedStep int            `json:"failed_step"`
	Frames     []FinanceFrame `json:"frames,omitempty"`
}

func executeFinance(raw []byte) []byte {
	fail := func(code string, step int) []byte {
		b, _ := canonical.Marshal(FinanceResponse{Schema: FinanceVersion, Error: code, FailedStep: step})
		return b
	}
	var r FinanceRequest
	if canonical.Decode(raw, &r) != nil || r.Schema != FinanceVersion || r.GameLimit < 1 || r.GameLimit > 3 || len(r.Commands) > 64 {
		return fail("invalid_request", -1)
	}
	if r.Ledger.Validate() != nil || (len(r.Ledger.Completed) == r.GameLimit && r.Ledger.Phase != "finalized") || len(r.Ledger.Completed) > r.GameLimit || len(r.Ledger.Debts) > 64 {
		return fail("invalid_state", -1)
	}
	l := r.Ledger
	out := FinanceResponse{Schema: FinanceVersion, FailedStep: -1}
	for i := -1; i < len(r.Commands); i++ {
		if i >= 0 {
			var e error
			l, e = applyFinance(l, r.Commands[i], r.GameLimit)
			if e != nil {
				return fail("command_rejected", i)
			}
		}
		if l.Validate() != nil {
			return fail("invalid_state", i)
		}
		totals := l.MatchScores()
		f := FinanceFrame{Ledger: l, Totals: totals, Complete: l.Phase == "finalized" && len(l.Completed) == r.GameLimit}
		for seat, a := range totals {
			rank := 1
			for _, b := range totals {
				if b.Cmp(a) > 0 {
					rank++
				}
			}
			f.Ranks = append(f.Ranks, rank)
			obs, e := bots.ProjectLedger(l, seat)
			if e != nil {
				return fail("observation_failed", i)
			}
			f.Observations = append(f.Observations, obs)
			var decision *bots.FinanceCommand
			if len(obs.Legal) > 0 {
				d, e := bots.ChooseFinance("heuristic@v1", obs, 1)
				if e != nil {
					return fail("decision_budget", i)
				}
				decision = &d
			}
			f.Decisions = append(f.Decisions, decision)
		}
		out.Frames = append(out.Frames, f)
	}
	b, e := canonical.Marshal(out)
	if e != nil {
		return fail("output_budget", -1)
	}
	return b
}
func applyFinance(l game.FinancialLedger, c FinanceCommand, limit int) (game.FinancialLedger, error) {
	if c.GameID != l.GameID || len(l.Debts) > 64 || len(c.Receipts) > 64 {
		return l, errors.New("probe admission")
	}
	switch c.Kind {
	case "charge":
		return l.Charge(c.ID, c.Seat, c.Other, c.Amount)
	case "receipt":
		if l.Phase != "playing" {
			return l, errors.New("receipt outside play")
		}
		cursor, e := game.BeginSettlement(l, c.Receipts)
		if e != nil {
			return l, e
		}
		cursor, done, e := game.ResumeSettlement(cursor, 4096, true)
		if e != nil {
			return l, e
		}
		if !done {
			return l, errors.New("probe work budget; no partial commitment")
		}
		return cursor.Committed, nil
	case "ponzi":
		return l.Ponzi(c.Seat, c.Other, c.Partner)
	case "coup":
		return l.Coup(c.Seat)
	case "forgive":
		return l.Forgive(c.ID, c.Debt, c.Amount, c.Consent)
	case "close":
		return l.CloseOrdinary(c.Receipts)
	case "finish":
		return l.FinishSettlement(c.Seat)
	case "finalize":
		if len(l.Completed) >= limit {
			return l, errors.New("fixed match complete")
		}
		return l.Finalize()
	case "nullify":
		if c.Start == nil {
			return l, errors.New("start snapshot required")
		}
		return l.Nullify(*c.Start, c.Consent)
	case "next":
		if len(l.Completed) >= limit {
			return l, errors.New("fixed match complete")
		}
		return l.NextGame(c.NextGame)
	default:
		return l, errors.New("unsupported financial probe command")
	}
}
