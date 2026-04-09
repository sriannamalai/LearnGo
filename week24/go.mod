module github.com/sri/learngo/week24

go 1.25.0

require (
	github.com/arangodb/go-driver/v2 v2.1.3
	github.com/jackc/pgx/v5 v5.7.4
	github.com/prometheus/client_golang v1.22.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.6
)

// Note: In a real project, running `go mod tidy` will generate a go.sum file
// with cryptographic checksums for all dependencies. Since we can't run
// `go mod tidy` without the actual module cache, you'll need to run:
//   cd week24 && go mod tidy
// This will download all transitive dependencies and create go.sum.
//
// For gRPC code generation, you'll also need:
//   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//   protoc --go_out=. --go-grpc_out=. proto/task.proto
