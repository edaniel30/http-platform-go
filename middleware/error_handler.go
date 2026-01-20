package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/edaniel30/http-platform-go/internal/constants"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const contentTypeJSON = "application/json; charset=utf-8"

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
	logFields := Fields{}
	addBaseRequestFields(logFields, ctx)
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
		ctx.Header("Content-Type", contentTypeJSON)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			NewApiError("Internal server error panic", http.StatusInternalServerError))
	}
}

// mapErrorToApiError maps different error types to API errors with metadata
// Returns: errorType, apiError, and additional log fields
func mapErrorToApiError(err error) (string, *ApiError, Fields) {
	additionalFields := Fields{}

	switch e := err.(type) {
	case *NotFoundError:
		return "NotFoundError", NewApiError(e.Error(), http.StatusNotFound), additionalFields

	case *UnauthorizedError:
		return "UnauthorizedError", NewApiError(e.Error(), http.StatusUnauthorized), additionalFields

	case *ConflictError:
		return "ConflictError", NewApiError(e.Error(), http.StatusConflict), additionalFields

	case *ExternalServiceError:
		additionalFields["external_status"] = e.Status()
		return "ExternalServiceError", NewApiError(e.Error(), e.Status()), additionalFields

	case *BadRequestError:
		return "BadRequestError", NewApiError(e.Error(), http.StatusBadRequest), additionalFields

	case *ForbiddenError:
		return "ForbiddenError", NewApiError(e.Error(), http.StatusForbidden), additionalFields

	case *UnprocessableEntityError:
		return "UnprocessableEntityError", NewApiError(e.Error(), http.StatusUnprocessableEntity), additionalFields

	case *TooManyRequestsError:
		return "TooManyRequestsError", NewApiError(e.Error(), http.StatusTooManyRequests), additionalFields

	case *InternalServerError:
		return "InternalServerError", NewApiError(e.Error(), http.StatusInternalServerError), additionalFields

	case *ServiceUnavailableError:
		return "ServiceUnavailableError", NewApiError(e.Error(), http.StatusServiceUnavailable), additionalFields

	case *json.UnmarshalTypeError:
		additionalFields["field"] = e.Field
		additionalFields["expected_type"] = e.Type.String()
		apiErr := NewApiError(
			fmt.Sprintf("Invalid type for field '%s', expected %s but got %s", e.Field, e.Type.String(), e.Value),
			http.StatusBadRequest,
		)
		return "UnmarshalTypeError", apiErr, additionalFields

	case validator.ValidationErrors:
		validationErrs := descriptiveValidationErrors(e)
		additionalFields["validation_errors"] = validationErrs
		return "ValidationError", NewApiError("Validation error", http.StatusBadRequest, validationErrs), additionalFields

	case *json.SyntaxError:
		additionalFields["offset"] = e.Offset
		additionalFields["syntax_error"] = e.Error()
		apiErr := NewApiError(fmt.Sprintf("Invalid JSON syntax at position %d", e.Offset), http.StatusBadRequest)
		return "JSONSyntaxError", apiErr, additionalFields

	default:
		return mapStandardError(err, additionalFields)
	}
}

// mapStandardError handles standard library errors using errors.Is
func mapStandardError(err error, additionalFields Fields) (string, *ApiError, Fields) {
	if errors.Is(err, io.EOF) {
		return "EmptyBody", NewApiError("Request body is empty", http.StatusBadRequest), additionalFields
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return "IncompleteBody", NewApiError("Request body is incomplete", http.StatusBadRequest), additionalFields
	}

	if errors.Is(err, context.Canceled) {
		additionalFields["reason"] = "context_canceled"
		return "RequestCanceled", NewApiError("Request was cancelled by client", constants.StatusClientClosedRequest), additionalFields
	}

	if errors.Is(err, context.DeadlineExceeded) {
		additionalFields["reason"] = "deadline_exceeded"
		return "RequestTimeout", NewApiError("Request timeout exceeded", http.StatusRequestTimeout), additionalFields
	}

	// Unknown error
	additionalFields["full_error"] = fmt.Sprintf("%+v", err)
	return "UnknownError", NewApiError("An error occurred", http.StatusInternalServerError), additionalFields
}

// logAndRespondWithError logs the error and sends the API error response
func logAndRespondWithError(ctx *gin.Context, apiErr *ApiError, errorType string, logFields Fields, logger Logger) {
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
	ctx.Header("Content-Type", contentTypeJSON)
	ctx.AbortWithStatusJSON(apiErr.Status, apiErr)
}

// handleBasicError handles different types of errors and converts them to appropriate HTTP responses
// This version only handles platform-specific errors, not database-specific errors
func handleBasicError(ctx *gin.Context, err error, logger Logger) {
	// Build log fields with request context
	logFields := buildLogFields(ctx)
	logFields["error"] = err.Error()

	// Map error to API error and get additional fields
	errorType, apiErr, additionalFields := mapErrorToApiError(err)

	// Merge additional fields into log fields
	for k, v := range additionalFields {
		logFields[k] = v
	}

	// Log and respond
	logAndRespondWithError(ctx, apiErr, errorType, logFields, logger)
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
