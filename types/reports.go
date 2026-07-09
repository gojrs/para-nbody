package types

// PatternProfile tracks the lifetime, location, and internal metrics of a single coherent wave-knot.
type PatternProfile struct {
	ID         string  `json:"id"`
	FirstSeen  int64   `json:"first_seen"`
	LastSeen   int64   `json:"last_seen"`
	Population int64   `json:"population"`
	AvgRadius  float64 `json:"avg_radius"`
	AvgState   float64 `json:"avg_state"`
	MaxState   float64 `json:"max_state"`
	CenterX    float64 `json:"center_x"`
	CenterY    float64 `json:"center_y"`
	CenterZ    float64 `json:"center_z"`
}

// SpectrumReport represents the global matrix summary for a specific timeline step.
type SpectrumReport struct {
	NumSteps       int64            `json:"num_steps"`
	OccupiedCells  int64            `json:"occupied_cells"`
	AvgRadius      float64          `json:"avg_radius"`
	AvgState       float64          `json:"avg_state"`
	MaxState       float64          `json:"max_state"`
	PatternCount   int64            `json:"pattern_count"`
	LargestPattern int64            `json:"largest_pattern"`
	ActivePatterns []PatternProfile `json:"active_patterns"`
}
