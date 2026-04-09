package handler

import (
	"encoding/json"
	"net/http"

	"github.com/lvermeire/dx-connect-ci-scaffold/services/api/internal/store"
)

// Handler holds dependencies for all HTTP handlers.
type Handler struct {
	items *store.ItemStore
}

// New returns a Handler wired to the given ItemStore.
func New(items *store.ItemStore) *Handler {
	return &Handler{items: items}
}

// Health responds with {"status":"ok"}.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) //nolint:errcheck
}
