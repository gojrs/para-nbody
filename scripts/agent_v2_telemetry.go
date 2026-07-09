package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gojrs/para-nbody/types"
)

func main() {
	protocol := "http"
	host := "172.20.192.10:42069" // Pointed directly to your local node [cite: 148]
	baseURL := fmt.Sprintf("%s://%s/api/v1/pnbody", protocol, host)

	fmt.Println("🤖 Initializing V2 Spatio-Temporal Pattern Audit...")
	fmt.Println("🔬 Target Matrix Window: 42x42x42 Absolute Coordinate Theater")
	fmt.Println("---------------------------------------------------------------------------------")

	// 1. Ignite the initial universe seed state
	cfg := types.NBodyConfig{
		N:                   types.SeedingModeChaos,
		Steps:               500, // Run initial relaxation steps
		GridSize:            42,  // Hard locked to hardware cache-friendly envelope
		KernelRadius:        1,   // Focus on tight localized interactions
		PhaseRelaxationRate: 0.0300,
		TwistFollowRate:     0.0150,
	}

	jsonData, _ := json.Marshal(cfg)
	resp, err := http.Post(baseURL+"/wave-sweep", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("❌ Initialization connection failed: %v", err)
	}

	var sweepRes map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&sweepRes)
	resp.Body.Close()

	universeID, _ := sweepRes["universe_id"].(string)
	if universeID == "" {
		log.Fatalf("❌ Failed to secure runtime Universe ID tokens from server")
	}

	fmt.Printf("🚀 Reality Deployed. Tracking Universe: [%s]\n", universeID)
	fmt.Println("---------------------------------------------------------------------------------")
	fmt.Printf("%-10s | %-12s | %-10s | %-12s | %-20s\n", "Master Step", "Pattern Count", "Max State", "Avg Radius", "Dominant Cluster Center")
	fmt.Println("---------------------------------------------------------------------------------")

	// 2. Continuous Evolutionary Increment & Tracking Loop
	totalSteps := 500
	incrementSize := 500
	maxLifespan := 4000

	for totalSteps <= maxLifespan {
		// Gather telemetry from current snapshot block
		invResp, err := http.Get(fmt.Sprintf("%s/%s/inventory", baseURL, universeID))
		if err != nil {
			log.Printf("⚠️ Telemetry gap at step %d\n", totalSteps)
			continue
		}

		var report types.SpectrumReport
		if err := json.NewDecoder(invResp.Body).Decode(&report); err != nil {
			invResp.Body.Close()
			log.Printf("⚠️ Parse failure on spectrum matrix payload: %v\n", err)
			continue
		}
		invResp.Body.Close()

		// Locate the largest coherent pattern cluster by population weight
		var dominantCenter string = "N/A (Vacuum Static)"
		var maxPop int64 = 0
		for _, p := range report.ActivePatterns {
			if p.Population > maxPop {
				maxPop = p.Population
				dominantCenter = fmt.Sprintf("(%.1f, %.1f, %.1f)", p.CenterX, p.CenterY, p.CenterZ)
			}
		}

		// Print clean snapshot log line
		fmt.Printf("Tick %-5d | %-13d | %-10.2f | %-12.4f | %-20s\n",
			totalSteps, report.PatternCount, report.MaxState, report.AvgRadius, dominantCenter)

		// 🎨 AUTOMATED VISUAL CATCH: Snag the overlapping Venn slice right here!
		downloadSliceImage(baseURL, universeID, totalSteps, 21)
		if totalSteps == maxLifespan {
			break
		}

		// Advance time step coordinates downstream
		runUrl := fmt.Sprintf("%s/%s/run?count=%d", baseURL, universeID, incrementSize)
		runResp, err := http.Post(runUrl, "application/json", nil)
		if err != nil {
			log.Fatalf("\n❌ Master frame sequence severed: %v", err)
		}
		runResp.Body.Close()

		totalSteps += incrementSize
		time.Sleep(200 * time.Millisecond) // Safety settle delay for background storage flushes [cite: 127]
	}
	fmt.Println("---------------------------------------------------------------------------------")
	fmt.Println("🔬 Deep Horizon Sweep Complete. Metrics persisted cleanly inside local cache.")

	// 🗑️ TELEMETRY SCRUB SYSTEM: Wipe the universe state off the server now that we have our frames
	deleteURL := fmt.Sprintf("%s/%s", baseURL, universeID)
	req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err == nil {
		client := &http.Client{}
		delResp, err := client.Do(req)
		if err == nil {
			if delResp.StatusCode == http.StatusOK {
				fmt.Printf("🗑️ Server footprint successfully scrubbed for universe: [%s]\n", universeID)
			}
			delResp.Body.Close()
		}
	}
}

func downloadSliceImage(baseURL, universeID string, step int, targetZ int) {
	// 1. Define the dedicated directory path for this specific reality thread
	dirPath := fmt.Sprintf("./renders/universe_%s", universeID)

	// 2. 🛡️ FOLDER CHECKPOINT: Create the path if it doesn't exist yet (0755 = standard read/write permissions)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Printf("⚠️ Failed to establish rendering pipeline directory: %v\n", err)
		return
	}

	// 3. Construct the URL to hit our V2 matrix visualizer route
	url := fmt.Sprintf("%s/%s/heatmap?slice=%d", baseURL, universeID, targetZ)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("⚠️ Image download skip at step %d: %v\n", step, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var imgBytes bytes.Buffer
	_, _ = imgBytes.ReadFrom(resp.Body)

	// 4. Clean, protected, sequential naming layout isolated inside the folder
	filename := fmt.Sprintf("%s/tick_%04d.png", dirPath, step)

	// Flush bytes safely to the isolated subdirectory
	_ = os.WriteFile(filename, imgBytes.Bytes(), 0644)
}
