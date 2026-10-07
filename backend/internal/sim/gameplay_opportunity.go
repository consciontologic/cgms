package sim

import (
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"slices"
)

// The offline scheduler solicits optional out-of-turn opportunities in fixed
// round order between atomic transitions. Declining is a policy choice, never a
// game pass, turn advance, response or inferred refusal. Normal required actors
// keep their own menu. This scheduling convention must be disclosed in experiments.
func planOutOfTurn(l game.Lifecycle, streams map[string]*randomstream.Stream, choose GameplayPolicy, o GameplayOptions, st *GameplayStep, declined []int) (bool, error) {
	s := l.Board
	if l.Automatic != nil || l.Ending != "" || s.Phase != "playing" {
		return false, nil
	}
	required := s.Active
	if p := s.Pending; p != nil {
		if p.Decision != nil {
			required = p.Decision.Actor
		} else if p.Cursor < len(p.Responders) {
			required = p.Responders[p.Cursor]
		} else {
			required = p.Actor
		}
	}
	order := slices.Clone(s.Order)
	if len(order) == 0 {
		for _, p := range s.Players {
			order = append(order, p.Seat)
		}
	}
	for _, seat := range order {
		if slices.Contains(declined, seat) || s.Players[seat-1].Departed || !s.Players[seat-1].Connected {
			continue
		}
		obs, e := game.ObserveWithoutMenu(s, seat)
		if e != nil {
			return false, e
		}
		obs.Legal = nil
		add := func(c game.Command) {
			c.GameID = s.GameID
			c.Actor = seat
			c.ID = fmt.Sprintf("opportunity/%d/%d/%s/%d", st.Index, seat, c.Kind, c.Value)
			if _, _, e := game.Apply(s, c); e == nil {
				obs.Legal = append(obs.Legal, c)
			}
		}
		add(game.Command{Kind: "coup"})
		if seat != required && l.TurnStarted && s.Pending == nil && !postWork(s) {
			for _, p := range s.Players {
				if p.Seat != seat {
					add(game.Command{Kind: "ponzi", Value: p.Seat})
				}
			}
		}
		if seat != required && l.TurnStarted && s.Pending == nil && !postWork(s) {
			for _, proposal := range obs.Proposals {
				if proposal.Status == "offered" && proposal.Terms.To == seat {
					add(game.Command{Kind: "accept-offer", OfferID: proposal.ID, Revision: proposal.Revision})
					add(game.Command{Kind: "decline-offer", OfferID: proposal.ID, Revision: proposal.Revision})
				}
			}
		}
		if len(obs.Legal) == 0 {
			continue
		}
		obs.Legal = append(obs.Legal, game.Command{GameID: s.GameID, Actor: seat, Kind: "decline-out-of-turn"})
		obs.MenuCoverage = "optional-out-of-turn-v1"
		rng := streams[fmt.Sprint("bot-", seat)]
		before, _ := rng.Snapshot()
		p := o.Job.Policies[seat-1]
		d, e := choose(p.Policy+"@"+p.Version, obs, rng, o.BotBudget)
		if e != nil {
			return true, e
		}
		if d.Command.Actor != seat || d.Command.GameID != s.GameID {
			return true, errors.New("optional policy actor mismatch")
		}
		after, _ := rng.Snapshot()
		st.BotBefore = &before
		st.BotAfter = &after
		st.Decision = &d
		c := d.Command
		c.ID = fmt.Sprintf("gameplay/%d", st.Index)
		st.Command = &c
		st.Kind = "command"
		if c.Kind == "decline-out-of-turn" {
			st.Kind = "policy-wait"
		}
		return true, nil
	}
	return false, nil
}

var _ bots.Decision

func planPlayingFinance(m game.MatchLifecycle, streams map[string]*randomstream.Stream, o GameplayOptions, st *GameplayStep, declined []int) (bool, error) {
	l := m.Game
	if l.Ending != "" || l.Automatic != nil || l.Board.Pending != nil || postWork(l.Board) || len(l.PendingReceipts) > 0 || !l.TurnStarted {
		return false, nil
	}
	for _, p := range l.Board.Players {
		if !p.Connected || p.Departed || slices.Contains(declined, -p.Seat) {
			continue
		}
		obs, e := bots.ProjectGameplayFinance(l, p.Seat)
		if e != nil {
			return false, e
		}
		if len(obs.Legal) == 0 {
			continue
		}
		key := financeOpportunityKey(m, obs)
		if o.financialDeclines[p.Seat] == key {
			continue
		}
		st.OpportunityKey = key
		obs.Legal = append(obs.Legal, game.LifecycleOperation{GameID: l.Board.GameID, Actor: p.Seat, Kind: "decline-financial-opportunity"})
		consented := slices.Contains(m.NullificationConsents, p.Seat)
		if !consented {
			candidate := game.LifecycleOperation{GameID: l.Board.GameID, Actor: p.Seat, ID: fmt.Sprintf("gameplay/%d", st.Index), Kind: "nullify"}
			if _, e := m.ApplyOperation(candidate); e == nil {
				obs.Legal = append(obs.Legal, candidate)
			}
		}
		return true, recordFinanceChoice(obs, streams, o, st)
	}
	return false, nil
}

func recordFinanceChoice(obs bots.GameplayFinanceObservation, streams map[string]*randomstream.Stream, o GameplayOptions, st *GameplayStep) error {
	rng := streams[fmt.Sprint("bot-", obs.Seat)]
	before, _ := rng.Snapshot()
	profile := o.Job.Policies[obs.Seat-1]
	d, e := bots.ChooseGameplayFinancial(profile.Policy+"@"+profile.Version, obs, rng, o.BotBudget)
	if e != nil {
		return e
	}
	after, _ := rng.Snapshot()
	operation := d.Operation
	operation.ID = fmt.Sprintf("gameplay/%d", st.Index)
	st.Operation = &operation
	st.FinanceDecision = &d
	st.BotBefore = &before
	st.BotAfter = &after
	st.Kind = "operation"
	if operation.Kind == "decline-financial-opportunity" {
		st.Kind = "policy-wait"
	}
	return nil
}

func financeOpportunityKey(m game.MatchLifecycle, obs bots.GameplayFinanceObservation) string {
	obs.Legal = nil
	return hash(struct {
		Observation bots.GameplayFinanceObservation
		Consents    []int
	}{obs, m.NullificationConsents})
}

// Only own cash and debts plus the public fixed match horizon cross the policy
// boundary. Receivables, opponents' finances and full ledger never do.
func projectPolicyEconomy(m game.MatchLifecycle, seat int) bots.GameplayEconomy {
	e := bots.GameplayEconomy{Cash: m.Game.Ledger.Cash[seat-1].Add(game.IntAmount(0)), Owed: game.IntAmount(0), FutureGames: m.GameLimit - len(m.Game.Ledger.Completed) - 1}
	for _, d := range m.Game.Ledger.Debts {
		if d.Debtor == seat-1 {
			e.Owed = e.Owed.Add(d.Remaining)
		}
	}
	return e
}
