package router

import (
	"context"
	"net/http"
	"time"

	loggerutil "github.com/edaniel30/http-platform-go/internal/logger"
	"github.com/edaniel30/http-platform-go/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Logger interface is defined here to avoid import cycles.
// Any type that implements httpplatform.Logger will automatically implement this interface
// because they have identical method signatures.
type Logger interface {
	Info(ctx context.Context, msg string, fields map[string]any)
	Error(ctx context.Context, msg string, fields map[string]any)
	Warn(ctx context.Context, msg string, fields map[string]any)
	Debug(ctx context.Context, msg string, fields map[string]any)
	Close() error
}

// ginRouter wraps gin.Engine to provide HTTP routing capabilities.
// This is an internal implementation that shouldn't be exported.
type GinRouter struct {
	engine    *gin.Engine
	baseGroup *gin.RouterGroup // Optional base group when BasePath is configured
}

// ginRouterGroup wraps gin.RouterGroup to provide route grouping capabilities.
// This is an internal implementation that shouldn't be exported.
type GinRouterGroup struct {
	group *gin.RouterGroup
}

type RouterConfig struct {
	Mode                      string
	TrustedProxies            []string
	EnableTraceID             bool
	EnableLogger              bool
	EnableContextCancellation bool
	CORS                      *middleware.CORSConfig
	ServiceName               string
	BasePath                  string
	Logger                    Logger
}

// newGinRouter creates a new Gin router with the given configuration.
// It initializes the Gin engine, applies middleware in the correct order, and sets up the base path if configured.
func NewGinRouter(cfg RouterConfig) *GinRouter {
	// Set gin mode
	gin.SetMode(cfg.Mode)

	// Create engine
	engine := gin.New()

	// Set trusted proxies
	if cfg.TrustedProxies != nil {
		_ = engine.SetTrustedProxies(cfg.TrustedProxies)
	}

	// Apply middleware to engine first
	// Order matters: TraceID -> ErrorHandler -> ContextCancellation -> CORS -> Telemetry -> Logger

	// 1. TraceID - for traceability across the entire pipeline
	if cfg.EnableTraceID {
		engine.Use(middleware.TraceID())
	}

	// 2. ErrorHandler - must be early to catch panics from other middleware
	// This replaces the old Recovery middleware and handles all errors
	engine.Use(middleware.ErrorHandler(cfg.Logger))

	// 3. ContextCancellation - detect client disconnections early to avoid wasted work
	if cfg.EnableContextCancellation {
		engine.Use(middleware.ContextCancellation())
	}

	// 4. CORS - handle CORS before processing requests
	if cfg.CORS != nil {
		engine.Use(middleware.CORS(*cfg.CORS))
	}

	// 5. Telemetry middleware (traces all HTTP requests)
	if cfg.ServiceName != "" {
		engine.Use(middleware.Telemetry(cfg.ServiceName))
	}

	// 6. Logger - log after all processing
	if cfg.EnableLogger {
		engine.Use(basicLogger(cfg.Logger))
	}

	router := &GinRouter{engine: engine}

	// If BasePath is configured, create a base group
	if cfg.BasePath != "" {
		router.baseGroup = engine.Group(cfg.BasePath)
	}

	return router
}

// Handler returns the underlying http.Handler
func (r *GinRouter) Handler() http.Handler {
	return r.engine
}

// getRouterGroup returns the appropriate gin.IRouter (baseGroup if set, otherwise engine)
func (r *GinRouter) getRouterGroup() gin.IRouter {
	if r.baseGroup != nil {
		return r.baseGroup
	}
	return r.engine
}

// Use adds middleware to the router
func (r *GinRouter) Use(middleware ...gin.HandlerFunc) {
	r.getRouterGroup().Use(middleware...)
}

// GET registers a GET route
func (r *GinRouter) GET(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().GET(relativePath, handlers...)
}

// POST registers a POST route
func (r *GinRouter) POST(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().POST(relativePath, handlers...)
}

// PUT registers a PUT route
func (r *GinRouter) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().PUT(relativePath, handlers...)
}

// DELETE registers a DELETE route
func (r *GinRouter) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().DELETE(relativePath, handlers...)
}

// PATCH registers a PATCH route
func (r *GinRouter) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().PATCH(relativePath, handlers...)
}

// OPTIONS registers an OPTIONS route
func (r *GinRouter) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().OPTIONS(relativePath, handlers...)
}

// HEAD registers a HEAD route
func (r *GinRouter) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().HEAD(relativePath, handlers...)
}

// Group creates a new route group with the given prefix
func (r *GinRouter) Group(relativePath string, handlers ...gin.HandlerFunc) *GinRouterGroup {
	group := r.getRouterGroup().Group(relativePath, handlers...)
	return &GinRouterGroup{group: group}
}

// Use adds middleware to the group
func (g *GinRouterGroup) Use(middleware ...gin.HandlerFunc) {
	g.group.Use(middleware...)
}

// GET registers a GET route in the group
func (g *GinRouterGroup) GET(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.GET(relativePath, handlers...)
}

// POST registers a POST route in the group
func (g *GinRouterGroup) POST(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.POST(relativePath, handlers...)
}

// PUT registers a PUT route in the group
func (g *GinRouterGroup) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.PUT(relativePath, handlers...)
}

// DELETE registers a DELETE route in the group
func (g *GinRouterGroup) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.DELETE(relativePath, handlers...)
}

// PATCH registers a PATCH route in the group
func (g *GinRouterGroup) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.PATCH(relativePath, handlers...)
}

// OPTIONS registers an OPTIONS route in the group
func (g *GinRouterGroup) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.OPTIONS(relativePath, handlers...)
}

// HEAD registers a HEAD route in the group
func (g *GinRouterGroup) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.HEAD(relativePath, handlers...)
}

// Group creates a nested route group
func (g *GinRouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *GinRouterGroup {
	nestedGroup := g.group.Group(relativePath, handlers...)
	return &GinRouterGroup{group: nestedGroup}
}

// Middleware functions

// basicLogger creates a request logger middleware
func basicLogger(logger Logger) gin.HandlerFunc {
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
		fields := map[string]any{
			"path":        path,
			"status":      c.Writer.Status(),
			"duration":    duration.String(),
			"duration_ms": duration.Milliseconds(),
		}

		// Add base request metadata
		loggerutil.AddBaseRequestFields(fields, c)

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
