package middleware

import (
	"regexp"
	"strings"
)

// HTTP Domain Errors
// These are convenience error types for common HTTP status codes.
// They help maintain consistency across services and reduce boilerplate code.
// Use these in your handlers to return standardized error responses.

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

// NotFoundError represents an HTTP 404 Not Found error.
// Use this when a requested resource cannot be found.
//
// Example:
//
//	return middleware.NewNotFoundError("user not found")
type NotFoundError struct {
	Message string // Human-readable error message
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// UnauthorizedError represents an HTTP 401 Unauthorized error.
// Use this when authentication is required but not provided or invalid.
//
// Example:
//
//	return middleware.NewUnauthorizedError("invalid credentials")
type UnauthorizedError struct {
	Message string // Human-readable error message
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

// ForbiddenError represents an HTTP 403 Forbidden error.
// Use this when the user is authenticated but lacks permission.
//
// Example:
//
//	return middleware.NewForbiddenError("insufficient permissions")
type ForbiddenError struct {
	Message string // Human-readable error message
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// BadRequestError represents an HTTP 400 Bad Request error.
// Use this when the client sends invalid or malformed data.
//
// Example:
//
//	return middleware.NewBadRequestError("invalid email format")
type BadRequestError struct {
	Message string // Human-readable error message
}

func (e *BadRequestError) Error() string {
	return e.Message
}

// ConflictError represents an HTTP 409 Conflict error.
// Use this when the request conflicts with the current state (e.g., duplicate resource).
//
// Example:
//
//	return middleware.NewConflictError("user already exists")
type ConflictError struct {
	Message string // Human-readable error message
}

func (e *ConflictError) Error() string {
	return e.Message
}

// UnprocessableEntityError represents an HTTP 422 Unprocessable Entity error.
// Use this when the request is well-formed but contains semantic errors.
//
// Example:
//
//	return middleware.NewUnprocessableEntityError("validation failed: age must be positive")
type UnprocessableEntityError struct {
	Message string // Human-readable error message
}

func (e *UnprocessableEntityError) Error() string {
	return e.Message
}

// TooManyRequestsError represents an HTTP 429 Too Many Requests error.
// Use this when rate limiting is triggered.
//
// Example:
//
//	return middleware.NewTooManyRequestsError("rate limit exceeded, try again later")
type TooManyRequestsError struct {
	Message string // Human-readable error message
}

func (e *TooManyRequestsError) Error() string {
	return e.Message
}

// InternalServerError represents an HTTP 500 Internal Server Error.
// Use this when an unexpected error occurs during processing.
//
// Example:
//
//	return middleware.NewInternalServerError("failed to process request")
type InternalServerError struct {
	Message string // Human-readable error message
}

func (e *InternalServerError) Error() string {
	return e.Message
}

// ServiceUnavailableError represents an HTTP 503 Service Unavailable error.
// Use this when the service is temporarily unavailable (maintenance, overload, etc.).
//
// Example:
//
//	return middleware.NewServiceUnavailableError("service under maintenance")
type ServiceUnavailableError struct {
	Message string // Human-readable error message
}

func (e *ServiceUnavailableError) Error() string {
	return e.Message
}

// ExternalServiceError represents an error from an external service/API call.
// The StatusCode field contains the HTTP status code from the external service.
//
// Example:
//
//	return middleware.NewExternalServiceError("payment API failed", 502)
type ExternalServiceError struct {
	Message    string // Human-readable error message
	StatusCode int    // HTTP status code from the external service
}

func (e *ExternalServiceError) Error() string {
	return e.Message
}

// Status returns the HTTP status code from the external service.
func (e *ExternalServiceError) Status() int {
	return e.StatusCode
}

// Public constructors for HTTP domain errors
// These are exported so users can create these errors in their handlers.

// NewNotFoundError creates a new NotFoundError with the given message.
func NewNotFoundError(msg string) error {
	return &NotFoundError{Message: msg}
}

// NewUnauthorizedError creates a new UnauthorizedError with the given message.
func NewUnauthorizedError(msg string) error {
	return &UnauthorizedError{Message: msg}
}

// NewForbiddenError creates a new ForbiddenError with the given message.
func NewForbiddenError(msg string) error {
	return &ForbiddenError{Message: msg}
}

// NewBadRequestError creates a new BadRequestError with the given message.
func NewBadRequestError(msg string) error {
	return &BadRequestError{Message: msg}
}

// NewConflictError creates a new ConflictError with the given message.
func NewConflictError(msg string) error {
	return &ConflictError{Message: msg}
}

// NewUnprocessableEntityError creates a new UnprocessableEntityError with the given message.
func NewUnprocessableEntityError(msg string) error {
	return &UnprocessableEntityError{Message: msg}
}

// NewTooManyRequestsError creates a new TooManyRequestsError with the given message.
func NewTooManyRequestsError(msg string) error {
	return &TooManyRequestsError{Message: msg}
}

// NewInternalServerError creates a new InternalServerError with the given message.
func NewInternalServerError(msg string) error {
	return &InternalServerError{Message: msg}
}

// NewServiceUnavailableError creates a new ServiceUnavailableError with the given message.
func NewServiceUnavailableError(msg string) error {
	return &ServiceUnavailableError{Message: msg}
}

// NewExternalServiceError creates a new ExternalServiceError with the given message and status code.
func NewExternalServiceError(msg string, status int) error {
	return &ExternalServiceError{Message: msg, StatusCode: status}
}
