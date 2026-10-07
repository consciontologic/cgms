package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"testing"
)

func TestOpportunityRoyalRetentionAblation(t *testing.T) {
	s, _ := game.NewState(3, "retain")
	s, _ = s.Move([]string{"deck-1-clubs-02", "deck-1-hearts-02"}, 1, game.Series, false)
	s.Players[0].History = []game.Suit{game.Clubs, game.Hearts}
	s, _ = s.Move([]string{"deck-1-clubs-11"}, 1, game.Hand, false)
	o, e := game.GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	for policy, want := range map[string]string{"opportunity@v1": "end-turn", "opportunity-no-retain@v1": "attach", "economic@v1": "attach"} {
		d, e := ChooseGameplay(policy, o, nil, 100000)
		if e != nil || d.Command.Kind != want {
			t.Fatalf("%s: want %s got %s %v", policy, want, d.Command.Kind, e)
		}
	}
	s.Round = 12
	o, e = game.GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	d, e := ChooseGameplay("opportunity@v1", o, nil, 100000)
	if e != nil || d.Command.Kind != "attach" {
		t.Fatal("late horizon should release retained royal", e)
	}
}

func TestOpportunityTakesFreeOrdinaryGiftButNotAllianceLoan(t *testing.T) {
	for _, loan := range []bool{false, true} {
		o := game.Observation{GameID: "g", Seat: 2, Proposals: []game.Proposal{{ID: "p", From: 1, Terms: game.ProposalTerms{To: 2, Give: []string{"authorized-private-card"}, Loan: loan}}}, Legal: []game.Command{{GameID: "g", Actor: 2, Kind: "accept-offer", OfferID: "p"}, {GameID: "g", Actor: 2, Kind: "decline-offer", OfferID: "p"}}}
		d, e := ChooseGameplay("opportunity@v1", o, nil, 100)
		want := "accept-offer"
		if loan {
			want = "decline-offer"
		}
		if e != nil || d.Command.Kind != want {
			t.Fatalf("loan %v: %s %v", loan, d.Command.Kind, e)
		}
	}
}

func TestOpportunityFinanceChoosesLargestAffordableReducedOffer(t *testing.T) {
	o := GameplayFinanceObservation{GameID: "g", Seat: 1, Cash: game.IntAmount(7), Legal: []game.LifecycleOperation{{GameID: "g", Actor: 1, Kind: "promise-final-offer", Amount: game.IntAmount(4)}, {GameID: "g", Actor: 1, Kind: "promise-final-offer", Amount: game.IntAmount(7)}, {GameID: "g", Actor: 1, Kind: "promise-final-offer", Amount: game.IntAmount(9)}, {GameID: "g", Actor: 1, Kind: "promise-refuse"}}}
	d, e := ChooseGameplayFinancial("opportunity@v1", o, nil, 100)
	if e != nil || d.Operation.Kind != "promise-final-offer" || d.Operation.Amount.Cmp(game.IntAmount(7)) != 0 {
		t.Fatal("must use exact affordable maximum to retain reputation opportunity", e, d.Operation.Kind)
	}
	o.Debts = []game.FinancialDebt{{Debtor: 0, Remaining: game.IntAmount(1)}}
	d, e = ChooseGameplayFinancial("opportunity@v1", o, nil, 100)
	if e != nil || d.Operation.Kind != "promise-refuse" {
		t.Fatal("must not offer blocked payment", e)
	}
}

func TestOpportunityLatePurchaseUsesNonScoringSpades(t *testing.T) {
	o := game.Observation{GameID: "g", Seat: 1, Round: 12, Legal: []game.Command{{GameID: "g", Actor: 1, Kind: "purchase", Value: 1}, {GameID: "g", Actor: 1, Kind: "purchase", Value: 2}, {GameID: "g", Actor: 1, Kind: "end-turn"}}}
	d, e := ChooseGameplay("opportunity@v1", o, nil, 100)
	if e != nil || d.Command.Kind != "purchase" || d.Command.Value != 2 {
		t.Fatal("missed larger legal late acquisition", e)
	}
}

func TestOpportunityObservationMenuPrivacyAndImmutability(t *testing.T) {
	s, _ := game.NewState(3, "private-opportunity")
	s, _ = s.Move([]string{"deck-1-clubs-02", "deck-1-hearts-02"}, 1, game.Series, false)
	s.Players[0].History = []game.Suit{game.Clubs, game.Hearts}
	s, _ = s.Move([]string{"deck-1-clubs-11"}, 1, game.Hand, false)
	s, _ = s.Move([]string{"deck-2-hearts-12"}, 2, game.Hand, false)
	s, _ = s.Move([]string{"deck-2-spades-03"}, 3, game.Hand, false)
	n := s.Clone()
	for i := range n.Cards {
		switch n.Cards[i].Card.ID {
		case "deck-2-hearts-12":
			n.Cards[i].Controller = 3
		case "deck-2-spades-03":
			n.Cards[i].Controller = 2
		}
	}
	a, e := game.GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	b, e := game.GameplayObservation(n, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	if game.Digest(a) != game.Digest(b) || game.Digest(a.Legal) != game.Digest(b.Legal) {
		t.Fatal("hidden swap leaked through observation or menu")
	}
	before := game.Digest(a)
	for _, policy := range []string{"opportunity@v1", "opportunity-no-retain@v1"} {
		x, e := ChooseGameplay(policy, a, nil, 100000)
		if e != nil {
			t.Fatal(e)
		}
		y, e := ChooseGameplay(policy, b, nil, 100000)
		if e != nil || game.Digest(x) != game.Digest(y) {
			t.Fatal("private swap changed decision", e)
		}
		if game.Digest(a) != before {
			t.Fatal("policy mutated observation")
		}
		if _, e = ChooseGameplay(policy, a, nil, len(a.Legal)); e != ErrBudget {
			t.Fatal("candidate bypassed score budget", e)
		}
	}
}

func TestOpportunityFinancialPrivacyAndRefusalAblation(t *testing.T) {
	l := triggeredFinance(t, 10, 10)
	a, e := ProjectGameplayFinance(l, 1)
	if e != nil {
		t.Fatal(e)
	}
	n := l.Clone()
	n.Ledger, e = n.Ledger.Charge("other-private", 2, 1, game.IntAmount(17))
	if e != nil {
		t.Fatal(e)
	}
	b, e := ProjectGameplayFinance(n, 1)
	if e != nil {
		t.Fatal(e)
	}
	if game.Digest(a) != game.Digest(b) {
		t.Fatal("unrelated debt or legality leaked")
	}
	for policy, want := range map[string]string{"opportunity@v1": "promise-pay", "opportunity-no-retain@v1": "promise-pay", "pressure@v1": "promise-refuse"} {
		x, e := ChooseGameplayFinancial(policy, a, nil, 10000)
		if e != nil || x.Operation.Kind != want {
			t.Fatal(policy, want, e)
		}
		y, e := ChooseGameplayFinancial(policy, b, nil, 10000)
		if e != nil || game.Digest(x) != game.Digest(y) {
			t.Fatal("private finance changed decision", e)
		}
	}
}

func TestOpportunityCarriedDebtChangesIncomePriorityOnlyWithFutureGames(t *testing.T) {
	o := game.Observation{GameID: "g", Seat: 1, Legal: []game.Command{{GameID: "g", Actor: 1, Kind: "ringleader"}, {GameID: "g", Actor: 1, Kind: "open-series", Suit: game.Clubs}}}
	ctx := GameplayEconomy{Cash: game.IntAmount(0), Owed: game.IntAmount(20), FutureGames: 2}
	for policy, want := range map[string]string{"opportunity@v1": "ringleader", "opportunity-no-finance@v1": "open-series", "economic@v1": "open-series"} {
		d, e := ChooseGameplayWithEconomy(policy, o, ctx, nil, 100)
		if e != nil || d.Command.Kind != want {
			t.Fatal(policy, want, d.Command.Kind, e)
		}
	}
	ctx.FutureGames = 0
	d, e := ChooseGameplayWithEconomy("opportunity@v1", o, ctx, nil, 100)
	if e != nil || d.Command.Kind != "open-series" {
		t.Fatal("future-debt priority improperly survives final game", e)
	}
}

func TestOpportunityManyPromisesDoNotDisplaceRequiredPayment(t *testing.T) {
	o := GameplayFinanceObservation{GameID: "g", Seat: 1, Cash: game.IntAmount(100), Legal: []game.LifecycleOperation{{GameID: "g", Actor: 1, Kind: "promise-pay", PromiseID: "pay"}, {GameID: "g", Actor: 1, Kind: "promise-final-offer", PromiseID: "reduce", Amount: game.IntAmount(90)}}}
	for i := 0; i < 60; i++ {
		o.Legal = append(o.Legal, game.LifecycleOperation{GameID: "g", Actor: 1, Kind: "promise-final-offer", PromiseID: "unrelated", Amount: game.IntAmount(1)})
	}
	d, e := ChooseGameplayFinancial("opportunity@v1", o, nil, 1000)
	if e != nil || d.Operation.Kind != "promise-pay" {
		t.Fatal("unrelated menu multiplicity displaced payment", e)
	}
}
