package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTelemetry(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("creates telemetry middleware", func(t *testing.T) {
		middleware := Telemetry("test-service")
		assert.NotNil(t, middleware)
	})

	t.Run("middleware can be registered on router", func(t *testing.T) {
		router := gin.New()
		router.Use(Telemetry("test-service"))

		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})
}
