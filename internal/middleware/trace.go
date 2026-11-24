package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

const (
	// TraceIDHeader is the HTTP header name for trace ID
	TraceIDHeader = "X-Trace-Id"

	// TraceIDKey is the context key for storing trace ID
	TraceIDKey = "trace_id"
)

// traceIDContextKey is the key under which the trace ID is stored in the request's context.Context.
type traceIDContextKey struct{}

// TraceID generates or extracts a trace ID for each request
// If the request already has a trace ID in the X-Trace-Id header, it will be used
// Otherwise, a new UUID will be generated
// The trace ID is stored in the gin context, in the request's context.Context and added to the response header
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(TraceIDHeader)

		if traceID == "" {
			traceID = uuid.New().String()
		}

		setTraceID(c, traceID)

		c.Next()
	}
}

// OTelTraceID replaces the trace ID with the OpenTelemetry trace ID of the current span.
// It must run after the Telemetry middleware, which starts the span.
// With it, logs carry the same ID as the trace in the tracing backend (e.g. Tempo),
// so logs and traces can be joined. Without an active span, the trace ID set by TraceID is kept.
func OTelTraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if sc := trace.SpanContextFromContext(c.Request.Context()); sc.HasTraceID() {
			setTraceID(c, sc.TraceID().String())
		}

		c.Next()
	}
}

// TraceIDFromContext returns the trace ID stored in ctx by the TraceID or OTelTraceID middleware.
// Outside an HTTP request (e.g. a message consumer), it falls back to the trace ID of the
// OpenTelemetry span in ctx. Returns an empty string if neither is present.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if traceID, ok := ctx.Value(traceIDContextKey{}).(string); ok && traceID != "" {
		return traceID
	}

	if sc := trace.SpanContextFromContext(ctx); sc.HasTraceID() {
		return sc.TraceID().String()
	}

	return ""
}

// setTraceID stores the trace ID in the gin context, the request's context.Context and the response header.
func setTraceID(c *gin.Context, traceID string) {
	c.Set(TraceIDKey, traceID)
	c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), traceIDContextKey{}, traceID))
	c.Header(TraceIDHeader, traceID)
}
