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

// ErrorResponse is the standard error response
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

// MessageResponse is the standard message response
type MessageResponse struct {
	Message string `json:"message"`
}

// GenericResponse is a generic map response for flexible endpoints
type GenericResponse map[string]interface{}

// Register handles user registration
//
//	@Summary		Register a new user
//	@Description	Create a new user account and receive access and refresh tokens
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RegisterRequest	true	"Registration credentials"
//	@Success		201		{object}	RegisterResponse
//	@Failure		409		{object}	ErrorResponse	"Email or username already taken"
//	@Failure		422		{object}	ErrorResponse	"Validation error"
//	@Router			/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, err.Error())
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
		RespondWithError(c, err)
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
//
//	@Summary		Login user
//	@Description	Authenticate user and receive access and refresh tokens
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginRequest	true	"Login credentials"
//	@Success		200		{object}	LoginResponse
//	@Failure		401		{object}	ErrorResponse	"Invalid credentials"
//	@Failure		422		{object}	ErrorResponse	"Validation error"
//	@Router			/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, err.Error())
		return
	}

	// Authenticate user
	user, tokens, err := h.authService.Login(c.Request.Context(), service.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		RespondWithError(c, err)
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
//
//	@Summary		Refresh access token
//	@Description	Exchange refresh token for a new access token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		RefreshRequest	true	"Refresh token"
//	@Success		200		{object}	RefreshResponse
//	@Failure		401		{object}	ErrorResponse	"Invalid refresh token"
//	@Failure		422		{object}	ErrorResponse	"Validation error"
//	@Router			/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, err.Error())
		return
	}

	// Refresh access token
	accessToken, expiresIn, err := h.authService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, RefreshResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	})
}

// Logout handles user logout
//
//	@Summary		Logout user
//	@Description	Logout the authenticated user. Note: For MVP, this is stateless and the client should discard the token.
//	@Tags			Auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{object}	MessageResponse
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Router			/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// For MVP, logout is stateless (just client discards token)
	// Future: implement token blacklist in Redis if needed
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
