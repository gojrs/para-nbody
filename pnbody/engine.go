package pnbody

import (
	"math"
	"math/rand"
	"runtime"
	"sync"

	"github.com/gojrs/para-nbody/types"
)

// Run is the main entry point for the simulation
func Run(cfg types.NBodyConfig, seed *types.SeedState) types.NBodyResult {
	s := newNbodySim(cfg)
	if seed != nil {
		s.particles = seed.Particles
		s.cfg.BoxSize = seed.BoxSize
	}

	for step := 0; step < cfg.Steps; step++ {
		s.step()
		s.annihilate()
		s.merge()
	}

	return generateResult(s)
}

func newNbodySim(cfg types.NBodyConfig) *nbodySim {
	half := cfg.BoxSize / 2
	ps := make([]types.Particle, cfg.N)
	for i := range ps {
		charge := 1.0
		if i%2 != 0 {
			charge = -1.0
		}
		ps[i] = types.Particle{
			ID:     int64(i),
			X:      (rand.Float64()*2 - 1) * half,
			Y:      (rand.Float64()*2 - 1) * half,
			Z:      (rand.Float64()*2 - 1) * half, // Random Z start
			VX:     (rand.Float64()*2 - 1) * cfg.MaxSpeed,
			VY:     (rand.Float64()*2 - 1) * cfg.MaxSpeed,
			VZ:     (rand.Float64()*2 - 1) * cfg.MaxSpeed, // Random VZ start
			Mass:   cfg.ParticleMass,
			Charge: charge,
		}
	}
	if cfg.CellSize <= 0 {
		cfg.CellSize = 50.0
	}

	return &nbodySim{
		particles: ps,
		cfg:       cfg,
		hash:      newNbodyHash(cfg.CellSize, cfg.BoxSize),
	}
}

func (s *nbodySim) step() {
	n := len(s.particles)
	if n == 0 {
		return
	}

	// 1. PAVE: Refresh the 3D grid
	s.hash.Clear()
	for i := 0; i < n; i++ {
		// Passing X, Y, and the new Z
		s.hash.insert(i, s.particles[i].X, s.particles[i].Y, s.particles[i].Z)
	}

	fx := make([]float64, n)
	fy := make([]float64, n)
	fz := make([]float64, n) // New Z-force slice

	workers := runtime.NumCPU()
	var wg sync.WaitGroup
	chunk := (n + workers - 1) / workers

	for w := 0; w < workers; w++ {
		lo := w * chunk
		hi := lo + chunk
		if hi > n {
			hi = n
		}
		if lo >= n {
			break
		}

		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				a := &s.particles[i]

				// 2. QUERY: Only check nearby 3D voxels
				indices := s.hash.candidates(a.X, a.Y, a.Z)

				for _, j := range indices {
					if i == j {
						continue
					}
					b := &s.particles[j]

					// 3. CALC: 3D Force components
					dfx, dfy, dfz := nbodyForce(
						a.X, a.Y, a.Z,
						b.X, b.Y, b.Z,
						a.Mass, b.Mass,
						a.Charge, b.Charge,
						s.cfg.UnlikeMassRepulsionStrength,
						s.cfg.BoxSize,
					)
					fx[i] += dfx
					fy[i] += dfy
					fz[i] += dfz
				}
			}
		}(lo, hi)
	}
	wg.Wait()

	// 4. INTEGRATE: Move the particles and check for "Escapes"
	L := s.cfg.BoxSize
	limit := L * 2.0 // The "Boundary"
	dead := make([]bool, len(s.particles))

	for i := range s.particles {
		p := &s.particles[i]

		// Apply the forces we calculated in the goroutines
		p.VX += (fx[i] / p.Mass) * nbodyDT
		p.VY += (fy[i] / p.Mass) * nbodyDT
		p.VZ += (fz[i] / p.Mass) * nbodyDT

		// Move them (Notice: no nbodyWrap here!)
		p.X += p.VX * nbodyDT
		p.Y += p.VY * nbodyDT
		p.Z += p.VZ * nbodyDT

		// WHACK THE MOLE: If they pass the limit, mark them dead
		if math.Abs(p.X) > limit || math.Abs(p.Y) > limit || math.Abs(p.Z) > limit {
			dead[i] = true
		}
	}

	// 5. THE CLEANUP: Remove the escaped particles from s.particles
	s.filterDead(dead)
}

func (s *nbodySim) annihilate() {
	n := len(s.particles)
	dead := make([]bool, n)
	// 1. We use a local hash 'h' tuned specifically for the Annihilation Radius
	h := newNbodyHash(nbodyAnnihilateR, s.cfg.BoxSize)

	for i := range s.particles {
		// FIX: Insert into 'h', not 's.hash'
		h.insert(i, s.particles[i].X, s.particles[i].Y, s.particles[i].Z)
	}

	for i := 0; i < n; i++ {
		if dead[i] {
			continue
		}
		for _, j := range h.candidates(s.particles[i].X, s.particles[i].Y, s.particles[i].Z) {
			if j <= i || dead[j] {
				continue
			}
			if s.particles[i].Charge == s.particles[j].Charge {
				continue
			}

			dx := nbodyMinImage(s.particles[i].X-s.particles[j].X, s.cfg.BoxSize)
			dy := nbodyMinImage(s.particles[i].Y-s.particles[j].Y, s.cfg.BoxSize)
			dz := nbodyMinImage(s.particles[i].Z-s.particles[j].Z, s.cfg.BoxSize) // Add Z

			// FIX: 3D Distance check (including dz)
			if math.Sqrt(dx*dx+dy*dy+dz*dz) < nbodyAnnihilateR {
				dead[i], dead[j] = true, true
				s.annihilationCount++
				break
			}
		}
	}
	s.filterDead(dead)
}

func (s *nbodySim) merge() {
	n := len(s.particles)
	dead := make([]bool, n)
	h := newNbodyHash(nbodyMergeR, s.cfg.BoxSize)
	for i := range s.particles {
		h.insert(i, s.particles[i].X, s.particles[i].Y, s.particles[i].Z)
	}
	for i := 0; i < n; i++ {
		if dead[i] {
			continue
		}
		for _, j := range h.candidates(s.particles[i].X, s.particles[i].Y, s.particles[i].Z) {
			if j <= i || dead[j] {
				continue
			}
			if s.particles[i].Charge != s.particles[j].Charge {
				continue
			}

			// WHACK THE MOLE: Use straight distance, not MinImage
			dx := s.particles[i].X - s.particles[j].X
			dy := s.particles[i].Y - s.particles[j].Y
			dz := s.particles[i].Z - s.particles[j].Z

			if math.Sqrt(dx*dx+dy*dy+dz*dz) >= nbodyMergeR {
				continue
			}

			pi, pj := &s.particles[i], &s.particles[j]
			total := pi.Mass + pj.Mass
			ratio := pi.Mass / pj.Mass
			if ratio < 1 {
				ratio = 1 / ratio
			}

			pi.VX = (pi.VX*pi.Mass + pj.VX*pj.Mass) / total
			pi.VY = (pi.VY*pi.Mass + pj.VY*pj.Mass) / total
			pi.VZ = (pi.VZ*pi.Mass + pj.VZ*pj.Mass) / total

			// WHACK THE MOLE: Direct displacement, not MinImage
			toJx := pj.X - pi.X
			toJy := pj.Y - pi.Y
			toJz := pj.Z - pi.Z

			// WHACK THE MOLE: Direct position update, NO nbodyWrap
			pi.X = pi.X + toJx*pj.Mass/total
			pi.Y = pi.Y + toJy*pj.Mass/total
			pi.Z = pi.Z + toJz*pj.Mass/total

			pi.Mass = total
			dead[j] = true
			s.mergeCount++
			if ratio >= nbodyPocketRatio {
				s.pocketCount++
			}
			break
		}
	}
	s.filterDead(dead)
}

func (s *nbodySim) filterDead(dead []bool) {
	live := s.particles[:0]
	for i, p := range s.particles {
		if !dead[i] {
			live = append(live, p)
		}
	}
	s.particles = live
}

func generateResult(s *nbodySim) types.NBodyResult {
	m, am, maxM := 0, 0, 0.0
	for _, p := range s.particles {
		if p.IsMatter() {
			m++
		} else {
			am++
		}
		// WHACK THE MOLE: Find the largest mass in the 3D void
		if p.Mass > maxM {
			maxM = p.Mass
		}
	}

	return types.NBodyResult{
		Config:            s.cfg,
		FinalCount:        len(s.particles),
		MatterCount:       m,
		AntimatterCount:   am,
		AnnihilationCount: s.annihilationCount,
		MergeCount:        s.mergeCount,
		PocketCount:       s.pocketCount,
		MaxMass:           maxM, // PASS THE DATA HERE
		StepsCompleted:    s.cfg.Steps,
		FinalParticles:    s.particles,
	}
}
