package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/nashirabbash/trackride/internal/config"
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/service"
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
	service *service.RidesService
}

// NewRidesHandler creates a new RidesHandler
func NewRidesHandler(queries *dbsqlc.Queries, cfg *config.Config, redis *redis.Client) *RidesHandler {
	return &RidesHandler{
		queries: queries,
		cfg:     cfg,
		redis:   redis,
		service: service.NewRidesService(queries),
	}
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

// Methods GetFeed, Follow, Unfollow, LikeRide, CommentRide are defined in social.go

// LeaderboardHandler manages leaderboard endpoints
type LeaderboardHandler struct {
	queries *dbsqlc.Queries
	cfg     *config.Config
}

// NewLeaderboardHandler creates a new LeaderboardHandler
func NewLeaderboardHandler(queries *dbsqlc.Queries, cfg *config.Config) *LeaderboardHandler {
	return &LeaderboardHandler{queries: queries, cfg: cfg}
}

// HealthHandler manages health check endpoints
type HealthHandler struct{}

// Check returns a simple health status response
//
//	@Summary		Health check
//	@Description	Returns the health status of the API server
//	@Tags			Health
//	@Produce		json
//	@Success		200		{object}	object
//	@Router			/health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
		"app":    "trackride-backend",
	})
}
