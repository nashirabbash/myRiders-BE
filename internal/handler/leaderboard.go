package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/middleware"
)

// LeaderboardEntryResponse is the DTO for a leaderboard entry
type LeaderboardEntryResponse struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	VehicleType string  `json:"vehicle_type"`
	TotalKm     float64 `json:"total_km"`
	TotalRides  int32   `json:"total_rides"`
	Rank        int32   `json:"rank"`
	PeriodType  string  `json:"period_type"`
	PeriodStart string  `json:"period_start"`
}

// GetGlobal retrieves the global leaderboard
// Query parameters:
//   - period_type (default: "weekly", options: "weekly", "monthly")
//   - vehicle_type (optional: "motor", "mobil", "sepeda")
//   - page (default: 1)
//   - limit (default: 20, max: 100)
//
// Note: "all-time" leaderboard is not yet implemented. Use "monthly" for longer-term rankings.
func (h *LeaderboardHandler) GetGlobal(c *gin.Context) {
	// Parse query parameters
	periodType := c.DefaultQuery("period_type", "weekly")
	vehicleType := c.Query("vehicle_type")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	// Validate and parse pagination with error handling
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_PAGE", "detail": "page must be a positive integer"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_LIMIT", "detail": "limit must be between 1 and 100"})
		return
	}

	// Validate period_type (only weekly and monthly are supported)
	if periodType != "weekly" && periodType != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "INVALID_PERIOD_TYPE",
			"detail": "period_type must be 'weekly' or 'monthly' (all-time not yet implemented)",
		})
		return
	}

	offset := int32((page - 1) * limit)
	limitInt32 := int32(limit)

	// Calculate period_start based on period_type
	periodStart := calculatePeriodStart(periodType)
	periodStartDate := pgtype.Date{
		Time:  periodStart,
		Valid: true,
	}

	ctx := c.Request.Context()
	var entries []sqlc.LeaderboardEntry
	var err2 error

	if vehicleType == "" {
		// Get global leaderboard without vehicle filter
		entries, err2 = h.queries.GetLeaderboardByPeriod(ctx, sqlc.GetLeaderboardByPeriodParams{
			PeriodType:  periodType,
			PeriodStart: periodStartDate,
			Limit:       limitInt32,
			Offset:      offset,
		})
	} else {
		// Validate vehicle type
		if vehicleType != "motor" && vehicleType != "mobil" && vehicleType != "sepeda" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "INVALID_VEHICLE_TYPE",
				"detail": "vehicle_type must be 'motor', 'mobil', or 'sepeda'",
			})
			return
		}

		// Get global leaderboard with vehicle filter
		entries, err2 = h.queries.GetLeaderboardByPeriodAndVehicle(ctx, sqlc.GetLeaderboardByPeriodAndVehicleParams{
			PeriodType:  periodType,
			PeriodStart: periodStartDate,
			VehicleType: sqlc.NullVehicleType{VehicleType: sqlc.VehicleType(vehicleType), Valid: true},
			Limit:       limitInt32,
			Offset:      offset,
		})
	}

	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Convert to response DTOs
	response := make([]LeaderboardEntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = LeaderboardEntryResponse{
			ID:          entry.ID.String(),
			UserID:      entry.UserID.String(),
			VehicleType: string(entry.VehicleType.VehicleType),
			TotalKm:     entry.TotalKm,
			TotalRides:  entry.TotalRides,
			Rank:        entry.Rank,
			PeriodType:  entry.PeriodType,
			PeriodStart: entry.PeriodStart.Time.Format("2006-01-02"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":          response,
		"page":          page,
		"limit":         limit,
		"period_type":   periodType,
		"vehicle_type":  vehicleType,
	})
}

// GetFriends retrieves the leaderboard for the authenticated user's friends
// Query parameters:
//   - period_type (default: "weekly", options: "weekly", "monthly")
//   - vehicle_type (optional: "motor", "mobil", "sepeda")
//   - page (default: 1)
//   - limit (default: 20, max: 100)
//
// Note: Only returns rankings of users you follow. "all-time" leaderboard not yet implemented.
func (h *LeaderboardHandler) GetFriends(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	// Parse UUID
	userUUID, err := parseUUID(userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	// Parse query parameters
	periodType := c.DefaultQuery("period_type", "weekly")
	vehicleType := c.Query("vehicle_type")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	// Validate and parse pagination with error handling
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_PAGE", "detail": "page must be a positive integer"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_LIMIT", "detail": "limit must be between 1 and 100"})
		return
	}

	// Validate period_type (only weekly and monthly are supported)
	if periodType != "weekly" && periodType != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "INVALID_PERIOD_TYPE",
			"detail": "period_type must be 'weekly' or 'monthly' (all-time not yet implemented)",
		})
		return
	}

	offset := int32((page - 1) * limit)
	limitInt32 := int32(limit)

	// Calculate period_start based on period_type
	periodStart := calculatePeriodStart(periodType)
	periodStartDate := pgtype.Date{
		Time:  periodStart,
		Valid: true,
	}

	ctx := c.Request.Context()
	var entries []sqlc.LeaderboardEntry
	var err2 error

	if vehicleType == "" {
		// Get friends leaderboard without vehicle filter
		entries, err2 = h.queries.GetFriendsLeaderboard(ctx, sqlc.GetFriendsLeaderboardParams{
			FollowerID:  userUUID,
			PeriodType:  periodType,
			PeriodStart: periodStartDate,
			Limit:       limitInt32,
			Offset:      offset,
		})
	} else {
		// Validate vehicle type
		if vehicleType != "motor" && vehicleType != "mobil" && vehicleType != "sepeda" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":  "INVALID_VEHICLE_TYPE",
				"detail": "vehicle_type must be 'motor', 'mobil', or 'sepeda'",
			})
			return
		}

		// Get friends leaderboard with vehicle filter
		entries, err2 = h.queries.GetFriendsLeaderboardByVehicle(ctx, sqlc.GetFriendsLeaderboardByVehicleParams{
			FollowerID:  userUUID,
			PeriodType:  periodType,
			PeriodStart: periodStartDate,
			VehicleType: sqlc.NullVehicleType{VehicleType: sqlc.VehicleType(vehicleType), Valid: true},
			Limit:       limitInt32,
			Offset:      offset,
		})
	}

	if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Convert to response DTOs
	response := make([]LeaderboardEntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = LeaderboardEntryResponse{
			ID:          entry.ID.String(),
			UserID:      entry.UserID.String(),
			VehicleType: string(entry.VehicleType.VehicleType),
			TotalKm:     entry.TotalKm,
			TotalRides:  entry.TotalRides,
			Rank:        entry.Rank,
			PeriodType:  entry.PeriodType,
			PeriodStart: entry.PeriodStart.Time.Format("2006-01-02"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":          response,
		"page":          page,
		"limit":         limit,
		"period_type":   periodType,
		"vehicle_type":  vehicleType,
	})
}

// calculatePeriodStart calculates the period_start date based on period_type
// Explicitly truncates timestamps to midnight in UTC
func calculatePeriodStart(periodType string) time.Time {
	now := time.Now().UTC()
	switch periodType {
	case "weekly":
		// Get the Monday of the current week, explicitly truncated to midnight
		daysBack := int(now.Weekday()) - int(time.Monday)
		if daysBack < 0 {
			daysBack += 7
		}
		d := now.AddDate(0, 0, -daysBack)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	case "monthly":
		// Get the first day of the current month, explicitly truncated to midnight
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		// Default to weekly
		daysBack := int(now.Weekday()) - int(time.Monday)
		if daysBack < 0 {
			daysBack += 7
		}
		d := now.AddDate(0, 0, -daysBack)
		return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	}
}
