package middleware

import (
	"github.com/gin-gonic/gin"
)

// ContextCancellation creates a middleware that checks if the request context is cancelled
// before processing the request. This prevents wasted processing when clients disconnect.
//
// This middleware should be registered AFTER ErrorHandler to properly handle cancellation errors.
// If the context is cancelled, it returns 499 (Client Closed Request) immediately.
//
// Usage:
//
//	platform.Use(middleware.ContextCancellation())
//
// Or enable it globally in config (recommended).
func ContextCancellation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if context is already cancelled before processing
		if err := c.Request.Context().Err(); err != nil {
			// Context is cancelled, don't process the request
			_ = c.Error(err)
			c.Abort()
			return
		}

		// Process request
		c.Next()
	}
}
