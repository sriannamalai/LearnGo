package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

// ========================================
// Week 18 — Lesson 1: Dockerizing a Go Service
// ========================================
//
// This is a simple HTTP service designed to be containerized.
// See the Dockerfile in this directory for the container build.
//
// Key points about Go and Docker:
//   - Go compiles to a single static binary (no runtime needed!)
//   - This means we can use scratch or alpine as the base image
//   - The resulting container can be as small as ~10 MB
//
// To build and run without Docker:
//   go run main.go
//
// To build and run with Docker:
//   docker build -t go-lesson-service .
//   docker run -p 8080:8080 go-lesson-service
//
// To test:
//   curl http://localhost:8080/
//   curl http://localhost:8080/info

// InfoResponse contains information about the running service.
type InfoResponse struct {
	Service     string `json:"service"`
	Version     string `json:"version"`
	GoVersion   string `json:"go_version"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	Hostname    string `json:"hostname"`
	Environment string `json:"environment"`
	Port        string `json:"port"`
	Uptime      string `json:"uptime"`
}

var startTime = time.Now()

func main() {
	fmt.Println("=== Week 18, Lesson 1: Dockerizing a Go Service ===")
	fmt.Println()

	// ========================================
	// Configuration from environment variables
	// ========================================
	// In containers, configuration comes from environment variables.
	// This is one of the 12-factor app principles.
	port := getEnv("PORT", "8080")
	env := getEnv("ENVIRONMENT", "development")

	fmt.Printf("Environment: %s\n", env)
	fmt.Printf("Port: %s\n", port)
	fmt.Println()

	// ========================================
	// HTTP handlers
	// ========================================

	// Root handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "Hello from a containerized Go service!\n")
		fmt.Fprintf(w, "Environment: %s\n", env)
		fmt.Fprintf(w, "Try /info for more details.\n")
	})

	// Info endpoint — shows runtime information
	http.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()

		info := InfoResponse{
			Service:     "go-lesson-service",
			Version:     "1.0.0",
			GoVersion:   runtime.Version(),
			OS:          runtime.GOOS,
			Arch:        runtime.GOARCH,
			Hostname:    hostname,
			Environment: env,
			Port:        port,
			Uptime:      time.Since(startTime).Round(time.Second).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	})

	// Health check endpoint (containers need this!)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "healthy",
			"uptime": time.Since(startTime).Round(time.Second).String(),
		})
	})

	// ========================================
	// Start server
	// ========================================
	addr := ":" + port
	fmt.Printf("Server starting on %s\n", addr)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  GET /       — Welcome message\n")
	fmt.Printf("  GET /info   — Service information\n")
	fmt.Printf("  GET /health — Health check\n")
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// getEnv returns the environment variable value or a default.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
