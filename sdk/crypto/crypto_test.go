package crypto

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("s3cret")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Fatalf("unexpected hash format: %s", hash)
	}
}

func TestVerifyPassword(t *testing.T) {
	hash, err := HashPassword("correct-horse")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	match, err := VerifyPassword("correct-horse", hash)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if !match {
		t.Fatal("expected password to match")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	hash, err := HashPassword("right-password")
	if err != nil {
		t.Fatalf("hash error: %v", err)
	}

	match, err := VerifyPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if match {
		t.Fatal("expected password NOT to match")
	}
}

func TestVerifyInvalidHash(t *testing.T) {
	_, err := VerifyPassword("anything", "not-a-valid-hash")
	if err == nil {
		t.Fatal("expected error for invalid hash, got nil")
	}
}
