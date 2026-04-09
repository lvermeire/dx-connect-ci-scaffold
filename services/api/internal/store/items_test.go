package store_test

import (
	"testing"

	"github.com/lvermeire/dx-connect-ci-scaffold/services/api/internal/store"
)

func TestItemStore_EmptyOnCreate(t *testing.T) {
	s := store.NewItemStore()

	items := s.List()
	if len(items) != 0 {
		t.Fatalf("expected empty store, got %d items", len(items))
	}
}

func TestItemStore_CreateAndList(t *testing.T) {
	s := store.NewItemStore()

	item := s.Create("widget")

	if item.Name != "widget" {
		t.Errorf("expected name %q, got %q", "widget", item.Name)
	}
	if item.ID == "" {
		t.Error("expected non-empty ID")
	}

	items := s.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ID != item.ID {
		t.Errorf("expected item ID %q, got %q", item.ID, items[0].ID)
	}
}

func TestItemStore_ListReturnsCopy(t *testing.T) {
	s := store.NewItemStore()
	s.Create("widget")

	items := s.List()
	items[0].Name = "mutated"

	fresh := s.List()
	if fresh[0].Name == "mutated" {
		t.Error("List() should return a copy, not a reference to internal state")
	}
}
