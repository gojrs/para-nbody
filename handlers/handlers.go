package handlers

import (
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gojrs/para-nbody/engine"
	"github.com/gojrs/para-nbody/types"
)

type Handler struct {
	WorldManager *engine.WorldManager
}

func NewHandler(worldManager *engine.WorldManager) *Handler {
	return &Handler{
		WorldManager: worldManager,
	}
}

type PNBodyInitRequest struct {
	Width  int               `json:"width"`
	Height int               `json:"height"`
	Depth  int               `json:"depth"`
	Recipe types.Multivector `json:"recipe"`
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/pnbody", h.PNBody)
	router.POST("/pnbody/init", h.PNBodyIni)
	router.GET("/pnbody/:id", h.PNBodyByID)
	router.POST("/pnbody/:id/run", h.RunSteps)
}

// PNBody initializes the Genesis world.
//
// It creates a fresh 50x50x50 universe through WorldManager, hydrates one
// Standard Pillar at the center, and returns the center voxel.
func (h *Handler) PNBody(c *gin.Context) {
	const (
		width  = 50
		height = 50
		depth  = 50

		centerX = 25
		centerY = 25
		centerZ = 25
	)

	universeID := h.WorldManager.CreateUniverse(width, height, depth)

	world, ok := h.WorldManager.GetUniverse(universeID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "created universe could not be retrieved",
		})
		return
	}

	standardPillar := types.Multivector{
		V: [5]float64{0, 0, 0, 100, 0},
	}

	world.HydratePillar(centerX, centerY, centerZ, standardPillar)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Genesis world initialized",
		"universe_id": universeID,
		"x":           centerX,
		"y":           centerY,
		"z":           centerZ,
		"voxel":       world.Cells[centerX][centerY][centerZ],
	})
}

// PNBodyIni initializes a laboratory world from a JSON request.
//
// Expected JSON:
//
//	{
//	  "width": 50,
//	  "height": 50,
//	  "depth": 50,
//	  "recipe": {
//	    "scalar": 0,
//	    "v": [0, 0, 0, 100, 0]
//	  }
//	}
func (h *Handler) PNBodyIni(c *gin.Context) {
	var req PNBodyInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Width <= 0 || req.Height <= 0 || req.Depth <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "width, height, and depth must all be greater than zero",
		})
		return
	}

	universeID := h.WorldManager.CreateUniverse(req.Width, req.Height, req.Depth)

	world, ok := h.WorldManager.GetUniverse(universeID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "created universe could not be retrieved",
		})
		return
	}

	x := rand.Intn(req.Width)
	y := rand.Intn(req.Height)
	z := rand.Intn(req.Depth)

	world.HydratePillar(x, y, z, req.Recipe)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Laboratory world initialized",
		"universe_id": universeID,
		"x":           x,
		"y":           y,
		"z":           z,
		"voxel":       world.Cells[x][y][z],
	})
}

// PNBodyByID resumes/retrieves a cached universe by ID.
func (h *Handler) PNBodyByID(c *gin.Context) {
	id := c.Param("id")

	world, ok := h.WorldManager.GetUniverse(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "universe not found",
			"universe_id": id,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Universe retrieved",
		"universe_id": id,
		"world":       world,
	})
}

func (h *Handler) RunSteps(c *gin.Context) {
	id := c.Param("id")

	stepsStr := c.DefaultQuery("count", "1")
	steps, err := strconv.Atoi(stepsStr)
	if err != nil || steps <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "count must be a positive integer",
		})
		return
	}

	world, exists := h.WorldManager.GetUniverse(id)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error":       "universe not found",
			"universe_id": id,
		})
		return
	}

	for i := 0; i < steps; i++ {
		world.Step()
	}

	c.JSON(http.StatusOK, gin.H{
		"universe_id":     id,
		"steps_completed": steps,
		"status":          "Evolution Success",
	})
}
