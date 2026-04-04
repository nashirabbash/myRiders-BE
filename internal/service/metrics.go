package service

import (
	"math"

	"github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/pkg/polyline"
)

// MetricsResult contains computed ride metrics
type MetricsResult struct {
	DistanceKm      float64
	DurationSeconds int32
	MaxSpeedKmh     float64
	AvgSpeedKmh     float64
	ElevationM      float64
	Calories        int32
}

// RouteSummary contains encoded polyline and route bounds
type RouteSummary struct {
	Polyline    string `json:"polyline"`
	BoundingBox BBox   `json:"bounding_box"`
}

// BBox represents geographical bounding box
type BBox struct {
	North float64 `json:"north"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	West  float64 `json:"west"`
}

const (
	earthRadiusKm = 6371.0
)

// ComputeMetrics calculates ride statistics from GPS points
// userWeightKg is used for calorie estimation (defaults to 70kg if <= 0)
func ComputeMetrics(points []sqlc.RideGpsPoint, userWeightKg float64) MetricsResult {
	if len(points) < 2 {
		return MetricsResult{}
	}

	var totalDistance float64
	var maxSpeed float64
	var totalElevationGain float64

	// Calculate distance, speed, and elevation gain
	for i := 1; i < len(points); i++ {
		prev := points[i-1]
		curr := points[i]

		// Haversine distance
		distance := haversineKm(
			float64(prev.Latitude), float64(prev.Longitude),
			float64(curr.Latitude), float64(curr.Longitude),
		)
		totalDistance += distance

		// Track max speed
		if curr.SpeedKmh > maxSpeed {
			maxSpeed = curr.SpeedKmh
		}

		// Elevation gain (only count positive changes)
		elevationDiff := curr.ElevationM - prev.ElevationM
		if elevationDiff > 0 {
			totalElevationGain += elevationDiff
		}
	}

	// Calculate duration
	startTime := points[0].RecordedAt.Time
	endTime := points[len(points)-1].RecordedAt.Time
	duration := endTime.Sub(startTime)
	durationSeconds := int32(duration.Seconds())
	durationHours := duration.Hours()

	// Calculate average speed
	avgSpeed := 0.0
	if durationHours > 0 {
		avgSpeed = totalDistance / durationHours
	}

	// Estimate calories (basic formula for cycling)
	calories := estimateCalories(totalDistance, duration.Minutes(), userWeightKg)

	return MetricsResult{
		DistanceKm:      math.Round(totalDistance*100) / 100,
		DurationSeconds: durationSeconds,
		MaxSpeedKmh:     math.Round(maxSpeed*100) / 100,
		AvgSpeedKmh:     math.Round(avgSpeed*100) / 100,
		ElevationM:      math.Round(totalElevationGain),
		Calories:        int32(calories),
	}
}

// haversineKm calculates distance between two coordinates in kilometers
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

// estimateCalories estimates calories burned during a ride
// Uses formula: MET * weight(kg) * duration(hours)
// Defaults to 70kg if user weight not provided
func estimateCalories(distKm float64, durationMinutes float64, userWeightKg float64) float64 {
	if durationMinutes <= 0 || distKm <= 0 {
		return 0
	}

	// Use provided weight or default to 70kg
	weight := userWeightKg
	if weight <= 0 {
		weight = 70.0
	}

	// Average speed in km/h
	avgSpeedKmh := (distKm / durationMinutes) * 60

	// MET (Metabolic Equivalent) estimates based on speed
	// More granular values for better accuracy across different ride intensities
	met := 4.0
	if avgSpeedKmh >= 30 {
		met = 16.0
	} else if avgSpeedKmh >= 26 {
		met = 12.0
	} else if avgSpeedKmh >= 22 {
		met = 10.0
	} else if avgSpeedKmh >= 19 {
		met = 8.5
	} else if avgSpeedKmh >= 16 {
		met = 8.0
	} else if avgSpeedKmh >= 13 {
		met = 6.0
	}

	durationHours := durationMinutes / 60.0
	return met * weight * durationHours
}

// BuildRouteSummary creates a summary of the route with polyline encoding
func BuildRouteSummary(points []sqlc.RideGpsPoint) RouteSummary {
	if len(points) == 0 {
		return RouteSummary{
			BoundingBox: BBox{North: 0, South: 0, East: 0, West: 0},
		}
	}

	// Downsample points to max 500 for polyline encoding
	sampledPoints := downsamplePoints(points, 500)

	// Convert to coordinate pairs for polyline encoding
	coords := make([][2]float64, len(sampledPoints))
	for i, p := range sampledPoints {
		coords[i] = [2]float64{float64(p.Latitude), float64(p.Longitude)}
	}

	// Encode polyline
	encodedPolyline := polyline.Encode(coords)

	// Calculate bounding box
	bbox := calculateBoundingBox(points)

	return RouteSummary{
		Polyline:    encodedPolyline,
		BoundingBox: bbox,
	}
}

// downsamplePoints reduces the number of points while maintaining route shape
func downsamplePoints(points []sqlc.RideGpsPoint, maxPoints int) []sqlc.RideGpsPoint {
	if len(points) <= maxPoints {
		return points
	}

	factor := (len(points) + maxPoints - 1) / maxPoints
	result := make([]sqlc.RideGpsPoint, 0, maxPoints)

	for i, p := range points {
		if i%factor == 0 {
			result = append(result, p)
		}
	}

	return result
}

// calculateBoundingBox finds the geographical bounds of a route
func calculateBoundingBox(points []sqlc.RideGpsPoint) BBox {
	if len(points) == 0 {
		return BBox{}
	}

	bbox := BBox{
		North: float64(points[0].Latitude),
		South: float64(points[0].Latitude),
		East:  float64(points[0].Longitude),
		West:  float64(points[0].Longitude),
	}

	for _, p := range points {
		lat := float64(p.Latitude)
		lng := float64(p.Longitude)

		if lat > bbox.North {
			bbox.North = lat
		}
		if lat < bbox.South {
			bbox.South = lat
		}
		if lng > bbox.East {
			bbox.East = lng
		}
		if lng < bbox.West {
			bbox.West = lng
		}
	}

	return bbox
}
