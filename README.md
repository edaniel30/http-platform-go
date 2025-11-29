# HTTP Platform Go

A powerful, flexible, and production-ready HTTP server platform for Go, built on top of Gin with automatic middleware setup, logger integration, and graceful shutdown.

## Features

- 🚀 **Simple API** - Clean and intuitive interface for building HTTP servers
- 🔌 **Automatic Middleware** - TraceID, CORS, Recovery, and Logger pre-configured
- ⚙️ **Functional Options** - Flexible configuration using functional options pattern
- 📝 **Logger Integration** - Built-in integration with [loki-logger-go](https://github.com/edaniel30/loki-logger-go)
- 🛡️ **Production Ready** - Graceful shutdown, error handling, and thread-safe operations
- ❌ **Error Handler Middleware** - Structured JSON error responses for HTTP errors
- 🎯 **Type Safe** - Strongly typed configuration and handlers
- 📚 **Well Documented** - Comprehensive godoc comments and examples

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
    "github.com/edaniel30/http-platform-go/config"
    "github.com/edaniel30/loki-logger-go"
    "github.com/edaniel30/loki-logger-go/models"
    "github.com/gin-gonic/gin"
)

func main() {
    // Initialize logger
    logger, _ := loki.New(
        models.DefaultConfig(),
        models.WithAppName("my-app"),
        models.WithOnlyConsole(true),
    )

    // Create platform
    platform, err := httpplatform.New(
        config.DefaultConfig(),
        config.WithPort(8080),
        config.WithLogger(logger),
    )
    if err != nil {
        panic(err)
    }

    // Register routes
    platform.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"status": "ok"})
    })

    // Start server
    ctx := context.Background()
    platform.Start(ctx)
}
```

## Configuration

### Default Configuration

The library provides sensible defaults for development:

```go
config.DefaultConfig()
// Returns:
// - Port: 8080
// - Mode: "debug"
// - ReadTimeout: 30s
// - WriteTimeout: 30s
// - IdleTimeout: 60s
// - MaxHeaderBytes: 1MB
// - CORS: Enabled with origins ["*"]
// - TraceID: Enabled
// - Recovery: Enabled
// - Logger: Enabled
```

### Configuration Options

All options use the functional options pattern:

#### Server Options

```go
config.WithPort(8080)                        // Set server port
config.WithMode("release")                   // Set Gin mode: debug, release, or test
config.WithReadTimeout(60 * time.Second)     // Set read timeout
config.WithWriteTimeout(60 * time.Second)    // Set write timeout
config.WithIdleTimeout(120 * time.Second)    // Set idle timeout
config.WithMaxHeaderBytes(2 << 20)           // Set max header bytes (2MB)
config.WithBasePath("/api/v1")               // Set base path for all routes
config.WithTrustedProxies([]string{"10.0.0.1"}) // Set trusted proxies
```

#### Logger Option

```go
config.WithLogger(myLogger)  // Inject loki-logger instance (required)
```

#### CORS Options

```go
config.WithCORS([]string{"https://example.com"})           // Set allowed origins
config.WithAllowedMethods([]string{"GET", "POST"})         // Set allowed methods
config.WithAllowedHeaders([]string{"Content-Type"})        // Set allowed headers
config.WithExposedHeaders([]string{"X-Trace-Id"})          // Set exposed headers
config.WithAllowCredentials(true)                          // Allow credentials
config.WithMaxAge(12 * time.Hour)                          // Set preflight cache duration
```

#### Middleware Toggles

```go
config.WithoutTraceID()   // Disable TraceID middleware
config.WithoutCORS()      // Disable CORS middleware
config.WithoutRecovery()  // Disable Recovery middleware
config.WithoutLogger()    // Disable Logger middleware
```

#### Base Path

Set a base path for all routes registered with the platform:

```go
config.WithBasePath("/api/v1")  // All routes will be prefixed with /api/v1

// Example:
platform.GET("/health", handler)  // Actual route: /api/v1/health
platform.GET("/users", handler)   // Actual route: /api/v1/users
```

**Note:** When using `WithBasePath`, all routes registered directly on the platform will be automatically prefixed. You can still create nested groups:

```go
platform, _ := httpplatform.New(
    config.DefaultConfig(),
    config.WithBasePath("/api/v1"),  // Base path
)

// Route: /api/v1/health
platform.GET("/health", healthHandler)

// Nested group: /api/v1/users/*
users := platform.Group("/users")
users.GET("", listUsers)       // Route: /api/v1/users
users.GET("/:id", getUser)     // Route: /api/v1/users/:id
```

### Full Configuration Example

```go
platform, err := httpplatform.New(
    config.DefaultConfig(),
    config.WithPort(8080),
    config.WithLogger(logger),
    config.WithMode("release"),
    config.WithReadTimeout(60 * time.Second),
    config.WithWriteTimeout(60 * time.Second),
    config.WithCORS([]string{"https://example.com"}),
    config.WithAllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
    config.WithBasePath("/api/v1"),  // All routes will be prefixed with /api/v1
)

// Example route registration
platform.GET("/health", healthHandler)  // Accessible at: /api/v1/health
```

## Middleware

The platform automatically applies middleware in the following order:

1. **TraceID** - Generates or extracts trace IDs for distributed tracing
2. **CORS** - Handles cross-origin resource sharing
3. **Recovery** - Recovers from panics and logs errors
4. **Logger** - Logs all HTTP requests with method, path, status, and duration
5. **Erros** - Error handler for Http default response

### Built-in Middleware

#### TraceID Middleware

Automatically generates or extracts trace IDs from the `X-Trace-Id` header:

```go
import "github.com/edaniel30/http-platform-go/middleware"

// Get trace ID in handler
traceID := middleware.GetTraceID(c)
```

#### Custom Middleware

Add custom middleware using the `Use` method:

```go
platform.Use(func(c *gin.Context) {
    // Pre-processing
    start := time.Now()

    c.Next()

    // Post-processing
    duration := time.Since(start)
    log.Printf("Request took %v", duration)
})
```

#### Error Handler Middleware

The platform provides a comprehensive error handling middleware that automatically converts errors into structured JSON responses.

```go
// Apply globally to all routes
platform.Use(httpplatform.ErrorHandler())

// Or apply to specific routes
platform.GET("/users/:id", httpplatform.ErrorHandler(), getUserHandler)

// Or apply to a route group
api := platform.Group("/api")
api.Use(httpplatform.ErrorHandler())
```

**Handled Error Types:**

The middleware handles the following error types and converts them to appropriate HTTP responses:

| Error Type | HTTP Status | Description |
|------------|-------------|-------------|
| `NotFoundError` | 404 | Resource not found |
| `UnauthorizedError` | 401 | Authentication required or failed |
| `ConflictError` | 409 | Resource conflict (e.g., duplicate) |
| `BadRequestError` | 400 | Invalid request data |
| `DomainError` | 400 | Business logic error |
| `ExternalServiceError` | Custom | External service error with custom status |
| `validator.ValidationErrors` | 400 | Request validation errors |
| `json.UnmarshalTypeError` | 400 | JSON type mismatch |
| `panic` | 500 | Recovered panic |
| Other errors | 500 | Generic internal server error |

**Using Domain Errors in Handlers:**

```go
import "github.com/edaniel30/http-platform-go"

func getUserHandler(c *gin.Context) {
    id := c.Param("id")

    user, err := userService.GetByID(id)
    if err != nil {
        // Return domain error - middleware will handle it
        c.Error(httpplatform.NewNotFoundError("User not found"))
        return
    }

    c.JSON(http.StatusOK, user)
}
```

**Error Response Format:**

All errors are returned in a consistent JSON format:

```json
{
  "message": "User not found",
  "error": "Not Found",
  "status": 404,
  "cause": []
}
```

For validation errors, the `cause` field contains detailed information:

```json
{
  "message": "Validation error",
  "error": "Bad Request",
  "status": 400,
  "cause": [
    {
      "field": "Email",
      "reason": "required"
    },
    {
      "field": "Age",
      "reason": "min=18"
    }
  ]
}
```

**Available Error Constructors:**

```go
// Create domain errors
httpplatform.NewNotFoundError("Resource not found")
httpplatform.NewUnauthorizedError("Invalid credentials")
httpplatform.NewConflictError("Resource already exists")
httpplatform.NewBadRequestError("Invalid input")
httpplatform.NewDomainError("Business rule violated")
httpplatform.NewExternalServiceError("Service unavailable", 503)
```

**Panic Recovery:**

The error handler middleware automatically recovers from panics and returns a 500 error:

```go
func panicHandler(c *gin.Context) {
    panic("something went wrong")
    // Middleware catches this and returns:
    // {"message": "Internal server error panic", "error": "Internal Server Error", "status": 500}
}
```

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

Extracts all query parameters and returns them as a `map[string]any`:

```go
func getUsersHandler(c *gin.Context) {
    // Request: GET /users?name=John&age=30&status=active&status=pending

    params := httpplatform.QueryParamsToMap(c)
    // Result: map[string]any{
    //   "name": "John",           // Single value as string
    //   "age": "30",              // Single value as string
    //   "status": []string{"active", "pending"}  // Multiple values as []string
    // }

    // Use the params in your business logic
    name, _ := params["name"].(string)
    age, _ := params["age"].(string)

    // Handle multiple values
    if statuses, ok := params["status"].([]string); ok {
        // Process multiple status values
        for _, status := range statuses {
            // ...
        }
    }
}
```

**Behavior:**
- Single-value parameters are returned as `string`
- Multi-value parameters are returned as `[]string`
- Empty map if no query parameters

### HeadersToMap

Extracts all request headers and returns them as a `map[string]any`:

```go
func logHeadersHandler(c *gin.Context) {
    // Request headers:
    // Content-Type: application/json
    // Accept: application/json, text/plain
    // X-Request-ID: abc123
    // Authorization: Bearer token

    headers := httpplatform.HeadersToMap(c)
    // Result: map[string]any{
    //   "Content-Type": "application/json",
    //   "Accept": []string{"application/json", "text/plain"},
    //   "X-Request-Id": "abc123",
    //   "Authorization": "Bearer token"
    // }

    // Access specific headers
    contentType, _ := headers["Content-Type"].(string)

    // Handle multiple header values
    if accepts, ok := headers["Accept"].([]string); ok {
        for _, accept := range accepts {
            // Process each accept type
        }
    }

    // Log for debugging
    logger.Info("Request headers", models.Fields{"headers": headers})
}
```

**Behavior:**
- Single-value headers are returned as `string`
- Multi-value headers are returned as `[]string`
- Header names are case-sensitive as received from the client
- Empty map if no headers

## Best Practices

### 1. Always Inject Logger

The logger is required and must be injected during platform creation:

```go
platform, err := httpplatform.New(
    config.DefaultConfig(),
    config.WithLogger(myLogger), // Required
)
```

### 2. Use Trace IDs for Distributed Tracing

Extract trace IDs in handlers for logging and debugging:

```go
func myHandler(c *gin.Context) {
    traceID := middleware.GetTraceID(c)
    logger.Info("Processing request", models.Fields{
        "trace_id": traceID,
        "user_id": c.GetString("user_id"),
    })
}
```

### 3. Configure CORS for Production

Avoid using wildcard origins in production:

```go
// ❌ Bad for production
config.WithCORS([]string{"*"})

// ✅ Good for production
config.WithCORS([]string{
    "https://app.example.com",
    "https://admin.example.com",
})
```

### 4. Use Release Mode in Production

Set Gin to release mode for better performance:

```go
config.WithMode("release")  // Disables verbose logging
```

### 5. Set Appropriate Timeouts

Configure timeouts based on your application needs:

```go
config.WithReadTimeout(30 * time.Second)
config.WithWriteTimeout(30 * time.Second)
config.WithIdleTimeout(60 * time.Second)
```

### 6. Organize Routes with Groups

Use route groups for better organization:

```go
api := platform.Group("/api/v1")
admin := platform.Group("/admin")

// Each group can have its own middleware
admin.Use(adminAuthMiddleware())
```

## Architecture

### Middleware Chain

```
Request
  ↓
TraceID Middleware (generates/extracts trace ID)
  ↓
CORS Middleware (handles cross-origin requests)
  ↓
Recovery Middleware (panic recovery with logging)
  ↓
Logger Middleware (request/response logging)
  ↓
Custom Middleware (your middleware)
  ↓
Handler (your route handler)
  ↓
Response
```

### Project Structure

```
http-platform-go/
├── config/          # Configuration with functional options
├── middleware/      # Built-in middleware (trace, cors, recovery, logger)
├── internal/        # Internal implementations (gin adapter)
├── examples/        # Usage examples
├── platform.go      # Main Platform type and API
├── types.go         # Public interfaces
└── errors.go        # Custom error types
```

## Error Handling

### Configuration and Runtime Errors

The library provides custom error types for platform configuration and runtime operations:

```go
platform, err := httpplatform.New(config.DefaultConfig())
if err != nil {
    // Check error type
    if configErr, ok := err.(*httpplatform.ConfigError); ok {
        log.Printf("Configuration error: %v", configErr)
    }
}
```

**Platform Errors:**
- `ConfigError` - Configuration validation errors
- `RuntimeError` - Runtime operation errors
- `ErrAlreadyStarted` - Platform already started
- `ErrNotStarted` - Platform not started
- `ErrNilLogger` - Logger not provided

### HTTP Domain Errors

For handling HTTP request errors in your handlers, use the **Error Handler Middleware** which provides structured error responses for common HTTP error scenarios.

See the [Error Handler Middleware](#error-handler-middleware) section for detailed documentation on:
- Available error types (NotFoundError, UnauthorizedError, ConflictError, etc.)
- How to use domain errors in your handlers
- Error response format
- Validation error handling

## Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [gin-contrib/cors](https://github.com/gin-contrib/cors) - CORS middleware
- [google/uuid](https://github.com/google/uuid) - UUID generation
- [go-playground/validator](https://github.com/go-playground/validator) - Struct validation
- [edaniel30/loki-logger-go](https://github.com/edaniel30/loki-logger-go) - Loki logger

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Credits

Created and maintained by [Daniel Rivera](https://github.com/edaniel30).

Inspired by:
- [guardian-auth](https://github.com/edaniel30/guardian-auth) - Authentication patterns
- [loki-logger-go](https://github.com/edaniel30/loki-logger-go) - Logger integration
