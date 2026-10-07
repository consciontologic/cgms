package game

import (
	"errors"
	"fmt"
	"slices"
)

type FinancialDebt struct {
	ID        string `json:"id"`
	ChargeID  string `json:"charge_id"`
	GameID    string `json:"game_id"`
	Debtor    int    `json:"debtor"`
	Creditor  int    `json:"creditor"`
	Remaining Amount `json:"remaining"`
	Sequence  uint64 `json:"sequence"`
}
type FinancialCharge struct {
	Creditor int    `json:"creditor"`
	ID       string `json:"id"`
	GameID   string `json:"game_id"`
	Debtor   int    `json:"debtor"`
	Amount   Amount `json:"amount"`
}
type FinancialCorrection struct {
	OperationID string `json:"operation_id"`
	DebtID      string `json:"debt_id"`
	ChargeID    string `json:"charge_id"`
	GameID      string `json:"game_id"`
	Debtor      int    `json:"debtor"`
	Amount      Amount `json:"amount"`
	MatchLevel  bool   `json:"match_level"`
}
type FinancialResult struct {
	GameID string   `json:"game_id"`
	Scores []Amount `json:"scores"`
}
type FinancialLedger struct {
	GameID       string                `json:"game_id"`
	Phase        string                `json:"phase"`
	Scores       []Amount              `json:"scores"`
	Cash         []Amount              `json:"cash"`
	Debts        []FinancialDebt       `json:"debts"`
	Charges      []FinancialCharge     `json:"charges"`
	Corrections  []FinancialCorrection `json:"corrections"`
	Completed    []FinancialResult     `json:"completed"`
	Departed     []bool                `json:"departed"`
	Finished     []bool                `json:"finished"`
	Operations   []string              `json:"operations"`
	NextSequence uint64                `json:"next_sequence"`
}

func NewFinancialLedger(game string, seats int) FinancialLedger {
	return FinancialLedger{GameID: game, Scores: make([]Amount, seats), Cash: make([]Amount, seats), Departed: make([]bool, seats), Finished: make([]bool, seats), Phase: "playing"}
}
func (l FinancialLedger) Clone() FinancialLedger {
	n := l
	n.Scores = slices.Clone(l.Scores)
	n.Cash = slices.Clone(l.Cash)
	n.Debts = slices.Clone(l.Debts)
	n.Charges = slices.Clone(l.Charges)
	n.Corrections = slices.Clone(l.Corrections)
	n.Completed = slices.Clone(l.Completed)
	for i := range n.Completed {
		n.Completed[i].Scores = slices.Clone(n.Completed[i].Scores)
	}
	n.Departed = slices.Clone(l.Departed)
	n.Finished = slices.Clone(l.Finished)
	n.Operations = slices.Clone(l.Operations)
	return n
}
func (l FinancialLedger) validSeat(s int) bool { return s >= 0 && s < len(l.Scores) }
func (l FinancialLedger) Charge(id string, debtor, creditor int, a Amount) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if id == "" || !l.validSeat(debtor) || (creditor != -1 && !l.validSeat(creditor)) || creditor == debtor || a.Sign() <= 0 || l.Phase != "playing" {
		return l, errors.New("invalid charge")
	}
	for _, c := range l.Charges {
		if c.ID == id {
			if c.Debtor == debtor && c.Creditor == creditor && c.GameID == l.GameID && c.Amount.Cmp(a) == 0 {
				return l.Clone(), nil
			}
			return l, errors.New("charge ID conflict")
		}
	}
	n := l.Clone()
	n.Scores[debtor] = n.Scores[debtor].Sub(a)
	n.NextSequence++
	n.Charges = append(n.Charges, FinancialCharge{creditor, id, n.GameID, debtor, a})
	n.Debts = append(n.Debts, FinancialDebt{id, id, n.GameID, debtor, creditor, a, n.NextSequence})
	if n.Cash[debtor].Sign() > 0 {
		cash := n.Cash[debtor]
		n.Cash[debtor] = Amount{}
		n.Scores[debtor] = n.Scores[debtor].Sub(cash)
		settled, e := SettleReceipts(n, []FinancialReceipt{{debtor, cash}})
		if e != nil {
			return l, e
		}
		return settled, nil
	}
	return n, nil
}

// Validate checks structural and provenance integrity before a financial snapshot
// crosses the trusted engine boundary. Historical score values themselves require
// replay verification; shape validation never substitutes for an authenticated trace.
func (l FinancialLedger) Validate() error {
	if l.GameID == "" || (len(l.Scores) != 3 && len(l.Scores) != 4) || len(l.Cash) != len(l.Scores) || len(l.Departed) != len(l.Scores) || len(l.Finished) != len(l.Scores) {
		return errors.New("invalid financial dimensions")
	}
	if l.Phase != "playing" && l.Phase != "settlement" && l.Phase != "finalized" && l.Phase != "void" {
		return errors.New("invalid financial phase")
	}
	gameOrder := map[string]int{}
	for i, r := range l.Completed {
		if r.GameID == "" || len(r.Scores) != len(l.Scores) {
			return errors.New("invalid completed financial result")
		}
		if _, ok := gameOrder[r.GameID]; ok {
			return errors.New("duplicate completed game")
		}
		gameOrder[r.GameID] = i
		if r.GameID == l.GameID && (l.Phase != "finalized" || i != len(l.Completed)-1 || !slices.Equal(r.Scores, l.Scores)) {
			return errors.New("current result conflicts with completed history")
		}
	}
	if l.Phase == "finalized" {
		if _, ok := gameOrder[l.GameID]; !ok {
			return errors.New("missing finalized result")
		}
	} else {
		gameOrder[l.GameID] = len(l.Completed)
	}
	charges := map[string]FinancialCharge{}
	for _, c := range l.Charges {
		if c.ID == "" || c.GameID == "" || !l.validSeat(c.Debtor) || (c.Creditor != -1 && !l.validSeat(c.Creditor)) || c.Creditor == c.Debtor || c.Amount.Sign() <= 0 {
			return errors.New("invalid charge provenance")
		}
		if _, ok := gameOrder[c.GameID]; !ok {
			return errors.New("unknown charge origin game")
		}
		if _, ok := charges[c.ID]; ok {
			return errors.New("duplicate charge")
		}
		charges[c.ID] = c
	}
	debts := map[string]FinancialDebt{}
	usedCharge := map[string]bool{}
	var lastSequence uint64
	for _, d := range l.Debts {
		if d.ID == "" || d.Sequence <= lastSequence || !l.validSeat(d.Debtor) || (d.Creditor != -1 && !l.validSeat(d.Creditor)) || d.Debtor == d.Creditor || d.Remaining.Sign() < 0 {
			return errors.New("invalid debt")
		}
		if _, ok := debts[d.ID]; ok {
			return errors.New("duplicate debt")
		}
		if usedCharge[d.ChargeID] {
			return errors.New("multiple debts for one charge")
		}
		c, ok := charges[d.ChargeID]
		if !ok || c.GameID != d.GameID || c.Debtor != d.Debtor || c.Creditor != d.Creditor || c.Amount.Cmp(d.Remaining) < 0 {
			return errors.New("broken debt charge lineage")
		}
		usedCharge[d.ChargeID] = true
		debts[d.ID] = d
		lastSequence = d.Sequence
	}
	if len(usedCharge) != len(charges) || lastSequence != l.NextSequence {
		return errors.New("incomplete charge/debt history")
	}
	operations := map[string]bool{}
	for _, id := range l.Operations {
		if id == "" || operations[id] {
			return errors.New("invalid operation identity")
		}
		operations[id] = true
	}
	corrected := map[string]Amount{}
	correctionIDs := map[string]bool{}
	for _, c := range l.Corrections {
		d, ok := debts[c.DebtID]
		if !ok || c.ChargeID != d.ChargeID || !l.validSeat(c.Debtor) || c.Debtor != d.Debtor || c.Amount.Sign() <= 0 || c.OperationID == "" || correctionIDs[c.OperationID] || !operations[c.OperationID] {
			return errors.New("invalid correction provenance")
		}
		creating, ok := gameOrder[c.GameID]
		if !ok {
			return errors.New("unknown correction creating game")
		}
		origin := gameOrder[d.GameID]
		if c.MatchLevel != (c.GameID != d.GameID) || creating < origin {
			return errors.New("invalid correction scope or chronology")
		}
		correctionIDs[c.OperationID] = true
		corrected[d.ID] = corrected[d.ID].Add(c.Amount)
		if corrected[d.ID].Add(d.Remaining).Cmp(charges[d.ChargeID].Amount) > 0 {
			return errors.New("forgiveness exceeds original unpaid charge")
		}
	}
	for _, c := range l.Cash {
		if c.Sign() < 0 || (l.Phase == "finalized" && c.Sign() != 0) {
			return errors.New("invalid available cash")
		}
	}
	return nil
}

// Coup implements only the scenario-level financial boundary; card eligibility is
// adjudicated by the game transition layer before calling this method.
func (l FinancialLedger) Coup(owner int) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if !l.validSeat(owner) || l.Departed[owner] || l.Phase != "playing" {
		return l, errors.New("invalid Coup financial boundary")
	}
	n := l.Clone()
	for seat := range n.Scores {
		if n.Departed[seat] {
			continue
		}
		n.Scores[seat] = Amount{}
		n.Cash[seat] = Amount{}
		for _, d := range n.Debts {
			if d.Debtor == seat && d.GameID == n.GameID {
				n.Scores[seat] = n.Scores[seat].Sub(d.Remaining)
			}
		}
	}
	// Current-game corrections were in the replaced result. Cross-game corrections
	// remain separate immutable match adjustments.
	keep := n.Corrections[:0]
	for _, c := range n.Corrections {
		if c.MatchLevel || n.Departed[c.Debtor] {
			keep = append(keep, c)
		}
	}
	n.Corrections = keep
	for seat := range n.Scores {
		if seat != owner && !n.Departed[seat] {
			var e error
			n, e = n.Charge(fmt.Sprintf("%s/coup/%d", n.GameID, seat), seat, -1, IntAmount(50))
			if e != nil {
				return l, e
			}
		}
	}
	n, e := SettleReceipts(n, []FinancialReceipt{{owner, IntAmount(50)}})
	if e != nil {
		return l, e
	}
	n.Phase = "settlement"
	return n, nil
}
func (l FinancialLedger) Forgive(operation, debt string, a Amount, consent []int) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if operation == "" || a.Sign() <= 0 || (l.Phase != "playing" && l.Phase != "settlement") {
		return l, errors.New("invalid forgiveness")
	}
	for _, c := range l.Corrections {
		if c.OperationID == operation {
			if c.DebtID == debt && c.Amount.Cmp(a) == 0 {
				return l.Clone(), nil
			}
			return l, errors.New("forgiveness ID conflict")
		}
	}
	if slices.Contains(l.Operations, operation) {
		return l, errors.New("operation already used")
	}
	i := -1
	for k, d := range l.Debts {
		if d.ID == debt {
			i = k
		}
	}
	if i < 0 || l.Debts[i].Remaining.Cmp(a) < 0 {
		return l, errors.New("invalid forgiven amount")
	}
	d := l.Debts[i]
	if d.Creditor >= 0 {
		if !slices.Contains(consent, d.Creditor) {
			return l, errors.New("creditor consent required")
		}
	} else {
		for seat := range l.Scores {
			if !slices.Contains(consent, seat) {
				return l, errors.New("unanimous consent required")
			}
		}
	}
	n := l.Clone()
	n.Debts[i].Remaining = n.Debts[i].Remaining.Sub(a)
	cross := d.GameID != n.GameID
	n.Corrections = append(n.Corrections, FinancialCorrection{operation, debt, d.ChargeID, n.GameID, d.Debtor, a, cross})
	n.Operations = append(n.Operations, operation)
	if !cross {
		n.Scores[d.Debtor] = n.Scores[d.Debtor].Add(a)
	}
	return n, nil
}
func (l FinancialLedger) FinishSettlement(seat int) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if l.Phase != "settlement" || !l.validSeat(seat) {
		return l, errors.New("invalid settlement finish")
	}
	n := l.Clone()
	n.Finished[seat] = true
	return n, nil
}
func (l FinancialLedger) Finalize() (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if l.Phase != "settlement" {
		return l, errors.New("board not closed")
	}
	for _, f := range l.Finished {
		if !f {
			return l, errors.New("explicit settlement finish required")
		}
	}
	n := l.Clone()
	n.Completed = append(n.Completed, FinancialResult{n.GameID, slices.Clone(n.Scores)})
	n.Cash = make([]Amount, len(n.Cash))
	n.Phase = "finalized"
	return n, nil
}
func (l FinancialLedger) NextGame(id string) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if (l.Phase != "finalized" && l.Phase != "void") || id == "" || id == l.GameID {
		return l, errors.New("invalid next-game admission")
	}
	for _, r := range l.Completed {
		if r.GameID == id {
			return l, errors.New("reused game identity")
		}
	}
	n := l.Clone()
	n.GameID = id
	n.Phase = "playing"
	n.Cash = make([]Amount, len(n.Cash))
	n.Scores = make([]Amount, len(n.Scores))
	n.Departed = make([]bool, len(n.Departed))
	n.Finished = make([]bool, len(n.Finished))
	return n, nil
}
func (l FinancialLedger) MatchScores() []Amount {
	if l.Validate() != nil {
		return nil
	}
	v := make([]Amount, len(l.Scores))
	for _, r := range l.Completed {
		for i, a := range r.Scores {
			v[i] = v[i].Add(a)
		}
	}
	if l.Phase != "finalized" && l.Phase != "void" {
		for i, a := range l.Scores {
			v[i] = v[i].Add(a)
		}
	}
	for _, c := range l.Corrections {
		if c.MatchLevel {
			v[c.Debtor] = v[c.Debtor].Add(c.Amount)
		}
	}
	return v
}
func (l FinancialLedger) CloseOrdinary(awards []FinancialReceipt) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if l.Phase != "playing" {
		return l, errors.New("board already closed")
	}
	for _, r := range awards {
		if !l.validSeat(r.Seat) || l.Departed[r.Seat] {
			return l, errors.New("departed award recipient")
		}
	}
	n, e := SettleReceipts(l, awards)
	if e != nil {
		return l, e
	}
	n.Phase = "settlement"
	return n, nil
}
func (l FinancialLedger) Depart(seat int) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if l.Phase != "playing" || !l.validSeat(seat) || l.Departed[seat] {
		return l, errors.New("invalid departure")
	}
	n := l.Clone()
	n.Departed[seat] = true
	n.Cash[seat] = Amount{}
	if n.Scores[seat].Sign() > 0 {
		n.Scores[seat] = Amount{}
	}
	return n, nil
}

// Nullify takes the retained start-of-game financial snapshot. It never rewrites
// a completed result; a replacement must receive a fresh game-instance identity.
func (l FinancialLedger) Nullify(start FinancialLedger, consent []int) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if start.Validate() != nil || len(start.Scores) != len(l.Scores) || l.GameID != start.GameID || l.Phase == "finalized" || start.Phase != "playing" {
		return l, errors.New("invalid nullification snapshot")
	}
	for seat := range l.Scores {
		if !slices.Contains(consent, seat) {
			return l, errors.New("unanimous consent required")
		}
	}
	n := start.Clone()
	n.Phase = "void"
	return n, nil
}
func (l FinancialLedger) Transfer(id string, from, to int, a Amount) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if id == "" || !l.validSeat(from) || !l.validSeat(to) || from == to || a.Sign() <= 0 || l.Cash[from].Cmp(a) < 0 || l.Finished[from] || (l.Phase != "playing" && l.Phase != "settlement") || debtHead(l, from) >= 0 {
		return l, errors.New("invalid voluntary transfer")
	}
	if slices.Contains(l.Operations, id) {
		return l, errors.New("duplicate transfer")
	}
	n := l.Clone()
	n.Operations = append(n.Operations, id)
	n.Cash[from] = n.Cash[from].Sub(a)
	n.Scores[from] = n.Scores[from].Sub(a)
	settled, e := SettleReceipts(n, []FinancialReceipt{{to, a}})
	if e != nil {
		return l, e
	}
	return settled, nil
}
func (l FinancialLedger) Ponzi(target, user, partner int) (FinancialLedger, error) {
	if err := l.Validate(); err != nil {
		return l, err
	}

	if l.Phase != "playing" || !l.validSeat(target) || !l.validSeat(user) || target == user || (partner != -1 && (!l.validSeat(partner) || partner == user || partner == target)) {
		return l, errors.New("invalid Ponzi pair")
	}
	n := l.Clone()
	a := n.Cash[target]
	n.Cash[target] = Amount{}
	n.Scores[target] = n.Scores[target].Sub(a)
	r := []FinancialReceipt{{user, a}}
	if partner >= 0 {
		half, _ := a.Quo(IntAmount(2))
		r = []FinancialReceipt{{user, half}, {partner, half}}
	}
	settled, e := SettleReceipts(n, r)
	if e != nil {
		return l, e
	}
	return settled, nil
}
