package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"net/http"

	"github.com/gojrs/para-nbody/types"
)

// Response wrapper to match your Gin output
type APIResponse struct {
	ID     int64             `json:"id"`
	Result types.NBodyResult `json:"result"`
}

func main() {
	baseUrl := "http://localhost:42069/api/pnbody/"
	repulsions := []float64{1.0, 10.0, 50.0, 100.0, 250.0, 500.0}

	fmt.Println("🚀 Starting Typed 3D Sweep...")

	for _, r := range repulsions {
		// 1. Setup Payload using the real Config struct
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
			fmt.Printf("❌ Strength %.1f: Server Offline\n", r)
			continue
		}

		// 2. Decode using our APIResponse wrapper
		var apiRes APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiRes); err != nil {
			fmt.Printf("❌ Strength %.1f: Decode Error: %v\n", r, err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		// 3. Print the Audit
		fmt.Printf("✅ Strength %5.1f | ID: %3d | Survivors: %4d | MaxMass: %4.1f\n",
			r, apiRes.ID, apiRes.Result.FinalCount, apiRes.Result.MaxMass)
	}
}
