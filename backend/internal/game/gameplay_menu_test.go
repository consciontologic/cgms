package game

import (
	"context"
	"errors"
	"testing"
)

type cancelDuringMenu struct {
	context.Context
	cancel context.CancelFunc
	checks int
}

func (c *cancelDuringMenu) Err() error {
	c.checks++
	if c.checks == 10 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestGameplayMenuCancellationDropsPartialMenuWithoutMutation(t *testing.T) {
	s := formationFixture(t)
	before := Digest(s)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	probe := &cancelDuringMenu{Context: ctx, cancel: cancel}
	o, err := GameplayObservationContext(probe, s, 1, 100000)
	if probe.checks < 10 || !errors.Is(err, context.Canceled) || len(o.Legal) != 0 || Digest(s) != before {
		t.Fatal("canceled enumeration returned a partial menu or mutated its state", probe.checks, err)
	}
	want, err := GameplayObservation(s, 1, 100000)
	if err != nil {
		t.Fatal(err)
	}
	got, err := GameplayObservationContext(context.Background(), s, 1, 100000)
	if err != nil || Digest(got) != Digest(want) {
		t.Fatal("context path changed the uncanceled menu", err)
	}
}

func TestGameplayMenuOpeningAndExplicitEnd(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-spades-03", "deck-1-clubs-13"}, 1, Hand, false)
	o, e := GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	end, open, attach := false, false, false
	for _, c := range o.Legal {
		switch c.Kind {
		case "end-turn":
			end = true
		case "open-series":
			open = true
		case "attach":
			attach = true
		}
	}
	if !end || !open || !attach {
		t.Fatalf("missing categories end=%v open=%v attach=%v", end, open, attach)
	}
}
func TestGameplayMenuBudgetNeverFabricatesPass(t *testing.T) {
	s := formationFixture(t)
	o, e := GameplayObservation(s, 1, 1)
	if e != ErrMenuBudget || len(o.Legal) != 0 {
		t.Fatal("budget generated partial legal list")
	}
}
func TestGameplayMenuHiddenCardPermutation(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-2-spades-04", "deck-2-hearts-12"}, 2, Hand, false)
	n := s.Clone()
	for i := range n.Cards {
		if n.Cards[i].Card.ID == "deck-2-spades-04" {
			n.Cards[i].Controller = 3
		}
		if n.Cards[i].Card.ID == "deck-2-hearts-12" {
			n.Cards[i].Controller = 3
		}
	} // Reassigning unseen equal-count hands must not affect seat-one menu.
	a, e := GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	b, e := GameplayObservation(n, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	if Digest(a.Legal) != Digest(b.Legal) {
		t.Fatal("hidden custody leaked through menu")
	}
}

func TestGameplayMenuIncludesOfferedProposalResponses(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-hearts-03"}, 1, Hand, false)
	s, _, e := Apply(s, Command{GameID: s.GameID, ID: "proposal", Kind: "offer", Actor: 1, OfferID: "gift", Revision: 1, Terms: &ProposalTerms{To: 2, Give: []string{"deck-1-hearts-03"}}})
	if e != nil {
		t.Fatal(e)
	}
	o, e := GameplayObservation(s, 2, 100000)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, c := range o.Legal {
		if c.Kind == "accept-offer" {
			found = true
		}
	}
	if !found {
		t.Fatal("offered proposal omitted")
	}
}

func TestGameplayObservationEquivalentHiddenRoyalSwap(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-2-hearts-12"}, 2, Hand, false)
	s, _ = s.Move([]string{"deck-2-spades-03"}, 3, Hand, false)
	n := s.Clone()
	for i := range n.Cards {
		if n.Cards[i].Card.ID == "deck-2-hearts-12" {
			n.Cards[i].Controller = 3
		}
		if n.Cards[i].Card.ID == "deck-2-spades-03" {
			n.Cards[i].Controller = 2
		}
	}
	a, e := GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	b, e := GameplayObservation(n, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	if Digest(a) != Digest(b) {
		t.Fatal("opponent hidden rank swap changed authorized observation")
	}
}

func TestGameplayLoanProjectionIsPartyOnlyHistoricalTerms(t *testing.T) {
	s := formationFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-11", "deck-2-clubs-11"}, 2, Hand, false)
	s.Loans = []FormationLoan{{Cards: []string{"deck-1-clubs-11", "deck-2-clubs-11"}, From: 1, To: 2, Allocation: "old-formation"}}
	a, e := GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	b, e := GameplayObservation(s, 3, 100000)
	if e != nil {
		t.Fatal(e)
	}
	if len(a.Loans) != 1 || len(b.Loans) != 0 {
		t.Fatal("loan party projection")
	}
	for _, c := range a.Cards {
		if c.Card.ID == "deck-1-clubs-11" {
			t.Fatal("historical terms exposed current concealed custody")
		}
	}
	a.Loans[0].Cards[0] = "tamper"
	if s.Loans[0].Cards[0] == "tamper" {
		t.Fatal("loan projection aliases state")
	}
}

// This is an audit of the frozen sampler's limitation, not a desired exhaustive
// menu contract. Same-total physical payments have different remaining cards.
func TestSampledMenuAuditDistinctUsefulPayments(t *testing.T) {
	s := formationFixture(t)
	var err error
	ids := []string{"deck-1-spades-02", "deck-1-spades-03", "deck-1-spades-05"}
	s, err = s.Move(ids, 1, Hand, false)
	if err != nil {
		t.Fatal(err)
	}
	alternatives := [][]string{{ids[0], ids[1]}, {ids[2]}}
	for i, cards := range alternatives {
		_, _, err := Apply(s, Command{GameID: s.GameID, ID: []string{"audit-a", "audit-b"}[i], Actor: 1, Kind: "purchase", Cards: cards, Value: 1})
		if err != nil {
			t.Fatalf("expected independently legal payment: %v", err)
		}
	}
	o, err := GameplayObservation(s, 1, 100000)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, c := range o.Legal {
		if c.Kind == "purchase" && c.Value == 1 && observationTotal(o, c.Cards).Cmp(IntAmount(5)) == 0 {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("frozen sampler changed: got %d exact-five payment alternatives, expected one of two legal alternatives", count)
	}
	t.Log("sampled-all-categories-v2: 2 legal exact-five payments, 1 represented; one-card payment preserves two future cards versus the two-card payment")
}
