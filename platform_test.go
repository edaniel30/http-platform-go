package httpplatform

import (
	"testing"

	"github.com/edaniel30/http-platform-go/internal/logger"
	"github.com/edaniel30/http-platform-go/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("successful creation with default config", func(t *testing.T) {
		cfg := DefaultConfig()
		platform, err := New(cfg)

		require.NoError(t, err)
		assert.NotNil(t, platform)
		assert.NotNil(t, platform.router)
		assert.Equal(t, cfg.Port, platform.config.Port)
	})

	t.Run("successful creation with options", func(t *testing.T) {
		platform, err := New(
			DefaultConfig(),
			WithPort(3000),
			WithMode("release"),
			WithBasePath("/api/v1"),
		)

		require.NoError(t, err)
		assert.NotNil(t, platform)
		assert.Equal(t, 3000, platform.config.Port)
		assert.Equal(t, "release", platform.config.Mode)
		assert.Equal(t, "/api/v1", platform.config.BasePath)
	})

	t.Run("successful creation with custom logger", func(t *testing.T) {
		customLogger := logger.NewDefaultLogger()

		platform, err := New(
			DefaultConfig(),
			WithLogger(customLogger),
		)

		require.NoError(t, err)
		assert.NotNil(t, platform)
		assert.Equal(t, customLogger, platform.config.Logger)
	})

	t.Run("successful creation with CORS", func(t *testing.T) {
		platform, err := New(
			DefaultConfig(),
			WithCORS(&middleware.CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "POST"},
			}),
		)

		require.NoError(t, err)
		assert.NotNil(t, platform)
		assert.NotNil(t, platform.config.CORS)
	})

	t.Run("successful creation with telemetry", func(t *testing.T) {
		platform, err := New(
			DefaultConfig(),
			WithTelemetry(&TelemetryConfig{
				ServiceName:  "test-service",
				Version:      "1.0.0",
				Environment:  "dev",
				OTLPEndpoint: "localhost:4318",
			}),
		)

		require.NoError(t, err)
		assert.NotNil(t, platform)
		assert.NotNil(t, platform.config.Telemetry)
		assert.Equal(t, "test-service", platform.config.Telemetry.ServiceName)
	})

	t.Run("successful creation with all features disabled", func(t *testing.T) {
		platform, err := New(
			DefaultConfig(),
			WithoutTraceID(),
			WithoutLogger(),
			WithoutContextCancellation(),
		)

		require.NoError(t, err)
		assert.NotNil(t, platform)
		assert.False(t, platform.config.EnableTraceID)
		assert.False(t, platform.config.EnableLogger)
		assert.False(t, platform.config.EnableContextCancellation)
	})

	t.Run("ensureCORS creates empty CORS when nil", func(t *testing.T) {
		cfg := DefaultConfig()
		cfg.CORS = nil
		cfg.ensureCORS()

		// ensureCORS just ensures CORS struct exists, doesn't populate it
		assert.NotNil(t, cfg.CORS)
	})

	t.Run("ensureCORS preserves existing CORS", func(t *testing.T) {
		cfg := DefaultConfig()
		originalCORS := &middleware.CORSConfig{
			AllowedOrigins: []string{"https://custom.com"},
			AllowedMethods: []string{"GET"},
		}
		cfg.CORS = originalCORS
		cfg.ensureCORS()

		assert.Equal(t, originalCORS, cfg.CORS)
		assert.Equal(t, []string{"https://custom.com"}, cfg.CORS.AllowedOrigins)
	})
}

func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		errorMsg  string
	}{
		{
			name: "invalid port",
			config: Config{
				Port:         0,
				Mode:         "debug",
				Logger:       logger.NewDefaultLogger(),
				ReadTimeout:  30,
				WriteTimeout: 30,
				IdleTimeout:  60,
			},
			errorMsg: "Port must be between 1 and 65535",
		},
		{
			name: "invalid mode",
			config: Config{
				Port:         8080,
				Mode:         "invalid",
				Logger:       logger.NewDefaultLogger(),
				ReadTimeout:  30,
				WriteTimeout: 30,
				IdleTimeout:  60,
			},
			errorMsg: "Mode must be 'debug', 'release', or 'test'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform, err := New(tt.config)

			assert.Error(t, err)
			assert.Nil(t, platform)
			assert.Contains(t, err.Error(), tt.errorMsg)

			// Verify it's a ConfigError
			var configErr *ConfigError
			assert.ErrorAs(t, err, &configErr)
		})
	}
}

func TestPlatformHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	platform, err := New(DefaultConfig())
	require.NoError(t, err)

	// Test that methods can be called without panicking
	assert.NotPanics(t, func() {
		platform.GET("/test", func(c *gin.Context) {})
		platform.POST("/test", func(c *gin.Context) {})
		platform.PUT("/test", func(c *gin.Context) {})
		platform.DELETE("/test", func(c *gin.Context) {})
		platform.PATCH("/test", func(c *gin.Context) {})
		platform.OPTIONS("/test", func(c *gin.Context) {})
		platform.HEAD("/test", func(c *gin.Context) {})
	})
}

func TestPlatformGroup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	platform, err := New(DefaultConfig())
	require.NoError(t, err)

	t.Run("create group", func(t *testing.T) {
		group := platform.Group("/api")
		assert.NotNil(t, group)
	})

	t.Run("register routes in group", func(t *testing.T) {
		group := platform.Group("/users")

		assert.NotPanics(t, func() {
			group.GET("/:id", func(c *gin.Context) {})
			group.POST("", func(c *gin.Context) {})
			group.PUT("/:id", func(c *gin.Context) {})
			group.DELETE("/:id", func(c *gin.Context) {})
		})
	})

	t.Run("nested groups", func(t *testing.T) {
		api := platform.Group("/api")
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

func TestPlatformUse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("adds middleware to router", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		testMiddleware := func(c *gin.Context) {
			c.Next()
		}

		// Verify Use() doesn't panic
		assert.NotPanics(t, func() {
			platform.Use(testMiddleware)
		})

		// Verify we can add a route after middleware
		assert.NotPanics(t, func() {
			platform.GET("/test", func(c *gin.Context) {
				c.String(200, "OK")
			})
		})
	})

	t.Run("adds multiple middlewares", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		middleware1 := func(c *gin.Context) { c.Next() }
		middleware2 := func(c *gin.Context) { c.Next() }
		middleware3 := func(c *gin.Context) { c.Next() }

		assert.NotPanics(t, func() {
			platform.Use(middleware1)
			platform.Use(middleware2)
			platform.Use(middleware3)
		})
	})
}

func TestPlatformRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns underlying router", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		router := platform.Router()
		assert.NotNil(t, router)
	})

	t.Run("returned router can be used directly", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		router := platform.Router()

		// Should be able to call router methods directly
		assert.NotPanics(t, func() {
			router.GET("/direct", func(c *gin.Context) {
				c.String(200, "direct route")
			})
		})
	})

	t.Run("changes to returned router affect platform", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		// Get router and add route directly
		router := platform.Router()
		router.GET("/via-router", func(c *gin.Context) {
			c.String(200, "via router")
		})

		// Verify it's the same router instance
		assert.Equal(t, router, platform.Router())
	})

	t.Run("router has handler", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		router := platform.Router()
		handler := router.Handler()
		assert.NotNil(t, handler)
	})
}

func TestPlatformValidateNotStarted(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns nil when server not started", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		// Should not error since server hasn't been started
		err = platform.validateNotStarted()
		assert.NoError(t, err)
	})

	t.Run("can add routes when not started", func(t *testing.T) {
		platform, err := New(DefaultConfig())
		require.NoError(t, err)

		// These should all work fine since server not started
		assert.NotPanics(t, func() {
			platform.GET("/test1", func(c *gin.Context) {})
			platform.POST("/test2", func(c *gin.Context) {})
			platform.Use(func(c *gin.Context) { c.Next() })
		})
	})
}
