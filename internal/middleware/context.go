package middleware

import (
	"context"
	"time"

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
			c.Error(err)
			c.Abort()
			return
		}

		// Process request
		c.Next()
	}
}

// IsContextCancelled checks if the request context has been cancelled (client disconnected)
// Use this in long-running handlers to avoid unnecessary work.
//
// Example:
//
//	func (h *Handler) ProcessData(c *gin.Context) {
//	    for i, item := range largeDataset {
//	        // Check if client disconnected every 100 items
//	        if i % 100 == 0 && middleware.IsContextCancelled(c) {
//	            c.Error(context.Canceled)
//	            return
//	        }
//	        // Process item...
//	    }
//	}
func IsContextCancelled(c *gin.Context) bool {
	return c.Request.Context().Err() != nil
}

// GetContextError returns the context error if any (context.Canceled or context.DeadlineExceeded)
// Returns nil if context is still valid.
//
// Example:
//
//	if err := middleware.GetContextError(c); err != nil {
//	    c.Error(err)
//	    return
//	}
func GetContextError(c *gin.Context) error {
	return c.Request.Context().Err()
}

// WithTimeout wraps a handler with a timeout using context.WithTimeout
// If the handler doesn't complete within the timeout, it returns 408 Request Timeout.
//
// Example:
//
//	// Set 5 second timeout for this specific endpoint
//	router.GET("/slow-endpoint", middleware.WithTimeout(5*time.Second), handler)
//
// Note: This is useful for specific endpoints that need stricter timeouts than the global server timeout.
func WithTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal completion
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Handler completed successfully
			return
		case <-ctx.Done():
			// Timeout exceeded
			c.Error(ctx.Err())
			c.Abort()
			return
		}
	}
}
