# Logger

Structured HTTP request logging with trace correlation and automatic severity levels.

## Quick Start

**Enabled by default** with DefaultLogger (stdout):

```go
platform, _ := httpplatform.New(httpplatform.DefaultConfig())
// Logs all HTTP requests to stdout
```

## Logged Fields

Every HTTP request logs:

| Field | Description | Example |
|-------|-------------|---------|
| `method` | HTTP method | `GET`, `POST` |
| `path` | Request path | `/api/users` |
| `status` | HTTP status code | `200`, `404`, `500` |
| `duration` | Human-readable time | `45ms`, `1.2s` |
| `duration_ms` | Milliseconds | `45`, `1200` |
| `client_ip` | Client IP address | `192.168.1.1` |
| `trace_id` | Request trace ID | `abc-123-def` |
| `query` | Query params (if any) | `page=1` |
| `errors` | Error messages (if any) | `User not found` |

## Log Levels (Automatic)

| Status Code | Level | Usage |
|-------------|-------|-------|
| 2xx, 3xx | **INFO** | Successful requests |
| 4xx | **WARN** | Client errors |
| 5xx | **ERROR** | Server errors |

Example output:

```
INFO: Request completed method=GET path=/api/users status=200 duration=45ms trace_id=abc-123
WARN: Request completed with client error method=POST path=/api/login status=401 duration=12ms trace_id=def-456
ERROR: Request failed method=GET path=/api/orders status=500 duration=230ms trace_id=ghi-789
```

## Custom Logger

### Logger Interface

The `Logger` interface is defined in the public `httpplatform` package, making it easy to create custom adapters:

```go
package httpplatform

type Logger interface {
    Info(ctx context.Context, msg string, fields map[string]any)
    Error(ctx context.Context, msg string, fields map[string]any)
    Warn(ctx context.Context, msg string, fields map[string]any)
    Debug(ctx context.Context, msg string, fields map[string]any)
    Close() error
}
```

### Using Logrus

Example adapter for Logrus:

```go
package main

import (
    "context"

    "github.com/edaniel30/http-platform-go"
    "github.com/sirupsen/logrus"
)

type LogrusAdapter struct {
    logger *logrus.Logger
}

func NewLogrusAdapter(logger *logrus.Logger) *LogrusAdapter {
    return &LogrusAdapter{logger: logger}
}

func (a *LogrusAdapter) Info(ctx context.Context, msg string, fields map[string]any) {
    a.logger.WithFields(logrus.Fields(fields)).Info(msg)
}

func (a *LogrusAdapter) Error(ctx context.Context, msg string, fields map[string]any) {
    a.logger.WithFields(logrus.Fields(fields)).Error(msg)
}

func (a *LogrusAdapter) Warn(ctx context.Context, msg string, fields map[string]any) {
    a.logger.WithFields(logrus.Fields(fields)).Warn(msg)
}

func (a *LogrusAdapter) Debug(ctx context.Context, msg string, fields map[string]any) {
    a.logger.WithFields(logrus.Fields(fields)).Debug(msg)
}

func (a *LogrusAdapter) Close() error {
    return nil // Logrus doesn't need explicit closing
}
```

Usage:

```go
logrusLogger := logrus.New()
logrusLogger.SetFormatter(&logrus.JSONFormatter{})

logger := NewLogrusAdapter(logrusLogger)

platform, _ := httpplatform.New(
    httpplatform.DefaultConfig(),
    httpplatform.WithLogger(logger),
)
```

### Using Loki Logger

If the logger already implements the `httpplatform.Logger` interface:

```go
import "github.com/edaniel30/loki-logger-go"

logger := loki.New(loki.Config{
    URL:       "http://loki:3100",
    BatchSize: 100,
    Labels: map[string]string{
        "app": "user-service",
        "env": "production",
    },
})
defer logger.Close()

platform, _ := httpplatform.New(
    httpplatform.DefaultConfig(),
    httpplatform.WithLogger(logger),
)
```

## Configuration

### Default (stdout)

```go
cfg := httpplatform.DefaultConfig()
// Uses DefaultLogger (stdlib log to stdout)
platform, _ := httpplatform.New(cfg)
```

### Custom Logger

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithLogger(myLogger),
)
```

### Disable Logger

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutLogger(),
)
```

## Logger Lifecycle

**Important**: You manage the logger lifecycle, not the platform.

```go
// ✅ You create it
logger := loki.New(config)

// ✅ You close it
defer logger.Close()

// Platform uses it but doesn't close it
platform, _ := httpplatform.New(cfg, httpplatform.WithLogger(logger))
```

## When to Use

| Scenario | Enable Logger | Notes |
|----------|---------------|-------|
| Production | ✅ Yes | Essential for monitoring |
| Staging | ✅ Yes | Debug issues |
| Development | Optional | Reduces noise |
| Testing | ❌ No | Use `WithoutLogger()` |
| High-perf APIs | ⚠️ Maybe | Consider sampling |

## Best Practices

### ✅ Do

```go
// Close logger on shutdown
logger := loki.New(config)
defer logger.Close()

// Use structured logging
logger.Info(ctx, "User created", map[string]any{
    "user_id": user.ID,
    "email":   user.Email,
})

// Disable in tests
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutLogger(), // Cleaner test output
)
```

### ❌ Don't

```go
// Don't let platform close logger
// (it won't - you'll leak resources)

// Don't log sensitive data
logger.Info(ctx, "Login", map[string]any{
    "password": user.Password, // Never!
})

// Don't enable in high-load tests
// (slows down benchmarks)
```

## Middleware Order

Logger runs **last** to capture final response:

```
1. TraceID       → Generates trace_id
2. ErrorHandler  → Catches errors
3. ContextCancel → Detects disconnects
4. CORS          → Adds headers
5. Telemetry     → Traces request
6. Logger        → Logs final result ✓
```

This order ensures:
- All logs have trace IDs
- Status codes include error handler results
- Duration includes all middleware overhead

## DefaultLogger Details

Built-in logger using Go's `log` package:

```go
// Format: [http-platform] TIMESTAMP LEVEL: message key=value
[http-platform] 2026/01/22 10:30:45 INFO: Request completed method=GET path=/users status=200
```

Features:
- No external dependencies
- Stdout output
- Simple key=value format
- Alphabetically sorted fields

Good for:
- Development
- Simple deployments
- Getting started

Consider upgrading to:
- **Loki** - for centralized logging
- **Zap** - for high performance
- **Logrus** - for structured JSON

## Example: Development vs Production

### Development

```go
// Simple stdout logging
cfg := httpplatform.DefaultConfig()
platform, _ := httpplatform.New(cfg)
```

### Production

```go
// Loki for centralized logs
logger := loki.New(loki.Config{
    URL:       os.Getenv("LOKI_URL"),
    BatchSize: 500,
    Labels: map[string]string{
        "app":     "user-service",
        "env":     "production",
        "region":  "us-east-1",
    },
})
defer logger.Close()

platform, _ := httpplatform.New(cfg,
    httpplatform.WithLogger(logger),
)
```
