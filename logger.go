package httpplatform

import (
	"context"
)

// Logger is the interface that any logger implementation must satisfy.
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
//	platform, _ := httpplatform.New(cfg, httpplatform.WithLogger(logger))
//	platform.Start(context.Background())
//
// Creating a custom adapter:
//
//	type MyLoggerAdapter struct {
//		logger *mylogger.Logger
//	}
//
//	func (a *MyLoggerAdapter) Info(ctx context.Context, msg string, fields httpplatform.Fields) {
//		a.logger.Info(msg, convertFields(fields))
//	}
//	// ... implement other methods
type Logger interface {
	// Info logs an informational message with optional fields
	Info(ctx context.Context, msg string, fields map[string]any)

	// Error logs an error message with optional fields
	Error(ctx context.Context, msg string, fields map[string]any)

	// Warn logs a warning message with optional fields
	Warn(ctx context.Context, msg string, fields map[string]any)

	// Debug logs a debug message with optional fields
	Debug(ctx context.Context, msg string, fields map[string]any)

	// Close closes the logger and flushes any pending logs.
	// Note: This should be called by the logger creator (usually in main()),
	// not by the platform. Use defer logger.Close() after creating the logger.
	// Returns an error if the logger fails to close or flush properly.
	Close() error
}
