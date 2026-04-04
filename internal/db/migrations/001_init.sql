-- Create extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create custom types
CREATE TYPE vehicle_type AS ENUM ('motor', 'mobil', 'sepeda');
CREATE TYPE ride_status AS ENUM ('active', 'completed', 'abandoned');

-- Users table
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username     TEXT NOT NULL,
    email        TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    display_name TEXT NOT NULL,
    avatar_url   TEXT,
    push_token   TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Vehicles table
CREATE TABLE vehicles (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type      vehicle_type NOT NULL,
    name      TEXT NOT NULL,
    brand     TEXT,
    color     TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rides table
CREATE TABLE rides (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vehicle_id       UUID NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT,
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
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- GPS points table (high-frequency data)
CREATE TABLE ride_gps_points (
    id          BIGSERIAL PRIMARY KEY,
    ride_id     UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    latitude    FLOAT8 NOT NULL,
    longitude   FLOAT8 NOT NULL,
    speed_kmh   FLOAT8 NOT NULL DEFAULT 0,
    elevation_m FLOAT8 NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Follows table (user-to-user relationships)
CREATE TABLE follows (
    follower_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, following_id),
    CONSTRAINT no_self_follow CHECK (follower_id != following_id)
);

-- Ride likes table
CREATE TABLE ride_likes (
    ride_id    UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (ride_id, user_id)
);

-- Ride comments table
CREATE TABLE ride_comments (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ride_id    UUID NOT NULL REFERENCES rides(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    VARCHAR(280) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Leaderboard entries table (pre-computed rankings)
CREATE TABLE leaderboard_entries (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vehicle_type vehicle_type,
    period_type  TEXT NOT NULL,
    period_start DATE NOT NULL,
    total_km     FLOAT8 NOT NULL DEFAULT 0,
    total_rides  INT NOT NULL DEFAULT 0,
    rank         INT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
