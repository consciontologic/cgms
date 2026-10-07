package game

import (
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"slices"
)

type Command struct {
	Exile       bool           `json:"exile,omitempty"`
	Infiltrator bool           `json:"infiltrator,omitempty"`
	Suit        Suit           `json:"suit,omitempty"`
	Formation   *FormationSpec `json:"formation,omitempty"`
	OfferID     string         `json:"offer_id,omitempty"`
	Revision    int            `json:"revision,omitempty"`
	Terms       *ProposalTerms `json:"terms,omitempty"`
	FormationID string         `json:"formation_id,omitempty"`
	TargetSeat  int            `json:"target_seat,omitempty"`
	Price       int            `json:"price,omitempty"`
	RandomWords []string       `json:"random_words,omitempty"`
	GameID      string         `json:"game_id"`
	ID          string         `json:"id"`
	Kind        string         `json:"kind"`
	Actor       int            `json:"actor"`
	WindowID    string         `json:"window_id"`
	DecisionID  string         `json:"decision_id"`
	Cards       []string       `json:"cards"`
	Targets     []string       `json:"targets"`
	AceID       string         `json:"ace_id"`
	Value       int            `json:"value"`
}
type Event struct {
	Kind   string `json:"kind"`
	Actor  int    `json:"actor"`
	Amount Amount `json:"amount"`
}
type PendingAction struct {
	Baron           bool            `json:"baron,omitempty"`
	Bomb            bool            `json:"bomb,omitempty"`
	Exile           bool            `json:"exile,omitempty"`
	Infiltrator     bool            `json:"infiltrator,omitempty"`
	TargetSuit      Suit            `json:"target_suit,omitempty"`
	Opening         *OpeningIntent  `json:"opening,omitempty"`
	Ability         *AbilityIntent  `json:"ability,omitempty"`
	ProposalID      string          `json:"proposal_id,omitempty"`
	Quantity        int             `json:"quantity,omitempty"`
	JusticeQueen    string          `json:"justice_queen"`
	Kind            string          `json:"kind"`
	ID              string          `json:"id"`
	Actor           int             `json:"actor"`
	Defender        int             `json:"defender"`
	Clubs           []string        `json:"clubs"`
	Targets         []string        `json:"targets"`
	Responders      []int           `json:"responders"`
	Cursor          int             `json:"cursor"`
	AttackModifier  int             `json:"attack_modifier"`
	DefenseModifier int             `json:"defense_modifier"`
	Aces            []string        `json:"aces"`
	Decision        *EffectDecision `json:"decision"`
}
type EffectDecision struct {
	Opponents []int      `json:"opponents"`
	Cursor    int        `json:"cursor"`
	Reserved  []string   `json:"reserved"`
	QueenID   string     `json:"queen_id"`
	ID        string     `json:"id"`
	EffectID  string     `json:"effect_id"`
	Actor     int        `json:"actor"`
	Kind      string     `json:"kind"`
	Stage     string     `json:"stage"`
	Choices   [][]string `json:"choices"`
	Selected  []string   `json:"selected"`
	AceID     string     `json:"ace_id"`
	AceActor  int        `json:"ace_actor"`
}
type DecisionReceipt struct {
	ID        string `json:"id"`
	CommandID string `json:"command_id"`
}

func cloneTransition(n *State, s State) {
	if s.Pending != nil {
		p := *s.Pending
		p.Opening = cloneOpening(s.Pending.Opening)
		p.Ability = cloneAbilityIntent(s.Pending.Ability)
		p.Clubs = slices.Clone(p.Clubs)
		p.Targets = slices.Clone(p.Targets)
		p.Responders = slices.Clone(p.Responders)
		p.Aces = slices.Clone(p.Aces)
		if p.Decision != nil {
			d := *p.Decision
			d.Selected = slices.Clone(d.Selected)
			d.Opponents = slices.Clone(d.Opponents)
			d.Reserved = slices.Clone(d.Reserved)
			d.Choices = make([][]string, len(d.Choices))
			for i := range d.Choices {
				d.Choices[i] = slices.Clone(p.Decision.Choices[i])
			}
			p.Decision = &d
		}
		n.Pending = &p
	}
	n.Decisions = slices.Clone(s.Decisions)
	n.Commands = map[string]string{}
	for k, v := range s.Commands {
		n.Commands[k] = v
	}
}
func Digest(v any) string {
	h, e := canonical.Hash(v)
	if e != nil {
		panic("nonserializable engine digest: " + e.Error())
	}
	return h
}

var ErrUnsupported = errors.New("unsupported transition")

// ErrStateBudget means the next atomic state exceeds the canonical replay
// representation budget; the prior committed state and events are retained.
var ErrStateBudget = errors.New("state serialization budget exhausted")

func Apply(s State, c Command) (State, []Event, error) {
	if c.GameID != s.GameID || c.ID == "" || c.Actor < 1 || c.Actor > len(s.Players) {
		return s, nil, errors.New("invalid command identity")
	}
	key, hashErr := canonical.Hash(c)
	if hashErr != nil {
		return s, nil, errors.New("invalid command encoding")
	}
	if old, ok := s.Commands[c.ID]; ok {
		if old != key {
			return s, nil, errors.New("command conflict")
		}
		return s.Clone(), nil, nil
	}
	if c.Kind != "coup" && (len(s.DrawQueue) > 0 || s.Pending == nil && postActionWorkPending(s)) {
		return s, nil, errors.New("deferred draws pending")
	}
	if s.Phase != "playing" {
		return s, nil, errors.New("board closed")
	}
	n := s
	var events []Event
	var err error
	switch c.Kind {
	case "resolve-ready":
		n, events, err = resolveReady(s, c)
	case "expose-unused-ace":
		n, events, err = exposeUnusedAce(s, c)
	case "return-loan":
		n, err = declareLoanReturn(s, c)
	case "open-series", "open-formation", "take-back", "attach":
		n, err = declareOpening(s, c)
	case "ringleader", "richer-sacrifice", "barricade-sacrifice", "dexter-assassination", "fate", "code", "compensation", "main-inflation", "kidnapper":
		n, err = DeclareAbility(s, c)
	case "kidnapper-outcome":
		if s.Pending == nil || s.Pending.ID != c.WindowID || s.Pending.Actor != c.Actor || len(c.Cards) > 1 {
			return s, nil, errors.New("stale Kidnapper outcome")
		}
		selected := ""
		if len(c.Cards) == 1 {
			selected = c.Cards[0]
		}
		n, events, err = ResolveKidnapper(s, selected)
	case "ability-outcome":
		if s.Pending == nil || s.Pending.ID != c.WindowID || s.Pending.Actor != c.Actor {
			return s, nil, errors.New("stale ability outcome")
		}
		n, events, err = ResolveAbility(s, c.Cards, c.Targets)
	case "offer", "withdraw-offer", "decline-offer", "accept-offer":
		n, err = proposalCommand(s, c)
	case "purchase":
		n, err = declarePurchase(s, c)
	case "purchase-outcome":
		n, events, err = resolvePurchase(s, c)
	case "ponzi":
		n, err = declarePonzi(s, c)
	case "attack", "infiltrator-attack", "baron-attack", "bomb-attack", "baron-bomb-attack":
		n, err = attack(s, c)
	case "justice":
		n, err = justice(s, c)
	case "coup":
		n, events, err = coup(s, c)
	case "decision-closed":
		n, events, err = answerClosedJustice(s, c)
	case "pass", "compensation-response", "numerical-defense", "advancement-defense", "confinement", "inflation", "negotiate", "decision":
		if s.Pending == nil || c.WindowID != s.Pending.ID {
			return s, nil, errors.New("stale window")
		}
		p := s.Pending
		if c.Kind == "decision" {
			n, events, err = answerDecision(s, c)
		} else {
			if p.Decision != nil || p.Cursor >= len(p.Responders) || p.Responders[p.Cursor] != c.Actor || !eligible(s, c.Actor) {
				return s, nil, errors.New("wrong responder")
			}
			switch c.Kind {
			case "pass":
				n, events, err = advanceResponse(s)
			case "negotiate":
				n, err = negotiate(s, c)
			default:
				n, events, err = aceResponse(s, c)
			}
		}
	default:
		return s, nil, ErrUnsupported
	}
	if err != nil {
		return s, nil, err
	}
	n = n.Clone()
	n.Version++
	n.Commands[c.ID] = key
	if _, e := canonical.Marshal(n); e != nil {
		return s, nil, fmt.Errorf("%w: %v", ErrStateBudget, e)
	}
	return n, events, nil
}
