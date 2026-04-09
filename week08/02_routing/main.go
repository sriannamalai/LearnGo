package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// ========================================
// Week 8, Lesson 2: Routing & Request Parsing
// ========================================
// Go 1.22 introduced enhanced routing in the standard library's ServeMux.
// Now you can match HTTP methods and extract path parameters — features
// that previously required a third-party router like gorilla/mux or chi.
//
// Run this program:
//   go run .
//
// Then try these with curl:
//   curl http://localhost:8080/
//   curl http://localhost:8080/users
//   curl http://localhost:8080/users/42
//   curl -X POST http://localhost:8080/users -d '{"name":"Sri","email":"sri@example.com"}'
//   curl "http://localhost:8080/search?q=golang&page=1&limit=10"
//   curl -X POST http://localhost:8080/form -d "username=sri&password=secret123"

// ========================================
// Data Types
// ========================================

// User represents a user in our system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// In-memory user storage
var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
	{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 8: Routing & Request Parsing")
	fmt.Println("========================================")

	// ========================================
	// 1. Create a new ServeMux (instead of using DefaultServeMux)
	// ========================================
	// Using your own ServeMux is better practice than the default.
	// It gives you explicit control and avoids global state.

	mux := http.NewServeMux()

	// ========================================
	// 2. Go 1.22+ Enhanced Routing: Method + Path Patterns
	// ========================================
	// New pattern syntax: "METHOD /path"
	// You can now specify the HTTP method directly in the pattern!

	// GET / — Home page
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Note: "GET /" only matches the root path exactly in Go 1.22+
		fmt.Fprintf(w, "Welcome to the User API\n")
		fmt.Fprintf(w, "Routes:\n")
		fmt.Fprintf(w, "  GET    /users        - List all users\n")
		fmt.Fprintf(w, "  GET    /users/{id}   - Get user by ID\n")
		fmt.Fprintf(w, "  POST   /users        - Create a user\n")
		fmt.Fprintf(w, "  GET    /search       - Search with query params\n")
		fmt.Fprintf(w, "  POST   /form         - Form submission\n")
	})

	// ========================================
	// 3. Path Parameters with {name} syntax (Go 1.22+)
	// ========================================
	// Use {paramName} in the pattern to capture path segments.
	// Access them with r.PathValue("paramName").

	// GET /users — List all users
	mux.HandleFunc("GET /users", listUsersHandler)

	// GET /users/{id} — Get a specific user by ID
	mux.HandleFunc("GET /users/{id}", getUserHandler)

	// POST /users — Create a new user (from JSON body)
	mux.HandleFunc("POST /users", createUserHandler)

	// ========================================
	// 4. Query Parameters
	// ========================================
	mux.HandleFunc("GET /search", searchHandler)

	// ========================================
	// 5. Form Data Parsing
	// ========================================
	mux.HandleFunc("POST /form", formHandler)

	// ========================================
	// 6. Demonstrating different response patterns
	// ========================================
	mux.HandleFunc("GET /demo/json", jsonResponseHandler)
	mux.HandleFunc("GET /demo/status", statusCodeHandler)
	mux.HandleFunc("GET /demo/redirect", redirectHandler)

	// ========================================
	// Start server with our custom mux
	// ========================================
	addr := ":8080"
	fmt.Printf("\nServer starting on http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")

	// Pass our mux as the handler (instead of nil for DefaultServeMux)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// ========================================
// Handler: List all users
// ========================================
func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] GET /users\n", time.Now().Format("15:04:05"))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ========================================
// Handler: Get user by ID (path parameter)
// ========================================
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// ========================================
	// Extract path parameter using r.PathValue (Go 1.22+)
	// ========================================
	// For the pattern "GET /users/{id}", PathValue("id") returns the
	// value captured from the URL. For /users/42, it returns "42".

	idStr := r.PathValue("id")
	fmt.Printf("[%s] GET /users/%s\n", time.Now().Format("15:04:05"), idStr)

	// Path values are always strings — convert to int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, `{"error": "invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Search for the user
	for _, user := range users {
		if user.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}
	}

	// User not found — return 404
	http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
}

// ========================================
// Handler: Create user (parse JSON request body)
// ========================================
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] POST /users\n", time.Now().Format("15:04:05"))

	// ========================================
	// Parse JSON request body
	// ========================================
	// json.NewDecoder reads from r.Body (the request body)
	// and decodes JSON into our struct.

	var newUser User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newUser); err != nil {
		http.Error(w, `{"error": "invalid JSON body"}`, http.StatusBadRequest)
		return
	}

	// Assign an ID (simple auto-increment)
	newUser.ID = len(users) + 1
	users = append(users, newUser)

	// Return 201 Created with the new user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201
	json.NewEncoder(w).Encode(newUser)
}

// ========================================
// Handler: Search with query parameters
// ========================================
func searchHandler(w http.ResponseWriter, r *http.Request) {
	// ========================================
	// Reading query parameters
	// ========================================
	// URL: /search?q=golang&page=1&limit=10
	// r.URL.Query() returns url.Values (a map[string][]string)

	query := r.URL.Query()

	// Get individual parameters with .Get() (returns first value or "")
	q := query.Get("q")
	page := query.Get("page")
	limit := query.Get("limit")

	fmt.Printf("[%s] GET /search?q=%s&page=%s&limit=%s\n",
		time.Now().Format("15:04:05"), q, page, limit)

	// Set defaults for missing parameters
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "20"
	}

	// Check required parameters
	if q == "" {
		http.Error(w, `{"error": "query parameter 'q' is required"}`, http.StatusBadRequest)
		return
	}

	// Convert numeric parameters
	pageNum, _ := strconv.Atoi(page)
	limitNum, _ := strconv.Atoi(limit)

	// Return search results (simulated)
	w.Header().Set("Content-Type", "application/json")
	result := map[string]any{
		"query":   q,
		"page":    pageNum,
		"limit":   limitNum,
		"results": []string{"result1", "result2", "result3"},
		"total":   42,
	}
	json.NewEncoder(w).Encode(result)
}

// ========================================
// Handler: Form data parsing
// ========================================
func formHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] POST /form\n", time.Now().Format("15:04:05"))

	// ========================================
	// Parse form data from request body
	// ========================================
	// r.ParseForm() must be called before accessing form values.
	// It parses both URL query parameters and POST form data.

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Access individual form fields with r.FormValue()
	username := r.FormValue("username")
	password := r.FormValue("password")

	// You can also access r.PostForm for POST-body-only values
	// r.Form includes both URL query params and POST body

	fmt.Fprintf(w, "Form Data Received:\n")
	fmt.Fprintf(w, "  Username: %s\n", username)
	fmt.Fprintf(w, "  Password: %s (length: %d)\n", "****", len(password))

	// Show all form values
	fmt.Fprintf(w, "\nAll Form Values:\n")
	for key, values := range r.Form {
		fmt.Fprintf(w, "  %s = %v\n", key, values)
	}
}

// ========================================
// Handler: JSON response pattern
// ========================================
func jsonResponseHandler(w http.ResponseWriter, r *http.Request) {
	// ========================================
	// Proper JSON response pattern
	// ========================================
	// Always set Content-Type BEFORE writing the body.
	// Use json.NewEncoder for streaming or json.Marshal for building first.

	type Response struct {
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
		Version   string `json:"version"`
	}

	resp := Response{
		Message:   "This is a properly formatted JSON response",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ========================================
// Handler: Status code examples
// ========================================
func statusCodeHandler(w http.ResponseWriter, r *http.Request) {
	// ========================================
	// HTTP Status Codes
	// ========================================
	// Go provides named constants for all standard HTTP status codes.

	code := r.URL.Query().Get("code")

	fmt.Fprintf(w, "Common HTTP Status Codes:\n\n")
	fmt.Fprintf(w, "2xx Success:\n")
	fmt.Fprintf(w, "  %d - %s\n", http.StatusOK, http.StatusText(http.StatusOK))
	fmt.Fprintf(w, "  %d - %s\n", http.StatusCreated, http.StatusText(http.StatusCreated))
	fmt.Fprintf(w, "  %d - %s\n", http.StatusNoContent, http.StatusText(http.StatusNoContent))
	fmt.Fprintf(w, "\n4xx Client Errors:\n")
	fmt.Fprintf(w, "  %d - %s\n", http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	fmt.Fprintf(w, "  %d - %s\n", http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	fmt.Fprintf(w, "  %d - %s\n", http.StatusForbidden, http.StatusText(http.StatusForbidden))
	fmt.Fprintf(w, "  %d - %s\n", http.StatusNotFound, http.StatusText(http.StatusNotFound))
	fmt.Fprintf(w, "\n5xx Server Errors:\n")
	fmt.Fprintf(w, "  %d - %s\n", http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	fmt.Fprintf(w, "  %d - %s\n", http.StatusBadGateway, http.StatusText(http.StatusBadGateway))
	fmt.Fprintf(w, "  %d - %s\n", http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable))

	if code != "" {
		fmt.Fprintf(w, "\nYou requested code: %s\n", code)
	}
}

// ========================================
// Handler: Redirect example
// ========================================
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// http.Redirect sends a redirect response to the client
	// StatusFound (302) = temporary redirect
	// StatusMovedPermanently (301) = permanent redirect
	http.Redirect(w, r, "/", http.StatusFound)
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Go 1.22+ Enhanced ServeMux Patterns:
//   "GET /users"        — matches GET requests to /users
//   "POST /users"       — matches POST requests to /users
//   "GET /users/{id}"   — captures path parameter "id"
//   "DELETE /users/{id}" — method + path param together
//
// Request Parsing:
//   r.PathValue("id")   — get path parameter (Go 1.22+)
//   r.URL.Query()       — get query parameters (?key=value)
//   r.FormValue("key")  — get form data (after ParseForm)
//   json.NewDecoder(r.Body) — parse JSON request body
//
// Response Writing:
//   w.Header().Set(...)  — set response headers
//   w.WriteHeader(code)  — set status code (call before Write)
//   json.NewEncoder(w)   — write JSON response
//   http.Error(w, msg, code) — write error response
//   http.Redirect(w, r, url, code) — redirect
