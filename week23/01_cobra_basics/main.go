package main

// ========================================
// Week 23, Lesson 1: Cobra CLI Basics
// ========================================
// Cobra is the most popular CLI framework in Go. It powers kubectl,
// Hugo, GitHub CLI, and hundreds of other tools. Cobra provides:
//   - Subcommands (app server, app config, etc.)
//   - Flags (persistent and local)
//   - Argument validation
//   - Auto-generated help text
//   - Shell completions
//   - Command aliases
//
// Architecture: Cobra apps follow a cmd/ pattern where each file
// in the cmd/ directory defines one command. The main.go file
// simply calls cmd.Execute().
//
// Run this program:
//   go run . --help
//   go run . greet --name Sri
//   go run . greet --name Sri --shout
//   go run . version
//   go run . version --short
//   go run . hi --name World
// ========================================

import "github.com/sri/learngo/week23/01_cobra_basics/cmd"

func main() {
	// ========================================
	// Entry Point Pattern
	// ========================================
	// The main function in a Cobra app is intentionally minimal.
	// All logic lives in the cmd package. This separation keeps
	// things clean and testable. The Execute() function is defined
	// in cmd/root.go and starts the Cobra command tree.
	cmd.Execute()
}
