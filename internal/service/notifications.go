package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// NotificationService handles push notifications via Expo Push API
type NotificationService struct {
	client *http.Client
}

// NewNotificationService creates a new notification service with timeout
func NewNotificationService() *NotificationService {
	return &NotificationService{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// expoPushPayload represents the request payload for Expo Push API
type expoPushPayload struct {
	To    string `json:"to"`
	Title string `json:"title"`
	Body  string `json:"body"`
	Sound string `json:"sound"`
	Badge int    `json:"badge"`
}

// SendLikeNotification sends a push notification when a ride receives a like
func (ns *NotificationService) SendLikeNotification(ctx context.Context, pushToken, likerName, rideTitle string) error {
	title := "New Like!"
	body := likerName + " liked your ride \"" + rideTitle + "\""

	payload := expoPushPayload{
		To:    pushToken,
		Title: title,
		Body:  body,
		Sound: "default",
		Badge: 1,
	}

	return ns.sendNotification(ctx, payload)
}

// SendCommentNotification sends a push notification when a ride receives a comment
func (ns *NotificationService) SendCommentNotification(ctx context.Context, pushToken, commenterName, rideTitle string) error {
	title := "New Comment!"
	body := commenterName + " commented on your ride \"" + rideTitle + "\""

	payload := expoPushPayload{
		To:    pushToken,
		Title: title,
		Body:  body,
		Sound: "default",
		Badge: 1,
	}

	return ns.sendNotification(ctx, payload)
}

// sendNotification sends the actual HTTP request to Expo Push API with proper error handling
func (ns *NotificationService) sendNotification(ctx context.Context, payload expoPushPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://exp.host/--/api/v2/push/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := ns.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	// Drain response body to prevent connection leaks
	_, _ = io.ReadAll(resp.Body)

	// Check response status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Expo API returned status %d", resp.StatusCode)
	}

	return nil
}
