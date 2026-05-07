package types

import "encoding/json"

// Particle is the core entity of the simulation.
// Exported fields allow JSON serialization and cross-package access.
type Particle struct {
	ID     int64   `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	VX     float64 `json:"vx"`
	VY     float64 `json:"vy"`
	VZ     float64 `json:"vz"`
	Mass   float64 `json:"mass"`
	Charge float64 `json:"type"` // Maps to your "type" in frontend
}

// NBodyConfig defines the simulation parameters.
type NBodyConfig struct {
	N                           int     `json:"n"`
	BoxSize                     float64 `json:"boxSize"`
	MaxSpeed                    float64 `json:"maxSpeed"`
	ParticleMass                float64 `json:"particleMass"`
	Steps                       int     `json:"steps"`
	UnlikeMassRepulsionStrength float64 `json:"unlikeMassRepulsionStrength"`
	CellSize                    float64 `json:"cellSize"`
}

// SeedState allows resuming a universe from a specific point.
type SeedState struct {
	Particles []Particle `json:"particles"`
	BoxSize   float64    `json:"boxSize"`
}

// NBodyResult is the standard output for the API and Database.
type NBodyResult struct {
	Config            NBodyConfig `json:"config"`
	FinalCount        int         `json:"finalCount"`
	MatterCount       int         `json:"matterCount"`
	AntimatterCount   int         `json:"antimatterCount"`
	AnnihilationCount int         `json:"annihilationCount"`
	MergeCount        int         `json:"mergeCount"`
	PocketCount       int         `json:"pocketCount"`
	MaxMass           float64     `json:"maxMass"`
	StepsCompleted    int         `json:"stepsCompleted"`
	FinalParticles    []Particle  `json:"finalParticles"`
}

// runRequest handles the incoming Gin JSON payload.
type RunRequest struct {
	Experiment  string          `json:"experiment"`
	Config      json.RawMessage `json:"config"`
	Label       string          `json:"label"`
	ParentRunID *int64          `json:"parentRunId"`
}

// Helper method for logic.
func (p *Particle) IsMatter() bool {
	return p.Charge > 0
}
