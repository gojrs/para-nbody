package engine

import (
	"fmt"

	"github.com/gojrs/para-nbody/types"
	"github.com/google/uuid"
)

// UniverseStore now saves and retrieves the master interface boundary contract
type UniverseStore interface {
	Create(id string, world types.Universe) error
	Get(id string) (types.Universe, bool, error)
	Update(id string, world types.Universe) error
	Delete(id string) error
	Close() error
}

type WorldManager struct {
	store UniverseStore
}

func NewWorldManager(store UniverseStore) *WorldManager {
	return &WorldManager{
		store: store,
	}
}

// CreateUniverse defaults to creating a V2World now, but outputs the universal interface token
func (m *WorldManager) CreateUniverse(w, h, d int) (string, error) {
	id := uuid.New().String()

	// Defaulting to your high-performance discrete engine matrix
	newWorld := NewV2World(id, w, h, d, 1, 0.0300, 0.0150, types.SeedingModeStandard)

	if err := m.store.Create(id, newWorld); err != nil {
		return "", fmt.Errorf("create universe %q: %w", id, err)
	}

	return id, nil
}

// GetUniverse now drops a clean types.Universe interface into the handlers floor!
func (m *WorldManager) GetUniverse(id string) (types.Universe, bool, error) {
	return m.store.Get(id)
}

// UpdateUniverse persists alterations to any interface-compliant world state
func (m *WorldManager) UpdateUniverse(id string, world types.Universe) error {
	if err := m.store.Update(id, world); err != nil {
		return fmt.Errorf("update universe %q: %w", id, err)
	}

	return nil
}

func (m *WorldManager) Stop() {
	if m == nil || m.store == nil {
		return
	}
	_ = m.store.Close()
}

// DeleteUniverse unlinks and permanently destroys a universe state by routing the task to the secure store boundary
func (m *WorldManager) DeleteUniverse(id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	return m.store.Delete(id)
}
