package model

import (
	"fmt"
	"time"
)

// ========================================
// Task Model
// ========================================
// The task model defines the core domain entity for our CLI app.
// It uses JSON tags for serialization (since we store tasks in a
// JSON file) and provides methods for display and status management.

// Priority represents the urgency level of a task.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// Status represents the current state of a task.
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
)

// Task represents a single task in our task manager.
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Priority    Priority  `json:"priority"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// NewTask creates a new task with the given title and priority.
// The ID is assigned by the store when saving.
func NewTask(title string, priority Priority) *Task {
	return &Task{
		Title:     title,
		Priority:  priority,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
}

// Complete marks the task as completed and records the completion time.
func (t *Task) Complete() {
	t.Status = StatusCompleted
	now := time.Now()
	t.CompletedAt = &now
}

// IsPending returns true if the task is not yet completed.
func (t *Task) IsPending() bool {
	return t.Status == StatusPending
}

// IsCompleted returns true if the task has been completed.
func (t *Task) IsCompleted() bool {
	return t.Status == StatusCompleted
}

// String returns a human-readable representation of the task.
func (t *Task) String() string {
	check := "[ ]"
	if t.IsCompleted() {
		check = "[x]"
	}

	priorityIndicator := ""
	switch t.Priority {
	case PriorityHigh:
		priorityIndicator = "!!!"
	case PriorityMedium:
		priorityIndicator = "!! "
	case PriorityLow:
		priorityIndicator = "!  "
	}

	return fmt.Sprintf("%s %s #%d %s (%s)",
		check, priorityIndicator, t.ID, t.Title,
		t.CreatedAt.Format("Jan 02, 15:04"))
}

// ValidatePriority checks if a string is a valid priority value.
func ValidatePriority(p string) (Priority, error) {
	switch Priority(p) {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return Priority(p), nil
	default:
		return "", fmt.Errorf("invalid priority %q: must be low, medium, or high", p)
	}
}
