package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	platformErrors "github.com/edaniel30/http-platform-go/errors"
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

// ErrorHandler creates a middleware that handles errors and panics, converting them to appropriate HTTP responses
// This middleware:
// - Recovers from panics and logs them with stack traces
// - Handles platform-specific errors (NotFound, Unauthorized, Forbidden, TooManyRequests, etc.)
// - Handles validation errors from go-playground/validator
// - Handles JSON parsing errors (syntax errors, type mismatches)
// - Handles request body errors (empty body, incomplete body)
// - Handles context cancellation (client disconnect, timeout)
// - Logs errors with appropriate severity levels and structured fields
//
// For more advanced error handling (e.g., database-specific errors), implement a custom error handler in your application
func ErrorHandler(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Setup panic recovery
		defer func() {
			if err := recover(); err != nil {
				handlePanic(c, err, logger)
			}
		}()

		// Process the request
		c.Next()

		// Handle any errors that were added during request processing
		// Only handle the first error to avoid multiple responses
		if len(c.Errors) > 0 {
			handleBasicError(c, c.Errors[0].Err, logger)
		}
	}
}

// buildLogFields creates base log fields with request context and trace ID
func buildLogFields(ctx *gin.Context) Fields {
	logFields := Fields{
		"client_ip": ctx.ClientIP(),
		"method":    ctx.Request.Method,
		"path":      ctx.Request.URL.Path,
	}

	// Add trace ID if available
	if traceID := GetTraceID(ctx); traceID != "" {
		logFields["trace_id"] = traceID
	}

	return logFields
}

// handlePanic handles panics and converts them to appropriate error responses
func handlePanic(ctx *gin.Context, err any, logger Logger) {
	// Build log fields with request context
	logFields := buildLogFields(ctx)

	reqCtx := ctx.Request.Context()
	switch er := err.(type) {
	case error:
		logFields["panic"] = er.Error()
		logFields["stack_trace"] = string(debug.Stack())
		logger.Error(reqCtx, "Panic recovered", logFields)
		handleBasicError(ctx, er, logger)
	default:
		logFields["panic"] = fmt.Sprintf("%v", err)
		logFields["stack_trace"] = string(debug.Stack())
		logger.Error(reqCtx, "Panic recovered (non-error type)", logFields)
		// Set Content-Type header before sending response
		ctx.Header("Content-Type", "application/json; charset=utf-8")
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			NewApiError("Internal server error panic", http.StatusInternalServerError))
	}
}

// handleBasicError handles different types of errors and converts them to appropriate HTTP responses
// This version only handles platform-specific errors, not database-specific errors
func handleBasicError(ctx *gin.Context, err error, logger Logger) {
	var apiErr *ApiError
	var errorType string

	// Build log fields with request context
	logFields := buildLogFields(ctx)
	logFields["error"] = err.Error()

	switch e := err.(type) {
	case *platformErrors.NotFoundError:
		errorType = "NotFoundError"
		apiErr = NewApiError(e.Error(), http.StatusNotFound)

	case *platformErrors.UnauthorizedError:
		errorType = "UnauthorizedError"
		apiErr = NewApiError(e.Error(), http.StatusUnauthorized)

	case *platformErrors.ConflictError:
		errorType = "ConflictError"
		apiErr = NewApiError(e.Error(), http.StatusConflict)

	case *platformErrors.ExternalServiceError:
		errorType = "ExternalServiceError"
		apiErr = NewApiError(e.Error(), e.Status())
		logFields["external_status"] = e.Status()

	case *platformErrors.BadRequestError:
		errorType = "BadRequestError"
		apiErr = NewApiError(e.Error(), http.StatusBadRequest)

	case *platformErrors.ForbiddenError:
		errorType = "ForbiddenError"
		apiErr = NewApiError(e.Error(), http.StatusForbidden)

	case *platformErrors.UnprocessableEntityError:
		errorType = "UnprocessableEntityError"
		apiErr = NewApiError(e.Error(), http.StatusUnprocessableEntity)

	case *platformErrors.TooManyRequestsError:
		errorType = "TooManyRequestsError"
		apiErr = NewApiError(e.Error(), http.StatusTooManyRequests)

	case *platformErrors.InternalServerError:
		errorType = "InternalServerError"
		apiErr = NewApiError(e.Error(), http.StatusInternalServerError)

	case *platformErrors.ServiceUnavailableError:
		errorType = "ServiceUnavailableError"
		apiErr = NewApiError(e.Error(), http.StatusServiceUnavailable)

	case *json.UnmarshalTypeError:
		errorType = "UnmarshalTypeError"
		apiErr = NewApiError(
			fmt.Sprintf("Invalid type for field '%s', expected %s but got %s",
				e.Field, e.Type.String(), e.Value),
			http.StatusBadRequest,
		)
		logFields["field"] = e.Field
		logFields["expected_type"] = e.Type.String()

	case validator.ValidationErrors:
		errorType = "ValidationError"
		validationErrs := descriptiveValidationErrors(e)
		apiErr = NewApiError("Validation error", http.StatusBadRequest, validationErrs)
		logFields["validation_errors"] = validationErrs

	case *json.SyntaxError:
		errorType = "JSONSyntaxError"
		apiErr = NewApiError(
			fmt.Sprintf("Invalid JSON syntax at position %d", e.Offset),
			http.StatusBadRequest,
		)
		logFields["offset"] = e.Offset
		logFields["syntax_error"] = e.Error()

	default:
		// Check for specific error types using errors.Is
		if errors.Is(err, io.EOF) {
			errorType = "EmptyBody"
			apiErr = NewApiError("Request body is empty", http.StatusBadRequest)
		} else if errors.Is(err, io.ErrUnexpectedEOF) {
			errorType = "IncompleteBody"
			apiErr = NewApiError("Request body is incomplete", http.StatusBadRequest)
		} else if err == context.Canceled {
			// Check for context cancellation errors
			errorType = "RequestCanceled"
			// 499 is nginx's non-standard status code for "Client Closed Request"
			// Since HTTP doesn't have a standard code, we use 499 or could use 408 Request Timeout
			apiErr = NewApiError("Request was cancelled by client", 499)
			logFields["reason"] = "context_canceled"
		} else if err == context.DeadlineExceeded {
			errorType = "RequestTimeout"
			apiErr = NewApiError("Request timeout exceeded", http.StatusRequestTimeout)
			logFields["reason"] = "deadline_exceeded"
		} else {
			errorType = "UnknownError"
			apiErr = NewApiError("An error occurred", http.StatusInternalServerError)
			// Log full error for unknown errors
			logFields["full_error"] = fmt.Sprintf("%+v", err)
		}
	}

	// Add error type and status to log
	logFields["error_type"] = errorType
	logFields["status"] = apiErr.Status

	// Log based on severity
	reqCtx := ctx.Request.Context()
	if apiErr.Status >= 500 {
		logger.Error(reqCtx, "Server error", logFields)
	} else {
		logger.Warn(reqCtx, "Client error", logFields)
	}

	// Set Content-Type header before sending response
	ctx.Header("Content-Type", "application/json; charset=utf-8")
	ctx.AbortWithStatusJSON(apiErr.Status, apiErr)
}

// descriptiveValidationErrors converts validator.ValidationErrors to a descriptive format
func descriptiveValidationErrors(validationErrs validator.ValidationErrors) []*validationError {
	var errs []*validationError
	for _, fieldErr := range validationErrs {
		tagName := fieldErr.ActualTag()
		if fieldErr.Param() != "" {
			tagName = fmt.Sprintf("%s=%s", tagName, fieldErr.Param())
		}
		errs = append(errs, newValidationError(fieldErr.Field(), tagName))
	}
	return errs
}
