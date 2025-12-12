// Package httpplatform provides a flexible HTTP server platform built on top of Gin
// It offers a clean API for creating HTTP servers with automatic middleware setup,
// logger integration, and graceful shutdown capabilities.
//
// Key features:
//   - Functional options pattern for configuration
//   - Automatic middleware chain (TraceID, ErrorHandler, ContextCancellation, CORS, Telemetry, Logger)
//   - Logger injection (any logger that implements middleware.Logger interface)
//   - Graceful shutdown with context support
//   - Clean API for route registration
//   - Context cancellation detection for client disconnections
//
// Logger Lifecycle:
// The platform does NOT close the logger during shutdown. The logger lifecycle is managed
// by the caller who created it. This allows the logger to be used by other parts of the
// application and prevents panics if the platform is started/stopped multiple times.
//
// Example:
//
//	logger := mylogger.New()
//	defer logger.Close() // Caller is responsible for closing
//
//	cfg := httpplatform.DefaultConfig()
//	cfg.Logger = logger
//	platform, _ := httpplatform.New(cfg)
//
//	platform.GET("/health", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"status": "ok"})
//	})
//
//	platform.Start(context.Background())
package httpplatform

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/edaniel30/http-platform-go/errors"
	"github.com/edaniel30/http-platform-go/internal/adapters"
	"github.com/edaniel30/http-platform-go/internal/telemetry"
	"github.com/edaniel30/http-platform-go/middleware"
	"github.com/gin-gonic/gin"
)

// Platform is the main HTTP server platform
// It encapsulates server lifecycle, routing, and middleware management
type Platform struct {
	config           Config
	router           *adapters.GinRouter
	server           *http.Server
	telemetryManager *telemetry.TelemetryManager
	mu               sync.RWMutex
	started          bool
}

// New creates a new HTTP platform with the given configuration and options
// The configuration uses the functional options pattern for flexibility
func New(cfg Config, opts ...Option) (*Platform, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Initialize telemetry if enabled
	var tm *telemetry.TelemetryManager
	if cfg.EnableTelemetry {
		telemetryCfg := telemetry.Config{
			ServiceName:    cfg.ServiceName,
			ServiceVersion: cfg.ServiceVersion,
			Environment:    cfg.Environment,
			OTLPEndpoint:   cfg.OTLPEndpoint,
			SampleAll:      cfg.TelemetrySampleAll,
		}

		ctx := context.Background()
		var err error
		tm, err = telemetry.Init(ctx, telemetryCfg)
		if err != nil {
			cfg.Logger.Error(ctx, "failed to initialize telemetry", middleware.Fields{"error": err})
			// Don't fail the entire platform startup, just log the error
			tm = nil
		} else {
			cfg.Logger.Info(ctx, "telemetry initialized successfully", middleware.Fields{
				"service":  cfg.ServiceName,
				"version":  cfg.ServiceVersion,
				"endpoint": cfg.OTLPEndpoint,
			})
		}
	}

	router := adapters.NewGinRouter(cfg)

	p := &Platform{
		config:           cfg,
		router:           router,
		telemetryManager: tm,
	}

	return p, nil
}

// Start begins listening for HTTP requests
// It starts the server and blocks until context is cancelled or an error occurs
// Graceful shutdown is handled automatically with a 5-second timeout
func (p *Platform) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.started {
		p.mu.Unlock()
		return errors.ErrAlreadyStarted()
	}
	p.started = true
	p.mu.Unlock()

	addr := fmt.Sprintf(":%d", p.config.Port)
	p.server = &http.Server{
		Addr:           addr,
		Handler:        p.router.Handler(),
		ReadTimeout:    p.config.ReadTimeout,
		WriteTimeout:   p.config.WriteTimeout,
		IdleTimeout:    p.config.IdleTimeout,
		MaxHeaderBytes: p.config.MaxHeaderBytes,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		p.config.Logger.Info(ctx, "server started", middleware.Fields{
			"port": p.config.Port,
			"mode": p.config.Mode,
		})

		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- errors.NewRuntimeError("server failed to start", err)
		}
	}()

	select {
	case <-quit:
		p.config.Logger.Info(ctx, "shutdown signal received", middleware.Fields{})
	case <-ctx.Done():
		p.config.Logger.Info(ctx, "context cancelled, shutting down", middleware.Fields{})
	case err := <-errChan:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Accumulate all shutdown errors instead of returning early
	var shutdownErrors []error

	// Shutdown server
	p.config.Logger.Info(shutdownCtx, "shutting down server...", middleware.Fields{})
	if err := p.server.Shutdown(shutdownCtx); err != nil {
		p.config.Logger.Error(shutdownCtx, "error during server shutdown", middleware.Fields{"error": err})
		shutdownErrors = append(shutdownErrors, errors.NewRuntimeError("server shutdown failed", err))
	}

	// Shutdown telemetry if initialized (always attempt even if server shutdown failed)
	if p.telemetryManager != nil {
		p.config.Logger.Info(shutdownCtx, "shutting down telemetry...", middleware.Fields{})
		if err := p.telemetryManager.Shutdown(shutdownCtx); err != nil {
			p.config.Logger.Error(shutdownCtx, "error shutting down telemetry", middleware.Fields{"error": err})
			shutdownErrors = append(shutdownErrors, errors.NewRuntimeError("telemetry shutdown failed", err))
		} else {
			p.config.Logger.Info(shutdownCtx, "telemetry shutdown complete", middleware.Fields{})
		}
	}

	// Return accumulated errors if any
	if len(shutdownErrors) > 0 {
		// Log summary of errors
		p.config.Logger.Error(shutdownCtx, "server stopped with errors", middleware.Fields{
			"error_count": len(shutdownErrors),
		})
		// Return first error (most critical: server shutdown)
		// Additional errors are already logged
		return shutdownErrors[0]
	}

	p.config.Logger.Info(shutdownCtx, "server stopped gracefully", middleware.Fields{})

	// Note: We do NOT call Logger.Close() here because:
	// 1. The logger may be used by other parts of the application
	// 2. The logger lifecycle should be managed by the caller who created it
	// 3. Calling Close() here could cause panics if Start() is called multiple times
	// The caller is responsible for closing the logger when the application exits.

	return nil
}

// Stop gracefully shuts down the platform
// If multiple components fail to shutdown, all errors are logged but only the first is returned.
func (p *Platform) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server == nil {
		return errors.ErrNotStarted()
	}

	// Accumulate all shutdown errors instead of returning early
	var shutdownErrors []error

	// Shutdown server
	if err := p.server.Shutdown(ctx); err != nil {
		p.config.Logger.Error(ctx, "error during server shutdown", middleware.Fields{"error": err})
		shutdownErrors = append(shutdownErrors, errors.NewRuntimeError("server shutdown failed", err))
	}

	// Shutdown telemetry if initialized (always attempt even if server shutdown failed)
	if p.telemetryManager != nil {
		if err := p.telemetryManager.Shutdown(ctx); err != nil {
			p.config.Logger.Error(ctx, "error shutting down telemetry", middleware.Fields{"error": err})
			shutdownErrors = append(shutdownErrors, errors.NewRuntimeError("telemetry shutdown failed", err))
		}
	}

	// Return accumulated errors if any
	if len(shutdownErrors) > 0 {
		// Return first error (most critical: server shutdown)
		// Additional errors are already logged
		return shutdownErrors[0]
	}

	return nil
}

// Use adds custom middleware to the platform
// Middleware is applied in the order it's registered
func (p *Platform) Use(middleware ...gin.HandlerFunc) {
	p.router.Use(middleware...)
}

// GET registers a GET route
func (p *Platform) GET(relativePath string, handlers ...gin.HandlerFunc) {
	p.router.GET(relativePath, handlers...)
}

// POST registers a POST route
func (p *Platform) POST(relativePath string, handlers ...gin.HandlerFunc) {
	p.router.POST(relativePath, handlers...)
}

// PUT registers a PUT route
func (p *Platform) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	p.router.PUT(relativePath, handlers...)
}

// DELETE registers a DELETE route
func (p *Platform) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	p.router.DELETE(relativePath, handlers...)
}

// PATCH registers a PATCH route
func (p *Platform) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	p.router.PATCH(relativePath, handlers...)
}

// OPTIONS registers an OPTIONS route
func (p *Platform) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	p.router.OPTIONS(relativePath, handlers...)
}

// HEAD registers a HEAD route
func (p *Platform) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	p.router.HEAD(relativePath, handlers...)
}

// Group creates a new route group with the given prefix
// Useful for organizing related routes under a common path
func (p *Platform) Group(relativePath string, handlers ...gin.HandlerFunc) *adapters.GinRouterGroup {
	return p.router.Group(relativePath, handlers...)
}

// Router returns the underlying router for advanced usage
func (p *Platform) Router() *adapters.GinRouter {
	return p.router
}
