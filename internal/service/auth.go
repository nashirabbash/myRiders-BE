package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/nashirabbash/trackride/internal/config"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
	jwtpkg "github.com/nashirabbash/trackride/pkg/jwt"
)

// AuthService handles authentication business logic
type AuthService struct {
	queries *sqlc.Queries
	cfg     *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(queries *sqlc.Queries, cfg *config.Config) *AuthService {
	return &AuthService{
		queries: queries,
		cfg:     cfg,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username    string
	Email       string
	Password    string
	DisplayName string
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string
	Password string
}

// AuthTokens represents access and refresh tokens
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword compares a password with its hash
func (s *AuthService) VerifyPassword(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*sqlc.CreateUserRow, *AuthTokens, error) {
	// Check if email already exists
	_, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, nil, fmt.Errorf("EMAIL_TAKEN")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("INTERNAL_ERROR")
	}

	// Check if username already exists
	_, err = s.queries.GetUserByUsername(ctx, req.Username)
	if err == nil {
		return nil, nil, fmt.Errorf("USERNAME_TAKEN")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, fmt.Errorf("INTERNAL_ERROR")
	}

	// Hash password
	passwordHash, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, nil, err
	}

	// Create user
	user, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
		DisplayName:  req.DisplayName,
	})
	if err != nil {
		return nil, nil, err
	}

	// Generate tokens
	tokens, err := s.GenerateTokens(user.ID.String())
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*sqlc.User, *AuthTokens, error) {
	// Get user by email
	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, fmt.Errorf("INVALID_CREDENTIALS")
	}

	// Verify password
	if !s.VerifyPassword(user.PasswordHash, req.Password) {
		return nil, nil, fmt.Errorf("INVALID_CREDENTIALS")
	}

	// Generate tokens
	tokens, err := s.GenerateTokens(user.ID.String())
	if err != nil {
		return nil, nil, err
	}

	return &user, tokens, nil
}

// GenerateTokens generates both access and refresh tokens
func (s *AuthService) GenerateTokens(userID string) (*AuthTokens, error) {
	// Use configured TTLs (already parsed in config)
	accessTTL := s.cfg.JWTAccessTTL
	if accessTTL == 0 {
		accessTTL = 1 * time.Hour
	}

	refreshTTL := s.cfg.JWTRefreshTTL
	if refreshTTL == 0 {
		refreshTTL = 720 * time.Hour // 30 days
	}

	// Generate access token
	accessToken, err := jwtpkg.GenerateAccessToken(userID, s.cfg.JWTAccessSecret, accessTTL)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := jwtpkg.GenerateRefreshToken(userID, s.cfg.JWTRefreshSecret, refreshTTL)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTTL.Seconds()),
	}, nil
}

// RefreshAccessToken validates a refresh token and generates a new access token
// Returns the access token and its expiry in seconds
func (s *AuthService) RefreshAccessToken(refreshTokenStr string) (string, int64, error) {
	// Parse refresh token
	claims, err := jwtpkg.ParseToken(refreshTokenStr, s.cfg.JWTRefreshSecret)
	if err != nil {
		return "", 0, fmt.Errorf("TOKEN_INVALID")
	}

	// Verify token type
	if claims.Type != "refresh" {
		return "", 0, fmt.Errorf("TOKEN_INVALID")
	}

	// Use configured access token TTL
	accessTTL := s.cfg.JWTAccessTTL
	if accessTTL == 0 {
		accessTTL = 1 * time.Hour
	}

	// Generate new access token
	accessToken, err := jwtpkg.GenerateAccessToken(claims.UserID(), s.cfg.JWTAccessSecret, accessTTL)
	if err != nil {
		return "", 0, err
	}

	return accessToken, int64(accessTTL.Seconds()), nil
}
