package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/nashirabbash/trackride/internal/db"
)

const (
	batchSize     = 10
	flushInterval = 5 * time.Second // tick interval for the flush goroutine
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
	points    []GPSPoint
	lastAdded time.Time
	connRefs  int // number of active WebSocket connections for this ride
	stopFlush chan struct{}
	mu        sync.Mutex
}

// GPSBuffer batches GPS points and flushes them to the database
type GPSBuffer struct {
	buffers map[string]*rideBuffer
	mu      sync.RWMutex
	queries db.Queries // nil in Phase 1; set in Phase 2
}

// NewGPSBuffer creates a new GPS buffer
func NewGPSBuffer(queries db.Queries) *GPSBuffer {
	return &GPSBuffer{
		buffers: make(map[string]*rideBuffer),
		queries: queries,
	}
}

// Connect increments the connection reference count for a ride and starts the
// flush goroutine on the first connection.
func (b *GPSBuffer) Connect(rideID string) {
	b.mu.Lock()
	buf, ok := b.buffers[rideID]
	if !ok {
		buf = &rideBuffer{
			points:    make([]GPSPoint, 0),
			stopFlush: make(chan struct{}),
		}
		b.buffers[rideID] = buf
		go b.flushLoop(rideID, buf)
	}
	buf.mu.Lock()
	buf.connRefs++
	buf.mu.Unlock()
	b.mu.Unlock()
}

// Disconnect decrements the reference count. Only flushes and removes the
// buffer when the last connection for this ride closes.
func (b *GPSBuffer) Disconnect(rideID string) {
	b.mu.Lock()
	buf, ok := b.buffers[rideID]
	b.mu.Unlock()
	if !ok {
		return
	}

	buf.mu.Lock()
	buf.connRefs--
	last := buf.connRefs <= 0
	buf.mu.Unlock()

	if last {
		// Signal the flush goroutine to stop, do a final flush, then remove.
		close(buf.stopFlush)
		b.flushOnce(rideID)
		b.mu.Lock()
		delete(b.buffers, rideID)
		b.mu.Unlock()
	}
}

// Add appends a GPS point to the buffer. A batch flush is triggered immediately
// when the batch size threshold is reached.
func (b *GPSBuffer) Add(rideID string, point GPSPoint) {
	b.mu.RLock()
	buf, ok := b.buffers[rideID]
	b.mu.RUnlock()
	if !ok {
		return
	}

	buf.mu.Lock()
	buf.points = append(buf.points, point)
	buf.lastAdded = time.Now()
	full := len(buf.points) >= batchSize
	buf.mu.Unlock()

	if full {
		b.flushOnce(rideID)
	}
}

// flushLoop runs a single background goroutine per ride, flushing on a fixed
// interval whenever there are unsaved points that haven't been batch-flushed.
func (b *GPSBuffer) flushLoop(rideID string, buf *rideBuffer) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			buf.mu.Lock()
			hasPoints := len(buf.points) > 0
			buf.mu.Unlock()
			if hasPoints {
				b.flushOnce(rideID)
			}
		case <-buf.stopFlush:
			return
		}
	}
}

// flushOnce drains the current buffer and persists points to the database.
func (b *GPSBuffer) flushOnce(rideID string) {
	b.mu.RLock()
	buf, ok := b.buffers[rideID]
	b.mu.RUnlock()
	if !ok {
		return
	}

	buf.mu.Lock()
	if len(buf.points) == 0 {
		buf.mu.Unlock()
		return
	}
	points := make([]GPSPoint, len(buf.points))
	copy(points, buf.points)
	buf.points = buf.points[:0]
	buf.mu.Unlock()

	// Phase 2: call queries.InsertGPSPointsBatch with points
	ctx := context.Background()
	_ = ctx
	_ = points
}
