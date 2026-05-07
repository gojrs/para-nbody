package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gojrs/para-nbody/db"
	"github.com/gojrs/para-nbody/pnbody"
	"github.com/gojrs/para-nbody/types"
)

// HandlePNBody handles Glass 1: The Standard Paving (Genesis)
func HandlePNBody(c *gin.Context) {
	var cfg types.NBodyConfig
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Run the Optimized Engine
	result := pnbody.Run(cfg, nil)

	// 2. Save to Accountant (Database)
	id, err := db.SaveRun("GENESIS_FLAT", "standard_init", cfg, result, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save run"})
		return
	}

	// 3. Return the result with the Accountant ID
	c.JSON(http.StatusOK, gin.H{
		"id":        id,
		"particles": result.FinalParticles,
		"status":    "GCC simulation complete",
	})
}

// HandlePNBodyByID handles Glass 2: The Time Machine (Resume & Override)
func HandlePNBodyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// 1. Fetch from Store
	oldCfg, oldRes, err := db.GetRunByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Universe not found"})
		return
	}

	// 2. Setup the Seed for the "Time Machine"
	seed := &types.SeedState{
		Particles: oldRes.FinalParticles,
		BoxSize:   oldCfg.BoxSize,
	}

	// 3. CRACK THE CRYSTAL: Apply manual overrides for the new run
	activeCfg := oldCfg
	activeCfg.MaxSpeed = 50.0
	activeCfg.UnlikeMassRepulsionStrength = 500.0
	activeCfg.Steps = 500 // Tuned for quick resolution

	// 4. Fire the Engine with the Seed
	result := pnbody.Run(activeCfg, seed)

	// 5. Accountant: Save the lineage
	newID, _ := db.SaveRun("CRYSTAL_CRACKER", "overridden_run", activeCfg, result, &id)

	c.JSON(http.StatusOK, gin.H{
		"id":     newID,
		"result": result,
	})
}

// HandlePNBodyIni handles Glass 3: The Laboratory (Custom Seed)
func HandlePNBodyIni(c *gin.Context) {
	var payload struct {
		Label  string            `json:"label"`
		Config types.NBodyConfig `json:"config"`
		Seed   types.SeedState   `json:"seed"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := pnbody.Run(payload.Config, &payload.Seed)
	newID, _ := db.SaveRun(payload.Label, "custom_seed", payload.Config, result, nil)

	c.JSON(http.StatusOK, gin.H{
		"id":     newID,
		"result": result,
	})
}
