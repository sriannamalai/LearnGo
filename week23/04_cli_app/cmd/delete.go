package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

// ========================================
// Delete Command
// ========================================
// Permanently removes a task from the store by its ID.
// In a production app, you might add a --force flag or
// confirmation prompt to prevent accidental deletion.
//
// Examples:
//   tasks delete 1
//   tasks delete 5 --force

var deleteForce bool

var deleteCmd = &cobra.Command{
	Use:   "delete [task ID]",
	Short: "Delete a task permanently",
	Long: `Permanently remove a task from your task list.

The task ID is shown in the list output (e.g., #1, #2, #3).

Examples:
  tasks delete 1
  tasks delete 5`,

	Aliases: []string{"rm", "remove"},

	// Require exactly one argument: the task ID
	Args: cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse the task ID
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid task ID %q: must be a number", args[0])
		}

		// Look up the task first to show its title in confirmation
		task, lookupErr := taskStore.GetByID(id)

		// Delete the task
		if err := taskStore.Delete(id); err != nil {
			return err
		}

		if lookupErr == nil {
			fmt.Printf("Deleted: #%d %s\n", task.ID, task.Title)
		} else {
			fmt.Printf("Deleted task #%d.\n", id)
		}

		if verbose {
			fmt.Printf("[verbose] Task permanently removed from store\n")
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false,
		"skip confirmation prompt")
}
