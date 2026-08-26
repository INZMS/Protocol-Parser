package auth

import (
	"testing"
	"time"
)

func TestIssueAndVerify(t *testing.T) {
	manager := &Manager{secret: []byte("test-secret"), ttl: time.Hour}
	token, err := manager.Issue(7, "admin")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 7 || claims.Username != "admin" {
		t.Fatalf("unexpected claims: %#v", claims)
	}
}

func TestRejectTamperedToken(t *testing.T) {
	manager := &Manager{secret: []byte("test-secret"), ttl: time.Hour}
	token, _ := manager.Issue(7, "admin")
	if _, err := manager.Verify(token + "x"); err == nil {
		t.Fatal("expected tampered token error")
	}
}
