package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Telemetry returns a middleware that traces HTTP requests using OpenTelemetry
// serviceName should match the service name configured in telemetry initialization
func Telemetry(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
