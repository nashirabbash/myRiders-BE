package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/internal/middleware"
	"github.com/nashirabbash/trackride/internal/service"
)

// Follow handles POST /users/:id/follow
func (h *SocialHandler) Follow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	targetUserID := c.Param("id")

	// Parse UUIDs for proper comparison
	followerID, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	followingID, err := parseUUID(targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}

	// Prevent self-follow using normalized UUID comparison
	if followerID.Bytes == followingID.Bytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CANNOT_FOLLOW_SELF"})
		return
	}

	// Check if target user exists
	_, err = h.queries.GetUserByID(c.Request.Context(), followingID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "USER_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Insert follow relationship (ON CONFLICT DO NOTHING handles duplicates)
	err = h.queries.FollowUser(c.Request.Context(), sqlc.FollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		log.Printf("Error following user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FOLLOW_SUCCESS"})
}

// Unfollow handles DELETE /users/:id/follow
func (h *SocialHandler) Unfollow(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	targetUserID := c.Param("id")

	// Parse UUIDs
	followerID, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	followingID, err := parseUUID(targetUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}

	// Delete follow relationship
	err = h.queries.UnfollowUser(c.Request.Context(), sqlc.UnfollowUserParams{
		FollowerID:  followerID,
		FollowingID: followingID,
	})
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "UNFOLLOW_SUCCESS"})
}

// LikeRide handles POST /rides/:id/like
func (h *SocialHandler) LikeRide(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	rideID := c.Param("id")

	// Parse UUIDs
	userUUID, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	rideUUID, err := parseUUID(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}

	// Verify ride exists
	ride, err := h.queries.GetRideByID(c.Request.Context(), rideUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "RIDE_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Check if user already liked the ride
	hasLiked, err := h.queries.HasUserLikedRide(c.Request.Context(), sqlc.HasUserLikedRideParams{
		RideID: rideUUID,
		UserID: userUUID,
	})
	if err != nil {
		log.Printf("Error checking like status: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	isNewLike := !hasLiked

	// Insert like (ON CONFLICT DO NOTHING handles duplicates)
	err = h.queries.LikeRide(c.Request.Context(), sqlc.LikeRideParams{
		RideID: rideUUID,
		UserID: userUUID,
	})
	if err != nil {
		log.Printf("Error liking ride: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Send push notification to ride owner asynchronously only if this is a new like
	// Pre-fetch user details in request context before spawning goroutine to reduce async DB load
	if isNewLike {
		rideOwner, err := h.queries.GetUserByID(c.Request.Context(), ride.UserID)
		if err != nil {
			log.Printf("Error fetching ride owner for notification: %v", err)
		} else {
			liker, err := h.queries.GetUserByID(c.Request.Context(), userUUID)
			if err != nil {
				log.Printf("Error fetching liker for notification: %v", err)
			} else {
				// Now spawn goroutine with pre-fetched data (no additional DB queries needed)
				go func(owner, user sqlc.GetUserByIDRow, rideTitle string) {
					if owner.PushToken.Valid {
						notificationService := service.NewNotificationService()
						if err := notificationService.SendLikeNotification(context.Background(), owner.PushToken.String, user.DisplayName, rideTitle); err != nil {
							log.Printf("Error sending like notification: %v", err)
						}
					}
				}(rideOwner, liker, ride.Title.String)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "LIKE_SUCCESS"})
}

// CommentRide handles POST /rides/:id/comments
func (h *SocialHandler) CommentRide(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	rideID := c.Param("id")

	var req struct {
		Content string `json:"content" binding:"required,min=1,max=280"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "VALIDATION_ERROR", "detail": err.Error()})
		return
	}

	// Parse UUIDs
	userUUID, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	rideUUID, err := parseUUID(rideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INVALID_ID"})
		return
	}

	// Verify ride exists
	ride, err := h.queries.GetRideByID(c.Request.Context(), rideUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "RIDE_NOT_FOUND"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		}
		return
	}

	// Create comment
	comment, err := h.queries.CreateComment(c.Request.Context(), sqlc.CreateCommentParams{
		RideID:  rideUUID,
		UserID:  userUUID,
		Content: req.Content,
	})
	if err != nil {
		log.Printf("Error creating comment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Send push notification to ride owner asynchronously
	// Pre-fetch user details in request context before spawning goroutine to reduce async DB load
	rideOwner, err := h.queries.GetUserByID(c.Request.Context(), ride.UserID)
	if err != nil {
		log.Printf("Error fetching ride owner for notification: %v", err)
	} else {
		commenter, err := h.queries.GetUserByID(c.Request.Context(), userUUID)
		if err != nil {
			log.Printf("Error fetching commenter for notification: %v", err)
		} else {
			// Now spawn goroutine with pre-fetched data (no additional DB queries needed)
			go func(owner, user sqlc.GetUserByIDRow, rideTitle string) {
				if owner.PushToken.Valid {
					notificationService := service.NewNotificationService()
					if err := notificationService.SendCommentNotification(context.Background(), owner.PushToken.String, user.DisplayName, rideTitle); err != nil {
						log.Printf("Error sending comment notification: %v", err)
					}
				}
			}(rideOwner, commenter, ride.Title.String)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         comment.ID.String(),
		"content":    comment.Content,
		"created_at": comment.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// FeedItemResponse represents a ride in the feed
type FeedItemResponse struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	StartedAt       string      `json:"started_at"`
	EndedAt         *string     `json:"ended_at"`
	DistanceKm      float64     `json:"distance_km"`
	DurationSeconds int32       `json:"duration_seconds"`
	MaxSpeedKmh     float64     `json:"max_speed_kmh"`
	AvgSpeedKmh     float64     `json:"avg_speed_kmh"`
	ElevationM      float64     `json:"elevation_m"`
	Calories        int32       `json:"calories"`
	RouteSummary    interface{} `json:"route_summary"`
	VehicleType     string      `json:"vehicle_type"`
	VehicleName     string      `json:"vehicle_name"`
	LikeCount       int64       `json:"like_count"`
	CommentCount    int64       `json:"comment_count"`
	UserHasLiked    bool        `json:"user_has_liked"`
	Owner           OwnerInfo   `json:"owner"`
	CreatedAt       string      `json:"created_at"`
}

type OwnerInfo struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

// GetFeed handles GET /feed
func (h *SocialHandler) GetFeed(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Validate pagination to prevent overflow
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Parse UUID
	userUUID, err := parseUUID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "TOKEN_INVALID"})
		return
	}

	offset := int32((page - 1) * limit)

	// Get feed with user's like status
	feeds, err := h.queries.GetFollowingFeedWithUserStatus(c.Request.Context(), sqlc.GetFollowingFeedWithUserStatusParams{
		UserID: userUUID,
		Limit:  int32(limit),
		Offset: offset,
	})
	if err != nil {
		log.Printf("Error fetching feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
		return
	}

	// Transform to response DTOs
	items := make([]FeedItemResponse, 0, len(feeds))
	for _, feed := range feeds {
		endedAtStr := ""
		if feed.EndedAt.Valid {
			endedAtStr = feed.EndedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}

		// Parse route summary
		var routeSummary interface{}
		if feed.RouteSummary != nil {
			_ = json.Unmarshal(feed.RouteSummary, &routeSummary)
		}

		// Convert like and comment counts to int64
		likeCount := int64(0)
		if feed.LikeCount != nil {
			switch v := feed.LikeCount.(type) {
			case float64:
				likeCount = int64(v)
			case int64:
				likeCount = v
			}
		}

		commentCount := int64(0)
		if feed.CommentCount != nil {
			switch v := feed.CommentCount.(type) {
			case float64:
				commentCount = int64(v)
			case int64:
				commentCount = v
			}
		}

		item := FeedItemResponse{
			ID:              feed.ID.String(),
			Title:           feed.Title.String,
			StartedAt:       feed.StartedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
			EndedAt:         func() *string { if endedAtStr == "" { return nil }; return &endedAtStr }(),
			DistanceKm:      feed.DistanceKm,
			DurationSeconds: feed.DurationSeconds,
			MaxSpeedKmh:     feed.MaxSpeedKmh,
			AvgSpeedKmh:     feed.AvgSpeedKmh,
			ElevationM:      feed.ElevationM,
			Calories:        feed.Calories,
			RouteSummary:    routeSummary,
			VehicleType:     string(feed.VehicleType),
			VehicleName:     feed.VehicleName,
			LikeCount:       likeCount,
			CommentCount:    commentCount,
			UserHasLiked:    feed.UserHasLiked,
			Owner: OwnerInfo{
				ID:          feed.UserID.String(),
				Username:    feed.Username,
				DisplayName: feed.DisplayName,
				AvatarURL:   feed.AvatarUrl.String,
			},
			CreatedAt: feed.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  items,
		"page":  page,
		"limit": limit,
	})
}
