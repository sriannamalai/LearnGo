package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/sri/learngo/week24/01_capstone/internal/model"
	"github.com/sri/learngo/week24/01_capstone/internal/store"
)

// ========================================
// User Service — Authentication & User Management
// ========================================
// The user service handles user registration, authentication,
// and profile management. It demonstrates:
//   - Week 21-22: Security patterns (password hashing, auth)
//   - Week 13-14: Service layer with dependency injection
//   - Week 2-3: Defensive programming and validation
//
// SECURITY NOTE: In production, use bcrypt or argon2 for password
// hashing. This example uses SHA-256 for simplicity — it is NOT
// suitable for production password storage.

// UserRepository defines the interface for user persistence.
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetUser(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
}

// UserService contains the business logic for user operations.
type UserService struct {
	repo   UserRepository
	logger *slog.Logger
}

// NewUserService creates a new UserService with the given dependencies.
func NewUserService(pgStore *store.PostgresStore, logger *slog.Logger) *UserService {
	var repo UserRepository
	if pgStore != nil {
		repo = pgStore
	}

	return &UserService{
		repo:   repo,
		logger: logger.With("component", "user_service"),
	}
}

// ========================================
// User Registration
// ========================================

// RegisterUser creates a new user account.
func (s *UserService) RegisterUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	s.logger.Info("registering user", "email", req.Email)

	// Validate input
	if req.Email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	// Check for existing user with the same email
	if s.repo != nil {
		existing, _ := s.repo.GetUserByEmail(ctx, req.Email)
		if existing != nil {
			return nil, fmt.Errorf("email %s is already registered", req.Email)
		}
	}

	// Hash the password
	// IMPORTANT: In production, use bcrypt:
	//   hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	passwordHash := hashPassword(req.Password)

	// Create the user entity
	user, err := model.NewUser(req.Email, req.Name, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	user.ID = generateID()

	// Persist to database
	if s.repo != nil {
		if err := s.repo.CreateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("saving user: %w", err)
		}
	}

	s.logger.Info("user registered", "id", user.ID, "email", user.Email)
	return user, nil
}

// ========================================
// User Authentication
// ========================================

// AuthenticateUser verifies credentials and returns the user if valid.
func (s *UserService) AuthenticateUser(ctx context.Context, email, password string) (*model.User, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("database not available")
	}

	// Look up the user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal whether the email exists (security best practice)
		s.logger.Debug("authentication failed: user not found", "email", email)
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if the account is active
	if !user.Active {
		s.logger.Warn("authentication attempt on deactivated account", "email", email)
		return nil, fmt.Errorf("account is deactivated")
	}

	// Verify the password
	// In production: bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if hashPassword(password) != user.PasswordHash {
		s.logger.Debug("authentication failed: wrong password", "email", email)
		return nil, fmt.Errorf("invalid email or password")
	}

	// Record the login
	user.RecordLogin()

	s.logger.Info("user authenticated", "id", user.ID, "email", user.Email)
	return user, nil
}

// ========================================
// User Profile
// ========================================

// GetUser retrieves a user by their ID.
func (s *UserService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("database not available")
	}

	return s.repo.GetUser(ctx, userID)
}

// ========================================
// Helpers
// ========================================

// hashPassword creates a SHA-256 hash of the password.
// WARNING: This is for educational purposes only.
// In production, always use bcrypt or argon2.
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}
