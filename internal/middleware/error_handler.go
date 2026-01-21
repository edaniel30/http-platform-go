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
		ctx.Header("Content-Type", constants.ContentTypeJSON)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			NewApiError("Internal server error panic", http.StatusInternalServerError, "InternalServerError"))
	}
}

// buildLogFields creates base log fields with request context and trace ID
func buildLogFields(ctx *gin.Context) Fields {
	logFields := Fields{}
	addBaseRequestFields(logFields, ctx)
	return logFields
}

// handleBasicError handles different types of errors and converts them to appropriate HTTP responses
// This version only handles platform-specific errors, not database-specific errors
func handleBasicError(ctx *gin.Context, err error, logger Logger) {
	// Build log fields with request context
	logFields := buildLogFields(ctx)
	logFields["error"] = err.Error()

	// Map error to API error and get additional fields
	apiErr, additionalFields := mapErrorToApiError(err)

	// Merge additional fields into log fields
	for k, v := range additionalFields {
		logFields[k] = v
	}

	// Log and respond
	logAndRespondWithError(ctx, apiErr, logFields, logger)
}

// logAndRespondWithError logs the error and sends the API error response
func logAndRespondWithError(ctx *gin.Context, apiErr *ApiError, logFields Fields, logger Logger) {
	// Add error code and status to log
	logFields["status"] = apiErr.Status
	logFields["code"] = apiErr.Code

	// Log based on severity
	reqCtx := ctx.Request.Context()
	if apiErr.Status >= 500 {
		logger.Error(reqCtx, "Server error", logFields)
	} else {
		logger.Warn(reqCtx, "Client error", logFields)
	}

	// Set Content-Type header before sending response
	ctx.Header("Content-Type", constants.ContentTypeJSON)
	ctx.AbortWithStatusJSON(apiErr.Status, apiErr)
}

// mapStandardError handles standard library errors using errors.Is
func mapStandardError(err error, additionalFields Fields) (*ApiError, Fields) {
	if errors.Is(err, io.EOF) {
		return NewApiError("Request body is empty", http.StatusBadRequest, "EmptyBody"), additionalFields
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		return NewApiError("Request body is incomplete", http.StatusBadRequest, "IncompleteBody"), additionalFields
	}

	if errors.Is(err, context.Canceled) {
		additionalFields["reason"] = "context_canceled"
		return NewApiError("Request was cancelled by client", constants.StatusClientClosedRequest, "RequestCanceled"), additionalFields
	}

	if errors.Is(err, context.DeadlineExceeded) {
		additionalFields["reason"] = "deadline_exceeded"
		return NewApiError("Request timeout exceeded", http.StatusRequestTimeout, "RequestTimeout"), additionalFields
	}

	// Unknown error
	additionalFields["full_error"] = fmt.Sprintf("%+v", err)
	return NewApiError("An error occurred", http.StatusInternalServerError, "UnknownError"), additionalFields
}

// mapErrorToApiError maps different error types to API errors with metadata
// Returns: errorType, apiError, and additional log fields
func mapErrorToApiError(err error) (*ApiError, Fields) {
	additionalFields := Fields{}

	switch e := err.(type) {
	case *NotFoundError:
		return NewApiError(e.Error(), http.StatusNotFound, "NotFoundError"), additionalFields

	case *UnauthorizedError:
		return NewApiError(e.Error(), http.StatusUnauthorized, "UnauthorizedError"), additionalFields

	case *ConflictError:
		return NewApiError(e.Error(), http.StatusConflict, "ConflictError"), additionalFields

	case *ExternalServiceError:
		additionalFields["external_status"] = e.Status()
		return NewApiError(e.Error(), e.Status(), "ExternalServiceError"), additionalFields

	case *BadRequestError:
		return NewApiError(e.Error(), http.StatusBadRequest, "BadRequestError"), additionalFields

	case *ForbiddenError:
		return NewApiError(e.Error(), http.StatusForbidden, "ForbiddenError"), additionalFields

	case *UnprocessableEntityError:
		return NewApiError(e.Error(), http.StatusUnprocessableEntity, "UnprocessableEntityError"), additionalFields

	case *TooManyRequestsError:
		return NewApiError(e.Error(), http.StatusTooManyRequests, "TooManyRequestsError"), additionalFields

	case *InternalServerError:
		return NewApiError(e.Error(), http.StatusInternalServerError, "InternalServerError"), additionalFields

	case *ServiceUnavailableError:
		return NewApiError(e.Error(), http.StatusServiceUnavailable, "ServiceUnavailableError"), additionalFields

	case *json.UnmarshalTypeError:
		additionalFields["field"] = e.Field
		additionalFields["expected_type"] = e.Type.String()
		apiErr := NewApiError(
			fmt.Sprintf("Invalid type for field '%s', expected %s but got %s", e.Field, e.Type.String(), e.Value),
			http.StatusBadRequest,
			"UnmarshalTypeError",
		)
		return apiErr, additionalFields

	case validator.ValidationErrors:
		validationErrs := descriptiveValidationErrors(e)
		additionalFields["validation_errors"] = validationErrs
		return NewApiError("Validation error", http.StatusBadRequest, "ValidationError", validationErrs), additionalFields

	case *json.SyntaxError:
		additionalFields["offset"] = e.Offset
		additionalFields["syntax_error"] = e.Error()
		apiErr := NewApiError(fmt.Sprintf("Invalid JSON syntax at position %d", e.Offset), http.StatusBadRequest, "JSONSyntaxError")
		return apiErr, additionalFields

	default:
		return mapStandardError(err, additionalFields)
	}
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
