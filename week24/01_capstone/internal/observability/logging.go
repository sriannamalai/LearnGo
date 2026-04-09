package observability

import (
	"log/slog"
	"os"
	"strings"
)

// ========================================
// Structured Logging with slog
// ========================================
// Go 1.21 introduced log/slog as the standard structured logging
// package. It replaces the older log package with:
//   - Structured key-value pairs instead of printf-style messages
//   - Configurable output format (JSON for machines, text for humans)
//   - Log levels (Debug, Info, Warn, Error)
//   - Contextual attributes via logger.With()
//   - Handler interface for custom outputs
//
// This demonstrates:
//   - Week 17-18: Structured logging patterns
//   - Week 1-5: Standard library usage
//
// Why structured logging matters:
//   - Machine-parseable (JSON) for log aggregation tools (ELK, Loki)
//   - Consistent format across the entire application
//   - Queryable fields (find all errors for user X, or all slow queries)
//   - Context propagation (request ID, user ID in every log line)

// NewLogger creates a configured slog.Logger based on the given
// level and format settings. This is the application's logging
// factory function — called once during startup.
func NewLogger(level, format string) *slog.Logger {
	// ========================================
	// Parse Log Level
	// ========================================
	// slog supports four levels:
	//   Debug (-4) — verbose debugging info, disabled in production
	//   Info  (0)  — normal operation events
	//   Warn  (4)  — something unexpected but recoverable
	//   Error (8)  — something failed, needs attention
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn", "warning":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// ========================================
	// Configure Handler
	// ========================================
	// The handler determines the output format.
	//   - JSONHandler: machine-readable JSON (for production)
	//   - TextHandler: human-readable text (for development)
	opts := &slog.HandlerOptions{
		Level: slogLevel,
		// AddSource adds the source file and line number to each log entry.
		// Useful for debugging but adds overhead — enable in development.
		AddSource: slogLevel == slog.LevelDebug,
	}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "json":
		// JSON format for production and log aggregation
		// Output:
		//   {"time":"2025-01-15T10:30:00Z","level":"INFO","msg":"task created","id":"abc123","user":"user001"}
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		// Text format for development
		// Output:
		//   time=2025-01-15T10:30:00Z level=INFO msg="task created" id=abc123 user=user001
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	// ========================================
	// Create Logger with Default Attributes
	// ========================================
	// logger.With() adds attributes that appear in EVERY log entry
	// from this logger. This is how you add context that's relevant
	// to all operations (service name, version, etc.).
	logger := slog.New(handler).With(
		"service", "taskflow",
		"version", "1.0.0",
	)

	return logger
}

// ========================================
// Logging Best Practices
// ========================================
//
// 1. Use structured fields, not string formatting:
//    GOOD: slog.Info("task created", "id", task.ID, "user", userID)
//    BAD:  slog.Info(fmt.Sprintf("task %s created by %s", task.ID, userID))
//
// 2. Use appropriate log levels:
//    Debug: Detailed info for troubleshooting (SQL queries, cache hits)
//    Info:  Normal operations (server started, request handled)
//    Warn:  Unexpected but handled (retrying connection, deprecated API)
//    Error: Something failed (query error, external service down)
//
// 3. Add context with logger.With():
//    reqLogger := logger.With("request_id", reqID, "user_id", userID)
//    reqLogger.Info("processing request")  // Both fields included
//
// 4. Group related attributes:
//    slog.Info("db query", slog.Group("query",
//        slog.String("table", "tasks"),
//        slog.Duration("duration", elapsed),
//        slog.Int("rows", count),
//    ))
//
// 5. Use Error level with error attribute:
//    slog.Error("failed to create task", "error", err, "user", userID)
//
// 6. Never log sensitive data:
//    BAD:  slog.Info("login", "password", password)
//    GOOD: slog.Info("login", "email", email)
