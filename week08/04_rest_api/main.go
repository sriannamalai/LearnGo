package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ========================================
// Week 8, Mini-Project: Books REST API
// ========================================
// A complete RESTful API for managing a "books" resource.
// Implements all CRUD operations with proper HTTP methods,
// status codes, JSON request/response, and error handling.
//
// Run this program:
//   go run .
//
// Test with curl:
//   # List all books
//   curl http://localhost:8080/books
//
//   # Get a single book
//   curl http://localhost:8080/books/1
//
//   # Create a book
//   curl -X POST http://localhost:8080/books \
//     -H "Content-Type: application/json" \
//     -d '{"title":"The Go Programming Language","author":"Donovan & Kernighan","year":2015,"isbn":"978-0134190440"}'
//
//   # Update a book
//   curl -X PUT http://localhost:8080/books/1 \
//     -H "Content-Type: application/json" \
//     -d '{"title":"Updated Title","author":"Updated Author","year":2024,"isbn":"000-0000000000"}'
//
//   # Delete a book
//   curl -X DELETE http://localhost:8080/books/1
//
//   # Try getting a deleted book (404)
//   curl http://localhost:8080/books/1

// ========================================
// Data Model
// ========================================

// Book represents a book resource in our API
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   int    `json:"year,omitempty"`
	ISBN   string `json:"isbn,omitempty"`
}

// ========================================
// In-Memory Store
// ========================================
// BookStore manages our collection of books.
// Uses a sync.RWMutex for safe concurrent access — important because
// HTTP handlers run in separate goroutines!

type BookStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

// NewBookStore creates a store pre-loaded with sample books
func NewBookStore() *BookStore {
	store := &BookStore{
		books:  make(map[int]Book),
		nextID: 1,
	}

	// Seed with some sample data
	store.Create(Book{Title: "The Go Programming Language", Author: "Donovan & Kernighan", Year: 2015, ISBN: "978-0134190440"})
	store.Create(Book{Title: "Learning Go", Author: "Jon Bodner", Year: 2021, ISBN: "978-1492077213"})
	store.Create(Book{Title: "Concurrency in Go", Author: "Katherine Cox-Buday", Year: 2017, ISBN: "978-1491941195"})

	return store
}

// List returns all books in the store
func (s *BookStore) List() []Book {
	s.mu.RLock()
	defer s.mu.RUnlock()

	books := make([]Book, 0, len(s.books))
	for _, book := range s.books {
		books = append(books, book)
	}
	return books
}

// Get returns a single book by ID
func (s *BookStore) Get(id int) (Book, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	book, ok := s.books[id]
	return book, ok
}

// Create adds a new book and returns it with the assigned ID
func (s *BookStore) Create(book Book) Book {
	s.mu.Lock()
	defer s.mu.Unlock()

	book.ID = s.nextID
	s.nextID++
	s.books[book.ID] = book
	return book
}

// Update replaces a book by ID. Returns false if not found.
func (s *BookStore) Update(id int, book Book) (Book, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; !ok {
		return Book{}, false
	}

	book.ID = id
	s.books[id] = book
	return book, true
}

// Delete removes a book by ID. Returns false if not found.
func (s *BookStore) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; !ok {
		return false
	}

	delete(s.books, id)
	return true
}

// ========================================
// API Response Types
// ========================================

// ErrorResponse is the standard error format for our API
type ErrorResponse struct {
	Error   string `json:"error"`
	Status  int    `json:"status"`
	Message string `json:"message,omitempty"`
}

// ========================================
// Helper Functions
// ========================================

// writeJSON sends a JSON response with the given status code
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}

// writeError sends a JSON error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Status:  status,
		Message: message,
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 8 Project: Books REST API")
	fmt.Println("========================================")

	// Initialize our data store
	store := NewBookStore()

	// ========================================
	// Set up routes using Go 1.22+ patterns
	// ========================================
	mux := http.NewServeMux()

	// API info endpoint
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":    "Books REST API",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"GET /books":        "List all books",
				"GET /books/{id}":   "Get a book by ID",
				"POST /books":       "Create a new book",
				"PUT /books/{id}":   "Update a book",
				"DELETE /books/{id}": "Delete a book",
			},
		})
	})

	// ========================================
	// CRUD Routes
	// ========================================

	// GET /books — List all books
	mux.HandleFunc("GET /books", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)
		books := store.List()
		writeJSON(w, http.StatusOK, books)
	})

	// GET /books/{id} — Get a single book
	mux.HandleFunc("GET /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "ID must be a number")
			return
		}

		book, ok := store.Get(id)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Sprintf("Book with ID %d not found", id))
			return
		}

		writeJSON(w, http.StatusOK, book)
	})

	// POST /books — Create a new book
	mux.HandleFunc("POST /books", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		// Parse the JSON request body
		var book Book
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		// Validate required fields
		if book.Title == "" {
			writeError(w, http.StatusBadRequest, "Title is required")
			return
		}
		if book.Author == "" {
			writeError(w, http.StatusBadRequest, "Author is required")
			return
		}

		// Create the book
		created := store.Create(book)

		// Return 201 Created with Location header
		w.Header().Set("Location", fmt.Sprintf("/books/%d", created.ID))
		writeJSON(w, http.StatusCreated, created)
	})

	// PUT /books/{id} — Update a book
	mux.HandleFunc("PUT /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "ID must be a number")
			return
		}

		// Parse the JSON request body
		var book Book
		if err := json.NewDecoder(r.Body).Decode(&book); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
			return
		}

		// Validate required fields
		if book.Title == "" {
			writeError(w, http.StatusBadRequest, "Title is required")
			return
		}
		if book.Author == "" {
			writeError(w, http.StatusBadRequest, "Author is required")
			return
		}

		// Update the book
		updated, ok := store.Update(id, book)
		if !ok {
			writeError(w, http.StatusNotFound, fmt.Sprintf("Book with ID %d not found", id))
			return
		}

		writeJSON(w, http.StatusOK, updated)
	})

	// DELETE /books/{id} — Delete a book
	mux.HandleFunc("DELETE /books/{id}", func(w http.ResponseWriter, r *http.Request) {
		logRequest(r)

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "ID must be a number")
			return
		}

		if !store.Delete(id) {
			writeError(w, http.StatusNotFound, fmt.Sprintf("Book with ID %d not found", id))
			return
		}

		// 204 No Content — successful deletion, no response body
		w.WriteHeader(http.StatusNoContent)
	})

	// ========================================
	// Apply middleware
	// ========================================
	var handler http.Handler = mux
	handler = loggingMiddleware(handler)

	// ========================================
	// Start the server
	// ========================================
	addr := ":8080"
	fmt.Printf("\nServer starting on http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("Quick test commands:")
	fmt.Println("  curl http://localhost:8080/books")
	fmt.Println("  curl http://localhost:8080/books/1")
	fmt.Println(`  curl -X POST http://localhost:8080/books -H "Content-Type: application/json" -d '{"title":"New Book","author":"Author"}'`)
	fmt.Println(`  curl -X PUT http://localhost:8080/books/1 -H "Content-Type: application/json" -d '{"title":"Updated","author":"Author"}'`)
	fmt.Println("  curl -X DELETE http://localhost:8080/books/1")

	log.Fatal(http.ListenAndServe(addr, handler))
}

// ========================================
// Request Logging
// ========================================

func logRequest(r *http.Request) {
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)
}

// loggingMiddleware logs each request with method, path, status, and duration
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		wrapped := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		fmt.Printf("[API] %s %s -> %d (%s)\n",
			r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.statusCode = code
	sw.ResponseWriter.WriteHeader(code)
}

// ========================================
// Sample curl session:
// ========================================
//
// $ curl -s http://localhost:8080/books | jq .
// [
//   {"id": 1, "title": "The Go Programming Language", "author": "Donovan & Kernighan", "year": 2015, "isbn": "978-0134190440"},
//   {"id": 2, "title": "Learning Go", "author": "Jon Bodner", "year": 2021, "isbn": "978-1492077213"},
//   {"id": 3, "title": "Concurrency in Go", "author": "Katherine Cox-Buday", "year": 2017, "isbn": "978-1491941195"}
// ]
//
// $ curl -s http://localhost:8080/books/1 | jq .
// {"id": 1, "title": "The Go Programming Language", "author": "Donovan & Kernighan", "year": 2015, "isbn": "978-0134190440"}
//
// $ curl -s -X POST http://localhost:8080/books -H "Content-Type: application/json" \
//     -d '{"title":"Go in Action","author":"Kennedy, Ketelsen, Martin","year":2015}' | jq .
// {"id": 4, "title": "Go in Action", "author": "Kennedy, Ketelsen, Martin", "year": 2015}
//
// $ curl -s -X PUT http://localhost:8080/books/4 -H "Content-Type: application/json" \
//     -d '{"title":"Go in Action, 2nd Ed","author":"Kennedy, Ketelsen, Martin","year":2024}' | jq .
// {"id": 4, "title": "Go in Action, 2nd Ed", "author": "Kennedy, Ketelsen, Martin", "year": 2024}
//
// $ curl -s -X DELETE http://localhost:8080/books/4 -w "\nHTTP Status: %{http_code}\n"
// HTTP Status: 204
//
// $ curl -s http://localhost:8080/books/999 | jq .
// {"error": "Not Found", "status": 404, "message": "Book with ID 999 not found"}
