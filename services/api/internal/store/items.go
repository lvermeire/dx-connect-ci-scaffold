package store

import (
	"fmt"
	"sync"
)

// Item is the domain model for a named item.
type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ItemStore is a thread-safe in-memory store for Items.
type ItemStore struct {
	mu    sync.Mutex
	items []Item
	seq   int
}

// NewItemStore returns an initialised empty ItemStore.
func NewItemStore() *ItemStore {
	return &ItemStore{items: []Item{}}
}

// List returns a copy of all stored items.
func (s *ItemStore) List() []Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make([]Item, len(s.items))
	copy(result, s.items)
	return result
}

// Create adds a new item with the given name and returns it.
func (s *ItemStore) Create(name string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seq++
	item := Item{ID: fmt.Sprintf("%d", s.seq), Name: name}
	s.items = append(s.items, item)
	return item
}
