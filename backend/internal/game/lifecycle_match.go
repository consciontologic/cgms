package game

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"maps"
	"slices"
)

// MatchLifecycle keeps fixed match length and every game-instance identity,
// including void instances, separate from the count of finalized results.
// It is private provenance, never a bot observation.
type MatchLifecycle struct {
	NullificationConsents []int                 `json:"nullification_consents,omitempty"`
	CommandHistory        []LifecycleReceiptSet `json:"command_history"`
	GameLimit             int                   `json:"game_limit"`
	Game                  Lifecycle             `json:"game"`
	Instances             []string              `json:"instances"`
	StartLedger           FinancialLedger       `json:"start_ledger"`
}
type LifecycleReceiptSet struct {
	GameID            string            `json:"game_id"`
	CardCommands      map[string]string `json:"card_commands"`
	LifecycleCommands map[string]string `json:"lifecycle_commands"`
}

// Clone detaches every mutable match-owned history and baseline, not only the
// current game. Those snapshots remain authoritative for retry and nullification.
func (m MatchLifecycle) Clone() MatchLifecycle {
	n := m
	n.Game = m.Game.Clone()
	n.StartLedger = m.StartLedger.Clone()
	n.Instances = slices.Clone(m.Instances)
	n.NullificationConsents = slices.Clone(m.NullificationConsents)
	n.CommandHistory = make([]LifecycleReceiptSet, len(m.CommandHistory))
	for i, r := range m.CommandHistory {
		n.CommandHistory[i] = LifecycleReceiptSet{GameID: r.GameID, CardCommands: maps.Clone(r.CardCommands), LifecycleCommands: maps.Clone(r.LifecycleCommands)}
	}
	return n
}

func NewMatchLifecycle(board State, limit int) (MatchLifecycle, error) {
	if limit < 1 {
		return MatchLifecycle{}, errors.New("positive fixed match length required")
	}
	ledger := NewFinancialLedger(board.GameID, len(board.Players))
	g, e := NewLifecycle(board, ledger)
	if e != nil {
		return MatchLifecycle{}, e
	}
	return MatchLifecycle{GameLimit: limit, Game: g, Instances: []string{board.GameID}, StartLedger: ledger.Clone()}, nil
}
func (m MatchLifecycle) NextGame(id string) (MatchLifecycle, error) {
	if m.Game.Validate() != nil || m.Game.Automatic != nil || (m.Game.Promises.Phase != "finalized" && m.Game.Promises.Phase != "void") || m.GameLimit < 1 || len(m.Game.Ledger.Completed) >= m.GameLimit || slices.Contains(m.Instances, id) {
		return m, errors.New("match cannot admit another game")
	}
	ledger, e := m.Game.Ledger.NextGame(id)
	if e != nil {
		return m, e
	}
	board, e := NewState(len(m.Game.Board.Players), id)
	if e != nil {
		return m, e
	}
	g, e := NewLifecycle(board, ledger)
	if e != nil {
		return m, e
	}
	n := m.Clone()
	n.CommandHistory = make([]LifecycleReceiptSet, len(m.CommandHistory))
	for i, r := range m.CommandHistory {
		n.CommandHistory[i] = LifecycleReceiptSet{GameID: r.GameID, CardCommands: maps.Clone(r.CardCommands), LifecycleCommands: maps.Clone(r.LifecycleCommands)}
	}
	n.CommandHistory = append(n.CommandHistory, LifecycleReceiptSet{GameID: m.Game.Board.GameID, CardCommands: m.Game.Board.Clone().Commands, LifecycleCommands: m.Game.Clone().Commands})
	n.Game = g
	n.NullificationConsents = nil
	n.StartLedger = ledger.Clone()
	n.Instances = append(slices.Clone(m.Instances), id)
	return n, nil
}

// Ranks uses competition ranking: equal cumulative exact totals share a rank.
func (m MatchLifecycle) Ranks() []int {
	scores := m.Game.Ledger.MatchScores()
	if scores == nil {
		return nil
	}
	ranks := make([]int, len(scores))
	for i, a := range scores {
		ranks[i] = 1
		for _, b := range scores {
			if b.Cmp(a) > 0 {
				ranks[i]++
			}
		}
	}
	return ranks
}

// ApplyOperation places nullification in the same identity/retry domain as
// other commands, while retaining the match-owned financial start snapshot.
func (m MatchLifecycle) ApplyOperation(op LifecycleOperation) (MatchLifecycle, error) {
	if op.GameID != m.Game.Board.GameID {
		digest, err := canonical.Hash(op)
		if err != nil {
			return m, err
		}
		for _, receipt := range m.CommandHistory {
			if receipt.GameID == op.GameID {
				if old, ok := receipt.LifecycleCommands[op.ID]; ok && old == digest {
					n := m.Clone()
					n.Game = m.Game.Clone()
					return n, nil
				}
				return m, errors.New("historic command missing or conflicting")
			}
		}
		return m, errors.New("unknown game instance")
	}

	if op.Kind != "nullify" {
		n := m.Clone()
		g, e := m.Game.ApplyOperation(op)
		if e != nil {
			return m, e
		}
		n.Game = g
		return n, nil
	}
	if op.ID == "" || op.GameID != m.Game.Board.GameID || op.Actor < 1 || op.Actor > len(m.Game.Board.Players) {
		return m, errors.New("invalid match operation")
	}
	hash, e := canonical.Hash(op)
	if e != nil {
		return m, e
	}
	if old, ok := m.Game.Commands[op.ID]; ok {
		if old != hash {
			return m, errors.New("match command conflict")
		}
		n := m.Clone()
		n.Game = m.Game.Clone()
		return n, nil
	}
	if len(op.Seats) > 1 || len(op.Seats) == 1 && op.Seats[0] != op.Actor {
		return m, errors.New("cannot assert another player's consent")
	}
	if m.Game.Validate() != nil || m.Game.Automatic != nil || (m.Game.Ledger.Phase != "playing" && m.Game.Ledger.Phase != "settlement") {
		return m, errors.New("nullification consent boundary")
	}
	n := m.Clone()
	n.Game = m.Game.Clone()
	n.NullificationConsents = slices.Clone(m.NullificationConsents)
	seen := map[int]bool{}
	for _, seat := range n.NullificationConsents {
		if seat < 1 || seat > len(m.Game.Board.Players) || seen[seat] {
			return m, errors.New("invalid saved nullification consent")
		}
		seen[seat] = true
	}
	if !seen[op.Actor] {
		n.NullificationConsents = append(n.NullificationConsents, op.Actor)
	}
	if len(n.NullificationConsents) == len(m.Game.Board.Players) {
		n, e = n.Nullify(n.NullificationConsents)
		if e != nil {
			return m, e
		}
	}
	n.Game.Commands[op.ID] = hash
	return n, nil
}
