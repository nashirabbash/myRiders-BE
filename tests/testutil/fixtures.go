package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	jwtpkg "github.com/nashirabbash/trackride/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

const (
	TestPassword     = "Test1234!"
	TestAccessSecret = "test-access-secret-at-least-32-characters-long"
)

// TestUser holds a created user and a valid access token for them.
type TestUser struct {
	ID          string
	Username    string
	Email       string
	DisplayName string
	AccessToken string
	RefreshToken string
}

// CreateUser inserts a user directly into the DB and returns a TestUser with
// a signed access token. suffix is appended to username/email to avoid
// collisions when multiple users are needed in a single test.
func CreateUser(t *testing.T, pool *pgxpool.Pool, suffix string) TestUser {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(TestPassword), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("fixtures: hash password: %v", err)
	}

	username := fmt.Sprintf("testuser_%s", suffix)
	email := fmt.Sprintf("test_%s@example.com", suffix)
	displayName := fmt.Sprintf("Test User %s", suffix)

	queries := dbsqlc.New(pool)
	user, err := queries.CreateUser(context.Background(), dbsqlc.CreateUserParams{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
	})
	if err != nil {
		t.Fatalf("fixtures: CreateUser: %v", err)
	}

	userID := user.ID.String()
	accessToken, err := jwtpkg.GenerateAccessToken(userID, TestAccessSecret, 1*time.Hour)
	if err != nil {
		t.Fatalf("fixtures: GenerateAccessToken: %v", err)
	}
	refreshToken, err := jwtpkg.GenerateRefreshToken(userID, "test-refresh-secret-at-least-32-characters-long", 720*time.Hour)
	if err != nil {
		t.Fatalf("fixtures: GenerateRefreshToken: %v", err)
	}

	return TestUser{
		ID:           userID,
		Username:     username,
		Email:        email,
		DisplayName:  displayName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
}

// TestVehicle holds a created vehicle.
type TestVehicle struct {
	ID     string
	UserID string
	Type   string
	Name   string
}

// CreateVehicle inserts a vehicle for the given user.
func CreateVehicle(t *testing.T, pool *pgxpool.Pool, userID string, vehicleType dbsqlc.VehicleType) TestVehicle {
	t.Helper()

	queries := dbsqlc.New(pool)

	var userUUID interface{}
	var uid dbsqlc.Vehicle
	_ = uid // unused; we use raw SQL via pgxpool for simplicity

	// Use the Queries method with a parsed UUID
	var pgUserID [16]byte
	if err := parseUUIDBytes(userID, &pgUserID); err != nil {
		t.Fatalf("fixtures: parse user UUID: %v", err)
	}

	from, _ := parseUUIDForQuery(userID)
	_ = userUUID

	vehicle, err := queries.CreateVehicle(context.Background(), dbsqlc.CreateVehicleParams{
		UserID: from,
		Type:   vehicleType,
		Name:   fmt.Sprintf("Test %s", string(vehicleType)),
	})
	if err != nil {
		t.Fatalf("fixtures: CreateVehicle: %v", err)
	}

	return TestVehicle{
		ID:     vehicle.ID.String(),
		UserID: userID,
		Type:   string(vehicleType),
		Name:   vehicle.Name,
	}
}

// TestRide holds a created active ride.
type TestRide struct {
	ID        string
	UserID    string
	VehicleID string
}

// CreateActiveRide inserts an active ride for the given user and vehicle.
func CreateActiveRide(t *testing.T, pool *pgxpool.Pool, userID, vehicleID string) TestRide {
	t.Helper()

	queries := dbsqlc.New(pool)

	userUUID, err := parseUUIDForQuery(userID)
	if err != nil {
		t.Fatalf("fixtures: parse user UUID: %v", err)
	}

	vehicleUUID, err := parseUUIDForQuery(vehicleID)
	if err != nil {
		t.Fatalf("fixtures: parse vehicle UUID: %v", err)
	}

	ride, err := queries.CreateRide(context.Background(), dbsqlc.CreateRideParams{
		UserID:    userUUID,
		VehicleID: vehicleUUID,
	})
	if err != nil {
		t.Fatalf("fixtures: CreateRide: %v", err)
	}

	return TestRide{
		ID:        ride.ID.String(),
		UserID:    userID,
		VehicleID: vehicleID,
	}
}

// ExpiredToken returns a JWT access token that is already expired.
func ExpiredToken(userID string) string {
	token, _ := jwtpkg.GenerateAccessToken(userID, TestAccessSecret, -1*time.Hour)
	return token
}
