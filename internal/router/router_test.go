package router

import (
	"net/http/httptest"
	"testing"

	"github.com/edaniel30/http-platform-go/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewGinRouter(t *testing.T) {
	t.Run("basic router creation", func(t *testing.T) {
		cfg := RouterConfig{
			Mode:                      gin.TestMode,
			EnableTraceID:             true,
			EnableLogger:              true,
			EnableContextCancellation: true,
			Logger:                    middleware.NewDefaultLogger(),
		}

		router := NewGinRouter(cfg)

		assert.NotNil(t, router)
		assert.NotNil(t, router.engine)
		assert.Nil(t, router.baseGroup)

		// Register a route and test it works
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "success")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.engine.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "success", w.Body.String())
	})
}

func TestGinRouterHTTPMethods(t *testing.T) {
	cfg := RouterConfig{
		Mode:   gin.TestMode,
		Logger: middleware.NewDefaultLogger(),
	}
	router := NewGinRouter(cfg)

	handler := func(c *gin.Context) {
		c.String(200, "OK")
	}

	t.Run("GET", func(t *testing.T) {
		assert.NotPanics(t, func() {
			router.GET("/test", handler)
		})
	})

	t.Run("POST", func(t *testing.T) {
		assert.NotPanics(t, func() {
			router.POST("/test", handler)
		})
	})

	t.Run("PUT", func(t *testing.T) {
		assert.NotPanics(t, func() {
			router.PUT("/test", handler)
		})
	})

	t.Run("DELETE", func(t *testing.T) {
		assert.NotPanics(t, func() {
			router.DELETE("/test", handler)
		})
	})

	t.Run("PATCH", func(t *testing.T) {
		assert.NotPanics(t, func() {
			router.PATCH("/test", handler)
		})
	})

	t.Run("OPTIONS", func(t *testing.T) {
		assert.NotPanics(t, func() {
			router.OPTIONS("/test", handler)
		})
	})

	t.Run("HEAD", func(t *testing.T) {
		assert.NotPanics(t, func() {
			router.HEAD("/test", handler)
		})
	})
}

func TestGinRouterHTTPMethodsWithBasePath(t *testing.T) {
	cfg := RouterConfig{
		Mode:     gin.TestMode,
		BasePath: "/api/v1",
		Logger:   middleware.NewDefaultLogger(),
	}
	router := NewGinRouter(cfg)

	// Register a test route
	router.GET("/users", func(c *gin.Context) {
		c.String(200, "users list")
	})

	// Create a test request
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/users", nil)

	router.engine.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "users list", w.Body.String())
}

func TestGinRouterGroup(t *testing.T) {
	cfg := RouterConfig{
		Mode:   gin.TestMode,
		Logger: middleware.NewDefaultLogger(),
	}
	router := NewGinRouter(cfg)

	t.Run("create group", func(t *testing.T) {
		group := router.Group("/api")
		assert.NotNil(t, group)
		assert.NotNil(t, group.group)
	})

	t.Run("register routes in group", func(t *testing.T) {
		group := router.Group("/users")

		assert.NotPanics(t, func() {
			group.GET("/:id", func(c *gin.Context) {})
			group.POST("", func(c *gin.Context) {})
			group.PUT("/:id", func(c *gin.Context) {})
			group.DELETE("/:id", func(c *gin.Context) {})
			group.PATCH("/:id", func(c *gin.Context) {})
			group.OPTIONS("/:id", func(c *gin.Context) {})
			group.HEAD("/:id", func(c *gin.Context) {})
		})
	})

	t.Run("nested groups", func(t *testing.T) {
		api := router.Group("/api")
		v1 := api.Group("/v1")
		users := v1.Group("/users")

		assert.NotNil(t, api)
		assert.NotNil(t, v1)
		assert.NotNil(t, users)

		assert.NotPanics(t, func() {
			users.GET("", func(c *gin.Context) {})
		})
	})
}

func TestGinRouterGroupWithBasePath(t *testing.T) {
	cfg := RouterConfig{
		Mode:     gin.TestMode,
		BasePath: "/api",
		Logger:   middleware.NewDefaultLogger(),
	}
	router := NewGinRouter(cfg)

	// Create group under base path
	users := router.Group("/users")
	users.GET("", func(c *gin.Context) {
		c.String(200, "users")
	})

	// Test that route is under /api/users
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/users", nil)

	router.engine.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "users", w.Body.String())
}

func TestGinRouterGroupMiddleware(t *testing.T) {
	cfg := RouterConfig{
		Mode:   gin.TestMode,
		Logger: middleware.NewDefaultLogger(),
	}
	router := NewGinRouter(cfg)

	middlewareCalled := false

	group := router.Group("/api")
	group.Use(func(c *gin.Context) {
		middlewareCalled = true
		c.Next()
	})

	group.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)

	router.engine.ServeHTTP(w, req)

	assert.True(t, middlewareCalled, "group middleware should have been called")
	assert.Equal(t, 200, w.Code)
}

func TestGinRouterMiddleware(t *testing.T) {
	cfg := RouterConfig{
		Mode:   gin.TestMode,
		Logger: middleware.NewDefaultLogger(),
	}
	router := NewGinRouter(cfg)

	middlewareCalled := false

	router.Use(func(c *gin.Context) {
		middlewareCalled = true
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.String(200, "OK")
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)

	router.engine.ServeHTTP(w, req)

	assert.True(t, middlewareCalled, "router middleware should have been called")
	assert.Equal(t, 200, w.Code)
}

func TestGinRouterGroupHTTPMethods(t *testing.T) {
	cfg := RouterConfig{
		Mode:   gin.TestMode,
		Logger: middleware.NewDefaultLogger(),
	}
	router := NewGinRouter(cfg)
	group := router.Group("/api")

	handler := func(c *gin.Context) {
		c.String(200, "OK")
	}

	tests := []struct {
		name   string
		method string
		setup  func()
	}{
		{
			name:   "GET",
			method: "GET",
			setup:  func() { group.GET("/get", handler) },
		},
		{
			name:   "POST",
			method: "POST",
			setup:  func() { group.POST("/post", handler) },
		},
		{
			name:   "PUT",
			method: "PUT",
			setup:  func() { group.PUT("/put", handler) },
		},
		{
			name:   "DELETE",
			method: "DELETE",
			setup:  func() { group.DELETE("/delete", handler) },
		},
		{
			name:   "PATCH",
			method: "PATCH",
			setup:  func() { group.PATCH("/patch", handler) },
		},
		{
			name:   "OPTIONS",
			method: "OPTIONS",
			setup:  func() { group.OPTIONS("/options", handler) },
		},
		{
			name:   "HEAD",
			method: "HEAD",
			setup:  func() { group.HEAD("/head", handler) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, tt.setup)
		})
	}
}
