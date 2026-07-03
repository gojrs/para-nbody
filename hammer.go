package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gojrs/para-nbody/types"
)

type APIResponse struct {
	ID     int64             `json:"id"`
	Result types.NBodyResult `json:"result"`
}

func main() {
	baseUrl := "http://localhost:42069/api/pnbody/"
	repulsions := []float64{1.0, 10.0, 50.0, 100.0, 250.0, 500.0}

	fmt.Println("🚀 Starting Advanced 3D Cosmic Sweep...")
	fmt.Println("-------------------------------------------------------------------------------------------------------")
	fmt.Printf("%-10s | %-6s | %-12s | %-14s | %-15s | %-15s\n",
		"Repulsion", "ID", "Total Active", "Max Peak Mass", "Matter Pockets", "Antimatter Voids")
	fmt.Println("-------------------------------------------------------------------------------------------------------")

	for _, r := range repulsions {
		cfg := types.NBodyConfig{
			N:                           500,
			BoxSize:                     1000.0,
			MaxSpeed:                    2.0,
			Steps:                       500,
			ParticleMass:                1.0,
			UnlikeMassRepulsionStrength: r,
		}

		jsonData, _ := json.Marshal(cfg)
		resp, err := http.Post(baseUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("❌ Strength %5.1f: Server Offline\n", r)
			continue
		}

		var apiRes struct {
			ID     int64 `json:"id"`
			Result struct {
				FinalCount      int64   `json:"final_count"`
				MaxMass         float64 `json:"max_mass"`
				MatterPockets   int64   `json:"matter_pockets"`
				AntimatterVoids int64   `json:"antimatter_voids"`
			} `json:"result"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&apiRes); err != nil {
			fmt.Printf("❌ Strength %.1f: Decode Error: %v\n", r, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		fmt.Printf("⚡ %8.1f | %3d  | %12d | %14.2f | %15d | %15d\n",
			r, apiRes.ID, apiRes.Result.FinalCount, apiRes.Result.MaxMass,
			apiRes.Result.MatterPockets, apiRes.Result.AntimatterVoids)
	}
}
