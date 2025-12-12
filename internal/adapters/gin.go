package adapters

import (
	"net/http"

	"github.com/edaniel30/http-platform-go/middleware"
	config "github.com/edaniel30/http-platform-go/models"
	"github.com/gin-gonic/gin"
)

// GinRouter wraps gin.Engine to implement the Router interface
type GinRouter struct {
	engine    *gin.Engine
	baseGroup *gin.RouterGroup // Optional base group when BasePath is configured
}

// GinRouterGroup wraps gin.RouterGroup to implement the RouterGroup interface
type GinRouterGroup struct {
	group *gin.RouterGroup
}

// NewGinRouter creates a new Gin router with the given configuration
func NewGinRouter(cfg config.Config) *GinRouter {
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
	if cfg.EnableCORS {
		corsMiddleware := middleware.CORS(middleware.CORSConfig{
			AllowedOrigins:   cfg.AllowedOrigins,
			AllowedMethods:   cfg.AllowedMethods,
			AllowedHeaders:   cfg.AllowedHeaders,
			ExposedHeaders:   cfg.ExposedHeaders,
			AllowCredentials: cfg.AllowCredentials,
			MaxAge:           cfg.MaxAge,
		})
		engine.Use(corsMiddleware)
	}

	// 5. Telemetry middleware (traces all HTTP requests)
	if cfg.EnableTelemetry {
		engine.Use(middleware.Telemetry(cfg.ServiceName))
	}

	// 6. Logger - log after all processing
	if cfg.EnableLogger {
		engine.Use(middleware.BasicLogger(cfg.Logger))
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

// Use adds middleware to the router
func (r *GinRouter) Use(middleware ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.Use(middleware...)
	} else {
		r.engine.Use(middleware...)
	}
}

// GET registers a GET route
func (r *GinRouter) GET(relativePath string, handlers ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.GET(relativePath, handlers...)
	} else {
		r.engine.GET(relativePath, handlers...)
	}
}

// POST registers a POST route
func (r *GinRouter) POST(relativePath string, handlers ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.POST(relativePath, handlers...)
	} else {
		r.engine.POST(relativePath, handlers...)
	}
}

// PUT registers a PUT route
func (r *GinRouter) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.PUT(relativePath, handlers...)
	} else {
		r.engine.PUT(relativePath, handlers...)
	}
}

// DELETE registers a DELETE route
func (r *GinRouter) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.DELETE(relativePath, handlers...)
	} else {
		r.engine.DELETE(relativePath, handlers...)
	}
}

// PATCH registers a PATCH route
func (r *GinRouter) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.PATCH(relativePath, handlers...)
	} else {
		r.engine.PATCH(relativePath, handlers...)
	}
}

// OPTIONS registers an OPTIONS route
func (r *GinRouter) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.OPTIONS(relativePath, handlers...)
	} else {
		r.engine.OPTIONS(relativePath, handlers...)
	}
}

// HEAD registers a HEAD route
func (r *GinRouter) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	if r.baseGroup != nil {
		r.baseGroup.HEAD(relativePath, handlers...)
	} else {
		r.engine.HEAD(relativePath, handlers...)
	}
}

// Group creates a new route group with the given prefix
func (r *GinRouter) Group(relativePath string, handlers ...gin.HandlerFunc) *GinRouterGroup {
	var group *gin.RouterGroup
	if r.baseGroup != nil {
		group = r.baseGroup.Group(relativePath, handlers...)
	} else {
		group = r.engine.Group(relativePath, handlers...)
	}
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
