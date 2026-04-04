package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nashirabbash/trackride/internal/middleware"
)

// Start handles POST /rides/start
func (h *RidesHandler) Start(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	var req struct {
		VehicleID string `json:"vehicle_id" binding:"required,uuid"`
		Title     string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	ride, wsToken, err := h.service.StartRide(c, userID, req.VehicleID, req.Title)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Store ws_token in Redis with 10-minute TTL
	h.redis.SetEx(c, "ws_token:"+wsToken, userID+":"+ride.ID.String(), 10*time.Minute)

	c.JSON(http.StatusCreated, gin.H{
		"ride_id":  ride.ID,
		"ws_token": wsToken,
		"started_at": ride.StartedAt,
	})
}

// Stop handles POST /rides/:id/stop
func (h *RidesHandler) Stop(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	rideID := c.Param("id")
	ride, err := h.service.StopRide(c, rideID, userID)
	if err != nil {
		statusCode := http.StatusNotFound
		if err.Error() == "RIDE_NOT_ACTIVE" {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ride)
}

// List handles GET /rides
func (h *RidesHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	vehicleType := c.Query("vehicle_type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	rides, total, err := h.service.ListRides(c, userID, vehicleType, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  rides,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// GetByID handles GET /rides/:id
func (h *RidesHandler) GetByID(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	rideID := c.Param("id")
	ride, err := h.service.GetRideByID(c, rideID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "RIDE_NOT_FOUND"})
		return
	}

	c.JSON(http.StatusOK, ride)
}
