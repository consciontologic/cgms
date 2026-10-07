//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/metaphy6/cgms/backend/internal/identity"
)

// The child executes the actual production startup/shutdown function. No database
// address, credential, token, key, or response body is printed on failures.
func TestServerHelperProcess(t *testing.T) {
	if os.Getenv("CGMS_SERVER_SMOKE_CHILD") != "1" {
		return
	}
	if err := run(); err != nil {
		if os.Getenv("CGMS_SERVER_EXPECT_DUPLICATE") == "1" && err.Error() == "cannot acquire single-process authority" {
			os.Exit(11)
		}
		os.Exit(12)
	}
	os.Exit(0)
}

func TestServerTCPShutdownRestart(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	dsn := os.Getenv("CGMS_TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("disposable PostgreSQL required")
	}
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal("database setup")
	}
	defer admin.Close()
	schema := fmt.Sprintf("server_smoke_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal("schema setup")
	}
	defer func() {
		cleanup, c := context.WithTimeout(context.Background(), 3*time.Second)
		defer c()
		if _, e := admin.Exec(cleanup, "DROP SCHEMA "+schema+" CASCADE"); e != nil {
			t.Error("schema cleanup")
		}
	}()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal("database config")
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	childDSN := dsn
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		u, e := url.Parse(dsn)
		if e != nil {
			t.Fatal("database URL")
		}
		q := u.Query()
		q.Set("search_path", schema)
		u.RawQuery = q.Encode()
		childDSN = u.String()
	} else {
		childDSN += " search_path=" + schema
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal("scoped pool")
	}
	defer pool.Close()
	rules, err := filepath.Abs("../../../config/rules/game-rules.json")
	if err != nil {
		t.Fatal("rules path")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal("port allocation")
	}
	address := listener.Addr().String()
	if err = listener.Close(); err != nil {
		t.Fatal("port release")
	}
	origin := "http://" + address
	var cleanups []func()
	defer func() {
		for i := len(cleanups) - 1; i >= 0; i-- {
			cleanups[i]()
		}
	}()
	start := func(duplicate bool) (*exec.Cmd, <-chan error) {
		t.Helper()
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestServerHelperProcess$")
		cmd.Env = append(os.Environ(), "CGMS_SERVER_SMOKE_CHILD=1", "CGMS_DATABASE_URL="+childDSN, "CGMS_ORIGIN="+origin, "CGMS_LISTEN="+address, "CGMS_RULES_FILE="+rules, "CGMS_REDIS_URL=", "CGMS_INSECURE_LOCAL=1", "CGMS_SERVER_EXPECT_DUPLICATE="+map[bool]string{true: "1", false: "0"}[duplicate])
		if err := cmd.Start(); err != nil {
			t.Fatal("child start")
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait(); close(done) }()
		cleanups = append(cleanups, func() {
			select {
			case <-done:
				return
			default:
				_ = cmd.Process.Kill()
				<-done
			}
		})
		return cmd, done
	}
	client := &http.Client{Timeout: time.Second}
	ready := func(done <-chan error) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			select {
			case <-done:
				t.Fatal("server exited before ready")
			default:
			}
			response, e := client.Get(origin + "/v1/session")
			if e == nil {
				_ = response.Body.Close()
				if response.StatusCode == http.StatusUnauthorized {
					return
				}
			}
			select {
			case <-ctx.Done():
				t.Fatal("startup deadline")
			case <-time.After(20 * time.Millisecond):
			}
		}
		t.Fatal("server did not become ready")
	}
	stop := func(cmd *exec.Cmd, done <-chan error) {
		t.Helper()
		if e := cmd.Process.Signal(syscall.SIGTERM); e != nil {
			t.Fatal("signal shutdown")
		}
		select {
		case e := <-done:
			if e != nil {
				t.Fatal("server did not exit gracefully")
			}
		case <-time.After(12 * time.Second):
			t.Fatal("shutdown deadline")
		}
	}
	first, done := start(false)
	ready(done)
	request, err := http.NewRequest("POST", origin+"/v1/guests", bytes.NewBufferString(`{}`))
	if err != nil {
		t.Fatal("guest request")
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		t.Fatal("guest HTTP")
	}
	var guest identity.Session
	err = json.NewDecoder(io.LimitReader(response.Body, 8192)).Decode(&guest)
	_ = response.Body.Close()
	if err != nil || response.StatusCode != 200 || guest.Token == "" || guest.Account.ID == "" {
		t.Fatal("guest creation")
	}
	var keyBefore, keyAfter []byte
	if err = pool.QueryRow(ctx, "SELECT value FROM online_settings WHERE name='handle-key-v1'").Scan(&keyBefore); err != nil {
		t.Fatal("initial persistent key")
	}
	_, duplicateDone := start(true)
	select {
	case e := <-duplicateDone:
		exit, ok := e.(*exec.ExitError)
		if !ok || exit.ExitCode() != 11 {
			t.Fatal("duplicate authority was not rejected")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("duplicate startup deadline")
	}
	stop(first, done)
	second, secondDone := start(false)
	ready(secondDone)
	request, err = http.NewRequest("GET", origin+"/v1/session", nil)
	if err != nil {
		t.Fatal("session request")
	}
	request.Header.Set("Authorization", "Bearer "+guest.Token)
	response, err = client.Do(request)
	if err != nil {
		t.Fatal("restored session HTTP")
	}
	var session struct {
		Account identity.Account `json:"account"`
	}
	err = json.NewDecoder(io.LimitReader(response.Body, 8192)).Decode(&session)
	_ = response.Body.Close()
	if err != nil || response.StatusCode != 200 || session.Account.ID != guest.Account.ID {
		t.Fatal("identity/session not preserved")
	}
	if err = pool.QueryRow(ctx, "SELECT value FROM online_settings WHERE name='handle-key-v1'").Scan(&keyAfter); err != nil || len(keyAfter) != 32 || !bytes.Equal(keyBefore, keyAfter) {
		t.Fatal("handle key not preserved")
	}
	stop(second, secondDone)
}
