package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ========================================
// Prometheus Metrics
// ========================================
// Prometheus is the standard for metrics collection in Go services.
// It uses a pull model — Prometheus scrapes the /metrics endpoint
// at regular intervals. This demonstrates:
//   - Week 17-18: Metrics types, labels, and registration
//
// Metric types:
//   - Counter:   monotonically increasing (requests, errors)
//   - Gauge:     value that goes up and down (active connections)
//   - Histogram: distribution of values (request durations)
//   - Summary:   similar to histogram, calculates quantiles client-side
//
// Naming Convention: <namespace>_<subsystem>_<name>_<unit>
//   Example: taskflow_http_requests_total
//            taskflow_http_request_duration_seconds

// ========================================
// HTTP Metrics
// ========================================

var (
	// HTTPRequestsTotal counts all HTTP requests by method, path, and status.
	// This is the most fundamental metric for any web service.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "taskflow",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests by method, path, and status code.",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration measures the latency of HTTP requests.
	// The histogram buckets are chosen for typical API response times.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "taskflow",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Duration of HTTP requests in seconds.",
			// Buckets from 5ms to 10s — covering fast and slow requests
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "path"},
	)

	// HTTPRequestsInFlight tracks the number of currently active requests.
	// This gauge helps detect overload conditions.
	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "taskflow",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed.",
		},
	)
)

// ========================================
// gRPC Metrics
// ========================================

var (
	// GRPCRequestsTotal counts all gRPC calls by method and status.
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "taskflow",
			Subsystem: "grpc",
			Name:      "requests_total",
			Help:      "Total number of gRPC requests by method and status.",
		},
		[]string{"method", "status"},
	)

	// GRPCRequestDuration measures gRPC call latency.
	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "taskflow",
			Subsystem: "grpc",
			Name:      "request_duration_seconds",
			Help:      "Duration of gRPC requests in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

// ========================================
// Database Metrics
// ========================================

var (
	// DBQueryDuration measures database query latency.
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "taskflow",
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Duration of database queries in seconds.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"store", "operation"},
	)

	// DBConnectionsActive tracks the number of active DB connections.
	DBConnectionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "taskflow",
			Subsystem: "db",
			Name:      "connections_active",
			Help:      "Number of active database connections.",
		},
		[]string{"store"},
	)

	// DBErrorsTotal counts database errors by store and operation.
	DBErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "taskflow",
			Subsystem: "db",
			Name:      "errors_total",
			Help:      "Total number of database errors.",
		},
		[]string{"store", "operation"},
	)
)

// ========================================
// Business Metrics
// ========================================
// Business metrics track domain-specific events. These are
// invaluable for understanding user behavior and system health
// beyond just infrastructure metrics.

var (
	// TasksCreatedTotal counts tasks created by priority.
	TasksCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "taskflow",
			Subsystem: "tasks",
			Name:      "created_total",
			Help:      "Total number of tasks created by priority.",
		},
		[]string{"priority"},
	)

	// TasksCompletedTotal counts tasks completed.
	TasksCompletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "taskflow",
			Subsystem: "tasks",
			Name:      "completed_total",
			Help:      "Total number of tasks completed.",
		},
	)

	// UsersRegisteredTotal counts user registrations.
	UsersRegisteredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "taskflow",
			Subsystem: "users",
			Name:      "registered_total",
			Help:      "Total number of users registered.",
		},
	)
)

// RegisterMetrics is a no-op since promauto handles registration.
// This function exists as a clear initialization point in the startup
// sequence, ensuring all metrics are registered before the server
// starts handling requests.
func RegisterMetrics() {
	// promauto registers metrics automatically when they're declared.
	// This function serves as documentation that metrics are ready.
	//
	// If you need custom collectors (like runtime stats), register
	// them here:
	//   prometheus.MustRegister(collectors.NewGoCollector())
	//   prometheus.MustRegister(collectors.NewProcessCollector(...))
}

// ========================================
// Usage Examples
// ========================================
//
// In HTTP handlers:
//   observability.HTTPRequestsTotal.WithLabelValues("GET", "/api/tasks", "200").Inc()
//   timer := prometheus.NewTimer(observability.HTTPRequestDuration.WithLabelValues("GET", "/api/tasks"))
//   defer timer.ObserveDuration()
//
// In database operations:
//   timer := prometheus.NewTimer(observability.DBQueryDuration.WithLabelValues("postgres", "insert_task"))
//   defer timer.ObserveDuration()
//
// In business logic:
//   observability.TasksCreatedTotal.WithLabelValues("high").Inc()
//   observability.TasksCompletedTotal.Inc()
//
// Prometheus query examples (PromQL):
//   rate(taskflow_http_requests_total[5m])              # Request rate
//   histogram_quantile(0.99, rate(taskflow_http_request_duration_seconds_bucket[5m]))  # p99 latency
//   taskflow_tasks_created_total                         # Total tasks created
//   taskflow_db_connections_active                       # Active DB connections
