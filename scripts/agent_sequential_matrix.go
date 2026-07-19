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
	// --- ⚙️ MASTER LIFESPAN & ENGINE STEP VARIABLES ---
	maxLifespan := 4000000 // Upper limit configuration
	incrementSize := 50000 // Discrete timeline block step size
	initialSteps := 500    // Initial baseline ticks pre-run on sweep

	// Load the environment file safely[cite: 6]
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, defaulting to system environment variables")
	}

	// Read the variable securely[cite: 6]
	apiUrl := os.Getenv("PNBODY_API_URL")
	if apiUrl == "" {
		apiUrl = "http://localhost" // Fallback local default
	}

	baseURL := fmt.Sprintf("%s/api/v1/pnbody", apiUrl)

	// 🎭 DEFINE THE 4-WAY CONFIGURATION COMBINATORICS ARRAY
	matrixRuns := []struct {
		engineMode  types.EngineMode
		seedingMode types.NBodyConfigMode
		label       string
	}{
		{types.EngineModeClockwise, types.SeedingModeParity, "V2_Engine_Parity"},
		{types.EngineModeClockwise, types.SeedingModeChaos, "V2_Engine_Chaos"},
		{types.EngineModeCounterRotating, types.SeedingModeParity, "V3_Engine_Parity_Natures_Choice"},
		{types.EngineModeCounterRotating, types.SeedingModeChaos, "V3_Engine_Chaos"},
	}

	fmt.Printf("🚀 GSON Systems Lab: Sequential Automation Engine Initiated.\n")
	fmt.Printf("📊 Matrix Layout: 4 distinct universes will execute strictly one-at-a-time.\n")
	fmt.Printf("⏱️ Target Parameters: Max Lifespan = %d Ticks | Intermittent Step Blocks = %d Ticks\n", maxLifespan, incrementSize)
	fmt.Println("---------------------------------------------------------------------------------")

	// 🏃 EXECUTE SEQUENTIAL TIMELINE SWEEPS
	for idx, matrixTarget := range matrixRuns {
		fmt.Printf("\n🌌 [Universe %d/4 Configuration Block Engage: %s]\n", idx+1, matrixTarget.label)
		fmt.Println("---------------------------------------------------------------------------------")

		// 1. Hydrate structural payload map natively
		cfg := types.NBodyConfig{
			Mode:                matrixTarget.engineMode,
			N:                   matrixTarget.seedingMode,
			Steps:               initialSteps, // Initial ticks computed during registration handler[cite: 6]
			GridSize:            42,
			KernelRadius:        1,
			PhaseRelaxationRate: 0.0300,
			TwistFollowRate:     0.0150,
		}

		jsonData, _ := json.Marshal(cfg)
		resp, err := http.Post(baseURL+"/wave-sweep", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("❌ Skipping sequence matrix %s - Link initialization failed: %v", matrixTarget.label, err)
			continue
		}

		var sweepRes map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&sweepRes)
		resp.Body.Close()

		universeID, _ := sweepRes["universe_id"].(string)
		if universeID == "" {
			log.Printf("⚠️ Warning: Could not resolve valid Universe ID for matrix target %s", matrixTarget.label)
			continue
		}

		// Establish isolated directory paths
		dirPath := fmt.Sprintf("./renders/universe_%s", universeID)
		_ = os.MkdirAll(dirPath, 0755)

		// 📝 PERSIST CONFIG SNAPSHOT FOR THE GIVEN RUN ENVIRONMENT
		configBytes, err := json.MarshalIndent(cfg, "", "    ")
		if err == nil {
			configFile := fmt.Sprintf("%s/config.json", dirPath)
			_ = os.WriteFile(configFile, configBytes, 0644)
		}

		logFile, _ := os.OpenFile(fmt.Sprintf("%s/telemetry_audit.txt", dirPath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

		// 🚀 DYNAMIC CONSOLE PRINTING VIA PASSED CONTEXT VARIABLES
		fmt.Printf("🌌 Local Workspace Bound: [%s]\n", universeID)
		fmt.Printf("🏃 Evolving timeline to %d frames... Monitoring visual renders directory.\n", maxLifespan)

		_, _ = logFile.WriteString("Master Step | Pattern Count | Max State  | Avg Radius    | Dominant Cluster Center\n")
		_, _ = logFile.WriteString("---------------------------------------------------------------------------------\n")

		// Track ticks tracking loops natively matching initial steps
		totalSteps := initialSteps

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

			logLine := fmt.Sprintf("Tick %-5d | %-13d | %-10.2f | %-12.4f | %-20s\n",
				totalSteps, report.PatternCount, report.MaxState, report.AvgRadius, dominantCenter)
			_, _ = logFile.WriteString(logLine)

			downloadSliceImage(baseURL, universeID, totalSteps, 21)

			if totalSteps == maxLifespan {
				break
			}

			// Advance time steps downstream safely via context variables
			runUrl := fmt.Sprintf("%s/%s/run?count=%d", baseURL, universeID, incrementSize)
			runResp, err := http.Post(runUrl, "application/json", nil)
			if err != nil {
				break
			}
			runResp.Body.Close()

			totalSteps += incrementSize
			time.Sleep(100 * time.Millisecond)
		}

		logFile.Close()
		fmt.Printf("🏁 Timeline Evolution stabilized cleanly inside memory cache for target %s.\n", matrixTarget.label)
		fmt.Println("---------------------------------------------------------------------------------")
	}

	fmt.Println("\n🎉 All 4 sequential matrix sweeps complete! Ready for comparative diagnostic mapping.")
}

func downloadSliceImage(baseURL, universeID string, step int, targetZ int) {
	dirPath := fmt.Sprintf("./renders/universe_%s", universeID)
	url := fmt.Sprintf("%s/%s/heatmap?slice=%d", baseURL, universeID, targetZ)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var imgBytes bytes.Buffer
	_, _ = imgBytes.ReadFrom(resp.Body)
	filename := fmt.Sprintf("%s/tick_%07d.png", dirPath, step)
	_ = os.WriteFile(filename, imgBytes.Bytes(), 0644)
}
