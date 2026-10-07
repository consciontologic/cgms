package offlineprobe

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"os"
	"path/filepath"
	"testing"
)

func financeFixture(seats int) FinanceRequest {
	r := FinanceRequest{Schema: FinanceVersion, Ledger: game.NewFinancialLedger("finance-1", seats), GameLimit: 3}
	add := func(c FinanceCommand) {
		if c.GameID == "" {
			c.GameID = fmt.Sprintf("finance-%d", 1+countNext(r.Commands))
		}
		r.Commands = append(r.Commands, c)
	}
	receipt := func(seat int, n int64) game.FinancialReceipt {
		return game.FinancialReceipt{Seat: seat, Amount: game.IntAmount(n)}
	}
	add(FinanceCommand{Kind: "charge", ID: "old-player", Seat: 0, Other: 1, Amount: game.IntAmount(10)})
	add(FinanceCommand{Kind: "receipt", Receipts: []game.FinancialReceipt{receipt(0, 6)}})
	add(FinanceCommand{Kind: "close"})
	finish := func() {
		for s := 0; s < seats; s++ {
			add(FinanceCommand{Kind: "finish", Seat: s})
		}
		add(FinanceCommand{Kind: "finalize"})
	}
	finish()
	add(FinanceCommand{Kind: "next", NextGame: "finance-2"})
	add(FinanceCommand{Kind: "forgive", ID: "cross-forgive", Debt: "old-player", Amount: game.IntAmount(2), Consent: []int{1}})
	// Equal Ponzi halves are computed before recipient debts (Q10), not rounded.
	add(FinanceCommand{Kind: "charge", ID: "system", Seat: 0, Other: -1, Amount: game.IntAmount(1)})
	add(FinanceCommand{Kind: "receipt", Receipts: []game.FinancialReceipt{receipt(2, 3)}})
	add(FinanceCommand{Kind: "ponzi", Seat: 2, Other: 0, Partner: 1})
	add(FinanceCommand{Kind: "coup", Seat: 0})
	finish()
	add(FinanceCommand{Kind: "next", NextGame: "finance-3"})
	add(FinanceCommand{Kind: "receipt", Receipts: []game.FinancialReceipt{receipt(1, 50), receipt(2, 50)}})
	add(FinanceCommand{Kind: "close"})
	finish()
	return r
}
func countNext(cs []FinanceCommand) int {
	n := 0
	for _, c := range cs {
		if c.Kind == "next" {
			n++
		}
	}
	return n
}
func financeInvoke(t *testing.T, r FinanceRequest) FinanceResponse {
	t.Helper()
	raw, e := canonical.Marshal(r)
	if e != nil {
		t.Fatal(e)
	}
	var out FinanceResponse
	if e = canonical.Decode(Execute(raw), &out); e != nil {
		t.Fatal(e)
	}
	return out
}
func TestFinanceThreeGameDifferentialFixtures(t *testing.T) {
	for _, seats := range []int{3, 4} {
		r := financeFixture(seats)
		before := game.Digest(r)
		out := financeInvoke(t, r)
		if out.Error != "" || len(out.Frames) != len(r.Commands)+1 {
			t.Fatalf("finance trace failed: %+v", out)
		}
		if before != game.Digest(r) {
			t.Fatal("mutated request")
		}
		last := out.Frames[len(out.Frames)-1]
		if !last.Complete || len(last.Ledger.Completed) != 3 || last.Ledger.Phase != "finalized" {
			t.Fatal("not a completed financial match")
		}
		// §4.2: first game's 6 receipt pays creditor; debtor has -4, creditor +6.
		if last.Ledger.Completed[0].Scores[0].String() != "-4" || last.Ledger.Completed[0].Scores[1].String() != "6" {
			t.Fatal("first result changed")
		}
		// Cross-game forgiveness is nonspendable and survives Coup replacement (N06).
		if len(last.Ledger.Corrections) != 1 || !last.Ledger.Corrections[0].MatchLevel || last.Ledger.Corrections[0].Amount.String() != "2" {
			t.Fatal("lost match correction")
		}
		// Coup owner repays the remaining 1/2 player debt as a fresh receipt.
		if last.Totals[0].String() != "47" || last.Totals[1].String() != "13/2" || last.Totals[2].String() != "0" {
			t.Fatalf("match totals %+v", last.Totals)
		}
		for _, f := range out.Frames {
			for seat, o := range f.Observations {
				for _, d := range o.Debts {
					if d.Debtor != seat && d.Creditor != seat {
						t.Fatal("unrelated debt leaked")
					}
				}
			}
		}
		// Resume every serialized financial frame; retain exact immutable history.
		for i, f := range out.Frames {
			rest := r
			rest.Ledger = f.Ledger
			rest.Commands = r.Commands[i:]
			resumed := financeInvoke(t, rest)
			if resumed.Error != "" || game.Digest(resumed.Frames[len(resumed.Frames)-1]) != game.Digest(last) {
				t.Fatalf("resume %d failed: %s", i, resumed.Error)
			}
		}
		if dir := os.Getenv("CGMS_PROBE_FIXTURES"); dir != "" {
			raw, _ := canonical.Marshal(r)
			expected := Execute(raw)
			for suffix, b := range map[string][]byte{"request": raw, "expected": expected} {
				if e := os.WriteFile(filepath.Join(dir, fmt.Sprintf("finance-%d.%s.json", seats, suffix)), b, 0600); e != nil {
					t.Fatal(e)
				}
			}
		}
	}
}
func TestFinanceBoundsAndAtomicRejection(t *testing.T) {
	r := financeFixture(3)
	r.Commands = []FinanceCommand{{GameID: "wrong", Kind: "receipt"}}
	if o := financeInvoke(t, r); o.Error != "command_rejected" || o.FailedStep != 0 || len(o.Frames) != 0 {
		t.Fatal("wrong game accepted")
	}
	r = financeFixture(3)
	r.Commands = append(r.Commands, FinanceCommand{GameID: "finance-3", Kind: "next", NextGame: "finance-4"})
	if o := financeInvoke(t, r); o.Error != "command_rejected" || len(o.Frames) != 0 {
		t.Fatal("fixed match limit bypass")
	}
	r = financeFixture(3)
	r.Commands = []FinanceCommand{{GameID: "finance-1", Kind: "finalize"}}
	if o := financeInvoke(t, r); o.Error != "command_rejected" {
		t.Fatal("early finalization accepted")
	}
	r = financeFixture(3)
	r.GameLimit = 0
	if o := financeInvoke(t, r); o.Error != "invalid_request" {
		t.Fatal("missing limit")
	}
}

// Q10's reciprocal thirds are accounting (not an illegal three-member alliance).
func TestFinanceFIFOExactCycleAndLargeRationals(t *testing.T) {
	third, _ := game.NewAmount("1", "3")
	huge, _ := game.NewAmount("9007199254740993", "2")
	for name, a := range map[string]game.Amount{"cycle": third, "large": huge} {
		r := FinanceRequest{Schema: FinanceVersion, Ledger: game.NewFinancialLedger("exact", 3), GameLimit: 1}
		if name == "cycle" {
			r.Commands = []FinanceCommand{
				{GameID: "exact", Kind: "charge", ID: "ab", Seat: 0, Other: 1, Amount: game.IntAmount(1)},
				{GameID: "exact", Kind: "charge", ID: "ba", Seat: 1, Other: 0, Amount: game.IntAmount(1)},
			}
		}
		r.Commands = append(r.Commands, FinanceCommand{GameID: "exact", Kind: "receipt", Receipts: []game.FinancialReceipt{{Seat: 0, Amount: a}}})
		out := financeInvoke(t, r)
		if out.Error != "" {
			t.Fatal(out.Error)
		}
		last := out.Frames[len(out.Frames)-1].Ledger
		if last.Cash[0].Cmp(a) != 0 || last.Scores[0].Cmp(a) != 0 {
			t.Fatal("exact cash or no-second-debit score failed")
		}
		if name == "cycle" && (!last.Debts[0].Remaining.IsZero() || !last.Debts[1].Remaining.IsZero() || !last.Scores[1].IsZero()) {
			t.Fatal("reciprocal debt was not settled exactly")
		}
		exportFinance(t, name, r)
	}
}
func exportFinance(t *testing.T, name string, r FinanceRequest) {
	t.Helper()
	if dir := os.Getenv("CGMS_PROBE_FIXTURES"); dir != "" {
		raw, e := canonical.Marshal(r)
		if e != nil {
			t.Fatal(e)
		}
		for suffix, b := range map[string][]byte{"request": raw, "expected": Execute(raw)} {
			if e := os.WriteFile(filepath.Join(dir, "finance-"+name+"."+suffix+".json"), b, 0600); e != nil {
				t.Fatal(e)
			}
		}
	}
}

func TestFinanceSameGameForgivenessAndNullification(t *testing.T) {
	start := game.NewFinancialLedger("same", 3)
	r := FinanceRequest{Schema: FinanceVersion, Ledger: start, GameLimit: 1, Commands: []FinanceCommand{
		{GameID: "same", Kind: "charge", ID: "charge", Seat: 1, Other: 0, Amount: game.IntAmount(10)},
		{GameID: "same", Kind: "receipt", Receipts: []game.FinancialReceipt{{Seat: 1, Amount: game.IntAmount(6)}}},
		{GameID: "same", Kind: "forgive", ID: "forgive", Debt: "charge", Amount: game.IntAmount(2), Consent: []int{0}},
		{GameID: "same", Kind: "forgive", ID: "forgive", Debt: "charge", Amount: game.IntAmount(2), Consent: []int{0}},
		{GameID: "same", Kind: "coup", Seat: 0},
	}}
	out := financeInvoke(t, r)
	if out.Error != "" {
		t.Fatal(out.Error)
	}
	last := out.Frames[len(out.Frames)-1].Ledger
	if last.Scores[0].String() != "50" || last.Scores[1].String() != "-52" || len(last.Corrections) != 0 || last.Debts[0].Remaining.String() != "2" {
		t.Fatal("same-game forgiveness or Coup double charge")
	}
	exportFinance(t, "same-game", r)
	r.Commands = append(r.Commands, FinanceCommand{GameID: "same", Kind: "nullify", Start: &start, Consent: []int{0, 1, 2}})
	out = financeInvoke(t, r)
	if out.Error != "" {
		t.Fatal(out.Error)
	}
	last = out.Frames[len(out.Frames)-1].Ledger
	if last.Phase != "void" || len(last.Debts) != 0 || len(last.Corrections) != 0 || len(last.Completed) != 0 {
		t.Fatal("nullification did not restore start")
	}
	exportFinance(t, "nullification", r)
}
