package engine

import (
	"math"
	"runtime"
	"sync"

	"github.com/gojrs/para-nbody/types"
)

// InitV2 provisions both high-density integer matrix layouts up front to allow double-buffering
func (w *World) InitV2() {
	w.CellsV2 = make([][][]types.IntegerState, w.Width)
	w.CellsV2Buffer = make([][][]types.IntegerState, w.Width)

	for x := 0; x < w.Width; x++ {
		w.CellsV2[x] = make([][]types.IntegerState, w.Height)
		w.CellsV2Buffer[x] = make([][]types.IntegerState, w.Height)
		for y := 0; y < w.Height; y++ {
			w.CellsV2[x][y] = make([]types.IntegerState, w.Depth)
			w.CellsV2Buffer[x][y] = make([]types.IntegerState, w.Depth)
		}
	}
}

// StepV2 executes the discrete Planck-scale mechanics with zero memory allocations
func (w *World) StepV2() {
	// 6 cardinal directional offsets [cite: 67, 68]
	dx := [6]int{1, -1, 0, 0, 0, 0}
	dy := [6]int{0, 0, 1, -1, 0, 0}
	dz := [6]int{0, 0, 0, 0, 1, -1}

	numCores := runtime.NumCPU()
	chunksPerAxis := int(math.Ceil(math.Cbrt(float64(numCores))))
	if chunksPerAxis < 1 {
		chunksPerAxis = 1
	}

	radius := w.KernelRadius
	if radius == 0 {
		radius = 3
	}

	chunkDimX := int(math.Ceil(float64(w.Width-(radius*2)) / float64(chunksPerAxis)))
	chunkDimY := int(math.Ceil(float64(w.Height-(radius*2)) / float64(chunksPerAxis)))
	chunkDimZ := int(math.Ceil(float64(w.Depth-(radius*2)) / float64(chunksPerAxis)))

	var wg sync.WaitGroup

	// 🧵 Concurrent 3D Partitioning Worker Grid Loop
	for cx := 0; cx < chunksPerAxis; cx++ {
		for cy := 0; cy < chunksPerAxis; cy++ {
			for cz := 0; cz < chunksPerAxis; cz++ {

				startX := radius + (cx * chunkDimX)
				endX := startX + chunkDimX
				if endX > w.Width-radius {
					endX = w.Width - radius
				}

				startY := radius + (cy * chunkDimY)
				endY := startY + chunkDimY
				if endY > w.Height-radius {
					endY = w.Height - radius
				}

				startZ := radius + (cz * chunkDimZ)
				endZ := startZ + chunkDimZ
				if endZ > w.Depth-radius {
					endZ = w.Depth - radius
				}

				if startX >= endX || startY >= endY || startZ >= endZ {
					continue
				}

				wg.Add(1)
				go func(sX, eX, sY, eY, sZ, eZ int) {
					defer wg.Done()

					for x := sX; x < eX; x++ {
						for y := sY; y < eY; y++ {
							for z := sZ; z < eZ; z++ {
								// Read from the persistent active CellsV2 layer
								cell := w.CellsV2[x][y][z]
								currentTension := cell.CalculateTension()

								var pullX, pullY, pullZ int64

								// 🎯 GATHER FIELD GRADIENT FORCE SLOPES
								for dxLocal := -radius; dxLocal <= radius; dxLocal++ {
									for dyLocal := -radius; dyLocal <= radius; dyLocal++ {
										for dzLocal := -radius; dzLocal <= radius; dzLocal++ {

											target := w.CellsV2[x+dxLocal][y+dyLocal][z+dzLocal]
											distSq := int64(dxLocal*dxLocal + dyLocal*dyLocal + dzLocal*dzLocal)

											if distSq > 0 {
												targetTension := target.CalculateTension()
												gradientForce := (targetTension - currentTension) / distSq

												pullX += int64(dxLocal) * gradientForce
												pullY += int64(dyLocal) * gradientForce
												pullZ += int64(dzLocal) * gradientForce
											}
										}
									}
								}

								// Base structural carry-over from current state directly into the pre-allocated write buffer
								w.CellsV2Buffer[x][y][z].Momentum[0] = cell.Momentum[0] + pullX
								w.CellsV2Buffer[x][y][z].Momentum[1] = cell.Momentum[1] + pullY
								w.CellsV2Buffer[x][y][z].Momentum[2] = cell.Momentum[2] + pullZ

								// 🍾 THE QUANTUM BOTTLE CONFIGURATION SHIFT
								const PlanckLimit = 12000
								if currentTension > PlanckLimit {
									w.CellsV2Buffer[x][y][z].Momentum[0] = (w.CellsV2Buffer[x][y][z].Momentum[0] + w.CellsV2Buffer[x][y][z].Momentum[1]) / 2
									w.CellsV2Buffer[x][y][z].Momentum[1] = -w.CellsV2Buffer[x][y][z].Momentum[0]
								}

								// 🎯 KINETIC ADVECTION: Push displacements downstream into w.CellsV2Buffer
								const ScalingDampener = 100
								moveX := w.CellsV2Buffer[x][y][z].Momentum[0] / ScalingDampener
								moveY := w.CellsV2Buffer[x][y][z].Momentum[1] / ScalingDampener
								moveZ := w.CellsV2Buffer[x][y][z].Momentum[2] / ScalingDampener

								for i := 0; i < 6; i++ {
									nx := x + dx[i]
									ny := y + dy[i]
									nz := z + dz[i]

									w.CellsV2Buffer[nx][ny][nz].Ax += moveX * int64(dx[i])
									w.CellsV2Buffer[nx][ny][nz].By += moveY * int64(dy[i])
									w.CellsV2Buffer[nx][ny][nz].Bx += moveZ * int64(dz[i])
									w.CellsV2Buffer[nx][ny][nz].By += (moveX * moveY) / (ScalingDampener * 10)
								}
							}
						}
					}
				}(startX, endX, startY, endY, startZ, endZ)
			}
		}
	}

	wg.Wait()

	// 🏓 THE PING-PONG SWAP: Flip the matrix pointers instantaneously with zero cost
	w.CellsV2, w.CellsV2Buffer = w.CellsV2Buffer, w.CellsV2
}
