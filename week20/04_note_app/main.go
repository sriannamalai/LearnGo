package main

// ========================================
// Week 20 — Lesson 4: Mini-Project
// Desktop Note-Taking App
// ========================================
// This project combines everything from Week 20:
//   - Fyne app setup and window management
//   - Widgets: Button, Entry, List, Label, Toolbar
//   - Layouts: Border, VBox, HBox, Split
//   - Data binding and callbacks
//
// Features:
//   - Create, edit, and delete notes
//   - List notes in a sidebar
//   - Text editor area for note content
//   - Save notes to a JSON file
//   - Search/filter notes by title or content
//
// Run:
//   go run .

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// Note represents a single note with a title, content,
// and timestamps for creation and last modification.
type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// NoteStore manages the collection of notes and handles
// persistence to a JSON file on disk.
type NoteStore struct {
	Notes    []Note `json:"notes"`
	FilePath string `json:"-"`
}

// NewNoteStore creates a new note store that saves to the given file path.
func NewNoteStore(filePath string) *NoteStore {
	return &NoteStore{
		Notes:    []Note{},
		FilePath: filePath,
	}
}

// Load reads notes from the JSON file on disk.
// If the file doesn't exist, it starts with an empty list.
func (ns *NoteStore) Load() error {
	data, err := os.ReadFile(ns.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet — that's fine, start fresh
			return nil
		}
		return fmt.Errorf("failed to read notes file: %w", err)
	}

	if err := json.Unmarshal(data, &ns.Notes); err != nil {
		return fmt.Errorf("failed to parse notes file: %w", err)
	}

	return nil
}

// Save writes all notes to the JSON file on disk.
func (ns *NoteStore) Save() error {
	// Ensure the directory exists
	dir := filepath.Dir(ns.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(ns.Notes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize notes: %w", err)
	}

	if err := os.WriteFile(ns.FilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write notes file: %w", err)
	}

	return nil
}

// Add creates a new note and appends it to the store.
func (ns *NoteStore) Add(title, content string) *Note {
	now := time.Now().Format(time.RFC3339)
	note := Note{
		ID:        fmt.Sprintf("note_%d", time.Now().UnixNano()),
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	ns.Notes = append(ns.Notes, note)
	return &ns.Notes[len(ns.Notes)-1]
}

// Update modifies an existing note's title and content.
func (ns *NoteStore) Update(id, title, content string) {
	for i := range ns.Notes {
		if ns.Notes[i].ID == id {
			ns.Notes[i].Title = title
			ns.Notes[i].Content = content
			ns.Notes[i].UpdatedAt = time.Now().Format(time.RFC3339)
			return
		}
	}
}

// Delete removes a note by its ID.
func (ns *NoteStore) Delete(id string) {
	for i, note := range ns.Notes {
		if note.ID == id {
			ns.Notes = append(ns.Notes[:i], ns.Notes[i+1:]...)
			return
		}
	}
}

// Search returns notes whose title or content contains the query.
// An empty query returns all notes.
func (ns *NoteStore) Search(query string) []Note {
	if query == "" {
		return ns.Notes
	}

	query = strings.ToLower(query)
	var results []Note
	for _, note := range ns.Notes {
		if strings.Contains(strings.ToLower(note.Title), query) ||
			strings.Contains(strings.ToLower(note.Content), query) {
			results = append(results, note)
		}
	}
	return results
}

// ========================================
// Application
// ========================================

// NoteApp holds all the UI components and state for the
// note-taking application.
type NoteApp struct {
	app    fyne.App
	window fyne.Window
	store  *NoteStore

	// UI components
	noteList      *widget.List
	titleEntry    *widget.Entry
	contentEntry  *widget.MultiLineEntry
	searchEntry   *widget.Entry
	statusLabel   *widget.Label
	deleteButton  *widget.Button
	saveButton    *widget.Button

	// State
	selectedIndex int
	filteredNotes []Note
}

// NewNoteApp creates and initializes the note-taking application.
func NewNoteApp() *NoteApp {
	na := &NoteApp{
		selectedIndex: -1,
	}

	// Determine save file location
	// Use the user's home directory for storing notes
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	notesPath := filepath.Join(homeDir, ".fyne-notes", "notes.json")

	na.store = NewNoteStore(notesPath)
	if err := na.store.Load(); err != nil {
		fmt.Printf("Warning: Could not load notes: %v\n", err)
	}
	na.filteredNotes = na.store.Notes

	// Create the Fyne app and window
	na.app = app.New()
	na.window = na.app.NewWindow("Note App")
	na.window.Resize(fyne.NewSize(800, 600))
	na.window.CenterOnScreen()

	// Build the UI
	na.buildUI()

	// Add sample notes if the store is empty
	if len(na.store.Notes) == 0 {
		na.addSampleNotes()
	}

	return na
}

// buildUI constructs the entire user interface.
func (na *NoteApp) buildUI() {
	// ========================================
	// Status Bar
	// ========================================
	na.statusLabel = widget.NewLabel("Ready")

	// ========================================
	// Search Bar
	// ========================================
	na.searchEntry = widget.NewEntry()
	na.searchEntry.SetPlaceHolder("Search notes...")
	na.searchEntry.OnChanged = func(query string) {
		na.filteredNotes = na.store.Search(query)
		na.selectedIndex = -1
		na.noteList.UnselectAll()
		na.noteList.Refresh()
		na.clearEditor()
		na.statusLabel.SetText(fmt.Sprintf("Found %d notes", len(na.filteredNotes)))
	}

	// ========================================
	// Note List (Sidebar)
	// ========================================
	na.noteList = widget.NewList(
		// Length: number of filtered notes
		func() int {
			return len(na.filteredNotes)
		},
		// CreateItem: template for each list item
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabelWithStyle("Title", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Updated: ..."),
			)
		},
		// UpdateItem: populate each list item
		func(index widget.ListItemID, item fyne.CanvasObject) {
			box := item.(*fyne.Container)
			titleLabel := box.Objects[0].(*widget.RichText)
			dateLabel := box.Objects[1].(*widget.Label)

			note := na.filteredNotes[index]
			title := note.Title
			if title == "" {
				title = "(Untitled)"
			}
			// Truncate long titles
			if len(title) > 30 {
				title = title[:27] + "..."
			}
			titleLabel.ParseMarkdown("**" + title + "**")
			dateLabel.SetText(formatTime(note.UpdatedAt))
		},
	)
	na.noteList.OnSelected = func(index widget.ListItemID) {
		na.selectedIndex = index
		if index >= 0 && index < len(na.filteredNotes) {
			note := na.filteredNotes[index]
			na.titleEntry.SetText(note.Title)
			na.contentEntry.SetText(note.Content)
			na.deleteButton.Enable()
			na.saveButton.Enable()
			na.statusLabel.SetText(fmt.Sprintf("Editing: %s", note.Title))
		}
	}

	// ========================================
	// Editor Area
	// ========================================
	na.titleEntry = widget.NewEntry()
	na.titleEntry.SetPlaceHolder("Note title...")

	na.contentEntry = widget.NewMultiLineEntry()
	na.contentEntry.SetPlaceHolder("Start typing your note here...")
	na.contentEntry.SetMinRowsVisible(15)

	// ========================================
	// Action Buttons
	// ========================================
	newButton := widget.NewButtonWithIcon("New Note", theme.DocumentCreateIcon(), func() {
		na.newNote()
	})
	newButton.Importance = widget.HighImportance

	na.saveButton = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {
		na.saveCurrentNote()
	})

	na.deleteButton = widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		na.deleteCurrentNote()
	})
	na.deleteButton.Importance = widget.DangerImportance
	na.deleteButton.Disable()

	// ========================================
	// Sidebar Layout
	// ========================================
	sidebar := container.NewBorder(
		// Top: search bar and new button
		container.NewVBox(
			widget.NewLabelWithStyle("Notes", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			na.searchEntry,
			newButton,
			widget.NewSeparator(),
		),
		// Bottom: note count
		widget.NewLabel(fmt.Sprintf("%d notes", len(na.store.Notes))),
		nil, nil,
		// Center: scrollable note list
		na.noteList,
	)

	// ========================================
	// Editor Layout
	// ========================================
	editorToolbar := container.NewHBox(
		na.saveButton,
		na.deleteButton,
		layout.NewSpacer(),
	)

	editor := container.NewBorder(
		// Top: title entry and toolbar
		container.NewVBox(
			widget.NewLabelWithStyle("Editor", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Title:"),
			na.titleEntry,
			editorToolbar,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		// Center: content area (fills remaining space)
		na.contentEntry,
	)

	// ========================================
	// Main Layout: Split Sidebar and Editor
	// ========================================
	// HSplit creates a horizontal split with a draggable divider.
	split := container.NewHSplit(sidebar, editor)
	split.SetOffset(0.3) // Sidebar takes 30% of the width

	// ========================================
	// Window Layout with Status Bar
	// ========================================
	mainLayout := container.NewBorder(
		nil,             // top
		na.statusLabel,  // bottom (status bar)
		nil, nil,
		split,           // center
	)

	na.window.SetContent(mainLayout)

	// ========================================
	// Keyboard Shortcut
	// ========================================
	// Ctrl+S to save, Ctrl+N for new note
	na.window.Canvas().AddShortcut(
		&fyne.ShortcutCopy{},
		func(shortcut fyne.Shortcut) {}, // Default handling
	)
}

// ========================================
// Note Operations
// ========================================

// newNote creates a new empty note and selects it in the list.
func (na *NoteApp) newNote() {
	note := na.store.Add("New Note", "")
	na.filteredNotes = na.store.Search(na.searchEntry.Text)
	na.noteList.Refresh()

	// Select the new note
	for i, n := range na.filteredNotes {
		if n.ID == note.ID {
			na.noteList.Select(i)
			break
		}
	}

	// Focus the title entry for immediate editing
	na.window.Canvas().Focus(na.titleEntry)
	na.titleEntry.SetText("New Note")

	// Auto-save
	if err := na.store.Save(); err != nil {
		na.statusLabel.SetText(fmt.Sprintf("Error saving: %v", err))
	} else {
		na.statusLabel.SetText("New note created")
	}
}

// saveCurrentNote saves the currently selected note with the
// editor's title and content.
func (na *NoteApp) saveCurrentNote() {
	if na.selectedIndex < 0 || na.selectedIndex >= len(na.filteredNotes) {
		na.statusLabel.SetText("No note selected")
		return
	}

	note := na.filteredNotes[na.selectedIndex]
	na.store.Update(note.ID, na.titleEntry.Text, na.contentEntry.Text)

	// Refresh the list to show updated title
	na.filteredNotes = na.store.Search(na.searchEntry.Text)
	na.noteList.Refresh()

	// Persist to disk
	if err := na.store.Save(); err != nil {
		na.statusLabel.SetText(fmt.Sprintf("Error saving: %v", err))
		return
	}

	na.statusLabel.SetText(fmt.Sprintf("Saved: %s", na.titleEntry.Text))
}

// deleteCurrentNote removes the currently selected note after
// showing a confirmation dialog.
func (na *NoteApp) deleteCurrentNote() {
	if na.selectedIndex < 0 || na.selectedIndex >= len(na.filteredNotes) {
		return
	}

	note := na.filteredNotes[na.selectedIndex]
	title := note.Title
	if title == "" {
		title = "(Untitled)"
	}

	// Show confirmation dialog
	dialog.ShowConfirm(
		"Delete Note",
		fmt.Sprintf("Are you sure you want to delete \"%s\"?", title),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			na.store.Delete(note.ID)
			na.filteredNotes = na.store.Search(na.searchEntry.Text)
			na.noteList.Refresh()
			na.clearEditor()
			na.selectedIndex = -1
			na.noteList.UnselectAll()

			if err := na.store.Save(); err != nil {
				na.statusLabel.SetText(fmt.Sprintf("Error saving: %v", err))
				return
			}

			na.statusLabel.SetText(fmt.Sprintf("Deleted: %s", title))
		},
		na.window,
	)
}

// clearEditor resets the editor fields to empty.
func (na *NoteApp) clearEditor() {
	na.titleEntry.SetText("")
	na.contentEntry.SetText("")
	na.deleteButton.Disable()
}

// addSampleNotes populates the store with a few example notes
// so the app doesn't start completely empty.
func (na *NoteApp) addSampleNotes() {
	na.store.Add("Welcome to Note App!",
		"This is your new note-taking application built with Fyne.\n\n"+
			"Features:\n"+
			"- Create new notes with the 'New Note' button\n"+
			"- Edit notes by selecting them from the sidebar\n"+
			"- Save changes with the 'Save' button\n"+
			"- Delete notes you no longer need\n"+
			"- Search notes by title or content\n\n"+
			"Your notes are automatically saved to a JSON file.")

	na.store.Add("Go + Fyne Tips",
		"Building desktop apps with Go and Fyne:\n\n"+
			"1. Use layouts (VBox, HBox, Border) for structure\n"+
			"2. Data binding keeps UI and data in sync\n"+
			"3. Fyne handles cross-platform differences\n"+
			"4. Use goroutines for long-running tasks\n"+
			"5. The canvas package provides custom drawing")

	na.store.Add("Todo List",
		"- [ ] Learn more Fyne widgets\n"+
			"- [ ] Build a custom theme\n"+
			"- [ ] Add file import/export\n"+
			"- [ ] Try Fyne on mobile\n"+
			"- [x] Create the note app")

	na.filteredNotes = na.store.Notes
	_ = na.store.Save()
}

// ========================================
// Run starts the application event loop.
// ========================================
func (na *NoteApp) Run() {
	na.window.ShowAndRun()
}

// ========================================
// Helpers
// ========================================

// formatTime converts an RFC3339 timestamp to a short display format.
func formatTime(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339
	}

	now := time.Now()
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return "Today " + t.Format("3:04 PM")
	}
	return t.Format("Jan 2, 2006")
}

// ========================================
// Main
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 20 - Mini-Project: Note App")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Starting desktop note-taking application...")
	fmt.Println("Notes are saved to: ~/.fyne-notes/notes.json")
	fmt.Println()

	noteApp := NewNoteApp()
	noteApp.Run()

	fmt.Println("Note App closed. Goodbye!")
}
