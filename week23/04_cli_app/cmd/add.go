package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sri/learngo/week23/04_cli_app/internal/model"
)

// ========================================
// Add Command
// ========================================
// Adds a new task to the store. The task title is provided as a
// positional argument, and optional flags set priority.
//
// Examples:
//   tasks add "Buy groceries"
//   tasks add "Deploy to production" --priority high
//   tasks add "Read a book" -p low

var addPriority string

var addCmd = &cobra.Command{
	Use:   "add [task title]",
	Short: "Add a new task",
	Long: `Add a new task to your task list.

The task title is provided as arguments (all args are joined).
Use --priority to set urgency level (low, medium, high).

Examples:
  tasks add Buy groceries
  tasks add "Deploy to production" --priority high
  tasks add Read a book -p low`,

	// Require at least one argument (the task title)
	Args: cobra.MinimumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		// Join all positional arguments into the title.
		// This lets users write: tasks add Buy groceries
		// instead of requiring: tasks add "Buy groceries"
		title := strings.Join(args, " ")

		// Validate priority
		priority, err := model.ValidatePriority(addPriority)
		if err != nil {
			return err
		}

		// Create and save the task
		task := model.NewTask(title, priority)
		if err := taskStore.Add(task); err != nil {
			return fmt.Errorf("saving task: %w", err)
		}

		if verbose {
			fmt.Printf("[verbose] Task saved to store\n")
		}

		fmt.Printf("Added task #%d: %s [%s]\n", task.ID, task.Title, task.Priority)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addPriority, "priority", "p", "medium",
		"task priority: low, medium, or high")
}
