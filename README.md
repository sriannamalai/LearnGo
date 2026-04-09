# LearnGo

A structured, hands-on journey to learn Go from scratch — covering fundamentals through advanced topics over 24 weeks. Every week includes numbered lessons with educational source code and a mini-project.

## Curriculum Overview

| Phase | Weeks | Topics |
|-------|-------|--------|
| **1. Go Fundamentals** | 1–7 | Variables, functions, structs, interfaces, collections, concurrency, file I/O |
| **2. Web & Databases** | 8–11 | HTTP/REST, testing, PostgreSQL, ArangoDB |
| **3. System & Network Programming** | 12–15 | Syscalls, processes, TCP/UDP, system utilities |
| **4. Microservices** | 16–19 | gRPC, NATS, Docker, observability (OpenTelemetry, Prometheus) |
| **5. Desktop & Mobile Apps** | 20–22 | Fyne, Wails, mobile development |
| **6. CLI & Capstone** | 23–24 | Cobra, Bubble Tea, full-stack capstone project |

See [CURRICULUM.md](CURRICULUM.md) for the full weekly breakdown.

## Repository Structure

```
LearnGo/
├── CURRICULUM.md
├── README.md
├── week01/                # Basics: hello, variables, control flow, calculator
├── week02/                # Functions & Error Handling
├── week03/                # Structs, Methods & Interfaces
├── week04/                # Collections: arrays, slices, maps
├── week05/                # Pointers, Packages & Modules
├── week06/                # Concurrency: goroutines, channels, mutex
├── week07/                # File I/O & JSON
├── week08/                # HTTP Servers & REST APIs
├── week09/                # Testing & Benchmarking
├── week10/                # PostgreSQL
├── week11/                # ArangoDB
├── week12/                # System Programming Fundamentals
├── week13/                # Advanced System Programming
├── week14/                # Network Programming (TCP/UDP)
├── week15/                # System Utilities Development
├── week16/                # gRPC & Protocol Buffers
├── week17/                # Messaging & Events (NATS)
├── week18/                # Containers & Deployment (Docker/K8s)
├── week19/                # Observability (OpenTelemetry, Prometheus)
├── week20/                # Desktop Apps (Fyne)
├── week21/                # Advanced Desktop (Wails)
├── week22/                # Mobile Apps
├── week23/                # Advanced CLI (Cobra, Bubble Tea)
└── week24/                # Capstone: Full-stack TaskFlow app
```

Each week has its own directory with numbered lessons and a mini-project.

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- A text editor or IDE (VS Code with Go extension recommended)
- Git
- Docker (for weeks 10–11, 17–19)

## Running the Code

Navigate to any lesson directory and run:

```bash
cd week01
go run 01_hello/main.go
```

For weeks with external dependencies, run `go mod tidy` first:

```bash
cd week10
go mod tidy
go run 01_connect/main.go
```

## Progress

### Phase 1: Go Fundamentals
- [x] Week 1: Basics (hello world, variables, control flow, calculator)
- [x] Week 2: Functions & Error Handling (functions, multiple returns, errors, temperature converter)
- [x] Week 3: Structs, Methods & Interfaces (structs, methods, interfaces, contact book)
- [x] Week 4: Collections (arrays, slices, maps, word frequency counter)
- [x] Week 5: Pointers, Packages & Modules (pointers, packages, multi-file project)
- [x] Week 6: Concurrency (goroutines, channels, mutex, parallel web fetcher)
- [x] Week 7: File I/O & JSON (file ops, JSON, HTTP client, config reader)

### Phase 2: Web & Databases
- [x] Week 8: HTTP Servers & REST (server, routing, middleware, books REST API)
- [x] Week 9: Testing & Benchmarking (unit tests, table tests, benchmarks, testable code)
- [x] Week 10: PostgreSQL (connect, CRUD, migrations, todo app)
- [x] Week 11: ArangoDB (documents, AQL, graphs, social network)

### Phase 3: System & Network Programming
- [x] Week 12: System Programming (os, processes, signals, process monitor)
- [x] Week 13: Advanced System Programming (pipes, filesystem, mmap, container demo)
- [x] Week 14: Network Programming (TCP, UDP, DNS, chat server)
- [x] Week 15: System Utilities (file watcher, log parser, disk usage, system toolkit)

### Phase 4: Microservices
- [x] Week 16: gRPC & Protobuf (protobuf, server, client, two-service app)
- [x] Week 17: Messaging & Events (NATS pub/sub, request/reply, JetStream, event pipeline)
- [x] Week 18: Containers & Deployment (Dockerfile, Compose, health checks, K8s manifests)
- [x] Week 19: Observability (circuit breaker, tracing, metrics, observable services)

### Phase 5: Desktop & Mobile Apps
- [x] Week 20: Desktop Apps with Fyne (widgets, layouts, note-taking app)
- [x] Week 21: Advanced Desktop with Wails (bindings, events, system dashboard)
- [x] Week 22: Mobile Apps (Fyne mobile, gomobile, task manager)

### Phase 6: CLI & Capstone
- [x] Week 23: Advanced CLI (Cobra, Viper, Bubble Tea, full CLI app)
- [x] Week 24: Capstone — TaskFlow (REST + gRPC, PostgreSQL + ArangoDB, Docker, observability)
