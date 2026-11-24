package router

import (
	"context"
	"errors"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/edaniel30/http-platform-go/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// recordingLogger keeps the fields of every log call.
type recordingLogger struct {
	mu      sync.Mutex
	entries []map[string]any
}

func (l *recordingLogger) record(fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.entries = append(l.entries, fields)
}

func (l *recordingLogger) Info(_ context.Context, _ string, f map[string]any)  { l.record(f) }
func (l *recordingLogger) Error(_ context.Context, _ string, f map[string]any) { l.record(f) }
func (l *recordingLogger) Warn(_ context.Context, _ string, f map[string]any)  { l.record(f) }
func (l *recordingLogger) Debug(_ context.Context, _ string, f map[string]any) { l.record(f) }
func (l *recordingLogger) Close() error                                        { return nil }

func TestTraceIDWithTelemetry(t *testing.T) {
	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(previous)
		_ = tp.Shutdown(context.Background())
	})

	log := &recordingLogger{}
	router := NewGinRouter(RouterConfig{
		Mode:          gin.TestMode,
		EnableTraceID: true,
		ServiceName:   "test-service",
		Logger:        log,
	})

	var spanTraceID, fromCtx string
	router.GET("/fail", func(c *gin.Context) {
		spanTraceID = trace.SpanContextFromContext(c.Request.Context()).TraceID().String()
		fromCtx = middleware.TraceIDFromContext(c.Request.Context())
		_ = c.Error(errors.New("boom"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/fail", nil)
	router.engine.ServeHTTP(w, req)

	assert.Len(t, spanTraceID, 32)
	assert.Equal(t, spanTraceID, fromCtx)
	assert.Equal(t, spanTraceID, w.Header().Get(middleware.TraceIDHeader))

	// The error handler's log carries the OpenTelemetry trace ID, not the UUID
	assert.NotEmpty(t, log.entries)
	for _, fields := range log.entries {
		assert.Equal(t, spanTraceID, fields["trace_id"])
	}
}
