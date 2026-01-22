# CORS (Cross-Origin Resource Sharing)

Enables browser-based applications from different domains to access your API.

## Quick Start

**Enabled by default** with permissive settings for public APIs:

```go
platform, _ := httpplatform.New(httpplatform.DefaultConfig())
// Allows all origins, methods, and headers
```

## Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `AllowedOrigins` | `["*"]` | Which domains can access the API |
| `AllowedMethods` | `["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"]` | Permitted HTTP methods |
| `AllowedHeaders` | `["*"]` | Which request headers are allowed |
| `ExposedHeaders` | `["Content-Length", "X-Trace-Id"]` | Headers browsers can read |
| `AllowCredentials` | `false` | Allow cookies/auth headers |
| `MaxAge` | `12h` | Preflight cache duration |

## Common Scenarios

### Public API (Default)

```go
cfg := httpplatform.DefaultConfig()
platform, _ := httpplatform.New(cfg)
// AllowedOrigins: ["*"] - any domain can access
```

### Authenticated API (Specific Origins)

```go
cfg := httpplatform.DefaultConfig()

platform, _ := httpplatform.New(cfg,
    httpplatform.WithCORSOrigins(
        "https://app.example.com",
        "https://admin.example.com",
    ),
    httpplatform.WithCORSCredentials(true), // Enable cookies/auth
)
```

### Development (Multiple Frontends)

```go
cfg := httpplatform.DefaultConfig()

platform, _ := httpplatform.New(cfg,
    httpplatform.WithCORSOrigins(
        "http://localhost:3000",  // React
        "http://localhost:5173",  // Vite
    ),
    httpplatform.WithCORSCredentials(true),
)
```

### Read-Only Public API

```go
cfg := httpplatform.DefaultConfig()

platform, _ := httpplatform.New(cfg,
    httpplatform.WithCORSMethods("GET", "HEAD", "OPTIONS"),
)
```

### Production API (Locked Down)

```go
cfg := httpplatform.DefaultConfig()

platform, _ := httpplatform.New(cfg,
    httpplatform.WithCORSOrigins("https://app.example.com"),
    httpplatform.WithCORSMethods("GET", "POST", "PUT", "DELETE"),
    httpplatform.WithCORSHeaders("Authorization", "Content-Type"),
    httpplatform.WithCORSCredentials(true),
)
```

## Security Rules

### ⚠️ Wildcard Cannot Use Credentials

CORS specification prohibits using wildcard origin with credentials:

```go
// ❌ INVALID - platform will reject this
cfg.CORS = &middleware.CORSConfig{
    AllowedOrigins:   []string{"*"},
    AllowCredentials: true, // Error!
}

// ✅ VALID - specific origin with credentials
cfg.CORS = &middleware.CORSConfig{
    AllowedOrigins:   []string{"https://app.example.com"},
    AllowCredentials: true,
}
```

### Preflight Requests

Browsers automatically send OPTIONS requests for complex requests:

```
OPTIONS /api/users
Access-Control-Request-Method: POST
Access-Control-Request-Headers: Authorization

→ 200 OK
  Access-Control-Allow-Origin: https://app.example.com
  Access-Control-Allow-Methods: POST
  Access-Control-Allow-Headers: Authorization
  Access-Control-Max-Age: 43200

POST /api/users (actual request proceeds)
```

Middleware handles this automatically.

## Best Practices

### Production vs Development

```go
// ✅ Development - permissive
platform, _ := httpplatform.New(cfg,
    httpplatform.WithCORSOrigins(
        "http://localhost:*", // All localhost ports
    ),
)

// ✅ Production - restrictive
platform, _ := httpplatform.New(cfg,
    httpplatform.WithCORSOrigins("https://app.example.com"),
    httpplatform.WithCORSHeaders("Authorization", "Content-Type"),
    httpplatform.WithCORSCredentials(true),
)
```

### MaxAge Tuning

```go
// Development - short cache (1 min)
httpplatform.WithCORS(&middleware.CORSConfig{
    MaxAge: 1 * time.Minute,
})

// Production - longer cache (24h)
httpplatform.WithCORS(&middleware.CORSConfig{
    MaxAge: 24 * time.Hour,
})
```

### Security Checklist

✅ **Do:**
- Use specific origins in production
- Only allow required methods
- Restrict headers to necessary ones
- Use credentials only when needed
- Set appropriate MaxAge for environment

❌ **Don't:**
- Use `["*"]` in production (unless truly public)
- Allow all methods unnecessarily
- Enable credentials with wildcard origin
- Allow all headers in production
- Set MaxAge too high in development

## When to Use CORS

| Use Case | Enable CORS |
|----------|-------------|
| Browser-based frontend | ✅ Yes |
| Mobile apps (native) | ❌ No |
| Server-to-server APIs | ❌ No |
| Public REST API | ✅ Yes |
| Internal microservices | ❌ No |
| GraphQL API (browser) | ✅ Yes |
| gRPC services | ❌ No |

## Disable CORS

For internal APIs not accessed by browsers:

```go
// CORS is nil by default, explicitly set to disable
cfg := httpplatform.DefaultConfig()
cfg.CORS = nil

platform, _ := httpplatform.New(cfg)
```
