package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

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

func TestIsContextCancelled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns false for active context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		assert.False(t, IsContextCancelled(c))
	})

	t.Run("returns true for cancelled context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request = c.Request.WithContext(ctx)

		assert.True(t, IsContextCancelled(c))
	})

	t.Run("returns true for deadline exceeded", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond) // Ensure timeout

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request = c.Request.WithContext(ctx)

		assert.True(t, IsContextCancelled(c))
	})
}

func TestGetContextError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns nil for active context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		err := GetContextError(c)
		assert.Nil(t, err)
	})

	t.Run("returns context.Canceled for cancelled context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request = c.Request.WithContext(ctx)

		err := GetContextError(c)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("returns context.DeadlineExceeded for timeout", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(2 * time.Millisecond)

		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request = c.Request.WithContext(ctx)

		err := GetContextError(c)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestWithTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("completes successfully within timeout", func(t *testing.T) {
		router := gin.New()

		handlerCalled := false
		router.GET("/test", WithTimeout(100*time.Millisecond), func(c *gin.Context) {
			handlerCalled = true
			time.Sleep(10 * time.Millisecond) // Quick operation
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.True(t, handlerCalled)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("aborts on timeout", func(t *testing.T) {
		router := gin.New()
		logger := logger.NewDefaultLogger()
		router.Use(ErrorHandler(logger))

		router.GET("/test", WithTimeout(10*time.Millisecond), func(c *gin.Context) {
			time.Sleep(50 * time.Millisecond) // Slow operation
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		// Should return 408 Request Timeout
		assert.Equal(t, 408, w.Code)
	})

	t.Run("context deadline is propagated to handler", func(t *testing.T) {
		router := gin.New()

		var contextHadDeadline bool
		router.GET("/test", WithTimeout(50*time.Millisecond), func(c *gin.Context) {
			_, contextHadDeadline = c.Request.Context().Deadline()
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.True(t, contextHadDeadline, "handler should receive context with deadline")
	})
}
