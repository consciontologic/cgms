package game

import (
	"errors"
	"slices"
)

type ProposalTerms struct {
	To      int      `json:"to"`
	Give    []string `json:"give"`
	Receive []string `json:"receive"`
	Loan    bool     `json:"loan"`
}
type Proposal struct {
	ID       string        `json:"id"`
	Revision int           `json:"revision"`
	GameID   string        `json:"game_id"`
	Turn     int           `json:"turn"`
	From     int           `json:"from"`
	Terms    ProposalTerms `json:"terms"`
	Status   string        `json:"status"`
}

func cloneProposals(in []Proposal) []Proposal {
	out := slices.Clone(in)
	for i := range out {
		out[i].Terms.Give = slices.Clone(in[i].Terms.Give)
		out[i].Terms.Receive = slices.Clone(in[i].Terms.Receive)
	}
	return out
}

func offerIndex(s State, id string) int {
	for i, p := range s.Proposals {
		if p.ID == id {
			return i
		}
	}
	return -1
}
func proposalCards(s State, p Proposal, checkQuota bool) error {
	if !eligible(s, p.From) || !eligible(s, p.Terms.To) || p.From == p.Terms.To || (len(p.Terms.Give) == 0 && len(p.Terms.Receive) == 0) {
		return errors.New("ineligible proposal parties")
	}
	if p.Terms.Loan {
		if len(p.Terms.Receive) != 0 || len(p.Terms.Give) == 0 || (s.Active != p.From && s.Active != p.Terms.To) {
			return errors.New("invalid loan terms")
		}
		if a := s.Players[p.From-1].Ally; a != 0 && a != p.Terms.To {
			return errors.New("exclusive alliance")
		}
		if a := s.Players[p.Terms.To-1].Ally; a != 0 && a != p.From {
			return errors.New("exclusive alliance")
		}
		intact := false
		first, _ := s.Card(p.Terms.Give[0])
		for _, r := range s.Formations {
			if r.ID == first.Allocation && r.Controller == p.From && ValidateFormationShape(s, p.From, r.Spec) == nil && containsSelection([][]string{r.Spec.Cards}, p.Terms.Give) {
				intact = true
				for _, id := range r.Spec.Cards {
					v, _ := s.Card(id)
					if v.Zone != Formation || v.Controller != p.From || v.Allocation != r.ID {
						intact = false
					}
				}
			}
		}
		if !intact {
			return errors.New("intact registered formation required")
		}
		if _, e := s.LoanFormation(p.Terms.Give, p.From, p.Terms.To); e != nil {
			return e
		}
	} else if s.Active != p.From {
		return errors.New("maker turn required")
	}
	seen := map[string]bool{}
	for j, ids := range [][]string{p.Terms.Give, p.Terms.Receive} {
		seat := p.From
		if j == 1 {
			seat = p.Terms.To
		}
		aces := 0
		for _, id := range ids {
			c, ok := s.Card(id)
			if !ok || seen[id] || c.Controller != seat || c.AvailableFromRound > s.Round || c.Zone == ActiveAce {
				return errors.New("unavailable transfer card")
			}
			seen[id] = true
			if !p.Terms.Loan && c.Zone != Hand && c.Zone != ConcealedAce && !s.Players[seat-1].Underground {
				return errors.New("open trade needs Underground")
			}
			if c.Card.Rank == 1 {
				aces++
				if c.Zone != ConcealedAce || len(s.CurrentSeries(seat)) < 2 || aces > 1 || checkQuota && s.Players[seat-1].AceRound == s.Round {
					return errors.New("Ace transfer allowance")
				}
			}
		}
	}
	return nil
}
func proposalCommand(s State, c Command) (State, error) {
	idx := offerIndex(s, c.OfferID)
	if c.OfferID == "" || c.Revision < 1 {
		return s, errors.New("proposal identity")
	}
	if c.Kind == "offer" {
		if s.Pending != nil || !eligible(s, c.Actor) || c.Terms == nil {
			return s, errors.New("proposal boundary")
		}
		p := Proposal{ID: c.OfferID, Revision: c.Revision, GameID: s.GameID, Turn: s.Turn, From: c.Actor, Terms: *c.Terms, Status: "offered"}
		if idx >= 0 {
			old := s.Proposals[idx]
			if old.Status != "offered" || old.From != c.Actor || old.Turn != s.Turn || old.Revision+1 != c.Revision {
				return s, errors.New("proposal revision")
			}
			p.Turn = old.Turn
		} else if c.Revision != 1 {
			return s, errors.New("initial proposal revision")
		}
		if e := proposalCards(s, p, true); e != nil {
			return s, e
		}
		n := s.Clone()
		if idx >= 0 {
			n.Proposals[idx] = cloneProposals([]Proposal{p})[0]
		} else {
			n.Proposals = append(n.Proposals, cloneProposals([]Proposal{p})[0])
		}
		return n, nil
	}
	if idx < 0 {
		return s, errors.New("unknown offer")
	}
	p := s.Proposals[idx]
	if p.Status != "offered" || p.Revision != c.Revision || p.GameID != s.GameID || p.Turn != s.Turn {
		return s, errors.New("closed or stale offer")
	}
	n := s.Clone()
	switch c.Kind {
	case "withdraw-offer":
		if c.Actor != p.From {
			return s, errors.New("not maker")
		}
		n.Proposals[idx].Status = "withdrawn"
	case "decline-offer":
		if c.Actor != p.Terms.To {
			return s, errors.New("not recipient")
		}
		n.Proposals[idx].Status = "declined"
	case "accept-offer":
		if s.Pending != nil || c.Actor != p.Terms.To {
			return s, errors.New("acceptance boundary")
		}
		if e := proposalCards(s, p, true); e != nil {
			return s, e
		}
		for _, id := range append(slices.Clone(p.Terms.Give), p.Terms.Receive...) {
			v, _ := s.Card(id)
			if v.Card.Rank == 1 {
				n.Players[v.Controller-1].AceRound = s.Round
			}
		}
		n.Proposals[idx].Status = "accepted"
		n.Pending = &PendingAction{ID: c.ID, Kind: "transfer", Actor: c.Actor, ProposalID: p.ID, Responders: responseSeats(s, c.Actor)}
	default:
		return s, errors.New("unknown proposal operation")
	}
	return n, nil
}
func finishTransfer(s State) (State, []Event, error) {
	p := s.Pending
	idx := offerIndex(s, p.ProposalID)
	if idx < 0 {
		return s, nil, errors.New("missing accepted proposal")
	}
	offer := s.Proposals[idx]
	n := s.Clone()
	if e := proposalCards(s, offer, false); e != nil {
		n.Pending = nil
		n.Proposals[idx].Status = "failed"
		return n, []Event{{Kind: "transfer-invalidated", Actor: p.Actor}}, nil
	}
	if offer.Terms.Loan {
		var e error
		n, e = n.LoanFormation(offer.Terms.Give, offer.From, offer.Terms.To)
		if e != nil {
			return s, nil, e
		}
		n.Players[offer.From-1].Ally = offer.Terms.To
		n.Players[offer.Terms.To-1].Ally = offer.From
		first, _ := n.Card(offer.Terms.Give[0])
		for i := range n.Formations {
			if n.Formations[i].ID == first.Allocation {
				n.Formations[i].Controller = offer.Terms.To
				if n.Formations[i].Spec.Kind == "underground" {
					n.Players[offer.Terms.To-1].Underground = true
				}
			}
		}
	} else {
		for j, ids := range [][]string{offer.Terms.Give, offer.Terms.Receive} {
			to := offer.Terms.To
			if j == 1 {
				to = offer.From
			}
			for _, id := range ids {
				v, _ := n.Card(id)
				z := Hand
				if v.Card.Rank == 1 {
					z = ConcealedAce
				}
				var e error
				n, e = n.Move([]string{id}, to, z, false)
				if e != nil {
					return s, nil, e
				}
			}
		}
	}
	if !offer.Terms.Loan {
		for j, ids := range [][]string{offer.Terms.Give, offer.Terms.Receive} {
			seat := offer.From
			if j == 1 {
				seat = offer.Terms.To
			}
			printed := 0
			for _, id := range ids {
				v, _ := s.Card(id)
				if v.Card.Suit == Spades && v.Card.Rank >= 2 && v.Card.Rank <= 10 {
					printed += v.Card.Rank
				}
			}
			if printed >= 20 {
				for i, e := range n.Effects {
					if e.Kind == "inflation" && e.Target == seat {
						for k := range n.Cards {
							if n.Cards[k].Card.ID == e.CardID {
								n.Cards[k].Zone = Discard
								n.Cards[k].Controller = 0
							}
						}
						n.Effects = append(n.Effects[:i], n.Effects[i+1:]...)
						break
					}
				}
			}
		}
	}
	n.Proposals[idx].Status = "completed"
	n.Pending = nil
	return n, []Event{{Kind: "transfer-completed", Actor: p.Actor}}, nil
}

// ExpireProposals is called at turn/closure/departure boundaries, never on silence.
func ExpireProposals(s State) State {
	n := s.Clone()
	for i, p := range n.Proposals {
		if p.Status == "offered" && (p.Turn != s.Turn || s.Phase != "playing" || s.Players[p.From-1].Departed || s.Players[p.Terms.To-1].Departed) {
			n.Proposals[i].Status = "expired"
		}
	}
	return n
}
