package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/middleware"
)

// CreateVehicleRequest represents a vehicle creation payload
type CreateVehicleRequest struct {
	Type  dbsqlc.VehicleType `json:"type" binding:"required,oneof=motor mobil sepeda"`
	Name  string             `json:"name" binding:"required,min=1,max=100"`
	Brand *string            `json:"brand" binding:"omitempty,max=50"`
	Color *string            `json:"color" binding:"omitempty,max=30"`
}

// UpdateVehicleRequest represents a vehicle update payload
type UpdateVehicleRequest struct {
	Type  *dbsqlc.VehicleType `json:"type" binding:"omitempty,oneof=motor mobil sepeda"`
	Name  *string             `json:"name" binding:"omitempty,min=1,max=100"`
	Brand *string             `json:"brand" binding:"omitempty,max=50"`
	Color *string             `json:"color" binding:"omitempty,max=30"`
}

// VehicleResponse is the DTO for vehicle responses
type VehicleResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Brand     string `json:"brand"`
	Color     string `json:"color"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// mapVehicleToResponse converts a database Vehicle to a VehicleResponse DTO
func mapVehicleToResponse(v dbsqlc.Vehicle) VehicleResponse {
	return VehicleResponse{
		ID:        v.ID.String(),
		Type:      string(v.Type),
		Name:      v.Name,
		Brand:     v.Brand.String,
		Color:     v.Color.String,
		IsActive:  v.IsActive,
		CreatedAt: v.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: v.UpdatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// List returns all vehicles belonging to the authenticated user
func (h *VehiclesHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	// Parse UUID
	id, err := parseUUID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	// Fetch vehicles
	vehicles, err := h.queries.ListVehiclesByUser(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Convert to response DTOs
	response := make([]VehicleResponse, len(vehicles))
	for i, v := range vehicles {
		response[i] = mapVehicleToResponse(v)
	}

	c.JSON(http.StatusOK, response)
}

// Create adds a new vehicle for the authenticated user
func (h *VehiclesHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	var req CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	// Parse UUID
	id, err := parseUUID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	// Create vehicle
	vehicle, err := h.queries.CreateVehicle(c.Request.Context(), dbsqlc.CreateVehicleParams{
		UserID: id,
		Type:   req.Type,
		Name:   req.Name,
		Brand:  optionalStringPgtype(derefString(req.Brand)),
		Color:  optionalStringPgtype(derefString(req.Color)),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusCreated, mapVehicleToResponse(vehicle))
}

// Update modifies an existing vehicle
func (h *VehiclesHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	vehicleID := c.Param("id")

	var req UpdateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	// Parse UUIDs
	userUUID, err := parseUUID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	vehicleUUID, err := parseUUID(vehicleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}

	// Verify vehicle exists and belongs to user
	existing, err := h.queries.GetVehicleByID(c.Request.Context(), vehicleUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "VEHICLE_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Security check: ensure vehicle belongs to authenticated user
	if existing.UserID.String() != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "VEHICLE_NOT_FOUND"})
		return
	}

	// Use existing values for fields not provided
	vehicleType := existing.Type
	if req.Type != nil {
		vehicleType = *req.Type
	}

	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	brand := existing.Brand
	if req.Brand != nil {
		brand = optionalStringPgtype(derefString(req.Brand))
	}

	color := existing.Color
	if req.Color != nil {
		color = optionalStringPgtype(derefString(req.Color))
	}

	// Update vehicle
	vehicle, err := h.queries.UpdateVehicle(c.Request.Context(), dbsqlc.UpdateVehicleParams{
		ID:     vehicleUUID,
		UserID: userUUID,
		Type:   vehicleType,
		Name:   name,
		Brand:  brand,
		Color:  color,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "VEHICLE_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	c.JSON(http.StatusOK, mapVehicleToResponse(vehicle))
}

// Delete removes a vehicle (after checking it's not being used in active rides)
func (h *VehiclesHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	vehicleID := c.Param("id")

	// Parse UUIDs
	userUUID, err := parseUUID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	vehicleUUID, err := parseUUID(vehicleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}

	// Verify vehicle exists and belongs to user
	existing, err := h.queries.GetVehicleByID(c.Request.Context(), vehicleUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "VEHICLE_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Security check: ensure vehicle belongs to authenticated user
	if existing.UserID.String() != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "VEHICLE_NOT_FOUND"})
		return
	}

	// Check if vehicle is being used in an active ride
	hasActive, err := h.queries.HasActiveRide(c.Request.Context(), vehicleUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	if hasActive {
		c.JSON(http.StatusConflict, gin.H{"error": "VEHICLE_IN_USE"})
		return
	}

	// Delete vehicle
	err = h.queries.DeleteVehicle(c.Request.Context(), dbsqlc.DeleteVehicleParams{
		ID:     vehicleUUID,
		UserID: userUUID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
