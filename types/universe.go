package types

// Universe is the master interface defining how the server handlers,
// automation agents, and database stores interact with a running reality.
type Universe interface {
	GetID() string
	GetDimensions() (int, int, int)
	Step()                   // Increments the master clock cycle by 1 tick
	ToJSON() ([]byte, error) // Serializes the entire world space for disk storage
	GenerateInventory(step int64) SpectrumReport
}

// V1WaveUniverse exposes the unique capabilities of your legacy float-field matrix.
type V1WaveUniverse interface {
	Universe
	HydratePillar(x, y, z int, recipe Multivector)
	CalculateElectronProtonDistances() []float64
}

// V2PlanckUniverse exposes the clean, discrete value assignments of your new entity framework.
type V2PlanckUniverse interface {
	Universe
	GetPixel(x, y, z int) Pixel
	SetPixel(x, y, z int, p Pixel)
}
