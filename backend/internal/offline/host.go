package offline

import (
	"errors"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"io/fs"
	"sync"
)

type Request struct {
	Kind       string              `json:"kind"`
	Slot       string              `json:"slot"`
	Version    int64               `json:"version"`
	Population int                 `json:"population,omitempty"`
	Games      int                 `json:"games,omitempty"`
	Difficulty string              `json:"difficulty,omitempty"`
	Action     *matchstore.Request `json:"action,omitempty"`
}

// Host has no sockets, service clients, accounts, dirt or authority credentials.
// Every mutation commits before returning a projection; reloading is mandatory
// for each request, including after any uncertain storage failure.
type Host struct {
	store *Store
	mu    sync.Mutex
}

func NewHost(store *Store) *Host { return &Host{store: store} }
func (h *Host) Handle(r Request) (View, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if !slotPattern.MatchString(r.Slot) {
		return View{}, errors.New("invalid slot")
	}
	if r.Kind == "new" || r.Kind == "tutorial" {
		if _, err := h.store.Load(r.Slot); !errors.Is(err, fs.ErrNotExist) {
			return View{}, errors.New("save slot already exists or is unreadable")
		}
		s, err := NewSession(r.Population, r.Games, 1, r.Difficulty, h.store.rules)
		if err != nil {
			return View{}, err
		}
		if r.Kind == "tutorial" {
			s.Tutorial = "first-turn-v1"
			for i := 0; i < 100; i++ {
				next, advanced, e := s.Advance()
				if e != nil {
					return View{}, e
				}
				s = next
				if !advanced {
					break
				}
			}
			if !s.Engine.Match.Game.TurnStarted {
				return View{}, errors.New("tutorial setup budget")
			}
			s.HumanSeat = s.Engine.Match.Game.Board.Active
		}
		view, err := s.View()
		if err != nil {
			return View{}, err
		}
		if err = h.store.Save(r.Slot, s); err != nil {
			return View{}, err
		}
		return view, nil
	}
	s, err := h.store.Load(r.Slot)
	if err != nil {
		return View{}, err
	}
	if r.Kind == "load" {
		return s.View()
	}
	if r.Version != s.Version {
		return View{}, errors.New("stale local version; reload")
	}
	var n *Session
	switch r.Kind {
	case "act":
		if r.Action == nil {
			return View{}, errors.New("missing human action")
		}
		n, err = s.Apply(s.HumanSeat, *r.Action)
		if err == nil && s.Tutorial == "first-turn-v1" {
			if s.TutorialStep == 0 && r.Action.Command != nil && r.Action.Command.Kind == "open-series" {
				n.TutorialStep = 1
			}
			if s.TutorialStep == 2 && r.Action.Operation != nil && r.Action.Operation.Kind == "end-turn" {
				n.TutorialStep = 3
			}
		}
	case "step":
		var advanced bool
		n, advanced, err = s.Advance()
		if err == nil && !advanced {
			found := false
			for seat := 1; seat <= len(s.Engine.Match.Game.Board.Players); seat++ {
				if seat == s.HumanSeat {
					continue
				}
				var choice matchstore.Request
				actions, e := s.Actions(seat)
				if e != nil {
					return View{}, e
				}
				if len(actions) == 0 {
					continue
				}
				if s.Tutorial == "first-turn-v1" && s.TutorialStep == 1 {
					for _, a := range actions {
						if a.Command != nil && a.Command.Kind == "pass" {
							choice = a
							break
						}
					}
				}
				if choice.Command == nil && choice.Operation == nil {
					choice, e = s.Suggest(seat)
					if errors.Is(e, ErrNoDecision) {
						continue
					}
					if e != nil {
						return View{}, e
					}
				}
				n, err = s.Apply(seat, choice)
				found = true
				break
			}
			if !found {
				return s.View()
			}
		}
		if err == nil && s.Tutorial == "first-turn-v1" && s.TutorialStep == 1 && n.Engine.Match.Game.Board.Pending == nil {
			n.TutorialStep = 2
		}
	default:
		return View{}, errors.New("unknown local request")
	}
	if err != nil {
		return View{}, err
	}
	view, err := n.View()
	if err != nil {
		return View{}, err
	}
	if err = h.store.Save(r.Slot, n); err != nil {
		return View{}, err
	}
	return view, nil
}
