package types

// Universe is the master interface defining how the server handlers,
// automation agents, and database stores interact with a running reality.
type Universe interface {
	GetID() string
	GetDimensions() (int, int, int)
	Step()
	ToJSON() ([]byte, error)
	GenerateInventory(step int64) SpectrumReport
}

// V1WaveUniverse exposes the unique capabilities of your legacy float-field matrix.
type V1WaveUniverse interface {
	Universe
	HydratePillar(x, y, z int, recipe Multivector)
	CalculateElectronProtonDistances() []float64
}
