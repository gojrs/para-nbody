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
	X      *int              `json:"x,omitempty"`
	Y      *int              `json:"y,omitempty"`
	Z      *int              `json:"z,omitempty"`
	Recipe types.Multivector `json:"recipe"`
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/pnbody", h.PNBody)
		apiV1.POST("/pnbody/init", h.PNBodyIni)
		apiV1.GET("/pnbody/:id", h.PNBodyByID)
		apiV1.POST("/pnbody/:id/run", h.RunSteps)
	}
}

// PNBody initializes the Genesis world.
//
// It creates a fresh 50x50x50 universe through WorldManager, hydrates one
// Standard Pillar at the center, and returns the center voxel.
// ... existing code ...

func (h *Handler) PNBody(c *gin.Context) {
	const (
		width  = 50
		height = 50
		depth  = 50

		centerX = 25
		centerY = 25
		centerZ = 25
	)

	universeID, err := h.WorldManager.CreateUniverse(width, height, depth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	world, ok, err := h.WorldManager.GetUniverse(universeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
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

	if err := h.WorldManager.UpdateUniverse(universeID, world); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Genesis world initialized",
		"universe_id": universeID,
		"x":           centerX,
		"y":           centerY,
		"z":           centerZ,
		"voxel":       world.Cells[centerX][centerY][centerZ],
	})
}

// ... existing code ...

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

	universeID, err := h.WorldManager.CreateUniverse(req.Width, req.Height, req.Depth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	world, ok, err := h.WorldManager.GetUniverse(universeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "created universe could not be retrieved",
		})
		return
	}

	x := rand.Intn(req.Width)
	y := rand.Intn(req.Height)
	z := rand.Intn(req.Depth)

	if req.X != nil || req.Y != nil || req.Z != nil {
		if req.X == nil || req.Y == nil || req.Z == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "x, y, and z must all be provided together",
			})
			return
		}

		x = *req.X
		y = *req.Y
		z = *req.Z
	}

	if x < 0 || x >= req.Width || y < 0 || y >= req.Height || z < 0 || z >= req.Depth {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "x, y, and z must be within world bounds",
		})
		return
	}

	world.HydratePillar(x, y, z, req.Recipe)

	if err := h.WorldManager.UpdateUniverse(universeID, world); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Laboratory world initialized",
		"universe_id": universeID,
		"x":           x,
		"y":           y,
		"z":           z,
		"voxel":       world.Cells[x][y][z],
	})
}

// ... existing code ...

func (h *Handler) PNBodyByID(c *gin.Context) {
	id := c.Param("id")

	world, ok, err := h.WorldManager.GetUniverse(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       err.Error(),
			"universe_id": id,
		})
		return
	}
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

	world, exists, err := h.WorldManager.GetUniverse(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       err.Error(),
			"universe_id": id,
		})
		return
	}
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

	if err := h.WorldManager.UpdateUniverse(id, world); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       err.Error(),
			"universe_id": id,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"universe_id":     id,
		"steps_completed": steps,
		"status":          "Evolution Success",
	})
}
