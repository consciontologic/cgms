package matchstore

import "github.com/metaphy6/cgms/backend/internal/game"

// OnlineDetails exposes lifecycle context without automatic financial workspaces.
type OnlineDetails struct {
	BotSeats         []int                      `json:"bot_seats,omitempty"`
	BotDifficulty    string                     `json:"bot_difficulty,omitempty"`
	RoundClosed      bool                       `json:"round_closed"`
	DeparturePending bool                       `json:"departure_pending"`
	RulesContext     game.RulesContext          `json:"rules_context"`
	Corrections      []game.FinancialCorrection `json:"corrections,omitempty"`
	GamesPerMatch    int                        `json:"games_per_match"`
	CompletedGames   int                        `json:"completed_games"`
	FinancialPhase   string                     `json:"financial_phase"`
	Ending           string                     `json:"ending,omitempty"`
	Declarer         int                        `json:"declarer,omitempty"`
	AutomaticPending bool                       `json:"automatic_pending"`
	ServerPending    bool                       `json:"server_pending"`
	TurnStarted      bool                       `json:"turn_started"`
	Finished         bool                       `json:"finished"`
	DecisionStage    string                     `json:"decision_stage,omitempty"`
	DrawQueue        []game.DrawEntitlement     `json:"draw_queue,omitempty"`
	Alliances        map[int]int                `json:"alliances"`
	Results          []game.FinancialResult     `json:"results"`
	OwnMatchScore    game.Amount                `json:"own_match_score"`
	Ranks            []int                      `json:"ranks"`
	Reputation       []game.PromiseOutcome      `json:"reputation,omitempty"`
}

func onlineProjection(e Envelope, seat int) OnlineDetails {
	m := e.Match.Clone()
	l := m.Game
	context, _ := game.ProjectRulesContext(l.Board, seat)
	o := OnlineDetails{GamesPerMatch: m.GameLimit, CompletedGames: len(l.Ledger.Completed), FinancialPhase: l.Ledger.Phase, Ending: l.Ending, Declarer: l.Declarer, AutomaticPending: l.Automatic != nil || l.PromisePayment != nil, Finished: l.Ledger.Finished[seat-1], DrawQueue: l.Board.DrawQueue, Alliances: map[int]int{}, Results: l.Ledger.Completed, OwnMatchScore: l.Ledger.MatchScores()[seat-1]}
	// Readiness follows the same bounded planner as the runtime, including deal,
	// initiative and begin-turn. A provisional Active seat is not a playable turn.
	_, o.ServerPending = NextServer(e, 0)
	o.RoundClosed = l.RoundClosed
	o.TurnStarted = l.TurnStarted
	o.DeparturePending = departurePending(l, seat)
	if e.BotDifficulty != "" {
		o.BotDifficulty = e.BotDifficulty
		for bot := 2; bot <= len(l.Board.Players); bot++ {
			o.BotSeats = append(o.BotSeats, bot)
		}
	}
	for _, correction := range l.Ledger.Corrections {
		if correction.Debtor == seat-1 {
			o.Corrections = append(o.Corrections, correction)
		}
	}
	o.RulesContext = context
	if l.Board.Pending != nil && l.Board.Pending.Decision != nil {
		o.DecisionStage = l.Board.Pending.Decision.Stage
	}
	for _, p := range l.Board.Players {
		o.Alliances[p.Seat] = p.Ally
	}
	if outcomes, err := l.Promises.Outcomes(); err == nil {
		for _, v := range outcomes {
			if v.Payer == seat-1 || v.Recipient == seat-1 {
				o.Reputation = append(o.Reputation, v)
			}
		}
	}
	if l.Ledger.Phase == "finalized" {
		o.Ranks = m.Ranks()
	}
	return o
}

func departurePending(l game.Lifecycle, seat int) bool {
	if !l.RoundClosed || l.DeparturesResolved || l.Automatic != nil || l.Ending != "" || l.Board.Round >= 13 || l.Board.Pending != nil || l.Board.Players[seat-1].Departed {
		return false
	}
	for _, choice := range l.DepartureChoices {
		if choice.Seat == seat {
			return false
		}
	}
	return true
}
