package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gojrs/para-nbody/types"
)

func main() {
	baseURL := "https://pnbody-api.codethematrix.dev/api/v1/pnbody"
	// UPGRADED TO LIVE SECURE HTTPS ENDPOINT
	targetUrl := fmt.Sprintf("%s/wave-sweep", baseURL)

	gravities := []float64{0.001, 0.02}
	repulsions := []float64{10.0, 500.0}

	fmt.Println("🤖 Initializing Automated HTTPS Wave-Interference Sweep...")
	fmt.Println("----------------------------------------------------------------------------------------------------")
	fmt.Printf("%-10s | %-10s | %-15s | %-13s | %-40s\n",
		"Gravity G", "Repulsion", "Condensed Cells", "Max Peak Mass", "Inventory Telemetry URL")
	fmt.Println("----------------------------------------------------------------------------------------------------")

	for _, g := range gravities {
		for _, r := range repulsions {

			cfg := types.NBodyConfig{
				Steps:                       150, // Evolution ticks
				GridSize:                    40,
				GravitySensitivity:          g,
				UnlikeMassRepulsionStrength: r,
				BaseMigrationRate:           0.25,
				StickyClumpRate:             0.05,
			}

			jsonData, _ := json.Marshal(cfg)

			resp, err := http.Post(targetUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("❌ Connection Failed to SSL Endpoint: %v\n", err)
				continue
			}

			// Capture the top-level "id" (the universe_id string) along with results
			var apiRes struct {
				ID     string `json:"universe_id"` // 🌟 Added to catch the raw creation ID from Gin
				Result struct {
					FinalCount    int64   `json:"final_count"`
					MaxMass       float64 `json:"max_mass"`
					MatterPockets int64   `json:"matter_pockets"`
				} `json:"result"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&apiRes); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			// Print the clean table row with a direct copy-pasteable link to your Proxmox API instance
			inventoryUrl := fmt.Sprintf("%s/%s/inventory", baseURL, apiRes.ID)
			fmt.Printf("⚡ %8.3f | %9.1f | %15d | %13.2f | %-40s\n",
				g, r, apiRes.Result.FinalCount, apiRes.Result.MaxMass, inventoryUrl)

			time.Sleep(200 * time.Millisecond)
		}
	}
	fmt.Println("----------------------------------------------------------------------------------------------------")
}
