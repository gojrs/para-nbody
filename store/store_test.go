package storage

import (
	"testing"
	"time"

	"github.com/gojrs/para-nbody/engine"
)

func TestTTLStore_Lifecycle(t *testing.T) {
	store := NewTTLStore(50 * time.Millisecond)
	defer store.Close()

	id := "ttl-instance"
	world := &engine.V3World{ID: id, Width: 4, Height: 4, Depth: 4}

	if err := store.Create(id, world); err != nil {
		t.Fatalf("TTLStore Create error: %v", err)
	}

	got, exists, err := store.Get(id)
	if err != nil || !exists || got == nil {
		t.Fatal("Expected active memory footprint to be retrievable")
	}

	// Validate Delete explicitly
	if err := store.Delete(id); err != nil {
		t.Fatalf("TTLStore Delete error: %v", err)
	}

	_, exists, _ = store.Get(id)
	if exists {
		t.Error("Universe footprint was not cleared from mem table registry")
	}
}

func TestSQLiteStore_MockErrors(t *testing.T) {
	store, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize memory SQLite provider: %v", err)
	}
	defer store.Close()

	// Direct input edge parameters confirmation
	err = store.Create("", nil)
	if err == nil {
		t.Error("Expected error when missing core configuration arguments")
	}

	_, exists, err := store.Get("non-existent-uuid")
	if err != nil || exists {
		t.Error("Expected no rows query mismatch fallback safety trigger")
	}
}
