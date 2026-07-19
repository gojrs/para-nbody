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
	"github.com/joho/godotenv"
)

func main() {
	// Load the environment file safely
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, defaulting to system environment variables")
	}

	// Read the variable securely
	apiUrl := os.Getenv("PNBODY_API_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost" // Fallback local default
	}

	//protocol := "http"
	//host := "172.20.192.10:42069"
	baseURL := fmt.Sprintf("%s/api/v1/pnbody", apiUrl)

	// 1. Ignite the initial universe seed state
	cfg := types.NBodyConfig{
		Mode:                types.EngineModeClockwise,
		N:                   types.SeedingModeChaos,
		Steps:               500,
		GridSize:            42,
		KernelRadius:        1,
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

	// Define directory path and establish text audit log destination
	// Define directory path and establish text audit log destination
	dirPath := fmt.Sprintf("./renders/universe_%s", universeID)
	_ = os.MkdirAll(dirPath, 0755)

	// 🌟 NEW: Pretty-print the configuration payload directly into the directory
	configBytes, err := json.MarshalIndent(cfg, "", "    ")
	if err == nil {
		configFile := fmt.Sprintf("%s/config.json", dirPath)
		_ = os.WriteFile(configFile, configBytes, 0644)
	} else {
		log.Printf("⚠️ Warning: Failed to marshal config snapshot: %v", err)
	}

	logFile, _ := os.OpenFile(fmt.Sprintf("%s/telemetry_audit.txt", dirPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer logFile.Close()

	// 🚀 VISUAL RADAR KICKOFF
	fmt.Println("🔬 GSON Systems Lab: V3 Chiral Simulation Engaged.")
	fmt.Printf("🌌 Universe ID: [%s]\n", universeID)
	fmt.Println("🏃 Evolving timeline to 4,000,000 frames... Monitoring visual renders directory.")

	// Write standard header info to your text file asset
	_, _ = logFile.WriteString("Master Step | Pattern Count | Max State  | Avg Radius    | Dominant Cluster Center\n")
	_, _ = logFile.WriteString("---------------------------------------------------------------------------------\n")

	totalSteps := 500
	incrementSize := 500
	maxLifespan := 4000

	for totalSteps <= maxLifespan {
		invResp, err := http.Get(fmt.Sprintf("%s/%s/inventory", baseURL, universeID))
		if err != nil {
			totalSteps += incrementSize
			continue
		}

		var report types.SpectrumReport
		if err := json.NewDecoder(invResp.Body).Decode(&report); err != nil {
			invResp.Body.Close()
			continue
		}
		invResp.Body.Close()

		var dominantCenter string = "N/A (Vacuum Static)"
		var maxPop int64 = 0
		for _, p := range report.ActivePatterns {
			if p.Population > maxPop {
				maxPop = p.Population
				dominantCenter = fmt.Sprintf("(%.1f, %.1f, %.1f)", p.CenterX, p.CenterY, p.CenterZ)
			}
		}

		// 📝 SILENT LOGGING: Append the statistics strictly to the text file asset
		logLine := fmt.Sprintf("Tick %-5d | %-13d | %-10.2f | %-12.4f | %-20s\n",
			totalSteps, report.PatternCount, report.MaxState, report.AvgRadius, dominantCenter)
		_, _ = logFile.WriteString(logLine)

		// 🎨 VISUAL SNAPSHOT: Fetch slice image straight to the isolated folder
		downloadSliceImage(baseURL, universeID, totalSteps, 21)

		if totalSteps == maxLifespan {
			break
		}

		// Advance time step coordinates downstream
		runUrl := fmt.Sprintf("%s/%s/run?count=%d", baseURL, universeID, incrementSize)
		runResp, err := http.Post(runUrl, "application/json", nil)
		if err != nil {
			break
		}
		runResp.Body.Close()

		totalSteps += incrementSize
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("🏁 4,000,000 Ticks Complete. Evolution stabilized safely in memory cache.")
}

//func downloadSliceImage(baseURL, universeID string, step int, targetZ int) {
//	dirPath := fmt.Sprintf("./renders/universe_%s", universeID)
//	url := fmt.Sprintf("%s/%s/heatmap?slice=%d", baseURL, universeID, targetZ)
//
//	resp, err := http.Get(url)
//	if err != nil {
//		return
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return
//	}
//
//	var imgBytes bytes.Buffer
//	_, _ = imgBytes.ReadFrom(resp.Body)
//	filename := fmt.Sprintf("%s/tick_%07d.png", dirPath, step) // Padded naming configuration
//	_ = os.WriteFile(filename, imgBytes.Bytes(), 0644)
//}
