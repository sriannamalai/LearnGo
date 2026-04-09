module github.com/sri/learngo/week11

go 1.25.0

require github.com/arangodb/go-driver/v2 v2.1.3

// Note: In a real project, running `go mod tidy` will generate a go.sum file
// with cryptographic checksums for all dependencies. Since we can't run
// `go mod tidy` without the actual module cache, you'll need to run:
//   cd week11 && go mod tidy
// This will download go-driver and all its transitive dependencies and create go.sum.
