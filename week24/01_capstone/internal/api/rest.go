package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sri/learngo/week24/01_capstone/internal/model"
	"github.com/sri/learngo/week24/01_capstone/internal/service"
)

// ========================================
// REST API Handlers
// ========================================
// These handlers implement the REST API for TaskFlow using Go's
// standard library net/http package with Go 1.22+ enhanced routing.
// This demonstrates:
//   - Week 8-9: HTTP handlers, JSON encoding/decoding
//   - Week 8: Go 1.22+ method-based routing (GET /api/tasks)
//   - Week 2-3: Error handling with proper HTTP status codes
//   - Week 13-14: Handler depends on service layer, not stores
//
// Architecture Decision: We use net/http rather than a framework
// (like Gin or Echo) because:
//   1. Go 1.22+ routing is powerful enough for most APIs
//   2. Zero external dependencies for the HTTP layer
//   3. Full control over request/response lifecycle
//   4. Easy to add middleware without framework lock-in

// RESTHandler holds the dependencies for REST API handlers.
type RESTHandler struct {
	taskService *service.TaskService
	userService *service.UserService
	logger      *slog.Logger
}

// NewRESTHandler creates a new REST handler with injected dependencies.
func NewRESTHandler(taskSvc *service.TaskService, userSvc *service.UserService, logger *slog.Logger) *RESTHandler {
	return &RESTHandler{
		taskService: taskSvc,
		userService: userSvc,
		logger:      logger.With("component", "rest_api"),
	}
}

// RegisterRoutes registers all REST API routes on the given mux.
// Go 1.22+ supports method-based routing: "GET /path" only matches GET.
func (h *RESTHandler) RegisterRoutes(mux *http.ServeMux) {
	// ========================================
	// Task Endpoints
	// ========================================
	mux.HandleFunc("POST /api/v1/tasks", h.CreateTask)
	mux.HandleFunc("GET /api/v1/tasks", h.ListTasks)
	mux.HandleFunc("GET /api/v1/tasks/{id}", h.GetTask)
	mux.HandleFunc("PUT /api/v1/tasks/{id}", h.UpdateTask)
	mux.HandleFunc("DELETE /api/v1/tasks/{id}", h.DeleteTask)
	mux.HandleFunc("POST /api/v1/tasks/{id}/complete", h.CompleteTask)

	// ========================================
	// User Endpoints
	// ========================================
	mux.HandleFunc("POST /api/v1/users/register", h.RegisterUser)
	mux.HandleFunc("POST /api/v1/users/login", h.LoginUser)
	mux.HandleFunc("GET /api/v1/users/me", h.GetCurrentUser)

	h.logger.Info("REST API routes registered")
}

// ========================================
// Task Handlers
// ========================================

// CreateTask handles POST /api/v1/tasks
func (h *RESTHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID := getUserIDFromContext(r)

	// Decode the request body
	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest,
			"invalid request body: "+err.Error())
		return
	}

	// Create the task via the service layer
	task, err := h.taskService.CreateTask(r.Context(), userID, req)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, task.ToResponse())
}

// ListTasks handles GET /api/v1/tasks
func (h *RESTHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	// Parse query parameters for filtering
	filter := model.TaskFilter{
		UserID: userID,
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = model.Status(status)
	}
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filter.Priority = model.Priority(priority)
	}

	tasks, err := h.taskService.ListTasks(r.Context(), filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert to response DTOs
	responses := make([]*model.TaskResponse, len(tasks))
	for i, t := range tasks {
		responses[i] = t.ToResponse()
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": responses,
		"total": len(responses),
	})
}

// GetTask handles GET /api/v1/tasks/{id}
func (h *RESTHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	// Go 1.22+ path value extraction
	taskID := r.PathValue("id")
	if taskID == "" {
		h.writeError(w, http.StatusBadRequest, "task ID is required")
		return
	}

	task, err := h.taskService.GetTask(r.Context(), userID, taskID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, task.ToResponse())
}

// UpdateTask handles PUT /api/v1/tasks/{id}
func (h *RESTHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	taskID := r.PathValue("id")

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest,
			"invalid request body: "+err.Error())
		return
	}

	task, err := h.taskService.UpdateTask(r.Context(), userID, taskID, req)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, task.ToResponse())
}

// DeleteTask handles DELETE /api/v1/tasks/{id}
func (h *RESTHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	taskID := r.PathValue("id")

	if err := h.taskService.DeleteTask(r.Context(), userID, taskID); err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("task %s deleted", taskID),
	})
}

// CompleteTask handles POST /api/v1/tasks/{id}/complete
func (h *RESTHandler) CompleteTask(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	taskID := r.PathValue("id")

	task, err := h.taskService.CompleteTask(r.Context(), userID, taskID)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, task.ToResponse())
}

// ========================================
// User Handlers
// ========================================

// RegisterUser handles POST /api/v1/users/register
func (h *RESTHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest,
			"invalid request body: "+err.Error())
		return
	}

	user, err := h.userService.RegisterUser(r.Context(), req)
	if err != nil {
		h.writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	h.writeJSON(w, http.StatusCreated, user.ToResponse())
}

// LoginUser handles POST /api/v1/users/login
func (h *RESTHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest,
			"invalid request body: "+err.Error())
		return
	}

	user, err := h.userService.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		// Use 401 for authentication failures
		h.writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// In production, you'd generate and return a JWT token here.
	// See Week 21-22 for JWT implementation patterns.
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"user":    user.ToResponse(),
		"token":   "jwt-token-would-go-here",
		"message": "authentication successful",
	})
}

// GetCurrentUser handles GET /api/v1/users/me
func (h *RESTHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)

	user, err := h.userService.GetUser(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeJSON(w, http.StatusOK, user.ToResponse())
}

// ========================================
// Response Helpers
// ========================================
// These helper methods ensure consistent JSON responses and error
// formats across all endpoints. This is a pattern from Week 8-9.

// writeJSON writes a JSON response with the given status code.
func (h *RESTHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes a JSON error response.
func (h *RESTHandler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error":  http.StatusText(status),
		"detail": message,
		"status": status,
	})
}

// ========================================
// Context Helpers
// ========================================

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const userIDKey contextKey = "user_id"

// getUserIDFromContext extracts the user ID from the request context.
// The auth middleware sets this value after validating the JWT token.
func getUserIDFromContext(r *http.Request) string {
	if userID, ok := r.Context().Value(userIDKey).(string); ok {
		return userID
	}
	// Default user for development/demo mode
	return "demo-user-001"
}
