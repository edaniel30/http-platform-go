package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// TelemetryManager manages the OpenTelemetry lifecycle.
type Manager struct {
	tp       *sdktrace.TracerProvider
	shutdown func(context.Context) error
}

// InitTelemetry initializes OpenTelemetry with OTLP exporter to Datadog Agent.
// Returns a TelemetryManager that should be shut down on application exit.
//
// Parameters:
//   - serviceName: Name of the service for tracing
//   - serviceVersion: Version of the service
//   - environment: Deployment environment (e.g., "production", "development")
//   - otlpEndpoint: OTLP endpoint (e.g., "localhost:4318" for Datadog Agent)
//   - sampleAll: If true, samples all traces; if false, samples 10% of traces
func InitTelemetry(ctx context.Context, serviceName, serviceVersion, environment, otlpEndpoint string, sampleAll bool) (*Manager, error) {
	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.DeploymentEnvironmentKey.String(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter to Datadog Agent
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(), // Agent is typically local, no TLS needed
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Determine sampler
	var sampler sdktrace.Sampler
	if sampleAll {
		sampler = sdktrace.AlwaysSample()
	} else {
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1)) // Sample 10% by default
	}

	// Create trace provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(512),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Set global trace provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Manager{
		tp:       tp,
		shutdown: tp.Shutdown,
	}, nil
}

// Shutdown gracefully shuts down the telemetry provider.
func (tm *Manager) Shutdown(ctx context.Context) error {
	if tm.shutdown != nil {
		return tm.shutdown(ctx)
	}
	return nil
}
