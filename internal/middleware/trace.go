package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TraceIDHeader is the HTTP header name for trace ID
	TraceIDHeader = "X-Trace-Id"

	// TraceIDKey is the context key for storing trace ID
	TraceIDKey = "trace_id"
)

// TraceID generates or extracts a trace ID for each request
// If the request already has a trace ID in the X-Trace-Id header, it will be used
// Otherwise, a new UUID will be generated
// The trace ID is stored in the gin context and added to the response header
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)

		if traceID == "" {
			traceID = uuid.New().String()
		}

		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDHeader, traceID)

		c.Next()
	}
}

// GetTraceID extracts the trace ID from the gin context
// Returns empty string if no trace ID is found
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}
