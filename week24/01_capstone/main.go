package main

// ========================================
// Week 24: Capstone Project — TaskFlow
// ========================================
// TaskFlow is a full-stack Go application that represents the
// culmination of all 24 weeks of the LearnGo curriculum. It
// demonstrates production-quality Go patterns including:
//
// From Week 1-5: Core Go (types, functions, error handling, structs, interfaces)
// From Week 6-7: Concurrency (goroutines, channels, context, graceful shutdown)
// From Week 8-9: HTTP servers (net/http, routing, middleware, JSON APIs)
// From Week 10: PostgreSQL (connection pooling, queries, transactions)
// From Week 11: ArangoDB (document & graph database)
// From Week 12: Testing (unit tests, integration tests, mocks)
// From Week 13-14: Design patterns (repository, service layer, dependency injection)
// From Week 15-16: gRPC (protobuf, services, streaming)
// From Week 17-18: Observability (OpenTelemetry tracing, Prometheus metrics, slog)
// From Week 19-20: Docker (multi-stage builds, docker-compose)
// From Week 21-22: Security (auth middleware, CORS, input validation)
// From Week 23: CLI tools (Cobra commands, Viper config)
//
// Architecture:
//   main.go ──> cmd/        (Cobra CLI commands)
//           ──> internal/api (REST + gRPC handlers)
//           ──> internal/service (business logic)
//           ──> internal/store   (database repositories)
//           ──> internal/model   (domain models)
//           ──> internal/observability (tracing, metrics, logging)
//
// Run:
//   go run . serve             # Start REST + gRPC servers
//   go run . migrate           # Run database migrations
//   go run . serve --help      # Show server options
//   docker-compose up          # Full stack with dependencies
// ========================================

import "github.com/sri/learngo/week24/01_capstone/cmd"

func main() {
	cmd.Execute()
}
