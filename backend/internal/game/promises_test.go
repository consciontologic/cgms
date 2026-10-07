package game

import "testing"

func TestPromiseReputationBoundaries(t *testing.T) {
	for _, tc := range []struct {
		paid int64
		want string
	}{{39, "none"}, {40, "anyhoo"}, {100, "trust"}} {
		if got := promiseClassification(IntAmount(100), IntAmount(tc.paid), false); got != tc.want {
			t.Fatalf("paid %d: got %s want %s", tc.paid, got, tc.want)
		}
	}
	if got := promiseClassification(IntAmount(100), IntAmount(100), true); got != "scam" {
		t.Fatal("refusal must override payment", got)
	}
}

func promisedBook(t *testing.T, amount Amount, mode string) PromiseBook {
	t.Helper()
	b := NewPromiseBook("g", 3)
	var e error
	b, e = b.Offer(Promise{ID: "p", Payer: 0, Recipient: 1, AwardID: "ringleader/0/1", Mode: mode, Amount: amount})
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.Accept("p", 1)
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.RecordAward(PromiseAward{"ringleader/0/1", 0, IntAmount(100)})
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func promiseLedger(t *testing.T) FinancialLedger {
	t.Helper()
	l := NewFinancialLedger("g", 3)
	l, e := SettleReceipts(l, []FinancialReceipt{{0, IntAmount(100)}})
	if e != nil {
		t.Fatal(e)
	}
	return l
}
func TestPromisePaymentDebtGrossAndExactPercentage(t *testing.T) {
	half, _ := NewAmount("1", "2")
	b := promisedBook(t, half, "percentage")
	l := promiseLedger(t)
	l, e := l.Charge("recipient-debt", 1, 2, IntAmount(50))
	if e != nil {
		t.Fatal(e)
	}
	old := Digest(b)
	b2, l2, e := b.Pay("p", "pay/1", 0, l)
	if e != nil {
		t.Fatal(e)
	}
	if Digest(b) != old || b2.Promises[0].Paid.Cmp(IntAmount(50)) != 0 || l2.Cash[1].Sign() != 0 || l2.Cash[2].Cmp(IntAmount(50)) != 0 {
		t.Fatal("gross not delivered through recipient debt or state aliased")
	}
	b2, e = b2.CloseBoard()
	if e != nil {
		t.Fatal(e)
	}
	out, e := b2.Outcomes()
	if e != nil || len(out) != 1 || out[0].Outcome != "trust" {
		t.Fatal(out, e)
	}
	retry, rl, e := b2.Pay("p", "pay/1", 0, l2)
	if e != nil || Digest(retry) != Digest(b2) || Digest(rl) != Digest(l2) {
		t.Fatal("duplicate changed payment", e)
	}
}
func TestPromiseFinalOffersAndPendingFinish(t *testing.T) {
	for _, tc := range []struct {
		pay    int64
		accept bool
		refuse bool
		want   string
	}{{39, true, false, "none"}, {40, true, false, "anyhoo"}, {40, false, false, "scam"}, {40, true, true, "scam"}, {0, true, false, "none"}} {
		b := promisedBook(t, IntAmount(100), "fixed")
		l := promiseLedger(t)
		var e error
		b, e = b.CloseBoard()
		if e != nil {
			t.Fatal(e)
		}
		l, e = l.CloseOrdinary(nil)
		if e != nil {
			t.Fatal(e)
		}
		if _, e = b.Finish(0); e == nil {
			t.Fatal("pending promise allowed finish")
		}
		b, e = b.OfferFinal("p", 0, IntAmount(tc.pay))
		if e != nil {
			t.Fatal(e)
		}
		if _, e = b.Finish(1); e == nil {
			t.Fatal("incoming offer allowed finish")
		}
		b, e = b.AnswerFinal("p", 1, tc.accept)
		if e != nil {
			t.Fatal(e)
		}
		if tc.accept {
			if tc.refuse {
				b, e = b.Refuse("p", 0)
			} else {
				b, l, e = b.Pay("p", "pay", 0, l)
			}
			if e != nil {
				t.Fatal(e)
			}
		}
		out, e := b.Outcomes()
		if e != nil || out[0].Outcome != tc.want {
			t.Fatal(tc, out, e)
		}
		for i := 0; i < 3; i++ {
			b, e = b.Finish(i)
			if e != nil {
				t.Fatal(e)
			}
		}
		b, e = b.Finalize()
		if e != nil {
			t.Fatal(e)
		}
		if _, e = b.OfferFinal("p", 0, IntAmount(1)); e == nil {
			t.Fatal("finalized mutation")
		}
	}
}
func TestPromiseUntriggeredCoupAndScamPrecedence(t *testing.T) {
	b := NewPromiseBook("g", 3)
	var e error
	for _, id := range []string{"ordinary-award", "ringleader", "coup"} {
		b, e = b.Offer(Promise{ID: id, Payer: 0, Recipient: 1, AwardID: id, Mode: "fixed", Amount: IntAmount(10)})
		if e != nil {
			t.Fatal(e)
		}
		b, e = b.Accept(id, 1)
		if e != nil {
			t.Fatal(e)
		}
	}
	b, e = b.RecordAward(PromiseAward{"ringleader", 0, IntAmount(10)})
	if e != nil {
		t.Fatal(e)
	}
	l := promiseLedger(t)
	b, l, e = b.Pay("ringleader", "r-payment", 0, l)
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.RecordAward(PromiseAward{"coup", 0, IntAmount(50)})
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.CloseBoard()
	if e != nil {
		t.Fatal(e)
	}
	out, e := b.Outcomes()
	if e != nil || out[0].Outcome != "pending" {
		t.Fatal("silence was refusal", out, e)
	}
	b, e = b.Refuse("coup", 0)
	if e != nil {
		t.Fatal(e)
	}
	out, e = b.Outcomes()
	if e != nil || len(out) != 1 || out[0].Outcome != "scam" || out[0].Promised.Cmp(IntAmount(20)) != 0 {
		t.Fatal(out, e)
	}
	out, e = b.Nullify().Outcomes()
	if e != nil || len(out) != 0 {
		t.Fatal("void reputation", out, e)
	}
}
func TestPromiseResumableCyclicRecipientSettlement(t *testing.T) {
	b := promisedBook(t, IntAmount(1), "fixed")
	l := promiseLedger(t)
	var e error
	l, e = l.Charge("b-c", 1, 2, IntAmount(100))
	if e != nil {
		t.Fatal(e)
	}
	l, e = l.Charge("c-b", 2, 1, IntAmount(100))
	if e != nil {
		t.Fatal(e)
	}
	c, e := b.PreparePayment("p", "cycle-pay", 0, l)
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = b.CommitPayment(c, l); e == nil {
		t.Fatal("unfinished gross recorded")
	}
	before := Digest(b)
	for !c.Settlement.Done {
		c.Settlement, _, e = ResumeSettlement(c.Settlement, 1, true)
		if e != nil {
			t.Fatal(e)
		}
	}
	n, paid, e := b.CommitPayment(c, l)
	if e != nil {
		t.Fatal(e)
	}
	if n.Promises[0].Paid.Cmp(IntAmount(1)) != 0 || Digest(b) != before || paid.Debts[0].Remaining.Sign() != 0 || paid.Debts[1].Remaining.Sign() != 0 {
		t.Fatal("cyclic payment mismatch")
	}
	bad := c
	bad.PromiseID = "missing"
	if _, _, e = b.CommitPayment(bad, l); e == nil {
		t.Fatal("forged missing promise accepted")
	}
}
func TestPromiseRejectCorruptAndWrongActors(t *testing.T) {
	b := promisedBook(t, IntAmount(10), "fixed")
	before := Digest(b)
	if _, e := b.OfferFinal("p", 1, IntAmount(4)); e == nil {
		t.Fatal("wrong payer")
	}
	if Digest(b) != before {
		t.Fatal("rejection mutation")
	}
	bad := b.Clone()
	bad.Promises[0].Promised = IntAmount(11)
	if bad.Validate() == nil {
		t.Fatal("forged amount")
	}
	bad = b.Clone()
	bad.Promises[0].Paid = IntAmount(10)
	if bad.Validate() == nil {
		t.Fatal("unrecorded gross")
	}
	bad = b.Clone()
	bad.Finished[0] = true
	if bad.Validate() == nil {
		t.Fatal("pending finished")
	}
	b, e := b.CloseBoard()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = b.Offer(Promise{ID: "late"}); e == nil {
		t.Fatal("new board closed promise")
	}
}

func TestPromiseLateAcceptanceAndZeroAward(t *testing.T) {
	b := NewPromiseBook("g", 3)
	var e error
	b, e = b.Offer(Promise{ID: "late", Payer: 0, Recipient: 1, AwardID: "a", Mode: "fixed", Amount: IntAmount(10)})
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.RecordAward(PromiseAward{"a", 0, IntAmount(10)})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = b.Accept("late", 1); e == nil {
		t.Fatal("retroactive promise accepted")
	}
	b = NewPromiseBook("g", 3)
	half, _ := NewAmount("1", "2")
	b, e = b.Offer(Promise{ID: "zero", Payer: 0, Recipient: 1, AwardID: "a", Mode: "percentage", Amount: half})
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.Accept("zero", 1)
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.RecordAward(PromiseAward{"a", 0, Amount{}})
	if e != nil {
		t.Fatal(e)
	}
	b, e = b.CloseBoard()
	if e != nil {
		t.Fatal(e)
	}
	out, e := b.Outcomes()
	if e != nil || len(out) != 0 {
		t.Fatal("zero realized reputation", out, e)
	}
	for i := 0; i < 3; i++ {
		b, e = b.Finish(i)
		if e != nil {
			t.Fatal(e)
		}
	}
}
