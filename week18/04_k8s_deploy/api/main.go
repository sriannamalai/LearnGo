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
// Week 18 — Mini-Project: API Gateway Service
// ========================================
//
// This is the API gateway for our Dockerized microservices stack.
// It acts as the entry point for all client requests and routes
// them to the appropriate backend services.
//
// Architecture:
//   Client ──> API Gateway (this service, port 8080)
//                  |
//                  ├──> Worker Service (internal, port 8081)
//                  └──> (future services...)
//
// To run standalone:
//   go run main.go
//
// To run with Docker Compose:
//   cd .. && docker-compose up --build

// Task represents a work item.
type Task struct {
	ID        string    `json:"id"`
	Data      string    `json:"data"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	tasks   = make(map[string]*Task)
	tasksMu sync.RWMutex
	counter int
)

func main() {
	fmt.Println("=== Week 18 Mini-Project: API Gateway ===")
	fmt.Println()

	port := getEnv("PORT", "8080")
	workerURL := getEnv("WORKER_URL", "http://localhost:8081")

	fmt.Printf("API Gateway starting on port %s\n", port)
	fmt.Printf("Worker service: %s\n", workerURL)
	fmt.Println()

	mux := http.NewServeMux()

	// ========================================
	// API Endpoints
	// ========================================

	// List all tasks
	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		tasksMu.RLock()
		defer tasksMu.RUnlock()

		taskList := make([]*Task, 0, len(tasks))
		for _, t := range tasks {
			taskList = append(taskList, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tasks": taskList,
			"count": len(taskList),
		})
	})

	// Create a task
	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}

		tasksMu.Lock()
		counter++
		id := fmt.Sprintf("task-%03d", counter)
		task := &Task{
			ID:        id,
			Data:      input.Data,
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		tasks[id] = task
		tasksMu.Unlock()

		log.Printf("[API] Created task %s", id)

		// Notify worker service (fire-and-forget)
		go notifyWorker(workerURL, task)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	})

	// Get a specific task
	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		tasksMu.RLock()
		task, exists := tasks[id]
		tasksMu.RUnlock()

		if !exists {
			http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(task)
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "api-gateway",
		})
	})

	// Readiness check
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ready": true})
	})

	// Root
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "api-gateway",
			"version": "1.0.0",
			"docs":    "POST /tasks, GET /tasks, GET /tasks/{id}",
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
		fmt.Printf("API Gateway listening on :%s\n", port)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nAPI Gateway shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	fmt.Println("API Gateway stopped")
}

// notifyWorker sends a task to the worker service for processing.
func notifyWorker(workerURL string, task *Task) {
	data, _ := json.Marshal(task)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Post(workerURL+"/process", "application/json",
		jsonReaderFrom(data))
	if err != nil {
		log.Printf("[API] Failed to notify worker: %v", err)
		return
	}
	defer resp.Body.Close()
	log.Printf("[API] Worker notified for task %s (status: %d)", task.ID, resp.StatusCode)
}

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}

type byteReader struct {
	data []byte
	pos  int
}

func jsonReaderFrom(data []byte) *byteReader {
	return &byteReader{data: data}
}

func (r *byteReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= len(r.data) {
		return n, fmt.Errorf("EOF")
	}
	return n, nil
}
