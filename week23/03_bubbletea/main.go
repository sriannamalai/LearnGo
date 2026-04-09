package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ========================================
// Week 23, Lesson 3: Bubble Tea TUI
// ========================================
// Bubble Tea is a Go framework for building terminal user interfaces
// (TUIs). It's based on The Elm Architecture (TEA):
//   - Model:  Your application state
//   - Update: A function that handles messages and updates the model
//   - View:   A function that renders the model as a string
//
// This pattern makes TUIs predictable and testable. All state changes
// flow through Update, and all rendering flows through View.
//
// We also use Lipgloss for styling — it brings CSS-like styling to
// the terminal with colors, borders, padding, and alignment.
//
// Run this program:
//   go run .
//
// Controls:
//   j/down  — move cursor down
//   k/up    — move cursor up
//   space   — toggle todo item
//   a       — add a new todo
//   d       — delete current item
//   q/esc   — quit
// ========================================

// ========================================
// Lipgloss Styles
// ========================================
// Lipgloss provides CSS-like styling for terminal output.
// Define styles as package-level variables for reuse.

var (
	// Title style: bold, colored, with bottom border
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginBottom(1)

	// Style for the currently selected item
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	// Style for completed items
	completedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Strikethrough(true)

	// Style for the help text at the bottom
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	// Style for the status bar
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#A550DF")).
			Padding(0, 1)

	// Checkbox styles
	checkboxChecked   = lipgloss.NewStyle().Foreground(lipgloss.Color("#73F59F")).Render("[x]")
	checkboxUnchecked = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6F61")).Render("[ ]")

	// Border style for the main content area
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)
)

// ========================================
// Model — Application State
// ========================================
// The model holds ALL application state. In Bubble Tea, the model
// is immutable-by-convention: Update returns a new model rather
// than mutating the existing one (though Go doesn't enforce this).

// todoItem represents a single todo item.
type todoItem struct {
	title     string
	completed bool
}

// model is the Bubble Tea model — it holds the entire application state.
type model struct {
	todos    []todoItem // The list of todo items
	cursor   int        // Which item the cursor is pointing at
	quitting bool       // Whether the user wants to quit
}

// ========================================
// Init — Initial Command
// ========================================
// Init returns an initial command for the application to run.
// It's called once when the program starts. Return nil if there's
// no initial command (most common case).

func (m model) Init() tea.Cmd {
	// No initial command — we just render the initial view.
	// If you needed to fetch data on startup, you'd return a Cmd here.
	// Example: return fetchDataCmd
	return nil
}

// ========================================
// Update — Message Handler
// ========================================
// Update is called when a message (event) occurs. It receives the
// current model and a message, and returns the updated model and
// an optional command. This is the ONLY place state changes happen.
//
// Messages can be:
//   - tea.KeyMsg: keyboard input
//   - tea.WindowSizeMsg: terminal was resized
//   - Custom messages from your own commands

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ========================================
	// Keyboard Input
	// ========================================
	case tea.KeyMsg:
		switch msg.String() {

		// Quit the application
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		// Move cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// Move cursor down
		case "down", "j":
			if m.cursor < len(m.todos)-1 {
				m.cursor++
			}

		// Toggle completion status
		case " ", "enter":
			if len(m.todos) > 0 {
				m.todos[m.cursor].completed = !m.todos[m.cursor].completed
			}

		// Add a new todo item
		case "a":
			newItem := todoItem{
				title:     fmt.Sprintf("New task #%d", len(m.todos)+1),
				completed: false,
			}
			m.todos = append(m.todos, newItem)
			m.cursor = len(m.todos) - 1 // Move cursor to new item

		// Delete current item
		case "d":
			if len(m.todos) > 0 {
				m.todos = append(m.todos[:m.cursor], m.todos[m.cursor+1:]...)
				// Adjust cursor if we deleted the last item
				if m.cursor >= len(m.todos) && m.cursor > 0 {
					m.cursor--
				}
			}
		}
	}

	return m, nil
}

// ========================================
// View — Rendering
// ========================================
// View returns a string that represents the entire UI. Bubble Tea
// calls this after every Update. The string is rendered to the
// terminal. This function should be pure — it should only read
// from the model, never modify it.

func (m model) View() string {
	if m.quitting {
		return "\nGoodbye! Your tasks have been saved (not really, this is a demo).\n\n"
	}

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render(" Todo List - Bubble Tea Demo "))
	b.WriteString("\n\n")

	// Count completed items for the status bar
	completed := 0
	for _, t := range m.todos {
		if t.completed {
			completed++
		}
	}

	// Render each todo item
	if len(m.todos) == 0 {
		b.WriteString("  No tasks yet! Press 'a' to add one.\n")
	} else {
		for i, todo := range m.todos {
			// Cursor indicator
			cursor := "  " // No cursor
			if m.cursor == i {
				cursor = selectedStyle.Render("> ")
			}

			// Checkbox
			checkbox := checkboxUnchecked
			if todo.completed {
				checkbox = checkboxChecked
			}

			// Item text (style differently if completed)
			text := todo.title
			if todo.completed {
				text = completedStyle.Render(text)
			} else if m.cursor == i {
				text = selectedStyle.Render(text)
			}

			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checkbox, text))
		}
	}

	// Status bar
	b.WriteString("\n")
	status := fmt.Sprintf(" %d/%d completed ", completed, len(m.todos))
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n")

	// Help text
	help := "j/k: navigate | space: toggle | a: add | d: delete | q: quit"
	b.WriteString(helpStyle.Render(help))
	b.WriteString("\n")

	// Wrap everything in a border
	return borderStyle.Render(b.String())
}

// ========================================
// Main — Entry Point
// ========================================

func main() {
	// Create the initial model with some sample todos.
	// This is the starting state of the application.
	initialModel := model{
		todos: []todoItem{
			{title: "Learn Bubble Tea basics", completed: true},
			{title: "Understand Model-Update-View pattern", completed: true},
			{title: "Handle keyboard events", completed: false},
			{title: "Style with Lipgloss", completed: false},
			{title: "Build interactive todo list", completed: false},
			{title: "Add colors and borders", completed: false},
			{title: "Create the Week 23 capstone CLI", completed: false},
		},
		cursor: 2, // Start at the first uncompleted item
	}

	// ========================================
	// Create and Start the Program
	// ========================================
	// tea.NewProgram creates a new Bubble Tea program.
	// Options:
	//   tea.WithAltScreen()     — use the alternate screen buffer
	//   tea.WithMouseCellMotion() — enable mouse support
	//
	// AltScreen gives you a full-screen TUI that doesn't pollute
	// the user's scrollback buffer when they quit.
	p := tea.NewProgram(initialModel, tea.WithAltScreen())

	// Run starts the event loop. It blocks until the program exits.
	// The final model is returned so you can inspect the state.
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// After the program exits, we can inspect the final state.
	if m, ok := finalModel.(model); ok {
		completed := 0
		for _, t := range m.todos {
			if t.completed {
				completed++
			}
		}
		fmt.Printf("Final state: %d/%d tasks completed\n", completed, len(m.todos))
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// The Elm Architecture (TEA):
//   1. Model  — holds ALL application state
//   2. Init   — returns initial command (or nil)
//   3. Update — handles messages, returns new model + optional command
//   4. View   — renders model to a string (pure function)
//
// Bubble Tea Messages:
//   - tea.KeyMsg       — keyboard input
//   - tea.WindowSizeMsg — terminal resize
//   - tea.MouseMsg     — mouse events (if enabled)
//   - Custom messages  — from your own Cmds
//
// Lipgloss Styling:
//   - lipgloss.NewStyle()     — create a new style
//   - .Bold(true)             — make text bold
//   - .Foreground(color)      — text color
//   - .Background(color)      — background color
//   - .Padding(top, right)    — add padding
//   - .Border(borderStyle)    — add a border
//   - .Render(text)           — apply style to text
//
// Color Types:
//   - lipgloss.Color("#hex")  — hex colors
//   - lipgloss.Color("201")   — ANSI 256 colors
//   - lipgloss.AdaptiveColor  — light/dark theme support
