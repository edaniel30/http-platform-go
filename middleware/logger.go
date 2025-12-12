package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Fields represents a map of structured log fields
type Fields map[string]any

// Logger is the interface that any logger implementation must satisfy
// This allows the platform to be agnostic about the logging implementation.
//
// Important: The platform does NOT call Close() during shutdown. The caller
// who creates the logger is responsible for calling Close() when appropriate.
// This allows the logger to be shared across multiple components and prevents
// issues if the platform is started/stopped multiple times.
//
// Example implementation:
//
//	logger := mylogger.New()
//	defer logger.Close() // Caller closes the logger
//
//	cfg := httpplatform.DefaultConfig()
//	cfg.Logger = logger
//	platform, _ := httpplatform.New(cfg)
//	platform.Start(context.Background())
type Logger interface {
	// Info logs an informational message with optional fields
	Info(ctx context.Context, msg string, fields Fields)

	// Error logs an error message with optional fields
	Error(ctx context.Context, msg string, fields Fields)

	// Warn logs a warning message with optional fields
	Warn(ctx context.Context, msg string, fields Fields)

	// Debug logs a debug message with optional fields
	Debug(ctx context.Context, msg string, fields Fields)

	// Close closes the logger and flushes any pending logs.
	// Note: This should be called by the logger creator (usually in main()),
	// not by the platform. Use defer logger.Close() after creating the logger.
	// Returns an error if the logger fails to close or flush properly.
	Close() error
}

// BasicLogger creates a request logger middleware using the platform logger interface
// This middleware logs all incoming HTTP requests with method, path, status, and duration
func BasicLogger(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Build log fields
		fields := Fields{
			"method":      c.Request.Method,
			"path":        path,
			"status":      c.Writer.Status(),
			"duration":    duration.String(),
			"duration_ms": duration.Milliseconds(),
			"client_ip":   c.ClientIP(),
		}

		// Add query params if present
		if raw != "" {
			fields["query"] = raw
		}

		// Add trace ID if available
		if traceID := GetTraceID(c); traceID != "" {
			fields["trace_id"] = traceID
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		status := c.Writer.Status()
		ctx := c.Request.Context()
		if status >= 500 {
			logger.Error(ctx, "Request completed with server error", fields)
		} else if status >= 400 {
			logger.Warn(ctx, "Request completed with client error", fields)
		} else {
			logger.Info(ctx, "Request completed", fields)
		}
	}
}
