package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/sri/learngo/week23/04_cli_app/internal/model"
)

// ========================================
// Interactive Command — Bubble Tea TUI
// ========================================
// Launches a full-screen interactive TUI for managing tasks.
// This combines Bubble Tea's Model-Update-View pattern with
// our task store for a seamless interactive experience.
//
// Usage:
//   tasks interactive
//   tasks tui
//
// Controls:
//   j/down  — move down
//   k/up    — move up
//   space   — toggle task completion
//   a       — add a new task
//   d       — delete current task
//   1/2/3   — set priority (low/med/high)
//   q/esc   — quit and save

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Launch interactive TUI mode",
	Long: `Launch a full-screen terminal UI for managing tasks.

The TUI provides a visual interface with keyboard navigation,
task toggling, and real-time updates.

Controls:
  j/k or arrows  Navigate up and down
  space/enter    Toggle task completion
  a              Add a new task
  d              Delete selected task
  1              Set priority to low
  2              Set priority to medium
  3              Set priority to high
  q/esc          Quit`,

	Aliases: []string{"tui", "ui"},
	Args:    cobra.NoArgs,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Load tasks from the store into the TUI model
		tasks := taskStore.GetAll()

		m := tuiModel{
			tasks:  tasks,
			cursor: 0,
		}

		// Run the Bubble Tea program with alternate screen
		p := tea.NewProgram(m, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("running TUI: %w", err)
		}

		// Display final stats after quitting
		if fm, ok := finalModel.(tuiModel); ok {
			completed := 0
			for _, t := range fm.tasks {
				if t.IsCompleted() {
					completed++
				}
			}
			fmt.Printf("Session ended: %d/%d tasks completed.\n",
				completed, len(fm.tasks))
		}

		return nil
	},
}

// ========================================
// TUI Model
// ========================================

// TUI styles
var (
	tuiTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2)

	tuiSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Bold(true)

	tuiCompletedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Strikethrough(true)

	tuiHighPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6F61")).
				Bold(true)

	tuiMedPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD93D"))

	tuiLowPriorityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6BCB77"))

	tuiHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	tuiStatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#A550DF")).
			Padding(0, 1)

	tuiBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)
)

type tuiModel struct {
	tasks    []*model.Task
	cursor   int
	quitting bool
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tasks)-1 {
				m.cursor++
			}

		case " ", "enter":
			if len(m.tasks) > 0 {
				t := m.tasks[m.cursor]
				if t.IsPending() {
					// Mark done in store and model
					taskStore.MarkDone(t.ID)
					t.Complete()
				}
			}

		case "a":
			// Add a new task
			newTask := model.NewTask(
				fmt.Sprintf("New task #%d", taskStore.Count()+1),
				model.PriorityMedium,
			)
			taskStore.Add(newTask)
			m.tasks = append(m.tasks, newTask)
			m.cursor = len(m.tasks) - 1

		case "d":
			// Delete current task
			if len(m.tasks) > 0 {
				t := m.tasks[m.cursor]
				taskStore.Delete(t.ID)
				m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)
				if m.cursor >= len(m.tasks) && m.cursor > 0 {
					m.cursor--
				}
			}

		case "1":
			// Set priority to low
			if len(m.tasks) > 0 {
				m.tasks[m.cursor].Priority = model.PriorityLow
			}

		case "2":
			// Set priority to medium
			if len(m.tasks) > 0 {
				m.tasks[m.cursor].Priority = model.PriorityMedium
			}

		case "3":
			// Set priority to high
			if len(m.tasks) > 0 {
				m.tasks[m.cursor].Priority = model.PriorityHigh
			}
		}
	}

	return m, nil
}

func (m tuiModel) View() string {
	if m.quitting {
		return "\n  Goodbye! Tasks have been saved.\n\n"
	}

	var b strings.Builder

	// Title
	b.WriteString(tuiTitleStyle.Render(" Task Manager - Interactive Mode "))
	b.WriteString("\n\n")

	// Task list
	if len(m.tasks) == 0 {
		b.WriteString("  No tasks yet! Press 'a' to add one.\n")
	} else {
		completed := 0
		for i, t := range m.tasks {
			// Cursor
			cursor := "  "
			if m.cursor == i {
				cursor = tuiSelectedStyle.Render("> ")
			}

			// Checkbox
			checkbox := "[ ]"
			if t.IsCompleted() {
				checkbox = "[x]"
				completed++
			}

			// Priority badge
			var priority string
			switch t.Priority {
			case model.PriorityHigh:
				priority = tuiHighPriorityStyle.Render("HIGH")
			case model.PriorityMedium:
				priority = tuiMedPriorityStyle.Render("MED ")
			case model.PriorityLow:
				priority = tuiLowPriorityStyle.Render("LOW ")
			}

			// Title
			title := t.Title
			if t.IsCompleted() {
				title = tuiCompletedStyle.Render(title)
			} else if m.cursor == i {
				title = tuiSelectedStyle.Render(title)
			}

			b.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, checkbox, priority, title))
		}

		// Status bar
		b.WriteString("\n")
		status := fmt.Sprintf(" %d/%d completed ", completed, len(m.tasks))
		b.WriteString(tuiStatusStyle.Render(status))
	}

	// Help text
	b.WriteString("\n\n")
	help := "j/k: navigate | space: complete | a: add | d: delete | 1/2/3: priority | q: quit"
	b.WriteString(tuiHelpStyle.Render(help))
	b.WriteString("\n")

	return tuiBorderStyle.Render(b.String())
}

func init() {
	// Silence usage errors for the interactive command
	// since it handles its own error display in the TUI.
	_ = os.Stderr
}
