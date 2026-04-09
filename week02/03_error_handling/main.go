package main

import (
	"errors"
	"fmt"
	"math"
	"strconv"
)

// ========================================
// Lesson 3: Error Handling in Go
// ========================================
// Go handles errors explicitly — no exceptions, no try/catch.
// Instead, functions return an error value alongside their result.
// This is one of Go's most important design decisions.
//
// The philosophy: errors are values, and you should handle them
// explicitly at the point they occur.

// ========================================
// The error interface
// ========================================
// The built-in error interface is simple:
//
//   type error interface {
//       Error() string
//   }
//
// Any type with an Error() string method satisfies this interface.

// ========================================
// Creating errors with errors.New
// ========================================
// errors.New creates a basic error from a string.
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("cannot divide by zero")
	}
	return a / b, nil // nil means "no error"
}

// ========================================
// Creating errors with fmt.Errorf
// ========================================
// fmt.Errorf lets you format error messages with context.
// This is more useful than errors.New when you need to include values.
func sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, fmt.Errorf("cannot compute square root of negative number: %.2f", n)
	}
	return math.Sqrt(n), nil
}

// ========================================
// Functions that parse and validate
// ========================================
func parseAge(input string) (int, error) {
	age, err := strconv.Atoi(input) // Atoi = ASCII to Integer
	if err != nil {
		return 0, fmt.Errorf("invalid age %q: %w", input, err)
		// %w wraps the original error — we'll explain this below
	}
	if age < 0 {
		return 0, fmt.Errorf("age cannot be negative: %d", age)
	}
	if age > 150 {
		return 0, fmt.Errorf("age seems unrealistic: %d", age)
	}
	return age, nil
}

// ========================================
// Custom error types
// ========================================
// Since error is just an interface, you can create your own
// error types with additional data.

// ValidationError carries the field name and a message
type ValidationError struct {
	Field   string
	Message string
}

// Implement the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %q: %s", e.Field, e.Message)
}

// Another custom error type
type InsufficientFundsError struct {
	Balance   float64
	Requested float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds: balance=$%.2f, requested=$%.2f",
		e.Balance, e.Requested)
}

// ========================================
// Using custom errors in functions
// ========================================
func validateUsername(username string) error {
	if len(username) == 0 {
		return &ValidationError{Field: "username", Message: "cannot be empty"}
	}
	if len(username) < 3 {
		return &ValidationError{Field: "username", Message: "must be at least 3 characters"}
	}
	if len(username) > 20 {
		return &ValidationError{Field: "username", Message: "must be 20 characters or less"}
	}
	return nil // nil = no error = success!
}

func withdraw(balance, amount float64) (float64, error) {
	if amount <= 0 {
		return balance, errors.New("withdrawal amount must be positive")
	}
	if amount > balance {
		return balance, &InsufficientFundsError{Balance: balance, Requested: amount}
	}
	return balance - amount, nil
}

// ========================================
// Sentinel errors (pre-defined error values)
// ========================================
// Sentinel errors are package-level variables that callers can
// compare against. This is a common Go pattern.
var (
	ErrNotFound   = errors.New("item not found")
	ErrOutOfStock = errors.New("item out of stock")
)

type Item struct {
	Name  string
	Stock int
}

func buyItem(inventory map[string]Item, name string) error {
	item, exists := inventory[name]
	if !exists {
		return fmt.Errorf("buying %q: %w", name, ErrNotFound)
	}
	if item.Stock <= 0 {
		return fmt.Errorf("buying %q: %w", name, ErrOutOfStock)
	}

	// Reduce stock
	item.Stock--
	inventory[name] = item
	fmt.Printf("    Purchased %q! Stock remaining: %d\n", name, item.Stock)
	return nil
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Lesson: Error Handling in Go")
	fmt.Println("========================================")

	// ========================================
	// Basic error checking pattern
	// ========================================
	fmt.Println("\n--- Basic Error Checking ---")

	// THE most common Go pattern: call, then check error
	result, err := divide(10, 3)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  10 / 3 = %.4f\n", result)
	}

	// Handle the error case
	result, err = divide(10, 0)
	if err != nil {
		fmt.Println("  10 / 0 -> Error:", err)
	} else {
		fmt.Printf("  10 / 0 = %.4f\n", result)
	}

	// ========================================
	// fmt.Errorf with context
	// ========================================
	fmt.Println("\n--- Errors with Context (fmt.Errorf) ---")

	val, err := sqrt(25)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  sqrt(25) = %.2f\n", val)
	}

	val, err = sqrt(-4)
	if err != nil {
		fmt.Println("  sqrt(-4) -> Error:", err)
	}

	// ========================================
	// Error wrapping with %w
	// ========================================
	fmt.Println("\n--- Error Wrapping ---")

	age, err := parseAge("25")
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Parsed age: %d\n", age)
	}

	// Invalid input — the original strconv error is wrapped inside
	_, err = parseAge("abc")
	if err != nil {
		fmt.Println("  parseAge(\"abc\") -> Error:", err)
	}

	_, err = parseAge("-5")
	if err != nil {
		fmt.Println("  parseAge(\"-5\") -> Error:", err)
	}

	_, err = parseAge("200")
	if err != nil {
		fmt.Println("  parseAge(\"200\") -> Error:", err)
	}

	// ========================================
	// Custom error types
	// ========================================
	fmt.Println("\n--- Custom Error Types ---")

	err = validateUsername("")
	if err != nil {
		fmt.Println("  validateUsername(\"\") ->", err)
	}

	err = validateUsername("ab")
	if err != nil {
		fmt.Println("  validateUsername(\"ab\") ->", err)
	}

	err = validateUsername("Sri")
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Println("  validateUsername(\"Sri\") -> OK!")
	}

	// ========================================
	// Type assertions on errors
	// ========================================
	fmt.Println("\n--- Type Assertions on Errors ---")

	// You can check if an error is a specific custom type
	// and extract additional information from it.
	balance := 100.0
	newBalance, err := withdraw(balance, 150)
	if err != nil {
		// Use errors.As to check if err (or any wrapped error) is
		// of a specific type
		var insuffErr *InsufficientFundsError
		if errors.As(err, &insuffErr) {
			// We can access the struct fields!
			fmt.Printf("  Withdrawal denied: have $%.2f, need $%.2f, short $%.2f\n",
				insuffErr.Balance, insuffErr.Requested,
				insuffErr.Requested-insuffErr.Balance)
		} else {
			fmt.Println("  Error:", err)
		}
	} else {
		fmt.Printf("  New balance: $%.2f\n", newBalance)
	}

	// Successful withdrawal
	newBalance, err = withdraw(balance, 30)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		fmt.Printf("  Withdrew $30 from $%.2f -> new balance: $%.2f\n", balance, newBalance)
	}

	// ========================================
	// Type assertion on ValidationError
	// ========================================
	fmt.Println("\n--- Checking Error Types ---")
	err = validateUsername("x")
	var valErr *ValidationError
	if errors.As(err, &valErr) {
		fmt.Printf("  Field: %s, Problem: %s\n", valErr.Field, valErr.Message)
	}

	// ========================================
	// Sentinel errors with errors.Is
	// ========================================
	fmt.Println("\n--- Sentinel Errors (errors.Is) ---")

	inventory := map[string]Item{
		"keyboard": {Name: "keyboard", Stock: 2},
		"mouse":    {Name: "mouse", Stock: 0},
	}

	// Successful purchase
	err = buyItem(inventory, "keyboard")
	if err != nil {
		fmt.Println("  Error:", err)
	}

	// Out of stock
	err = buyItem(inventory, "mouse")
	if err != nil {
		// errors.Is checks if err (or any wrapped error) matches a sentinel
		if errors.Is(err, ErrOutOfStock) {
			fmt.Println("  Mouse is out of stock — check back later!")
		} else {
			fmt.Println("  Error:", err)
		}
	}

	// Not found
	err = buyItem(inventory, "webcam")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			fmt.Println("  Webcam is not in our inventory.")
		} else {
			fmt.Println("  Error:", err)
		}
	}

	// ========================================
	// The error handling mantra
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("  Key Takeaways:")
	fmt.Println("  - Errors are values, not exceptions")
	fmt.Println("  - Always check: if err != nil { handle it }")
	fmt.Println("  - errors.New for simple errors")
	fmt.Println("  - fmt.Errorf for errors with context")
	fmt.Println("  - %w wraps errors (preserves the chain)")
	fmt.Println("  - errors.Is checks sentinel errors")
	fmt.Println("  - errors.As extracts custom error types")
	fmt.Println("  - nil means success (no error)")
	fmt.Println("========================================")
}
