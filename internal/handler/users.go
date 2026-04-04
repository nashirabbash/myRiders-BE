package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/middleware"
)

// UsersHandler handles user profile endpoints
type UsersHandler struct {
	queries *sqlc.Queries
}

// NewUsersHandler creates a new users handler
func NewUsersHandler(queries *sqlc.Queries) *UsersHandler {
	return &UsersHandler{
		queries: queries,
	}
}

// UpdateProfileRequest represents a profile update payload
type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,min=1,max=100"`
	AvatarURL   *string `json:"avatar_url" binding:"omitempty,url"`
}

// UserProfileResponse is the DTO for user profile (never includes password_hash)
type UserProfileResponse struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GetMe retrieves the authenticated user's profile
func (h *UsersHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	// Parse UUID
	id, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Fetch user
	user, err := h.queries.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "USER_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Return profile (DTO excludes password_hash)
	c.JSON(http.StatusOK, UserProfileResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl.String,
		CreatedAt:   user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   user.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// UpdateMe updates the authenticated user's profile
func (h *UsersHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	// Parse UUID
	id, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Update profile
	user, err := h.queries.UpdateUserProfile(c.Request.Context(), sqlc.UpdateUserProfileParams{
		ID:          id,
		DisplayName: derefString(req.DisplayName),
		AvatarUrl:   optionalStringPgtype2(req.AvatarURL),
		PushToken:   pgtype.Text{Valid: false},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Return updated profile
	c.JSON(http.StatusOK, UserProfileResponse{
		ID:          user.ID.String(),
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl.String,
		CreatedAt:   user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   user.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetProfile retrieves a public user profile by ID
func (h *UsersHandler) GetProfile(c *gin.Context) {
	userID := c.Param("id")

	// Parse UUID
	id, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}

	// Fetch user
	user, err := h.queries.GetUserByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "USER_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Return public profile (DTO excludes password_hash and email)
	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID.String(),
		"username":     user.Username,
		"display_name": user.DisplayName,
		"avatar_url":   user.AvatarUrl.String,
		"created_at":   user.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	})
}
