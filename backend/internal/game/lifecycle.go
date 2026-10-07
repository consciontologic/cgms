package game

import (
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
)

// Lifecycle is private replayable orchestration state. It is not an observation.
// Random outcomes are explicit inputs; automatic accounting pauses durably.
type ForgivenessConsent struct {
	DebtID string `json:"debt_id"`
	Amount Amount `json:"amount"`
	Seats  []int  `json:"seats"`
}

type DepartureChoice struct {
	Seat  int  `json:"seat"`
	Leave bool `json:"leave"`
}

type Lifecycle struct {
	DepartureChoices    []DepartureChoice     `json:"departure_choices,omitempty"`
	DeparturesResolved  bool                  `json:"departures_resolved,omitempty"`
	ForgivenessConsents []ForgivenessConsent  `json:"forgiveness_consents,omitempty"`
	AutomaticPurpose    string                `json:"automatic_purpose,omitempty"`
	PromisePayment      *PromisePaymentCursor `json:"promise_payment,omitempty"`
	PaymentPrior        *FinancialLedger      `json:"payment_prior,omitempty"`
	Commands            map[string]string     `json:"commands"`
	Schema              string                `json:"schema"`
	Board               State                 `json:"board"`
	Ledger              FinancialLedger       `json:"ledger"`
	Promises            PromiseBook           `json:"promises"`
	TurnIndex           int                   `json:"turn_index"`
	TurnStarted         bool                  `json:"turn_started"`
	RoundClosed         bool                  `json:"round_closed"`
	Ending              string                `json:"ending"`
	Declarer            int                   `json:"declarer"`
	Automatic           *SettlementCursor     `json:"automatic,omitempty"`
	PendingReceipts     []FinancialReceipt    `json:"pending_receipts,omitempty"`
	EndTurnPending      bool                  `json:"end_turn_pending,omitempty"`
	CodeCursor          int                   `json:"code_cursor"`
}

func NewLifecycle(s State, ledger FinancialLedger) (Lifecycle, error) {
	l := Lifecycle{Schema: "cgms-lifecycle-v1", Commands: map[string]string{}, Board: s.Clone(), Ledger: ledger.Clone(), Promises: NewPromiseBook(s.GameID, len(s.Players))}
	if len(s.Order) > 0 {
		l.TurnIndex = slices.Index(s.Order, s.Active)
	}
	if e := l.Validate(); e != nil {
		return Lifecycle{}, e
	}
	return l, nil
}
func (l Lifecycle) Clone() Lifecycle {
	n := l
	n.DepartureChoices = slices.Clone(l.DepartureChoices)
	n.ForgivenessConsents = slices.Clone(l.ForgivenessConsents)
	for i := range n.ForgivenessConsents {
		n.ForgivenessConsents[i].Seats = slices.Clone(l.ForgivenessConsents[i].Seats)
	}
	if l.PromisePayment != nil {
		p := *l.PromisePayment
		p.Settlement = p.Settlement.clone()
		n.PromisePayment = &p
	}
	if l.PaymentPrior != nil {
		p := l.PaymentPrior.Clone()
		n.PaymentPrior = &p
	}
	n.PendingReceipts = slices.Clone(l.PendingReceipts)
	n.Commands = map[string]string{}
	for k, v := range l.Commands {
		n.Commands[k] = v
	}
	n.Board = l.Board.Clone()
	n.Ledger = l.Ledger.Clone()
	n.Promises = l.Promises.Clone()
	if l.Automatic != nil {
		c := l.Automatic.clone()
		n.Automatic = &c
	}
	return n
}
func (l Lifecycle) Validate() error {
	departureSeen := map[int]bool{}
	for _, choice := range l.DepartureChoices {
		if !l.RoundClosed || choice.Seat < 1 || choice.Seat > len(l.Board.Players) || departureSeen[choice.Seat] {
			return errors.New("invalid saved departure choice")
		}
		departureSeen[choice.Seat] = true
	}
	if l.DeparturesResolved && !l.RoundClosed {
		return errors.New("departure result outside round boundary")
	}

	consentKeys := map[string]bool{}
	for _, c := range l.ForgivenessConsents {
		key := c.DebtID + "/" + c.Amount.String()
		if c.DebtID == "" || c.Amount.Sign() <= 0 || consentKeys[key] || len(c.Seats) == 0 || len(c.Seats) >= len(l.Board.Players) {
			return errors.New("invalid saved forgiveness consent")
		}
		consentKeys[key] = true
		seen := map[int]bool{}
		for _, seat := range c.Seats {
			if seat < 1 || seat > len(l.Board.Players) || seen[seat] {
				return errors.New("invalid forgiveness consent actor")
			}
			seen[seat] = true
		}
	}

	if l.Schema != "cgms-lifecycle-v1" || l.Promises.Validate() != nil || l.Promises.GameID != l.Board.GameID || l.Promises.Seats != len(l.Board.Players) || l.Board.Validate() != nil || l.Ledger.Validate() != nil || l.Board.GameID != l.Ledger.GameID || len(l.Board.Players) != len(l.Ledger.Scores) || l.TurnIndex < 0 || l.TurnIndex >= len(l.Board.Players) {
		return errors.New("invalid lifecycle")
	}
	if l.Board.Round > 13 || l.CodeCursor < 0 || l.CodeCursor > len(l.Board.Players) {
		return errors.New("invalid lifecycle counters")
	}
	if len(l.Board.Order) > 0 && (l.TurnIndex >= len(l.Board.Order) || l.Board.Order[l.TurnIndex] != l.Board.Active || l.RoundClosed && l.TurnIndex != len(l.Board.Order)-1) {
		return errors.New("invalid frozen round cursor")
	}
	if l.Ending == "" && (l.Board.Phase != "playing" || l.Ledger.Phase != "playing" || l.Promises.Phase != "playing") {
		return errors.New("playing phase disagreement")
	}
	if l.Ending == "coup-pending" && (l.Automatic == nil || l.Ledger.Phase != "playing" || l.Promises.Phase != "playing") {
		return errors.New("pending Coup accounting disagreement")
	}
	if l.Ending != "" && l.Ending != "coup-pending" {
		if l.Automatic != nil {
			if l.AutomaticPurpose == "" && (l.Ledger.Phase != "playing" || l.Promises.Phase != "settlement") || l.AutomaticPurpose != "" && l.Ledger.Phase != l.Promises.Phase {
				return errors.New("automatic ending phase disagreement")
			}
		} else if l.Ledger.Phase != l.Promises.Phase || l.Ledger.Phase == "playing" {
			return errors.New("closed phase disagreement")
		}
	}
	for _, r := range l.PendingReceipts {
		if !l.Ledger.validSeat(r.Seat) || r.Amount.Sign() < 0 {
			return errors.New("invalid pending receipt")
		}
	}
	for id, h := range l.Commands {
		raw, e := hex.DecodeString(h)
		if id == "" || e != nil || len(raw) != 32 || hex.EncodeToString(raw) != h {
			return errors.New("invalid lifecycle command receipt")
		}
	}
	if l.AutomaticPurpose != "" && l.AutomaticPurpose != "promise-payment" && l.AutomaticPurpose != "transfer" {
		return errors.New("unknown automatic financial purpose")
	}
	if l.AutomaticPurpose != "" && l.Automatic == nil {
		return errors.New("missing automatic financial cursor")
	}
	if l.AutomaticPurpose == "promise-payment" {
		if l.PromisePayment == nil || l.PaymentPrior == nil || l.PaymentPrior.Validate() != nil || l.PromisePayment.GameID != l.Board.GameID || l.PaymentPrior.GameID != l.Board.GameID || l.PromisePayment.BookBefore != Digest(l.Promises) || l.PromisePayment.LedgerBefore != Digest(*l.PaymentPrior) {
			return errors.New("invalid promised payment envelope")
		}
	} else if l.PromisePayment != nil || l.PaymentPrior != nil {
		return errors.New("orphan payment envelope")
	}
	if l.TurnStarted && (l.RoundClosed || l.Ending != "") {
		return errors.New("invalid lifecycle turn")
	}
	if l.Ending != "" && l.Board.Phase != "settlement" {
		return errors.New("ending without closed board")
	}
	if l.Automatic != nil && (l.Automatic.Digest == "" || l.Automatic.Digest != l.Automatic.hash() || Digest(l.Automatic.Committed) != Digest(l.Ledger)) {
		return errors.New("invalid automatic settlement boundary")
	}
	return nil
}
func (l Lifecycle) BeginTurn(recycle []string) (Lifecycle, []string, error) {
	if e := l.Validate(); e != nil {
		return l, nil, e
	}
	s := l.Board
	if l.TurnStarted || l.RoundClosed || l.Automatic != nil || l.Ending != "" || s.Pending != nil || postActionWorkPending(s) || len(s.Order) == 0 || s.Order[l.TurnIndex] != s.Active || s.Players[s.Active-1].Departed {
		return l, nil, errors.New("not a scheduled turn start")
	}
	n := l.Clone()
	n.Board.Players[s.Active-1].Confined = false
	n.Board.Players[s.Active-1].OpeningUsed = false
	var drawn []string
	var e error
	n.Board, drawn, e = DrawCards(n.Board, s.Active, 1, false, recycle)
	if e != nil {
		return l, nil, e
	}
	n.TurnStarted = true
	return n, drawn, nil
}
func (l Lifecycle) EndTurn() (Lifecycle, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if !l.TurnStarted || l.Automatic != nil || l.Ending != "" || l.Board.Pending != nil || postActionWorkPending(l.Board) || len(l.PendingReceipts) > 0 {
		return l, errors.New("turn action remains pending")
	}
	n := l.Clone()
	n.TurnStarted = false
	for i := range n.Board.Proposals {
		if n.Board.Proposals[i].Status == "offered" {
			n.Board.Proposals[i].Status = "expired"
		}
	}
	if n.TurnIndex+1 < len(n.Board.Order) {
		n.TurnIndex++
		n.Board.Active = n.Board.Order[n.TurnIndex]
		n.Board.Turn++
		return n, nil
	}
	n.RoundClosed = true
	awards, e := roundReceipts(n.Board)
	if e != nil {
		return l, e
	}
	c, e := BeginSettlement(n.Ledger, awards)
	if e != nil {
		return l, e
	}
	n.Automatic = &c
	return n, nil
}

// BeginRound requires completed prior income and explicit initiative outcomes.
func (l Lifecycle) BeginRound(draws, returnOrder []string) (Lifecycle, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if !l.RoundClosed || l.Automatic != nil || l.Ending != "" || l.Board.Round >= 13 || postActionWorkPending(l.Board) || len(l.DepartureChoices) > 0 && !l.DeparturesResolved {
		return l, errors.New("round not ready")
	}
	n := l.Clone()
	n.Board.Round++
	n.Board.Turn++
	for i := range n.Board.Cards {
		if n.Board.Cards[i].AvailableFromRound <= n.Board.Round {
			n.Board.Cards[i].AvailableFromRound = 0
		}
	}
	var e error
	n.Board, e = ResolveInitiative(n.Board, draws, returnOrder)
	if e != nil {
		return l, e
	}
	n.TurnIndex = 0
	n.RoundClosed = false
	n.DepartureChoices = nil
	n.DeparturesResolved = false
	n.CodeCursor = 0
	for i := range n.Board.Cards {
		if n.Board.Cards[i].AvailableFromRound <= n.Board.Round {
			n.Board.Cards[i].AvailableFromRound = 0
		}
	}
	return n, nil
}
func (l Lifecycle) ResumeAutomatic(budget int) (Lifecycle, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if l.Automatic == nil {
		return l, errors.New("no automatic settlement")
	}
	c, done, e := ResumeSettlement(*l.Automatic, budget, true)
	if e != nil {
		return l, e
	}
	n := l.Clone()
	n.Automatic = &c
	if !done {
		return n, nil
	}
	if n.AutomaticPurpose == "promise-payment" {
		if n.PromisePayment == nil || n.PaymentPrior == nil {
			return l, errors.New("missing payment continuation")
		}
		payment := *n.PromisePayment
		payment.Settlement = c
		n.Promises, n.Ledger, e = n.Promises.CommitPayment(payment, *n.PaymentPrior)
		if e != nil {
			return l, e
		}
		n.Automatic = nil
		n.AutomaticPurpose = ""
		n.PromisePayment = nil
		n.PaymentPrior = nil
		return n, nil
	}
	n.Ledger = c.Committed
	n.Automatic = nil
	if n.AutomaticPurpose == "transfer" {
		n.AutomaticPurpose = ""
		return n, nil
	}
	if n.Ending == "coup-pending" {
		return n.beginCoupAccounting()
	}
	if n.Ending != "" {
		n.Ledger.Phase = "settlement"
		return n, nil
	}
	if n.EndTurnPending {
		n.EndTurnPending = false
		return n.EndTurn()
	}
	// Round awards settle fully before Code creates any system obligation.
	if n.RoundClosed {
		var e error
		n, e = n.nextCodeCharge()
		if e != nil {
			return l, e
		}
		if n.Automatic != nil {
			return n, nil
		}
	}
	if n.RoundClosed && n.Board.Round == 13 {
		return n.closeOrdinary("round-limit", 0)
	}
	return n, nil
}
func (l Lifecycle) DeclareOrdinary(seat int, threshold string) (Lifecycle, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if l.Automatic != nil || l.Ending != "" || l.Board.Pending != nil || postActionWorkPending(l.Board) || seat < 1 || seat > len(l.Board.Players) || l.Board.Players[seat-1].Departed || l.Board.Players[seat-1].Confined || len(l.Board.Players[seat-1].History) != 4 {
		return l, errors.New("invalid declaration boundary/history")
	}
	count := 0
	for _, c := range l.Board.Cards {
		if c.Controller == seat && (threshold == "diamonds" && c.Card.Suit == Diamonds && c.Card.Rank >= 2 && c.Card.Rank <= 10 || threshold == "royals" && c.Card.Rank >= 11) {
			count++
		}
	}
	if threshold != "diamonds" && threshold != "royals" || threshold == "diamonds" && count < 13 || threshold == "royals" && count < 15 {
		return l, errors.New("unproved threshold")
	}
	n, e := l.closeOrdinary(threshold, seat)
	if e != nil {
		return l, e
	}
	needed := 13
	if threshold == "royals" {
		needed = 15
	}
	for _, c := range l.Board.Cards {
		if needed > 0 && c.Controller == seat && (threshold == "diamonds" && c.Card.Suit == Diamonds && c.Card.Rank >= 2 && c.Card.Rank <= 10 || threshold == "royals" && c.Card.Rank >= 11) {
			if !slices.Contains(n.Board.PublicHistory, c.Card.ID) {
				n.Board.PublicHistory = append(n.Board.PublicHistory, c.Card.ID)
			}
			needed--
		}
	}
	return n, nil
}
func (l Lifecycle) closeOrdinary(reason string, declarer int) (Lifecycle, error) {
	awards := ordinaryReceipts(l.Board, declarer)
	c, e := BeginSettlement(l.Ledger, awards)
	if e != nil {
		return l, e
	}
	n := l.Clone()
	n.Board, e = CloseCardBoard(n.Board)
	if e != nil {
		return l, e
	}
	for _, a := range awards {
		if n.Board.Players[a.Seat].Departed {
			continue
		}
		amount := a.Amount
		if a.Seat+1 == declarer {
			amount = amount.Sub(IntAmount(50))
		}
		n.Promises, e = n.Promises.RecordAward(PromiseAward{ID: fmt.Sprintf("%s/ordinary-end/seat-%d", n.Board.GameID, a.Seat+1), Payer: a.Seat, Amount: amount})
		if e != nil {
			return l, e
		}
	}
	if declarer > 0 {
		n.Promises, e = n.Promises.RecordAward(PromiseAward{ID: fmt.Sprintf("%s/threshold/seat-%d", n.Board.GameID, declarer), Payer: declarer - 1, Amount: IntAmount(50)})
		if e != nil {
			return l, e
		}
	}
	n.Promises, e = n.Promises.CloseBoard()
	if e != nil {
		return l, e
	}
	n.Board = ExpireProposals(n.Board)
	n.Automatic = &c
	n.Ending = reason
	n.Declarer = declarer
	n.TurnStarted = false
	return n, nil
}
func ordinaryReceipts(s State, declarer int) []FinancialReceipt {
	awards := make([]FinancialReceipt, len(s.Players))
	for i := range awards {
		awards[i].Seat = i
		if i+1 == declarer {
			awards[i].Amount = IntAmount(50)
		}
	}
	for _, c := range s.Cards {
		if c.Controller == 0 || s.Players[c.Controller-1].Departed {
			continue
		}
		amount := 0
		if c.Card.Suit == Diamonds && c.Card.Rank >= 2 && c.Card.Rank <= 10 {
			bonus := 5
			if s.Code != nil && s.Code.Declarer != c.Controller {
				bonus = s.Code.DiamondBonus
			}
			amount = bonus + c.Card.Rank
		}
		if c.Card.Suit == Spades && c.Card.Rank >= 11 {
			amount = 10
		}
		awards[c.Controller-1].Amount = awards[c.Controller-1].Amount.Add(IntAmount(int64(amount)))
	}
	return awards
}
func roundReceipts(s State) ([]FinancialReceipt, error) {
	awards := make([]FinancialReceipt, len(s.Players))
	for i, p := range s.Players {
		awards[i].Seat = i
		if p.Departed {
			continue
		}
		for _, suit := range []Suit{Hearts, Clubs, Spades, Diamonds} {
			if !slices.Contains(s.CurrentSeries(p.Seat), suit) {
				continue
			}
			kings := []PlacedCard{}
			for _, c := range s.Cards {
				if c.Controller == p.Seat && c.Zone == Attachment && c.Card.Rank == 13 && c.Allocation == string(suit) {
					kings = append(kings, c)
				}
			}
			if len(kings) == 1 && kings[0].Card.Suit != Diamonds && kings[0].AvailableFromRound <= s.Round {
				awards[i].Amount = awards[i].Amount.Add(IntAmount(2))
			}
		}
		for _, f := range s.Formations {
			if f.Controller != p.Seat || f.Spec.Kind != "underground" || !FormationFunctioning(s, f.ID) {
				continue
			}
			kings := 0
			for _, id := range f.Spec.Cards {
				c, _ := s.Card(id)
				bound := false
				for _, b := range s.Bindings {
					if slices.Contains(b.Kings, id) {
						bound = true
					}
				}
				if c.Card.Rank == 13 && !bound {
					kings++
				}
			}
			if kings >= 5 {
				awards[i].Amount = awards[i].Amount.Add(IntAmount(int64(10 + kings - 5)))
			}
		}
	}
	return awards, nil
}

// nextCodeCharge creates one obligation only after the previous receipt queue
// drains. Reusing existing cash for repayment must not count as fresh earnings.
func (l Lifecycle) nextCodeCharge() (Lifecycle, error) {
	s := l.Board
	if s.Code == nil || !eligible(s, s.Code.Declarer) || len(s.CurrentSeries(s.Code.Declarer)) < 2 || !FormationFunctioning(s, s.Code.Formation) {
		return l, nil
	}
	valid := false
	for _, f := range s.Formations {
		if f.ID == s.Code.Formation && f.Controller == s.Code.Declarer && f.Spec.Kind == "code" {
			valid = true
		}
	}
	if !valid {
		return l, nil
	}
	n := l.Clone()
	for n.CodeCursor < len(s.Players) {
		i := n.CodeCursor
		n.CodeCursor++
		if i+1 == s.Code.Declarer || !eligible(s, i+1) || len(s.Players[i].History) < 2 {
			continue
		}
		id := fmt.Sprintf("%s/round-%d/code/%d", s.GameID, s.Round, i+1)
		amount := IntAmount(5)
		n.Ledger.Scores[i] = n.Ledger.Scores[i].Sub(amount)
		n.Ledger.NextSequence++
		n.Ledger.Charges = append(n.Ledger.Charges, FinancialCharge{Creditor: -1, ID: id, GameID: s.GameID, Debtor: i, Amount: amount})
		n.Ledger.Debts = append(n.Ledger.Debts, FinancialDebt{ID: id, ChargeID: id, GameID: s.GameID, Debtor: i, Creditor: -1, Remaining: amount, Sequence: n.Ledger.NextSequence})
		cash := n.Ledger.Cash[i]
		n.Ledger.Cash[i] = Amount{}
		n.Ledger.Scores[i] = n.Ledger.Scores[i].Sub(cash)
		c, e := BeginSettlement(n.Ledger, []FinancialReceipt{{Seat: i, Amount: cash}})
		if e != nil {
			return l, e
		}
		n.Automatic = &c
		return n, nil
	}
	return n, nil
}
