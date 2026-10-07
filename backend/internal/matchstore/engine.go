// Package matchstore adapts the pure game engine to private durable checkpoints.
package matchstore

import (
	"crypto/rand"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/game"
	"math/big"
	"slices"
	"strconv"
)

const EnvelopeSchema = "cgms-match-envelope-v1"

// EconomyEnvelopeSchema fences old readers which do not enforce pinned economy.
// The game engine is unchanged; the online persistence contract is different.
const EconomyEnvelopeSchema = "cgms-match-envelope-economy-v1"
const EngineVersion = "cgms-lifecycle-adapter-v1"

// Envelope is private storage, never a client response. Chance contains committed
// server selections, not a reproducible public seed. The store must atomically
// commit it with the engine transition and idempotency receipt.
type chanceReplay struct{ used map[string]bool }

type Envelope struct {
	Schema        string              `json:"schema"`
	EngineVersion string              `json:"engine_version"`
	RulesHash     string              `json:"rules_hash"`
	Match         game.MatchLifecycle `json:"match"`
	Chance        map[string][]string `json:"chance"`
	BotDifficulty string              `json:"bot_difficulty,omitempty"`
	replay        *chanceReplay
}

func NewEnvelope(m game.MatchLifecycle, rulesHash string) (Envelope, error) {
	e := Envelope{Schema: EnvelopeSchema, EngineVersion: EngineVersion, RulesHash: rulesHash, Match: m.Clone(), Chance: map[string][]string{}}
	return e, e.Validate()
}
func (e Envelope) Validate() error {
	if e.BotDifficulty != "" && e.BotDifficulty != "beginner" && e.BotDifficulty != "standard" && e.BotDifficulty != "advanced" {
		return errors.New("invalid bot difficulty")
	}
	if (e.Schema != EnvelopeSchema && e.Schema != EconomyEnvelopeSchema) || e.EngineVersion != EngineVersion || e.RulesHash == "" || e.Match.Game.Schema != "cgms-lifecycle-v1" {
		return errors.New("incompatible lifecycle checkpoint")
	}
	if e.Match.GameLimit < 1 || len(e.Match.Instances) == 0 || !slices.Contains(e.Match.Instances, e.Match.Game.Board.GameID) {
		return errors.New("invalid match identity")
	}
	seen := map[string]bool{}
	for _, id := range e.Match.Instances {
		if id == "" || seen[id] {
			return errors.New("duplicate match instance")
		}
		seen[id] = true
	}
	if err := e.Match.StartLedger.Validate(); err != nil {
		return err
	}
	return e.Match.Game.Validate()
}
func (e Envelope) Clone() Envelope {
	n := e
	n.Match = e.Match.Clone()
	n.Chance = make(map[string][]string, len(e.Chance))
	for k, v := range e.Chance {
		n.Chance[k] = slices.Clone(v)
	}
	return n
}

type Request struct {
	Command   *game.Command            `json:"command,omitempty"`
	Operation *game.LifecycleOperation `json:"operation,omitempty"`
}

// ValidateActorRequest checks a generated candidate using the same authority
// boundary as Submit. It commits nothing and returns no private outcome. Bot
// policy sees only authorized observations; its chosen intent still uses Submit.
func ValidateActorRequest(e Envelope, seat int, r Request) error {
	_, err := applyRequest(e, seat, r)
	return err
}

// applyRequest never substitutes actor/game/window identities. The store checks
// its authenticated actor receipt before invoking this new-intent boundary.
func applyRequest(e Envelope, seat int, r Request) (Envelope, error) {
	if err := e.Validate(); err != nil {
		return e, err
	}
	if seat < 1 || seat > len(e.Match.Game.Board.Players) || (r.Command == nil) == (r.Operation == nil) {
		return e, errors.New("invalid actor request")
	}
	n := e.Clone()
	var err error
	if r.Command != nil {
		c := *r.Command
		if c.Actor != seat || c.GameID != e.Match.Game.Board.GameID || len(c.RandomWords) > 0 {
			return e, errors.New("unauthorized command identity or randomness")
		}
		switch c.Kind {
		case "resolve-ready", "purchase-outcome", "ability-outcome":
			return e, errors.New("server command required")
		}

		if c.Kind == "kidnapper-outcome" {
			p := n.Match.Game.Board.Pending
			if p == nil || p.Ability == nil || len(game.KidnapperClosedCandidates(n.Match.Game.Board, p.Ability.TargetSeat)) > 0 {
				return e, errors.New("not an exposed Kidnapper choice")
			}
		}
		if c.Kind == "decision-closed" {
			p := n.Match.Game.Board.Pending
			if p == nil || p.Decision == nil || p.Decision.Kind != "justice" || c.WindowID != p.ID || c.DecisionID != p.Decision.ID || c.Actor != p.Decision.Actor || len(c.Cards) > 0 {
				return e, errors.New("invalid closed decision")
			}
			var pool []string
			for _, choice := range p.Decision.Choices {
				if len(choice) != 1 {
					return e, errors.New("invalid hidden choice")
				}
				card, _ := n.Match.Game.Board.Card(choice[0])
				if card.Zone == game.Hand || card.Zone == game.ConcealedAce {
					pool = append(pool, choice[0])
				}
			}
			if len(pool) > 0 {
				var picked []string
				n, picked, err = SecurePermutation(n, c.GameID+"/"+c.ID+"/justice", pool)
				if err != nil {
					return e, err
				}
				count := uint64(len(pool))
				word := uint64(slices.Index(pool, picked[0]))
				threshold := -count % count
				if word < threshold {
					word += count
				}
				c.RandomWords = []string{strconv.FormatUint(word, 10)}
			}
		}
		n.Match.Game, _, err = n.Match.Game.ApplyBoardCommand(c)
	} else {
		op := *r.Operation
		if op.Actor != seat || op.GameID != e.Match.Game.Board.GameID || len(op.Recycle) > 0 || len(op.InitiativeDraws) > 0 || len(op.ReturnOrder) > 0 || len(op.Seats) > 0 || op.Budget != 0 {
			return e, errors.New("unauthorized operation fields")
		}
		switch op.Kind {
		case "end-turn", "declare-ordinary", "coup", "departure-choice", "finish-settlement", "promise-offer", "promise-accept", "promise-final-offer", "promise-final-answer", "promise-pay", "promise-refuse", "voluntary-transfer", "forgive", "nullify":
		default:
			return e, errors.New("server operation required")
		}
		n.Match, err = n.Match.ApplyOperation(op)
	}
	if err != nil {
		return e, err
	}
	if err = n.Validate(); err != nil {
		return e, err
	}
	return n, nil
}

// ServerRequest is an internal scheduler capability, not an actor request. All
// random fields are generated below from current engine state using crypto/rand.
type ServerRequest struct {
	ID         string                   `json:"id"`
	GameID     string                   `json:"game_id"`
	Kind       string                   `json:"kind"`
	Operation  *game.LifecycleOperation `json:"operation,omitempty"`
	NextGameID string                   `json:"next_game_id,omitempty"`
}

func applyServerRequest(e Envelope, r ServerRequest) (Envelope, error) {
	if err := e.Validate(); err != nil {
		return e, err
	}
	if r.ID == "" || r.GameID != e.Match.Game.Board.GameID {
		return e, errors.New("invalid server identity")
	}
	n := e.Clone()
	s := n.Match.Game.Board
	var err error
	perm := func(label string, cards []string) []string {
		if err != nil {
			return nil
		}
		var p []string
		n, p, err = SecurePermutation(n, r.GameID+"/"+r.ID+"/"+label, cards)
		return p
	}
	initiative := func(board game.State) ([]string, []string) {
		draws := slices.Clone(board.DrawOrder)
		pool := zoneIDs(board, game.Discard)
		if len(pool) > 0 {
			draws = append(draws, perm("initiative-discard", pool)...)
		}
		returned, x := game.InitiativeReturnSupply(board, draws)
		if x != nil {
			err = x
			return nil, nil
		}
		var order []string
		if len(returned) > 0 {
			order = perm("initiative-return", returned)
		}
		return draws, order
	}
	switch r.Kind {
	case "deal":
		p := perm("deal", zoneIDs(s, game.Draw))
		if err == nil {
			n.Match.Game.Board, _, err = game.DealAttempt(s, p)
		}
	case "initiative":
		draws, order := initiative(s)
		if err == nil {
			n.Match.Game.Board, err = game.ResolveInitiative(s, draws, order)
		}
	case "after-action":
		var shuffle, recycle []string
		if s.NeedsShuffle {
			shuffle = perm("draw", zoneIDs(s, game.Draw))
		} else if len(s.DrawOrder) == 0 {
			recycle = perm("recycle", zoneIDs(s, game.Discard))
		}
		if err == nil {
			n.Match.Game, err = n.Match.Game.ContinueAfterAction(shuffle, recycle)
		}
	case "operation":
		if r.Operation == nil {
			return e, errors.New("missing server operation")
		}
		op := *r.Operation
		if op.ID != r.ID || op.GameID != r.GameID || len(op.Recycle) > 0 || len(op.InitiativeDraws) > 0 || len(op.ReturnOrder) > 0 || len(op.Seats) > 0 {
			return e, errors.New("server operation retarget or supplied randomness")
		}
		switch op.Kind {
		case "begin-turn":
			if len(s.DrawOrder) == 0 {
				op.Recycle = perm("recycle", zoneIDs(s, game.Discard))
			}
		case "begin-round":
			s.Round++
			op.InitiativeDraws, op.ReturnOrder = initiative(s)
		case "resume-automatic":
			if op.Budget < 1 || op.Budget > 10000 {
				return e, errors.New("invalid bounded settlement budget")
			}
		case "finalize":
		default:
			return e, errors.New("not a scheduler operation")
		}
		if err == nil {
			n.Match, err = n.Match.ApplyOperation(op)
		}
	case "next-game":
		n.Match, err = n.Match.NextGame(r.NextGameID)
	case "pending-chance":
		p := s.Pending
		if p == nil {
			return e, errors.New("missing pending action")
		}
		c := game.Command{ID: r.ID, GameID: r.GameID, Actor: p.Actor, WindowID: p.ID, Kind: "resolve-ready"}
		candidate := s
		quantity := 0
		switch p.Kind {
		case "purchase":
			c.Kind = "purchase-outcome"
			candidate, err = s.Move(p.Targets, 0, game.Draw, false)
			quantity = p.Quantity
		case "ability":
			if p.Ability == nil {
				return e, errors.New("missing ability")
			}
			a := p.Ability
			c.Kind = "ability-outcome"
			switch a.Kind {
			case "kidnapper":
				c.Kind = "kidnapper-outcome"
				pool := game.KidnapperClosedCandidates(s, a.TargetSeat)
				if len(pool) > 0 {
					c.Cards = perm("kidnapper", pool)
					if len(c.Cards) > 0 {
						c.Cards = c.Cards[:1]
					}
				} else if len(game.KidnapperOpenCandidates(s, a.TargetSeat)) > 0 {
					return e, errors.New("requires actor exposed choice")
				}
			case "richer-sacrifice":
				candidate, err = s.Move(a.Cards, 0, game.Discard, false)
				quantity = 2
			case "barricade-sacrifice":
				if len(a.Cards) < 1 {
					return e, errors.New("missing sacrifice")
				}
				candidate, err = s.Move(a.Cards[:1], 0, game.Discard, false)
				if err == nil {
					candidate, err = candidate.Move(a.Cards[1:], 0, game.Draw, false)
				}
				quantity = 5
			case "fate":
				candidate, err = s.Move(a.Cards, 0, game.Draw, false)
				if err == nil {
					candidate, quantity, err = candidate.FateReturn(a.TargetSeat)
				}
			default:
				c.Kind = "resolve-ready"
			}
		}
		if err == nil && quantity > 0 {
			if candidate.NeedsShuffle {
				c.Cards = perm("effect-draw", zoneIDs(candidate, game.Draw))
			}
			if quantity > len(candidate.DrawOrder) && len(zoneIDs(candidate, game.Discard)) > 0 {
				c.Targets = perm("effect-recycle", zoneIDs(candidate, game.Discard))
			}
		}
		if err == nil {
			n.Match.Game, _, err = n.Match.Game.ApplyBoardCommand(c)
		}
	default:
		return e, errors.New("unsupported server continuation")
	}
	if err != nil {
		return e, err
	}
	if err = n.Validate(); err != nil {
		return e, err
	}
	return n, nil
}
func zoneIDs(s game.State, z game.Zone) []string {
	var out []string
	for _, c := range s.Cards {
		if c.Zone == z {
			out = append(out, c.Card.ID)
		}
	}
	return out
}

// SecurePermutation binds a private choice to its candidate set. Reusing the
// choice key after restart returns the committed order or rejects retargeting.
func SecurePermutation(e Envelope, key string, input []string) (Envelope, []string, error) {
	if key == "" {
		return e, nil, errors.New("empty chance identity")
	}
	sorted := slices.Clone(input)
	slices.Sort(sorted)
	for i, v := range sorted {
		if v == "" || (i > 0 && sorted[i-1] == v) {
			return e, nil, errors.New("invalid chance candidates")
		}
	}
	if old, ok := e.Chance[key]; ok {
		if e.replay != nil {
			e.replay.used[key] = true
		}
		copy := slices.Clone(old)
		slices.Sort(copy)
		if !slices.Equal(copy, sorted) {
			return e, nil, errors.New("chance identity retarget")
		}
		return e.Clone(), slices.Clone(old), nil
	}
	if e.replay != nil {
		return e, nil, errors.New("missing committed replay outcome")
	}
	out := slices.Clone(input)
	for i := len(out) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return e, nil, err
		}
		out[i], out[j.Int64()] = out[j.Int64()], out[i]
	}
	n := e.Clone()
	n.Chance[key] = slices.Clone(out)
	return n, out, nil
}

// Projection contains only authorized seat information, not the private envelope.
type Projection struct {
	Online   OnlineDetails        `json:"online"`
	Board    game.Observation     `json:"board"`
	Cash     game.Amount          `json:"cash"`
	Score    game.Amount          `json:"score"`
	Debts    []game.FinancialDebt `json:"debts"`
	Promises []game.Promise       `json:"promises"`
}

func SeatProjection(e Envelope, seat int) (Projection, error) {
	if err := e.Validate(); err != nil {
		return Projection{}, err
	}
	o, err := game.ObserveWithoutMenu(e.Match.Game.Board, seat)
	if err != nil {
		return Projection{}, err
	}
	l := e.Match.Game.Clone()
	p := Projection{Online: onlineProjection(e, seat), Board: o, Cash: l.Ledger.Cash[seat-1], Score: l.Ledger.Scores[seat-1]}
	for _, d := range l.Ledger.Debts {
		if d.Debtor == seat-1 || d.Creditor == seat-1 {
			p.Debts = append(p.Debts, d)
		}
	}
	for _, v := range l.Promises.Promises {
		if v.Payer == seat-1 || v.Recipient == seat-1 {
			p.Promises = append(p.Promises, v)
		}
	}
	return p, nil
}
