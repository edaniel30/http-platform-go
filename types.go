package httpplatform

import (
	"net/http"

	config "github.com/edaniel30/http-platform-go/models"
	"github.com/gin-gonic/gin"
)

// Config type from config package
// This allows users to use httpplatform.Config instead of importing the config package separately
type Config = config.Config

// Option type
type Option = config.Option

// Gin Framework Types
// These types are exported to avoid direct gin-gonic/gin imports in consuming applications

// Context wraps gin.Context to provide HTTP request/response handling capabilities.
// It includes methods for binding request data, setting response status, headers, and more.
// Use this type in handlers instead of directly importing gin.Context from gin-gonic/gin.
type Context = gin.Context

// HandlerFunc defines the handler function signature for HTTP endpoints.
// Use this type when creating middleware or handler functions instead of gin.HandlerFunc.
type HandlerFunc = gin.HandlerFunc

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
