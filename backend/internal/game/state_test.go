package game

import (
	"reflect"
	"testing"
)

func cid(s Suit, r int) string {
	for _, c := range Deck() {
		if c.Deck == 1 && c.Suit == s && c.Rank == r {
			return c.ID
		}
	}
	panic("bad fixture")
}
func position(t *testing.T) State {
	t.Helper()
	s, e := NewState(3, "fixture-g1")
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestStatePartitionAndMovement(t *testing.T) {
	s := position(t)
	if e := s.Validate(); e != nil {
		t.Fatal(e)
	}
	before := s.Clone()
	ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
	n, e := s.Move(ids, 1, Formation, true)
	if e != nil {
		t.Fatal(e)
	}
	loan, e := n.Move(ids, 2, Formation, true)
	if e != nil {
		t.Fatal(e)
	}
	back, e := loan.Move(ids[:1], 1, Unassigned, true)
	if e != nil {
		t.Fatal(e)
	}
	for _, x := range []State{n, loan, back} {
		if e = x.Validate(); e != nil {
			t.Fatal(e)
		}
	}
	if !reflect.DeepEqual(s, before) {
		t.Fatal("predecessor changed")
	}
	c, _ := back.Card(ids[1])
	if c.Controller != 2 {
		t.Fatal("remote survivor reclaimed")
	}
	broken := back.Clone()
	broken.Cards[0] = broken.Cards[1]
	if broken.Validate() == nil {
		t.Fatal("duplicate physical card accepted")
	}
	broken = back.Clone()
	broken.Cards = broken.Cards[:103]
	if broken.Validate() == nil {
		t.Fatal("missing physical card accepted")
	}
}
func TestBindingsAvailabilityAndFate(t *testing.T) {
	s := position(t)
	ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
	s, _ = s.Move(ids, 1, Formation, true)
	var e error
	s, e = s.Bind(Binding{"b1", ids, "queen-hearts", "people", true})
	if e != nil {
		t.Fatal(e)
	}
	if s.BindingStatus("b1") != "active" {
		t.Fatal("binding not active")
	}
	if _, e = s.Bind(Binding{"b2", []string{ids[0], cid(Spades, 13)}, "queen-clubs", "justice", false}); e == nil {
		t.Fatal("third binding accepted")
	}
	s, _ = s.Move(ids[:1], 2, Hand, true)
	if s.BindingStatus("b1") != "dormant-separated" {
		t.Fatal("split lost binding")
	}
	s, _ = s.Move(ids, 1, Hand, true)
	if s.BindingStatus("b1") != "dormant-together" {
		t.Fatal("reunion activated free")
	}
	for i := range s.Cards {
		if s.Cards[i].Card.ID == ids[0] {
			s.Cards[i].AvailableFromRound = 2
		}
	}
	if _, e = s.Move(ids[:1], 2, Hand, false); e == nil {
		t.Fatal("restricted voluntary move allowed")
	}
	s, _ = s.Move([]string{cid(Spades, 1)}, 2, ConcealedAce, true)
	s, e = s.ActivateEffect(AceEffect{"inflation1", cid(Spades, 1), "inflation", 2, 1, 1, "qualifying-transaction-or-target-departure-or-closure", 0})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Move([]string{cid(Spades, 1)}, 1, Hand, false); e == nil {
		t.Fatal("active ace reusable")
	}
	n, count, e := s.FateReturn(1)
	if e != nil || count != 2 {
		t.Fatalf("Fate personal count=%d err=%v", count, e)
	}
	c, _ := n.Card(cid(Spades, 1))
	if c.Zone != ActiveAce || !reflect.DeepEqual(s.Effects, n.Effects) {
		t.Fatal("Fate reset active effect")
	}
	c, _ = n.Card(ids[0])
	if c.AvailableFromRound != 2 {
		t.Fatal("restriction lost")
	}
	if len(n.Bindings) != 1 {
		t.Fatal("Fate lost binding")
	}
	if e = n.Validate(); e != nil {
		t.Fatal(e)
	}
}
func TestStateCloneNoAliasing(t *testing.T) {
	s := position(t)
	s.Players[0].History = []Suit{Diamonds}
	s.Players[0].Quotas["fate"] = 1
	s.Bindings = []Binding{{"b", []string{"a", "b"}, "slot", "formation", false}}
	s.Effects = []AceEffect{{ID: "e"}}
	s.PublicHistory = []string{"known"}
	n := s.Clone()
	n.Cards[0].Controller = 2
	n.Players[0].History[0] = Hearts
	n.Players[0].Quotas["fate"] = 99
	n.Bindings[0].Kings[0] = "changed"
	n.Effects[0].ID = "changed"
	n.PublicHistory[0] = "changed"
	if s.Cards[0].Controller != 0 || s.Players[0].History[0] != Diamonds || s.Players[0].Quotas["fate"] != 1 || s.Bindings[0].Kings[0] != "a" || s.Effects[0].ID != "e" || s.PublicHistory[0] != "known" {
		t.Fatal("mutable aliases escape")
	}
}
func FuzzStatePartition(f *testing.F) {
	f.Add(uint8(3), uint8(7))
	f.Fuzz(func(t *testing.T, seat, rank uint8) {
		s := position(t)
		n, e := s.Move([]string{cid(Hearts, int(rank%9)+2)}, int(seat%3)+1, Hand, true)
		if e != nil {
			t.Fatal(e)
		}
		if e = n.Validate(); e != nil {
			t.Fatal(e)
		}
		if s.Cards[0].Controller != 0 {
			t.Fatal("aliased")
		}
	})
}
func TestBindingReunionNeedsExplicitAllocation(t *testing.T) {
	s := position(t)
	ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
	s, _ = s.Move(ids, 1, Formation, true)
	s, _ = s.Bind(Binding{"b1", ids, "queen-hearts", "people", true})
	s, _ = s.Move(ids[:1], 2, Hand, true)
	s, _ = s.Move(ids, 1, Formation, true)
	if s.BindingStatus("b1") == "active" {
		t.Fatal("reunion grants free activation")
	}
}
func TestFateReturnRequiresShuffleAndPreservesMarkerOnRedraw(t *testing.T) {
	s := position(t)
	id := cid(Hearts, 13)
	s, _ = s.Move([]string{id}, 1, Hand, true)
	for i := range s.Cards {
		if s.Cards[i].Card.ID == id {
			s.Cards[i].AvailableFromRound = 2
		}
	}
	n, _, e := s.FateReturn(1)
	if e != nil {
		t.Fatal(e)
	}
	if _, _, e = DrawCards(n, 1, 1, false, nil); e == nil {
		t.Fatal("unshuffled Fate return can draw")
	}
	order := append([]string{id}, n.DrawOrder[:len(n.DrawOrder)-1]...)
	n, e = ShuffleSupply(n, order)
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = DrawCards(n, 1, 1, false, nil)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card(id)
	if c.AvailableFromRound != 2 {
		t.Fatal("redraw cleared physical marker")
	}
}
func TestIntactLoanPreservesBindingAllocation(t *testing.T) {
	s := position(t)
	ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
	s, _ = s.Move(ids, 1, Formation, true)
	s, _ = s.Move([]string{cid(Hearts, 2)}, 1, Series, true)
	s, _ = s.Move([]string{cid(Hearts, 3)}, 2, Series, true)
	s, _ = s.Bind(Binding{"b", ids, "queen-hearts", "people", true})
	n, e := s.LoanFormation(ids, 1, 2)
	if e != nil {
		t.Fatal(e)
	}
	if n.BindingStatus("b") != "active" {
		t.Fatal("intact loan erased allocation")
	}
	n, e = n.ReturnLoan(ids, 2, 1)
	if e != nil || n.BindingStatus("b") != "active" {
		t.Fatal("intact return erased allocation", e)
	}
}
func TestAssassinationReleasesRemoteBinding(t *testing.T) {
	s := position(t)
	ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
	s, _ = s.Move(ids, 1, Formation, true)
	s, _ = s.Bind(Binding{"b", ids, "queen-hearts", "people", true})
	s, _ = s.Move(ids[1:], 2, Hand, true)
	n, e := s.AssassinateBinding(ids[0])
	if e != nil {
		t.Fatal(e)
	}
	remote, _ := n.Card(ids[1])
	dead, _ := n.Card(ids[0])
	if len(n.Bindings) != 0 || remote.Controller != 2 || remote.Zone != Hand || dead.Zone != Discard {
		t.Fatal("assassination failed release or moved remote survivor")
	}
}
func TestSourceDepartureKeepsTargetInflation(t *testing.T) {
	s := position(t)
	s, _ = s.Move([]string{cid(Spades, 1)}, 1, ConcealedAce, true)
	s, effectErr := s.ActivateEffect(AceEffect{ID: "i", CardID: cid(Spades, 1), Kind: "inflation", Source: 1, Target: 2, Custodian: 2, Expiry: "qualifying-transaction-or-target-departure-or-closure"})
	if effectErr != nil {
		t.Fatal(effectErr)
	}
	n, e := s.DepartPlayer(1)
	if e != nil {
		t.Fatal(e)
	}
	if !n.Players[0].Departed || len(n.Effects) != 1 {
		t.Fatal("source departure effect lifecycle")
	}
	n, e = n.DepartPlayer(2)
	if e != nil {
		t.Fatal(e)
	}
	a, _ := n.Card(cid(Spades, 1))
	if len(n.Effects) != 0 || a.Zone != Discard {
		t.Fatal("target departure keeps effect")
	}
}
func TestCapturedDiamondForcedExposureKeepsRestriction(t *testing.T) {
	s := position(t)
	id := cid(Diamonds, 7)
	s, _ = s.Move([]string{id}, 2, Hand, true)
	for i := range s.Cards {
		if s.Cards[i].Card.ID == id {
			s.Cards[i].AvailableFromRound = 3
		}
	}
	n, e := s.AcquireCaptured([]string{id}, 1)
	if e != nil {
		t.Fatal(e)
	}
	c, _ := n.Card(id)
	if c.Zone != Series || c.Controller != 1 || c.AvailableFromRound != 3 {
		t.Fatal("capture must expose restricted Diamond without lifting marker")
	}
}
func TestLoanDormancyPartialReturnAndRemoteCustody(t *testing.T) {
	s := position(t)
	ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
	s, _ = s.Move(ids, 1, Formation, true)
	s, _ = s.Move([]string{cid(Hearts, 2)}, 1, Series, true)
	s, _ = s.Bind(Binding{"b", ids, "queen-hearts", "people", true})
	n, e := s.LoanFormation(ids, 1, 2)
	if e != nil {
		t.Fatal(e)
	}
	if n.BindingStatus("b") != "dormant-together" {
		t.Fatal("loan fabricated missing support")
	}
	if _, e = n.Move(ids[:1], 2, Attachment, true); e == nil {
		t.Fatal("dormant bound king attached")
	}
	n, _ = n.Move(ids[1:], 3, Hand, true)
	n, e = n.ReturnLoan(ids[:1], 2, 1)
	if e != nil {
		t.Fatal(e)
	}
	returned, _ := n.Card(ids[0])
	remote, _ := n.Card(ids[1])
	if returned.Zone != Unassigned || remote.Controller != 3 || remote.Zone != Hand || n.BindingStatus("b") != "dormant-separated" {
		t.Fatal("partial return rebuilt formation or reclaimed remote card")
	}
	view, _ := Observe(n, 1)
	for _, c := range view.Cards {
		if c.Card.ID == ids[1] {
			t.Fatal("remote hidden survivor disclosed")
		}
	}
	if e = n.Validate(); e != nil {
		t.Fatal(e)
	}
}
func TestAssassinationReleasesExposedSurvivorUnassigned(t *testing.T) {
	s := position(t)
	ids := []string{cid(Hearts, 13), cid(Clubs, 13)}
	s, _ = s.Move(ids, 1, Formation, true)
	s, _ = s.Bind(Binding{"b", ids, "queen-hearts", "people", true})
	n, e := s.AssassinateBinding(ids[0])
	if e != nil {
		t.Fatal(e)
	}
	survivor, _ := n.Card(ids[1])
	if survivor.Controller != 1 || survivor.Zone != Unassigned || survivor.Allocation != "" {
		t.Fatal("exposed survivor not released")
	}
	if s.BindingStatus("b") != "active" {
		t.Fatal("release aliased predecessor")
	}
}
func TestCompensationDepartureExpiresRemainder(t *testing.T) {
	s := position(t)
	s, _ = s.Move([]string{cid(Hearts, 1)}, 2, ConcealedAce, true)
	s, effectErr := s.ActivateEffect(AceEffect{ID: "c", CardID: cid(Hearts, 1), Kind: "compensation", Source: 2, Target: 2, Custodian: 2, Losses: 4, Expiry: "target-departure-or-closure"})
	if effectErr != nil {
		t.Fatal(effectErr)
	}
	n, e := s.DepartPlayer(2)
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Effects) != 0 || len(n.DrawQueue) != 0 {
		t.Fatal("departure retained remainder entitlement")
	}
	a, _ := n.Card(cid(Hearts, 1))
	if a.Zone != Discard {
		t.Fatal("persistent Ace was returned to supply")
	}
}
