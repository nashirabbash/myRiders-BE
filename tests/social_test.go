package tests

import (
	"fmt"
	"net/http"
	"testing"

	dbsqlc "github.com/nashirabbash/trackride/internal/db/sqlc"
	"github.com/nashirabbash/trackride/tests/testutil"
)

// GET /v1/feed

func TestGetFeed_ReturnsOK(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "feed_user")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/feed", nil, user.AccessToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestGetFeed_EmptyWhenNoFollows(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "feed_empty")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/feed", nil, user.AccessToken)
	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	items := body["data"].([]interface{})
	if len(items) != 0 {
		t.Errorf("expected empty feed, got %d items", len(items))
	}
}

func TestGetFeed_DefaultPagination(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "feed_pagination")

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/feed", nil, user.AccessToken)
	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["page"].(float64) != 1 {
		t.Errorf("expected default page=1, got %v", body["page"])
	}
}

func TestGetFeed_Unauthorized_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/v1/feed", nil, "bad.token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// POST /v1/users/:id/follow

func TestFollow_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	follower := testutil.CreateUser(t, env.Pool, "follow_a")
	target := testutil.CreateUser(t, env.Pool, "follow_b")

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/follow", target.ID), nil, follower.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestFollow_Idempotent(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	follower := testutil.CreateUser(t, env.Pool, "follow_idem_a")
	target := testutil.CreateUser(t, env.Pool, "follow_idem_b")

	// Follow twice; second request should also succeed (ON CONFLICT DO NOTHING)
	testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/follow", target.ID), nil, follower.AccessToken)

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/follow", target.ID), nil, follower.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for duplicate follow, got %d", resp.StatusCode)
	}
}

func TestFollow_CannotFollowSelf(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "follow_self")

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/follow", user.ID), nil, user.AccessToken)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for self-follow, got %d", resp.StatusCode)
	}
}

func TestFollow_InvalidTargetID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "follow_badid")

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/users/not-a-uuid/follow", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestFollow_TargetNotFound(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "follow_notfound")

	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/users/00000000-0000-0000-0000-000000000000/follow", nil, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestFollow_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/users/00000000-0000-0000-0000-000000000000/follow", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// DELETE /v1/users/:id/follow

func TestUnfollow_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	follower := testutil.CreateUser(t, env.Pool, "unfollow_a")
	target := testutil.CreateUser(t, env.Pool, "unfollow_b")

	// Follow first
	testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/users/%s/follow", target.ID), nil, follower.AccessToken)

	// Then unfollow
	resp := testutil.Do(env.Router, http.MethodDelete,
		fmt.Sprintf("/v1/users/%s/follow", target.ID), nil, follower.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestUnfollow_SafeWhenNotFollowing(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "unfollow_safe_a")
	target := testutil.CreateUser(t, env.Pool, "unfollow_safe_b")

	resp := testutil.Do(env.Router, http.MethodDelete,
		fmt.Sprintf("/v1/users/%s/follow", target.ID), nil, user.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for idempotent unfollow, got %d", resp.StatusCode)
	}
}

func TestUnfollow_InvalidTargetID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "unfollow_badid")

	resp := testutil.Do(env.Router, http.MethodDelete, "/v1/users/not-a-uuid/follow", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUnfollow_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodDelete,
		"/v1/users/00000000-0000-0000-0000-000000000000/follow", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// POST /v1/rides/:id/like

func TestLikeRide_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "like_owner")
	liker := testutil.CreateUser(t, env.Pool, "like_user")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, owner.ID, vehicle.ID)

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/rides/%s/like", ride.ID), nil, liker.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}
}

func TestLikeRide_DuplicateLikeIdempotent(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "like_idem_owner")
	liker := testutil.CreateUser(t, env.Pool, "like_idem_user")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, owner.ID, vehicle.ID)

	testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/rides/%s/like", ride.ID), nil, liker.AccessToken)

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/rides/%s/like", ride.ID), nil, liker.AccessToken)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for duplicate like, got %d", resp.StatusCode)
	}
}

func TestLikeRide_InvalidRideID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "like_badid")

	resp := testutil.Do(env.Router, http.MethodPost, "/v1/rides/not-a-uuid/like", nil, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestLikeRide_RideNotFound(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "like_notfound")

	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/rides/00000000-0000-0000-0000-000000000000/like", nil, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestLikeRide_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/rides/00000000-0000-0000-0000-000000000000/like", nil, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

// POST /v1/rides/:id/comments

func TestCommentRide_Success(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "comment_owner")
	commenter := testutil.CreateUser(t, env.Pool, "comment_user")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, owner.ID, vehicle.ID)

	body := map[string]string{"content": "Great ride!"}
	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/rides/%s/comments", ride.ID), body, commenter.AccessToken)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, testutil.BodyString(resp))
	}

	var result map[string]interface{}
	testutil.ParseJSON(t, resp, &result)

	if result["id"] == nil {
		t.Error("expected id in comment response")
	}
	if result["content"] != "Great ride!" {
		t.Errorf("expected content='Great ride!', got %v", result["content"])
	}
}

func TestCommentRide_MissingContent(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	owner := testutil.CreateUser(t, env.Pool, "comment_missing_owner")
	commenter := testutil.CreateUser(t, env.Pool, "comment_missing_user")
	vehicle := testutil.CreateVehicle(t, env.Pool, owner.ID, dbsqlc.VehicleTypeMotor)
	ride := testutil.CreateActiveRide(t, env.Pool, owner.ID, vehicle.ID)

	resp := testutil.Do(env.Router, http.MethodPost,
		fmt.Sprintf("/v1/rides/%s/comments", ride.ID),
		map[string]string{}, commenter.AccessToken)

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected 422 for missing content, got %d", resp.StatusCode)
	}
}

func TestCommentRide_InvalidRideID(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "comment_badid")

	body := map[string]string{"content": "Hello"}
	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/rides/not-a-uuid/comments", body, user.AccessToken)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCommentRide_RideNotFound(t *testing.T) {
	env := testutil.Setup(t)
	testutil.TruncateAll(t, env.Pool)

	user := testutil.CreateUser(t, env.Pool, "comment_notfound")

	body := map[string]string{"content": "Hello"}
	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/rides/00000000-0000-0000-0000-000000000000/comments", body, user.AccessToken)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCommentRide_Unauthorized_MissingToken(t *testing.T) {
	env := testutil.Setup(t)

	body := map[string]string{"content": "Hello"}
	resp := testutil.Do(env.Router, http.MethodPost,
		"/v1/rides/00000000-0000-0000-0000-000000000000/comments", body, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}
