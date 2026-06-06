package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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
