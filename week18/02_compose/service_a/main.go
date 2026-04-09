package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ========================================
// Week 18 — Lesson 2: Service A (HTTP API)
// ========================================
//
// This is the "API" service in our Docker Compose setup.
// It receives HTTP requests from clients and communicates
// with Service B for processing.
//
// In Docker Compose, services communicate by name:
//   Service A calls http://service-b:8081/process
//   (Docker Compose DNS resolves "service-b" to the container IP)
//
// To run standalone:
//   go run main.go
//
// To run with Docker Compose:
//   cd .. && docker-compose up

// ProcessRequest is sent to Service B.
type ProcessRequest struct {
	TaskID string `json:"task_id"`
	Data   string `json:"data"`
}

// ProcessResponse is received from Service B.
type ProcessResponse struct {
	TaskID    string `json:"task_id"`
	Result    string `json:"result"`
	Processor string `json:"processor"`
}

var startTime = time.Now()

func main() {
	fmt.Println("=== Week 18, Lesson 2: Service A (HTTP API) ===")
	fmt.Println()

	port := getEnv("PORT", "8080")
	serviceBURL := getEnv("SERVICE_B_URL", "http://localhost:8081")

	fmt.Printf("Service A starting on port %s\n", port)
	fmt.Printf("Service B URL: %s\n", serviceBURL)
	fmt.Println()

	// ========================================
	// HTTP Handlers
	// ========================================

	// Root endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"service": "service-a",
			"status":  "running",
			"uptime":  time.Since(startTime).Round(time.Second).String(),
		})
	})

	// Process endpoint — delegates to Service B
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		taskID := fmt.Sprintf("task-%d", time.Now().UnixNano())
		log.Printf("[Service A] Received request, assigning task ID: %s", taskID)

		// ========================================
		// Call Service B for processing
		// ========================================
		// In Docker Compose, "service-b" resolves via Docker DNS.
		// This is how containers communicate within a compose network.
		processURL := serviceBURL + "/process"
		log.Printf("[Service A] Forwarding to Service B: %s", processURL)

		req := ProcessRequest{
			TaskID: taskID,
			Data:   string(body),
		}
		reqData, _ := json.Marshal(req)

		// Make the HTTP call to Service B
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(processURL, "application/json",
			io.NopCloser(jsonReader(reqData)))
		if err != nil {
			log.Printf("[Service A] Service B error: %v", err)
			http.Error(w, fmt.Sprintf("service B unavailable: %v", err),
				http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Forward Service B's response
		respBody, _ := io.ReadAll(resp.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)

		log.Printf("[Service A] Task %s completed", taskID)
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "service-a",
		})
	})

	addr := ":" + port
	fmt.Printf("Service A listening on %s\n", addr)
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

// jsonReader wraps a byte slice in a reader for http.Post.
func jsonReader(data []byte) io.Reader {
	return io.NopCloser(bytesReader(data))
}

type bytesReaderImpl struct {
	data []byte
	pos  int
}

func bytesReader(data []byte) *bytesReaderImpl {
	return &bytesReaderImpl{data: data}
}

func (r *bytesReaderImpl) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *bytesReaderImpl) Close() error { return nil }
