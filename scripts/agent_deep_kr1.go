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
	host := "172.20.192.10:42069" // Pointed directly to your local network node
	baseURL := fmt.Sprintf("%s://%s/api/v1/pnbody", protocol, host)

	fmt.Println("🤖 Initializing Continuous Timeline Evolution Agent...")
	fmt.Println("🔬 Tracking a single KR=1 universe across an extended lifespan...")
	fmt.Println("---------------------------------------------------------------------------------")
	fmt.Printf("%-12s | %-12s | %-18s\n", "Total Steps", "Quarks (u/d)", "Avg Orbit Radius")
	fmt.Println("---------------------------------------------------------------------------------")

	// 1. SPIKE THE INITIAL UNIVERSE (Baseline 1,000 Steps)
	cfg := SweepConfig{
		Steps:               1000,
		GridSize:            42, // V2 integer routing flag
		GravitySensitivity:  0.001,
		BaseMigrationRate:   0.25,
		StickyClumpRate:     0.05,
		PhaseRelaxationRate: 0.0300,
		TwistFollowRate:     0.0150,
		KernelRadius:        1, // Hard-locked to local face-to-face metric
	}

	jsonData, _ := json.Marshal(cfg)
	resp, err := http.Post(baseURL+"/wave-sweep", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Initialization failed: %v", err)
	}

	var sweepRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&sweepRes)
	resp.Body.Close()

	universeID := ""
	if val, ok := sweepRes["universe_id"]; ok {
		universeID = fmt.Sprintf("%v", val)
	} else if val, ok := sweepRes["id"]; ok {
		universeID = fmt.Sprintf("%v", val)
	}

	if universeID == "" {
		log.Fatalf("Failed to retrieve Universe ID from server")
	}

	// Brief pause to allow background async disk routines to finalize
	time.Sleep(500 * time.Millisecond)

	// Fetch baseline inventory (1,000 steps)
	printTelemetryRow(baseURL, universeID, 1000)

	// 2. TIMELINE INCREMENT LOOP (Evolve the same world step-by-step)
	totalAccumulatedSteps := 1000
	stepsPerIncrement := 500
	maxTargetSteps := 4000

	for totalAccumulatedSteps < maxTargetSteps {
		// Hit the query run endpoint: POST /api/v1/pnbody/:id/run?count=500
		runUrl := fmt.Sprintf("%s/%s/run?count=%d", baseURL, universeID, stepsPerIncrement)
		runResp, err := http.Post(runUrl, "application/json", nil)
		if err != nil {
			log.Fatalf("\n❌ Server disconnected during evolution loop: %v", err)
		}
		runResp.Body.Close()

		totalAccumulatedSteps += stepsPerIncrement

		// Short pause for database flushing stability
		time.Sleep(500 * time.Millisecond)

		// Audit the live changes to the exact same universe ID
		printTelemetryRow(baseURL, universeID, totalAccumulatedSteps)
	}
	fmt.Println("---------------------------------------------------------------------------------")
}

func printTelemetryRow(baseURL, universeID string, stepCount int) {
	invResp, err := http.Get(fmt.Sprintf("%s/%s/inventory", baseURL, universeID))
	if err != nil {
		return
	}
	defer invResp.Body.Close()

	var invData types.InventoryResponse
	if err := json.NewDecoder(invResp.Body).Decode(&invData); err != nil {
		return
	}

	ups := invData.Metrics[types.KeyUpQuark]
	downs := invData.Metrics[types.KeyDownQuark]
	avgDistance := invData.AverageDistance

	quarkTelemetry := fmt.Sprintf("%d / %d", ups, downs)
	fmt.Printf("Steps=%-5d | %-12s | %.4f voxels\n",
		stepCount, quarkTelemetry, avgDistance)
}
