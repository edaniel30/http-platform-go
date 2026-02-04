package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/edaniel30/http-platform-go/internal/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestContextCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("continues when context is not cancelled", func(t *testing.T) {
		router := gin.New()
		router.Use(ContextCancellation())

		handlerCalled := false
		router.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.True(t, handlerCalled, "handler should be called when context is not cancelled")
		assert.Equal(t, 200, w.Code)
	})

	t.Run("aborts when context is already cancelled", func(t *testing.T) {
		router := gin.New()
		logger := logger.NewDefaultLogger()
		router.Use(ErrorHandler(logger))
		router.Use(ContextCancellation())

		handlerCalled := false
		router.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.String(200, "OK")
		})

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		req = req.WithContext(ctx)
		router.ServeHTTP(w, req)

		assert.False(t, handlerCalled, "handler should not be called when context is cancelled")
		assert.Equal(t, 499, w.Code) // Client closed request
	})
}
