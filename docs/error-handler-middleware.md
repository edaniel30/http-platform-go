# ErrorHandler Middleware

The ErrorHandler middleware provides centralized error handling and panic recovery, converting all errors into consistent JSON responses.

## What It Does

The ErrorHandler middleware helps your application:

- **Recover from panics**: Catches panics and converts them to proper HTTP responses
- **Centralize error handling**: Single place to handle all error types consistently
- **Structure error responses**: Returns well-formatted JSON error responses
- **Log errors appropriately**: Logs errors with context, severity levels, and trace IDs
- **Handle multiple error types**: Supports platform errors, validation errors, JSON errors, and context errors

## Components

### 1. ErrorHandler Middleware

**Purpose**: Catches all errors and panics, converting them to appropriate HTTP responses.

**Enabled by default** - runs as the second middleware in the chain (after TraceID).

**How it works**:
- Sets up panic recovery with `defer recover()`
- Processes the request chain with `c.Next()`
- Checks for errors added to the context
- Handles the first error (if any)
- Logs errors with appropriate severity
- Returns structured JSON responses

### 2. Platform Error Types

The middleware handles these custom error types:

| Error Type | Status | Usage |
|------------|--------|-------|
| `NotFoundError` | 404 | Resource not found |
| `UnauthorizedError` | 401 | Authentication required |
| `ForbiddenError` | 403 | Insufficient permissions |
| `BadRequestError` | 400 | Invalid request data |
| `ConflictError` | 409 | Resource conflict |
| `UnprocessableEntityError` | 422 | Semantic errors |
| `TooManyRequestsError` | 429 | Rate limit exceeded |
| `InternalServerError` | 500 | Server-side errors |
| `ServiceUnavailableError` | 503 | Service temporarily down |
| `ExternalServiceError` | varies | External API failures |

**Example usage**:
```go
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.repo.FindByID(id)
    if err != nil {
        c.Error(httpplatform.NewNotFoundError("User not found"))
        return
    }
    c.JSON(200, user)
}
```

### 3. Automatic Error Detection

**Validation errors** (from `go-playground/validator`):
```go
type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(err) // Auto-formatted by ErrorHandler
        return
    }
}
```

**Response**:
```json
{
    "message": "Validation error",
    "error": "Bad Request",
    "status": 400,
    "cause": [
        {"field": "Email", "reason": "required"},
        {"field": "Password", "reason": "min=8"}
    ]
}
```

**Other auto-detected errors**:
- JSON syntax errors → 400 with position
- Empty body → 400
- Context cancellation → 499 (client disconnect) or 408 (timeout)

## When to Use

### Use c.Error() in handlers when:
- Business logic errors occur
- Validation fails
- External services fail
- Resources not found
- Authorization fails

### The middleware automatically handles:
- Panics (no manual recovery needed)
- JSON binding errors
- Context cancellation
- Unknown errors (converts to 500)

## Error Response Format

All errors follow this structure:

```json
{
    "message": "Human-readable error message",
    "error": "HTTP status text",
    "status": 400,
    "cause": ["Optional array of causes"]
}
```

## Usage Examples

### Example 1: Simple Error

```go
func (h *Handler) DeleteUser(c *gin.Context) {
    err := h.repo.Delete(userID)
    if err != nil {
        c.Error(httpplatform.NewInternalServerError("Failed to delete user"))
        return
    }
    c.JSON(200, gin.H{"message": "User deleted"})
}
```

### Example 2: Multiple Error Checks

```go
func (h *Handler) ProcessOrder(c *gin.Context) {
    order, err := h.repo.FindByID(orderID)
    if err != nil {
        c.Error(httpplatform.NewNotFoundError("Order not found"))
        return
    }

    if order.UserID != currentUser.ID {
        c.Error(httpplatform.NewForbiddenError("Access denied"))
        return
    }

    result, err := h.service.Process(order)
    if err != nil {
        c.Error(httpplatform.NewInternalServerError("Processing failed"))
        return
    }

    c.JSON(200, result)
}
```

## Best Practices

**1. Always use c.Error() for errors:**
```go
// ✅ Good
c.Error(httpplatform.NewBadRequestError("Invalid data"))
return

// ❌ Bad - inconsistent format
c.JSON(400, gin.H{"error": "Invalid data"})
```

**2. Use appropriate error types:**
```go
// ✅ Good
c.Error(httpplatform.NewNotFoundError("User not found"))

// ❌ Bad
c.Error(httpplatform.NewInternalServerError("User not found"))
```

**3. Return after c.Error():**
```go
// ✅ Good
if err != nil {
    c.Error(err)
    return
}

// ❌ Bad - continues executing
if err != nil {
    c.Error(err)
}
```

## Logging

**Server errors (5xx)** logged as **ERROR**:
```
ERROR Server error trace_id=abc-123 status=500 error_type=InternalServerError
```

**Client errors (4xx)** logged as **WARN**:
```
WARN Client error trace_id=abc-123 status=404 error_type=NotFoundError
```

**Panics** logged with **stack traces**:
```
ERROR Panic recovered trace_id=abc-123 panic="runtime error" stack_trace="..."
```

## HTTP Status Codes

| Status | Description |
|--------|-------------|
| 400 | Bad Request (validation, JSON errors) |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 408 | Request Timeout |
| 409 | Conflict |
| 422 | Unprocessable Entity |
| 429 | Too Many Requests |
| 499 | Client Closed Request |
| 500 | Internal Server Error |
| 503 | Service Unavailable |

## Configuration

Always enabled by default:
```go
cfg := httpplatform.DefaultConfig()
// ErrorHandler is always enabled and cannot be disabled
```

## Available Error Constructors

```go
httpplatform.NewNotFoundError("Resource not found")
httpplatform.NewUnauthorizedError("Authentication required")
httpplatform.NewForbiddenError("Access denied")
httpplatform.NewBadRequestError("Invalid input")
httpplatform.NewConflictError("Resource already exists")
httpplatform.NewUnprocessableEntityError("Invalid data structure")
httpplatform.NewTooManyRequestsError("Rate limit exceeded")
httpplatform.NewInternalServerError("Operation failed")
httpplatform.NewServiceUnavailableError("Service temporarily down")
httpplatform.NewExternalServiceError("API call failed", statusCode)
```
