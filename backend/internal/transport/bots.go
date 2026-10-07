package transport

import (
	"context"
	"errors"
	"fmt"

	"github.com/metaphy6/cgms/backend/internal/botplay"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
)

// botIntent plans at most one bounded decision from a persisted roster and an
// authorized observation. No goroutine or transient game state lives here.
// Human turns and required responses remain pending until their owner acts,
// except for the rules' immediate, out-of-turn Coup victory.
func (s *Server) botIntent(ctx context.Context, match string, env matchstore.Envelope, version int64) (string, matchstore.Intent, bool, error) {
	if env.BotDifficulty == "" {
		return "", matchstore.Intent{}, false, nil
	}
	l := env.Match.Game
	if l.Ledger.Phase == "playing" && l.Ending == "" {
		for seat := 2; seat <= len(l.Board.Players); seat++ {
			if err := ctx.Err(); err != nil {
				return "", matchstore.Intent{}, false, err
			}
			request, found, err := botplay.Coup(env, seat)
			if err != nil {
				return "", matchstore.Intent{}, false, err
			}
			if found {
				return s.botActorIntent(ctx, match, l.Board.GameID, version, seat, request)
			}
		}
	}
	if l.RoundClosed && l.Ending == "" && !l.DeparturesResolved && !l.Board.Players[0].Departed {
		// Bot policies independently stay. Collect the human's explicit round
		// choice first so background votes cannot stale the visible prompt. The
		// engine still applies all departures simultaneously; Coup remains first.
		humanAnswered := false
		for _, choice := range l.DepartureChoices {
			if choice.Seat == 1 {
				humanAnswered = true
				break
			}
		}
		if !humanAnswered {
			return "", matchstore.Intent{}, false, nil
		}
	}
	seats := []int{}
	if l.Ledger.Phase == "playing" && !l.RoundClosed {
		seat := l.Board.Active
		if l.Board.Pending != nil {
			o, err := game.ObserveWithoutMenu(l.Board, 1)
			if err != nil {
				return "", matchstore.Intent{}, false, err
			}
			seat = o.RequiredActor
			if seat == 0 {
				seat = l.Board.Pending.Actor
			}
		}
		if seat > 1 {
			seats = append(seats, seat)
		}
	} else {
		for seat := 2; seat <= len(l.Board.Players); seat++ {
			seats = append(seats, seat)
		}
	}
	for _, seat := range seats {
		if err := ctx.Err(); err != nil {
			return "", matchstore.Intent{}, false, err
		}
		request, err := botplay.SuggestOnlineContext(ctx, env, version, seat, env.BotDifficulty)
		if errors.Is(err, botplay.ErrNoDecision) {
			continue
		}
		if err != nil {
			return "", matchstore.Intent{}, false, err
		}
		return s.botActorIntent(ctx, match, l.Board.GameID, version, seat, request)
	}
	return "", matchstore.Intent{}, false, nil
}

func (s *Server) botActorIntent(ctx context.Context, match, gameID string, version int64, seat int, request matchstore.Request) (string, matchstore.Intent, bool, error) {
	var actor string
	if err := s.cfg.Pool.QueryRow(ctx, `SELECT actor_id FROM members JOIN identity_accounts ON account_id=actor_id WHERE match_id=$1 AND seat=$2 AND kind='bot'`, match, seat).Scan(&actor); err != nil {
		return "", matchstore.Intent{}, false, err
	}
	id := fmt.Sprintf("bot/%d/%d", version, seat)
	if request.Command != nil {
		request.Command.ID = id
	}
	if request.Operation != nil {
		request.Operation.ID = id
	}
	return actor, matchstore.Intent{ID: id, GameID: gameID, ExpectedVersion: version, Request: request}, true, nil
}

// automaticStep executes inside the existing per-match dispatcher. Receipts and
// version/actor checks are identical to human commands; an uncertain commit is
// recovered from durable state on the next scan or after process restart.
func (s *Server) automaticStep(ctx context.Context, id string) (any, error) {
	env, version, err := s.matches.Restore(ctx, id)
	if err != nil {
		return nil, err
	}
	if next, ok := matchstore.NextServer(env, version); ok {
		return s.matches.Continue(ctx, id, next)
	}
	actor, intent, ok, err := s.botIntent(ctx, id, env, version)
	if err != nil || !ok {
		return nil, err
	}
	return s.matches.Submit(ctx, id, actor, intent)
}
