package jwt

import (
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

func TestGenerateAccessToken(t *testing.T) {
	userID := "test-user-123"
	secret := "my-secret-key-that-is-long-enough"
	ttl := 1 * time.Hour

	token, err := GenerateAccessToken(userID, secret, ttl)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("GenerateAccessToken returned empty token")
	}

	// Parse and verify the token
	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("Expected Subject %q, got %q", userID, claims.Subject)
	}

	if claims.Type != "access" {
		t.Errorf("Expected Type 'access', got %q", claims.Type)
	}

	if claims.UserID() != userID {
		t.Errorf("Expected UserID() to return %q, got %q", userID, claims.UserID())
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	userID := "test-user-456"
	secret := "my-secret-key-that-is-long-enough"
	ttl := 30 * 24 * time.Hour

	token, err := GenerateRefreshToken(userID, secret, ttl)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("GenerateRefreshToken returned empty token")
	}

	// Parse and verify the token
	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("Expected Subject %q, got %q", userID, claims.Subject)
	}

	if claims.Type != "refresh" {
		t.Errorf("Expected Type 'refresh', got %q", claims.Type)
	}
}

func TestParseToken_ValidToken(t *testing.T) {
	userID := "user-789"
	secret := "test-secret-key-for-tokens"
	ttl := 1 * time.Hour

	// Generate a token
	token, err := GenerateAccessToken(userID, secret, ttl)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Parse it back
	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.Subject != userID {
		t.Errorf("Expected Subject %q, got %q", userID, claims.Subject)
	}
}

func TestParseToken_InvalidSignature(t *testing.T) {
	userID := "user-signature-test"
	secret := "original-secret"
	ttl := 1 * time.Hour

	// Generate a token with one secret
	token, err := GenerateAccessToken(userID, secret, ttl)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Try to parse with a different secret
	_, err = ParseToken(token, "different-secret")
	if err == nil {
		t.Fatal("ParseToken should fail with wrong secret")
	}

	tokenErr, ok := err.(TokenError)
	if !ok {
		t.Fatalf("Expected TokenError, got %T: %v", err, err)
	}
	if tokenErr.Code != "TOKEN_INVALID" {
		t.Errorf("Expected error code 'TOKEN_INVALID', got %q", tokenErr.Code)
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	userID := "user-expiry-test"
	secret := "test-secret-key"
	ttl := -1 * time.Hour // Already expired

	// Generate a token with negative TTL
	token, err := GenerateAccessToken(userID, secret, ttl)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Try to parse it
	_, err = ParseToken(token, secret)
	if err == nil {
		t.Fatal("ParseToken should fail for expired token")
	}

	tokenErr, ok := err.(TokenError)
	if !ok {
		t.Fatalf("Expected TokenError, got %T: %v", err, err)
	}
	if tokenErr.Code != "TOKEN_EXPIRED" {
		t.Errorf("Expected error code 'TOKEN_EXPIRED', got %q", tokenErr.Code)
	}
}

func TestParseToken_MalformedToken(t *testing.T) {
	secret := "test-secret-key"
	malformedToken := "invalid.malformed.token"

	_, err := ParseToken(malformedToken, secret)
	if err == nil {
		t.Fatal("ParseToken should fail for malformed token")
	}

	tokenErr, ok := err.(TokenError)
	if !ok {
		t.Fatalf("Expected TokenError, got %T: %v", err, err)
	}
	if tokenErr.Code != "TOKEN_INVALID" {
		t.Errorf("Expected error code 'TOKEN_INVALID', got %q", tokenErr.Code)
	}
}

func TestClaimsUserID(t *testing.T) {
	claims := &Claims{
		Type: "access",
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject: "user-id-test",
		},
	}

	if claims.UserID() != "user-id-test" {
		t.Errorf("Expected UserID() to return 'user-id-test', got %q", claims.UserID())
	}
}

func TestTokenError_Error(t *testing.T) {
	err := TokenError{
		Code:    "TEST_ERROR",
		Message: "test error message",
	}

	if err.Error() != "test error message" {
		t.Errorf("Expected Error() to return 'test error message', got %q", err.Error())
	}
}
