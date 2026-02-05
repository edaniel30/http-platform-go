package middleware

import (
	"regexp"
	"strings"
)

// ApiError represents a structured API error response
type ApiError struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Cause   []any  `json:"cause,omitempty"`
}

// NewApiError creates a new ApiError with the given message, status code, error code, and optional causes
func NewApiError(message string, status int, code string, cause ...any) *ApiError {
	return &ApiError{
		Message: message,
		Status:  status,
		Code:    toScreamingSnakeCase(code),
		Cause:   cause,
	}
}

// toScreamingSnakeCase converts a PascalCase string to SCREAMING_SNAKE_CASE
// Example: "NotFoundError" -> "NOT_FOUND", "InternalServerError" -> "INTERNAL_SERVER_ERROR"
func toScreamingSnakeCase(s string) string {
	// Remove "Error" suffix if present
	s = strings.TrimSuffix(s, "Error")

	// Insert underscore before each uppercase letter (except the first one)
	// This regex matches uppercase letters that are followed by lowercase letters or preceded by lowercase letters
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	s = re.ReplaceAllString(s, "${1}_${2}")

	// Convert to uppercase
	return strings.ToUpper(s)
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
