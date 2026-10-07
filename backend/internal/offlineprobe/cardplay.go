package offlineprobe

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/bots"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/game"
	"slices"
)

const CardplayVersion = "cgms-offline-cardplay-v1"

type CardplayStep struct {
	Kind     string        `json:"kind"`
	GameID   string        `json:"game_id"`
	Command  *game.Command `json:"command"`
	Order    []string      `json:"order"`
	NextGame string        `json:"next_game"`
	Seat     int           `json:"seat"`
}
type CardplayRequest struct {
	Schema string              `json:"schema"`
	Match  game.MatchLifecycle `json:"match"`
	Steps  []CardplayStep      `json:"steps"`
}
type CardplayObservation struct {
	Board       game.Observation      `json:"board"`
	Ledger      bots.LedgerProjection `json:"ledger"`
	OpeningUsed bool                  `json:"opening_used"`
	TurnStarted bool                  `json:"turn_started"`
	RoundClosed bool                  `json:"round_closed"`
	Ending      string                `json:"ending"`
}
type CardplayDecision struct {
	Kind     string    `json:"kind"`
	Actor    int       `json:"actor"`
	Suit     game.Suit `json:"suit"`
	WindowID string    `json:"window_id"`
}

// CardplayChoice is deliberately a passive, observation-only series policy.
// It is comparison workload coverage, never a competent/full-menu bot claim.
func CardplayChoice(o game.Observation, used, started, closed bool) *CardplayDecision {
	if o.Phase != "playing" || !started || closed {
		return nil
	}
	if o.WindowID != "" {
		if o.RequiredActor == o.Seat && o.DecisionID == "" {
			return &CardplayDecision{Kind: "pass", Actor: o.Seat, WindowID: o.WindowID}
		}
		return nil
	}
	if o.Active != o.Seat {
		return nil
	}
	for _, suit := range []game.Suit{game.Hearts, game.Clubs, game.Spades, game.Diamonds} {
		if used && !slices.Contains(o.Players[o.Seat-1].History, suit) {
			continue
		}
		for _, c := range o.Cards {
			if c.Controller == o.Seat && c.Zone == game.Hand && c.Card.Suit == suit && c.Card.Rank >= 2 && c.Card.Rank <= 10 && c.AvailableFromRound <= o.Round {
				return &CardplayDecision{Kind: "open-series", Actor: o.Seat, Suit: suit}
			}
		}
	}
	return &CardplayDecision{Kind: "end-turn", Actor: o.Seat}
}

type CardplayFrame struct {
	Match        game.MatchLifecycle   `json:"match"`
	Events       []game.Event          `json:"events"`
	Observations []CardplayObservation `json:"observations"`
	Decisions    []*CardplayDecision   `json:"decisions"`
}

func cardplayFrame(m game.MatchLifecycle, events []game.Event) (CardplayFrame, error) {
	f := CardplayFrame{Match: m, Events: events}
	for i, p := range m.Game.Board.Players {
		o, e := game.ObserveWithoutMenu(m.Game.Board, i+1)
		if e != nil {
			return f, e
		}
		l, e := bots.ProjectLedger(m.Game.Ledger, i)
		if e != nil {
			return f, e
		}
		f.Observations = append(f.Observations, CardplayObservation{o, l, p.OpeningUsed, m.Game.TurnStarted, m.Game.RoundClosed, m.Game.Ending})
		f.Decisions = append(f.Decisions, CardplayChoice(o, p.OpeningUsed, m.Game.TurnStarted, m.Game.RoundClosed))
	}
	return f, nil
}
func cardplayStep(m game.MatchLifecycle, s CardplayStep) (game.MatchLifecycle, []game.Event, error) {
	if m.Game.Validate() != nil || s.GameID != m.Game.Board.GameID {
		return m, nil, errors.New("invalid cardplay boundary")
	}
	n := m.Clone()
	var e error
	var events []game.Event
	switch s.Kind {
	case "deal":
		var retry bool
		n.Game.Board, retry, e = game.DealAttempt(n.Game.Board, s.Order)
		if retry {
			return m, nil, errors.New("deal guarantee")
		}
	case "initiative":
		if len(n.Game.Board.Order) > 0 {
			return m, nil, errors.New("already initialized")
		}
		n.Game.Board, e = game.ResolveInitiative(n.Game.Board, nil, s.Order)
	case "begin-turn":
		n.Game, _, e = n.Game.BeginTurn(nil)
	case "command":
		if s.Command == nil || (s.Command.Kind != "open-series" && s.Command.Kind != "pass") {
			return m, nil, errors.New("unsupported cardplay policy action")
		}
		n.Game, events, e = n.Game.ApplyBoardCommand(*s.Command)
	case "end-turn":
		n.Game, e = n.Game.EndTurn()
	case "begin-round":
		n.Game, e = n.Game.BeginRound(nil, s.Order)
	case "finish":
		n.Game, e = n.Game.FinishSettlement(s.Seat)
	case "finalize":
		n.Game, e = n.Game.Finalize()
	case "next":
		n, e = n.NextGame(s.NextGame)
	default:
		return m, nil, errors.New("unsupported cardplay operation")
	}
	if e != nil {
		return m, nil, e
	}
	// The comparison boundary groups bounded automatic work, never player inputs.
	for work := 0; n.Game.Automatic != nil; work++ {
		if work >= 8 {
			return m, nil, errors.New("automatic budget")
		}
		n.Game, e = n.Game.ResumeAutomatic(4096)
		if e != nil {
			return m, nil, e
		}
	}
	if e = n.Game.Validate(); e != nil {
		return m, nil, e
	}
	return n, events, nil
}

type CardplayResponse struct {
	Schema     string         `json:"schema"`
	Error      string         `json:"error,omitempty"`
	FailedStep int            `json:"failed_step"`
	Digests    []string       `json:"digests,omitempty"`
	Final      *CardplayFrame `json:"final,omitempty"`
}

func executeCardplay(raw []byte) []byte {
	fail := func(code string, i int) []byte {
		b, _ := canonical.Marshal(CardplayResponse{Schema: CardplayVersion, Error: code, FailedStep: i})
		return b
	}
	var r CardplayRequest
	if canonical.Decode(raw, &r) != nil || r.Schema != CardplayVersion || len(r.Steps) > 2048 || r.Match.GameLimit < 1 || r.Match.GameLimit > 3 {
		return fail("invalid_request", -1)
	}
	if !cardplayProfile(r.Match) {
		return fail("invalid_state", -1)
	}
	m := r.Match
	out := CardplayResponse{Schema: CardplayVersion, FailedStep: -1}
	var events []game.Event
	for i := -1; i < len(r.Steps); i++ {
		var e error
		if i >= 0 {
			m, events, e = cardplayStep(m, r.Steps[i])
			if e != nil {
				return fail("command_rejected", i)
			}
		}
		f, e := cardplayFrame(m, events)
		if e != nil {
			return fail("observation_failed", i)
		}
		out.Digests = append(out.Digests, game.Digest(f))
		out.Final = &f
	}
	b, e := canonical.Marshal(out)
	if e != nil {
		return fail("output_budget", -1)
	}
	return b
}

// Only the explicitly measured reachable profile is accepted; no unsupported
// saved effect may silently survive as if this candidate had adjudicated it.
func cardplayProfile(m game.MatchLifecycle) bool {
	l, b := m.Game, m.Game.Board
	if l.Validate() != nil || len(b.Formations) > 0 || len(b.Proposals) > 0 || b.Code != nil || len(b.Loans) > 0 || len(b.Effects) > 0 || len(b.Bindings) > 0 || len(b.DrawQueue) > 0 || b.NeedsShuffle || len(b.Decisions) > 0 || l.Automatic != nil || len(l.PendingReceipts) > 0 || len(l.DepartureChoices) > 0 || len(l.ForgivenessConsents) > 0 || l.EndTurnPending || l.PromisePayment != nil || l.PaymentPrior != nil || l.AutomaticPurpose != "" || len(l.Promises.Promises) > 0 || len(l.Promises.Payments) > 0 || len(l.Ledger.Debts) > 0 || len(l.Ledger.Charges) > 0 || len(l.Ledger.Corrections) > 0 || len(l.Ledger.Operations) > 0 || len(m.NullificationConsents) > 0 {
		return false
	}
	for _, p := range b.Players {
		if p.Confined || p.Departed || p.Underground || p.Ally != 0 || p.AceRound != 0 || p.CompensationUsed || len(p.Quotas) > 0 {
			return false
		}
	}
	for _, c := range b.Cards {
		if c.AvailableFromRound != 0 || c.UsedTurn != 0 || c.Allocation != "" || (c.Zone != game.Draw && c.Zone != game.Hand && c.Zone != game.ConcealedAce && c.Zone != game.Series) {
			return false
		}
	}
	if b.Pending != nil && (b.Pending.Kind != "opening" || b.Pending.Opening == nil || b.Pending.Opening.Kind != "open-series" || b.Pending.Decision != nil) {
		return false
	}
	return l.Ending == "" || l.Ending == "round-limit"
}

const CardplayChoiceVersion = "cgms-offline-cardplay-choice-v1"

type cardplayChoiceRequest struct {
	Schema      string              `json:"schema"`
	Observation CardplayObservation `json:"observation"`
}

func executeCardplayChoice(raw []byte) []byte {
	var r cardplayChoiceRequest
	if canonical.Decode(raw, &r) != nil || r.Schema != CardplayChoiceVersion || r.Observation.Board.Seat < 1 || r.Observation.Board.Seat > len(r.Observation.Board.Players) {
		return []byte(`{"error":"invalid_request"}`)
	}
	o := r.Observation
	b, e := canonical.Marshal(CardplayChoice(o.Board, o.OpeningUsed, o.TurnStarted, o.RoundClosed))
	if e != nil {
		return []byte(`{"error":"invalid_request"}`)
	}
	return b
}
