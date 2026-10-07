//go:build integration

package transport

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthenticationDependencyFailureIsUnavailable(t *testing.T) {
	s := journeyServer(t)
	guest := journeyGuests(t, s, 1)[0]
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := httptest.NewRequest("GET", "/v1/session", nil).WithContext(ctx)
	r.Header.Set("Authorization", "Bearer "+guest.Token)
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 503 || !strings.Contains(w.Body.String(), "UNAVAILABLE") {
		t.Fatalf("dependency failure misclassified: status %d", w.Code)
	}
}
