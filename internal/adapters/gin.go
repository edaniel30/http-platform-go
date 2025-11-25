package adapters

import (
	"net/http"

	"github.com/edaniel30/http-platform-go/middleware"
	config "github.com/edaniel30/http-platform-go/models"
	"github.com/gin-gonic/gin"
)

// GinRouter wraps gin.Engine to implement the Router interface
type GinRouter struct {
	engine *gin.Engine
}

// GinRouterGroup wraps gin.RouterGroup to implement the RouterGroup interface
type GinRouterGroup struct {
	group *gin.RouterGroup
}

// NewGinRouter creates a new Gin router with the given configuration
func NewGinRouter(cfg config.Config) *GinRouter {
	// Set gin mode
	gin.SetMode(cfg.Mode)

	// Create engine with default middleware or without based on config
	var engine *gin.Engine
	if cfg.EnableRecovery || cfg.EnableLogger {
		engine = gin.New() // Start clean, we'll add our own middleware
	} else {
		engine = gin.New()
	}

	// Set trusted proxies
	if cfg.TrustedProxies != nil {
		engine.SetTrustedProxies(cfg.TrustedProxies)
	}

	router := &GinRouter{engine: engine}

	// Apply middleware in order: TraceID -> CORS -> Recovery -> Logger
	if cfg.EnableTraceID {
		router.Use(middleware.TraceID())
	}

	if cfg.EnableCORS {
		corsMiddleware := middleware.CORS(middleware.CORSConfig{
			AllowedOrigins:   cfg.AllowedOrigins,
			AllowedMethods:   cfg.AllowedMethods,
			AllowedHeaders:   cfg.AllowedHeaders,
			ExposedHeaders:   cfg.ExposedHeaders,
			AllowCredentials: cfg.AllowCredentials,
			MaxAge:           cfg.MaxAge,
		})
		router.Use(corsMiddleware)
	}

	if cfg.EnableRecovery {
		router.Use(middleware.Recovery(cfg.Logger))
	}

	if cfg.EnableLogger {
		router.Use(middleware.Logger(cfg.Logger))
	}

	return router
}

// Handler returns the underlying http.Handler
func (r *GinRouter) Handler() http.Handler {
	return r.engine
}

// Use adds middleware to the router
func (r *GinRouter) Use(middleware ...gin.HandlerFunc) {
	r.engine.Use(middleware...)
}

// GET registers a GET route
func (r *GinRouter) GET(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.GET(relativePath, handlers...)
}

// POST registers a POST route
func (r *GinRouter) POST(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.POST(relativePath, handlers...)
}

// PUT registers a PUT route
func (r *GinRouter) PUT(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.PUT(relativePath, handlers...)
}

// DELETE registers a DELETE route
func (r *GinRouter) DELETE(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.DELETE(relativePath, handlers...)
}

// PATCH registers a PATCH route
func (r *GinRouter) PATCH(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.PATCH(relativePath, handlers...)
}

// OPTIONS registers an OPTIONS route
func (r *GinRouter) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.OPTIONS(relativePath, handlers...)
}

// HEAD registers a HEAD route
func (r *GinRouter) HEAD(relativePath string, handlers ...gin.HandlerFunc) {
	r.engine.HEAD(relativePath, handlers...)
}

// Group creates a new route group with the given prefix
func (r *GinRouter) Group(relativePath string, handlers ...gin.HandlerFunc) *GinRouterGroup {
	group := r.engine.Group(relativePath, handlers...)
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
