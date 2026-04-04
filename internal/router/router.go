package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nashirabbash/trackride/internal/config"
	"github.com/nashirabbash/trackride/internal/handler"
	"github.com/nashirabbash/trackride/internal/middleware"
	"github.com/nashirabbash/trackride/internal/websocket"
)

// Setup initializes and returns the Gin router with all routes mounted
func Setup(cfg *config.Config, h *handler.Handlers, hub *websocket.Hub) *gin.Engine {
	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.CORS())

	// Health check (no auth required)
	r.GET("/health", h.Health.Check)

	v1 := r.Group("/v1")

	// Public routes — Auth
	auth := v1.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
	}

	// Protected routes — require JWT access token
	authed := v1.Group("")
	authed.Use(middleware.Auth(cfg.JWTAccessSecret))
	{
		// Auth
		authed.POST("/auth/logout", h.Auth.Logout)

		// Users
		authed.GET("/users/me", h.Users.GetMe)
		authed.PUT("/users/me", h.Users.UpdateMe)
		authed.GET("/users/:id", h.Users.GetProfile)

		// Vehicles
		authed.GET("/vehicles", h.Vehicles.List)
		authed.POST("/vehicles", h.Vehicles.Create)
		authed.PUT("/vehicles/:id", h.Vehicles.Update)
		authed.DELETE("/vehicles/:id", h.Vehicles.Delete)

		// Rides
		authed.POST("/rides/start", h.Rides.Start)
		authed.POST("/rides/:id/stop", h.Rides.Stop)
		authed.GET("/rides", h.Rides.List)
		authed.GET("/rides/:id", h.Rides.GetByID)

		// Social
		authed.GET("/feed", h.Social.GetFeed)
		authed.POST("/users/:id/follow", h.Social.Follow)
		authed.DELETE("/users/:id/follow", h.Social.Unfollow)
		authed.POST("/rides/:id/like", h.Social.LikeRide)
		authed.POST("/rides/:id/comments", h.Social.CommentRide)

		// Leaderboard
		authed.GET("/leaderboard", h.Leaderboard.GetGlobal)
		authed.GET("/leaderboard/friends", h.Leaderboard.GetFriends)
	}

	// WebSocket route — auth via query param token
	v1.GET("/rides/:id/stream", hub.HandleWS)

	return r
}
