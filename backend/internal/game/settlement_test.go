package game

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestSettlementWorkedDebt(t *testing.T) {
	l := NewFinancialLedger("g1", 3)
	l, _ = l.Charge("kidnap", 0, 1, IntAmount(10))
	c, e := BeginSettlement(l, []FinancialReceipt{{0, IntAmount(6)}})
	if e != nil {
		t.Fatal(e)
	}
	c, done, e := ResumeSettlement(c, 100, false)
	if e != nil || !done {
		t.Fatal(e)
	}
	if c.Ledger.Scores[0].String() != "-4" || c.Ledger.Scores[1].String() != "6" || c.Ledger.Debts[0].Remaining.String() != "4" {
		t.Fatalf("worked ledger wrong: %+v", c.Ledger)
	}
}
func finishCursor(t *testing.T, c SettlementCursor, opt bool) SettlementCursor {
	t.Helper()
	for !c.Done {
		var e error
		c, _, e = ResumeSettlement(c, 1, opt)
		if e != nil {
			t.Fatal(e)
		}
	}
	return c
}
func cycleLedger(t *testing.T, seats int, den string) SettlementCursor {
	t.Helper()
	l := NewFinancialLedger("g1", 3)
	for i := 0; i < seats; i++ {
		var e error
		l, e = l.Charge(fmt.Sprintf("d%d", i), i, (i+1)%seats, IntAmount(1))
		if e != nil {
			t.Fatal(e)
		}
	}
	a, e := NewAmount("1", den)
	if e != nil {
		t.Fatal(e)
	}
	c, e := BeginSettlement(l, []FinancialReceipt{{0, a}})
	if e != nil {
		t.Fatal(e)
	}
	return c
}
func TestSettlementBatchOracleAndResume(t *testing.T) {
	for _, seats := range []int{2, 3} {
		for _, d := range []string{"1", "2", "3", "7"} {
			t.Run(fmt.Sprintf("%d/%s", seats, d), func(t *testing.T) {
				c := cycleLedger(t, seats, d)
				ref := finishCursor(t, c, false)
				opt := finishCursor(t, c, true)
				if !reflect.DeepEqual(ref.Ledger, opt.Ledger) {
					t.Fatal("ledger mismatch")
				}
				rp, e := ExpandSettlementAudit(ref.Audit, 1000)
				if e != nil {
					t.Fatal(e)
				}
				op, e := ExpandSettlementAudit(opt.Audit, 1000)
				if e != nil || !reflect.DeepEqual(rp, op) {
					t.Fatal("ordered audit mismatch", e)
				}
				if opt.Ledger.Cash[0].Cmp(c.Queue[0].Amount) != 0 {
					t.Fatal("cycle cash")
				}
				cursor := c
				for !cursor.Done {
					b, _ := json.Marshal(cursor)
					var restored SettlementCursor
					if e = json.Unmarshal(b, &restored); e != nil {
						t.Fatal(e)
					}
					cursor, _, e = ResumeSettlement(restored, 1, true)
					if e != nil {
						t.Fatal(e)
					}
				}
				if cursor.Digest != opt.Digest {
					t.Fatal("resume mismatch")
				}
			})
		}
	}
}
func TestSettlementLargeDenominatorUsefulProgress(t *testing.T) {
	c := cycleLedger(t, 2, "100000000000000000000000000000000000000000000000003")
	n, done, e := ResumeSettlement(c, 1, true)
	if e != nil || done {
		t.Fatal("expected retained remainder", e)
	}
	if n.Ledger.Debts[0].Remaining.Sign() != 0 || len(n.Audit) != 1 || n.Audit[0].Repeats != "100000000000000000000000000000000000000000000000003" {
		t.Fatal("no useful exact progress")
	}
	if c.Ledger.Debts[0].Remaining.String() != "1" || n.Committed.Debts[0].Remaining.String() != "1" {
		t.Fatal("public boundary mutated")
	}
	raw, _ := json.Marshal(n)
	var restored SettlementCursor
	if e = json.Unmarshal(raw, &restored); e != nil {
		t.Fatal(e)
	}
	out, done, e := ResumeSettlement(restored, 1, true)
	if e != nil || !done || out.Operations != 2 {
		t.Fatal("prefix was repeated", e)
	}
	if _, e = ExpandSettlementAudit(out.Audit, 100); e == nil {
		t.Fatal("expansion should remain bounded")
	}
	t.Logf("denominator_digits=51 durable_operations=%d compressed_segments=%d checkpoint_bytes=%d", out.Operations, len(out.Audit), len(raw))
}
func TestSettlementOldestSystemAndInterleaving(t *testing.T) {
	l := NewFinancialLedger("g", 3)
	l, _ = l.Charge("system", 0, -1, IntAmount(5))
	l, _ = l.Charge("player", 0, 1, IntAmount(10))
	c, _ := BeginSettlement(l, []FinancialReceipt{{0, IntAmount(8)}})
	r := finishCursor(t, c, false)
	if r.Ledger.Debts[1].Remaining.String() != "7" || r.Ledger.Scores[1].String() != "3" {
		t.Fatal("oldest system first")
	}
	for a := int64(0); a < 5; a++ {
		for b := int64(0); b < 5; b++ {
			l := NewFinancialLedger("g", 3)
			l, _ = l.Charge("a", 0, 1, IntAmount(2))
			l, _ = l.Charge("b", 1, 2, IntAmount(3))
			l, _ = l.Charge("c", 2, 0, IntAmount(2))
			l, _ = l.Charge("d", 0, -1, IntAmount(1))
			c, _ := BeginSettlement(l, []FinancialReceipt{{1, IntAmount(b)}, {0, IntAmount(a)}})
			ref := finishCursor(t, c, false)
			opt := finishCursor(t, c, true)
			rp, _ := ExpandSettlementAudit(ref.Audit, 1000)
			op, _ := ExpandSettlementAudit(opt.Audit, 1000)
			if !reflect.DeepEqual(ref.Ledger, opt.Ledger) || !reflect.DeepEqual(rp, op) {
				t.Fatalf("differential %d/%d", a, b)
			}
		}
	}
}
func TestSettlementCorruptCheckpoint(t *testing.T) {
	c := cycleLedger(t, 2, "3")
	c.Ledger.Scores[0] = IntAmount(99)
	if _, _, e := ResumeSettlement(c, 1, true); e == nil {
		t.Fatal("accepted corrupt checkpoint")
	}
}
func FuzzSettlementEquivalence(f *testing.F) {
	f.Add(uint8(2), uint8(3), uint8(1))
	f.Fuzz(func(t *testing.T, a, b, c uint8) {
		l := NewFinancialLedger("g", 3)
		for i, v := range []uint8{a, b, c} {
			l, _ = l.Charge(fmt.Sprint(i), i, (i+1)%3, IntAmount(int64(v%7+1)))
		}
		v, _ := NewAmount("1", fmt.Sprint(int(a%7+1)))
		cur, e := BeginSettlement(l, []FinancialReceipt{{0, v}, {1, IntAmount(int64(b % 3))}})
		if e != nil {
			t.Fatal(e)
		}
		ref := finishCursor(t, cur, false)
		opt := finishCursor(t, cur, true)
		rp, _ := ExpandSettlementAudit(ref.Audit, 10000)
		op, _ := ExpandSettlementAudit(opt.Audit, 10000)
		if !reflect.DeepEqual(ref.Ledger, opt.Ledger) || !reflect.DeepEqual(rp, op) {
			t.Fatal("oracle disagreement")
		}
	})
}
func TestSettlementGoldenQ10(t *testing.T) {
	raw, e := os.ReadFile("../../testdata/ledger/q10-cycle.json")
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Denominator string   `json:"denominator"`
		PaymentIDs  []string `json:"expected_payment_ids"`
		Scores      []string `json:"expected_scores"`
		Cash        []string `json:"expected_cash"`
	}
	if e = json.Unmarshal(raw, &v); e != nil {
		t.Fatal(e)
	}
	c := finishCursor(t, cycleLedger(t, 2, v.Denominator), true)
	p, e := ExpandSettlementAudit(c.Audit, 100)
	if e != nil {
		t.Fatal(e)
	}
	var ids []string
	for _, x := range p {
		ids = append(ids, x.DebtID)
	}
	if !reflect.DeepEqual(ids, v.PaymentIDs) {
		t.Fatal("ordered golden payments")
	}
	for i := range c.Ledger.Scores {
		if c.Ledger.Scores[i].String() != v.Scores[i] || c.Ledger.Cash[i].String() != v.Cash[i] {
			t.Fatal("golden balances")
		}
	}
}
func TestSettlementRejectsBrokenChargeLineage(t *testing.T) {
	c := cycleLedger(t, 2, "3")
	c.Ledger.Debts[0].ChargeID = "missing"
	if _, e := BeginSettlement(c.Ledger, nil); e == nil {
		t.Fatal("accepted missing original charge")
	}
}
func TestSettlementPartialHeadResume(t *testing.T) {
	l := NewFinancialLedger("g", 3)
	l, _ = l.Charge("first", 0, -1, IntAmount(5))
	l, _ = l.Charge("second", 0, 1, IntAmount(10))
	c, _ := BeginSettlement(l, []FinancialReceipt{{0, IntAmount(8)}})
	c, _, e := ResumeSettlement(c, 1, false)
	if e != nil {
		t.Fatal(e)
	}
	if len(c.Queue) == 0 || c.Queue[0].Seat != 0 || c.Queue[0].Amount.String() != "3" {
		t.Fatal("partially consumed receipt head must be durable")
	}
	end := finishCursor(t, c, false)
	if end.Ledger.Scores[0].String() != "-7" {
		t.Fatal("head receipt credited twice")
	}
}
func TestSynchronousSettlementRejectsCycleBeforeMutation(t *testing.T) {
	c := cycleLedger(t, 2, "3")
	before := c.Ledger.Clone()
	out, e := SettleReceipts(c.Ledger, c.Queue)
	if e == nil {
		t.Fatal("unbounded synchronous cycle must require continuation API")
	}
	if !reflect.DeepEqual(out, before) {
		t.Fatal("rejection changed committed ledger")
	}
	c = finishCursor(t, c, true)
	if c.Ledger.Cash[0].String() != "1/3" {
		t.Fatal("bounded continuation remains legal")
	}
}
func TestSettlementResumeRejectsInvalidRehashedState(t *testing.T) {
	for _, mutate := range []func(*SettlementCursor){func(c *SettlementCursor) { c.Queue[0].Seat = 9 }, func(c *SettlementCursor) { c.Ledger.Debts[0].ChargeID = "missing" }, func(c *SettlementCursor) { c.Queue[0].Amount = IntAmount(-1) }, func(c *SettlementCursor) { c.Done = true }} {
		c := cycleLedger(t, 2, "3")
		mutate(&c)
		c.Digest = c.hash()
		if _, _, e := ResumeSettlement(c, 0, true); e == nil {
			t.Fatal("accepted invalid rehashed cursor")
		}
	}
}
