package game

import (
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"slices"
)

// ApplyBoardCommand bridges card effects and exact finances. Its State-only
// counterpart remains useful for narrow fixtures but cannot orchestrate a game.
func (l Lifecycle) ApplyBoardCommand(c Command) (Lifecycle, []Event, error) {
	if e := l.Validate(); e != nil {
		return l, nil, e
	}
	hash, e := canonical.Hash(c)
	if e != nil {
		return l, nil, e
	}
	if old, ok := l.Board.Commands[c.ID]; ok {
		if old != hash {
			return l, nil, errors.New("command conflict")
		}
		return l.Clone(), nil, nil
	}
	if c.Kind == "coup" {
		n, e := l.DeclareCoup(c)
		if e != nil {
			return l, nil, e
		}
		return n, []Event{{Kind: "coup", Actor: c.Actor, Amount: IntAmount(50)}}, nil
	}
	if !l.TurnStarted || l.Automatic != nil || l.EndTurnPending || len(l.PendingReceipts) > 0 {
		return l, nil, errors.New("card command outside actionable turn")
	}
	board, events, e := Apply(l.Board, c)
	if e != nil {
		return l, nil, e
	}
	n := l.Clone()
	n.Board = board
	for _, event := range events {
		switch event.Kind {
		case "ponzi-ready":
			n, e = n.resolvePonzi()
			if e != nil {
				return l, nil, e
			}
		case "kidnapper-obligation":
			if l.Board.Pending == nil || l.Board.Pending.Ability == nil {
				return l, nil, errors.New("missing kidnapping provenance")
			}
			n, e = n.recordObligation(l.Board.Pending.ID+"/doppelganger", l.Board.Pending.Ability.TargetSeat-1, event.Actor-1, event.Amount)
			if e != nil {
				return l, nil, e
			}
		case "combat-score", "ringleader":
			n.PendingReceipts = append(n.PendingReceipts, FinancialReceipt{Seat: event.Actor - 1, Amount: event.Amount})
			if event.Kind == "ringleader" {
				n.Promises, e = n.Promises.RecordAward(PromiseAward{ID: fmt.Sprintf("%s/ringleader/seat-%d", board.GameID, event.Actor), Payer: event.Actor - 1, Amount: event.Amount})
				if e != nil {
					return l, nil, e
				}
			}
		case "confined":
			if event.Actor == l.Board.Active {
				// Restore the old State-only scheduler and route through the full round
				// boundary after completed response costs and deferred work are retained.
				n.Board.Active = l.Board.Active
				n.Board.Turn = l.Board.Turn
				n.Board.Round = l.Board.Round
				n.Board.Pending = nil
				for i := range n.Board.Players {
					if i+1 != event.Actor {
						n.Board.Players[i].Confined = l.Board.Players[i].Confined
					}
				}
				n.EndTurnPending = true
			}
		}
	}
	if n.Board.Pending == nil && !postActionWorkPending(n.Board) {
		n, e = n.finishActionAccounting()
		if e != nil {
			return l, nil, e
		}
	}
	return n, events, nil
}
func (l Lifecycle) finishActionAccounting() (Lifecycle, error) {
	if len(l.PendingReceipts) > 0 {
		n := l.Clone()
		c, e := BeginSettlement(n.Ledger, n.PendingReceipts)
		if e != nil {
			return l, e
		}
		n.PendingReceipts = nil
		n.Automatic = &c
		return n, nil
	}
	if l.EndTurnPending {
		n := l.Clone()
		n.EndTurnPending = false
		return n.EndTurn()
	}
	return l, nil
}

// ContinueAfterAction consumes at most one deferred entitlement per call. An
// adapter records supplied permutations and can cancel at each durable boundary.
func (l Lifecycle) ContinueAfterAction(shuffle, recycle []string) (Lifecycle, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if l.Board.Pending != nil || l.Automatic != nil || l.Ending != "" {
		return l, errors.New("post-action boundary")
	}
	n := l.Clone()
	var e error
	if n.Board.NeedsShuffle {
		n.Board, e = ShuffleSupply(n.Board, shuffle)
		if e != nil {
			return l, e
		}
	} else if len(shuffle) > 0 {
		return l, errors.New("unexpected shuffle outcome")
	}
	if len(n.Board.DrawQueue) == 0 {
		n.Board, e = QueueCompensation(n.Board)
		if e != nil {
			return l, e
		}
	}
	if len(n.Board.DrawQueue) > 0 {
		n.Board, _, e = StepCompensation(n.Board, recycle)
		if e != nil {
			return l, e
		}
	} else if len(recycle) > 0 {
		return l, errors.New("unexpected recycle outcome")
	}
	if !postActionWorkPending(n.Board) {
		n, e = n.finishActionAccounting()
		if e != nil {
			return l, e
		}
	}
	return n, nil
}

func (l Lifecycle) resolvePonzi() (Lifecycle, error) {
	p := l.Board.Pending
	if p == nil || p.Kind != "ponzi" || p.Cursor != len(p.Responders) {
		return l, errors.New("Ponzi not resolved")
	}
	n := l.Clone()
	ids := []string{}
	for _, card := range n.Board.Cards {
		isPonzi := card.Allocation == "ponzi"
		for _, f := range n.Board.Formations {
			if f.ID == card.Allocation && f.Controller == p.Actor && f.Spec.Kind == "ponzi" {
				isPonzi = FormationFunctioning(n.Board, f.ID)
			}
		}
		if card.Controller == p.Actor && card.Zone == Formation && isPonzi && card.Card.Rank == 11 && card.AvailableFromRound <= n.Board.Round && (card.Card.Suit == Clubs || card.Card.Suit == Spades) {
			ids = append(ids, card.Card.ID)
		}
	}
	n.Board.Pending = nil
	if len(ids) != 4 || !eligible(n.Board, p.Actor) || !eligible(n.Board, p.Defender) || len(n.Board.CurrentSeries(p.Actor)) < 2 || !slices.Contains(n.Board.CurrentSeries(p.Actor), Clubs) || len(n.Board.Players[p.Defender-1].History) < 2 {
		return n, nil
	}
	var e error
	n.Board, e = n.Board.Move(ids, 0, Draw, false)
	if e != nil {
		return l, e
	}
	target := p.Defender - 1
	amount := n.Ledger.Cash[target]
	n.Ledger.Cash[target] = Amount{}
	n.Ledger.Scores[target] = n.Ledger.Scores[target].Sub(amount)
	partner := n.Board.Players[p.Actor-1].Ally
	if partner == 0 {
		n.PendingReceipts = append(n.PendingReceipts, FinancialReceipt{Seat: p.Actor - 1, Amount: amount})
	} else {
		half, _ := amount.Quo(IntAmount(2))
		n.PendingReceipts = append(n.PendingReceipts, FinancialReceipt{Seat: p.Actor - 1, Amount: half}, FinancialReceipt{Seat: partner - 1, Amount: half})
	}
	return n, nil
}

// recordObligation charges once; recycling available cash into the repayment
// queue is not a fresh score award. All inputs and charge provenance stay exact.
func (l Lifecycle) recordObligation(id string, debtor, creditor int, amount Amount) (Lifecycle, error) {
	if id == "" || !l.Ledger.validSeat(debtor) || creditor == debtor || creditor < 0 || !l.Ledger.validSeat(creditor) || amount.Sign() <= 0 {
		return l, errors.New("invalid player obligation")
	}
	for _, c := range l.Ledger.Charges {
		if c.ID == id {
			return l, errors.New("duplicate obligation")
		}
	}
	n := l.Clone()
	n.Ledger.NextSequence++
	n.Ledger.Scores[debtor] = n.Ledger.Scores[debtor].Sub(amount)
	n.Ledger.Charges = append(n.Ledger.Charges, FinancialCharge{Creditor: creditor, ID: id, GameID: n.Board.GameID, Debtor: debtor, Amount: amount})
	n.Ledger.Debts = append(n.Ledger.Debts, FinancialDebt{ID: id, ChargeID: id, GameID: n.Board.GameID, Debtor: debtor, Creditor: creditor, Remaining: amount, Sequence: n.Ledger.NextSequence})
	cash := n.Ledger.Cash[debtor]
	n.Ledger.Cash[debtor] = Amount{}
	n.Ledger.Scores[debtor] = n.Ledger.Scores[debtor].Sub(cash)
	n.PendingReceipts = append(n.PendingReceipts, FinancialReceipt{Seat: debtor, Amount: cash})
	return n, nil
}
