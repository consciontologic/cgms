package bots

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"slices"
)

// These are development policies, not replacements for frozen v1 baselines.
// The ablation changes only retention; both policies use exactly the same menu.
func opportunityPolicy(policy string) bool {
	return policy == "bargaining@v1" || policy == "opportunity@v1" || policy == "opportunity-no-retain@v1" || policy == "opportunity-no-finance@v1"
}

func opportunityScore(policy string, o game.Observation, c game.Command) (int, []string) {
	score, reasons := gameplayScore("economic@v1", o, c)
	reason := reasons[0]
	if c.Kind == "attach" && policy != "opportunity-no-retain@v1" && o.Round < 10 {
		for _, v := range o.Cards {
			if slices.Contains(c.Cards, v.Card.ID) && v.Card.Rank >= 11 && v.Card.Rank <= 12 {
				score = -2
				reason = "reserve early jack/queen for future formation; release from round ten"
			}
		}
	}
	// A free ordinary gift has no alliance/Coup cost. Loan acceptance deliberately
	// remains conservative: its permanent alliance needs a separate value model.
	if c.Kind == "accept-offer" {
		for _, p := range o.Proposals {
			if p.ID != c.OfferID || p.Terms.To != o.Seat || p.Terms.Loan {
				continue
			}
			if len(p.Terms.Give) > 0 && len(p.Terms.Receive) == 0 {
				score = 90
				reason = "accept free ordinary cards without creating an alliance"
			}
		}
	}
	if c.Kind == "purchase" {
		// Number Spades have no direct end score. More draws preserve opportunities
		// even late; price/Inflation legality remains the menu's responsibility.
		score = 4 + c.Value
		reason = "buy authorized card opportunities; number Spades have no end award"
	}
	return score, []string{fmt.Sprintf("%s: priority %d", reason, score), "development opportunity policy; own and public board observation only", "royal reservation and acquisition are heuristics, not inferred hidden draws", "no cross-game debt model in board observation"}
}

func opportunityFinanceScore(o GameplayFinanceObservation, op game.LifecycleOperation, score int, reason string) (int, string) {
	if op.Kind != "promise-final-offer" {
		return score, reason
	}
	if op.Amount.Cmp(o.Cash) > 0 {
		return -200, "reject unaffordable reduced offer"
	}
	for _, d := range o.Debts {
		if d.Debtor == o.Seat-1 && d.Remaining.Sign() > 0 {
			return -200, "older own debts prohibit voluntary payment"
		}
	}
	// Exact ordering, not float conversion or money-dependent operation loops.
	// Rank legal amounts to choose the greatest affordable delivery. Higher delivery
	// can reach the canonical 40% reputation threshold; acceptance is still required.
	score = 150
	for _, other := range o.Legal {
		if other.Kind == op.Kind && other.PromiseID == op.PromiseID && other.Amount.Cmp(o.Cash) <= 0 && other.Amount.Cmp(op.Amount) < 0 {
			score++
		}
	}
	return score, "maximize affordable reduced delivery after older debts; no invented acceptance"
}

// GameplayEconomy is a detached authorized projection, never a ledger or match.
// FutureGames excludes the current game; cash expires but debt carries forward.
type GameplayEconomy struct {
	Cash        game.Amount
	Owed        game.Amount
	FutureGames int
}

func ChooseGameplayWithEconomy(policy string, o game.Observation, e GameplayEconomy, r *randomstream.Stream, budget int) (Decision, error) {
	return chooseGameplay(policy, o, &e, r, budget)
}
func opportunityEconomyScore(e GameplayEconomy, c game.Command, score int, reasons []string) (int, []string) {
	if reasons[len(reasons)-1] == "no cross-game debt model in board observation" {
		reasons[len(reasons)-1] = "uses detached own cash, owed debt and public remaining game count"
	} else {
		reasons = append(reasons, "uses detached own cash, owed debt and public remaining game count")
	}
	if e.FutureGames > 0 && e.Owed.Cmp(e.Cash) > 0 {
		bonus := 0
		switch c.Kind {
		case "ringleader":
			bonus = 40
		case "ponzi":
			bonus = 30
		case "open-formation":
			if c.Formation != nil && c.Formation.Kind == "underground" {
				bonus = 25
			}
		}
		if bonus > 0 {
			score += bonus
			reasons = append(reasons, fmt.Sprintf("own debt exceeds cash and persists into later games: income priority +%d, final priority %d", bonus, score))
		}
	}
	return score, reasons
}
