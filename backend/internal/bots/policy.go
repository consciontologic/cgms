// Package bots implements fair policies over detached authorized observations only.
package bots

import (
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/randomstream"
	"slices"
	"sort"
)

var ErrBudget = errors.New("bot operation budget exhausted")
var ErrUnsupported = errors.New("no supported legal action; not an implicit pass")

const MenuVersion = game.LegalMenuVersion

type Decision struct {
	Command     game.Command `json:"command"`
	Reasons     []string     `json:"reasons"`
	Operations  int          `json:"operations"`
	Policy      string       `json:"policy"`
	MenuVersion string       `json:"menu_version"`
}
type Weights struct {
	ImmediateScore   int
	ExposedHoldings  int
	Protection       int
	KnownDebt        int
	LegalDeclaration int
}

var initialWeights = Weights{10, 1, 2, -1, 20}

func Choose(policy string, o game.Observation, r *randomstream.Stream, budget int) (Decision, error) {
	return ChooseWeighted(policy, o, r, budget, initialWeights)
}
func ChooseWeighted(policy string, o game.Observation, r *randomstream.Stream, budget int, w Weights) (Decision, error) {
	d := Decision{Policy: policy, MenuVersion: MenuVersion}
	if policy != "legal-random@v1" && policy != "heuristic@v1" {
		return d, fmt.Errorf("unknown policy")
	}
	if len(o.Legal) == 0 {
		return d, ErrUnsupported
	}
	if budget < len(o.Legal) {
		return d, ErrBudget
	}
	type entry struct {
		key string
		c   game.Command
	}
	menu := []entry{}
	seen := map[string]bool{}
	for _, c := range o.Legal {
		if c.Actor != o.Seat || c.GameID != o.GameID {
			return d, fmt.Errorf("unauthorized menu identity")
		}
		b, e := canonical.Marshal(c)
		if e != nil {
			return d, e
		}
		key := string(b)
		if !seen[key] {
			c.Cards = slices.Clone(c.Cards)
			c.Targets = slices.Clone(c.Targets)
			menu = append(menu, entry{key, c})
			seen[key] = true
		}
		d.Operations++
	}
	sort.Slice(menu, func(i, j int) bool { return menu[i].key < menu[j].key })
	chosen := 0
	if policy == "legal-random@v1" {
		if r == nil {
			return d, fmt.Errorf("missing bot stream")
		}
		j, used, e := boundedSample(r.Uint64, uint64(len(menu)), budget-d.Operations)
		if e != nil {
			return d, e
		}
		d.Operations += used
		if d.Operations > budget {
			return Decision{}, ErrBudget
		}
		chosen = int(j)
		d.Reasons = []string{"uniform over canonical distinct " + MenuVersion + " actions; no full-rule coverage claim"}
	} else {
		if budget < d.Operations+len(menu) {
			return Decision{}, ErrBudget
		}
		best := -int(^uint(0) >> 1)
		for i, m := range menu {
			score, reasons := features(o, m.c, w)
			d.Operations++
			if score > best {
				best = score
				chosen = i
				d.Reasons = reasons
			}
		}
	}
	d.Command = menu[chosen].c
	return d, nil
}
func features(o game.Observation, c game.Command, w Weights) (int, []string) {
	immediate, holdings, protect, declare := 0, 0, 0, 0
	switch c.Kind {
	case "attack":
		holdings = len(c.Targets)
		for _, id := range c.Targets {
			for _, card := range o.Cards {
				if card.Card.ID == id && card.Card.Suit == game.Clubs {
					immediate++
				}
			}
		}
	case "confinement", "negotiate", "advancement-defense", "numerical-defense":
		protect = 1
	case "coup", "ordinary-victory":
		declare = 1
	}
	score := w.ImmediateScore*immediate + w.ExposedHoldings*holdings + w.Protection*protect + w.LegalDeclaration*declare
	return score, []string{fmt.Sprintf("immediate-score proxy (exposed target Clubs): %d x %d", immediate, w.ImmediateScore), fmt.Sprintf("exposed-holdings target cards: %d x %d", holdings, w.ExposedHoldings), fmt.Sprintf("protection response: %d x %d", protect, w.Protection), "known-debt: unavailable in this combat observation; contribution 0", fmt.Sprintf("legal-declaration: %d x %d", declare, w.LegalDeclaration), "ties resolved by canonical command bytes"}
}
func boundedSample(next func() uint64, n uint64, budget int) (uint64, int, error) {
	used := 0
	exhausted := false
	v, e := randomstream.Sample(func() uint64 {
		if used >= budget {
			exhausted = true
			return ^uint64(0)
		}
		used++
		return next()
	}, n)
	if exhausted {
		return 0, used, ErrBudget
	}
	return v, used, e
}

// DefaultWeights returns the versioned initial heuristic profile by value.
func DefaultWeights() Weights { return initialWeights }
