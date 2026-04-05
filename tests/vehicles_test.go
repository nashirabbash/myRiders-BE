package tests

import (
	"fmt"
	"net/http"
	"testing"

	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/tests/testutil"
)

// GET /v1/vehicles

func TestListVehicles_ReturnsOwnedVehicles(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "list_vehicles")
	testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)
	testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMobil)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/vehicles", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var body []map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if len(body) != 2 {
		t.Errorf("expected 2 vehicles, got %d", len(body))
	}
}

func TestListVehicles_EmptyWhenNoVehicles(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "no_vehicles")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/vehicles", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var body []map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if len(body) != 0 {
		t.Errorf("expected empty list, got %d items", len(body))
	}
}

func TestListVehicles_DoesNotReturnOtherUsersVehicles(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "vehicle_owner")
	other := testutil.CreateUser(t, env.Pool, "vehicle_other")
	testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/vehicles", nil, other.AccessToken)
	var body []map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if len(body) != 0 {
		t.Errorf("expected 0 vehicles for other user, got %d", len(body))
	}
}

func TestListVehicles_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/vehicles", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// POST /v1/vehicles

func TestCreateVehicle_MinimumPayload(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "create_vehicle_min")

	body := map[string]interface{}{"type": "motor", "name": "Honda Beat"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/vehicles", body, user.AccessToken)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["id"] == nil {
		t.Error("expected id in response")
	}
	if result["type"] != "motor" {
		t.Errorf("expected type=motor, got %v", result["type"])
	}
}

func TestCreateVehicle_WithOptionalFields(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "create_vehicle_opt")

	body := map[string]interface{}{
		"type": "sepeda", "name": "Trek Marlin", "brand": "Trek", "color": "Red",
	}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/vehicles", body, user.AccessToken)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["brand"] != "Trek" {
		t.Errorf("expected brand=Trek, got %v", result["brand"])
	}
}

func TestCreateVehicle_MissingRequiredFields(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "create_vehicle_missing")

	cases := []map[string]interface{}{
		{"name": "Honda Beat"},           // missing type
		{"type": "motor"},                // missing name
	}
	for _, body := range cases {
		resp := testutil.Do(env.Router, http.MethodPost, "/v1/vehicles", body, user.AccessToken)
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d for body %v", resp.StatusCode, body)
		}
	}
}

func TestCreateVehicle_InvalidType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "create_vehicle_badtype")

	body := map[string]interface{}{"type": "helicopter", "name": "Airwolf"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/vehicles", body, user.AccessToken)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for invalid type, got %d", resp.StatusCode)
	}
}

func TestCreateVehicle_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	body := map[string]interface{}{"type": "motor", "name": "X"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/vehicles", body, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// PUT /v1/vehicles/:id

func TestUpdateVehicle_FullUpdate(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "update_vehicle")
	vehicle := testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)

	body := map[string]interface{}{"type": "mobil", "name": "Avanza", "brand": "Toyota", "color": "White"}
	resp := testutil.Do(env.Router, http.MethodPut,
		fmt.Sprintf("/v1/vehicles/%s", vehicle.ID), body, user.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["type"] != "mobil" {
		t.Errorf("expected type=mobil, got %v", result["type"])
	}
}

func TestUpdateVehicle_InvalidID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "update_vehicle_invalid_id")

	body := map[string]interface{}{"name": "New Name"}
	resp := testutil.Do(env.Router, http.MethodPut, "/v1/vehicles/not-a-uuid", body, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateVehicle_NotFound_UnknownID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "update_vehicle_notfound")

	body := map[string]interface{}{"name": "Ghost"}
	resp := testutil.Do(env.Router, http.MethodPut,
		"/v1/vehicles/00000000-0000-0000-0000-000000000000", body, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestUpdateVehicle_NotFound_OtherUsersVehicle(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "update_vehicle_owner")
	attacker := testutil.CreateUser(t, env.Pool, "update_vehicle_attacker")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)

	body := map[string]interface{}{"name": "Stolen"}
	resp := testutil.Do(env.Router, http.MethodPut,
		fmt.Sprintf("/v1/vehicles/%s", vehicle.ID), body, attacker.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 when updating another user's vehicle, got %d", resp.StatusCode)
	}
}

func TestUpdateVehicle_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPut,
		"/v1/vehicles/00000000-0000-0000-0000-000000000000", map[string]interface{}{}, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// DELETE /v1/vehicles/:id

func TestDeleteVehicle_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "delete_vehicle")
	vehicle := testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)

	resp := testutil.Do(env.Router, http.MethodDelete,
		fmt.Sprintf("/v1/vehicles/%s", vehicle.ID), nil, user.AccessToken)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204, got %d", resp.StatusCode)
	}
}

func TestDeleteVehicle_InvalidID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "delete_vehicle_invalid_id")

	resp := testutil.Do(env.Router, http.MethodDelete, "/v1/vehicles/not-a-uuid", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestDeleteVehicle_NotFound_UnknownID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "delete_vehicle_notfound")

	resp := testutil.Do(env.Router, http.MethodDelete,
		"/v1/vehicles/00000000-0000-0000-0000-000000000000", nil, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestDeleteVehicle_NotFound_OtherUsersVehicle(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "delete_vehicle_owner")
	attacker := testutil.CreateUser(t, env.Pool, "delete_vehicle_attacker")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)

	resp := testutil.Do(env.Router, http.MethodDelete,
		fmt.Sprintf("/v1/vehicles/%s", vehicle.ID), nil, attacker.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 when deleting another user's vehicle, got %d", resp.StatusCode)
	}
}

func TestDeleteVehicle_FailsWhenVehicleHasActiveRide(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "delete_vehicle_active_ride")
	vehicle := testutil.CreateVehicle(t, env.Pool, user.ID, dbsqlc.VehicleTypeMotor)
	testutil.CreateActiveRide(t, env.Pool, user.ID, vehicle.ID)

	resp := testutil.Do(env.Router, http.MethodDelete,
		fmt.Sprintf("/v1/vehicles/%s", vehicle.ID), nil, user.AccessToken)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 for vehicle in use, got %d", resp.StatusCode)
	}
}

func TestDeleteVehicle_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodDelete,
		"/v1/vehicles/00000000-0000-0000-0000-000000000000", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
