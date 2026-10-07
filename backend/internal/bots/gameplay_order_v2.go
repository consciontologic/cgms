package bots

import (
	"github.com/metaphy6/cgms/backend/internal/game"
	"slices"
	"strings"
)

// GameplayOrderV2 identifies only the six deterministic policies whose corrected
// semantic action ordering was explicitly versioned. All v1 ordering is frozen.
func GameplayOrderV2(policy string) bool {
	return slices.Contains([]string{"economic@v2", "pressure@v2", "opportunity@v2", "opportunity-no-retain@v2", "opportunity-no-finance@v2", "bargaining@v2"}, policy)
}
func gameplayScoringPolicy(policy string) string {
	if GameplayOrderV2(policy) {
		return strings.TrimSuffix(policy, "@v2") + "@v1"
	}
	return policy
}
func gameplayActionKeyV2(o game.Observation, c game.Command) string {
	c = detachCommand(c)
	c.ID = ""
	c.GameID = ""
	c.WindowID = ""
	c.DecisionID = ""
	// Preserve the request and fingerprint its referenced existing formation
	// separately. Physical membership and kind do not recursively follow protective
	// references, so cycles and administrative identifiers cannot enter the key.
	identity := func(id string) string {
		for _, f := range o.Formations {
			if f.ID == id {
				cards := slices.Clone(f.Spec.Cards)
				slices.Sort(cards)
				return game.Digest(struct {
					Kind  string
					Cards []string
				}{f.Spec.Kind, cards})
			}
		}
		return "unresolved-formation"
	}
	if c.FormationID != "" {
		c.FormationID = identity(c.FormationID)
	}
	if c.Formation != nil && c.Formation.Protection != nil && c.Formation.Protection.Formation != "" {
		c.Formation.Protection.Formation = identity(c.Formation.Protection.Formation)
	}
	if c.OfferID != "" {
		for _, p := range o.Proposals {
			if p.ID == c.OfferID {
				terms := p.Terms
				terms.Give = slices.Clone(terms.Give)
				terms.Receive = slices.Clone(terms.Receive)
				c.Terms = &terms
				break
			}
		}
		c.OfferID = ""
	}
	return game.Digest(c)
}
