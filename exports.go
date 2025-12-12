package httpplatform

import (
	"github.com/edaniel30/http-platform-go/errors"
	"github.com/edaniel30/http-platform-go/middleware"
	config "github.com/edaniel30/http-platform-go/models"
)

// DefaultConfig function
// Returns a Config with sensible defaults
var DefaultConfig = config.DefaultConfig

var (
	// WithPort sets the HTTP server port (default: 8080)
	WithPort = config.WithPort

	// WithMode sets the Gin mode: "debug", "release", or "test" (default: "debug")
	WithMode = config.WithMode

	// WithLogger injects a loki-logger instance (required)
	WithLogger = config.WithLogger

	// WithReadTimeout sets the maximum duration for reading the entire request, including the body
	WithReadTimeout = config.WithReadTimeout

	// WithWriteTimeout sets the maximum duration before timing out writes of the response
	WithWriteTimeout = config.WithWriteTimeout

	// WithIdleTimeout sets the maximum amount of time to wait for the next request when keep-alives are enabled
	WithIdleTimeout = config.WithIdleTimeout

	// WithMaxHeaderBytes sets the maximum number of bytes the server will read parsing the request header's keys and values
	WithMaxHeaderBytes = config.WithMaxHeaderBytes

	// WithCORS sets the allowed origins for CORS requests (e.g., []string{"https://example.com"})
	WithCORS = config.WithCORS

	// WithAllowedMethods sets the allowed HTTP methods for CORS requests (e.g., []string{"GET", "POST"})
	WithAllowedMethods = config.WithAllowedMethods

	// WithAllowedHeaders sets the allowed request headers for CORS requests (e.g., []string{"Content-Type", "Authorization"})
	WithAllowedHeaders = config.WithAllowedHeaders

	// WithExposedHeaders sets the response headers exposed to the client (e.g., []string{"X-Trace-Id"})
	WithExposedHeaders = config.WithExposedHeaders

	// WithAllowCredentials sets whether CORS requests can include credentials (cookies, HTTP authentication)
	WithAllowCredentials = config.WithAllowCredentials

	// WithMaxAge sets the maximum time (in seconds) that preflight request results can be cached
	WithMaxAge = config.WithMaxAge

	// WithoutTraceID disables the TraceID middleware
	WithoutTraceID = config.WithoutTraceID

	// WithoutCORS disables the CORS middleware
	WithoutCORS = config.WithoutCORS

	// WithoutLogger disables the Logger middleware
	WithoutLogger = config.WithoutLogger

	// WithoutContextCancellation disables the ContextCancellation middleware
	WithoutContextCancellation = config.WithoutContextCancellation

	// WithBasePath sets a base path prefix for all routes (e.g., "/api/v1")
	WithBasePath = config.WithBasePath

	// WithTrustedProxies sets the list of trusted proxy IP addresses
	WithTrustedProxies = config.WithTrustedProxies

	// WithTelemetry enables OpenTelemetry tracing with Datadog
	// serviceName: name of the service (e.g., "guardian-auth")
	// version: service version (e.g., "1.0.0")
	// environment: deployment environment (e.g., "production", "staging", "development")
	// otlpEndpoint: Datadog Agent endpoint (e.g., "192.168.1.100:4318")
	WithTelemetry = config.WithTelemetry

	// WithTelemetrySampling configures trace sampling
	// sampleAll: if true, samples all traces. If false, uses default sampling (10%)
	WithTelemetrySampling = config.WithTelemetrySampling

	// WithoutTelemetry disables telemetry (default is disabled)
	WithoutTelemetry = config.WithoutTelemetry
)

// Error functions from errors package
// This allows users to work with platform errors without importing the errors package separately
var (
	// Configuration and Runtime Errors

	// NewConfigError creates a new configuration error with a custom message
	NewConfigError = errors.NewConfigError

	// ErrNilLogger returns an error when logger is not provided (required)
	ErrNilLogger = errors.ErrNilLogger

	// ErrInvalidPort returns an error when the port number is invalid
	ErrInvalidPort = errors.ErrInvalidPort

	// ErrInvalidMode returns an error when the Gin mode is invalid (must be debug, release, or test)
	ErrInvalidMode = errors.ErrInvalidMode

	// NewRuntimeError creates a new runtime error with a message and optional cause
	NewRuntimeError = errors.NewRuntimeError

	// ErrAlreadyStarted returns an error when attempting to start an already running platform
	ErrAlreadyStarted = errors.ErrAlreadyStarted

	// ErrNotStarted returns an error when attempting to stop a platform that is not running
	ErrNotStarted = errors.ErrNotStarted

	// HTTP Domain Errors

	// NewNotFoundError creates a 404 Not Found error with a custom message
	NewNotFoundError = errors.NewNotFoundError

	// NewUnauthorizedError creates a 401 Unauthorized error with a custom message
	NewUnauthorizedError = errors.NewUnauthorizedError

	// NewConflictError creates a 409 Conflict error with a custom message (e.g., duplicate resource)
	NewConflictError = errors.NewConflictError

	// NewBadRequestError creates a 400 Bad Request error with a custom message
	NewBadRequestError = errors.NewBadRequestError

	// NewForbiddenError creates a 403 Forbidden error with a custom message (user lacks permissions)
	NewForbiddenError = errors.NewForbiddenError

	// NewUnprocessableEntityError creates a 422 Unprocessable Entity error with a custom message (semantic validation errors)
	NewUnprocessableEntityError = errors.NewUnprocessableEntityError

	// NewTooManyRequestsError creates a 429 Too Many Requests error with a custom message (rate limiting)
	NewTooManyRequestsError = errors.NewTooManyRequestsError

	// NewInternalServerError creates a 500 Internal Server Error with a custom message
	NewInternalServerError = errors.NewInternalServerError

	// NewServiceUnavailableError creates a 503 Service Unavailable error with a custom message (temporary unavailability)
	NewServiceUnavailableError = errors.NewServiceUnavailableError

	// NewExternalServiceError creates an error for external service failures with a custom status code
	NewExternalServiceError = errors.NewExternalServiceError
)

// Middleware functions
var (
	// ErrorHandler creates a middleware that handles errors and converts them to structured JSON responses.
	// It automatically handles domain errors (NotFoundError, UnauthorizedError, etc.), validation errors,
	// JSON unmarshaling errors, and panics. Requires a logger instance for automatic error logging with
	// request context (method, path, client IP, trace ID, error type, status).
	// Apply globally with platform.Use(ErrorHandler(logger)) or to specific routes/groups.
	// Returns consistent JSON format: {"message": "...", "error": "...", "status": 400, "cause": [...]}
	ErrorHandler = middleware.ErrorHandler

	// ContextCancellation creates a middleware that detects client disconnections early.
	// Enabled by default via cfg.EnableContextCancellation. Use this directly for specific routes only.
	ContextCancellation = middleware.ContextCancellation

	// WithTimeout creates a middleware that enforces a timeout for specific endpoints.
	// Example: router.GET("/slow", httpplatform.WithTimeout(5*time.Second), handler)
	WithTimeout = middleware.WithTimeout
)

// Context helper functions for checking request cancellation in handlers
var (
	// IsContextCancelled checks if the client has disconnected (context cancelled).
	// Use in long-running handlers to avoid wasted work:
	//   if httpplatform.IsContextCancelled(c) { return }
	IsContextCancelled = middleware.IsContextCancelled

	// GetContextError returns context.Canceled or context.DeadlineExceeded if applicable.
	// Returns nil if the context is still valid.
	GetContextError = middleware.GetContextError
)

// Logger interface and Fields type from middleware package
// This allows users to implement custom loggers without importing middleware directly
type (
	// Logger is the interface that any logger implementation must satisfy.
	// The platform expects a logger that implements this interface for structured logging.
	Logger = middleware.Logger

	// Fields represents a map of structured log fields for adding metadata to log entries.
	Fields = middleware.Fields
)
