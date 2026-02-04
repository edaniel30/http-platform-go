package httpplatform

import "github.com/edaniel30/http-platform-go/internal/middleware"

// HTTP Error Constructors
// These are convenience functions for creating HTTP domain errors in your handlers.
// They provide type-safe error construction with automatic HTTP status code mapping.
//
// Example usage in a handler:
//
//	func GetUser(c *gin.Context) {
//	    user, err := userService.Find(id)
//	    if err != nil {
//	        c.Error(httpplatform.NewNotFoundError("user not found"))
//	        return
//	    }
//	    c.JSON(200, user)
//	}
//
// The ErrorHandler middleware will automatically convert these errors
// to appropriate HTTP responses with structured JSON format.

// NewNotFoundError creates an HTTP 404 Not Found error.
// Use this when a requested resource cannot be found.
func NewNotFoundError(msg string) error {
	return &middleware.NotFoundError{Message: msg}
}

// NewUnauthorizedError creates an HTTP 401 Unauthorized error.
// Use this when authentication is required but not provided or invalid.
func NewUnauthorizedError(msg string) error {
	return &middleware.UnauthorizedError{Message: msg}
}

// NewForbiddenError creates an HTTP 403 Forbidden error.
// Use this when the user is authenticated but lacks permission.
func NewForbiddenError(msg string) error {
	return &middleware.ForbiddenError{Message: msg}
}

// NewBadRequestError creates an HTTP 400 Bad Request error.
// Use this when the client sends invalid or malformed data.
func NewBadRequestError(msg string) error {
	return &middleware.BadRequestError{Message: msg}
}

// NewConflictError creates an HTTP 409 Conflict error.
// Use this when the request conflicts with the current state (e.g., duplicate resource).
func NewConflictError(msg string) error {
	return &middleware.ConflictError{Message: msg}
}

// NewUnprocessableEntityError creates an HTTP 422 Unprocessable Entity error.
// Use this when the request is well-formed but contains semantic errors.
func NewUnprocessableEntityError(msg string) error {
	return &middleware.UnprocessableEntityError{Message: msg}
}

// NewTooManyRequestsError creates an HTTP 429 Too Many Requests error.
// Use this when rate limiting is triggered.
func NewTooManyRequestsError(msg string) error {
	return &middleware.TooManyRequestsError{Message: msg}
}

// NewInternalServerError creates an HTTP 500 Internal Server Error.
// Use this when an unexpected error occurs during processing.
func NewInternalServerError(msg string) error {
	return &middleware.InternalServerError{Message: msg}
}

// NewServiceUnavailableError creates an HTTP 503 Service Unavailable error.
// Use this when the service is temporarily unavailable (maintenance, overload, etc.).
func NewServiceUnavailableError(msg string) error {
	return &middleware.ServiceUnavailableError{Message: msg}
}

// NewExternalServiceError creates an error representing a failure from an external service.
// The status parameter should be the HTTP status code returned by the external service.
func NewExternalServiceError(msg string, status int) error {
	return &middleware.ExternalServiceError{Message: msg, StatusCode: status}
}
