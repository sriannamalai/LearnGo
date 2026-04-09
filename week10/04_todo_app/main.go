package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ========================================
// Week 10, Mini-Project: TODO App with PostgreSQL
// ========================================
// A complete command-line TODO application backed by PostgreSQL.
// Demonstrates all CRUD operations with a real database.
//
// Prerequisites:
//   1. PostgreSQL running with "learngo" database
//   2. export DATABASE_URL="postgres://postgres:postgres@localhost:5432/learngo?sslmode=disable"
//   3. cd week10 && go mod tidy
//
// Run:
//   go run ./04_todo_app/
//
// Commands:
//   add <text>    — Add a new todo
//   list          — List all todos
//   done <id>     — Mark a todo as complete
//   undone <id>   — Mark a todo as incomplete
//   edit <id>     — Edit a todo's text
//   delete <id>   — Delete a todo
//   search <text> — Search todos by text
//   clear         — Delete all completed todos
//   stats         — Show todo statistics
//   quit          — Exit the application

// ========================================
// Data Model
// ========================================

// Todo represents a single todo item.
type Todo struct {
	ID          int
	Title       string
	Completed   bool
	CreatedAt   time.Time
	CompletedAt *time.Time // nil if not completed (NULL in SQL)
}

// ========================================
// TodoStore handles all database operations
// ========================================

type TodoStore struct {
	pool *pgxpool.Pool
}

func NewTodoStore(pool *pgxpool.Pool) *TodoStore {
	return &TodoStore{pool: pool}
}

// SetupSchema creates the todos table if it doesn't exist.
func (s *TodoStore) SetupSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS todos (
			id           SERIAL PRIMARY KEY,
			title        TEXT NOT NULL,
			completed    BOOLEAN NOT NULL DEFAULT false,
			created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE
		)
	`
	_, err := s.pool.Exec(ctx, query)
	return err
}

// Add creates a new todo and returns it.
func (s *TodoStore) Add(ctx context.Context, title string) (*Todo, error) {
	query := `
		INSERT INTO todos (title)
		VALUES ($1)
		RETURNING id, title, completed, created_at, completed_at
	`

	var todo Todo
	err := s.pool.QueryRow(ctx, query, title).Scan(
		&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.CompletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}
	return &todo, nil
}

// List returns all todos, ordered by completion status then creation date.
func (s *TodoStore) List(ctx context.Context) ([]Todo, error) {
	query := `
		SELECT id, title, completed, created_at, completed_at
		FROM todos
		ORDER BY completed ASC, created_at DESC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list todos: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, t)
	}

	return todos, rows.Err()
}

// SetCompleted marks a todo as completed or incomplete.
func (s *TodoStore) SetCompleted(ctx context.Context, id int, completed bool) (*Todo, error) {
	var completedAt any
	if completed {
		completedAt = time.Now()
	} else {
		completedAt = nil // SQL NULL
	}

	query := `
		UPDATE todos
		SET completed = $1, completed_at = $2
		WHERE id = $3
		RETURNING id, title, completed, created_at, completed_at
	`

	var todo Todo
	err := s.pool.QueryRow(ctx, query, completed, completedAt, id).Scan(
		&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.CompletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("todo #%d not found", id)
		}
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}
	return &todo, nil
}

// Edit updates a todo's title.
func (s *TodoStore) Edit(ctx context.Context, id int, newTitle string) (*Todo, error) {
	query := `
		UPDATE todos
		SET title = $1
		WHERE id = $2
		RETURNING id, title, completed, created_at, completed_at
	`

	var todo Todo
	err := s.pool.QueryRow(ctx, query, newTitle, id).Scan(
		&todo.ID, &todo.Title, &todo.Completed, &todo.CreatedAt, &todo.CompletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("todo #%d not found", id)
		}
		return nil, fmt.Errorf("failed to edit todo: %w", err)
	}
	return &todo, nil
}

// Delete removes a todo by ID.
func (s *TodoStore) Delete(ctx context.Context, id int) error {
	result, err := s.pool.Exec(ctx, "DELETE FROM todos WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("todo #%d not found", id)
	}
	return nil
}

// Search finds todos matching the given text (case-insensitive).
func (s *TodoStore) Search(ctx context.Context, query string) ([]Todo, error) {
	sql := `
		SELECT id, title, completed, created_at, completed_at
		FROM todos
		WHERE title ILIKE '%' || $1 || '%'
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, sql, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search todos: %w", err)
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed, &t.CreatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, t)
	}

	return todos, rows.Err()
}

// ClearCompleted deletes all completed todos.
func (s *TodoStore) ClearCompleted(ctx context.Context) (int, error) {
	result, err := s.pool.Exec(ctx, "DELETE FROM todos WHERE completed = true")
	if err != nil {
		return 0, fmt.Errorf("failed to clear completed: %w", err)
	}
	return int(result.RowsAffected()), nil
}

// Stats returns statistics about todos.
func (s *TodoStore) Stats(ctx context.Context) (total, completed, pending int, err error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE completed = true) as completed,
			COUNT(*) FILTER (WHERE completed = false) as pending
		FROM todos
	`
	err = s.pool.QueryRow(ctx, query).Scan(&total, &completed, &pending)
	return
}

// DropTable removes the todos table (for cleanup).
func (s *TodoStore) DropTable(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, "DROP TABLE IF EXISTS todos")
	return err
}

// ========================================
// Display Helpers
// ========================================

func displayTodo(t Todo) {
	status := "[ ]"
	if t.Completed {
		status = "[x]"
	}
	age := time.Since(t.CreatedAt).Round(time.Minute)
	fmt.Printf("  #%-4d %s %s  (%s ago)\n", t.ID, status, t.Title, age)
}

func displayTodos(todos []Todo) {
	if len(todos) == 0 {
		fmt.Println("  No todos found.")
		return
	}
	for _, t := range todos {
		displayTodo(t)
	}
	fmt.Printf("  --- %d todo(s) ---\n", len(todos))
}

// ========================================
// Main — Interactive CLI Loop
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 10 Project: TODO App")
	fmt.Println("========================================")

	ctx := context.Background()

	// Connect to database
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/learngo?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer pool.Close()

	store := NewTodoStore(pool)

	// Setup the database schema
	if err := store.SetupSchema(ctx); err != nil {
		log.Fatalf("Failed to setup schema: %v\n", err)
	}

	fmt.Println("Connected to PostgreSQL!")
	fmt.Println()
	printHelp()

	// ========================================
	// Interactive Command Loop
	// ========================================
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\ntodo> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse command and arguments
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToLower(parts[0])
		arg := ""
		if len(parts) > 1 {
			arg = strings.TrimSpace(parts[1])
		}

		switch cmd {
		case "add":
			if arg == "" {
				fmt.Println("Usage: add <todo text>")
				continue
			}
			todo, err := store.Add(ctx, arg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Added todo #%d: %s\n", todo.ID, todo.Title)

		case "list", "ls":
			todos, err := store.List(ctx)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			displayTodos(todos)

		case "done":
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Println("Usage: done <id>")
				continue
			}
			todo, err := store.SetCompleted(ctx, id, true)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Completed: #%d %s\n", todo.ID, todo.Title)

		case "undone":
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Println("Usage: undone <id>")
				continue
			}
			todo, err := store.SetCompleted(ctx, id, false)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Reopened: #%d %s\n", todo.ID, todo.Title)

		case "edit":
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Println("Usage: edit <id>")
				continue
			}
			fmt.Print("New text: ")
			if !scanner.Scan() {
				break
			}
			newTitle := strings.TrimSpace(scanner.Text())
			if newTitle == "" {
				fmt.Println("Title cannot be empty.")
				continue
			}
			todo, err := store.Edit(ctx, id, newTitle)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Updated: #%d %s\n", todo.ID, todo.Title)

		case "delete", "rm":
			id, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Println("Usage: delete <id>")
				continue
			}
			if err := store.Delete(ctx, id); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Deleted todo #%d\n", id)

		case "search":
			if arg == "" {
				fmt.Println("Usage: search <text>")
				continue
			}
			todos, err := store.Search(ctx, arg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Results for %q:\n", arg)
			displayTodos(todos)

		case "clear":
			count, err := store.ClearCompleted(ctx)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Cleared %d completed todo(s)\n", count)

		case "stats":
			total, completed, pending, err := store.Stats(ctx)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("  Total:     %d\n", total)
			fmt.Printf("  Completed: %d\n", completed)
			fmt.Printf("  Pending:   %d\n", pending)
			if total > 0 {
				pct := float64(completed) / float64(total) * 100
				fmt.Printf("  Progress:  %.0f%%\n", pct)
			}

		case "help":
			printHelp()

		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmd)
		}
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  add <text>    - Add a new todo")
	fmt.Println("  list          - List all todos")
	fmt.Println("  done <id>     - Mark as complete")
	fmt.Println("  undone <id>   - Mark as incomplete")
	fmt.Println("  edit <id>     - Edit todo text")
	fmt.Println("  delete <id>   - Delete a todo")
	fmt.Println("  search <text> - Search todos")
	fmt.Println("  clear         - Remove completed todos")
	fmt.Println("  stats         - Show statistics")
	fmt.Println("  help          - Show this help")
	fmt.Println("  quit          - Exit")
}

// ========================================
// Sample Session
// ========================================
//
// todo> add Buy groceries
// Added todo #1: Buy groceries
//
// todo> add Learn Go testing
// Added todo #2: Learn Go testing
//
// todo> add Read Go blog
// Added todo #3: Read Go blog
//
// todo> list
//   #3   [ ] Read Go blog  (0m ago)
//   #2   [ ] Learn Go testing  (0m ago)
//   #1   [ ] Buy groceries  (1m ago)
//   --- 3 todo(s) ---
//
// todo> done 1
// Completed: #1 Buy groceries
//
// todo> stats
//   Total:     3
//   Completed: 1
//   Pending:   2
//   Progress:  33%
//
// todo> search Go
// Results for "Go":
//   #3   [ ] Read Go blog  (1m ago)
//   #2   [ ] Learn Go testing  (1m ago)
//   --- 2 todo(s) ---
//
// todo> quit
// Goodbye!
