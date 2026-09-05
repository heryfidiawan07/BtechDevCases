package security

import (
	"testing"
	"time"
)

func TestJWTIssueAndParse(t *testing.T) {
	t.Parallel()

	mgr := NewJWTManager("this-is-a-very-long-test-secret-key", 15*time.Minute)

	token, expiresAt, err := mgr.Issue(42, "alice@demo.com")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("token already expired")
	}

	claims, err := mgr.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Email != "alice@demo.com" {
		t.Fatalf("email %q", claims.Email)
	}

	id, err := UserIDFromClaims(claims)
	if err != nil {
		t.Fatalf("user id: %v", err)
	}
	if id != 42 {
		t.Fatalf("id %d", id)
	}
}

func TestJWTRejectsTamperedToken(t *testing.T) {
	t.Parallel()

	mgr := NewJWTManager("this-is-a-very-long-test-secret-key", time.Minute)
	token, _, err := mgr.Issue(1, "a@b.com")
	if err != nil {
		t.Fatal(err)
	}

	other := NewJWTManager("another-very-long-test-secret-key!", time.Minute)
	if _, err := other.Parse(token); err == nil {
		t.Fatal("expected unauthorized")
	}
}
