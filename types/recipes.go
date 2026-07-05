package types

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
