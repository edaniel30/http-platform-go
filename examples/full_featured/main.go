// Production-ready HTTP server with all features enabled
//
// Environment variables:
// - PORT: Server port (default: 8081)
// - OTEL_ENDPOINT: OpenTelemetry collector endpoint (optional)
// - ENVIRONMENT: Environment name (default: production)
//
// Docker setup:
//   # Jaeger (with OTLP)
//   docker run -d --name jaeger -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one:latest

package main

import (
	"context"
	"log"
	"os"
	"time"

	httpplatform "github.com/edaniel30/http-platform-go"
	"github.com/edaniel30/http-platform-go/httperrors"
	"github.com/edaniel30/http-platform-go/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment
	otelEndpoint := os.Getenv("OTEL_ENDPOINT")
	environment := getEnv("ENVIRONMENT", "production")

	// Build configuration with all features
	cfg := httpplatform.DefaultConfig()

	// Create platform with all options
	platform, err := httpplatform.New(
		cfg,
		httpplatform.WithPort(8081),
		httpplatform.WithMode("release"),
		httpplatform.WithBasePath("/api/v1"),

		// CORS configuration
		httpplatform.WithCORS(&middleware.CORSConfig{
			AllowedOrigins:   []string{"https://example.com", "https://app.example.com"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
			ExposedHeaders:   []string{"X-Trace-Id"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),

		// OpenTelemetry (if endpoint provided)
		func() httpplatform.Option {
			if otelEndpoint != "" {
				return httpplatform.WithTelemetry(&httpplatform.TelemetryConfig{
					ServiceName:  "full-featured-api",
					Version:      "1.0.0",
					Environment:  environment,
					OTLPEndpoint: otelEndpoint,
					SampleAll:    environment != "production",
				})
			}
			return func(c *httpplatform.Config) {} // No-op
		}(),
	)
	if err != nil {
		log.Fatalf("Failed to create platform: %v", err)
	}

	// Health check (not under /api/v1)
	platform.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"environment": environment,
			"version":     "1.0.0",
		})
	})

	// Users routes
	users := platform.Group("/users")
	{
		users.GET("", listUsersHandler)
		users.GET("/:id", getUserHandler)
		users.POST("", createUserHandler)
		users.PUT("/:id", updateUserHandler)
		users.DELETE("/:id", deleteUserHandler)
	}

	// Products routes
	products := platform.Group("/products")
	{
		products.GET("", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"products": []gin.H{
					{"id": 1, "name": "Product A", "price": 100},
					{"id": 2, "name": "Product B", "price": 200},
				},
			})
		})
	}

	// Admin routes
	admin := platform.Group("/admin")
	{
		admin.GET("/stats", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"total_requests": 1234,
				"active_users":   56,
				"uptime_hours":   24,
			})
		})
	}

	// Start server
	log.Printf("Server starting on http://localhost:%d\n", 8081)
	log.Println("Environment:", environment)
	log.Println("CORS enabled")
	if otelEndpoint != "" {
		log.Println("Telemetry enabled:", otelEndpoint)
		log.Println("View traces at: http://localhost:16686")
	}
	log.Println()
	log.Println("Available routes:")
	log.Println("  GET    /api/v1/health")
	log.Println("  GET    /api/v1/users")
	log.Println("  GET    /api/v1/users/:id")
	log.Println("  POST   /api/v1/users")
	log.Println("  PUT    /api/v1/users/:id")
	log.Println("  DELETE /api/v1/users/:id")
	log.Println("  GET    /api/v1/products")
	log.Println("  GET    /api/v1/admin/stats")

	if err := platform.Start(context.Background()); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Handler functions

func listUsersHandler(c *gin.Context) {
	params := httpplatform.QueryParamsToMap(c)

	c.JSON(200, gin.H{
		"users": []gin.H{
			{"id": 1, "name": "John Doe", "email": "john@example.com"},
			{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
		},
		"filters": params,
	})
}

func getUserHandler(c *gin.Context) {
	userID := c.Param("id")

	// Simulate user not found
	if userID == "999" {
		_ = c.Error(httperrors.NewNotFoundError("User not found"))
		return
	}

	c.JSON(200, gin.H{
		"id":    userID,
		"name":  "User " + userID,
		"email": "user" + userID + "@example.com",
	})
}

func createUserHandler(c *gin.Context) {
	var user struct {
		Name  string `json:"name" binding:"required,min=2,max=50"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		_ = c.Error(httperrors.NewBadRequestError("Invalid input: " + err.Error()))
		return
	}

	c.JSON(201, gin.H{
		"id":    "123",
		"name":  user.Name,
		"email": user.Email,
	})
}

func updateUserHandler(c *gin.Context) {
	userID := c.Param("id")

	var user struct {
		Name  string `json:"name"`
		Email string `json:"email" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		_ = c.Error(httperrors.NewBadRequestError("Invalid input: " + err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"id":      userID,
		"updated": true,
	})
}

func deleteUserHandler(c *gin.Context) {
	userID := c.Param("id")

	c.JSON(200, gin.H{
		"id":      userID,
		"deleted": true,
	})
}

// Helper functions

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
