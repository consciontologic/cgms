package transport

import (
	"bytes"
	"fmt"
	"github.com/metaphy6/cgms/backend/internal/telemetry"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOriginAndBoundedRequestAdmission(t *testing.T) {
	s := &Server{cfg: Config{Origin: "https://game.example"}}
	r := httptest.NewRequest(http.MethodPost, "/v1/guests", strings.NewReader("{}"))
	r.Header.Set("Origin", "https://evil.example")
	if s.allowedOrigin(r) {
		t.Fatal("foreign origin allowed")
	}
	r.Header.Set("Origin", "https://game.example")
	if !s.allowedOrigin(r) {
		t.Fatal("own origin rejected")
	}
}
func TestStrictJSON(t *testing.T) {
	for _, body := range []string{`{"value":1,"value":2}`, `{"Value":1}`, `{"value":1,"extra":1}`, `{"value":1} {}`, `null`} {
		r := httptest.NewRequest("POST", "/", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		var v struct {
			Value int `json:"value"`
		}
		if decode(r, &v) == nil {
			t.Fatalf("accepted malformed %s", body)
		}
	}
}
func TestStreamPreflightAdmissionIsBounded(t *testing.T) {
	s := &Server{cfg: Config{Origin: "https://game.example"}, rates: map[string]rate{}, active: make(chan struct{}, 1), mux: http.NewServeMux()}
	s.active <- struct{}{}
	r := httptest.NewRequest("GET", "/v1/matches/m/stream?after=0", nil)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 429 {
		t.Fatal(w.Code)
	}
}

func TestTelemetryRequestFailureAndPrivacy(t *testing.T) {
	var log bytes.Buffer
	recorder := telemetry.New(&log, false)
	s := &Server{cfg: Config{Origin: "https://game.example", Telemetry: recorder}}
	r := httptest.NewRequest("GET", "/private-card?seed=PRIVATE", nil)
	r.Header.Set("Origin", "https://other.example")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 403 || !strings.Contains(log.String(), `"outcome":"error"`) {
		t.Fatal("failure not measured")
	}
	for _, secret := range []string{"private-card", "PRIVATE", "other.example"} {
		if strings.Contains(log.String(), secret) {
			t.Fatal("private request logged")
		}
	}
}

func TestTrustedProxyIsolationAndSpoofDefense(t *testing.T) {
	s := &Server{cfg: Config{ProxyClientIP: true}, rates: map[string]rate{}, active: make(chan struct{}, 32), mux: http.NewServeMux()}
	for range 21 {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "172.16.0.2:1000"
		r.Header.Set("X-Real-IP", "192.0.2.1")
		s.ServeHTTP(w, r)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "172.16.0.2:1000"
	r.Header.Set("X-Real-IP", "192.0.2.2")
	s.ServeHTTP(w, r)
	if w.Code == 429 {
		t.Fatal("proxy merges unrelated clients")
	}
	s.cfg.ProxyClientIP = false
	for i := range 21 {
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "192.0.2.3:1000"
		r.Header.Set("X-Real-IP", fmt.Sprintf("192.0.2.%d", i+5))
		s.ServeHTTP(w, r)
	}
	if w.Code != 429 {
		t.Fatal("direct client spoof bypasses limit")
	}
	s.cfg.ProxyClientIP = true
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Real-IP", "bad,header")
	s.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal("invalid proxy identity accepted")
	}
}
