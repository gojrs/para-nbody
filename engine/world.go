package engine

import (
	"math"
	"sync"

	"github.com/gojrs/para-nbody/types"
)

type World struct {
	Width             int                     `json:"width"`
	Height            int                     `json:"height"`
	Depth             int                     `json:"depth"`
	RepulsionStrength float64                 `json:"repulsion_strength"`
	Cells             [][][]types.LedgerState `json:"cells"`
	Tracers           []types.TracerBody      `json:"tracers"`

	// --- AUTOMATION RUNTIME FIELDS ---
	GravitySensitivity float64 `json:"gravity_sensitivity"`
	BaseMigrationRate  float64 `json:"base_migration_rate"`
	StickyClumpRate    float64 `json:"sticky_clump_rate"`
}

func NewWorld(width, height, depth int) World {
	world := World{
		Width:             width,
		Height:            height,
		Depth:             depth,
		RepulsionStrength: 1.0,
		Cells:             make([][][]types.LedgerState, width),
		Tracers:           make([]types.TracerBody, 0),
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

func (w *World) HydratePillar(x, y, z int, recipe types.Multivector) {
	if x < 0 || x >= w.Width || y < 0 || y >= w.Height || z < 0 || z >= w.Depth {
		return
	}
	current := w.Cells[x][y][z]
	current.Fields = current.Fields.Add(recipe)
	current.Energy = current.Fields.NetCurvature()
	current.Commit = false
	w.Cells[x][y][z] = current
}

func (w *World) CloneCells() [][][]types.LedgerState {
	cloned := make([][][]types.LedgerState, w.Width)
	for x := 0; x < w.Width; x++ {
		cloned[x] = make([][]types.LedgerState, w.Height)
		for y := 0; y < w.Height; y++ {
			cloned[x][y] = make([]types.LedgerState, w.Depth)
			copy(cloned[x][y], w.Cells[x][y])
		}
	}
	return cloned
}

func (w *World) Step() {
	// NEW FIX: Clear out previous tick's raw tracer mass injections to prevent stack explosions
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			for z := 0; z < w.Depth; z++ {
				w.Cells[x][y][z].Fields.V[3] = 0 // Reset temporary matter field line
				w.Cells[x][y][z].Fields.V[4] = 0 // Reset temporary antimatter field line
			}
		}
	}
	// 1. Tracer-to-Grid Deposition Phase
	// Bottled units inject their mass onto the field grid before background migrations compute
	for _, tracer := range w.Tracers {
		tx := int(math.Round(tracer.Position[0]))
		ty := int(math.Round(tracer.Position[1]))
		tz := int(math.Round(tracer.Position[2]))

		if tx > 0 && tx < w.Width-1 && ty > 0 && ty < w.Height-1 && tz > 0 && tz < w.Depth-1 {
			if tracer.IsMatter {
				w.Cells[tx][ty][tz].Fields.V[3] += tracer.BaseMass
			} else {
				w.Cells[tx][ty][tz].Fields.V[4] += tracer.BaseMass
			}
		}
	}

	// 2. Initialize nextCells fresh to zero values
	nextCells := make([][][]types.LedgerState, w.Width)
	for i := range nextCells {
		nextCells[i] = make([][]types.LedgerState, w.Height)
		for j := range nextCells[i] {
			nextCells[i][j] = make([]types.LedgerState, w.Depth)
			for k := range nextCells[i][j] {
				nextCells[i][j][k].Fields.Scalar = w.Cells[i][j][k].Fields.Scalar
				for f := 0; f < 3; f++ {
					nextCells[i][j][k].Fields.V[f] = w.Cells[i][j][k].Fields.V[f]
				}
			}
		}
	}

	dx := [6]int{1, -1, 0, 0, 0, 0}
	dy := [6]int{0, 0, 1, -1, 0, 0}
	dz := [6]int{0, 0, 0, 0, 1, -1}

	// 3. Dispatch Spatial Worker Pool
	numWorkers := 4
	xRangePerWorker := (w.Width - 2) / numWorkers
	if xRangePerWorker < 1 {
		xRangePerWorker = w.Width - 2
		numWorkers = 1
	}

	var wg sync.WaitGroup

	for workerID := 0; workerID < numWorkers; workerID++ {
		startX := 1 + (workerID * xRangePerWorker)
		endX := startX + xRangePerWorker
		if workerID == numWorkers-1 {
			endX = w.Width - 1
		}

		wg.Add(1)
		go func(sX, eX int) {
			defer wg.Done()

			for x := sX; x < eX; x++ {
				for y := 1; y < w.Height-1; y++ {
					for z := 1; z < w.Depth-1; z++ {
						curr := w.Cells[x][y][z]
						mVal := curr.Fields.V[3]
						aVal := curr.Fields.V[4]

						if mVal <= 0 && aVal <= 0 {
							continue
						}

						// Depth-Based Grid Tracking (Sinking Mass-Drag)
						// Highly concentrated points sink deep, dramatically reducing field bleed
						dragFactor := 1.0 / (1.0 + 0.15*(mVal+aVal))

						var mWeights [6]float64
						var aWeights [6]float64
						var mTotalWeight float64 = 0
						var aTotalWeight float64 = 0

						for i := 0; i < 6; i++ {
							nx, ny, nz := x+dx[i], y+dy[i], z+dz[i]
							neigh := w.Cells[nx][ny][nz]

							mWeights[i] = 1.0 + (neigh.Fields.V[3] * 0.5) - (neigh.Fields.V[4] * 0.2 * w.RepulsionStrength)
							if mWeights[i] < 0.001 {
								mWeights[i] = 0.001
							}
							mTotalWeight += mWeights[i]

							aWeights[i] = 1.0 + (neigh.Fields.V[4] * 0.5) - (neigh.Fields.V[3] * 0.2 * w.RepulsionStrength)
							if aWeights[i] < 0.001 {
								aWeights[i] = 0.001
							}
							aTotalWeight += aWeights[i]
						}

						mMigrationRate := w.BaseMigrationRate
						if mVal > 5.0 {
							mMigrationRate = w.StickyClumpRate
						}

						aMigrationRate := w.BaseMigrationRate
						if aVal > 5.0 {
							aMigrationRate = w.StickyClumpRate
						}

						mOutTotal := mVal * mMigrationRate * dragFactor
						aOutTotal := aVal * aMigrationRate * dragFactor

						nextCells[x][y][z].Fields.V[3] += mVal - mOutTotal
						nextCells[x][y][z].Fields.V[4] += aVal - aOutTotal

						for i := 0; i < 6; i++ {
							nx, ny, nz := x+dx[i], y+dy[i], z+dz[i]

							if mOutTotal > 0 && mTotalWeight > 0 {
								mMove := mOutTotal * (mWeights[i] / mTotalWeight)
								nextCells[nx][ny][nz].Fields.V[3] += mMove
							}

							if aOutTotal > 0 && aTotalWeight > 0 {
								aMove := aOutTotal * (aWeights[i] / aTotalWeight)
								nextCells[nx][ny][nz].Fields.V[4] += aMove
							}
						}
					}
				}
			}
		}(startX, endX)
	}

	wg.Wait()

	// 4. Boundary Sorting Phase
	for x := 1; x < w.Width-1; x++ {
		for y := 1; y < w.Height-1; y++ {
			for z := 1; z < w.Depth-1; z++ {
				cell := &nextCells[x][y][z]

				m := cell.Fields.V[3]
				a := cell.Fields.V[4]

				if m > 0 && a > 0 {
					annihilationAmt := math.Min(m, a) * 0.5
					cell.Fields.V[3] -= annihilationAmt
					cell.Fields.V[4] -= annihilationAmt
				}

				if cell.Fields.V[3] < 1e-4 {
					cell.Fields.V[3] = 0
				}
				if cell.Fields.V[4] < 1e-4 {
					cell.Fields.V[4] = 0
				}

				cell.Energy = cell.Fields.NetCurvature()
				cell.Commit = true
			}
		}
	}

	w.Cells = nextCells

	// 5. Grid-to-Tracer Projection Phase (With Dynamic Asymmetric Scaling)
	for idx := range w.Tracers {
		tracer := &w.Tracers[idx]
		tx := int(math.Round(tracer.Position[0]))
		ty := int(math.Round(tracer.Position[1]))
		tz := int(math.Round(tracer.Position[2]))

		bounceDampening := -0.9
		if tx <= 1 || tx >= w.Width-2 {
			tracer.Velocity[0] *= bounceDampening
			tracer.Position[0] += tracer.Velocity[0]
		}
		if ty <= 1 || ty >= w.Height-2 {
			tracer.Velocity[1] *= bounceDampening
			tracer.Position[1] += tracer.Velocity[1]
		}
		if tz <= 1 || tz >= w.Depth-2 {
			tracer.Velocity[2] *= bounceDampening
			tracer.Position[2] += tracer.Velocity[2]
		}

		tx = int(math.Round(tracer.Position[0]))
		ty = int(math.Round(tracer.Position[1]))
		tz = int(math.Round(tracer.Position[2]))

		if tx <= 0 || tx >= w.Width-1 || ty <= 0 || ty >= w.Height-1 || tz <= 0 || tz >= w.Depth-1 {
			continue
		}

		// Read base ledger gradients
		gradX := w.Cells[tx+1][ty][tz].Energy - w.Cells[tx-1][ty][tz].Energy
		gradY := w.Cells[tx][ty+1][tz].Energy - w.Cells[tx][ty-1][tz].Energy
		gradZ := w.Cells[tx][ty][tz+1].Energy - w.Cells[tx][ty][tz-1].Energy

		// Apply the asymmetrical charge conjugation scaling factors
		chargeSign := 1.0

		if !tracer.IsMatter {
			// If an antimatter body is interacting with a matter-dominated grid,
			// scale the repulsive acceleration by your custom configuration factor!
			chargeSign = -1.0 * w.RepulsionStrength
		}

		//gravitySensitivity := 0.02
		tracer.Velocity[0] += gradX * w.GravitySensitivity * chargeSign
		tracer.Velocity[1] += gradY * w.GravitySensitivity * chargeSign
		tracer.Velocity[2] += gradZ * w.GravitySensitivity * chargeSign

		tracer.Position[0] += tracer.Velocity[0]
		tracer.Position[1] += tracer.Velocity[1]
		tracer.Position[2] += tracer.Velocity[2]
	}
}
