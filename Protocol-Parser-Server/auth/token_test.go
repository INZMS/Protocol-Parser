package auth

import (
	"strings"
	"testing"
)

func TestNewManagerFromEnvRejectsMissingAndWeakSecret(t *testing.T) {
	t.Setenv("AUTH_SECRET", "")
	if _, err := NewManagerFromEnv(); err == nil {
		t.Fatal("expected missing AUTH_SECRET to be rejected")
	}

	t.Setenv("AUTH_SECRET", "replace_with_a_long_random_secret")
	if _, err := NewManagerFromEnv(); err == nil {
		t.Fatal("expected example AUTH_SECRET to be rejected")
	}

	t.Setenv("AUTH_SECRET", strings.Repeat("a", 31))
	if _, err := NewManagerFromEnv(); err == nil {
		t.Fatal("expected short AUTH_SECRET to be rejected")
	}
}

func TestManagerIssuesAndVerifiesToken(t *testing.T) {
	t.Setenv("AUTH_SECRET", "test-only-secret-with-at-least-32-characters")
	t.Setenv("AUTH_EXPIRE_HOURS", "2")
	manager, err := NewManagerFromEnv()
	if err != nil {
		t.Fatalf("NewManagerFromEnv() error = %v", err)
	}
	token, err := manager.Issue(42, "tester")
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.UserID != 42 || claims.Username != "tester" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}
