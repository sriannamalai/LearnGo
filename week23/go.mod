module github.com/sri/learngo/week23

go 1.25.0

require (
	github.com/charmbracelet/bubbletea v1.3.4
	github.com/charmbracelet/lipgloss v1.1.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
)

// Note: In a real project, running `go mod tidy` will generate a go.sum file
// with cryptographic checksums for all dependencies. Since we can't run
// `go mod tidy` without the actual module cache, you'll need to run:
//   cd week23 && go mod tidy
// This will download all transitive dependencies and create go.sum.
