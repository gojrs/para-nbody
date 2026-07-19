package engine

import (
	"testing"

	"github.com/gojrs/para-nbody/types"
)

type MockUniverseStore struct {
	store map[string]types.Universe
}

func NewMockUniverseStore() *MockUniverseStore {
	return &MockUniverseStore{store: make(map[string]types.Universe)}
}

func (m *MockUniverseStore) Create(id string, world types.Universe) error {
	m.store[id] = world
	return nil
}

func (m *MockUniverseStore) Get(id string) (types.Universe, bool, error) {
	w, ok := m.store[id]
	return w, ok, nil
}

func (m *MockUniverseStore) Update(id string, world types.Universe) error {
	m.store[id] = world
	return nil
}

func (m *MockUniverseStore) Delete(id string) error {
	delete(m.store, id)
	return nil
}

func (m *MockUniverseStore) Close() error { return nil }

func TestV3World_StepAndInventory(t *testing.T) {
	size := 8
	world := &V3World{
		ID:           "test-fsm",
		Width:        size,
		Height:       size,
		Depth:        size,
		KernelRadius: 1,
		Cells:        make([][][]types.Pixel, size),
		Buffer:       make([][][]types.Pixel, size),
	}

	for x := 0; x < size; x++ {
		world.Cells[x] = make([][]types.Pixel, size)
		world.Buffer[x] = make([][]types.Pixel, size)
		for y := 0; y < size; y++ {
			world.Cells[x][y] = make([]types.Pixel, size)
			world.Buffer[x][y] = make([]types.Pixel, size)
			for z := 0; z < size; z++ {
				world.Cells[x][y][z] = types.NewPixel(10, 0, -5, 0, 0, 0)
			}
		}
	}

	// Run step verification iteration
	world.Step()

	report := world.GenerateInventory(1)
	if report.NumSteps != 1 {
		t.Errorf("Expected report steps to be 1, got %d", report.NumSteps)
	}
}

func TestWorldManager_Lifecycle(t *testing.T) {
	mockStore := NewMockUniverseStore()
	wm := NewWorldManager(mockStore)
	defer wm.Stop()

	id, err := wm.CreateUniverse(10, 10, 10)
	if err != nil {
		t.Fatalf("Failed to create universe: %v", err)
	}

	if id == "" {
		t.Fatal("Expected non-empty universe UUID string token")
	}

	world, exists, err := wm.GetUniverse(id)
	if err != nil || !exists || world == nil {
		t.Errorf("Failed to retrieve managed universe framework")
	}

	err = wm.DeleteUniverse(id)
	if err != nil {
		t.Errorf("Failed deleting tracking entry: %v", err)
	}
}
