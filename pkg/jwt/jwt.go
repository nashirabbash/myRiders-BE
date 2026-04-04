package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenError represents a JWT parsing/validation error
type TokenError struct {
	Code    string // "expired", "invalid", "malformed"
	Message string
}

func (e *TokenError) Error() string {
	return e.Message
}

// Claims represents custom JWT claims for TrackRide
type Claims struct {
	UserID string `json:"sub"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new JWT access token
func GenerateAccessToken(userID string, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken generates a new JWT refresh token
func GenerateRefreshToken(userID string, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken parses and validates a JWT token, returning specific error types
func ParseToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		// Check if token is expired by looking for the specific error
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, &TokenError{Code: "expired", Message: "token has expired"}
		}

		// All other errors are treated as invalid
		return nil, &TokenError{Code: "invalid", Message: "token is invalid or malformed"}
	}

	if !token.Valid {
		return nil, &TokenError{Code: "invalid", Message: "token is invalid"}
	}

	return claims, nil
}
