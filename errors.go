package httpplatform

import "fmt"

// Public Error Types
// These types are exported so users can use errors.As() to inspect them
// and access their fields for better error handling.

// ConfigError represents a configuration validation error.
// Users can access the Field and Message to understand what configuration is invalid.
type ConfigError struct {
	Field   string // The configuration field that caused the error (optional)
	Message string // Human-readable error message
}

func (e *ConfigError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("httpplatform: config error [%s]: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("httpplatform: config error: %s", e.Message)
}

// RuntimeError represents an error that occurred during platform runtime operations.
// The Cause field contains the underlying error if any.
//
// Example:
//
//	var runtimeErr *httpplatform.RuntimeError
//	if errors.As(err, &runtimeErr) {
//		fmt.Printf("Runtime error: %s\n", runtimeErr.Message)
//		if runtimeErr.Cause != nil {
//			fmt.Printf("Caused by: %v\n", runtimeErr.Cause)
//		}
//	}
type RuntimeError struct {
	Message string // Human-readable error message
	Cause   error  // The underlying error that caused this error (optional)
}

func (e *RuntimeError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("httpplatform: runtime error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("httpplatform: runtime error: %s", e.Message)
}

func (e *RuntimeError) Unwrap() error {
	return e.Cause
}

// Internal constructor functions for library errors
// These are not exported as they're only used internally by the library.

// newConfigFieldError creates a configuration error with a specific field.
func newConfigFieldError(field, message string) error {
	return &ConfigError{Field: field, Message: message}
}

// newRuntimeError creates a runtime error with a specific message and cause.
func newRuntimeError(message string, cause error) error {
	return &RuntimeError{Message: message, Cause: cause}
}
