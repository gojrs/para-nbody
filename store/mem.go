package storage

import (
	"fmt"
	"time"

	"github.com/gojrs/para-nbody/engine"
	"github.com/jellydator/ttlcache/v3"
)

type TTLStore struct {
	cache *ttlcache.Cache[string, *engine.World]
}

func NewTTLStore(ttl time.Duration) *TTLStore {
	cache := ttlcache.New[string, *engine.World](
		ttlcache.WithTTL[string, *engine.World](ttl),
	)

	go cache.Start()

	return &TTLStore{
		cache: cache,
	}
}

func (s *TTLStore) Create(id string, world *engine.World) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if world == nil {
		return fmt.Errorf("world is required")
	}

	s.cache.Set(id, world, ttlcache.DefaultTTL)
	return nil
}

func (s *TTLStore) Get(id string) (*engine.World, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("id is required")
	}

	item := s.cache.Get(id)
	if item == nil || item.Value() == nil {
		return nil, false, nil
	}

	return item.Value(), true, nil
}

func (s *TTLStore) Update(id string, world *engine.World) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if world == nil {
		return fmt.Errorf("world is required")
	}

	s.cache.Set(id, world, ttlcache.DefaultTTL)
	return nil
}

func (s *TTLStore) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}

	s.cache.Delete(id)
	return nil
}

func (s *TTLStore) Close() error {
	if s == nil || s.cache == nil {
		return nil
	}

	s.cache.Stop()
	return nil
}
