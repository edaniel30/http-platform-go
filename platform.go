// Package httpplatform provides a flexible HTTP server platform built on top of Gin
// It offers a clean API for creating HTTP servers with automatic middleware setup,
// logger integration, and graceful shutdown capabilities.
//
// Key features:
//   - Functional options pattern for configuration
//   - Automatic middleware chain (TraceID, CORS, Recovery, Logger)
//   - Logger injection with loki-logger-go
//   - Graceful shutdown with context support
//   - Clean API for route registration
//
// Basic usage:
//
//	logger, _ := loki.New(models.DefaultConfig(), models.WithAppName("my-app"))
//	platform, err := httpplatform.New(
//	    config.DefaultConfig(),
//	    config.WithPort(8080),
//	    config.WithLogger(logger),
//	)
//	if err != nil {
//	    panic(err)
//	}
//
//	platform.GET("/health", func(c *gin.Context) {
//	    c.JSON(200, gin.H{"status": "ok"})
//	})
//
//	ctx := context.Background()
//	platform.Start(ctx)
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
	internalmodels "github.com/edaniel30/http-platform-go/models"
	"github.com/edaniel30/loki-logger-go/models"
	"github.com/gin-gonic/gin"
)

// Platform is the main HTTP server platform
// It encapsulates server lifecycle, routing, and middleware management
type Platform struct {
	config  internalmodels.Config
	router  *adapters.GinRouter
	server  *http.Server
	mu      sync.RWMutex
	started bool
}

// New creates a new HTTP platform with the given configuration and options
// The configuration uses the functional options pattern for flexibility
//
// Example:
//
//	platform, err := httpplatform.New(
//	    config.DefaultConfig(),
//	    config.WithPort(8080),
//	    config.WithLogger(myLogger),
//	    config.WithMode("release"),
//	)
func New(cfg internalmodels.Config, opts ...internalmodels.Option) (*Platform, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	router := adapters.NewGinRouter(cfg)

	p := &Platform{
		config: cfg,
		router: router,
	}

	return p, nil
}

// Start begins listening for HTTP requests
// It starts the server and blocks until context is cancelled or an error occurs
// Graceful shutdown is handled automatically with a 5-second timeout
//
// Example:
//
//	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
//	defer cancel()
//	if err := platform.Start(ctx); err != nil {
//	    log.Fatal(err)
//	}
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
		p.config.Logger.Info("server started", models.Fields{
			"port": p.config.Port,
			"mode": p.config.Mode,
		})

		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- errors.NewRuntimeError("server failed to start", err)
		}
	}()

	select {
	case <-quit:
		p.config.Logger.Info("shutdown signal received")
	case <-ctx.Done():
		p.config.Logger.Info("context cancelled, shutting down")
	case err := <-errChan:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p.config.Logger.Info("shutting down server...")
	if err := p.server.Shutdown(shutdownCtx); err != nil {
		p.config.Logger.Error("error during shutdown", models.Fields{"error": err})
		return errors.NewRuntimeError("shutdown failed", err)
	}

	p.config.Logger.Info("server stopped gracefully")
	p.config.Logger.Close()

	return nil
}

// Stop gracefully shuts down the platform
func (p *Platform) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.server == nil {
		return errors.ErrNotStarted()
	}

	return p.server.Shutdown(ctx)
}

// Use adds custom middleware to the platform
// Middleware is applied in the order it's registered
//
// Example:
//
//	platform.Use(func(c *gin.Context) {
//	    c.Set("custom", "value")
//	    c.Next()
//	})
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
//
// Example:
//
//	api := platform.Group("/api/v1")
//	api.GET("/users", listUsers)
//	api.POST("/users", createUser)
func (p *Platform) Group(relativePath string, handlers ...gin.HandlerFunc) *adapters.GinRouterGroup {
	return p.router.Group(relativePath, handlers...)
}

// Router returns the underlying router for advanced usage
func (p *Platform) Router() *adapters.GinRouter {
	return p.router
}
