package game

import (
	"errors"
	"testing"
)

func TestLedgerChargeOnce(t *testing.T) {
	l := NewFinancialLedger("g1", 3)
	n, e := l.Charge("kidnap", 0, 1, IntAmount(10))
	if e != nil {
		t.Fatal(e)
	}
	if n.Scores[0].String() != "-10" || len(n.Debts) != 1 || n.Cash[1].Sign() != 0 {
		t.Fatalf("charge must debit once without creditor income: %+v", n)
	}
	if l.Scores[0].Sign() != 0 || len(l.Debts) != 0 {
		t.Fatal("input mutated")
	}
	again, e := n.Charge("kidnap", 0, 1, IntAmount(10))
	if e != nil || len(again.Debts) != 1 || again.Scores[0].String() != "-10" {
		t.Fatal("retry charges twice")
	}
}
func TestFinancialBoundaries(t *testing.T) {
	l := NewFinancialLedger("g1", 3)
	l.Cash[1] = IntAmount(20)
	l.Scores[1] = IntAmount(20)
	n, e := l.Coup(0)
	if e != nil {
		t.Fatal(e)
	}
	if n.Scores[1].String() != "-50" || n.Cash[1].Sign() != 0 || len(n.Debts) != 2 || n.Debts[0].Remaining.String() != "50" {
		t.Fatalf("F04 cash clear before Coup charge: %+v", n)
	}
}
func TestForgivenessCurrentCharge(t *testing.T) {
	l := NewFinancialLedger("g1", 3)
	l, _ = l.Charge("d", 0, 1, IntAmount(10))
	n, e := l.Forgive("f", "d", IntAmount(10), []int{1})
	if e != nil {
		t.Fatal(e)
	}
	if n.Scores[0].Sign() != 0 || n.Cash[0].Sign() != 0 || n.Debts[0].Remaining.Sign() != 0 {
		t.Fatalf("nonspendable correction: %+v", n)
	}
}
func finalizeTest(t *testing.T, l FinancialLedger) FinancialLedger {
	t.Helper()
	var e error
	for i := range l.Scores {
		l, e = l.FinishSettlement(i)
		if e != nil {
			t.Fatal(e)
		}
	}
	l, e = l.Finalize()
	if e != nil {
		t.Fatal(e)
	}
	return l
}
func TestLedgerWorkedExamplesAndPonzi(t *testing.T) {
	l := NewFinancialLedger("g1", 3)
	l, _ = l.Charge("kidnap", 0, 1, IntAmount(10))
	l, _ = SettleReceipts(l, []FinancialReceipt{{0, IntAmount(6)}})
	l, _ = SettleReceipts(l, []FinancialReceipt{{0, IntAmount(8)}})
	if l.Scores[0].String() != "4" || l.Cash[0].String() != "4" || l.Scores[1].String() != "10" || l.Debts[0].Remaining.Sign() != 0 {
		t.Fatal("6 then 8 worked ledger")
	}
	l = NewFinancialLedger("g1", 3)
	l, _ = l.Charge("kidnap", 1, 2, IntAmount(10))
	l, _ = SettleReceipts(l, []FinancialReceipt{{1, IntAmount(6)}})
	l, _ = l.Coup(0)
	if l.Scores[1].String() != "-54" || l.Debts[0].Remaining.String() != "4" {
		t.Fatal("Coup carried adjustment")
	}
	p := NewFinancialLedger("p", 3)
	p, _ = SettleReceipts(p, []FinancialReceipt{{0, IntAmount(11)}})
	p, e := p.Ponzi(0, 1, 2)
	if e != nil || p.Cash[1].String() != "11/2" || p.Cash[2].String() != "11/2" {
		t.Fatal("exclusive paired halves", e)
	}
}
func TestFinancialCrossGameCorrectionAndVoid(t *testing.T) {
	l := NewFinancialLedger("g1", 3)
	l, _ = l.Charge("d", 0, 1, IntAmount(10))
	l, _ = l.CloseOrdinary(nil)
	l = finalizeTest(t, l)
	g1 := l.Clone()
	l, e := l.NextGame("g2")
	if e != nil {
		t.Fatal(e)
	}
	start := l.Clone()
	l, e = l.Forgive("f", "d", IntAmount(10), []int{1})
	if e != nil || l.Cash[0].Sign() != 0 || l.MatchScores()[0].Sign() != 0 {
		t.Fatal("cross-game correction", e)
	}
	retry, e := l.Forgive("f", "d", IntAmount(10), []int{1})
	if e != nil || len(retry.Corrections) != 1 {
		t.Fatal("duplicate correction")
	}
	void, e := l.Nullify(start, []int{0, 1, 2})
	if e != nil || void.Debts[0].Remaining.String() != "10" || len(void.Corrections) != 0 {
		t.Fatal("nullification restoration", e)
	}
	l, e = l.Coup(0)
	if e != nil || l.MatchScores()[0].String() != "50" || l.Completed[0].Scores[0].String() != "-10" || g1.Completed[0].Scores[0].String() != "-10" {
		t.Fatal("N06 correction survives Coup", e)
	}
	same := NewFinancialLedger("s", 3)
	same, _ = same.Charge("d", 0, 1, IntAmount(10))
	same, _ = same.Forgive("f", "d", IntAmount(10), []int{1})
	same, _ = same.Coup(0)
	if same.MatchScores()[0].String() != "50" {
		t.Fatal("same-game extra bonus")
	}
}
func TestFinancialCoupNextGameAndDeparture(t *testing.T) {
	l := NewFinancialLedger("g1", 3)
	l, _ = l.Coup(0)
	l = finalizeTest(t, l)
	l, _ = l.NextGame("g2")
	for _, c := range l.Cash {
		if c.Sign() != 0 {
			t.Fatal("cash survived game")
		}
	}
	l, _ = SettleReceipts(l, []FinancialReceipt{{1, IntAmount(50)}})
	if l.Scores[1].String() != "50" || l.MatchScores()[1].Sign() != 0 || l.Cash[1].Sign() != 0 {
		t.Fatal("Coup next-game receipt")
	}
	d := NewFinancialLedger("d", 3)
	d, _ = d.Charge("d", 1, 2, IntAmount(10))
	d, _ = d.Depart(1)
	d, _ = d.Coup(0)
	if d.Scores[1].String() != "-10" || len(d.Debts) != 2 {
		t.Fatal("departed Coup exemption")
	}
	o := NewFinancialLedger("o", 3)
	o, _ = o.CloseOrdinary([]FinancialReceipt{{0, IntAmount(12)}})
	if _, e := o.Finalize(); e == nil {
		t.Fatal("silence finalized")
	}
	o = finalizeTest(t, o)
	if o.Scores[0].String() != "12" || o.Cash[0].Sign() != 0 {
		t.Fatal("expiration changes score")
	}
	fresh := NewFinancialLedger("newmatch", 3)
	if len(fresh.Debts) != 0 {
		t.Fatal("debt leaked between matches")
	}
}
func TestImmediateChargeAndPartialForgiveness(t *testing.T) {
	l := NewFinancialLedger("g", 3)
	l, _ = SettleReceipts(l, []FinancialReceipt{{0, IntAmount(6)}})
	l, e := l.Charge("d", 0, 1, IntAmount(10))
	if e != nil || l.Scores[0].String() != "-4" || l.Cash[1].String() != "6" || l.Debts[0].Remaining.String() != "4" {
		t.Fatal("immediate payment", e)
	}
	l, e = l.Forgive("f", "d", IntAmount(4), []int{1})
	if e != nil || l.Scores[0].Sign() != 0 || l.Cash[0].Sign() != 0 || l.Cash[1].String() != "6" {
		t.Fatal("partial correction manufactured cash", e)
	}
}
func TestChargeRetryRejectsChangedCreditor(t *testing.T) {
	l := NewFinancialLedger("g", 3)
	l, _ = l.Charge("d", 0, 1, IntAmount(10))
	if _, e := l.Charge("d", 0, 2, IntAmount(10)); e == nil {
		t.Fatal("retry changed original creditor")
	}
}

func TestFinancialValidationRejectsCorruptProvenance(t *testing.T) {
	base := NewFinancialLedger("g1", 3)
	base, _ = base.Charge("d", 0, 1, IntAmount(10))
	base, _ = base.Forgive("f", "d", IntAmount(4), []int{1})
	cases := map[string]func(*FinancialLedger){
		"duplicate charge debt": func(l *FinancialLedger) {
			d := l.Debts[0]
			d.ID = "duplicate"
			d.Sequence = 2
			l.NextSequence = 2
			l.Debts = append(l.Debts, d)
		},
		"orphan charge":                func(l *FinancialLedger) { l.Debts = nil },
		"unknown correction charge":    func(l *FinancialLedger) { l.Corrections[0].ChargeID = "unknown" },
		"unknown correction debt":      func(l *FinancialLedger) { l.Corrections[0].DebtID = "unknown" },
		"correction seat":              func(l *FinancialLedger) { l.Corrections[0].Debtor = 9 },
		"wrong correction debtor":      func(l *FinancialLedger) { l.Corrections[0].Debtor = 2 },
		"correction amount":            func(l *FinancialLedger) { l.Corrections[0].Amount = IntAmount(5) },
		"negative correction":          func(l *FinancialLedger) { l.Corrections[0].Amount = IntAmount(-1) },
		"missing correction operation": func(l *FinancialLedger) { l.Operations = nil },
		"duplicate correction":         func(l *FinancialLedger) { l.Corrections = append(l.Corrections, l.Corrections[0]) },
		"forged match correction":      func(l *FinancialLedger) { l.Corrections[0].MatchLevel = true },
		"unknown creating game":        func(l *FinancialLedger) { l.Corrections[0].GameID = "unknown" },
		"completed dimension":          func(l *FinancialLedger) { l.Completed = []FinancialResult{{"prior", make([]Amount, 4)}} },
		"duplicate completed game": func(l *FinancialLedger) {
			l.Completed = []FinancialResult{{"prior", make([]Amount, 3)}, {"prior", make([]Amount, 3)}}
		},
		"live game also completed": func(l *FinancialLedger) { l.Completed = []FinancialResult{{l.GameID, make([]Amount, 3)}} },
		"unknown original game":    func(l *FinancialLedger) { l.Charges[0].GameID = "unknown"; l.Debts[0].GameID = "unknown" },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			l := base.Clone()
			mutate(&l)
			if l.Validate() == nil {
				t.Fatal("accepted corrupt ledger")
			}
			if _, e := BeginSettlement(l, nil); e == nil {
				t.Fatal("accepted corrupt settlement ledger")
			}
		})
	}
}
func TestFinancialWrappersPreserveRejectedLedger(t *testing.T) {
	base := cycleLedger(t, 2, "3").Ledger
	base.Cash[2] = IntAmount(3)
	base.Scores[2] = IntAmount(3)
	for name, call := range map[string]func(FinancialLedger) (FinancialLedger, error){"Coup": func(l FinancialLedger) (FinancialLedger, error) { return l.Coup(2) }, "Transfer": func(l FinancialLedger) (FinancialLedger, error) { return l.Transfer("t", 2, 0, IntAmount(1)) }, "Ponzi": func(l FinancialLedger) (FinancialLedger, error) { return l.Ponzi(2, 0, 1) }, "Charge": func(l FinancialLedger) (FinancialLedger, error) { return l.Charge("new", 2, 0, IntAmount(1)) }} {
		t.Run(name, func(t *testing.T) {
			before := financialHash(base)
			out, e := call(base)
			if !errors.Is(e, ErrSettlementRequiresContinuation) || financialHash(out) != before || financialHash(base) != before {
				t.Fatal("shape refusal changed financial state", e)
			}
		})
	}
}
