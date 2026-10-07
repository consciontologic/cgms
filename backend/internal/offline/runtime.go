package offline

import (
	"github.com/metaphy6/cgms/backend/internal/botplay"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
)

var ErrNoDecision = botplay.ErrNoDecision

// View is the sole presentation response. No private checkpoint/chance is here.
type View struct {
	Instruction  string                `json:"instruction,omitempty"`
	Version      int64                 `json:"version"`
	Difficulty   string                `json:"difficulty"`
	Projection   matchstore.Projection `json:"projection"`
	Actions      []matchstore.Request  `json:"actions"`
	Tutorial     string                `json:"tutorial,omitempty"`
	TutorialStep int                   `json:"tutorial_step"`
}

func (s *Session) View() (View, error) {
	if err := s.Validate(s.Engine.RulesHash); err != nil {
		return View{}, err
	}
	p, err := matchstore.SeatProjection(s.Engine, s.HumanSeat)
	if err != nil {
		return View{}, err
	}
	actions, err := s.Actions(s.HumanSeat)
	if err != nil {
		return View{}, err
	}
	instruction := ""
	if s.Tutorial == "first-turn-v1" {
		instruction = []string{"Open a series from your hand. Choose an available open-series action; opponents then respond in order.", "Other players may respond before your series opens. Advance the lesson to see each explicit pass.", "Your series is now on the table. End your turn to let the next player act.", "First turn complete. Your progress is saved; continue playing with the same legal actions."}[s.TutorialStep]
	}
	return View{Instruction: instruction, Version: s.Version, Difficulty: s.Difficulty, Projection: p, Actions: actions, Tutorial: s.Tutorial, TutorialStep: s.TutorialStep}, nil
}
func (s *Session) Advance() (*Session, bool, error) {
	if err := s.Validate(s.Engine.RulesHash); err != nil {
		return s, false, err
	}
	plan, ok := matchstore.NextServer(s.Engine, s.Version)
	if !ok {
		return s, false, nil
	}
	plan.Request.ID = plan.ID
	if plan.Request.Operation != nil {
		plan.Request.Operation.ID = plan.ID
	}
	next, err := matchstore.ApplyLocalAutomatic(s.Engine, plan.Request)
	if err != nil {
		return s, false, err
	}
	n := s.clone()
	n.Engine = next
	n.Version++
	return n, true, nil
}
func (s *Session) Apply(seat int, request matchstore.Request) (*Session, error) {
	if err := s.Validate(s.Engine.RulesHash); err != nil {
		return s, err
	}
	next, err := matchstore.ApplyLocalRequest(s.Engine, seat, request)
	if err != nil {
		return s, err
	}
	n := s.clone()
	n.Engine = next
	if game.Digest(next) != game.Digest(s.Engine) {
		n.Version++
	}
	return n, nil
}
func (s *Session) Actions(seat int) ([]matchstore.Request, error) {
	return botplay.Actions(s.Engine, s.Version, seat)
}

func (s *Session) Suggest(seat int) (matchstore.Request, error) {
	return botplay.Suggest(s.Engine, s.Version, seat, s.Difficulty)
}
