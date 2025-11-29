package httpplatform

import (
	"github.com/edaniel30/http-platform-go/errors"
	"github.com/edaniel30/http-platform-go/middleware"
	config "github.com/edaniel30/http-platform-go/models"
)

// Re-export Config type from config package
// This allows users to use httpplatform.Config instead of importing the config package separately
type Config = config.Config

// Re-export Option type
type Option = config.Option

// Re-export DefaultConfig function
// Returns a Config with sensible defaults
var DefaultConfig = config.DefaultConfig

// Re-export configuration option functions
// These allow customizing the platform behavior using the functional options pattern
var (
	WithPort              = config.WithPort
	WithMode              = config.WithMode
	WithLogger            = config.WithLogger
	WithReadTimeout       = config.WithReadTimeout
	WithWriteTimeout      = config.WithWriteTimeout
	WithIdleTimeout       = config.WithIdleTimeout
	WithMaxHeaderBytes    = config.WithMaxHeaderBytes
	WithCORS              = config.WithCORS
	WithAllowedMethods    = config.WithAllowedMethods
	WithAllowedHeaders    = config.WithAllowedHeaders
	WithExposedHeaders    = config.WithExposedHeaders
	WithAllowCredentials  = config.WithAllowCredentials
	WithMaxAge            = config.WithMaxAge
	WithoutTraceID        = config.WithoutTraceID
	WithoutCORS           = config.WithoutCORS
	WithoutRecovery       = config.WithoutRecovery
	WithoutLogger         = config.WithoutLogger
	WithBasePath          = config.WithBasePath
	WithTrustedProxies    = config.WithTrustedProxies
)

// Re-export error functions from errors package
// This allows users to work with platform errors without importing the errors package separately
var (
	// Configuration errors
	NewConfigError    = errors.NewConfigError
	ErrNilLogger      = errors.ErrNilLogger
	ErrInvalidPort    = errors.ErrInvalidPort
	ErrInvalidMode    = errors.ErrInvalidMode
	NewRuntimeError   = errors.NewRuntimeError
	ErrAlreadyStarted = errors.ErrAlreadyStarted
	ErrNotStarted     = errors.ErrNotStarted

	// HTTP domain errors
	NewNotFoundError         = errors.NewNotFoundError
	NewUnauthorizedError     = errors.NewUnauthorizedError
	NewConflictError         = errors.NewConflictError
	NewBadRequestError       = errors.NewBadRequestError
	NewExternalServiceError  = errors.NewExternalServiceError
	NewDomainError           = errors.NewDomainError
)

// Re-export middleware functions
var (
	ErrorHandler = middleware.ErrorHandler
)
