package token_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"auth-as-a-service/app/token"
)

// mockCache is an in-memory redis.Service for testing.
type mockCache struct {
	data map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string]string)}
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	v, ok := m.data[key]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (m *mockCache) Set(_ context.Context, key string, value any, _ time.Duration) error {
	m.data[key] = fmt.Sprintf("%v", value)
	return nil
}

func (m *mockCache) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCache) Health() map[string]string { return nil }
func (m *mockCache) Close() error              { return nil }

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-tests")
	os.Setenv("JWT_EXPIRY_HOURS", "24")
	os.Setenv("REFRESH_TOKEN_EXPIRY_DAYS", "30")
	os.Exit(m.Run())
}

func TestValidToken(t *testing.T) {
	cache := newMockCache()
	tok, err := token.Generate("user-123")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	userID, err := token.Validate(context.Background(), tok, cache)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("expected user-123, got %s", userID)
	}
}

func TestExpiredToken(t *testing.T) {
	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"sub": "user-123",
		"jti": "expired-jti",
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	_, err = token.Validate(context.Background(), tok, newMockCache())
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestTamperedSignature(t *testing.T) {
	tok, err := token.Generate("user-123")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	tampered := tok[:len(tok)-4] + "xxxx"

	_, err = token.Validate(context.Background(), tampered, newMockCache())
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}

func TestRevokedToken(t *testing.T) {
	cache := newMockCache()
	tok, err := token.Generate("user-123")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if err := token.Revoke(context.Background(), tok, cache); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	_, err = token.Validate(context.Background(), tok, cache)
	if err == nil {
		t.Fatal("expected error for revoked token, got nil")
	}
}

func TestValidRefreshToken(t *testing.T) {
	cache := newMockCache()
	tok, err := token.GenerateRefresh("user-456")
	if err != nil {
		t.Fatalf("generate refresh: %v", err)
	}

	userID, err := token.ValidateRefresh(context.Background(), tok, cache)
	if err != nil {
		t.Fatalf("validate refresh: %v", err)
	}
	if userID != "user-456" {
		t.Errorf("expected user-456, got %s", userID)
	}
}

func TestRefreshTokenRejectedAsAccessToken(t *testing.T) {
	cache := newMockCache()
	tok, err := token.GenerateRefresh("user-456")
	if err != nil {
		t.Fatalf("generate refresh: %v", err)
	}

	_, err = token.Validate(context.Background(), tok, cache)
	if err == nil {
		t.Fatal("expected error using refresh token as access token, got nil")
	}
}

func TestAccessTokenRejectedAsRefreshToken(t *testing.T) {
	cache := newMockCache()
	tok, err := token.Generate("user-456")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = token.ValidateRefresh(context.Background(), tok, cache)
	if err == nil {
		t.Fatal("expected error using access token as refresh token, got nil")
	}
}

func TestRevokedRefreshToken(t *testing.T) {
	cache := newMockCache()
	tok, err := token.GenerateRefresh("user-456")
	if err != nil {
		t.Fatalf("generate refresh: %v", err)
	}

	if err := token.Revoke(context.Background(), tok, cache); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	_, err = token.ValidateRefresh(context.Background(), tok, cache)
	if err == nil {
		t.Fatal("expected error for revoked refresh token, got nil")
	}
}

func TestRotatedRefreshTokenRejected(t *testing.T) {
	cache := newMockCache()
	oldRefresh, err := token.GenerateRefresh("user-456")
	if err != nil {
		t.Fatalf("generate refresh: %v", err)
	}

	// Simulate rotation: revoke old refresh token.
	if err := token.Revoke(context.Background(), oldRefresh, cache); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	// Old refresh token must no longer be usable.
	_, err = token.ValidateRefresh(context.Background(), oldRefresh, cache)
	if err == nil {
		t.Fatal("expected error reusing rotated refresh token, got nil")
	}
}
