package auth

import (
	"testing"
	"github.com/google/uuid"
	"time"
)

func TestValidToken(t *testing.T) {
	newID := uuid.New()
	newJWT, err := MakeJWT(newID, "test", 60 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id, err := ValidateJWT(newJWT, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != newID {
		t.Fatalf("expected id %v, got %v", newID, id)
	}
}

func TestExpiredToken(t *testing.T) {
	newID := uuid.New()
	newJWT, err := MakeJWT(newID, "test", -60 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = ValidateJWT(newJWT, "test")
	if err == nil {
		t.Errorf("expected expired token error, got nil")
	}
}

func TestWrongSecret(t *testing.T) {
	newID := uuid.New()
	newJWT, err := MakeJWT(newID, "test", 60 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = ValidateJWT(newJWT, "wrongtest")
	if err == nil {
		t.Errorf("expected error for wrong secret, got nil")
	}
}