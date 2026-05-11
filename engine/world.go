package engine

import "github.com/gojrs/para-nbody/types"

// World represents the 3D GSIM ledger space.
//
// Cells are indexed as:
//
//	world.Cells[x][y][z]
//
// Each cell contains a LedgerState. The zero value of LedgerState is treated
// as Vacuum.
type World struct {
	Width  int                     `json:"width"`
	Height int                     `json:"height"`
	Depth  int                     `json:"depth"`
	Cells  [][][]types.LedgerState `json:"cells"`
}

// NewWorld initializes a World with the requested dimensions.
//
// Every cell is populated with the zero value of LedgerState, representing
// Vacuum.
func NewWorld(width, height, depth int) World {
	world := World{
		Width:  width,
		Height: height,
		Depth:  depth,
		Cells:  make([][][]types.LedgerState, width),
	}

	for x := 0; x < width; x++ {
		world.Cells[x] = make([][]types.LedgerState, height)

		for y := 0; y < height; y++ {
			world.Cells[x][y] = make([]types.LedgerState, depth)

			for z := 0; z < depth; z++ {
				world.Cells[x][y][z] = types.LedgerState{}
			}
		}
	}

	return world
}

// HydratePillar injects a "Matter Recipe" into a specific coordinate.
// In GSIM, this represents the 4-Simplex displacement (V3).
func (w *World) HydratePillar(x, y, z int, recipe types.Multivector) {
	// Boundary check to prevent a Spleef/Panic
	if x < 0 || x >= w.Width || y < 0 || y >= w.Height || z < 0 || z >= w.Depth {
		return
	}

	// Fetch current state
	current := w.Cells[x][y][z]

	// Add the recipe to the existing field
	// This represents 'Hydrating' the vacuum with mass
	current.Fields = current.Fields.Add(recipe)

	// Update the energy balance (The Referee will audit this later)
	current.Energy = current.Fields.Density()
	current.Commit = false // Mark as 'Dirty' so the Ref knows to recalculate

	// Write back to the grid
	w.Cells[x][y][z] = current
}
func (w *World) Step() {
	// 1. Create the buffer for the next state
	// We use the same dimensions as the current world
	nextCells := make([][][]types.LedgerState, w.Width)
	for i := range nextCells {
		nextCells[i] = make([][]types.LedgerState, w.Height)
		for j := range nextCells[i] {
			nextCells[i][j] = make([]types.LedgerState, w.Depth)
		}
	}

	// 2. Iterate through the "Inner Volumetric" space
	// We skip the 1-voxel thick "skin" of the world to stay safe
	for x := 1; x < w.Width-1; x++ {
		for y := 1; y < w.Height-1; y++ {
			for z := 1; z < w.Depth-1; z++ {
				current := w.Cells[x][y][z]

				// 3. Simple Diffusion: Average the 6 cardinal neighbors
				// (Up, Down, Left, Right, Forward, Back)
				var neighborSum types.Multivector
				neighborSum = neighborSum.Add(w.Cells[x+1][y][z].Fields)
				neighborSum = neighborSum.Add(w.Cells[x-1][y][z].Fields)
				neighborSum = neighborSum.Add(w.Cells[x][y+1][z].Fields)
				neighborSum = neighborSum.Add(w.Cells[x][y-1][z].Fields)
				neighborSum = neighborSum.Add(w.Cells[x][y][z+1].Fields)
				neighborSum = neighborSum.Add(w.Cells[x][y][z-1].Fields)

				// 4. Calculate the 'Local Gradient'
				// We average the neighbors (1/6) to find the 'Relaxed' state
				avgField := neighborSum.Scale(1.0 / 6.0)

				// Apply the update: 50% current state, 50% neighbor average
				// This prevents the universe from exploding/jittering
				current.Fields = current.Fields.Add(avgField).Scale(0.5)

				// Sync the Ledger columns
				current.Energy = current.Fields.Density()
				current.Commit = true

				nextCells[x][y][z] = current
			}
		}
	}

	// 5. Swap the buffer: The 'Next' becomes the 'Now'
	w.Cells = nextCells
}
