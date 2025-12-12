# CORS Middleware

The CORS middleware handles Cross-Origin Resource Sharing, allowing your API to be accessed from web applications running on different domains.

## What It Does

The CORS middleware helps your application:

- **Allow cross-origin requests**: Enable browsers to access your API from different domains
- **Control access methods**: Specify which HTTP methods are allowed
- **Manage headers**: Define which headers can be sent and exposed
- **Handle credentials**: Control whether cookies and authentication headers can be sent
- **Enforce CORS specification**: Automatically enforces security rules (e.g., wildcard origin cannot use credentials)
- **Handle preflight requests**: Automatically responds to OPTIONS requests

## Components

### 1. CORS Middleware

**Purpose**: Adds appropriate CORS headers to responses and handles preflight requests.

**Enabled by default** - runs as the fourth middleware in the chain.

**How it works**:
- Browser sends OPTIONS preflight (for complex requests)
- Middleware validates origin, method, and headers
- Adds CORS headers to response
- Browser validates and sends actual request

### 2. Configuration Options

**AllowedOrigins** - Which domains can access your API
```go
cfg.AllowedOrigins = []string{"*"} // All origins (default)
cfg.AllowedOrigins = []string{"https://myapp.com"} // Specific domain
```

**AllowedMethods** - Which HTTP methods are allowed
```go
cfg.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"} // Default
```

**AllowedHeaders** - Which request headers are allowed
```go
cfg.AllowedHeaders = []string{"*"} // All headers (default)
cfg.AllowedHeaders = []string{"Authorization", "Content-Type"} // Specific headers
```

**ExposedHeaders** - Which response headers can be read by browser
```go
cfg.ExposedHeaders = []string{"Content-Length", "X-Trace-Id"} // Default
```

**AllowCredentials** - Whether cookies/auth can be sent
```go
cfg.AllowCredentials = false // Default (required with wildcard origin)
cfg.AllowCredentials = true  // Only valid with specific origins
```

**MaxAge** - How long browser caches preflight response
```go
cfg.MaxAge = 12 * time.Hour // Default
```

## Configuration

### Default Configuration (Public API)

```go
cfg := httpplatform.DefaultConfig()
// AllowedOrigins: ["*"]
// AllowedMethods: ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"]
// AllowedHeaders: ["*"]
// ExposedHeaders: ["Content-Length", "X-Trace-Id"]
// AllowCredentials: false
// MaxAge: 12 hours
```

### Custom Configuration 

```go
cfg := httpplatform.DefaultConfig()
cfg.AllowedOrigins = []string{
    "https://myapp.com",
    "https://admin.myapp.com",
}
cfg.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE"}
cfg.AllowedHeaders = []string{"Authorization", "Content-Type"}
cfg.AllowCredentials = true // Can be true with specific origins
cfg.MaxAge = 24 * time.Hour

platform, _ := httpplatform.New(cfg)
```

## Usage Examples

### Example 1: Public API (Default)

```go
cfg := httpplatform.DefaultConfig()
platform, _ := httpplatform.New(cfg)
// Allows any domain to access
```

### Example 2: Authentication with Credentials

```go
cfg := httpplatform.DefaultConfig()
cfg.AllowedOrigins = []string{"https://myapp.com"}
cfg.AllowCredentials = true // Can use cookies/auth
cfg.AllowedHeaders = []string{"Authorization", "Content-Type"}

platform, _ := httpplatform.New(cfg)
```

### Example 3: Development (Multiple Frontends)

```go
cfg := httpplatform.DefaultConfig()
cfg.AllowedOrigins = []string{
    "http://localhost:3000",    // React
    "http://localhost:5173",    // Vite
}
cfg.AllowCredentials = true

platform, _ := httpplatform.New(cfg)
```

### Example 4: Read-Only Public API

```go
cfg := httpplatform.DefaultConfig()
cfg.AllowedOrigins = []string{"*"}
cfg.AllowedMethods = []string{"GET", "HEAD", "OPTIONS"}

platform, _ := httpplatform.New(cfg)
```

## CORS Specification Rules

### Rule 1: Wildcard Cannot Use Credentials

**Important**: When `AllowedOrigins` is `["*"]`, `AllowCredentials` must be `false`.

```go
// ❌ Invalid (platform validates and rejects)
cfg.AllowedOrigins = []string{"*"}
cfg.AllowCredentials = true

// ✅ Valid
cfg.AllowedOrigins = []string{"https://myapp.com"}
cfg.AllowCredentials = true
```

### Rule 2: Preflight Requests

For complex requests (custom headers, credentials), browsers send OPTIONS preflight:

```
Browser → OPTIONS /api/users
         Access-Control-Request-Method: POST

Server → 200 OK
         Access-Control-Allow-Origin: https://myapp.com
         Access-Control-Allow-Methods: POST

Browser → POST /api/users (actual request)
```

The middleware handles this automatically.

## When to Use

### Enable CORS when:
- Building APIs accessed from web browsers
- Supporting multiple frontend applications
- Creating public APIs

### Disable CORS when:
- Internal microservices (no browser access)
- Server-to-server APIs
- gRPC-only services

## Best Practices

**1. Use specific origins in production:**
```go
// ✅ Good - secure
cfg.AllowedOrigins = []string{"https://myapp.com"}

// ❌ Bad - allows any domain
cfg.AllowedOrigins = []string{"*"}
```

**2. Only allow necessary methods:**
```go
// ✅ Good
cfg.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE"}

// ❌ Too permissive
cfg.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "TRACE"}
```

**3. Restrict headers in production:**
```go
// ✅ Good
cfg.AllowedHeaders = []string{"Authorization", "Content-Type"}

// ❌ Less secure
cfg.AllowedHeaders = []string{"*"}
```

**4. Use credentials only when needed:**
```go
// ✅ Good - only for auth endpoints
cfg.AllowCredentials = true

// ❌ Bad - when not using cookies
cfg.AllowCredentials = true // Unnecessarily permissive
```

**5. Adjust MaxAge for environment:**
```go
// Development - short cache
cfg.MaxAge = 1 * time.Minute

// Production - longer cache
cfg.MaxAge = 24 * time.Hour
```

## Configuration

Default (enabled):
```go
cfg := httpplatform.DefaultConfig()
// EnableCORS is true by default
```

To disable:
```go
platform, _ := httpplatform.New(cfg,
    httpplatform.WithoutCORS(),
)
```
