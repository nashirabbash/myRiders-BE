// Package testutil provides shared test infrastructure for API integration tests.
// Tests require a running PostgreSQL and Redis instance.
// Configure via environment variables:
//
//	TEST_DATABASE_URL  – PostgreSQL DSN (falls back to DATABASE_URL)
//	TEST_REDIS_URL     – Redis URL     (falls back to REDIS_URL)
//
// Tests are skipped automatically when neither variable is set.
package testutil

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	dbpkg "github.com/nashirabbash/trackride/internal/db"
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/config"
	"github.com/nashirabbash/trackride/internal/handler"
	"github.com/nashirabbash/trackride/internal/router"
	"github.com/nashirabbash/trackride/internal/websocket"
	"github.com/redis/go-redis/v9"
)

// Env holds live test dependencies.
type Env struct {
	Router  *gin.Engine
	Pool    *pgxpool.Pool
	Redis   *redis.Client
	Queries *dbsqlc.Queries
	Cfg     *config.Config
}

// Setup connects to the test database and Redis, sets up the Gin router, and
// returns an Env ready for use in tests. It calls t.Skip when credentials are
// missing so the suite degrades gracefully in environments without a DB.
func Setup(t *testing.T) *Env {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dbURL := firstNonEmpty(os.Getenv("TEST_DATABASE_URL"), os.Getenv("DATABASE_URL"))
	redisURL := firstNonEmpty(os.Getenv("TEST_REDIS_URL"), os.Getenv("REDIS_URL"))

	if dbURL == "" || redisURL == "" {
		t.Skip("TEST_DATABASE_URL and TEST_REDIS_URL must be set to run integration tests")
	}

	cfg := &config.Config{
		DatabaseURL:         dbURL,
		RedisURL:            redisURL,
		JWTAccessSecret:     "test-access-secret-at-least-32-characters-long",
		JWTRefreshSecret:    "test-refresh-secret-at-least-32-characters-long",
		JWTAccessTTL:        1 * time.Hour,
		JWTRefreshTTL:       720 * time.Hour,
		WsTokenTTL:          600,
		LeaderboardTimezone: "UTC",
		AppEnv:              "test",
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("testutil: pgxpool.New: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("testutil: DB ping failed: %v", err)
	}

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		t.Fatalf("testutil: redis.ParseURL: %v", err)
	}
	redisClient := redis.NewClient(redisOpts)

	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf("testutil: Redis ping failed: %v", err)
	}

	queries := dbpkg.NewQueries(pool)
	gpsBuffer := websocket.NewGPSBuffer(queries)
	wsHub := websocket.NewHub(gpsBuffer, redisClient)

	handlers := &handler.Handlers{
		Auth:        handler.NewAuthHandler(queries, cfg),
		Users:       handler.NewUsersHandler(queries),
		Vehicles:    handler.NewVehiclesHandler(queries, cfg),
		Rides:       handler.NewRidesHandler(queries, cfg, redisClient),
		Social:      handler.NewSocialHandler(queries, cfg),
		Leaderboard: handler.NewLeaderboardHandler(queries, cfg),
		Health:      &handler.HealthHandler{},
	}

	r := router.Setup(cfg, handlers, wsHub)

	env := &Env{
		Router:  r,
		Pool:    pool,
		Redis:   redisClient,
		Queries: queries,
		Cfg:     cfg,
	}

	t.Cleanup(func() {
		pool.Close()
		redisClient.Close()
	})

	return env
}

// Truncate deletes all rows from the given tables in a single statement so
// each test scenario starts from a clean slate. Pass tables in dependency order
// (children before parents) since TRUNCATE … CASCADE is used.
func Truncate(t *testing.T, pool *pgxpool.Pool, tables ...string) {
	t.Helper()
	for _, tbl := range tables {
		_, err := pool.Exec(context.Background(),
			"TRUNCATE TABLE "+tbl+" RESTART IDENTITY CASCADE")
		if err != nil {
			t.Fatalf("testutil: truncate %s: %v", tbl, err)
		}
	}
}

// TruncateAll resets every user-data table. Call this at the start of any
// test that touches multiple domains.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	Truncate(t, pool,
		"leaderboard_entries",
		"ride_comments",
		"ride_likes",
		"ride_gps_points",
		"rides",
		"follows",
		"vehicles",
		"users",
	)
}

// StatusIs asserts that the HTTP response has the expected status code.
func StatusIs(t *testing.T, expected int, actual *http.Response) {
	t.Helper()
	if actual.StatusCode != expected {
		t.Errorf("expected status %d, got %d", expected, actual.StatusCode)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
