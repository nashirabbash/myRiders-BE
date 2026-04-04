-- Follows Queries

-- name: FollowUser :exec
INSERT INTO follows (follower_id, following_id)
VALUES ($1, $2)
ON CONFLICT (follower_id, following_id) DO NOTHING;

-- name: UnfollowUser :exec
DELETE FROM follows
WHERE follower_id = $1 AND following_id = $2;

-- name: IsFollowing :one
SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = $2) as is_following;

-- name: GetFollowerCount :one
SELECT COUNT(*) as count
FROM follows
WHERE following_id = $1;

-- name: GetFollowingCount :one
SELECT COUNT(*) as count
FROM follows
WHERE follower_id = $1;

-- name: ListFollowers :many
SELECT u.id, u.username, u.display_name, u.avatar_url
FROM follows f
JOIN users u ON f.follower_id = u.id
WHERE f.following_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListFollowing :many
SELECT u.id, u.username, u.display_name, u.avatar_url
FROM follows f
JOIN users u ON f.following_id = u.id
WHERE f.follower_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- Likes Queries

-- name: LikeRide :exec
INSERT INTO ride_likes (ride_id, user_id)
VALUES ($1, $2)
ON CONFLICT (ride_id, user_id) DO NOTHING;

-- name: UnlikeRide :exec
DELETE FROM ride_likes
WHERE ride_id = $1 AND user_id = $2;

-- name: HasUserLikedRide :one
SELECT EXISTS(SELECT 1 FROM ride_likes WHERE ride_id = $1 AND user_id = $2) as has_liked;

-- name: GetRideLikeCount :one
SELECT COUNT(*) as count
FROM ride_likes
WHERE ride_id = $1;

-- Comments Queries

-- name: CreateComment :one
INSERT INTO ride_comments (ride_id, user_id, content)
VALUES ($1, $2, $3)
RETURNING id, ride_id, user_id, content, created_at, updated_at;

-- name: GetCommentByID :one
SELECT id, ride_id, user_id, content, created_at, updated_at
FROM ride_comments
WHERE id = $1;

-- name: ListRideComments :many
SELECT id, ride_id, user_id, content, created_at, updated_at
FROM ride_comments
WHERE ride_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRideCommentCount :one
SELECT COUNT(*) as count
FROM ride_comments
WHERE ride_id = $1;

-- name: UpdateComment :one
UPDATE ride_comments
SET content = $2, updated_at = NOW()
WHERE id = $1 AND user_id = $3
RETURNING id, ride_id, user_id, content, created_at, updated_at;

-- name: DeleteComment :exec
DELETE FROM ride_comments
WHERE id = $1 AND user_id = $2;

-- Feed Query (recent rides from followed users)

-- name: GetFollowingFeed :many
SELECT
    r.id, r.user_id, r.vehicle_id, r.title, r.started_at, r.ended_at,
    r.distance_km, r.duration_seconds, r.max_speed_kmh, r.avg_speed_kmh,
    r.elevation_m, r.calories, r.route_summary, r.status,
    r.created_at, r.updated_at,
    u.username, u.display_name, u.avatar_url,
    v.type as vehicle_type, v.name as vehicle_name,
    COALESCE(COUNT(DISTINCT rl.user_id), 0) as like_count,
    COALESCE(COUNT(DISTINCT rc.id), 0) as comment_count
FROM rides r
JOIN users u ON r.user_id = u.id
JOIN vehicles v ON r.vehicle_id = v.id
LEFT JOIN ride_likes rl ON r.id = rl.ride_id
LEFT JOIN ride_comments rc ON r.id = rc.ride_id
WHERE r.status = 'completed' AND EXISTS(
    SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = r.user_id
)
GROUP BY r.id, u.id, v.id
ORDER BY r.started_at DESC
LIMIT $2 OFFSET $3;

-- Get feed with user's like status

-- name: GetFollowingFeedWithUserStatus :many
SELECT
    r.id, r.user_id, r.vehicle_id, r.title, r.started_at, r.ended_at,
    r.distance_km, r.duration_seconds, r.max_speed_kmh, r.avg_speed_kmh,
    r.elevation_m, r.calories, r.route_summary, r.status,
    r.created_at, r.updated_at,
    u.username, u.display_name, u.avatar_url,
    v.type as vehicle_type, v.name as vehicle_name,
    COALESCE(COUNT(DISTINCT rl.user_id), 0) as like_count,
    COALESCE(COUNT(DISTINCT rc.id), 0) as comment_count,
    EXISTS(SELECT 1 FROM ride_likes WHERE ride_id = r.id AND ride_likes.user_id = $1) as user_has_liked
FROM rides r
JOIN users u ON r.user_id = u.id
JOIN vehicles v ON r.vehicle_id = v.id
LEFT JOIN ride_likes rl ON r.id = rl.ride_id
LEFT JOIN ride_comments rc ON r.id = rc.ride_id
WHERE r.status = 'completed' AND EXISTS(
    SELECT 1 FROM follows WHERE follower_id = $1 AND following_id = r.user_id
)
GROUP BY r.id, u.id, v.id
ORDER BY r.started_at DESC
LIMIT $2 OFFSET $3;
