package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ========================================
// Week 18 — Mini-Project: Worker Service
// ========================================
//
// This is the background worker that processes tasks
// submitted through the API gateway.
//
// In a real system, this might:
//   - Process images
//   - Send emails
//   - Generate reports
//   - Run ML inference
//
// To run standalone:
//   PORT=8081 go run main.go
//
// To run with Docker Compose:
//   cd .. && docker-compose up --build

// Task represents a work item received from the API.
type Task struct {
	ID        string    `json:"id"`
	Data      string    `json:"data"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ProcessedTask includes the processing result.
type ProcessedTask struct {
	Task
	Result      string    `json:"result"`
	ProcessedAt time.Time `json:"processed_at"`
	Duration    string    `json:"duration"`
	WorkerID    string    `json:"worker_id"`
}

var (
	processed   = make(map[string]*ProcessedTask)
	processedMu sync.RWMutex
	workerID    string
)

func main() {
	fmt.Println("=== Week 18 Mini-Project: Worker Service ===")
	fmt.Println()

	port := getEnv("PORT", "8081")
	hostname, _ := os.Hostname()
	workerID = fmt.Sprintf("worker-%s", hostname[:8])

	fmt.Printf("Worker starting on port %s\n", port)
	fmt.Printf("Worker ID: %s\n", workerID)
	fmt.Println()

	mux := http.NewServeMux()

	// Process endpoint — receives tasks from the API gateway
	mux.HandleFunc("POST /process", func(w http.ResponseWriter, r *http.Request) {
		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		log.Printf("[Worker] Processing task %s: %s", task.ID, task.Data)

		// Simulate processing
		start := time.Now()
		time.Sleep(200 * time.Millisecond)

		result := &ProcessedTask{
			Task:        task,
			Result:      fmt.Sprintf("Processed: %s", strings.ToUpper(task.Data)),
			ProcessedAt: time.Now(),
			Duration:    time.Since(start).String(),
			WorkerID:    workerID,
		}
		result.Status = "completed"

		processedMu.Lock()
		processed[task.ID] = result
		processedMu.Unlock()

		log.Printf("[Worker] Task %s completed in %s", task.ID, result.Duration)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	})

	// List processed tasks
	mux.HandleFunc("GET /processed", func(w http.ResponseWriter, r *http.Request) {
		processedMu.RLock()
		defer processedMu.RUnlock()

		list := make([]*ProcessedTask, 0, len(processed))
		for _, t := range processed {
			list = append(list, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"processed": list,
			"count":     len(list),
			"worker_id": workerID,
		})
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "healthy",
			"service":   "worker",
			"worker_id": workerID,
		})
	})

	// ========================================
	// Server with graceful shutdown
	// ========================================
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		fmt.Printf("Worker listening on :%s\n", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nWorker shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	fmt.Println("Worker stopped")
}

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}
