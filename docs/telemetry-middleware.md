# Telemetry Middleware

The Telemetry middleware provides OpenTelemetry distributed tracing for monitoring request flows across microservices.

## What It Does

The Telemetry middleware helps your application:

- **Trace distributed requests**: Track requests across multiple services with spans
- **Monitor performance**: Measure request duration, latency, and bottlenecks
- **Debug production issues**: Visualize request flow through your system
- **Integrate with observability tools**: Works with Jaeger, Zipkin, Datadog, and other OTLP-compatible backends

## Components

### 1. Telemetry Middleware

**Purpose**: Automatically creates OpenTelemetry spans for every HTTP request.

**Disabled by default** - must be explicitly enabled.

**How it works**:
- Intercepts each incoming request
- Creates a span with request metadata (method, path, status)
- Propagates trace context to downstream services
- Records span duration and attributes
- Sends traces to OTLP endpoint

**Example flow**:
```
Request arrives
  ↓
Telemetry creates span: "GET /api/users"
  ↓
Handler processes request
  ↓
Span records: status=200, duration=45ms
  ↓
Span sent to OTLP endpoint (Jaeger, Datadog, etc.)
```

### 2. Span Attributes

The middleware automatically records these attributes:

- `http.method` - HTTP method (GET, POST, etc.)
- `http.route` - Request path (/api/users/:id)
- `http.status_code` - Response status (200, 404, 500, etc.)
- `http.request_content_length` - Request body size
- `http.response_content_length` - Response body size

## Configuration

### Enable Telemetry

```go
cfg := httpplatform.DefaultConfig()
cfg.EnableTelemetry = true
cfg.ServiceName = "user-service"
cfg.OTLPEndpoint = "localhost:4318" // Your OTLP collector
cfg.TelemetrySampleAll = true // Sample all traces

platform, _ := httpplatform.New(cfg)
```

### With Functional Options

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithTelemetry("user-service", "localhost:4318", true),
)
```

### Configuration Options

- **ServiceName**: Name of your service in traces (e.g., "user-service", "api-gateway")
- **OTLPEndpoint**: OTLP collector endpoint (e.g., "localhost:4318" for Jaeger/Datadog)
- **TelemetrySampleAll**:
  - `true` - Sample all traces (recommended for development/staging)
  - `false` - Use default sampling (recommended for high-traffic production)

## When to Use

### Enable Telemetry when:
- Running in production or staging environments
- Debugging performance issues across microservices
- Monitoring distributed systems
- Tracking request flows through multiple services

### Disable Telemetry when:
- Local development (unless testing tracing)
- Running tests
- Services that don't need observability (internal tools)


## Configuration

Default (disabled):
```go
cfg := httpplatform.DefaultConfig()
// EnableTelemetry is false by default
```

To enable:
```go
cfg.EnableTelemetry = true
cfg.ServiceName = "my-service"
cfg.OTLPEndpoint = "localhost:4318"
```

To disable:
```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutTelemetry(),
)
```
