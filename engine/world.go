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
	// 1. Only subtract mass from cells occupied by actual tracers
	for _, tracer := range w.Tracers {
		tx := int(math.Round(tracer.Position[0]))
		ty := int(math.Round(tracer.Position[1]))
		tz := int(math.Round(tracer.Position[2]))

		if tx > 0 && tx < w.Width-1 && ty > 0 && ty < w.Height-1 && tz > 0 && tz < w.Depth-1 {
			if tracer.IsMatter {
				w.Cells[tx][ty][tz].Fields.V[3] -= tracer.BaseMass
				if w.Cells[tx][ty][tz].Fields.V[3] < 0 {
					w.Cells[tx][ty][tz].Fields.V[3] = 0
				}
			} else {
				w.Cells[tx][ty][tz].Fields.V[4] -= tracer.BaseMass
				if w.Cells[tx][ty][tz].Fields.V[4] < 0 {
					w.Cells[tx][ty][tz].Fields.V[4] = 0
				}
			}
		}
	}

	// 2. Tracer-to-Grid Deposition Phase
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

	// 3. Initialize transactional nextCells buffer
	nextCells := w.CloneCells()

	dx := [6]int{1, -1, 0, 0, 0, 0}
	dy := [6]int{0, 0, 1, -1, 0, 0}
	dz := [6]int{0, 0, 0, 0, 1, -1}

	// 4. Dispatch Concurrent Spatial Workers
	numWorkers := 4
	xRangePerWorker := (w.Width - 2) / numWorkers
	if xRangePerWorker < 1 {
		xRangePerWorker = w.Width - 2
		numWorkers = 1
	}

	var wg sync.WaitGroup

	const ConstructiveThreshold = 6.5
	const DestructiveThreshold = -2.0

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

						// --- PASS 1: WAVE INTERFERENCE & SPIN-LOCK TRAPPING ---
						cell := w.Cells[x][y][z]
						var netInterference float64 = 0.0

						for i := 0; i < 6; i++ {
							nx, ny, nz := x+dx[i], y+dy[i], z+dz[i]
							neigh := w.Cells[nx][ny][nz]

							phaseDiff := cell.Fields.Phase - neigh.Fields.Phase
							interference := cell.Fields.Amplitude * neigh.Fields.Amplitude * math.Cos(phaseDiff)
							netInterference += interference
						}

						if netInterference >= ConstructiveThreshold {
							nextCells[x][y][z].Fields.V[3] += netInterference * 0.15
							nextCells[x][y][z].Fields.V[0] += netInterference * 0.5
							nextCells[x][y][z].Fields.Amplitude *= 0.1
						} else if netInterference <= DestructiveThreshold {
							nextCells[x][y][z].Fields.V[3] += 0.5
							nextCells[x][y][z].Fields.Phase *= 0.5
						}

						// --- PASS 2: GEOMETRIC SURFACE DISTANCE IMPEDANCE & NEIGHBORHOOD ATTRACTION ---
						mVal := nextCells[x][y][z].Fields.V[3]
						aVal := nextCells[x][y][z].Fields.V[4]

						if mVal <= 0 && aVal <= 0 {
							continue
						}

						netCurvature := mVal - aVal
						distanceFromSurface := math.Abs(netCurvature)

						const SpatialRigidityConstant = 0.15
						dragFactor := 1.0 / (1.0 + (distanceFromSurface * SpatialRigidityConstant))

						// Local Neighborhood Gravity Pull (2-voxel envelope)
						var pullX, pullY, pullZ float64
						// NEW: Electrical Twist Acceleration Gradient vectors
						var elecX, elecY, elecZ float64

						for dxLocal := -3; dxLocal <= 3; dxLocal++ {
							for dyLocal := -3; dyLocal <= 3; dyLocal++ {
								for dzLocal := -3; dzLocal <= 3; dzLocal++ {
									nx, ny, nz := x+dxLocal, y+dyLocal, z+dzLocal

									if nx >= 0 && nx < w.Width && ny >= 0 && ny < w.Height && nz >= 0 && nz < w.Depth {
										targetCell := w.Cells[nx][ny][nz]
										distSq := float64(dxLocal*dxLocal + dyLocal*dyLocal + dzLocal*dzLocal)

										if distSq > 0 {
											// Mass gravity contribution (up to 2 voxels out)
											if math.Abs(float64(dxLocal)) <= 2 && math.Abs(float64(dyLocal)) <= 2 && math.Abs(float64(dzLocal)) <= 2 {
												if targetCell.Fields.V[3] > 0 {
													force := targetCell.Fields.V[3] / distSq
													pullX += (float64(dxLocal) / math.Sqrt(distSq)) * force
													pullY += (float64(dyLocal) / math.Sqrt(distSq)) * force
													pullZ += (float64(dzLocal) / math.Sqrt(distSq)) * force
												}
											}

											// LONG-RANGE ELECTRICAL TWIST contributed up to 3 voxels out
											neighborTwist := targetCell.Fields.V[1]
											if neighborTwist != 0 {
												// Coulomb imitation rule: Opposite charges attract, like charges repel
												// We multiply active cell twist by neighbor twist to establish orientation
												chargeInteraction := cell.Fields.V[1] * neighborTwist

												var twistForce float64
												if chargeInteraction < 0 {
													twistForce = math.Abs(chargeInteraction) / distSq // Attraction
												} else {
													twistForce = -(chargeInteraction) / distSq // Repulsion
												}

												elecX += (float64(dxLocal) / math.Sqrt(distSq)) * twistForce
												elecY += (float64(dyLocal) / math.Sqrt(distSq)) * twistForce
												elecZ += (float64(dzLocal) / math.Sqrt(distSq)) * twistForce
											}
										}
									}
								}
							}
						}

						var mWeights, aWeights [6]float64
						var mTotalWeight, aTotalWeight float64

						for i := 0; i < 6; i++ {
							nx, ny, nz := x+dx[i], y+dy[i], z+dz[i]
							neigh := w.Cells[nx][ny][nz]

							localPullFactor := 0.0
							if dx[i] != 0 {
								localPullFactor = pullX * float64(dx[i])
							}
							if dy[i] != 0 {
								localPullFactor = pullY * float64(dy[i])
							}
							if dz[i] != 0 {
								localPullFactor = pullZ * float64(dz[i])
							}

							localTwistFactor := 0.0
							if dx[i] != 0 {
								localTwistFactor = elecX * float64(dx[i])
							}
							if dy[i] != 0 {
								localTwistFactor = elecY * float64(dy[i])
							}
							if dz[i] != 0 {
								localTwistFactor = elecZ * float64(dz[i])
							}

							// Merge both mass attraction and electrical gradient changes into fluid weights
							mWeights[i] = 1.0 + (neigh.Fields.V[3] * 0.5) - (neigh.Fields.V[4] * 0.2 * w.RepulsionStrength) + (localPullFactor * 0.1) + (localTwistFactor * 0.15)
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

						nextCells[x][y][z].Fields.V[3] -= mOutTotal
						nextCells[x][y][z].Fields.V[4] -= aOutTotal

						for i := 0; i < 6; i++ {
							nx, ny, nz := x+dx[i], y+dy[i], z+dz[i]
							if mOutTotal > 0 && mTotalWeight > 0 {
								nextCells[nx][ny][nz].Fields.V[3] += mOutTotal * (mWeights[i] / mTotalWeight)
							}
							if aOutTotal > 0 && aTotalWeight > 0 {
								nextCells[nx][ny][nz].Fields.V[4] += aOutTotal * (aWeights[i] / aTotalWeight)
							}
						}
					}
				}
			}
		}(startX, endX)
	}

	wg.Wait()

	// 5. Boundary Sorting Phase
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

	// 6. Grid-to-Tracer Projection Phase
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

		gradX := w.Cells[tx+1][ty][tz].Energy - w.Cells[tx-1][ty][tz].Energy
		gradY := w.Cells[tx][ty+1][tz].Energy - w.Cells[tx][ty-1][tz].Energy
		gradZ := w.Cells[tx][ty][tz+1].Energy - w.Cells[tx][ty][tz-1].Energy

		chargeSign := 1.0
		if !tracer.IsMatter {
			chargeSign = -1.0 * w.RepulsionStrength
		}

		tracer.Velocity[0] += gradX * w.GravitySensitivity * chargeSign
		tracer.Velocity[1] += gradY * w.GravitySensitivity * chargeSign
		tracer.Velocity[2] += gradZ * w.GravitySensitivity * chargeSign

		tracer.Position[0] += tracer.Velocity[0]
		tracer.Position[1] += tracer.Velocity[1]
		tracer.Position[2] += tracer.Velocity[2]
	}
}

const (
	ConstructiveThreshold = 15.0 // Amplitude spike required to "bottle" a particle
	DestructiveThreshold  = -5.0 // Cancellation depth required to trigger "Reverse Spleef" injection
)

func (w *World) ProcessWaveMechanics() {
	dx := [6]int{1, -1, 0, 0, 0, 0}
	dy := [6]int{0, 0, 1, -1, 0, 0}
	dz := [6]int{0, 0, 0, 0, 1, -1}

	// We can process this within your parallelized worker segments safely
	for x := 1; x < w.Width-1; x++ {
		for y := 1; y < w.Height-1; y++ {
			for z := 1; z < w.Depth-1; z++ {
				cell := &w.Cells[x][y][z]

				var netInterference float64 = 0.0

				// Sample all 6 spatial directions to calculate interference patterns
				for i := 0; i < 6; i++ {
					nx, ny, nz := x+dx[i], y+dy[i], z+dz[i]
					neighbor := w.Cells[nx][ny][nz]

					// Standard wave interference equation: A1 * A2 * cos(phase_difference)
					phaseDiff := cell.Fields.Phase - neighbor.Fields.Phase
					interference := cell.Fields.Amplitude * neighbor.Fields.Amplitude * math.Cos(phaseDiff)
					netInterference += interference
				}

				// Execute your structural state updates based on your Wave Rules:
				if netInterference >= ConstructiveThreshold {
					// 1. Constructive Max: Waves stack and condense into a localized "bottled" unit!
					// We transfer loose grid energy into a locked mass channel
					cell.Fields.V[3] += netInterference * 0.1
					cell.Fields.Amplitude = 0 // The wave collapses/flattens into the mass node

				} else if netInterference <= DestructiveThreshold {
					// 2. Destructive Max: Waves cancel! Reverse Spleef mechanism activates
					// Space injects a fresh baseline field unit to restore ledger balance
					cell.Fields.V[3] += 1.0 // Injection unit padding
					cell.Fields.Phase = 0   // Reset phase structure locally
				}
			}
		}
	}
}
