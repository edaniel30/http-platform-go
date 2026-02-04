# Context Cancellation

Detects client disconnections and prevents wasted processing on abandoned requests.

## Quick Start

**Enabled by default** - no configuration needed.

```go
platform, _ := httpplatform.New(httpplatform.DefaultConfig())
// ContextCancellation middleware runs automatically
```

## Features

### Automatic Detection
- Returns **499** (Client Closed Request) when client disconnects
- Returns **408** (Request Timeout) when timeout exceeded
- Stops processing immediately to free resources

### Manual Checks

Check for disconnection in long-running operations using Go's standard `context` API:

```go
import "context"

func ProcessLargeDataset(c *gin.Context) {
    for i, item := range dataset {
        // Check every 100 items using standard context API
        if i%100 == 0 && c.Request.Context().Err() != nil {
            c.Error(c.Request.Context().Err())
            return
        }
        process(item)
    }
}
```

### Per-Endpoint Timeouts

Create custom timeout middleware for specific endpoints using Go's `context.WithTimeout`:

```go
import (
    "context"
    "time"
    "github.com/gin-gonic/gin"
)

// Custom timeout middleware
func withTimeout(timeout time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()

        c.Request = c.Request.WithContext(ctx)
        c.Next()

        // Check if timeout occurred
        if ctx.Err() == context.DeadlineExceeded {
            c.Error(ctx.Err())
            c.Abort()
        }
    }
}

// Use it on specific endpoints
platform.GET("/report",
    withTimeout(30*time.Second),
    handler.GenerateReport,
)

platform.GET("/health",
    withTimeout(5*time.Second),
    handler.HealthCheck,
)
```

### Error Inspection

Use Go's standard context error checking:

```go
import (
    "context"
    "errors"
)

func handler(c *gin.Context) {
    if err := c.Request.Context().Err(); err != nil {
        if errors.Is(err, context.Canceled) {
            // Client disconnected
        } else if errors.Is(err, context.DeadlineExceeded) {
            // Timeout exceeded
        }
        c.Error(err)
        return
    }
    // Continue processing...
}
```

## When to Use

| Scenario | How to Check | Frequency |
|----------|-------------|-----------|
| Large file processing | `c.Request.Context().Err() != nil` | Every N iterations |
| Expensive computations | `c.Request.Context().Err() != nil` | Before each step |
| Report generation | Custom `withTimeout` middleware | Per endpoint |
| Quick APIs | Custom `withTimeout` middleware | Per endpoint |

## Configuration

### Global timeout (server-level):

```go
cfg := httpplatform.DefaultConfig()
cfg.ReadTimeout = 60  // seconds
cfg.WriteTimeout = 60 // seconds

platform, _ := httpplatform.New(cfg)
```

## Best Practices

✅ **Do:**
- Check cancellation in loops processing > 100 items
- Use shorter timeouts for critical endpoints
- Use longer timeouts for reports/exports
- Handle `context.Canceled` gracefully

❌ **Don't:**
- Check cancellation on every iteration (overhead)
- Disable globally (wastes resources)
- Use timeouts < 1 second (too aggressive)
- Ignore context errors in handlers
