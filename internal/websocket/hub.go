package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow WebSocket connections from any origin for MVP
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Hub manages WebSocket connections for GPS streaming
type Hub struct {
	buffer *GPSBuffer
	redis  *redis.Client
}

// NewHub creates a new WebSocket hub
func NewHub(buffer *GPSBuffer, redis *redis.Client) *Hub {
	return &Hub{
		buffer: buffer,
		redis:  redis,
	}
}

type incomingMessage struct {
	Type      string  `json:"type"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	SpeedKmh  float64 `json:"speed_kmh"`
	ElevM     float64 `json:"elevation_m"`
	Timestamp string  `json:"timestamp"`
}

// HandleWS handles incoming WebSocket connections for a ride
func (h *Hub) HandleWS(c *gin.Context) {
	rideID := c.Param("id")
	wsToken := c.Query("token")

	// Validate WebSocket token from Redis
	val, err := h.redis.Get(c, "ws_token:"+wsToken).Result()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	// Verify that the token is for this ride (format: "userID:rideID")
	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 || parts[1] != rideID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	userID := parts[0]

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] Upgrade error for ride=%s: %v", rideID, err)
		return
	}
	defer func() {
		conn.Close()
		h.buffer.FlushAndClear(rideID)
		log.Printf("[WS] Disconnected: ride=%s user=%s", rideID, userID)
	}()

	log.Printf("[WS] Connected: ride=%s user=%s", rideID, userID)
	var pointCount int

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Error: %v", err)
			}
			break
		}

		var msg incomingMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "ping":
			_ = conn.WriteJSON(map[string]string{
				"type":        "pong",
				"server_time": time.Now().UTC().Format(time.RFC3339),
			})

		case "gps_point":
			// Validate coordinates
			if msg.Lat < -90 || msg.Lat > 90 || msg.Lng < -180 || msg.Lng > 180 {
				continue
			}

			ts, err := time.Parse(time.RFC3339, msg.Timestamp)
			if err != nil || ts.IsZero() {
				ts = time.Now().UTC()
			}

			h.buffer.Add(rideID, GPSPoint{
				Lat:       msg.Lat,
				Lng:       msg.Lng,
				SpeedKmh:  msg.SpeedKmh,
				ElevM:     msg.ElevM,
				Timestamp: ts,
			})
			pointCount++

			_ = conn.WriteJSON(map[string]interface{}{
				"type":            "ack",
				"points_received": pointCount,
			})
		}
	}
}
