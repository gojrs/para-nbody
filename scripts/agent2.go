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

func main() {
	protocol := "http"            // Direct local network protocol profile bypassing reverse proxies
	host := "172.20.192.10:42069" // Pointed straight to your Proxmox network address
	sweepUrl := fmt.Sprintf("%s://%s/api/v1/pnbody/wave-sweep", protocol, host)

	// Operational window bounds for phase relaxation
	currentRelaxation := 0.0350
	lowerBound := 0.0250
	relaxationStep := 0.0050

	fmt.Println("🤖 Initializing Deep Horizon Spectrum Audit...")
	fmt.Println("🎯 Automated Volumetric Sweep: KR 1 -> 10 Sequential Scan")
	fmt.Println("---------------------------------------------------------------------------------")

	for currentRelaxation >= lowerBound {
		twistFollow := currentRelaxation * 0.5

		fmt.Printf("\n🌊 [Phase Relaxation: %.4f | Twist: %.4f]\n", currentRelaxation, twistFollow)
		fmt.Println("---------------------------------------------------------------------------------")
		fmt.Printf("%-6s | %-12s | %-16s | %-10s | %-18s\n", "KR", "Universe ID", "Engine Mode", "Quarks (u/d)", "Avg Orbit Radius")
		fmt.Println("---------------------------------------------------------------------------------")

		// 🎯 THE FULL RADIAL SPECTRUM LOOP: 1 to 10
		for testKR := 1; testKR <= 10; testKR++ {

			cfg := types.NBodyConfig{
				N:                   types.SeedingModeChaos,
				Steps:               1000,
				GridSize:            42, // Trigger for Engine V2 Phase-Space paths
				GravitySensitivity:  0.001,
				BaseMigrationRate:   0.25,
				StickyClumpRate:     0.05,
				PhaseRelaxationRate: currentRelaxation,
				TwistFollowRate:     twistFollow,
				KernelRadius:        testKR, // Dynamically stepping across the spectrum
			}

			jsonData, _ := json.Marshal(cfg)
			resp, err := http.Post(sweepUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Fatalf("Server connection lost on KR=%d: %v", testKR, err)
			}

			var sweepRes map[string]interface{}
			if err = json.NewDecoder(resp.Body).Decode(&sweepRes); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			universeID := ""
			if val, ok := sweepRes["universe_id"]; ok {
				universeID = fmt.Sprintf("%v", val)
			} else if val, ok := sweepRes["id"]; ok {
				universeID = fmt.Sprintf("%v", val)
			}

			if universeID == "" {
				continue
			}

			// Buffer padding to let asynchronous background updates settle completely
			time.Sleep(500 * time.Millisecond)

			// Fetch the telemetry metrics ledger inventory
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

			quarkTelemetry := fmt.Sprintf("%d / %d", ups, downs)
			fmt.Printf("KR=%-2d | %s | %-16s | %-10s | %.4f voxels\n",
				testKR, displayID, engineMode, quarkTelemetry, avgDistance)

			// Steady cooling pause between massive runs
			time.Sleep(500 * time.Millisecond)
		}

		currentRelaxation -= relaxationStep
		fmt.Println("---------------------------------------------------------------------------------")
	}
}
