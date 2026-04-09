package model

import (
	"fmt"
	"strings"
	"time"
)

// ========================================
// User Domain Model
// ========================================
// The user model represents authenticated users of the TaskFlow
// system. It demonstrates:
//   - Input validation (Week 2-3 error handling)
//   - Value objects for email (Week 5 types)
//   - Separation of domain model from API representation

// Role represents a user's permission level.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// IsValid checks if the role value is recognized.
func (r Role) IsValid() bool {
	return r == RoleUser || r == RoleAdmin
}

// ========================================
// User Entity
// ========================================

// User represents a registered user of the TaskFlow system.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never serialized to JSON
	Role         Role      `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// NewUser creates a new user with validated fields.
func NewUser(email, name, passwordHash string) (*User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, fmt.Errorf("invalid email format: %s", email)
	}
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if passwordHash == "" {
		return nil, fmt.Errorf("password hash is required")
	}

	now := time.Now().UTC()
	return &User{
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		Role:         RoleUser,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// RecordLogin updates the last login timestamp.
func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// Deactivate marks the user as inactive.
func (u *User) Deactivate() {
	u.Active = false
	u.UpdatedAt = time.Now().UTC()
}

// IsAdmin checks if the user has admin privileges.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// ========================================
// User DTOs
// ========================================

// CreateUserRequest is the payload for user registration.
type CreateUserRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// UserResponse is the API representation (excludes sensitive fields).
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      Role      `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts a domain User to an API UserResponse.
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
	}
}
