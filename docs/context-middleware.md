# Context Middleware

The context middleware provides request cancellation detection and timeout management to prevent wasted processing when clients disconnect.

## What It Does

The context middleware helps your application:

- **Detect client disconnections**: Stop processing when the client is no longer waiting for a response
- **Save server resources**: Avoid wasting CPU, memory, and database connections on abandoned requests
- **Handle timeouts**: Set specific time limits for slow operations
- **Improve throughput**: Free up resources faster by detecting cancelled requests early

## Components

### 1. ContextCancellation Middleware

**Purpose**: Checks if the client has disconnected before processing the request.

**Enabled by default** - runs automatically on every request.

**How it works**:
- Intercepts each incoming request
- Checks if the context is already cancelled (`context.Canceled`)
- If cancelled: aborts processing and returns 499 (Client Closed Request)
- If active: continues to the handler

### 2. IsContextCancelled(c)

**Purpose**: Check if the client disconnected during long-running operations.

**Use case**: Processing large datasets, loops, or expensive operations.

```go
func (h *Handler) ProcessData(c *gin.Context) {
    for i, item := range largeDataset {
        // Check every 100 items if client disconnected
        if i % 100 == 0 && httpplatform.IsContextCancelled(c) {
            c.Error(context.Canceled)
            return
        }
        process(item)
    }
}
```

### 3. GetContextError(c)

**Purpose**: Get the specific context error (Canceled or DeadlineExceeded).

**Returns**:
- `context.Canceled` - client disconnected
- `context.DeadlineExceeded` - timeout exceeded
- `nil` - context still valid

```go
if err := httpplatform.GetContextError(c); err != nil {
    c.Error(err)
    return
}
```

### 4. WithTimeout(duration)

**Purpose**: Set a specific timeout for an endpoint (overrides global timeout).

**Use case**: Endpoints that need more or less time than the default.

```go
// 5 second timeout for this endpoint
platform.GET("/slow-operation",
    httpplatform.WithTimeout(5*time.Second),
    handler.SlowOperation,
)
```

## When to Use

### Use IsContextCancelled when:
- Processing large files or datasets
- Running long loops (check every N iterations)
- Performing expensive computations
- Making multiple sequential operations

### Use WithTimeout when:
- An endpoint needs a different timeout than the global default
- Generating reports or exports (longer timeout)
- Quick operations that should fail fast (shorter timeout)

### Use GetContextError when:
- You need to differentiate between cancellation types
- Handling errors from context-aware operations (DB queries, HTTP calls)

## HTTP Status Codes

- **499**: Client Closed Request (client disconnected)
- **408**: Request Timeout (timeout exceeded)

## Configuration

Enabled by default:
```go
cfg := httpplatform.DefaultConfig()
// EnableContextCancellation is true by default
```

To disable:
```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutContextCancellation(),
)
```
