package polyline

import "math"

// Encode converts a slice of [lat, lng] coordinate pairs to Google Encoded Polyline format
func Encode(coords [][2]float64) string {
	var result []byte
	var prevLat, prevLng int

	for _, c := range coords {
		lat := int(math.Round(c[0] * 1e5))
		lng := int(math.Round(c[1] * 1e5))

		result = append(result, encodeValue(lat-prevLat)...)
		result = append(result, encodeValue(lng-prevLng)...)

		prevLat = lat
		prevLng = lng
	}

	return string(result)
}

// encodeValue encodes a single value using variable-length encoding
func encodeValue(v int) []byte {
	v <<= 1
	if v < 0 {
		v = ^v
	}

	var chunks []byte
	for v >= 0x20 {
		chunks = append(chunks, byte((0x20|(v&0x1f))+63))
		v >>= 5
	}
	chunks = append(chunks, byte(v+63))

	return chunks
}

// Decode converts a Google Encoded Polyline string to [lat, lng] coordinate pairs
func Decode(encoded string) [][2]float64 {
	var coords [][2]float64
	var lat, lng int
	i := 0

	decodeValue := func() int {
		result := 0
		shift := 0
		for {
			if i >= len(encoded) {
				break
			}
			b := int(encoded[i]) - 63
			i++
			result |= (b & 0x1f) << shift
			shift += 5
			if b < 0x20 {
				break
			}
		}
		if result&1 != 0 {
			return ^(result >> 1)
		}
		return result >> 1
	}

	for i < len(encoded) {
		lat += decodeValue()
		lng += decodeValue()
		coords = append(coords, [2]float64{float64(lat) / 1e5, float64(lng) / 1e5})
	}

	return coords
}
