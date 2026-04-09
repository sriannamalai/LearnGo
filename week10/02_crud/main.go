package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ========================================
// Week 10, Lesson 2: CRUD Operations
// ========================================
// Prerequisites:
//   1. PostgreSQL running with a "learngo" database
//   2. export DATABASE_URL="postgres://postgres:postgres@localhost:5432/learngo?sslmode=disable"
//   3. cd week10 && go mod tidy
//
// Run:
//   go run ./02_crud/

// ========================================
// Data Model
// ========================================

// Product represents a product in our database.
type Product struct {
	ID        int
	Name      string
	Price     float64
	Quantity  int
	CreatedAt time.Time
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 10: CRUD Operations")
	fmt.Println("========================================")

	// ========================================
	// Connect to the database
	// ========================================
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
	// 1. CREATE TABLE
	// ========================================
	fmt.Println("--- CREATE TABLE ---")
	createTable(ctx, pool)

	// ========================================
	// 2. INSERT (Create)
	// ========================================
	fmt.Println("\n--- INSERT ---")
	id1 := insertProduct(ctx, pool, "Mechanical Keyboard", 89.99, 50)
	id2 := insertProduct(ctx, pool, "USB-C Hub", 34.99, 100)
	id3 := insertProduct(ctx, pool, "Monitor Stand", 45.00, 30)
	fmt.Printf("Inserted products with IDs: %d, %d, %d\n", id1, id2, id3)

	// ========================================
	// 3. SELECT Single Row (Read)
	// ========================================
	fmt.Println("\n--- SELECT Single Row ---")
	selectOne(ctx, pool, id1)

	// ========================================
	// 4. SELECT Multiple Rows (Read)
	// ========================================
	fmt.Println("\n--- SELECT Multiple Rows ---")
	selectAll(ctx, pool)

	// ========================================
	// 5. SELECT with WHERE clause
	// ========================================
	fmt.Println("\n--- SELECT with WHERE ---")
	selectByPriceRange(ctx, pool, 30.0, 50.0)

	// ========================================
	// 6. UPDATE
	// ========================================
	fmt.Println("\n--- UPDATE ---")
	updateProduct(ctx, pool, id1, 99.99, 45)

	// ========================================
	// 7. DELETE
	// ========================================
	fmt.Println("\n--- DELETE ---")
	deleteProduct(ctx, pool, id3)

	// ========================================
	// 8. Verify final state
	// ========================================
	fmt.Println("\n--- Final State ---")
	selectAll(ctx, pool)

	// ========================================
	// 9. Cleanup
	// ========================================
	fmt.Println("\n--- Cleanup ---")
	dropTable(ctx, pool)
}

// ========================================
// CREATE TABLE
// ========================================

func createTable(ctx context.Context, pool *pgxpool.Pool) {
	// DROP TABLE IF EXISTS for clean re-runs
	_, err := pool.Exec(ctx, "DROP TABLE IF EXISTS products")
	if err != nil {
		log.Fatalf("Failed to drop table: %v\n", err)
	}

	// CREATE TABLE with various column types
	query := `
		CREATE TABLE products (
			id         SERIAL PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			price      DECIMAL(10, 2) NOT NULL,
			quantity   INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	// Exec is used for queries that don't return rows
	_, err = pool.Exec(ctx, query)
	if err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}

	fmt.Println("Table 'products' created successfully!")
}

// ========================================
// INSERT — Create
// ========================================

func insertProduct(ctx context.Context, pool *pgxpool.Pool, name string, price float64, quantity int) int {
	// ========================================
	// Parameter Binding with $1, $2, $3
	// ========================================
	// NEVER use string formatting to build SQL queries!
	// Always use parameter placeholders ($1, $2, ...) to prevent SQL injection.
	//
	// Bad:  fmt.Sprintf("INSERT INTO products (name) VALUES ('%s')", name) // SQL INJECTION!
	// Good: pool.QueryRow(ctx, "INSERT INTO products (name) VALUES ($1)", name)

	query := `
		INSERT INTO products (name, price, quantity)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	// RETURNING id gives us the auto-generated ID
	// QueryRow is used when we expect exactly one row back
	var id int
	err := pool.QueryRow(ctx, query, name, price, quantity).Scan(&id)
	if err != nil {
		log.Fatalf("Failed to insert product: %v\n", err)
	}

	fmt.Printf("Inserted: %s (price: $%.2f, qty: %d) -> ID: %d\n", name, price, quantity, id)
	return id
}

// ========================================
// SELECT Single Row — Read One
// ========================================

func selectOne(ctx context.Context, pool *pgxpool.Pool, id int) {
	query := `
		SELECT id, name, price, quantity, created_at
		FROM products
		WHERE id = $1
	`

	// QueryRow returns a single row
	// Scan reads columns into Go variables (must match column order!)
	var p Product
	err := pool.QueryRow(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Price,
		&p.Quantity,
		&p.CreatedAt,
	)
	if err != nil {
		log.Printf("Failed to find product %d: %v\n", id, err)
		return
	}

	fmt.Printf("Found: ID=%d, Name=%s, Price=$%.2f, Qty=%d, Created=%s\n",
		p.ID, p.Name, p.Price, p.Quantity, p.CreatedAt.Format("2006-01-02 15:04:05"))
}

// ========================================
// SELECT Multiple Rows — Read Many
// ========================================

func selectAll(ctx context.Context, pool *pgxpool.Pool) {
	query := `
		SELECT id, name, price, quantity, created_at
		FROM products
		ORDER BY id
	`

	// Query returns multiple rows — you must iterate with rows.Next()
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Failed to query products: %v\n", err)
	}
	// ALWAYS defer rows.Close() to release the connection back to the pool
	defer rows.Close()

	fmt.Printf("%-4s %-25s %10s %8s %s\n", "ID", "Name", "Price", "Qty", "Created")
	fmt.Println("---- ------------------------- ---------- -------- --------------------")

	count := 0
	for rows.Next() {
		var p Product
		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity, &p.CreatedAt)
		if err != nil {
			log.Printf("Failed to scan row: %v\n", err)
			continue
		}
		fmt.Printf("%-4d %-25s $%9.2f %8d %s\n",
			p.ID, p.Name, p.Price, p.Quantity, p.CreatedAt.Format("2006-01-02 15:04"))
		count++
	}

	// Check for errors that occurred during iteration
	if err := rows.Err(); err != nil {
		log.Printf("Row iteration error: %v\n", err)
	}

	fmt.Printf("Total: %d products\n", count)
}

// ========================================
// SELECT with WHERE — Filtered Read
// ========================================

func selectByPriceRange(ctx context.Context, pool *pgxpool.Pool, minPrice, maxPrice float64) {
	query := `
		SELECT id, name, price, quantity
		FROM products
		WHERE price BETWEEN $1 AND $2
		ORDER BY price
	`

	rows, err := pool.Query(ctx, query, minPrice, maxPrice)
	if err != nil {
		log.Fatalf("Failed to query by price range: %v\n", err)
	}
	defer rows.Close()

	fmt.Printf("Products priced between $%.2f and $%.2f:\n", minPrice, maxPrice)
	for rows.Next() {
		var id, qty int
		var name string
		var price float64
		if err := rows.Scan(&id, &name, &price, &qty); err != nil {
			log.Printf("Scan error: %v\n", err)
			continue
		}
		fmt.Printf("  [%d] %s - $%.2f (qty: %d)\n", id, name, price, qty)
	}
}

// ========================================
// UPDATE
// ========================================

func updateProduct(ctx context.Context, pool *pgxpool.Pool, id int, newPrice float64, newQty int) {
	query := `
		UPDATE products
		SET price = $1, quantity = $2
		WHERE id = $3
	`

	// Exec returns a command tag with the number of rows affected
	result, err := pool.Exec(ctx, query, newPrice, newQty, id)
	if err != nil {
		log.Fatalf("Failed to update product: %v\n", err)
	}

	// RowsAffected tells you how many rows were updated
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Printf("No product found with ID %d\n", id)
		return
	}

	fmt.Printf("Updated product %d: price=$%.2f, quantity=%d (%d row affected)\n",
		id, newPrice, newQty, rowsAffected)

	// Show the updated product
	selectOne(ctx, pool, id)
}

// ========================================
// DELETE
// ========================================

func deleteProduct(ctx context.Context, pool *pgxpool.Pool, id int) {
	query := `DELETE FROM products WHERE id = $1`

	result, err := pool.Exec(ctx, query, id)
	if err != nil {
		log.Fatalf("Failed to delete product: %v\n", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		fmt.Printf("No product found with ID %d\n", id)
		return
	}

	fmt.Printf("Deleted product %d (%d row affected)\n", id, rowsAffected)
}

// ========================================
// DROP TABLE (cleanup)
// ========================================

func dropTable(ctx context.Context, pool *pgxpool.Pool) {
	_, err := pool.Exec(ctx, "DROP TABLE IF EXISTS products")
	if err != nil {
		log.Printf("Failed to drop table: %v\n", err)
		return
	}
	fmt.Println("Table 'products' dropped. Clean slate!")
}

// ========================================
// Key Concepts Recap
// ========================================
//
// SQL Operations → pgx Methods:
//   CREATE/DROP/INSERT (no return)  → pool.Exec(ctx, sql, args...)
//   SELECT one row                  → pool.QueryRow(ctx, sql, args...).Scan(...)
//   SELECT multiple rows            → pool.Query(ctx, sql, args...)
//   INSERT with RETURNING           → pool.QueryRow(ctx, sql, args...).Scan(&id)
//
// Parameter Binding:
//   PostgreSQL uses $1, $2, $3... (NOT ? like MySQL)
//   pool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", userID)
//   NEVER use fmt.Sprintf for SQL — always use parameter binding!
//
// Row Iteration Pattern:
//   rows, err := pool.Query(ctx, sql)
//   defer rows.Close()              // Always close!
//   for rows.Next() {               // Iterate
//       err := rows.Scan(&col1, &col2)
//   }
//   err = rows.Err()                // Check for iteration errors
//
// Error Handling:
//   pgx.ErrNoRows — returned by QueryRow when no row matches
//   Use errors.Is(err, pgx.ErrNoRows) to check for "not found"
