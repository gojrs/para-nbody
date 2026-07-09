package engine

import (
	"crypto/md5"
	"fmt"
	"math"

	"github.com/gojrs/para-nbody/types"
)

func (w *V2World) GenerateInventory(currentStep int64) types.SpectrumReport {
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

	// 🔬 THE STRUCTURAL FIREWALL GATE
	// Lowered dynamic cutoff threshold to account for V3 scaling dampeners
	const MaterialDensityThreshold = 5.0

	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			for z := 0; z < w.Depth; z++ {
				p := w.Cells[x][y][z]

				// 🌟 V3 POINTER TRANSITION: Safely read tension using the object method
				cellMagnitude := float64(p.CalculateTension())

				if math.IsNaN(cellMagnitude) || math.IsInf(cellMagnitude, 0) {
					continue
				}

				if cellMagnitude < MaterialDensityThreshold {
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

					if math.IsNaN(avgRadius) || math.IsInf(avgRadius, 0) {
						avgRadius = 0.0
					}
					if math.IsNaN(cX) {
						cX = 0.0
					}
					if math.IsNaN(cY) {
						cY = 0.0
					}
					if math.IsNaN(cZ) {
						cZ = 0.0
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
	if math.IsNaN(report.AvgState) {
		report.AvgState = 0.0
	}

	var globalRadiusSum float64
	for _, p := range activePatterns {
		if !math.IsNaN(p.AvgRadius) {
			globalRadiusSum += p.AvgRadius
		}
	}

	if len(activePatterns) > 0 {
		report.AvgRadius = globalRadiusSum / float64(len(activePatterns))
	}
	if math.IsNaN(report.AvgRadius) {
		report.AvgRadius = 0.0
	}

	report.ActivePatterns = activePatterns
	return report
}
