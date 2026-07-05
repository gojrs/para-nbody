package types

type LedgerState struct {
	Fields Multivector `json:"fields"`
	Energy float64     `json:"energy"`
	Commit bool        `json:"commit"`
}

type Multivector struct {
	Scalar float64    `json:"scalar"`
	V      [5]float64 `json:"v"` // [0]:Spin, [1]:Electric, [2]:Magnetic, [3]:Matter (+W), [4]:Antimatter (-W)
	// --- NEW WAVE GEOMETRY CHANNELS ---
	Amplitude float64 `json:"amplitude"` // Wave peak height
	Phase     float64 `json:"phase"`     // Current phase angle in radians (0 to 2*Pi)
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

// ---- GEOMETRIC COOKBOOK REGISTRY ----

// SpatialArchitecture represents the pure geometric requirements for a state
type SpatialArchitecture struct {
	MinCompression float64 // Depth into the +W side
	RequiredTwist  float64 // Electrical phase alignment on the V axis
}

// HumanLabel is our interface translation layer
type HumanLabel interface {
	HumanName() string
	CheckGeometry(mv Multivector) bool
}

// ProtonClass satisfies the HumanLabel interface via implicit composition
type ProtonClass struct {
	Architecture SpatialArchitecture
}

func NewProtonClass() ProtonClass {
	return ProtonClass{
		Architecture: SpatialArchitecture{
			MinCompression: 50.0,
			RequiredTwist:  1.0,
		},
	}
}

func (p ProtonClass) HumanName() string {
	return "Hydrogen Proton Core (H+)"
}

func (p ProtonClass) CheckGeometry(mv Multivector) bool {
	// Translates our human label to raw geometric validation
	return mv.V[3] >= p.Architecture.MinCompression && mv.V[1] >= p.Architecture.RequiredTwist
}

// ElectronClass satisfies the HumanLabel interface
type ElectronClass struct {
	Architecture SpatialArchitecture
}

func NewElectronClass() ElectronClass {
	return ElectronClass{
		Architecture: SpatialArchitecture{
			MinCompression: 0.1,  // Extremely light spatial compression depth
			RequiredTwist:  -1.0, // Negative orthogonal phase twist (opposite of Proton)
		},
	}
}

func (e ElectronClass) HumanName() string {
	return "Hydrogen Electron Shell (e-)"
}

func (e ElectronClass) CheckGeometry(mv Multivector) bool {
	// Validates if a coordinate voxel matches the geometric footprint of an electron
	// It must have a minimum baseline presence, but its Lattice Twist must be deeply negative
	return mv.V[3] >= e.Architecture.MinCompression && mv.V[1] <= e.Architecture.RequiredTwist
}

type ParticleCookbook struct {
	Decoders []HumanLabel
}

func NewParticleCookbook() *ParticleCookbook {
	return &ParticleCookbook{
		Decoders: []HumanLabel{
			NewProtonClass(),
			NewElectronClass(),
			UpQuarkClass{},   // 🌟 Registered from fundamental_recipes.go
			DownQuarkClass{}, // 🌟 Registered from fundamental_recipes.go
			NeutrinoClass{},  // 🌟 Registered from fundamental_recipes.go
		},
	}
}

// IdentifyVoxel checks the raw geometry against the cookbook recipes
func (cb *ParticleCookbook) IdentifyVoxel(mv Multivector) string {
	for _, decoder := range cb.Decoders {
		if decoder.CheckGeometry(mv) {
			return decoder.HumanName()
		}
	}
	return "Unperturbed Vacuum (Surface 0)"
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
	GridSize            int     `json:"grid_size"`             // Dynamic spatial resolution (e.g., 30, 40, 50)
	SunMass             float64 `json:"sun_mass"`              // Central anchor mass allocation
	MercuryMass         float64 `json:"mercury_mass"`          // Tracer mass allocation
	MercuryVelocityZ    float64 `json:"mercury_velocity_z"`    // Tangential speed kick
	MercuryIsMatter     bool    `json:"mercury_is_matter"`     // Structural field sign flag (True/False)
	GravitySensitivity  float64 `json:"gravity_sensitivity"`   // Weak force constant (The G-scaling lever)
	BaseMigrationRate   float64 `json:"base_migration_rate"`   // Fluid bleed rate (Normal vacuum diffusion)
	StickyClumpRate     float64 `json:"sticky_clump_rate"`     // Cohesive surface tension rate for dense knots
	PhaseRelaxationRate float64 `json:"phase_relaxation_rate"` // 🌟 Ensure this is here
	TwistFollowRate     float64 `json:"twist_follow_rate"`
}

type NBodyResult struct {
	FinalCount int64   `json:"final_count"`
	MaxMass    float64 `json:"max_mass"`
}
