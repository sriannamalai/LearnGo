package model

import (
	"fmt"
	"time"
)

// ========================================
// Task Domain Model
// ========================================
// The task model represents the core domain entity. Following
// domain-driven design principles (Week 13-14), the model:
//   - Contains domain logic (validation, state transitions)
//   - Is independent of storage (no SQL or DB-specific tags)
//   - Uses value objects for constrained types (Priority, Status)
//   - Provides factory functions for creation

// Version of the application, set at build time.
var Version = "1.0.0"

// Priority represents the urgency level of a task.
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// IsValid checks if the priority value is recognized.
func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}

// Status represents the lifecycle state of a task.
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusArchived   Status = "archived"
)

// IsValid checks if the status value is recognized.
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusCompleted, StatusArchived:
		return true
	}
	return false
}

// ========================================
// Task Entity
// ========================================

// Task represents a unit of work in the TaskFlow system.
// It is the primary domain entity and the aggregate root for
// task-related operations.
type Task struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Priority    Priority   `json:"priority"`
	Status      Status     `json:"status"`
	Tags        []string   `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// NewTask creates a new task with validated fields.
// This factory function ensures tasks are always created in a
// valid state — a pattern from Week 5 (structs and constructors).
func NewTask(userID, title string, priority Priority) (*Task, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if !priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s", priority)
	}

	now := time.Now().UTC()
	return &Task{
		UserID:    userID,
		Title:     title,
		Priority:  priority,
		Status:    StatusPending,
		Tags:      make([]string, 0),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// ========================================
// Domain Logic — State Transitions
// ========================================
// Tasks follow a state machine: pending -> in_progress -> completed -> archived.
// These methods enforce valid transitions and record timestamps.

// Start moves the task from pending to in_progress.
func (t *Task) Start() error {
	if t.Status != StatusPending {
		return fmt.Errorf("cannot start task: current status is %s (must be pending)", t.Status)
	}
	t.Status = StatusInProgress
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// Complete moves the task to completed status.
func (t *Task) Complete() error {
	if t.Status != StatusPending && t.Status != StatusInProgress {
		return fmt.Errorf("cannot complete task: current status is %s", t.Status)
	}
	t.Status = StatusCompleted
	now := time.Now().UTC()
	t.CompletedAt = &now
	t.UpdatedAt = now
	return nil
}

// Archive moves a completed task to archived status.
func (t *Task) Archive() error {
	if t.Status != StatusCompleted {
		return fmt.Errorf("cannot archive task: must be completed first (current: %s)", t.Status)
	}
	t.Status = StatusArchived
	t.UpdatedAt = time.Now().UTC()
	return nil
}

// IsOverdue checks if the task is past its due date.
func (t *Task) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	if t.Status == StatusCompleted || t.Status == StatusArchived {
		return false
	}
	return time.Now().After(*t.DueDate)
}

// ========================================
// Task Filters
// ========================================
// Filter types for querying tasks. These are used by the service
// layer and passed to the repository for database queries.

// TaskFilter defines criteria for searching tasks.
type TaskFilter struct {
	UserID   string   `json:"user_id,omitempty"`
	Status   Status   `json:"status,omitempty"`
	Priority Priority `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}

// ========================================
// Request/Response DTOs
// ========================================
// Data Transfer Objects for the API layer. These decouple the
// API representation from the domain model, allowing each to
// evolve independently (Week 8-9 patterns).

// CreateTaskRequest is the payload for creating a new task.
type CreateTaskRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Priority    Priority `json:"priority"`
	Tags        []string `json:"tags,omitempty"`
	DueDate     string   `json:"due_date,omitempty"` // RFC3339 format
}

// UpdateTaskRequest is the payload for updating a task.
type UpdateTaskRequest struct {
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	Priority    *Priority `json:"priority,omitempty"`
	Status      *Status   `json:"status,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	DueDate     *string   `json:"due_date,omitempty"`
}

// TaskResponse is the API representation of a task.
type TaskResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Priority    Priority   `json:"priority"`
	Status      Status     `json:"status"`
	Tags        []string   `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	IsOverdue   bool       `json:"is_overdue"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// ToResponse converts a domain Task to an API TaskResponse.
func (t *Task) ToResponse() *TaskResponse {
	return &TaskResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Status:      t.Status,
		Tags:        t.Tags,
		DueDate:     t.DueDate,
		IsOverdue:   t.IsOverdue(),
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		CompletedAt: t.CompletedAt,
	}
}
