package bots

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"slices"
)

// bargainingScore is a versioned marginal-utility heuristic, not expected score.
// It only values available menu terms and authenticated party-visible proposals.
func bargainingScore(o game.Observation, c game.Command, score int, reasons []string) (int, []string) {
	if c.Kind != "offer" && c.Kind != "accept-offer" {
		return score, reasons
	}
	reason := "decline unvalued exchange or permanent alliance"
	score = -50
	var give, receive []string
	faces := map[string]game.Card{}
	for _, v := range o.Cards {
		faces[v.Card.ID] = v.Card
	}
	if c.Kind == "offer" {
		if c.Terms == nil || c.Terms.Loan || len(c.Terms.Receive) == 0 {
			return score, []string{reason}
		}
		for _, p := range o.Proposals {
			if p.From == o.Seat && p.GameID == o.GameID && p.Turn == o.Turn {
				return score, []string{"one proposal per turn including declined or withdrawn proposals"}
			}
		}
		give, receive = c.Terms.Give, c.Terms.Receive
	} else {
		var proposal *game.Proposal
		for _, p := range o.Proposals {
			if p.ID == c.OfferID && p.Revision == c.Revision && p.GameID == o.GameID && p.Turn == o.Turn && p.Status == "offered" && p.Terms.To == o.Seat && p.From != o.Seat {
				v := p
				proposal = &v
				break
			}
		}
		if proposal == nil {
			return score, []string{"no authenticated current named-party proposal"}
		}
		// Exact offered card identities are disclosed to these named parties by T01.
		// Resolve faces only, never private custody, availability, or unoffered cards.
		for _, card := range game.Deck() {
			if slices.Contains(proposal.Terms.Give, card.ID) || slices.Contains(proposal.Terms.Receive, card.ID) {
				faces[card.ID] = card
			}
		}
		give, receive = proposal.Terms.Receive, proposal.Terms.Give
		if proposal.Terms.Loan {
			if !usefulIncomingLoan(o, *proposal) {
				return score, []string{reason}
			}
			return 145, []string{"supported incoming formation utility exceeds fixed alliance reservation; fewer than three Coup queens", "permanent alliance disables Coup; no promise of reciprocal payment"}
		}
	}
	gain := 0
	for _, id := range receive {
		card, ok := faces[id]
		if !ok {
			return -50, []string{"offered face unavailable in authorized terms"}
		}
		gain += bargainingCardValue(o, card, give)
	}
	for _, id := range give {
		card, ok := faces[id]
		if !ok {
			return -50, []string{"payment face unavailable"}
		}
		gain -= bargainingCardValue(o, card, []string{id})
	}
	if gain > 0 {
		score = 120 + min(gain, 25)
		reason = "positive marginal card utility, including owned twin opportunity"
	}
	return score, []string{fmt.Sprintf("%s: priority %d; utility difference %d", reason, score, gain), "heuristic utility is not expected financial score; authenticated proposal faces only"}
}

func bargainingCardValue(o game.Observation, card game.Card, excluded []string) int {
	value := 1
	if card.Rank >= 2 && card.Rank <= 10 {
		if card.Suit == game.Diamonds {
			bonus := 5
			if o.Code != nil && o.Code.Declarer != o.Seat {
				bonus = o.Code.DiamondBonus
			}
			return bonus + card.Rank
		}
		if card.Suit == game.Spades {
			return card.Rank/2 + 1
		}
		return value
	}
	if card.Rank == 1 {
		return 5
	}
	value = 3
	if card.Suit == game.Spades {
		value = 10
	}
	if card.Rank == 13 {
		value += 2
	}
	for _, v := range o.Cards {
		if v.Controller == o.Seat && v.Card.ID != card.ID && !slices.Contains(excluded, v.Card.ID) && v.Card.Rank == card.Rank && v.Card.Suit == card.Suit && (v.Zone == game.Hand || v.Zone == game.Attachment) {
			value += 12
			break
		}
	}
	return value
}

func usefulIncomingLoan(o game.Observation, p game.Proposal) bool {
	queens := 0
	series := map[game.Suit]bool{}
	for _, v := range o.Cards {
		if v.Controller != o.Seat {
			continue
		}
		if v.Card.Rank == 12 && (v.Card.Suit == game.Hearts || v.Card.Suit == game.Diamonds) {
			queens++
		}
		if v.Zone == game.Series && v.Card.Rank >= 2 && v.Card.Rank <= 10 {
			series[v.Card.Suit] = true
		}
	}
	if queens >= 3 {
		return false
	}
	for _, l := range o.Loans {
		if l.To == o.Seat {
			return false
		}
	}
	for _, f := range o.Formations {
		if f.Controller != p.From || len(f.Spec.Cards) != len(p.Terms.Give) {
			continue
		}
		same := true
		for _, id := range f.Spec.Cards {
			if !slices.Contains(p.Terms.Give, id) {
				same = false
			}
		}
		if !same {
			continue
		}
		// Conservative natural-formation support; no inferred Underground privilege.
		if f.Spec.Substitute != nil {
			return false
		}
		return f.Spec.Kind == "underground" || f.Spec.Kind == "kidnapper" && len(series) >= 2 && series[game.Clubs]
	}
	return false
}
