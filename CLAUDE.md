# http-platform-go - Architecture & Development Guide

**Audience**: AI assistants and developers working on this codebase
**Purpose**: Architectural context, design patterns, and development guidelines
**Note**: For feature documentation, see `docs/`. For usage, see `README.md` and `examples/`.

## Project Overview

Production-ready HTTP server platform for Go built on Gin with automatic middleware setup, distributed tracing, and graceful shutdown.

**Tech Stack:**
- Go 1.25.7+
- `gin-gonic/gin` - HTTP framework
- `go.opentelemetry.io/otel` - Distributed tracing (optional)
- `gin-contrib/cors` - CORS middleware
- `go-playground/validator` - Request validation

**Typical Workflow:**
```go
platform, _ := httpplatform.New(httpplatform.DefaultConfig(), options...)
platform.GET("/route", handler)
platform.Start(ctx)  // Blocks until shutdown signal
```

See `examples/basic_server/` and `examples/full_featured/` for complete examples.

## Core Architecture

### Design Pattern: Functional Options

Primary configuration pattern throughout the codebase:

```go
// From config.go
type Option func(*Config)

func WithPort(port int) Option {
    return func(c *Config) {
        c.Port = port
    }
}

// Usage
platform, err := httpplatform.New(
    httpplatform.DefaultConfig(),
    httpplatform.WithPort(3000),
    httpplatform.WithLogger(myLogger),
)
```

**Benefits:**
- Backward compatible (new options don't break existing code)
- Self-documenting API
- Optional parameters with sensible defaults
- Type-safe configuration

**Validation:** All options validated in `Config.validate()` before platform creation.

### Package Structure: Export Interfaces, Hide Implementations

```
httpplatform/              # Public API only
├── platform.go           # Platform type, lifecycle methods
├── config.go             # Configuration + functional options
├── errors.go             # Public error types (ConfigError, RuntimeError)
├── logger.go             # Logger interface (public) + NewDefaultLogger()
├── utils.go              # Utility functions
└── internal/             # Implementation details (not importable)
    ├── logger/           # Logger implementation (DefaultLogger, helpers)
    ├── middleware/       # All middleware implementations
    ├── router/           # Gin wrapper/abstraction
    ├── telemetry/        # OpenTelemetry setup
    └── constants/        # Internal constants
```

**Key Principle:** Users interact with `httpplatform` package only. Implementation details are encapsulated in `internal/`.

**Logger Design:**
- **Public API** (`logger.go`): Interface definition that users can implement
- **Implementation** (`internal/logger/`): DefaultLogger and internal helpers
- **Reason**: Allows users to create custom logger adapters without importing internal packages

### Middleware Chain: Fixed Execution Order

**From `internal/router/router.go:87-110`:**

```go
// Order is CRITICAL - each position has a specific reason
// 1. TraceID - establishes request identity for all subsequent logging
if cfg.EnableTraceID {
    engine.Use(middleware.TraceID())
}

// 2. ErrorHandler - must be early to catch panics from all other middleware
engine.Use(middleware.ErrorHandler(cfg.Logger))

// 3. ContextCancellation - detect client disconnects early
if cfg.EnableContextCancellation {
    engine.Use(middleware.ContextCancellation())
}

// 4. CORS - handle CORS before business logic
if cfg.CORS != nil {
    engine.Use(middleware.CORS(*cfg.CORS))
}

// 5. Telemetry - traces the full request processing
if cfg.ServiceName != "" {
    engine.Use(middleware.Telemetry(cfg.ServiceName))
}

// 6. Logger - logs AFTER all processing with final status
if cfg.EnableLogger {
    engine.Use(middleware.BasicLogger(cfg.Logger))
}
```

**Why This Order:**
- **TraceID first**: All logs need correlation ID
- **ErrorHandler second**: Must catch panics from everything else
- **ContextCancellation**: Early detection saves processing
- **CORS**: Handles preflight before business logic
- **Telemetry**: Traces complete request lifecycle
- **Logger last**: Captures final state (status, duration, errors)

**⚠️ CRITICAL:** This order cannot be changed without breaking assumptions. Each middleware depends on previous ones.

## Key Design Patterns

### 1. Factory Pattern
- `New()` - Creates Platform instances
- `NewGinRouter()` - Creates router with middleware chain
- `NewDefaultLogger()` - Creates stdout logger

### 2. Wrapper/Adapter Pattern
**From `internal/router/router.go:41-58`:**

```go
type GinRouter struct {
    engine    *gin.Engine      // Wrapped Gin engine
    baseGroup *gin.RouterGroup // Optional for BasePath
}

func (r *GinRouter) getRouterGroup() gin.IRouter {
    if r.baseGroup != nil {
        return r.baseGroup   // Routes under base path
    }
    return r.engine          // Routes at root
}
```

**Purpose:**
- Hides Gin internals from users
- Allows BasePath feature transparently
- Could swap Gin for another router in future

### 3. Strategy Pattern
**From `logger.go` (public API):**

```go
type Logger interface {
    Info(ctx context.Context, msg string, fields Fields)
    Error(ctx context.Context, msg string, fields Fields)
    Warn(ctx context.Context, msg string, fields Fields)
    Debug(ctx context.Context, msg string, fields Fields)
    Close() error
}

// Fields type removed - use map[string]any directly  // Type alias for convenience
```

**Location:**
- **Public API**: `logger.go` (root package) - Interface and Fields type
- **Implementation**: `internal/logger/logger.go` - DefaultLogger and helpers
- **Usage**: Users can import and implement the interface directly

**Implementations:**
- `DefaultLogger` (stdlib log, stdout) - via `httpplatform.NewDefaultLogger()`
- User can provide: Loki, Zap, Logrus, etc. by implementing `httpplatform.Logger`

**Example custom adapter:**
```go
type MyAdapter struct { logger *mylogger.Logger }

func (a *MyAdapter) Info(ctx context.Context, msg string, fields map[string]any) {
    a.logger.Info(msg, convertFields(fields))
}
// ... implement other methods
```

### 4. Type-Based Error Mapping

**From `internal/middleware/error_handler.go:97-170`:**

```go
func mapErrorToApiError(err error) (*ApiError, Fields) {
    switch e := err.(type) {
    case *NotFoundError:
        return NewApiError(e.Error(), 404, "NotFoundError"), fields
    case *UnauthorizedError:
        return NewApiError(e.Error(), 401, "UnauthorizedError"), fields
    case validator.ValidationErrors:
        return handleValidationErrors(e), fields
    case *json.SyntaxError:
        return NewApiError("Invalid JSON syntax", 400, "JSONSyntaxError"), fields
    default:
        return mapStandardError(err, fields)
    }
}
```

**Error Types:**
- **Public** (`httpplatform` package): `ConfigError`, `RuntimeError`
- **HTTP Domain** (`middleware` package): `NotFoundError`, `UnauthorizedError`, etc.

See `docs/ERROR_HANDLER.md` for complete error types list.

## Critical Implementation Details

### Logger Lifecycle: User Owns It

**From `platform.go:206-229`:**

```go
func (p *Platform) shutdownComponents() {
    // Close telemetry if enabled
    if p.telemetryManager != nil {
        if err := p.telemetryManager.Shutdown(); err != nil {
            p.config.Logger.Error(ctx, "Failed to shutdown telemetry", Fields{"error": err.Error()})
        }
    }

    // INTENTIONALLY NOT CLOSING LOGGER
    // Logger may be shared across components
    // User is responsible for logger lifecycle
}
```

**⚠️ IMPORTANT:** Platform does NOT call `logger.Close()`. Reasons:
1. Logger may be shared across multiple components
2. Logger lifecycle is caller's responsibility
3. Prevents panic if Start() called multiple times
4. Follows Go idiom: "Who creates it, closes it"

### Thread Safety

**From `platform.go:54-62`:**

```go
type Platform struct {
    config           Config
    router           *router.GinRouter
    server           *http.Server
    telemetryManager *telemetry.Manager
    mu               sync.RWMutex  // Protects 'started'
    started          bool
}
```

Uses `sync.RWMutex` to prevent:
- Multiple simultaneous `Start()` calls
- Configuration changes after startup
- Race conditions in lifecycle management

### CORS Specification Enforcement

**From `config.go:229-236`:**

```go
func (c *Config) validateCORS() error {
    if c.CORS == nil {
        return nil
    }
    // CORS spec: Cannot use credentials with wildcard origin
    if c.CORS.AllowCredentials && slices.Contains(c.CORS.AllowedOrigins, "*") {
        return NewConfigError("CORS: cannot use credentials with wildcard origin '*'")
    }
    return nil
}
```

**Why:** CORS specification explicitly prohibits this combination for security reasons.

## Code Organization Conventions

### Naming Conventions

| Element | Convention | Examples |
|---------|-----------|----------|
| **Files** | `snake_case.go` | `error_handler.go`, `http_errors.go` |
| **Tests** | `*_test.go` | `platform_test.go`, `config_test.go` |
| **Docs** | `UPPERCASE.md` | `README.md`, `CORS.md`, `LOGGER.md` |
| **Exported types** | `PascalCase` | `Platform`, `Config`, `Logger` |
| **Unexported types** | `camelCase` or `PascalCase` (internal) | `ginRouter`, `telemetryManager` |
| **Exported functions** | `PascalCase` | `New()`, `Start()`, `WithPort()` |
| **Unexported functions** | `camelCase` | `validateNotStarted()`, `createHTTPServer()` |

### File Organization Within Packages

**Typical structure:**
```go
// 1. Package declaration
package httpplatform

// 2. Imports (ordered by goimports)
import (
    "context"
    "net/http"

    "github.com/gin-gonic/gin"

    "github.com/edaniel30/http-platform-go/internal/router"
)

// 3. Type definitions
type Platform struct { ... }

// 4. Constructors/factories
func New(...) (*Platform, error) { ... }

// 5. Public methods (alphabetically)
func (p *Platform) DELETE(...) { ... }
func (p *Platform) GET(...) { ... }
func (p *Platform) Start(...) { ... }

// 6. Private methods (alphabetically)
func (p *Platform) createHTTPServer() *http.Server { ... }
func (p *Platform) validateNotStarted() error { ... }
```

### Import Order (enforced by `goimports`)

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "net/http"

    // 2. External dependencies
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    // 3. Internal packages (project-local)
    "github.com/edaniel30/http-platform-go/internal/middleware"
    "github.com/edaniel30/http-platform-go/internal/router"
)
```

Configured in `.golangci.yml`:
```yaml
goimports:
  local-prefixes: github.com/edaniel30/http-platform-go
```

## Testing Strategy

### Test Patterns

**1. Table-Driven Tests (for similar cases):**

```go
tests := []struct {
    name      string
    config    Config
    wantError bool
    errorMsg  string
}{
    {name: "valid config", config: DefaultConfig(), wantError: false},
    {name: "invalid port", config: Config{Port: 0}, wantError: true, errorMsg: "Port must be between"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        err := tt.config.validate()
        // assertions
    })
}
```

**2. Subtests with `t.Run()` (for different scenarios):**

```go
func TestNew(t *testing.T) {
    t.Run("successful creation with defaults", func(t *testing.T) {...})
    t.Run("fails with invalid config", func(t *testing.T) {...})
    t.Run("applies options correctly", func(t *testing.T) {...})
}
```

**3. HTTP Testing:**

```go
router := gin.New()
router.Use(middleware.ErrorHandler(logger))
router.GET("/test", handler)

w := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/test", nil)
router.ServeHTTP(w, req)

assert.Equal(t, 200, w.Code)
```

### Coverage Requirements

**From `Makefile`:**
```makefile
COVERAGE_THRESHOLD=80
```

**Excluded from coverage:**
- `/internal/mocks` - Auto-generated mocks
- `/examples` - Example applications
- `/internal/telemetry` - Requires external OTLP endpoint

See `.coverignore` and `COVERAGE_EXCEPTIONS.md` for detailed justification.

**Run coverage:**
```bash
make test-coverage        # Check threshold
make test-coverage-html   # Visual coverage report
```

## Development Guidelines

### Adding New Middleware

**Checklist:**
1. Implement in `internal/middleware/your_middleware.go`
2. Add configuration field to `RouterConfig` in `internal/router/router.go`
3. Add to middleware chain in `NewGinRouter()` - **carefully consider position**
4. Justify order in code comments
5. Create comprehensive docs in `docs/YOUR_MIDDLEWARE.md`
6. Write tests in `internal/middleware/your_middleware_test.go`
7. Update README.md middleware table
8. Update this CLAUDE.md if architectural implications

**Example:**
```go
// internal/middleware/your_middleware.go
func YourMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Pre-processing
        c.Next()
        // Post-processing
    }
}

// internal/router/router.go
if cfg.EnableYourFeature {
    engine.Use(middleware.YourMiddleware())
}
```

### Adding New Configuration Options

**Checklist:**
1. Add field to `Config` struct in `config.go`
2. Document field with comment explaining purpose and default
3. Create `WithXXX()` option function
4. Add validation in `Config.validate()` if needed
5. Write table-driven tests in `config_test.go`
6. Update README.md configuration section
7. Update default in `DefaultConfig()` if applicable

**Example:**
```go
// config.go
type Config struct {
    // ... existing fields

    // YourFeature enables XYZ functionality (default: false)
    YourFeature bool
}

func WithYourFeature(enabled bool) Option {
    return func(c *Config) {
        c.YourFeature = enabled
    }
}

// Validation (if needed)
func (c *Config) validate() error {
    // ... existing validations

    if c.YourFeature && c.SomeRequiredField == "" {
        return NewConfigError("YourFeature requires SomeRequiredField")
    }

    return nil
}
```

### Adding New Error Types

**Checklist:**
1. Define struct in `internal/middleware/http_errors.go`
2. Implement `Error()` method
3. Create constructor function `NewXXXError(message string)`
4. Add case in `mapErrorToApiError()` in `error_handler.go`
5. Write tests in `error_handler_test.go`
6. Document in `docs/ERROR_HANDLER.md`

**Example:**
```go
// internal/middleware/http_errors.go
type YourError struct {
    message string
}

func (e *YourError) Error() string {
    return e.message
}

func NewYourError(message string) *YourError {
    return &YourError{message: message}
}

// internal/middleware/error_handler.go
func mapErrorToApiError(err error) (*ApiError, Fields) {
    switch e := err.(type) {
    // ... existing cases
    case *YourError:
        return NewApiError(e.Error(), 418, "YourError"), additionalFields
    // ...
    }
}
```

### Adding New Utility Functions

**Checklist:**
1. Add to `utils.go` in root package
2. Document with clear godoc comment
3. Keep functions pure (no side effects)
4. Add tests in `utils_test.go`
5. Update README.md utility section if user-facing

**Example:**
```go
// utils.go

// ExtractBearerToken extracts the Bearer token from Authorization header.
// Returns empty string if header is missing or not in "Bearer <token>" format.
func ExtractBearerToken(c *gin.Context) string {
    auth := c.GetHeader("Authorization")
    if strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    return ""
}
```

## Important Rules & Conventions

### ✅ Do

- **Use `DefaultConfig()` as starting point** - Has sensible defaults
- **Always validate configuration** - In `validate()` method
- **Keep internal/ packages internal** - Don't expose implementation
- **Document middleware order** - When adding to chain
- **Use functional options** - For extensibility
- **Write table-driven tests** - For configuration and mappings
- **Use subtests** - For organizing test scenarios
- **Reference docs/** - Instead of duplicating in code comments
- **Close resources you create** - User's logger, user's resources

### ❌ Don't

- **Don't close user's logger** - They manage lifecycle
- **Don't change middleware order** - Without thorough analysis
- **Don't expose gin.Engine** - Keep wrapped in GinRouter
- **Don't use wildcard CORS with credentials** - Spec violation, will fail validation
- **Don't skip validation** - Always call `validate()` on configs
- **Don't break backward compatibility** - Use new options instead
- **Don't add to root package unnecessarily** - Prefer internal/
- **Don't test infrastructure code** - Server lifecycle, signals (see COVERAGE_EXCEPTIONS.md)

### Security Considerations

1. **CORS validation** - Enforces spec at config level
2. **Panic recovery** - ErrorHandler catches and logs all panics
3. **Context cancellation** - Prevents wasted work on disconnected clients
4. **Timeout enforcement** - ReadTimeout, WriteTimeout, IdleTimeout
5. **Header validation** - Gin's built-in + go-playground/validator
6. **No secrets in logs** - Logger doesn't auto-log headers

### Performance Considerations

1. **Middleware order** - Optimized for early exit (context cancellation early)
2. **Lazy initialization** - Telemetry only if configured
3. **Connection pooling** - Gin's built-in HTTP/1.1 keep-alive
4. **Graceful shutdown** - 5-second timeout to drain connections
5. **Logger buffering** - Implementation-dependent (user's choice)

## Common Commands

### Development
```bash
# Setup
make setup                    # Install pre-commit hooks

# Testing
make test                     # Run all tests
make test-coverage            # Check 80% threshold
make test-coverage-html       # Visual coverage report
make test-race               # Race condition detection

# Cleanup
make clean                   # Remove coverage files

# Linting (automatic via pre-commit)
golangci-lint run            # Manual linting
```

### Go Commands
```bash
# Dependencies
go mod download              # Download dependencies
go mod tidy                  # Clean up go.mod/go.sum

# Testing
go test ./...                # Run all tests
go test -v ./...             # Verbose output
go test -race ./...          # With race detector
go test -coverprofile=c.out ./...  # Generate coverage

# Validation
go vet ./...                 # Static analysis
go fmt ./...                 # Format code
```

### CI/CD
Runs automatically on push to `main`/`develop`:
- Go vet
- Tests with coverage (≥80% required)
- Security scan (govulncheck)
- Codecov upload

## Documentation Strategy

### Documentation Locations

| Topic | Location | Purpose |
|-------|----------|---------|
| **Architecture** | `CLAUDE.md` (this file) | Design patterns, conventions, development guidelines |
| **User Guide** | `README.md` | Quick start, API reference, configuration |
| **Feature Details** | `docs/*.md` | Deep dive into each middleware (CORS, Logger, etc.) |
| **Examples** | `examples/` | Working code demonstrating usage |
| **Coverage** | `COVERAGE_EXCEPTIONS.md` | Justification for untested code |

### Documentation Principles

1. **DRY (Don't Repeat Yourself)**
   - README.md → What/How to use
   - docs/*.md → Deep dive features
   - CLAUDE.md → Why/Architecture
   - No duplication between them

2. **Link Don't Duplicate**
   ```markdown
   For CORS configuration details, see [docs/CORS.md](docs/CORS.md).
   ```

3. **Code Examples from Real Code**
   - Reference actual file paths
   - Use real code snippets
   - Keep examples up-to-date

4. **Update Together**
   - Add feature → Update README + docs/ + CLAUDE.md
   - Change architecture → Update CLAUDE.md
   - Add option → Update README configuration section

## Project Lifecycle

### Current State
- **Version**: Check git tags for current version
- **Go Version**: 1.25.7+
- **Stability**: Production-ready
- **Breaking Changes**: Use new major version

### Release Process
- Automated via `.github/workflows/auto-release.yml`
- Triggered by pushing git tags
- Semantic versioning (v1.0.0, v1.1.0, etc.)

### Future Considerations
- Middleware is extensible via functional options
- Router abstraction allows framework swap
- Logger interface supports any implementation
- Telemetry supports any OTLP-compatible backend

## Quick Reference

### File Location Guide

**Need to change:**
- Middleware behavior → `internal/middleware/`
- Router logic → `internal/router/router.go`
- Configuration → `config.go`
- Public API → `platform.go`
- Logger interface (public) → `logger.go`
- Logger implementation → `internal/logger/logger.go`
- Error types → `internal/middleware/http_errors.go`
- Telemetry → `internal/telemetry/telemetry.go`

**Need to document:**
- User-facing features → `README.md`
- Middleware details → `docs/`
- Architecture decisions → `CLAUDE.md` (this file)
- Untested code → `COVERAGE_EXCEPTIONS.md`

### Key Files Reference

| File | Purpose | Modify When |
|------|---------|-------------|
| `platform.go:63-126` | Platform creation, validation | Adding platform-level features |
| `platform.go:128-259` | Server lifecycle (Start/Stop) | Changing startup/shutdown behavior |
| `config.go:29-78` | Configuration struct | Adding new config options |
| `config.go:80-249` | Functional options + validation | Adding new options or validation rules |
| `internal/router/router.go:68-110` | Middleware chain setup | Adding/reordering middleware |
| `internal/middleware/error_handler.go:97-170` | Error type mapping | Adding new HTTP error types |
| `logger.go` | Logger interface (public API) | Changing logging contract (users depend on this) |
| `internal/logger/logger.go` | DefaultLogger implementation + helpers | Internal logging implementation |

### Import Paths

```go
// Public API
import "github.com/edaniel30/http-platform-go"

// Middleware (for error types and utilities)
import "github.com/edaniel30/http-platform-go/internal/middleware"

// Everything else is internal - don't import
```

## Getting Help

1. **Feature usage** → Check `README.md` and `examples/`
2. **Middleware details** → Check `docs/*.md`
3. **Architecture questions** → This file (CLAUDE.md)
4. **Contributing** → Follow guidelines in this file
5. **Issues** → GitHub issues

---

**Last Updated**: This file should be updated when:
- Adding new design patterns
- Changing architecture
- Updating development guidelines
- Adding important conventions
- Modifying critical implementation details

**For feature documentation updates**, modify the appropriate `docs/*.md` file instead.
