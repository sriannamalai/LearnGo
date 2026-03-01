# Go Learning Curriculum — 24 Weeks

---

## Phase 1: Go Fundamentals (Weeks 1-7)

### Week 1: Basics
- Variables, constants, types (int, float64, string, bool)
- fmt package (Println, Printf, Sprintf)
- Control flow: if/else, for loops, switch
- Projects: Hello World, simple calculator

### Week 2: Functions & Error Handling
- Function syntax, multiple return values
- Named returns, variadic functions
- Error handling patterns (error interface)
- Projects: Temperature converter, basic CLI tool

### Week 3: Structs, Methods & Interfaces
- Defining structs, struct methods (value vs pointer receivers)
- Interfaces, implicit implementation
- Type assertions, empty interface
- Project: Contact book

### Week 4: Collections
- Arrays, slices (append, copy, slicing)
- Maps (CRUD, iteration)
- Range keyword
- Project: Word frequency counter

### Week 5: Pointers, Packages & Modules
- Pointers (& and *), nil
- Creating packages, go modules (go mod init)
- Exported vs unexported identifiers
- Project: Multi-file organized project

### Week 6: Concurrency
- Goroutines, sync.WaitGroup
- Channels (buffered/unbuffered), select
- Mutex, common patterns
- Project: Parallel web fetcher

### Week 7: File I/O & JSON
- Reading/writing files (os, bufio)
- JSON marshal/unmarshal, struct tags
- Working with HTTP client
- Project: Config reader, JSON API client

---

## Phase 2: Web & Databases (Weeks 8-11)

### Week 8: HTTP Servers & REST
- net/http package, handlers, middleware
- Routing, request/response handling
- JSON APIs
- Project: Simple REST API

### Week 9: Testing & Benchmarking
- testing package, table-driven tests
- Benchmarks, test coverage
- Testable code patterns, mocking
- Project: Tests for all prior weeks

### Week 10: Databases with PostgreSQL
- database/sql interface, pgx driver
- Connecting to PostgreSQL, connection pooling
- CRUD operations, prepared statements
- Migrations with golang-migrate
- Project: Todo app with PostgreSQL persistence

### Week 11: ArangoDB (Multi-Model Database)
- Introduction to ArangoDB — document, graph, and key-value in one
- go-driver for ArangoDB (arangodb/go-driver)
- Document collections: create, read, update, delete
- AQL (ArangoDB Query Language) from Go
- Graph databases: vertices, edges, traversals
- Project: Social network graph (users + relationships)

---

## Phase 3: System & Network Programming (Weeks 12-15)

### Week 12: System Programming Fundamentals
- The os and syscall packages, golang.org/x/sys
- Process management: exec.Command, process lifecycle
- Signal handling with os/signal (SIGTERM, SIGINT, SIGHUP)
- Environment variables, file permissions, filesystem ops
- Project: Process monitor (list/watch system processes)

### Week 13: Advanced System Programming
- Working with cgroups and namespaces (Linux concepts)
- Pipes, IPC (inter-process communication)
- Memory-mapped files, low-level file operations
- Project: Simple container runtime (namespace isolation demo)

### Week 14: Network Programming
- net package deep dive: TCP and UDP servers/clients
- Connection handling with goroutines
- Building custom protocols over TCP
- DNS lookups and resolution
- Project: TCP chat server (multi-client)

### Week 15: System Utilities Development
- Building tools like: file watcher, log parser, disk usage analyzer
- Working with filepath, os, bufio for real-world tooling
- Daemon patterns, graceful shutdown
- Project: Multi-tool system utility (file watcher + log tail + disk monitor)

---

## Phase 4: Microservices — Beginner to Advanced (Weeks 16-19)

### Week 16: Microservices Fundamentals
- Microservices architecture principles
- gRPC and Protocol Buffers (protobuf)
- Generating Go code from .proto files
- Unary and streaming RPCs
- Project: Two-service gRPC app (user service + order service)

### Week 17: Microservices — Messaging & Events
- Message queues and event-driven architecture
- NATS / NATS JetStream for pub/sub and streaming
- Async communication patterns between services
- Project: Event-driven order processing pipeline

### Week 18: Microservices — Containers & Deployment
- Writing Dockerfiles for Go services
- Docker Compose for multi-service orchestration
- Health checks, configuration management
- Intro to Kubernetes concepts (pods, services, deployments)
- Project: Dockerized microservices stack

### Week 19: Microservices — Advanced Patterns & Observability
- Circuit breakers, retries, timeouts (resilience patterns)
- OpenTelemetry for distributed tracing
- Prometheus metrics + Grafana dashboards
- Structured logging (slog)
- Service mesh concepts (Istio/Linkerd overview)
- Project: Observable microservices with tracing & metrics

---

## Phase 5: Desktop & Mobile Apps (Weeks 20-22)

### Week 20: Desktop Apps with Fyne
- Fyne framework setup and architecture
- Widgets, layouts, and theming
- Event handling, data binding
- Packaging for macOS, Windows, Linux
- Project: Desktop note-taking app

### Week 21: Advanced Desktop — Wails (Go + Web UI)
- Wails framework: Go backend + HTML/CSS/JS frontend
- Binding Go functions to the frontend
- Building modern UIs with web technologies
- Project: Desktop dashboard app (system stats viewer)

### Week 22: Mobile Apps with Go
- Fyne mobile: targeting iOS and Android
- Go Mobile (golang.org/x/mobile) overview
- Platform-specific considerations and limitations
- Building and packaging for mobile
- Project: Cross-platform mobile task manager

---

## Phase 6: CLI Mastery & Capstone (Weeks 23-24)

### Week 23: Advanced CLI Tools
- Cobra library for CLI apps
- Project layout conventions
- Config management (Viper, environment variables)
- Interactive prompts, progress bars, TUI with Bubble Tea
- Project: Full-featured CLI application

### Week 24: Review & Capstone Project
- Code review of all projects
- Refactoring exercises, best practices audit
- Capstone: Full-stack Go application combining:
  - REST/gRPC API backend
  - PostgreSQL + ArangoDB data layer
  - Desktop or CLI frontend
  - Docker deployment
  - Observability (metrics, logging, tracing)
