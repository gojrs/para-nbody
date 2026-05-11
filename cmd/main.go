package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	// Replace with your actual module path if different
	"github.com/gojrs/para-nbody/db"
	"github.com/gojrs/para-nbody/handlers"
)

func main() {
	// 1. Initialize the Accountant (Database)
	dbPath := filepath.Join(".", "dataset", "para-nbody-v2.store")

	// Open the local connection
	store, err := initLocalDB(dbPath)
	if err != nil {
		log.Fatalf("Critical: Could not initialize database: %v", err)
	}

	// Link the connection to our modular db package
	if err := db.SetDB(store); err != nil {
		log.Fatalf("Critical: Failed to link modular store: %v", err)
	}

	// 2. Setup the Gin Engine
	router := gin.Default()

	// Routes matching your existing API structure
	apiV1Pn := router.Group("/api/v1/pnbody")
	{
		apiV1Pn.POST("/", handlers.HandlePNBody)
		apiV1Pn.POST("/by/:id", handlers.HandlePNBodyByID)
		apiV1Pn.POST("/with/", handlers.HandlePNBodyIni)
	}

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
	<-quit
	log.Println("Shutdown signal received. Closing the GSON Lab...")

	// Create a context with a 10-second timeout for the shutdown process
	// This gives active simulations time to finish saving to the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Close the database connection last
	if store != nil {
		log.Println("Closing database connection...")
		store.Close()
	}

	log.Println("GSON Lab offline. Safe travels.")
}

// initLocalDB is your existing local initialization helper
func initLocalDB(path string) (*sql.DB, error) {
	// 1. Open the connection using the sqlite3 driver
	dbConn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// 2. Test the connection immediately
	if err := dbConn.Ping(); err != nil {
		return nil, err
	}

	// 3. Optional: Set Concurrency Tuning for SQLite
	dbConn.SetMaxOpenConns(1)

	return dbConn, nil
}
