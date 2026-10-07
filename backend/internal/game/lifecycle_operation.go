package game

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/canonical"
)

type LifecycleOperation struct {
	PromiseID       string   `json:"promise_id,omitempty"`
	Recipient       int      `json:"recipient,omitempty"`
	AwardID         string   `json:"award_id,omitempty"`
	Mode            string   `json:"mode,omitempty"`
	Amount          Amount   `json:"amount"`
	Accept          bool     `json:"accept,omitempty"`
	DebtID          string   `json:"debt_id,omitempty"`
	ID              string   `json:"id"`
	GameID          string   `json:"game_id"`
	Kind            string   `json:"kind"`
	Actor           int      `json:"actor"`
	Threshold       string   `json:"threshold,omitempty"`
	Recycle         []string `json:"recycle,omitempty"`
	InitiativeDraws []string `json:"initiative_draws,omitempty"`
	ReturnOrder     []string `json:"return_order,omitempty"`
	Seats           []int    `json:"seats,omitempty"`
	Budget          int      `json:"budget,omitempty"`
}

// ApplyOperation records accepted boundary commands independently of card-command
// receipts. Retrying an identical command returns the same committed boundary.
func (l Lifecycle) ApplyOperation(op LifecycleOperation) (Lifecycle, error) {
	if l.Validate() != nil || op.ID == "" || op.GameID != l.Board.GameID {
		return l, errors.New("invalid lifecycle operation identity")
	}
	hash, e := canonical.Hash(op)
	if e != nil {
		return l, e
	}
	if old, ok := l.Commands[op.ID]; ok {
		if old != hash {
			return l, errors.New("lifecycle operation conflict")
		}
		return l.Clone(), nil
	}
	n := l
	switch op.Kind {
	case "promise-offer", "promise-accept", "promise-final-offer", "promise-final-answer", "promise-pay", "promise-refuse", "voluntary-transfer", "forgive":
		n, e = l.applyFinancialOperation(op)
	case "begin-turn":
		if op.Actor != l.Board.Active {
			return l, errors.New("wrong scheduled actor")
		}
		n, _, e = l.BeginTurn(op.Recycle)
	case "end-turn":
		if op.Actor != l.Board.Active {
			return l, errors.New("wrong turn actor")
		}
		n, e = l.EndTurn()
	case "begin-round":
		if !l.DeparturesResolved {
			return l, errors.New("departure choices unresolved")
		}
		n, e = l.BeginRound(op.InitiativeDraws, op.ReturnOrder)
	case "declare-ordinary":
		n, e = l.DeclareOrdinary(op.Actor, op.Threshold)
	case "coup":
		n, e = l.DeclareCoup(Command{GameID: op.GameID, ID: op.ID, Kind: "coup", Actor: op.Actor})
	case "resume-automatic":
		n, e = l.ResumeAutomatic(op.Budget)
	case "departure-choice":
		if len(op.Seats) != 0 {
			return l, errors.New("departure choice must name only authenticated actor")
		}
		n, e = l.RecordDepartureChoice(op.Actor, op.Accept)
	case "depart":
		return l, errors.New("collect explicit departure-choice operations")
	case "finish-settlement":
		n, e = l.FinishSettlement(op.Actor)
	case "finalize":
		n, e = l.Finalize()
	default:
		return l, ErrUnsupported
	}
	if e != nil {
		return l, e
	}
	n = n.Clone()
	n.Commands[op.ID] = hash
	if e = n.Validate(); e != nil {
		return l, e
	}
	return n, nil
}
