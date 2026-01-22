# Telemetry

OpenTelemetry distributed tracing for monitoring request flows across microservices.

## Quick Start

**Disabled by default** - must be explicitly enabled:

```go
cfg := httpplatform.DefaultConfig()

platform, _ := httpplatform.New(cfg,
    httpplatform.WithTelemetry("user-service", "localhost:4318", true),
)
```

## Features

- **Distributed tracing** - Track requests across multiple services with spans
- **Performance monitoring** - Measure request duration, latency, and bottlenecks
- **Production debugging** - Visualize request flow through your system
- **Observability integration** - Works with Jaeger, Zipkin, Datadog, and OTLP-compatible backends

## Span Attributes

Automatically recorded for every HTTP request:

| Attribute | Description | Example |
|-----------|-------------|---------|
| `http.method` | HTTP method | `GET`, `POST` |
| `http.route` | Request path | `/api/users/:id` |
| `http.status_code` | Response status | `200`, `404`, `500` |
| `http.request_content_length` | Request body size | `1024` |
| `http.response_content_length` | Response body size | `2048` |

## Configuration

| Option | Description | Example |
|--------|-------------|---------|
| `ServiceName` | Service name in traces | `"user-service"`, `"api-gateway"` |
| `OTLPEndpoint` | OTLP collector endpoint | `"localhost:4318"` (Jaeger/Datadog) |
| `TelemetrySampleAll` | Sample all traces (true) or use default sampling (false) | `true` for dev/staging, `false` for production |

### Basic Setup

```go
cfg := httpplatform.DefaultConfig()
cfg.EnableTelemetry = true
cfg.ServiceName = "user-service"
cfg.OTLPEndpoint = "localhost:4318"
cfg.TelemetrySampleAll = true

platform, _ := httpplatform.New(cfg)
```

### With Functional Options

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithTelemetry(
        "user-service",     // ServiceName
        "localhost:4318",   // OTLPEndpoint
        true,               // SampleAll
    ),
)
```

### Disable Telemetry

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutTelemetry(),
)
```

## How It Works

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

## When to Use

| Scenario | Enable Telemetry | Notes |
|----------|------------------|-------|
| Production | ✅ Yes | Essential for monitoring |
| Staging | ✅ Yes | Debug performance issues |
| Microservices | ✅ Yes | Track distributed requests |
| Development | ⚠️ Optional | Enable when testing tracing |
| Testing | ❌ No | Unnecessary overhead |
| Single service | ⚠️ Optional | Useful for performance profiling |

## Sampling Strategies

### Development/Staging

```go
cfg.TelemetrySampleAll = true  // Capture all traces
```

Good for:
- Low traffic environments
- Debugging specific issues
- Complete visibility

### Production (High Traffic)

```go
cfg.TelemetrySampleAll = false  // Use default sampling
```

Good for:
- High-traffic services
- Reduced overhead
- Cost optimization

## Integration with TraceID

Telemetry middleware works seamlessly with TraceID middleware:

```go
cfg := httpplatform.DefaultConfig()

// Both enabled
platform, _ := httpplatform.New(cfg,
    httpplatform.WithTelemetry("my-service", "localhost:4318", true),
    // TraceID is enabled by default
)
```

**TraceID**: Simple correlation via `X-Trace-Id` header
**Telemetry**: Full distributed tracing with spans and attributes

## Backend Integration Examples

### Jaeger

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithTelemetry("user-service", "localhost:4318", true),
)
```

Run Jaeger:
```bash
docker run -d --name jaeger \
  -p 4318:4318 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest
```

Access UI: http://localhost:16686

### Datadog

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithTelemetry(
        "user-service",
        os.Getenv("DD_OTLP_ENDPOINT"), // Datadog OTLP endpoint
        false, // Use sampling for production
    ),
)
```

### Zipkin

```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithTelemetry("user-service", "localhost:9411", true),
)
```

## Best Practices

### ✅ Do

```go
// Enable in production/staging
cfg.EnableTelemetry = true
cfg.ServiceName = "user-service"  // Clear service name
cfg.OTLPEndpoint = os.Getenv("OTLP_ENDPOINT")

// Use sampling in high-traffic production
cfg.TelemetrySampleAll = false

// Set meaningful service names
cfg.ServiceName = "api-gateway"  // Not "service1"
```

### ❌ Don't

```go
// Don't enable in tests (unnecessary overhead)
cfg.EnableTelemetry = true  // in test setup

// Don't sample all in high-traffic production
cfg.TelemetrySampleAll = true  // on service with 10k req/s

// Don't use generic service names
cfg.ServiceName = "service"  // Too generic

// Don't hardcode endpoints
cfg.OTLPEndpoint = "localhost:4318"  // Use env vars
```

## Middleware Order

Telemetry runs **fifth** in the middleware chain:

```
1. TraceID       → Generates/extracts trace_id
2. ErrorHandler  → Catches errors
3. ContextCancel → Detects disconnects
4. CORS          → Handles preflight
5. Telemetry     → Creates spans ✓
6. Logger        → Logs requests
```

This order ensures:
- All spans have trace IDs
- Error spans are properly recorded
- CORS preflight requests are handled before tracing
- Logger captures telemetry overhead

## Troubleshooting

### No traces appearing

Check:
1. `EnableTelemetry = true`
2. OTLP endpoint is reachable
3. Backend is running (Jaeger, Datadog)
4. Firewall/network allows connections to port 4318

### High overhead

Solution:
```go
cfg.TelemetrySampleAll = false  // Reduce sampling
```

### Missing trace IDs in spans

Ensure TraceID middleware is enabled:
```go
// TraceID is enabled by default
// Don't call WithoutTraceID()
```
