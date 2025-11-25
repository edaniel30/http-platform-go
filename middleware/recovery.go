package middleware

import (
	"github.com/edaniel30/loki-logger-go"
	"github.com/edaniel30/loki-logger-go/middleware"
	"github.com/gin-gonic/gin"
)

// Recovery creates a recovery middleware that logs panics using loki logger
// This middleware recovers from any panics and logs the error with a stack trace
func Recovery(logger *loki.Logger) gin.HandlerFunc {
	return middleware.GinRecovery(logger)
}
