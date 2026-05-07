package pnbody

import (
	"math"

	"github.com/gojrs/para-nbody/types" // Replace 'yourproject' with your module name
)

const (
	nbodyAnnihilateR = 3.0
	nbodyMergeR      = 5.0
	nbodyDT          = 0.016
	nbodySoften      = 2.0
	nbodyG           = 1.0
	nbodyPocketRatio = 10.0
)

// nbodySim handles the localized simulation state
type nbodySim struct {
	particles         []types.Particle
	annihilationCount int
	mergeCount        int
	pocketCount       int
	cfg               types.NBodyConfig
	hash              nbodyHash
}

// ── spatial hash ─────────────────────────────────────────────────────────────

type nbodyHash struct {
	cells      map[int][]int
	bucketSize float64
	gridWidth  int
	half       float64
}

func newNbodyHash(bucketSize, boxSize float64) nbodyHash {
	gw := int(math.Ceil(boxSize/bucketSize)) + 1
	return nbodyHash{
		cells:      make(map[int][]int),
		bucketSize: bucketSize,
		gridWidth:  gw,
		half:       boxSize / 2,
	}
}

func (h *nbodyHash) Clear() {
	h.cells = make(map[int][]int)
}

func (h *nbodyHash) key(cx, cy, cz int) int {
	cx = ((cx % h.gridWidth) + h.gridWidth) % h.gridWidth
	cy = ((cy % h.gridWidth) + h.gridWidth) % h.gridWidth
	cz = ((cz % h.gridWidth) + h.gridWidth) % h.gridWidth
	return cx + (cy * h.gridWidth) + (cz * h.gridWidth * h.gridWidth)
}

func (h *nbodyHash) insert(idx int, x, y, z float64) {
	cx := int(math.Floor((x + h.half) / h.bucketSize))
	cy := int(math.Floor((y + h.half) / h.bucketSize))
	cz := int(math.Floor((z + h.half) / h.bucketSize))
	h.cells[h.key(cx, cy, cz)] = append(h.cells[h.key(cx, cy, cz)], idx)
}

func (h *nbodyHash) candidates(x, y, z float64) []int {
	cx := int(math.Floor((x + h.half) / h.bucketSize))
	cy := int(math.Floor((y + h.half) / h.bucketSize))
	cz := int(math.Floor((z + h.half) / h.bucketSize))
	var out []int
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			for dz := -1; dz <= 1; dz++ { // Check the 26 surrounding voxels
				out = append(out, h.cells[h.key(cx+dx, cy+dy, cz+dz)]...)
			}
		}
	}
	return out
}

// ── physics helpers ──────────────────────────────────────────────────────────

func nbodyMinImage(d, L float64) float64 {
	if d > L/2 {
		return d - L
	}
	if d < -L/2 {
		return d + L
	}
	return d
}

func nbodyWrap(x, L float64) float64 {
	return x - L*math.Floor((x+L/2)/L)
}

func nbodyForce(ax, ay, az, bx, by, bz, am, bm, ac, bc, repScale, boxSize float64) (fx, fy, fz float64) {
	// WHACK THE MOLE: Remove nbodyMinImage to disable "Ghost Forces" from across the box
	dx := bx - ax
	dy := by - ay
	dz := bz - az

	d2 := dx*dx + dy*dy + dz*dz // True 3D distance

	if d2 < nbodySoften*nbodySoften {
		d2 = nbodySoften * nbodySoften
	}
	d := math.Sqrt(d2)
	cp := ac * bc
	scale := 1.0
	if cp < 0 {
		scale = repScale
	}

	f := nbodyG * am * bm * cp * scale / d2
	return f * dx / d, f * dy / d, f * dz / d
}
