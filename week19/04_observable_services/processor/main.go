package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
// Week 19 — Mini-Project: Observable Processor Service
// ========================================
//
// This is the background processing service. It receives work from
// the API service and processes it, with full observability:
//
//   1. TRACING:  Every processing step is traced
//   2. METRICS:  Processing rate, duration, errors tracked
//   3. LOGGING:  Structured JSON logs with context
//
// To run:
//   PORT=8081 go run main.go

// ========================================
// Prometheus Metrics
// ========================================

var (
	tasksProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "processor",
			Name:      "tasks_processed_total",
			Help:      "Total tasks processed.",
		},
		[]string{"status", "priority"},
	)

	processingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "processor",
			Name:      "task_duration_seconds",
			Help:      "Task processing duration.",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5},
		},
		[]string{"priority"},
	)

	taskQueueDepth = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "processor",
			Name:      "queue_depth",
			Help:      "Current queue depth.",
		},
	)
)

// ========================================
// Logger and Tracer
// ========================================

var (
	logger *slog.Logger
	tracer trace.Tracer
)

func initLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger = slog.New(handler).With(
		slog.String("service", "processor"),
		slog.String("version", "1.0.0"),
	)
}

func initTracer() func() {
	exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		logger.Error("Failed to create exporter", slog.String("error", err.Error()))
		os.Exit(1)
	}

	res, _ := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("processor-service"),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	tracer = otel.Tracer("processor-service")

	return func() { tp.Shutdown(context.Background()) }
}

// ========================================
// Request/Response types
// ========================================

type ProcessRequest struct {
	RequestID string `json:"request_id"`
	Data      string `json:"data"`
	Priority  string `json:"priority"`
}

type ProcessResponse struct {
	RequestID   string `json:"request_id"`
	Status      string `json:"status"`
	Result      string `json:"result"`
	WorkerID    string `json:"worker_id"`
	ProcessedAt string `json:"processed_at"`
	Duration    string `json:"duration"`
}

func main() {
	fmt.Println("=== Week 19 Mini-Project: Observable Processor Service ===")
	fmt.Println()

	initLogger()
	shutdownTracer := initTracer()
	defer shutdownTracer()

	port := getEnv("PORT", "8081")
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("processor-%s", hostname)

	logger.Info("Starting processor service",
		slog.String("port", port),
		slog.String("worker_id", workerID),
	)

	mux := http.NewServeMux()

	// ========================================
	// Process endpoint — the main work handler
	// ========================================
	mux.HandleFunc("POST /process", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// ---- TRACING ----
		ctx, span := tracer.Start(r.Context(), "ProcessTask",
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Parse request
		var req ProcessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "invalid request")
			logger.WarnContext(ctx, "Invalid request", slog.String("error", err.Error()))
			http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
			return
		}

		span.SetAttributes(
			attribute.String("request.id", req.RequestID),
			attribute.String("request.priority", req.Priority),
		)

		// ---- LOGGING ----
		logger.InfoContext(ctx, "Processing task",
			slog.String("request_id", req.RequestID),
			slog.String("priority", req.Priority),
			slog.Int("data_length", len(req.Data)),
		)

		// ---- METRICS: Track queue depth ----
		taskQueueDepth.Inc()
		defer taskQueueDepth.Dec()

		// ========================================
		// Step 1: Validate
		// ========================================
		_, validateSpan := tracer.Start(ctx, "Validate")
		time.Sleep(5 * time.Millisecond)
		validateSpan.End()

		// ========================================
		// Step 2: Transform
		// ========================================
		ctx2, transformSpan := tracer.Start(ctx, "Transform",
			trace.WithAttributes(
				attribute.Int("input.length", len(req.Data)),
			),
		)

		// Simulate processing with variable duration based on priority
		var processingTime time.Duration
		switch req.Priority {
		case "high":
			processingTime = time.Duration(20+rand.Intn(30)) * time.Millisecond
		case "low":
			processingTime = time.Duration(100+rand.Intn(200)) * time.Millisecond
		default:
			processingTime = time.Duration(50+rand.Intn(100)) * time.Millisecond
		}
		time.Sleep(processingTime)

		result := fmt.Sprintf("PROCESSED[%s]: %s",
			strings.ToUpper(req.Priority), strings.ToUpper(req.Data))

		transformSpan.SetAttributes(
			attribute.Int("output.length", len(result)),
			attribute.String("processing.time", processingTime.String()),
		)
		transformSpan.End()
		_ = ctx2

		// ========================================
		// Step 3: Store result
		// ========================================
		_, storeSpan := tracer.Start(ctx, "StoreResult",
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.system", "redis"),
				attribute.String("db.operation", "SET"),
			),
		)
		time.Sleep(10 * time.Millisecond)
		storeSpan.End()

		// ---- METRICS ----
		duration := time.Since(start)
		priority := req.Priority
		if priority == "" {
			priority = "normal"
		}
		tasksProcessed.WithLabelValues("success", priority).Inc()
		processingDuration.WithLabelValues(priority).Observe(duration.Seconds())

		// ---- LOGGING ----
		logger.InfoContext(ctx, "Task processed successfully",
			slog.String("request_id", req.RequestID),
			slog.Duration("duration", duration),
			slog.String("worker_id", workerID),
		)

		span.SetStatus(codes.Ok, "processed")

		writeJSON(w, http.StatusOK, ProcessResponse{
			RequestID:   req.RequestID,
			Status:      "success",
			Result:      result,
			WorkerID:    workerID,
			ProcessedAt: time.Now().Format(time.RFC3339),
			Duration:    duration.String(),
		})
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":    "healthy",
			"service":   "processor",
			"worker_id": workerID,
		})
	})

	// Prometheus metrics
	mux.Handle("GET /metrics", promhttp.Handler())

	// ========================================
	// Server
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

	fmt.Printf("Processor Service listening on :%s\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /process — Process a task")
	fmt.Println("  GET  /health  — Health check")
	fmt.Println("  GET  /metrics — Prometheus metrics")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down processor service")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	logger.Info("Processor service stopped")
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
