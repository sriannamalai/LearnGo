package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ========================================
// Week 18 — Lesson 3: Health Checks & Graceful Shutdown
// ========================================
//
// Production services need:
//   1. Health checks:     So orchestrators know if the service is alive
//   2. Readiness checks:  So load balancers know when to send traffic
//   3. Graceful shutdown: So in-flight requests complete before stopping
//   4. Env configuration: So the same binary works in any environment
//
// Health check endpoints are used by:
//   - Docker (HEALTHCHECK in Dockerfile)
//   - Kubernetes (livenessProbe, readinessProbe)
//   - Load balancers (AWS ALB, nginx)
//   - Monitoring systems (Prometheus, Datadog)
//
// To run:
//   go run main.go
//   # Then press Ctrl+C to see graceful shutdown in action
//
// Test endpoints:
//   curl http://localhost:8080/health
//   curl http://localhost:8080/ready
//   curl http://localhost:8080/info

// Config holds all service configuration, loaded from environment.
type Config struct {
	Port              string
	Environment       string
	ShutdownTimeout   time.Duration
	ReadHeaderTimeout time.Duration
	DatabaseURL       string
	LogLevel          string
}

// loadConfig reads configuration from environment variables.
// This follows the 12-Factor App methodology.
func loadConfig() Config {
	return Config{
		Port:              getEnv("PORT", "8080"),
		Environment:       getEnv("ENVIRONMENT", "development"),
		ShutdownTimeout:   getDurationEnv("SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadHeaderTimeout: getDurationEnv("READ_HEADER_TIMEOUT", 5*time.Second),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://localhost:5432/mydb"),
		LogLevel:          getEnv("LOG_LEVEL", "info"),
	}
}

// ========================================
// Health and readiness state
// ========================================

// ServiceHealth tracks the health state of the service.
type ServiceHealth struct {
	mu            sync.RWMutex
	ready         bool
	healthy       bool
	startTime     time.Time
	lastCheck     time.Time
	dependencies  map[string]DependencyStatus
}

// DependencyStatus tracks the health of an external dependency.
type DependencyStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "up" or "down"
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse is returned by the /health endpoint.
type HealthResponse struct {
	Status       string                      `json:"status"` // "healthy" or "unhealthy"
	Uptime       string                      `json:"uptime"`
	Timestamp    string                      `json:"timestamp"`
	Dependencies map[string]DependencyStatus `json:"dependencies,omitempty"`
}

// ReadyResponse is returned by the /ready endpoint.
type ReadyResponse struct {
	Ready   bool   `json:"ready"`
	Message string `json:"message"`
}

func newServiceHealth() *ServiceHealth {
	return &ServiceHealth{
		healthy:      true,
		ready:        false, // Not ready until initialization is complete
		startTime:    time.Now(),
		dependencies: make(map[string]DependencyStatus),
	}
}

func (h *ServiceHealth) setReady(ready bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.ready = ready
}

func (h *ServiceHealth) isReady() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.ready
}

func (h *ServiceHealth) isHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.healthy
}

func (h *ServiceHealth) setDependency(name, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dependencies[name] = DependencyStatus{
		Name:   name,
		Status: status,
	}
	h.lastCheck = time.Now()
}

func main() {
	fmt.Println("=== Week 18, Lesson 3: Health Checks & Graceful Shutdown ===")
	fmt.Println()

	// ========================================
	// Load configuration from environment
	// ========================================
	cfg := loadConfig()

	fmt.Println("--- Configuration (from environment) ---")
	fmt.Printf("  PORT:             %s\n", cfg.Port)
	fmt.Printf("  ENVIRONMENT:      %s\n", cfg.Environment)
	fmt.Printf("  SHUTDOWN_TIMEOUT: %s\n", cfg.ShutdownTimeout)
	fmt.Printf("  DATABASE_URL:     %s\n", cfg.DatabaseURL)
	fmt.Printf("  LOG_LEVEL:        %s\n", cfg.LogLevel)
	fmt.Println()

	// ========================================
	// Initialize health tracking
	// ========================================
	health := newServiceHealth()

	// ========================================
	// Set up HTTP handlers
	// ========================================
	mux := http.NewServeMux()

	// ----------------------------------------
	// /health — Liveness probe
	// ----------------------------------------
	// Purpose: "Is this process alive and running?"
	// Used by: Docker HEALTHCHECK, Kubernetes livenessProbe
	// If unhealthy: container/pod is RESTARTED
	//
	// Should check: basic process health
	// Should NOT check: external dependencies (that's /ready)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		resp := HealthResponse{
			Uptime:    time.Since(health.startTime).Round(time.Second).String(),
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if health.isHealthy() {
			resp.Status = "healthy"
			w.WriteHeader(http.StatusOK)
		} else {
			resp.Status = "unhealthy"
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// ----------------------------------------
	// /ready — Readiness probe
	// ----------------------------------------
	// Purpose: "Is this service ready to handle traffic?"
	// Used by: Kubernetes readinessProbe, load balancers
	// If not ready: traffic is NOT sent to this instance
	//   (but the container is NOT restarted)
	//
	// Should check: database connection, cache connection,
	//   required services, initialization complete
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if health.isReady() {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ReadyResponse{
				Ready:   true,
				Message: "service is ready",
			})
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ReadyResponse{
				Ready:   false,
				Message: "service is initializing",
			})
		}
	})

	// ----------------------------------------
	// /info — Detailed service information
	// ----------------------------------------
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"service":     "health-checks-demo",
			"version":     "1.0.0",
			"environment": cfg.Environment,
			"hostname":    hostname,
			"uptime":      time.Since(health.startTime).Round(time.Second).String(),
			"healthy":     health.isHealthy(),
			"ready":       health.isReady(),
		})
	})

	// ----------------------------------------
	// / — Application endpoint
	// ----------------------------------------
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Simulate some work
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Hello from a production-ready Go service!",
		})
	})

	// ========================================
	// Create the HTTP server
	// ========================================
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// ========================================
	// Simulate initialization (warming up)
	// ========================================
	// Start the server first, then mark as ready.
	// This way, health check passes but readiness doesn't
	// until we're fully initialized.
	go func() {
		fmt.Println("--- Simulating startup initialization ---")
		fmt.Println("  Connecting to database...")
		time.Sleep(500 * time.Millisecond)
		health.setDependency("database", "up")
		fmt.Println("  Database connected")

		fmt.Println("  Loading configuration...")
		time.Sleep(300 * time.Millisecond)
		fmt.Println("  Configuration loaded")

		fmt.Println("  Warming caches...")
		time.Sleep(200 * time.Millisecond)
		health.setDependency("cache", "up")
		fmt.Println("  Caches warmed")

		// Now mark as ready — load balancer can send traffic
		health.setReady(true)
		fmt.Println()
		fmt.Println("  Service is READY for traffic!")
		fmt.Println()
	}()

	// ========================================
	// Start server in a goroutine
	// ========================================
	go func() {
		fmt.Printf("Server starting on :%s\n", cfg.Port)
		fmt.Println("Endpoints:")
		fmt.Println("  GET /       — Application endpoint")
		fmt.Println("  GET /health — Liveness probe")
		fmt.Println("  GET /ready  — Readiness probe")
		fmt.Println("  GET /info   — Service information")
		fmt.Println()
		fmt.Println("Press Ctrl+C to see graceful shutdown")
		fmt.Println()

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// ========================================
	// Graceful Shutdown
	// ========================================
	// Wait for interrupt signal (Ctrl+C or SIGTERM from orchestrator)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	fmt.Printf("\n--- Graceful Shutdown (received %s) ---\n", sig)
	fmt.Println()

	// Step 1: Mark as NOT ready (stop receiving new traffic)
	fmt.Println("Step 1: Marking service as NOT ready...")
	health.setReady(false)
	fmt.Println("  Load balancer will stop sending new requests")
	fmt.Println()

	// Step 2: Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Step 3: Gracefully shut down the server
	// This waits for in-flight requests to complete, up to the timeout.
	fmt.Println("Step 2: Waiting for in-flight requests to complete...")
	fmt.Printf("  Timeout: %s\n", cfg.ShutdownTimeout)

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("  Forced shutdown: %v", err)
	} else {
		fmt.Println("  All requests completed!")
	}
	fmt.Println()

	// Step 4: Clean up resources
	fmt.Println("Step 3: Cleaning up resources...")
	fmt.Println("  Closing database connections...")
	time.Sleep(100 * time.Millisecond) // Simulate cleanup
	fmt.Println("  Flushing metrics...")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("  Done!")
	fmt.Println()

	fmt.Println("Server shut down cleanly. Goodbye!")
}

// ========================================
// Helper functions
// ========================================

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		d, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		return d
	}
	return defaultValue
}
