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

	// Operational window bounds for phase relaxation
	currentRelaxation := 0.0350
	lowerBound := 0.0250 // Slightly narrowed to keep the nested loop runtime safe
	relaxationStep := 0.0050

	fmt.Println("🤖 Initializing Automated Horizon Discovery Agent...")
	fmt.Println("🎯 Scanning Radius Horizons (KR 5 -> 8) across Phase Window...")
	fmt.Println("---------------------------------------------------------------------------------")

	iteration := 1
	for currentRelaxation >= lowerBound {
		twistFollow := currentRelaxation * 0.5

		fmt.Printf("\n🌊 [Phase Relaxation: %.4f | Twist: %.4f]\n", currentRelaxation, twistFollow)
		fmt.Println("---------------------------------------------------------------------------------")
		fmt.Printf("%-6s | %-12s | %-16s | %-10s | %-18s\n", "KR", "Universe ID", "Engine Mode", "Quarks (u/d)", "Avg Orbit Radius")
		fmt.Println("---------------------------------------------------------------------------------")

		// 🎯 NESTED RADIUS HORIZON SWEEP
		// We step from KR=5 up to KR=8 to locate where the orbit size plateaus
		for testKR := 5; testKR <= 8; testKR++ {

			cfg := SweepConfig{
				Steps:               1000,
				GridSize:            42, // 🌟 Explicit code trigger to route into Engine V2 Phase-Space
				GravitySensitivity:  0.001,
				BaseMigrationRate:   0.25,
				StickyClumpRate:     0.05,
				PhaseRelaxationRate: currentRelaxation,
				TwistFollowRate:     twistFollow,
				KernelRadius:        testKR, // Dynamically modifying the horizon bounds
			}

			jsonData, _ := json.Marshal(cfg)
			resp, err := http.Post(sweepUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Fatalf("Server connection lost: %v", err)
			}

			var sweepRes map[string]interface{}
			if err = json.NewDecoder(resp.Body).Decode(&sweepRes); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			// Extract string token safely [cite: 26, 27, 28]
			// Extract string token safely
			universeID := ""
			if val, ok := sweepRes["universe_id"]; ok {
				universeID = fmt.Sprintf("%v", val)
			} else if val, ok := sweepRes["id"]; ok {
				universeID = fmt.Sprintf("%v", val) // 🎯 assign the value here!
			}

			if universeID == "" {
				continue
			}

			// Short pause to allow background async disk tasks to stabilize
			time.Sleep(200 * time.Millisecond)

			// Fetch the calculated metrics ledger inventory [cite: 28, 29]
			invUrl := fmt.Sprintf("%s://%s/api/v1/pnbody/%s/inventory", protocol, host, universeID)
			invResp, err := http.Get(invUrl)
			if err != nil {
				continue
			}

			var invData types.InventoryResponse
			if err := json.NewDecoder(invResp.Body).Decode(&invData); err != nil {
				invResp.Body.Close()
				continue
			}
			invResp.Body.Close()

			// Extract derived telemetry descriptors [cite: 29]
			ups := invData.Metrics[types.KeyUpQuark]
			downs := invData.Metrics[types.KeyDownQuark]
			avgDistance := invData.AverageDistance

			displayID := universeID
			if len(universeID) >= 8 {
				displayID = universeID[:8]
			}

			engineMode := "V2_PHASE_SPACE"
			if val, ok := sweepRes["engine_mode"]; ok {
				engineMode = fmt.Sprintf("%v", val)
			}

			// Print clean row layout for visual delta checking
			quarkTelemetry := fmt.Sprintf("%d / %d", ups, downs)
			fmt.Printf("KR=%-2d | %s | %-16s | %-10s | %.4f voxels\n",
				testKR, displayID, engineMode, quarkTelemetry, avgDistance)

			// Cooling window buffer to keep the pipeline steady
			time.Sleep(400 * time.Millisecond)
		}

		currentRelaxation -= relaxationStep
		iteration++
		fmt.Println("---------------------------------------------------------------------------------")
	}
}
