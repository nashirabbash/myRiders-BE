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
	dbpkg "github.com/nashirabbash/trackride/internal/db"
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

	pgConfig.MaxConns = cfg.DBMaxConns
	pgConfig.MinConns = cfg.DBMinConns

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

	// Initialize database store
	// In Phase 2, this will be initialized with actual sqlc queries:
	// queries := sqlc.New(pgPool)
	store := dbpkg.NewStore(pgPool)
	_ = store

	// Initialize WebSocket components
	// In Phase 2, pass store to GPS buffer for actual GPS point persistence
	gpsBuffer := websocket.NewGPSBuffer(nil)
	wsHub := websocket.NewHub(gpsBuffer, redisClient)

	// Initialize handlers
	// In Phase 2, pass store to handlers for actual database operations
	handlers := &handler.Handlers{
		Auth:        handler.NewAuthHandler(nil, cfg),
		Users:       handler.NewUsersHandler(nil, cfg),
		Vehicles:    handler.NewVehiclesHandler(nil, cfg),
		Rides:       handler.NewRidesHandler(nil, cfg, redisClient),
		Social:      handler.NewSocialHandler(nil, cfg),
		Leaderboard: handler.NewLeaderboardHandler(nil, cfg),
		Health:      &handler.HealthHandler{},
	}

	// Initialize leaderboard cron job
	// In Phase 2, pass store to leaderboard job for actual ranking computation
	leaderboardJob := jobs.NewLeaderboardJob(nil, cfg.LeaderboardTimezone)
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

	// Start server in a goroutine; send fatal errors to errChan so the main
	// goroutine can trigger a graceful shutdown instead of os.Exit via Fatalf.
	errChan := make(chan error, 1)
	go func() {
		log.Printf("[Server] Starting on http://0.0.0.0:%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for OS signal or a server startup error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("[Shutdown] Received signal: %s", sig)
	case err := <-errChan:
		log.Printf("[Shutdown] Server error: %v — shutting down", err)
	}

	log.Println("[Shutdown] Graceful shutdown initiated...")

	// Stop cron scheduler first and wait for any in-flight job to finish
	cronCtx := leaderboardJob.Stop()
	<-cronCtx.Done()
	log.Println("[Shutdown] Leaderboard job stopped ✓")

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
