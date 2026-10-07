package main

import (
	"context"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/economy"
	"github.com/metaphy6/cgms/backend/internal/game"
	"github.com/metaphy6/cgms/backend/internal/identity"
	"github.com/metaphy6/cgms/backend/internal/matchstore"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestFixturesAreValidAndPopulationSpecific(t *testing.T) {
	for _, n := range []int{3, 4} {
		for _, name := range fixtureNames {
			m, err := fixture(name, n)
			if err != nil {
				t.Fatal(name, n, err)
			}
			if err = m.Game.Validate(); err != nil {
				t.Fatal(name, n, err)
			}
			if len(m.Game.Board.Players) != n {
				t.Fatal("population")
			}
			if name == "trade" && !m.Game.TurnStarted {
				t.Fatal("not at human turn")
			}
			if name == "settlement" && m.Game.Ledger.Phase != "settlement" {
				t.Fatal("not settlement")
			}
			if name == "final-settlement" && (m.GameLimit != 1 || m.Game.Ledger.Phase != "settlement") {
				t.Fatal("final standings require a one-game settlement fixture")
			}
			if name == "loan" && (!m.Game.Board.Players[0].Underground || len(m.Game.Board.Formations) != 1) {
				t.Fatal("loan requires registered Underground")
			}
			if name == "attack" && m.Game.Board.Pending != nil {
				t.Fatal("attack declaration must remain a human action")
			}
			if name == "justice" && (m.Game.Board.Pending == nil || m.Game.Board.Pending.Decision == nil || m.Game.Board.Pending.Decision.Kind != "justice") {
				t.Fatal("Justice requires staged selection")
			}
			if name == "confinement" {
				ace, ok := m.Game.Board.Card("deck-1-diamonds-01")
				if !ok || ace.Controller != 2 || ace.Zone != "concealed-ace" || m.Game.Board.Pending == nil || m.Game.Board.Pending.Responders[0] != 2 {
					t.Fatal("Confinement requires responder's concealed Diamond Ace")
				}
			}
			env, err := matchstore.NewEnvelope(m, fixtureRules)
			if err != nil {
				t.Fatal(err)
			}
			if _, automatic := matchstore.NextServer(env, 0); automatic {
				t.Fatal("fixture must begin at human boundary", name)
			}
		}
	}
	for _, n := range []int{0, 2, 5} {
		if _, err := fixture("trade", n); err == nil {
			t.Fatal("invalid population")
		}
	}
	if _, err := fixture("unknown", 3); err == nil {
		t.Fatal("unknown fixture")
	}
}

func TestTurnDrawFixturesUseScheduledAuthorityAndConcealedDestination(t *testing.T) {
	for _, tc := range []struct {
		name, card string
		zone       game.Zone
	}{
		{"turn-draw-ace", "deck-1-hearts-01", game.ConcealedAce},
		{"turn-draw-diamond", "deck-1-diamonds-04", game.Hand},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := fixture(tc.name, 4)
			if err != nil {
				t.Fatal(err)
			}
			m.Game, err = m.Game.EndTurn()
			if err != nil {
				t.Fatal(err)
			}
			if m.Game.Board.Active != 2 || m.Game.TurnStarted {
				t.Fatal("expected next scheduled turn before draw")
			}
			var drawn []string
			m.Game, drawn, err = m.Game.BeginTurn(nil)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(drawn, []string{tc.card}) {
				t.Fatalf("draw = %v", drawn)
			}
			card, ok := m.Game.Board.Card(tc.card)
			if !ok || card.Controller != 2 || card.Zone != tc.zone {
				t.Fatal("wrong acquisition destination")
			}
			if _, _, err = m.Game.BeginTurn(nil); err == nil {
				t.Fatal("scheduled draw repeated")
			}
			for viewer := 1; viewer <= 4; viewer++ {
				view, err := game.ObserveWithoutMenu(m.Game.Board, viewer)
				if err != nil {
					t.Fatal(err)
				}
				for _, c := range view.Cards {
					if c.Card.ID == tc.card && viewer != 2 {
						t.Fatal("opponent saw the drawn face")
					}
				}
			}
		})
	}
}

func TestLargeOpeningFixturePreservesEveryPhysicalNumberCopy(t *testing.T) {
	for _, players := range []int{3, 4} {
		m, err := fixture("opening-large", players)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, c := range m.Game.Board.Cards {
			if c.Controller != 1 || c.Zone != game.Hand {
				continue
			}
			if c.Card.Suit != game.Clubs || c.Card.Rank < 2 || c.Card.Rank > 10 || seen[c.Card.ID] {
				t.Fatal("large opening contains a wrong or duplicate physical card")
			}
			seen[c.Card.ID] = true
		}
		if len(seen) != 18 {
			t.Fatalf("opening has %d cards, want 18", len(seen))
		}
	}
}

func TestPaidFixturesPreserveOpeningGameAndPrivateCards(t *testing.T) {
	for _, players := range []int{3, 4} {
		for _, scenario := range []string{"ad-free", "paid-daily", "paid-weekly", "paid-monthly", "paid-yearly"} {
			id := fmt.Sprintf("same-fixture-%d", players)
			opening, err := fixtureWithID("opening", players, id)
			if err != nil {
				t.Fatal(err)
			}
			adFree, err := fixtureWithID(scenario, players, id)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(opening, adFree) {
				t.Fatal("Paid account fixture must not alter game identity, cards, rules, or turn state")
			}
			for viewer := 1; viewer <= players; viewer++ {
				view, err := game.ObserveWithoutMenu(adFree.Game.Board, viewer)
				if err != nil {
					t.Fatal(err)
				}
				for _, card := range view.Cards {
					if (card.Zone == game.Hand || card.Zone == game.ConcealedAce) && card.Controller != viewer {
						t.Fatal("Paid fixture exposed an opponent's hidden card")
					}
				}
			}
		}
	}
}

func TestPaidFixtureSessionRejectsEntitlementPayloads(t *testing.T) {
	for _, players := range []int{3, 4} {
		for _, scenario := range []string{"ad-free", "paid-daily", "paid-weekly", "paid-monthly", "paid-yearly"} {
			id := fmt.Sprintf("fixture-%s-%d", scenario, players)
			h := &harness{origin: "http://127.0.0.1:8081", sessions: map[string][]identity.Session{id: {{Token: "private-test-token", ExpiresAt: time.Now().Add(time.Hour)}}}}
			for _, extra := range []string{"", `,"product":"ad_free"`, `,"expires_at":"2099-01-01T00:00:00Z"`, `,"confirmed_at":"2026-10-01T00:00:00Z"`, `,"account":"other"`} {
				r := httptest.NewRequest("POST", h.origin+"/__fixture/session", strings.NewReader(fmt.Sprintf(`{"scenario":%q,"players":%d,"seat":1%s}`, scenario, players, extra)))
				r.Header.Set("Origin", h.origin)
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				want := http.StatusBadRequest
				if extra == "" {
					want = http.StatusOK
				}
				if w.Code != want {
					t.Fatalf("fixture status = %d, want %d", w.Code, want)
				}
				if strings.Contains(w.Body.String(), "private-test-token") || strings.Contains(w.Body.String(), "expires_at") {
					t.Fatal("fixture response disclosed credentials or entitlement data")
				}
			}
		}
	}
}

func TestPaidFixturePurchaseUsesConfirmedTermsAndAccountBinding(t *testing.T) {
	confirmed := time.Date(2026, 10, 1, 15, 0, 0, 123456789, time.FixedZone("fixture", 3*60*60))
	for _, tc := range []struct {
		scenario, kind string
		duration       time.Duration
	}{
		{"ad-free", "ad_free", 30 * 24 * time.Hour},
		{"paid-daily", "daily", 24 * time.Hour},
		// These are synthetic provider expiry windows, not offer durations.
		{"paid-weekly", "weekly", 48 * time.Hour},
		{"paid-monthly", "monthly", 48 * time.Hour},
		{"paid-yearly", "yearly", 48 * time.Hour},
	} {
		contract, purchase, ok := fixturePurchase("test-match", "seat-one", tc.scenario, confirmed)
		if !ok || purchase.Account != "seat-one" || contract.Products[purchase.Product].Kind != tc.kind {
			t.Fatal("fixture product/account binding", tc.scenario)
		}
		anchor := confirmed.UTC().Truncate(time.Microsecond)
		if !purchase.ConfirmedAt.Equal(anchor) || !purchase.ExpiresAt.Equal(anchor.Add(tc.duration)) {
			t.Fatal("fixture expiry must be anchored once to confirmed purchase time", tc.scenario)
		}
		grant, err := economy.ReconcilePurchase(contract, purchase.Account, nil, purchase)
		if err != nil || !grant.Active(anchor) || grant.Active(purchase.ExpiresAt) {
			t.Fatal("fixture must use the trusted real product contract", tc.scenario, err)
		}
		if _, err = economy.ReconcilePurchase(contract, "other-seat", nil, purchase); err == nil {
			t.Fatal("fixture grant accepted for another account")
		}
	}
	if _, _, ok := fixturePurchase("test-match", "seat-one", "opening", confirmed); ok {
		t.Fatal("ordinary fixture minted a paid grant")
	}
}

func TestCompensationFixtureResolvesInSeatOrder(t *testing.T) {
	for _, n := range []int{3, 4} {
		m, e := fixture("compensation", n)
		if e != nil {
			t.Fatal(e)
		}
		for seat := 2; seat <= n; seat++ {
			m.Game, _, e = m.Game.ApplyBoardCommand(game.Command{GameID: m.Game.Board.GameID, ID: fmt.Sprintf("pass-%d", seat), Kind: "pass", Actor: seat, WindowID: m.Game.Board.Pending.ID})
			if e != nil {
				t.Fatal(e)
			}
		}
		m.Game, e = m.Game.ContinueAfterAction(m.Game.Board.DrawOrder, nil)
		if e != nil {
			t.Fatal(e)
		}
		if len(m.Game.Board.DrawQueue) != 1 || m.Game.Board.DrawQueue[0].Seat != 3 {
			t.Fatal("first recipient must finish before third seat", m.Game.Board.DrawQueue)
		}
		m.Game, e = m.Game.ContinueAfterAction(nil, nil)
		if e != nil || len(m.Game.Board.DrawQueue) != 0 {
			t.Fatal("second recipient did not finish", e)
		}
		for _, eff := range m.Game.Board.Effects {
			if eff.Losses != 0 {
				t.Fatal("entitlement not consumed", eff)
			}
		}
	}
}

func TestCardStyleFixtureSupportsPublicDiamondCaptureAndHiddenCards(t *testing.T) {
	for _, players := range []int{3, 4} {
		m, err := fixture("card-style", players)
		if err != nil {
			t.Fatal(err)
		}
		if m.Game.Board.Pending != nil || !m.Game.TurnStarted {
			t.Fatal("capture must begin at a human action boundary")
		}
		for viewer := 1; viewer <= players; viewer++ {
			view, err := game.ObserveWithoutMenu(m.Game.Board, viewer)
			if err != nil {
				t.Fatal(err)
			}
			ownHand, ownAce := 0, 0
			for _, card := range view.Cards {
				if card.Zone == game.Hand || card.Zone == game.ConcealedAce {
					if card.Controller != viewer {
						t.Fatal("fixture projection disclosed another seat's hidden identity")
					}
					if card.Zone == game.Hand {
						ownHand++
					} else {
						ownAce++
					}
				}
			}
			if ownHand != 1 || ownAce != 1 {
				t.Fatal("each viewer needs authorized own hand and Ace examples")
			}
		}
		m.Game, _, err = m.Game.ApplyBoardCommand(game.Command{GameID: m.Game.Board.GameID, ID: "capture", Kind: "attack", Actor: 1, Cards: []string{"deck-1-clubs-08"}, Targets: []string{"deck-1-diamonds-08"}})
		if err != nil {
			t.Fatal(err)
		}
		for seat := 2; seat <= players; seat++ {
			m.Game, _, err = m.Game.ApplyBoardCommand(game.Command{GameID: m.Game.Board.GameID, ID: fmt.Sprintf("style-pass-%d", seat), Kind: "pass", Actor: seat, WindowID: m.Game.Board.Pending.ID})
			if err != nil {
				t.Fatal(err)
			}
		}
		card, _ := m.Game.Board.Card("deck-1-diamonds-08")
		if card.Controller != 1 || card.Zone != game.Series || card.Card.ID != "deck-1-diamonds-08" {
			t.Fatal("legal capture must preserve physical identity and expose Diamond for new controller")
		}
		if err = m.Game.Validate(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFixtureRunGroupsAreBoundedAndReused(t *testing.T) {
	calls := 0
	h := &harness{origin: "http://127.0.0.1:8081", sessions: map[string][]identity.Session{}}
	h.create = func(_ context.Context, id, name string, n int) ([]identity.Session, error) {
		calls++
		return []identity.Session{{Token: "one"}, {Token: "two"}, {Token: "three"}}, nil
	}
	for _, run := range []string{"chromium", "chromium", "firefox"} {
		req := httptest.NewRequest("POST", h.origin+"/__fixture/session", strings.NewReader(`{"scenario":"trade","players":3,"seat":1,"run":"`+run+`"}`))
		req.Header.Set("Origin", h.origin)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != 200 || !strings.Contains(w.Body.String(), "fixture-trade-3-"+run) {
			t.Fatal("run isolation", w.Code, w.Body.String())
		}
	}
	if calls != 2 {
		t.Fatal("group recreated", calls)
	}
	for i := len(h.sessions); i < 128; i++ {
		h.sessions[fmt.Sprintf("occupied-%d", i)] = []identity.Session{{Token: "fixture"}}
	}
	req := httptest.NewRequest("POST", h.origin+"/__fixture/session", strings.NewReader(`{"scenario":"trade","players":3,"seat":1,"run":"overflow"}`))
	req.Header.Set("Origin", h.origin)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 400 || calls != 2 || len(h.sessions) != 128 {
		t.Fatal("fixture group cap bypassed", w.Code, calls, len(h.sessions))
	}
}

func TestFixtureSessionIsSameOriginAndCookieOnly(t *testing.T) {
	h := &harness{origin: "http://127.0.0.1:8081", sessions: map[string][]identity.Session{"fixture-trade-3": {{Token: "private-test-token", ExpiresAt: time.Now().Add(time.Hour)}}}}
	for _, tc := range []struct {
		origin, host, body string
		want               int
	}{
		{h.origin, "127.0.0.1:8081", `{"scenario":"trade","players":3,"seat":1}`, 200},
		{"https://evil.test", "127.0.0.1:8081", `{"scenario":"trade","players":3,"seat":1}`, 403},
		{h.origin, "evil.test", `{"scenario":"trade","players":3,"seat":1}`, 403},
		{h.origin, "127.0.0.1:8081", `{"scenario":"trade","players":3,"seat":2}`, 400},
		{h.origin, "127.0.0.1:8081", `{"scenario":"trade","players":3,"seat":1,"extra":true}`, 400},
	} {
		r := httptest.NewRequest("POST", h.origin+"/__fixture/session", strings.NewReader(tc.body))
		r.Host = tc.host
		r.Header.Set("Origin", tc.origin)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatal(w.Code, tc.want)
		}
		if strings.Contains(w.Body.String(), "private-test-token") {
			t.Fatal("token exposed")
		}
		if tc.want == 200 {
			c := w.Result().Cookies()
			if len(c) != 1 || !c[0].HttpOnly || c[0].SameSite != http.SameSiteStrictMode {
				t.Fatal("unsafe cookie")
			}
		}
	}
}

func TestLoopbackBindIsMandatory(t *testing.T) {
	for _, s := range []string{"0.0.0.0:8081", ":8081", "example.com:8081", "localhost:8081"} {
		if validListen(s) {
			t.Fatalf("accepted %s", s)
		}
	}
	for _, s := range []string{"127.0.0.1:8081", "[::1]:8081"} {
		if !validListen(s) {
			t.Fatal(s)
		}
	}
}
