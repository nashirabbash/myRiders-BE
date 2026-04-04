-- name: CreateVehicle :one
INSERT INTO vehicles (user_id, type, name, brand, color)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, type, name, brand, color, is_active, created_at, updated_at;

-- name: GetVehicleByID :one
SELECT id, user_id, type, name, brand, color, is_active, created_at, updated_at
FROM vehicles
WHERE id = $1;

-- name: ListVehiclesByUser :many
SELECT id, user_id, type, name, brand, color, is_active, created_at, updated_at
FROM vehicles
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListActiveVehiclesByUser :many
SELECT id, user_id, type, name, brand, color, is_active, created_at, updated_at
FROM vehicles
WHERE user_id = $1 AND is_active = TRUE
ORDER BY created_at DESC;

-- name: UpdateVehicle :one
UPDATE vehicles
SET type = COALESCE($3, type),
    name = COALESCE($4, name),
    brand = COALESCE($5, brand),
    color = COALESCE($6, color),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING id, user_id, type, name, brand, color, is_active, created_at, updated_at;

-- name: DeactivateVehicle :exec
UPDATE vehicles
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: DeleteVehicle :exec
DELETE FROM vehicles
WHERE id = $1 AND user_id = $2;

-- name: HasActiveRide :one
SELECT EXISTS(
    SELECT 1 FROM rides
    WHERE vehicle_id = $1 AND status = 'active'
) as has_active_ride;
