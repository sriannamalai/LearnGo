package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ========================================
// Week 10, Lesson 3: Database Migrations
// ========================================
// Migrations are versioned SQL scripts that evolve your database schema.
// Each migration runs once, in order, and is tracked so it's never re-run.
//
// Prerequisites:
//   1. PostgreSQL running with "learngo" database
//   2. export DATABASE_URL="postgres://postgres:postgres@localhost:5432/learngo?sslmode=disable"
//   3. Migration SQL files in ./migrations/ directory
//
// Run:
//   go run ./03_migrations/
//
// Migration files live in: week10/03_migrations/migrations/
//   001_create_users.sql
//   002_add_email.sql
//
// The program will:
//   1. Create a migration_version table (if it doesn't exist)
//   2. Check which migrations have already been applied
//   3. Run any new migrations in order
//   4. Record each migration in the tracking table

// ========================================
// Migration Tracking
// ========================================

// Migration represents a single migration file.
type Migration struct {
	Version  int
	Name     string
	Filename string
	SQL      string
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 10: Database Migrations")
	fmt.Println("========================================")

	ctx := context.Background()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/learngo?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer pool.Close()

	fmt.Println("Connected to PostgreSQL!\n")

	// ========================================
	// Step 1: Ensure migration tracking table exists
	// ========================================
	fmt.Println("--- Step 1: Initialize Migration Tracking ---")
	if err := createMigrationTable(ctx, pool); err != nil {
		log.Fatalf("Failed to create migration table: %v\n", err)
	}

	// ========================================
	// Step 2: Load migration files from disk
	// ========================================
	fmt.Println("\n--- Step 2: Load Migration Files ---")
	migrationsDir := filepath.Join("03_migrations", "migrations")
	migrations, err := loadMigrations(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to load migrations: %v\n", err)
	}

	for _, m := range migrations {
		fmt.Printf("  Found: %s (version %d)\n", m.Filename, m.Version)
	}

	// ========================================
	// Step 3: Get current migration version
	// ========================================
	fmt.Println("\n--- Step 3: Check Current Version ---")
	currentVersion, err := getCurrentVersion(ctx, pool)
	if err != nil {
		log.Fatalf("Failed to get current version: %v\n", err)
	}
	fmt.Printf("Current migration version: %d\n", currentVersion)

	// ========================================
	// Step 4: Run pending migrations
	// ========================================
	fmt.Println("\n--- Step 4: Run Pending Migrations ---")
	applied := 0
	for _, m := range migrations {
		if m.Version <= currentVersion {
			fmt.Printf("  SKIP: %s (already applied)\n", m.Filename)
			continue
		}

		fmt.Printf("  APPLYING: %s ...\n", m.Filename)
		if err := applyMigration(ctx, pool, m); err != nil {
			log.Fatalf("Migration %s failed: %v\n", m.Filename, err)
		}
		fmt.Printf("  DONE: %s applied successfully\n", m.Filename)
		applied++
	}

	if applied == 0 {
		fmt.Println("  No pending migrations — database is up to date!")
	} else {
		fmt.Printf("\n  Applied %d migration(s)\n", applied)
	}

	// ========================================
	// Step 5: Show migration history
	// ========================================
	fmt.Println("\n--- Migration History ---")
	showMigrationHistory(ctx, pool)

	// ========================================
	// Step 6: Verify the schema
	// ========================================
	fmt.Println("\n--- Verify Schema ---")
	verifySchema(ctx, pool)

	// ========================================
	// Cleanup (optional — comment out to keep tables)
	// ========================================
	fmt.Println("\n--- Cleanup ---")
	cleanup(ctx, pool)
}

// ========================================
// Create Migration Tracking Table
// ========================================

func createMigrationTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS migration_version (
			version    INTEGER PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	_, err := pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	fmt.Println("Migration tracking table ready.")
	return nil
}

// ========================================
// Load Migrations from SQL Files
// ========================================

func loadMigrations(dir string) ([]Migration, error) {
	// Read all .sql files from the migrations directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %s: %w", dir, err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Parse version number from filename (e.g., "001_create_users.sql" -> 1)
		filename := entry.Name()
		var version int
		_, err := fmt.Sscanf(filename, "%d_", &version)
		if err != nil {
			return nil, fmt.Errorf("invalid migration filename %s: must start with a number", filename)
		}

		// Read the SQL content
		content, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filename, err)
		}

		// Extract a human-readable name from the filename
		name := strings.TrimSuffix(filename, ".sql")

		migrations = append(migrations, Migration{
			Version:  version,
			Name:     name,
			Filename: filename,
			SQL:      string(content),
		})
	}

	// Sort by version number
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// ========================================
// Get Current Migration Version
// ========================================

func getCurrentVersion(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var version int
	err := pool.QueryRow(ctx,
		"SELECT COALESCE(MAX(version), 0) FROM migration_version",
	).Scan(&version)
	return version, err
}

// ========================================
// Apply a Single Migration (in a transaction)
// ========================================

func applyMigration(ctx context.Context, pool *pgxpool.Pool, m Migration) error {
	// ========================================
	// Use a Transaction!
	// ========================================
	// Transactions ensure that either ALL statements in the migration
	// succeed, or NONE of them do. This prevents half-applied migrations.

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	// If anything fails, rollback the transaction
	defer tx.Rollback(ctx)

	// Execute the migration SQL
	_, err = tx.Exec(ctx, m.SQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration in the tracking table
	_, err = tx.Exec(ctx,
		"INSERT INTO migration_version (version, name) VALUES ($1, $2)",
		m.Version, m.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit the transaction — makes all changes permanent
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ========================================
// Show Migration History
// ========================================

func showMigrationHistory(ctx context.Context, pool *pgxpool.Pool) {
	rows, err := pool.Query(ctx,
		"SELECT version, name, applied_at FROM migration_version ORDER BY version",
	)
	if err != nil {
		log.Printf("Failed to query history: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Printf("%-8s %-30s %s\n", "Version", "Name", "Applied At")
	fmt.Println("-------- ------------------------------ -------------------------")

	for rows.Next() {
		var version int
		var name string
		var appliedAt time.Time
		if err := rows.Scan(&version, &name, &appliedAt); err != nil {
			log.Printf("Scan error: %v\n", err)
			continue
		}
		fmt.Printf("%-8d %-30s %s\n", version, name, appliedAt.Format("2006-01-02 15:04:05"))
	}
}

// ========================================
// Verify Schema
// ========================================

func verifySchema(ctx context.Context, pool *pgxpool.Pool) {
	// Check what tables exist (excluding system tables)
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		ORDER BY table_name
	`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Printf("Failed to query schema: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Tables in public schema:")
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		fmt.Printf("  - %s\n", tableName)
	}

	// Check columns in the users table
	colQuery := `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_name = 'users'
		ORDER BY ordinal_position
	`

	colRows, err := pool.Query(ctx, colQuery)
	if err != nil {
		log.Printf("Failed to query columns: %v\n", err)
		return
	}
	defer colRows.Close()

	fmt.Println("\nColumns in 'users' table:")
	fmt.Printf("  %-20s %-20s %s\n", "Column", "Type", "Nullable")
	fmt.Println("  -------------------- -------------------- --------")
	for colRows.Next() {
		var name, dataType, nullable string
		colRows.Scan(&name, &dataType, &nullable)
		fmt.Printf("  %-20s %-20s %s\n", name, dataType, nullable)
	}
}

// ========================================
// Cleanup
// ========================================

func cleanup(ctx context.Context, pool *pgxpool.Pool) {
	// Drop tables created by migrations (in reverse dependency order)
	tables := []string{"users", "migration_version"}
	for _, table := range tables {
		_, err := pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			log.Printf("Failed to drop %s: %v\n", table, err)
		} else {
			fmt.Printf("Dropped table: %s\n", table)
		}
	}
	fmt.Println("Cleanup complete!")
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Why Migrations?
//   - Track schema changes over time (like git for your database)
//   - Apply changes consistently across environments (dev, staging, prod)
//   - Each migration runs exactly once, in version order
//
// Migration Pattern:
//   1. Create numbered SQL files: 001_create_users.sql, 002_add_email.sql
//   2. Track applied versions in a migration_version table
//   3. On startup, run any migrations newer than the current version
//   4. Wrap each migration in a transaction for atomicity
//
// Transactions:
//   tx, err := pool.Begin(ctx)    // Start transaction
//   defer tx.Rollback(ctx)        // Rollback if not committed
//   tx.Exec(ctx, sql, args...)    // Execute within transaction
//   tx.Commit(ctx)                // Commit all changes
//
// Production Tips:
//   - Use a proper migration tool (golang-migrate, goose, atlas)
//   - Always test migrations on a copy of production data
//   - Keep migrations small and focused
//   - Never modify an already-applied migration
