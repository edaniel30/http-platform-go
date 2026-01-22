package middleware

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Fields represents a map of structured log fields
type Fields map[string]any

// addBaseRequestFields adds standard request metadata to log fields
// This includes: client_ip, method, path, and trace_id (if available)
func addBaseRequestFields(fields Fields, c *gin.Context) {
	fields["client_ip"] = c.ClientIP()
	fields["method"] = c.Request.Method
	fields["path"] = c.Request.URL.Path

	// Add trace ID if available
	if traceID := GetTraceID(c); traceID != "" {
		fields["trace_id"] = traceID
	}
}

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
			"path":        path,
			"status":      c.Writer.Status(),
			"duration":    duration.String(),
			"duration_ms": duration.Milliseconds(),
		}

		// Add base request metadata (client_ip, method, path, trace_id)
		addBaseRequestFields(fields, c)

		// Add query params if present
		if raw != "" {
			fields["query"] = raw
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields["errors"] = c.Errors.String()
		}

		// Log based on status code
		status := c.Writer.Status()
		ctx := c.Request.Context()
		switch {
		case status >= 500:
			logger.Error(ctx, "Request completed with server error", fields)
		case status >= 400:
			logger.Warn(ctx, "Request completed with client error", fields)
		default:
			logger.Info(ctx, "Request completed", fields)
		}
	}
}

// DefaultLogger is a simple logger implementation using Go's standard log package.
// It's used as the default logger when no custom logger is provided.
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger that writes to stdout.
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "[http-platform] ", log.LstdFlags),
	}
}

// Info logs an informational message with optional fields.
func (l *DefaultLogger) Info(ctx context.Context, msg string, fields Fields) {
	l.log("INFO", msg, fields)
}

// Error logs an error message with optional fields.
func (l *DefaultLogger) Error(ctx context.Context, msg string, fields Fields) {
	l.log("ERROR", msg, fields)
}

// Warn logs a warning message with optional fields.
func (l *DefaultLogger) Warn(ctx context.Context, msg string, fields Fields) {
	l.log("WARN", msg, fields)
}

// Debug logs a debug message with optional fields.
func (l *DefaultLogger) Debug(ctx context.Context, msg string, fields Fields) {
	l.log("DEBUG", msg, fields)
}

// Close closes the logger. Since the standard log package doesn't require cleanup,
// this method is a no-op.
func (l *DefaultLogger) Close() error {
	return nil
}

// log is a helper method that formats and writes log messages.
func (l *DefaultLogger) log(level, msg string, fields Fields) {
	if len(fields) > 0 {
		l.logger.Printf("%s: %s %s", level, msg, formatFields(fields))
	} else {
		l.logger.Printf("%s: %s", level, msg)
	}
}

// formatFields converts Fields map to a readable key=value string format.
func formatFields(fields Fields) string {
	if len(fields) == 0 {
		return ""
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build key=value pairs
	pairs := make([]string, 0, len(fields))
	for _, k := range keys {
		v := fields[k]
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(pairs, " ")
}
