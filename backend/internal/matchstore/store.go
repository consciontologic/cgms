// Package matchstore persists private authoritative match aggregates. It is an
// internal application boundary, not a transport or authentication provider.
package matchstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
)

var (
	ErrUnauthorized   = errors.New("match membership required")
	ErrConflict       = errors.New("command identity conflict")
	ErrStale          = errors.New("stale game or version")
	ErrInvalid        = errors.New("invalid intent")
	ErrCompatibility  = errors.New("unsupported persisted match")
	ErrOutcomeUnknown = telemetry.UnknownCommit
)

// Intent is private application input. Authentication maps an external identity
// to actorID before calling Submit. Physical engine card IDs are not a wire API.
type Intent struct {
	Transport       json.RawMessage `json:"transport,omitempty"`
	ID              string          `json:"id"`
	GameID          string          `json:"game_id"`
	ExpectedVersion int64           `json:"expected_version"`
	Request         Request         `json:"request"`
}
type ServerIntent struct {
	ID              string        `json:"id"`
	GameID          string        `json:"game_id"`
	ExpectedVersion int64         `json:"expected_version"`
	Request         ServerRequest `json:"request"`
}
type Result struct {
	Version    int64           `json:"version"`
	GameID     string          `json:"game_id"`
	Status     string          `json:"status"`
	Projection json.RawMessage `json:"projection,omitempty"`
}
type View struct {
	Version    int64           `json:"version"`
	Cursor     int64           `json:"cursor"`
	Projection json.RawMessage `json:"projection"`
	// Snapshot-only advisory status. Entitlements can change without a game event.
	AdmissionWaiting bool `json:"admission_waiting,omitempty"`
}
type Store struct {
	pool          *pgxpool.Pool
	rulesHash     string
	fault         func(string) error
	economyPolicy *economy.Policy
	now           func() time.Time
}

func New(pool *pgxpool.Pool, rulesHash string) *Store {
	return &Store{pool: pool, rulesHash: rulesHash, now: time.Now}
}
func (s *Store) fail(stage string) error {
	if s.fault != nil {
		return s.fault(stage)
	}
	return nil
}
func rollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
func encode(v any) ([]byte, error) {
	b, e := json.Marshal(v)
	if e == nil && len(b) > 64<<20 {
		return nil, errors.New("snapshot size limit")
	}
	return b, e
}
func (s *Store) decode(b []byte) (Envelope, error) {
	var e Envelope
	if len(b) > 64<<20 || json.Unmarshal(b, &e) != nil {
		return e, ErrCompatibility
	}
	if e.RulesHash != s.rulesHash || e.Validate() != nil {
		return e, ErrCompatibility
	}
	return e, nil
}
func bodyHash(v any) ([]byte, error) {
	if in, ok := v.(Intent); ok && len(in.Transport) > 0 {
		in.Request = Request{}
		v = in
	}
	b, e := canonical.Marshal(v)
	if e != nil {
		return nil, ErrInvalid
	}
	h := sha256.Sum256(b)
	return h[:], nil
}
func scopedID(actor, id string) string {
	b, _ := json.Marshal([]string{actor, id})
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func (s *Store) Create(ctx context.Context, id string, e Envelope, actors []string) error {
	if id == "" || len(id) > 256 || len(actors) != len(e.Match.Game.Board.Players) || e.RulesHash != s.rulesHash || e.Validate() != nil {
		return ErrInvalid
	}
	seen := map[string]bool{}
	for _, a := range actors {
		if a == "" || len(a) > 256 || seen[a] {
			return ErrInvalid
		}
		seen[a] = true
	}
	if _, err := encode(e); err != nil {
		return err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if err = s.CreateTx(ctx, tx, id, e, actors); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return ErrOutcomeUnknown
	}
	return nil
}

// CreateTx participates in a caller-owned room/start transaction.
func (s *Store) CreateTx(ctx context.Context, tx pgx.Tx, id string, e Envelope, actors []string) error {
	if id == "" || len(id) > 256 || len(actors) != len(e.Match.Game.Board.Players) || e.RulesHash != s.rulesHash || e.Validate() != nil {
		return ErrInvalid
	}
	if s.economyPolicy != nil {
		e.Schema = EconomyEnvelopeSchema
	} else if e.Schema == EconomyEnvelopeSchema {
		return ErrInvalid
	}
	seen := map[string]bool{}
	for _, actor := range actors {
		if actor == "" || len(actor) > 256 || seen[actor] {
			return ErrInvalid
		}
		seen[actor] = true
	}
	if e.BotDifficulty != "" {
		for i, actor := range actors {
			var bot bool
			if err := tx.QueryRow(ctx, "SELECT kind='bot' FROM identity_accounts WHERE account_id=$1", actor).Scan(&bot); err != nil || bot != (i > 0) {
				return ErrInvalid
			}
		}
	}
	b, err := encode(e)
	if err != nil {
		return err
	}
	config, err := encode(persistedConfig{GameLimit: e.Match.GameLimit, RulesHash: e.RulesHash, Initial: e, Economy: s.economyPolicy})
	if err != nil {
		return err
	}
	metadata, err := encode([]string{e.Schema, e.EngineVersion})
	if err != nil {
		return err
	}
	random, err := encode(e.Chance)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, "INSERT INTO matches(match_id,snapshot,state_version,current_game_id,config,random_state,metadata) VALUES($1,$2,0,$3,$4,$5,$6)", id, b, e.Match.Game.Board.GameID, config, random, metadata)
	if err != nil {
		return err
	}
	for i, a := range actors {
		if _, err = tx.Exec(ctx, "INSERT INTO members(match_id,actor_id,seat) VALUES($1,$2,$3)", id, a, i+1); err != nil {
			return err
		}
	}
	if s.economyPolicy != nil {
		if err = economy.ReserveStartTx(ctx, tx, e.Match.Game.Board.GameID, actors, s.now()); err != nil {
			return err
		}
	}
	if err = syncLedger(ctx, tx, id, e, 0); err != nil {
		return err
	}
	if err = s.writeViews(ctx, tx, id, e, 0, false); err != nil {
		return err
	}

	return nil
}

type storedMatch struct {
	body    []byte
	config  []byte
	version int64
	gameID  string
}

func lockMatch(ctx context.Context, tx pgx.Tx, id string) (result storedMatch, resultErr error) {
	ctx, finish := telemetry.Start(ctx, telemetry.Lock)
	defer func() { finish(resultErr) }()
	var row storedMatch
	err := tx.QueryRow(ctx, "SELECT snapshot,state_version,current_game_id,jsonb_build_object('economy',convert_from(config,'UTF8')::jsonb->'economy') FROM matches WHERE match_id=$1 FOR UPDATE", id).Scan(&row.body, &row.version, &row.gameID, &row.config)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrUnauthorized
	}
	return row, err
}
func (s *Store) decodeRow(row storedMatch) (Envelope, error) {
	e, err := s.decode(row.body)
	if err == nil && e.Match.Game.Board.GameID != row.gameID {
		err = ErrCompatibility
	}
	if err == nil {
		var cfg struct {
			Economy *economy.Policy `json:"economy"`
		}
		if json.Unmarshal(row.config, &cfg) != nil || (e.Schema == EconomyEnvelopeSchema) != (cfg.Economy != nil) || (cfg.Economy != nil && cfg.Economy.Validate() != nil) {
			err = ErrCompatibility
		}
	}
	return e, err
}
func (s *Store) locked(ctx context.Context, tx pgx.Tx, id string) (Envelope, int64, error) {
	row, err := lockMatch(ctx, tx, id)
	if err != nil {
		return Envelope{}, 0, err
	}
	e, err := s.decodeRow(row)
	return e, row.version, err
}
func member(ctx context.Context, tx pgx.Tx, id, actor string) (int, error) {
	var seat int
	err := tx.QueryRow(ctx, "SELECT seat FROM members WHERE match_id=$1 AND actor_id=$2", id, actor).Scan(&seat)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrUnauthorized
	}
	return seat, err
}
func receipt(ctx context.Context, tx pgx.Tx, id, actor, command string, hash []byte, server bool) (Result, bool, error) {
	var h, b []byte
	var err error
	if server {
		err = tx.QueryRow(ctx, "SELECT body_hash,result FROM server_results WHERE match_id=$1 AND command_id=$2", id, command).Scan(&h, &b)
	} else {
		err = tx.QueryRow(ctx, "SELECT body_hash,result FROM command_results WHERE match_id=$1 AND actor_id=$2 AND command_id=$3", id, actor, command).Scan(&h, &b)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return Result{}, false, nil
	}
	if err != nil {
		return Result{}, false, err
	}
	if !bytes.Equal(h, hash) {
		return Result{}, true, ErrConflict
	}
	var r Result
	if json.Unmarshal(b, &r) != nil {
		return r, true, ErrCompatibility
	}
	return r, true, nil
}
func (s *Store) Submit(ctx context.Context, id, actor string, in Intent) (Result, error) {
	return s.SubmitChecked(ctx, id, actor, in, nil)
}

// SubmitChecked authorizes transport capabilities under the same lock, after retries.
func (s *Store) SubmitChecked(ctx context.Context, id, actor string, in Intent, check func(Envelope, int, Intent) (Intent, error)) (result Result, resultErr error) {
	ctx, finish := telemetry.Start(ctx, telemetry.Transaction)
	defer func() { finish(resultErr) }()
	if (len(in.Transport) > 0 && check == nil) || in.ID == "" || len(in.ID) > 256 || in.GameID == "" || in.ExpectedVersion < 0 {
		return Result{}, ErrInvalid
	}
	hash, err := bodyHash(in)
	if err != nil {
		return Result{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer rollback(tx)
	// Lock precedes every state-dependent read, including duplicate resolution.
	row, err := lockMatch(ctx, tx, id)
	if err != nil {
		return Result{}, err
	}
	seat, err := member(ctx, tx, id, actor)
	if err != nil {
		return Result{}, err
	}
	if r, found, err := receipt(ctx, tx, id, actor, in.ID, hash, false); found || err != nil {
		return r, err
	}
	env, err := s.decodeRow(row)
	if err != nil {
		return Result{}, err
	}
	if check != nil {
		in, err = check(env, seat, in)
		if err != nil {
			return Result{}, ErrInvalid
		}
		verified, e := bodyHash(in)
		if e != nil || !bytes.Equal(verified, hash) {
			return Result{}, ErrInvalid
		}
	}
	version := row.version
	coup := (in.Request.Command != nil && in.Request.Command.Kind == "coup") || (in.Request.Operation != nil && (in.Request.Operation.Kind == "coup" || in.Request.Operation.Kind == "declare-ordinary"))
	if in.GameID != env.Match.Game.Board.GameID || (!coup && in.ExpectedVersion != version) {
		return Result{}, ErrStale
	}
	req := in.Request
	if req.Command != nil {
		c := *req.Command
		if c.GameID != in.GameID || c.ID != in.ID {
			return Result{}, ErrInvalid
		}
		c.ID = scopedID("actor:"+actor, in.ID)
		req.Command = &c
	}
	if req.Operation != nil {
		o := *req.Operation
		if o.GameID != in.GameID || o.ID != in.ID {
			return Result{}, ErrInvalid
		}
		o.ID = scopedID("actor:"+actor, in.ID)
		req.Operation = &o
	}
	_, engineDone := telemetry.Start(ctx, telemetry.Engine)
	next, err := applyRequest(env, seat, req)
	engineDone(err)
	if err != nil {
		return Result{}, ErrInvalid
	}
	return s.persist(ctx, tx, id, actor, in.ID, hash, next, version+1, seat, false, env.Match.Game.Board.GameID, in)
}

// Continue is a trusted scheduler capability. Never expose it as member input.
func (s *Store) Continue(ctx context.Context, id string, in ServerIntent) (result Result, resultErr error) {
	ctx, finish := telemetry.Start(ctx, telemetry.Transaction)
	defer func() { finish(resultErr) }()
	if in.ID == "" || len(in.ID) > 256 || in.GameID == "" || in.ExpectedVersion < 0 {
		return Result{}, ErrInvalid
	}
	hash, err := bodyHash(in)
	if err != nil {
		return Result{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer rollback(tx)
	row, err := lockMatch(ctx, tx, id)
	if err != nil {
		return Result{}, err
	}
	if r, found, err := receipt(ctx, tx, id, "", in.ID, hash, true); found || err != nil {
		return r, err
	}
	env, err := s.decodeRow(row)
	if err != nil {
		return Result{}, err
	}
	v := row.version
	if in.GameID != env.Match.Game.Board.GameID || in.ExpectedVersion != v || in.Request.GameID != in.GameID {
		return Result{}, ErrStale
	}
	req := in.Request
	req.ID = scopedID("scheduler:", in.ID)
	if req.Operation != nil {
		o := *req.Operation
		o.ID = scopedID("scheduler:", in.ID)
		req.Operation = &o
	}
	_, engineDone := telemetry.Start(ctx, telemetry.Engine)
	next, err := applyServerRequest(env, req)
	engineDone(err)
	if err != nil {
		return Result{}, ErrInvalid
	}
	return s.persist(ctx, tx, id, "", in.ID, hash, next, v+1, 0, true, env.Match.Game.Board.GameID, in)
}
func (s *Store) persist(ctx context.Context, tx pgx.Tx, id, actor, command string, hash []byte, e Envelope, v int64, seat int, server bool, previousGame string, input any) (Result, error) {
	if err := s.syncEconomy(ctx, tx, id, previousGame, e); err != nil {
		return Result{}, err
	}
	b, err := encode(e)
	if err != nil {
		return Result{}, err
	}
	random, err := encode(e.Chance)
	if err != nil {
		return Result{}, err
	}
	if _, err = tx.Exec(ctx, "UPDATE matches SET snapshot=$2,state_version=$3,current_game_id=$4,random_state=$5 WHERE match_id=$1", id, b, v, e.Match.Game.Board.GameID, random); err != nil {
		return Result{}, err
	}
	if err = s.fail("snapshot"); err != nil {
		return Result{}, err
	}
	if err = syncReputation(ctx, tx, id, e); err != nil {
		return Result{}, err
	}
	if err = syncLedger(ctx, tx, id, e, v); err != nil {
		return Result{}, err
	}
	if err = s.fail("ledger"); err != nil {
		return Result{}, err
	}
	// Journal is private forensic state, never returned to a member.
	inputBytes, err := encode(input)
	if err != nil {
		return Result{}, err
	}
	journalBytes, err := encode(JournalEntry{ActorID: actor, CommandID: command, Input: inputBytes, Snapshot: e, Version: v, Server: server})
	if err != nil {
		return Result{}, err
	}
	if _, err = tx.Exec(ctx, "INSERT INTO match_events(match_id,seq,game_id,event) VALUES($1,$2,$3,$4)", id, v, e.Match.Game.Board.GameID, journalBytes); err != nil {
		return Result{}, err
	}
	if err = s.writeViews(ctx, tx, id, e, v, true); err != nil {
		return Result{}, err
	}
	if err = s.fail("events"); err != nil {
		return Result{}, err
	}
	r := Result{Version: v, GameID: e.Match.Game.Board.GameID, Status: "committed"}
	if e.Match.Game.Automatic != nil || e.Match.Game.PromisePayment != nil {
		r.Status = "committed-pending-settlement"
	}
	if !server {
		r.Projection, err = project(e, seat)
		if err != nil {
			return Result{}, err
		}
	}
	rb, err := encode(r)
	if err != nil {
		return Result{}, err
	}
	if server {
		_, err = tx.Exec(ctx, "INSERT INTO server_results(match_id,command_id,body_hash,result) VALUES($1,$2,$3,$4)", id, command, hash, rb)
	} else {
		_, err = tx.Exec(ctx, "INSERT INTO command_results(match_id,actor_id,command_id,body_hash,result) VALUES($1,$2,$3,$4,$5)", id, actor, command, hash, rb)
	}
	if err != nil {
		return Result{}, err
	}
	if err = s.fail("receipt"); err != nil {
		return Result{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Result{}, ErrOutcomeUnknown
	}
	if err = s.fail("acknowledgement"); err != nil {
		return Result{}, ErrOutcomeUnknown
	}
	return r, nil
}
func project(e Envelope, seat int) ([]byte, error) {
	o, err := SeatProjection(e, seat)
	if err != nil {
		return nil, err
	}
	return encode(o)
}
func (s *Store) writeViews(ctx context.Context, tx pgx.Tx, id string, e Envelope, v int64, event bool) error {
	for seat := 1; seat <= len(e.Match.Game.Board.Players); seat++ {
		b, err := project(e, seat)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO seat_projections(match_id,seat,state_version,cursor,projection) VALUES($1,$2,$3,$3,$4) ON CONFLICT(match_id,seat) DO UPDATE SET state_version=excluded.state_version,cursor=excluded.cursor,projection=excluded.projection", id, seat, v, b); err != nil {
			return err
		}
		if event {
			if _, err = tx.Exec(ctx, "INSERT INTO seat_events(match_id,seat,seq,projection) VALUES($1,$2,$3,$4)", id, seat, v, b); err != nil {
				return err
			}
		}
	}
	return nil
}
func (s *Store) Snapshot(ctx context.Context, id, actor string) (View, error) {
	var v View
	var checkpoint []byte
	var seat int
	var head int64
	// The cached view and private checkpoint are read at one PostgreSQL statement
	// snapshot. An upgraded server can refresh additive presentation context from
	// that exact state without a command, event, cache write or version advance.
	err := s.pool.QueryRow(ctx, "SELECT p.state_version,p.cursor,p.projection,m.seat,x.state_version,x.snapshot FROM seat_projections p JOIN members m USING(match_id,seat) JOIN matches x ON x.match_id=p.match_id WHERE m.match_id=$1 AND m.actor_id=$2", id, actor).Scan(&v.Version, &v.Cursor, &v.Projection, &seat, &head, &checkpoint)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrUnauthorized
	}
	if err == nil {
		if head != v.Version {
			return View{}, ErrCompatibility
		}
		var fields map[string]json.RawMessage
		var online map[string]json.RawMessage
		if json.Unmarshal(v.Projection, &fields) != nil || json.Unmarshal(fields["online"], &online) != nil || online == nil {
			return View{}, ErrCompatibility
		}
		_, hasRound := online["round_closed"]
		_, hasDeparture := online["departure_pending"]
		_, hasServerPending := online["server_pending"]
		var projection Projection
		if json.Unmarshal(v.Projection, &projection) != nil {
			return View{}, ErrCompatibility
		}
		rules := projection.Online.RulesContext
		missingPending := projection.Board.WindowID != "" && rules.Pending == nil
		missingOperands := rules.Combat != nil && (rules.Combat.AttackCards == nil || rules.Combat.TargetCards == nil)
		missingOutcome := rules.Pending != nil && rules.Pending.Type == "kidnapper" && projection.Board.RequiredActor == 0
		if !hasRound || !hasDeparture || !hasServerPending || missingPending || missingOperands || missingOutcome {
			env, decodeErr := s.decode(checkpoint)
			if decodeErr != nil || seat < 1 || seat > len(env.Match.Game.Board.Players) {
				return View{}, ErrCompatibility
			}
			projection, err = SeatProjection(env, seat)
			if err != nil {
				return View{}, err
			}
			v.Projection, err = encode(projection)
			if err != nil {
				return View{}, err
			}
		}
		status := projection.Online
		if (status.FinancialPhase == "finalized" || status.FinancialPhase == "void") && status.CompletedGames < status.GamesPerMatch {
			v.AdmissionWaiting, err = s.admissionWaiting(ctx, id, actor, projection.Board.GameID)
		}
	}
	return v, err
}

// Reconcile queries the primary after waiting for the same match lock. Absence
// remains unknown: callers must retry the identical intent, never mint a new ID.
func (s *Store) Reconcile(ctx context.Context, id, actor string, in Intent) (Result, error) {
	hash, err := bodyHash(in)
	if err != nil {
		return Result{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Result{}, err
	}
	defer rollback(tx)
	if _, err = lockMatch(ctx, tx, id); err != nil {
		return Result{}, err
	}
	if _, err = member(ctx, tx, id, actor); err != nil {
		return Result{}, err
	}
	r, found, err := receipt(ctx, tx, id, actor, in.ID, hash, false)
	if err != nil {
		return r, err
	}
	if !found {
		return r, ErrOutcomeUnknown
	}
	return r, nil
}

// Restore is private administrative recovery, never an actor-facing endpoint.
func (s *Store) Restore(ctx context.Context, id string) (Envelope, int64, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Envelope{}, 0, err
	}
	defer rollback(tx)
	return s.locked(ctx, tx, id)
}
func (s *Store) String() string { return fmt.Sprintf("matchstore(%s)", s.rulesHash) }

// JournalEntry contains private replay inputs and committed chance outcomes.
type JournalEntry struct {
	ActorID   string          `json:"actor_id"`
	CommandID string          `json:"command_id"`
	Input     json.RawMessage `json:"input"`
	Snapshot  Envelope        `json:"snapshot"`
	Version   int64           `json:"version"`
	Server    bool            `json:"server"`
}
type persistedConfig struct {
	GameLimit int             `json:"game_limit"`
	RulesHash string          `json:"rules_hash"`
	Initial   Envelope        `json:"initial"`
	Economy   *economy.Policy `json:"economy,omitempty"`
}
