package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ========================================
// Week 18 — Lesson 2: Service B (Worker/Processor)
// ========================================
//
// This is the "Worker" service in our Docker Compose setup.
// It receives processing requests from Service A and returns results.
//
// In Docker Compose:
//   Service A calls this service at http://service-b:8081
//   Docker Compose DNS handles the name resolution.
//
// To run standalone:
//   PORT=8081 go run main.go
//
// To run with Docker Compose:
//   cd .. && docker-compose up

// ProcessRequest is received from Service A.
type ProcessRequest struct {
	TaskID string `json:"task_id"`
	Data   string `json:"data"`
}

// ProcessResponse is sent back to Service A.
type ProcessResponse struct {
	TaskID      string `json:"task_id"`
	Result      string `json:"result"`
	Processor   string `json:"processor"`
	ProcessedAt string `json:"processed_at"`
	Duration    string `json:"duration"`
}

var startTime = time.Now()

func main() {
	fmt.Println("=== Week 18, Lesson 2: Service B (Worker) ===")
	fmt.Println()

	port := getEnv("PORT", "8081")
	hostname, _ := os.Hostname()

	fmt.Printf("Service B starting on port %s\n", port)
	fmt.Printf("Hostname: %s\n", hostname)
	fmt.Println()

	// ========================================
	// HTTP Handlers
	// ========================================

	// Process endpoint — does the actual work
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		start := time.Now()

		var req ProcessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("[Service B] Processing task: %s", req.TaskID)

		// ========================================
		// Simulate processing work
		// ========================================
		// In a real system, this could be:
		//   - Image processing
		//   - Data transformation
		//   - ML inference
		//   - Heavy computation
		time.Sleep(100 * time.Millisecond) // Simulate work

		result := fmt.Sprintf("Processed: %s", strings.ToUpper(req.Data))

		resp := ProcessResponse{
			TaskID:      req.TaskID,
			Result:      result,
			Processor:   hostname, // Useful to see which container handled it
			ProcessedAt: time.Now().Format(time.RFC3339),
			Duration:    time.Since(start).String(),
		}

		log.Printf("[Service B] Task %s completed in %s", req.TaskID, resp.Duration)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "service-b",
			"uptime":  time.Since(startTime).Round(time.Second).String(),
		})
	})

	addr := ":" + port
	fmt.Printf("Service B listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
