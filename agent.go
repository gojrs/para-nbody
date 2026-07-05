package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type SweepConfig struct {
	Steps               int     `json:"steps"`
	GridSize            int     `json:"grid_size"`
	GravitySensitivity  float64 `json:"gravity_sensitivity"`
	BaseMigrationRate   float64 `json:"base_migration_rate"`
	StickyClumpRate     float64 `json:"sticky_clump_rate"`
	PhaseRelaxationRate float64 `json:"phase_relaxation_rate"`
	TwistFollowRate     float64 `json:"twist_follow_rate"`
}

type InventoryResponse struct {
	ID      string           `json:"universe_id"`
	Metrics map[string]int64 `json:"metrics"`
}

func main() {
	sweepUrl := "https://pnbody-api.codethematrix.dev/api/v1/pnbody/wave-sweep"

	// Start at the upper bound of our discovered stable pocket
	currentRelaxation := 0.0350
	lowerBound := 0.0150
	stepSize := 0.0020

	fmt.Println("🤖 Initializing High-Resolution Telemetry Sweep...")
	fmt.Println("🎯 Targeting Operational Window: 0.0350 -> 0.0150")
	fmt.Println("-------------------------------------------------------------------------")

	iteration := 1
	for currentRelaxation >= lowerBound {
		// Maintain a proportional scale for the electrical twist follow-through
		twistFollow := currentRelaxation * 0.5

		fmt.Printf("🏃 Iteration %2d | Testing Phase Relaxation: %.4f (Twist: %.4f)...\n",
			iteration, currentRelaxation, twistFollow)

		cfg := SweepConfig{
			Steps:               1000,
			GridSize:            40,
			GravitySensitivity:  0.001,
			BaseMigrationRate:   0.25,
			StickyClumpRate:     0.05,
			PhaseRelaxationRate: currentRelaxation,
			TwistFollowRate:     twistFollow,
		}

		jsonData, _ := json.Marshal(cfg)
		resp, err := http.Post(sweepUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatalf("Server connection lost: %v", err)
		}

		var sweepRes struct {
			UniverseID string `json:"universe_id"`
		}
		json.NewDecoder(resp.Body).Decode(&sweepRes)
		resp.Body.Close()

		time.Sleep(300 * time.Millisecond)

		invUrl := fmt.Sprintf("https://pnbody-api.codethematrix.dev/api/v1/pnbody/%s/inventory", sweepRes.UniverseID)
		invResp, _ := http.Get(invUrl)

		var invData InventoryResponse
		json.NewDecoder(invResp.Body).Decode(&invData)
		invResp.Body.Close()

		protons := invData.Metrics["Hydrogen Proton Core (H+)"]
		electrons := invData.Metrics["Hydrogen Electron Shell (e-)"]
		ups := invData.Metrics["Up Quark (u)"]
		downs := invData.Metrics["Down Quark (d)"]

		fmt.Printf("   📊 Census -> P: %d | e-: %d | u: %d | d: %d\n",
			protons, electrons, ups, downs)

		currentRelaxation -= stepSize
		iteration++

		time.Sleep(500 * time.Millisecond)
	}
}
