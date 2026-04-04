package polyline

import "math"

// Encode converts coordinates to Google Encoded Polyline format
// Takes a slice of [2]float64 arrays where each contains [latitude, longitude]
func Encode(coords [][2]float64) string {
	if len(coords) == 0 {
		return ""
	}

	var result []byte
	var prevLat, prevLng int32

	for _, coord := range coords {
		lat := int32(math.Round(coord[0] * 1e5))
		lng := int32(math.Round(coord[1] * 1e5))

		latDelta := lat - prevLat
		lngDelta := lng - prevLng

		result = append(result, encodeValue(latDelta)...)
		result = append(result, encodeValue(lngDelta)...)

		prevLat = lat
		prevLng = lng
	}

	return string(result)
}

// Decode converts a Google Encoded Polyline string back to coordinates
func Decode(encoded string) [][2]float64 {
	var coords [][2]float64
	var lat, lng int32
	var index int

	for index < len(encoded) {
		latDelta := decodeValue(encoded, &index)
		lngDelta := decodeValue(encoded, &index)

		lat += latDelta
		lng += lngDelta

		coords = append(coords, [2]float64{
			float64(lat) / 1e5,
			float64(lng) / 1e5,
		})
	}

	return coords
}

// encodeValue encodes a single coordinate delta value
func encodeValue(value int32) []byte {
	// Left shift by 1 and invert if negative
	value <<= 1
	if value < 0 {
		value = ^value
	}

	var chunks []byte

	// Break value into 5-bit chunks
	for value >= 0x20 {
		chunk := byte((0x20 | (value & 0x1f)) + 63)
		chunks = append(chunks, chunk)
		value >>= 5
	}

	// Add final chunk
	chunks = append(chunks, byte(value+63))

	return chunks
}

// decodeValue decodes a single coordinate delta value from the encoded string
func decodeValue(encoded string, index *int) int32 {
	var result int32
	var shift uint

	for {
		if *index >= len(encoded) {
			break
		}

		b := int32(encoded[*index]) - 63
		*index++

		result |= (b & 0x1f) << shift
		shift += 5

		if b < 0x20 {
			break
		}
	}

	// Invert if the least significant bit is set
	if result&1 != 0 {
		return ^(result >> 1)
	}

	return result >> 1
}
