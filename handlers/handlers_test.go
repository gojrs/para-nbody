package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gojrs/para-nbody/engine"
	"github.com/gojrs/para-nbody/types"
)

// Re-using the MockUniverseStore declared in engine package tests context if combined,
// otherwise redeclare cleanly here for isolation.
type LocalMockStore struct {
	m map[string]types.Universe
}

func (l *LocalMockStore) Create(id string, world types.Universe) error { l.m[id] = world; return nil }
func (l *LocalMockStore) Get(id string) (types.Universe, bool, error) {
	w, ok := l.m[id]
	return w, ok, nil
}
func (l *LocalMockStore) Update(id string, world types.Universe) error { l.m[id] = world; return nil }
func (l *LocalMockStore) Delete(id string) error                       { delete(l.m, id); return nil }
func (l *LocalMockStore) Close() error                                 { return nil }

func TestHandler_PNBodyByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockStore := &LocalMockStore{m: make(map[string]types.Universe)}
	wm := engine.NewWorldManager(mockStore)
	h := NewHandler(wm)
	h.RegisterRoutes(r)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/pnbody/missing-id", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandler_WaveSweep_ValidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	mockStore := &LocalMockStore{m: make(map[string]types.Universe)}
	wm := engine.NewWorldManager(mockStore)
	h := NewHandler(wm)
	h.RegisterRoutes(r)

	cfg := types.NBodyConfig{
		GridSize:     8,
		Steps:        2,
		KernelRadius: 1,
		N:            types.SeedingModeChaos,
		Mode:         types.EngineModeClockwise,
	}

	body, _ := json.Marshal(cfg)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/pnbody/wave-sweep", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to decode response payload: %v", err)
	}

	if _, ok := resp["universe_id"]; !ok {
		t.Error("Response did not contain a newly registered universe_id structural token")
	}
}
