package polyline

import (
	"math"
	"testing"
)

func TestEncode_EmptyCoordinates(t *testing.T) {
	result := Encode([][2]float64{})
	if result != "" {
		t.Errorf("Expected empty string for empty coords, got %q", result)
	}
}

func TestEncode_SinglePoint(t *testing.T) {
	coords := [][2]float64{{0, 0}}
	result := Encode(coords)
	if result == "" {
		t.Fatal("Encode should return non-empty string for single point")
	}
}

func TestEncode_MultiplePoints(t *testing.T) {
	// Test with a simple route
	coords := [][2]float64{
		{38.5, -120.2},
		{40.7, -120.95},
		{43.252, -126.453},
	}
	result := Encode(coords)
	if result == "" {
		t.Fatal("Encode should return non-empty string")
	}
}

func TestEncode_NegativeCoordinates(t *testing.T) {
	coords := [][2]float64{
		{-38.5, -120.2},
		{-40.7, -120.95},
	}
	result := Encode(coords)
	if result == "" {
		t.Fatal("Encode should handle negative coordinates")
	}
}

func TestDecode_EmptyString(t *testing.T) {
	result := Decode("")
	if len(result) != 0 {
		t.Errorf("Expected empty slice for empty string, got %d coords", len(result))
	}
}

func TestEncodeDecode_RoundTrip(t *testing.T) {
	// Google's example from their documentation
	original := [][2]float64{
		{38.5, -120.2},
		{40.7, -120.95},
		{43.252, -126.453},
	}

	// Encode
	encoded := Encode(original)

	// Decode
	decoded := Decode(encoded)

	// Check that we have the same number of points
	if len(decoded) != len(original) {
		t.Errorf("Expected %d points after round-trip, got %d", len(original), len(decoded))
	}

	// Check that points are close to original (within rounding error)
	tolerance := 0.00001 // About 1 meter at equator
	for i, coord := range decoded {
		latDiff := math.Abs(coord[0] - original[i][0])
		lngDiff := math.Abs(coord[1] - original[i][1])

		if latDiff > tolerance {
			t.Errorf("Point %d latitude mismatch: expected %f, got %f (diff: %f)", i, original[i][0], coord[0], latDiff)
		}
		if lngDiff > tolerance {
			t.Errorf("Point %d longitude mismatch: expected %f, got %f (diff: %f)", i, original[i][1], coord[1], lngDiff)
		}
	}
}

func TestEncodeDecode_SinglePoint(t *testing.T) {
	original := [][2]float64{{0.0, 0.0}}

	encoded := Encode(original)
	decoded := Decode(encoded)

	if len(decoded) != 1 {
		t.Errorf("Expected 1 point, got %d", len(decoded))
	}

	if decoded[0][0] != 0.0 || decoded[0][1] != 0.0 {
		t.Errorf("Expected (0, 0), got (%f, %f)", decoded[0][0], decoded[0][1])
	}
}

func TestEncodeDecode_NegativeCoordinates(t *testing.T) {
	original := [][2]float64{
		{-10.5, -50.25},
		{-11.3, -51.75},
	}

	encoded := Encode(original)
	decoded := Decode(encoded)

	if len(decoded) != len(original) {
		t.Errorf("Expected %d points, got %d", len(original), len(decoded))
	}

	tolerance := 0.00001
	for i, coord := range decoded {
		latDiff := math.Abs(coord[0] - original[i][0])
		lngDiff := math.Abs(coord[1] - original[i][1])

		if latDiff > tolerance || lngDiff > tolerance {
			t.Errorf("Point %d mismatch: expected (%f, %f), got (%f, %f)",
				i, original[i][0], original[i][1], coord[0], coord[1])
		}
	}
}

func TestEncodeDecode_LargeCoordinates(t *testing.T) {
	original := [][2]float64{
		{89.9, 179.9},
		{-89.9, -179.9},
	}

	encoded := Encode(original)
	decoded := Decode(encoded)

	if len(decoded) != len(original) {
		t.Errorf("Expected %d points, got %d", len(original), len(decoded))
	}

	tolerance := 0.00001
	for i, coord := range decoded {
		latDiff := math.Abs(coord[0] - original[i][0])
		lngDiff := math.Abs(coord[1] - original[i][1])

		if latDiff > tolerance || lngDiff > tolerance {
			t.Errorf("Point %d mismatch: expected (%f, %f), got (%f, %f)",
				i, original[i][0], original[i][1], coord[0], coord[1])
		}
	}
}

func TestEncode_PrecisionPreserved(t *testing.T) {
	// Test that encoding preserves precision to ~5 decimal places
	coords := [][2]float64{
		{37.4419, -122.1430}, // Mountain View, CA
	}

	encoded := Encode(coords)
	decoded := Decode(encoded)

	// At 5 decimal places, precision is about 1.1 meters
	tolerance := 0.00001
	latDiff := math.Abs(decoded[0][0] - coords[0][0])
	lngDiff := math.Abs(decoded[0][1] - coords[0][1])

	if latDiff > tolerance {
		t.Errorf("Latitude precision lost: expected %f, got %f", coords[0][0], decoded[0][0])
	}
	if lngDiff > tolerance {
		t.Errorf("Longitude precision lost: expected %f, got %f", coords[0][1], decoded[0][1])
	}
}
