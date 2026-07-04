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
	// UPGRADED TO LIVE SECURE HTTPS ENDPOINT
	targetUrl := "https://pnbody-api.codethematrix.dev/api/v1/pnbody/wave-sweep"

	gravities := []float64{0.001, 0.02}
	repulsions := []float64{10.0, 500.0}

	fmt.Println("🤖 Initializing Automated HTTPS Wave-Interference Sweep...")
	fmt.Println("---------------------------------------------------------------------------------")
	fmt.Printf("%-12s | %-12s | %-12s | %-14s | %-12s\n",
		"Gravity G", "Repulsion", "Condensed Cells", "Max Peak Mass", "Tracer Pockets")
	fmt.Println("---------------------------------------------------------------------------------")

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

			var apiRes struct {
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

			fmt.Printf("⚡ %10.3f | %10.1f | %12d | %14.2f | %12d\n",
				g, r, apiRes.Result.FinalCount, apiRes.Result.MaxMass, apiRes.Result.MatterPockets)

			time.Sleep(200 * time.Millisecond)
		}
	}
	fmt.Println("---------------------------------------------------------------------------------")
}
