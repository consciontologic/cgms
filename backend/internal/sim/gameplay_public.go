package sim

import (
	"context"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"sort"
	"strings"
)

type GameplayPublicCounter struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
type GameplayPublicGame struct {
	Instance             int           `json:"instance"`
	Voided               bool          `json:"voided"`
	Slot                 int           `json:"slot"`
	BoardEnded           bool          `json:"board_ended"`
	FinanciallyFinalized bool          `json:"financially_finalized"`
	Ending               string        `json:"ending"`
	Declarer             int           `json:"declarer"`
	FinalRound           int           `json:"final_round"`
	TurnsStarted         int           `json:"turns_started"`
	Transitions          int           `json:"transitions"`
	Redeals              int           `json:"redeals"`
	Scores               []game.Amount `json:"scores"`
	ScoreLeaders         []int         `json:"score_leaders"`
}

// GameplayPublicEconomy is retained only for in-process compatibility with v1.
// It is neither populated nor serialized or rendered by the public v2 surface.
type GameplayPublicEconomy struct {
	Slot           int         `json:"slot"`
	Round          int         `json:"round"`
	RecordedTotal  game.Amount `json:"recorded_total"`
	AvailableTotal game.Amount `json:"available_total"`
	UnpaidTotal    game.Amount `json:"unpaid_total"`
}
type GameplayPublicStats struct {
	StartedInstances             int                     `json:"started_instances"`
	VoidedGames                  int                     `json:"voided_games"`
	Schema                       string                  `json:"schema"`
	Population                   int                     `json:"population"`
	PlannedGames                 int                     `json:"planned_games"`
	StartedGames                 int                     `json:"started_games"`
	BoardEndedGames              int                     `json:"board_ended_games"`
	FinanciallyFinalizedGames    int                     `json:"financially_finalized_games"`
	FinalizedMatches             int                     `json:"finalized_matches"`
	NeverStartedGames            int                     `json:"never_started_games"`
	IncompleteReason             string                  `json:"incomplete_reason"`
	Transitions                  int                     `json:"transitions"`
	PolicyCalls                  int                     `json:"policy_calls"`
	PolicyOperations             int                     `json:"policy_operations"`
	MatchScores                  []game.Amount           `json:"match_scores"`
	MatchRanks                   []int                   `json:"match_ranks"`
	Games                        []GameplayPublicGame    `json:"games"`
	ActionUsage                  []GameplayPublicCounter `json:"action_usage"`
	PublicFormationOpportunities []GameplayPublicCounter `json:"public_formation_opportunities"`
	Economy                      []GameplayPublicEconomy `json:"-"`
	Narrative                    []string                `json:"narrative"`
}

func publicGameplayStop(r GameplayRun) string {
	switch r.ExitCode {
	case 0:
		return "none"
	case 2:
		return "invalid-input"
	case 3:
		return "unsupported-or-waiting"
	case 4:
		return "replay-or-invariant-failure"
	case 5:
		return "resource-budget-exhausted"
	case 6:
		return "io-failure"
	case 130:
		return "canceled"
	}
	return "unknown-stop"
}
func sortedPublicCounters(m map[string]int) []GameplayPublicCounter {
	out := []GameplayPublicCounter{}
	for k, v := range m {
		out = append(out, GameplayPublicCounter{k, v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
func publicEnding(s string) string {
	switch s {
	case "round-limit", "diamonds", "royals", "departures", "coup", "void":
		return s
	case "coup-pending":
		return "coup"
	}
	return "incomplete"
}
func publicAction(kind string) string {
	switch kind {
	case "attack", "baron-attack", "bomb-attack", "baron-bomb-attack", "infiltrator-attack", "purchase", "ponzi", "justice", "coup", "open-series", "open-formation", "attach", "take-back", "ringleader", "richer-sacrifice", "barricade-sacrifice", "dexter-assassination", "fate", "code", "kidnapper", "compensation", "main-inflation", "confinement", "inflation", "compensation-response", "advancement-defense", "numerical-defense", "negotiate", "pass", "end-turn", "declare-ordinary", "finish-settlement":
		return kind
	}
	return "other-boundary-or-private-decision"
}

// GameplayStats validates the private transcript, then extracts explicit public
// aggregates. Never serialize this function's private source alongside the report.
// Formation opportunities count visible functioning formations at their owner's
// idle board-policy decisions, not all legal activations or hidden-card chances.
func GameplayStats(ctx context.Context, r GameplayRun) (GameplayPublicStats, error) {
	return gameplayStatsWithFinance(ctx, r, nil)
}
func gameplayStatsWithFinance(ctx context.Context, r GameplayRun, private *GameplayPrivateFinance) (GameplayPublicStats, error) {
	out := GameplayPublicStats{Schema: "cgms-gameplay-public-v2", Population: r.Job.Population, PlannedGames: r.Job.Games, IncompleteReason: publicGameplayStop(r), Transitions: len(r.Trace), Games: []GameplayPublicGame{}, Narrative: []string{}, Economy: []GameplayPublicEconomy{}}
	if e := ReplayGameplayContext(ctx, r); e != nil {
		return out, e
	}
	m := r.Initial
	slots := map[string]int{}
	usage := map[string]int{}
	opps := map[string]int{}
	for _, st := range r.Trace {
		if ctx.Err() != nil {
			return out, ctx.Err()
		}
		if st.Kind == "deal" {
			if _, ok := slots[st.GameID]; !ok {
				slots[st.GameID] = len(out.Games)
				out.Games = append(out.Games, GameplayPublicGame{Slot: len(m.Game.Ledger.Completed) + 1, Instance: len(out.Games) + 1, Ending: "incomplete"})
			}
		}
		gi, started := slots[st.GameID]
		if !started {
			return out, errors.New("gameplay event before planned deal")
		}
		g := &out.Games[gi]
		g.Transitions++
		if st.Decision != nil {
			out.PolicyCalls++
			out.PolicyOperations += st.Decision.Operations
		}
		if st.FinanceDecision != nil {
			out.PolicyCalls++
			out.PolicyOperations += st.FinanceDecision.Operations
		}
		if st.Command != nil {
			usage[publicAction(st.Command.Kind)]++
		}
		if st.Operation != nil {
			usage[publicAction(st.Operation.Kind)]++
			if st.Operation.Kind == "begin-turn" {
				g.TurnsStarted++
			}
		}
		decisionActor := 0
		if st.Command != nil {
			decisionActor = st.Command.Actor
		} else if st.Operation != nil {
			decisionActor = st.Operation.Actor
		}
		if st.Decision != nil && st.Kind != "policy-wait" && m.Game.Board.Pending == nil && decisionActor == m.Game.Board.Active {
			for _, f := range m.Game.Board.Formations {
				if f.Controller == decisionActor && game.FormationFunctioning(m.Game.Board, f.ID) {
					opps[publicFormation(f.Spec.Kind)]++
				}
			}
		}
		next, _, e := applyGameplayStep(m, st)
		if e != nil {
			return out, e
		}
		if st.Kind == "deal" && len(next.Game.Board.Players[0].History) == 0 {
			g.Redeals++
		}
		if m.Game.Ending == "" && next.Game.Ending != "" {
			g.BoardEnded = true
			g.Ending = publicEnding(next.Game.Ending)
			g.Declarer = next.Game.Declarer
			out.Narrative = append(out.Narrative, fmt.Sprintf("Game %d board ended in round %d: %s; declaration seat %d.", g.Slot, next.Game.Board.Round, g.Ending, g.Declarer))
		}
		if len(next.Game.Ledger.Completed) > len(m.Game.Ledger.Completed) {
			g.FinanciallyFinalized = true
			last := next.Game.Ledger.Completed[len(next.Game.Ledger.Completed)-1]
			g.Scores = append([]game.Amount{}, last.Scores...)
			for i, a := range g.Scores {
				best := true
				for _, b := range g.Scores {
					if b.Cmp(a) > 0 {
						best = false
					}
				}
				if best {
					g.ScoreLeaders = append(g.ScoreLeaders, i+1)
				}
			}
			out.Narrative = append(out.Narrative, fmt.Sprintf("Game %d financial settlement finalized after explicit participant decisions; highest game-score seats %v.", g.Slot, g.ScoreLeaders))
		}

		if st.Kind == "deal" || st.Operation != nil && st.Operation.Kind == "begin-round" || len(next.Game.Ledger.Completed) > len(m.Game.Ledger.Completed) {
			boundary := st.Kind
			if st.Operation != nil {
				boundary = st.Operation.Kind
			}
			if len(next.Game.Ledger.Completed) > len(m.Game.Ledger.Completed) {
				boundary = "financially-finalized"
			}
			private.capture(boundary, g.Slot, next.Game.Board.Round, next.Game.Ledger)
		}
		if next.Game.Ending == "void" {
			g.Voided = true
			g.Ending = "void"
			g.BoardEnded = false
		}
		if st.Kind != "next-game" {
			g.FinalRound = next.Game.Board.Round
		}
		m = next
	}
	for _, g := range out.Games {
		if g.BoardEnded {
			out.BoardEndedGames++
		}
		if g.FinanciallyFinalized {
			out.FinanciallyFinalizedGames++
		}
	}
	checkpointSlot := 1
	if len(out.Games) > 0 {
		checkpointSlot = out.Games[len(out.Games)-1].Slot
	}
	private.capture("checkpoint", checkpointSlot, m.Game.Board.Round, m.Game.Ledger)
	out.StartedGames, out.StartedInstances, out.VoidedGames = committedGameplayStarts(r)
	out.NeverStartedGames = out.PlannedGames - out.StartedGames
	if out.NeverStartedGames < 0 {
		out.NeverStartedGames = 0
	}
	if out.FinanciallyFinalizedGames == out.PlannedGames && r.ExitCode == 0 {
		out.FinalizedMatches = 1
	}
	if r.Match.Game.Ledger.Phase == "finalized" {
		out.MatchScores = r.Match.Game.Ledger.MatchScores()
		out.MatchRanks = r.Match.Ranks()
	}
	out.ActionUsage = sortedPublicCounters(usage)
	out.PublicFormationOpportunities = sortedPublicCounters(opps)
	return out, nil
}
func GameplayPublicReport(ctx context.Context, r GameplayRun) (string, error) {
	s, e := GameplayStats(ctx, r)
	if e != nil {
		return "", e
	}
	return RenderGameplayPublicReport(s), nil
}

// RenderGameplayPublicReport renders an already validated public projection.
func RenderGameplayPublicReport(s GameplayPublicStats) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Gameplay findings\n\nAccepted-rule development run; %d-player population. This run supplies integration evidence, not a held-out strength or balance estimate.\n\nPlanned games: %d; started: %d; board-ended: %d; financially finalized: %d; finalized matches: %d; never started: %d. Incomplete reason: %s.\n\n", s.Population, s.PlannedGames, s.StartedGames, s.BoardEndedGames, s.FinanciallyFinalizedGames, s.FinalizedMatches, s.NeverStartedGames, s.IncompleteReason)
	fmt.Fprintf(&b, "Physical game instances started: %d; voided instances: %d. Started/planned games count configured nonvoid slots.\n\n", s.StartedInstances, s.VoidedGames)
	fmt.Fprintln(&b, "| Slot / instance | Ending | Declarer | Score leaders | Rounds | Turns started | Transitions | Redeals | Final scores |\n|---|---|---:|---|---:|---:|---:|---:|---|")
	for _, g := range s.Games {
		fmt.Fprintf(&b, "| %d / %d | %s | %d | %v | %d | %d | %d | %d | %v |\n", g.Slot, g.Instance, g.Ending, g.Declarer, g.ScoreLeaders, g.FinalRound, g.TurnsStarted, g.Transitions, g.Redeals, g.Scores)
	}
	fmt.Fprintf(&b, "\nCumulative match scores: %v. Shared competition ranks: %v. Declaration outcomes, game score leaders and cumulative ranks are distinct.\n\n%d transitions; %d policy decisions; %d recorded policy operations. Deterministic budgets are resource limits, never draws or losses.\n\n", s.MatchScores, s.MatchRanks, s.Transitions, s.PolicyCalls, s.PolicyOperations)
	fmt.Fprintln(&b, "| Action category | Recorded uses |\n|---|---:|")
	for _, c := range s.ActionUsage {
		fmt.Fprintf(&b, "| %s | %d |\n", c.Name, c.Count)
	}
	fmt.Fprintln(&b, "\nPublic formation opportunities count functioning exposed formations at their owner's recorded idle card decisions. They exclude hidden-card opportunities and are not denominators for causal effectiveness. Raw usage does not establish usefulness or balance.\n\n| Public formation | Observed supported decision opportunities |\n|---|---:|")
	for _, c := range s.PublicFormationOpportunities {
		fmt.Fprintf(&b, "| %s | %d |\n", c.Name, c.Count)
	}
	fmt.Fprintln(&b, "\nCurrent cash, debt and unfinalized cumulative scores/ranks are private and omitted. Completed game scores remain public; cumulative scores/ranks are shown only at a finalized boundary.")
	fmt.Fprintln(&b, "\nRepresentative public ending narrative:")
	for _, line := range s.Narrative {
		fmt.Fprintf(&b, "\n- %s\n", line)
	}
	fmt.Fprintln(&b, "\nNo hidden cards, private seeds, proposal terms or forensic policy traces are included. Full rule conformance requires the separate rule fixture matrix. Human enjoyment, comprehension and social bargaining require focused human playtests.")
	return b.String()
}

func publicFormation(kind string) string {
	switch kind {
	case "fate", "baron", "underground", "people", "great-people", "coup", "justice", "code", "kidnapper", "ponzi":
		return kind
	}
	return "unclassified-public-formation"
}
