package httpplatform

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandlerFunc is the standard HTTP handler function signature
// It wraps http.HandlerFunc for use with Platform
type HandlerFunc func(c *gin.Context)

// MiddlewareFunc is the signature for HTTP middleware
// Middleware can be added to the platform to intercept and process requests
type MiddlewareFunc gin.HandlerFunc

// Router defines the HTTP routing capabilities
// This interface allows for different router implementations while maintaining a consistent API
type Router interface {
	// Handler returns the underlying http.Handler
	Handler() http.Handler

	// Use adds middleware to the router
	Use(middleware ...MiddlewareFunc)

	// GET registers a GET route
	GET(relativePath string, handlers ...HandlerFunc)

	// POST registers a POST route
	POST(relativePath string, handlers ...HandlerFunc)

	// PUT registers a PUT route
	PUT(relativePath string, handlers ...HandlerFunc)

	// DELETE registers a DELETE route
	DELETE(relativePath string, handlers ...HandlerFunc)

	// PATCH registers a PATCH route
	PATCH(relativePath string, handlers ...HandlerFunc)

	// OPTIONS registers an OPTIONS route
	OPTIONS(relativePath string, handlers ...HandlerFunc)

	// HEAD registers a HEAD route
	HEAD(relativePath string, handlers ...HandlerFunc)

	// Group creates a new route group with the given prefix
	Group(relativePath string, handlers ...HandlerFunc) RouterGroup
}

// RouterGroup defines route grouping capabilities
// Groups allow organizing related routes under a common prefix
type RouterGroup interface {
	// Use adds middleware to the group
	Use(middleware ...MiddlewareFunc)

	// GET registers a GET route in the group
	GET(relativePath string, handlers ...HandlerFunc)

	// POST registers a POST route in the group
	POST(relativePath string, handlers ...HandlerFunc)

	// PUT registers a PUT route in the group
	PUT(relativePath string, handlers ...HandlerFunc)

	// DELETE registers a DELETE route in the group
	DELETE(relativePath string, handlers ...HandlerFunc)

	// PATCH registers a PATCH route in the group
	PATCH(relativePath string, handlers ...HandlerFunc)

	// OPTIONS registers an OPTIONS route in the group
	OPTIONS(relativePath string, handlers ...HandlerFunc)

	// HEAD registers a HEAD route in the group
	HEAD(relativePath string, handlers ...HandlerFunc)

	// Group creates a nested route group
	Group(relativePath string, handlers ...HandlerFunc) RouterGroup
}
