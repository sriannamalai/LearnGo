package observability

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// ========================================
// OpenTelemetry Tracing Setup
// ========================================
// Distributed tracing allows you to follow a request as it flows
// through your system — from the HTTP handler, through the service
// layer, into the database, and back. This demonstrates:
//   - Week 17-18: OpenTelemetry setup, span creation, context propagation
//   - Week 6-7: Context as the carrier for trace data
//
// How tracing works:
//   1. A trace represents an entire request lifecycle
//   2. A trace contains multiple spans (one per operation)
//   3. Spans have parent-child relationships forming a tree
//   4. Context propagation carries the trace ID across boundaries
//
// Example trace for "Create Task":
//   [HTTP Handler] POST /api/v1/tasks (50ms)
//     [Service] CreateTask (45ms)
//       [Postgres] INSERT task (10ms)
//       [ArangoDB] Create node (8ms)
//
// This setup sends traces to an OTLP-compatible collector (like
// Jaeger, Grafana Tempo, or Honeycomb).

// InitTracer initializes the OpenTelemetry trace provider.
// It returns a shutdown function that must be called when the
// application exits to flush any remaining spans.
func InitTracer(ctx context.Context, serviceName, endpoint string) (func(context.Context) error, error) {
	// ========================================
	// 1. Create the Exporter
	// ========================================
	// The exporter sends span data to a collection backend.
	// OTLP (OpenTelemetry Protocol) is the standard format
	// supported by most observability platforms.
	//
	// We use otlptracehttp for HTTP-based export. For gRPC export:
	//   otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint))

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // Use TLS in production
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTLP exporter: %w", err)
	}

	// ========================================
	// 2. Create the Resource
	// ========================================
	// A resource describes the entity producing telemetry.
	// It includes the service name, version, and environment.
	// All spans from this service will carry these attributes.

	res, err := resource.New(ctx,
		resource.WithFromEnv(),   // Read OTEL_RESOURCE_ATTRIBUTES env var
		resource.WithProcess(),   // Add process info (PID, runtime)
		resource.WithOS(),        // Add OS info
		resource.WithHost(),      // Add host info
	)
	if err != nil {
		slog.Warn("failed to create trace resource, using default", "error", err)
		res = resource.Default()
	}

	// ========================================
	// 3. Create the Trace Provider
	// ========================================
	// The trace provider manages span processors and exporters.
	// BatchSpanProcessor batches spans before sending them to the
	// exporter, which is more efficient than sending one at a time.

	tp := sdktrace.NewTracerProvider(
		// BatchSpanProcessor sends spans in batches for efficiency.
		// In production, you can configure batch size and timeout.
		sdktrace.WithBatcher(exporter),

		// The resource identifies this service in trace data
		sdktrace.WithResource(res),

		// Sampler controls which traces are recorded.
		// AlwaysSample is good for development; use
		// TraceIDRatioBased(0.1) for production (sample 10%).
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// ========================================
	// 4. Register as the Global Provider
	// ========================================
	// Setting the global provider allows all packages to create
	// spans using otel.Tracer("package-name").
	otel.SetTracerProvider(tp)

	slog.Info("OpenTelemetry tracer initialized",
		"service", serviceName,
		"endpoint", endpoint,
	)

	// Return a shutdown function that flushes pending spans
	return func(ctx context.Context) error {
		slog.Info("shutting down trace provider")
		return tp.Shutdown(ctx)
	}, nil
}

// ========================================
// Tracer Helper
// ========================================
// Use this to create spans in your code:
//
//   ctx, span := observability.Tracer().Start(ctx, "operation-name")
//   defer span.End()
//   // ... do work ...
//   span.SetAttributes(attribute.String("key", "value"))

// Tracer returns a named tracer for the TaskFlow application.
// Each package should use its own tracer name for clear span attribution.
//
// Usage in other packages:
//   tracer := otel.Tracer("github.com/sri/learngo/week24/internal/service")
//   ctx, span := tracer.Start(ctx, "TaskService.CreateTask")
//   defer span.End()
func Tracer() interface{} {
	return otel.Tracer("github.com/sri/learngo/week24/01_capstone")
}
