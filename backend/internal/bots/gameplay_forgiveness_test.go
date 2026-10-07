package bots

import (
	"encoding/json"
	"github.com/metaphy6/cgms/backend/internal/game"
	"strings"
	"testing"
)

func TestSystemForgivenessOnlyDisclosesDebtorRequestedTerms(t *testing.T) {
	l := financeFixture(t)
	var e error
	l.Ledger, e = l.Ledger.Charge("requested-debt", 0, -1, game.IntAmount(9))
	if e != nil {
		t.Fatal(e)
	}
	l.Ledger, e = l.Ledger.Charge("unrelated-private-debt", 2, -1, game.IntAmount(17))
	if e != nil {
		t.Fatal(e)
	}
	before, e := ProjectGameplayFinance(l, 2)
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(before)
	if strings.Contains(string(b), "requested-debt") {
		t.Fatal("unrequested debt leaked")
	}
	own, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	var request *game.LifecycleOperation
	for _, op := range own.Legal {
		if op.Kind == "forgive" && op.DebtID == "requested-debt" {
			v := op
			request = &v
		}
	}
	if request == nil {
		t.Fatal("debtor cannot request system forgiveness")
	}
	l, e = l.ApplyOperation(*request)
	if e != nil {
		t.Fatal(e)
	}
	other, e := ProjectGameplayFinance(l, 2)
	if e != nil {
		t.Fatal(e)
	}
	b, _ = json.Marshal(other)
	if !strings.Contains(string(b), "requested-debt") || strings.Contains(string(b), "unrelated-private-debt") {
		t.Fatal("request projection leaks or hides wrong debt")
	}
	found := false
	for _, op := range other.Legal {
		if op.Kind == "forgive" && op.DebtID == "requested-debt" && op.Actor == 2 {
			found = true
		}
	}
	if !found {
		t.Fatal("explicit consent absent")
	}
}
