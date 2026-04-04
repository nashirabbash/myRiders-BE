-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, display_name)
VALUES ($1, $2, $3, $4)
RETURNING id, username, email, display_name, avatar_url, push_token, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, username, email, password_hash, display_name, avatar_url, push_token, created_at, updated_at
FROM users
WHERE LOWER(email) = LOWER($1);

-- name: GetUserByUsername :one
SELECT id, username, email, display_name, avatar_url, push_token, created_at, updated_at
FROM users
WHERE LOWER(username) = LOWER($1);

-- name: GetUserByID :one
SELECT id, username, email, display_name, avatar_url, push_token, created_at, updated_at
FROM users
WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET display_name = COALESCE($2, display_name),
    avatar_url = COALESCE($3, avatar_url),
    push_token = COALESCE($4, push_token),
    updated_at = NOW()
WHERE id = $1
RETURNING id, username, email, display_name, avatar_url, push_token, created_at, updated_at;

-- name: UpdateUserPushToken :exec
UPDATE users
SET push_token = $2, updated_at = NOW()
WHERE id = $1;
