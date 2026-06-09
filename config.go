package httpplatform

import (
	"time"

	"github.com/edaniel30/http-platform-go/internal/constants"
	"github.com/edaniel30/http-platform-go/internal/logger"
	"github.com/edaniel30/http-platform-go/internal/middleware"
)

// TelemetryConfig holds OpenTelemetry configuration for distributed tracing.
type TelemetryConfig struct {
	ServiceName  string // Service name for tracing
	Version      string // Service version
	Environment  string // Environment (e.g., "production", "development")
	OTLPEndpoint string // OTLP endpoint (e.g., "localhost:4318" for Datadog Agent)
	SampleAll    bool   // If true, samples all traces; if false, uses default sampling
}

// Config holds all configuration for the HTTP platform.
// Use DefaultConfig() to get sensible defaults, then customize with Option functions.
type Config struct {
	// Required fields
	Logger Logger // Logger instance (required, any logger that implements Logger interface)

	// HTTP server settings
	Port           int           // Port number to listen on (default: 8080)
	Mode           string        // Gin mode: "debug", "release", or "test" (default: "debug")
	BasePath       string        // Base path for all routes (e.g., "/api/v1")
	TrustedProxies []string      // List of trusted proxy IPs
	ReadTimeout    time.Duration // Maximum duration for reading the entire request (default: 30s)
	WriteTimeout   time.Duration // Maximum duration for writing the response (default: 30s)
	IdleTimeout    time.Duration // Maximum time to wait for the next request (default: 60s)
	MaxHeaderBytes int           // Maximum bytes for request headers (default: 1MB)

	// Middleware configuration
	CORS      *middleware.CORSConfig // CORS configuration (nil to disable)
	Telemetry *TelemetryConfig       // Telemetry configuration (nil to disable)

	// Middleware toggles
	EnableTraceID             bool // Add X-Trace-ID header to requests/responses (default: true)
	EnableLogger              bool // Enable request/response logging (default: true)
	EnableContextCancellation bool // Detect and handle client disconnections early (default: true)
}

// Option is a function that modifies a Config.
// Use Option functions with New() to customize the configuration.
type Option func(*Config)

// ensureCORS initializes CORS configuration if it's nil
func (c *Config) ensureCORS() {
	if c.CORS == nil {
		c.CORS = &middleware.CORSConfig{}
	}
}

// DefaultConfig returns a Config with sensible default values.
// This is the recommended starting point for most applications.
//
// Default values:
//   - Port: 8080
//   - Mode: "debug"
//   - ReadTimeout: 30 seconds
//   - WriteTimeout: 30 seconds
//   - IdleTimeout: 60 seconds
//   - MaxHeaderBytes: 1 MB
//   - Logger: nil (must be set by user)
//   - CORS: Enabled with permissive defaults (["*"] origin)
//   - Telemetry: Disabled by default
//   - EnableTraceID: true
//   - EnableLogger: true
//   - EnableContextCancellation: true
//
// Example:
//
//	cfg := config.DefaultConfig()
//	platform, err := New(cfg,
//		config.WithPort(3000),
//		config.WithLogger(myLogger),
//	)
func DefaultConfig() Config {
	return Config{
		Port:           8080,
		Mode:           constants.GinModeDebug,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: constants.DefaultMaxHeaderBytes,
		Logger:         logger.NewDefaultLogger(), // Uses default stdlib logger
		CORS: &middleware.CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
			AllowedHeaders:   []string{"*"},
			ExposedHeaders:   []string{"Content-Length", "X-Trace-Id"},
			AllowCredentials: false, // Must be false when using wildcard origin "*"
			MaxAge:           12 * time.Hour,
		},
		Telemetry:                 nil, // Disabled by default
		EnableTraceID:             true,
		EnableLogger:              true,
		EnableContextCancellation: true,
		BasePath:                  "",
		TrustedProxies:            nil,
	}
}

// WithPort sets the port number to listen on.
//
// Example:
//
//	config.WithPort(3000)
//	config.WithPort(8443)
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithMode sets the Gin mode: "debug", "release", or "test".
//
// Example:
//
//	config.WithMode("release")
//	config.WithMode("debug")
func WithMode(mode string) Option {
	return func(c *Config) {
		c.Mode = mode
	}
}

// WithLogger sets the logger instance (required).
// Any logger that implements the Logger interface can be used.
//
// Example:
//
//	config.WithLogger(myLogger)
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithReadTimeout sets the maximum duration for reading the entire request.
// Default is 30 seconds.
//
// Example:
//
//	config.WithReadTimeout(60 * time.Second)
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the maximum duration for writing the response.
// Default is 30 seconds.
//
// Example:
//
//	config.WithWriteTimeout(60 * time.Second)
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

// WithIdleTimeout sets the maximum time to wait for the next request.
// Default is 60 seconds.
//
// Example:
//
//	config.WithIdleTimeout(120 * time.Second)
func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.IdleTimeout = timeout
	}
}

// WithMaxHeaderBytes sets the maximum bytes for request headers.
// Default is 1 MB (1 << 20).
//
// Example:
//
//	config.WithMaxHeaderBytes(2 << 20) // 2 MB
func WithMaxHeaderBytes(bytes int) Option {
	return func(c *Config) {
		c.MaxHeaderBytes = bytes
	}
}

// WithBasePath sets the base path for all routes.
//
// Example:
//
//	config.WithBasePath("/api/v1")
//	config.WithBasePath("/v2")
func WithBasePath(basePath string) Option {
	return func(c *Config) {
		c.BasePath = basePath
	}
}

// WithTrustedProxies sets the list of trusted proxy IPs.
//
// Example:
//
//	config.WithTrustedProxies([]string{"192.168.1.1", "10.0.0.1"})
func WithTrustedProxies(proxies []string) Option {
	return func(c *Config) {
		c.TrustedProxies = proxies
	}
}

// WithCORS configures CORS with custom settings.
// Pass nil to disable CORS.
//
// Example:
//
//	config.WithCORS(&middleware.CORSConfig{
//		AllowedOrigins: []string{"https://example.com"},
//		AllowedMethods: []string{"GET", "POST"},
//		AllowedHeaders: []string{"Content-Type"},
//		AllowCredentials: true,
//		MaxAge: 12 * time.Hour,
//	})
//
// To disable CORS:
//
//	config.WithCORS(nil)
func WithCORS(cors *middleware.CORSConfig) Option {
	return func(c *Config) {
		c.CORS = cors
	}
}

// WithCORSOrigins sets the allowed origins for CORS.
// Initializes CORS if not already configured.
//
// Example:
//
//	config.WithCORSOrigins([]string{"https://example.com", "https://app.example.com"})
//	config.WithCORSOrigins([]string{"*"})
func WithCORSOrigins(origins []string) Option {
	return func(c *Config) {
		c.ensureCORS()
		c.CORS.AllowedOrigins = origins
	}
}

// WithCORSMethods sets the allowed HTTP methods for CORS.
// Initializes CORS if not already configured.
//
// Example:
//
//	config.WithCORSMethods([]string{"GET", "POST", "PUT"})
func WithCORSMethods(methods []string) Option {
	return func(c *Config) {
		c.ensureCORS()
		c.CORS.AllowedMethods = methods
	}
}

// WithCORSHeaders sets the allowed request headers for CORS.
// Initializes CORS if not already configured.
//
// Example:
//
//	config.WithCORSHeaders([]string{"Content-Type", "Authorization"})
func WithCORSHeaders(headers []string) Option {
	return func(c *Config) {
		c.ensureCORS()
		c.CORS.AllowedHeaders = headers
	}
}

// WithCORSCredentials enables or disables credentials for CORS.
// Cannot be true when AllowedOrigins is ["*"].
// Initializes CORS if not already configured.
//
// Example:
//
//	config.WithCORSCredentials(true)
func WithCORSCredentials(allow bool) Option {
	return func(c *Config) {
		c.ensureCORS()
		c.CORS.AllowCredentials = allow
	}
}

// WithCORSWildcard enables pattern matching in AllowedOrigins entries.
// Allows patterns like "http://localhost:*" or "https://*.example.com".
// Unlike the global "*" wildcard, this is compatible with AllowCredentials.
// Initializes CORS if not already configured.
//
// Example:
//
//	config.WithCORSWildcard(true)
//	config.WithCORSOrigins([]string{"http://localhost:*", "https://app.example.com"})
func WithCORSWildcard(allow bool) Option {
	return func(c *Config) {
		c.ensureCORS()
		c.CORS.AllowWildcard = allow
	}
}

// WithTelemetry enables OpenTelemetry distributed tracing.
// Pass nil to disable telemetry.
//
// Example:
//
//	config.WithTelemetry(&config.TelemetryConfig{
//		ServiceName: "my-service",
//		Version: "1.0.0",
//		Environment: "production",
//		OTLPEndpoint: "localhost:4318",
//		SampleAll: false,
//	})
//
// To disable telemetry:
//
//	config.WithTelemetry(nil)
func WithTelemetry(telemetry *TelemetryConfig) Option {
	return func(c *Config) {
		c.Telemetry = telemetry
	}
}

// WithoutTraceID disables the automatic X-Trace-ID header.
//
// Example:
//
//	config.WithoutTraceID()
func WithoutTraceID() Option {
	return func(c *Config) {
		c.EnableTraceID = false
	}
}

// WithoutLogger disables the request/response logging middleware.
//
// Example:
//
//	config.WithoutLogger()
func WithoutLogger() Option {
	return func(c *Config) {
		c.EnableLogger = false
	}
}

// WithoutContextCancellation disables early detection of client disconnections.
//
// Example:
//
//	config.WithoutContextCancellation()
func WithoutContextCancellation() Option {
	return func(c *Config) {
		c.EnableContextCancellation = false
	}
}

// Validate checks if the configuration is valid.
// Returns a ConfigError if any required field is missing or invalid.
func (c *Config) validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return newConfigFieldError("Port", "Port must be between 1 and 65535")
	}

	if c.Mode != constants.GinModeDebug && c.Mode != constants.GinModeRelease && c.Mode != constants.GinModeTest {
		return newConfigFieldError("Mode", "Mode must be 'debug', 'release', or 'test'")
	}

	if c.ReadTimeout <= 0 {
		return newConfigFieldError("ReadTimeout", "ReadTimeout must be positive")
	}

	if c.WriteTimeout <= 0 {
		return newConfigFieldError("WriteTimeout", "WriteTimeout must be positive")
	}

	if c.IdleTimeout <= 0 {
		return newConfigFieldError("IdleTimeout", "IdleTimeout must be positive")
	}

	// Validate CORS configuration if enabled
	if c.CORS != nil {
		// CORS spec: wildcard origin "*" cannot be used with credentials
		if middleware.IsWildcardOrigin(c.CORS.AllowedOrigins) && c.CORS.AllowCredentials {
			return newConfigFieldError("CORS", "AllowCredentials cannot be true when AllowedOrigins is [\"*\"]. Either set AllowCredentials to false or specify explicit origins")
		}
	}

	return nil
}
