# TraceID Middleware

The TraceID middleware provides request traceability by generating or propagating unique identifiers across your distributed system.

## What It Does

The TraceID middleware helps your application:

- **Track requests across services**: Follow a request through multiple microservices
- **Enable distributed tracing**: Correlate logs from different components of the same request
- **Debug production issues**: Search logs by trace ID to see the complete request flow
- **Monitor request paths**: Understand how requests flow through your system

## Components

### 1. TraceID Middleware

**Purpose**: Ensures every request has a unique trace ID for tracking and correlation.

**Enabled by default** - runs as the first middleware in the chain.

**How it works**:
- Checks if the request has an `X-Trace-Id` header
- If present: uses the existing trace ID (propagates from upstream services)
- If absent: generates a new UUID v4 trace ID
- Stores the trace ID in the Gin context
- Adds the trace ID to the response header `X-Trace-Id`

**Example flow**:
```
Request without X-Trace-Id
  ↓
TraceID middleware generates: "550e8400-e29b-41d4-a716-446655440000"
  ↓
Stored in context & added to response header
  ↓
Available to all handlers and other middleware
```

### 2. GetTraceID(c)

**Purpose**: Retrieve the trace ID from the context in handlers or middleware.

**Returns**: The trace ID string, or empty string if not found.

```go
func (h *Handler) CreateUser(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)

    // Use in logs
    h.logger.Info("Creating user", Fields{
        "trace_id": traceID,
        "username": username,
    })

    // Pass to downstream services
    req.Header.Set("X-Trace-Id", traceID)
}
```

## When to Use

### Use GetTraceID when:
- **Logging operations**: Include trace ID in all log entries for correlation
- **Calling downstream services**: Propagate trace ID to maintain request tracking
- **Database operations**: Tag queries with trace ID for debugging
- **Error reporting**: Include trace ID in error messages for easier troubleshooting
- **Metrics**: Tag metrics with trace ID for detailed analysis

## HTTP Headers

### Request Header
- **`X-Trace-Id`**: Optional incoming trace ID from upstream services

### Response Header
- **`X-Trace-Id`**: The trace ID for this request (either received or generated)

## Usage Examples

### Example 1: Logging with Trace ID

```go
func (h *Handler) ProcessOrder(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)
    orderID := c.Param("id")

    h.logger.Info("Processing order", Fields{
        "trace_id": traceID,
        "order_id": orderID,
    })

    order, err := h.orderService.Process(orderID)
    if err != nil {
        h.logger.Error("Failed to process order", Fields{
            "trace_id": traceID,
            "order_id": orderID,
            "error": err,
        })
        c.Error(err)
        return
    }

    h.logger.Info("Order processed successfully", Fields{
        "trace_id": traceID,
        "order_id": orderID,
    })

    c.JSON(200, order)
}
```

**Resulting logs**:
```
INFO  Processing order trace_id=550e8400-e29b-41d4-a716-446655440000 order_id=12345
INFO  Order processed successfully trace_id=550e8400-e29b-41d4-a716-446655440000 order_id=12345
```

Now you can search logs by trace ID to see the complete flow!

### Example 2: Propagating to Downstream Services

```go
func (h *Handler) FetchUserData(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)
    userID := c.Param("id")

    // Call downstream service with same trace ID
    req, _ := http.NewRequest("GET",
        fmt.Sprintf("http://user-service/users/%s", userID), nil)

    // Propagate trace ID
    req.Header.Set("X-Trace-Id", traceID)

    resp, err := h.client.Do(req)
    if err != nil {
        c.Error(err)
        return
    }
    defer resp.Body.Close()

    // Now both services share the same trace ID
    c.JSON(resp.StatusCode, resp.Body)
}
```

### Example 3: Database Query Tagging

```go
func (h *Handler) GetUser(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)
    userID := c.Param("id")

    // Tag query with trace ID for debugging
    query := fmt.Sprintf("/* trace_id: %s */ SELECT * FROM users WHERE id = ?", traceID)

    var user User
    err := h.db.QueryRow(query, userID).Scan(&user.ID, &user.Name, &user.Email)
    if err != nil {
        c.Error(err)
        return
    }

    c.JSON(200, user)
}
```

### Example 4: Error Context

```go
func (h *Handler) ComplexOperation(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)

    result, err := h.service.DoComplexWork()
    if err != nil {
        // Include trace ID in error response
        c.JSON(500, gin.H{
            "error": "Operation failed",
            "trace_id": traceID,
            "message": "Please include this trace ID when contacting support",
        })
        return
    }

    c.JSON(200, result)
}
```

**Error response**:
```json
{
    "error": "Operation failed",
    "trace_id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Please include this trace ID when contacting support"
}
```

## Best Practices

**1. Always include trace ID in logs:**
```go
// ✅ Good - traceable
h.logger.Info("Event occurred", Fields{
    "trace_id": httpplatform.GetTraceID(c),
    "user_id": userID,
})

// ❌ Bad - no correlation
h.logger.Info("Event occurred", Fields{
    "user_id": userID,
})
```

**2. Propagate to all downstream services:**
```go
// ✅ Good - maintains trace across services
req.Header.Set("X-Trace-Id", httpplatform.GetTraceID(c))

// ❌ Bad - trace chain breaks
// No trace ID header set
```

**3. Include in error messages for support:**
```go
// ✅ Good - easy to debug
c.JSON(500, gin.H{
    "error": "Internal error",
    "trace_id": httpplatform.GetTraceID(c),
})

// ❌ Bad - hard to debug
c.JSON(500, gin.H{"error": "Internal error"})
```

**4. Use in database comments for query tracing:**
```go
// ✅ Good - can trace slow queries
query := fmt.Sprintf("/* trace_id: %s */ %s", traceID, baseQuery)

// ❌ Bad - can't correlate slow queries with requests
query := baseQuery
```

## Distributed Tracing Flow

```
Client Request
  ↓ (no X-Trace-Id header)
Service A (API Gateway)
  ↓ TraceID middleware generates: abc-123
  ↓ (X-Trace-Id: abc-123)
Service B (User Service)
  ↓ TraceID middleware extracts: abc-123
  ↓ (X-Trace-Id: abc-123)
Service C (Auth Service)
  ↓ TraceID middleware extracts: abc-123

All logs from these services have trace_id=abc-123
```

## Integration with OpenTelemetry

The TraceID middleware works seamlessly with the Telemetry middleware for full distributed tracing:

```go
cfg := httpplatform.DefaultConfig()
cfg.EnableTelemetry = true
cfg.ServiceName = "my-service"

// Both TraceID and Telemetry work together
// TraceID provides simple correlation
// Telemetry provides full distributed tracing with spans
```

## Configuration

Enabled by default:
```go
cfg := httpplatform.DefaultConfig()
// EnableTraceID is true by default
```

To disable (not recommended):
```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutTraceID(),
)
```

## Header Constants

Available for use in your code:

```go
import "github.com/edaniel30/http-platform-go/middleware"

// Header name
middleware.TraceIDHeader // "X-Trace-Id"

// Context key
middleware.TraceIDKey // "trace_id"
```
