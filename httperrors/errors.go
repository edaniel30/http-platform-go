package httperrors

// HTTP Error Types
// These are convenience types for creating HTTP domain errors in your handlers.
// They provide type-safe error construction with automatic HTTP status code mapping.
//
// Example usage in a handler:
//
//	func GetUser(c *gin.Context) {
//	    user, err := userService.Find(id)
//	    if err != nil {
//	        // Check if it's a specific error type
//	        var notFoundErr *NotFoundError
//	        if errors.As(err, &notFoundErr) {
//	            c.Error(notFoundErr)
//	            return
//	        }
//	        c.Error(NewInternalServerError("unexpected error"))
//	        return
//	    }
//	    c.JSON(200, user)
//	}
//
// The ErrorHandler middleware will automatically convert these errors
// to appropriate HTTP responses with structured JSON format.

// NotFoundError represents an HTTP 404 Not Found error.
// Use this when a requested resource cannot be found.
//
// Example:
//
//	return NewNotFoundError("user not found")
//
// Type assertion example:
//
//	var notFoundErr *NotFoundError
//	if errors.As(err, &notFoundErr) {
//	    // Handle not found case
//	}
type NotFoundError struct {
	Message string // Human-readable error message
}

func (e *NotFoundError) Error() string {
	return e.Message
}

// NewNotFoundError creates an HTTP 404 Not Found error.
func NewNotFoundError(msg string) error {
	return &NotFoundError{Message: msg}
}

// UnauthorizedError represents an HTTP 401 Unauthorized error.
// Use this when authentication is required but not provided or invalid.
type UnauthorizedError struct {
	Message string // Human-readable error message
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

// NewUnauthorizedError creates an HTTP 401 Unauthorized error.
func NewUnauthorizedError(msg string) error {
	return &UnauthorizedError{Message: msg}
}

// ForbiddenError represents an HTTP 403 Forbidden error.
// Use this when the user is authenticated but lacks permission.
type ForbiddenError struct {
	Message string // Human-readable error message
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

// NewForbiddenError creates an HTTP 403 Forbidden error.
func NewForbiddenError(msg string) error {
	return &ForbiddenError{Message: msg}
}

// BadRequestError represents an HTTP 400 Bad Request error.
// Use this when the client sends invalid or malformed data.
type BadRequestError struct {
	Message string // Human-readable error message
}

func (e *BadRequestError) Error() string {
	return e.Message
}

// NewBadRequestError creates an HTTP 400 Bad Request error.
func NewBadRequestError(msg string) error {
	return &BadRequestError{Message: msg}
}

// ConflictError represents an HTTP 409 Conflict error.
// Use this when the request conflicts with the current state (e.g., duplicate resource).
//
// Example:
//
//	var conflictErr *ConflictError
//	if errors.As(err, &conflictErr) {
//	    log.Info("Resource already exists", "message", conflictErr.Message)
//	    return nil
//	}
type ConflictError struct {
	Message string // Human-readable error message
}

func (e *ConflictError) Error() string {
	return e.Message
}

// NewConflictError creates an HTTP 409 Conflict error.
func NewConflictError(msg string) error {
	return &ConflictError{Message: msg}
}

// UnprocessableEntityError represents an HTTP 422 Unprocessable Entity error.
// Use this when the request is well-formed but contains semantic errors.
type UnprocessableEntityError struct {
	Message string // Human-readable error message
}

func (e *UnprocessableEntityError) Error() string {
	return e.Message
}

// NewUnprocessableEntityError creates an HTTP 422 Unprocessable Entity error.
func NewUnprocessableEntityError(msg string) error {
	return &UnprocessableEntityError{Message: msg}
}

// TooManyRequestsError represents an HTTP 429 Too Many Requests error.
// Use this when rate limiting is triggered.
type TooManyRequestsError struct {
	Message string // Human-readable error message
}

func (e *TooManyRequestsError) Error() string {
	return e.Message
}

// NewTooManyRequestsError creates an HTTP 429 Too Many Requests error.
func NewTooManyRequestsError(msg string) error {
	return &TooManyRequestsError{Message: msg}
}

// InternalServerError represents an HTTP 500 Internal Server Error.
// Use this when an unexpected error occurs during processing.
type InternalServerError struct {
	Message string // Human-readable error message
}

func (e *InternalServerError) Error() string {
	return e.Message
}

// NewInternalServerError creates an HTTP 500 Internal Server Error.
func NewInternalServerError(msg string) error {
	return &InternalServerError{Message: msg}
}

// ServiceUnavailableError represents an HTTP 503 Service Unavailable error.
// Use this when the service is temporarily unavailable (maintenance, overload, etc.).
type ServiceUnavailableError struct {
	Message string // Human-readable error message
}

func (e *ServiceUnavailableError) Error() string {
	return e.Message
}

// NewServiceUnavailableError creates an HTTP 503 Service Unavailable error.
func NewServiceUnavailableError(msg string) error {
	return &ServiceUnavailableError{Message: msg}
}

// ExternalServiceError represents an error from an external service/API call.
// The StatusCode field contains the HTTP status code from the external service.
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

// NewExternalServiceError creates an error representing a failure from an external service.
// The status parameter should be the HTTP status code returned by the external service.
func NewExternalServiceError(msg string, status int) error {
	return &ExternalServiceError{Message: msg, StatusCode: status}
}
