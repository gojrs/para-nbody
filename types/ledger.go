package types

type LedgerState struct {
	// The "Fields" are the actual 5D physical values (The Clifford Multivector)
	// This tracks V1-V5: Spin, Electric, Magnetic, Matter, Expansion
	Fields Multivector `json:"fields"`

	// Energy represents the "Balance" of the voxel.
	// In a vacuum, this should net to zero across the sector.
	Energy float64 `json:"energy"`

	// Metadata for the Referee
	Commit bool `json:"commit"` // Has this state been validated for this tick?
}

type Multivector struct {
	Scalar float64    `json:"scalar"`
	V      [5]float64 `json:"v"` // The 5 Fields
}

// Add is a method of Multivector - must be in the same package!
func (m Multivector) Add(other Multivector) Multivector {
	res := Multivector{Scalar: m.Scalar + other.Scalar}
	for i := 0; i < 5; i++ {
		res.V[i] = m.V[i] + other.V[i]
	}
	return res
}

// Scale returns a new Multivector with each component multiplied by factor.
func (m Multivector) Scale(factor float64) Multivector {
	res := Multivector{Scalar: m.Scalar * factor}
	for i := 0; i < 5; i++ {
		res.V[i] = m.V[i] * factor
	}
	return res
}

// Density returns the Matter Pillar value (V3)
func (m Multivector) Density() float64 {
	return m.V[3]
}
