package errors

import "fmt"

type configError struct {
	message string
}

func (e *configError) Error() string {
	return "config error: " + e.message
}

func NewConfigError(msg string) error {
	return &configError{message: msg}
}

func ErrNilLogger() error {
	return &configError{message: "logger cannot be nil"}
}

func ErrInvalidPort(port int) error {
	return &configError{message: "invalid port"}
}

func ErrInvalidMode(mode string) error {
	return &configError{message: "invalid mode: must be debug, release, or test"}
}

type RuntimeError struct {
	message string
	cause   error
}

func (e *RuntimeError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("runtime error: %s: %v", e.message, e.cause)
	}
	return fmt.Sprintf("runtime error: %s", e.message)
}

func (e *RuntimeError) Unwrap() error {
	return e.cause
}

func NewRuntimeError(msg string, cause error) error {
	return &RuntimeError{message: msg, cause: cause}
}

func ErrAlreadyStarted() error {
	return &RuntimeError{message: "platform already started"}
}

func ErrNotStarted() error {
	return &RuntimeError{message: "platform not started"}
}
