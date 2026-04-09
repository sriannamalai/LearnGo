module github.com/sri/learngo/week10

go 1.25.0

require github.com/jackc/pgx/v5 v5.7.4

// Note: In a real project, running `go mod tidy` will generate a go.sum file
// with cryptographic checksums for all dependencies. Since we can't run
// `go mod tidy` without the actual module cache, you'll need to run:
//   cd week10 && go mod tidy
// This will download pgx and all its transitive dependencies and create go.sum.
