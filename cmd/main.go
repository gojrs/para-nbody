package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gojrs/para-nbody/engine"
	"github.com/gojrs/para-nbody/handlers"
	storage "github.com/gojrs/para-nbody/store"
)

func main() {
	storeMode := os.Getenv("PNBODY_STORE")

	var universeStore engine.UniverseStore
	var err error

	switch storeMode {
	case "sqlite":
		dbPath := os.Getenv("PNBODY_DB")
		if dbPath == "" {
			dbPath = filepath.Join(".", "dataset", "para-nbody-v2.store")
		}

		universeStore, err = storage.NewSQLiteStore(dbPath)
		if err != nil {
			log.Fatalf("Critical: Could not initialize SQLite store: %v", err)
		}

		log.Printf("Using SQLite universe store: %s", dbPath)

	default:
		universeStore = storage.NewTTLStore(30 * time.Minute)
		log.Println("Using TTL universe store")
	}

	worldManager := engine.NewWorldManager(universeStore)
	handler := handlers.NewHandler(worldManager)

	// 2. Setup the Gin Engine
	router := gin.Default()
	handler.RegisterRoutes(router)

	// 3. Configure the HTTP Server
	srv := &http.Server{
		Addr:    ":42069",
		Handler: router,
	}

	// 4. Start the Server in a Goroutine
	// This prevents the server from blocking the main thread
	go func() {
		log.Printf("GSON Systems Lab online at http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// 5. SAFE SHUTDOWN MOLE: Listen for interrupt signals
	quit := make(chan os.Signal, 1)
	// SIGINT = Ctrl+C, SIGTERM = Kill command
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block here until a signal is received
	// ... existing code ...

	// Block here until a signal is received
	<-quit
	log.Println("Shutdown signal received. Closing the GSON Lab...")

	// Create a context with a 10-second timeout for the shutdown process
	// This gives active simulations time to finish saving to the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}
}
