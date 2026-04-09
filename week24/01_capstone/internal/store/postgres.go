package store

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sri/learngo/week24/01_capstone/internal/model"
)

// ========================================
// PostgreSQL Repository
// ========================================
// This file implements the repository pattern for PostgreSQL.
// It demonstrates:
//   - Week 10: pgx connection pooling, parameterized queries
//   - Week 13-14: Repository pattern, interface-based design
//   - Week 6: Context propagation for timeouts and cancellation
//   - Week 2-3: Comprehensive error handling with wrapping
//
// Architecture Decision: We use pgx directly (not an ORM) because:
//   1. Full control over SQL queries for performance tuning
//   2. No magic — every query is explicit and auditable
//   3. pgx is the fastest PostgreSQL driver for Go
//   4. Connection pooling is built-in via pgxpool

// PostgresConfig holds the PostgreSQL connection parameters.
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
}

// DSN constructs the connection string from the config.
func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// PostgresStore implements task and user storage using PostgreSQL.
type PostgresStore struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPostgresStore creates a new PostgreSQL store with a connection pool.
func NewPostgresStore(ctx context.Context, cfg PostgresConfig) (*PostgresStore, error) {
	// Configure the connection pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parsing postgres config: %w", err)
	}

	// Pool settings for production
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// Verify connectivity with a ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	logger := slog.Default().With("component", "postgres")
	logger.Info("PostgreSQL connection pool established",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
		"max_conns", cfg.MaxConns,
	)

	return &PostgresStore{pool: pool, logger: logger}, nil
}

// Close shuts down the connection pool.
func (s *PostgresStore) Close() {
	if s.pool != nil {
		s.pool.Close()
		s.logger.Info("PostgreSQL connection pool closed")
	}
}

// ========================================
// Task Repository Operations
// ========================================

// CreateTask inserts a new task into the database.
func (s *PostgresStore) CreateTask(ctx context.Context, task *model.Task) error {
	query := `
		INSERT INTO tasks (id, user_id, title, description, priority, status, tags, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := s.pool.QueryRow(ctx, query,
		task.ID,
		task.UserID,
		task.Title,
		task.Description,
		task.Priority,
		task.Status,
		task.Tags,
		task.DueDate,
		task.CreatedAt,
		task.UpdatedAt,
	).Scan(&task.ID)

	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}

	s.logger.Debug("task created", "id", task.ID, "title", task.Title)
	return nil
}

// GetTask retrieves a single task by ID.
func (s *PostgresStore) GetTask(ctx context.Context, id string) (*model.Task, error) {
	query := `
		SELECT id, user_id, title, description, priority, status, tags,
		       due_date, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = $1`

	task := &model.Task{}
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Priority,
		&task.Status,
		&task.Tags,
		&task.DueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
		&task.CompletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("task %s not found", id)
		}
		return nil, fmt.Errorf("querying task: %w", err)
	}

	return task, nil
}

// ListTasks retrieves tasks matching the given filter criteria.
// This demonstrates dynamic query building with parameterized queries
// to prevent SQL injection (Week 22 security patterns).
func (s *PostgresStore) ListTasks(ctx context.Context, filter model.TaskFilter) ([]*model.Task, error) {
	// ========================================
	// Dynamic Query Builder
	// ========================================
	// Build the WHERE clause dynamically based on which filters are set.
	// Use parameterized queries ($1, $2, ...) to prevent SQL injection.
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}
	if filter.Priority != "" {
		conditions = append(conditions, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, filter.Priority)
		argIndex++
	}

	query := `
		SELECT id, user_id, title, description, priority, status, tags,
		       due_date, created_at, updated_at, completed_at
		FROM tasks`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	// Execute the query
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying tasks: %w", err)
	}
	defer rows.Close()

	// Scan results
	var tasks []*model.Task
	for rows.Next() {
		task := &model.Task{}
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.Status,
			&task.Tags,
			&task.DueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning task row: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating task rows: %w", err)
	}

	return tasks, nil
}

// UpdateTask updates an existing task in the database.
func (s *PostgresStore) UpdateTask(ctx context.Context, task *model.Task) error {
	query := `
		UPDATE tasks
		SET title = $1, description = $2, priority = $3, status = $4,
		    tags = $5, due_date = $6, updated_at = $7, completed_at = $8
		WHERE id = $9`

	result, err := s.pool.Exec(ctx, query,
		task.Title,
		task.Description,
		task.Priority,
		task.Status,
		task.Tags,
		task.DueDate,
		task.UpdatedAt,
		task.CompletedAt,
		task.ID,
	)

	if err != nil {
		return fmt.Errorf("updating task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task %s not found", task.ID)
	}

	s.logger.Debug("task updated", "id", task.ID)
	return nil
}

// DeleteTask removes a task from the database.
func (s *PostgresStore) DeleteTask(ctx context.Context, id string) error {
	result, err := s.pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	s.logger.Debug("task deleted", "id", id)
	return nil
}

// ========================================
// User Repository Operations
// ========================================

// CreateUser inserts a new user into the database.
func (s *PostgresStore) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, email, name, password_hash, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := s.pool.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.PasswordHash,
		user.Role,
		user.Active,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}

	s.logger.Debug("user created", "id", user.ID, "email", user.Email)
	return nil
}

// GetUser retrieves a user by ID.
func (s *PostgresStore) GetUser(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, name, password_hash, role, active, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1`

	user := &model.User{}
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user %s not found", id)
		}
		return nil, fmt.Errorf("querying user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email address.
func (s *PostgresStore) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, name, password_hash, role, active, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1`

	user := &model.User{}
	err := s.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.Role,
		&user.Active,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return user, nil
}
