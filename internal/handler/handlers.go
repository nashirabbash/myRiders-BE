package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/nashirabbash/trackride/internal/config"
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
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

// AuthHandler is defined in auth.go with NewAuthHandler constructor
// UsersHandler is defined in users.go with NewUsersHandler constructor

// VehiclesHandler manages vehicle-related endpoints
type VehiclesHandler struct {
	queries *dbsqlc.Queries
	cfg     *config.Config
}

// NewVehiclesHandler creates a new VehiclesHandler
func NewVehiclesHandler(queries *dbsqlc.Queries, cfg *config.Config) *VehiclesHandler {
	return &VehiclesHandler{queries: queries, cfg: cfg}
}

// RidesHandler manages ride-related endpoints
type RidesHandler struct {
	queries *dbsqlc.Queries
	cfg     *config.Config
	redis   *redis.Client
}

// NewRidesHandler creates a new RidesHandler
func NewRidesHandler(queries *dbsqlc.Queries, cfg *config.Config, redis *redis.Client) *RidesHandler {
	return &RidesHandler{queries: queries, cfg: cfg, redis: redis}
}

// Methods Start, Stop, List, GetByID are defined in rides.go

// SocialHandler manages social features (follows, likes, comments)
type SocialHandler struct {
	queries *dbsqlc.Queries
	cfg     *config.Config
}

// NewSocialHandler creates a new SocialHandler
func NewSocialHandler(queries *dbsqlc.Queries, cfg *config.Config) *SocialHandler {
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
	queries *dbsqlc.Queries
	cfg     *config.Config
}

// NewLeaderboardHandler creates a new LeaderboardHandler
func NewLeaderboardHandler(queries *dbsqlc.Queries, cfg *config.Config) *LeaderboardHandler {
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
