# HTTP Platform Go - Examples

This directory contains practical examples demonstrating how to use the `http-platform-go` library. The examples progress from basic to advanced usage.

## Prerequisites

- Go 1.21+
- For telemetry example: Running OpenTelemetry collector (optional)

## Running Examples

Navigate to any example directory and run:

```bash
cd examples/basic_server
go run main.go
```

All examples will start an HTTP server on port 8080 by default.

---

## Examples Overview

### 1. basic_server

**Difficulty:** ⭐ Beginner

Minimal HTTP server with a single health check endpoint.

**Features:**
- Basic server setup
- Single GET endpoint
- Default configuration

**Run:**
```bash
cd examples/basic_server
go run main.go
# Visit: http://localhost:8080/health
```

---

### 2. with_cors

**Difficulty:** ⭐ Beginner

HTTP server with CORS configuration for cross-origin requests.

**Features:**
- CORS middleware enabled
- Custom allowed origins
- Credentials support

**Run:**
```bash
cd examples/with_cors
go run main.go
# Test CORS: curl -H "Origin: https://example.com" http://localhost:8080/api/users
```

---

### 3. with_telemetry

**Difficulty:** ⭐⭐ Intermediate

HTTP server with OpenTelemetry distributed tracing.

**Features:**
- OpenTelemetry integration
- Automatic request tracing
- Trace ID propagation
- OTLP exporter

**Environment Variables:**
```bash
export OTEL_ENDPOINT="localhost:4318"  # OpenTelemetry collector endpoint
```

**Run:**
```bash
cd examples/with_telemetry
go run main.go
```

**View traces:**
- Jaeger UI: http://localhost:16686
- Datadog APM: https://app.datadoghq.com/apm/traces

---

### 4. with_routes

**Difficulty:** ⭐⭐ Intermediate

HTTP server demonstrating route organization and grouping.

**Features:**
- Route groups with prefixes
- Multiple HTTP methods
- Path parameters
- Query parameters
- Utility functions (QueryParamsToMap, HeadersToMap)

**Run:**
```bash
cd examples/with_routes
go run main.go

# Test endpoints:
curl http://localhost:8080/api/v1/users
curl http://localhost:8080/api/v1/users/123
curl -X POST http://localhost:8080/api/v1/users -d '{"name":"John"}'
curl http://localhost:8080/api/v1/users/search?name=John&age=30
```

---

### 5. full_featured

**Difficulty:** ⭐⭐⭐ Advanced

Production-ready HTTP server with all features enabled.

**Features:**
- OpenTelemetry tracing
- CORS configuration
- Custom error handling
- Graceful shutdown
- All middleware enabled
- Route groups
- Health checks

**Environment Variables:**
```bash
# Telemetry (optional)
export OTEL_ENDPOINT="localhost:4318"

# Server
export PORT="8080"
export ENVIRONMENT="production"
```

**Run:**
```bash
cd examples/full_featured
go run main.go
```

---

## Common Patterns

### Testing Endpoints

```bash
# Health check
curl http://localhost:8080/health

# With headers
curl -H "X-Request-ID: abc123" http://localhost:8080/api/users

# POST with JSON
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John","email":"john@example.com"}'

# With query parameters
curl "http://localhost:8080/api/users?name=John&age=30"
```

### Graceful Shutdown

All examples support graceful shutdown with `Ctrl+C`:
- Server stops accepting new connections
- Existing requests complete within 5 seconds
- Resources are cleaned up properly

---

## Development Setup

### Running OpenTelemetry Collector (for telemetry examples)

```bash
docker run -d --name otel-collector \
  -p 4318:4318 \
  otel/opentelemetry-collector:latest
```

### Running Jaeger (for viewing traces)

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

---

## Troubleshooting

### Port already in use

If port 8080 is already in use, change it in the example:

```go
httpplatform.New(cfg,
    httpplatform.WithPort(3000),  // Change port
)
```

### Logger connection issues

Run in console-only mode if Loki is not available:

```go
logger, _ := loki.New(
    loki.DefaultConfig(),
    loki.WithOnlyConsole(true),  // Console only
)
```

### Telemetry not working

Ensure OTLP collector is running and accessible:

```bash
curl http://localhost:4318/v1/traces
```

---

## Next Steps

- Read the [main README](../README.md) for full API documentation
- Check [middleware documentation](../middleware/) for custom middleware
- Review [error handling patterns](../middleware/http_errors.go)

## Questions?

Open an issue at: https://github.com/edaniel30/http-platform-go/issues
