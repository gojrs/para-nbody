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
	// Your brand new live public FQDN!
	targetUrl := "http://pnbody-api.codethematrix.dev:42069/api/pnbody/"

	// We are sweeping across tiny, weak-force style gravity sensitivities
	// to see how it balances against high repulsion values!
	gravities := []float64{0.001, 0.005, 0.02}
	repulsions := []float64{10.0, 100.0, 500.0}

	fmt.Println("🤖 Initializing Automated AI Parameter Sweep...")
	fmt.Println("---------------------------------------------------------------------------------")
	fmt.Printf("%-12s | %-12s | %-12s | %-14s | %-12s\n",
		"Gravity G", "Repulsion", "Active Cells", "Max Peak Mass", "Tracer Count")
	fmt.Println("---------------------------------------------------------------------------------")

	for _, g := range gravities {
		for _, r := range repulsions {

			// Constructing the parameterized JSON payload
			cfg := types.NBodyConfig{
				Steps:                       300,
				GridSize:                    40,
				SunMass:                     1000.0,
				MercuryMass:                 1.0,
				MercuryVelocityZ:            0.35,
				MercuryIsMatter:             false, // Testing the antimatter dipole bounce!
				GravitySensitivity:          g,
				UnlikeMassRepulsionStrength: r,
				BaseMigrationRate:           0.25,
				StickyClumpRate:             0.05,
			}

			jsonData, _ := json.Marshal(cfg)

			resp, err := http.Post(targetUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("❌ G: %5.3f | R: %5.1f | Connection Failed: %v\n", g, r, err)
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

			// Small sleep to be respectful to your server resources
			time.Sleep(200 * time.Millisecond)
		}
	}
	fmt.Println("---------------------------------------------------------------------------------")
}
