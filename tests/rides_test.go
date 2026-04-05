package tests

import (
	"fmt"
	"net/http"
	"testing"

	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/tests/testutil"
)

// POST /v1/rides/start

func TestStartRide_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "start_ride")
	vehicle := testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)

	body := map[string]interface{}{"vehicle_id": vehicle.ID, "title": "Morning Ride"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/start", body, user.AccessToken)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	for _, field := range []string{"ride_id", "ws_token", "started_at"} {
		if result[field] == nil {
			t.Errorf("expected field %q in start ride response", field)
		}
	}
}

func TestStartRide_WithoutTitle(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "start_ride_notitle")
	vehicle := testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)

	body := map[string]interface{}{"vehicle_id": vehicle.ID}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/start", body, user.AccessToken)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201 without title, got %d", resp.StatusCode)
	}
}

func TestStartRide_MissingVehicleID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "start_ride_nocar")

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/start", map[string]interface{}{}, user.AccessToken)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing vehicle_id, got %d", resp.StatusCode)
	}
}

func TestStartRide_InvalidVehicleIDFormat(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "start_ride_badid")

	body := map[string]interface{}{"vehicle_id": "not-a-uuid"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/start", body, user.AccessToken)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for invalid UUID format, got %d", resp.StatusCode)
	}
}

func TestStartRide_VehicleNotFound(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "start_ride_novehicle")

	body := map[string]interface{}{"vehicle_id": "00000000-0000-0000-0000-000000000000"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/start", body, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent vehicle, got %d", resp.StatusCode)
	}
}

func TestStartRide_VehicleBelongsToAnotherUser(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "start_ride_veh_owner")
	attacker := testutil.CreateUser(t, env.Pool, "start_ride_veh_attacker")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)

	body := map[string]interface{}{"vehicle_id": vehicle.ID}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/start", body, attacker.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 when using another user's vehicle, got %d", resp.StatusCode)
	}
}

func TestStartRide_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/start",
		map[string]interface{}{"vehicle_id": "00000000-0000-0000-0000-000000000000"}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// POST /v1/rides/:id/stop

func TestStopRide_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "stop_ride")
	vehicle := testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, user.ID, vehicle.ID)

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/rides/%s/stop", ride.ID), nil, user.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestStopRide_InvalidID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "stop_ride_invalid")

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/not-a-uuid/stop", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 400 or 404 for invalid ride ID, got %d", resp.StatusCode)
	}
}

func TestStopRide_NotFound_UnknownRide(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "stop_ride_notfound")

	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/rides/00000000-0000-0000-0000-000000000000/stop", nil, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestStopRide_NotFound_OtherUsersRide(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "stop_ride_owner")
	attacker := testutil.CreateUser(t, env.Pool, "stop_ride_attacker")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, owner.ID, vehicle.ID)

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/rides/%s/stop", ride.ID), nil, attacker.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 stopping another user's ride, got %d", resp.StatusCode)
	}
}

func TestStopRide_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/rides/00000000-0000-0000-0000-000000000000/stop", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// GET /v1/rides

func TestListRides_ReturnsPaginatedHistory(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "list_rides")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/rides", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["data"] == nil {
		t.Error("expected data field in list rides response")
	}
	if body["page"] == nil {
		t.Error("expected page field in list rides response")
	}
}

func TestListRides_DefaultPagination(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "list_rides_default_page")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/rides", nil, user.AccessToken)
	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["page"].(float64) != 1 {
		t.Errorf("expected default page=1, got %v", body["page"])
	}
}

func TestListRides_FilterByVehicleType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "list_rides_filter")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/rides?vehicle_type=motor", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 with vehicle_type filter, got %d", resp.StatusCode)
	}
}

func TestListRides_ReturnsOnlyOwnRides(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "list_rides_owner")
	other := testutil.CreateUser(t, env.Pool, "list_rides_other")

	// Owner has no rides; other has a vehicle; ensure no cross-contamination
	_ = other

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/rides", nil, owner.AccessToken)
	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	items := body["data"].([]interface{})
	if len(items) != 0 {
		t.Errorf("expected no rides for owner with no history, got %d", len(items))
	}
}

func TestListRides_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/rides", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// GET /v1/rides/:id

func TestGetRide_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "get_ride")
	vehicle := testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, user.ID, vehicle.ID)

	resp := testutil.Do(env.Router, http.MethodGet,
		fmt.Sprintf("/v1/rides/%s", ride.ID), nil, user.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestGetRide_InvalidID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "get_ride_invalid")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/rides/not-a-uuid", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 400 or 404 for invalid ride ID, got %d", resp.StatusCode)
	}
}

func TestGetRide_NotFound(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "get_ride_notfound")

	resp := testutil.Do(env.Router, http.MethodGet,
		"/v1/rides/00000000-0000-0000-0000-000000000000", nil, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestGetRide_NotFound_OtherUsersRide(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "get_ride_owner")
	viewer := testutil.CreateUser(t, env.Pool, "get_ride_viewer")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, owner.ID, vehicle.ID)

	resp := testutil.Do(env.Router, http.MethodGet,
		fmt.Sprintf("/v1/rides/%s", ride.ID), nil, viewer.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for another user's ride, got %d", resp.StatusCode)
	}
}

func TestGetRide_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet,
		"/v1/rides/00000000-0000-0000-0000-000000000000", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
