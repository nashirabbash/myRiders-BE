package errors

import (
	"errors"
	"net/http"
)

// DomainError represents an error with a domain-specific code and HTTP status
type DomainError struct {
	Code       string
	Message    string
	StatusCode int
}

// Error implements the error interface
func (e *DomainError) Error() string {
	return e.Code
}

// New creates a new domain error with a code and message
func New(code string, message string, statusCode int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Common domain errors
var (
	// Auth errors
	ErrEmailTaken = New("EMAIL_TAKEN", "This email is already registered", http.StatusConflict)
	ErrUsernameTaken = New("USERNAME_TAKEN", "This username is already taken", http.StatusConflict)
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrTokenInvalid = New("TOKEN_INVALID", "Invalid or expired token", http.StatusUnauthorized)
	ErrTokenExpired = New("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized)
	ErrUnauthorized = New("UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)

	// Vehicle errors
	ErrVehicleNotFound = New("VEHICLE_NOT_FOUND", "Vehicle not found", http.StatusNotFound)
	ErrVehicleInUse = New("VEHICLE_IN_USE", "Cannot delete vehicle with active rides", http.StatusConflict)

	// Ride errors
	ErrRideNotFound = New("RIDE_NOT_FOUND", "Ride not found", http.StatusNotFound)
	ErrRideNotActive = New("RIDE_NOT_ACTIVE", "Ride is not currently active", http.StatusBadRequest)
	ErrRideAlreadyActive = New("RIDE_ALREADY_ACTIVE", "User already has an active ride", http.StatusBadRequest)

	// Validation errors
	ErrValidationFailed = New("VALIDATION_ERROR", "Invalid request data", http.StatusUnprocessableEntity)
	ErrInvalidID = New("INVALID_ID", "Invalid ID format", http.StatusBadRequest)

	// Generic errors
	ErrInternalServerError = New("INTERNAL_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrNotFound = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrForbidden = New("FORBIDDEN", "You do not have permission to access this resource", http.StatusForbidden)
)

// IsDomainError checks if an error is a DomainError
func IsDomainError(err error) bool {
	_, ok := err.(*DomainError)
	return ok
}

// AsDomainError returns the DomainError if err is one, otherwise returns nil
func AsDomainError(err error) *DomainError {
	var de *DomainError
	if errors.As(err, &de) {
		return de
	}
	return nil
}
