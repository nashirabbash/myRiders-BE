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
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	VehicleType string `json:"vehicle_type"`
	TotalKm     float64 `json:"total_km"`
	TotalRides  int32 `json:"total_rides"`
	Rank        int32 `json:"rank"`
	PeriodType  string `json:"period_type"`
	PeriodStart string `json:"period_start"`
}

// GetGlobal retrieves the global leaderboard
// Query parameters:
//   - period_type (default: "weekly", options: "weekly", "monthly", "all-time")
//   - vehicle_type (optional: "motor", "mobil", "sepeda")
//   - page (default: 1)
//   - limit (default: 20, max: 100)
func (h *LeaderboardHandler) GetGlobal(c *gin.Context) {
	// Parse query parameters
	periodType := c.DefaultQuery("period_type", "weekly")
	vehicleType := c.Query("vehicle_type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and constrain pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Validate period_type
	if periodType != "weekly" && periodType != "monthly" && periodType != "all-time" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_PERIOD_TYPE"})
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
	var err error

	if vehicleType == "" {
		// Get global leaderboard without vehicle filter
		entries, err = h.queries.GetLeaderboardByPeriod(ctx, sqlc.GetLeaderboardByPeriodParams{
			PeriodType:  periodType,
			PeriodStart: periodStartDate,
			Limit:       limitInt32,
			Offset:      offset,
		})
	} else {
		// Get global leaderboard with vehicle filter
		entries, err = h.queries.GetLeaderboardByPeriodAndVehicle(ctx, sqlc.GetLeaderboardByPeriodAndVehicleParams{
			PeriodType:  periodType,
			PeriodStart: periodStartDate,
			VehicleType: sqlc.NullVehicleType{VehicleType: sqlc.VehicleType(vehicleType), Valid: true},
			Limit:       limitInt32,
			Offset:      offset,
		})
	}

	if err != nil {
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
//   - period_type (default: "weekly", options: "weekly", "monthly", "all-time")
//   - vehicle_type (optional: "motor", "mobil", "sepeda")
//   - page (default: 1)
//   - limit (default: 20, max: 100)
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate and constrain pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Validate period_type
	if periodType != "weekly" && periodType != "monthly" && periodType != "all-time" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_PERIOD_TYPE"})
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
func calculatePeriodStart(periodType string) time.Time {
	now := time.Now().UTC()
	switch periodType {
	case "weekly":
		// Get the Monday of the current week
		daysBack := int(now.Weekday()) - int(time.Monday)
		if daysBack < 0 {
			daysBack += 7
		}
		return now.AddDate(0, 0, -daysBack)
	case "monthly":
		// Get the first day of the current month
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "all-time":
		// For all-time, use a date far in the past
		return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		// Default to weekly
		daysBack := int(now.Weekday()) - int(time.Monday)
		if daysBack < 0 {
			daysBack += 7
		}
		return now.AddDate(0, 0, -daysBack)
	}
}
