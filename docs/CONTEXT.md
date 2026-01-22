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

Check for disconnection in long-running operations:

```go
import "github.com/edaniel30/http-platform-go/internal/middleware"

func ProcessLargeDataset(c *gin.Context) {
    for i, item := range dataset {
        // Check every 100 items
        if i%100 == 0 && middleware.IsContextCancelled(c) {
            c.Error(context.Canceled)
            return
        }
        process(item)
    }
}
```

### Per-Endpoint Timeouts

Override global timeout for specific endpoints:

```go
import "github.com/edaniel30/http-platform-go/internal/middleware"

// 30 second timeout for reports
platform.GET("/report",
    middleware.WithTimeout(30*time.Second),
    handler.GenerateReport,
)

// 5 second timeout for quick operations
platform.GET("/health",
    middleware.WithTimeout(5*time.Second),
    handler.HealthCheck,
)
```

### Error Inspection

Get specific error type:

```go
if err := middleware.GetContextError(c); err != nil {
    if errors.Is(err, context.Canceled) {
        // Client disconnected
    } else if errors.Is(err, context.DeadlineExceeded) {
        // Timeout exceeded
    }
}
```

## When to Use

| Scenario | Function | Frequency |
|----------|----------|-----------|
| Large file processing | `IsContextCancelled(c)` | Every N iterations |
| Expensive computations | `IsContextCancelled(c)` | Before each step |
| Report generation | `WithTimeout(duration)` | Per endpoint |
| Quick APIs | `WithTimeout(duration)` | Per endpoint |

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
