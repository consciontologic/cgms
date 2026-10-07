package game

import "testing"

func TestProposalNonblockingRevisionAndAtomicTransfer(t *testing.T) {
	s, _ := NewState(3, "offer")
	s, _ = s.Move([]string{"deck-1-spades-05"}, 1, Hand, true)
	s, _ = s.Move([]string{"deck-1-clubs-05"}, 2, Hand, true)
	c := Command{GameID: s.GameID, ID: "propose", Kind: "offer", Actor: 1, OfferID: "o", Revision: 1, Terms: &ProposalTerms{To: 2, Give: []string{"deck-1-spades-05"}, Receive: []string{"deck-1-clubs-05"}}}
	n, _, e := Apply(s, c)
	if e != nil {
		t.Fatal(e)
	}
	if n.Pending != nil || n.Players[0].OpeningUsed {
		t.Fatal("offer blocks/reserves")
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "accept", Kind: "accept-offer", Actor: 2, OfferID: "o", Revision: 1})
	if e != nil {
		t.Fatal(e)
	}
	if n.Pending == nil || n.Pending.Actor != 2 {
		t.Fatal("acceptor must be actor")
	}
	v, _ := n.Card(c.Terms.Give[0])
	if v.Controller != 1 {
		t.Fatal("early transfer")
	}
	for _, seat := range []int{3, 1} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "accept"})
		if e != nil {
			t.Fatal(e)
		}
	}
	v, _ = n.Card(c.Terms.Give[0])
	if v.Controller != 2 || v.Zone != Hand || n.Proposals[0].Status != "completed" {
		t.Fatal("transfer not committed")
	}
	if n.Players[0].Ally != 0 {
		t.Fatal("trade formed alliance")
	}
}
func TestProposalStaleRevisionAndExpiry(t *testing.T) {
	s, _ := NewState(3, "offer")
	s, _ = s.Move([]string{"deck-1-spades-05"}, 1, Hand, true)
	c := Command{GameID: s.GameID, ID: "propose", Kind: "offer", Actor: 1, OfferID: "o", Revision: 1, Terms: &ProposalTerms{To: 2, Give: []string{"deck-1-spades-05"}}}
	n, _, e := Apply(s, c)
	if e != nil {
		t.Fatal(e)
	}
	c.ID = "revise"
	c.Revision = 2
	n, _, e = Apply(n, c)
	if e != nil {
		t.Fatal(e)
	}
	before := Digest(n)
	bad, _, e := Apply(n, Command{GameID: s.GameID, ID: "accept", Kind: "accept-offer", Actor: 2, OfferID: "o", Revision: 1})
	if e == nil || Digest(bad) != before {
		t.Fatal("stale acceptance")
	}
	n.Turn++
	bad, _, e = Apply(n, Command{GameID: s.GameID, ID: "late", Kind: "accept-offer", Actor: 2, OfferID: "o", Revision: 2})
	if e == nil {
		t.Fatal("expired turn accepted")
	}
	_ = bad
}
func TestBrokenFormationCannotGrantLoanPrivilege(t *testing.T) {
	s, _ := NewState(3, "loan")
	ids := []string{"deck-1-hearts-13", "deck-2-hearts-13", "deck-1-clubs-13", "deck-2-clubs-13", "deck-1-spades-13"}
	s, _ = s.Move(ids, 1, Hand, true)
	var e error
	s, e = OpenFormation(s, 1, "u", FormationSpec{Kind: "underground", Cards: ids})
	if e != nil {
		t.Fatal(e)
	}
	s, _ = s.Move(ids[:1], 2, Hand, true)
	_, _, e = Apply(s, Command{GameID: s.GameID, ID: "loan", Actor: 1, Kind: "offer", OfferID: "loan", Revision: 1, Terms: &ProposalTerms{To: 3, Give: ids[1:], Loan: true}})
	if e == nil {
		t.Fatal("broken formation loan accepted")
	}
}
func TestVoluntaryTradeClearsInflationAfterTwentyPrinted(t *testing.T) {
	s, _ := NewState(3, "trade-inflation")
	ids := []string{"deck-1-spades-10", "deck-2-spades-10"}
	s, _ = s.Move(ids, 1, Hand, true)
	s, _ = s.Move([]string{"deck-1-spades-01"}, 2, ConcealedAce, true)
	s, _ = s.ActivateEffect(AceEffect{ID: "infl", CardID: "deck-1-spades-01", Kind: "inflation", Source: 2, Target: 1, Custodian: 1, Expiry: "spend"})
	n, _, e := Apply(s, Command{GameID: s.GameID, ID: "offer", Actor: 1, Kind: "offer", OfferID: "o", Revision: 1, Terms: &ProposalTerms{To: 2, Give: ids}})
	if e != nil {
		t.Fatal(e)
	}
	n, _, e = Apply(n, Command{GameID: s.GameID, ID: "accept", Actor: 2, Kind: "accept-offer", OfferID: "o", Revision: 1})
	if e != nil {
		t.Fatal(e)
	}
	if len(n.Effects) != 1 {
		t.Fatal("cleared before completion")
	}
	for _, seat := range []int{3, 1} {
		n, _, e = Apply(n, Command{GameID: s.GameID, ID: string(rune('a' + seat)), Kind: "pass", Actor: seat, WindowID: "accept"})
		if e != nil {
			t.Fatal(e)
		}
	}
	if len(n.Effects) != 0 {
		t.Fatal("qualifying completed trade retains Inflation")
	}
}
