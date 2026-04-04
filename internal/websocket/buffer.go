package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/nashirabbash/trackride/internal/db"
)

const (
	batchSize     = 10
	flushInterval = 30 * time.Second
)

// GPSPoint represents a single GPS coordinate with metadata
type GPSPoint struct {
	Lat       float64
	Lng       float64
	SpeedKmh  float64
	ElevM     float64
	Timestamp time.Time
}

type rideBuffer struct {
	points []*GPSPoint
	timer  *time.Timer
	mu     sync.Mutex
}

// GPSBuffer batches GPS points and flushes them to the database
type GPSBuffer struct {
	buffers map[string]*rideBuffer
	mu      sync.RWMutex
	queries db.Queries // Can be nil in Phase 1, will be set in Phase 2
}

// NewGPSBuffer creates a new GPS buffer
func NewGPSBuffer(queries db.Queries) *GPSBuffer {
	return &GPSBuffer{
		buffers: make(map[string]*rideBuffer),
		queries: queries,
	}
}

// Add adds a GPS point to the buffer for a ride
func (b *GPSBuffer) Add(rideID string, point GPSPoint) {
	b.mu.Lock()
	buf, ok := b.buffers[rideID]
	if !ok {
		buf = &rideBuffer{points: make([]*GPSPoint, 0)}
		b.buffers[rideID] = buf
	}
	b.mu.Unlock()

	buf.mu.Lock()
	defer buf.mu.Unlock()

	buf.points = append(buf.points, &point)

	// Flush if batch is full
	if len(buf.points) >= batchSize {
		if buf.timer != nil {
			buf.timer.Stop()
		}
		go b.flush(rideID)
		return
	}

	// Reset flush timer
	if buf.timer != nil {
		buf.timer.Stop()
	}
	buf.timer = time.AfterFunc(flushInterval, func() {
		b.flush(rideID)
	})
}

// flush persists buffered GPS points to the database
func (b *GPSBuffer) flush(rideID string) {
	b.mu.Lock()
	buf, ok := b.buffers[rideID]
	b.mu.Unlock()
	if !ok {
		return
	}

	buf.mu.Lock()
	if len(buf.points) == 0 {
		buf.mu.Unlock()
		return
	}

	// Copy points and clear buffer
	points := make([]*GPSPoint, len(buf.points))
	copy(points, buf.points)
	buf.points = buf.points[:0]

	// Stop timer
	if buf.timer != nil {
		buf.timer.Stop()
		buf.timer = nil
	}
	buf.mu.Unlock()

	// Batch insert to database (placeholder for now)
	// This would call queries.InsertGPSPointsBatch with the points
	ctx := context.Background()
	_ = ctx
	_ = points
}

// FlushAndClear flushes any remaining points and removes the buffer
func (b *GPSBuffer) FlushAndClear(rideID string) {
	b.flush(rideID)
	b.mu.Lock()
	delete(b.buffers, rideID)
	b.mu.Unlock()
}
