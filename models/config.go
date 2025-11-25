package config

import (
	"time"

	"github.com/edaniel30/http-platform-go/errors"
	"github.com/edaniel30/loki-logger-go"
)

// Config holds all configuration for the HTTP platform
type Config struct {
	// Port is the port number to listen on
	Port int

	// Mode sets the Gin mode: "debug", "release", or "test"
	Mode string

	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the response
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header
	MaxHeaderBytes int

	// Logger is the loki logger instance (required)
	Logger *loki.Logger

	// CORS configuration
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           time.Duration

	// Middleware toggles
	EnableTraceID  bool
	EnableCORS     bool
	EnableRecovery bool
	EnableLogger   bool

	// BasePath is the base path for all routes (e.g., "/api/v1")
	BasePath string

	// TrustedProxies defines a list of trusted proxies
	TrustedProxies []string
}

type Option func(*Config)

func DefaultConfig() Config {
	return Config{
		Port:             8080,
		Mode:             "debug",
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		IdleTimeout:      60 * time.Second,
		MaxHeaderBytes:   1 << 20, // 1 MB
		Logger:           nil,     // Must be set by user
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Content-Length", "X-Trace-Id"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		EnableTraceID:    true,
		EnableCORS:       true,
		EnableRecovery:   true,
		EnableLogger:     true,
		BasePath:         "",
		TrustedProxies:   nil,
	}
}

func (c *Config) Validate() error {
	if c.Logger == nil {
		return errors.ErrNilLogger()
	}

	if c.Port <= 0 || c.Port > 65535 {
		return errors.ErrInvalidPort(c.Port)
	}

	if c.Mode != "debug" && c.Mode != "release" && c.Mode != "test" {
		return errors.ErrInvalidMode(c.Mode)
	}

	if c.ReadTimeout <= 0 {
		return errors.NewConfigError("readTimeout must be positive")
	}

	if c.WriteTimeout <= 0 {
		return errors.NewConfigError("writeTimeout must be positive")
	}

	if c.IdleTimeout <= 0 {
		return errors.NewConfigError("idleTimeout must be positive")
	}

	return nil
}

func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

func WithMode(mode string) Option {
	return func(c *Config) {
		c.Mode = mode
	}
}

func WithLogger(logger *loki.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

func WithIdleTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.IdleTimeout = timeout
	}
}

func WithMaxHeaderBytes(bytes int) Option {
	return func(c *Config) {
		c.MaxHeaderBytes = bytes
	}
}

func WithCORS(origins []string) Option {
	return func(c *Config) {
		c.AllowedOrigins = origins
	}
}

func WithAllowedMethods(methods []string) Option {
	return func(c *Config) {
		c.AllowedMethods = methods
	}
}

func WithAllowedHeaders(headers []string) Option {
	return func(c *Config) {
		c.AllowedHeaders = headers
	}
}

func WithExposedHeaders(headers []string) Option {
	return func(c *Config) {
		c.ExposedHeaders = headers
	}
}

func WithAllowCredentials(allow bool) Option {
	return func(c *Config) {
		c.AllowCredentials = allow
	}
}

func WithMaxAge(maxAge time.Duration) Option {
	return func(c *Config) {
		c.MaxAge = maxAge
	}
}

func WithoutTraceID() Option {
	return func(c *Config) {
		c.EnableTraceID = false
	}
}

func WithoutCORS() Option {
	return func(c *Config) {
		c.EnableCORS = false
	}
}

func WithoutRecovery() Option {
	return func(c *Config) {
		c.EnableRecovery = false
	}
}

func WithoutLogger() Option {
	return func(c *Config) {
		c.EnableLogger = false
	}
}

func WithBasePath(basePath string) Option {
	return func(c *Config) {
		c.BasePath = basePath
	}
}

func WithTrustedProxies(proxies []string) Option {
	return func(c *Config) {
		c.TrustedProxies = proxies
	}
}
