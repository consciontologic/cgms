package game

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"math/big"
	"slices"
	"sort"
)

const SettlementVersion = "cgms-fifo-v1"

type FinancialReceipt struct {
	Seat   int    `json:"seat"`
	Amount Amount `json:"amount"`
}
type Payment struct {
	DebtID   string `json:"debt_id"`
	Debtor   int    `json:"debtor"`
	Creditor int    `json:"creditor"`
	Amount   Amount `json:"amount"`
}
type SettlementAudit struct {
	Algorithm string             `json:"algorithm"`
	Before    string             `json:"before"`
	After     string             `json:"after"`
	Previous  string             `json:"previous"`
	Payments  []Payment          `json:"payments"`
	Receipts  []FinancialReceipt `json:"receipts"`
	Repeats   string             `json:"repeats"`
}

// SettlementCursor is private working state. Committed is never changed until Done;
// callers must not publish Ledger or admit another game command while !Done.
type SettlementCursor struct {
	HeadCredited bool               `json:"head_credited"`
	Version      string             `json:"version"`
	Committed    FinancialLedger    `json:"committed"`
	Ledger       FinancialLedger    `json:"ledger"`
	Queue        []FinancialReceipt `json:"queue"`
	Audit        []SettlementAudit  `json:"audit"`
	Operations   uint64             `json:"operations"`
	Done         bool               `json:"done"`
	Digest       string             `json:"digest"`
}

// An empty digest is never valid and is rejected before commitment.
func financialHash(v any) string {
	h, e := canonical.Hash(v)
	if e != nil {
		return ""
	}
	return h
}
func (c SettlementCursor) hash() string { c.Digest = ""; return financialHash(c) }
func (c SettlementCursor) boundaryHash() string {
	return financialHash(struct {
		HeadCredited bool               `json:"head_credited"`
		Ledger       FinancialLedger    `json:"ledger"`
		Queue        []FinancialReceipt `json:"queue"`
	}{c.HeadCredited, c.Ledger, c.Queue})
}
func (c SettlementCursor) clone() SettlementCursor {
	c.Committed = c.Committed.Clone()
	c.Ledger = c.Ledger.Clone()
	c.Queue = slices.Clone(c.Queue)
	c.Audit = slices.Clone(c.Audit)
	for i := range c.Audit {
		c.Audit[i].Payments = slices.Clone(c.Audit[i].Payments)
		c.Audit[i].Receipts = slices.Clone(c.Audit[i].Receipts)
	}
	return c
}
func BeginSettlement(l FinancialLedger, r []FinancialReceipt) (SettlementCursor, error) {
	if err := l.Validate(); err != nil {
		return SettlementCursor{}, err
	}
	r = slices.Clone(r)
	for _, x := range r {
		if !l.validSeat(x.Seat) || x.Amount.Sign() < 0 {
			return SettlementCursor{}, errors.New("invalid gross receipt")
		}
	}
	sort.SliceStable(r, func(i, j int) bool { return r[i].Seat < r[j].Seat })
	c := SettlementCursor{Version: SettlementVersion, Committed: l.Clone(), Ledger: l.Clone(), Queue: r}
	c.Done = len(r) == 0
	c.Digest = c.hash()
	if c.Digest == "" {
		return SettlementCursor{}, errors.New("settlement canonical encoding limit")
	}
	return c, nil
}
func debtHead(l FinancialLedger, seat int) int {
	best := -1
	for i, d := range l.Debts {
		if d.Debtor == seat && d.Remaining.Sign() > 0 && (best < 0 || d.Sequence < l.Debts[best].Sequence) {
			best = i
		}
	}
	return best
}
func minAmount(a, b Amount) Amount {
	if a.Cmp(b) < 0 {
		return a
	}
	return b
}

// referenceStep is the unbatched Q10 oracle: consume the head fully before
// creditor receipts appended to the FIFO tail are considered.
func referenceStep(c *SettlementCursor) SettlementAudit {
	seg := SettlementAudit{Algorithm: "reference", Before: c.boundaryHash(), Repeats: "1"}
	r := c.Queue[0]
	if !c.HeadCredited {
		seg.Receipts = []FinancialReceipt{r}
		c.Ledger.Scores[r.Seat] = c.Ledger.Scores[r.Seat].Add(r.Amount)
		c.HeadCredited = true
	}
	i := debtHead(c.Ledger, r.Seat)
	if i >= 0 && r.Amount.Sign() > 0 {
		d := &c.Ledger.Debts[i]
		paid := minAmount(r.Amount, d.Remaining)
		d.Remaining = d.Remaining.Sub(paid)
		r.Amount = r.Amount.Sub(paid)
		seg.Payments = append(seg.Payments, Payment{d.ID, d.Debtor, d.Creditor, paid})
		if d.Creditor >= 0 {
			c.Queue = append(c.Queue, FinancialReceipt{d.Creditor, paid})
		}
		c.Queue[0] = r
	}
	if r.Amount.Sign() == 0 || debtHead(c.Ledger, r.Seat) < 0 {
		c.Ledger.Cash[r.Seat] = c.Ledger.Cash[r.Seat].Add(r.Amount)
		c.Queue = c.Queue[1:]
		c.HeadCredited = false
	}
	seg.After = c.boundaryHash()
	return seg
}

// cycleStep applies only a single circulating receipt and a simple directed cycle
// of current oldest debts. Every cycle preserves the queue and all transfer sizes.
// floor(min(debt / receipt)) complete cycles end no later than the first exhausted
// head. This proves equivalence including receipt totals and ordered payments;
// branching, system sinks and interleaved receipts fall back to referenceStep.
func cycleStep(c *SettlementCursor) (SettlementAudit, bool) {
	if c.HeadCredited || len(c.Queue) != 1 || c.Queue[0].Amount.Sign() <= 0 {
		return SettlementAudit{}, false
	}
	r := c.Queue[0]
	seat := r.Seat
	seen := map[int]bool{}
	var heads []int
	var repeat *big.Int
	for !seen[seat] {
		seen[seat] = true
		i := debtHead(c.Ledger, seat)
		if i < 0 {
			return SettlementAudit{}, false
		}
		d := c.Ledger.Debts[i]
		if d.Creditor < 0 || d.Remaining.Cmp(r.Amount) < 0 {
			return SettlementAudit{}, false
		}
		ratio, _ := d.Remaining.Quo(r.Amount)
		rr := ratio.rat()
		k := new(big.Int).Quo(rr.Num(), rr.Denom())
		if repeat == nil || k.Cmp(repeat) < 0 {
			repeat = k
		}
		heads = append(heads, i)
		seat = d.Creditor
	}
	if seat != r.Seat || repeat.Cmp(big.NewInt(2)) < 0 {
		return SettlementAudit{}, false
	}
	seg := SettlementAudit{Algorithm: "simple-cycle-v1", Before: c.boundaryHash(), Repeats: repeat.String()}
	factor := amountRat(new(big.Rat).SetInt(repeat))
	delta := r.Amount.Mul(factor)
	for _, i := range heads {
		d := &c.Ledger.Debts[i]
		seg.Payments = append(seg.Payments, Payment{d.ID, d.Debtor, d.Creditor, r.Amount})
		seg.Receipts = append(seg.Receipts, FinancialReceipt{d.Debtor, r.Amount})
		d.Remaining = d.Remaining.Sub(delta)
		c.Ledger.Scores[d.Debtor] = c.Ledger.Scores[d.Debtor].Add(delta)
	}
	seg.After = c.boundaryHash()
	return seg, true
}

// budget counts durable segments, not a financial cutoff. A zero budget is a
// checkpoint-only call. A negative budget is invalid. Resume never replays a prefix.
func ResumeSettlement(c SettlementCursor, budget int, optimized bool) (SettlementCursor, bool, error) {
	if c.Version != SettlementVersion || c.Digest == "" || c.Digest != c.hash() {
		return c, false, errors.New("incompatible or corrupt settlement checkpoint")
	}
	if err := c.Ledger.Validate(); err != nil {
		return c, false, err
	}
	if err := c.Committed.Validate(); err != nil {
		return c, false, err
	}
	if c.Done != (len(c.Queue) == 0) || (c.HeadCredited && len(c.Queue) == 0) || c.Committed.GameID != c.Ledger.GameID || len(c.Audit) != int(c.Operations) {
		return c, false, errors.New("invalid settlement progress")
	}
	for _, r := range c.Queue {
		if !c.Ledger.validSeat(r.Seat) || r.Amount.Sign() < 0 {
			return c, false, errors.New("invalid pending receipt")
		}
	}
	if budget < 0 {
		return c, false, errors.New("negative work budget")
	}
	original := c
	c = c.clone()
	for work := 0; work < budget && len(c.Queue) > 0; work++ {
		var seg SettlementAudit
		var ok bool
		if optimized {
			seg, ok = cycleStep(&c)
		}
		if !ok {
			seg = referenceStep(&c)
		}
		if len(c.Audit) > 0 {
			seg.Previous = financialHash(c.Audit[len(c.Audit)-1])
		}
		c.Audit = append(c.Audit, seg)
		c.Operations++
	}
	c.Done = len(c.Queue) == 0
	if c.Done {
		c.Committed = c.Ledger.Clone()
	}
	c.Digest = c.hash()
	if c.Digest == "" {
		return original, false, errors.New("settlement canonical encoding limit; retained prior checkpoint")
	}
	return c, c.Done, nil
}

// ExpandSettlementAudit is deliberately bounded for diagnostics; a refusal does
// not alter any obligations or the compressed audit.
func ExpandSettlementAudit(a []SettlementAudit, limit int) ([]Payment, error) {
	var out []Payment
	for _, s := range a {
		k, ok := new(big.Int).SetString(s.Repeats, 10)
		if !ok || k.Sign() < 0 {
			return nil, errors.New("invalid audit count")
		}
		n := new(big.Int).Mul(k, big.NewInt(int64(len(s.Payments))))
		if !n.IsInt64() || n.Int64() > int64(limit-len(out)) {
			return nil, errors.New("audit expansion budget exhausted")
		}
		for j := int64(0); j < k.Int64(); j++ {
			out = append(out, s.Payments...)
		}
	}
	return out, nil
}

// ErrSettlementRequiresContinuation means the synchronous convenience API cannot
// handle this shape. The obligation is legal and must use the durable cursor API.
var ErrSettlementRequiresContinuation = errors.New("settlement requires bounded BeginSettlement/ResumeSettlement continuation")

// SettleReceipts handles only small acyclic obligation graphs. This is an API
// admission bound, not a cutoff in the rules: rejected calls preserve all inputs,
// and arbitrary exact cycles remain supported through the durable cursor API.
// An acyclic path crosses at most four seats. Each positive repayment either
// exhausts a debt or consumes a receipt; a split can be charged to an exhausted
// debt. Thus the small fixed shape has bounded work independent of denominators.
func SettleReceipts(l FinancialLedger, r []FinancialReceipt) (FinancialLedger, error) {
	if e := l.Validate(); e != nil {
		return l, e
	}
	if len(l.Debts) > 64 || len(r) > 64 {
		return l, ErrSettlementRequiresContinuation
	}
	colors := make([]uint8, len(l.Scores))
	var visit func(int) bool
	visit = func(seat int) bool {
		if colors[seat] == 1 {
			return false
		}
		if colors[seat] == 2 {
			return true
		}
		colors[seat] = 1
		for _, d := range l.Debts {
			if d.Debtor == seat && d.Creditor >= 0 && d.Remaining.Sign() > 0 && !visit(d.Creditor) {
				return false
			}
		}
		colors[seat] = 2
		return true
	}
	for seat := range l.Scores {
		if !visit(seat) {
			return l, ErrSettlementRequiresContinuation
		}
	}
	c, e := BeginSettlement(l, r)
	if e != nil {
		return l, e
	}
	c, done, e := ResumeSettlement(c, 16*(len(l.Debts)+len(r)+1), true)
	if e != nil {
		return l, e
	}
	if !done {
		return l, ErrSettlementRequiresContinuation
	}
	return c.Committed, nil
}
