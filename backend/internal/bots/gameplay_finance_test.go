package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func financeFixture(t *testing.T) game.Lifecycle {
	t.Helper()
	s, _ := game.NewState(3, "finance-bot")
	l, e := game.NewLifecycle(s, game.NewFinancialLedger("finance-bot", 3))
	if e != nil {
		t.Fatal(e)
	}
	return l
}
func TestGameplayFinancePaysAffordableTriggeredPromise(t *testing.T) {
	l := financeFixture(t)
	var e error
	l.Promises, e = l.Promises.Offer(game.Promise{ID: "promise", Payer: 0, Recipient: 1, AwardID: "award", Mode: "fixed", Amount: game.IntAmount(6)})
	if e != nil {
		t.Fatal(e)
	}
	l.Promises, _ = l.Promises.Accept("promise", 1)
	l.Promises, e = l.Promises.RecordAward(game.PromiseAward{ID: "award", Payer: 0, Amount: game.IntAmount(10)})
	if e != nil {
		t.Fatal(e)
	}
	l.Ledger.Cash[0] = game.IntAmount(10)
	l.Ledger.Scores[0] = game.IntAmount(10)
	o, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	d, e := ChooseGameplayFinancial("economic@v1", o, nil, 1000)
	if e != nil {
		t.Fatal(e)
	}
	if d.Operation.Kind != "promise-pay" {
		t.Fatalf("affordable accepted promise chose %s", d.Operation.Kind)
	}
	if d.Operation.Actor != 1 {
		t.Fatal("financial seat convention")
	}
}

func triggeredFinance(t *testing.T, amount, cash int64) game.Lifecycle {
	t.Helper()
	l := financeFixture(t)
	l.Promises, _ = l.Promises.Offer(game.Promise{ID: "p", Payer: 0, Recipient: 1, AwardID: "a", Mode: "fixed", Amount: game.IntAmount(amount)})
	l.Promises, _ = l.Promises.Accept("p", 1)
	l.Promises, _ = l.Promises.RecordAward(game.PromiseAward{ID: "a", Payer: 0, Amount: game.IntAmount(amount)})
	l.Ledger.Cash[0] = game.IntAmount(cash)
	l.Ledger.Scores[0] = game.IntAmount(cash)
	return l
}
func TestGameplayFinanceReducedOfferAnswerAndActualPayment(t *testing.T) {
	l := triggeredFinance(t, 10, 4)
	o, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	d, e := ChooseGameplayFinancial("economic@v1", o, nil, 1000)
	if e != nil {
		t.Fatal(e)
	}
	if d.Operation.Kind != "promise-final-offer" || d.Operation.Amount.Cmp(game.IntAmount(4)) != 0 {
		t.Fatalf("expected exact affordable final offer, got %+v", d.Operation)
	}
	l, e = l.ApplyOperation(d.Operation)
	if e != nil {
		t.Fatal(e)
	}
	o, e = ProjectGameplayFinance(l, 2)
	if e != nil {
		t.Fatal(e)
	}
	d, e = ChooseGameplayFinancial("economic@v1", o, nil, 1000)
	if e != nil || d.Operation.Kind != "promise-final-answer" || !d.Operation.Accept {
		t.Fatal("incoming final decision missing", e)
	}
	l, e = l.ApplyOperation(d.Operation)
	if e != nil {
		t.Fatal(e)
	}
	if l.Promises.Promises[0].Paid.Sign() != 0 {
		t.Fatal("acceptance invented receipt")
	}
	o, _ = ProjectGameplayFinance(l, 1)
	d, e = ChooseGameplayFinancial("economic@v1", o, nil, 1000)
	if e != nil || d.Operation.Kind != "promise-pay" {
		t.Fatal("accepted final offer not paid", e)
	}
	l, e = l.ApplyOperation(d.Operation)
	if e != nil {
		t.Fatal(e)
	}
	o, e = ProjectGameplayFinance(l, 1)
	if e != nil || o.Waiting != "automatic-settlement" || len(o.Legal) != 0 {
		t.Fatal("bot interleaved automatic settlement", e)
	}
}
func TestGameplayFinanceNoFabricatedSystemConsent(t *testing.T) {
	l := financeFixture(t)
	var e error
	l.Ledger, e = l.Ledger.Charge("system", 0, -1, game.IntAmount(8))
	if e != nil {
		t.Fatal(e)
	}
	o, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	for _, op := range o.Legal {
		if op.Kind == "forgive" {
			n, err := l.ApplyOperation(op)
			if err != nil || n.Ledger.Debts[0].Remaining.Cmp(game.IntAmount(8)) != 0 || len(n.ForgivenessConsents) != 1 || len(n.ForgivenessConsents[0].Seats) != 1 || n.ForgivenessConsents[0].Seats[0] != 1 {
				t.Fatal("request fabricated another participant consent or payment", err)
			}
		}
	}
}
func TestGameplayFinancePartyOnlyPrivacyAndExplicitRefusal(t *testing.T) {
	l := triggeredFinance(t, 10, 10)
	o, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	n := l.Clone()
	n.Ledger, e = n.Ledger.Charge("private-other-debt", 2, 1, game.IntAmount(7))
	if e != nil {
		t.Fatal(e)
	}
	other, e := ProjectGameplayFinance(n, 1)
	if e != nil {
		t.Fatal(e)
	}
	if game.Digest(o) != game.Digest(other) {
		t.Fatal("unrelated debt leaked into financial observation")
	}
	d, e := ChooseGameplayFinancial("pressure@v1", o, nil, 1000)
	if e != nil || d.Operation.Kind != "promise-refuse" {
		t.Fatal("pressure policy refusal should be explicit", e)
	}
	n, e = l.ApplyOperation(d.Operation)
	if e != nil || n.Promises.Promises[0].Status != "refused" {
		t.Fatal("explicit refusal not recorded", e)
	}
}

func TestGameplayFinanceDoesNotOfferUnpayableZeroWhileInDebt(t *testing.T) {
	l := triggeredFinance(t, 10, 0)
	var e error
	l.Ledger, e = l.Ledger.Charge("prior-system", 0, -1, game.IntAmount(3))
	if e != nil {
		t.Fatal(e)
	}
	o, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	d, e := ChooseGameplayFinancial("economic@v1", o, nil, 1000)
	if e != nil {
		t.Fatal(e)
	}
	if d.Operation.Kind != "promise-refuse" {
		t.Fatalf("existing debts make voluntary payments unavailable; selected %s", d.Operation.Kind)
	}
}
