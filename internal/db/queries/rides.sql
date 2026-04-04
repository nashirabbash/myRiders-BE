-- name: CreateRide :one
INSERT INTO rides (user_id, vehicle_id, started_at, status)
VALUES ($1, $2, $3, 'active')
RETURNING id, user_id, vehicle_id, title, started_at, ended_at, distance_km, duration_seconds, max_speed_kmh, avg_speed_kmh, elevation_m, calories, route_summary, status, created_at, updated_at;

-- name: GetRideByID :one
SELECT id, user_id, vehicle_id, title, started_at, ended_at, distance_km, duration_seconds, max_speed_kmh, avg_speed_kmh, elevation_m, calories, route_summary, status, created_at, updated_at
FROM rides
WHERE id = $1;

-- name: GetActiveRide :one
SELECT id, user_id, vehicle_id, title, started_at, ended_at, distance_km, duration_seconds, max_speed_kmh, avg_speed_kmh, elevation_m, calories, route_summary, status, created_at, updated_at
FROM rides
WHERE id = $1 AND user_id = $2 AND status = 'active';

-- name: UpdateRideCompleted :one
UPDATE rides
SET ended_at = $2,
    status = 'completed',
    distance_km = $3,
    duration_seconds = $4,
    max_speed_kmh = $5,
    avg_speed_kmh = $6,
    elevation_m = $7,
    calories = $8,
    route_summary = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING id, user_id, vehicle_id, title, started_at, ended_at, distance_km, duration_seconds, max_speed_kmh, avg_speed_kmh, elevation_m, calories, route_summary, status, created_at, updated_at;

-- name: AbandonRide :one
UPDATE rides
SET status = 'abandoned', ended_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, vehicle_id, title, started_at, ended_at, distance_km, duration_seconds, max_speed_kmh, avg_speed_kmh, elevation_m, calories, route_summary, status, created_at, updated_at;

-- name: ListRidesByUser :many
SELECT id, user_id, vehicle_id, title, started_at, ended_at, distance_km, duration_seconds, max_speed_kmh, avg_speed_kmh, elevation_m, calories, route_summary, status, created_at, updated_at
FROM rides
WHERE user_id = $1 AND status = 'completed'
ORDER BY started_at DESC
LIMIT $2 OFFSET $3;

-- name: ListRidesByUserAndVehicle :many
SELECT id, user_id, vehicle_id, title, started_at, ended_at, distance_km, duration_seconds, max_speed_kmh, avg_speed_kmh, elevation_m, calories, route_summary, status, created_at, updated_at
FROM rides
WHERE user_id = $1 AND vehicle_id = $2 AND status = 'completed'
ORDER BY started_at DESC
LIMIT $3 OFFSET $4;

-- name: GetRideCount :one
SELECT COUNT(*) as count
FROM rides
WHERE user_id = $1 AND status = 'completed';

-- name: GetUserRideStats :one
SELECT
    COUNT(*) as total_rides,
    COALESCE(SUM(distance_km), 0) as total_distance_km,
    COALESCE(SUM(calories), 0) as total_calories,
    COALESCE(SUM(elevation_m), 0) as total_elevation_m,
    COALESCE(SUM(EXTRACT(EPOCH FROM (ended_at - started_at))), 0)::INT as total_seconds
FROM rides
WHERE user_id = $1 AND status = 'completed';

-- GPS Points Queries

-- name: InsertGPSPoint :exec
INSERT INTO ride_gps_points (ride_id, latitude, longitude, speed_kmh, elevation_m, recorded_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetGPSPointsByRide :many
SELECT id, ride_id, latitude, longitude, speed_kmh, elevation_m, recorded_at, created_at
FROM ride_gps_points
WHERE ride_id = $1
ORDER BY recorded_at ASC;

-- name: GetGPSPointCount :one
SELECT COUNT(*) as count
FROM ride_gps_points
WHERE ride_id = $1;

-- name: DeleteGPSPointsByRide :exec
DELETE FROM ride_gps_points
WHERE ride_id = $1;
