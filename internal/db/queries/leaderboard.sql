-- name: InsertLeaderboardEntry :exec
INSERT INTO leaderboard_entries (user_id, vehicle_type, period_type, period_start, total_km, total_rides, rank)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: DeleteLeaderboardEntries :exec
DELETE FROM leaderboard_entries
WHERE period_type = $1 AND period_start = $2
  AND (vehicle_type IS NULL OR vehicle_type = $3);

-- name: GetLeaderboardByPeriod :many
SELECT id, user_id, vehicle_type, period_type, period_start, total_km, total_rides, rank, created_at, updated_at
FROM leaderboard_entries
WHERE period_type = $1 AND period_start = $2
ORDER BY rank ASC
LIMIT $3 OFFSET $4;

-- name: GetLeaderboardByPeriodAndVehicle :many
SELECT id, user_id, vehicle_type, period_type, period_start, total_km, total_rides, rank, created_at, updated_at
FROM leaderboard_entries
WHERE period_type = $1 AND period_start = $2 AND vehicle_type = $3
ORDER BY rank ASC
LIMIT $4 OFFSET $5;

-- name: GetLeaderboardEntryByRank :one
SELECT id, user_id, vehicle_type, period_type, period_start, total_km, total_rides, rank, created_at, updated_at
FROM leaderboard_entries
WHERE period_type = $1 AND period_start = $2 AND rank = $3;

-- name: GetUserLeaderboardRank :one
SELECT id, user_id, vehicle_type, period_type, period_start, total_km, total_rides, rank, created_at, updated_at
FROM leaderboard_entries
WHERE period_type = $1 AND period_start = $2 AND user_id = $3 AND (vehicle_type IS NULL OR vehicle_type = $4)
LIMIT 1;

-- name: GetUserRankingHistory :many
SELECT id, user_id, vehicle_type, period_type, period_start, total_km, total_rides, rank, created_at, updated_at
FROM leaderboard_entries
WHERE user_id = $1 AND period_type = $2
ORDER BY period_start DESC
LIMIT $3;

-- name: GetLeaderboardStats :one
SELECT
    COUNT(DISTINCT user_id) as unique_users,
    COALESCE(SUM(total_km), 0) as total_platform_km,
    COALESCE(SUM(total_rides), 0) as total_platform_rides,
    COALESCE(MAX(rank), 0) as max_rank
FROM leaderboard_entries
WHERE period_type = $1 AND period_start = $2;

-- Compute rankings for current period (used by cron job)

-- name: ComputeWeeklyRankings :many
SELECT
    u.id as user_id,
    COALESCE(SUM(r.distance_km), 0) as total_km,
    COUNT(DISTINCT r.id) as total_rides
FROM users u
LEFT JOIN rides r ON u.id = r.user_id AND r.status = 'completed'
    AND r.started_at >= $1 AND r.started_at < $1 + INTERVAL '7 days'
GROUP BY u.id
HAVING COUNT(DISTINCT r.id) > 0 OR SUM(r.distance_km) > 0
ORDER BY total_km DESC, total_rides DESC;

-- name: ComputeWeeklyRankingsByVehicle :many
SELECT
    u.id as user_id,
    v.type as vehicle_type,
    COALESCE(SUM(r.distance_km), 0) as total_km,
    COUNT(DISTINCT r.id) as total_rides
FROM users u
LEFT JOIN rides r ON u.id = r.user_id AND r.status = 'completed'
    AND r.started_at >= $1 AND r.started_at < $1 + INTERVAL '7 days'
LEFT JOIN vehicles v ON r.vehicle_id = v.id
WHERE v.type = $2
GROUP BY u.id, v.type
HAVING COUNT(DISTINCT r.id) > 0 OR SUM(r.distance_km) > 0
ORDER BY total_km DESC, total_rides DESC;

-- name: ComputeMonthlyRankings :many
SELECT
    u.id as user_id,
    COALESCE(SUM(r.distance_km), 0) as total_km,
    COUNT(DISTINCT r.id) as total_rides
FROM users u
LEFT JOIN rides r ON u.id = r.user_id AND r.status = 'completed'
    AND r.started_at >= $1 AND r.started_at < $1 + INTERVAL '1 month'
GROUP BY u.id
HAVING COUNT(DISTINCT r.id) > 0 OR SUM(r.distance_km) > 0
ORDER BY total_km DESC, total_rides DESC;

-- name: ComputeMonthlyRankingsByVehicle :many
SELECT
    u.id as user_id,
    v.type as vehicle_type,
    COALESCE(SUM(r.distance_km), 0) as total_km,
    COUNT(DISTINCT r.id) as total_rides
FROM users u
LEFT JOIN rides r ON u.id = r.user_id AND r.status = 'completed'
    AND r.started_at >= $1 AND r.started_at < $1 + INTERVAL '1 month'
LEFT JOIN vehicles v ON r.vehicle_id = v.id
WHERE v.type = $2
GROUP BY u.id, v.type
HAVING COUNT(DISTINCT r.id) > 0 OR SUM(r.distance_km) > 0
ORDER BY total_km DESC, total_rides DESC;

-- name: ComputeAllTimeRankings :many
SELECT
    u.id as user_id,
    COALESCE(SUM(r.distance_km), 0) as total_km,
    COUNT(DISTINCT r.id) as total_rides
FROM users u
LEFT JOIN rides r ON u.id = r.user_id AND r.status = 'completed'
GROUP BY u.id
HAVING COUNT(DISTINCT r.id) > 0 OR SUM(r.distance_km) > 0
ORDER BY total_km DESC, total_rides DESC;

-- name: ComputeAllTimeRankingsByVehicle :many
SELECT
    u.id as user_id,
    v.type as vehicle_type,
    COALESCE(SUM(r.distance_km), 0) as total_km,
    COUNT(DISTINCT r.id) as total_rides
FROM users u
LEFT JOIN rides r ON u.id = r.user_id AND r.status = 'completed'
LEFT JOIN vehicles v ON r.vehicle_id = v.id
WHERE v.type = $1
GROUP BY u.id, v.type
HAVING COUNT(DISTINCT r.id) > 0 OR SUM(r.distance_km) > 0
ORDER BY total_km DESC, total_rides DESC;
