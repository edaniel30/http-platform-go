package main

import (
	"context"
	"log"

	httpplatform "github.com/edaniel30/http-platform-go"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create platform with minimal configuration
	platform, err := httpplatform.New(
		httpplatform.DefaultConfig(),
		httpplatform.WithPort(8081),
	)
	if err != nil {
		log.Fatalf("Failed to create platform: %v", err)
	}

	// Register a simple health check endpoint
	platform.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Server is running",
		})
	})

	// Start the server (blocks until shutdown signal)
	log.Println("Server starting on http://localhost:8080")
	log.Println("Try: curl http://localhost:8080/health")

	if err := platform.Start(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
