package tests

import (
	"net/http"
	"testing"

	"github.com/nashirabbash/trackride/tests/testutil"
)

func TestHealth_ReturnsOK(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/health", nil, "")

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestHealth_ResponseShape(t *testing.T) {
	env := testutil.Setup(t)

	resp := testutil.Do(env.Router, http.MethodGet, "/health", nil, "")

	var body map[string]interface{}
	testutil.ParseJSON(t, resp, &body)

	if body["status"] != "healthy" {
		t.Errorf("expected status=healthy, got %v", body["status"])
	}
	if body["app"] == nil {
		t.Error("expected app field in response")
	}
}
