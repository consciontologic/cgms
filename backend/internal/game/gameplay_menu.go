package game

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
)

const GameplayMenuVersion = "sampled-all-categories-v2"

var ErrMenuBudget = errors.New("gameplay candidate budget exhausted")

// GameplayObservation contains only authorized information. The finite menu uses
// one deterministic physical subset per exact total, not every physical subset.
// A budget failure returns no menu and must never be interpreted as a pass.
func GameplayObservation(s State, seat, budget int) (Observation, error) {
	return GameplayObservationContext(context.Background(), s, seat, budget)
}

// GameplayObservationContext preserves the finite menu contract while allowing
// an online worker to stop expensive candidate validation when its job expires.
// Cancellation, like budget exhaustion, never returns a partial usable menu.
func GameplayObservationContext(ctx context.Context, s State, seat, budget int) (Observation, error) {
	if err := ctx.Err(); err != nil {
		return Observation{}, err
	}
	o, e := observeBase(s, seat)
	if e != nil {
		return o, e
	}
	o.MenuCoverage = GameplayMenuVersion
	o.Legal, e = gameplayCommandsContext(ctx, s, o, budget)
	if err := ctx.Err(); err != nil {
		o.Legal = nil
		return o, err
	}
	return o, e
}
func GameplayCommands(s State, seat, budget int) ([]Command, error) {
	o, e := observeBase(s, seat)
	if e != nil {
		return nil, e
	}
	return gameplayCommands(s, o, budget)
}
func gameplayCommands(s State, o Observation, budget int) ([]Command, error) {
	return gameplayCommandsContext(context.Background(), s, o, budget)
}

func gameplayCommandsContext(ctx context.Context, s State, o Observation, budget int) ([]Command, error) {
	out := []Command{}
	seen := map[string]bool{}
	rejected := map[string]bool{}
	sortKeys := map[string]string{}
	used := 0
	exhausted := false
	add := func(c Command, synthetic bool) {
		if ctx.Err() != nil {
			return
		}
		used++
		if used > budget {
			exhausted = true
			return
		}
		if c.Kind == "offer" && c.OfferID == "" {
			c.Revision = 1
			c.OfferID = fmt.Sprintf("offer/%d/%s", o.Version, Digest(c.Terms))
		}
		c.GameID = o.GameID
		c.Actor = o.Seat
		c.ID = fmt.Sprintf("menu/%d/%s", o.Version, Digest(c))
		key := c.ID // This already hashes every semantic command field and the fixed menu context.
		if seen[key] || rejected[key] {
			return
		}
		if !synthetic {
			if _, _, e := Apply(s, c); e != nil {
				rejected[key] = true
				return
			}
		}
		seen[key] = true
		sortKeys[key] = Digest(c)
		out = append(out, c)
	}
	// Coup is an own-inventory exception even while another actor must decide.
	add(Command{Kind: "coup"}, false)
	if s.Pending != nil && s.Pending.Kind == "ability" && s.Pending.Ability != nil && s.Pending.Ability.Kind == "kidnapper" && s.Pending.Actor == o.Seat && s.Pending.Cursor == len(s.Pending.Responders) && s.Pending.Decision == nil {
		for _, id := range KidnapperOpenCandidates(s, s.Pending.Ability.TargetSeat) {
			add(Command{Kind: "kidnapper-outcome", WindowID: s.Pending.ID, Cards: []string{id}}, false)
		}
		if exhausted {
			return nil, ErrMenuBudget
		}
		return out, nil
	}
	if o.WindowID != "" {
		if o.RequiredActor == o.Seat && o.DecisionID != "" {
			for _, v := range o.Choices {
				add(Command{Kind: "decision", WindowID: o.WindowID, DecisionID: o.DecisionID, Cards: slices.Clone(v)}, false)
			}
			if o.DecisionKind == "justice" { // Presence of hidden eligibility is not exposed before the authorized choice.
				add(Command{Kind: "decision-closed", WindowID: o.WindowID, DecisionID: o.DecisionID}, true)
			}
		} else if o.RequiredActor == o.Seat {
			add(Command{Kind: "pass", WindowID: o.WindowID}, false)
			add(Command{Kind: "negotiate", WindowID: o.WindowID}, false)
			for _, v := range o.Cards {
				if v.Controller == o.Seat && v.Zone == ConcealedAce {
					for _, k := range []string{"confinement", "inflation", "compensation-response", "advancement-defense"} {
						add(Command{Kind: k, WindowID: o.WindowID, AceID: v.Card.ID}, false)
					}
					for value := 2; value <= 10; value++ {
						add(Command{Kind: "numerical-defense", WindowID: o.WindowID, AceID: v.Card.ID, Value: value}, false)
					}
				}
			}
		}
		if exhausted {
			return nil, ErrMenuBudget
		}
		return out, nil
	}
	if o.Phase != "playing" {
		return out, nil
	}
	// Visible proposal terms are available only to the named parties.
	for _, loan := range o.Loans {
		if loan.To != o.Seat {
			continue
		}
		owned := []string{}
		for _, id := range loan.Cards {
			for _, c := range o.Cards {
				if c.Card.ID == id && c.Controller == o.Seat {
					owned = append(owned, id)
				}
			}
		}
		if len(owned) > 0 {
			add(Command{Kind: "return-loan", TargetSeat: loan.From, Cards: owned}, false)
			for _, id := range owned {
				add(Command{Kind: "return-loan", TargetSeat: loan.From, Cards: []string{id}}, false)
			}
		}
	}
	for _, p := range o.Proposals {
		if p.Status != "offered" {
			continue
		}
		if p.From == o.Seat {
			add(Command{Kind: "withdraw-offer", OfferID: p.ID, Revision: p.Revision}, false)
		}
		if p.Terms.To == o.Seat {
			add(Command{Kind: "accept-offer", OfferID: p.ID, Revision: p.Revision}, false)
			add(Command{Kind: "decline-offer", OfferID: p.ID, Revision: p.Revision}, false)
		}
	}
	for _, p := range o.Players {
		if p.Seat != o.Seat {
			add(Command{Kind: "ponzi", Value: p.Seat}, false)
		}
	}
	if o.Active != o.Seat {
		if exhausted {
			return nil, ErrMenuBudget
		}
		return out, nil
	}
	if formationBoundary(s, o.Seat) {
		add(Command{Kind: "end-turn"}, true)
	}
	history := []Suit{}
	for _, p := range o.Players {
		if p.Seat == o.Seat {
			history = p.History
		}
	}
	if len(history) == 4 {
		diamonds, royals := 0, 0
		for _, c := range o.Cards {
			if c.Controller == o.Seat && c.Zone != ActiveAce {
				if c.Card.Rank >= 11 {
					royals++
				}
				if c.Card.Suit == Diamonds && c.Card.Rank >= 2 && c.Card.Rank <= 10 {
					diamonds++
				}
			}
		}
		if diamonds >= 13 {
			add(Command{Kind: "ordinary-victory", Value: 13}, true)
		}
		if royals >= 15 {
			add(Command{Kind: "ordinary-victory", Value: 15}, true)
		}
	}
	for _, su := range []Suit{Hearts, Clubs, Spades, Diamonds} {
		add(Command{Kind: "open-series", Suit: su}, false)
	}
	own := []PlacedCard{}
	for _, c := range o.Cards {
		if c.Controller == o.Seat && c.AvailableFromRound <= o.Round {
			own = append(own, c)
		}
	}
	for _, c := range own {
		if c.Zone == Hand || c.Zone == Unassigned {
			for _, su := range []Suit{Hearts, Clubs, Spades, Diamonds} {
				add(Command{Kind: "attach", Suit: su, Cards: []string{c.Card.ID}}, false)
			}
		}
		if c.Zone == Attachment {
			for _, k := range []string{"ringleader", "richer-sacrifice"} {
				add(Command{Kind: k, Cards: []string{c.Card.ID}}, false)
			}
			if c.Card.Rank == 12 && c.Card.Suit == Hearts {
				others := []string{}
				for _, v := range own {
					if v.Card.ID != c.Card.ID && v.Zone != ActiveAce {
						others = append(others, v.Card.ID)
					}
				}
				if len(others) >= 5 {
					add(Command{Kind: "barricade-sacrifice", Cards: append([]string{c.Card.ID}, others[:5]...)}, false)
				}
			}
		}
		if c.Zone == ConcealedAce {
			add(Command{Kind: "compensation", AceID: c.Card.ID}, false)
			for _, p := range o.Players {
				if p.Seat != o.Seat {
					add(Command{Kind: "main-inflation", AceID: c.Card.ID, TargetSeat: p.Seat}, false)
				}
			}
		}
	}
	for _, r := range o.Formations {
		if r.Controller != o.Seat {
			continue
		}
		add(Command{Kind: "take-back", FormationID: r.ID}, false)
		switch r.Spec.Kind {
		case "code":
			for _, bonus := range []int{1, 5, 10} {
				for _, price := range []int{5, 10} {
					add(Command{Kind: "code", FormationID: r.ID, Value: bonus, Price: price}, false)
				}
			}
		case "fate":
			for _, p := range o.Players {
				if p.Seat != o.Seat {
					for _, id := range r.Spec.Cards {
						add(Command{Kind: "fate", FormationID: r.ID, TargetSeat: p.Seat, Cards: []string{id}}, false)
					}
				}
			}
		case "justice":
			for _, id := range r.Spec.Cards {
				add(Command{Kind: "justice", AceID: id, FormationID: r.ID}, false)
			}
		case "kidnapper":
			for _, p := range o.Players {
				if p.Seat != o.Seat {
					add(Command{Kind: "kidnapper", FormationID: r.ID, TargetSeat: p.Seat}, false)
				}
			}
		}
		for _, p := range o.Players {
			if p.Seat != o.Seat {
				terms := ProposalTerms{To: p.Seat, Give: slices.Clone(r.Spec.Cards), Loan: true}
				add(Command{Kind: "offer", Terms: &terms}, false)
			}
		}
	}
	// Public gifts and trades never guess a concealed opponent identity.
	for _, v := range own {
		if v.Zone != Hand && v.Zone != ConcealedAce {
			continue
		}
		for _, p := range o.Players {
			if p.Seat == o.Seat {
				continue
			}
			terms := ProposalTerms{To: p.Seat, Give: []string{v.Card.ID}}
			add(Command{Kind: "offer", Terms: &terms}, false)
			for _, w := range o.Cards {
				if w.Controller == p.Seat && publicZone(w.Zone) && w.Zone != ActiveAce {
					tr := terms
					tr.Receive = []string{w.Card.ID}
					add(Command{Kind: "offer", Terms: &tr}, false)
				}
			}
		}
	}
	for _, r := range o.Formations {
		if r.Controller != o.Seat {
			continue
		}
		if r.Spec.Kind == "people" || r.Spec.Kind == "great-people" {
			for _, su := range []Suit{Clubs, Spades, Diamonds} {
				spec := r.Spec
				spec.Protection = &FormationProtection{Series: su}
				add(Command{Kind: "open-formation", FormationID: r.ID, Formation: &spec}, false)
			}
		}
	}
	for _, spec := range sampledFormationSpecs(own) {
		if spec.Kind == "people" || spec.Kind == "great-people" {
			for _, su := range []Suit{Clubs, Spades, Diamonds} {
				v := spec
				v.Protection = &FormationProtection{Series: su}
				add(Command{Kind: "open-formation", FormationID: "formed/" + Digest(v), Formation: &v}, false)
			}
			if spec.Kind == "great-people" {
				for _, r := range o.Formations {
					if r.Controller == o.Seat {
						v := spec
						v.Protection = &FormationProtection{Formation: r.ID}
						add(Command{Kind: "open-formation", FormationID: "formed/" + Digest(v), Formation: &v}, false)
					}
				}
			}
		} else {
			v := spec
			add(Command{Kind: "open-formation", FormationID: "formed/" + Digest(v), Formation: &v}, false)
		}
	}
	// Representative exact-value subsets are independent of hidden opponent cards.
	clubs := []PlacedCard{}
	spades := []PlacedCard{}
	diamonds := []string{}
	for _, c := range own {
		if c.Zone == Series && c.Card.Suit == Clubs && c.UsedTurn != o.Turn {
			clubs = append(clubs, c)
		}
		if (c.Zone == Series || c.Zone == Hand) && c.Card.Suit == Spades && c.Card.Rank >= 2 && c.Card.Rank <= 10 {
			spades = append(spades, c)
		}
		if c.Zone == Series && c.Card.Suit == Diamonds {
			diamonds = append(diamonds, c.Card.ID)
		}
	}
	for _, subset := range representativeSubsets(o, spades) {
		total := observationTotal(o, subset)
		price := 5
		if o.Code != nil && o.Code.Declarer != o.Seat {
			price = o.Code.PurchasePrice
		}
		for q := 1; q <= 104 && IntAmount(int64(q*price)).Cmp(total) <= 0; q++ {
			add(Command{Kind: "purchase", Cards: subset, Value: q}, false)
		}
	}
	for _, target := range o.Cards {
		if target.Controller != o.Seat && target.Controller > 0 && target.Card.Rank >= 11 && publicZone(target.Zone) {
			for _, id := range diamonds {
				add(Command{Kind: "dexter-assassination", Cards: []string{id}, Targets: []string{target.Card.ID}}, false)
			}
		}
	}
	attackFlags := [][2]bool{{false, false}}
	hasInfiltrator, hasExile := false, false
	for _, c := range own {
		if c.Zone == Attachment && c.Card.Suit == Clubs {
			hasInfiltrator = hasInfiltrator || c.Card.Rank == 11
			hasExile = hasExile || c.Card.Rank == 12
		}
	}
	if hasInfiltrator {
		attackFlags = append(attackFlags, [2]bool{true, false})
	}
	if hasExile {
		attackFlags = append(attackFlags, [2]bool{false, true})
	}
	if hasInfiltrator && hasExile {
		attackFlags = append(attackFlags, [2]bool{true, true})
	}
	cs := representativeSubsets(o, clubs)
	// Prune impossible modifiers before charging the candidate budget. Dense
	// ordinary boards otherwise spend it enumerating Baron attacks without a
	// usable Baron, partial-series Baron targets and already-used Ace powers.
	canBaron := hasBaron(s, o.Seat) && s.Players[o.Seat-1].Quotas["baron"] != o.Round
	canAttackAce := s.Players[o.Seat-1].AceRound != o.Round
	for _, p := range o.Players {
		if p.Seat == o.Seat {
			continue
		}
		for _, su := range []Suit{Hearts, Clubs, Spades, Diamonds} {
			targets := []PlacedCard{}
			for _, v := range o.Cards {
				if v.Controller == p.Seat && v.Zone == Series && v.Card.Suit == su {
					targets = append(targets, v)
				}
			}
			ts := representativeSubsets(o, targets)
			if len(targets) > 0 {
				all := []string{}
				for _, v := range targets {
					all = append(all, v.Card.ID)
				}
				ts = append(ts, all)
			}
			for _, a := range cs {
				for _, b := range ts {
					av, bv := observationTotal(o, a), observationTotal(o, b)
					baronTarget := canBaron && len(b) == len(targets)
					for _, ace := range own {
						if ace.Zone != ConcealedAce || !canAttackAce {
							continue
						}
						for value := 2; value <= 10; value++ {
							for _, kind := range []string{"attack", "baron-attack"} {
								if kind == "baron-attack" && !baronTarget {
									continue
								}
								cmp := av.Add(IntAmount(int64(value))).Cmp(bv)
								if kind == "attack" && cmp != 0 || kind == "baron-attack" && cmp <= 0 {
									continue
								}
								for _, flags := range attackFlags {
									add(Command{Kind: kind, Cards: a, Targets: b, AceID: ace.Card.ID, Value: value, Infiltrator: flags[0], Exile: flags[1]}, false)
									if value == 10 && ace.Card.Suit == Clubs {
										add(Command{Kind: kind, Cards: a, Targets: b, AceID: ace.Card.ID, Infiltrator: flags[0], Exile: flags[1]}, false)
									}
								}
							}
						}
					}
					for _, kind := range []string{"attack", "baron-attack"} {
						if kind == "baron-attack" && !baronTarget {
							continue
						}
						if kind == "attack" && av.Cmp(bv) != 0 || kind == "baron-attack" && av.Cmp(bv) <= 0 {
							continue
						}
						for _, flags := range attackFlags {
							add(Command{Kind: kind, Cards: a, Targets: b, Infiltrator: flags[0], Exile: flags[1]}, false)
						}
					}
				}
			}
		}
		for _, v := range o.Cards {
			if v.Controller == p.Seat && v.Zone == Attachment && v.Card.Rank == 11 && v.Card.Suit == Hearts {
				for _, a := range cs {
					for _, k := range []string{"bomb-attack", "baron-bomb-attack"} {
						if k == "baron-bomb-attack" && !canBaron {
							continue
						}
						for _, flags := range attackFlags {
							add(Command{Kind: k, Cards: a, Targets: []string{v.Card.ID}, Infiltrator: flags[0], Exile: flags[1]}, false)
							for _, ace := range own {
								if ace.Zone != ConcealedAce || !canAttackAce {
									continue
								}
								for value := 2; value <= 10; value++ {
									if observationTotal(o, a).Add(IntAmount(int64(value))).Cmp(IntAmount(5)) < 0 {
										continue
									}
									add(Command{Kind: k, Cards: a, Targets: []string{v.Card.ID}, AceID: ace.Card.ID, Value: value, Infiltrator: flags[0], Exile: flags[1]}, false)
									if value == 10 && ace.Card.Suit == Clubs {
										add(Command{Kind: k, Cards: a, Targets: []string{v.Card.ID}, AceID: ace.Card.ID, Infiltrator: flags[0], Exile: flags[1]}, false)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if exhausted {
		return nil, ErrMenuBudget
	}
	sort.Slice(out, func(i, j int) bool { return sortKeys[out[i].ID] < sortKeys[out[j].ID] })
	return out, nil
}

func observationTotal(o Observation, ids []string) Amount {
	sum := IntAmount(0)
	for _, id := range ids {
		for _, c := range o.Cards {
			if c.Card.ID == id {
				a := IntAmount(int64(c.Card.Rank))
				for _, e := range o.Effects {
					if e.Kind == "inflation" && e.Target == c.Controller && c.Card.Suit == Spades {
						a, _ = a.Quo(IntAmount(2))
					}
				}
				sum = sum.Add(a)
			}
		}
	}
	return sum
}
func representativeSubsets(o Observation, cards []PlacedCard) [][]string {
	states := map[int][]string{0: nil}
	for _, c := range cards {
		v := c.Card.Rank * 2
		for _, e := range o.Effects {
			if e.Kind == "inflation" && e.Target == c.Controller && c.Card.Suit == Spades {
				v = c.Card.Rank
			}
		}
		keys := []int{}
		for k := range states {
			keys = append(keys, k)
		}
		sort.Sort(sort.Reverse(sort.IntSlice(keys)))
		for _, k := range keys {
			if _, ok := states[k+v]; !ok {
				states[k+v] = append(slices.Clone(states[k]), c.Card.ID)
			}
		}
	}
	keys := []int{}
	for k := range states {
		if k > 0 {
			keys = append(keys, k)
		}
	}
	sort.Ints(keys)
	out := [][]string{}
	for _, k := range keys {
		out = append(out, states[k])
	}
	return out
}
func sampledFormationSpecs(cards []PlacedCard) []FormationSpec {
	pool := []PlacedCard{}
	for _, c := range cards {
		if (c.Zone == Hand || c.Zone == Unassigned) && c.Card.Rank >= 11 {
			pool = append(pool, c)
		}
	}
	sort.Slice(pool, func(i, j int) bool { return pool[i].Card.ID < pool[j].Card.ID })
	type shape struct {
		kind  string
		slots []FormationSlot
	}
	shapes := []shape{}
	for _, r := range []int{13, 12} {
		slots := []FormationSlot{}
		for _, su := range []Suit{Hearts, Clubs, Spades, Diamonds} {
			slots = append(slots, FormationSlot{r, su})
		}
		kind := "fate"
		if r == 12 {
			kind = "justice"
		}
		shapes = append(shapes, shape{kind, slots})
	}
	for _, su := range []Suit{Hearts, Clubs, Spades, Diamonds} {
		shapes = append(shapes, shape{"baron", []FormationSlot{{13, su}, {13, su}, {13, Diamonds}}}, shape{"kidnapper", []FormationSlot{{11, su}, {11, su}}})
	}
	shapes = append(shapes, shape{"great-people", []FormationSlot{{12, Hearts}, {12, Hearts}}}, shape{"ponzi", []FormationSlot{{11, Clubs}, {11, Clubs}, {11, Spades}, {11, Spades}}})
	out := []FormationSpec{}
	build := func(sh shape, replace int, kings []string) {
		ids := slices.Clone(kings)
		for j, slot := range sh.slots {
			if j == replace {
				continue
			}
			found := ""
			for _, c := range pool {
				if c.Card.Rank == slot.Rank && c.Card.Suit == slot.Suit && !slices.Contains(ids, c.Card.ID) {
					found = c.Card.ID
					break
				}
			}
			if found == "" {
				return
			}
			ids = append(ids, found)
		}
		spec := FormationSpec{Kind: sh.kind, Cards: ids}
		if replace >= 0 {
			spec.Substitute = &FormationSubstitution{Kings: slices.Clone(kings), Slot: sh.slots[replace]}
			if spec.Kind == "great-people" {
				spec.Kind = "people"
			}
		}
		out = append(out, spec)
	}
	kings := []string{}
	for _, c := range pool {
		if c.Card.Rank == 13 {
			kings = append(kings, c.Card.ID)
		}
	}
	for _, sh := range shapes {
		build(sh, -1, nil)
		if sh.kind == "ponzi" {
			continue
		}
		for slot := range sh.slots {
			for i := 0; i < len(kings); i++ {
				for j := i + 1; j < len(kings); j++ {
					build(sh, slot, []string{kings[i], kings[j]})
				}
			}
		}
	}
	for _, r := range []int{12, 13} {
		ids := []string{}
		for _, c := range pool {
			if c.Card.Rank == r {
				ids = append(ids, c.Card.ID)
			}
		}
		kind := "code"
		if r == 13 {
			kind = "underground"
		}
		for size := 5; size <= len(ids); size++ {
			out = append(out, FormationSpec{Kind: kind, Cards: slices.Clone(ids[:size])})
		}
	}
	return out
}
