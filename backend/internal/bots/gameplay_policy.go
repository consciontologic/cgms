package bots

import (
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"slices"
	"sort"
)

// ChooseGameplay is a transparent observation-only policy. The random policy is
// uniform over the versioned finite menu, not over the exponential action space.
// Economic and pressure use fixed distinct priorities; neither is a strength claim.
func ChooseGameplay(policy string, o game.Observation, r *randomstream.Stream, budget int) (Decision, error) {
	return chooseGameplay(policy, o, nil, r, budget)
}

// GameplayV1OrderingHasTies identifies the frozen v1 scorer's ambiguous ordering
// without changing it. Callers that remove adapter-invalid candidates must use
// the original eager order in this case because the historical sort is unstable.
func GameplayV1OrderingHasTies(o game.Observation) bool {
	seen := make(map[string]bool, len(o.Legal))
	for _, c := range o.Legal {
		key := gameplayActionKey(o, c)
		if seen[key] {
			return true
		}
		seen[key] = true
	}
	return false
}

func chooseGameplay(policy string, o game.Observation, economy *GameplayEconomy, r *randomstream.Stream, budget int) (Decision, error) {
	d := Decision{Policy: policy, MenuVersion: o.MenuCoverage}
	orderV2 := GameplayOrderV2(policy)
	policy = gameplayScoringPolicy(policy)
	if !opportunityPolicy(policy) && policy != "gameplay-random@v1" && policy != "gameplay-random@v2" && policy != "economic@v1" && policy != "pressure@v1" && policy != "legal-random@v1" && policy != "heuristic@v1" {
		return d, fmt.Errorf("unknown gameplay policy")
	}
	if len(o.Legal) == 0 {
		return d, ErrUnsupported
	}
	if budget < len(o.Legal) {
		return d, ErrBudget
	}
	menu := slices.Clone(o.Legal)
	for _, c := range menu {
		if c.Actor != o.Seat || c.GameID != o.GameID {
			return d, fmt.Errorf("unauthorized legal command")
		}
		d.Operations++
	}
	key := gameplayActionKey
	if orderV2 {
		key = gameplayActionKeyV2
	}
	sort.Slice(menu, func(i, j int) bool { return key(o, menu[i]) < key(o, menu[j]) })
	selected := 0
	if policy == "gameplay-random@v1" || policy == "gameplay-random@v2" || policy == "legal-random@v1" {
		if r == nil {
			return d, fmt.Errorf("missing policy stream")
		}
		j, ops, e := boundedSample(r.Uint64, uint64(len(menu)), budget-d.Operations)
		if e != nil {
			return d, e
		}
		selected = int(j)
		d.Operations += ops
		d.Reasons = []string{"uniform choice from " + game.GameplayMenuVersion, "physical subset representative sampling; not uniform over all legal physical combinations"}
	} else {
		if budget < d.Operations+len(menu) {
			return d, ErrBudget
		}
		best := -int(^uint(0) >> 1)
		for i, c := range menu {
			score, reasons := gameplayScore(policy, o, c)
			if opportunityPolicy(policy) {
				score, reasons = opportunityScore(policy, o, c)
				if policy == "bargaining@v1" {
					score, reasons = bargainingScore(o, c, score, reasons)
				}
				if economy != nil && policy != "opportunity-no-finance@v1" {
					score, reasons = opportunityEconomyScore(*economy, c, score, reasons)
				}
			}
			d.Operations++
			if score > best {
				best = score
				selected = i
				d.Reasons = reasons
			}
		}
	}
	d.Command = detachCommand(menu[selected])
	return d, nil
}
func detachCommand(c game.Command) game.Command {
	c.Cards = slices.Clone(c.Cards)
	c.Targets = slices.Clone(c.Targets)
	c.RandomWords = slices.Clone(c.RandomWords)
	if c.Terms != nil {
		v := *c.Terms
		v.Give = slices.Clone(v.Give)
		v.Receive = slices.Clone(v.Receive)
		c.Terms = &v
	}
	if c.Formation != nil {
		v := *c.Formation
		v.Cards = slices.Clone(v.Cards)
		if v.Protection != nil {
			p := *v.Protection
			v.Protection = &p
		}
		if v.Substitute != nil {
			p := *v.Substitute
			p.Kings = slices.Clone(p.Kings)
			v.Substitute = &p
		}
		c.Formation = &v
	}
	return c
}
func gameplayScore(policy string, o game.Observation, c game.Command) (int, []string) {
	score := -20
	reason := "unprioritized legal category"
	pressure := policy == "pressure@v1"
	switch c.Kind {
	case "coup", "ordinary-victory":
		score = 100000
		reason = "take an immediately valid game ending"
	case "end-turn":
		score = 0
		reason = "explicit turn completion after useful actions"
	case "open-series":
		score = 100
		reason = "build support and historical victory eligibility"
		if c.Suit == game.Clubs && pressure {
			score = 130
		}
	case "attach":
		score = 40
		reason = "activate supported non-combo royal"
		for _, v := range o.Cards {
			if slices.Contains(c.Cards, v.Card.ID) && v.Card.Rank == 13 {
				score = 25
			}
		}
	case "open-formation":
		score = 60
		reason = "establish supported combo"
		if c.Formation != nil && c.Formation.Kind == "underground" {
			score = 160
			reason = "earn lasting Underground privilege and income"
		}
	case "take-back":
		score = -100
		reason = "avoid deterministic takeback/reopen loops absent a planned improvement"
	case "code":
		score = 50 - c.Value + c.Price
		reason = "choose lower enemy diamond bonus and higher enemy purchase price"
	case "ringleader":
		score = 85
		reason = "realize guaranteed ten-point award"
	case "richer-sacrifice":
		score = -5
		reason = "preserve known round income rather than blindly sacrifice"
	case "barricade-sacrifice":
		score = -15
		reason = "preserve defense unless a search demonstrates net draw value"
	case "purchase":
		score = 5
		reason = "convert spendable Spades into card opportunities"
		if pressure {
			score = 2
		}
		if o.Round >= 11 {
			score = -10
			reason = "preserve late-game Spade holdings"
		}
	case "attack", "baron-attack", "infiltrator-attack":
		score = 15 + len(c.Targets)*3
		if pressure {
			score += 40
		}
		reason = "reduce eligible enemy exposed holdings"
		for _, v := range o.Cards {
			if slices.Contains(c.Targets, v.Card.ID) {
				if v.Card.Suit == game.Diamonds {
					score += 10 + v.Card.Rank
				}
				if v.Card.Suit == game.Clubs {
					score -= len(c.Cards) * 2
				}
			}
		}
	case "bomb-attack", "baron-bomb-attack":
		score = 8
		if pressure {
			score = 25
		}
		reason = "remove defense while accounting for committed Club cost"
	case "dexter-assassination":
		score = 20
		if pressure {
			score = 45
		}
		reason = "remove exposed enemy royal for the lowest printed Diamond payment"
		for _, v := range o.Cards {
			if slices.Contains(c.Cards, v.Card.ID) {
				score -= v.Card.Rank
			}
		}
	case "fate":
		score = 3
		if pressure {
			score = 12
		}
		reason = "disrupt eligible opponent at the known formation cost"
	case "justice", "kidnapper", "ponzi":
		score = 50
		reason = "use supported acquisition/economic opportunity"
	case "compensation":
		score = 30
		reason = "activate once-per-game protection before future losses"
	case "main-inflation":
		score = 5
		if pressure {
			score = 18
		}
		reason = "reduce opponent Spade effectiveness"
	case "pass":
		score = 0
		reason = "explicitly decline current response"
	case "confinement":
		score = 70
		reason = "cancel the current hostile action with a legal defense"
	case "negotiate":
		score = 40
		reason = "cancel incoming combat with legal exact payment"
	case "inflation", "advancement-defense", "numerical-defense", "compensation-response":
		score = 35
		reason = "use authorized defense before resolution"
	case "decision":
		score = 20
		reason = "resolve required choice deterministically"
		if o.DecisionKind == "dexter" && len(c.Cards) > 0 {
			score = 80
			reason = "prevent hostile Ace with legal Diamond payment"
		}
		if o.DecisionKind == "negotiator" {
			score -= len(c.Cards)
		}
	case "kidnapper-outcome":
		score = 20
		reason = "select authorized exposed theft target"
	case "decision-closed":
		score = 19
		reason = "request authorized hidden sampling without inspecting identities"
	case "accept-offer":
		score = -1
		reason = "no unexamined trade value assumption"
	case "decline-offer":
		score = 1
		reason = "explicitly decline an unvalued proposal"
	case "offer", "withdraw-offer":
		score = -50
		reason = "do not churn proposals without a beneficial transaction model"
	}
	if c.Kind == "open-formation" && c.Formation != nil {
		for _, r := range o.Formations {
			if r.ID == c.FormationID && game.Digest(r.Spec) == game.Digest(*c.Formation) {
				score = -100
				reason = "avoid unchanged reconfiguration"
			}
		}
	}
	if c.Kind == "attach" {
		for _, v := range o.Cards {
			if slices.Contains(c.Cards, v.Card.ID) && v.Card.Rank == 13 {
				for _, a := range o.Cards {
					if a.Controller == o.Seat && a.Zone == game.Attachment && a.Allocation == string(c.Suit) && a.Card.Rank == 13 {
						score = -30
						reason = "avoid disabling single-king Richer income"
					}
				}
			}
		}
	}
	return score, []string{fmt.Sprintf("%s: priority %d", reason, score), "uses authorized visible cards and menu only; no future draws or opponent hands", "debt consequences are not available in this board observation; no debt-aware strength claim", "semantic action hash, excluding administrative identities, breaks ties"}
}

// ChooseGameplayFinance remains an explicit finish policy over the authorized
// financial menu. It never converts a pending promise into refusal or completion.
func ChooseGameplayFinance(policy string, o LedgerProjection, budget int) (FinanceCommand, error) {
	if GameplayOrderV2(policy) {
		return ChooseFinance("heuristic@v1", o, budget)
	}
	switch policy {
	case "gameplay-random@v1", "economic@v1", "pressure@v1":
		return ChooseFinance("heuristic@v1", o, budget)
	default:
		return ChooseFinance(policy, o, budget)
	}
}

func gameplayActionKey(o game.Observation, c game.Command) string {
	c = detachCommand(c)
	c.ID = ""
	c.GameID = ""
	c.WindowID = ""
	c.DecisionID = ""
	if c.FormationID != "" {
		for _, r := range o.Formations {
			if r.ID == c.FormationID {
				v := detachCommand(game.Command{Formation: &r.Spec})
				c.Formation = v.Formation
			}
		}
		c.FormationID = ""
	}
	if c.Formation != nil && c.Formation.Protection != nil && c.Formation.Protection.Formation != "" {
		for _, r := range o.Formations {
			if r.ID == c.Formation.Protection.Formation {
				c.Formation.Protection.Formation = game.Digest(r.Spec.Cards)
				break
			}
		}
	}
	if c.OfferID != "" {
		for _, p := range o.Proposals {
			if p.ID == c.OfferID {
				terms := p.Terms
				c.Terms = &terms
			}
		}
		c.OfferID = ""
	}
	return game.Digest(c)
}
