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

// HTTP Domain Errors

type NotFoundError struct {
	message string
}

func (e *NotFoundError) Error() string {
	return e.message
}

func NewNotFoundError(msg string) error {
	return &NotFoundError{message: msg}
}

type UnauthorizedError struct {
	message string
}

func (e *UnauthorizedError) Error() string {
	return e.message
}

func NewUnauthorizedError(msg string) error {
	return &UnauthorizedError{message: msg}
}

type ConflictError struct {
	message string
}

func (e *ConflictError) Error() string {
	return e.message
}

func NewConflictError(msg string) error {
	return &ConflictError{message: msg}
}

type BadRequestError struct {
	message string
}

func (e *BadRequestError) Error() string {
	return e.message
}

func NewBadRequestError(msg string) error {
	return &BadRequestError{message: msg}
}

type ExternalServiceError struct {
	message string
	status  int
}

func (e *ExternalServiceError) Error() string {
	return e.message
}

func (e *ExternalServiceError) Status() int {
	return e.status
}

func NewExternalServiceError(msg string, status int) error {
	return &ExternalServiceError{message: msg, status: status}
}

type DomainError struct {
	message string
}

func (e *DomainError) Error() string {
	return e.message
}

func NewDomainError(msg string) error {
	return &DomainError{message: msg}
}
