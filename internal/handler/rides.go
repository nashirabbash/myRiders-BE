package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	domainerrors "github.com/nashirabbash/trackride/internal/errors"
	"github.com/nashirabbash/trackride/internal/middleware"
)

// Start handles POST /rides/start
func (h *RidesHandler) Start(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		RespondWithError(c, domainerrors.ErrUnauthorized)
		return
	}

	var req struct {
		VehicleID string `json:"vehicle_id" binding:"required,uuid"`
		Title     string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, err.Error())
		return
	}

	ride, wsToken, err := h.service.StartRide(c.Request.Context(), userID, req.VehicleID, req.Title)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	// Store ws_token in Redis with 10-minute TTL
	h.redis.SetEx(c.Request.Context(), "ws_token:"+wsToken, userID+":"+ride.ID.String(), 10*time.Minute)

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
		RespondWithError(c, domainerrors.ErrUnauthorized)
		return
	}

	rideID := c.Param("id")
	ride, err := h.service.StopRide(c.Request.Context(), rideID, userID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, ride)
}

// List handles GET /rides
func (h *RidesHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		RespondWithError(c, domainerrors.ErrUnauthorized)
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

	rides, total, err := h.service.ListRides(c.Request.Context(), userID, vehicleType, page, limit)
	if err != nil {
		RespondWithError(c, err)
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
		RespondWithError(c, domainerrors.ErrUnauthorized)
		return
	}

	rideID := c.Param("id")
	ride, err := h.service.GetRideByID(c.Request.Context(), rideID, userID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, ride)
}
