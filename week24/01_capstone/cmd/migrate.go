package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ========================================
// Migrate Command
// ========================================
// Handles database schema migrations. This demonstrates:
//   - Week 10: PostgreSQL operations with pgx
//   - Week 2-3: Error handling, file I/O
//   - Week 23: Cobra CLI patterns
//
// In production, you'd use a migration tool like golang-migrate
// or goose. This simplified implementation shows the core concept:
// track which migrations have run and execute new ones in order.
//
// Usage:
//   taskflow migrate up      # Run pending migrations
//   taskflow migrate status  # Show migration status

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long: `Manage database schema migrations.

Subcommands:
  up      Run all pending migrations
  status  Show which migrations have been applied

Examples:
  taskflow migrate up
  taskflow migrate status`,

	// Default action when no subcommand given
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	RunE:  runMigrateUp,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	RunE:  runMigrateStatus,
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
}

// buildDSN constructs a PostgreSQL connection string from Viper config.
func buildDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		viper.GetString("database.postgres.user"),
		viper.GetString("database.postgres.password"),
		viper.GetString("database.postgres.host"),
		viper.GetInt("database.postgres.port"),
		viper.GetString("database.postgres.dbname"),
		viper.GetString("database.postgres.sslmode"),
	)
}

// runMigrateUp executes all pending database migrations.
func runMigrateUp(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	slog.Info("connecting to PostgreSQL for migrations")

	// Connect to the database
	pool, err := pgxpool.New(ctx, buildDSN())
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	// Verify connectivity
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}
	slog.Info("database connection established")

	// ========================================
	// Ensure Migration Tracking Table Exists
	// ========================================
	// We need a table to track which migrations have been applied.
	// This is the bootstrap problem — the first thing the migrator
	// does is create its own tracking table.
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	// ========================================
	// Discover Migration Files
	// ========================================
	// Look for .sql files in the migrations directory.
	migrationsDir := findMigrationsDir()
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory %s: %w", migrationsDir, err)
	}

	// Filter and sort SQL files
	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	if len(migrationFiles) == 0 {
		fmt.Println("No migration files found.")
		return nil
	}

	// ========================================
	// Get Already-Applied Migrations
	// ========================================
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("querying applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scanning migration version: %w", err)
		}
		applied[version] = true
	}

	// ========================================
	// Run Pending Migrations
	// ========================================
	pending := 0
	for _, filename := range migrationFiles {
		if applied[filename] {
			continue
		}

		pending++
		slog.Info("applying migration", "file", filename)

		// Read the migration SQL
		sqlBytes, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", filename, err)
		}

		// Execute the migration in a transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("beginning transaction: %w", err)
		}

		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("executing migration %s: %w", filename, err)
		}

		// Record the migration
		if _, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			filename,
		); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("recording migration %s: %w", filename, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("committing migration %s: %w", filename, err)
		}

		fmt.Printf("  Applied: %s\n", filename)
	}

	if pending == 0 {
		fmt.Println("All migrations are up to date.")
	} else {
		fmt.Printf("\nApplied %d migration(s) successfully.\n", pending)
	}

	return nil
}

// runMigrateStatus shows which migrations have been applied.
func runMigrateStatus(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, buildDSN())
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	// Get applied migrations
	rows, err := pool.Query(ctx,
		"SELECT version, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		// Table might not exist yet
		fmt.Println("No migrations have been applied yet.")
		fmt.Println("Run 'taskflow migrate up' to apply migrations.")
		return nil
	}
	defer rows.Close()

	fmt.Println("========================================")
	fmt.Println("  Migration Status")
	fmt.Println("========================================")

	count := 0
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return fmt.Errorf("scanning row: %w", err)
		}
		fmt.Printf("  [x] %s (applied: %s)\n",
			version, appliedAt.Format("2006-01-02 15:04:05"))
		count++
	}

	if count == 0 {
		fmt.Println("  No migrations applied yet.")
	}

	return nil
}

// findMigrationsDir locates the migrations directory.
func findMigrationsDir() string {
	// Check common locations
	candidates := []string{
		"migrations",
		"01_capstone/migrations",
		"../migrations",
	}
	for _, dir := range candidates {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return "migrations" // Default
}
