package middleware

import (
	"github.com/edaniel30/loki-logger-go"
	"github.com/edaniel30/loki-logger-go/middleware"
	"github.com/gin-gonic/gin"
)

// Logger creates a request logger middleware using loki logger
// This middleware logs all incoming HTTP requests with method, path, status, and duration
func Logger(logger *loki.Logger) gin.HandlerFunc {
	return middleware.GinLogger(logger)
}
