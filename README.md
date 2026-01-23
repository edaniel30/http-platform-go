# http-platform-go

A powerful, flexible, and production-ready HTTP server platform for Go, built on top of Gin with automatic middleware setup, logger integration, and graceful shutdown.

## Installation

```bash
go get github.com/edaniel30/http-platform-go
```

## Quick Start

```go
package main

import (
    "context"
    "net/http"

    "github.com/edaniel30/http-platform-go"
    "github.com/gin-gonic/gin"
)

func main() {
    // Create platform with default configuration
    // Uses DefaultLogger (stdout) by default
    platform, err := httpplatform.New(httpplatform.DefaultConfig())
    if err != nil {
        panic(err)
    }

    // Register routes
    platform.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Start server with graceful shutdown
    ctx := context.Background()
    platform.Start(ctx)
}
```

## Configuration

### Default Configuration

The library provides sensible defaults for development:

```go
httpplatform.DefaultConfig()
// Returns:
// - Port: 8080
// - Mode: "debug"
// - ReadTimeout: 30s
// - WriteTimeout: 30s
// - IdleTimeout: 60s
// - MaxHeaderBytes: 1MB
// - Logger: DefaultLogger (stdout)
// - ErrorHandler: Enabled
// - TraceID: Enabled
// - ContextCancellation: Enabled
// - CORS: Enabled with origins ["*"]
// - Telemetry: Disabled
```

### Configuration Options

All options use the functional options pattern:

#### Server Options

```go
httpplatform.WithPort(8080)                        // Set server port
httpplatform.WithMode("release")                   // Set Gin mode: debug, release, or test
httpplatform.WithReadTimeout(60 * time.Second)     // Set read timeout
httpplatform.WithWriteTimeout(60 * time.Second)    // Set write timeout
httpplatform.WithIdleTimeout(120 * time.Second)    // Set idle timeout
httpplatform.WithMaxHeaderBytes(2 << 20)           // Set max header bytes (2MB)
httpplatform.WithBasePath("/api/v1")               // Set base path for all routes
httpplatform.WithTrustedProxies([]string{"10.0.0.1"}) // Set trusted proxies
```

#### Logger Options

```go
// Use DefaultLogger (stdout) - enabled by default
httpplatform.DefaultConfig()

// Use custom logger (Loki, Zap, Logrus, etc.)
httpplatform.WithLogger(myLogger)

// Disable logger
httpplatform.WithoutLogger()
```

**Creating a custom logger adapter:**

The `Logger` interface is public, allowing you to create adapters for any logger:

```go
type MyLoggerAdapter struct {
    logger *mylogger.Logger
}

func (a *MyLoggerAdapter) Info(ctx context.Context, msg string, fields map[string]any) {
    a.logger.Info(msg, convertFields(fields))
}
// ... implement Error, Warn, Debug, Close
```

See [Logger Documentation](docs/LOGGER.md) for complete examples with Zap, Logrus, and Loki.

#### CORS Options

```go
httpplatform.WithCORSOrigins("https://example.com", "https://app.example.com")
httpplatform.WithCORSMethods("GET", "POST", "PUT", "DELETE")
httpplatform.WithCORSHeaders("Authorization", "Content-Type")
httpplatform.WithCORSCredentials(true)              // Allow cookies/auth headers
httpplatform.WithCORS(&middleware.CORSConfig{       // Full control
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST"},
    AllowCredentials: true,
    MaxAge:          12 * time.Hour,
})
```

See [CORS Documentation](docs/CORS.md) for detailed configuration.

#### Telemetry Options

```go
// Enable OpenTelemetry tracing
httpplatform.WithTelemetry("service-name", "localhost:4318", true)

// Disable telemetry (default)
httpplatform.WithoutTelemetry()
```

See [Telemetry Documentation](docs/TELEMETRY.md) for OpenTelemetry setup.

#### Middleware Toggles

```go
httpplatform.WithoutTraceID()   // Disable TraceID middleware (not recommended)
httpplatform.WithoutLogger()    // Disable Logger middleware
```

#### Base Path

Set a base path for all routes registered with the platform:

```go
httpplatform.WithBasePath("/api/v1")  // All routes will be prefixed with /api/v1

// Example:
platform.GET("/health", handler)  // Actual route: /api/v1/health
platform.GET("/users", handler)   // Actual route: /api/v1/users
```

**Note:** When using `WithBasePath`, all routes registered directly on the platform will be automatically prefixed. You can still create nested groups:

```go
platform, _ := httpplatform.New(
    httpplatform.DefaultConfig(),
    httpplatform.WithBasePath("/api/v1"),  // Base path
)

// Route: /api/v1/health
platform.GET("/health", healthHandler)

// Nested group: /api/v1/users/*
users := platform.Group("/users")
users.GET("", listUsers)       // Route: /api/v1/users
users.GET("/:id", getUser)     // Route: /api/v1/users/:id
```

## Middlewares

The platform automatically applies middleware in the following order:

| Order | Middleware | Status | Description |
|-------|------------|--------|-------------|
| 1 | [**TraceID**](docs/TRACE.md) | Always enabled | Generates or extracts `X-Trace-Id` for request correlation |
| 2 | [**ErrorHandler**](docs/ERROR_HANDLER.md) | Always enabled | Catches panics and formats errors with consistent JSON responses |
| 3 | [**ContextCancellation**](docs/CONTEXT.md) | Always enabled | Detects client disconnections (499) and timeouts (408) |
| 4 | [**CORS**](docs/CORS.md) | Enabled by default | Handles cross-origin requests for browser-based clients |
| 5 | [**Telemetry**](docs/TELEMETRY.md) | Disabled by default | OpenTelemetry distributed tracing (Jaeger, Datadog, Zipkin) |
| 6 | [**Logger**](docs/LOGGER.md) | Enabled by default | Logs HTTP requests with status, duration, and trace ID |


## Route Registration

### Simple Routes

```go
platform.GET("/users", listUsers)
platform.POST("/users", createUser)
platform.PUT("/users/:id", updateUser)
platform.DELETE("/users/:id", deleteUser)
platform.PATCH("/users/:id", patchUser)
platform.OPTIONS("/users", optionsUsers)
platform.HEAD("/users", headUsers)
```

### Route Groups

Organize related routes under a common prefix:

```go
api := platform.Group("/api/v1")
{
    api.GET("/health", healthCheck)

    users := api.Group("/users")
    {
        users.GET("", listUsers)
        users.GET("/:id", getUser)
        users.POST("", createUser)
    }

    products := api.Group("/products")
    {
        products.GET("", listProducts)
        products.POST("", createProduct)
    }
}
```

## Graceful Shutdown

The platform handles graceful shutdown automatically with a 5-second timeout:

```go
import (
    "context"
    "os/signal"
    "syscall"
)

// Automatic graceful shutdown on SIGINT/SIGTERM
ctx := context.Background()
platform.Start(ctx)

// OR with signal handling
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
platform.Start(ctx)
```

## Utility Functions

The platform provides utility functions to simplify common request processing tasks.

### QueryParamsToMap

Extracts all query parameters and returns them as a `map[string]string`:

```go
func getUsersHandler(c *gin.Context) {
    // Request: GET /users?name=John&age=30&status=active&status=pending

    params := httpplatform.QueryParamsToMap(c)
    // Result: map[string]string{
    //   "name": "John",
    //   "age": "30",
    //   "status": "active"  // First value only
    // }

    // Use the params directly (type-safe, no type assertions needed)
    name := params["name"]
    age := params["age"]
    status := params["status"]

    // For multiple values, use c.QueryArray() or c.Request.URL.Query()
    statuses := c.QueryArray("status")  // []string{"active", "pending"}
}
```

**Behavior:**
- All parameters are returned as `string`
- Multi-value parameters return only the **first value**
- Empty map if no query parameters
- For accessing all values, use `c.QueryArray()` or `c.Request.URL.Query()`

### HeadersToMap

Extracts all request headers and returns them as a `map[string]string`:

```go
func logHeadersHandler(c *gin.Context) {
    // Request headers:
    // Content-Type: application/json
    // Accept: application/json, text/plain
    // X-Request-ID: abc123
    // Authorization: Bearer token

    headers := httpplatform.HeadersToMap(c)
    // Result: map[string]string{
    //   "Content-Type": "application/json",
    //   "Accept": "application/json",  // First value only
    //   "X-Request-Id": "abc123",
    //   "Authorization": "Bearer token"
    // }

    // Access specific headers directly (type-safe)
    contentType := headers["Content-Type"]
    accept := headers["Accept"]
    requestID := headers["X-Request-Id"]

    // For multiple header values, use c.Request.Header directly
    acceptHeaders := c.Request.Header["Accept"]  // []string{"application/json", "text/plain"}

    // Log for debugging
    logger.Info("Request headers", map[string]any{"headers": headers})
}
```

**Behavior:**
- All headers are returned as `string`
- Multi-value headers return only the **first value**
- Header names are case-sensitive as received from the client
- Empty map if no headers
- For accessing all values, use `c.Request.Header` directly

## Dependencies

### Required
- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [gin-contrib/cors](https://github.com/gin-contrib/cors) - CORS middleware
- [google/uuid](https://github.com/google/uuid) - UUID generation for TraceID
- [go-playground/validator](https://github.com/go-playground/validator) - Request validation

### Optional
- [edaniel30/loki-logger-go](https://github.com/edaniel30/loki-logger-go) - Loki logger integration
- [go.opentelemetry.io/otel](https://github.com/open-telemetry/opentelemetry-go) - OpenTelemetry tracing

**Note:** The platform works out-of-the-box with `DefaultLogger` (no external logger required).


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
