package tests

import (
	"net/http"
	"testing"

	"github.com/nashirabbash/trackride/tests/testutil"
)

// GET /v1/leaderboard (public — no auth required by router, but covered with authed user)

func TestLeaderboard_DefaultParams(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "leaderboard_default")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/leaderboard", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["data"] == nil {
		t.Error("expected data field in leaderboard response")
	}
	if body["period_type"] != "weekly" {
		t.Errorf("expected default period_type=weekly, got %v", body["period_type"])
	}
}

func TestLeaderboard_FilterByVehicleType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "leaderboard_vtype")

	cases := []string{"motor", "mobil", "sepeda"}
	for _, vt := range cases {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard?vehicle_type="+vt, nil, user.AccessToken)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for vehicle_type=%s, got %d", vt, resp.StatusCode)
		}
	}
}

func TestLeaderboard_FilterByPeriodType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "leaderboard_period")

	for _, period := range []string{"weekly", "monthly"} {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard?period_type="+period, nil, user.AccessToken)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for period_type=%s, got %d", period, resp.StatusCode)
		}
	}
}

func TestLeaderboard_InvalidPage(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "leaderboard_bad_page")

	for _, page := range []string{"0", "-1", "abc"} {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard?page="+page, nil, user.AccessToken)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for page=%s, got %d", page, resp.StatusCode)
		}
	}
}

func TestLeaderboard_InvalidLimit(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "leaderboard_bad_limit")

	for _, limit := range []string{"0", "-5", "abc", "101"} {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard?limit="+limit, nil, user.AccessToken)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for limit=%s, got %d", limit, resp.StatusCode)
		}
	}
}

func TestLeaderboard_InvalidPeriodType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "leaderboard_bad_period")

	resp := testutil.Do(env.Router, http.MethodGet,
		"/v1/leaderboard?period_type=all-time", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid period_type, got %d", resp.StatusCode)
	}
}

func TestLeaderboard_InvalidVehicleType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "leaderboard_bad_vtype")

	resp := testutil.Do(env.Router, http.MethodGet,
		"/v1/leaderboard?vehicle_type=helicopter", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid vehicle_type, got %d", resp.StatusCode)
	}
}

// GET /v1/leaderboard/friends

func TestLeaderboardFriends_ReturnsOK(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/leaderboard/friends", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestLeaderboardFriends_EmptyWhenNoFollows(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb_empty")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/leaderboard/friends", nil, user.AccessToken)
	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	items := body["data"].([]interface{})
	if len(items) != 0 {
		t.Errorf("expected empty friends leaderboard, got %d entries", len(items))
	}
}

func TestLeaderboardFriends_FilterByVehicleType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb_vtype")

	for _, vt := range []string{"motor", "mobil", "sepeda"} {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard/friends?vehicle_type="+vt, nil, user.AccessToken)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for vehicle_type=%s, got %d", vt, resp.StatusCode)
		}
	}
}

func TestLeaderboardFriends_FilterByPeriodType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb_period")

	for _, period := range []string{"weekly", "monthly"} {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard/friends?period_type="+period, nil, user.AccessToken)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for period_type=%s, got %d", period, resp.StatusCode)
		}
	}
}

func TestLeaderboardFriends_InvalidPage(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb_bad_page")

	for _, page := range []string{"0", "-1", "abc"} {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard/friends?page="+page, nil, user.AccessToken)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for page=%s, got %d", page, resp.StatusCode)
		}
	}
}

func TestLeaderboardFriends_InvalidLimit(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb_bad_limit")

	for _, limit := range []string{"0", "-5", "101", "abc"} {
		resp := testutil.Do(env.Router, http.MethodGet,
			"/v1/leaderboard/friends?limit="+limit, nil, user.AccessToken)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 for limit=%s, got %d", limit, resp.StatusCode)
		}
	}
}

func TestLeaderboardFriends_InvalidPeriodType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb_bad_period")

	resp := testutil.Do(env.Router, http.MethodGet,
		"/v1/leaderboard/friends?period_type=all-time", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid period_type, got %d", resp.StatusCode)
	}
}

func TestLeaderboardFriends_InvalidVehicleType(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "friends_lb_bad_vtype")

	resp := testutil.Do(env.Router, http.MethodGet,
		"/v1/leaderboard/friends?vehicle_type=tank", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid vehicle_type, got %d", resp.StatusCode)
	}
}

func TestLeaderboardFriends_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/leaderboard/friends", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestLeaderboardFriends_Unauthorized_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/leaderboard/friends", nil, "bad.token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
