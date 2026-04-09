package main

// ========================================
// Week 22 — Lesson 4: Mini-Project
// Cross-Platform Mobile Task Manager
// ========================================
// This project combines everything from Week 22:
//   - Fyne mobile-friendly UI patterns
//   - Responsive layouts for phone screens
//   - Persistent storage to JSON file
//   - Touch-friendly widgets and navigation
//
// Features:
//   - Add, edit, and delete tasks
//   - Mark tasks as complete
//   - Categories: Personal, Work, Shopping, Health, Other
//   - Priorities: Low, Medium, High
//   - Persistent storage to JSON file
//   - Responsive layout for phone screens
//   - Bottom tab navigation (mobile pattern)
//   - Task filtering by category and status
//
// Build commands:
//   Desktop:  go run .
//   iOS:      fyne package -os ios -appID com.example.taskmanager
//   Android:  fyne package -os android -appID com.example.taskmanager
//
// Run:
//   go run .

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ========================================
// Data Model
// ========================================

// Task represents a single task with all its attributes.
type Task struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Notes     string `json:"notes"`
	Category  string `json:"category"`
	Priority  string `json:"priority"` // "low", "medium", "high"
	Completed bool   `json:"completed"`
	CreatedAt string `json:"createdAt"`
	DueDate   string `json:"dueDate,omitempty"`
}

// TaskStore manages the collection of tasks with persistence.
type TaskStore struct {
	Tasks    []Task `json:"tasks"`
	FilePath string `json:"-"`
	nextID   int
}

// NewTaskStore creates a new task store.
func NewTaskStore(filePath string) *TaskStore {
	ts := &TaskStore{
		Tasks:    []Task{},
		FilePath: filePath,
		nextID:   1,
	}
	ts.Load()
	return ts
}

// Load reads tasks from disk.
func (ts *TaskStore) Load() error {
	data, err := os.ReadFile(ts.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &ts.Tasks); err != nil {
		return err
	}

	// Find the highest ID to continue numbering
	for _, t := range ts.Tasks {
		var id int
		fmt.Sscanf(t.ID, "task_%d", &id)
		if id >= ts.nextID {
			ts.nextID = id + 1
		}
	}

	return nil
}

// Save writes tasks to disk.
func (ts *TaskStore) Save() error {
	dir := filepath.Dir(ts.FilePath)
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(ts.Tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ts.FilePath, data, 0644)
}

// Add creates a new task.
func (ts *TaskStore) Add(title, notes, category, priority string) *Task {
	task := Task{
		ID:        fmt.Sprintf("task_%d", ts.nextID),
		Title:     title,
		Notes:     notes,
		Category:  category,
		Priority:  priority,
		Completed: false,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	ts.Tasks = append(ts.Tasks, task)
	ts.nextID++
	return &ts.Tasks[len(ts.Tasks)-1]
}

// Update modifies an existing task.
func (ts *TaskStore) Update(id, title, notes, category, priority string) {
	for i := range ts.Tasks {
		if ts.Tasks[i].ID == id {
			ts.Tasks[i].Title = title
			ts.Tasks[i].Notes = notes
			ts.Tasks[i].Category = category
			ts.Tasks[i].Priority = priority
			return
		}
	}
}

// ToggleComplete toggles the completed status of a task.
func (ts *TaskStore) ToggleComplete(id string) {
	for i := range ts.Tasks {
		if ts.Tasks[i].ID == id {
			ts.Tasks[i].Completed = !ts.Tasks[i].Completed
			return
		}
	}
}

// Delete removes a task by ID.
func (ts *TaskStore) Delete(id string) {
	for i, t := range ts.Tasks {
		if t.ID == id {
			ts.Tasks = append(ts.Tasks[:i], ts.Tasks[i+1:]...)
			return
		}
	}
}

// Filter returns tasks matching the given criteria.
func (ts *TaskStore) Filter(category, status string) []Task {
	var result []Task

	for _, t := range ts.Tasks {
		// Category filter
		if category != "All" && category != "" && t.Category != category {
			continue
		}
		// Status filter
		if status == "Active" && t.Completed {
			continue
		}
		if status == "Completed" && !t.Completed {
			continue
		}
		result = append(result, t)
	}

	// Sort: incomplete first, then by priority, then by date
	sort.Slice(result, func(i, j int) bool {
		// Incomplete tasks first
		if result[i].Completed != result[j].Completed {
			return !result[i].Completed
		}
		// High priority first
		pi := priorityWeight(result[i].Priority)
		pj := priorityWeight(result[j].Priority)
		if pi != pj {
			return pi > pj
		}
		// Newer first
		return result[i].CreatedAt > result[j].CreatedAt
	})

	return result
}

// Stats returns task statistics.
func (ts *TaskStore) Stats() map[string]int {
	stats := map[string]int{
		"total":     len(ts.Tasks),
		"completed": 0,
		"active":    0,
		"high":      0,
	}

	for _, t := range ts.Tasks {
		if t.Completed {
			stats["completed"]++
		} else {
			stats["active"]++
		}
		if t.Priority == "high" && !t.Completed {
			stats["high"]++
		}
	}

	return stats
}

func priorityWeight(priority string) int {
	switch priority {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// ========================================
// Application
// ========================================

var (
	categories = []string{"Personal", "Work", "Shopping", "Health", "Other"}
	priorities = []string{"Low", "Medium", "High"}
)

// TaskApp holds the application state and UI components.
type TaskApp struct {
	app    fyne.App
	window fyne.Window
	store  *TaskStore

	// UI state
	currentFilter   string // category filter
	currentStatus   string // "All", "Active", "Completed"
	filteredTasks   []Task
	selectedTaskID  string

	// UI components
	taskList        *widget.List
	statusLabel     *widget.Label
	statsLabel      *widget.Label
	filterSelect    *widget.Select
	statusSelect    *widget.Select
}

// NewTaskApp creates and initializes the task manager app.
func NewTaskApp() *TaskApp {
	ta := &TaskApp{
		currentFilter: "All",
		currentStatus: "All",
	}

	// Set up storage path
	homeDir, _ := os.UserHomeDir()
	storePath := filepath.Join(homeDir, ".task-manager", "tasks.json")
	ta.store = NewTaskStore(storePath)

	// Add sample tasks if empty
	if len(ta.store.Tasks) == 0 {
		ta.addSampleTasks()
	}
	ta.refreshFilter()

	// Create Fyne app
	ta.app = app.NewWithID("com.example.taskmanager")
	ta.window = ta.app.NewWindow("Task Manager")
	ta.window.Resize(fyne.NewSize(380, 680)) // Phone-like size
	ta.window.CenterOnScreen()

	ta.buildUI()
	return ta
}

// ========================================
// UI Construction
// ========================================

func (ta *TaskApp) buildUI() {
	// ========================================
	// Status Bar
	// ========================================
	ta.statusLabel = widget.NewLabel("Ready")
	ta.statsLabel = widget.NewLabel("")
	ta.updateStats()

	// ========================================
	// Task List
	// ========================================
	ta.taskList = widget.NewList(
		func() int { return len(ta.filteredTasks) },
		func() fyne.CanvasObject {
			// Template: checkbox + title + priority badge
			check := widget.NewCheck("", nil)
			title := widget.NewLabel("Task title")
			title.Wrapping = fyne.TextWrapWord
			category := widget.NewLabel("Category")
			category.TextStyle = fyne.TextStyle{Italic: true}
			priorityLabel := widget.NewLabel("[P]")
			priorityLabel.TextStyle = fyne.TextStyle{Bold: true}

			return container.NewBorder(
				nil, nil,
				check,
				priorityLabel,
				container.NewVBox(title, category),
			)
		},
		func(index widget.ListItemID, item fyne.CanvasObject) {
			if index >= len(ta.filteredTasks) {
				return
			}
			task := ta.filteredTasks[index]

			border := item.(*fyne.Container)
			check := border.Objects[1].(*widget.Check)
			priorityLabel := border.Objects[2].(*widget.Label)
			center := border.Objects[0].(*fyne.Container)
			titleLabel := center.Objects[0].(*widget.Label)
			categoryLabel := center.Objects[1].(*widget.Label)

			// Update checkbox
			check.SetChecked(task.Completed)
			check.OnChanged = func(checked bool) {
				ta.store.ToggleComplete(task.ID)
				ta.store.Save()
				ta.refreshFilter()
				ta.taskList.Refresh()
				ta.updateStats()
			}

			// Update title
			title := task.Title
			if task.Completed {
				title = "~" + title + "~" // Strikethrough indicator
			}
			titleLabel.SetText(title)

			// Update category
			categoryLabel.SetText(task.Category)

			// Update priority indicator
			switch task.Priority {
			case "high":
				priorityLabel.SetText("!!!")
			case "medium":
				priorityLabel.SetText("!!")
			default:
				priorityLabel.SetText("!")
			}
		},
	)
	ta.taskList.OnSelected = func(id widget.ListItemID) {
		if id < len(ta.filteredTasks) {
			ta.selectedTaskID = ta.filteredTasks[id].ID
			ta.showTaskDetail(ta.filteredTasks[id])
		}
		ta.taskList.UnselectAll()
	}

	// ========================================
	// Filter Controls
	// ========================================
	filterCategories := append([]string{"All"}, categories...)
	ta.filterSelect = widget.NewSelect(filterCategories, func(selected string) {
		ta.currentFilter = selected
		ta.refreshFilter()
		ta.taskList.Refresh()
		ta.updateStats()
	})
	ta.filterSelect.SetSelected("All")

	ta.statusSelect = widget.NewSelect([]string{"All", "Active", "Completed"}, func(selected string) {
		ta.currentStatus = selected
		ta.refreshFilter()
		ta.taskList.Refresh()
		ta.updateStats()
	})
	ta.statusSelect.SetSelected("All")

	filterBar := container.NewGridWithColumns(2,
		ta.filterSelect,
		ta.statusSelect,
	)

	// ========================================
	// Add Task Button
	// ========================================
	addButton := widget.NewButtonWithIcon("New Task", theme.ContentAddIcon(), func() {
		ta.showAddTaskDialog()
	})
	addButton.Importance = widget.HighImportance

	// ========================================
	// Home Tab — Task List
	// ========================================
	homeContent := container.NewBorder(
		container.NewVBox(
			ta.statsLabel,
			filterBar,
			addButton,
			widget.NewSeparator(),
		),
		ta.statusLabel,
		nil, nil,
		ta.taskList,
	)

	homeTab := container.NewTabItemWithIcon("Tasks", theme.ListIcon(), homeContent)

	// ========================================
	// Summary Tab — Statistics
	// ========================================
	summaryTab := container.NewTabItemWithIcon("Summary", theme.InfoIcon(),
		ta.buildSummaryTab(),
	)

	// ========================================
	// Settings Tab
	// ========================================
	settingsTab := container.NewTabItemWithIcon("Settings", theme.SettingsIcon(),
		ta.buildSettingsTab(),
	)

	// ========================================
	// Tab Navigation (bottom tabs for mobile)
	// ========================================
	tabs := container.NewAppTabs(homeTab, summaryTab, settingsTab)
	tabs.SetTabLocation(container.TabLocationBottom)

	ta.window.SetContent(tabs)
}

// ========================================
// Summary Tab
// ========================================

func (ta *TaskApp) buildSummaryTab() fyne.CanvasObject {
	stats := ta.store.Stats()

	totalLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Total Tasks: %d", stats["total"]),
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	activeCard := widget.NewCard("Active", "",
		widget.NewLabelWithStyle(
			fmt.Sprintf("%d tasks remaining", stats["active"]),
			fyne.TextAlignCenter,
			fyne.TextStyle{},
		),
	)

	completedCard := widget.NewCard("Completed", "",
		widget.NewLabelWithStyle(
			fmt.Sprintf("%d tasks done", stats["completed"]),
			fyne.TextAlignCenter,
			fyne.TextStyle{},
		),
	)

	highPriorityCard := widget.NewCard("High Priority", "",
		widget.NewLabelWithStyle(
			fmt.Sprintf("%d urgent tasks", stats["high"]),
			fyne.TextAlignCenter,
			fyne.TextStyle{},
		),
	)

	// Category breakdown
	categoryBreakdown := container.NewVBox(
		widget.NewLabelWithStyle("By Category", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
	)

	for _, cat := range categories {
		count := 0
		for _, t := range ta.store.Tasks {
			if t.Category == cat {
				count++
			}
		}
		if count > 0 {
			categoryBreakdown.Add(
				widget.NewLabel(fmt.Sprintf("  %s: %d tasks", cat, count)),
			)
		}
	}

	// Completion rate
	completionRate := 0.0
	if stats["total"] > 0 {
		completionRate = float64(stats["completed"]) / float64(stats["total"]) * 100
	}

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(completionRate / 100.0)

	return container.NewScroll(container.NewVBox(
		widget.NewLabelWithStyle("Task Summary", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		totalLabel,
		container.NewGridWithColumns(2,
			activeCard,
			completedCard,
		),
		highPriorityCard,
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Completion Rate: %.0f%%", completionRate)),
		progressBar,
		widget.NewSeparator(),
		categoryBreakdown,
	))
}

// ========================================
// Settings Tab
// ========================================

func (ta *TaskApp) buildSettingsTab() fyne.CanvasObject {
	// Clear completed tasks
	clearCompletedBtn := widget.NewButton("Clear Completed Tasks", func() {
		dialog.ShowConfirm("Clear Completed",
			"Remove all completed tasks?",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				var active []Task
				for _, t := range ta.store.Tasks {
					if !t.Completed {
						active = append(active, t)
					}
				}
				ta.store.Tasks = active
				ta.store.Save()
				ta.refreshFilter()
				ta.taskList.Refresh()
				ta.updateStats()
				ta.statusLabel.SetText("Cleared completed tasks")
			},
			ta.window,
		)
	})

	// Delete all tasks
	deleteAllBtn := widget.NewButton("Delete All Tasks", func() {
		dialog.ShowConfirm("Delete All",
			"This will permanently delete ALL tasks. Are you sure?",
			func(confirmed bool) {
				if !confirmed {
					return
				}
				ta.store.Tasks = []Task{}
				ta.store.Save()
				ta.refreshFilter()
				ta.taskList.Refresh()
				ta.updateStats()
				ta.statusLabel.SetText("All tasks deleted")
			},
			ta.window,
		)
	})
	deleteAllBtn.Importance = widget.DangerImportance

	// Add sample tasks
	sampleBtn := widget.NewButton("Load Sample Tasks", func() {
		ta.addSampleTasks()
		ta.store.Save()
		ta.refreshFilter()
		ta.taskList.Refresh()
		ta.updateStats()
		ta.statusLabel.SetText("Sample tasks loaded")
	})

	// Storage info
	storageLabel := widget.NewLabel(
		fmt.Sprintf("Storage: %s", ta.store.FilePath),
	)
	storageLabel.Wrapping = fyne.TextWrapWord

	return container.NewScroll(container.NewVBox(
		widget.NewLabelWithStyle("Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),

		widget.NewCard("Data Management", "", container.NewVBox(
			clearCompletedBtn,
			sampleBtn,
			widget.NewSeparator(),
			deleteAllBtn,
		)),

		widget.NewCard("Storage", "", container.NewVBox(
			storageLabel,
			widget.NewLabel(fmt.Sprintf("Tasks: %d", len(ta.store.Tasks))),
		)),

		widget.NewCard("About", "", container.NewVBox(
			widget.NewLabel("Task Manager v1.0"),
			widget.NewLabel("Built with Go + Fyne"),
			widget.NewLabel("Week 22 Mini-Project"),
		)),
	))
}

// ========================================
// Dialogs
// ========================================

// showAddTaskDialog shows a form dialog to create a new task.
func (ta *TaskApp) showAddTaskDialog() {
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Task title...")

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetPlaceHolder("Notes (optional)...")
	notesEntry.SetMinRowsVisible(3)

	categorySelect := widget.NewSelect(categories, nil)
	categorySelect.SetSelected("Personal")

	prioritySelect := widget.NewSelect(priorities, nil)
	prioritySelect.SetSelected("Medium")

	form := dialog.NewForm(
		"New Task",
		"Add",
		"Cancel",
		[]*widget.FormItem{
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Notes", notesEntry),
			widget.NewFormItem("Category", categorySelect),
			widget.NewFormItem("Priority", prioritySelect),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}

			title := strings.TrimSpace(titleEntry.Text)
			if title == "" {
				ta.statusLabel.SetText("Error: title is required")
				return
			}

			ta.store.Add(
				title,
				notesEntry.Text,
				categorySelect.Selected,
				strings.ToLower(prioritySelect.Selected),
			)
			ta.store.Save()
			ta.refreshFilter()
			ta.taskList.Refresh()
			ta.updateStats()
			ta.statusLabel.SetText(fmt.Sprintf("Added: %s", title))
		},
		ta.window,
	)
	form.Resize(fyne.NewSize(350, 400))
	form.Show()
}

// showTaskDetail shows task details with edit and delete options.
func (ta *TaskApp) showTaskDetail(task Task) {
	titleEntry := widget.NewEntry()
	titleEntry.SetText(task.Title)

	notesEntry := widget.NewMultiLineEntry()
	notesEntry.SetText(task.Notes)
	notesEntry.SetMinRowsVisible(3)

	categorySelect := widget.NewSelect(categories, nil)
	categorySelect.SetSelected(task.Category)

	prioritySelect := widget.NewSelect(priorities, nil)
	prioritySelect.SetSelected(strings.Title(task.Priority))

	// Status display
	status := "Active"
	if task.Completed {
		status = "Completed"
	}
	statusLabel := widget.NewLabel(fmt.Sprintf("Status: %s", status))

	createdLabel := widget.NewLabel(
		fmt.Sprintf("Created: %s", formatTaskTime(task.CreatedAt)),
	)

	// Delete button inside the dialog
	deleteBtn := widget.NewButton("Delete Task", func() {
		dialog.ShowConfirm("Delete Task",
			fmt.Sprintf("Delete \"%s\"?", task.Title),
			func(confirmed bool) {
				if confirmed {
					ta.store.Delete(task.ID)
					ta.store.Save()
					ta.refreshFilter()
					ta.taskList.Refresh()
					ta.updateStats()
					ta.statusLabel.SetText(fmt.Sprintf("Deleted: %s", task.Title))
				}
			},
			ta.window,
		)
	})
	deleteBtn.Importance = widget.DangerImportance

	form := dialog.NewForm(
		"Task Details",
		"Save",
		"Close",
		[]*widget.FormItem{
			widget.NewFormItem("Title", titleEntry),
			widget.NewFormItem("Notes", notesEntry),
			widget.NewFormItem("Category", categorySelect),
			widget.NewFormItem("Priority", prioritySelect),
			widget.NewFormItem("", statusLabel),
			widget.NewFormItem("", createdLabel),
			widget.NewFormItem("", deleteBtn),
		},
		func(confirmed bool) {
			if !confirmed {
				return
			}

			title := strings.TrimSpace(titleEntry.Text)
			if title == "" {
				return
			}

			ta.store.Update(
				task.ID,
				title,
				notesEntry.Text,
				categorySelect.Selected,
				strings.ToLower(prioritySelect.Selected),
			)
			ta.store.Save()
			ta.refreshFilter()
			ta.taskList.Refresh()
			ta.updateStats()
			ta.statusLabel.SetText(fmt.Sprintf("Updated: %s", title))
		},
		ta.window,
	)
	form.Resize(fyne.NewSize(350, 500))
	form.Show()
}

// ========================================
// State Management
// ========================================

func (ta *TaskApp) refreshFilter() {
	ta.filteredTasks = ta.store.Filter(ta.currentFilter, ta.currentStatus)
}

func (ta *TaskApp) updateStats() {
	stats := ta.store.Stats()
	ta.statsLabel.SetText(fmt.Sprintf(
		"%d active  |  %d completed  |  %d high priority",
		stats["active"], stats["completed"], stats["high"],
	))
}

// ========================================
// Sample Data
// ========================================

func (ta *TaskApp) addSampleTasks() {
	samples := []struct {
		title    string
		notes    string
		category string
		priority string
	}{
		{"Buy groceries", "Milk, eggs, bread, vegetables", "Shopping", "high"},
		{"Finish Go project", "Complete Week 22 mini-project", "Work", "high"},
		{"Morning run", "30 minutes in the park", "Health", "medium"},
		{"Read Go book", "Chapters 10-12", "Personal", "medium"},
		{"Team meeting notes", "Prepare agenda for Monday", "Work", "high"},
		{"Call dentist", "Schedule annual checkup", "Health", "low"},
		{"Update resume", "Add recent Go projects", "Personal", "medium"},
		{"Fix kitchen light", "Replace LED bulb", "Personal", "low"},
		{"Order vitamins", "Vitamin D and Omega-3", "Shopping", "low"},
		{"Code review", "Review PRs from teammates", "Work", "medium"},
	}

	for _, s := range samples {
		ta.store.Add(s.title, s.notes, s.category, s.priority)
	}
}

// ========================================
// Helpers
// ========================================

func formatTaskTime(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}

// ========================================
// Run
// ========================================

func (ta *TaskApp) Run() {
	ta.window.ShowAndRun()
}

// ========================================
// Main
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 22 - Mini-Project: Task Manager")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Cross-platform mobile task manager")
	fmt.Println("Built with Go + Fyne")
	fmt.Println()
	fmt.Println("Build for mobile:")
	fmt.Println("  iOS:     fyne package -os ios -appID com.example.taskmanager")
	fmt.Println("  Android: fyne package -os android -appID com.example.taskmanager")
	fmt.Println()

	taskApp := NewTaskApp()
	taskApp.Run()

	fmt.Println("Task Manager closed. Goodbye!")
}
