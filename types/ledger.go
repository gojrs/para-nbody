package types

type LedgerState struct {
	Fields Multivector `json:"fields"`
	Energy float64     `json:"energy"`
	Commit bool        `json:"commit"`
}

type Multivector struct {
	Scalar float64    `json:"scalar"`
	V      [5]float64 `json:"v"` // [0]:Spin, [1]:Electric, [2]:Magnetic, [3]:Matter (+W), [4]:Antimatter (-W)
}

func (m Multivector) Add(other Multivector) Multivector {
	res := Multivector{Scalar: m.Scalar + other.Scalar}
	for i := 0; i < 5; i++ {
		res.V[i] = m.V[i] + other.V[i]
	}
	return res
}

func (m Multivector) Scale(factor float64) Multivector {
	res := Multivector{Scalar: m.Scalar * factor}
	for i := 0; i < 5; i++ {
		res.V[i] = m.V[i] * factor
	}
	return res
}

func (m Multivector) Matter() float64 {
	return m.V[3]
}

func (m Multivector) Antimatter() float64 {
	return m.V[4]
}

func (m Multivector) NetCurvature() float64 {
	return m.V[3] - m.V[4]
}

// TracerBody represents a continuous "bottled" unit traveling through the discrete grid.
type TracerBody struct {
	ID       string     `json:"id"`
	IsMatter bool       `json:"is_matter"` // True = Matter (+W tier), False = Antimatter (-W tier)
	Position [3]float64 `json:"position"`  // Continuous X, Y, Z coordinates
	Velocity [3]float64 `json:"velocity"`  // Smooth motion vector
	BaseMass float64    `json:"base_mass"` // Intrinsic mass payload
}

type NBodyConfig struct {
	N                           int     `json:"n"`
	BoxSize                     float64 `json:"box_size"`
	MaxSpeed                    float64 `json:"max_speed"`
	Steps                       int     `json:"steps"`
	ParticleMass                float64 `json:"particle_mass"`
	UnlikeMassRepulsionStrength float64 `json:"unlike_mass_repulsion_strength"`

	// --- NEW COSMIC AUTOMATION PARAMETERS ---
	GridSize           int     `json:"grid_size"`           // Dynamic spatial resolution (e.g., 30, 40, 50)
	SunMass            float64 `json:"sun_mass"`            // Central anchor mass allocation
	MercuryMass        float64 `json:"mercury_mass"`        // Tracer mass allocation
	MercuryVelocityZ   float64 `json:"mercury_velocity_z"`  // Tangential speed kick
	MercuryIsMatter    bool    `json:"mercury_is_matter"`   // Structural field sign flag (True/False)
	GravitySensitivity float64 `json:"gravity_sensitivity"` // Weak force constant (The G-scaling lever)
	BaseMigrationRate  float64 `json:"base_migration_rate"` // Fluid bleed rate (Normal vacuum diffusion)
	StickyClumpRate    float64 `json:"sticky_clump_rate"`   // Cohesive surface tension rate for dense knots
}

type NBodyResult struct {
	FinalCount int64   `json:"final_count"`
	MaxMass    float64 `json:"max_mass"`
}
