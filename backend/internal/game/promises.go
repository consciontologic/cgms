package game

import (
	"errors"
	"slices"
)

// Promise seats use the financial ledger's zero-based seat indices. All methods
// return detached state. Named award IDs are supplied by adjudication, never bots.
type Promise struct {
	ID        string `json:"id"`
	Payer     int    `json:"payer"`
	Recipient int    `json:"recipient"`
	AwardID   string `json:"award_id"`
	Mode      string `json:"mode"` // fixed or percentage; percentage is a fraction in [0,1].
	Amount    Amount `json:"amount"`
	Status    string `json:"status"`
	Triggered bool   `json:"triggered"`
	Promised  Amount `json:"promised"`
	Paid      Amount `json:"paid"`
	Final     Amount `json:"final"`
}
type PromiseAward struct {
	ID     string `json:"id"`
	Payer  int    `json:"payer"`
	Amount Amount `json:"amount"`
}
type PromisePayment struct {
	ID          string `json:"id"`
	PromiseID   string `json:"promise_id"`
	Amount      Amount `json:"amount"`
	LedgerAfter string `json:"ledger_after"`
}
type PromiseBook struct {
	GameID   string           `json:"game_id"`
	Seats    int              `json:"seats"`
	Phase    string           `json:"phase"`
	Promises []Promise        `json:"promises"`
	Awards   []PromiseAward   `json:"awards"`
	Payments []PromisePayment `json:"payments"`
	Finished []bool           `json:"finished"`
}
type PromiseOutcome struct {
	Payer     int    `json:"payer"`
	Recipient int    `json:"recipient"`
	Outcome   string `json:"outcome"`
	Promised  Amount `json:"promised"`
	Delivered Amount `json:"delivered"`
}

func NewPromiseBook(gameID string, seats int) PromiseBook {
	return PromiseBook{GameID: gameID, Seats: seats, Phase: "playing", Finished: make([]bool, seats)}
}
func (b PromiseBook) Clone() PromiseBook {
	b.Promises = slices.Clone(b.Promises)
	b.Awards = slices.Clone(b.Awards)
	b.Payments = slices.Clone(b.Payments)
	b.Finished = slices.Clone(b.Finished)
	return b
}
func (b PromiseBook) seat(s int) bool { return s >= 0 && s < b.Seats }
func (b PromiseBook) index(id string) int {
	for i, p := range b.Promises {
		if p.ID == id {
			return i
		}
	}
	return -1
}
func promiseClassification(promised, paid Amount, refused bool) string {
	if promised.Sign() <= 0 {
		return "none"
	}
	if refused {
		return "scam"
	}
	if paid.Cmp(promised) >= 0 {
		return "trust"
	}
	if paid.Mul(IntAmount(5)).Cmp(promised.Mul(IntAmount(2))) >= 0 {
		return "anyhoo"
	}
	return "none"
}
func pendingPromise(p Promise) bool {
	return p.Triggered && p.Promised.Sign() > 0 && (p.Status == "accepted" || p.Status == "final-offered" || p.Status == "final-accepted")
}
func (b PromiseBook) Validate() error {
	if b.GameID == "" || (b.Seats != 3 && b.Seats != 4) || len(b.Finished) != b.Seats || (b.Phase != "playing" && b.Phase != "settlement" && b.Phase != "finalized" && b.Phase != "void") {
		return errors.New("invalid promise book")
	}
	awards := map[string]PromiseAward{}
	for _, a := range b.Awards {
		if a.ID == "" || !b.seat(a.Payer) || a.Amount.Sign() < 0 {
			return errors.New("invalid award")
		}
		if _, ok := awards[a.ID]; ok {
			return errors.New("duplicate award")
		}
		awards[a.ID] = a
	}
	ids := map[string]bool{}
	payments := map[string]Amount{}
	ops := map[string]bool{}
	for _, p := range b.Payments {
		if p.ID == "" || ops[p.ID] || p.Amount.Sign() < 0 || p.LedgerAfter == "" {
			return errors.New("invalid promise payment")
		}
		ops[p.ID] = true
		payments[p.PromiseID] = payments[p.PromiseID].Add(p.Amount)
	}
	for _, p := range b.Promises {
		if p.ID == "" || ids[p.ID] || !b.seat(p.Payer) || !b.seat(p.Recipient) || p.Payer == p.Recipient || p.AwardID == "" || p.Amount.Sign() < 0 || (p.Mode != "fixed" && p.Mode != "percentage") || (p.Mode == "percentage" && p.Amount.Cmp(IntAmount(1)) > 0) {
			return errors.New("invalid promise terms")
		}
		ids[p.ID] = true
		if !slices.Contains([]string{"offered", "accepted", "final-offered", "final-accepted", "paid", "refused", "rejected", "untriggered"}, p.Status) || p.Paid.Sign() < 0 || p.Promised.Sign() < 0 || p.Paid.Cmp(p.Promised) > 0 || p.Paid.Cmp(payments[p.ID]) != 0 {
			return errors.New("invalid promise status or payment")
		}
		if p.Triggered {
			a, ok := awards[p.AwardID]
			if !ok || a.Payer != p.Payer || p.Status == "offered" || p.Status == "untriggered" {
				return errors.New("invalid award condition")
			}
			want := p.Amount
			if p.Mode == "percentage" {
				want = want.Mul(a.Amount)
			}
			if want.Cmp(p.Promised) != 0 {
				return errors.New("promise amount mismatch")
			}
		} else if p.Promised.Sign() != 0 || p.Paid.Sign() != 0 || (p.Status != "offered" && p.Status != "accepted" && p.Status != "untriggered") {
			return errors.New("untriggered promise mutation")
		}
		if (p.Status == "final-offered" || p.Status == "final-accepted") && (p.Final.Sign() < 0 || p.Final.Cmp(p.Promised) >= 0 || p.Paid.Sign() != 0) {
			return errors.New("invalid reduced final offer")
		}
		if p.Status == "paid" && p.Paid.Cmp(p.Promised) != 0 && p.Paid.Cmp(p.Final) != 0 {
			return errors.New("unfulfilled paid promise")
		}
		if pendingPromise(p) && (b.Finished[p.Payer] || b.Finished[p.Recipient]) {
			return errors.New("finished pending participant")
		}
	}
	for id := range payments {
		if !ids[id] {
			return errors.New("payment without promise")
		}
	}
	for _, f := range b.Finished {
		if b.Phase == "playing" && f || b.Phase == "finalized" && !f {
			return errors.New("invalid finish phase")
		}
	}
	return nil
}
func (b PromiseBook) Offer(p Promise) (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	if b.Phase != "playing" || p.ID == "" || b.index(p.ID) >= 0 || !b.seat(p.Payer) || !b.seat(p.Recipient) {
		return b, errors.New("invalid promise offer")
	}
	for _, a := range b.Awards {
		if a.ID == p.AwardID {
			return b, errors.New("promise after award")
		}
	}
	p.Status = "offered"
	p.Triggered = false
	p.Promised = Amount{}
	p.Paid = Amount{}
	p.Final = Amount{}
	n := b.Clone()
	n.Promises = append(n.Promises, p)
	if e := n.Validate(); e != nil {
		return b, e
	}
	return n, nil
}
func (b PromiseBook) Accept(id string, actor int) (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	i := b.index(id)
	if i < 0 || b.Phase != "playing" || b.Promises[i].Recipient != actor || b.Promises[i].Status != "offered" {
		return b, errors.New("invalid promise acceptance")
	}
	for _, a := range b.Awards {
		if a.ID == b.Promises[i].AwardID {
			return b, errors.New("acceptance after award")
		}
	}
	n := b.Clone()
	n.Promises[i].Status = "accepted"
	return n, nil
}
func (b PromiseBook) RecordAward(a PromiseAward) (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	if b.Phase != "playing" || a.ID == "" || !b.seat(a.Payer) || a.Amount.Sign() < 0 {
		return b, errors.New("invalid award boundary")
	}
	for _, old := range b.Awards {
		if old.ID == a.ID {
			if old == a {
				return b.Clone(), nil
			}
			return b, errors.New("award conflict")
		}
	}
	n := b.Clone()
	n.Awards = append(n.Awards, a)
	for i, p := range n.Promises {
		if p.Status == "accepted" && p.AwardID == a.ID && p.Payer == a.Payer {
			p.Triggered = true
			p.Promised = p.Amount
			if p.Mode == "percentage" {
				p.Promised = p.Amount.Mul(a.Amount)
			}
			if p.Promised.IsZero() {
				p.Status = "paid"
			}
			n.Promises[i] = p
		}
	}
	return n, nil
}
func (b PromiseBook) OfferFinal(id string, actor int, amount Amount) (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	i := b.index(id)
	if i < 0 || (b.Phase != "playing" && b.Phase != "settlement") || !b.seat(actor) || b.Finished[actor] {
		return b, errors.New("invalid final offer boundary")
	}
	p := b.Promises[i]
	if p.Payer != actor || p.Status != "accepted" || !p.Triggered || amount.Sign() < 0 || amount.Cmp(p.Promised) >= 0 {
		return b, errors.New("invalid final offer")
	}
	n := b.Clone()
	n.Promises[i].Final = amount
	n.Promises[i].Status = "final-offered"
	return n, nil
}
func (b PromiseBook) AnswerFinal(id string, actor int, accept bool) (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	i := b.index(id)
	if i < 0 || b.Promises[i].Recipient != actor || b.Promises[i].Status != "final-offered" || (b.Phase != "playing" && b.Phase != "settlement") {
		return b, errors.New("invalid final answer")
	}
	n := b.Clone()
	n.Promises[i].Status = "rejected"
	if accept {
		n.Promises[i].Status = "final-accepted"
	}
	return n, nil
}
func (b PromiseBook) Refuse(id string, actor int) (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	i := b.index(id)
	if i < 0 || b.Promises[i].Payer != actor || !pendingPromise(b.Promises[i]) || b.Promises[i].Status == "final-offered" || (b.Phase != "playing" && b.Phase != "settlement") {
		return b, errors.New("invalid refusal")
	}
	n := b.Clone()
	n.Promises[i].Status = "refused"
	return n, nil
}

// PreparePayment debits the payer and returns a resumable recipient receipt.
// CommitPayment alone records delivered gross value after the FIFO cursor finishes.
type PromisePaymentCursor struct {
	GameID       string           `json:"game_id"`
	PromiseID    string           `json:"promise_id"`
	Operation    string           `json:"operation"`
	Amount       Amount           `json:"amount"`
	BookBefore   string           `json:"book_before"`
	LedgerBefore string           `json:"ledger_before"`
	Settlement   SettlementCursor `json:"settlement"`
}

func (b PromiseBook) PreparePayment(id, operation string, actor int, l FinancialLedger) (PromisePaymentCursor, error) {
	fail := func() (PromisePaymentCursor, error) {
		return PromisePaymentCursor{}, errors.New("invalid promise payment")
	}
	if e := b.Validate(); e != nil {
		return PromisePaymentCursor{}, e
	}
	if e := l.Validate(); e != nil {
		return PromisePaymentCursor{}, e
	}
	i := b.index(id)
	if i < 0 || operation == "" || !b.seat(actor) || b.Finished[actor] || l.GameID != b.GameID || l.Phase != b.Phase || (b.Phase != "playing" && b.Phase != "settlement") || slices.Contains(l.Operations, operation) {
		return fail()
	}
	p := b.Promises[i]
	if p.Payer != actor || !p.Triggered || (p.Status != "accepted" && p.Status != "final-accepted") || l.Finished[actor] || debtHead(l, actor) >= 0 {
		return fail()
	}
	a := p.Promised
	if p.Status == "final-accepted" {
		a = p.Final
	}
	if l.Cash[actor].Cmp(a) < 0 {
		return fail()
	}
	n := l.Clone()
	n.Operations = append(n.Operations, operation)
	n.Cash[actor] = n.Cash[actor].Sub(a)
	n.Scores[actor] = n.Scores[actor].Sub(a)
	receipts := []FinancialReceipt{}
	if a.Sign() > 0 {
		receipts = append(receipts, FinancialReceipt{p.Recipient, a})
	}
	c, e := BeginSettlement(n, receipts)
	if e != nil {
		return PromisePaymentCursor{}, e
	}
	return PromisePaymentCursor{b.GameID, id, operation, a, Digest(b), Digest(l), c}, nil
}
func (b PromiseBook) CommitPayment(c PromisePaymentCursor, l FinancialLedger) (PromiseBook, FinancialLedger, error) {
	if c.GameID != b.GameID || c.BookBefore != Digest(b) || c.LedgerBefore != Digest(l) || !c.Settlement.Done {
		return b, l, errors.New("stale or unfinished promise payment")
	}
	// Reconstruct the admitted cursor to reject a forged payment envelope. Resume's
	// digest validation checks persisted progress; replay authenticates its history.
	i0 := b.index(c.PromiseID)
	if i0 < 0 {
		return b, l, errors.New("unknown payment promise")
	}
	expected, e := b.PreparePayment(c.PromiseID, c.Operation, b.Promises[i0].Payer, l)
	if e != nil {
		return b, l, e
	}
	if c.Amount.Cmp(expected.Amount) != 0 {
		return b, l, errors.New("payment continuation mismatch")
	}
	// A recomputable hash is integrity metadata, not authority. Re-adjudicate
	// each compressed/reference segment from the admitted input; never expand
	// cycle repetitions into individual transfers.
	verified := expected.Settlement
	for _, segment := range c.Settlement.Audit {
		if segment.Algorithm != "reference" && segment.Algorithm != "simple-cycle-v1" {
			return b, l, errors.New("unknown payment audit algorithm")
		}
		var stepErr error
		verified, _, stepErr = ResumeSettlement(verified, 1, segment.Algorithm == "simple-cycle-v1")
		if stepErr != nil || len(verified.Audit) == 0 || Digest(verified.Audit[len(verified.Audit)-1]) != Digest(segment) {
			return b, l, errors.New("payment audit transition mismatch")
		}
	}
	if !verified.Done || Digest(verified) != Digest(c.Settlement) {
		return b, l, errors.New("payment final settlement mismatch")
	}
	n := b.Clone()
	i := n.index(c.PromiseID)
	n.Promises[i].Paid = c.Amount
	n.Promises[i].Status = "paid"
	n.Payments = append(n.Payments, PromisePayment{c.Operation, c.PromiseID, c.Amount, Digest(verified.Committed)})
	return n, verified.Committed, nil
}
func (b PromiseBook) Pay(id, operation string, actor int, l FinancialLedger) (PromiseBook, FinancialLedger, error) {
	for _, p := range b.Payments {
		if p.ID == operation {
			i := b.index(id)
			if p.PromiseID == id && i >= 0 && b.Promises[i].Payer == actor && l.GameID == b.GameID && slices.Contains(l.Operations, operation) {
				return b.Clone(), l.Clone(), nil
			}
			return b, l, errors.New("payment operation conflict")
		}
	}
	c, e := b.PreparePayment(id, operation, actor, l)
	if e != nil {
		return b, l, e
	}
	c.Settlement, _, e = ResumeSettlement(c.Settlement, 256, true)
	if e != nil {
		return b, l, e
	}
	if !c.Settlement.Done {
		return b, l, ErrSettlementRequiresContinuation
	}
	return b.CommitPayment(c, l)
}
func (b PromiseBook) CloseBoard() (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	if b.Phase != "playing" {
		return b, errors.New("promise board already closed")
	}
	n := b.Clone()
	n.Phase = "settlement"
	for i, p := range n.Promises {
		if !p.Triggered {
			n.Promises[i].Status = "untriggered"
		}
	}
	return n, nil
}
func (b PromiseBook) Finish(actor int) (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	if b.Phase != "settlement" || !b.seat(actor) {
		return b, errors.New("invalid promise finish")
	}
	for _, p := range b.Promises {
		if pendingPromise(p) && (p.Payer == actor || p.Recipient == actor) {
			return b, errors.New("pending promise or final offer")
		}
	}
	n := b.Clone()
	n.Finished[actor] = true
	return n, nil
}
func (b PromiseBook) Finalize() (PromiseBook, error) {
	if e := b.Validate(); e != nil {
		return b, e
	}
	if b.Phase != "settlement" {
		return b, errors.New("invalid finalization")
	}
	for _, f := range b.Finished {
		if !f {
			return b, errors.New("finish required")
		}
	}
	n := b.Clone()
	n.Phase = "finalized"
	return n, nil
}
func (b PromiseBook) Nullify() PromiseBook { n := b.Clone(); n.Phase = "void"; return n }
func (b PromiseBook) Outcomes() ([]PromiseOutcome, error) {
	if e := b.Validate(); e != nil {
		return nil, e
	}
	if b.Phase == "void" {
		return []PromiseOutcome{}, nil
	}
	if b.Phase == "playing" {
		return nil, errors.New("board still playing")
	}
	out := []PromiseOutcome{}
	for payer := 0; payer < b.Seats; payer++ {
		for recipient := 0; recipient < b.Seats; recipient++ {
			if payer == recipient {
				continue
			}
			a, paid := Amount{}, Amount{}
			refused, pending := false, false
			for _, p := range b.Promises {
				if p.Payer == payer && p.Recipient == recipient && p.Triggered {
					a = a.Add(p.Promised)
					paid = paid.Add(p.Paid)
					refused = refused || p.Status == "refused" || p.Status == "rejected"
					pending = pending || pendingPromise(p)
				}
			}
			label := promiseClassification(a, paid, refused)
			if pending {
				label = "pending"
			}
			if a.Sign() > 0 {
				out = append(out, PromiseOutcome{payer, recipient, label, a, paid})
			}
		}
	}
	return out, nil
}
