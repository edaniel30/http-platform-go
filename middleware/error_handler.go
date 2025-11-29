package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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

// ErrorHandler creates a middleware that handles errors and converts them to appropriate HTTP responses
// This middleware should be added to the middleware chain to handle errors consistently across the application
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

		// Setup panic recovery
		defer func() {
			if err := recover(); err != nil {
				handlePanic(c, err)
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
func handlePanic(ctx *gin.Context, err interface{}) {
	switch er := err.(type) {
	case error:
		handleError(ctx, er)
	default:
		log.Printf("[ErrorHandler] Recovered from panic: %v", err)
		ctx.AbortWithStatusJSON(
			http.StatusInternalServerError,
			NewApiError("Internal server error panic", http.StatusInternalServerError))
	}
}

// handleError handles different types of errors and converts them to appropriate HTTP responses
func handleError(ctx *gin.Context, err error) {
	var apiErr *ApiError

	switch error := err.(type) {
	case *platformErrors.NotFoundError:
		apiErr = NewApiError(error.Error(), http.StatusNotFound)

	case *platformErrors.UnauthorizedError:
		apiErr = NewApiError(error.Error(), http.StatusUnauthorized)

	case validator.ValidationErrors:
		apiErr = NewApiError("Validation error", http.StatusBadRequest, descriptiveValidationErrors(error))

	case *platformErrors.DomainError:
		apiErr = NewApiError(error.Error(), http.StatusBadRequest)

	case *platformErrors.ConflictError:
		apiErr = NewApiError(error.Error(), http.StatusConflict)

	case *platformErrors.ExternalServiceError:
		apiErr = NewApiError(error.Error(), error.Status())

	case *platformErrors.BadRequestError:
		apiErr = NewApiError(error.Error(), http.StatusBadRequest)

	case *json.UnmarshalTypeError:
		apiErr = NewApiError(
			fmt.Sprintf("Invalid type for field '%s', expected %s but got %s",
				error.Field, error.Type.String(), error.Value),
			http.StatusBadRequest,
		)

	default:
		// Generic error
		log.Printf("[ErrorHandler] Error: %v", error)
		apiErr = NewApiError("An error occurred", http.StatusInternalServerError)
	}

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
