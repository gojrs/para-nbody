package storage

import (
	"fmt"
	"time"

	"github.com/gojrs/para-nbody/types"
	"github.com/jellydator/ttlcache/v3"
)

type TTLStore struct {
	cache *ttlcache.Cache[string, types.Universe]
}

func NewTTLStore(ttl time.Duration) *TTLStore {
	cache := ttlcache.New[string, types.Universe](
		ttlcache.WithTTL[string, types.Universe](ttl),
	)

	go cache.Start()

	return &TTLStore{
		cache: cache,
	}
}

func (s *TTLStore) Create(id string, world types.Universe) error {
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if world == nil {
		return fmt.Errorf("world is required")
	}

	s.cache.Set(id, world, ttlcache.DefaultTTL)
	return nil
}

func (s *TTLStore) Get(id string) (types.Universe, bool, error) {
	if id == "" {
		return nil, false, fmt.Errorf("id is required")
	}

	item := s.cache.Get(id)
	if item == nil || item.Value() == nil {
		return nil, false, nil
	}

	return item.Value(), true, nil
}

func (s *TTLStore) Update(id string, world types.Universe) error {
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
