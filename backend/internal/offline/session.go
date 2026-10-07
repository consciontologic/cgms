// Package offline owns network-independent local sessions. Private snapshots are
// storage capabilities, never presentation DTOs or authoritative online results.
package offline

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
)

const Schema = "cgms-offline-session-v1"
const EngineVersion = "cgms-offline-engine-v1"
const MaxSaveBytes = 4 << 20

type Session struct {
	Schema        string              `json:"schema"`
	EngineVersion string              `json:"engine_version"`
	Version       int64               `json:"version"`
	HumanSeat     int                 `json:"human_seat"`
	Difficulty    string              `json:"difficulty"`
	Tutorial      string              `json:"tutorial"`
	TutorialStep  int                 `json:"tutorial_step"`
	Engine        matchstore.Envelope `json:"engine"`
}

func newID() (string, error) {
	var b [16]byte
	_, err := rand.Read(b[:])
	return "offline-" + hex.EncodeToString(b[:]), err
}
func NewSession(population, games, seat int, difficulty, rules string) (*Session, error) {
	if games < 1 || games > 3 || seat < 1 || seat > population || rules == "" {
		return nil, errors.New("invalid local configuration")
	}
	id, err := newID()
	if err != nil {
		return nil, err
	}
	board, err := game.NewState(population, id)
	if err != nil {
		return nil, err
	}
	match, err := game.NewMatchLifecycle(board, games)
	if err != nil {
		return nil, err
	}
	env, err := matchstore.NewEnvelope(match, rules)
	if err != nil {
		return nil, err
	}
	s := &Session{Schema: Schema, EngineVersion: EngineVersion, HumanSeat: seat, Difficulty: difficulty, Engine: env}
	return s, s.Validate(rules)
}
func (s *Session) Validate(rules string) error {
	if s == nil || s.Schema != Schema || s.EngineVersion != EngineVersion || s.Version < 0 || s.HumanSeat < 1 || s.HumanSeat > len(s.Engine.Match.Game.Board.Players) || s.Engine.Schema != matchstore.EnvelopeSchema || s.Engine.RulesHash != rules || rules == "" {
		return errors.New("incompatible local session")
	}
	switch s.Difficulty {
	case "beginner", "standard", "advanced":
	default:
		return errors.New("unknown difficulty")
	}
	if s.Tutorial != "" && s.Tutorial != "first-turn-v1" {
		return errors.New("unknown tutorial")
	}
	if s.TutorialStep < 0 || s.TutorialStep > 3 {
		return errors.New("invalid tutorial progress")
	}
	return s.Engine.Validate()
}
func (s *Session) clone() *Session { n := *s; n.Engine = s.Engine.Clone(); return &n }
