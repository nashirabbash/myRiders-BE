package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/nashirabbash/trackride/internal/config"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/handler"
	"github.com/nashirabbash/trackride/internal/jobs"
	"github.com/nashirabbash/trackride/internal/router"
	"github.com/nashirabbash/trackride/internal/websocket"
)

func main() {
	// Load config first
	cfg := config.Load()

	log.Printf("[Startup] TrackRide Backend v1.0 (env=%s)", cfg.AppEnv)

	// Initialize PostgreSQL connection pool
	ctx := context.Background()
	pgConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[DB] Failed to parse database URL: %v", err)
	}

	pgConfig.MaxConns = 25
	pgConfig.MinConns = 5

	pgPool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		log.Fatalf("[DB] Failed to create connection pool: %v", err)
	}

	// Ping PostgreSQL with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	if err := pgPool.Ping(pingCtx); err != nil {
		cancel()
		log.Fatalf("[DB] PostgreSQL ping failed: %v", err)
	}
	cancel()
	log.Println("[DB] PostgreSQL connected ✓")

	// Initialize Redis client from URL
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("[Redis] Failed to parse Redis URL: %v", err)
	}
	redisOpts.MaxRetries = 3
	redisOpts.PoolSize = 10
	redisOpts.MinIdleConns = 2

	redisClient := redis.NewClient(redisOpts)

	// Ping Redis with timeout
	pingCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		cancel()
		log.Fatalf("[Redis] Redis ping failed: %v", err)
	}
	cancel()
	log.Println("[Redis] Redis connected ✓")

	// Initialize query layer
	queries := sqlc.New(pgPool)

	// Initialize WebSocket components
	gpsBuffer := websocket.NewGPSBuffer(queries)
	wsHub := websocket.NewHub(gpsBuffer, redisClient)

	// Initialize handlers
	handlers := &handler.Handlers{
		Auth:        handler.NewAuthHandler(queries, cfg),
		Users:       handler.NewUsersHandler(queries, cfg),
		Vehicles:    handler.NewVehiclesHandler(queries, cfg),
		Rides:       handler.NewRidesHandler(queries, cfg, redisClient),
		Social:      handler.NewSocialHandler(queries, cfg),
		Leaderboard: handler.NewLeaderboardHandler(queries, cfg),
		Health:      &handler.HealthHandler{},
	}

	// Initialize leaderboard cron job
	leaderboardJob := jobs.NewLeaderboardJob(queries)
	leaderboardJob.Start()

	// Setup router
	ginRouter := router.Setup(cfg, handlers, wsHub)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      ginRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("[Server] Starting on http://0.0.0.0:%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Server] Failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("[Shutdown] Graceful shutdown initiated...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[Shutdown] HTTP server shutdown error: %v", err)
	}
	log.Println("[Shutdown] HTTP server stopped ✓")

	// Close database connections
	pgPool.Close()
	log.Println("[Shutdown] PostgreSQL closed ✓")

	// Close Redis connection
	redisClient.Close()
	log.Println("[Shutdown] Redis closed ✓")

	log.Println("[Shutdown] Complete")
}
