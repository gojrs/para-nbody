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

	id, _ := h.WorldManager.CreateUniverse(req.Width, req.Height, req.Depth)
	world, _, _ := h.WorldManager.GetUniverse(id)

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
		size = 42
	}

	// 🪐 Instantiating clean V3 FSM World matrix
	v3World := &engine.V3World{
		ID:           uuid.New().String(),
		Width:        size,
		Height:       size,
		Depth:        size,
		KernelRadius: cfg.KernelRadius,
		Cells:        make([][][]types.Pixel, size),
		Buffer:       make([][][]types.Pixel, size),
	}

	for x := 0; x < size; x++ {
		v3World.Cells[x] = make([][]types.Pixel, size)
		v3World.Buffer[x] = make([][]types.Pixel, size)
		for y := 0; y < size; y++ {
			v3World.Cells[x][y] = make([]types.Pixel, size)
			v3World.Buffer[x][y] = make([]types.Pixel, size)
			for z := 0; z < size; z++ {
				wX, wY, wTension := int64(0), int64(0), int64(0)

				// 🧪 STRONG-TYPED SWITCH GATES: Evaluation strictly based on enum assignments
				switch cfg.N {
				case types.SeedingModeChaos:
					hashSeed := int64(x*y*z + x + y + z)

					switch cfg.Mode {
					case types.EngineModeCounterRotating:
						if hashSeed%2 == 0 {
							wX, wY, wTension = 35, 20, -10
						} else {
							wX, wY, wTension = -35, -20, 10
						}
					case types.EngineModeClockwise:
						if hashSeed%2 == 0 {
							wX, wY, wTension = int64(35*math.Sin(float64(x))), int64(20*math.Cos(float64(y))), -10
						} else {
							wX, wY, wTension = int64(20*math.Sin(float64(x))), int64(35*math.Cos(float64(y))), -5
						}
					}

				case types.SeedingModeParity:
					switch cfg.Mode {
					case types.EngineModeCounterRotating:
						if x < size/2 {
							wX, wY, wTension = 25, 25, -10
						} else {
							wX, wY, wTension = -25, -25, 10
						}
					case types.EngineModeClockwise:
						wX, wY, wTension = 25, 25, -5
					}
				}

				v3World.Cells[x][y][z] = types.NewPixel(wX, wY, wTension, 0, 0, 0)
			}
		}
	}

	for i := 0; i < cfg.Steps; i++ {
		v3World.Step()
	}

	_ = h.WorldManager.UpdateUniverse(v3World.ID, v3World)

	// 🏷️ DYNAMIC METHOD INVOCATION: Uses your specified String() method implementation natively
	c.JSON(http.StatusOK, gin.H{
		"universe_id":  v3World.ID,
		"engine_mode":  cfg.Mode.String(),
		"seeding_mode": cfg.N.String(),
		"message":      "Unified V3 finite state matrix configured successfully.",
	})
}

// ... rest of the original handlers (PNBodyIni, GetUniverseInventory, etc.) remain intact ...

func (h *Handler) GetUniverseInventory(c *gin.Context) {
	id := c.Param("id")
	world, exists, _ := h.WorldManager.GetUniverse(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe missing"})
		return
	}

	var currentStep int64 = 1000
	if v3, ok := world.(*engine.V3World); ok {
		report := v3.GenerateInventory(currentStep)
		c.JSON(http.StatusOK, report)
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Unrecognized grid engine interface type"})
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

	for i := 0; i < steps; i++ {
		world.Step()
	}

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

	universeID := uuid.New().String()
	v1World := engine.NewWorld(universeID, size, size, size)
	v1World.RepulsionStrength = cfg.UnlikeMassRepulsionStrength
	v1World.GravitySensitivity = cfg.GravitySensitivity
	v1World.BaseMigrationRate = cfg.BaseMigrationRate
	v1World.StickyClumpRate = cfg.StickyClumpRate

	centerF := float64(size) / 2.0
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

	_ = h.WorldManager.UpdateUniverse(universeID, &v1World)

	for i := 1; i <= cfg.Steps; i++ {
		v1World.Step()
	}

	_ = h.WorldManager.UpdateUniverse(universeID, &v1World)

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

	var w, hDim, d int
	var getPixel func(x, y, z int) types.Pixel

	if v3World, isV3 := world.(*engine.V3World); isV3 {
		w, hDim, d = v3World.Width, v3World.Height, v3World.Depth
		getPixel = func(x, y, z int) types.Pixel {
			return v3World.Cells[x][y][z]
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Heatmaps are exclusive to active discrete V3 integer grids"})
		return
	}

	sliceStr := c.DefaultQuery("slice", strconv.Itoa(d/2))
	sliceZ, err := strconv.Atoi(sliceStr)
	if err != nil || sliceZ < 0 || sliceZ >= d {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Slice index out of bounds (0-%d)", d-1)})
		return
	}

	var maxObserved float64 = 1.0
	for x := 0; x < w; x++ {
		for y := 0; y < hDim; y++ {
			p := getPixel(x, y, sliceZ)
			if p.Alice != nil {
				mag := float64(p.CalculateTension())
				if mag > maxObserved {
					maxObserved = mag
				}
			}
		}
	}

	upLeft := image.Point{0, 0}
	lowRight := image.Point{w, hDim}
	img := image.NewRGBA(image.Rectangle{Min: upLeft, Max: lowRight})

	for x := 0; x < w; x++ {
		for y := 0; y < hDim; y++ {
			p := getPixel(x, y, sliceZ)

			if p.Alice != nil {
				aliceMag := math.Sqrt(float64(p.Alice.X*p.Alice.X + p.Alice.Y*p.Alice.Y))
				bobMag := aliceMag

				rVal := uint8((aliceMag / maxObserved) * 255)
				bVal := uint8((bobMag / maxObserved) * 255)

				overlapFactor := (aliceMag * bobMag) / (maxObserved * maxObserved)
				gVal := uint8(overlapFactor * 255)

				if aliceMag > 1000 && bobMag > 1000 {
					img.Set(x, y, color.RGBA{R: rVal, G: 255, B: bVal, A: 255})
				} else {
					img.Set(x, y, color.RGBA{R: rVal, G: gVal, B: bVal, A: 255})
				}
			} else {
				img.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}

	var imgBuffer bytes.Buffer
	if err := png.Encode(&imgBuffer, img); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to compile PNG matrix payload"})
		return
	}

	pngBytes := imgBuffer.Bytes()
	c.Header("Content-Type", "image/png")
	c.Header("Content-Length", strconv.Itoa(len(pngBytes)))
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(pngBytes)
}

func (h *Handler) DeleteUniverse(c *gin.Context) {
	id := c.Param("id")
	_, exists, err := h.WorldManager.GetUniverse(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed verifying universe presence: %v", err)})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe not found"})
		return
	}

	if err := h.WorldManager.DeleteUniverse(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to drop target universe: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"universe_id": id,
		"status":      "Purged",
		"message":     "All coordinate buffers successfully vaporized.",
	})
}
