# TraceID

Generates or propagates unique identifiers for tracking requests across distributed systems.

## Quick Start

**Enabled by default** - automatically runs as the first middleware:

```go
platform, _ := httpplatform.New(httpplatform.DefaultConfig())
// Every request gets a trace ID
```

## Features

- **Request tracking** - Follow requests through multiple microservices
- **Log correlation** - Connect logs from different services for the same request
- **Production debugging** - Search logs by trace ID to see complete request flow
- **Upstream propagation** - Respects incoming `X-Trace-Id` headers

## How It Works

```
Request without X-Trace-Id
  ↓
TraceID middleware generates: "550e8400-e29b-41d4-a716-446655440000"
  ↓
Stored in Gin context + added to response header
  ↓
Available to handlers, middleware, and downstream services
```

If request already has `X-Trace-Id` header, that value is used instead.

## HTTP Headers

| Header | Direction | Description |
|--------|-----------|-------------|
| `X-Trace-Id` | Request | Optional - incoming trace ID from upstream |
| `X-Trace-Id` | Response | Always present - generated or propagated ID |

## Usage

### Get TraceID in Handler

```go
import "github.com/edaniel30/http-platform-go"

func (h *Handler) CreateUser(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)

    // Use in logs
    h.logger.Info("Creating user", map[string]any{
        "trace_id": traceID,
        "username": username,
    })

    // Pass to downstream services
    req.Header.Set("X-Trace-Id", traceID)
}
```

## Common Patterns

### 1. Logging with TraceID

```go
func (h *Handler) ProcessOrder(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)

    h.logger.Info("Processing order", map[string]any{
        "trace_id": traceID,
        "order_id": orderID,
    })

    result, err := h.service.Process(orderID)
    if err != nil {
        h.logger.Error("Failed", map[string]any{
            "trace_id": traceID,
            "error": err,
        })
        c.Error(err)
        return
    }

    c.JSON(200, result)
}
```

**Resulting logs**:
```
INFO  Processing order trace_id=550e8400... order_id=12345
INFO  Order completed trace_id=550e8400... order_id=12345
```

### 2. Propagate to Downstream Services

```go
func (h *Handler) FetchUserData(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)

    // Create request to downstream service
    req, _ := http.NewRequest("GET", "http://user-service/users/123", nil)

    // Propagate trace ID
    req.Header.Set("X-Trace-Id", traceID)

    resp, err := h.client.Do(req)
    // ...
}
```

### 3. Include in Error Responses

```go
func (h *Handler) ComplexOperation(c *gin.Context) {
    traceID := httpplatform.GetTraceID(c)

    result, err := h.service.DoWork()
    if err != nil {
        c.JSON(500, gin.H{
            "error": "Operation failed",
            "trace_id": traceID,
            "message": "Include this trace ID when contacting support",
        })
        return
    }

    c.JSON(200, result)
}
```

**Response**:
```json
{
    "error": "Operation failed",
    "trace_id": "550e8400-e29b-41d4-a716-446655440000",
    "message": "Include this trace ID when contacting support"
}
```

## Distributed Tracing Flow

```
Client Request (no X-Trace-Id)
  ↓
Service A (API Gateway)
  ↓ Generates: abc-123
  ↓ Sets header: X-Trace-Id: abc-123
Service B (User Service)
  ↓ Extracts: abc-123
  ↓ Sets header: X-Trace-Id: abc-123
Service C (Auth Service)
  ↓ Extracts: abc-123

All services log with trace_id=abc-123
```

## When to Use GetTraceID

| Use Case | Example |
|----------|---------|
| **Logging** | Include in all log entries for correlation |
| **Downstream calls** | Propagate to maintain request tracking |
| **Database ops** | Tag queries with trace ID for debugging |
| **Error reporting** | Include in error messages for troubleshooting |
| **Metrics** | Tag metrics with trace ID for analysis |
| **User support** | Return in API errors so users can reference |

## Configuration

### Default (Enabled)

```go
cfg := httpplatform.DefaultConfig()
platform, _ := httpplatform.New(cfg)
// TraceID is enabled by default
```

### Disable (Not Recommended)

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutTraceID(),
)
```

**Warning**: Disabling TraceID makes debugging production issues significantly harder.

## Best Practices

### ✅ Do

```go
// Always include trace ID in logs
h.logger.Info("Event occurred", map[string]any{
    "trace_id": httpplatform.GetTraceID(c),
    "user_id": userID,
})

// Propagate to all downstream services
req.Header.Set("X-Trace-Id", httpplatform.GetTraceID(c))

// Include in error responses for support
c.JSON(500, gin.H{
    "error": "Internal error",
    "trace_id": httpplatform.GetTraceID(c),
})

// Use in database comments for query tracing
query := fmt.Sprintf("/* trace_id: %s */ %s", traceID, baseQuery)
```

### ❌ Don't

```go
// Don't log without trace ID
h.logger.Info("Event occurred", map[string]any{
    "user_id": userID,  // Missing trace_id
})

// Don't break trace chain
// (Not propagating to downstream services)

// Don't disable in production
httpplatform.WithoutTraceID()  // Hard to debug

// Don't use custom trace ID format
// (Always use X-Trace-Id header standard)
```

## Constants

Available for use in your code:

```go
import "github.com/edaniel30/http-platform-go/internal/middleware"

// Header name
middleware.TraceIDHeader  // "X-Trace-Id"

// Context key
middleware.TraceIDKey     // "trace_id"
```

## Integration with Telemetry

TraceID works seamlessly with OpenTelemetry:

```go
cfg := httpplatform.DefaultConfig()

platform, _ := httpplatform.New(cfg,
    // TraceID enabled by default - provides X-Trace-Id header
    httpplatform.WithTelemetry("my-service", "localhost:4318", true),
    // Telemetry provides full distributed tracing with spans
)
```

**TraceID**: Simple correlation via headers
**Telemetry**: Full OpenTelemetry tracing with spans and attributes

## Middleware Order

TraceID runs **first** in the middleware chain:

```
1. TraceID       → Generates/extracts trace_id ✓
2. ErrorHandler  → Catches errors
3. ContextCancel → Detects disconnects
4. CORS          → Handles preflight
5. Telemetry     → Creates spans
6. Logger        → Logs requests
```

This order ensures:
- All logs have trace IDs
- Error logs include trace IDs
- Telemetry spans are tagged with trace IDs
- HTTP requests are logged with trace IDs
