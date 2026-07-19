package engine

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"sync"

	"github.com/gojrs/para-nbody/types"
)

type V3World struct {
	ID           string            `json:"id"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	Depth        int               `json:"depth"`
	KernelRadius int               `json:"kernel_radius"`
	Cells        [][][]types.Pixel `json:"cells"`
	Buffer       [][][]types.Pixel `json:"buffer"`
}

// Step runs the discrete Alice and Bob mechanics with zero runtime heap allocations.
func (w *V3World) Step() {
	dx := [6]int{1, -1, 0, 0, 0, 0}
	dy := [6]int{0, 0, 1, -1, 0, 0}
	dz := [6]int{0, 0, 0, 0, 1, -1}

	radius := w.KernelRadius
	if radius == 0 {
		radius = 1
	}

	chunksPerAxis := int(math.Ceil(math.Cbrt(float64(runtime.NumCPU()))))
	chunkDimX := int(math.Ceil(float64(w.Width-(radius*2)) / float64(chunksPerAxis)))
	chunkDimY := int(math.Ceil(float64(w.Height-(radius*2)) / float64(chunksPerAxis)))
	chunkDimZ := int(math.Ceil(float64(w.Depth-(radius*2)) / float64(chunksPerAxis)))

	var wg sync.WaitGroup

	// Reset write buffers to isolate the frame calculation
	for x := radius; x < w.Width-radius; x++ {
		for y := radius; y < w.Height-radius; y++ {
			for z := radius; z < w.Depth-radius; z++ {
				w.Buffer[x][y][z] = types.NewPixel(0, 0, 0, 0, 0, 0)
			}
		}
	}

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
					var a types.Alice
					var b types.Bob

					for x := sX; x < eX; x++ {
						for y := sY; y < eY; y++ {
							for z := sZ; z < eZ; z++ {
								room := w.Cells[x][y][z]
								currentTension := room.CalculateTension()

								var expX, expY, expZ int64
								var conX, conY, conZ int64

								for dxL := -radius; dxL <= radius; dxL++ {
									for dyL := -radius; dyL <= radius; dyL++ {
										for dzL := -radius; dzL <= radius; dzL++ {
											target := w.Cells[x+dxL][y+dyL][z+dzL]
											distSq := int64(dxL*dxL + dyL*dyL + dzL*dzL)
											if distSq == 0 {
												continue
											}

											gradientForce := (target.CalculateTension() - currentTension) / distSq

											ax, ay, az := a.Expand(int64(dxL), int64(dyL), int64(dzL), gradientForce)
											expX += ax
											expY += ay
											expZ += az

											bx, by, bz := b.Contract(int64(dxL), int64(dyL), int64(dzL), gradientForce)
											conX += bx
											conY += by
											conZ += bz
										}
									}
								}

								w.Buffer[x][y][z].Destination.X = room.Destination.X + uint64(expX+conX)
								w.Buffer[x][y][z].Destination.Y = room.Destination.Y + uint64(expY+conY)
								w.Buffer[x][y][z].Destination.Z = room.Destination.Z + uint64(expZ+conZ)

								if room.Alice != nil {
									w.Buffer[x][y][z].Alice.X = room.Alice.X
									w.Buffer[x][y][z].Alice.Y = room.Alice.Y
									w.Buffer[x][y][z].Mal.Tension = room.Mal.Wiggle(currentTension)
								}

								chirality := room.DetermineChirality()
								const ScalingDampener = 100
								moveToken := int64(w.Buffer[x][y][z].Destination.X) / ScalingDampener

								if moveToken > 0 {
									for i := 0; i < 6; i++ {
										nx, ny, nz := x+dx[i], y+dy[i], z+dz[i]
										if w.Buffer[nx][ny][nz].Alice == nil {
											continue
										}

										switch chirality {
										case types.ShiftClockwise:
											w.Buffer[nx][ny][nz].Alice.Y += moveToken * int64(dx[i])
											w.Buffer[nx][ny][nz].Alice.X += moveToken * int64(dy[i])
										case types.ShiftCounterClockwise:
											w.Buffer[nx][ny][nz].Alice.Y -= moveToken * int64(dx[i])
											w.Buffer[nx][ny][nz].Alice.X -= moveToken * int64(dy[i])
										case types.ShiftStatic:
											w.Buffer[nx][ny][nz].Mal.Tension += moveToken * int64(dz[i])
										}
									}
								}
							}
						}
					}
				}(startX, endX, startY, endY, startZ, endZ)
			}
		}
	}
	wg.Wait()
	w.Cells, w.Buffer = w.Buffer, w.Cells
}

func (w *V3World) GetID() string                  { return w.ID }
func (w *V3World) GetDimensions() (int, int, int) { return w.Width, w.Height, w.Depth }
func (w *V3World) ToJSON() ([]byte, error)        { return json.Marshal(w) }

func (w *V3World) GenerateInventory(currentStep int64) types.SpectrumReport {
	visited := make([][][]bool, w.Width)
	for i := range visited {
		visited[i] = make([][]bool, w.Height)
		for j := range visited[i] {
			visited[i][j] = make([]bool, w.Depth)
		}
	}

	var report types.SpectrumReport
	report.NumSteps = currentStep

	type Coord struct{ X, Y, Z int }
	var activePatterns []types.PatternProfile
	const MaterialDensityThreshold = 5.0

	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			for z := 0; z < w.Depth; z++ {
				p := w.Cells[x][y][z]
				cellMagnitude := float64(p.CalculateTension())

				if math.IsNaN(cellMagnitude) || math.IsInf(cellMagnitude, 0) || cellMagnitude < MaterialDensityThreshold {
					continue
				}

				report.OccupiedCells++
				report.AvgState += cellMagnitude
				if cellMagnitude > report.MaxState {
					report.MaxState = cellMagnitude
				}

				if !visited[x][y][z] {
					report.PatternCount++
					queue := []Coord{{X: x, Y: y, Z: z}}
					visited[x][y][z] = true

					var clusterVoxels []Coord
					var cSumX, cSumY, cSumZ float64
					var cMaxState float64
					var cSumState float64

					for len(queue) > 0 {
						curr := queue[0]
						queue = queue[1:]
						clusterVoxels = append(clusterVoxels, curr)

						room := w.Cells[curr.X][curr.Y][curr.Z]
						mag := float64(room.CalculateTension())

						if !math.IsNaN(mag) && !math.IsInf(mag, 0) {
							cSumState += mag
							if mag > cMaxState {
								cMaxState = mag
							}
						}

						cSumX += float64(curr.X)
						cSumY += float64(curr.Y)
						cSumZ += float64(curr.Z)

						neighbors := []Coord{
							{X: (curr.X + 1) % w.Width, Y: curr.Y, Z: curr.Z},
							{X: (curr.X - 1 + w.Width) % w.Width, Y: curr.Y, Z: curr.Z},
							{X: curr.X, Y: (curr.Y + 1) % w.Height, Z: curr.Z},
							{X: curr.X, Y: (curr.Y - 1 + w.Height) % w.Height, Z: curr.Z},
							{X: curr.X, Y: curr.Y, Z: (curr.Z + 1) % w.Depth},
							{X: curr.X, Y: curr.Y, Z: (curr.Z - 1 + w.Depth) % w.Depth},
						}

						for _, n := range neighbors {
							if !visited[n.X][n.Y][n.Z] {
								np := w.Cells[n.X][n.Y][n.Z]
								nMag := float64(np.CalculateTension())
								if !math.IsNaN(nMag) && nMag >= MaterialDensityThreshold {
									visited[n.X][n.Y][n.Z] = true
									queue = append(queue, n)
								}
							}
						}
					}

					pop := int64(len(clusterVoxels))
					if pop > report.LargestPattern {
						report.LargestPattern = pop
					}

					var cX, cY, cZ float64
					if pop > 0 {
						cX = cSumX / float64(pop)
						cY = cSumY / float64(pop)
						cZ = cSumZ / float64(pop)
					}

					var clusterRadiusSum float64
					for _, v := range clusterVoxels {
						dx := float64(v.X) - cX
						dy := float64(v.Y) - cY
						dz := float64(v.Z) - cZ
						distSq := dx*dx + dy*dy + dz*dz
						if distSq > 0 {
							clusterRadiusSum += math.Sqrt(distSq)
						}
					}

					avgRadius := 0.0
					if pop > 0 {
						avgRadius = clusterRadiusSum / float64(pop)
					}

					hasher := md5.New()
					hasher.Write([]byte(fmt.Sprintf("%d-%.3f-%.3f", currentStep, cX, cY)))
					pID := fmt.Sprintf("%x", hasher.Sum(nil))[:8]

					activePatterns = append(activePatterns, types.PatternProfile{
						ID:         pID,
						FirstSeen:  currentStep,
						LastSeen:   currentStep,
						Population: pop,
						AvgRadius:  avgRadius,
						AvgState:   cSumState / float64(pop),
						MaxState:   cMaxState,
						CenterX:    cX,
						CenterY:    cY,
						CenterZ:    cZ,
					})
				}
			}
		}
	}

	if report.OccupiedCells > 0 {
		report.AvgState /= float64(report.OccupiedCells)
	}
	var globalRadiusSum float64
	for _, p := range activePatterns {
		globalRadiusSum += p.AvgRadius
	}
	if len(activePatterns) > 0 {
		report.AvgRadius = globalRadiusSum / float64(len(activePatterns))
	}

	report.ActivePatterns = activePatterns
	return report
}
