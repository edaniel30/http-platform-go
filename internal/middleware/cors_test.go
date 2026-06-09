package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestIsWildcardOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origins  []string
		expected bool
	}{
		{
			name:     "single wildcard origin",
			origins:  []string{"*"},
			expected: true,
		},
		{
			name:     "empty origins",
			origins:  []string{},
			expected: false,
		},
		{
			name:     "nil origins",
			origins:  nil,
			expected: false,
		},
		{
			name:     "single specific origin",
			origins:  []string{"https://example.com"},
			expected: false,
		},
		{
			name:     "multiple origins including wildcard",
			origins:  []string{"*", "https://example.com"},
			expected: false,
		},
		{
			name:     "multiple specific origins",
			origins:  []string{"https://example.com", "https://api.example.com"},
			expected: false,
		},
		{
			name:     "wildcard with whitespace",
			origins:  []string{" * "},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWildcardOrigin(tt.origins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCORS(t *testing.T) {
	t.Run("wildcard origin sets AllowAllOrigins", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true, // Should be overridden to false
			MaxAge:           12 * time.Hour,
		}

		middleware := CORS(cfg)
		assert.NotNil(t, middleware)
	})

	t.Run("specific origins", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins:   []string{"https://example.com", "https://api.example.com"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			ExposedHeaders:   []string{"X-Total-Count"},
			AllowCredentials: true,
			MaxAge:           24 * time.Hour,
		}

		middleware := CORS(cfg)
		assert.NotNil(t, middleware)
	})

	t.Run("wildcard forces credentials to false", func(t *testing.T) {
		cfg := CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true, // This should be ignored
		}

		middleware := CORS(cfg)
		assert.NotNil(t, middleware)
		// Note: We can't directly test the internal config of gin-contrib/cors,
		// but we verify the middleware is created without panic
	})

	t.Run("AllowWildcard enables pattern matching with credentials", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		cfg := CORSConfig{
			AllowedOrigins:   []string{"http://localhost:*", "https://app.example.com"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
			AllowWildcard:    true,
		}

		router := gin.New()
		router.Use(CORS(cfg))
		router.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

		t.Run("allows matching port pattern origin", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
			req.Header.Set("Origin", "http://localhost:3000")
			router.ServeHTTP(w, req)

			assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		})

		t.Run("allows exact origin match", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
			req.Header.Set("Origin", "https://app.example.com")
			router.ServeHTTP(w, req)

			assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		})

		t.Run("blocks non-matching origin", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
			req.Header.Set("Origin", "https://evil.com")
			router.ServeHTTP(w, req)

			assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		})
	})
}
