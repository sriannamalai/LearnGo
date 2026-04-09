# TaskFlow - Capstone Project

**Week 24 of the LearnGo Curriculum**

TaskFlow is a full-stack Go application that demonstrates production-quality patterns accumulated over 24 weeks of Go learning. It combines REST APIs, gRPC services, dual database storage, distributed tracing, metrics, and containerized deployment.

## Architecture

```
                    +-----------+
                    |  Client   |
                    +-----+-----+
                          |
              +-----------+-----------+
              |                       |
         REST (8080)             gRPC (9090)
              |                       |
    +---------+---------+   +---------+---------+
    | HTTP Middleware    |   | gRPC Interceptors |
    | - Request ID      |   | - Logging         |
    | - Logging         |   | - Recovery        |
    | - CORS            |   | - Auth            |
    | - Auth            |   +--------+----------+
    | - Recovery        |            |
    +---------+---------+            |
              |                      |
              +----------+-----------+
                         |
                  +------+------+
                  | Service     |
                  | Layer       |
                  | (Business   |
                  |  Logic)     |
                  +------+------+
                         |
              +----------+----------+
              |                     |
     +--------+--------+  +--------+--------+
     | PostgreSQL       |  | ArangoDB        |
     | (Tasks, Users)   |  | (Graph, Tags)   |
     +---------+--------+  +--------+--------+
               |                    |
     +---------+--------+  +-------+---------+
     | Observability                          |
     | - OpenTelemetry Tracing (Jaeger)       |
     | - Prometheus Metrics                   |
     | - Structured Logging (slog)            |
     +----------------------------------------+
```

## Curriculum Concepts Demonstrated

| Week | Topic | Where Used |
|------|-------|------------|
| 1-5 | Core Go (types, functions, errors, structs, interfaces) | Throughout all files |
| 6-7 | Concurrency (goroutines, channels, context) | `cmd/server.go` — concurrent servers, graceful shutdown |
| 8-9 | HTTP (net/http, routing, middleware, JSON) | `internal/api/rest.go`, `internal/api/middleware.go` |
| 10 | PostgreSQL (pgx, pooling, queries) | `internal/store/postgres.go` |
| 11 | ArangoDB (documents, graphs, AQL) | `internal/store/arango.go` |
| 12 | Testing (unit, integration, mocks) | Interface-based design enables testability |
| 13-14 | Design Patterns (repository, service, DI) | `internal/service/`, `internal/store/` |
| 15-16 | gRPC (protobuf, services) | `internal/api/grpc.go`, `proto/task.proto` |
| 17-18 | Observability (tracing, metrics, logging) | `internal/observability/` |
| 19-20 | Docker (multi-stage builds, compose) | `Dockerfile`, `docker-compose.yml` |
| 21-22 | Security (auth, CORS, validation) | `internal/api/middleware.go` |
| 23 | CLI Tools (Cobra, Viper) | `cmd/server.go`, `cmd/migrate.go` |

## Project Structure

```
01_capstone/
├── main.go                          # Entry point
├── cmd/
│   ├── server.go                    # Server startup with graceful shutdown
│   └── migrate.go                   # Database migration command
├── internal/
│   ├── api/
│   │   ├── rest.go                  # REST API handlers (Go 1.22+ routing)
│   │   ├── middleware.go            # HTTP middleware stack
│   │   └── grpc.go                  # gRPC service implementation
│   ├── model/
│   │   ├── task.go                  # Task domain model with state machine
│   │   └── user.go                  # User domain model
│   ├── store/
│   │   ├── postgres.go              # PostgreSQL repository (pgx)
│   │   └── arango.go                # ArangoDB repository (graph)
│   ├── service/
│   │   ├── task_service.go          # Task business logic
│   │   └── user_service.go          # User authentication and management
│   └── observability/
│       ├── tracing.go               # OpenTelemetry tracing setup
│       ├── metrics.go               # Prometheus metrics definitions
│       └── logging.go               # Structured logging with slog
├── proto/
│   └── task.proto                   # Protobuf service definitions
├── migrations/
│   └── 001_create_tables.sql        # Database schema
├── config.yaml                      # Application configuration
├── prometheus.yml                   # Prometheus scrape configuration
├── Dockerfile                       # Multi-stage Docker build
├── docker-compose.yml               # Full stack orchestration
└── README.md                        # This file
```

## Getting Started

### Prerequisites

- Go 1.25+
- Docker and Docker Compose (for full stack)
- PostgreSQL 17+ (for local development without Docker)
- ArangoDB 3.12+ (optional, for graph features)

### Quick Start with Docker

```bash
# Start the full stack (app + PostgreSQL + ArangoDB + Prometheus + Jaeger)
cd week24/01_capstone
docker-compose up

# The following services will be available:
#   REST API:    http://localhost:8080
#   gRPC:        localhost:9090
#   Metrics:     http://localhost:8080/metrics
#   Health:      http://localhost:8080/health
#   Jaeger UI:   http://localhost:16686
#   Prometheus:  http://localhost:9091
#   ArangoDB UI: http://localhost:8529
```

### Local Development

```bash
# Install dependencies
cd week24
go mod tidy

# Run database migrations
go run ./01_capstone migrate up

# Start the server
go run ./01_capstone serve

# With custom ports
go run ./01_capstone serve --http-port 3000 --grpc-port 50051
```

## API Reference

### REST Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/tasks` | Create a new task |
| `GET` | `/api/v1/tasks` | List tasks (with filters) |
| `GET` | `/api/v1/tasks/{id}` | Get a specific task |
| `PUT` | `/api/v1/tasks/{id}` | Update a task |
| `DELETE` | `/api/v1/tasks/{id}` | Delete a task |
| `POST` | `/api/v1/tasks/{id}/complete` | Mark task as completed |
| `POST` | `/api/v1/users/register` | Register a new user |
| `POST` | `/api/v1/users/login` | Authenticate a user |
| `GET` | `/api/v1/users/me` | Get current user profile |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |

### Example API Calls

```bash
# Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go", "priority": "high", "tags": ["learning", "go"]}'

# List pending tasks
curl http://localhost:8080/api/v1/tasks?status=pending

# Complete a task
curl -X POST http://localhost:8080/api/v1/tasks/{id}/complete

# Register a user
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"email": "dev@example.com", "name": "Go Developer", "password": "securepassword123"}'
```

## Configuration

Configuration is loaded from (highest to lowest priority):

1. Environment variables (`TASKFLOW_*`)
2. Config file (`config.yaml`)
3. Default values

Key configuration options:

| Variable | Default | Description |
|----------|---------|-------------|
| `TASKFLOW_SERVER_HTTP_PORT` | 8080 | HTTP server port |
| `TASKFLOW_SERVER_GRPC_PORT` | 9090 | gRPC server port |
| `TASKFLOW_DATABASE_POSTGRES_HOST` | localhost | PostgreSQL host |
| `TASKFLOW_OBSERVABILITY_LOGGING_LEVEL` | info | Log level (debug/info/warn/error) |
| `TASKFLOW_OBSERVABILITY_LOGGING_FORMAT` | json | Log format (json/text) |

## Design Decisions

**Why net/http over Gin/Echo?** Go 1.22+ enhanced routing is powerful enough for most APIs. Using the standard library means zero HTTP framework dependencies and full control over the request lifecycle.

**Why pgx over GORM?** Direct SQL with pgx provides full query control, better performance, and no ORM magic. Every query is explicit and auditable.

**Why both PostgreSQL and ArangoDB?** Each database excels at different access patterns. PostgreSQL handles transactional CRUD operations, while ArangoDB provides efficient graph traversals for task dependencies.

**Why the service layer pattern?** Separating business logic from transport (HTTP/gRPC) and storage allows each layer to be tested independently and evolve without affecting others.

**Why slog over zerolog/zap?** slog is part of the Go standard library (since Go 1.21), meaning zero external dependencies for logging. It provides structured logging with good performance for most use cases.
