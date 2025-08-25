package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"pepo/internal/middleware"
)

func TestFormToJSONAdapterHandlesAPIPrefix(t *testing.T) {
	form := url.Values{}
	form.Set("person_id", "abc123")
	form.Set("description", "test action")
	form.Set("valence", "positive")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/actions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var captured *http.Request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r
	})

	middleware.NewFormToJSONAdapter(handler).ServeHTTP(httptest.NewRecorder(), req)

	if captured == nil {
		t.Fatalf("handler was not called")
	}

	if ct := captured.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}

	var payload map[string]any
	if err := json.NewDecoder(captured.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if payload["person_id"] != "abc123" {
		t.Errorf("unexpected person_id: %v", payload["person_id"])
	}
	if payload["description"] != "test action" {
		t.Errorf("unexpected description: %v", payload["description"])
	}
	if payload["valence"] != "positive" {
		t.Errorf("unexpected valence: %v", payload["valence"])
	}
	if _, ok := payload["occurred_at"]; !ok {
		t.Errorf("occurred_at missing from payload")
	}
}
