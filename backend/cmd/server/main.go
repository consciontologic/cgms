// Command server runs the single-process CGMS online authority. Deployment,
// TLS termination and allowance policy are separate release gates.
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/config"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"github.com/metaphy6/cgms/backend/internal/rooms"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
	"github.com/metaphy6/cgms/backend/internal/transport"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) == 3 && os.Args[1] == "check-config" {
		if _, e := config.Load(os.Args[2]); e != nil {
			fmt.Fprintln(os.Stderr, "invalid configuration")
			os.Exit(1)
		}
		return
	}
	if e := run(); e != nil {
		slog.Error("server stopped", "reason", e.Error())
		os.Exit(1)
	}
}
func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var settings *config.Config
	if path := os.Getenv("CGMS_CONFIG"); path != "" {
		c, err := config.Load(path)
		if err != nil {
			return err
		}
		settings = &c
	}
	recorder := telemetry.New(os.Stdout, settings != nil && settings.InsecureLocal)
	defer func() {
		end, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = recorder.Close(end)
	}()
	if settings != nil && settings.Observability.Enabled {
		if err := recorder.Enable(ctx, settings.Observability.Endpoint); err != nil {
			return err
		}
	}
	file := os.Getenv("CGMS_RULES_FILE")
	if settings != nil {
		file = settings.RulesFile
	}
	if file == "" {
		file = "config/rules/game-rules.json"
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		return errors.New("rules file unavailable")
	}
	if _, e = game.LoadRules(raw, nil); e != nil {
		return e
	}
	dsn := os.Getenv("CGMS_DATABASE_URL")
	origin, redisURL, insecure := os.Getenv("CGMS_ORIGIN"), os.Getenv("CGMS_REDIS_URL"), os.Getenv("CGMS_INSECURE_LOCAL") == "1"
	if settings != nil {
		secret, err := settings.Secrets()
		if err != nil {
			return err
		}
		dsn = secret.DatabaseURL
		redisURL = secret.RedisURL
		origin = settings.Origin
		insecure = settings.InsecureLocal
	}
	if dsn == "" {
		return errors.New("CGMS_DATABASE_URL required")
	}
	var economyPolicy *economy.Policy
	economyEnabled := os.Getenv("CGMS_ECONOMY_ENABLED")
	if settings == nil && economyEnabled != "" && economyEnabled != "0" && economyEnabled != "1" {
		return errors.New("invalid economy activation")
	}
	if (settings != nil && settings.EconomyEnabled) || (settings == nil && economyEnabled == "1") {
		policy := economy.AdoptedPolicy()
		economyPolicy = &policy
	}
	cfg, e := pgxpool.ParseConfig(dsn)
	if e != nil {
		return errors.New("invalid database configuration")
	}
	cfg.MaxConns = 16
	cfg.ConnConfig.ConnectTimeout = 2 * time.Second
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		return errors.New("database unavailable")
	}
	defer pool.Close()
	boot, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	lease, e := transport.AcquireAuthority(boot, pool)
	if e != nil {
		return errors.New("cannot acquire single-process authority")
	}
	defer lease.Close()
	for _, up := range []func(context.Context, *pgxpool.Pool) error{matchstore.Up, identity.Up, rooms.Up, economy.Up} {
		if e = up(boot, pool); e != nil {
			return errors.New("migration failed")
		}
	}
	key, e := transport.PersistentKey(boot, pool)
	if e != nil {
		return errors.New("handle key unavailable")
	}
	maxStreams := 64
	if settings != nil {
		maxStreams = settings.MaxStreams
	}
	proxyClientIP := settings != nil && settings.ProxyClientIP
	s, e := transport.New(transport.Config{EconomyPolicy: economyPolicy, MaxStreams: maxStreams, ProxyClientIP: proxyClientIP, Pool: pool, RulesHash: game.AcceptedRulesHash, Origin: origin, RedisURL: redisURL, HandleKey: key, InsecureLocal: insecure, Telemetry: recorder})
	if e != nil {
		return errors.New("invalid online configuration")
	}
	s.Start()
	address := os.Getenv("CGMS_LISTEN")
	if settings != nil {
		address = settings.Listen
	}
	if address == "" {
		address = "127.0.0.1:8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/live", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	mux.HandleFunc("GET /health/ready", func(w http.ResponseWriter, r *http.Request) {
		ping, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if pool.Ping(ping) != nil {
			http.Error(w, "not ready", 503)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /internal/metrics", recorder.Metrics)
	mux.Handle("/", s)
	server := &http.Server{Addr: address, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServe() }()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var cause error
running:
	for {
		select {
		case <-ctx.Done():
			break running
		case e = <-done:
			if !errors.Is(e, http.ErrServerClosed) {
				cause = fmt.Errorf("HTTP listener unavailable")
			}
			break running
		case <-ticker.C:
			ping, pc := context.WithTimeout(ctx, time.Second)
			e = lease.Ping(ping)
			pc()
			if e != nil {
				cause = errors.New("authority connection lost")
				break running
			}
		}
	}
	shutdown, sc := context.WithTimeout(context.Background(), 10*time.Second)
	defer sc()
	if e = s.Close(shutdown); e != nil && cause == nil {
		cause = errors.New("online shutdown incomplete")
	}
	if e = server.Shutdown(shutdown); e != nil {
		_ = server.Close()
		if cause == nil {
			cause = errors.New("HTTP shutdown incomplete")
		}
	}
	return cause
}
