// Package rooms owns private lobbies. Match creation and room start share one
// PostgreSQL transaction; identity, room and match records use that lock order.
package rooms

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
)

var (
	ErrInvalid   = errors.New("invalid room input")
	ErrForbidden = errors.New("room access denied")
	ErrConflict  = errors.New("room command conflict")
	ErrFull      = errors.New("room is full")
	ErrClosed    = errors.New("room is already started")
	ErrNotReady  = errors.New("room requires all configured seats")
)

//go:embed migrations/001_rooms.sql
var schema string

//go:embed migrations/002_bots.sql
var botsMigration string

type Member struct {
	ActorID string `json:"actor_id"`
	Seat    int    `json:"seat"`
	Bot     bool   `json:"bot,omitempty"`
}
type Room struct {
	ID            string   `json:"id"`
	OwnerID       string   `json:"owner_id"`
	Capacity      int      `json:"capacity"`
	Games         int      `json:"games"`
	Status        string   `json:"status"`
	MatchID       string   `json:"match_id,omitempty"`
	Members       []Member `json:"members"`
	BotDifficulty string   `json:"bot_difficulty,omitempty"`
}
type Store struct {
	pool      *pgxpool.Pool
	matches   *matchstore.Store
	rulesHash string
}

func New(pool *pgxpool.Pool, matches *matchstore.Store, rulesHash string) *Store {
	return &Store{pool, matches, rulesHash}
}
func rollback(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
func Up(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(734982615); CREATE TABLE IF NOT EXISTS rooms_schema(checksum bytea PRIMARY KEY)`); err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(schema))
	var old []byte
	err = tx.QueryRow(ctx, `SELECT checksum FROM rooms_schema`).Scan(&old)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err = tx.Exec(ctx, schema); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO rooms_schema VALUES($1)`, sum[:])
	} else if err == nil && subtle.ConstantTimeCompare(old, sum[:]) != 1 {
		return errors.New("unsupported room schema")
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS rooms_migrations(name text PRIMARY KEY,checksum bytea NOT NULL)`); err != nil {
		return err
	}
	botSum := sha256.Sum256([]byte(botsMigration))
	err = tx.QueryRow(ctx, `SELECT checksum FROM rooms_migrations WHERE name='bots-v1'`).Scan(&old)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err = tx.Exec(ctx, botsMigration); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO rooms_migrations VALUES('bots-v1',$1)`, botSum[:])
	} else if err == nil && subtle.ConstantTimeCompare(old, botSum[:]) != 1 {
		return errors.New("unsupported room bot migration")
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func randomID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func active(ctx context.Context, tx pgx.Tx, actor string) error {
	var id string
	err := tx.QueryRow(ctx, `SELECT account_id FROM identity_accounts WHERE account_id=$1 AND kind IN ('guest','member') FOR SHARE`, actor).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrForbidden
	}
	return err
}
func load(ctx context.Context, tx pgx.Tx, id string, exclusive bool) (Room, error) {
	var r Room
	lock := " FOR SHARE"
	if exclusive {
		lock = " FOR UPDATE"
	}
	err := tx.QueryRow(ctx, `SELECT room_id,owner_id,capacity,games,status,coalesce(match_id,''),bot_difficulty FROM rooms WHERE room_id=$1`+lock, id).Scan(&r.ID, &r.OwnerID, &r.Capacity, &r.Games, &r.Status, &r.MatchID, &r.BotDifficulty)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrForbidden
	}
	if err != nil {
		return r, err
	}
	rows, err := tx.Query(ctx, `SELECT actor_id,seat,kind='bot' FROM room_members JOIN identity_accounts ON account_id=actor_id WHERE room_id=$1 ORDER BY seat`, id)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Member
		if err = rows.Scan(&m.ActorID, &m.Seat, &m.Bot); err != nil {
			return r, err
		}
		r.Members = append(r.Members, m)
	}
	return r, rows.Err()
}
func fingerprint(v any) []byte { b, _ := json.Marshal(v); sum := sha256.Sum256(b); return sum[:] }
func lookup(ctx context.Context, tx pgx.Tx, actor, id string, hash []byte, out any) (bool, error) {
	var prior, result []byte
	err := tx.QueryRow(ctx, `SELECT body_hash,result FROM room_commands WHERE actor_id=$1 AND command_id=$2`, actor, id).Scan(&prior, &result)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !bytes.Equal(prior, hash) {
		return true, ErrConflict
	}
	if json.Unmarshal(result, out) != nil {
		return true, ErrConflict
	}
	return true, nil
}
func receipt(ctx context.Context, tx pgx.Tx, actor, id string, hash []byte, result any) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO room_commands VALUES($1,$2,$3,$4)`, actor, id, hash, b)
	return err
}
func (s *Store) Create(ctx context.Context, actor, command string, capacity, games int) (Room, error) {
	return s.CreateWithBots(ctx, actor, command, capacity, games, "")
}

// CreateWithBots atomically owns the opponent identities; callers never choose
// their IDs, credentials or seats. Empty difficulty preserves human-only rooms.
func (s *Store) CreateWithBots(ctx context.Context, actor, command string, capacity, games int, difficulty string) (Room, error) {
	if command == "" || len(command) > 256 || capacity != 3 && capacity != 4 || games < 0 || games > 100 {
		return Room{}, ErrInvalid
	}
	if difficulty != "" && difficulty != "beginner" && difficulty != "standard" && difficulty != "advanced" {
		return Room{}, ErrInvalid
	}
	hash := fingerprint(struct {
		Kind            string
		Capacity, Games int
	}{"create", capacity, games})
	// Keep old human-room receipt hashes byte-for-byte compatible.
	if difficulty != "" {
		hash = fingerprint([]any{"create-bots-v1", capacity, games, difficulty})
	}
	if games == 0 {
		games = 3
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Room{}, err
	}
	defer rollback(tx)
	if err = active(ctx, tx, actor); err != nil {
		return Room{}, err
	}
	// A per-actor command key also serializes concurrent room creations, including
	// distinct room IDs generated after an identical lost request.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,734982615))`, actor); err != nil {
		return Room{}, err
	}
	var r Room
	if found, err := lookup(ctx, tx, actor, command, hash, &r); found || err != nil {
		return r, err
	}
	id, err := randomID()
	if err != nil {
		return Room{}, err
	}
	r = Room{ID: id, OwnerID: actor, Capacity: capacity, Games: games, Status: "open", Members: []Member{{ActorID: actor, Seat: 1}}, BotDifficulty: difficulty}
	if _, err = tx.Exec(ctx, `INSERT INTO rooms(room_id,owner_id,capacity,games,status,match_id,bot_difficulty) VALUES($1,$2,$3,$4,'open',NULL,$5)`, id, actor, capacity, games, difficulty); err != nil {
		return Room{}, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO room_members VALUES($1,$2,1)`, id, actor); err != nil {
		return Room{}, err
	}
	if difficulty != "" {
		for seat := 2; seat <= capacity; seat++ {
			botID, e := randomID()
			if e != nil {
				return Room{}, e
			}
			if _, err = tx.Exec(ctx, `INSERT INTO identity_accounts(account_id,kind) VALUES($1,'bot')`, botID); err != nil {
				return Room{}, err
			}
			if _, err = tx.Exec(ctx, `INSERT INTO room_members VALUES($1,$2,$3)`, id, botID, seat); err != nil {
				return Room{}, err
			}
			r.Members = append(r.Members, Member{ActorID: botID, Seat: seat, Bot: true})
		}
	}
	if err = receipt(ctx, tx, actor, command, hash, r); err != nil {
		return Room{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Room{}, matchstore.ErrOutcomeUnknown
	}
	return r, nil
}
func (s *Store) Get(ctx context.Context, actor, id string) (Room, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Room{}, err
	}
	defer rollback(tx)
	if err = active(ctx, tx, actor); err != nil {
		return Room{}, err
	}
	r, err := load(ctx, tx, id, false)
	if err != nil {
		return r, err
	}
	for _, m := range r.Members {
		if m.ActorID == actor {
			return r, nil
		}
	}
	return Room{}, ErrForbidden
}
func (s *Store) Invite(ctx context.Context, actor, id string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer rollback(tx)
	if err = active(ctx, tx, actor); err != nil {
		return "", err
	}
	r, err := load(ctx, tx, id, true)
	if err != nil {
		return "", err
	}
	if r.OwnerID != actor {
		return "", ErrForbidden
	}
	if r.Status != "open" {
		return "", ErrClosed
	}
	token, err := randomID()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(token))
	if _, err = tx.Exec(ctx, `INSERT INTO room_invitations VALUES($1,$2,$3) ON CONFLICT(room_id) DO UPDATE SET token_hash=EXCLUDED.token_hash,expires_at=EXCLUDED.expires_at`, id, hash[:], time.Now().UTC().Add(24*time.Hour)); err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", matchstore.ErrOutcomeUnknown
	}
	return token, nil
}
func (s *Store) Join(ctx context.Context, actor, id, invite string) (Room, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Room{}, err
	}
	defer rollback(tx)
	if err = active(ctx, tx, actor); err != nil {
		return Room{}, err
	}
	r, err := load(ctx, tx, id, true)
	if err != nil {
		return Room{}, err
	}
	if r.Status != "open" {
		return Room{}, ErrClosed
	}
	for _, m := range r.Members {
		if m.ActorID == actor {
			return r, nil
		}
	}
	hash := sha256.Sum256([]byte(invite))
	var accepted bool
	if len(invite) > 128 {
		return Room{}, ErrForbidden
	}
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM room_invitations WHERE room_id=$1 AND token_hash=$2 AND expires_at>now())`, id, hash[:]).Scan(&accepted); err != nil {
		return Room{}, err
	}
	if !accepted {
		return Room{}, ErrForbidden
	}
	if len(r.Members) >= r.Capacity {
		return Room{}, ErrFull
	}
	seat := len(r.Members) + 1
	if _, err = tx.Exec(ctx, `INSERT INTO room_members VALUES($1,$2,$3)`, id, actor, seat); err != nil {
		return Room{}, err
	}
	r.Members = append(r.Members, Member{ActorID: actor, Seat: seat})
	if err = tx.Commit(ctx); err != nil {
		return Room{}, matchstore.ErrOutcomeUnknown
	}
	return r, nil
}
func (s *Store) Start(ctx context.Context, actor, id, command string) (string, error) {
	if command == "" || len(command) > 256 {
		return "", ErrInvalid
	}
	hash := fingerprint(struct{ Kind, RoomID string }{"start", id})
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer rollback(tx)
	// Observe candidate membership, exclusively lock all account rows in sorted order, then
	// lock/re-read room. A concurrent join requires a fresh start attempt; never
	// acquire a newly discovered account lock after the room lock. Economy admission
	// also writes these accounts: starting with SHARE would allow overlapping room
	// starts to hold each other's locks while both attempt an UPDATE upgrade.
	rows, err := tx.Query(ctx, `SELECT actor_id FROM room_members WHERE room_id=$1 ORDER BY actor_id`, id)
	if err != nil {
		return "", err
	}
	var actors []string
	for rows.Next() {
		var a string
		if err = rows.Scan(&a); err != nil {
			rows.Close()
			return "", err
		}
		actors = append(actors, a)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return "", err
	}
	if !slices.Contains(actors, actor) {
		return "", ErrForbidden
	}
	allActive := true
	for _, a := range actors {
		var kind string
		err = tx.QueryRow(ctx, `SELECT kind FROM identity_accounts WHERE account_id=$1 FOR UPDATE`, a).Scan(&kind)
		if err != nil {
			return "", ErrForbidden
		}
		if kind == "deleted" {
			allActive = false
			if a == actor {
				return "", ErrForbidden
			}
		}
	}
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,734982615))`, actor); err != nil {
		return "", err
	}
	r, err := load(ctx, tx, id, true)
	if err != nil {
		return "", err
	}
	if r.OwnerID != actor {
		return "", ErrForbidden
	}
	var matchID string
	if found, err := lookup(ctx, tx, actor, command, hash, &matchID); found || err != nil {
		return matchID, err
	}
	if !allActive {
		return "", ErrForbidden
	}
	if r.Status != "open" {
		return "", ErrClosed
	}
	if len(r.Members) != r.Capacity {
		return "", ErrNotReady
	}
	var ordered, sorted []string
	for _, m := range r.Members {
		ordered = append(ordered, m.ActorID)
		sorted = append(sorted, m.ActorID)
	}
	slices.Sort(sorted)
	if !slices.Equal(sorted, actors) {
		return "", ErrConflict
	}
	matchID, err = randomID()
	if err != nil {
		return "", err
	}
	gameID, err := randomID()
	if err != nil {
		return "", err
	}
	board, err := game.NewState(r.Capacity, gameID)
	if err != nil {
		return "", err
	}
	match, err := game.NewMatchLifecycle(board, r.Games)
	if err != nil {
		return "", err
	}
	env, err := matchstore.NewEnvelope(match, s.rulesHash)
	if err != nil {
		return "", err
	}
	env.BotDifficulty = r.BotDifficulty
	if err = s.matches.CreateTx(ctx, tx, matchID, env, ordered); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `UPDATE rooms SET status='started',match_id=$2 WHERE room_id=$1`, id, matchID); err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM room_invitations WHERE room_id=$1`, id); err != nil {
		return "", err
	}
	if err = receipt(ctx, tx, actor, command, hash, matchID); err != nil {
		return "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return "", matchstore.ErrOutcomeUnknown
	}
	return matchID, nil
}
