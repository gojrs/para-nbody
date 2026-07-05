package types

import "math"

// UpQuarkClass represents a fractional, high-compression twist configuration
type UpQuarkClass struct{}

func (u UpQuarkClass) HumanName() string { return "Up Quark (u)" }
func (u UpQuarkClass) CheckGeometry(mv Multivector) bool {
	// High compression depth, positive electrical lattice twist
	return mv.V[3] > 15.0 && mv.V[1] > 0.3
}

// DownQuarkClass represents the opposite fractional twist configuration
type DownQuarkClass struct{}

func (u DownQuarkClass) HumanName() string { return "Down Quark (d)" }
func (u DownQuarkClass) CheckGeometry(mv Multivector) bool {
	// High compression depth, light negative electrical lattice twist
	return mv.V[3] > 15.0 && mv.V[1] < -0.1 && mv.V[1] > -0.4
}

// NeutrinoClass represents the Off-Surface Neutrality State
type NeutrinoClass struct{}

func (n NeutrinoClass) HumanName() string { return "Electron Neutrino (v_e)" }
func (n NeutrinoClass) CheckGeometry(mv Multivector) bool {
	netCurvature := mv.V[3] - mv.V[4] // Net W displacement
	totalEnergy := mv.V[3] + mv.V[4]  // Total physical payload presence

	// It is neutral (Net Curvature is close to 0) AND it is off the surface (Total Energy exists)
	return math.Abs(netCurvature) < 0.01 && totalEnergy > 1.0 && math.Abs(mv.V[1]) < 0.05
}
