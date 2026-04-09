package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lvermeire/dx-connect-ci-scaffold/services/api/internal/handler"
	"github.com/lvermeire/dx-connect-ci-scaffold/services/api/internal/store"
)

func newTestHandler() *handler.Handler {
	return handler.New(store.NewItemStore())
}

func TestHealth_Returns200WithStatusOK(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	h.Health(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status %q, got %q", "ok", body["status"])
	}
}
