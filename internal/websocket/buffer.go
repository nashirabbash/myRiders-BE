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
	connRefs  int // number of active WebSocket connections for this ride
	stopFlush chan struct{}
	mu        sync.Mutex
}

// GPSBuffer batches GPS points and flushes them to the database
type GPSBuffer struct {
	buffers map[string]*rideBuffer
	mu      sync.RWMutex
	queries *db.Queries // nil in Phase 1; set in Phase 2
}

// NewGPSBuffer creates a new GPS buffer
func NewGPSBuffer(queries *db.Queries) *GPSBuffer {
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

// Disconnect decrements the reference count. Holds the map lock for the entire
// check-and-delete so a concurrent reconnect cannot observe an inconsistent state.
func (b *GPSBuffer) Disconnect(rideID string) {
	b.mu.Lock()
	buf, ok := b.buffers[rideID]
	if !ok {
		b.mu.Unlock()
		return
	}

	buf.mu.Lock()
	buf.connRefs--
	last := buf.connRefs <= 0
	buf.mu.Unlock()

	if last {
		// Remove from map while still holding b.mu so no new Connect can
		// re-use the same rideID entry before we clean up.
		delete(b.buffers, rideID)
		b.mu.Unlock()

		// Stop flush goroutine and drain remaining points directly from buf
		// (cannot use flushOnce here — buf is no longer in the map).
		close(buf.stopFlush)
		b.flushBuf(buf)
		return
	}
	b.mu.Unlock()
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
			b.flushBuf(buf) // call flushBuf directly; no map lookup needed
		case <-buf.stopFlush:
			return
		}
	}
}

// flushBuf drains buf directly without a map lookup. Used by Disconnect after
// the entry has already been removed from b.buffers, and by flushLoop.
func (b *GPSBuffer) flushBuf(buf *rideBuffer) {
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

// flushOnce resolves the buffer from the map and delegates to flushBuf.
func (b *GPSBuffer) flushOnce(rideID string) {
	b.mu.RLock()
	buf, ok := b.buffers[rideID]
	b.mu.RUnlock()
	if ok {
		b.flushBuf(buf)
	}
}
