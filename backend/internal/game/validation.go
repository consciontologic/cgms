package game

import (
	"encoding/hex"
	"errors"
	"slices"
)

// validateMetadata protects every index consumed by observations, policies and
// continuations. Rule eligibility still belongs to transitions; intermediate
// atomic movement may change a referenced card's zone before clearing Pending.
func validateMetadata(s State) error {
	seats := len(s.Players)
	if s.Code != nil && (s.Code.Declarer < 1 || s.Code.Declarer > seats || s.Code.Formation == "" || s.Code.DiamondBonus < 1 || s.Code.DiamondBonus > 10 || s.Code.PurchasePrice < 5 || s.Code.PurchasePrice > 10) {
		return errors.New("invalid Code settings")
	}
	seatOK := func(n int) bool { return n >= 1 && n <= seats }
	if s.Round < 1 || s.Turn < 1 || s.Version < 0 || !seatOK(s.Active) || (s.Phase != "playing" && s.Phase != "settlement") {
		return errors.New("invalid state phase/counters")
	}
	cards := map[string]Card{}
	for _, c := range Deck() {
		cards[c.ID] = c
	}
	cardList := func(ids []string) bool {
		seen := map[string]bool{}
		for _, id := range ids {
			if _, ok := cards[id]; !ok || seen[id] {
				return false
			}
			seen[id] = true
		}
		return true
	}
	seatList := func(ids []int) bool {
		seen := map[int]bool{}
		for _, id := range ids {
			if !seatOK(id) || seen[id] {
				return false
			}
			seen[id] = true
		}
		return true
	}
	if len(s.Order) > seats || !seatList(s.Order) || !cardList(s.PublicHistory) {
		return errors.New("invalid order or public identity history")
	}
	if len(s.Order) > 0 {
		for _, p := range s.Players {
			if !p.Departed && !slices.Contains(s.Order, p.Seat) {
				return errors.New("missing active initiative seat")
			}
		}
	}
	for _, p := range s.Players {
		seen := map[Suit]bool{}
		for _, suit := range p.History {
			if (suit != Hearts && suit != Clubs && suit != Spades && suit != Diamonds) || seen[suit] {
				return errors.New("invalid opening history")
			}
			seen[suit] = true
		}
		if p.Ally < 0 || p.Ally > seats || p.Ally == p.Seat || p.AceRound < 0 {
			return errors.New("invalid player counters")
		}
		if p.Ally != 0 && s.Players[p.Ally-1].Ally != p.Seat {
			return errors.New("asymmetric alliance")
		}
		for key, v := range p.Quotas {
			if key == "" || v < 0 {
				return errors.New("invalid quota")
			}
		}
	}
	for _, c := range s.Cards {
		if c.UsedTurn < 0 || c.UsedTurn > s.Turn {
			return errors.New("invalid physical usage")
		}
	}
	bindings := map[string]bool{}
	for _, b := range s.Bindings {
		if b.ID == "" || bindings[b.ID] {
			return errors.New("duplicate binding identity")
		}
		bindings[b.ID] = true
		for _, id := range b.Kings {
			c, ok := s.Card(id)
			if !ok || c.Zone == Attachment {
				return errors.New("bound king cannot attach")
			}
			if c.Zone == Formation && c.Allocation != "" && c.Allocation != b.Formation {
				return errors.New("bound king allocated outside original formation")
			}
		}
	}
	effects := map[string]bool{}
	targets := map[string]map[int]bool{}
	for _, e := range s.Effects {
		if e.ID == "" || effects[e.ID] || e.Expiry == "" || e.Losses < 0 {
			return errors.New("invalid effect identity/counter")
		}
		effects[e.ID] = true
		if targets[e.Kind] == nil {
			targets[e.Kind] = map[int]bool{}
		}
		if targets[e.Kind][e.Target] {
			return errors.New("duplicate target effect")
		}
		targets[e.Kind][e.Target] = true
		c, ok := cards[e.CardID]
		if !ok || e.Kind == "inflation" && (c.Suit != Spades || e.Losses != 0) || e.Kind == "compensation" && c.Suit != Hearts {
			return errors.New("effect card mismatch")
		}
	}
	// Formation records are historical declarations. Validate immutable metadata,
	// not current custody/support: destruction and transfers legitimately break them.
	validSuit := func(v Suit) bool { return slices.Contains([]Suit{Hearts, Clubs, Spades, Diamonds}, v) }
	specOK := func(v FormationSpec) bool {
		if !slices.Contains([]string{"fate", "baron", "underground", "great-people", "people", "coup", "justice", "code", "kidnapper", "ponzi"}, v.Kind) || len(v.Cards) == 0 || !cardList(v.Cards) {
			return false
		}
		for _, id := range v.Cards {
			if cards[id].Rank < 11 {
				return false
			}
		}
		if sub := v.Substitute; sub != nil {
			if len(sub.Kings) != 2 || !cardList(sub.Kings) || sub.Slot.Rank < 11 || sub.Slot.Rank > 13 || !validSuit(sub.Slot.Suit) {
				return false
			}
			for _, id := range sub.Kings {
				if cards[id].Rank != 13 || !slices.Contains(v.Cards, id) {
					return false
				}
			}
		}
		if protection := v.Protection; protection != nil {
			if (protection.Series == "") == (protection.Formation == "") || protection.Series != "" && !validSuit(protection.Series) {
				return false
			}
		}
		return true
	}
	formationIDs := map[string]bool{}
	for _, f := range s.Formations {
		if f.ID == "" || formationIDs[f.ID] || !seatOK(f.Controller) || !specOK(f.Spec) {
			return errors.New("invalid historical formation metadata")
		}
		formationIDs[f.ID] = true
	}
	proposalIDs := map[string]bool{}
	for _, offer := range s.Proposals {
		if offer.ID == "" || proposalIDs[offer.ID] || offer.GameID != s.GameID || offer.Revision < 1 || offer.Turn < 1 || offer.Turn > s.Turn || !seatOK(offer.From) || !seatOK(offer.Terms.To) || offer.From == offer.Terms.To || !slices.Contains([]string{"offered", "accepted", "withdrawn", "declined", "failed", "completed", "expired", "canceled"}, offer.Status) || len(offer.Terms.Give)+len(offer.Terms.Receive) == 0 || !cardList(append(slices.Clone(offer.Terms.Give), offer.Terms.Receive...)) || offer.Terms.Loan && (len(offer.Terms.Give) == 0 || len(offer.Terms.Receive) != 0) {
			return errors.New("invalid proposal metadata")
		}
		proposalIDs[offer.ID] = true
	}
	for _, l := range s.Loans {
		if !seatOK(l.From) || !seatOK(l.To) || l.From == l.To || len(l.Cards) == 0 || l.Allocation == "" || !cardList(l.Cards) {
			return errors.New("invalid loan metadata")
		}
	}
	for id, hash := range s.Commands {
		b, e := hex.DecodeString(hash)
		if id == "" || e != nil || len(b) != 32 || hex.EncodeToString(b) != hash {
			return errors.New("invalid command receipt")
		}
	}
	decisions := map[string]bool{}
	for _, d := range s.Decisions {
		if d.ID == "" || d.CommandID == "" || decisions[d.ID] {
			return errors.New("invalid decision receipt")
		}
		decisions[d.ID] = true
	}
	p := s.Pending
	if p == nil {
		return nil
	}
	if s.Phase != "playing" || len(s.DrawQueue) > 0 || p.ID == "" || !seatOK(p.Actor) || p.Cursor < 0 || p.Cursor > len(p.Responders) || !seatList(p.Responders) || !cardList(p.Clubs) || !cardList(p.Targets) || !cardList(p.Aces) || p.AttackModifier < 0 || p.DefenseModifier < 0 {
		return errors.New("invalid pending action metadata")
	}
	for _, seat := range p.Responders {
		if seat == p.Actor {
			return errors.New("actor in response sequence")
		}
	}
	switch p.Kind {
	case "loan-return":
		if !seatOK(p.Defender) || p.Defender == p.Actor || len(p.Targets) == 0 {
			return errors.New("invalid loan return metadata")
		}
	case "":
		if !seatOK(p.Defender) || len(p.Clubs) == 0 || len(p.Targets) == 0 {
			return errors.New("invalid combat continuation")
		}
	case "opening":
		if p.Opening == nil {
			return errors.New("missing opening intent")
		}
		v := p.Opening
		switch v.Kind {
		case "open-series":
			if !validSuit(v.Suit) {
				return errors.New("invalid opening suit")
			}
		case "open-formation":
			if v.FormationID == "" || v.Formation == nil || !specOK(*v.Formation) {
				return errors.New("invalid formation opening")
			}
		case "take-back":
			if v.FormationID == "" {
				return errors.New("invalid take-back identity")
			}
		case "attach":
			if !validSuit(v.Suit) || cards[v.CardID].Rank < 11 {
				return errors.New("invalid attachment metadata")
			}
		default:
			return errors.New("invalid opening kind")
		}
	case "ability":
		if p.Ability == nil {
			return errors.New("missing ability intent")
		}
		v := p.Ability
		if !slices.Contains([]string{"kidnapper", "compensation", "main-inflation", "ringleader", "richer-sacrifice", "barricade-sacrifice", "dexter-assassination", "fate", "code"}, v.Kind) || !cardList(v.Cards) || !cardList(v.Targets) || v.TargetSeat != 0 && !seatOK(v.TargetSeat) || v.AceID != "" && cards[v.AceID].Rank != 1 {
			return errors.New("invalid ability metadata")
		}
		if (v.Kind == "kidnapper" || v.Kind == "main-inflation") && (!seatOK(v.TargetSeat) || v.TargetSeat == p.Actor) {
			return errors.New("invalid hostile ability target")
		}
		if v.Kind == "code" && (v.DiamondBonus < 1 || v.DiamondBonus > 10 || v.PurchasePrice < 5 || v.PurchasePrice > 10) {
			return errors.New("invalid Code declaration metadata")
		}
	case "transfer":
		idx := offerIndex(s, p.ProposalID)
		if idx < 0 {
			return errors.New("unknown accepted offer")
		}
		offer := s.Proposals[idx]
		if offer.Terms.To != p.Actor || (offer.Status != "accepted" && offer.Status != "completed" && offer.Status != "failed" && offer.Status != "canceled") {
			return errors.New("invalid accepted transfer actor/status")
		}
	case "purchase":
		if p.Quantity < 1 || p.Quantity > 104 || len(p.Targets) == 0 {
			return errors.New("invalid purchase continuation")
		}
	case "ponzi":
		if !seatOK(p.Defender) {
			return errors.New("invalid Ponzi continuation")
		}
	case "justice":
		if cards[p.JusticeQueen].Rank != 12 {
			return errors.New("invalid Justice continuation")
		}
	case "scheduled-turn-draw":
		if p.Decision != nil || len(p.Responders) > 0 || len(p.Clubs) > 0 || len(p.Targets) > 0 {
			return errors.New("invalid scheduled draw")
		}
	default:
		return errors.New("unsupported pending kind")
	}
	d := p.Decision
	if d == nil {
		return nil
	}
	if d.ID == "" || d.EffectID == "" || !seatOK(d.Actor) || d.Cursor < 0 || d.Cursor > len(d.Opponents) || !seatList(d.Opponents) || !cardList(d.Selected) || !cardList(d.Reserved) {
		return errors.New("invalid effect decision metadata")
	}
	for _, choice := range d.Choices {
		if !cardList(choice) {
			return errors.New("invalid decision physical selection")
		}
	}
	switch d.Kind {
	case "justice":
		if p.Kind != "justice" || d.Stage != "selection" || d.Actor != p.Actor || cards[d.QueenID].Rank != 12 {
			return errors.New("invalid Justice decision")
		}
		for _, choice := range d.Choices {
			if len(choice) != 1 {
				return errors.New("Justice requires singleton choices")
			}
		}
		if d.Cursor < len(d.Opponents) && len(d.Choices) == 0 {
			return errors.New("missing Justice choices")
		}
	case "negotiator":
		if p.Kind != "" || (d.Stage == "clubs" && d.Actor != p.Actor) || (d.Stage == "payment" && d.Actor != p.Defender) {
			return errors.New("wrong Negotiator decision actor")
		}
		if d.Stage != "clubs" && d.Stage != "payment" {
			return errors.New("invalid Negotiator stage")
		}
		if len(d.Choices) == 0 {
			return errors.New("missing Negotiator choices")
		}
		for _, choice := range d.Choices {
			if len(choice) == 0 {
				return errors.New("empty Negotiator choice")
			}
		}
	case "dexter":
		target := p.Actor
		if d.Stage == "ability-inflation" && p.Ability != nil {
			target = p.Ability.TargetSeat
		}
		if d.Stage == "attack-modifier" {
			target = p.Defender
		}
		if d.Actor != target || d.AceActor == d.Actor {
			return errors.New("wrong Dexter decision actor")
		}
		if d.Stage == "attack-modifier" && d.AceActor != p.Actor {
			return errors.New("wrong attacking ace actor")
		}
		if (d.Stage == "numerical-defense" || d.Stage == "advancement-defense") && d.AceActor != p.Defender {
			return errors.New("wrong defending ace actor")
		}
		switch d.Stage {
		case "confinement", "inflation", "numerical-defense", "advancement-defense", "attack-modifier", "ability-inflation":
		default:
			return errors.New("invalid Dexter stage")
		}
		if !seatOK(d.AceActor) || cards[d.AceID].Rank != 1 || len(d.Choices) == 0 {
			return errors.New("invalid Dexter decision")
		}
		passes := 0
		for _, choice := range d.Choices {
			if len(choice) == 0 {
				passes++
			} else if len(choice) != 1 || cards[choice[0]].Suit != Diamonds || cards[choice[0]].Rank < 2 || cards[choice[0]].Rank > 10 {
				return errors.New("invalid Dexter payment")
			}
		}
		if passes != 1 {
			return errors.New("missing/duplicate Dexter pass")
		}
	default:
		return errors.New("unsupported effect decision")
	}
	return nil
}

// validateRecordedChoices rechecks saved authorization before any atomic mutation.
// It is deliberately separate from Validate: movement inside an accepted atomic
// effect temporarily changes ownership before the pending record is cleared.
func validateRecordedChoices(s State) error {
	if e := s.Validate(); e != nil {
		return e
	}
	p := s.Pending
	if p == nil || p.Decision == nil {
		return errors.New("missing decision")
	}
	d := p.Decision
	expected := [][]string{}
	switch d.Kind {
	case "negotiator":
		n := s.Clone()
		n.Pending.Decision = nil
		n.Players[p.Defender-1].Quotas["negotiator"] = 0
		n, e := negotiate(n, Command{ID: d.EffectID, Actor: p.Defender})
		if e != nil {
			return errors.New("saved Negotiator no longer legal")
		}
		base := n.Pending.Decision
		if d.Stage == "clubs" {
			if base.Stage != "clubs" {
				return errors.New("invalid saved club adjustment")
			}
			expected = base.Choices
		} else if len(d.Selected) == 0 {
			if base.Stage != "payment" {
				return errors.New("missing club adjustment")
			}
			expected = base.Choices
		} else {
			if base.Stage != "clubs" || !containsSelection(base.Choices, d.Selected) {
				return errors.New("nonmaximal saved clubs")
			}
			total := sumCards(s, d.Selected).Add(IntAmount(int64(p.AttackModifier)))
			for _, v := range subsets(availableNumbers(s, p.Defender, Spades)) {
				if sumCards(s, v).Cmp(total) == 0 {
					expected = append(expected, v)
				}
			}
		}
	case "dexter":
		ace, ok := s.Card(d.AceID)
		if !ok || ace.Controller != d.AceActor || ace.Zone != ConcealedAce {
			return errors.New("missing committed hostile ace")
		}
		if d.Stage == "confinement" && ace.Card.Suit != Diamonds || d.Stage == "inflation" && ace.Card.Suit != Spades || d.Stage == "advancement-defense" && ace.Card.Suit != Clubs {
			return errors.New("wrong committed named ace")
		}
		diamonds := availableNumbers(s, d.Actor, Diamonds)
		total := sumCards(s, targetNumbers(s, d.Actor, Diamonds))
		if attached(s, d.Actor, Diamonds) == "" || total.Sign() <= 0 || total.Cmp(IntAmount(10)) > 0 || s.Players[d.Actor-1].Quotas["dexter"] == s.Round || len(diamonds) == 0 {
			return errors.New("saved Dexter no longer payable")
		}
		expected = append(expected, []string{})
		for _, id := range diamonds {
			expected = append(expected, []string{id})
		}
	case "justice":
		if d.Cursor >= len(d.Opponents) {
			return errors.New("finished Justice cannot await choice")
		}
		used := map[int]bool{}
		reservedSeats := map[int]bool{}
		for _, id := range d.Reserved {
			card, _ := s.Card(id)
			if !slices.Contains(d.Opponents[:d.Cursor], card.Controller) || reservedSeats[card.Controller] || card.Zone == ActiveAce || (card.Card.Rank == 1 && card.Zone != ConcealedAce) {
				return errors.New("invalid reserved Justice custody")
			}
			reservedSeats[card.Controller] = true
			if used[card.Card.Rank] {
				return errors.New("duplicate Justice reserved rank")
			}
			used[card.Card.Rank] = true
		}
		target := d.Opponents[d.Cursor]
		if !eligible(s, target) {
			return errors.New("ineligible Justice opponent")
		}
		for _, card := range s.Cards {
			if card.Controller == target && card.Zone != ActiveAce && !(card.Card.Rank == 1 && card.Zone != ConcealedAce) && !used[card.Card.Rank] {
				expected = append(expected, []string{card.Card.ID})
			}
		}
	default:
		return errors.New("unsupported decision kind")
	}
	if len(expected) != len(d.Choices) {
		return errors.New("saved choice set mismatch")
	}
	for _, choice := range expected {
		if !containsSelection(d.Choices, choice) {
			return errors.New("saved choice omitted authorized option")
		}
	}
	for _, choice := range d.Choices {
		if !containsSelection(expected, choice) {
			return errors.New("unauthorized saved choice")
		}
	}
	return nil
}
