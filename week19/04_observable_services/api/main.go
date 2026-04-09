package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
// Week 19 — Mini-Project: Observable API Service
// ========================================
//
// This is the API service in our observable microservices stack.
// It demonstrates the "three pillars of observability":
//
//   1. TRACING:  Follow requests across services (OpenTelemetry)
//   2. METRICS:  Quantitative measurements (Prometheus)
//   3. LOGGING:  Structured event records (log/slog)
//
// Architecture:
//   Client ──> API Service (port 8080) ──> Processor Service (port 8081)
//
// To run:
//   go run main.go
//
// Endpoints:
//   GET  /              — Service info
//   POST /api/process   — Submit work (calls processor)
//   GET  /health        — Health check
//   GET  /metrics       — Prometheus metrics

// ========================================
// Prometheus Metrics
// ========================================

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "api",
			Name:      "requests_total",
			Help:      "Total HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "api",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	activeRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "api",
			Name:      "active_requests",
			Help:      "Currently active requests.",
		},
	)

	processorCallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "api",
			Name:      "processor_calls_total",
			Help:      "Total calls to processor service.",
		},
		[]string{"status"},
	)
)

// ========================================
// Structured Logger (slog)
// ========================================
//
// log/slog (added in Go 1.21) is the standard library's structured
// logging package. It outputs JSON by default, making logs easy to
// parse and search in log aggregation systems.

var logger *slog.Logger

func initLogger() {
	// JSON handler for production (machine-readable).
	// Use slog.NewTextHandler for human-readable development logs.
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// AddSource: true, // Uncomment to include file:line in logs
	})

	logger = slog.New(handler).With(
		slog.String("service", "api"),
		slog.String("version", "1.0.0"),
	)
}

// ========================================
// Tracer Setup
// ========================================

var tracer trace.Tracer

func initTracer() func() {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		logger.Error("Failed to create trace exporter", slog.String("error", err.Error()))
		os.Exit(1)
	}

	res, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("api-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	tracer = otel.Tracer("api-service")

	return func() {
		tp.Shutdown(context.Background())
	}
}

// ========================================
// Request/Response types
// ========================================

type ProcessRequest struct {
	Data     string `json:"data"`
	Priority string `json:"priority"`
}

type ProcessResponse struct {
	RequestID   string `json:"request_id"`
	Status      string `json:"status"`
	Result      string `json:"result,omitempty"`
	ProcessedBy string `json:"processed_by,omitempty"`
	Duration    string `json:"duration"`
	Error       string `json:"error,omitempty"`
}

func main() {
	fmt.Println("=== Week 19 Mini-Project: Observable API Service ===")
	fmt.Println()

	// Initialize the three pillars
	initLogger()
	shutdownTracer := initTracer()
	defer shutdownTracer()

	port := getEnv("PORT", "8080")
	processorURL := getEnv("PROCESSOR_URL", "http://localhost:8081")

	logger.Info("Starting API service",
		slog.String("port", port),
		slog.String("processor_url", processorURL),
	)

	mux := http.NewServeMux()

	// ========================================
	// Application endpoints
	// ========================================

	mux.HandleFunc("POST /api/process", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())

		// ---- TRACING: Start a span ----
		ctx, span := tracer.Start(r.Context(), "ProcessRequest",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("request.id", requestID),
				attribute.String("http.method", r.Method),
				attribute.String("http.path", r.URL.Path),
			),
		)
		defer span.End()

		// ---- METRICS: Track active requests ----
		activeRequests.Inc()
		defer activeRequests.Dec()

		// ---- LOGGING: Log the request ----
		logger.InfoContext(ctx, "Received process request",
			slog.String("request_id", requestID),
			slog.String("remote_addr", r.RemoteAddr),
		)

		// Parse request
		var req ProcessRequest
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &req); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid request")
			logger.WarnContext(ctx, "Invalid request body",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)
			writeJSON(w, http.StatusBadRequest, ProcessResponse{
				RequestID: requestID,
				Status:    "error",
				Error:     "invalid JSON request",
				Duration:  time.Since(start).String(),
			})
			requestsTotal.WithLabelValues("POST", "/api/process", "400").Inc()
			return
		}

		span.SetAttributes(
			attribute.String("request.priority", req.Priority),
			attribute.Int("request.data_length", len(req.Data)),
		)

		// ---- Call processor service ----
		result, err := callProcessor(ctx, processorURL, requestID, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "processor call failed")
			processorCallsTotal.WithLabelValues("error").Inc()

			logger.ErrorContext(ctx, "Processor call failed",
				slog.String("request_id", requestID),
				slog.String("error", err.Error()),
			)

			writeJSON(w, http.StatusServiceUnavailable, ProcessResponse{
				RequestID: requestID,
				Status:    "error",
				Error:     "processor unavailable",
				Duration:  time.Since(start).String(),
			})
			requestsTotal.WithLabelValues("POST", "/api/process", "503").Inc()
			return
		}

		processorCallsTotal.WithLabelValues("success").Inc()
		span.SetStatus(codes.Ok, "processed successfully")

		duration := time.Since(start)
		requestDuration.WithLabelValues("POST", "/api/process").Observe(duration.Seconds())
		requestsTotal.WithLabelValues("POST", "/api/process", "200").Inc()

		logger.InfoContext(ctx, "Request processed successfully",
			slog.String("request_id", requestID),
			slog.Duration("duration", duration),
		)

		writeJSON(w, http.StatusOK, ProcessResponse{
			RequestID:   requestID,
			Status:      "success",
			Result:      result,
			ProcessedBy: "processor-service",
			Duration:    duration.String(),
		})
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "api",
		})
	})

	// Prometheus metrics
	mux.Handle("GET /metrics", promhttp.Handler())

	// Root
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"service":  "observable-api",
			"version":  "1.0.0",
			"docs":     "POST /api/process, GET /health, GET /metrics",
		})
	})

	// ========================================
	// Start server with graceful shutdown
	// ========================================
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("Server listening", slog.String("addr", ":"+port))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	fmt.Printf("API Service listening on :%s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /api/process — Submit work (traced, metered, logged)")
	fmt.Println("  GET  /health      — Health check")
	fmt.Println("  GET  /metrics     — Prometheus metrics")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API service")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	logger.Info("API service stopped")
}

// callProcessor makes an HTTP call to the processor service.
func callProcessor(ctx context.Context, baseURL, requestID string, req ProcessRequest) (string, error) {
	_, span := tracer.Start(ctx, "CallProcessor",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("peer.service", "processor"),
			attribute.String("http.url", baseURL+"/process"),
		),
	)
	defer span.End()

	// In a real implementation, we'd:
	// 1. Marshal the request
	// 2. Inject trace context into HTTP headers
	// 3. Make the HTTP call
	// 4. Parse the response

	// Simulate the call
	time.Sleep(time.Duration(50+rand.Intn(200)) * time.Millisecond)

	// Simulate occasional failures
	if rand.Float64() < 0.1 {
		return "", fmt.Errorf("processor returned 500")
	}

	span.SetAttributes(attribute.String("processor.result", "success"))
	return fmt.Sprintf("Processed: %s (priority: %s)", req.Data, req.Priority), nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}
