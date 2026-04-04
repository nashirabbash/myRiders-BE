package jwt

import (
	"testing"
	"time"
)

const testSecret = "test-secret-key-at-least-32-characters-long-for-hs256"

func TestGenerateAccessToken(t *testing.T) {
	userID := "test-user-id"
	ttl := 1 * time.Hour

	token, err := GenerateAccessToken(userID, testSecret, ttl)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := "test-user-id"
	ttl := 7 * 24 * time.Hour

	token, err := GenerateRefreshToken(userID, testSecret, ttl)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestParseValidAccessToken(t *testing.T) {
	userID := "test-user-id"
	ttl := 1 * time.Hour

	// Generate token
	tokenString, err := GenerateAccessToken(userID, testSecret, ttl)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Parse token
	claims, err := ParseToken(tokenString, testSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected user_id %q, got %q", userID, claims.UserID)
	}

	if claims.Type != "access" {
		t.Fatalf("expected type 'access', got %q", claims.Type)
	}
}

func TestParseValidRefreshToken(t *testing.T) {
	userID := "test-user-id"
	ttl := 7 * 24 * time.Hour

	// Generate token
	tokenString, err := GenerateRefreshToken(userID, testSecret, ttl)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Parse token
	claims, err := ParseToken(tokenString, testSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected user_id %q, got %q", userID, claims.UserID)
	}

	if claims.Type != "refresh" {
		t.Fatalf("expected type 'refresh', got %q", claims.Type)
	}
}

func TestParseExpiredToken(t *testing.T) {
	userID := "test-user-id"
	ttl := -1 * time.Hour // Negative duration = already expired

	// Generate expired token
	tokenString, err := GenerateAccessToken(userID, testSecret, ttl)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Parse should fail
	_, err = ParseToken(tokenString, testSecret)
	if err == nil {
		t.Fatal("expected error for expired token")
	}

	tokenErr, ok := err.(TokenError)
	if !ok {
		t.Fatalf("expected TokenError, got %T", err)
	}

	if tokenErr.Code != "TOKEN_EXPIRED" {
		t.Fatalf("expected TOKEN_EXPIRED, got %s", tokenErr.Code)
	}
}

func TestParseInvalidSignature(t *testing.T) {
	userID := "test-user-id"
	ttl := 1 * time.Hour

	// Generate token with testSecret
	tokenString, err := GenerateAccessToken(userID, testSecret, ttl)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Try to parse with different secret
	wrongSecret := "wrong-secret-key"
	_, err = ParseToken(tokenString, wrongSecret)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}

	tokenErr, ok := err.(TokenError)
	if !ok {
		t.Fatalf("expected TokenError, got %T", err)
	}

	if tokenErr.Code != "TOKEN_INVALID" {
		t.Fatalf("expected TOKEN_INVALID, got %s", tokenErr.Code)
	}
}

func TestParseMalformedToken(t *testing.T) {
	malformedToken := "this.is.not.a.valid.jwt.token"

	_, err := ParseToken(malformedToken, testSecret)
	if err == nil {
		t.Fatal("expected error for malformed token")
	}

	tokenErr, ok := err.(TokenError)
	if !ok {
		t.Fatalf("expected TokenError, got %T", err)
	}

	if tokenErr.Code != "TOKEN_INVALID" {
		t.Fatalf("expected TOKEN_INVALID, got %s", tokenErr.Code)
	}
}
