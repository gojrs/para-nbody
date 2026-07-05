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

	// The Agent's hunting parameters
	currentRelaxation := 0.10
	twistFollow := 0.05

	fmt.Println("🤖 Starting GSON Lab Tuning Agent... Target: Atomic Electron Capture")
	fmt.Println("-------------------------------------------------------------------------")

	for iteration := 1; iteration <= 10; iteration++ {
		fmt.Printf("🏃 Iteration %d | Testing Phase Relaxation Rate: %.4f...\n", iteration, currentRelaxation)

		cfg := SweepConfig{
			Steps:               1000, // Long run to let chemistry cook
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

		// Allow the memory store on the server a brief window to synchronize states
		time.Sleep(500 * time.Millisecond)

		// Query our custom inventory endpoint to look at the resulting census
		invUrl := fmt.Sprintf("https://pnbody-api.codethematrix.dev/api/v1/pnbody/%s/inventory", sweepRes.UniverseID)
		invResp, _ := http.Get(invUrl)

		var invData InventoryResponse
		json.NewDecoder(invResp.Body).Decode(&invData)
		invResp.Body.Close()

		protons := invData.Metrics["Hydrogen Proton Core (H+)"]
		electrons := invData.Metrics["Hydrogen Electron Shell (e-)"]
		ups := invData.Metrics["Up Quark (u)"]
		downs := invData.Metrics["Down Quark (d)"]

		fmt.Printf("   📊 Census Result -> Protons: %d | Electrons: %d | Up Quarks: %d | Down Quarks: %d\n",
			protons, electrons, ups, downs)

		// 🧠 Agent Logic: If electrons are remaining completely uncaptured or bleeding off too fast,
		// cool down the vacuum sheet elasticity to slow down the field propagation velocity
		if electrons > 400 {
			fmt.Println("   📉 Vacuum is too energetic. Dropping relaxation rate to decelerate wave fronts...")
			currentRelaxation -= 0.01
			twistFollow -= 0.003
		} else {
			fmt.Println("   🎯 Electron envelope containment shifting! Locking parameter profile.")
			break
		}

		time.Sleep(1 * time.Second)
	}
}
