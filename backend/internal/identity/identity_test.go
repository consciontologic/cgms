package identity

import (
	"context"
	"strings"
	"testing"
)

func TestPasswordCredentialsAndBounds(t *testing.T) {
	s := New(nil)
	credential, err := s.passwordHash(context.Background(), "long secret password 123")
	if err != nil {
		t.Fatal(err)
	}
	if string(credential) == "long secret password 123" {
		t.Fatal("plaintext credential")
	}
	if !s.passwordMatches(context.Background(), credential, "long secret password 123") || s.passwordMatches(context.Background(), credential, "wrong secret password 123") {
		t.Fatal("password verification")
	}
	if s.passwordMatches(context.Background(), []byte("malformed"), "long secret password 123") {
		t.Fatal("invalid hash accepted")
	}
	if _, err = s.passwordHash(context.Background(), "short"); err == nil {
		t.Fatal("short password accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = s.passwordHash(ctx, "long secret password 123"); err == nil {
		t.Fatal("canceled hashing accepted")
	}
}

func TestSessionTokenValidation(t *testing.T) {
	for _, token := range []string{"", "invalid", strings.Repeat("a", 10000), strings.Repeat("!", 43)} {
		if _, err := tokenHash(token); err == nil {
			t.Fatal("malformed token accepted")
		}
	}
	token, err := opaque()
	if err != nil {
		t.Fatal(err)
	}
	hash, err := tokenHash(token)
	if err != nil || len(hash) != 32 {
		t.Fatal("opaque token digest", err)
	}
}
