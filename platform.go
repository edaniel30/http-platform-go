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

	"github.com/edaniel30/http-platform-go/internal/constants"
	"github.com/edaniel30/http-platform-go/internal/router"
	"github.com/edaniel30/http-platform-go/internal/telemetry"
	"github.com/gin-gonic/gin"
)

// Platform is the main HTTP server platform
// It encapsulates server lifecycle, routing, and middleware management
type Platform struct {
	config           Config
	router           *router.GinRouter
	server           *http.Server
	telemetryManager *telemetry.Manager
	mu               sync.RWMutex
	started          bool
}

// New creates a new HTTP platform with the given configuration and options
// The configuration uses the functional options pattern for flexibility
func New(cfg Config, opts ...Option) (*Platform, error) {
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Initialize telemetry if enabled
	var tm *telemetry.Manager
	if cfg.Telemetry != nil {
		ctx := context.Background()
		var err error
		tm, err = telemetry.InitTelemetry(
			ctx,
			cfg.Telemetry.ServiceName,
			cfg.Telemetry.Version,
			cfg.Telemetry.Environment,
			cfg.Telemetry.OTLPEndpoint,
			cfg.Telemetry.SampleAll,
		)
		if err != nil {
			cfg.Logger.Error(ctx, "failed to initialize telemetry", map[string]any{"error": err})
			// Don't fail the entire platform startup, just log the error
			tm = nil
		} else {
			cfg.Logger.Info(ctx, "Telemetry initialized successfully", map[string]any{})
		}
	}

	// Create router with configuration
	var serviceName string
	if cfg.Telemetry != nil {
		serviceName = cfg.Telemetry.ServiceName
	}

	router := router.NewGinRouter(router.RouterConfig{
		Mode:                      cfg.Mode,
		TrustedProxies:            cfg.TrustedProxies,
		EnableTraceID:             cfg.EnableTraceID,
		EnableLogger:              cfg.EnableLogger,
		EnableContextCancellation: cfg.EnableContextCancellation,
		CORS:                      cfg.CORS,
		ServiceName:               serviceName,
		BasePath:                  cfg.BasePath,
		Logger:                    cfg.Logger,
	})

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
	if err := p.validateNotStarted(); err != nil {
		return err
	}

	p.createHTTPServer()
	quit := p.setupSignalHandling()
	errChan := p.startServerAsync(ctx)

	if err := p.waitForShutdownSignal(ctx, quit, errChan); err != nil {
		return err
	}

	return p.gracefulShutdown()
}

// validateNotStarted checks if the server is already started and marks it as started
func (p *Platform) validateNotStarted() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return newRuntimeError("server already started", nil)
	}
	p.started = true
	return nil
}

// createHTTPServer initializes the HTTP server with the configured settings
func (p *Platform) createHTTPServer() {
	addr := fmt.Sprintf(":%d", p.config.Port)
	p.server = &http.Server{
		Addr:           addr,
		Handler:        p.router.Handler(),
		ReadTimeout:    p.config.ReadTimeout,
		WriteTimeout:   p.config.WriteTimeout,
		IdleTimeout:    p.config.IdleTimeout,
		MaxHeaderBytes: p.config.MaxHeaderBytes,
	}
}

// setupSignalHandling configures OS signal handling for graceful shutdown
func (p *Platform) setupSignalHandling() chan os.Signal {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	return quit
}

// startServerAsync starts the HTTP server in a goroutine and returns an error channel
func (p *Platform) startServerAsync(ctx context.Context) chan error {
	errChan := make(chan error, 1)
	go func() {
		p.config.Logger.Info(ctx, "Server started", map[string]any{
			"port": p.config.Port,
			"mode": p.config.Mode,
		})

		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- newRuntimeError("server failed to start", err)
		}
	}()
	return errChan
}

// waitForShutdownSignal blocks until a shutdown signal is received or an error occurs
func (p *Platform) waitForShutdownSignal(ctx context.Context, quit chan os.Signal, errChan chan error) error {
	select {
	case <-quit:
		p.config.Logger.Info(ctx, "Shutdown signal received", map[string]any{})
	case <-ctx.Done():
		p.config.Logger.Info(ctx, "Context cancelled, shutting down", map[string]any{})
	case err := <-errChan:
		return err
	}
	return nil
}

// shutdownComponents performs shutdown of server and telemetry, accumulating errors
func (p *Platform) shutdownComponents(ctx context.Context) []error {
	var shutdownErrors []error

	// Shutdown server
	if err := p.server.Shutdown(ctx); err != nil {
		p.config.Logger.Error(ctx, "Error during server shutdown", map[string]any{"error": err})
		shutdownErrors = append(shutdownErrors, newRuntimeError("server shutdown failed", err))
	}

	// Shutdown telemetry if initialized (always attempt even if server shutdown failed)
	if p.telemetryManager != nil {
		if err := p.telemetryManager.Shutdown(ctx); err != nil {
			p.config.Logger.Error(ctx, "Error shutting down telemetry", map[string]any{"error": err})
			shutdownErrors = append(shutdownErrors, newRuntimeError("telemetry shutdown failed", err))
		}
	}

	return shutdownErrors
}

// gracefulShutdown performs a graceful shutdown of the server and telemetry
func (p *Platform) gracefulShutdown() error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer cancel()

	shutdownErrors := p.shutdownComponents(shutdownCtx)

	// Return accumulated errors if any
	if len(shutdownErrors) > 0 {
		// Log summary of errors
		p.config.Logger.Error(shutdownCtx, "Server stopped with errors", map[string]any{
			"error_count": len(shutdownErrors),
		})
		// Return first error (most critical: server shutdown)
		// Additional errors are already logged
		return shutdownErrors[0]
	}

	p.config.Logger.Info(shutdownCtx, "Server stopped gracefully", map[string]any{})

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
		return newRuntimeError("server not started", nil)
	}

	shutdownErrors := p.shutdownComponents(ctx)

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
func (p *Platform) Group(relativePath string, handlers ...gin.HandlerFunc) *router.GinRouterGroup {
	return p.router.Group(relativePath, handlers...)
}

// Router returns the underlying router for advanced usage
func (p *Platform) Router() *router.GinRouter {
	return p.router
}
