//go:build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBrowserFixtureRealAuthority(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		t.Run(fmt.Sprint("economy-", enabled), func(t *testing.T) { testBrowserFixture(t, enabled) })
	}
}

func testBrowserFixture(t *testing.T, enabled bool) {
	if os.Getenv("CGMS_TEST_DATABASE_URL") == "" {
		t.Fatal("CGMS_TEST_DATABASE_URL required")
	}
	assets := t.TempDir()
	if e := os.WriteFile(filepath.Join(assets, "index.html"), []byte("fixture UI"), 0600); e != nil {
		t.Fatal(e)
	}
	reservation, e := net.Listen("tcp", "127.0.0.1:0")
	if e != nil {
		t.Fatal(e)
	}
	address := reservation.Addr().String()
	reservation.Close()
	origin := "http://" + address
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- runWithEconomy(ctx, address, assets, enabled) }()
	defer func() {
		cancel()
		select {
		case e := <-done:
			if e != nil {
				t.Error(e)
			}
		case <-time.After(15 * time.Second):
			t.Error("fixture shutdown timed out")
		}
	}()
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(20 * time.Second)
	for {
		r, e := client.Get(origin)
		if e == nil {
			r.Body.Close()
			if r.StatusCode == 200 {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("fixture readiness timed out")
		}
		time.Sleep(50 * time.Millisecond)
	}
	// Every fixture uses the same loopback client address. Respect the real
	// authority's 5 requests/second admission policy as the scenario set grows.
	pace := time.NewTicker(250 * time.Millisecond)
	defer pace.Stop()
	for _, n := range []int{3, 4} {
		for _, name := range fixtureNames {
			confirmedAfter := time.Now().UTC().Truncate(time.Microsecond)
			body := fmt.Sprintf(`{"scenario":%q,"players":%d,"seat":1}`, name, n)
			req, _ := http.NewRequest("POST", origin+"/__fixture/session", strings.NewReader(body))
			req.Header.Set("Origin", origin)
			req.Header.Set("Content-Type", "application/json")
			response, e := client.Do(req)
			if e != nil {
				t.Fatal(e)
			}
			var selected map[string]string
			e = json.NewDecoder(response.Body).Decode(&selected)
			response.Body.Close()
			if e != nil || response.StatusCode != 200 || len(response.Cookies()) != 1 {
				t.Fatal("fixture session failed", name, n, e)
			}
			confirmedBefore := time.Now().UTC()
			cookie := response.Cookies()[0]
			req, _ = http.NewRequest("GET", origin+"/v1/matches/"+selected["match_id"], nil)
			req.AddCookie(cookie)
			<-pace.C
			response, e = client.Do(req)
			if e != nil {
				t.Fatal(e)
			}
			var snap struct {
				Projection struct {
					Board struct {
						Seat   int    `json:"seat"`
						GameID string `json:"game_id"`
					} `json:"board"`
				} `json:"projection"`
			}
			e = json.NewDecoder(response.Body).Decode(&snap)
			response.Body.Close()
			if e != nil || response.StatusCode != 200 || snap.Projection.Board.Seat != 1 || snap.Projection.Board.GameID == "" {
				t.Fatal("real authorized snapshot failed", name, n, response.StatusCode, e)
			}
			if enabled {
				req, _ = http.NewRequest("GET", origin+"/v1/economy", nil)
				req.AddCookie(cookie)
				<-pace.C
				response, e = client.Do(req)
				if e != nil {
					t.Fatal(e)
				}
				var account economy.AccountView
				e = json.NewDecoder(response.Body).Decode(&account)
				response.Body.Close()
				wantFree := 2
				if strings.HasPrefix(name, "paid-") {
					wantFree = 3
				}
				if e != nil || response.StatusCode != 200 || !account.Enabled || account.Benefits.Dirt != 30 || account.FreeStartsRemaining != wantFree {
					t.Fatal("synthetic completed rewards fixture", name, n, account, e)
				}
			}
			if name == "ad-free" || strings.HasPrefix(name, "paid-") {
				testPaidAccountBoundary(t, client, origin, selected["match_id"], name, n, enabled, confirmedAfter, confirmedBefore, pace.C)
			}
		}
	}
}

func testPaidAccountBoundary(t *testing.T, client *http.Client, origin, matchID, scenario string, players int, enabled bool, confirmedAfter, confirmedBefore time.Time, pace <-chan time.Time) {
	t.Helper()
	read := func(path string, cookie *http.Cookie, want int) []byte {
		t.Helper()
		req, err := http.NewRequest("GET", origin+path, nil)
		if err != nil {
			t.Fatal(err)
		}
		if cookie != nil {
			req.AddCookie(cookie)
		}
		<-pace
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil || response.StatusCode != want {
			t.Fatalf("Paid fixture GET %s status = %d, want %d; read error = %v", path, response.StatusCode, want, err)
		}
		return body
	}
	selectSeat := func(seat int) *http.Cookie {
		t.Helper()
		req, err := http.NewRequest("POST", origin+"/__fixture/session", strings.NewReader(fmt.Sprintf(`{"scenario":%q,"players":%d,"seat":%d}`, scenario, players, seat)))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", "application/json")
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		var selected map[string]string
		err = json.NewDecoder(response.Body).Decode(&selected)
		response.Body.Close()
		if err != nil || response.StatusCode != http.StatusOK || len(response.Cookies()) != 1 || len(selected) != 1 || selected["match_id"] != matchID {
			t.Fatal("Paid seat reselection changed match identity or disclosed extra data")
		}
		return response.Cookies()[0]
	}
	read("/v1/economy", nil, http.StatusUnauthorized)
	for seat := 1; seat <= players; seat++ {
		cookie := selectSeat(seat)
		before := read("/v1/matches/"+matchID, cookie, http.StatusOK)
		for _, privateKey := range []string{`"benefits"`, `"ad_free_until"`, `"transaction"`, `"confirmed_at"`, `"ConfirmedAt"`, `"qa-` + scenario + `"`} {
			if strings.Contains(string(before), privateKey) {
				t.Fatal("game projection disclosed private entitlement data")
			}
		}
		var snap struct {
			Projection struct {
				Board struct {
					Seat   int    `json:"seat"`
					GameID string `json:"game_id"`
				} `json:"board"`
			} `json:"projection"`
		}
		if err := json.Unmarshal(before, &snap); err != nil || snap.Projection.Board.Seat != seat || snap.Projection.Board.GameID != matchID+"-game" {
			t.Fatal("Paid viewer seat or game identity changed")
		}
		body := read("/v1/economy", cookie, http.StatusOK)
		var account economy.AccountView
		if err := json.Unmarshal(body, &account); err != nil {
			t.Fatal(err)
		}
		if enabled {
			wantUnlimited := seat == 1 && scenario != "ad-free"
			wantFree := 2
			if wantUnlimited {
				wantFree = 3
			}
			if !account.Enabled || account.Benefits.Dirt != 30 || account.FreeStartsRemaining != wantFree || account.Benefits.Unlimited != wantUnlimited || account.Benefits.NoAds != (seat == 1) {
				t.Fatal("Paid fixture changed another seat, earned balance, or start allowance", scenario, seat)
			}
			if seat == 1 {
				duration := 48 * time.Hour // Other tiers use a synthetic provider QA expiry.
				if scenario == "ad-free" {
					duration = 30 * 24 * time.Hour
				} else if scenario == "paid-daily" {
					duration = 24 * time.Hour
				}
				if account.Benefits.AdFreeUntil.Before(confirmedAfter.Add(duration)) || account.Benefits.AdFreeUntil.After(confirmedBefore.Add(duration)) {
					t.Fatal("Paid fixture expiry is not anchored to initial confirmed purchase", scenario)
				}
				if wantUnlimited && !account.Benefits.PremiumUntil.Equal(account.Benefits.AdFreeUntil) {
					t.Fatal("Paid unlimited and ad-free expiry diverged")
				}
				if !wantUnlimited && account.Benefits.PremiumUntil.After(time.Now()) {
					t.Fatal("Standalone Ad-free altered unlimited-play expiry")
				}
			} else if account.Benefits.AdFreeUntil.After(time.Now()) {
				t.Fatal("another seat received the Ad-free grant")
			}
		} else {
			var disabled map[string]bool
			if err := json.Unmarshal(body, &disabled); err != nil || len(disabled) != 1 || disabled["enabled"] {
				t.Fatal("disabled economy exposed synthetic grant data")
			}
		}
		// Re-selecting the same synthetic identity models a reload/reconnect.
		// The receipt and its absolute expiry are not minted again.
		reloaded := read("/v1/economy", selectSeat(seat), http.StatusOK)
		if string(reloaded) != string(body) {
			t.Fatal("reselection changed the account or extended the QA grant")
		}
		after := read("/v1/matches/"+matchID, cookie, http.StatusOK)
		if string(after) != string(before) {
			t.Fatal("account inspection changed the authoritative game snapshot")
		}
	}
}
