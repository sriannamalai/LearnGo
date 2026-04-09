package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ========================================
// Done Command
// ========================================
// Marks a task as completed by its ID. This demonstrates Cobra's
// positional argument handling and validation.
//
// Examples:
//   tasks done 1
//   tasks done 3

var doneCmd = &cobra.Command{
	Use:   "done [task ID]",
	Short: "Mark a task as completed",
	Long: `Mark a task as completed by providing its ID.

The task ID is shown in the list output (e.g., #1, #2, #3).

Examples:
  tasks done 1
  tasks done 3`,

	Aliases: []string{"complete", "finish"},

	// Require exactly one argument: the task ID
	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse the task ID from the argument
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID %q: must be a number", args[0])
		}

		// Mark the task as done
		if err := taskStore.MarkDone(id); err != nil {
			return err
		}

		// Fetch the task to show its title in the confirmation
		task, err := taskStore.GetByID(id)
		if err != nil {
			// Task was marked done successfully, just can't fetch for display
			fmt.Printf("Task #%d marked as completed.\n", id)
			return nil
		}

		fmt.Printf("Completed: #%d %s\n", task.ID, task.Title)

		if verbose {
			fmt.Printf("[verbose] Completed at: %s\n",
				task.CompletedAt.Format("2006-01-02 15:04:05"))
		}

		return nil
	},
}
