// Package identity owns local credentials and opaque revocable sessions. Match
// identity survives guest upgrade and account tombstoning; no game action is implied.
package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

var (
	ErrCredentials = errors.New("invalid or expired credentials")
	ErrInvalid     = errors.New("invalid identity input")
	ErrConflict    = errors.New("identity already exists or cannot be upgraded")
)

const SessionTTL = 24 * time.Hour
const passwordPrefix = "argon2id-v19-m19456-t2-p1"

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,32}$`)

//go:embed migrations/001_identity.sql
var schema string

//go:embed migrations/002_bots.sql
var botsMigration string

type Account struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Username string `json:"username,omitempty"`
}
type Session struct {
	Token     string    `json:"token"`
	Account   Account   `json:"account"`
	ExpiresAt time.Time `json:"expires_at"`
}
type Store struct {
	pool      *pgxpool.Pool
	hashSlots chan struct{}
	now       func() time.Time
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool, hashSlots: make(chan struct{}, 2), now: time.Now}
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
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(734982614); CREATE TABLE IF NOT EXISTS identity_schema(checksum bytea PRIMARY KEY)`); err != nil {
		return err
	}
	sum := sha256.Sum256([]byte(schema))
	var old []byte
	err = tx.QueryRow(ctx, `SELECT checksum FROM identity_schema`).Scan(&old)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err = tx.Exec(ctx, schema); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO identity_schema VALUES($1)`, sum[:])
	} else if err == nil && subtle.ConstantTimeCompare(old, sum[:]) != 1 {
		return errors.New("unsupported identity schema")
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS identity_migrations(name text PRIMARY KEY,checksum bytea NOT NULL)`); err != nil {
		return err
	}
	botSum := sha256.Sum256([]byte(botsMigration))
	err = tx.QueryRow(ctx, `SELECT checksum FROM identity_migrations WHERE name='bots-v1'`).Scan(&old)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, err = tx.Exec(ctx, botsMigration); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO identity_migrations VALUES('bots-v1',$1)`, botSum[:])
	} else if err == nil && subtle.ConstantTimeCompare(old, botSum[:]) != 1 {
		return errors.New("unsupported identity bot migration")
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func opaque() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func tokenHash(token string) ([]byte, error) {
	if len(token) != 43 {
		return nil, ErrCredentials
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(b) != 32 || base64.RawURLEncoding.EncodeToString(b) != token {
		return nil, ErrCredentials
	}
	h := sha256.Sum256([]byte(token))
	return h[:], nil
}
func normalizeUsername(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if !usernamePattern.MatchString(v) {
		return "", ErrInvalid
	}
	return v, nil
}
func validPassword(v string) bool {
	return utf8.ValidString(v) && len(v) <= 128 && utf8.RuneCountInString(v) >= 15
}
func (s *Store) derive(ctx context.Context, password string, salt []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case s.hashSlots <- struct{}{}:
		defer func() { <-s.hashSlots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	out := argon2.IDKey([]byte(password), salt, 2, 19456, 1, 32)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
func (s *Store) passwordHash(ctx context.Context, password string) ([]byte, error) {
	if !validPassword(password) {
		return nil, ErrInvalid
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	h, err := s.derive(ctx, password, salt)
	if err != nil {
		return nil, err
	}
	return []byte(passwordPrefix + "$" + base64.RawStdEncoding.EncodeToString(salt) + "$" + base64.RawStdEncoding.EncodeToString(h)), nil
}
func (s *Store) passwordMatches(ctx context.Context, encoded []byte, password string) bool {
	if !validPassword(password) {
		return false
	}
	parts := strings.Split(string(encoded), "$")
	if len(parts) != 3 || parts[0] != passwordPrefix {
		return false
	}
	salt, e := base64.RawStdEncoding.DecodeString(parts[1])
	if e != nil || len(salt) != 16 {
		return false
	}
	want, e := base64.RawStdEncoding.DecodeString(parts[2])
	if e != nil || len(want) != 32 {
		return false
	}
	got, e := s.derive(ctx, password, salt)
	return e == nil && subtle.ConstantTimeCompare(got, want) == 1
}
func conflict(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return ErrConflict
	}
	return err
}
func (s *Store) issue(ctx context.Context, tx pgx.Tx, a Account) (Session, error) {
	token, err := opaque()
	if err != nil {
		return Session{}, err
	}
	hash, _ := tokenHash(token)
	out := Session{token, a, s.now().UTC().Add(SessionTTL)}
	_, err = tx.Exec(ctx, `INSERT INTO identity_sessions(token_hash,account_id,expires_at) VALUES($1,$2,$3)`, hash, a.ID, out.ExpiresAt)
	return out, err
}
func (s *Store) CreateGuest(ctx context.Context) (Session, error) { return s.create(ctx, "", nil) }
func (s *Store) Register(ctx context.Context, username, password string) (Session, error) {
	name, err := normalizeUsername(username)
	if err != nil {
		return Session{}, err
	}
	h, err := s.passwordHash(ctx, password)
	if err != nil {
		return Session{}, err
	}
	return s.create(ctx, name, h)
}
func (s *Store) create(ctx context.Context, name string, password []byte) (Session, error) {
	id, err := opaque()
	if err != nil {
		return Session{}, err
	}
	a := Account{ID: id, Kind: "guest"}
	var username any
	if name != "" {
		a.Kind = "member"
		a.Username = name
		username = name
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Session{}, err
	}
	defer rollback(tx)
	if _, err = tx.Exec(ctx, `INSERT INTO identity_accounts VALUES($1,$2,$3,$4)`, id, a.Kind, username, password); err != nil {
		return Session{}, conflict(err)
	}
	out, err := s.issue(ctx, tx, a)
	if err != nil {
		return Session{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Session{}, err
	}
	return out, nil
}
func (s *Store) Authenticate(ctx context.Context, token string) (Account, error) {
	hash, err := tokenHash(token)
	if err != nil {
		return Account{}, err
	}
	var a Account
	err = s.pool.QueryRow(ctx, `SELECT a.account_id,a.kind,coalesce(a.username,'') FROM identity_sessions s JOIN identity_accounts a USING(account_id) WHERE s.token_hash=$1 AND s.expires_at>$2 AND a.kind IN ('guest','member')`, hash, s.now().UTC()).Scan(&a.ID, &a.Kind, &a.Username)
	if errors.Is(err, pgx.ErrNoRows) {
		err = ErrCredentials
	}
	return a, err
}
func (s *Store) Login(ctx context.Context, username, password string) (Session, error) {
	name, err := normalizeUsername(username)
	if err != nil || !validPassword(password) {
		return Session{}, ErrCredentials
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Session{}, err
	}
	defer rollback(tx)
	var a Account
	var encoded []byte
	err = tx.QueryRow(ctx, `SELECT account_id,kind,username,password_hash FROM identity_accounts WHERE username=$1 AND kind='member' FOR UPDATE`, name).Scan(&a.ID, &a.Kind, &a.Username, &encoded)
	if errors.Is(err, pgx.ErrNoRows) {
		_, _ = s.derive(ctx, password, make([]byte, 16))
		return Session{}, ErrCredentials
	}
	if err != nil {
		return Session{}, err
	}
	if !s.passwordMatches(ctx, encoded, password) {
		return Session{}, ErrCredentials
	}
	out, err := s.issue(ctx, tx, a)
	if err != nil {
		return Session{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Session{}, err
	}
	return out, nil
}
func (s *Store) lockSession(ctx context.Context, tx pgx.Tx, token string) (Account, error) {
	hash, err := tokenHash(token)
	if err != nil {
		return Account{}, err
	}
	var id string
	if err = tx.QueryRow(ctx, `SELECT account_id FROM identity_sessions WHERE token_hash=$1`, hash).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = ErrCredentials
		}
		return Account{}, err
	}
	var a Account
	if err = tx.QueryRow(ctx, `SELECT account_id,kind,coalesce(username,'') FROM identity_accounts WHERE account_id=$1 FOR UPDATE`, id).Scan(&a.ID, &a.Kind, &a.Username); err != nil {
		return a, err
	}
	var valid bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identity_sessions WHERE token_hash=$1 AND account_id=$2 AND expires_at>$3)`, hash, id, s.now().UTC()).Scan(&valid)
	if err == nil && (!valid || a.Kind == "deleted") {
		err = ErrCredentials
	}
	return a, err
}
func (s *Store) Upgrade(ctx context.Context, token, username, password string) (Session, error) {
	name, err := normalizeUsername(username)
	if err != nil {
		return Session{}, err
	}
	h, err := s.passwordHash(ctx, password)
	if err != nil {
		return Session{}, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Session{}, err
	}
	defer rollback(tx)
	a, err := s.lockSession(ctx, tx, token)
	if err != nil {
		return Session{}, err
	}
	if a.Kind != "guest" {
		return Session{}, ErrConflict
	}
	if _, err = tx.Exec(ctx, `UPDATE identity_accounts SET kind='member',username=$2,password_hash=$3 WHERE account_id=$1`, a.ID, name, h); err != nil {
		return Session{}, conflict(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM identity_sessions WHERE account_id=$1`, a.ID); err != nil {
		return Session{}, err
	}
	a.Kind = "member"
	a.Username = name
	out, err := s.issue(ctx, tx, a)
	if err != nil {
		return Session{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return Session{}, err
	}
	return out, nil
}
func (s *Store) Logout(ctx context.Context, token string) error {
	hash, err := tokenHash(token)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `DELETE FROM identity_sessions WHERE token_hash=$1`, hash)
	return err
}
func (s *Store) Delete(ctx context.Context, token string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer rollback(tx)
	a, err := s.lockSession(ctx, tx, token)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE identity_accounts SET kind='deleted',username=NULL,password_hash=NULL WHERE account_id=$1`, a.ID); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM identity_sessions WHERE account_id=$1`, a.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
