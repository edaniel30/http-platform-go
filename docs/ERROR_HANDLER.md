# Error Handler

Centralizes error handling with consistent JSON responses and automatic panic recovery.

## Quick Start

**Always enabled** - automatically handles all errors and panics:

```go
func (h *Handler) GetUser(c *gin.Context) {
    user, err := h.repo.FindByID(id)
    if err != nil {
        c.Error(middleware.NewNotFoundError("User not found"))
        return
    }
    c.JSON(200, user)
}
```

## Error Types

| Constructor | Status | Use Case |
|-------------|--------|----------|
| `NewNotFoundError()` | 404 | Resource doesn't exist |
| `NewBadRequestError()` | 400 | Invalid input data |
| `NewUnauthorizedError()` | 401 | Missing/invalid auth |
| `NewForbiddenError()` | 403 | Insufficient permissions |
| `NewConflictError()` | 409 | Duplicate resource |
| `NewUnprocessableEntityError()` | 422 | Semantic validation errors |
| `NewTooManyRequestsError()` | 429 | Rate limit exceeded |
| `NewInternalServerError()` | 500 | Server-side errors |
| `NewServiceUnavailableError()` | 503 | Temporary unavailability |
| `NewExternalServiceError(msg, code)` | varies | External API failures |

## Response Format

All errors return consistent JSON:

```json
{
    "message": "User not found",
    "status": 404,
    "code": "NOT_FOUND",
    "cause": []
}
```

With validation errors:

```json
{
    "message": "Validation error",
    "status": 400,
    "code": "VALIDATION",
    "cause": [
        {"field": "Email", "reason": "required"},
        {"field": "Password", "reason": "min=8"}
    ]
}
```

## Automatic Error Detection

### Validation Errors (go-playground/validator)

```go
type CreateUserRequest struct {
    Email    string `binding:"required,email"`
    Password string `binding:"required,min=8"`
}

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.Error(err) // Auto-formatted as 400 with field details
        return
    }
}
```

### Other Auto-Handled Errors

| Error Type | Status | Example |
|------------|--------|---------|
| JSON syntax error | 400 | `{"key": invalid}` |
| Empty request body | 400 | `io.EOF` |
| Malformed JSON | 400 | `json.UnmarshalTypeError` |
| Client disconnect | 499 | `context.Canceled` |
| Request timeout | 408 | `context.DeadlineExceeded` |
| Panic (any type) | 500 | With stack trace |

## Common Patterns

### Resource Not Found

```go
user, err := repo.FindByID(id)
if err != nil {
    c.Error(middleware.NewNotFoundError("User not found"))
    return
}
```

### Authorization Check

```go
if order.UserID != currentUser.ID {
    c.Error(middleware.NewForbiddenError("Access denied"))
    return
}
```

### External Service Failure

```go
resp, err := paymentAPI.Charge(amount)
if err != nil {
    c.Error(middleware.NewExternalServiceError("Payment failed", 502))
    return
}
```

### Complex Validation

```go
if amount < 0 {
    c.Error(middleware.NewUnprocessableEntityError("Amount must be positive"))
    return
}
```

### Multiple Error Checks

```go
func (h *Handler) ProcessOrder(c *gin.Context) {
    // 1. Find resource
    order, err := h.repo.FindByID(orderID)
    if err != nil {
        c.Error(middleware.NewNotFoundError("Order not found"))
        return
    }

    // 2. Check permissions
    if order.UserID != currentUser.ID {
        c.Error(middleware.NewForbiddenError("Access denied"))
        return
    }

    // 3. Process
    result, err := h.service.Process(order)
    if err != nil {
        c.Error(middleware.NewInternalServerError("Processing failed"))
        return
    }

    c.JSON(200, result)
}
```

## Logging

Errors are automatically logged with appropriate severity:

**Server errors (5xx)** → `ERROR` level:
```
ERROR: Server error status=500 code=INTERNAL_SERVER trace_id=abc-123
```

**Client errors (4xx)** → `WARN` level:
```
WARN: Client error status=404 code=NOT_FOUND trace_id=abc-123
```

**Panics** → `ERROR` with stack trace:
```
ERROR: Panic recovered panic="nil pointer" stack_trace="..." trace_id=abc-123
```

## Best Practices

### ✅ Do

```go
// Use appropriate error types
c.Error(middleware.NewNotFoundError("User not found"))

// Always return after error
if err != nil {
    c.Error(err)
    return // Important!
}

// Provide helpful messages
c.Error(middleware.NewBadRequestError("Email format is invalid"))
```

### ❌ Don't

```go
// Don't use wrong status codes
c.Error(middleware.NewInternalServerError("User not found")) // Wrong!

// Don't forget to return
if err != nil {
    c.Error(err)
    // Missing return - handler continues!
}

// Don't use direct JSON responses
c.JSON(400, gin.H{"error": "bad request"}) // Inconsistent format

// Don't swallow errors
if err != nil {
    log.Println(err) // Error not sent to client
}
```

## Error Code Format

Error codes are automatically converted to SCREAMING_SNAKE_CASE:

| Error Type | Code |
|------------|------|
| `NotFoundError` | `NOT_FOUND` |
| `BadRequestError` | `BAD_REQUEST` |
| `InternalServerError` | `INTERNAL_SERVER` |
| `UnprocessableEntityError` | `UNPROCESSABLE_ENTITY` |
| `ValidationErrors` | `VALIDATION` |
| `JSONSyntaxError` | `JSONSYNTAX` |

## HTTP Status Codes Reference

| Code | Name | Usage |
|------|------|-------|
| 400 | Bad Request | Invalid input, validation errors |
| 401 | Unauthorized | Missing/invalid authentication |
| 403 | Forbidden | Authenticated but insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 408 | Request Timeout | Request took too long |
| 409 | Conflict | Resource already exists |
| 422 | Unprocessable Entity | Valid syntax, invalid semantics |
| 429 | Too Many Requests | Rate limit exceeded |
| 499 | Client Closed Request | Client disconnected |
| 500 | Internal Server Error | Unexpected server error |
| 503 | Service Unavailable | Temporary outage |

## Integration with Other Middleware

ErrorHandler runs **second** in the middleware chain (after TraceID):

```
1. TraceID          → Adds trace_id to logs
2. ErrorHandler     → Catches all errors/panics
3. ContextCancel    → Detects disconnects
4. CORS             → Handles preflight
5. Telemetry        → Traces requests
6. Logger           → Logs requests
```

This order ensures:
- All errors have trace IDs for debugging
- Panics in any middleware are caught
- Error logs include full request context
