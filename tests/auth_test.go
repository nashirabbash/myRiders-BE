package tests

import (
	"net/http"
	"testing"

	"github.com/nashirabbash/trackride/tests/testutil"
)

// POST /v1/auth/register

func TestRegister_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	body := map[string]string{
		"username":     "newuser",
		"email":        "newuser@example.com",
		"password":     testutil.TestPassword,
		"display_name": "New User",
	}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/register", body, "")

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	for _, field := range []string{"id", "username", "email", "display_name", "access_token", "refresh_token", "expires_in"} {
		if result[field] == nil {
			t.Errorf("expected field %q in register response", field)
		}
	}
}

func TestRegister_MissingRequiredFields(t *testing.T) {
	env := testutil.Setup(t)

	cases := []struct {
		name string
		body map[string]string
	}{
		{"missing_username", map[string]string{"email": "a@example.com", "password": testutil.TestPassword, "display_name": "A"}},
		{"missing_email", map[string]string{"username": "userx", "password": testutil.TestPassword, "display_name": "A"}},
		{"missing_password", map[string]string{"username": "userx", "email": "a@example.com", "display_name": "A"}},
		{"missing_display_name", map[string]string{"username": "userx", "email": "a@example.com", "password": testutil.TestPassword}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/register", tc.body, "")
			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Errorf("expected 422, got %d", resp.StatusCode)
			}
		})
	}
}

func TestRegister_InvalidEmailFormat(t *testing.T) {
	env := testutil.Setup(t)

	body := map[string]string{
		"username": "userx", "email": "not-an-email",
		"password": testutil.TestPassword, "display_name": "A",
	}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/register", body, "")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	env := testutil.Setup(t)

	body := map[string]string{
		"username": "userx", "email": "a@example.com",
		"password": "short", "display_name": "A",
	}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/register", body, "")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	testutil.CreateUser(t, env.Pool, "dup_email")

	body := map[string]string{
		"username": "different_user", "email": "test_dup_email@example.com",
		"password": testutil.TestPassword, "display_name": "Dup",
	}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/register", body, "")
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 for duplicate email, got %d", resp.StatusCode)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	testutil.CreateUser(t, env.Pool, "dup_user")

	body := map[string]string{
		"username": "testuser_dup_user", "email": "unique@example.com",
		"password": testutil.TestPassword, "display_name": "Dup",
	}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/register", body, "")
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409 for duplicate username, got %d", resp.StatusCode)
	}
}

// POST /v1/auth/login

func TestLogin_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "login_ok")

	body := map[string]string{"email": user.Email, "password": testutil.TestPassword}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/login", body, "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	for _, field := range []string{"id", "username", "email", "access_token", "refresh_token", "expires_in"} {
		if result[field] == nil {
			t.Errorf("expected field %q in login response", field)
		}
	}
}

func TestLogin_MissingFields(t *testing.T) {
	env := testutil.Setup(t)

	cases := []map[string]string{
		{"password": testutil.TestPassword},
		{"email": "a@example.com"},
	}
	for _, body := range cases {
		resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/login", body, "")
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("expected 422, got %d", resp.StatusCode)
		}
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	env := testutil.Setup(t)

	body := map[string]string{"email": "nobody@example.com", "password": testutil.TestPassword}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/login", body, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "bad_pass")

	body := map[string]string{"email": user.Email, "password": "WrongPass999!"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/login", body, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// POST /v1/auth/refresh

func TestRefresh_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "refresh_ok")

	body := map[string]string{"refresh_token": user.RefreshToken}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/refresh", body, "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["access_token"] == nil {
		t.Error("expected access_token in refresh response")
	}
	if result["expires_in"] == nil {
		t.Error("expected expires_in in refresh response")
	}
}

func TestRefresh_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/refresh", map[string]string{}, "")
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", resp.StatusCode)
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	body := map[string]string{"refresh_token": "totally.invalid.token"}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/refresh", body, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRefresh_ExpiredToken(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "expired_refresh")
	expired := testutil.ExpiredToken(user.ID)

	body := map[string]string{"refresh_token": expired}
	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/refresh", body, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", resp.StatusCode)
	}
}

// POST /v1/auth/logout

func TestLogout_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "logout_ok")

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/logout", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLogout_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/logout", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLogout_Unauthorized_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/auth/logout", nil, "invalid.token.here")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
