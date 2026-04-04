package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nashirabbash/trackride/internal/config"
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler with queries and config
func NewAuthHandler(queries *dbsqlc.Queries, cfg *config.Config) *AuthHandler {
	authService := service.NewAuthService(queries, cfg)
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents a registration payload
type RegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
}

// LoginRequest represents a login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshRequest represents a refresh token payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterResponse is the DTO for successful registration
type RegisterResponse struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	AvatarURL    string `json:"avatar_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LoginResponse is the DTO for successful login
type LoginResponse struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	DisplayName  string `json:"display_name"`
	AvatarURL    string `json:"avatar_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// RefreshResponse is the DTO for token refresh
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	// Register user
	user, tokens, err := h.authService.Register(c.Request.Context(), service.RegisterRequest{
		Username:    req.Username,
		Email:       req.Email,
		Password:    req.Password,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		switch err.Error() {
		case "EMAIL_TAKEN":
			c.JSON(http.StatusConflict, gin.H{"error": "EMAIL_TAKEN"})
		case "USERNAME_TAKEN":
			c.JSON(http.StatusConflict, gin.H{"error": "USERNAME_TAKEN"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Return response
	c.JSON(http.StatusCreated, RegisterResponse{
		ID:           user.ID.String(),
		Username:     user.Username,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarUrl.String,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	// Authenticate user
	user, tokens, err := h.authService.Login(c.Request.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if err.Error() == "INVALID_CREDENTIALS" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_CREDENTIALS"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Return response
	c.JSON(http.StatusOK, LoginResponse{
		ID:           user.ID.String(),
		Username:     user.Username,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarUrl.String,
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	})
}

// Refresh handles refresh token exchange
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	// Refresh access token
	accessToken, expiresIn, err := h.authService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// For MVP, logout is stateless (just client discards token)
	// Future: implement token blacklist in Redis if needed
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
