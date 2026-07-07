package types

import "math"

// GeometricRecipe defines the "shopping list" of spatial dimensions
// required to cook or stabilize a specific sub-atomic particle architecture.
type GeometricRecipe struct {
	Name               string
	TargetCompression  float64 // Minimum +W or -W depth required
	RequiredTwistPhase float64 // The electrical orientation on the orthogonal axis
	RequiredSpin       float64 // Intrinsic angular matrix rotation
}

// ParticleClass defines the interface for how a bottled knot interacts with the world
type ParticleClass interface {
	GetName() string
	EvaluateStability(currentCurvature float64, localTwist float64) bool
}

// Proton Configuration
type Proton struct {
	BaseMass float64
	Charge   float64 // +1 Human Label
}

func (p Proton) GetName() string { return "Proton Core" }

func (p Proton) EvaluateStability(wDepth float64, vTwist float64) bool {
	// A Proton is stable only if it maintains a deep compression valley
	// AND a positive orthogonal twist phase
	return wDepth > 50.0 && vTwist > 0.5
}

// Electron Configuration
type Electron struct {
	BaseMass float64
	Charge   float64 // -1 Human Label
}

func (e Electron) GetName() string { return "Electron Shell Envelope" }
func (e Electron) EvaluateStability(wDepth float64, vTwist float64) bool {
	// An Electron has low compression depth but a highly agile negative twist phase
	return wDepth > 0.1 && vTwist < -0.5
}

// --- ENGINE V2: PHASE-SPACE PLANCK SCALE ARCHITECTURE ---

// IntegerState represents the 4-coordinate integer phase space layout
type IntegerState struct {
	Ax int64 `json:"ax"` // Field A, Axis X
	Ay int64 `json:"ay"` // Field A, Axis Y
	Bx int64 `json:"bx"` // Field B, Axis X
	By int64 `json:"by"` // Field B, Axis Y

	// 🏃 The 3D Momentum Vector (The wave-packet ripple)
	Momentum [3]int64 `json:"momentum"`
}

// CalculateTension derives the collective "drag" or scalar tension from the origin (0,0)
func (s IntegerState) CalculateTension() int64 {
	// Pythagorean hypotenuse across your 4 core dimensions
	sqSum := float64(s.Ax*s.Ax + s.Ay*s.Ay + s.Bx*s.Bx + s.By*s.By)
	return int64(math.Sqrt(sqSum))
}

// CalculateCharge extracts the net horizontal phase twist across the planes
func (s IntegerState) CalculateCharge() int64 {
	return s.Ay - s.By
}
