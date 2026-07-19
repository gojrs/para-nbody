package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gojrs/para-nbody/engine"
	storage "github.com/gojrs/para-nbody/store"
	"github.com/gojrs/para-nbody/types"
)

func main() {
	fmt.Println("🔬 Initializing V3 Finite State Machine Conservation Audit...")
	fmt.Println("-----------------------------------------------------------------")

	// 1. Establish the clean, in-memory TTL Store
	ttlStore := storage.NewTTLStore(10 * time.Minute)
	defer ttlStore.Close()

	universeID := "v3-test-reality"
	size := 32 // Compact matrix window for quick multi-frame evaluation

	// 2. Instantiate our new V3 state machine world
	v3World := &engine.V3World{
		ID:           universeID,
		Width:        size,
		Height:       size,
		Depth:        size,
		KernelRadius: 1,
		Cells:        make([][][]types.Pixel, size),
		Buffer:       make([][][]types.Pixel, size),
	}

	// Allocate nodes & pre-populate with mixed chirality seeds to stress test routing
	for x := 0; x < size; x++ {
		v3World.Cells[x] = make([][]types.Pixel, size)
		v3World.Buffer[x] = make([][]types.Pixel, size)
		for y := 0; y < size; y++ {
			v3World.Cells[x][y] = make([]types.Pixel, size)
			v3World.Buffer[x][y] = make([]types.Pixel, size)
			for z := 0; z < size; z++ {
				// Allocate safe non-nil pointer references for Alice
				wX, wY, wTension := int64(0), int64(0), int64(0)

				// Inject high-tension matter/antimatter collision vortex signatures in the center
				if x > 10 && x < 22 && y > 10 && y < 22 && z > 10 && z < 22 {
					hashSeed := int64(x*y*z + x + y + z)
					if hashSeed%2 == 0 {
						wX, wY, wTension = 35, 20, -10 // Clockwise Matter profile
					} else {
						wX, wY, wTension = -35, -20, 10 // Counter-Clockwise Antimatter profile
					}
				}
				v3World.Cells[x][y][z] = types.NewPixel(wX, wY, wTension, 0, 0, 0)
			}
		}
	}

	// 3. Commit the initial state to the TTL cache
	err := ttlStore.Create(universeID, v3World)
	if err != nil {
		log.Fatalf("❌ Failed to seed TTL Store: %v", err)
	}

	// 4. Calculate Baseline Integer Sums across the grid
	initAliceX, initAliceY, initMalTension := calculateGlobalLedger(v3World)
	initNetTension := calculateTotalScalarTension(v3World)

	fmt.Printf("📊 Frame 0 Baseline Ledger:\n")
	fmt.Printf("   • Sum(Alice.X): %d\n", initAliceX)
	fmt.Printf("   • Sum(Alice.Y): %d\n", initAliceY)
	fmt.Printf("   • Sum(Mal.Z):   %d\n", initMalTension)
	fmt.Printf("   • Net Structural Tension: %d\n", initNetTension)
	fmt.Println("-----------------------------------------------------------------")

	// 5. Evolve the system through the FSM timeline
	testFrames := 500
	fmt.Printf("🏃 Processing %d state machine update loops...", testFrames)

	start := time.Now()
	for frame := 1; frame <= testFrames; frame++ {
		// Retrieve current tick state token from TTL interface
		worldInterface, exists, _ := ttlStore.Get(universeID)
		if !exists {
			log.Fatalf("❌ Reality unlinked from TTL store mid-run!")
		}

		// Type assert onto our concrete V3 state engine layout
		activeWorld := worldInterface.(*engine.V3World)

		// Run the asymmetric step algorithm
		activeWorld.Step()

		// Re-save to flash frame memory
		_ = ttlStore.Update(universeID, activeWorld)
	}
	duration := time.Since(start)
	fmt.Printf(" Done in %v\n", duration)
	fmt.Println("-----------------------------------------------------------------")

	// 6. Pull final snapshot from TTL and audit conservation balances
	finalInterface, _, _ := ttlStore.Get(universeID)
	finalWorld := finalInterface.(*engine.V3World)
	finalAliceX, finalAliceY, finalMalTension := calculateGlobalLedger(finalWorld)
	finalNetTension := calculateTotalScalarTension(finalWorld)

	fmt.Printf("📊 Frame %d Audit Ledger:\n", testFrames)
	fmt.Printf("   • Sum(Alice.X): %d  (Delta: %d)\n", finalAliceX, finalAliceX-initAliceX)
	fmt.Printf("   • Sum(Alice.Y): %d  (Delta: %d)\n", finalAliceY, finalAliceY-initAliceY)
	fmt.Printf("   • Sum(Mal.Z):   %d  (Delta: %d)\n", finalMalTension, finalMalTension-initMalTension)
	fmt.Printf("   • Final Structural Tension: %d\n", finalNetTension)
	fmt.Println("-----------------------------------------------------------------")

	// Strict bit assertion validation check
	if finalAliceX-initAliceX == 0 && finalAliceY-initAliceY == 0 && finalMalTension-initMalTension == 0 {
		fmt.Println("✅ SUCCESS: The FSM rules are mathematically sound! Complete whole-integer conservation achieved.")
	} else {
		fmt.Println("❌ ERROR: Ledger leak detected. State transitions are dropping integer bits.")
	}
}

func calculateGlobalLedger(w *engine.V3World) (int64, int64, int64) {
	var totalX, totalY, totalZ int64
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			for z := 0; z < w.Depth; z++ {
				p := w.Cells[x][y][z]
				if p.Alice != nil {
					totalX += p.Alice.X
					totalY += p.Alice.Y
				}
				totalZ += p.Mal.Tension
			}
		}
	}
	return totalX, totalY, totalZ
}

func calculateTotalScalarTension(w *engine.V3World) int64 {
	var totalTension int64
	for x := 0; x < w.Width; x++ {
		for y := 0; y < w.Height; y++ {
			for z := 0; z < w.Depth; z++ {
				totalTension += w.Cells[x][y][z].CalculateTension()
			}
		}
	}
	return totalTension
}
