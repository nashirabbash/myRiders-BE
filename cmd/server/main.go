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

	// Ping PostgreSQL
	if err := pgPool.Ping(ctx); err != nil {
		log.Fatalf("[DB] PostgreSQL ping failed: %v", err)
	}
	log.Println("[DB] PostgreSQL connected ✓")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Ping Redis
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("[Redis] Redis ping failed: %v", err)
	}
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
