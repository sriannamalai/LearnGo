package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ========================================
// Week 9, Mini-Project: Testing Testable Code
// ========================================
// This test file demonstrates:
//   1. Mocking interfaces
//   2. Testing services with mock dependencies
//   3. Testing HTTP handlers with httptest
//   4. Testing pure functions
//
// Run: go test -v ./04_testable_code/

// ========================================
// 1. Mock Implementation (implements UserStore)
// ========================================
// A mock lets you control exactly what the store returns.
// This makes tests predictable and fast (no database needed!).

type MockUserStore struct {
	users      map[string]*User
	createErr  error // Set this to simulate store errors
	getUserErr error
}

func NewMockUserStore() *MockUserStore {
	return &MockUserStore{users: make(map[string]*User)}
}

func (m *MockUserStore) GetUser(id string) (*User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	user, ok := m.users[id]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", id)
	}
	return user, nil
}

func (m *MockUserStore) CreateUser(user *User) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.users[user.ID] = user
	return nil
}

func (m *MockUserStore) ListUsers() ([]*User, error) {
	users := make([]*User, 0, len(m.users))
	for _, u := range m.users {
		users = append(users, u)
	}
	return users, nil
}

// ========================================
// 2. Testing the Service Layer
// ========================================
// With the mock, we test business logic without any real database.

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		userName  string
		email     string
		storeErr  error
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid user",
			userName: "Alice",
			email:    "alice@example.com",
			wantErr:  false,
		},
		{
			name:      "empty name",
			userName:  "",
			email:     "alice@example.com",
			wantErr:   true,
			errSubstr: "name is required",
		},
		{
			name:      "empty email",
			userName:  "Alice",
			email:     "",
			wantErr:   true,
			errSubstr: "email is required",
		},
		{
			name:      "invalid email",
			userName:  "Alice",
			email:     "notanemail",
			wantErr:   true,
			errSubstr: "invalid email",
		},
		{
			name:      "store error",
			userName:  "Alice",
			email:     "alice@example.com",
			storeErr:  fmt.Errorf("database connection failed"),
			wantErr:   true,
			errSubstr: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh mock for each test
			mock := NewMockUserStore()
			mock.createErr = tt.storeErr

			service := NewUserService(mock)
			user, err := service.CreateUser(tt.userName, tt.email)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if user.Name != tt.userName {
				t.Errorf("user.Name = %q; want %q", user.Name, tt.userName)
			}
			if user.Email != tt.email {
				t.Errorf("user.Email = %q; want %q", user.Email, tt.email)
			}
			if user.ID == "" {
				t.Error("user.ID should not be empty")
			}
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	mock := NewMockUserStore()
	mock.users["user_123"] = &User{
		ID:    "user_123",
		Name:  "Alice",
		Email: "alice@example.com",
	}

	service := NewUserService(mock)

	t.Run("existing user", func(t *testing.T) {
		user, err := service.GetUser("user_123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("user.Name = %q; want %q", user.Name, "Alice")
		}
	})

	t.Run("empty ID", func(t *testing.T) {
		_, err := service.GetUser("")
		if err == nil {
			t.Fatal("expected error for empty ID")
		}
	})

	t.Run("nonexistent user", func(t *testing.T) {
		_, err := service.GetUser("user_999")
		if err == nil {
			t.Fatal("expected error for nonexistent user")
		}
	})
}

// ========================================
// 3. Testing HTTP Handlers with httptest
// ========================================
// httptest provides tools to test HTTP handlers WITHOUT starting a real server.
// httptest.NewRequest creates a fake request.
// httptest.NewRecorder captures the response.

func TestHandleGetUser(t *testing.T) {
	// Set up the mock with a test user
	mock := NewMockUserStore()
	mock.users["user_123"] = &User{
		ID:        "user_123",
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	service := NewUserService(mock)
	handler := NewUserHandler(service)

	// We need a mux for path value extraction to work
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", handler.HandleGetUser)

	t.Run("existing user returns 200", func(t *testing.T) {
		// Create a test request
		req := httptest.NewRequest("GET", "/users/user_123", nil)
		// Create a response recorder (implements http.ResponseWriter)
		rec := httptest.NewRecorder()

		// Call the handler through the mux (so PathValue works)
		mux.ServeHTTP(rec, req)

		// Check status code
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d; want %d", rec.Code, http.StatusOK)
		}

		// Check Content-Type header
		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Content-Type = %q; want %q", contentType, "application/json")
		}

		// Check response body
		var user User
		if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("user.Name = %q; want %q", user.Name, "Alice")
		}
	})

	t.Run("nonexistent user returns 404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/user_999", nil)
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d; want %d", rec.Code, http.StatusNotFound)
		}
	})
}

func TestHandleCreateUser(t *testing.T) {
	mock := NewMockUserStore()
	service := NewUserService(mock)
	handler := NewUserHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /users", handler.HandleCreateUser)

	t.Run("valid user returns 201", func(t *testing.T) {
		body := `{"name": "Bob", "email": "bob@example.com"}`
		req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d; want %d", rec.Code, http.StatusCreated)
		}

		var user User
		if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if user.Name != "Bob" {
			t.Errorf("user.Name = %q; want %q", user.Name, "Bob")
		}
		if user.ID == "" {
			t.Error("user.ID should not be empty")
		}
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		body := `{"email": "bob@example.com"}`
		req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		body := `{this is not json}`
		req := httptest.NewRequest("POST", "/users", strings.NewReader(body))
		rec := httptest.NewRecorder()

		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d; want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

// ========================================
// 4. Testing with httptest.Server (full HTTP server)
// ========================================
// httptest.NewServer starts a real HTTP server on a random port.
// Useful for integration-style tests.

func TestWithHTTPServer(t *testing.T) {
	mock := NewMockUserStore()
	mock.users["user_1"] = &User{
		ID:    "user_1",
		Name:  "Alice",
		Email: "alice@example.com",
	}

	service := NewUserService(mock)
	handler := NewUserHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", handler.HandleGetUser)
	mux.HandleFunc("POST /users", handler.HandleCreateUser)

	// Start a test server
	server := httptest.NewServer(mux)
	defer server.Close() // Always close the server when done

	t.Run("GET user via real HTTP", func(t *testing.T) {
		// Make a real HTTP request to the test server
		resp, err := http.Get(server.URL + "/users/user_1")
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d; want %d", resp.StatusCode, http.StatusOK)
		}

		var user User
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("user.Name = %q; want %q", user.Name, "Alice")
		}
	})
}

// ========================================
// 5. Testing Pure Functions
// ========================================
// Pure functions are the easiest to test — no mocking needed!

func TestFormatUserDisplay(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want string
	}{
		{
			name: "normal user",
			user: &User{Name: "Alice", Email: "alice@example.com"},
			want: "Alice (alice@example.com)",
		},
		{
			name: "nil user",
			user: nil,
			want: "<no user>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatUserDisplay(tt.user)
			if got != tt.want {
				t.Errorf("FormatUserDisplay() = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{name: "valid", email: "user@example.com", want: ""},
		{name: "empty", email: "", want: "email is required"},
		{name: "no @", email: "userexample.com", want: "email must contain @"},
		{name: "no domain", email: "user@", want: "invalid email format"},
		{name: "no local", email: "@example.com", want: "invalid email format"},
		{name: "no dot in domain", email: "user@example", want: "email domain must contain a dot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got != tt.want {
				t.Errorf("ValidateEmail(%q) = %q; want %q", tt.email, got, tt.want)
			}
		})
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Testable Code Patterns:
//
// 1. Dependency Injection via Interfaces
//    - Define an interface for what your code needs
//    - Accept the interface in constructors
//    - In tests, pass a mock; in production, pass the real thing
//
// 2. Mock Implementations
//    - Create a struct that implements the interface
//    - Add fields to control behavior (e.g., createErr error)
//    - Keeps tests fast and deterministic
//
// 3. httptest.NewRequest + httptest.NewRecorder
//    - Test handlers without starting a server
//    - Create a request: httptest.NewRequest("GET", "/path", body)
//    - Capture response: httptest.NewRecorder()
//    - Call handler: handler.ServeHTTP(recorder, request)
//    - Check: recorder.Code, recorder.Body, recorder.Header()
//
// 4. httptest.NewServer
//    - Starts a real HTTP server on localhost with random port
//    - Use server.URL to get the base URL
//    - Always defer server.Close()
//    - Great for integration tests
//
// 5. Pure Functions
//    - Take inputs, return outputs, no side effects
//    - Easiest to test — just check output vs expected
//    - Extract business logic into pure functions when possible
