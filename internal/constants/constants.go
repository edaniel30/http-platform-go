package constants

import "time"

// HTTP Server Constants (internal use only)

// ShutdownTimeout is the default timeout for graceful server shutdown
const ShutdownTimeout = 5 * time.Second

// DefaultMaxHeaderBytes is the default maximum size for request headers (1 MB)
const DefaultMaxHeaderBytes = 1 << 20 // 1 MB

// StatusClientClosedRequest is a non-standard HTTP status code (used by nginx)
// indicating that the client closed the connection before the server could send a response
const StatusClientClosedRequest = 499

// Gin Mode Constants

// GinModeDebug enables debug mode with verbose logging
const GinModeDebug = "debug"

// GinModeRelease enables production mode with optimized performance
const GinModeRelease = "release"

// GinModeTest enables test mode (minimal output)
const GinModeTest = "test"

