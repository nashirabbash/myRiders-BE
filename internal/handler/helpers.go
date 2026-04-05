package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nashirabbash/trackride/internal/errors"
)

// parseUUID converts a string to pgtype.UUID
func parseUUID(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(s)
	return id, err
}

// optionalString returns a pgtype.Text for optional string values
func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// optionalStringPgtype converts an optional string to pgtype.Text
func optionalStringPgtype(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// optionalStringPgtype2 converts a pointer to string to pgtype.Text (nil pointer = NULL)
func optionalStringPgtype2(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

// derefString dereferences a pointer to string, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// RespondWithError responds to a request with an error, using domain error types for proper status codes
func RespondWithError(c *gin.Context, err error) {
	if de := errors.AsDomainError(err); de != nil {
		c.JSON(de.StatusCode, gin.H{"error": de.Code})
		return
	}

	// Default to internal server error for unknown error types
	c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR"})
}

// RespondWithValidationError responds with a validation error
func RespondWithValidationError(c *gin.Context, details string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"error":  "VALIDATION_ERROR",
		"detail": details,
	})
}
