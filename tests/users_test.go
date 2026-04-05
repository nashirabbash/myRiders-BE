package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nashirabbash/trackride/tests/testutil"
)

// GET /v1/users/me

func TestGetMe_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "getme")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/users/me", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["id"] != user.ID {
		t.Errorf("expected id=%s, got %v", user.ID, body["id"])
	}
	if body["username"] != user.Username {
		t.Errorf("expected username=%s, got %v", user.Username, body["username"])
	}
}

func TestGetMe_ExcludesSensitiveFields(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "getme_sensitive")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/users/me", nil, user.AccessToken)
	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["password_hash"] != nil {
		t.Error("response must not include password_hash")
	}
}

func TestGetMe_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/users/me", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestGetMe_Unauthorized_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/users/me", nil, "bad.token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// PUT /v1/users/me

func TestUpdateMe_DisplayName(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "update_name")

	body := map[string]string{"display_name": "Updated Name"}
	resp := testutil.Do(env.Router, http.MethodPut, "/v1/users/me", body, user.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["display_name"] != "Updated Name" {
		t.Errorf("expected display_name='Updated Name', got %v", result["display_name"])
	}
}

func TestUpdateMe_AvatarURL(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "update_avatar")

	body := map[string]string{"avatar_url": "https://example.com/avatar.png"}
	resp := testutil.Do(env.Router, http.MethodPut, "/v1/users/me", body, user.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["avatar_url"] != "https://example.com/avatar.png" {
		t.Errorf("expected avatar_url to be updated, got %v", result["avatar_url"])
	}
}

func TestUpdateMe_PartialUpdate_KeepsExistingValues(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "partial_update")

	// Only update display_name — avatar_url should remain unchanged
	body := map[string]string{"display_name": "Only Name Changed"}
	resp := testutil.Do(env.Router, http.MethodPut, "/v1/users/me", body, user.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["display_name"] != "Only Name Changed" {
		t.Errorf("display_name not updated correctly: %v", result["display_name"])
	}
}

func TestUpdateMe_InvalidAvatarURL(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "invalid_avatar")

	body := map[string]string{"avatar_url": "not-a-url"}
	resp := testutil.Do(env.Router, http.MethodPut, "/v1/users/me", body, user.AccessToken)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for invalid URL, got %d", resp.StatusCode)
	}
}

func TestUpdateMe_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	body := map[string]string{"display_name": "Hacker"}
	resp := testutil.Do(env.Router, http.MethodPut, "/v1/users/me", body, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// GET /v1/users/:id

func TestGetProfile_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "profile_owner")
	viewer := testutil.CreateUser(t, env.Pool, "profile_viewer")

	resp := testutil.Do(env.Router, http.MethodGet, fmt.Sprintf("/v1/users/%s", owner.ID), nil, viewer.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestGetProfile_DoesNotExposePrivateFields(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "profile_private")
	viewer := testutil.CreateUser(t, env.Pool, "profile_viewer2")

	resp := testutil.Do(env.Router, http.MethodGet, fmt.Sprintf("/v1/users/%s", owner.ID), nil, viewer.AccessToken)

	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["email"] != nil {
		t.Error("public profile must not expose email")
	}
	if body["password_hash"] != nil {
		t.Error("public profile must not expose password_hash")
	}
}

func TestGetProfile_InvalidID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	viewer := testutil.CreateUser(t, env.Pool, "profile_invalid")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/users/not-a-uuid", nil, viewer.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	viewer := testutil.CreateUser(t, env.Pool, "profile_notfound")

	resp := testutil.Do(env.Router, http.MethodGet,
		"/v1/users/00000000-0000-0000-0000-000000000000", nil, viewer.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
