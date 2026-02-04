package httpplatform

import (
	"errors"
	"testing"

	"github.com/edaniel30/http-platform-go/internal/middleware"
)

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("resource not found")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "resource not found" {
		t.Errorf("expected message 'resource not found', got '%s'", err.Error())
	}

	// Verify it returns the correct internal type using errors.As
	var notFoundErr *middleware.NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected *middleware.NotFoundError, got %T", err)
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("invalid credentials")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "invalid credentials" {
		t.Errorf("expected message 'invalid credentials', got '%s'", err.Error())
	}

	var unauthorizedErr *middleware.UnauthorizedError
	if !errors.As(err, &unauthorizedErr) {
		t.Errorf("expected *middleware.UnauthorizedError, got %T", err)
	}
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("insufficient permissions")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "insufficient permissions" {
		t.Errorf("expected message 'insufficient permissions', got '%s'", err.Error())
	}

	var forbiddenErr *middleware.ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Errorf("expected *middleware.ForbiddenError, got %T", err)
	}
}

func TestNewBadRequestError(t *testing.T) {
	err := NewBadRequestError("invalid email format")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "invalid email format" {
		t.Errorf("expected message 'invalid email format', got '%s'", err.Error())
	}

	var badRequestErr *middleware.BadRequestError
	if !errors.As(err, &badRequestErr) {
		t.Errorf("expected *middleware.BadRequestError, got %T", err)
	}
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("user already exists")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "user already exists" {
		t.Errorf("expected message 'user already exists', got '%s'", err.Error())
	}

	var conflictErr *middleware.ConflictError
	if !errors.As(err, &conflictErr) {
		t.Errorf("expected *middleware.ConflictError, got %T", err)
	}
}

func TestNewUnprocessableEntityError(t *testing.T) {
	err := NewUnprocessableEntityError("validation failed: age must be positive")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "validation failed: age must be positive" {
		t.Errorf("expected message 'validation failed: age must be positive', got '%s'", err.Error())
	}

	var unprocessableErr *middleware.UnprocessableEntityError
	if !errors.As(err, &unprocessableErr) {
		t.Errorf("expected *middleware.UnprocessableEntityError, got %T", err)
	}
}

func TestNewTooManyRequestsError(t *testing.T) {
	err := NewTooManyRequestsError("rate limit exceeded")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "rate limit exceeded" {
		t.Errorf("expected message 'rate limit exceeded', got '%s'", err.Error())
	}

	var tooManyReqErr *middleware.TooManyRequestsError
	if !errors.As(err, &tooManyReqErr) {
		t.Errorf("expected *middleware.TooManyRequestsError, got %T", err)
	}
}

func TestNewInternalServerError(t *testing.T) {
	err := NewInternalServerError("failed to process request")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "failed to process request" {
		t.Errorf("expected message 'failed to process request', got '%s'", err.Error())
	}

	var internalErr *middleware.InternalServerError
	if !errors.As(err, &internalErr) {
		t.Errorf("expected *middleware.InternalServerError, got %T", err)
	}
}

func TestNewServiceUnavailableError(t *testing.T) {
	err := NewServiceUnavailableError("service under maintenance")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "service under maintenance" {
		t.Errorf("expected message 'service under maintenance', got '%s'", err.Error())
	}

	var serviceErr *middleware.ServiceUnavailableError
	if !errors.As(err, &serviceErr) {
		t.Errorf("expected *middleware.ServiceUnavailableError, got %T", err)
	}
}

func TestNewExternalServiceError(t *testing.T) {
	err := NewExternalServiceError("payment API failed", 502)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "payment API failed" {
		t.Errorf("expected message 'payment API failed', got '%s'", err.Error())
	}

	var extErr *middleware.ExternalServiceError
	if !errors.As(err, &extErr) {
		t.Fatalf("expected *middleware.ExternalServiceError, got %T", err)
	}

	if extErr.Status() != 502 {
		t.Errorf("expected status code 502, got %d", extErr.Status())
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		errFunc  func(string) error
		message  string
		expected string
	}{
		{"NotFound", NewNotFoundError, "test message", "test message"},
		{"Unauthorized", NewUnauthorizedError, "auth failed", "auth failed"},
		{"Forbidden", NewForbiddenError, "no access", "no access"},
		{"BadRequest", NewBadRequestError, "bad data", "bad data"},
		{"Conflict", NewConflictError, "duplicate", "duplicate"},
		{"UnprocessableEntity", NewUnprocessableEntityError, "invalid", "invalid"},
		{"TooManyRequests", NewTooManyRequestsError, "rate limited", "rate limited"},
		{"InternalServer", NewInternalServerError, "server error", "server error"},
		{"ServiceUnavailable", NewServiceUnavailableError, "down", "down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errFunc(tt.message)
			if err.Error() != tt.expected {
				t.Errorf("expected message '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}
