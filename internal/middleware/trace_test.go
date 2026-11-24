package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func TestTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates new trace ID when not provided", func(t *testing.T) {
		router := gin.New()
		router.Use(TraceID())

		var capturedTraceID string
		router.GET("/test", func(c *gin.Context) {
			if traceID, exists := c.Get(TraceIDKey); exists {
				if id, ok := traceID.(string); ok {
					capturedTraceID = id
				}
			}
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.NotEmpty(t, capturedTraceID)

		_, err := uuid.Parse(capturedTraceID)
		assert.NoError(t, err)

		assert.Equal(t, capturedTraceID, w.Header().Get(TraceIDHeader))
	})

	t.Run("uses existing UUID trace ID from request header", func(t *testing.T) {
		router := gin.New()
		router.Use(TraceID())

		existingTraceID := uuid.New().String()
		var capturedTraceID string

		router.GET("/test", func(c *gin.Context) {
			if traceID, exists := c.Get(TraceIDKey); exists {
				if id, ok := traceID.(string); ok {
					capturedTraceID = id
				}
			}
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
		req.Header.Set(TraceIDHeader, existingTraceID)
		router.ServeHTTP(w, req)

		assert.Equal(t, existingTraceID, capturedTraceID)
		assert.Equal(t, existingTraceID, w.Header().Get(TraceIDHeader))
	})
}

func TestTraceIDStoresInRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceID())

	var fromGin, fromCtx string
	router.GET("/test", func(c *gin.Context) {
		fromGin = c.GetString(TraceIDKey)
		fromCtx = TraceIDFromContext(c.Request.Context())
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.NotEmpty(t, fromCtx)
	assert.Equal(t, fromGin, fromCtx)
}

func TestOTelTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("replaces the trace ID with the OpenTelemetry one", func(t *testing.T) {
		tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
		defer func() { _ = tp.Shutdown(context.Background()) }()

		router := gin.New()
		router.Use(TraceID())
		router.Use(otelgin.Middleware("test-service", otelgin.WithTracerProvider(tp)))
		router.Use(OTelTraceID())

		var spanTraceID, fromGin, fromCtx string
		router.GET("/test", func(c *gin.Context) {
			spanTraceID = trace.SpanContextFromContext(c.Request.Context()).TraceID().String()
			fromGin = c.GetString(TraceIDKey)
			fromCtx = TraceIDFromContext(c.Request.Context())
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Len(t, spanTraceID, 32)
		assert.Equal(t, spanTraceID, fromGin)
		assert.Equal(t, spanTraceID, fromCtx)
		assert.Equal(t, spanTraceID, w.Header().Get(TraceIDHeader))
	})

	t.Run("keeps the trace ID when there is no span", func(t *testing.T) {
		router := gin.New()
		router.Use(TraceID())
		router.Use(OTelTraceID())

		existingTraceID := uuid.New().String()
		var fromGin, fromCtx string
		router.GET("/test", func(c *gin.Context) {
			fromGin = c.GetString(TraceIDKey)
			fromCtx = TraceIDFromContext(c.Request.Context())
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), "GET", "/test", nil)
		req.Header.Set(TraceIDHeader, existingTraceID)
		router.ServeHTTP(w, req)

		assert.Equal(t, existingTraceID, fromGin)
		assert.Equal(t, existingTraceID, fromCtx)
		assert.Equal(t, existingTraceID, w.Header().Get(TraceIDHeader))
	})
}

func TestTraceIDFromContext(t *testing.T) {
	t.Run("empty without trace ID or span", func(t *testing.T) {
		assert.Empty(t, TraceIDFromContext(context.Background()))
	})

	t.Run("empty for a nil context", func(t *testing.T) {
		//nolint:staticcheck // nil context is checked on purpose
		assert.Empty(t, TraceIDFromContext(nil))
	})

	t.Run("falls back to the span's trace ID", func(t *testing.T) {
		traceID, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
		assert.NoError(t, err)
		spanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
		assert.NoError(t, err)

		sc := trace.NewSpanContext(trace.SpanContextConfig{TraceID: traceID, SpanID: spanID})
		ctx := trace.ContextWithSpanContext(context.Background(), sc)

		assert.Equal(t, "4bf92f3577b34da6a3ce929d0e0e4736", TraceIDFromContext(ctx))
	})
}
