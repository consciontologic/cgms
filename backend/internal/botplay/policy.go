// Package botplay shares bounded, observation-only opponent decisions between
// the offline host and the authoritative online scheduler. It never persists
// state, starts workers, authorizes identities or publishes private envelopes.
package botplay

import (
	"context"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
)

const menuBudget = 50000

var ErrNoDecision = errors.New("no bot decision")

// Coup checks the policy's immediate-win exception at any actor boundary using
// only its own authorized inventory and the authoritative Coup validator. The
// fixed deck scan avoids regenerating every bot's complete action menu on each
// human wait; unrelated decisions remain owned by their player.
func Coup(env matchstore.Envelope, seat int) (matchstore.Request, bool, error) {
	o, err := game.ObserveWithoutMenu(env.Match.Game.Board, seat)
	if err != nil {
		return matchstore.Request{}, false, err
	}
	queens := 0
	for _, card := range o.Cards {
		if card.Controller == seat && card.Card.Rank == 12 && (card.Card.Suit == game.Hearts || card.Card.Suit == game.Diamonds) {
			queens++
		}
	}
	if queens < 4 {
		return matchstore.Request{}, false, nil
	}
	op := game.LifecycleOperation{ID: "bot-coup-check", GameID: o.GameID, Actor: seat, Kind: "coup"}
	request := matchstore.Request{Operation: &op}
	if err = matchstore.ValidateActorRequest(env, seat, request); err != nil {
		return matchstore.Request{}, false, nil
	}
	return request, true, nil
}

func Actions(env matchstore.Envelope, version int64, seat int) ([]matchstore.Request, error) {
	return projectedActions(context.Background(), env, version, seat, true)
}

func projectedActions(ctx context.Context, env matchstore.Envelope, version int64, seat int, validateAll bool) ([]matchstore.Request, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if seat < 1 || seat > len(env.Match.Game.Board.Players) {
		return nil, errors.New("invalid local actor")
	}
	if _, auto := matchstore.NextServer(env, version); auto {
		return nil, nil
	}
	l := env.Match.Game
	out := []matchstore.Request{}
	add := func(r matchstore.Request) {
		if ctx.Err() != nil {
			return
		}
		if !validateAll && r.Operation == nil {
			out = append(out, r)
			return
		}
		if e := matchstore.ValidateActorRequest(env, seat, r); e == nil {
			out = append(out, r)
		}
	}
	op := func(kind string) game.LifecycleOperation {
		return game.LifecycleOperation{ID: fmt.Sprintf("offline/%d/%d/%s", version, seat, kind), GameID: l.Board.GameID, Actor: seat, Kind: kind}
	}
	if l.RoundClosed && l.Ending == "" && !l.DeparturesResolved {
		stay := op("departure-choice")
		add(matchstore.Request{Operation: &stay})
		leave := stay
		leave.ID += "/leave"
		leave.Accept = true
		add(matchstore.Request{Operation: &leave})
		return out, nil
	}
	if l.Ledger.Phase == "playing" && l.TurnStarted && !l.RoundClosed {
		o, err := game.GameplayObservationContext(ctx, l.Board, seat, menuBudget)
		if err != nil {
			return nil, err
		}
		for _, c := range o.Legal {
			switch c.Kind {
			case "end-turn", "coup", "ordinary-victory":
				kind := c.Kind
				if kind == "ordinary-victory" {
					kind = "declare-ordinary"
				}
				v := op(kind)
				if c.Kind == "ordinary-victory" {
					v.Threshold = "diamonds"
					if c.Value == 15 {
						v.Threshold = "royals"
					}
					v.ID += "/" + v.Threshold
				}
				add(matchstore.Request{Operation: &v})
			default:
				v := c
				add(matchstore.Request{Command: &v})
			}
		}
	}
	financial, err := bots.ProjectGameplayFinance(l, seat)
	if err != nil {
		return nil, err
	}
	for _, v := range financial.Legal {
		v.Seats = nil
		add(matchstore.Request{Operation: &v})
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Suggest evaluates only projected legal actions. Difficulty selects distinct
// transparent heuristics; ordinal playing-strength claims require P03 evidence.
func Suggest(env matchstore.Envelope, version int64, seat int, difficulty string) (matchstore.Request, error) {
	return suggest(env, version, seat, difficulty, false)
}

func suggest(env matchstore.Envelope, version int64, seat int, difficulty string, acquisitionProgress bool) (matchstore.Request, error) {
	actions, err := Actions(env, version, seat)
	if err != nil {
		return matchstore.Request{}, err
	}
	return chooseProjected(env, seat, difficulty, acquisitionProgress, actions)
}

func chooseProjected(env matchstore.Envelope, seat int, difficulty string, acquisitionProgress bool, actions []matchstore.Request) (matchstore.Request, error) {
	if len(actions) == 0 {
		return matchstore.Request{}, ErrNoDecision
	}
	policy := "economic@v1"
	if difficulty == "standard" {
		policy = "pressure@v1"
	}
	if difficulty == "advanced" {
		policy = "opportunity@v1"
	}
	board, err := game.ObserveWithoutMenu(env.Match.Game.Board, seat)
	if err != nil {
		return matchstore.Request{}, err
	}
	for _, a := range actions {
		if a.Command != nil {
			// Returning payment to an empty supply can redraw those same Spades,
			// then reopen them indefinitely. Online bots deliberately buy only when
			// quantity exceeds their own paid physical-card count. This uses legal
			// own-payment information, not hidden draw identities or a retry timer.
			// Humans and frozen/offline policies retain every legal purchase.
			if acquisitionProgress && a.Command.Kind == "purchase" && a.Command.Value <= len(a.Command.Cards) {
				continue
			}
			board.Legal = append(board.Legal, *a.Command)
		}
		if a.Operation != nil {
			op := a.Operation
			switch op.Kind {
			case "departure-choice":
				if !op.Accept {
					return a, nil
				}
			case "end-turn", "coup", "declare-ordinary":
				kind := op.Kind
				if kind == "declare-ordinary" {
					kind = "ordinary-victory"
				}
				board.Legal = append(board.Legal, game.Command{ID: op.ID, GameID: op.GameID, Actor: seat, Kind: kind})
			}
		}
	}
	if len(board.Legal) > 0 {
		d, e := bots.ChooseGameplay(policy, board, nil, menuBudget*2)
		if e != nil {
			return matchstore.Request{}, e
		}
		for _, a := range actions {
			if a.Command != nil && a.Command.ID == d.Command.ID || a.Operation != nil && a.Operation.ID == d.Command.ID {
				return a, nil
			}
		}
	}
	f, e := bots.ProjectGameplayFinance(env.Match.Game, seat)
	if e != nil {
		return matchstore.Request{}, e
	}
	f.Legal = nil
	for _, a := range actions {
		if a.Operation != nil {
			switch a.Operation.Kind {
			case "promise-accept", "promise-pay", "promise-refuse", "promise-final-offer", "promise-final-answer", "finish-settlement":
				f.Legal = append(f.Legal, *a.Operation)
			}
		}
	}
	if len(f.Legal) == 0 {
		return matchstore.Request{}, ErrNoDecision
	}
	d, e := bots.ChooseGameplayFinancial(policy, f, nil, menuBudget*2)
	if e != nil {
		return matchstore.Request{}, e
	}
	return matchstore.Request{Operation: &d.Operation}, nil
}

// SuggestOnline uses the online runtime's fair-progress safeguards without
// changing the frozen strategy policies or offline replay behavior.
func SuggestOnline(env matchstore.Envelope, version int64, seat int, difficulty string) (matchstore.Request, error) {
	return SuggestOnlineContext(context.Background(), env, version, seat, difficulty)
}

func SuggestOnlineContext(ctx context.Context, env matchstore.Envelope, version int64, seat int, difficulty string) (matchstore.Request, error) {
	// The engine has already validated its bounded gameplay menu. Score that
	// authorized projection first, then apply the lifecycle validator to the
	// selected request. Eagerly cloning the entire lifecycle for every candidate
	// can exhaust the runtime deadline before its first database write. Removing
	// a rejected candidate and choosing again retains the same deterministic
	// winner as eager validation; no unchecked request reaches the dispatcher.
	actions, err := projectedActions(ctx, env, version, seat, false)
	if err != nil {
		return matchstore.Request{}, err
	}
	return chooseOnline(ctx, env, seat, difficulty, actions)
}

func chooseOnline(ctx context.Context, env matchstore.Envelope, seat int, difficulty string, actions []matchstore.Request) (matchstore.Request, error) {
	board, err := game.ObserveWithoutMenu(env.Match.Game.Board, seat)
	if err != nil {
		return matchstore.Request{}, err
	}
	for _, a := range actions {
		if a.Command != nil {
			if a.Command.Kind == "purchase" && a.Command.Value <= len(a.Command.Cards) {
				continue
			}
			board.Legal = append(board.Legal, *a.Command)
		} else if a.Operation != nil {
			op := a.Operation
			kind := op.Kind
			if kind == "declare-ordinary" {
				kind = "ordinary-victory"
			}
			if kind == "end-turn" || kind == "coup" || kind == "ordinary-victory" {
				board.Legal = append(board.Legal, game.Command{ID: op.ID, GameID: op.GameID, Actor: seat, Kind: kind})
			}
		}
	}
	if bots.GameplayV1OrderingHasTies(board) {
		// Preserve the exact historical tie outcome, including removal order.
		valid := actions[:0]
		for _, a := range actions {
			if err := ctx.Err(); err != nil {
				return matchstore.Request{}, err
			}
			if matchstore.ValidateActorRequest(env, seat, a) == nil {
				valid = append(valid, a)
			}
		}
		actions = valid
	}
	for len(actions) > 0 {
		if err := ctx.Err(); err != nil {
			return matchstore.Request{}, err
		}
		choice, err := chooseProjected(env, seat, difficulty, true, actions)
		if err != nil {
			return matchstore.Request{}, err
		}
		if err = matchstore.ValidateActorRequest(env, seat, choice); err == nil {
			if err := ctx.Err(); err != nil {
				return matchstore.Request{}, err
			}
			return choice, nil
		}
		removed := false
		for i, candidate := range actions {
			if candidate.Command != nil && choice.Command != nil && candidate.Command.ID == choice.Command.ID ||
				candidate.Operation != nil && choice.Operation != nil && candidate.Operation.ID == choice.Operation.ID {
				actions = append(actions[:i], actions[i+1:]...)
				removed = true
				break
			}
		}
		if !removed {
			return matchstore.Request{}, ErrNoDecision
		}
	}
	return matchstore.Request{}, ErrNoDecision
}
