// Package transport exposes the authenticated online authority. Private engine
// snapshots and SQL errors never cross this boundary.
package transport

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/canonical"
	"github.com/metaphy6/cgms/backend/internal/delivery"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/presence"
	"github.com/metaphy6/cgms/backend/internal/rooms"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
	"github.com/metaphy6/cgms/backend/internal/wire"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	EconomyPolicy *economy.Policy
	MaxStreams    int
	// ProxyClientIP is only safe behind an isolated proxy that overwrites X-Real-IP.
	ProxyClientIP               bool
	Telemetry                   *telemetry.Recorder
	Pool                        *pgxpool.Pool
	RulesHash, Origin, RedisURL string
	HandleKey                   []byte
	InsecureLocal               bool
}
type Server struct {
	cfg           Config
	identity      *identity.Store
	rooms         *rooms.Store
	economy       *economy.Service
	matches       *matchstore.Store
	codec         *wire.Codec
	dispatcher    *delivery.Dispatcher
	streams       *delivery.Streams
	presence      *presence.Tracker
	mux           *http.ServeMux
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.Mutex
	rates         map[string]rate
	active        chan struct{}
	startOnce     sync.Once
	lifecycleMu   sync.Mutex
	closed        bool
	closeOnce     sync.Once
	done          chan struct{}
	automaticWake chan struct{}
}
type rate struct {
	tokens float64
	at     time.Time
}
type APIError struct {
	Code      string            `json:"code"`
	Args      map[string]string `json:"args"`
	RequestID string            `json:"request_id"`
}

func New(cfg Config) (*Server, error) {
	if cfg.MaxStreams == 0 {
		cfg.MaxStreams = 64
	}
	if cfg.MaxStreams < 1 || cfg.MaxStreams > 4096 {
		return nil, errors.New("invalid stream capacity")
	}
	if cfg.Pool == nil || cfg.Origin == "" || cfg.RulesHash == "" {
		return nil, errors.New("missing server configuration")
	}
	origin, e := url.Parse(cfg.Origin)
	if e != nil || origin.Host == "" || (origin.Scheme != "https" && !(cfg.InsecureLocal && origin.Scheme == "http")) || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" || origin.User != nil {
		return nil, errors.New("invalid origin")
	}
	codec, e := wire.New(cfg.HandleKey)
	if e != nil {
		return nil, e
	}
	matches := matchstore.New(cfg.Pool, cfg.RulesHash)
	var accounts *economy.Service
	if cfg.EconomyPolicy != nil {
		// Copy the value so caller mutation cannot change active prices after boot.
		policy := *cfg.EconomyPolicy
		cfg.EconomyPolicy = &policy
		matches, e = matchstore.NewWithEconomy(cfg.Pool, cfg.RulesHash, policy)
		if e != nil {
			return nil, e
		}
		accounts, e = economy.NewService(cfg.Pool, policy)
		if e != nil {
			return nil, e
		}
	}
	p, e := presence.New(cfg.RedisURL, presence.Options{})
	if e != nil {
		return nil, e
	}
	base := context.Background()
	if cfg.Telemetry != nil {
		base = cfg.Telemetry.Context(base)
	}
	ctx, cancel := context.WithCancel(base)
	s := &Server{cfg: cfg, codec: codec, identity: identity.New(cfg.Pool), matches: matches, economy: accounts, dispatcher: delivery.New(128, 32), streams: delivery.NewStreams(delivery.Options{MaxConnections: cfg.MaxStreams, OriginPatterns: []string{origin.Host}}), presence: p, ctx: ctx, cancel: cancel, rates: map[string]rate{}, done: make(chan struct{}), automaticWake: make(chan struct{}, 1), active: make(chan struct{}, 32)}
	s.rooms = rooms.New(cfg.Pool, s.matches, cfg.RulesHash)
	s.routes()
	return s, nil
}
func (s *Server) allowedOrigin(r *http.Request) bool {
	return r.Header.Get("Origin") == "" || r.Header.Get("Origin") == s.cfg.Origin
}
func (s *Server) admit(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	v, ok := s.rates[key]
	if !ok {
		if len(s.rates) >= 1024 {
			for k, x := range s.rates {
				if now.Sub(x.at) > time.Minute {
					delete(s.rates, k)
				}
			}
			if len(s.rates) >= 1024 {
				return false
			}
		}
		v = rate{20, now}
	}
	v.tokens = min(20, v.tokens+now.Sub(v.at).Seconds()*5)
	v.at = now
	allowed := v.tokens >= 1
	if allowed {
		v.tokens--
	}
	s.rates[key] = v
	return allowed
}

type admissionKey struct{}

type responseStatus struct {
	http.ResponseWriter
	status int
}

func (w *responseStatus) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (w *responseStatus) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Telemetry != nil {
		ctx, done := s.cfg.Telemetry.Start(r.Context(), telemetry.Request)
		observed := &responseStatus{ResponseWriter: w}
		w = observed
		defer func() {
			if observed.status >= 400 {
				done(errors.New("request rejected"))
			} else {
				done(nil)
			}
		}()
		r = r.WithContext(ctx)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	id := make([]byte, 12)
	_, _ = rand.Read(id)
	w.Header().Set("X-Request-ID", hex.EncodeToString(id))
	if !s.allowedOrigin(r) {
		s.fail(w, 403, "FORBIDDEN")
		return
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	if s.cfg.ProxyClientIP {
		ip := net.ParseIP(r.Header.Get("X-Real-IP"))
		if ip == nil {
			s.fail(w, 400, "INVALID_PROXY_IDENTITY")
			return
		}
		host = ip.String()
	}
	if !s.admit("ip:" + host) {
		s.fail(w, 429, "OVERLOADED")
		return
	}
	select {
	case s.active <- struct{}{}:
	default:
		s.fail(w, 429, "OVERLOADED")
		return
	}
	var once sync.Once
	release := func() { once.Do(func() { <-s.active }) }
	defer release()
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	r = r.WithContext(context.WithValue(ctx, admissionKey{}, release))
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	s.mux.ServeHTTP(w, r)
}
func (s *Server) fail(w http.ResponseWriter, status int, code string) {
	s.write(w, status, APIError{code, map[string]string{}, w.Header().Get("X-Request-ID")})
}
func (s *Server) write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func decode(r *http.Request, v any) error {
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		return errors.New("content type")
	}
	raw, e := io.ReadAll(r.Body)
	if e != nil {
		return e
	}
	if !strings.HasPrefix(strings.TrimSpace(string(raw)), "{") {
		return errors.New("object required")
	}
	return canonical.DecodeLimit(raw, v, 16<<10)
}
func (s *Server) csrf(token string) string {
	h := hmac.New(sha256.New, s.cfg.HandleKey)
	h.Write([]byte("csrf:" + token))
	return hex.EncodeToString(h.Sum(nil))
}
func (s *Server) token(r *http.Request) (string, bool, error) {
	auth := r.Header.Get("Authorization")
	cookie, e := r.Cookie("cgms_session")
	if auth != "" {
		if e == nil || !strings.HasPrefix(auth, "Bearer ") {
			return "", false, errors.New("ambiguous credentials")
		}
		return strings.TrimPrefix(auth, "Bearer "), false, nil
	}
	if e != nil {
		return "", false, e
	}
	return cookie.Value, true, nil
}
func (s *Server) account(w http.ResponseWriter, r *http.Request) (identity.Account, string, bool) {
	token, cookie, e := s.token(r)
	if e == nil && cookie && r.Method != "GET" {
		if r.Header.Get("Origin") != s.cfg.Origin || !hmac.Equal([]byte(r.Header.Get("X-CSRF-Token")), []byte(s.csrf(token))) {
			s.fail(w, 403, "FORBIDDEN")
			return identity.Account{}, "", false
		}
	}
	var a identity.Account
	if e == nil {
		a, e = s.identity.Authenticate(r.Context(), token)
		if e != nil && !errors.Is(e, identity.ErrCredentials) {
			s.fail(w, http.StatusServiceUnavailable, "UNAVAILABLE")
			return a, "", false
		}
	}
	if e != nil {
		slog.Warn("authentication failed")
		s.fail(w, 401, "UNAUTHENTICATED")
		return a, "", false
	}
	if !s.admit("account:" + a.ID) {
		s.fail(w, 429, "OVERLOADED")
		return a, "", false
	}
	return a, token, true
}
func (s *Server) issue(w http.ResponseWriter, r *http.Request, v identity.Session) {
	if r.Header.Get("Origin") != "" {
		http.SetCookie(w, &http.Cookie{Name: "cgms_session", Value: v.Token, Path: "/", HttpOnly: true, Secure: !s.cfg.InsecureLocal, SameSite: http.SameSiteStrictMode, Expires: v.ExpiresAt})
		s.write(w, 200, map[string]any{"account": v.Account, "expires_at": v.ExpiresAt, "csrf_token": s.csrf(v.Token)})
	} else {
		s.write(w, 200, v)
	}
}
func (s *Server) err(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, economy.ErrAllowance):
		s.fail(w, 409, "ALLOWANCE_EXHAUSTED")
	case errors.Is(e, matchstore.ErrUnauthorized):
		s.fail(w, 403, "FORBIDDEN")
	case errors.Is(e, matchstore.ErrConflict):
		s.fail(w, 409, "COMMAND_CONFLICT")
	case errors.Is(e, matchstore.ErrStale):
		s.fail(w, 409, "STATE_CONFLICT")
	case errors.Is(e, matchstore.ErrInvalid), errors.Is(e, wire.ErrInvalid):
		s.fail(w, 400, "INVALID_COMMAND")
	case errors.Is(e, matchstore.ErrCursor):
		s.fail(w, 409, "RESNAPSHOT_REQUIRED")
	case errors.Is(e, matchstore.ErrOutcomeUnknown), errors.Is(e, context.DeadlineExceeded):
		s.fail(w, 503, "OUTCOME_UNKNOWN")
	case errors.Is(e, delivery.ErrOverloaded), errors.Is(e, delivery.ErrClosed):
		s.fail(w, 429, "OVERLOADED")
	default:
		s.fail(w, 503, "UNAVAILABLE")
	}
}
func (s *Server) routes() {
	m := http.NewServeMux()
	s.mux = m
	s.economyRoutes(m)
	m.HandleFunc("POST /v1/guests", func(w http.ResponseWriter, r *http.Request) {
		var body struct{}
		if decode(r, &body) != nil {
			s.fail(w, 400, "INVALID_REQUEST")
			return
		}
		v, e := s.identity.CreateGuest(r.Context())
		if e != nil {
			s.err(w, e)
			return
		}
		s.issue(w, r, v)
	})
	for _, action := range []string{"register", "login", "upgrade"} {
		m.HandleFunc("POST /v1/"+action, func(w http.ResponseWriter, r *http.Request) {
			var b struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if decode(r, &b) != nil {
				s.fail(w, 400, "INVALID_REQUEST")
				return
			}
			var v identity.Session
			var e error
			switch action {
			case "register":
				v, e = s.identity.Register(r.Context(), b.Username, b.Password)
			case "login":
				v, e = s.identity.Login(r.Context(), b.Username, b.Password)
			case "upgrade":
				_, token, ok := s.account(w, r)
				if !ok {
					return
				}
				v, e = s.identity.Upgrade(r.Context(), token, b.Username, b.Password)
			}
			if e != nil {
				slog.Warn("credential operation rejected")
				s.fail(w, 400, "CREDENTIALS_REJECTED")
				return
			}
			s.issue(w, r, v)
		})
	}
	m.HandleFunc("GET /v1/session", func(w http.ResponseWriter, r *http.Request) {
		a, t, ok := s.account(w, r)
		if ok {
			s.write(w, 200, map[string]any{"account": a, "csrf_token": s.csrf(t)})
		}
	})
	for _, action := range []string{"logout", "account"} {
		method := "POST"
		if action == "account" {
			method = "DELETE"
		}
		m.HandleFunc(method+" /v1/"+action, func(w http.ResponseWriter, r *http.Request) {
			_, token, ok := s.account(w, r)
			if !ok {
				return
			}
			var e error
			if action == "logout" {
				e = s.identity.Logout(r.Context(), token)
			} else {
				e = s.identity.Delete(r.Context(), token)
			}
			if e != nil {
				s.err(w, e)
				return
			}
			http.SetCookie(w, &http.Cookie{Name: "cgms_session", Value: "", Path: "/", HttpOnly: true, Secure: !s.cfg.InsecureLocal, MaxAge: -1, SameSite: http.SameSiteStrictMode})
			s.write(w, 200, map[string]bool{"ok": true})
		})
	}
	m.HandleFunc("POST /v1/rooms", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		var b struct {
			CommandID     string `json:"command_id"`
			Capacity      int    `json:"capacity"`
			Games         int    `json:"games_per_match"`
			BotDifficulty string `json:"bot_difficulty"`
		}
		if decode(r, &b) != nil {
			s.fail(w, 400, "INVALID_REQUEST")
			return
		}
		v, e := s.rooms.CreateWithBots(r.Context(), a.ID, b.CommandID, b.Capacity, b.Games, b.BotDifficulty)
		if e != nil {
			s.roomError(w, e)
			return
		}
		s.write(w, 200, v)
	})
	m.HandleFunc("GET /v1/rooms/{room}", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		v, e := s.rooms.Get(r.Context(), a.ID, r.PathValue("room"))
		if e != nil {
			s.roomError(w, e)
			return
		}
		s.write(w, 200, v)
	})
	m.HandleFunc("POST /v1/rooms/{room}/invitations", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		v, e := s.rooms.Invite(r.Context(), a.ID, r.PathValue("room"))
		if e != nil {
			s.roomError(w, e)
			return
		}
		s.write(w, 200, map[string]string{"invitation": v})
	})
	m.HandleFunc("POST /v1/rooms/{room}/join", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		var b struct {
			Invitation string `json:"invitation"`
		}
		if decode(r, &b) != nil {
			s.fail(w, 400, "INVALID_REQUEST")
			return
		}
		v, e := s.rooms.Join(r.Context(), a.ID, r.PathValue("room"), b.Invitation)
		if e != nil {
			s.roomError(w, e)
			return
		}
		s.write(w, 200, v)
	})
	m.HandleFunc("POST /v1/rooms/{room}/matches", func(w http.ResponseWriter, r *http.Request) {
		a, _, ok := s.account(w, r)
		if !ok {
			return
		}
		var b struct {
			CommandID string `json:"command_id"`
		}
		if decode(r, &b) != nil {
			s.fail(w, 400, "INVALID_REQUEST")
			return
		}
		id, e := s.rooms.Start(r.Context(), a.ID, r.PathValue("room"), b.CommandID)
		if e != nil {
			s.roomError(w, e)
			return
		}
		s.write(w, 200, map[string]string{"match_id": id})
	})
	m.HandleFunc("GET /v1/matches/{match}", s.snapshot)
	m.HandleFunc("POST /v1/matches/{match}/commands", s.command)
	m.HandleFunc("GET /v1/matches/{match}/commands/{command}", s.reconcile)
	m.HandleFunc("GET /v1/matches/{match}/events", s.events)
	m.HandleFunc("GET /v1/matches/{match}/stream", s.stream)
}
func (s *Server) roomError(w http.ResponseWriter, e error) {
	if errors.Is(e, matchstore.ErrOutcomeUnknown) || errors.Is(e, context.DeadlineExceeded) || errors.Is(e, economy.ErrAllowance) {
		s.err(w, e)
		return
	}
	s.fail(w, 409, "ROOM_REJECTED")
}
func (s *Server) snapshot(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.account(w, r)
	if !ok {
		return
	}
	id := r.PathValue("match")
	v, e := s.matches.Snapshot(r.Context(), id, a.ID)
	if e == nil {
		v.Projection, e = s.codec.EncodeProjection(id, a.ID, v.Projection)
	}
	if e != nil {
		s.err(w, e)
		return
	}
	s.write(w, 200, v)
}
func (s *Server) command(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.account(w, r)
	if !ok {
		return
	}
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		s.fail(w, 400, "INVALID_COMMAND")
		return
	}
	raw, e := io.ReadAll(r.Body)
	if e != nil {
		s.fail(w, 400, "INVALID_COMMAND")
		return
	}
	id := r.PathValue("match")
	in, e := s.codec.DecodeIntent(id, a.ID, raw, nil)
	if e != nil {
		s.err(w, e)
		return
	}
	var meta struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(raw, &meta)
	victory := meta.Type == "coup" || meta.Type == "declare-ordinary"
	work := func(ctx context.Context) (any, error) {
		return s.matches.SubmitChecked(ctx, id, a.ID, in, func(env matchstore.Envelope, seat int, _ matchstore.Intent) (matchstore.Intent, error) {
			p, e := matchstore.SeatProjection(env, seat)
			if e != nil {
				return in, e
			}
			b, e := json.Marshal(p)
			if e != nil {
				return in, e
			}
			return s.codec.DecodeIntent(id, a.ID, raw, b)
		})
	}
	var v any
	if victory {
		v, e = s.dispatcher.Submit(r.Context(), id, -1, work)
	} else {
		v, e = s.dispatcher.SubmitRanked(r.Context(), id, func(ctx context.Context) (int, error) { return s.matches.Priority(ctx, id, a.ID, false) }, work)
	}
	if e != nil {
		s.err(w, e)
		return
	}
	result := v.(matchstore.Result)
	result.Projection, e = s.codec.EncodeProjection(id, a.ID, result.Projection)
	if e != nil {
		s.err(w, e)
		return
	}
	s.write(w, 200, result)
}
func (s *Server) reconcile(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.account(w, r)
	if !ok {
		return
	}
	id := r.PathValue("match")
	v, e := s.matches.Lookup(r.Context(), id, a.ID, r.PathValue("command"))
	if e == nil {
		v.Projection, e = s.codec.EncodeProjection(id, a.ID, v.Projection)
	}
	if e != nil {
		s.err(w, e)
		return
	}
	s.write(w, 200, v)
}
func (s *Server) catchup(ctx context.Context, id, actor string, after int64) ([]delivery.Event, error) {
	v, e := s.matches.Events(ctx, id, actor, after, 128)
	if e != nil {
		return nil, e
	}
	out := make([]delivery.Event, 0, len(v))
	for _, x := range v {
		p, e := s.codec.EncodeProjection(id, actor, x.Projection)
		if e != nil {
			return nil, e
		}
		b, _ := json.Marshal(matchstore.View{Version: x.Cursor, Cursor: x.Cursor, Projection: p})
		out = append(out, delivery.Event{Cursor: x.Cursor, Data: b})
	}
	return out, nil
}
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	a, _, ok := s.account(w, r)
	if !ok {
		return
	}
	after, e := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	if e != nil {
		s.fail(w, 400, "INVALID_CURSOR")
		return
	}
	v, e := s.catchup(r.Context(), r.PathValue("match"), a.ID, after)
	if e != nil {
		s.err(w, e)
		return
	}
	s.write(w, 200, v)
}
func (s *Server) stream(w http.ResponseWriter, r *http.Request) {
	for k, v := range r.URL.Query() {
		if k != "after" || len(v) != 1 {
			s.fail(w, 400, "INVALID_REQUEST")
			return
		}
	}
	a, token, ok := s.account(w, r)
	if !ok {
		return
	}
	id := r.PathValue("match")
	after, e := strconv.ParseInt(r.URL.Query().Get("after"), 10, 64)
	if e != nil {
		s.fail(w, 400, "INVALID_CURSOR")
		return
	}
	if _, e = s.catchup(r.Context(), id, a.ID, after); e != nil {
		s.err(w, e)
		return
	}
	authorize := func(ctx context.Context) error {
		_, e := s.identity.Authenticate(ctx, token)
		if e == nil {
			_, e = s.matches.Snapshot(ctx, id, a.ID)
		}
		return e
	}
	if release, ok := r.Context().Value(admissionKey{}).(func()); ok {
		release()
	}
	_ = s.streams.Serve(w, r, s.ctx, a.ID, after, authorize, func(ctx context.Context, c int64) ([]delivery.Event, error) {
		s.presence.Touch(ctx, id+":"+a.ID)
		v, e := s.catchup(ctx, id, a.ID, c)
		if errors.Is(e, matchstore.ErrCursor) {
			e = delivery.ErrCursor
		}
		return v, e
	})
}
func (s *Server) Close(ctx context.Context) error {
	s.lifecycleMu.Lock()
	s.closed = true
	s.cancel()
	s.closeOnce.Do(func() { go func() { s.wg.Wait(); close(s.done) }() })
	s.lifecycleMu.Unlock()
	e := s.streams.Close(ctx)
	d := s.dispatcher.Close(ctx)
	select {
	case <-s.done:
	case <-ctx.Done():
		return ctx.Err()
	}
	if e != nil {
		return e
	}
	return d
}
