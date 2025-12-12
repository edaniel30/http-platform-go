# Logger Middleware

The Logger middleware provides structured HTTP request logging with automatic trace ID correlation and severity-based log levels.

## What It Does

The Logger middleware helps your application:

- **Log all HTTP requests**: Records method, path, status, duration, and client IP
- **Structured logging**: Uses key-value fields for easy parsing and querying
- **Automatic trace correlation**: Includes trace IDs in all log entries
- **Severity-based logging**: Different log levels based on status codes (INFO, WARN, ERROR)
- **Performance tracking**: Records request duration in milliseconds
- **Pluggable logger**: Works with any logger implementation (Loki, Zap, Logrus, etc.)

## Components

### 1. BasicLogger Middleware

**Purpose**: Logs every HTTP request with timing and context information.

**Enabled by default** - runs as the last middleware in the chain.

**How it works**:
- Records request start time
- Processes the request
- Calculates duration
- Logs with appropriate severity based on status code

**Logged fields**:
- `method` - HTTP method (GET, POST, etc.)
- `path` - Request path (/api/users)
- `status` - Response status code (200, 404, 500, etc.)
- `duration` - Human-readable duration (45ms, 1.2s)
- `duration_ms` - Duration in milliseconds (45, 1200)
- `client_ip` - Client IP address
- `trace_id` - Trace ID (if available)
- `query` - Query parameters (if present)
- `errors` - Error messages (if any)

### 2. Logger Interface

**Purpose**: Allows using any logger implementation with the platform.

**Required methods**:
```go
type Logger interface {
    Info(msg string, fields Fields)
    Error(msg string, fields Fields)
    Warn(msg string, fields Fields)
    Debug(msg string, fields Fields)
    Close()
}
```

**Note**: The platform does NOT call `Close()`. You must close the logger yourself:
```go
logger := mylogger.New()
defer logger.Close() // You close it, not the platform

cfg.Logger = logger
platform, _ := httpplatform.New(cfg)
```

### 3. Log Levels

Logs are automatically assigned severity based on status code:

| Status Code | Log Level | Example |
|-------------|-----------|---------|
| 2xx, 3xx | **INFO** | Successful requests |
| 4xx | **WARN** | Client errors (404, 400, 401) |
| 5xx | **ERROR** | Server errors (500, 503) |

## Configuration

### Default Configuration

```go
cfg := httpplatform.DefaultConfig()
cfg.Logger = myLogger // Required - no default logger
cfg.EnableLogger = true // Enabled by default

platform, _ := httpplatform.New(cfg)
```

### Disable Logger

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutLogger(),
)
```

## Usage Examples

### Example 2: Loki Logger

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

cfg := httpplatform.DefaultConfig()
cfg.Logger = logger

platform, _ := httpplatform.New(cfg)
```


## When to Use

### Enable Logger when:
- Running in production or staging
- Debugging issues
- Monitoring performance
- Tracking request patterns

### Disable Logger when:
- Running tests (reduces noise)
- Local development (optional)
- High-performance scenarios where logging overhead matters