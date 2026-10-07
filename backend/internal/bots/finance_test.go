package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestFinanceExplicitFinish(t *testing.T) {
	l := game.NewFinancialLedger("finance", 3)
	l.Phase = "settlement"
	o, e := ProjectLedger(l, 0)
	if e != nil {
		t.Fatal(e)
	}
	c, e := ChooseFinance("legal-random@v1", o, 10)
	if e != nil {
		t.Fatal(e)
	}
	if c.Kind != "finish" || c.Seat != 0 || c.GameID != l.GameID {
		t.Fatal(c)
	}
	n, e := l.FinishSettlement(c.Seat)
	if e != nil || !n.Finished[0] || l.Finished[0] {
		t.Fatal("finish not independently legal and immutable")
	}
}
func TestFinanceProjectionAndWait(t *testing.T) {
	l := game.NewFinancialLedger("finance", 3)
	l, e := l.Charge("other", 1, 2, game.IntAmount(1))
	if e != nil {
		t.Fatal(e)
	}
	a, _ := ProjectLedger(l, 0)
	if len(a.Debts) != 0 {
		t.Fatal("unrelated debt projection")
	}
	if _, e = ChooseFinance("heuristic@v1", a, 10); e != ErrUnsupported {
		t.Fatal("playing fabricated finish")
	}
	l.Phase = "settlement"
	a, _ = ProjectLedger(l, 0)
	a.Legal[0].Seat = 2
	if _, e = ChooseFinance("heuristic@v1", a, 10); e == nil {
		t.Fatal("altered seat")
	}
	a, _ = ProjectLedger(l, 0)
	if _, e = ChooseFinance("heuristic@v1", a, 0); e != ErrBudget {
		t.Fatal("ignored budget")
	}
	l, _ = l.FinishSettlement(0)
	a, _ = ProjectLedger(l, 0)
	if len(a.Legal) != 0 {
		t.Fatal("finished seat menu")
	}
}
