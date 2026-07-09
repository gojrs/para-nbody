package engine

import (
	"encoding/json"
	"math"
	"runtime"
	"sync"

	"github.com/gojrs/para-nbody/types"
)

// V2World satisfies the types.V2PlanckUniverse interface completely.
type V2World struct {
	ID                  string                `json:"id"`
	Width               int                   `json:"width"`
	Height              int                   `json:"height"`
	Depth               int                   `json:"depth"`
	KernelRadius        int                   `json:"kernel_radius"`
	PhaseRelaxationRate float64               `json:"phase_relaxation_rate"`
	TwistFollowRate     float64               `json:"twist_follow_rate"`
	Cells               [][][]types.Pixel     `json:"cells"`
	Buffer              [][][]types.Pixel     `json:"buffer"`
	SeedingMode         types.NBodyConfigMode `json:"seeding_mode"`
}

// NewV2World instantiates the high-performance discrete matrix layer.
func NewV2World(id string, w, h, d, kr int, relaxation, twist float64, mode types.NBodyConfigMode) *V2World {
	world := &V2World{
		ID:                  id,
		Width:               w,
		Height:              h,
		Depth:               d,
		KernelRadius:        kr,
		PhaseRelaxationRate: relaxation,
		TwistFollowRate:     twist,
		Cells:               make([][][]types.Pixel, w),
		Buffer:              make([][][]types.Pixel, w),
		SeedingMode:         mode,
	}

	for x := 0; x < w; x++ {
		world.Cells[x] = make([][]types.Pixel, h)
		world.Buffer[x] = make([][]types.Pixel, h)
		for y := 0; y < h; y++ {
			world.Cells[x][y] = make([]types.Pixel, d)
			world.Buffer[x][y] = make([]types.Pixel, d)
		}
	}
	return world
}

func (w *V2World) GetID() string                  { return w.ID }
func (w *V2World) GetDimensions() (int, int, int) { return w.Width, w.Height, w.Depth }
func (w *V2World) ToJSON() ([]byte, error)        { return json.Marshal(w) }

func (w *V2World) GetPixel(x, y, z int) types.Pixel    { return w.Cells[x][y][z] }
func (w *V2World) SetPixel(x, y, z int, p types.Pixel) { w.Cells[x][y][z] = p }

// Step runs the discrete Alice and Bob mechanics with zero runtime heap allocations.
func (w *V2World) Step() {
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

	// Clear write buffer up front to avoid artifact ghosting across frames
	w.clearBuffer(radius)

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

					// Initialize operational type domains locally inside core chunk allocations
					var a types.Alice
					var b types.Bob

					for x := sX; x < eX; x++ {
						for y := sY; y < eY; y++ {
							for z := sZ; z < eZ; z++ {
								room := w.Cells[x][y][z]
								currentTension := room.CalculateTension()

								var aliceExpandX, aliceExpandY, aliceExpandZ int64
								var bobContractX, bobContractY, bobContractZ int64

								// Scan the Kernel Horizon to map density gradients
								for dxLocal := -radius; dxLocal <= radius; dxLocal++ {
									for dyLocal := -radius; dyLocal <= radius; dyLocal++ {
										for dzLocal := -radius; dzLocal <= radius; dzLocal++ {
											target := w.Cells[x+dxLocal][y+dyLocal][z+dzLocal]
											distSq := int64(dxLocal*dxLocal + dyLocal*dyLocal + dzLocal*dzLocal)

											if distSq > 0 {
												targetTension := target.CalculateTension()

												// --- V3 NATURE DUALITY CONSTRAINTS ---
												gradientForce := (targetTension - currentTension) / distSq

												// 🔴 ALICE: Seeks lower density zones (Inflationary Expansion) via component domain
												ax, ay, az := a.Expand(int64(dxLocal), int64(dyLocal), int64(dzLocal), gradientForce)
												aliceExpandX += ax
												aliceExpandY += ay
												aliceExpandZ += az

												// 🔵 BOB: Seeks higher density zones (Gravitational Contraction) via component domain
												bx, by, bz := b.Contract(int64(dxLocal), int64(dyLocal), int64(dzLocal), gradientForce)
												bobContractX += bx
												bobContractY += by
												bobContractZ += bz
											}
										}
									}
								}

								// Initialize the destination buffer slice
								w.Buffer[x][y][z].Destination.X = room.Destination.X + uint64(aliceExpandX+bobContractX)
								w.Buffer[x][y][z].Destination.Y = room.Destination.Y + uint64(aliceExpandY+bobContractY)
								w.Buffer[x][y][z].Destination.Z = room.Destination.Z + uint64(aliceExpandZ+bobContractZ)

								// Maintain structural integrity boundaries via direct Pointer targets
								const PlanckLimit = 12000
								if room.Alice != nil {
									if currentTension > PlanckLimit {
										w.Buffer[x][y][z].Alice.X = room.Alice.X / 2
									} else {
										w.Buffer[x][y][z].Alice.X = room.Alice.X
									}
								}

								// Apply dampening and shift weights to neighbor pixels
								const ScalingDampener = 100
								moveX := int64(w.Buffer[x][y][z].Destination.X) / ScalingDampener
								moveY := int64(w.Buffer[x][y][z].Destination.Y) / ScalingDampener
								moveZ := int64(w.Buffer[x][y][z].Destination.Z) / ScalingDampener

								for i := 0; i < 6; i++ {
									nx := x + dx[i]
									ny := y + dy[i]
									nz := z + dz[i]

									// Alice updates now absorb the collective advection pushes natively
									if w.Buffer[nx][ny][nz].Alice != nil {
										w.Buffer[nx][ny][nz].Alice.Y += moveX * int64(dx[i])
										w.Buffer[nx][ny][nz].Alice.X += moveY * int64(dy[i])
										w.Buffer[nx][ny][nz].Alice.X += moveZ * int64(dz[i])
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

	// Swap matrix sheets
	w.Cells, w.Buffer = w.Buffer, w.Cells
}

func (w *V2World) clearBuffer(radius int) {
	for x := radius; x < w.Width-radius; x++ {
		for y := radius; y < w.Height-radius; y++ {
			for z := radius; z < w.Depth-radius; z++ {
				// Initialize clean default allocations using our explicit constructor
				w.Buffer[x][y][z] = types.NewPixel(0, 0, 0, 0, 0)
			}
		}
	}
}
