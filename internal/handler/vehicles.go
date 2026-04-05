package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jackc/pgx/v5"
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	domainerrors "github.com/nashirabbash/trackride/internal/errors"
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
//
//	@Summary		List user vehicles
//	@Description	Retrieve all vehicles owned by the authenticated user
//	@Tags			Vehicles
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{array}		VehicleResponse
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Router			/v1/vehicles [get]
func (h *VehiclesHandler) List(c *gin.Context) {
	userUUID, ok := middleware.GetUserUUID(c)
	if !ok {
		RespondWithError(c, domainerrors.ErrUnauthorized)
		return
	}

	// Fetch vehicles
	vehicles, err := h.queries.ListVehiclesByUser(c.Request.Context(), userUUID)
	if err != nil {
		RespondWithError(c, domainerrors.ErrInternalServerError)
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
//
//	@Summary		Create new vehicle
//	@Description	Create a new vehicle for the authenticated user. Type must be one of: motor, mobil, sepeda
//	@Tags			Vehicles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateVehicleRequest	true	"Vehicle data"
//	@Success		201		{object}	VehicleResponse
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Failure		422		{object}	ErrorResponse	"Validation error"
//	@Router			/v1/vehicles [post]
func (h *VehiclesHandler) Create(c *gin.Context) {
	userUUID, ok := middleware.GetUserUUID(c)
	if !ok {
		RespondWithError(c, domainerrors.ErrUnauthorized)
		return
	}

	var req CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithValidationError(c, err.Error())
		return
	}

	// Create vehicle
	vehicle, err := h.queries.CreateVehicle(c.Request.Context(), dbsqlc.CreateVehicleParams{
		UserID: userUUID,
		Type:   req.Type,
		Name:   req.Name,
		Brand:  optionalStringPgtype(derefString(req.Brand)),
		Color:  optionalStringPgtype(derefString(req.Color)),
	})
	if err != nil {
		RespondWithError(c, domainerrors.ErrInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, mapVehicleToResponse(vehicle))
}

// Update modifies an existing vehicle
//
//	@Summary		Update vehicle
//	@Description	Update a vehicle owned by the authenticated user. All fields are optional.
//	@Tags			Vehicles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Vehicle ID (UUID format)"
//	@Param			body	body		UpdateVehicleRequest	true	"Vehicle update data"
//	@Success		200		{object}	VehicleResponse
//	@Failure		400		{object}	ErrorResponse	"Invalid ID format"
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	ErrorResponse	"Vehicle not found"
//	@Failure		422		{object}	ErrorResponse	"Validation error"
//	@Router			/v1/vehicles/{id} [put]
func (h *VehiclesHandler) Update(c *gin.Context) {
	userUUID, ok := middleware.GetUserUUID(c)
	if !ok {
		RespondWithError(c, domainerrors.ErrUnauthorized)
		return
	}

	vehicleID := c.Param("id")

	// Get raw JSON to detect omitted vs explicitly null fields
	var rawJSON map[string]interface{}
	if err := c.ShouldBindBodyWith(&rawJSON, binding.JSON); err != nil {
		RespondWithValidationError(c, err.Error())
		return
	}

	var req UpdateVehicleRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		RespondWithValidationError(c, err.Error())
		return
	}

	vehicleUUID, err := parseUUID(vehicleID)
	if err != nil {
		RespondWithError(c, domainerrors.ErrInvalidID)
		return
	}

	// Verify vehicle exists and belongs to user
	existing, err := h.queries.GetVehicleByIDAndUser(c.Request.Context(), dbsqlc.GetVehicleByIDAndUserParams{
		ID:     vehicleUUID,
		UserID: userUUID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondWithError(c, domainerrors.ErrVehicleNotFound)
		} else {
			RespondWithError(c, domainerrors.ErrInternalServerError)
		}
		return
	}

	// Use existing values for fields not provided; allow clearing fields if explicitly sent as null
	vehicleType := existing.Type
	if req.Type != nil {
		vehicleType = *req.Type
	}

	name := existing.Name
	if req.Name != nil {
		name = *req.Name
	}

	// For optional fields, check if they were explicitly sent (even if null)
	brand := existing.Brand
	if _, exists := rawJSON["brand"]; exists {
		// Field was sent (could be null or a string)
		brand = optionalStringPgtype(derefString(req.Brand))
	}

	color := existing.Color
	if _, exists := rawJSON["color"]; exists {
		// Field was sent (could be null or a string)
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
			RespondWithError(c, domainerrors.ErrVehicleNotFound)
		} else {
			RespondWithError(c, domainerrors.ErrInternalServerError)
		}
		return
	}

	c.JSON(http.StatusOK, mapVehicleToResponse(vehicle))
}

// Delete removes a vehicle (after checking it's not being used in active rides)
//
//	@Summary		Delete vehicle
//	@Description	Delete a vehicle owned by the authenticated user. Cannot delete if vehicle is in use in active rides.
//	@Tags			Vehicles
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Vehicle ID (UUID format)"
//	@Success		204
//	@Failure		400		{object}	ErrorResponse	"Invalid ID format"
//	@Failure		401		{object}	ErrorResponse	"Unauthorized"
//	@Failure		404		{object}	ErrorResponse	"Vehicle not found"
//	@Failure		409		{object}	ErrorResponse	"Vehicle is in use"
//	@Router			/v1/vehicles/{id} [delete]
func (h *VehiclesHandler) Delete(c *gin.Context) {
	userUUID, ok := middleware.GetUserUUID(c)
	if !ok {
		RespondWithError(c, domainerrors.ErrUnauthorized)
		return
	}

	vehicleID := c.Param("id")

	vehicleUUID, err := parseUUID(vehicleID)
	if err != nil {
		RespondWithError(c, domainerrors.ErrInvalidID)
		return
	}

	// Delete vehicle (atomic query prevents race condition with active rides check)
	rowsAffected, err := h.queries.DeleteVehicle(c.Request.Context(), dbsqlc.DeleteVehicleParams{
		ID:     vehicleUUID,
		UserID: userUUID,
	})
	if err != nil {
		RespondWithError(c, domainerrors.ErrInternalServerError)
		return
	}

	// If no rows were deleted, differentiate between not found vs. in use
	if rowsAffected == 0 {
		// Check if vehicle exists at all
		_, err := h.queries.GetVehicleByIDAndUser(c.Request.Context(), dbsqlc.GetVehicleByIDAndUserParams{
			ID:     vehicleUUID,
			UserID: userUUID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				RespondWithError(c, domainerrors.ErrVehicleNotFound)
			} else {
				RespondWithError(c, domainerrors.ErrInternalServerError)
			}
			return
		}

		// Vehicle exists but deletion failed, so it must have an active ride
		RespondWithError(c, domainerrors.ErrVehicleInUse)
		return
	}

	c.Status(http.StatusNoContent)
}
