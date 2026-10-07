package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestFinancialRandomV2ClassCardinalityIndependence(t *testing.T) {
	for _, offers := range []int{1, 100} {
		menu := []game.LifecycleOperation{{Kind: "decline-financial-opportunity"}, {Kind: "promise-accept"}}
		for i := 0; i < offers; i++ {
			menu = append(menu, game.LifecycleOperation{Kind: "promise-offer"})
		}
		for class, want := range []string{"decline-financial-opportunity", "promise-accept", "promise-offer"} {
			// n=3 rejection threshold1: 3,4,5 select indices0,1,2.
			words := []uint64{uint64(3 + class), 100}
			cursor := 0
			pick, used, err := sampleFinancialKinds(menu, func() uint64 { v := words[cursor]; cursor++; return v }, len(menu)+2)
			if err != nil || menu[pick].Kind != want || used != len(menu)+2 {
				t.Fatal("class probability depends on entry cardinality", offers, want, used, err)
			}
		}
	}
}
func TestFinancialRandomV2BudgetCannotFabricateChoice(t *testing.T) {
	menu := []game.LifecycleOperation{{Kind: "promise-offer"}, {Kind: "decline-financial-opportunity"}}
	calls := 0
	_, _, err := sampleFinancialKinds(menu, func() uint64 { calls++; return 2 }, len(menu)+1)
	if err != ErrBudget || calls != 1 {
		t.Fatal("missing second draw must be explicit budget exhaustion", err, calls)
	}
}
