package httpplatform

import (
	"net/http"

	"github.com/edaniel30/http-platform-go/middleware"
	"github.com/gin-gonic/gin"
)

// ginRouter wraps gin.Engine to provide HTTP routing capabilities.
// This is an internal implementation that shouldn't be exported.
type ginRouter struct {
	engine    *gin.Engine
	baseGroup *gin.RouterGroup // Optional base group when BasePath is configured
}

// ginRouterGroup wraps gin.RouterGroup to provide route grouping capabilities.
// This is an internal implementation that shouldn't be exported.
type ginRouterGroup struct {
	group *gin.RouterGroup
}

// newGinRouter creates a new Gin router with the given configuration.
// It initializes the Gin engine, applies middleware in the correct order, and sets up the base path if configured.
func newGinRouter(cfg Config) *ginRouter {
	// Set gin mode
	gin.SetMode(cfg.Mode)

	// Create engine
	engine := gin.New()

	// Set trusted proxies
	if cfg.TrustedProxies != nil {
		engine.SetTrustedProxies(cfg.TrustedProxies)
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
	if cfg.Telemetry != nil {
		engine.Use(middleware.Telemetry(cfg.Telemetry.ServiceName))
	}

	// 6. Logger - log after all processing
	if cfg.EnableLogger {
		engine.Use(middleware.BasicLogger(cfg.Logger))
	}

	router := &ginRouter{engine: engine}

	// If BasePath is configured, create a base group
	if cfg.BasePath != "" {
		router.baseGroup = engine.Group(cfg.BasePath)
	}

	return router
}

// Handler returns the underlying http.Handler
func (r *ginRouter) Handler() http.Handler {
	return r.engine
}

// getRouterGroup returns the appropriate gin.IRouter (baseGroup if set, otherwise engine)
func (r *ginRouter) getRouterGroup() gin.IRouter {
	if r.baseGroup != nil {
		return r.baseGroup
	}
	return r.engine
}

// Use adds middleware to the router
func (r *ginRouter) Use(middleware ...gin.HandlerFunc) {
	r.getRouterGroup().Use(middleware...)
}

// GET registers a GET route
func (r *ginRouter) GET(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().GET(relativePath, handlers...)
}

// POST registers a POST route
func (r *ginRouter) POST(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().POST(relativePath, handlers...)
}

// PUT registers a PUT route
func (r *ginRouter) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().PUT(relativePath, handlers...)
}

// DELETE registers a DELETE route
func (r *ginRouter) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().DELETE(relativePath, handlers...)
}

// PATCH registers a PATCH route
func (r *ginRouter) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().PATCH(relativePath, handlers...)
}

// OPTIONS registers an OPTIONS route
func (r *ginRouter) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().OPTIONS(relativePath, handlers...)
}

// HEAD registers a HEAD route
func (r *ginRouter) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	r.getRouterGroup().HEAD(relativePath, handlers...)
}

// Group creates a new route group with the given prefix
func (r *ginRouter) Group(relativePath string, handlers ...gin.HandlerFunc) *ginRouterGroup {
	group := r.getRouterGroup().Group(relativePath, handlers...)
	return &ginRouterGroup{group: group}
}

// Use adds middleware to the group
func (g *ginRouterGroup) Use(middleware ...gin.HandlerFunc) {
	g.group.Use(middleware...)
}

// GET registers a GET route in the group
func (g *ginRouterGroup) GET(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.GET(relativePath, handlers...)
}

// POST registers a POST route in the group
func (g *ginRouterGroup) POST(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.POST(relativePath, handlers...)
}

// PUT registers a PUT route in the group
func (g *ginRouterGroup) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.PUT(relativePath, handlers...)
}

// DELETE registers a DELETE route in the group
func (g *ginRouterGroup) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.DELETE(relativePath, handlers...)
}

// PATCH registers a PATCH route in the group
func (g *ginRouterGroup) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.PATCH(relativePath, handlers...)
}

// OPTIONS registers an OPTIONS route in the group
func (g *ginRouterGroup) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.OPTIONS(relativePath, handlers...)
}

// HEAD registers a HEAD route in the group
func (g *ginRouterGroup) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	g.group.HEAD(relativePath, handlers...)
}

// Group creates a nested route group
func (g *ginRouterGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *ginRouterGroup {
	nestedGroup := g.group.Group(relativePath, handlers...)
	return &ginRouterGroup{group: nestedGroup}
}
