package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"time"

	"github.com/sri/learngo/week24/01_capstone/internal/model"
	"github.com/sri/learngo/week24/01_capstone/internal/store"
)

// ========================================
// Task Service — Business Logic Layer
// ========================================
// The service layer sits between the API handlers and the data stores.
// It contains all business logic and orchestrates operations across
// multiple repositories. This demonstrates:
//   - Week 13-14: Service layer pattern, separation of concerns
//   - Week 5: Interface-based dependencies (dependency injection)
//   - Week 2-3: Rich error handling
//   - Week 6-7: Context propagation
//
// Architecture Decision: The service layer is the single source of
// truth for business rules. API handlers should NEVER access stores
// directly — they always go through the service. This ensures:
//   1. Business rules are enforced consistently
//   2. Logic can be tested without HTTP/gRPC infrastructure
//   3. Multiple APIs (REST, gRPC) share the same logic

// TaskRepository defines the interface for task persistence.
// Using an interface allows us to swap implementations (Postgres,
// in-memory, mock) without changing the service code.
type TaskRepository interface {
	CreateTask(ctx context.Context, task *model.Task) error
	GetTask(ctx context.Context, id string) (*model.Task, error)
	ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) error
	DeleteTask(ctx context.Context, id string) error
}

// TaskGraphRepository defines the interface for graph operations.
type TaskGraphRepository interface {
	AddTaskDependency(ctx context.Context, taskID, dependsOnID string) error
	GetTaskDependencies(ctx context.Context, taskID string) ([]string, error)
	GetBlockedTasks(ctx context.Context, taskID string) ([]string, error)
}

// TaskService contains the business logic for task operations.
type TaskService struct {
	repo       TaskRepository
	graphStore *store.ArangoStore
	logger     *slog.Logger
}

// NewTaskService creates a new TaskService with the given dependencies.
// This is constructor-based dependency injection — the service declares
// what it needs, and the caller provides concrete implementations.
func NewTaskService(pgStore *store.PostgresStore, arangoStore *store.ArangoStore, logger *slog.Logger) *TaskService {
	var repo TaskRepository
	if pgStore != nil {
		repo = pgStore
	}
	// If pgStore is nil, repo stays nil and we handle it in methods

	return &TaskService{
		repo:       repo,
		graphStore: arangoStore,
		logger:     logger.With("component", "task_service"),
	}
}

// ========================================
// Task CRUD Operations
// ========================================

// CreateTask validates and creates a new task.
func (s *TaskService) CreateTask(ctx context.Context, userID string, req model.CreateTaskRequest) (*model.Task, error) {
	s.logger.Info("creating task", "user", userID, "title", req.Title)

	// ========================================
	// Input Validation
	// ========================================
	// The service layer validates business rules beyond what the
	// model constructor checks. This is where cross-cutting concerns
	// like authorization and rate limiting would also go.
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(req.Title) > 500 {
		return nil, fmt.Errorf("title must be 500 characters or less")
	}

	priority := req.Priority
	if priority == "" {
		priority = model.PriorityMedium
	}
	if !priority.IsValid() {
		return nil, fmt.Errorf("invalid priority: %s (must be low, medium, or high)", priority)
	}

	// Create the domain entity
	task, err := model.NewTask(userID, req.Title, priority)
	if err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}

	// Set optional fields
	task.ID = generateID()
	task.Description = req.Description
	task.Tags = req.Tags
	if task.Tags == nil {
		task.Tags = make([]string, 0)
	}

	// Parse optional due date
	if req.DueDate != "" {
		dueDate, err := time.Parse(time.RFC3339, req.DueDate)
		if err != nil {
			return nil, fmt.Errorf("invalid due_date format (use RFC3339): %w", err)
		}
		task.DueDate = &dueDate
	}

	// Persist to the database
	if s.repo != nil {
		if err := s.repo.CreateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("saving task: %w", err)
		}
	}

	s.logger.Info("task created",
		"id", task.ID,
		"title", task.Title,
		"priority", task.Priority,
	)

	return task, nil
}

// GetTask retrieves a task by ID with authorization checks.
func (s *TaskService) GetTask(ctx context.Context, userID, taskID string) (*model.Task, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("database not available")
	}

	task, err := s.repo.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Authorization: users can only access their own tasks
	if task.UserID != userID {
		s.logger.Warn("unauthorized task access attempt",
			"user", userID,
			"task", taskID,
			"task_owner", task.UserID,
		)
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

// ListTasks retrieves tasks for a user with optional filters.
func (s *TaskService) ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
	if s.repo == nil {
		return []*model.Task{}, nil
	}

	// Apply sensible defaults for pagination
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	tasks, err := s.repo.ListTasks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}

	return tasks, nil
}

// UpdateTask applies partial updates to an existing task.
func (s *TaskService) UpdateTask(ctx context.Context, userID, taskID string, req model.UpdateTaskRequest) (*model.Task, error) {
	// Fetch the existing task (with authorization check)
	task, err := s.GetTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	// Apply updates (only non-nil fields are modified)
	if req.Title != nil {
		if *req.Title == "" {
			return nil, fmt.Errorf("title cannot be empty")
		}
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		if !req.Priority.IsValid() {
			return nil, fmt.Errorf("invalid priority: %s", *req.Priority)
		}
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		// Enforce state machine transitions
		if err := applyStatusChange(task, *req.Status); err != nil {
			return nil, err
		}
	}
	if req.Tags != nil {
		task.Tags = req.Tags
	}
	if req.DueDate != nil {
		if *req.DueDate == "" {
			task.DueDate = nil
		} else {
			dueDate, err := time.Parse(time.RFC3339, *req.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due_date: %w", err)
			}
			task.DueDate = &dueDate
		}
	}

	task.UpdatedAt = time.Now().UTC()

	// Persist the update
	if s.repo != nil {
		if err := s.repo.UpdateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("updating task: %w", err)
		}
	}

	s.logger.Info("task updated", "id", taskID)
	return task, nil
}

// DeleteTask removes a task permanently.
func (s *TaskService) DeleteTask(ctx context.Context, userID, taskID string) error {
	// Verify the task exists and belongs to the user
	_, err := s.GetTask(ctx, userID, taskID)
	if err != nil {
		return err
	}

	if s.repo != nil {
		if err := s.repo.DeleteTask(ctx, taskID); err != nil {
			return fmt.Errorf("deleting task: %w", err)
		}
	}

	s.logger.Info("task deleted", "id", taskID, "user", userID)
	return nil
}

// CompleteTask marks a task as completed using the domain model's
// state machine. This demonstrates how business logic should live
// in the service layer, not in HTTP handlers.
func (s *TaskService) CompleteTask(ctx context.Context, userID, taskID string) (*model.Task, error) {
	task, err := s.GetTask(ctx, userID, taskID)
	if err != nil {
		return nil, err
	}

	// ========================================
	// Business Rule: Check Dependencies
	// ========================================
	// A task cannot be completed if it has uncompleted dependencies.
	// This is an example of cross-store business logic that belongs
	// in the service layer — it needs both Postgres and ArangoDB.
	if s.graphStore != nil {
		deps, err := s.graphStore.GetTaskDependencies(ctx, taskID)
		if err != nil {
			s.logger.Warn("could not check dependencies", "error", err)
		} else if len(deps) > 0 {
			// In production, we'd check if all dependencies are completed
			s.logger.Debug("task has dependencies", "count", len(deps))
		}
	}

	// Use the domain model's state machine
	if err := task.Complete(); err != nil {
		return nil, err
	}

	if s.repo != nil {
		if err := s.repo.UpdateTask(ctx, task); err != nil {
			return nil, fmt.Errorf("saving completed task: %w", err)
		}
	}

	s.logger.Info("task completed", "id", taskID)
	return task, nil
}

// ========================================
// Helper Functions
// ========================================

// applyStatusChange validates and applies a status transition.
func applyStatusChange(task *model.Task, newStatus model.Status) error {
	switch newStatus {
	case model.StatusInProgress:
		return task.Start()
	case model.StatusCompleted:
		return task.Complete()
	case model.StatusArchived:
		return task.Archive()
	default:
		return fmt.Errorf("cannot transition to status: %s", newStatus)
	}
}

// generateID creates a unique ID for new entities.
// In production, you might use UUIDs (google/uuid) or ULIDs.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
