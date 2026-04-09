package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// ========================================
// Week 19 — Lesson 2: Distributed Tracing with OpenTelemetry
// ========================================
//
// Distributed tracing lets you follow a request as it flows through
// multiple services. It's essential for debugging microservices.
//
// Key concepts:
//   - Trace:  The entire journey of a request through the system
//   - Span:   A single operation within a trace (e.g., "database query")
//   - Context: Carries trace information between functions and services
//
// A trace is a tree of spans:
//   [HTTP Request]                          ← Root span
//     └── [Validate Input]                  ← Child span
//     └── [Query Database]                  ← Child span
//          └── [Connection Pool]            ← Grandchild span
//     └── [Call Payment Service]            ← Child span (crosses service boundary!)
//          └── [Process Payment]            ← Span in another service
//
// OpenTelemetry (OTel) is the standard for observability.
// It provides APIs for tracing, metrics, and logging.
//
// To run:
//   go run main.go
//
// This example exports traces to stdout for learning.
// In production, you'd export to Jaeger, Zipkin, or a cloud provider.

func main() {
	fmt.Println("=== Week 19, Lesson 2: Distributed Tracing ===")
	fmt.Println()

	// ========================================
	// Step 1: Initialize the tracer
	// ========================================
	fmt.Println("--- Setting up OpenTelemetry ---")
	fmt.Println()

	// Create a stdout exporter (prints traces to console).
	// In production, use:
	//   - OTLP exporter (to Jaeger, Grafana Tempo, etc.)
	//   - Zipkin exporter
	//   - Cloud provider exporter (AWS X-Ray, GCP Cloud Trace)
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithWriter(os.Stdout),
	)
	if err != nil {
		log.Fatalf("Failed to create exporter: %v", err)
	}

	// Resource describes the service producing traces.
	// This metadata is attached to every span.
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("order-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
			attribute.String("environment", "development"),
			attribute.String("team", "backend"),
		),
	)
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}

	// TracerProvider manages tracers and exports spans.
	// It's configured with:
	//   - Exporters (where to send spans)
	//   - Samplers (which requests to trace — all, 10%, etc.)
	//   - Resource (service metadata)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter), // Batch spans for efficiency
		sdktrace.WithResource(res),
		// Sampler: trace 100% of requests (for learning).
		// In production: sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1))
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Register as the global tracer provider
	otel.SetTracerProvider(tp)

	// Ensure all spans are flushed when the program exits
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer: %v", err)
		}
	}()

	fmt.Println("Tracer initialized with stdout exporter")
	fmt.Println()

	// ========================================
	// Step 2: Create a tracer
	// ========================================
	// A tracer is used to create spans. Name it after the package
	// or component that's being traced.
	tracer := otel.Tracer("order-service",
		trace.WithInstrumentationVersion("1.0.0"),
	)

	// ========================================
	// Step 3: Create spans
	// ========================================
	fmt.Println("--- Creating Spans ---")
	fmt.Println()
	fmt.Println("Simulating an order processing request...")
	fmt.Println("(Trace output will appear below)")
	fmt.Println()

	// Start the root span — this represents the incoming request
	ctx, rootSpan := tracer.Start(context.Background(), "ProcessOrder",
		// SpanKind indicates the type of operation:
		//   Server:   incoming request (HTTP handler, gRPC method)
		//   Client:   outgoing request (HTTP call, DB query)
		//   Internal: internal operation
		//   Producer: sending a message
		//   Consumer: receiving a message
		trace.WithSpanKind(trace.SpanKindServer),
		// Attributes are key-value pairs attached to the span
		trace.WithAttributes(
			attribute.String("order.id", "ORD-001"),
			attribute.String("customer.id", "CUST-042"),
			attribute.String("http.method", "POST"),
			attribute.String("http.url", "/api/orders"),
		),
	)

	// Process the order (creates child spans)
	err = processOrder(ctx, tracer, "ORD-001")

	if err != nil {
		// Record the error on the span
		rootSpan.RecordError(err)
		rootSpan.SetStatus(codes.Error, err.Error())
	} else {
		rootSpan.SetStatus(codes.Ok, "order processed successfully")
	}

	// End the span — ALWAYS end your spans!
	// Defer is the standard pattern.
	rootSpan.End()

	// Flush to ensure all spans are exported
	tp.ForceFlush(context.Background())

	fmt.Println()
	fmt.Println("--- Tracing Concepts ---")
	fmt.Println()
	fmt.Println("1. TRACE: the full journey (one trace ID)")
	fmt.Println("   All spans in a request share the same trace ID.")
	fmt.Println()
	fmt.Println("2. SPAN: a single operation (with start/end time)")
	fmt.Println("   Spans form a tree: parent -> child -> grandchild")
	fmt.Println()
	fmt.Println("3. CONTEXT: carries trace info between functions")
	fmt.Println("   Always pass ctx — it connects parent and child spans.")
	fmt.Println("   When calling another service, inject trace context into headers.")
	fmt.Println()
	fmt.Println("4. ATTRIBUTES: add data to spans")
	fmt.Println("   span.SetAttributes(attribute.String('key', 'value'))")
	fmt.Println()
	fmt.Println("5. EVENTS: log notable moments within a span")
	fmt.Println("   span.AddEvent('cache miss', ...)")
	fmt.Println()
	fmt.Println("6. STATUS: record success or failure")
	fmt.Println("   span.SetStatus(codes.Error, 'something failed')")
	fmt.Println()
	fmt.Println("7. CONTEXT PROPAGATION across services:")
	fmt.Println("   - HTTP: inject/extract trace headers (traceparent)")
	fmt.Println("   - gRPC: use otelgrpc interceptors")
	fmt.Println("   - NATS: inject/extract in message headers")
}

// processOrder simulates order processing with child spans.
func processOrder(ctx context.Context, tracer trace.Tracer, orderID string) error {
	// ========================================
	// Step: Validate input
	// ========================================
	ctx, validateSpan := tracer.Start(ctx, "ValidateInput",
		trace.WithAttributes(
			attribute.String("order.id", orderID),
		),
	)
	time.Sleep(5 * time.Millisecond)
	validateSpan.AddEvent("validation passed",
		trace.WithAttributes(attribute.Bool("valid", true)),
	)
	validateSpan.End()

	// ========================================
	// Step: Check inventory
	// ========================================
	if err := checkInventory(ctx, tracer, orderID); err != nil {
		return err
	}

	// ========================================
	// Step: Process payment (simulates calling another service)
	// ========================================
	if err := processPayment(ctx, tracer, orderID, 149.99); err != nil {
		return err
	}

	// ========================================
	// Step: Send confirmation
	// ========================================
	ctx, confirmSpan := tracer.Start(ctx, "SendConfirmation",
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(
			attribute.String("order.id", orderID),
			attribute.String("notification.type", "email"),
		),
	)
	time.Sleep(10 * time.Millisecond)
	confirmSpan.AddEvent("email sent",
		trace.WithAttributes(attribute.String("recipient", "customer@example.com")),
	)
	confirmSpan.End()

	return nil
}

// checkInventory simulates a database query with spans.
func checkInventory(ctx context.Context, tracer trace.Tracer, orderID string) error {
	ctx, span := tracer.Start(ctx, "CheckInventory",
		trace.WithSpanKind(trace.SpanKindClient), // We're calling a database
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.statement", "SELECT stock FROM products WHERE order_id = ?"),
		),
	)
	defer span.End()

	// Simulate database query
	time.Sleep(15 * time.Millisecond)

	// Add an event for the query result
	span.AddEvent("inventory check complete",
		trace.WithAttributes(
			attribute.Int("items.available", 42),
			attribute.Bool("in_stock", true),
		),
	)

	return nil
}

// processPayment simulates calling an external payment service.
func processPayment(ctx context.Context, tracer trace.Tracer, orderID string, amount float64) error {
	// This span represents an outgoing HTTP call to another service.
	// SpanKindClient tells the tracing system this is an outbound call.
	ctx, span := tracer.Start(ctx, "ProcessPayment",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("payment.order_id", orderID),
			attribute.Float64("payment.amount", amount),
			attribute.String("payment.currency", "USD"),
			attribute.String("http.method", "POST"),
			attribute.String("http.url", "https://payments.example.com/charge"),
		),
	)
	defer span.End()

	// In real code, you'd inject the trace context into HTTP headers:
	//   otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	//
	// The receiving service extracts it:
	//   ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))
	//
	// This links spans across service boundaries!

	// Simulate payment processing
	time.Sleep(30 * time.Millisecond)

	// Simulate occasional payment failures
	if rand.Float64() < 0.1 {
		err := fmt.Errorf("payment declined for order %s", orderID)
		span.RecordError(err)
		span.SetStatus(codes.Error, "payment declined")
		return err
	}

	span.SetAttributes(
		attribute.String("payment.id", fmt.Sprintf("PAY-%s", orderID[4:])),
		attribute.String("payment.status", "success"),
	)
	span.SetStatus(codes.Ok, "payment successful")

	return nil
}
