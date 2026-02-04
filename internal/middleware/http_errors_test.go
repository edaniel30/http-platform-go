package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToScreamingSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NotFoundError",
			input:    "NotFoundError",
			expected: "NOT_FOUND",
		},
		{
			name:     "InternalServerError",
			input:    "InternalServerError",
			expected: "INTERNAL_SERVER",
		},
		{
			name:     "BadRequestError",
			input:    "BadRequestError",
			expected: "BAD_REQUEST",
		},
		{
			name:     "UnauthorizedError",
			input:    "UnauthorizedError",
			expected: "UNAUTHORIZED",
		},
		{
			name:     "ServiceUnavailableError",
			input:    "ServiceUnavailableError",
			expected: "SERVICE_UNAVAILABLE",
		},
		{
			name:     "UnprocessableEntityError",
			input:    "UnprocessableEntityError",
			expected: "UNPROCESSABLE_ENTITY",
		},
		{
			name:     "already SCREAMING_SNAKE_CASE",
			input:    "NOT_FOUND",
			expected: "NOT_FOUND",
		},
		{
			name:     "without Error suffix",
			input:    "NotFound",
			expected: "NOT_FOUND",
		},
		{
			name:     "single word",
			input:    "Error",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toScreamingSnakeCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewApiError(t *testing.T) {
	t.Run("basic error without cause", func(t *testing.T) {
		err := NewApiError("Resource not found", 404, "NotFoundError")

		assert.Equal(t, "Resource not found", err.Message)
		assert.Equal(t, 404, err.Status)
		assert.Equal(t, "NOT_FOUND", err.Code) // Converted to SCREAMING_SNAKE_CASE
		assert.Empty(t, err.Cause)
	})

	t.Run("error with single cause", func(t *testing.T) {
		err := NewApiError("Validation failed", 400, "ValidationError", "field is required")

		assert.Equal(t, "Validation failed", err.Message)
		assert.Equal(t, 400, err.Status)
		assert.Equal(t, "VALIDATION", err.Code)
		assert.Len(t, err.Cause, 1)
		assert.Equal(t, "field is required", err.Cause[0])
	})

	t.Run("error with multiple causes", func(t *testing.T) {
		causes := []any{"error1", "error2", "error3"}
		err := NewApiError("Multiple errors", 500, "InternalServerError", causes...)

		assert.Equal(t, "Multiple errors", err.Message)
		assert.Equal(t, 500, err.Status)
		assert.Equal(t, "INTERNAL_SERVER", err.Code)
		assert.Len(t, err.Cause, 3)
	})

	t.Run("code already in SCREAMING_SNAKE_CASE", func(t *testing.T) {
		err := NewApiError("Error", 400, "ALREADY_FORMATTED")

		assert.Equal(t, "ALREADY_FORMATTED", err.Code)
	})
}

func TestHTTPErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "NotFoundError",
			err:      &NotFoundError{Message: "user not found"},
			expected: "user not found",
		},
		{
			name:     "UnauthorizedError",
			err:      &UnauthorizedError{Message: "invalid credentials"},
			expected: "invalid credentials",
		},
		{
			name:     "ForbiddenError",
			err:      &ForbiddenError{Message: "access denied"},
			expected: "access denied",
		},
		{
			name:     "BadRequestError",
			err:      &BadRequestError{Message: "invalid input"},
			expected: "invalid input",
		},
		{
			name:     "ConflictError",
			err:      &ConflictError{Message: "resource already exists"},
			expected: "resource already exists",
		},
		{
			name:     "UnprocessableEntityError",
			err:      &UnprocessableEntityError{Message: "validation failed"},
			expected: "validation failed",
		},
		{
			name:     "TooManyRequestsError",
			err:      &TooManyRequestsError{Message: "rate limit exceeded"},
			expected: "rate limit exceeded",
		},
		{
			name:     "InternalServerError",
			err:      &InternalServerError{Message: "internal error"},
			expected: "internal error",
		},
		{
			name:     "ServiceUnavailableError",
			err:      &ServiceUnavailableError{Message: "service unavailable"},
			expected: "service unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestExternalServiceError(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		err := &ExternalServiceError{
			Message:    "payment service failed",
			StatusCode: 502,
		}
		assert.Equal(t, "payment service failed", err.Error())
	})

	t.Run("Status method", func(t *testing.T) {
		err := &ExternalServiceError{
			Message:    "database unavailable",
			StatusCode: 503,
		}
		assert.Equal(t, 503, err.Status())
	})
}

// TestHTTPErrorConstructors removed - constructors are now public in httpplatform package
// and tested in http_errors_test.go at the root level.
