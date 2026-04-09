package api

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// ========================================
// HTTP Middleware
// ========================================
// Middleware functions wrap HTTP handlers to add cross-cutting
// concerns like logging, authentication, CORS, and panic recovery.
// This demonstrates:
//   - Week 8-9: HTTP middleware pattern
//   - Week 21-22: Security middleware (auth, CORS)
//   - Week 17-18: Observability (structured logging, request tracing)
//   - Week 5: Function types and closures
//
// Architecture Decision: We use the standard middleware pattern
// (func(http.Handler) http.Handler) rather than framework-specific
// middleware. This makes our middleware composable and reusable
// with any HTTP framework or the standard library.

// Middleware is a function that wraps an HTTP handler.
type Middleware func(http.Handler) http.Handler

// ChainMiddleware applies middlewares in order (first middleware wraps outermost).
// The first middleware in the list is the outermost (runs first).
func ChainMiddleware(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply in reverse order so the first middleware wraps outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// ========================================
// Request ID Middleware
// ========================================
// Assigns a unique ID to every request for tracing through logs.
// This is essential for correlating log entries in a distributed system.

// requestIDKey is the context key for the request ID.
const requestIDKey contextKey = "request_id"

// RequestIDMiddleware generates a unique request ID and adds it to
// the request context and response headers.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client already provided a request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add to response header for client-side correlation
		w.Header().Set("X-Request-ID", requestID)

		// Add to context for downstream handlers and logging
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ========================================
// Logging Middleware
// ========================================
// Logs every request with structured fields including method,
// path, status code, duration, and request ID. This is the
// observability pattern from Week 17-18.

// LoggingMiddleware creates a middleware that logs HTTP requests.
func LoggingMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the ResponseWriter to capture the status code
			wrapped := &statusRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Log the request with structured fields
			duration := time.Since(start)
			requestID, _ := r.Context().Value(requestIDKey).(string)

			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"request_id", requestID,
			)
		})
	}
}

// statusRecorder wraps http.ResponseWriter to capture the status code.
// This is necessary because http.ResponseWriter doesn't expose the
// status code after WriteHeader is called.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// ========================================
// CORS Middleware
// ========================================
// Cross-Origin Resource Sharing allows web browsers to make
// requests to our API from different domains. This is essential
// when the frontend and backend are served from different origins.

// CORSMiddleware creates a middleware that adds CORS headers.
func CORSMiddleware(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if the origin is allowed
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed || len(allowedOrigins) == 0 {
				// Default to allowing all origins in development
				if origin == "" {
					origin = "*"
				}
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Standard CORS headers
			w.Header().Set("Access-Control-Allow-Methods",
				"GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers",
				"Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers",
				"X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			// Browsers send OPTIONS requests before cross-origin requests
			// with custom headers or non-simple methods.
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ========================================
// Auth Middleware
// ========================================
// Validates the Authorization header and extracts the user ID.
// In production, this would verify JWT tokens.

// AuthMiddleware creates a middleware that validates authentication.
func AuthMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for public endpoints
			publicPaths := []string{
				"/health",
				"/metrics",
				"/api/v1/users/register",
				"/api/v1/users/login",
			}
			for _, path := range publicPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract the Bearer token
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// In demo mode, allow unauthenticated access
				next.ServeHTTP(w, r)
				return
			}

			// Parse "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization header"}`,
					http.StatusUnauthorized)
				return
			}
			token := parts[1]

			// In production, validate the JWT token here:
			//   claims, err := jwt.Verify(token)
			//   userID = claims.Subject
			//
			// For this educational example, we extract a demo user ID.
			userID := "demo-user-001"
			_ = token // In production, validate the token

			// Add user ID to context for downstream handlers
			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ========================================
// Recovery Middleware
// ========================================
// Catches panics in HTTP handlers and converts them to 500 errors.
// Without this, a panic would crash the entire server. This is
// especially important in production where unhandled panics in
// a single request would terminate all connections.

// RecoveryMiddleware creates a middleware that recovers from panics.
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					stack := debug.Stack()
					logger.Error("panic recovered",
						"error", err,
						"stack", string(stack),
						"method", r.Method,
						"path", r.URL.Path,
					)

					// Return a generic 500 error
					// Never expose panic details to clients (security)
					http.Error(w,
						`{"error":"internal server error","detail":"an unexpected error occurred"}`,
						http.StatusInternalServerError,
					)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// ========================================
// Helpers
// ========================================

// generateRequestID creates a unique request identifier.
func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("req_%x", b)
}
