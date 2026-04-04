package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
)

// RidesService handles ride business logic
type RidesService struct {
	queries *sqlc.Queries
}

// NewRidesService creates a new rides service
func NewRidesService(queries *sqlc.Queries) *RidesService {
	return &RidesService{queries: queries}
}

// StartRide creates a new ride and returns the ride and ws_token
func (s *RidesService) StartRide(ctx context.Context, userID, vehicleID, title string) (*sqlc.Ride, string, error) {
	// Parse UUIDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, "", fmt.Errorf("INVALID_USER_ID")
	}
	vehicleUUID, err := uuid.Parse(vehicleID)
	if err != nil {
		return nil, "", fmt.Errorf("INVALID_VEHICLE_ID")
	}

	userPGType := pgtype.UUID{Bytes: userUUID, Valid: true}
	vehiclePGType := pgtype.UUID{Bytes: vehicleUUID, Valid: true}

	// Validate vehicle exists and belongs to user
	vehicle, err := s.queries.GetVehicleByID(ctx, vehiclePGType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", fmt.Errorf("VEHICLE_NOT_FOUND")
		}
		return nil, "", fmt.Errorf("INTERNAL_ERROR")
	}

	// Check if vehicle belongs to user
	if vehicle.UserID.String() != userID {
		return nil, "", fmt.Errorf("FORBIDDEN")
	}

	// Create new ride
	ride, err := s.queries.CreateRide(ctx, sqlc.CreateRideParams{
		UserID:    userPGType,
		VehicleID: vehiclePGType,
		StartedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	})
	if err != nil {
		return nil, "", fmt.Errorf("INTERNAL_ERROR")
	}

	// Generate ws_token
	wsToken := uuid.NewString()

	return &ride, wsToken, nil
}

// StopRide completes an active ride and computes metrics
// TODO: Wrap the entire operation in a database transaction with SELECT...FOR UPDATE
// to ensure atomicity and prevent concurrent stop requests from creating duplicate metrics
func (s *RidesService) StopRide(ctx context.Context, rideID, userID string) (*sqlc.Ride, error) {
	rideUUID, err := uuid.Parse(rideID)
	if err != nil {
		return nil, fmt.Errorf("INVALID_RIDE_ID")
	}
	rideUUIDPGType := pgtype.UUID{Bytes: rideUUID, Valid: true}

	// Get ride with FOR UPDATE lock to prevent concurrent modifications
	// Note: This requires transaction support to be fully effective
	ride, err := s.queries.GetRideForUpdate(ctx, rideUUIDPGType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("RIDE_NOT_FOUND")
		}
		return nil, fmt.Errorf("INTERNAL_ERROR")
	}

	// Verify ride belongs to user
	if ride.UserID.String() != userID {
		return nil, fmt.Errorf("FORBIDDEN")
	}

	// Verify ride is active before processing
	if ride.Status != "active" {
		return nil, fmt.Errorf("RIDE_NOT_ACTIVE")
	}

	// Get all GPS points for the ride
	points, err := s.queries.GetGPSPointsByRide(ctx, rideUUIDPGType)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("INTERNAL_ERROR")
	}

	// TODO: Fetch user weight from user profile for accurate calorie calculation
	// For now, using default 70kg. Update when user weight field is added to database.
	const defaultUserWeightKg = 70.0

	// Compute metrics from GPS points
	var metrics MetricsResult
	endedAt := time.Now().UTC()

	if len(points) >= 2 {
		metrics = ComputeMetrics(points, defaultUserWeightKg)
		// Build route summary
		summary := BuildRouteSummary(points)
		summaryJSON, _ := json.Marshal(summary)

		// Update ride with computed metrics
		ride, err = s.queries.UpdateRideCompleted(ctx, sqlc.UpdateRideCompletedParams{
			ID:              rideUUIDPGType,
			EndedAt:         pgtype.Timestamptz{Time: endedAt, Valid: true},
			DistanceKm:      metrics.DistanceKm,
			DurationSeconds: metrics.DurationSeconds,
			MaxSpeedKmh:     metrics.MaxSpeedKmh,
			AvgSpeedKmh:     metrics.AvgSpeedKmh,
			ElevationM:      metrics.ElevationM,
			Calories:        metrics.Calories,
			RouteSummary:    summaryJSON,
		})
	} else {
		// Not enough GPS points, just mark as completed with zero metrics
		ride, err = s.queries.UpdateRideCompleted(ctx, sqlc.UpdateRideCompletedParams{
			ID:      rideUUIDPGType,
			EndedAt: pgtype.Timestamptz{Time: endedAt, Valid: true},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("INTERNAL_ERROR")
	}

	// TODO: Resource Cleanup Strategy
	// Consider implementing archival/purge of raw GPS points after ride completion:
	// - Keep points for 30 days for debugging/audit purposes
	// - Archive historical points to a separate table or S3
	// - Implement incremental cleanup job to prevent ride_gps_points table growth
	// - Keep summary polyline in route_summary JSONB for long-term retention

	return &ride, nil
}

// ListRides lists completed rides for a user with pagination
func (s *RidesService) ListRides(ctx context.Context, userID string, vehicleType string, page, limit int) ([]sqlc.Ride, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("INVALID_USER_ID")
	}
	userPGType := pgtype.UUID{Bytes: userUUID, Valid: true}
	offset := int32((page - 1) * limit)

	var rides []sqlc.Ride
	var total int64

	// Get all completed rides
	rides, err = s.queries.ListRidesByUser(ctx, sqlc.ListRidesByUserParams{
		UserID: userPGType,
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil && err != pgx.ErrNoRows {
		return nil, 0, fmt.Errorf("INTERNAL_ERROR")
	}

	// Get total count
	total, err = s.queries.GetRideCount(ctx, userPGType)
	if err != nil {
		return nil, 0, fmt.Errorf("INTERNAL_ERROR")
	}

	if rides == nil {
		rides = []sqlc.Ride{}
	}

	return rides, total, nil
}

// GetRideByID retrieves a specific ride ensuring user can access it
func (s *RidesService) GetRideByID(ctx context.Context, rideID, userID string) (*sqlc.Ride, error) {
	rideUUID, err := uuid.Parse(rideID)
	if err != nil {
		return nil, fmt.Errorf("INVALID_RIDE_ID")
	}
	rideUUIDPGType := pgtype.UUID{Bytes: rideUUID, Valid: true}

	ride, err := s.queries.GetRideByID(ctx, rideUUIDPGType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("RIDE_NOT_FOUND")
		}
		return nil, fmt.Errorf("INTERNAL_ERROR")
	}

	// Verify user owns the ride
	if ride.UserID.String() != userID {
		return nil, fmt.Errorf("FORBIDDEN")
	}

	return &ride, nil
}
