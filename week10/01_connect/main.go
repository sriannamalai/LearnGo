package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ========================================
// Week 10, Lesson 1: Connecting to PostgreSQL
// ========================================
// Prerequisites:
//   1. PostgreSQL installed and running
//      - macOS: brew install postgresql@17 && brew services start postgresql@17
//      - Docker: docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:17
//   2. A database created:
//      - createdb learngo (or: psql -c "CREATE DATABASE learngo;")
//   3. Run: cd week10 && go mod tidy (to download pgx)
//
// Set the connection string:
//   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/learngo"
//
// Run:
//   go run .

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 10: Connecting to PostgreSQL")
	fmt.Println("========================================")

	// ========================================
	// 1. Connection String
	// ========================================
	// The connection string (DSN) tells pgx how to connect.
	// Format: postgres://user:password@host:port/dbname?sslmode=disable
	//
	// Best practice: read from environment variable, never hardcode!

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/learngo?sslmode=disable"
		fmt.Println("Using default DATABASE_URL (set env var to override)")
	}
	fmt.Printf("Connecting to: %s\n\n", databaseURL)

	// ========================================
	// 2. Simple Connection with pgx.Connect
	// ========================================
	// pgx.Connect creates a SINGLE database connection.
	// Good for simple scripts, but not for web servers
	// (use a pool for that — see section 4).

	fmt.Println("--- Single Connection ---")
	demonstrateSingleConnection(databaseURL)

	// ========================================
	// 3. Context with Timeout
	// ========================================
	// Always use context for database operations!
	// This prevents queries from hanging forever.

	fmt.Println("\n--- Connection with Timeout ---")
	demonstrateConnectionWithTimeout(databaseURL)

	// ========================================
	// 4. Connection Pool with pgxpool
	// ========================================
	// For web servers, use a connection POOL.
	// A pool maintains multiple connections and reuses them.
	// This is much more efficient than connecting/disconnecting per request.

	fmt.Println("\n--- Connection Pool ---")
	demonstrateConnectionPool(databaseURL)
}

// ========================================
// Single Connection
// ========================================

func demonstrateSingleConnection(databaseURL string) {
	// Create a context — this is Go's way of handling timeouts and cancellation.
	// context.Background() is the "root" context with no deadline.
	ctx := context.Background()

	// pgx.Connect establishes a single connection to PostgreSQL
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Printf("Failed to connect: %v\n", err)
		fmt.Println("Make sure PostgreSQL is running and the connection string is correct.")
		return
	}
	// ALWAYS defer Close to release the connection when done
	defer conn.Close(ctx)

	fmt.Println("Connected successfully!")

	// ========================================
	// Ping — verify the connection is alive
	// ========================================
	err = conn.Ping(ctx)
	if err != nil {
		log.Printf("Ping failed: %v\n", err)
		return
	}
	fmt.Println("Ping successful — database is reachable!")

	// ========================================
	// Simple query to verify everything works
	// ========================================
	var greeting string
	err = conn.QueryRow(ctx, "SELECT 'Hello from PostgreSQL!'").Scan(&greeting)
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	fmt.Printf("Database says: %s\n", greeting)

	// Get the PostgreSQL version
	var version string
	err = conn.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		log.Printf("Version query failed: %v\n", err)
		return
	}
	fmt.Printf("PostgreSQL version: %s\n", version)

	// Get the current database name
	var dbName string
	err = conn.QueryRow(ctx, "SELECT current_database()").Scan(&dbName)
	if err != nil {
		log.Printf("Database name query failed: %v\n", err)
		return
	}
	fmt.Printf("Current database: %s\n", dbName)

	// Get the current user
	var currentUser string
	err = conn.QueryRow(ctx, "SELECT current_user").Scan(&currentUser)
	if err != nil {
		log.Printf("User query failed: %v\n", err)
		return
	}
	fmt.Printf("Connected as: %s\n", currentUser)
}

// ========================================
// Connection with Timeout
// ========================================

func demonstrateConnectionWithTimeout(databaseURL string) {
	// Create a context with a 5-second timeout.
	// If the connection takes longer than 5 seconds, it's cancelled.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Always call cancel to release resources

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Printf("Connection with timeout failed: %v\n", err)
		return
	}
	defer conn.Close(ctx)

	fmt.Println("Connected with 5-second timeout!")

	// You can also set timeouts per query
	queryCtx, queryCancel := context.WithTimeout(ctx, 2*time.Second)
	defer queryCancel()

	var now time.Time
	err = conn.QueryRow(queryCtx, "SELECT NOW()").Scan(&now)
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	fmt.Printf("Server time: %s\n", now.Format(time.RFC3339))
}

// ========================================
// Connection Pool
// ========================================

func demonstrateConnectionPool(databaseURL string) {
	ctx := context.Background()

	// ========================================
	// Configure the pool
	// ========================================
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Printf("Failed to parse config: %v\n", err)
		return
	}

	// Pool configuration options
	config.MaxConns = 10                       // Maximum connections in the pool
	config.MinConns = 2                        // Minimum idle connections to keep
	config.MaxConnLifetime = 30 * time.Minute  // Max time a connection can be reused
	config.MaxConnIdleTime = 5 * time.Minute   // Max time an idle connection is kept
	config.HealthCheckPeriod = 1 * time.Minute // How often to check connection health

	fmt.Printf("Pool config: max=%d, min=%d\n", config.MaxConns, config.MinConns)

	// ========================================
	// Create the pool
	// ========================================
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Failed to create pool: %v\n", err)
		return
	}
	// ALWAYS defer Close — this closes ALL connections in the pool
	defer pool.Close()

	fmt.Println("Connection pool created!")

	// ========================================
	// Use the pool
	// ========================================
	// You don't need to "get" a connection — just call methods on the pool.
	// The pool automatically manages connections for you.

	err = pool.Ping(ctx)
	if err != nil {
		log.Printf("Pool ping failed: %v\n", err)
		return
	}
	fmt.Println("Pool ping successful!")

	// Query using the pool (same API as a single connection)
	var result int
	err = pool.QueryRow(ctx, "SELECT 1 + 1").Scan(&result)
	if err != nil {
		log.Printf("Pool query failed: %v\n", err)
		return
	}
	fmt.Printf("1 + 1 = %d (from pool)\n", result)

	// ========================================
	// Pool statistics
	// ========================================
	stat := pool.Stat()
	fmt.Printf("\nPool Statistics:\n")
	fmt.Printf("  Total connections:    %d\n", stat.TotalConns())
	fmt.Printf("  Idle connections:     %d\n", stat.IdleConns())
	fmt.Printf("  Acquired connections: %d\n", stat.AcquiredConns())
	fmt.Printf("  Max connections:      %d\n", stat.MaxConns())

	fmt.Println("\nDone! Connection pool closed on exit (via defer).")
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Connection Types:
//   pgx.Connect()       — single connection (scripts, simple tools)
//   pgxpool.New()       — connection pool (web servers, concurrent apps)
//
// Always use:
//   defer conn.Close(ctx)  — release single connection
//   defer pool.Close()     — release all pooled connections
//
// Context:
//   context.Background()                  — no timeout (use sparingly)
//   context.WithTimeout(ctx, 5*time.Second) — 5-second timeout
//   context.WithCancel(ctx)                — manually cancellable
//
// Connection String Format:
//   postgres://user:password@host:port/dbname?sslmode=disable
//
// Pool Configuration:
//   MaxConns         — maximum connections (default: 4)
//   MinConns         — minimum idle connections
//   MaxConnLifetime  — connection recycle time
//   MaxConnIdleTime  — idle connection timeout
//
// Environment variable pattern:
//   os.Getenv("DATABASE_URL")
