package main

// ========================================
// Week 23, Lesson 4: Full CLI Application
// ========================================
// This is the capstone project for Week 23. It combines everything
// we learned about CLI development in Go:
//   - Cobra for command structure and flag parsing
//   - Viper for configuration management
//   - Bubble Tea for an interactive TUI mode
//   - Lipgloss for terminal styling
//   - JSON file storage for persistence
//
// This task manager supports both traditional CLI commands and an
// interactive TUI mode, demonstrating professional CLI patterns.
//
// Usage:
//   go run . --help                    # Show help
//   go run . add "Buy groceries"       # Add a task
//   go run . add "Deploy app" -p high  # Add high-priority task
//   go run . list                      # List all tasks
//   go run . list --status pending     # Filter by status
//   go run . list --priority high      # Filter by priority
//   go run . done 1                    # Mark task 1 complete
//   go run . delete 1                  # Delete task 1
//   go run . interactive               # Launch TUI mode
// ========================================

import "github.com/sri/learngo/week23/04_cli_app/cmd"

func main() {
	cmd.Execute()
}
