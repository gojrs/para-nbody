package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gojrs/para-nbody/types"
)

type SweepConfig struct {
	Steps               int     `json:"steps"`
	GridSize            int     `json:"grid_size"`
	GravitySensitivity  float64 `json:"gravity_sensitivity"`
	BaseMigrationRate   float64 `json:"base_migration_rate"`
	StickyClumpRate     float64 `json:"sticky_clump_rate"`
	PhaseRelaxationRate float64 `json:"phase_relaxation_rate"`
	TwistFollowRate     float64 `json:"twist_follow_rate"`
	KernelRadius        int     `json:"kernel_radius"`
}

func main() {
	protocol := "http"
	host := "172.20.192.10:42069"
	sweepUrl := fmt.Sprintf("%s://%s/api/v1/pnbody/wave-sweep", protocol, host)

	// Operational window bounds
	currentRelaxation := 0.0350
	lowerBound := 0.0150
	stepSize := 0.0020

	fmt.Println("🤖 Initializing High-Resolution Telemetry Sweep...")
	fmt.Println("🎯 Targeting Operational Window: 0.0350 -> 0.0150")
	fmt.Println("-------------------------------------------------------------------------")

	iteration := 1
	for currentRelaxation >= lowerBound {
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
			KernelRadius:        6, // 🔥 FORCE THE GEOMETRIC EXPANSION 🔥
		}

		jsonData, _ := json.Marshal(cfg)
		resp, err := http.Post(sweepUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatalf("Server connection lost: %v", err)
		}

		// 🎯 Read raw map directly to avoid contract naming friction
		var sweepRes map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&sweepRes); err != nil {
			fmt.Printf("   ❌ JSON Decode Error: %v\n", err)
			resp.Body.Close()
			currentRelaxation -= stepSize
			iteration++
			continue
		}
		resp.Body.Close()

		// Extract the string token safely whether the key is "universe_id" or "id"
		universeID := ""
		if val, ok := sweepRes["universe_id"]; ok {
			universeID = fmt.Sprintf("%v", val)
		} else if val, ok := sweepRes["id"]; ok {
			universeID = fmt.Sprintf("%v", val)
		}

		// Guard check against empty identifiers
		if universeID == "" {
			fmt.Println("   ⚠️ Server returned a blank ID. Skipping iteration...")
			currentRelaxation -= stepSize
			iteration++
			continue
		}

		time.Sleep(300 * time.Millisecond)

		// Formulate the path to query the specific universe metrics ledger
		invUrl := fmt.Sprintf("%s://%s/api/v1/pnbody/%s/inventory", protocol, host, universeID)
		invResp, err := http.Get(invUrl)
		if err != nil {
			fmt.Printf("   ❌ Inventory Fetch Failed: %v\n", err)
			currentRelaxation -= stepSize
			iteration++
			continue
		}

		var invData types.InventoryResponse
		if err := json.NewDecoder(invResp.Body).Decode(&invData); err != nil {
			fmt.Printf("   ❌ Inventory Decode Error: %v\n", err)
			invResp.Body.Close()
			currentRelaxation -= stepSize
			iteration++
			continue
		}
		invResp.Body.Close()

		// Extract types from the standardized shared contract
		protons := invData.Metrics[types.KeyProton]
		electrons := invData.Metrics[types.KeyElectron]
		ups := invData.Metrics[types.KeyUpQuark]
		downs := invData.Metrics[types.KeyDownQuark]
		avgDistance := invData.AverageDistance

		// Shorten long UUID strings down to 8 characters for terminal scannability
		displayID := universeID
		if len(universeID) >= 8 {
			displayID = universeID[:8]
		}

		// 📺 TELEMETRY SHOWCASE WINDOW
		fmt.Printf("   📊 Census -> P: %3d | e-: %3d | u: %3d | d: %3d\n",
			protons, electrons, ups, downs)
		fmt.Printf("   🌌 Matrix -> ID: %s | Avg Orbit Radius: %.2f voxels\n",
			displayID, avgDistance)

		currentRelaxation -= stepSize
		iteration++

		time.Sleep(500 * time.Millisecond)
	}
}
