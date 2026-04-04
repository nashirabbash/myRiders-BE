package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Environment
	AppEnv string
	Port   string

	// Database
	DatabaseURL string
	DBMaxConns  int32
	DBMinConns  int32

	// Redis
	RedisURL string

	// JWT
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	// WebSocket
	WsTokenTTL int

	// Leaderboard
	LeaderboardTimezone string

	// Google Maps
	GoogleMapsKey string
}

// Load reads environment variables and returns a validated Config
func Load() *Config {
	// Load .env file if it exists (non-fatal if missing)
	_ = godotenv.Load()

	c := &Config{
		AppEnv:              getEnv("APP_ENV", "development"),
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         mustEnv("DATABASE_URL"),
		RedisURL:            mustEnv("REDIS_URL"),
		JWTAccessSecret:     mustEnv("JWT_ACCESS_SECRET"),
		JWTRefreshSecret:    mustEnv("JWT_REFRESH_SECRET"),
		GoogleMapsKey:       os.Getenv("GOOGLE_MAPS_API_KEY"),
		LeaderboardTimezone: getEnv("LEADERBOARD_TIMEZONE", "Asia/Jakarta"),
	}

	// Parse database pool sizes
	c.DBMaxConns = int32(parseIntEnv("DB_MAX_CONNS", 25))
	c.DBMinConns = int32(parseIntEnv("DB_MIN_CONNS", 5))

	// Parse JWT TTL durations
	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "1h"))
	if err != nil {
		log.Fatalf("invalid JWT_ACCESS_TTL: %v", err)
	}
	c.JWTAccessTTL = accessTTL

	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		log.Fatalf("invalid JWT_REFRESH_TTL: %v", err)
	}
	c.JWTRefreshTTL = refreshTTL

	// Parse WebSocket token TTL
	wsTokenTTL, err := strconv.Atoi(getEnv("WS_TOKEN_TTL", "600"))
	if err != nil {
		log.Fatalf("invalid WS_TOKEN_TTL: %v", err)
	}
	c.WsTokenTTL = wsTokenTTL

	return c
}

// mustEnv reads a required environment variable and exits if not found
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return v
}

// getEnv reads an optional environment variable with a fallback
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parseIntEnv parses an optional integer environment variable with a fallback
func parseIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("invalid %s: %v", key, err)
	}
	return n
}
