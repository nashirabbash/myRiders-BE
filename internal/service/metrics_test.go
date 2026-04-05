package service

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
)

func createGPSPoint(lat, lng, speed, elev float64, timestamp time.Time) sqlc.RideGpsPoint {
	return sqlc.RideGpsPoint{
		ID:         0,
		RideID:     pgtype.UUID{Bytes: uuid.Nil, Valid: true},
		Latitude:   lat,
		Longitude:  lng,
		SpeedKmh:   speed,
		ElevationM: elev,
		RecordedAt: pgtype.Timestamptz{Time: timestamp, Valid: true},
		CreatedAt:  pgtype.Timestamptz{Time: timestamp, Valid: true},
	}
}

func TestComputeMetrics_Empty(t *testing.T) {
	points := []sqlc.RideGpsPoint{}
	result := ComputeMetrics(points, 70)

	if result.DistanceKm != 0 || result.DurationSeconds != 0 {
		t.Errorf("Empty points should result in zero metrics")
	}
}

func TestComputeMetrics_SinglePoint(t *testing.T) {
	now := time.Now()
	points := []sqlc.RideGpsPoint{
		createGPSPoint(0, 0, 0, 0, now),
	}

	result := ComputeMetrics(points, 70)
	if result.DistanceKm != 0 || result.DurationSeconds != 0 {
		t.Errorf("Single point should result in zero metrics")
	}
}

func TestComputeMetrics_TwoPoints(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	// Two points 1 degree apart (approximately 111 km)
	points := []sqlc.RideGpsPoint{
		createGPSPoint(0, 0, 10, 0, startTime),
		createGPSPoint(1, 0, 15, 0, endTime),
	}

	result := ComputeMetrics(points, 70)

	// Distance should be approximately 111 km
	if result.DistanceKm < 100 || result.DistanceKm > 120 {
		t.Errorf("Expected distance ~111 km, got %.2f km", result.DistanceKm)
	}

	// Duration should be 3600 seconds (1 hour)
	if result.DurationSeconds != 3600 {
		t.Errorf("Expected duration 3600 seconds, got %d", result.DurationSeconds)
	}

	// Max speed should be 15 km/h
	if result.MaxSpeedKmh != 15 {
		t.Errorf("Expected max speed 15 km/h, got %.2f", result.MaxSpeedKmh)
	}

	// Average speed should be ~111 km/h
	if result.AvgSpeedKmh < 100 || result.AvgSpeedKmh > 120 {
		t.Errorf("Expected avg speed ~111 km/h, got %.2f", result.AvgSpeedKmh)
	}
}

func TestComputeMetrics_ElevationGain(t *testing.T) {
	startTime := time.Now()

	points := []sqlc.RideGpsPoint{
		createGPSPoint(0, 0, 10, 0, startTime),
		createGPSPoint(0.001, 0, 10, 50, startTime.Add(1*time.Minute)),
		createGPSPoint(0.002, 0, 10, 100, startTime.Add(2*time.Minute)),
		createGPSPoint(0.003, 0, 10, 80, startTime.Add(3*time.Minute)), // Elevation loss
	}

	result := ComputeMetrics(points, 70)

	// Only count elevation gains, not losses
	// 0->50 = 50m, 50->100 = 50m, 100->80 = -20m (not counted)
	// Total = 100m
	if result.ElevationM != 100 {
		t.Errorf("Expected elevation gain 100m, got %.0f", result.ElevationM)
	}
}

func TestComputeMetrics_Calories(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	// Simple 10 km ride in 1 hour = 10 km/h average speed
	// At 10 km/h (< 13 km/h), MET is 4.0
	// Calories = 4.0 * 70 kg * 1 hour = 280 calories
	points := []sqlc.RideGpsPoint{
		createGPSPoint(0, 0, 8, 0, startTime),
		createGPSPoint(0.089, 0, 12, 0, endTime),
	}

	result := ComputeMetrics(points, 70)

	// Should be approximately 280 calories (MET 4.0 * 70 kg * 1 hour)
	if result.Calories < 260 || result.Calories > 300 {
		t.Errorf("Expected calories ~280, got %d", result.Calories)
	}
}

func TestComputeMetrics_DefaultWeight(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(1 * time.Hour)

	points := []sqlc.RideGpsPoint{
		createGPSPoint(0, 0, 10, 0, startTime),
		createGPSPoint(0.089, 0, 10, 0, endTime),
	}

	result := ComputeMetrics(points, 0) // Default weight

	// Should compute calories with default 70kg weight
	if result.Calories == 0 {
		t.Errorf("Should compute calories even with zero weight (uses default 70kg)")
	}
}

func TestBuildRouteSummary_Empty(t *testing.T) {
	points := []sqlc.RideGpsPoint{}
	summary := BuildRouteSummary(points)

	if summary.Polyline != "" {
		t.Errorf("Expected empty polyline for no points")
	}

	bbox := summary.BoundingBox
	if bbox.North != 0 || bbox.South != 0 || bbox.East != 0 || bbox.West != 0 {
		t.Errorf("Expected zero bounding box for no points")
	}
}

func TestBuildRouteSummary_SinglePoint(t *testing.T) {
	now := time.Now()
	points := []sqlc.RideGpsPoint{
		createGPSPoint(37.5, -122.5, 10, 0, now),
	}

	summary := BuildRouteSummary(points)

	bbox := summary.BoundingBox
	if bbox.North != 37.5 || bbox.South != 37.5 {
		t.Errorf("Expected bounding box bounds at 37.5, got North: %f, South: %f", bbox.North, bbox.South)
	}
}

func TestBuildRouteSummary_MultiplePoints(t *testing.T) {
	now := time.Now()
	points := []sqlc.RideGpsPoint{
		createGPSPoint(37.0, -122.0, 10, 0, now),
		createGPSPoint(38.0, -121.0, 10, 0, now.Add(1*time.Minute)),
		createGPSPoint(36.0, -123.0, 10, 0, now.Add(2*time.Minute)),
	}

	summary := BuildRouteSummary(points)

	bbox := summary.BoundingBox
	if bbox.North != 38.0 {
		t.Errorf("Expected North: 38.0, got %f", bbox.North)
	}
	if bbox.South != 36.0 {
		t.Errorf("Expected South: 36.0, got %f", bbox.South)
	}
	if bbox.East != -121.0 {
		t.Errorf("Expected East: -121.0, got %f", bbox.East)
	}
	if bbox.West != -123.0 {
		t.Errorf("Expected West: -123.0, got %f", bbox.West)
	}

	if summary.Polyline == "" {
		t.Errorf("Expected polyline to be encoded")
	}
}

func TestHaversineKm_SamePoint(t *testing.T) {
	distance := haversineKm(0, 0, 0, 0)
	if distance != 0 {
		t.Errorf("Same point should have 0 distance, got %f", distance)
	}
}

func TestHaversineKm_KnownDistance(t *testing.T) {
	// New York to Los Angeles: ~3944 km
	ny := [2]float64{40.7128, -74.0060}
	la := [2]float64{34.0522, -118.2437}

	distance := haversineKm(ny[0], ny[1], la[0], la[1])

	// Should be approximately 3944 km
	if distance < 3900 || distance > 4000 {
		t.Errorf("Expected distance ~3944 km, got %.2f km", distance)
	}
}

func TestEstimateCalories_ZeroDuration(t *testing.T) {
	calories := estimateCalories(10, 0, 70)
	if calories != 0 {
		t.Errorf("Expected 0 calories for 0 duration, got %f", calories)
	}
}

func TestEstimateCalories_ZeroDistance(t *testing.T) {
	calories := estimateCalories(0, 60, 70)
	if calories != 0 {
		t.Errorf("Expected 0 calories for 0 distance, got %f", calories)
	}
}

func TestEstimateCalories_SlowRide(t *testing.T) {
	// 10 km in 1 hour = 10 km/h average speed (MET 4.0 for < 13 km/h)
	// Calories = 4.0 * 70 * 1 = 280
	calories := estimateCalories(10, 60, 70)
	expectedMin := 4.0 * 70 * 1 * 0.95
	expectedMax := 4.0 * 70 * 1 * 1.05
	if calories < expectedMin || calories > expectedMax {
		t.Errorf("Expected calories ~280, got %.0f", calories)
	}
}

func TestEstimateCalories_FastRide(t *testing.T) {
	// 60 km in 2 hours = 30 km/h average speed (MET 16.0)
	// Calories = 16.0 * 70 * 2 = 2240
	calories := estimateCalories(60, 120, 70)
	expectedMin := 16.0 * 70 * 2 * 0.95
	expectedMax := 16.0 * 70 * 2 * 1.05
	if calories < expectedMin || calories > expectedMax {
		t.Errorf("Expected calories ~2240, got %.0f", calories)
	}
}

func TestDownsamplePoints_NoDownsampling(t *testing.T) {
	now := time.Now()
	points := []sqlc.RideGpsPoint{
		createGPSPoint(0, 0, 10, 0, now),
		createGPSPoint(1, 0, 10, 0, now.Add(1*time.Minute)),
		createGPSPoint(2, 0, 10, 0, now.Add(2*time.Minute)),
	}

	result := downsamplePoints(points, 500)
	if len(result) != 3 {
		t.Errorf("Expected 3 points (no downsampling), got %d", len(result))
	}
}

func TestDownsamplePoints_HeavyDownsampling(t *testing.T) {
	now := time.Now()
	// Create 100 points
	points := make([]sqlc.RideGpsPoint, 100)
	for i := 0; i < 100; i++ {
		points[i] = createGPSPoint(float64(i)*0.01, 0, 10, 0, now.Add(time.Duration(i)*time.Second))
	}

	result := downsamplePoints(points, 10)
	if len(result) > 10 {
		t.Errorf("Expected at most 10 points, got %d", len(result))
	}
	if len(result) < 9 {
		t.Errorf("Expected at least 9 points, got %d", len(result))
	}
}

func TestCalculateBoundingBox_SinglePoint(t *testing.T) {
	now := time.Now()
	points := []sqlc.RideGpsPoint{
		createGPSPoint(40.7, -74.0, 10, 0, now),
	}

	bbox := calculateBoundingBox(points)
	if math.Abs(bbox.North-40.7) > 0.0001 {
		t.Errorf("Expected North 40.7, got %f", bbox.North)
	}
	if math.Abs(bbox.South-40.7) > 0.0001 {
		t.Errorf("Expected South 40.7, got %f", bbox.South)
	}
	if math.Abs(bbox.East-(-74.0)) > 0.0001 {
		t.Errorf("Expected East -74.0, got %f", bbox.East)
	}
	if math.Abs(bbox.West-(-74.0)) > 0.0001 {
		t.Errorf("Expected West -74.0, got %f", bbox.West)
	}
}
