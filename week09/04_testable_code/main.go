package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ========================================
// Week 9, Mini-Project: Writing Testable Code
// ========================================
// This lesson teaches the patterns that make Go code easy to test:
//   1. Dependency injection via interfaces
//   2. Interfaces for mocking external dependencies
//   3. Testing HTTP handlers with httptest
//   4. Separating pure logic from I/O
//
// Run the program:
//   go run .
//
// Run the tests:
//   cd week09
//   go test -v ./04_testable_code/

// ========================================
// 1. Interfaces for Dependency Injection
// ========================================
// Instead of depending on a concrete database, our service depends
// on an interface. This means we can swap in a mock for testing.

// UserStore defines what our service needs from a data store.
// Any type that implements these methods satisfies the interface.
type UserStore interface {
	GetUser(id string) (*User, error)
	CreateUser(user *User) error
	ListUsers() ([]*User, error)
}

// User represents a user in our system.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// ========================================
// 2. Service layer — contains business logic
// ========================================
// The UserService depends on the UserStore INTERFACE, not a concrete type.
// This is dependency injection: we "inject" the dependency via the constructor.

// UserService handles business logic for users.
type UserService struct {
	store UserStore
}

// NewUserService creates a UserService with the given store.
// This is the "constructor" pattern in Go.
func NewUserService(store UserStore) *UserService {
	return &UserService{store: store}
}

// GetUser retrieves a user by ID with validation.
func (s *UserService) GetUser(id string) (*User, error) {
	if id == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.store.GetUser(id)
}

// CreateUser creates a new user with validation.
func (s *UserService) CreateUser(name, email string) (*User, error) {
	// Validate inputs (pure logic — easy to test!)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") {
		return nil, fmt.Errorf("invalid email format")
	}

	user := &User{
		ID:        fmt.Sprintf("user_%d", time.Now().UnixNano()),
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}

	if err := s.store.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// ========================================
// 3. In-memory store (implements UserStore)
// ========================================
// This is the "real" implementation used in production.

type MemoryStore struct {
	users map[string]*User
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{users: make(map[string]*User)}
}

func (m *MemoryStore) GetUser(id string) (*User, error) {
	user, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return user, nil
}

func (m *MemoryStore) CreateUser(user *User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MemoryStore) ListUsers() ([]*User, error) {
	users := make([]*User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

// ========================================
// 4. HTTP Handlers — separate from business logic
// ========================================
// Handlers only deal with HTTP concerns (parsing requests, writing responses).
// Business logic lives in the service layer.

// UserHandler contains HTTP handlers for user operations.
type UserHandler struct {
	service *UserService
}

// NewUserHandler creates a UserHandler with the given service.
func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

// HandleGetUser handles GET /users/{id}
func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	user, err := h.service.GetUser(id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, user)
}

// HandleCreateUser handles POST /users
func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &input); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Delegate to service (which handles validation and business logic)
	user, err := h.service.CreateUser(input.Name, input.Email)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusCreated, user)
}

// HandleListUsers handles GET /users
func (h *UserHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.store.ListUsers()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, users)
}

// ========================================
// 5. JSON response helpers
// ========================================
// These are pure functions — they just transform data.

func writeJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// ========================================
// 6. Pure functions — easiest to test
// ========================================
// Functions that take inputs and return outputs without side effects.

// FormatUserDisplay creates a display string for a user.
func FormatUserDisplay(user *User) string {
	if user == nil {
		return "<no user>"
	}
	return fmt.Sprintf("%s (%s)", user.Name, user.Email)
}

// ValidateEmail checks if an email address is valid.
// Returns an error message or empty string.
func ValidateEmail(email string) string {
	if email == "" {
		return "email is required"
	}
	if !strings.Contains(email, "@") {
		return "email must contain @"
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "invalid email format"
	}
	if !strings.Contains(parts[1], ".") {
		return "email domain must contain a dot"
	}
	return ""
}

// ========================================
// Main — wire everything together
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 9 Project: Testable Code Patterns")
	fmt.Println("========================================")

	// Wire up dependencies (dependency injection)
	store := NewMemoryStore()
	service := NewUserService(store)
	handler := NewUserHandler(service)

	// Create some sample users
	user1, _ := service.CreateUser("Alice", "alice@example.com")
	user2, _ := service.CreateUser("Bob", "bob@example.com")

	fmt.Println("\nCreated users:")
	fmt.Printf("  %s\n", FormatUserDisplay(user1))
	fmt.Printf("  %s\n", FormatUserDisplay(user2))

	// Set up HTTP routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users", handler.HandleListUsers)
	mux.HandleFunc("GET /users/{id}", handler.HandleGetUser)
	mux.HandleFunc("POST /users", handler.HandleCreateUser)

	fmt.Println("\nTestable Code Patterns Demonstrated:")
	fmt.Println("  1. Interface-based dependency injection (UserStore)")
	fmt.Println("  2. Service layer with business logic (UserService)")
	fmt.Println("  3. HTTP handlers separated from logic (UserHandler)")
	fmt.Println("  4. Pure helper functions (FormatUserDisplay, ValidateEmail)")
	fmt.Println()
	fmt.Println("Run the tests to see these patterns in action:")
	fmt.Println("  go test -v ./04_testable_code/")
	fmt.Println()
	fmt.Println("Starting server on http://localhost:8080")
	fmt.Println("  curl http://localhost:8080/users")

	http.ListenAndServe(":8080", mux)
}
