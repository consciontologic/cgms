package matchstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/metaphy6/cgms/backend/internal/game"
)

var ErrCursor = errors.New("snapshot required")

type StreamEvent struct {
	Cursor     int64           `json:"cursor"`
	Projection json.RawMessage `json:"projection"`
}

func (s *Store) Events(ctx context.Context, id, actor string, after int64, limit int) ([]StreamEvent, error) {
	if after < 0 || limit < 1 || limit > 128 {
		return nil, ErrCursor
	}
	tx, e := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if e != nil {
		return nil, e
	}
	defer rollback(tx)
	seat, e := member(ctx, tx, id, actor)
	if e != nil {
		return nil, e
	}
	var head int64
	if e = tx.QueryRow(ctx, "SELECT state_version FROM matches WHERE match_id=$1", id).Scan(&head); e != nil {
		return nil, e
	}
	if after > head || after < head-1000 {
		return nil, ErrCursor
	}
	rows, e := tx.Query(ctx, "SELECT seq,projection FROM seat_events WHERE match_id=$1 AND seat=$2 AND seq>$3 ORDER BY seq LIMIT $4", id, seat, after, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []StreamEvent{}
	size := 0
	for rows.Next() {
		var v StreamEvent
		if e = rows.Scan(&v.Cursor, &v.Projection); e != nil {
			return nil, e
		}
		if v.Cursor != after+int64(len(out))+1 {
			return nil, ErrCursor
		}
		size += len(v.Projection)
		if size > 1<<20 {
			return nil, ErrCursor
		}
		out = append(out, v)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	if len(out) == 0 && after < head {
		return nil, ErrCursor
	}
	return out, nil
}
func (s *Store) Lookup(ctx context.Context, id, actor, command string) (Result, error) {
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return Result{}, e
	}
	defer rollback(tx)
	if _, e = lockMatch(ctx, tx, id); e != nil {
		return Result{}, e
	}
	if _, e = member(ctx, tx, id, actor); e != nil {
		return Result{}, e
	}
	var b []byte
	e = tx.QueryRow(ctx, "SELECT result FROM command_results WHERE match_id=$1 AND actor_id=$2 AND command_id=$3", id, actor, command).Scan(&b)
	if errors.Is(e, pgx.ErrNoRows) {
		return Result{}, ErrOutcomeUnknown
	}
	var r Result
	if e == nil {
		e = json.Unmarshal(b, &r)
	}
	return r, e
}
func (s *Store) Priority(ctx context.Context, id, actor string, victory bool) (int, error) {
	if victory {
		return -1, nil
	}
	e, _, err := s.Restore(ctx, id)
	if err != nil {
		return 0, err
	}
	var seat int
	err = s.pool.QueryRow(ctx, "SELECT seat FROM members WHERE match_id=$1 AND actor_id=$2", id, actor).Scan(&seat)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrUnauthorized
	}
	if err != nil {
		return 0, err
	}
	if seat == e.Match.Game.Board.Active {
		return 0, nil
	}
	for i, x := range e.Match.Game.Board.Order {
		if x == seat {
			return i + 1, nil
		}
	}
	return seat + 10, nil
}

// NextServer plans one bounded automatic transition, never a human response.
func NextServer(e Envelope, version int64) (ServerIntent, bool) {
	m := e.Match
	l := m.Game
	b := l.Board
	r := ServerRequest{GameID: b.GameID}
	op := func(kind string) {
		r.Kind = "operation"
		r.Operation = &game.LifecycleOperation{GameID: b.GameID, Actor: b.Active, Kind: kind, Budget: 128}
		if kind != "resume-automatic" {
			r.Operation.Budget = 0
		}
	}
	switch {
	case l.Ledger.Phase == "finalized" || l.Ledger.Phase == "void":
		if len(l.Ledger.Completed) >= m.GameLimit {
			return ServerIntent{}, false
		}
		r.Kind = "next-game"
		r.NextGameID = scopedID("next-game:"+b.GameID, fmt.Sprint(version))
	case l.Automatic != nil:
		op("resume-automatic")
	case l.Ending != "":
		for _, f := range l.Ledger.Finished {
			if !f {
				return ServerIntent{}, false
			}
		}
		op("finalize")
	case len(b.Players[0].History) == 0 && len(b.Order) == 0:
		r.Kind = "deal"
	case len(b.Order) == 0:
		r.Kind = "initiative"
	case b.Pending != nil:
		p := b.Pending
		if p.Decision != nil || p.Cursor < len(p.Responders) {
			return ServerIntent{}, false
		}
		if p.Kind == "ability" && p.Ability != nil && p.Ability.Kind == "kidnapper" && len(game.KidnapperClosedCandidates(b, p.Ability.TargetSeat)) == 0 && len(game.KidnapperOpenCandidates(b, p.Ability.TargetSeat)) > 0 {
			return ServerIntent{}, false
		}
		r.Kind = "pending-chance"
	default:
		post := b.NeedsShuffle || len(b.DrawQueue) > 0 || l.EndTurnPending || len(l.PendingReceipts) > 0
		for _, x := range b.Effects {
			post = post || (x.Kind == "compensation" && x.Losses >= 5)
		}
		if post {
			r.Kind = "after-action"
		} else if l.RoundClosed {
			if !l.DeparturesResolved {
				return ServerIntent{}, false
			}
			op("begin-round")
		} else if !l.TurnStarted {
			op("begin-turn")
		} else {
			return ServerIntent{}, false
		}
	}
	return ServerIntent{ID: fmt.Sprintf("automatic/%d", version), GameID: b.GameID, ExpectedVersion: version, Request: r}, true
}
