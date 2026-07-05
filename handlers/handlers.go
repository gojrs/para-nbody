package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gojrs/para-nbody/engine"
	"github.com/gojrs/para-nbody/types"
)

type Handler struct {
	WorldManager *engine.WorldManager
}

func NewHandler(worldManager *engine.WorldManager) *Handler {
	return &Handler{WorldManager: worldManager}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/pnbody/init", h.PNBodyIni)
		apiV1.GET("/pnbody/:id", h.PNBodyByID)
		apiV1.POST("/pnbody/:id/run", h.RunSteps)
		apiV1.POST("/pnbody/wave-sweep", h.HandleWaveSweepRequest)
		apiV1.GET("/pnbody/:id/inventory", h.GetUniverseInventory)
	}

	// Support your legacy payload layout sent by hammer.go
	router.POST("/api/pnbody/", h.HandleHammerRequest)
}

func (h *Handler) HandleHammerRequest(c *gin.Context) {
	var cfg types.NBodyConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Dynamic sizing based on automation needs (fallback to 40 if zero)
	size := cfg.GridSize
	if size == 0 {
		size = 40
	}

	universeID, _ := h.WorldManager.CreateUniverse(size, size, size)
	world, _, _ := h.WorldManager.GetUniverse(universeID)

	// Bind dynamic cosmic levers from the API request
	world.RepulsionStrength = cfg.UnlikeMassRepulsionStrength
	world.GravitySensitivity = cfg.GravitySensitivity
	world.BaseMigrationRate = cfg.BaseMigrationRate
	world.StickyClumpRate = cfg.StickyClumpRate

	centerF := float64(size) / 2.0

	// Dynamic Tracer Seeding
	sun := types.TracerBody{
		ID:       "sun",
		IsMatter: true,
		Position: [3]float64{centerF, centerF, centerF},
		Velocity: [3]float64{0, 0, 0},
		BaseMass: cfg.SunMass,
	}

	mercury := types.TracerBody{
		ID:       "mercury",
		IsMatter: cfg.MercuryIsMatter,
		Position: [3]float64{centerF + 8.0, centerF, centerF},
		Velocity: [3]float64{0, 0, cfg.MercuryVelocityZ},
		BaseMass: cfg.MercuryMass,
	}

	world.Tracers = append(world.Tracers, sun, mercury)

	// 3. Evolve the system over time
	fmt.Println("\n🪐 Launching Continuous Planetary Tracer Experiment:")
	for i := 1; i <= cfg.Steps; i++ {
		world.Step()

		// Print a trace route tracking log every 50 ticks to watch the path shape
		if i%50 == 0 && len(world.Tracers) >= 2 {
			pM := world.Tracers[1].Position // Current coordinates of Mercury
			vM := world.Tracers[1].Velocity // Current velocity vector of Mercury

			// Calculate current radius/distance from the central sun voxel
			dx := pM[0] - centerF
			dy := pM[1] - centerF
			dz := pM[2] - centerF
			distance := math.Sqrt(dx*dx + dy*dy + dz*dz)

			fmt.Printf(" [Tick %3d] -> Mercury Position: (%5.2f, %5.2f, %5.2f) | Distance to Sun: %5.2f voxels | Speed: %5.3f\n",
				i, pM[0], pM[1], pM[2], distance, math.Sqrt(vM[0]*vM[0]+vM[1]*vM[1]+vM[2]*vM[2]))
		}
	}

	_ = h.WorldManager.UpdateUniverse(universeID, world)

	// Audit final matrix status for reporting metrics back to hammer.go
	var totalSurvivors int64 = 0
	var maxObservedMass float64 = 0.0
	for x := 0; x < world.Width; x++ {
		for y := 0; y < world.Height; y++ {
			for z := 0; z < world.Depth; z++ {
				m := world.Cells[x][y][z].Fields.Matter() + world.Cells[x][y][z].Fields.Antimatter()
				if m > 0.1 {
					totalSurvivors++
					if m > maxObservedMass {
						maxObservedMass = m
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id": time.Now().UnixNano() % 1000,
		"result": gin.H{
			"final_count":      totalSurvivors,
			"max_mass":         maxObservedMass,
			"matter_pockets":   int64(len(world.Tracers)), // Re-purposed to report active tracers
			"antimatter_voids": int64(0),
		},
	})
}

func (h *Handler) PNBodyIni(c *gin.Context) {
	var req struct {
		Width  int               `json:"width"`
		Height int               `json:"height"`
		Depth  int               `json:"depth"`
		Recipe types.Multivector `json:"recipe"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, _ := h.WorldManager.CreateUniverse(req.Width, req.Height, req.Depth)
	world, _, _ := h.WorldManager.GetUniverse(id)
	world.HydratePillar(req.Width/2, req.Height/2, req.Depth/2, req.Recipe)
	_ = h.WorldManager.UpdateUniverse(id, world)

	c.JSON(http.StatusOK, gin.H{"universe_id": id, "message": "Laboratory active"})
}

func (h *Handler) PNBodyByID(c *gin.Context) {
	id := c.Param("id")
	world, ok, _ := h.WorldManager.GetUniverse(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe missing"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"world": world})
}

func (h *Handler) RunSteps(c *gin.Context) {
	id := c.Param("id")
	stepsStr := c.DefaultQuery("count", "1")
	steps, err := strconv.Atoi(stepsStr)
	if err != nil || steps <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid count parameter"})
		return
	}

	world, exists, _ := h.WorldManager.GetUniverse(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe missing"})
		return
	}

	// Run calculations sequentially on the master structural object
	for i := 0; i < steps; i++ {
		world.Step()
	}

	// Safe thread block optimization for database flushing
	clonedWorld := *world
	clonedWorld.Cells = world.CloneCells()
	err = h.WorldManager.UpdateUniverse(id, &clonedWorld)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"universe_id":     id,
		"steps_completed": steps,
		"status":          "Evolution Success",
	})
}

func (h *Handler) HandleWaveSweepRequest(c *gin.Context) {
	var cfg types.NBodyConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	size := cfg.GridSize
	if size == 0 {
		size = 40
	}

	universeID, _ := h.WorldManager.CreateUniverse(size, size, size)
	world, _, _ := h.WorldManager.GetUniverse(universeID)

	world.RepulsionStrength = cfg.UnlikeMassRepulsionStrength
	world.GravitySensitivity = cfg.GravitySensitivity
	world.BaseMigrationRate = cfg.BaseMigrationRate
	world.StickyClumpRate = cfg.StickyClumpRate

	// Fallback to defaults if the agent leaves them unassigned
	world.PhaseRelaxationRate = cfg.PhaseRelaxationRate
	if world.PhaseRelaxationRate == 0 {
		world.PhaseRelaxationRate = 0.1
	}
	world.TwistFollowRate = cfg.TwistFollowRate
	if world.TwistFollowRate == 0 {
		world.TwistFollowRate = 0.05
	}

	// --- UPGRADED SEEDING: OVERLAPPING DUAL-LAYER TUG-OF-WAR ---
	// We dismantle the 1D X-axis fence. Space is now initialized by letting two
	// distinct phase sheets collide and overlap across every single shared voxel coordinate.

	maxDist := math.Sqrt(float64(world.Width*world.Width + world.Height*world.Height + world.Depth*world.Depth))

	for x := 1; x < world.Width-1; x++ {
		for y := 1; y < world.Height-1; y++ {
			for z := 1; z < world.Depth-1; z++ {
				// Base structural amplitude
				world.Cells[x][y][z].Fields.Amplitude = 3.5 * math.Sin(float64(x)*0.5)

				// Layer A: Distance from the origin corner (0,0,0) -> Positive Pull
				distOrigin := math.Sqrt(float64(x*x + y*y + z*z))
				phaseA := distOrigin * 0.2
				twistA := 1.0

				// Layer B: Distance from the opposite far corner -> Negative Pull
				dxFar := world.Width - x
				dyFar := world.Height - y
				dzFar := world.Depth - z
				distFar := math.Sqrt(float64(dxFar*dxFar + dyFar*dyFar + dzFar*dzFar))
				phaseB := distFar * -0.2
				twistB := -1.0

				// 🪢 The Shared Voxel Intersection: Summing the overlapping fields
				// This creates the emergent "rope" that your simulation engine will pull on.
				world.Cells[x][y][z].Fields.Phase = phaseA + phaseB

				// Calculate a normalized spatial weight based on where the voxel sits between the two origins
				weightA := 1.0 - (distOrigin / maxDist)
				weightB := 1.0 - (distFar / maxDist)

				// The baseline electrical twist V[1] is the net result of the multi-directional tug-of-war
				world.Cells[x][y][z].Fields.V[1] = (twistA * weightA) + (twistB * weightB)
			}
		}
	}

	// Evolve the wave framework across your spatial workers
	for i := 0; i < cfg.Steps; i++ {
		world.Step()
	}

	_ = h.WorldManager.UpdateUniverse(universeID, world)

	// Audit final matrix status
	var totalActive int64 = 0
	var maxObservedMass float64 = 0.0
	for x := 0; x < world.Width; x++ {
		for y := 0; y < world.Height; y++ {
			for z := 0; z < world.Depth; z++ {
				m := world.Cells[x][y][z].Fields.Matter() + world.Cells[x][y][z].Fields.Antimatter()
				if m > 0.1 {
					totalActive++
					if m > maxObservedMass {
						maxObservedMass = m
					}
				}
			}
		}
	}

	// 🌟 FIX: Return the genuine uuid string under the "universe_id" key
	c.JSON(http.StatusOK, gin.H{
		"universe_id": universeID,
		"result": gin.H{
			"final_count":    totalActive,
			"max_mass":       maxObservedMass,
			"matter_pockets": int64(len(world.Tracers)),
		},
	})
}

func (h *Handler) GetUniverseInventory(c *gin.Context) {
	id := c.Param("id")
	world, exists, _ := h.WorldManager.GetUniverse(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe missing"})
		return
	}

	cookbook := types.NewParticleCookbook()

	// Initialize inventory tally sheets
	inventory := map[string]int64{
		"Hydrogen Proton Core (H+)":      0,
		"Hydrogen Electron Shell (e-)":   0,
		"Up Quark (u)":                   0,
		"Down Quark (d)":                 0,
		"Electron Neutrino (v_e)":        0,
		"Unperturbed Vacuum (Surface 0)": 0,
	}

	// Audit the entire 3D voxel grid space
	for x := 0; x < world.Width; x++ {
		for y := 0; y < world.Height; y++ {
			for z := 0; z < world.Depth; z++ {
				cell := world.Cells[x][y][z]

				// Run the multivector properties against the recipe rules
				identity := cookbook.IdentifyVoxel(cell.Fields)
				inventory[identity]++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"universe_id": id,
		"grid_size":   world.Width,
		"metrics":     inventory,
	})
}
