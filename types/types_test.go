package types

import (
	"math"
	"testing"
)

func TestNBodyConfigMode_String(t *testing.T) {
	tests := []struct {
		mode NBodyConfigMode
		want string
	}{
		{SeedingModeStandard, "INVARIANT_CORE"},
		{SeedingModeChaos, "QUANTUM_CHAOS"},
		{SeedingModeSplit, "PHASE_SPLIT"},
		{SeedingModeParity, "EVEN_PARITY"},
	}

	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("NBodyConfigMode.String() = %q, want %q", got, tt.want)
		}
	}
}

func TestMultivector_Operations(t *testing.T) {
	mv1 := Multivector{Scalar: 1.0, V: [5]float64{1, 2, 3, 10, 5}}
	mv2 := Multivector{Scalar: 2.0, V: [5]float64{0, 1, 2, 4, 2}}

	// Test Add
	added := mv1.Add(mv2)
	if added.Scalar != 3.0 || added.Matter() != 14.0 || added.Antimatter() != 7.0 {
		t.Errorf("Multivector.Add failed: got %+v", added)
	}

	if got := added.NetCurvature(); got != 7.0 {
		t.Errorf("NetCurvature() = %f, want 7.0", got)
	}

	// Test Scale
	scaled := mv1.Scale(2.0)
	if scaled.Scalar != 2.0 || scaled.V[3] != 20.0 {
		t.Errorf("Multivector.Scale failed: got %+v", scaled)
	}
}

func TestParticleCookbook_Decoders(t *testing.T) {
	cookbook := NewParticleCookbook()

	// Proton geometry: MinCompression >= 50.0, RequiredTwist >= 1.0
	protonMv := Multivector{V: [5]float64{0, 1.5, 0, 55.0, 0}}
	if label := cookbook.IdentifyVoxel(protonMv); label != KeyProton {
		t.Errorf("Expected Proton, got %q", label)
	}

	// Electron geometry: MinCompression >= 0.1, RequiredTwist <= -1.0
	electronMv := Multivector{V: [5]float64{0, -1.2, 0, 0.5, 0}}
	if label := cookbook.IdentifyVoxel(electronMv); label != KeyElectron {
		t.Errorf("Expected Electron, got %q", label)
	}

	// Vacuum
	vacuumMv := Multivector{}
	if label := cookbook.IdentifyVoxel(vacuumMv); label != KeyVacuum {
		t.Errorf("Expected Vacuum, got %q", label)
	}
}

func TestPixel_Metrics(t *testing.T) {
	p := NewPixel(30, 40, -10, 0, 0, 0)

	// Tension check: sqrt(30^2 + 40^2 + (-10)^2) = sqrt(900+1600+100) = sqrt(2600) = ~50
	expectedTension := int64(math.Sqrt(2600))
	if got := p.CalculateTension(); got != expectedTension {
		t.Errorf("CalculateTension() = %d, want %d", got, expectedTension)
	}

	// Chirality check: X > 0 && Tension < 0 -> ShiftClockwise
	if got := p.DetermineChirality(); got != ShiftClockwise {
		t.Errorf("DetermineChirality() = %v, want %v", got, ShiftClockwise)
	}
}
