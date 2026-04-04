# TrackRide — Backend Developer Guide
**Go · PostgreSQL · WebSocket · MVP v1.0**

> Dokumen ini adalah panduan teknis khusus untuk backend developer.
> Untuk spesifikasi frontend/mobile, lihat `CLAUDE-frontend.md`.

---

## Quick Reference

| Item | Detail |
|---|---|
| Language | Go 1.22+ |
| HTTP Framework | Gin (`github.com/gin-gonic/gin`) |
| Database | PostgreSQL 16 |
| ORM / Query | sqlc + pgx/v5 |
| Auth | JWT (golang-jwt/jwt/v5) |
| WebSocket | gorilla/websocket |
| Cache / Token store | Redis (go-redis/v9) |
| Job scheduler | robfig/cron/v3 |
| Port | 8080 |
| Base URL | `https://api.trackride.app/v1` |
| WS URL | `wss://api.trackride.app/v1` |

---

## 1. Project Setup

### 1.1 Init & dependencies

```bash
mkdir trackride-backend && cd trackride-backend
go mod init github.com/yourorg/trackride

go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get github.com/gorilla/websocket
go get github.com/redis/go-redis/v9
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/stdlib
go get github.com/jmoiron/sqlx
go get github.com/robfig/cron/v3
go get golang.org/x/crypto
go get github.com/google/uuid
go get github.com/joho/godotenv

# sqlc untuk generate query code dari SQL
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 1.2 Struktur folder

```
cmd/
└── server/
    └── main.go             # Entry point

internal/
├── config/
│   └── config.go           # Load env, validasi
├── db/
│   ├── migrations/         # SQL migration files
│   ├── queries/            # SQL queries untuk sqlc
│   └── sqlc/               # Generated code dari sqlc
├── handler/                # HTTP handlers (satu per domain)
│   ├── auth.go
│   ├── users.go
│   ├── vehicles.go
│   ├── rides.go
│   ├── social.go
│   └── leaderboard.go
├── middleware/
│   ├── auth.go             # JWT verify middleware
│   └── cors.go
├── websocket/
│   ├── hub.go              # GPS WebSocket server
│   └── buffer.go           # In-memory GPS batch buffer
├── service/                # Business logic
│   ├── auth.go
│   ├── rides.go
│   ├── metrics.go
│   ├── geocoding.go
│   └── notifications.go
├── jobs/
│   └── leaderboard.go      # Cron: compute ranking tiap malam
└── router/
    └── router.go           # Route registration

pkg/
├── polyline/               # Google Encoded Polyline encode/decode
│   └── polyline.go
└── jwt/
    └── jwt.go              # JWT helper
```

### 1.3 Environment variables

```bash
# .env
DATABASE_URL=postgres://user:password@localhost:5432/trackride?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_ACCESS_SECRET=your-access-secret-min-32-chars
JWT_REFRESH_SECRET=your-refresh-secret-min-32-chars
JWT_ACCESS_TTL=1h
JWT_REFRESH_TTL=720h
WS_TOKEN_TTL=600
GOOGLE_MAPS_API_KEY=your-key
PORT=8080
APP_ENV=development
```

```go
// internal/config/config.go
package config

import (
    "log"
    "os"
    "github.com/joho/godotenv"
)

type Config struct {
    DatabaseURL      string
    RedisURL         string
    JWTAccessSecret  string
    JWTRefreshSecret string
    JWTAccessTTL     string
    JWTRefreshTTL    string
    WsTokenTTL       int
    GoogleMapsKey    string
    Port             string
}

func Load() *Config {
    _ = godotenv.Load()
    c := &Config{
        DatabaseURL:      mustEnv("DATABASE_URL"),
        RedisURL:         mustEnv("REDIS_URL"),
        JWTAccessSecret:  mustEnv("JWT_ACCESS_SECRET"),
        JWTRefreshSecret: mustEnv("JWT_REFRESH_SECRET"),
        JWTAccessTTL:     getEnv("JWT_ACCESS_TTL", "1h"),
        JWTRefreshTTL:    getEnv("JWT_REFRESH_TTL", "720h"),
        WsTokenTTL:       600,
        GoogleMapsKey:    os.Getenv("GOOGLE_MAPS_API_KEY"),
        Port:             getEnv("PORT", "8080"),
    }
    return c
}

func mustEnv(k string) string {
    v := os.Getenv(k)
    if v == "" { log.Fatalf("env %s is required", k) }
    return v
}

func getEnv(k, fallback string) string {
    if v := os.Getenv(k); v != "" { return v }
    return fallback
}
```

---

## 2. Database Schema & Migrations

### 2.1 Migration files

```sql
-- db/migrations/001_init.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE vehicle_type AS ENUM ('motor', 'mobil', 'sepeda');
CREATE TYPE ride_status AS ENUM ('active', 'completed', 'abandoned');

CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username     TEXT UNIQUE NOT NULL,
    email        TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url   TEXT,
    push_token   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vehicles (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type      vehicle_type NOT NULL,
    name      TEXT NOT NULL,
    brand     TEXT,
    color     TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rides (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id),
    vehicle_id       UUID NOT NULL REFERENCES vehicles(id),
    title            TEXT,
    started_at       TIMESTAMPTZ NOT NULL,
    ended_at         TIMESTAMPTZ,
    distance_km      FLOAT8 NOT NULL DEFAULT 0,
    duration_seconds INT NOT NULL DEFAULT 0,
    max_speed_kmh    FLOAT8 NOT NULL DEFAULT 0,
    avg_speed_kmh    FLOAT8 NOT NULL DEFAULT 0,
    elevation_m      FLOAT8 NOT NULL DEFAULT 0,
    calories         INT NOT NULL DEFAULT 0,
    route_summary    JSONB,
    status           ride_status NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ride_gps_points (
    id          BIGSERIAL PRIMARY KEY,
    ride_id     UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    latitude    FLOAT8 NOT NULL,
    longitude   FLOAT8 NOT NULL,
    speed_kmh   FLOAT8 NOT NULL DEFAULT 0,
    elevation_m FLOAT8 NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE follows (
    follower_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id)
);

CREATE TABLE ride_likes (
    ride_id    UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ride_id, user_id)
);

CREATE TABLE ride_comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id    UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    VARCHAR(280) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE leaderboard_entries (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id),
    vehicle_type vehicle_type,
    period_type  TEXT NOT NULL,
    period_start DATE NOT NULL,
    total_km     FLOAT8 NOT NULL DEFAULT 0,
    total_rides  INT NOT NULL DEFAULT 0,
    rank         INT NOT NULL
);
```

```sql
-- db/migrations/002_indexes.sql

CREATE INDEX idx_rides_user_completed ON rides(user_id, started_at DESC)
    WHERE status = 'completed';
CREATE INDEX idx_rides_vehicle ON rides(vehicle_id);
CREATE INDEX idx_gps_ride_time ON ride_gps_points(ride_id, recorded_at ASC);
CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_leaderboard_period ON leaderboard_entries(period_type, period_start, rank ASC);
```

### 2.2 sqlc config

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/queries/"
    schema: "internal/db/migrations/"
    gen:
      go:
        package: "db"
        out: "internal/db/sqlc"
        emit_json_tags: true
        emit_prepared_queries: false
```

```sql
-- internal/db/queries/rides.sql

-- name: CreateRide :one
INSERT INTO rides (user_id, vehicle_id, started_at, status)
VALUES ($1, $2, $3, 'active')
RETURNING *;

-- name: GetRideByID :one
SELECT r.*, v.type as vehicle_type, v.name as vehicle_name
FROM rides r
JOIN vehicles v ON v.id = r.vehicle_id
WHERE r.id = $1;

-- name: UpdateRideCompleted :one
UPDATE rides SET
    ended_at = $2, status = 'completed',
    distance_km = $3, duration_seconds = $4,
    max_speed_kmh = $5, avg_speed_kmh = $6,
    elevation_m = $7, calories = $8,
    route_summary = $9
WHERE id = $1
RETURNING *;

-- name: ListRidesByUser :many
SELECT r.*, v.type as vehicle_type, v.name as vehicle_name
FROM rides r
JOIN vehicles v ON v.id = r.vehicle_id
WHERE r.user_id = $1 AND r.status = 'completed'
ORDER BY r.started_at DESC
LIMIT $2 OFFSET $3;

-- name: InsertGPSPointsBatch :exec
INSERT INTO ride_gps_points (ride_id, latitude, longitude, speed_kmh, elevation_m, recorded_at)
SELECT unnest($1::uuid[]), unnest($2::float8[]), unnest($3::float8[]),
       unnest($4::float8[]), unnest($5::float8[]), unnest($6::timestamptz[]);

-- name: GetGPSPointsByRide :many
SELECT * FROM ride_gps_points
WHERE ride_id = $1
ORDER BY recorded_at ASC;
```

```bash
# Generate code
sqlc generate
```

---

## 3. JWT & Authentication

```go
// pkg/jwt/jwt.go
package jwtpkg

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID string `json:"sub"`
    Type   string `json:"type"`
    jwt.RegisteredClaims
}

func GenerateAccessToken(userID, secret string) (string, error) {
    claims := Claims{
        UserID: userID,
        Type:   "access",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func GenerateRefreshToken(userID, secret string) (string, error) {
    claims := Claims{
        UserID: userID,
        Type:   "refresh",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })
    if err != nil { return nil, err }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, jwt.ErrSignatureInvalid
}
```

### 3.1 Auth middleware

```go
// internal/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    jwtpkg "github.com/yourorg/trackride/pkg/jwt"
)

func Auth(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
            return
        }

        claims, err := jwtpkg.ParseToken(strings.TrimPrefix(authHeader, "Bearer "), secret)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_EXPIRED"})
            return
        }

        c.Set("user_id", claims.UserID)
        c.Next()
    }
}

// Helper untuk ambil user_id dari context
func GetUserID(c *gin.Context) string {
    return c.GetString("user_id")
}
```

---

## 4. Router

```go
// internal/router/router.go
package router

import (
    "github.com/gin-gonic/gin"
    "github.com/yourorg/trackride/internal/handler"
    "github.com/yourorg/trackride/internal/middleware"
    "github.com/yourorg/trackride/internal/websocket"
    "github.com/yourorg/trackride/internal/config"
)

func Setup(cfg *config.Config, h *handler.Handlers, hub *websocket.Hub) *gin.Engine {
    r := gin.Default()
    r.Use(middleware.CORS())

    v1 := r.Group("/v1")

    // Public
    auth := v1.Group("/auth")
    {
        auth.POST("/register", h.Auth.Register)
        auth.POST("/login", h.Auth.Login)
        auth.POST("/refresh", h.Auth.Refresh)
    }

    // Protected
    authed := v1.Group("")
    authed.Use(middleware.Auth(cfg.JWTAccessSecret))
    {
        authed.POST("/auth/logout", h.Auth.Logout)

        authed.GET("/users/me", h.Users.GetMe)
        authed.PUT("/users/me", h.Users.UpdateMe)
        authed.GET("/users/:id", h.Users.GetProfile)

        authed.GET("/vehicles", h.Vehicles.List)
        authed.POST("/vehicles", h.Vehicles.Create)
        authed.PUT("/vehicles/:id", h.Vehicles.Update)
        authed.DELETE("/vehicles/:id", h.Vehicles.Delete)

        authed.POST("/rides/start", h.Rides.Start)
        authed.POST("/rides/:id/stop", h.Rides.Stop)
        authed.GET("/rides", h.Rides.List)
        authed.GET("/rides/:id", h.Rides.GetByID)

        authed.GET("/feed", h.Social.GetFeed)
        authed.POST("/users/:id/follow", h.Social.Follow)
        authed.DELETE("/users/:id/follow", h.Social.Unfollow)
        authed.POST("/rides/:id/like", h.Social.LikeRide)
        authed.POST("/rides/:id/comments", h.Social.CommentRide)

        authed.GET("/leaderboard", h.Leaderboard.GetGlobal)
        authed.GET("/leaderboard/friends", h.Leaderboard.GetFriends)
    }

    // WebSocket — auth via query param token
    v1.GET("/rides/:id/stream", hub.HandleWS)

    return r
}
```

---

## 5. Handlers

### 5.1 Auth handler

```go
// internal/handler/auth.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "golang.org/x/crypto/bcrypt"
    "github.com/yourorg/trackride/internal/db/sqlc"
    jwtpkg "github.com/yourorg/trackride/pkg/jwt"
    "github.com/yourorg/trackride/internal/config"
)

type AuthHandler struct {
    queries *db.Queries
    cfg     *config.Config
}

type registerRequest struct {
    Username    string `json:"username" binding:"required,min=3,max=30"`
    Email       string `json:"email"    binding:"required,email"`
    Password    string `json:"password" binding:"required,min=8"`
    DisplayName string `json:"display_name" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req registerRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
        return
    }

    // Cek email / username sudah ada
    if _, err := h.queries.GetUserByEmail(c, req.Email); err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "EMAIL_TAKEN"})
        return
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
        return
    }

    user, err := h.queries.CreateUser(c, db.CreateUserParams{
        Username:     req.Username,
        Email:        req.Email,
        PasswordHash: string(hash),
        DisplayName:  req.DisplayName,
    })
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
        return
    }

    accessToken, _ := jwtpkg.GenerateAccessToken(user.ID.String(), h.cfg.JWTAccessSecret)
    refreshToken, _ := jwtpkg.GenerateRefreshToken(user.ID.String(), h.cfg.JWTRefreshSecret)

    c.JSON(http.StatusCreated, gin.H{
        "user":          user,
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "expires_in":    3600,
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req struct {
        Email    string `json:"email"    binding:"required,email"`
        Password string `json:"password" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR"})
        return
    }

    user, err := h.queries.GetUserByEmail(c, req.Email)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_CREDENTIALS"})
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "INVALID_CREDENTIALS"})
        return
    }

    accessToken, _ := jwtpkg.GenerateAccessToken(user.ID.String(), h.cfg.JWTAccessSecret)
    refreshToken, _ := jwtpkg.GenerateRefreshToken(user.ID.String(), h.cfg.JWTRefreshSecret)

    c.JSON(http.StatusOK, gin.H{
        "user":          user,
        "access_token":  accessToken,
        "refresh_token": refreshToken,
        "expires_in":    3600,
    })
}

func (h *AuthHandler) Refresh(c *gin.Context) {
    var req struct{ RefreshToken string `json:"refresh_token" binding:"required"` }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "VALIDATION_ERROR"})
        return
    }

    claims, err := jwtpkg.ParseToken(req.RefreshToken, h.cfg.JWTRefreshSecret)
    if err != nil || claims.Type != "refresh" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
        return
    }

    accessToken, _ := jwtpkg.GenerateAccessToken(claims.UserID, h.cfg.JWTAccessSecret)
    c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "expires_in": 3600})
}
```

### 5.2 Rides handler

```go
// internal/handler/rides.go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/yourorg/trackride/internal/middleware"
    "github.com/yourorg/trackride/internal/service"
)

type RidesHandler struct {
    svc   *service.RidesService
    redis *redis.Client
    cfg   *config.Config
}

func (h *RidesHandler) Start(c *gin.Context) {
    var req struct{ VehicleID string `json:"vehicle_id" binding:"required,uuid"` }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR"})
        return
    }

    userID := middleware.GetUserID(c)
    result, err := h.svc.StartRide(c, userID, req.VehicleID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Generate ws_token di Redis (TTL 10 menit)
    wsToken := uuid.NewString()
    h.redis.SetEx(c, "ws_token:"+wsToken, userID+":"+result.ID.String(), 600*time.Second)

    c.JSON(http.StatusCreated, gin.H{
        "ride_id":    result.ID,
        "started_at": result.StartedAt,
        "ws_token":   wsToken,
    })
}

func (h *RidesHandler) Stop(c *gin.Context) {
    rideID := c.Param("id")
    userID := middleware.GetUserID(c)

    ride, err := h.svc.StopRide(c, rideID, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, ride)
}

func (h *RidesHandler) List(c *gin.Context) {
    userID := middleware.GetUserID(c)
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    vehicleType := c.Query("vehicle_type")

    rides, total, err := h.svc.ListRides(c, userID, vehicleType, page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"data": rides, "total": total, "page": page})
}
```

---

## 6. WebSocket — GPS Streaming

### 6.1 GPS point buffer

```go
// internal/websocket/buffer.go
package websocket

import (
    "context"
    "sync"
    "time"
    "github.com/yourorg/trackride/internal/db/sqlc"
)

const (
    batchSize     = 10
    flushInterval = 30 * time.Second
)

type GPSPoint struct {
    Lat       float64
    Lng       float64
    SpeedKmh  float64
    ElevM     float64
    Timestamp time.Time
}

type rideBuffer struct {
    points []GPSPoint
    timer  *time.Timer
    mu     sync.Mutex
}

type GPSBuffer struct {
    buffers map[string]*rideBuffer
    mu      sync.RWMutex
    queries *db.Queries
}

func NewGPSBuffer(queries *db.Queries) *GPSBuffer {
    return &GPSBuffer{
        buffers: make(map[string]*rideBuffer),
        queries: queries,
    }
}

func (b *GPSBuffer) Add(rideID string, point GPSPoint) {
    b.mu.Lock()
    buf, ok := b.buffers[rideID]
    if !ok {
        buf = &rideBuffer{}
        b.buffers[rideID] = buf
    }
    b.mu.Unlock()

    buf.mu.Lock()
    defer buf.mu.Unlock()

    buf.points = append(buf.points, point)

    if len(buf.points) >= batchSize {
        // Reset timer & flush
        if buf.timer != nil { buf.timer.Stop() }
        go b.flush(rideID)
        return
    }

    // Reschedule flush timer
    if buf.timer != nil { buf.timer.Stop() }
    buf.timer = time.AfterFunc(flushInterval, func() { b.flush(rideID) })
}

func (b *GPSBuffer) flush(rideID string) {
    b.mu.Lock()
    buf, ok := b.buffers[rideID]
    b.mu.Unlock()
    if !ok { return }

    buf.mu.Lock()
    if len(buf.points) == 0 { buf.mu.Unlock(); return }
    points := make([]GPSPoint, len(buf.points))
    copy(points, buf.points)
    buf.points = buf.points[:0]
    buf.mu.Unlock()

    // Batch insert ke PostgreSQL
    ctx := context.Background()
    rideUUID, _ := uuid.Parse(rideID)

    lats := make([]float64, len(points))
    lngs := make([]float64, len(points))
    speeds := make([]float64, len(points))
    elevs := make([]float64, len(points))
    times := make([]time.Time, len(points))
    ids := make([]uuid.UUID, len(points))

    for i, p := range points {
        ids[i] = rideUUID
        lats[i] = p.Lat
        lngs[i] = p.Lng
        speeds[i] = p.SpeedKmh
        elevs[i] = p.ElevM
        times[i] = p.Timestamp
    }

    _ = b.queries.InsertGPSPointsBatch(ctx, db.InsertGPSPointsBatchParams{
        RideIds:    ids,
        Latitudes:  lats,
        Longitudes: lngs,
        Speeds:     speeds,
        Elevations: elevs,
        Times:      times,
    })
}

// Panggil saat ride stop untuk flush sisa buffer
func (b *GPSBuffer) FlushAndClear(rideID string) {
    b.flush(rideID)
    b.mu.Lock()
    delete(b.buffers, rideID)
    b.mu.Unlock()
}
```

### 6.2 WebSocket hub

```go
// internal/websocket/hub.go
package websocket

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
    "github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

type Hub struct {
    buffer *GPSBuffer
    redis  *redis.Client
}

func NewHub(buffer *GPSBuffer, redis *redis.Client) *Hub {
    return &Hub{buffer: buffer, redis: redis}
}

type incomingMsg struct {
    Type      string  `json:"type"`
    Lat       float64 `json:"lat"`
    Lng       float64 `json:"lng"`
    SpeedKmh  float64 `json:"speed_kmh"`
    ElevM     float64 `json:"elevation_m"`
    Timestamp string  `json:"timestamp"`
}

func (h *Hub) HandleWS(c *gin.Context) {
    rideID := c.Param("id")
    wsToken := c.Query("token")

    // Validasi ws_token dari Redis
    val, err := h.redis.Get(c, "ws_token:"+wsToken).Result()
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
        return
    }
    // val = "userID:rideID"
    parts := strings.SplitN(val, ":", 2)
    if len(parts) != 2 || parts[1] != rideID {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
        return
    }

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Printf("WS upgrade error: %v", err)
        return
    }
    defer func() {
        conn.Close()
        h.buffer.FlushAndClear(rideID)
        log.Printf("WS disconnected: ride=%s", rideID)
    }()

    log.Printf("WS connected: ride=%s user=%s", rideID, parts[0])
    var pointCount int

    for {
        _, msg, err := conn.ReadMessage()
        if err != nil { break }

        var incoming incomingMsg
        if err := json.Unmarshal(msg, &incoming); err != nil { continue }

        switch incoming.Type {
        case "ping":
            _ = conn.WriteJSON(map[string]string{
                "type":        "pong",
                "server_time": time.Now().UTC().Format(time.RFC3339),
            })

        case "gps_point":
            // Validasi koordinat dasar
            if incoming.Lat < -90 || incoming.Lat > 90 || incoming.Lng < -180 || incoming.Lng > 180 {
                continue
            }

            ts, _ := time.Parse(time.RFC3339, incoming.Timestamp)
            if ts.IsZero() { ts = time.Now().UTC() }

            h.buffer.Add(rideID, GPSPoint{
                Lat: incoming.Lat, Lng: incoming.Lng,
                SpeedKmh: incoming.SpeedKmh, ElevM: incoming.ElevM,
                Timestamp: ts,
            })
            pointCount++

            _ = conn.WriteJSON(map[string]interface{}{
                "type":            "ack",
                "points_received": pointCount,
            })
        }
    }
}
```

---

## 7. Service Layer

### 7.1 Rides service — hitung metrik saat stop

```go
// internal/service/rides.go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "math"
    "github.com/yourorg/trackride/internal/db/sqlc"
    "github.com/yourorg/trackride/pkg/polyline"
)

type RidesService struct {
    queries *db.Queries
}

func (s *RidesService) StopRide(ctx context.Context, rideID, userID string) (*db.Ride, error) {
    ride, err := s.queries.GetActiveRide(ctx, db.GetActiveRideParams{ID: rideID, UserID: userID})
    if err != nil {
        return nil, fmt.Errorf("ACTIVE_RIDE_NOT_FOUND")
    }

    points, err := s.queries.GetGPSPointsByRide(ctx, ride.ID)
    if err != nil || len(points) < 2 {
        return s.queries.UpdateRideCompleted(ctx, db.UpdateRideCompletedParams{ID: ride.ID})
    }

    metrics := computeMetrics(points)
    summary := buildRouteSummary(points)
    summaryJSON, _ := json.Marshal(summary)

    return s.queries.UpdateRideCompleted(ctx, db.UpdateRideCompletedParams{
        ID:              ride.ID,
        DistanceKm:      metrics.DistanceKm,
        DurationSeconds: metrics.DurationSeconds,
        MaxSpeedKmh:     metrics.MaxSpeedKmh,
        AvgSpeedKmh:     metrics.AvgSpeedKmh,
        ElevationM:      metrics.ElevationM,
        Calories:        metrics.Calories,
        RouteSummary:    summaryJSON,
    })
}
```

### 7.2 Metrics computation

```go
// internal/service/metrics.go
package service

import (
    "math"
    "time"
    "github.com/yourorg/trackride/internal/db/sqlc"
    "github.com/yourorg/trackride/pkg/polyline"
)

type RideMetrics struct {
    DistanceKm      float64
    DurationSeconds int32
    MaxSpeedKmh     float64
    AvgSpeedKmh     float64
    ElevationM      float64
    Calories        int32
}

func computeMetrics(points []db.RideGpsPoint) RideMetrics {
    var totalDist, maxSpeed, totalElev float64

    for i := 1; i < len(points); i++ {
        prev, curr := points[i-1], points[i]
        totalDist += haversineKm(prev.Latitude, prev.Longitude, curr.Latitude, curr.Longitude)
        if curr.SpeedKmh > maxSpeed { maxSpeed = curr.SpeedKmh }
        gain := curr.ElevationM - prev.ElevationM
        if gain > 0 { totalElev += gain }
    }

    start := points[0].RecordedAt
    end := points[len(points)-1].RecordedAt
    durationSec := int32(end.Sub(start).Seconds())
    durationHours := float64(durationSec) / 3600.0
    avgSpeed := 0.0
    if durationHours > 0 { avgSpeed = totalDist / durationHours }

    durationMin := float64(durationSec) / 60.0
    calories := estimateCalories(totalDist, durationMin, 70)

    return RideMetrics{
        DistanceKm:      math.Round(totalDist*10) / 10,
        DurationSeconds: durationSec,
        MaxSpeedKmh:     math.Round(maxSpeed*10) / 10,
        AvgSpeedKmh:     math.Round(avgSpeed*10) / 10,
        ElevationM:      math.Round(totalElev),
        Calories:        int32(calories),
    }
}

// Haversine formula — jarak antara dua koordinat GPS dalam km
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
    const R = 6371.0
    dLat := (lat2 - lat1) * math.Pi / 180
    dLng := (lng2 - lng1) * math.Pi / 180
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
        math.Sin(dLng/2)*math.Sin(dLng/2)
    return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// Estimasi kalori sepeda — MET-based (Howarth et al., 2021)
func estimateCalories(distKm, durationMin, weightKg float64) float64 {
    if durationMin <= 0 { return 0 }
    avgSpeed := (distKm / durationMin) * 60
    met := 6.0
    if avgSpeed >= 20 { met = 10.0 } else if avgSpeed >= 15 { met = 8.0 }
    return met * weightKg * (durationMin / 60)
}
```

### 7.3 Route summary & polyline encoding

```go
// internal/service/metrics.go (lanjutan)

type RouteSummary struct {
    Polyline    string   `json:"polyline"`
    Cities      []string `json:"cities"`
    BoundingBox BBox     `json:"bounding_box"`
}

type BBox struct {
    North float64 `json:"north"`
    South float64 `json:"south"`
    East  float64 `json:"east"`
    West  float64 `json:"west"`
}

func buildRouteSummary(points []db.RideGpsPoint) RouteSummary {
    // Downsample ke max 500 titik untuk polyline
    sampled := downsample(points, 500)

    coords := make([][2]float64, len(sampled))
    for i, p := range sampled {
        coords[i] = [2]float64{p.Latitude, p.Longitude}
    }

    encoded := polyline.Encode(coords)

    bbox := BBox{
        North: -90, South: 90, East: -180, West: 180,
    }
    for _, p := range points {
        if p.Latitude > bbox.North  { bbox.North = p.Latitude }
        if p.Latitude < bbox.South  { bbox.South = p.Latitude }
        if p.Longitude > bbox.East  { bbox.East = p.Longitude }
        if p.Longitude < bbox.West  { bbox.West = p.Longitude }
    }

    return RouteSummary{Polyline: encoded, Cities: []string{}, BoundingBox: bbox}
}

func downsample[T any](s []T, max int) []T {
    if len(s) <= max { return s }
    factor := (len(s) + max - 1) / max
    out := make([]T, 0, max)
    for i, v := range s {
        if i%factor == 0 { out = append(out, v) }
    }
    return out
}
```

### 7.4 Google Encoded Polyline

```go
// pkg/polyline/polyline.go
package polyline

import "math"

// Encode array of [lat, lng] pairs ke Google Encoded Polyline format
func Encode(coords [][2]float64) string {
    var result []byte
    var prevLat, prevLng int

    for _, c := range coords {
        lat := int(math.Round(c[0] * 1e5))
        lng := int(math.Round(c[1] * 1e5))
        result = append(result, encodeValue(lat-prevLat)...)
        result = append(result, encodeValue(lng-prevLng)...)
        prevLat, prevLng = lat, lng
    }
    return string(result)
}

func encodeValue(v int) []byte {
    v <<= 1
    if v < 0 { v = ^v }
    var chunks []byte
    for v >= 0x20 {
        chunks = append(chunks, byte((0x20|(v&0x1f))+63))
        v >>= 5
    }
    chunks = append(chunks, byte(v+63))
    return chunks
}

// Decode encoded polyline string → array of [lat, lng]
func Decode(encoded string) [][2]float64 {
    var coords [][2]float64
    var lat, lng int
    i := 0

    decode := func() int {
        result, shift := 0, 0
        for {
            b := int(encoded[i]) - 63
            i++
            result |= (b & 0x1f) << shift
            shift += 5
            if b < 0x20 { break }
        }
        if result&1 != 0 { return ^(result >> 1) }
        return result >> 1
    }

    for i < len(encoded) {
        lat += decode()
        lng += decode()
        coords = append(coords, [2]float64{float64(lat) / 1e5, float64(lng) / 1e5})
    }
    return coords
}
```

---

## 8. Leaderboard Cron Job

```go
// internal/jobs/leaderboard.go
package jobs

import (
    "context"
    "log"
    "time"

    "github.com/robfig/cron/v3"
    "github.com/yourorg/trackride/internal/db/sqlc"
)

type LeaderboardJob struct {
    queries *db.Queries
}

func NewLeaderboardJob(queries *db.Queries) *LeaderboardJob {
    return &LeaderboardJob{queries: queries}
}

func (j *LeaderboardJob) Start() {
    c := cron.New(cron.WithLocation(mustLoadLocation("Asia/Jakarta")))
    // Setiap Senin 00:01 WIB
    c.AddFunc("1 0 * * MON", j.computeWeekly)
    c.Start()
    log.Println("[Leaderboard] Cron job started")
}

func (j *LeaderboardJob) computeWeekly() {
    ctx := context.Background()
    periodStart := lastMonday()
    log.Printf("[Leaderboard] Computing weekly rankings for %s", periodStart.Format("2006-01-02"))

    vehicleTypes := []string{"", "motor", "mobil", "sepeda"} // "" = semua kendaraan

    for _, vt := range vehicleTypes {
        // Hapus entries lama untuk periode ini
        j.queries.DeleteLeaderboardEntries(ctx, db.DeleteLeaderboardEntriesParams{
            PeriodType:  "weekly",
            PeriodStart: periodStart,
            VehicleType: vt,
        })

        // Hitung rankings
        rows, err := j.queries.ComputeWeeklyRankings(ctx, db.ComputeWeeklyRankingsParams{
            PeriodStart: periodStart,
            VehicleType: vt,
        })
        if err != nil {
            log.Printf("[Leaderboard] Error computing vt=%s: %v", vt, err)
            continue
        }

        // Insert entries baru
        for rank, row := range rows {
            j.queries.InsertLeaderboardEntry(ctx, db.InsertLeaderboardEntryParams{
                UserID:      row.UserID,
                VehicleType: vt,
                PeriodType:  "weekly",
                PeriodStart: periodStart,
                TotalKm:     row.TotalKm,
                TotalRides:  row.TotalRides,
                Rank:        int32(rank + 1),
            })
        }
    }

    log.Println("[Leaderboard] Done")
}

func lastMonday() time.Time {
    now := time.Now().In(mustLoadLocation("Asia/Jakarta"))
    daysBack := int(now.Weekday()) - int(time.Monday)
    if daysBack < 0 { daysBack += 7 }
    d := now.AddDate(0, 0, -daysBack)
    return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
}

func mustLoadLocation(name string) *time.Location {
    loc, err := time.LoadLocation(name)
    if err != nil { return time.UTC }
    return loc
}
```

---

## 9. Push Notifications

```go
// internal/service/notifications.go
package service

import (
    "bytes"
    "encoding/json"
    "net/http"
)

// Menggunakan Expo Push API — tidak perlu setup FCM/APNs langsung
func SendPushNotification(pushToken, title, body string) error {
    payload := map[string]interface{}{
        "to":    pushToken,
        "title": title,
        "body":  body,
        "sound": "default",
        "badge": 1,
    }
    b, _ := json.Marshal(payload)

    resp, err := http.Post(
        "https://exp.host/--/api/v2/push/send",
        "application/json",
        bytes.NewReader(b),
    )
    if err != nil { return err }
    defer resp.Body.Close()
    return nil
}
```

---

## 10. Error Response Standard

Semua error menggunakan format:

```json
{ "error": "ERROR_CODE" }
```

| Code | HTTP | Situasi |
|---|---|---|
| `UNAUTHORIZED` | 401 | Token tidak ada |
| `TOKEN_EXPIRED` | 401 | Access token expired |
| `TOKEN_INVALID` | 401 | Refresh token invalid |
| `FORBIDDEN` | 403 | Tidak punya akses |
| `NOT_FOUND` | 404 | Resource tidak ditemukan |
| `EMAIL_TAKEN` | 409 | Email sudah terdaftar |
| `USERNAME_TAKEN` | 409 | Username sudah dipakai |
| `VALIDATION_ERROR` | 422 | Input tidak valid |
| `VEHICLE_NOT_FOUND` | 404 | Kendaraan tidak ditemukan |
| `ACTIVE_RIDE_NOT_FOUND` | 404 | Tidak ada ride aktif |
| `INTERNAL_ERROR` | 500 | Server error |

---

## 11. Performance Notes

**`ride_gps_points` tumbuh cepat** — estimasi 720.000 baris/hari untuk 1.000 user aktif. Mitigasi:

- Downsampling ke max 500 titik per ride sudah dilakukan di `buildRouteSummary` — polyline disimpan di `route_summary` JSONB
- `ride_gps_points` hanya diakses saat `POST /rides/:id/stop` untuk komputasi metrik, lalu jarang disentuh lagi
- Pertimbangkan table partitioning by month saat user > 10k: `PARTITION BY RANGE (recorded_at)`

**Leaderboard** — tidak pernah dihitung on-demand. Selalu dari `leaderboard_entries` yang sudah pre-computed.

**WebSocket** — satu goroutine per koneksi. Go handles ini dengan sangat efisien; 10.000 koneksi concurrent tidak masalah dengan `goroutine`.

---

## 12. Open Questions (Backend)

| # | Pertanyaan | Impact |
|---|---|---|
| 1 | Reverse geocoding (kota dilintasi) pakai Google Maps API atau Mapbox? | `geocoding.go` |
| 2 | Redis wajib dari awal, atau `sync.Map` in-memory untuk ws_token MVP? | Deployment complexity |
| 3 | Push notification synchronous atau background goroutine? | Latency `POST /rides/:id/like` |
| 4 | Deployment: single binary di VPS, atau Docker + Compose? | `cmd/server/main.go` setup |
| 5 | GPS points retention policy? | Storage planning jangka panjang |

---

*TrackRide Backend Guide v1.0 — Last updated: April 2026*
*Untuk spesifikasi frontend/mobile, lihat `CLAUDE-frontend.md`*
