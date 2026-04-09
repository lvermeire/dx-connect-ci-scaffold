package handler

import (
	"encoding/json"
	"net/http"
)

type createRequest struct {
	Name string `json:"name"`
}

// ListItems responds with the full list of items as JSON.
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.items.List()) //nolint:errcheck
}

// CreateItem parses a JSON body and creates a new item.
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid JSON"}`)) //nolint:errcheck
		return
	}
	if req.Name == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"name is required"}`)) //nolint:errcheck
		return
	}
	item := h.items.Create(req.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item) //nolint:errcheck
}
