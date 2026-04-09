package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// ========================================
// Week 8, Lesson 3: Middleware
// ========================================
// Middleware is code that runs before (and/or after) your handler.
// Common uses: logging, authentication, CORS, rate limiting, etc.
//
// In Go, middleware is just a function that takes an http.Handler
// and returns a new http.Handler — wrapping the original.
//
// Run this program:
//   go run .
//
// Then try:
//   curl -v http://localhost:8080/
//   curl -v http://localhost:8080/api/data
//   curl -v http://localhost:8080/admin/dashboard
//   curl -v -H "Authorization: Bearer secret-token-123" http://localhost:8080/admin/dashboard
//   curl -v -H "Origin: http://example.com" http://localhost:8080/api/data

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 8: Middleware")
	fmt.Println("========================================")

	mux := http.NewServeMux()

	// ========================================
	// 1. Basic routes
	// ========================================
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /api/data", apiDataHandler)
	mux.HandleFunc("GET /admin/dashboard", adminDashboardHandler)
	mux.HandleFunc("GET /slow", slowHandler)

	// ========================================
	// 2. Apply middleware by wrapping the mux
	// ========================================
	// Middleware wraps around handlers like layers of an onion.
	// The outermost middleware runs first on the request,
	// and last on the response.
	//
	// Request flow:  Logging -> CORS -> RequestID -> Auth -> Handler
	// Response flow: Handler -> Auth -> RequestID -> CORS -> Logging

	// Chain middleware: each wraps the previous result
	var handler http.Handler = mux
	handler = authMiddleware(handler)       // Innermost (closest to handler)
	handler = requestIDMiddleware(handler)  // Adds request ID
	handler = corsMiddleware(handler)       // Adds CORS headers
	handler = loggingMiddleware(handler)    // Outermost (runs first)

	// ========================================
	// 3. Start server with middleware-wrapped handler
	// ========================================
	addr := ":8080"
	fmt.Printf("\nServer starting on http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("Try these:")
	fmt.Println("  curl -v http://localhost:8080/")
	fmt.Println("  curl -v http://localhost:8080/api/data")
	fmt.Println("  curl -v http://localhost:8080/admin/dashboard")
	fmt.Println("  curl -v -H 'Authorization: Bearer secret-token-123' http://localhost:8080/admin/dashboard")

	log.Fatal(http.ListenAndServe(addr, handler))
}

// ========================================
// Middleware 1: Logging
// ========================================
// Logs every request with method, path, duration, and status code.
//
// This is the most common middleware — almost every web app needs it.

func loggingMiddleware(next http.Handler) http.Handler {
	// http.HandlerFunc is an adapter that lets you use a regular function
	// as an http.Handler. It implements the Handler interface by calling itself.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture the status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler in the chain
		next.ServeHTTP(wrapped, r)

		// Log after the handler has finished
		duration := time.Since(start)
		fmt.Printf("[LOG] %s %s %d %s\n",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code.
// This is a common pattern because http.ResponseWriter doesn't expose
// the status code after WriteHeader is called.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ========================================
// Middleware 2: CORS (Cross-Origin Resource Sharing)
// ========================================
// CORS headers allow web browsers to make requests from different domains.
// Without these headers, browsers block cross-origin requests.

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers on every response
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS requests
		// Browsers send OPTIONS before certain cross-origin requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// ========================================
// Middleware 3: Request ID
// ========================================
// Assigns a unique ID to each request for tracing and debugging.
// The ID is added to the response header and to the request context.

// contextKey is a custom type to avoid key collisions in context
type contextKey string

const requestIDKey contextKey = "requestID"

func requestIDMiddleware(next http.Handler) http.Handler {
	var counter int
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a simple request ID (in production, use UUID)
		counter++
		requestID := fmt.Sprintf("req-%d-%d", time.Now().UnixMilli(), counter)

		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to the request context
		// This makes it available to all downstream handlers
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		// Call the next handler with the enriched request
		next.ServeHTTP(w, r)
	})
}

// ========================================
// Middleware 4: Authentication
// ========================================
// Checks for a valid Authorization header on protected routes.
// Only applies to routes starting with /admin/.

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only protect /admin/ routes
		if !strings.HasPrefix(r.URL.Path, "/admin") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Validate the token (simplified — in production use JWT or similar)
		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "invalid authorization format, use: Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != "secret-token-123" {
			http.Error(w, `{"error": "invalid token"}`, http.StatusForbidden)
			return
		}

		// Token is valid — proceed to the handler
		fmt.Printf("[AUTH] Authenticated request to %s\n", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ========================================
// Route Handlers
// ========================================

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve request ID from context
	requestID, _ := r.Context().Value(requestIDKey).(string)

	fmt.Fprintf(w, "Welcome! Your request ID: %s\n", requestID)
	fmt.Fprintf(w, "\nRoutes:\n")
	fmt.Fprintf(w, "  GET /           - This page\n")
	fmt.Fprintf(w, "  GET /api/data   - Public API endpoint\n")
	fmt.Fprintf(w, "  GET /admin/dashboard - Protected (needs auth)\n")
	fmt.Fprintf(w, "  GET /slow       - Slow endpoint (2s)\n")
}

func apiDataHandler(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(requestIDKey).(string)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"data": "Hello from the API!", "request_id": "%s"}`, requestID)
	fmt.Fprintln(w)
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(requestIDKey).(string)

	fmt.Fprintf(w, "Welcome to the Admin Dashboard!\n")
	fmt.Fprintf(w, "Request ID: %s\n", requestID)
	fmt.Fprintf(w, "You are authenticated.\n")
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	// This handler takes 2 seconds — the logging middleware will show the duration
	time.Sleep(2 * time.Second)
	fmt.Fprintf(w, "This response took 2 seconds. Check the server logs for timing.\n")
}

// ========================================
// http.Handler vs http.HandlerFunc
// ========================================
//
// http.Handler is an INTERFACE with one method:
//   type Handler interface {
//       ServeHTTP(ResponseWriter, *Request)
//   }
//
// http.HandlerFunc is a TYPE (function adapter):
//   type HandlerFunc func(ResponseWriter, *Request)
//   func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }
//
// This means any function with the right signature can be used as a Handler
// by casting it: http.HandlerFunc(myFunc)
//
// Middleware pattern:
//   func myMiddleware(next http.Handler) http.Handler {
//       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//           // Before handler
//           next.ServeHTTP(w, r)
//           // After handler
//       })
//   }
//
// Chaining:
//   handler := middlewareA(middlewareB(middlewareC(finalHandler)))
//   // Request passes through A -> B -> C -> finalHandler
//   // Response passes through finalHandler -> C -> B -> A

// ========================================
// Bonus: Generic middleware chainer
// ========================================
// In larger apps, you might use a helper to chain middleware:
//
//   type Middleware func(http.Handler) http.Handler
//
//   func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
//       // Apply in reverse order so first middleware is outermost
//       for i := len(middlewares) - 1; i >= 0; i-- {
//           handler = middlewares[i](handler)
//       }
//       return handler
//   }
//
//   // Usage:
//   handler := Chain(mux, loggingMiddleware, corsMiddleware, authMiddleware)
