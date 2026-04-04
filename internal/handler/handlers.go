package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/nashirabbash/trackride/internal/config"
	"github.com/nashirabbash/trackride/internal/db"
)

// Handlers holds all HTTP handlers for the application.
//
// In Phase 2 implementations, always use middleware.GetUserID() with explicit checking:
//
//	userID, ok := middleware.GetUserID(c)
//	if !ok {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
//		return
//	}
type Handlers struct {
	Auth        *AuthHandler
	Users       *UsersHandler
	Vehicles    *VehiclesHandler
	Rides       *RidesHandler
	Social      *SocialHandler
	Leaderboard *LeaderboardHandler
	Health      *HealthHandler
}

// AuthHandler manages authentication endpoints
type AuthHandler struct {
	queries db.Queries
	cfg     *config.Config
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(queries db.Queries, cfg *config.Config) *AuthHandler {
	return &AuthHandler{queries: queries, cfg: cfg}
}

// Register placeholder
func (h *AuthHandler) Register(c *gin.Context) {}

// Login placeholder
func (h *AuthHandler) Login(c *gin.Context) {}

// Refresh placeholder
func (h *AuthHandler) Refresh(c *gin.Context) {}

// Logout placeholder
func (h *AuthHandler) Logout(c *gin.Context) {}

// UsersHandler manages user-related endpoints
type UsersHandler struct {
	queries db.Queries
	cfg     *config.Config
}

// NewUsersHandler creates a new UsersHandler
func NewUsersHandler(queries db.Queries, cfg *config.Config) *UsersHandler {
	return &UsersHandler{queries: queries, cfg: cfg}
}

// GetMe placeholder
func (h *UsersHandler) GetMe(c *gin.Context) {}

// UpdateMe placeholder
func (h *UsersHandler) UpdateMe(c *gin.Context) {}

// GetProfile placeholder
func (h *UsersHandler) GetProfile(c *gin.Context) {}

// VehiclesHandler manages vehicle-related endpoints
type VehiclesHandler struct {
	queries db.Queries
	cfg     *config.Config
}

// NewVehiclesHandler creates a new VehiclesHandler
func NewVehiclesHandler(queries db.Queries, cfg *config.Config) *VehiclesHandler {
	return &VehiclesHandler{queries: queries, cfg: cfg}
}

// List placeholder
func (h *VehiclesHandler) List(c *gin.Context) {}

// Create placeholder
func (h *VehiclesHandler) Create(c *gin.Context) {}

// Update placeholder
func (h *VehiclesHandler) Update(c *gin.Context) {}

// Delete placeholder
func (h *VehiclesHandler) Delete(c *gin.Context) {}

// RidesHandler manages ride-related endpoints
type RidesHandler struct {
	queries db.Queries
	cfg     *config.Config
	redis   *redis.Client
}

// NewRidesHandler creates a new RidesHandler
func NewRidesHandler(queries db.Queries, cfg *config.Config, redis *redis.Client) *RidesHandler {
	return &RidesHandler{queries: queries, cfg: cfg, redis: redis}
}

// Start placeholder
func (h *RidesHandler) Start(c *gin.Context) {}

// Stop placeholder
func (h *RidesHandler) Stop(c *gin.Context) {}

// List placeholder
func (h *RidesHandler) List(c *gin.Context) {}

// GetByID placeholder
func (h *RidesHandler) GetByID(c *gin.Context) {}

// SocialHandler manages social features (follows, likes, comments)
type SocialHandler struct {
	queries db.Queries
	cfg     *config.Config
}

// NewSocialHandler creates a new SocialHandler
func NewSocialHandler(queries db.Queries, cfg *config.Config) *SocialHandler {
	return &SocialHandler{queries: queries, cfg: cfg}
}

// GetFeed placeholder
func (h *SocialHandler) GetFeed(c *gin.Context) {}

// Follow placeholder
func (h *SocialHandler) Follow(c *gin.Context) {}

// Unfollow placeholder
func (h *SocialHandler) Unfollow(c *gin.Context) {}

// LikeRide placeholder
func (h *SocialHandler) LikeRide(c *gin.Context) {}

// CommentRide placeholder
func (h *SocialHandler) CommentRide(c *gin.Context) {}

// LeaderboardHandler manages leaderboard endpoints
type LeaderboardHandler struct {
	queries db.Queries
	cfg     *config.Config
}

// NewLeaderboardHandler creates a new LeaderboardHandler
func NewLeaderboardHandler(queries db.Queries, cfg *config.Config) *LeaderboardHandler {
	return &LeaderboardHandler{queries: queries, cfg: cfg}
}

// GetGlobal placeholder
func (h *LeaderboardHandler) GetGlobal(c *gin.Context) {}

// GetFriends placeholder
func (h *LeaderboardHandler) GetFriends(c *gin.Context) {}

// HealthHandler manages health check endpoints
type HealthHandler struct{}

// Check returns a simple health status response
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
		"app":    "trackride-backend",
	})
}
