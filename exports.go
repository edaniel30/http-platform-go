package httpplatform

import (
	"github.com/edaniel30/http-platform-go/errors"
	"github.com/edaniel30/http-platform-go/middleware"
	config "github.com/edaniel30/http-platform-go/models"
)

// Config type from config package
// This allows users to use httpplatform.Config instead of importing the config package separately
type Config = config.Config

// Option type
type Option = config.Option

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

	// WithoutRecovery disables the Recovery middleware
	WithoutRecovery = config.WithoutRecovery

	// WithoutLogger disables the Logger middleware
	WithoutLogger = config.WithoutLogger

	// WithBasePath sets a base path prefix for all routes (e.g., "/api/v1")
	WithBasePath = config.WithBasePath

	// WithTrustedProxies sets the list of trusted proxy IP addresses
	WithTrustedProxies = config.WithTrustedProxies
)

// Re-export error functions from errors package
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

	// NewExternalServiceError creates an error for external service failures with a custom status code
	NewExternalServiceError = errors.NewExternalServiceError

	// NewDomainError creates a 400 error for business logic/domain rule violations
	NewDomainError = errors.NewDomainError
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
)
