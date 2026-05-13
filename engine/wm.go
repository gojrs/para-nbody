package engine

import (
	"fmt"

	"github.com/google/uuid"
)

type UniverseStore interface {
	Create(id string, world *World) error
	Get(id string) (*World, bool, error)
	Update(id string, world *World) error
	Delete(id string) error
	Close() error
}

type WorldManager struct {
	store UniverseStore
}

// NewWorldManager creates the manager with a pluggable universe store.
func NewWorldManager(store UniverseStore) *WorldManager {
	return &WorldManager{
		store: store,
	}
}

// CreateUniverse initializes a world, saves it, and returns the ID.
func (m *WorldManager) CreateUniverse(w, h, d int) (string, error) {
	id := uuid.New().String()
	newWorld := NewWorld(w, h, d)

	if err := m.store.Create(id, &newWorld); err != nil {
		return "", fmt.Errorf("create universe %q: %w", id, err)
	}

	return id, nil
}

// GetUniverse retrieves a world by ID from the store.
func (m *WorldManager) GetUniverse(id string) (*World, bool, error) {
	return m.store.Get(id)
}

// UpdateUniverse persists changes to an existing universe.
func (m *WorldManager) UpdateUniverse(id string, world *World) error {
	if err := m.store.Update(id, world); err != nil {
		return fmt.Errorf("update universe %q: %w", id, err)
	}

	return nil
}

// Stop shuts down the WorldManager store.
func (m *WorldManager) Stop() {
	if m == nil || m.store == nil {
		return
	}

	_ = m.store.Close()
}
