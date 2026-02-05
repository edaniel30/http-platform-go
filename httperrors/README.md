# HTTP Errors Package

This package provides HTTP error types that can be used throughout your application to return consistent error responses.

## Installation

```go
import "github.com/edaniel30/http-platform-go/httperrors"
```

## Available Error Types

| Error Type | HTTP Status | Use Case |
|------------|-------------|----------|
| `NotFoundError` | 404 | Resource not found |
| `UnauthorizedError` | 401 | Authentication required or failed |
| `ForbiddenError` | 403 | Authenticated but lacks permission |
| `BadRequestError` | 400 | Invalid or malformed request |
| `ConflictError` | 409 | Resource conflict (e.g., duplicate) |
| `UnprocessableEntityError` | 422 | Semantic validation error |
| `TooManyRequestsError` | 429 | Rate limit exceeded |
| `InternalServerError` | 500 | Unexpected server error |
| `ServiceUnavailableError` | 503 | Service temporarily unavailable |
| `ExternalServiceError` | Custom | External service failure |

## Usage

### Basic Usage

```go
import "github.com/edaniel30/http-platform-go/httperrors"

func GetUser(c *gin.Context) {
    user, err := userService.Find(id)
    if err != nil {
        c.Error(httperrors.NewNotFoundError("user not found"))
        return
    }
    c.JSON(200, user)
}
```

### Type Assertion with errors.As()

You can use `errors.As()` to check for specific error types:

```go
import (
    "errors"
    "github.com/edaniel30/http-platform-go/httperrors"
)

func ProcessBilling(fileEvent FileEvent) error {
    err := billingService.Create(fileEvent)
    if err != nil {
        var conflictErr *httperrors.ConflictError
        if errors.As(err, &conflictErr) {
            // Handle duplicate billing gracefully
            log.Info("Billing already exists, skipping", "message", conflictErr.Message)
            return nil
        }
        return err
    }
    return nil
}
```

### Wrapping Errors

You can wrap these errors with additional context:

```go
import (
    "fmt"
    "github.com/edaniel30/http-platform-go/httperrors"
)

func DeleteUser(id string) error {
    if err := validateUserExists(id); err != nil {
        return fmt.Errorf("delete user: %w",
            httperrors.NewNotFoundError("user not found"))
    }
    // ... deletion logic
}

// Later, you can still use errors.As() to detect the NotFoundError
```

### External Service Errors

For errors from external APIs, use `ExternalServiceError` to preserve the original status code:

```go
resp, err := paymentAPI.Charge(amount)
if err != nil {
    return httperrors.NewExternalServiceError(
        "payment service failed",
        resp.StatusCode)
}
```

## Integration with ErrorHandler Middleware

The platform's ErrorHandler middleware automatically converts these errors to appropriate HTTP responses:

```json
{
  "message": "user not found",
  "status": 404,
  "code": "NOT_FOUND"
}
```

## Creating Custom Error Types

If you need custom error types, you can create them following the same pattern:

```go
type CustomError struct {
    Message string
}

func (e *CustomError) Error() string {
    return e.Message
}

func NewCustomError(msg string) error {
    return &CustomError{Message: msg}
}
```

Then add handling in your error handler middleware for automatic HTTP response mapping.
