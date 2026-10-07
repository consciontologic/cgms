package game

import (
	"errors"
	"fmt"
	"slices"
)

// DeclareCoup joins card adjudication to resumable replacement accounting. No
// automatic queue may be interrupted, even by an otherwise immediate Coup.
func (l Lifecycle) DeclareCoup(c Command) (Lifecycle, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if l.Automatic != nil || c.Kind != "coup" || l.Ending != "" {
		return l, errors.New("Coup boundary")
	}
	board, _, e := Apply(l.Board, c)
	if e != nil {
		return l, e
	}
	n := l.Clone()
	n.Board = board
	n.TurnStarted = false
	n.Ending = "coup"
	n.Declarer = c.Actor
	if len(n.PendingReceipts) > 0 {
		cursor, e := BeginSettlement(n.Ledger, n.PendingReceipts)
		if e != nil {
			return l, e
		}
		n.PendingReceipts = nil
		n.Automatic = &cursor
		n.Ending = "coup-pending"
		n.EndTurnPending = false
		return n, nil
	}
	return n.beginCoupAccounting()
}
func (l Lifecycle) beginCoupAccounting() (Lifecycle, error) {
	n := l.Clone()
	n.Ending = "coup"
	n.EndTurnPending = false
	c := Command{Actor: n.Declarer}
	board := n.Board
	ledger := n.Ledger
	for seat := range ledger.Scores {
		if ledger.Departed[seat] {
			continue
		}
		ledger.Scores[seat] = Amount{}
		ledger.Cash[seat] = Amount{}
		for _, d := range ledger.Debts {
			if d.Debtor == seat && d.GameID == ledger.GameID {
				ledger.Scores[seat] = ledger.Scores[seat].Sub(d.Remaining)
			}
		}
	}
	ledger.Corrections = slices.DeleteFunc(ledger.Corrections, func(c FinancialCorrection) bool { return !c.MatchLevel && !ledger.Departed[c.Debtor] })
	for seat := range ledger.Scores {
		if seat == c.Actor-1 || ledger.Departed[seat] {
			continue
		}
		id := fmt.Sprintf("%s/coup/%d", ledger.GameID, seat)
		amount := IntAmount(50)
		ledger.Scores[seat] = ledger.Scores[seat].Sub(amount)
		ledger.NextSequence++
		ledger.Charges = append(ledger.Charges, FinancialCharge{Creditor: -1, ID: id, GameID: ledger.GameID, Debtor: seat, Amount: amount})
		ledger.Debts = append(ledger.Debts, FinancialDebt{ID: id, ChargeID: id, GameID: ledger.GameID, Debtor: seat, Creditor: -1, Remaining: amount, Sequence: ledger.NextSequence})
	}
	cursor, e := BeginSettlement(ledger, []FinancialReceipt{{Seat: c.Actor - 1, Amount: IntAmount(50)}})
	if e != nil {
		return l, e
	}
	n.Promises, e = n.Promises.RecordAward(PromiseAward{ID: fmt.Sprintf("%s/coup/seat-%d", board.GameID, c.Actor), Payer: c.Actor - 1, Amount: IntAmount(50)})
	if e != nil {
		return l, e
	}
	n.Promises, e = n.Promises.CloseBoard()
	if e != nil {
		return l, e
	}
	n.Board = ExpireProposals(n.Board)
	n.Ledger = ledger
	n.Automatic = &cursor
	return n, nil
}

// DepartSeats applies the whole declared boundary set atomically. It leaves a
// returned-card shuffle pending; no arbitrary insertion order or draw is implied.
func (l Lifecycle) DepartSeats(seats []int) (Lifecycle, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if !l.RoundClosed || l.Automatic != nil || l.Ending != "" || l.Board.Round >= 13 || l.Board.Pending != nil {
		return l, errors.New("departure boundary")
	}
	seen := map[int]bool{}
	for _, seat := range seats {
		if seat < 1 || seat > len(l.Board.Players) || seen[seat] || l.Board.Players[seat-1].Departed {
			return l, errors.New("invalid departure set")
		}
		seen[seat] = true
	}
	n := l.Clone()
	for _, seat := range seats {
		var e error
		n.Board, e = n.Board.DepartPlayer(seat)
		if e != nil {
			return l, e
		}
		n.Ledger, e = n.Ledger.Depart(seat - 1)
		if e != nil {
			return l, e
		}
	}
	n.Board = ExpireProposals(n.Board)
	active := 0
	for _, p := range n.Board.Players {
		if !p.Departed {
			active++
		}
	}
	if active < 2 {
		return n.closeOrdinary("departures", 0)
	}
	return n, nil
}

// Nullify requires every affected participant, including departed seats, and
// restores the exact start snapshot. Consent uses public one-based seats.
func (m MatchLifecycle) Nullify(consent []int) (MatchLifecycle, error) {
	if m.Game.Validate() != nil || m.Game.Automatic != nil {
		return m, errors.New("nullification boundary")
	}
	zero := []int{}
	seen := map[int]bool{}
	for _, seat := range consent {
		if seat < 1 || seat > len(m.Game.Board.Players) || seen[seat] {
			return m, errors.New("invalid consent")
		}
		seen[seat] = true
		zero = append(zero, seat-1)
	}
	ledger, e := m.Game.Ledger.Nullify(m.StartLedger, zero)
	if e != nil {
		return m, e
	}
	board, e := CloseCardBoard(m.Game.Board)
	if e != nil {
		return m, e
	}
	n := m.Clone()
	n.Game = m.Game.Clone()
	n.Game.Board = board
	n.Game.Ledger = ledger
	n.Game.Ending = "void"
	n.Game.PendingReceipts = nil
	n.Game.ForgivenessConsents = nil
	n.Game.EndTurnPending = false
	n.Game.Declarer = 0
	n.Game.TurnStarted = false
	n.Game.Promises = n.Game.Promises.Nullify()
	n.Game.Board = ExpireProposals(n.Game.Board)
	return n, nil
}

func (l Lifecycle) FinishSettlement(seat int) (Lifecycle, error) {
	if l.Validate() != nil || l.Automatic != nil || l.Ending == "" || seat < 1 || seat > len(l.Board.Players) {
		return l, errors.New("finish boundary")
	}
	n := l.Clone()
	var e error
	n.Promises, e = n.Promises.Finish(seat - 1)
	if e != nil {
		return l, e
	}
	n.Ledger, e = n.Ledger.FinishSettlement(seat - 1)
	if e != nil {
		return l, e
	}
	return n, nil
}
func (l Lifecycle) Finalize() (Lifecycle, error) {
	if l.Validate() != nil || l.Automatic != nil {
		return l, errors.New("finalization boundary")
	}
	n := l.Clone()
	var e error
	n.Promises, e = n.Promises.Finalize()
	if e != nil {
		return l, e
	}
	n.Ledger, e = n.Ledger.Finalize()
	if e != nil {
		return l, e
	}
	return n, nil
}

func (l Lifecycle) applyFinancialOperation(op LifecycleOperation) (Lifecycle, error) {
	if op.Actor < 1 || op.Actor > len(l.Board.Players) || l.Automatic != nil || l.Board.Pending != nil || postActionWorkPending(l.Board) || len(l.PendingReceipts) > 0 || (l.Ledger.Phase != "playing" && l.Ledger.Phase != "settlement") {
		return l, errors.New("financial command boundary")
	}
	if op.Kind != "forgive" && l.Ledger.Phase == "playing" && (l.Board.Players[op.Actor-1].Departed || l.Board.Players[op.Actor-1].Confined) {
		return l, errors.New("ineligible financial actor")
	}
	n := l.Clone()
	var e error
	switch op.Kind {
	case "promise-offer":
		n.Promises, e = n.Promises.Offer(Promise{ID: op.PromiseID, Payer: op.Actor - 1, Recipient: op.Recipient - 1, AwardID: op.AwardID, Mode: op.Mode, Amount: op.Amount})
	case "promise-accept":
		n.Promises, e = n.Promises.Accept(op.PromiseID, op.Actor-1)
	case "promise-final-offer":
		n.Promises, e = n.Promises.OfferFinal(op.PromiseID, op.Actor-1, op.Amount)
	case "promise-final-answer":
		n.Promises, e = n.Promises.AnswerFinal(op.PromiseID, op.Actor-1, op.Accept)
	case "promise-refuse":
		n.Promises, e = n.Promises.Refuse(op.PromiseID, op.Actor-1)
	case "promise-pay":
		var c PromisePaymentCursor
		c, e = n.Promises.PreparePayment(op.PromiseID, op.ID, op.Actor-1, n.Ledger)
		if e != nil {
			return l, e
		}
		prior := n.Ledger.Clone()
		n.PaymentPrior = &prior
		n.PromisePayment = &c
		n.Ledger = c.Settlement.Committed.Clone()
		cursor := c.Settlement.clone()
		n.Automatic = &cursor
		n.AutomaticPurpose = "promise-payment"
	case "voluntary-transfer":
		from, to := op.Actor-1, op.Recipient-1
		if !n.Ledger.validSeat(to) || from == to || op.Amount.Sign() <= 0 || n.Ledger.Finished[from] || debtHead(n.Ledger, from) >= 0 || n.Ledger.Cash[from].Cmp(op.Amount) < 0 || slices.Contains(n.Ledger.Operations, op.ID) {
			return l, errors.New("invalid voluntary transfer")
		}
		n.Ledger.Operations = append(n.Ledger.Operations, op.ID)
		n.Ledger.Cash[from] = n.Ledger.Cash[from].Sub(op.Amount)
		n.Ledger.Scores[from] = n.Ledger.Scores[from].Sub(op.Amount)
		var c SettlementCursor
		c, e = BeginSettlement(n.Ledger, []FinancialReceipt{{Seat: to, Amount: op.Amount}})
		if e != nil {
			return l, e
		}
		n.Automatic = &c
		n.AutomaticPurpose = "transfer"
	case "forgive":
		if len(op.Seats) > 1 || len(op.Seats) == 1 && op.Seats[0] != op.Actor {
			return l, errors.New("cannot assert another player's consent")
		}
		found := false
		for _, d := range n.Ledger.Debts {
			if d.ID != op.DebtID {
				continue
			}
			found = true
			if d.Creditor >= 0 {
				if d.Creditor != op.Actor-1 {
					return l, errors.New("forgiveness requires creditor authority")
				}
				n.Ledger, e = n.Ledger.Forgive(op.ID, op.DebtID, op.Amount, []int{op.Actor - 1})
			} else {
				if op.Amount.Sign() <= 0 || op.Amount.Cmp(d.Remaining) > 0 {
					return l, errors.New("invalid forgiveness amount")
				}
				idx := -1
				for i, c := range n.ForgivenessConsents {
					if c.DebtID == op.DebtID && c.Amount.Cmp(op.Amount) == 0 {
						idx = i
					}
				}
				if idx < 0 {
					if op.Actor-1 != d.Debtor {
						return l, errors.New("only debtor may disclose system forgiveness request")
					}
					n.ForgivenessConsents = append(n.ForgivenessConsents, ForgivenessConsent{DebtID: op.DebtID, Amount: op.Amount})
					idx = len(n.ForgivenessConsents) - 1
				}
				entry := &n.ForgivenessConsents[idx]
				if !slices.Contains(entry.Seats, op.Actor) {
					entry.Seats = append(entry.Seats, op.Actor)
				}
				if len(entry.Seats) == len(n.Board.Players) {
					zero := []int{}
					for _, seat := range entry.Seats {
						zero = append(zero, seat-1)
					}
					n.Ledger, e = n.Ledger.Forgive(op.ID, op.DebtID, op.Amount, zero)
					n.ForgivenessConsents = slices.Delete(n.ForgivenessConsents, idx, idx+1)
				}
			}
			break
		}
		if !found {
			return l, errors.New("unknown forgiveness debt")
		}

	default:
		return l, ErrUnsupported
	}
	if e != nil {
		return l, e
	}
	return n, nil
}

// RecordDepartureChoice collects the simultaneous between-round boundary. A
// missing answer remains pending; only all explicit answers trigger departure.
func (l Lifecycle) RecordDepartureChoice(seat int, leave bool) (Lifecycle, error) {
	if l.Validate() != nil || !l.RoundClosed || l.DeparturesResolved || l.Automatic != nil || l.Ending != "" || l.Board.Round >= 13 || l.Board.Pending != nil || seat < 1 || seat > len(l.Board.Players) || l.Board.Players[seat-1].Departed {
		return l, errors.New("departure choice boundary")
	}
	for _, c := range l.DepartureChoices {
		if c.Seat == seat {
			return l, errors.New("departure choice already recorded")
		}
	}
	n := l.Clone()
	n.DepartureChoices = append(n.DepartureChoices, DepartureChoice{Seat: seat, Leave: leave})
	required := 0
	for _, p := range l.Board.Players {
		if !p.Departed {
			required++
		}
	}
	if len(n.DepartureChoices) != required {
		return n, nil
	}
	leaving := []int{}
	for _, c := range n.DepartureChoices {
		if c.Leave {
			leaving = append(leaving, c.Seat)
		}
	}
	n.DeparturesResolved = true
	if len(leaving) == 0 {
		return n, nil
	}
	return n.DepartSeats(leaving)
}
