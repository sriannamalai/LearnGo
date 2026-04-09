package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ========================================
// Week 19 — Lesson 3: Prometheus Metrics
// ========================================
//
// Prometheus is the standard for metrics in the cloud-native world.
// It works on a PULL model: Prometheus server scrapes a /metrics
// endpoint on your service at regular intervals.
//
// Metric types:
//   1. Counter:   Only goes up (requests served, errors occurred)
//   2. Gauge:     Goes up and down (temperature, active connections)
//   3. Histogram: Distribution of values (request latency, response sizes)
//   4. Summary:   Similar to histogram but calculates quantiles client-side
//
// To run:
//   go run main.go
//
// Then visit:
//   http://localhost:8080/         — Application endpoints
//   http://localhost:8080/metrics  — Prometheus metrics endpoint
//
// To set up Prometheus (optional):
//   # prometheus.yml configuration:
//   scrape_configs:
//     - job_name: 'go-app'
//       scrape_interval: 15s
//       static_configs:
//         - targets: ['localhost:8080']
//
//   docker run -d -p 9090:9090 \
//     -v ./prometheus.yml:/etc/prometheus/prometheus.yml \
//     prom/prometheus

// ========================================
// Define Metrics
// ========================================
//
// Metrics are defined as package-level variables.
// promauto.NewXxx automatically registers with the default registry.

// Counter: counts total HTTP requests, labeled by method, path, and status.
// Labels let you filter and group metrics in queries.
var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		// Namespace and subsystem are optional prefixes:
		// metric name = namespace_subsystem_name
		Namespace: "myapp",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests handled.",
	},
	// Labels — each unique combination of label values creates a time series
	[]string{"method", "path", "status"},
)

// Counter: counts total errors.
var errorsTotal = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "myapp",
		Name:      "errors_total",
		Help:      "Total number of errors encountered.",
	},
)

// Gauge: tracks currently active requests.
var activeRequests = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "myapp",
		Subsystem: "http",
		Name:      "active_requests",
		Help:      "Number of HTTP requests currently being processed.",
	},
)

// Gauge: tracks a business metric (items in queue).
var queueSize = promauto.NewGauge(
	prometheus.GaugeOpts{
		Namespace: "myapp",
		Name:      "queue_size",
		Help:      "Number of items currently in the processing queue.",
	},
)

// Histogram: tracks request duration distribution.
// Buckets define the boundaries for duration buckets.
var requestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "myapp",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		// Buckets define the histogram boundaries.
		// Choose buckets based on your SLO requirements.
		// These are for a typical web service:
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
	},
	[]string{"method", "path"},
)

// Histogram: tracks response sizes.
var responseSize = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "myapp",
		Subsystem: "http",
		Name:      "response_size_bytes",
		Help:      "HTTP response size in bytes.",
		Buckets:   prometheus.ExponentialBuckets(100, 2, 10), // 100, 200, 400, ... bytes
	},
	[]string{"method", "path"},
)

func main() {
	fmt.Println("=== Week 19, Lesson 3: Prometheus Metrics ===")
	fmt.Println()

	// ========================================
	// Instrumented HTTP handlers
	// ========================================

	// Application endpoint — instrumented with metrics
	http.HandleFunc("/api/orders", instrumentHandler("GET", "/api/orders",
		func(w http.ResponseWriter, r *http.Request) {
			// Simulate varying response times
			delay := time.Duration(rand.Intn(200)+10) * time.Millisecond
			time.Sleep(delay)

			response := `{"orders": [{"id": "ORD-001"}, {"id": "ORD-002"}]}`
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		},
	))

	// Another endpoint
	http.HandleFunc("/api/users", instrumentHandler("GET", "/api/users",
		func(w http.ResponseWriter, r *http.Request) {
			delay := time.Duration(rand.Intn(100)+5) * time.Millisecond
			time.Sleep(delay)

			response := `{"users": [{"id": "USR-001", "name": "Alice"}]}`
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(response))
		},
	))

	// Endpoint that sometimes errors
	http.HandleFunc("/api/process", instrumentHandler("POST", "/api/process",
		func(w http.ResponseWriter, r *http.Request) {
			delay := time.Duration(rand.Intn(500)+50) * time.Millisecond
			time.Sleep(delay)

			// Simulate 20% error rate
			if rand.Float64() < 0.2 {
				errorsTotal.Inc()
				http.Error(w, `{"error": "processing failed"}`, http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status": "processed"}`))
		},
	))

	// ========================================
	// Prometheus metrics endpoint
	// ========================================
	// promhttp.Handler() serves the /metrics endpoint that Prometheus scrapes.
	// It includes all registered metrics in the Prometheus text format.
	http.Handle("/metrics", promhttp.Handler())

	// Info endpoint explaining the metrics
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, "Prometheus Metrics Demo")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Endpoints:")
		fmt.Fprintln(w, "  GET  /api/orders  — List orders (instrumented)")
		fmt.Fprintln(w, "  GET  /api/users   — List users (instrumented)")
		fmt.Fprintln(w, "  POST /api/process — Process something (20% error rate)")
		fmt.Fprintln(w, "  GET  /metrics     — Prometheus metrics endpoint")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Generate some traffic, then check /metrics!")
	})

	// ========================================
	// Background metrics simulation
	// ========================================
	// Simulate queue activity for the gauge metric
	go simulateQueue()

	// Generate some initial traffic for demonstration
	go generateTraffic()

	fmt.Println("Server starting on :8080")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  http://localhost:8080/          — Info")
	fmt.Println("  http://localhost:8080/api/orders — Orders (instrumented)")
	fmt.Println("  http://localhost:8080/api/users  — Users (instrumented)")
	fmt.Println("  http://localhost:8080/metrics    — Prometheus metrics")
	fmt.Println()
	fmt.Println("--- Prometheus Metric Types ---")
	fmt.Println()
	fmt.Println("1. COUNTER (myapp_http_requests_total):")
	fmt.Println("   Only goes up. Use for: request counts, error counts, bytes sent")
	fmt.Println("   Query: rate(myapp_http_requests_total[5m]) — requests per second")
	fmt.Println()
	fmt.Println("2. GAUGE (myapp_http_active_requests):")
	fmt.Println("   Goes up and down. Use for: queue size, temperature, connections")
	fmt.Println("   Query: myapp_queue_size — current queue depth")
	fmt.Println()
	fmt.Println("3. HISTOGRAM (myapp_http_request_duration_seconds):")
	fmt.Println("   Tracks distributions. Prometheus stores data in buckets.")
	fmt.Println("   Query: histogram_quantile(0.95, rate(myapp_http_request_duration_seconds_bucket[5m]))")
	fmt.Println("          — 95th percentile latency")
	fmt.Println()
	fmt.Println("--- Naming Conventions ---")
	fmt.Println("  Use snake_case: myapp_http_requests_total")
	fmt.Println("  Include unit:   _seconds, _bytes, _total")
	fmt.Println("  Counters end in _total")
	fmt.Println()
	fmt.Println("Check /metrics after generating some traffic!")
	fmt.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// ========================================
// Instrumentation middleware
// ========================================

// instrumentHandler wraps an HTTP handler with Prometheus metrics.
// This is the pattern for instrumenting all your endpoints.
func instrumentHandler(method, path string, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Track active requests (gauge)
		activeRequests.Inc()
		defer activeRequests.Dec()

		// Wrap the ResponseWriter to capture the status code
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: 200}

		// Call the actual handler
		handler(rw, r)

		// Record metrics after the handler completes
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", rw.statusCode)

		// Increment request counter
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()

		// Record duration in histogram
		requestDuration.WithLabelValues(method, path).Observe(duration)

		// Record response size
		responseSize.WithLabelValues(method, path).Observe(float64(rw.bytesWritten))
	}
}

// responseWriterWrapper captures the status code and bytes written.
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// simulateQueue simulates changing queue depth for gauge demonstration.
func simulateQueue() {
	for {
		size := float64(rand.Intn(50))
		queueSize.Set(size)
		time.Sleep(2 * time.Second)
	}
}

// generateTraffic creates some initial requests for metrics.
func generateTraffic() {
	time.Sleep(1 * time.Second) // Wait for server to start

	client := &http.Client{Timeout: 5 * time.Second}
	endpoints := []string{
		"http://localhost:8080/api/orders",
		"http://localhost:8080/api/users",
	}

	for i := 0; i < 20; i++ {
		url := endpoints[rand.Intn(len(endpoints))]
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Also generate some POST requests
	for i := 0; i < 10; i++ {
		resp, err := client.Post("http://localhost:8080/api/process",
			"application/json", nil)
		if err == nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
}
