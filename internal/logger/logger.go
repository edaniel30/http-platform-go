package logger

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// Logger interface is defined here to avoid import cycles.
// This is the internal version used by middleware packages.
// The public version is re-exported in the httpplatform package.
type Logger interface {
	Info(ctx context.Context, msg string, fields map[string]any)
	Error(ctx context.Context, msg string, fields map[string]any)
	Warn(ctx context.Context, msg string, fields map[string]any)
	Debug(ctx context.Context, msg string, fields map[string]any)
	Close() error
}

// DefaultLogger is a simple logger implementation using Go's standard log package.
// It's used as the default logger when no custom logger is provided.
// Writes to stdout with a "[http-platform]" prefix.
//
// Example:
//
//	logger := httpplatform.NewDefaultLogger()
//	defer logger.Close()
//
//	cfg := httpplatform.DefaultConfig()
//	platform, _ := httpplatform.New(cfg, httpplatform.WithLogger(logger))
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger creates a new default logger that writes to stdout.
// This is the default logger used by DefaultConfig().
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "[http-platform] ", log.LstdFlags),
	}
}

// NewTestLogger creates a new logger with a custom log.Logger for testing purposes.
// This allows tests to capture log output to a buffer.
func NewTestLogger(l *log.Logger) *DefaultLogger {
	return &DefaultLogger{
		logger: l,
	}
}

// Info logs an informational message with optional fields.
func (l *DefaultLogger) Info(ctx context.Context, msg string, fields map[string]any) {
	l.log("INFO", msg, fields)
}

// Error logs an error message with optional fields.
func (l *DefaultLogger) Error(ctx context.Context, msg string, fields map[string]any) {
	l.log("ERROR", msg, fields)
}

// Warn logs a warning message with optional fields.
func (l *DefaultLogger) Warn(ctx context.Context, msg string, fields map[string]any) {
	l.log("WARN", msg, fields)
}

// Debug logs a debug message with optional fields.
func (l *DefaultLogger) Debug(ctx context.Context, msg string, fields map[string]any) {
	l.log("DEBUG", msg, fields)
}

// Close closes the logger. Since the standard log package doesn't require cleanup,
// this method is a no-op.
func (l *DefaultLogger) Close() error {
	return nil
}

// log is a helper method that formats and writes log messages.
func (l *DefaultLogger) log(level, msg string, fields map[string]any) {
	if len(fields) > 0 {
		l.logger.Printf("%s: %s %s", level, msg, FormatFields(fields))
	} else {
		l.logger.Printf("%s: %s", level, msg)
	}
}

// FormatFields converts fields map to a readable key=value string format.
// Exported for testing purposes.
func FormatFields(fields map[string]any) string {
	if len(fields) == 0 {
		return ""
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build key=value pairs
	pairs := make([]string, 0, len(fields))
	for _, k := range keys {
		v := fields[k]
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}

	return strings.Join(pairs, " ")
}

// AddBaseRequestFields adds standard request metadata to log fields
// This includes: client_ip, method, path, and trace_id (if available)
func AddBaseRequestFields(fields map[string]any, c *gin.Context) {
	fields["client_ip"] = c.ClientIP()
	fields["method"] = c.Request.Method
	fields["path"] = c.Request.URL.Path

	// Add trace ID if available
	if traceID, exists := c.Get("trace_id"); exists {
		if traceIDStr, ok := traceID.(string); ok && traceIDStr != "" {
			fields["trace_id"] = traceIDStr
		}
	}
}
