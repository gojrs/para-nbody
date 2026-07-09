package handlers

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gojrs/para-nbody/engine"
	"github.com/gojrs/para-nbody/types"
	"github.com/google/uuid"
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
		apiV1.GET("/pnbody/:id/heatmap", h.GetUniverseHeatmap)
		apiV1.DELETE("/pnbody/:id", h.DeleteUniverse)
	}

	// Support your legacy payload layout sent by hammer.go
	router.POST("/api/pnbody/", h.HandleHammerRequest)
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

	// 1. Create a clean legacy instance explicitly if this initialization endpoint is hit
	id, _ := h.WorldManager.CreateUniverse(req.Width, req.Height, req.Depth)
	world, _, _ := h.WorldManager.GetUniverse(id)

	// 2. 🌟 TYPE ASSERTION: Safely assert onto the concrete legacy *engine.World struct
	if v1World, ok := world.(*engine.World); ok {
		v1World.HydratePillar(req.Width/2, req.Height/2, req.Depth/2, req.Recipe)
		_ = h.WorldManager.UpdateUniverse(id, v1World)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assert universe to legacy V1 layout"})
		return
	}

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

	// 🎭 INSTANTIATE THE NEW V2 PLANCK ENGINE WORLD DIRECTLY
	// 🎭 INSTANTIATE THE NEW V2 PLANCK ENGINE WORLD DIRECTLY
	v2World := engine.NewV2World(
		uuid.New().String(),
		size, size, size,
		cfg.KernelRadius,
		cfg.PhaseRelaxationRate,
		cfg.TwistFollowRate,
		cfg.N,
	)

	// Determine Seeding Mode from unused config flags
	// Let's repurpose N (0 = default formula, 1 = random, 2 = 50/50 split, 3 = even parity)
	seedingMode := cfg.N

	// Seeding Phase: Populate the coordinate theater based on selected macro profile
	for x := 1; x < size-1; x++ {
		for y := 1; y < size-1; y++ {
			for z := 1; z < size-1; z++ {

				var p types.Pixel

				switch seedingMode {
				case types.SeedingModeChaos: // 🎲 MODE 1: PURE QUANTUM CHAOS [cite: 136]
					hashSeed := int64(x*1000 + y*100 + z)
					p = types.NewPixel(
						(hashSeed%71)-35,       // Alice X [cite: 136]
						((hashSeed*13)%41)-20,  // Alice Y [cite: 136]
						uint64(hashSeed%3),     // Dest X [cite: 136]
						uint64((hashSeed/3)%3), // Dest Y [cite: 136]
						uint64((hashSeed/9)%3), // Dest Z [cite: 136]
					)

				case types.SeedingModeSplit: // 🌗 MODE 2: THE 50/50 PHASE SPLIT [cite: 137]
					var bias int64 = 35
					if x > size/2 { //
						bias = -35
					}
					p = types.NewPixel(
						bias,                                  // Alice X [cite: 137]
						int64(20*math.Cos(float64(y)*0.3)),    // Alice Y [cite: 137]
						uint64(x%3), uint64(y%3), uint64(z%3), // Destination Vector [cite: 137]
					)

				default: // 🧘 MODE 0: STANDARD INVARIANT CORE [cite: 137, 138]
					p = types.NewPixel(
						int64(35*math.Sin(float64(x)*0.5)),    // Alice X [cite: 138]
						int64(20*math.Cos(float64(y)*0.3)),    // Alice Y [cite: 138]
						uint64(x%3), uint64(y%3), uint64(z%3), // Destination Vector [cite: 138]
					)
				}

				v2World.SetPixel(x, y, z, p) // [cite: 138]
			}
		}
	}

	// Evolve world step transactions
	for i := 0; i < cfg.Steps; i++ {
		v2World.Step()
	}

	// Persist asynchronously to avoid blocking the network transaction [cite: 127]
	go func(id string, w types.Universe) {
		_ = h.WorldManager.UpdateUniverse(id, w)
	}(v2World.ID, v2World)

	// Fetch immediate status for synchronous return payload delivery
	report := v2World.GenerateInventory(int64(cfg.Steps))

	c.JSON(http.StatusOK, gin.H{
		"universe_id": v2World.ID,
		"engine_mode": "V2_PHASE_SPACE",
		"result": gin.H{
			"final_count":    report.OccupiedCells,
			"max_mass":       report.MaxState,
			"matter_pockets": report.PatternCount,
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

	// Compute current step (defaulted here for snapshot evaluation)
	var currentStep int64 = 1000

	// Check if this matches our V2 architecture blueprint
	if v2, ok := world.(types.V2PlanckUniverse); ok {
		// Type assertion succeeded: Route to the new tracker reporting layer
		if concreteV2, isConcrete := v2.(*engine.V2World); isConcrete {
			report := concreteV2.GenerateInventory(currentStep)
			c.JSON(http.StatusOK, report)
			return
		}
	}

	// 📜 Fallback to legacy V1 Float layout tracking response logic if needed
	c.JSON(http.StatusBadRequest, gin.H{"error": "Legacy V1 inventory fallback not refactored"})
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

	// 🧵 Execute the Step loop directly on the interface token!
	for i := 0; i < steps; i++ {
		world.Step()
	}

	// Persist the updated state back to database storage cleanly
	err = h.WorldManager.UpdateUniverse(id, world)
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

func (h *Handler) HandleHammerRequest(c *gin.Context) {
	var cfg types.NBodyConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	size := cfg.GridSize
	if size == 0 {
		size = 40
	}

	// Create a clean legacy instance explicitly if this endpoint is hit
	universeID := uuid.New().String()
	v1World := engine.NewWorld(universeID, size, size, size)

	// Bind dynamic cosmic levers safely directly to the concrete object
	v1World.RepulsionStrength = cfg.UnlikeMassRepulsionStrength
	v1World.GravitySensitivity = cfg.GravitySensitivity
	v1World.BaseMigrationRate = cfg.BaseMigrationRate
	v1World.StickyClumpRate = cfg.StickyClumpRate

	centerF := float64(size) / 2.0

	// Dynamic Tracer Seeding
	sun := types.TracerBody{
		ID: "sun", IsMatter: true,
		Position: [3]float64{centerF, centerF, centerF}, BaseMass: cfg.SunMass,
	}
	mercury := types.TracerBody{
		ID: "mercury", IsMatter: cfg.MercuryIsMatter,
		Position: [3]float64{centerF + 8.0, centerF, centerF}, Velocity: [3]float64{0, 0, cfg.MercuryVelocityZ},
		BaseMass: cfg.MercuryMass,
	}
	v1World.Tracers = append(v1World.Tracers, sun, mercury)

	// Save the initial footprint to the master store interface
	_ = h.WorldManager.UpdateUniverse(universeID, &v1World)

	// Evolve the system over time
	fmt.Println("\n🪐 Launching Continuous Planetary Tracer Experiment:")
	for i := 1; i <= cfg.Steps; i++ {
		v1World.Step()

		if i%50 == 0 && len(v1World.Tracers) >= 2 {
			pM := v1World.Tracers[1].Position
			vM := v1World.Tracers[1].Velocity
			dx, dy, dz := pM[0]-centerF, pM[1]-centerF, pM[2]-centerF
			distance := math.Sqrt(dx*dx + dy*dy + dz*dz)
			fmt.Printf(" [Tick %3d] -> Mercury Position: (%5.2f, %5.2f, %5.2f) | Distance to Sun: %5.2f voxels | Speed: %5.3f\n",
				i, pM[0], pM[1], pM[2], distance, math.Sqrt(vM[0]*vM[0]+vM[1]*vM[1]+vM[2]*vM[2]))
		}
	}

	// Re-save post execution
	_ = h.WorldManager.UpdateUniverse(universeID, &v1World)

	// Audit final matrix status
	var totalSurvivors int64 = 0
	var maxObservedMass float64 = 0.0
	for x := 0; x < v1World.Width; x++ {
		for y := 0; y < v1World.Height; y++ {
			for z := 0; z < v1World.Depth; z++ {
				m := v1World.Cells[x][y][z].Fields.Matter() + v1World.Cells[x][y][z].Fields.Antimatter()
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
			"matter_pockets":   int64(len(v1World.Tracers)),
			"antimatter_voids": int64(0),
		},
	})
}

func (h *Handler) GetUniverseHeatmap(c *gin.Context) {
	id := c.Param("id")
	world, exists, _ := h.WorldManager.GetUniverse(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe missing"})
		return
	}

	// 1. Enforce high-performance V2 space verification
	v2World, isV2 := world.(*engine.V2World)
	if !isV2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Heatmaps are exclusive to V2 discrete spaces"})
		return
	}

	w, hDim, d := v2World.GetDimensions()

	// 2. Extract slice coordinate parameter or default to exact midpoint plane
	sliceStr := c.DefaultQuery("slice", strconv.Itoa(d/2))
	sliceZ, err := strconv.Atoi(sliceStr)
	if err != nil || sliceZ < 0 || sliceZ >= d {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Slice index out of bounds (0-%d)", d-1)})
		return
	}

	// 3. Scan matrix bounds to find the maximum localized state for color scaling normalization
	var maxObserved float64 = 1.0
	for x := 0; x < w; x++ {
		for y := 0; y < hDim; y++ {
			p := v2World.GetPixel(x, y, sliceZ)
			if p.Alice != nil {
				// Bob's tension is derived directly from Alice's pointer properties
				mag := float64(p.CalculateTension())
				if mag > maxObserved {
					maxObserved = mag
				}
			}
		}
	}

	// 4. Initialize native Go 2D image drawing canvas
	upLeft := image.Point{0, 0}
	lowRight := image.Point{w, hDim}
	img := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})

	// 5. Compute the RGB Venn Overlap Matrix
	for x := 0; x < w; x++ {
		for y := 0; y < hDim; y++ {
			p := v2World.GetPixel(x, y, sliceZ)

			if p.Alice != nil {
				// Calculate individual vector magnitudes cleanly
				aliceMag := math.Sqrt(float64(p.Alice.X*p.Alice.X + p.Alice.Y*p.Alice.Y))
				bobMag := aliceMag // Bob acts as a beautiful virtual calculation of Alice!

				// Normalize scalars to standard 8-bit depth ranges (0-255)
				rVal := uint8((aliceMag / maxObserved) * 255)
				bVal := uint8((bobMag / maxObserved) * 255)

				// Green Channel represents structural field intersection (Venn Diagram sweet spot)
				overlapFactor := (aliceMag * bobMag) / (maxObserved * maxObserved)
				gVal := uint8(overlapFactor * 255)

				// If fields lock together into intense tension, light up the canvas with clear contrast
				if aliceMag > 1000 && bobMag > 1000 {
					img.Set(x, y, color.RGBA{R: rVal, G: 255, B: bVal, A: 255})
				} else {
					img.Set(x, y, color.RGBA{R: rVal, G: gVal, B: bVal, A: 255})
				}
			} else {
				// Empty rooms fall back to pure black vacuum
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}

	// 6. 🛡️ FIREWALL BUFFER: Write PNG to an in-memory buffer first to prevent file corruption
	var imgBuffer bytes.Buffer
	if err := png.Encode(&imgBuffer, img); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compile PNG matrix payload"})
		return
	}

	// Extract the clean, complete byte payload
	pngBytes := imgBuffer.Bytes()

	// Explicitly declare the exact content type and byte size to the incoming network adapter
	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", strconv.Itoa(len(pngBytes)))

	// Write the exact, uncorrupted byte stream down the socket line safely
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(pngBytes)
}

// DeleteUniverse handles dropping an entire timeline from active memory or SQLite arrays
// DeleteUniverse handles dropping an entire timeline from active memory or SQLite arrays
// DeleteUniverse handles dropping an entire timeline from active memory or SQLite arrays
func (h *Handler) DeleteUniverse(c *gin.Context) {
	id := c.Param("id")

	// 1. Check if the target universe exists before trying to destroy it
	_, exists, err := h.WorldManager.GetUniverse(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed verifying universe presence: %v", err)})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe not found; it may have already timed out or been deleted"})
		return
	}

	// 2. 🚀 FIXED: Invoke the public manager wrapper instead of h.WorldManager.store.Delete(id)
	if err := h.WorldManager.DeleteUniverse(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to drop target universe: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"universe_id": id,
		"status":      "Purged",
		"message":     "All coordinate buffers, payloads, and snapshots successfully vaporized.",
	})
}
