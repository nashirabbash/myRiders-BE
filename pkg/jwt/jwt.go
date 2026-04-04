package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims with standard and custom fields
type Claims struct {
	UserID string `json:"sub"`
	Type   string `json:"type"` // "access" or "refresh"
	jwtlib.RegisteredClaims
}

// TokenError represents JWT validation errors
type TokenError struct {
	Code    string // e.g., "TOKEN_EXPIRED", "TOKEN_INVALID"
	Message string
}

func (e TokenError) Error() string {
	return e.Message
}

// GenerateAccessToken creates a JWT access token
func GenerateAccessToken(userID string, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   "access",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			NotBefore: jwtlib.NewNumericDate(time.Now()),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a JWT refresh token
func GenerateRefreshToken(userID string, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Type:   "refresh",
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			NotBefore: jwtlib.NewNumericDate(time.Now()),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken validates and parses a token string
func ParseToken(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwtlib.ParseWithClaims(tokenString, claims, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, TokenError{
				Code:    "TOKEN_INVALID",
				Message: "invalid signing method",
			}
		}
		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return nil, TokenError{
				Code:    "TOKEN_EXPIRED",
				Message: "token has expired",
			}
		}
		return nil, TokenError{
			Code:    "TOKEN_INVALID",
			Message: "invalid or malformed token",
		}
	}

	if !token.Valid {
		return nil, TokenError{
			Code:    "TOKEN_INVALID",
			Message: "token is not valid",
		}
	}

	return claims, nil
}
