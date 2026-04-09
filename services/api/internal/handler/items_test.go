package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/loic-vermeire/dx-connect-ci-scaffold/services/api/internal/handler"
	"github.com/loic-vermeire/dx-connect-ci-scaffold/services/api/internal/store"
)

func TestListItems_ReturnsEmptyArray(t *testing.T) {
	h := handler.New(store.NewItemStore())
	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	w := httptest.NewRecorder()

	h.ListItems(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []store.Item
	if err := json.NewDecoder(w.Body).Decode(&items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty list, got %d items", len(items))
	}
}

func TestCreateItem_Returns201WithItem(t *testing.T) {
	h := handler.New(store.NewItemStore())
	body := bytes.NewBufferString(`{"name":"widget"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/items", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var item store.Item
	if err := json.NewDecoder(w.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.Name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", item.Name)
	}
	if item.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateItem_MissingName_Returns400(t *testing.T) {
	h := handler.New(store.NewItemStore())
	body := bytes.NewBufferString(`{"name":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/items", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateItem_InvalidJSON_Returns400(t *testing.T) {
	h := handler.New(store.NewItemStore())
	body := bytes.NewBufferString(`not-json`)
	req := httptest.NewRequest(http.MethodPost, "/api/items", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateItem(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
