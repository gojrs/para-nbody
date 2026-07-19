package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	// Make sure to reference your generated templ views path here:
	// "github.com/gojrs/para-nbody/web/views"
)

func main() {
	router := gin.Default()

	// 🌐 POINT TO YOUR RUNNING PROXMOX CORE NODE
	engineHost := os.Getenv("GSON_ENGINE_HOST")
	if engineHost == "" {
		engineHost = "http://172.20.192.10:42069"
	}

	log.Printf("GSON Portal Server linking telemetry target to: %s", engineHost)

	// 🛠️ REGISTER RE-THEMED WEB SERVER VIEW ROUTS
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to GSON Systems Hub. Future Blog & Particle Lab coming here soon!")
	})

	// Add your hx-get proxy routes or Templ rendering calls here asynchronously...

	srv := &http.Server{
		Addr:    ":8080", // Web traffic operates on standard alternative ports
		Handler: router,
	}

	go func() {
		log.Printf("🤖 Standalone Web Portal active at http://localhost%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Portal listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Web Portal safely...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
