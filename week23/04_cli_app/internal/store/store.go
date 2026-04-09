package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/sri/learngo/week23/04_cli_app/internal/model"
)

// ========================================
// Task Store — JSON File Persistence
// ========================================
// The store manages CRUD operations for tasks, persisting them
// to a JSON file on disk. It uses a mutex for thread safety
// (important when the TUI and CLI might access concurrently)
// and maintains an auto-incrementing ID counter.
//
// Architecture Decision: We use a simple JSON file rather than
// a database because:
//   1. Zero setup — works immediately after install
//   2. Human-readable — users can inspect/edit the file
//   3. Portable — easy to backup, sync, or move
//   4. Appropriate for a CLI tool's typical data volume

// Store manages task persistence using a JSON file.
type Store struct {
	filePath string
	mu       sync.RWMutex
	data     storeData
}

// storeData is the structure persisted to the JSON file.
type storeData struct {
	NextID int           `json:"next_id"`
	Tasks  []*model.Task `json:"tasks"`
}

// New creates a new Store backed by the given file path.
// If the file exists, it loads existing data. If not, it
// creates a new empty store.
func New(filePath string) (*Store, error) {
	s := &Store{
		filePath: filePath,
		data: storeData{
			NextID: 1,
			Tasks:  make([]*model.Task, 0),
		},
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating store directory: %w", err)
	}

	// Load existing data if the file exists
	if _, err := os.Stat(filePath); err == nil {
		if err := s.load(); err != nil {
			return nil, fmt.Errorf("loading store: %w", err)
		}
	}

	return s, nil
}

// ========================================
// CRUD Operations
// ========================================

// Add creates a new task and persists it to the store.
// It assigns an auto-incrementing ID to the task.
func (s *Store) Add(task *model.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.ID = s.data.NextID
	s.data.NextID++
	s.data.Tasks = append(s.data.Tasks, task)

	return s.save()
}

// GetAll returns all tasks in the store, sorted by ID.
func (s *Store) GetAll() []*model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external mutation
	result := make([]*model.Task, len(s.data.Tasks))
	copy(result, s.data.Tasks)

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result
}

// GetByID returns a single task by its ID.
func (s *Store) GetByID(id int) (*model.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.data.Tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("task #%d not found", id)
}

// GetByStatus returns all tasks with the given status.
func (s *Store) GetByStatus(status model.Status) []*model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Task
	for _, t := range s.data.Tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

// GetByPriority returns all tasks with the given priority.
func (s *Store) GetByPriority(priority model.Priority) []*model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Task
	for _, t := range s.data.Tasks {
		if t.Priority == priority {
			result = append(result, t)
		}
	}
	return result
}

// MarkDone marks a task as completed by its ID.
func (s *Store) MarkDone(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.data.Tasks {
		if t.ID == id {
			if t.IsCompleted() {
				return fmt.Errorf("task #%d is already completed", id)
			}
			t.Complete()
			return s.save()
		}
	}
	return fmt.Errorf("task #%d not found", id)
}

// Delete removes a task from the store by its ID.
func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.data.Tasks {
		if t.ID == id {
			// Remove the task by slicing it out
			s.data.Tasks = append(s.data.Tasks[:i], s.data.Tasks[i+1:]...)
			return s.save()
		}
	}
	return fmt.Errorf("task #%d not found", id)
}

// Count returns the total number of tasks.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data.Tasks)
}

// ========================================
// Persistence — Load and Save
// ========================================

// load reads the JSON file and deserializes it into the store data.
func (s *Store) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	if err := json.Unmarshal(data, &s.data); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
	}

	return nil
}

// save serializes the store data to JSON and writes it to the file.
// It uses indented JSON for human readability.
func (s *Store) save() error {
	data, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return fmt.Errorf("serializing JSON: %w", err)
	}

	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
