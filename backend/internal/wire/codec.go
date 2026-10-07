// Package wire converts authorized internal projections to online capabilities.
package wire

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"sort"
	"strings"
)

var ErrInvalid = errors.New("INVALID_COMMAND")

type Codec struct{ key []byte }

func New(key []byte) (*Codec, error) {
	if len(key) < 32 {
		return nil, errors.New("handle key requires at least32bytes")
	}
	return &Codec{append([]byte(nil), key...)}, nil
}
func (c *Codec) handle(match, actor, gameID, scope, id string) string {
	h := hmac.New(sha256.New, c.key)
	b, _ := json.Marshal([]string{match, actor, gameID, scope, id})
	h.Write(b)
	return "h_" + base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// References are issued only from an authorized projection. Historical references
// have a distinct namespace and never authorize current card selection.
func (c *Codec) project(match, actor string, raw []byte) (matchstore.Projection, map[string]string, error) {
	var p matchstore.Projection
	if json.Unmarshal(raw, &p) != nil || p.Board.Seat < 1 || p.Board.GameID == "" {
		return p, nil, ErrInvalid
	}
	refs := map[string]string{}
	o := &p.Board
	scope := fmt.Sprintf("view/%d/%s/%s", o.Version, o.WindowID, o.DecisionID)
	card := func(id string, live bool) string {
		if id == "" {
			return ""
		}
		s := scope
		if !live {
			s = "history"
		}
		h := c.handle(match, actor, o.GameID, s, id)
		if live {
			refs[h] = id
		}
		return h
	}
	ids := func(v []string, live bool) {
		for i := range v {
			v[i] = card(v[i], live)
		}
	}
	// Every current authorized physical location is issued separately from history.
	ownCards := map[string]bool{}
	for i := range o.Cards {
		if o.Cards[i].Controller == o.Seat {
			ownCards[o.Cards[i].Card.ID] = true
		}
		o.Cards[i].Card.ID = card(o.Cards[i].Card.ID, true)
	}
	for i := range o.OwnBindings {
		for j, id := range o.OwnBindings[i].Kings {
			// Metadata may reuse an already authorized own physical capability;
			// it must never mint a capability for a concealed counterpart.
			if !ownCards[id] {
				return p, nil, ErrInvalid
			}
			o.OwnBindings[i].Kings[j] = c.handle(match, actor, o.GameID, scope, id)
		}
	}
	if combat := p.Online.RulesContext.Combat; combat != nil {
		// Reuse capabilities already issued for this board. Combat context must
		// never turn a hidden pending reference into a new selectable card.
		for _, operands := range [][]string{combat.AttackCards, combat.TargetCards} {
			for i, id := range operands {
				h := c.handle(match, actor, o.GameID, scope, id)
				if _, ok := refs[h]; !ok {
					return p, nil, ErrInvalid
				}
				operands[i] = h
			}
		}
	}
	for i := range o.KnownReserved {
		o.KnownReserved[i].ID = card(o.KnownReserved[i].ID, true)
	}
	for i := range o.Choices {
		ids(o.Choices[i], true)
	}
	ids(o.PublicHistory, false)
	for i := range o.Effects {
		o.Effects[i].CardID = card(o.Effects[i].CardID, true)
	}
	for i := range o.Formations {
		f := &o.Formations[i]
		ids(f.Spec.Cards, true)
		if f.Spec.Substitute != nil {
			ids(f.Spec.Substitute.Kings, true)
		}
	}
	for i := range o.Loans {
		ids(o.Loans[i].Cards, true)
	}
	// Private offers reveal exact terms only to their named parties. They are not
	// capabilities to act on the counterparty's hidden inventory.
	for i := range o.Proposals {
		ids(o.Proposals[i].Terms.Give, false)
		ids(o.Proposals[i].Terms.Receive, false)
	}
	o.Legal = nil
	return p, refs, nil
}
func (c *Codec) EncodeProjection(match, actor string, raw []byte) (json.RawMessage, error) {
	p, _, err := c.project(match, actor, raw)
	if err != nil {
		return nil, err
	}
	var source matchstore.Projection
	if json.Unmarshal(raw, &source) != nil {
		return nil, ErrInvalid
	}
	type disclosed struct {
		Handle string    `json:"handle"`
		Suit   game.Suit `json:"suit"`
		Rank   int       `json:"rank"`
	}
	faces := []disclosed{}
	known := map[string]bool{}
	for _, id := range source.Board.PublicHistory {
		known[id] = true
	}
	for _, offer := range source.Board.Proposals {
		for _, id := range append(append([]string{}, offer.Terms.Give...), offer.Terms.Receive...) {
			known[id] = true
		}
	}
	for _, card := range game.Deck() {
		if known[card.ID] {
			faces = append(faces, disclosed{c.handle(match, actor, source.Board.GameID, "history", card.ID), card.Suit, card.Rank})
		}
	}
	players, err := matchCardStyles(match, p.Board.Players)
	if err != nil {
		return nil, err
	}
	type styledBoard struct {
		game.Observation
		Players []styledPlayer `json:"players"`
	}
	return json.Marshal(struct {
		matchstore.Projection
		Board      styledBoard `json:"board"`
		KnownFaces []disclosed `json:"known_faces"`
	}{p, styledBoard{p.Board, players}, faces})
}

type styledPlayer struct {
	game.PlayerView
	CardStyle string `json:"card_style"`
}

// matchCardStyles decorates only the authorized online response. The immutable,
// public match ID supplies cosmetic variation; private chance, card identity,
// game number and viewer identity never enter this separate-purpose hash.
// Keeping this out of checkpoints/engine observations preserves replay digests.
func matchCardStyles(match string, players []game.PlayerView) ([]styledPlayer, error) {
	if len(players) < 3 || len(players) > 4 {
		return nil, ErrInvalid
	}
	editions := []string{"first-light", "rain-glaze", "ember-glaze", "drypoint"}
	digest := func(edition string) [sha256.Size]byte {
		return sha256.Sum256([]byte("cgms/match-card-style/v1\x00" + match + "\x00" + edition))
	}
	sort.SliceStable(editions, func(i, j int) bool {
		a, b := digest(editions[i]), digest(editions[j])
		return bytes.Compare(a[:], b[:]) < 0
	})
	seen := map[int]bool{}
	out := make([]styledPlayer, len(players))
	for i, player := range players {
		if player.Seat < 1 || player.Seat > len(players) || seen[player.Seat] {
			return nil, ErrInvalid
		}
		seen[player.Seat] = true
		out[i] = styledPlayer{player, editions[player.Seat-1]}
	}
	return out, nil
}

type request struct {
	CommandID       string          `json:"command_id"`
	GameID          string          `json:"game_id"`
	ExpectedVersion int64           `json:"expected_version"`
	Type            string          `json:"type"`
	Payload         json.RawMessage `json:"payload"`
}

// payload deliberately has no actor, game identity, command identity, random
// words, settlement budget, or scheduler-controlled fields.
type payload struct {
	WindowID    string              `json:"pending_action_id,omitempty"`
	DecisionID  string              `json:"decision_id,omitempty"`
	Cards       []string            `json:"selection,omitempty"`
	Targets     []string            `json:"targets,omitempty"`
	AceID       string              `json:"ace,omitempty"`
	Value       int                 `json:"value,omitempty"`
	Quantity    int                 `json:"quantity,omitempty"`
	Payment     []string            `json:"payment,omitempty"`
	Suit        game.Suit           `json:"suit,omitempty"`
	Formation   *game.FormationSpec `json:"formation,omitempty"`
	FormationID string              `json:"formation_id,omitempty"`
	OfferID     string              `json:"offer_id,omitempty"`
	Revision    int                 `json:"revision,omitempty"`
	Terms       *game.ProposalTerms `json:"terms,omitempty"`
	TargetSeat  int                 `json:"target_seat,omitempty"`
	Price       int                 `json:"price,omitempty"`
	Exile       bool                `json:"exile,omitempty"`
	Infiltrator bool                `json:"infiltrator,omitempty"`
	PromiseID   string              `json:"promise_id,omitempty"`
	Recipient   int                 `json:"recipient,omitempty"`
	AwardID     string              `json:"award_id,omitempty"`
	Mode        string              `json:"mode,omitempty"`
	Amount      game.Amount         `json:"amount,omitempty"`
	Accept      bool                `json:"accept,omitempty"`
	DebtID      string              `json:"debt_id,omitempty"`
	Threshold   string              `json:"threshold,omitempty"`
	Condition   *condition          `json:"condition,omitempty"`
}
type condition struct {
	Kind    string `json:"kind"`
	GameID  string `json:"game_id"`
	AwardID string `json:"award_id"`
	Payer   int    `json:"payer"`
}

func strict(b []byte, v any) error {
	if canonical.DecodeLimit(b, v, 65536) != nil {
		return ErrInvalid
	}
	return nil
}

func (c *Codec) DecodeIntent(match, actor string, body, authorizedProjection []byte) (matchstore.Intent, error) {
	var r request
	var p payload
	if strict(body, &r) != nil || strict(r.Payload, &p) != nil || r.CommandID == "" || r.GameID == "" || r.ExpectedVersion < 0 {
		return matchstore.Intent{}, ErrInvalid
	}

	var outer map[string]json.RawMessage
	if json.Unmarshal(body, &outer) != nil {
		return matchstore.Intent{}, ErrInvalid
	}
	for _, key := range []string{"command_id", "game_id", "expected_version", "type", "payload"} {
		value, ok := outer[key]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return matchstore.Intent{}, ErrInvalid
		}
	}
	if trimmed := bytes.TrimSpace(r.Payload); len(trimmed) == 0 || trimmed[0] != '{' {
		return matchstore.Intent{}, ErrInvalid
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(r.Payload, &fields) != nil {
		return matchstore.Intent{}, ErrInvalid
	}
	allowed, known := commandFields[r.Type]
	if !known {
		return matchstore.Intent{}, ErrInvalid
	}
	for key := range fields {
		if !strings.Contains(" "+allowed+" ", " "+key+" ") {
			return matchstore.Intent{}, ErrInvalid
		}
	}
	if authorizedProjection == nil {
		return matchstore.Intent{ID: r.CommandID, GameID: r.GameID, ExpectedVersion: r.ExpectedVersion, Transport: append(json.RawMessage(nil), body...)}, nil
	}
	view, refs, err := c.project(match, actor, authorizedProjection)
	if err != nil || view.Board.GameID != r.GameID {
		return matchstore.Intent{}, ErrInvalid
	}
	resolve := func(v *string) error {
		if *v == "" {
			return nil
		}
		id, ok := refs[*v]
		if !ok {
			return ErrInvalid
		}
		*v = id
		return nil
	}
	resolveAll := func(v []string) error {
		for i := range v {
			if resolve(&v[i]) != nil {
				return ErrInvalid
			}
		}
		return nil
	}
	for _, v := range [][]string{p.Cards, p.Targets, p.Payment} {
		if resolveAll(v) != nil {
			return matchstore.Intent{}, ErrInvalid
		}
	}
	if resolve(&p.AceID) != nil {
		return matchstore.Intent{}, ErrInvalid
	}
	if p.Formation != nil {
		if resolveAll(p.Formation.Cards) != nil {
			return matchstore.Intent{}, ErrInvalid
		}
		if p.Formation.Substitute != nil && resolveAll(p.Formation.Substitute.Kings) != nil {
			return matchstore.Intent{}, ErrInvalid
		}
	}
	if p.Terms != nil {
		if resolveAll(p.Terms.Give) != nil || resolveAll(p.Terms.Receive) != nil {
			return matchstore.Intent{}, ErrInvalid
		}
	}
	out := matchstore.Intent{ID: r.CommandID, GameID: r.GameID, ExpectedVersion: r.ExpectedVersion, Transport: append(json.RawMessage(nil), body...)}
	seat := view.Board.Seat
	switch r.Type {
	case "end-turn", "declare-ordinary", "coup", "departure-choice", "finish-settlement", "promise-offer", "promise-accept", "promise-final-offer", "promise-final-answer", "promise-pay", "promise-refuse", "voluntary-transfer", "forgive", "nullify":
		if r.Type == "promise-offer" && (p.Condition == nil || p.Condition.Kind != "award-occurrence" || p.Condition.GameID != r.GameID || p.Condition.AwardID != p.AwardID || p.Condition.Payer != seat) {
			return matchstore.Intent{}, ErrInvalid
		}
		out.Request.Operation = &game.LifecycleOperation{ID: r.CommandID, GameID: r.GameID, Actor: seat, Kind: r.Type, PromiseID: p.PromiseID, Recipient: p.Recipient, AwardID: p.AwardID, Mode: p.Mode, Amount: p.Amount, Accept: p.Accept, DebtID: p.DebtID, Threshold: p.Threshold}
	case "open-series", "open-formation", "take-back", "attach", "ringleader", "richer-sacrifice", "barricade-sacrifice", "dexter-assassination", "fate", "code", "compensation", "main-inflation", "kidnapper", "kidnapper-outcome", "offer", "withdraw-offer", "decline-offer", "accept-offer", "purchase", "ponzi", "attack", "infiltrator-attack", "baron-attack", "bomb-attack", "baron-bomb-attack", "justice", "decision-closed", "pass", "compensation-response", "numerical-defense", "advancement-defense", "confinement", "inflation", "negotiate", "decision", "return-loan", "expose-unused-ace":
		if r.Type == "purchase" {
			if p.Quantity < 1 || len(p.Payment) == 0 || len(p.Cards) > 0 || p.Value != 0 {
				return matchstore.Intent{}, ErrInvalid
			}
			p.Value = p.Quantity
			p.Cards = p.Payment
		}
		out.Request.Command = &game.Command{ID: r.CommandID, GameID: r.GameID, Actor: seat, Kind: r.Type, WindowID: p.WindowID, DecisionID: p.DecisionID, Cards: p.Cards, Targets: p.Targets, AceID: p.AceID, Value: p.Value, Suit: p.Suit, Formation: p.Formation, FormationID: p.FormationID, OfferID: p.OfferID, Revision: p.Revision, Terms: p.Terms, TargetSeat: p.TargetSeat, Price: p.Price, Exile: p.Exile, Infiltrator: p.Infiltrator}
	default:
		return matchstore.Intent{}, ErrInvalid
	}
	return out, nil
}

// Each wire action has a closed field set; ignored extras are never accepted.
var commandFields = map[string]string{
	"end-turn": "", "coup": "", "nullify": "", "finish-settlement": "", "declare-ordinary": "threshold", "departure-choice": "accept",
	"promise-offer": "promise_id recipient award_id mode amount condition", "promise-accept": "promise_id", "promise-final-offer": "promise_id amount", "promise-final-answer": "promise_id accept", "promise-pay": "promise_id", "promise-refuse": "promise_id", "voluntary-transfer": "recipient amount", "forgive": "debt_id amount",
	"open-series": "suit selection", "open-formation": "formation formation_id", "take-back": "selection formation_id", "attach": "selection suit", "ringleader": "selection ace value", "richer-sacrifice": "selection", "barricade-sacrifice": "selection", "dexter-assassination": "selection targets", "fate": "selection target_seat formation_id", "code": "formation_id value price", "compensation": "ace", "main-inflation": "ace target_seat", "kidnapper": "formation_id target_seat", "kidnapper-outcome": "selection pending_action_id",
	"offer": "offer_id revision terms", "withdraw-offer": "offer_id revision", "decline-offer": "offer_id revision", "accept-offer": "offer_id revision", "purchase": "quantity payment", "ponzi": "value", "attack": "selection targets ace value suit exile infiltrator", "infiltrator-attack": "selection targets ace value suit exile infiltrator", "baron-attack": "selection targets ace value suit exile infiltrator", "bomb-attack": "selection targets ace value suit exile infiltrator", "baron-bomb-attack": "selection targets ace value suit exile infiltrator", "justice": "ace", "decision-closed": "pending_action_id decision_id", "pass": "pending_action_id", "compensation-response": "pending_action_id ace", "numerical-defense": "pending_action_id ace value", "advancement-defense": "pending_action_id ace", "confinement": "pending_action_id ace", "inflation": "pending_action_id ace", "negotiate": "pending_action_id", "decision": "pending_action_id decision_id selection ace value suit target_seat", "return-loan": "selection target_seat", "expose-unused-ace": "ace",
}
