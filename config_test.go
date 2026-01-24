package httpplatform

import (
	"errors"
	"testing"
	"time"

	"github.com/edaniel30/http-platform-go/internal/constants"
	"github.com/edaniel30/http-platform-go/internal/logger"
	"github.com/edaniel30/http-platform-go/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Test all default values
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, constants.GinModeDebug, cfg.Mode)
	assert.Equal(t, 30*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
	assert.Equal(t, 60*time.Second, cfg.IdleTimeout)
	assert.Equal(t, constants.DefaultMaxHeaderBytes, cfg.MaxHeaderBytes)
	assert.NotNil(t, cfg.Logger)
	assert.True(t, cfg.EnableTraceID)
	assert.True(t, cfg.EnableLogger)
	assert.True(t, cfg.EnableContextCancellation)
	assert.Empty(t, cfg.BasePath)
	assert.Nil(t, cfg.TrustedProxies)
	assert.Nil(t, cfg.Telemetry)

	// Test CORS defaults
	require.NotNil(t, cfg.CORS)
	assert.Equal(t, []string{"*"}, cfg.CORS.AllowedOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}, cfg.CORS.AllowedMethods)
	assert.Equal(t, []string{"*"}, cfg.CORS.AllowedHeaders)
	assert.Equal(t, []string{"Content-Length", "X-Trace-Id"}, cfg.CORS.ExposedHeaders)
	assert.False(t, cfg.CORS.AllowCredentials)
	assert.Equal(t, 12*time.Hour, cfg.CORS.MaxAge)
}

func TestConfigOptions(t *testing.T) {
	tests := []struct {
		name   string
		option Option
		verify func(*testing.T, *Config)
	}{
		{
			name:   "WithPort",
			option: WithPort(3000),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, 3000, c.Port)
			},
		},
		{
			name:   "WithMode",
			option: WithMode(constants.GinModeRelease),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, constants.GinModeRelease, c.Mode)
			},
		},
		{
			name:   "WithLogger",
			option: WithLogger(logger.NewDefaultLogger()),
			verify: func(t *testing.T, c *Config) {
				assert.NotNil(t, c.Logger)
			},
		},
		{
			name:   "WithReadTimeout",
			option: WithReadTimeout(60 * time.Second),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, 60*time.Second, c.ReadTimeout)
			},
		},
		{
			name:   "WithWriteTimeout",
			option: WithWriteTimeout(45 * time.Second),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, 45*time.Second, c.WriteTimeout)
			},
		},
		{
			name:   "WithIdleTimeout",
			option: WithIdleTimeout(120 * time.Second),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, 120*time.Second, c.IdleTimeout)
			},
		},
		{
			name:   "WithMaxHeaderBytes",
			option: WithMaxHeaderBytes(2 << 20),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, 2<<20, c.MaxHeaderBytes)
			},
		},
		{
			name:   "WithBasePath",
			option: WithBasePath("/api/v1"),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, "/api/v1", c.BasePath)
			},
		},
		{
			name:   "WithTrustedProxies",
			option: WithTrustedProxies([]string{"192.168.1.1", "10.0.0.1"}),
			verify: func(t *testing.T, c *Config) {
				assert.Equal(t, []string{"192.168.1.1", "10.0.0.1"}, c.TrustedProxies)
			},
		},
		{
			name: "WithCORS",
			option: WithCORS(&middleware.CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
			}),
			verify: func(t *testing.T, c *Config) {
				require.NotNil(t, c.CORS)
				assert.Equal(t, []string{"https://example.com"}, c.CORS.AllowedOrigins)
			},
		},
		{
			name:   "WithCORSOrigins",
			option: WithCORSOrigins([]string{"https://example.com", "https://api.example.com"}),
			verify: func(t *testing.T, c *Config) {
				require.NotNil(t, c.CORS)
				assert.Equal(t, []string{"https://example.com", "https://api.example.com"}, c.CORS.AllowedOrigins)
			},
		},
		{
			name:   "WithCORSMethods",
			option: WithCORSMethods([]string{"GET", "POST"}),
			verify: func(t *testing.T, c *Config) {
				require.NotNil(t, c.CORS)
				assert.Equal(t, []string{"GET", "POST"}, c.CORS.AllowedMethods)
			},
		},
		{
			name:   "WithCORSHeaders",
			option: WithCORSHeaders([]string{"Content-Type", "Authorization"}),
			verify: func(t *testing.T, c *Config) {
				require.NotNil(t, c.CORS)
				assert.Equal(t, []string{"Content-Type", "Authorization"}, c.CORS.AllowedHeaders)
			},
		},
		{
			name:   "WithCORSCredentials",
			option: WithCORSCredentials(true),
			verify: func(t *testing.T, c *Config) {
				require.NotNil(t, c.CORS)
				assert.True(t, c.CORS.AllowCredentials)
			},
		},
		{
			name: "WithTelemetry",
			option: WithTelemetry(&TelemetryConfig{
				ServiceName:  "test-service",
				Version:      "1.0.0",
				Environment:  "test",
				OTLPEndpoint: "localhost:4318",
				SampleAll:    true,
			}),
			verify: func(t *testing.T, c *Config) {
				require.NotNil(t, c.Telemetry)
				assert.Equal(t, "test-service", c.Telemetry.ServiceName)
				assert.Equal(t, "1.0.0", c.Telemetry.Version)
				assert.Equal(t, "test", c.Telemetry.Environment)
				assert.Equal(t, "localhost:4318", c.Telemetry.OTLPEndpoint)
				assert.True(t, c.Telemetry.SampleAll)
			},
		},
		{
			name:   "WithoutTraceID",
			option: WithoutTraceID(),
			verify: func(t *testing.T, c *Config) {
				assert.False(t, c.EnableTraceID)
			},
		},
		{
			name:   "WithoutLogger",
			option: WithoutLogger(),
			verify: func(t *testing.T, c *Config) {
				assert.False(t, c.EnableLogger)
			},
		},
		{
			name:   "WithoutContextCancellation",
			option: WithoutContextCancellation(),
			verify: func(t *testing.T, c *Config) {
				assert.False(t, c.EnableContextCancellation)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.option(&cfg)
			tt.verify(t, &cfg)
		})
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid config",
			config:    DefaultConfig(),
			wantError: false,
		},
		{
			name: "invalid port - zero",
			config: Config{
				Port:         0,
				Mode:         constants.GinModeDebug,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				Logger:       logger.NewDefaultLogger(),
			},
			wantError: true,
			errorMsg:  "Port must be between 1 and 65535",
		},
		{
			name: "invalid port - negative",
			config: Config{
				Port:         -1,
				Mode:         constants.GinModeDebug,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				Logger:       logger.NewDefaultLogger(),
			},
			wantError: true,
			errorMsg:  "Port must be between 1 and 65535",
		},
		{
			name: "invalid port - too high",
			config: Config{
				Port:         65536,
				Mode:         constants.GinModeDebug,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				Logger:       logger.NewDefaultLogger(),
			},
			wantError: true,
			errorMsg:  "Port must be between 1 and 65535",
		},
		{
			name: "invalid mode",
			config: Config{
				Port:         8080,
				Mode:         "invalid",
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				Logger:       logger.NewDefaultLogger(),
			},
			wantError: true,
			errorMsg:  "Mode must be 'debug', 'release', or 'test'",
		},
		{
			name: "invalid read timeout",
			config: Config{
				Port:         8080,
				Mode:         constants.GinModeDebug,
				ReadTimeout:  0,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				Logger:       logger.NewDefaultLogger(),
			},
			wantError: true,
			errorMsg:  "ReadTimeout must be positive",
		},
		{
			name: "invalid write timeout",
			config: Config{
				Port:         8080,
				Mode:         constants.GinModeDebug,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 0,
				IdleTimeout:  60 * time.Second,
				Logger:       logger.NewDefaultLogger(),
			},
			wantError: true,
			errorMsg:  "WriteTimeout must be positive",
		},
		{
			name: "invalid idle timeout",
			config: Config{
				Port:         8080,
				Mode:         constants.GinModeDebug,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  0,
				Logger:       logger.NewDefaultLogger(),
			},
			wantError: true,
			errorMsg:  "IdleTimeout must be positive",
		},
		{
			name: "invalid CORS - wildcard with credentials",
			config: Config{
				Port:         8080,
				Mode:         constants.GinModeDebug,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				IdleTimeout:  60 * time.Second,
				Logger:       logger.NewDefaultLogger(),
				CORS: &middleware.CORSConfig{
					AllowedOrigins:   []string{"*"},
					AllowCredentials: true,
				},
			},
			wantError: true,
			errorMsg:  "AllowCredentials cannot be true when AllowedOrigins is [\"*\"]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)

				// Verify it's a ConfigError
				var configErr *ConfigError
				assert.True(t, errors.As(err, &configErr))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
