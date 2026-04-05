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
//
//	@Summary		Start a new ride
//	@Description	Create and start a new ride for a specific vehicle. Returns a ws_token for WebSocket connection.
//	@Tags			Rides
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body	object	true	"Ride start data (vehicle_id required, title optional)"
//	@Success		201		{object}	object
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	ErrorResponse	"Vehicle not found"
//	@Failure		422		{object}	ErrorResponse	"Validation error"
//	@Router			/v1/rides/start [post]
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
//
//	@Summary		Stop a ride
//	@Description	Stop an active ride and compute metrics (distance, duration, calories, elevation gain)
//	@Tags			Rides
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Ride ID (UUID format)"
//	@Success		200		{object}	object
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	ErrorResponse	"Ride not found"
//	@Router			/v1/rides/{id}/stop [post]
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
//
//	@Summary		List user rides
//	@Description	Retrieve completed rides with pagination and optional filtering by vehicle type
//	@Tags			Rides
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page			query	int		false	"Page number (default: 1)"
//	@Param			limit			query	int		false	"Items per page, max 100 (default: 20)"
//	@Param			vehicle_type	query	string	false	"Filter by vehicle type (motor, mobil, sepeda)"
//	@Success		200		{object}	object
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Router			/v1/rides [get]
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
//
//	@Summary		Get ride details
//	@Description	Retrieve detailed information about a specific ride
//	@Tags			Rides
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Ride ID (UUID format)"
//	@Success		200		{object}	object
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	ErrorResponse	"Ride not found"
//	@Router			/v1/rides/{id} [get]
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
