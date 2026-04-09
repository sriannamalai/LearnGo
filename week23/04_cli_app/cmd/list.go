package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/sri/learngo/week23/04_cli_app/internal/model"
)

// ========================================
// List Command
// ========================================
// Lists all tasks with optional filtering by status and priority.
// Uses Lipgloss styling for a polished terminal output.
//
// Examples:
//   tasks list                    # Show all tasks
//   tasks list --status pending   # Show only pending tasks
//   tasks list --priority high    # Show only high-priority tasks
//   tasks list -s completed       # Show completed tasks

var (
	listStatus   string
	listPriority string
)

// Lipgloss styles for the list display
var (
	listTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	listHighStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6F61")).
			Bold(true)

	listMediumStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD93D"))

	listLowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6BCB77"))

	listDoneStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Strikethrough(true)

	listCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long: `Display all tasks in your task list.

Use flags to filter by status or priority.

Examples:
  tasks list                    # Show all
  tasks list --status pending   # Only pending
  tasks list --status completed # Only completed
  tasks list --priority high    # Only high priority`,

	Aliases: []string{"ls", "l"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		tasks := taskStore.GetAll()

		// Apply filters
		if listStatus != "" {
			status := model.Status(listStatus)
			if status != model.StatusPending && status != model.StatusCompleted {
				return fmt.Errorf("invalid status %q: must be pending or completed", listStatus)
			}
			tasks = filterByStatus(tasks, status)
		}

		if listPriority != "" {
			priority, err := model.ValidatePriority(listPriority)
			if err != nil {
				return err
			}
			tasks = filterByPriority(tasks, priority)
		}

		// Display header
		header := " Task List "
		if listStatus != "" {
			header = fmt.Sprintf(" Tasks [%s] ", listStatus)
		}
		if listPriority != "" {
			header = fmt.Sprintf(" Tasks [%s priority] ", listPriority)
		}
		fmt.Println(listTitleStyle.Render(header))
		fmt.Println()

		if len(tasks) == 0 {
			fmt.Println("  No tasks found.")
			fmt.Println("  Use 'tasks add \"My task\"' to create one.")
			return nil
		}

		// Display each task with styling
		pending := 0
		completed := 0

		for _, t := range tasks {
			checkbox := "[ ]"
			if t.IsCompleted() {
				checkbox = "[x]"
				completed++
			} else {
				pending++
			}

			// Style the priority indicator
			var priorityStr string
			switch t.Priority {
			case model.PriorityHigh:
				priorityStr = listHighStyle.Render("HIGH")
			case model.PriorityMedium:
				priorityStr = listMediumStyle.Render("MED ")
			case model.PriorityLow:
				priorityStr = listLowStyle.Render("LOW ")
			}

			// Style the title
			title := t.Title
			if t.IsCompleted() {
				title = listDoneStyle.Render(title)
			}

			// Format the date
			date := t.CreatedAt.Format("Jan 02")

			fmt.Printf("  %s %s #%-3d %s  %s\n",
				checkbox, priorityStr, t.ID, title, date)
		}

		// Summary line
		fmt.Println()
		summary := fmt.Sprintf("  %d pending, %d completed, %d total",
			pending, completed, len(tasks))
		fmt.Println(listCountStyle.Render(summary))

		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "",
		"filter by status: pending or completed")
	listCmd.Flags().StringVarP(&listPriority, "priority", "p", "",
		"filter by priority: low, medium, or high")
}

// filterByStatus returns only tasks with the given status.
func filterByStatus(tasks []*model.Task, status model.Status) []*model.Task {
	var filtered []*model.Task
	for _, t := range tasks {
		if t.Status == status {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// filterByPriority returns only tasks with the given priority.
func filterByPriority(tasks []*model.Task, priority model.Priority) []*model.Task {
	var filtered []*model.Task
	for _, t := range tasks {
		if t.Priority == priority {
			filtered = append(filtered, t)
		}
	}
	return filtered
}
