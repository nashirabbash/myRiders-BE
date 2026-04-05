package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// Do sends an HTTP request to the given Gin router and returns the response.
// body may be nil for requests without a payload.
func Do(router *gin.Engine, method, path string, body interface{}, authToken string) *http.Response {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w.Result()
}

// ParseJSON decodes the JSON body of an HTTP response into target.
func ParseJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("ParseJSON: %v", err)
	}
}

// BodyString reads the entire response body as a string for debugging.
func BodyString(resp *http.Response) string {
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

// parseUUIDForQuery parses a UUID string into pgtype.UUID.
func parseUUIDForQuery(s string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(s)
	return id, err
}

// parseUUIDBytes fills a [16]byte from a UUID string (internal helper).
func parseUUIDBytes(s string, dst *[16]byte) error {
	var id pgtype.UUID
	if err := id.Scan(s); err != nil {
		return err
	}
	*dst = id.Bytes
	return nil
}
