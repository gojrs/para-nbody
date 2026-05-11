package engine

import (
	"time"

	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
)

type WorldManager struct {
	cache *ttlcache.Cache[string, *World]
}

// NewWorldManager creates the manager and starts the TTL janitor
func NewWorldManager() *WorldManager {
	c := ttlcache.New[string, *World](
		ttlcache.WithTTL[string, *World](30 * time.Minute),
	)

	// Start the expiration janitor in the background
	go c.Start()

	return &WorldManager{
		cache: c,
	}
}

// CreateUniverse initializes a world, saves it, and returns the ID
func (m *WorldManager) CreateUniverse(w, h, d int) string {
	id := uuid.New().String()
	newWorld := NewWorld(w, h, d)

	m.cache.Set(id, &newWorld, ttlcache.DefaultTTL)
	return id
}

// GetUniverse retrieves a world by ID from the cache
func (m *WorldManager) GetUniverse(id string) (*World, bool) {
	item := m.cache.Get(id)
	if item == nil || item.Value() == nil {
		return nil, false
	}
	return item.Value(), true
}
