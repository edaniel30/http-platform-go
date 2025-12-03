package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	platformErrors "github.com/edaniel30/http-platform-go/errors"
	"github.com/edaniel30/loki-logger-go"
	"github.com/edaniel30/loki-logger-go/models"
	mongokit "github.com/edaniel30/mongo-kit-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ApiError represents a structured API error response
type ApiError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Cause   []any  `json:"cause,omitempty"`
}

// NewApiError creates a new ApiError with the given message, status code, and optional causes
func NewApiError(message string, status int, cause ...any) *ApiError {
	return &ApiError{
		Message: message,
		Error:   http.StatusText(status),
		Status:  status,
		Cause:   cause,
	}
}

// validationError represents a single field validation error
type validationError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

// newValidationError creates a new validation error for a specific field
func newValidationError(field, reason string) *validationError {
	return &validationError{
		Field:  field,
		Reason: reason,
	}
}

// ErrorHandler creates a middleware that handles errors and converts them to appropriate HTTP responses
// This middleware should be added to the middleware chain to handle errors consistently across the application
// It automatically logs all errors with structured logging including request context
func ErrorHandler(logger *loki.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

		// Setup panic recovery
		defer func() {
			if err := recover(); err != nil {
				handlePanic(c, err, logger)
			}
		}()

		// Process the request
		c.Next()

		// Handle any errors that were added during request processing
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				handleError(c, err.Err)
			}
		}
	}
}

// handlePanic handles panics and converts them to appropriate error responses
func handlePanic(ctx *gin.Context, err any, logger *loki.Logger) {
	// Build log fields with request context
	logFields := models.Fields{
		"client_ip":          ctx.ClientIP(),
		"method":             ctx.Request.Method,
		"path":               ctx.Request.URL.Path,
		"_skip_stack_trace": true, // Disable automatic stack trace
	}

	// Add trace ID if available
	if traceID := GetTraceID(ctx); traceID != "" {
		logFields["trace_id"] = traceID
	}

	switch er := err.(type) {
	case error:
		logFields["panic"] = er.Error()
		logFields["stack_trace"] = string(debug.Stack())
		logger.Error("Panic recovered", logFields)
		handleError(ctx, er)
	default:
		logFields["panic"] = fmt.Sprintf("%v", err)
		logFields["stack_trace"] = string(debug.Stack())
		logger.Error("Panic recovered (non-error type)", logFields)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			NewApiError("Internal server error panic", http.StatusInternalServerError))
	}
}

// handleError handles different types of errors and converts them to appropriate HTTP responses
func handleError(ctx *gin.Context, err error) {
	var apiErr *ApiError
	var errorType string

	// Build log fields with request context
	logFields := models.Fields{
		"method":    ctx.Request.Method,
		"path":      ctx.Request.URL.Path,
		"client_ip": ctx.ClientIP(),
		"error":     err.Error(),
	}

	// Add trace ID if available
	if traceID := GetTraceID(ctx); traceID != "" {
		logFields["trace_id"] = traceID
	}

	switch error := err.(type) {
	case *platformErrors.NotFoundError:
		errorType = "NotFoundError"
		apiErr = NewApiError(error.Error(), http.StatusNotFound)

	case *platformErrors.UnauthorizedError:
		errorType = "UnauthorizedError"
		apiErr = NewApiError(error.Error(), http.StatusUnauthorized)

	case validator.ValidationErrors:
		errorType = "ValidationError"
		validationErrs := descriptiveValidationErrors(error)
		apiErr = NewApiError("Validation error", http.StatusBadRequest, validationErrs)
		logFields["validation_errors"] = validationErrs

	case *platformErrors.DomainError:
		errorType = "DomainError"
		apiErr = NewApiError(error.Error(), http.StatusBadRequest)

	case *platformErrors.ConflictError:
		errorType = "ConflictError"
		apiErr = NewApiError(error.Error(), http.StatusConflict)

	case *platformErrors.ExternalServiceError:
		errorType = "ExternalServiceError"
		apiErr = NewApiError(error.Error(), error.Status())
		logFields["external_status"] = error.Status()

	case *platformErrors.BadRequestError:
		errorType = "BadRequestError"
		apiErr = NewApiError(error.Error(), http.StatusBadRequest)

	case *json.UnmarshalTypeError:
		errorType = "UnmarshalTypeError"
		apiErr = NewApiError(
			fmt.Sprintf("Invalid type for field '%s', expected %s but got %s",
				error.Field, error.Type.String(), error.Value),
			http.StatusBadRequest,
		)
		logFields["field"] = error.Field
		logFields["expected_type"] = error.Type.String()

	default:
		// Check for mongo-kit/MongoDB driver errors
		if errors.Is(err, mongokit.ErrNoDocuments) {
			errorType = "DocumentNotFoundError"
			apiErr = NewApiError(err.Error(), http.StatusNotFound)
		} else if mongokit.IsDuplicateKeyError(err) {
			errorType = "DuplicateKeyError"
			apiErr = NewApiError(err.Error(), http.StatusConflict)
		} else if errors.Is(err, mongokit.ErrInvalidObjectID) {
			errorType = "InvalidObjectIDError"
			apiErr = NewApiError(err.Error(), http.StatusBadRequest)
		} else if errors.Is(err, mongokit.ErrClientDisconnected) {
			errorType = "DatabaseConnectionError"
			apiErr = NewApiError(err.Error(), http.StatusServiceUnavailable)
		} else if mongokit.IsTimeout(err) {
			errorType = "DatabaseTimeoutError"
			apiErr = NewApiError(err.Error(), http.StatusGatewayTimeout)
		} else if mongokit.IsNetworkError(err) {
			errorType = "DatabaseNetworkError"
			apiErr = NewApiError(err.Error(), http.StatusServiceUnavailable)
		} else {
			errorType = "UnknownError"
			apiErr = NewApiError("An error occurred", http.StatusInternalServerError)
		}
	}

	// Add error type and status to log
	logFields["error_type"] = errorType
	logFields["status"] = apiErr.Status

	ctx.AbortWithStatusJSON(apiErr.Status, apiErr)
}

// descriptiveValidationErrors converts validator.ValidationErrors to a descriptive format
func descriptiveValidationErrors(veer validator.ValidationErrors) []*validationError {
	var errs []*validationError
	for _, ve := range veer {
		err := ve.ActualTag()
		if ve.Param() != "" {
			err = fmt.Sprintf("%s=%s", err, ve.Param())
		}
		errs = append(errs, newValidationError(ve.Field(), err))
	}
	return errs
}
