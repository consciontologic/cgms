package bots

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"slices"
	"testing"
)

func bargainingFixture(t *testing.T) game.State {
	t.Helper()
	s, _ := game.NewState(3, "bargain")
	for seat, ids := range map[int][]string{1: {"deck-1-clubs-02", "deck-1-hearts-02"}, 2: {"deck-2-clubs-02", "deck-2-hearts-02", "deck-2-spades-02"}} {
		var e error
		s, e = s.Move(ids, seat, game.Series, false)
		if e != nil {
			t.Fatal(e)
		}
		s.Players[seat-1].History = []game.Suit{game.Clubs, game.Hearts, game.Spades}
	}
	return s
}

func TestBargainingActualMenuMutualTwinTradeAndNoChurn(t *testing.T) {
	s := bargainingFixture(t)
	var e error
	s, _ = s.Move([]string{"deck-1-clubs-11", "deck-1-spades-11"}, 1, game.Hand, false)
	s, _ = s.Move([]string{"deck-2-clubs-11", "deck-2-spades-11"}, 2, game.Hand, false)
	s.Active = 2
	s, e = game.AttachRoyal(s, 2, "deck-2-spades-11", game.Spades)
	if e != nil {
		t.Fatal(e)
	}
	s.Players[1].Underground = true
	s.Active = 1
	o, e := game.GameplayObservation(s, 1, 100000)
	if e != nil {
		t.Fatal(e)
	}
	d, e := ChooseGameplay("bargaining@v1", o, nil, 100000)
	if e != nil || d.Command.Kind != "offer" {
		t.Fatal("missing positive twin offer", d.Command.Kind, e)
	}
	s, _, e = game.Apply(s, d.Command)
	if e != nil {
		t.Fatal(e)
	}
	offered := s
	other, e := game.GameplayObservation(s, 2, 100000)
	if e != nil {
		t.Fatal(e)
	}
	accept, e := ChooseGameplay("bargaining@v1", other, nil, 100000)
	if e != nil || accept.Command.Kind != "accept-offer" {
		t.Fatal("recipient failed authorized hidden twin valuation", accept.Command.Kind, e)
	}
	resolved, _, e := game.Apply(s, accept.Command)
	if e != nil {
		t.Fatal("selected acceptance not legal", e)
	}
	for resolved.Pending != nil {
		pending := resolved.Pending
		seat := pending.Responders[pending.Cursor]
		resolved, _, e = game.Apply(resolved, game.Command{GameID: s.GameID, ID: fmt.Sprintf("reply-%d", seat), Actor: seat, Kind: "pass", WindowID: pending.ID})
		if e != nil {
			t.Fatal(e)
		}
	}
	for _, id := range d.Command.Terms.Give {
		card, _ := resolved.Card(id)
		if card.Controller != 2 {
			t.Fatal("accepted trade failed physical giving")
		}
	}
	for _, id := range d.Command.Terms.Receive {
		card, _ := resolved.Card(id)
		if card.Controller != 1 {
			t.Fatal("accepted trade failed physical receipt")
		}
	}
	// Party-only terms reveal the offered face, not unrelated hidden holdings.
	s, _ = s.Move([]string{"deck-1-hearts-08"}, 1, game.Hand, false)
	s, _ = s.Move([]string{"deck-1-hearts-09"}, 3, game.Hand, false)
	n := s.Clone()
	for i := range n.Cards {
		if n.Cards[i].Card.ID == "deck-1-hearts-08" {
			n.Cards[i].Controller = 3
		}
		if n.Cards[i].Card.ID == "deck-1-hearts-09" {
			n.Cards[i].Controller = 1
		}
	}
	a, _ := game.GameplayObservation(s, 2, 100000)
	b, _ := game.GameplayObservation(n, 2, 100000)
	if game.Digest(a) != game.Digest(b) {
		t.Fatal("unrelated hidden swap leaked into recipient observation/menu")
	}
	before := game.Digest(a)
	x, e := ChooseGameplay("bargaining@v1", a, nil, 100000)
	if e != nil {
		t.Fatal(e)
	}
	y, e := ChooseGameplay("bargaining@v1", b, nil, 100000)
	if e != nil || game.Digest(x) != game.Digest(y) || game.Digest(a) != before {
		t.Fatal("hidden state or mutation affected bargaining decision", e)
	}
	outsider, _ := game.GameplayObservation(s, 3, 100000)
	if len(outsider.Proposals) != 0 {
		t.Fatal("outsider received private proposal")
	}
	for _, v := range outsider.Cards {
		if slices.Contains(d.Command.Terms.Give, v.Card.ID) {
			t.Fatal("outsider received hidden offered face")
		}
	}
	for _, bad := range []string{"party", "revision", "game", "turn"} {
		invalid := a
		invalid.Proposals = append([]game.Proposal{}, a.Proposals...)
		switch bad {
		case "party":
			invalid.Proposals[0].Terms.To = 3
		case "revision":
			invalid.Proposals[0].Revision++
		case "game":
			invalid.Proposals[0].GameID = "other"
		case "turn":
			invalid.Proposals[0].Turn++
		}
		z, e := ChooseGameplay("bargaining@v1", invalid, nil, 100000)
		if e != nil || z.Command.Kind == "accept-offer" {
			t.Fatal("unauthenticated proposal valuation", bad, e)
		}
	}
	if _, e := ChooseGameplay("bargaining@v1", a, nil, len(a.Legal)); e != ErrBudget {
		t.Fatal("bargaining bypassed operation budget", e)
	}
	for _, status := range []string{"offered", "declined", "withdrawn"} {
		n := offered.Clone()
		n.Proposals[0].Status = status
		next, e := game.GameplayObservation(n, 1, 100000)
		if e != nil {
			t.Fatal(e)
		}
		answer, e := ChooseGameplay("bargaining@v1", next, nil, 100000)
		if e != nil || answer.Command.Kind == "offer" {
			t.Fatal("repeated proposal in same turn", status, e)
		}
	}
}

func TestBargainingActualLoanSupportAndCoupOption(t *testing.T) {
	s := bargainingFixture(t)
	s, _ = s.Move([]string{"deck-1-clubs-11", "deck-2-clubs-11"}, 2, game.Hand, false)
	s.Active = 2
	var e error
	s, e = game.OpenFormation(s, 2, "loan-k", game.FormationSpec{Kind: "kidnapper", Cards: []string{"deck-1-clubs-11", "deck-2-clubs-11"}})
	if e != nil {
		t.Fatal(e)
	}
	s.Active = 1
	s, _, e = game.Apply(s, game.Command{GameID: s.GameID, Actor: 2, ID: "loan-offer", Kind: "offer", OfferID: "loan", Revision: 1, Terms: &game.ProposalTerms{To: 1, Give: []string{"deck-1-clubs-11", "deck-2-clubs-11"}, Loan: true}})
	if e != nil {
		t.Fatal(e)
	}
	for _, near := range []bool{false, true} {
		n := s.Clone()
		if near {
			n, _ = n.Move([]string{"deck-1-hearts-12", "deck-2-hearts-12", "deck-1-diamonds-12"}, 1, game.Hand, false)
		}
		o, e := game.GameplayObservation(n, 1, 100000)
		if e != nil {
			t.Fatal(e)
		}
		d, e := ChooseGameplay("bargaining@v1", o, nil, 100000)
		if e != nil {
			t.Fatal(e)
		}
		if !near && d.Command.Kind != "accept-offer" {
			t.Fatal("supported useful loan declined", d.Command.Kind)
		}
		if near && d.Command.Kind == "accept-offer" {
			t.Fatal("ignored permanent alliance Coup cost")
		}
	}
}
