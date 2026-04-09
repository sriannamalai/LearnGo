package main

import "fmt"

// ========================================
// Lesson 1: Functions in Go
// ========================================
// Functions are the building blocks of Go programs.
// They let you organize code into reusable, named pieces.
//
// Syntax:  func name(parameters) returnType { body }

// ========================================
// Basic function with no parameters, no return value
// ========================================
// This is the simplest kind of function — it just does something.
func greet() {
	fmt.Println("Hello from the greet() function!")
}

// ========================================
// Function with parameters
// ========================================
// Parameters let you pass data into a function.
// Each parameter has a name and a type.
func greetPerson(name string) {
	fmt.Printf("Hello, %s! Welcome to Go.\n", name)
}

// When consecutive parameters share a type, you can shorten it:
//   func add(a int, b int) → func add(a, b int)
func add(a, b int) {
	fmt.Printf("  %d + %d = %d\n", a, b, a+b)
}

// ========================================
// Function with a return value
// ========================================
// Use the return type after the parameter list.
// The function MUST return a value of that type.
func multiply(a, b int) int {
	return a * b
}

// ========================================
// Function with multiple parameter types
// ========================================
func formatPrice(item string, price float64) string {
	// Sprintf is like Printf but returns the string instead of printing it
	return fmt.Sprintf("%s: $%.2f", item, price)
}

// ========================================
// Functions are first-class citizens
// ========================================
// You can assign functions to variables and pass them around.
// This function takes another function as a parameter!
func applyOperation(a, b int, operation func(int, int) int) int {
	return operation(a, b)
}

// Some operations to pass around
func subtract(a, b int) int {
	return a - b
}

func multiplyNums(a, b int) int {
	return a * b
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Lesson: Functions in Go")
	fmt.Println("========================================")

	// ========================================
	// Calling basic functions
	// ========================================
	fmt.Println("\n--- Basic Function Calls ---")
	greet()
	greetPerson("Sri")
	greetPerson("Go Learner")

	// ========================================
	// Functions with parameters
	// ========================================
	fmt.Println("\n--- Functions with Parameters ---")
	add(3, 5)
	add(100, 200)

	// ========================================
	// Using return values
	// ========================================
	fmt.Println("\n--- Functions with Return Values ---")

	// The return value can be stored in a variable
	product := multiply(7, 8)
	fmt.Printf("  7 * 8 = %d\n", product)

	// Or used directly in an expression
	fmt.Printf("  3 * 4 * 5 = %d\n", multiply(3, multiply(4, 5)))

	// ========================================
	// String return values
	// ========================================
	fmt.Println("\n--- String Return Values ---")
	fmt.Println(" ", formatPrice("Coffee", 4.99))
	fmt.Println(" ", formatPrice("Go Book", 39.95))
	fmt.Println(" ", formatPrice("Laptop", 1299.00))

	// ========================================
	// Functions as values (first-class functions)
	// ========================================
	fmt.Println("\n--- Functions as Values ---")

	// Store a function in a variable
	var mathFunc func(int, int) int
	mathFunc = multiply
	fmt.Printf("  Using mathFunc (multiply): %d\n", mathFunc(6, 7))

	// Pass functions to other functions
	fmt.Printf("  applyOperation(10, 3, subtract): %d\n", applyOperation(10, 3, subtract))
	fmt.Printf("  applyOperation(10, 3, multiplyNums): %d\n", applyOperation(10, 3, multiplyNums))

	// ========================================
	// Anonymous functions (function literals)
	// ========================================
	fmt.Println("\n--- Anonymous Functions ---")

	// Define and call a function in one step
	result := func(x, y int) int {
		return x + y
	}(5, 10) // The (5, 10) immediately calls it
	fmt.Printf("  Anonymous add: %d\n", result)

	// Store an anonymous function for later use
	double := func(n int) int {
		return n * 2
	}
	fmt.Printf("  double(21) = %d\n", double(21))
	fmt.Printf("  double(50) = %d\n", double(50))

	// ========================================
	// Closures: functions that capture variables
	// ========================================
	fmt.Println("\n--- Closures ---")

	// A closure "closes over" variables from its surrounding scope.
	// The counter function returns a function that increments and
	// returns a count each time it's called.
	counter := makeCounter()
	fmt.Printf("  counter() = %d\n", counter())
	fmt.Printf("  counter() = %d\n", counter())
	fmt.Printf("  counter() = %d\n", counter())

	// Each call to makeCounter creates a new, independent counter
	anotherCounter := makeCounter()
	fmt.Printf("  anotherCounter() = %d\n", anotherCounter())
	fmt.Printf("  counter() still = %d\n", counter()) // continues from 3

	fmt.Println("\n========================================")
	fmt.Println("  Key Takeaways:")
	fmt.Println("  - Functions start with 'func' keyword")
	fmt.Println("  - Parameters: name first, then type")
	fmt.Println("  - Return type comes after the parameters")
	fmt.Println("  - Functions are values: assign, pass, return them")
	fmt.Println("  - Closures capture surrounding variables")
	fmt.Println("========================================")
}

// ========================================
// Closure example: a counter factory
// ========================================
// makeCounter returns a function. That function has access to
// the 'count' variable even after makeCounter has finished.
// This is a closure — the returned function "closes over" count.
func makeCounter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// Sample output:
//
// ========================================
//   Lesson: Functions in Go
// ========================================
//
// --- Basic Function Calls ---
// Hello from the greet() function!
// Hello, Sri! Welcome to Go.
// Hello, Go Learner! Welcome to Go.
//
// --- Functions with Parameters ---
//   3 + 5 = 8
//   100 + 200 = 300
//
// --- Functions with Return Values ---
//   7 * 8 = 56
//   3 * 4 * 5 = 60
//
// --- String Return Values ---
//   Coffee: $4.99
//   Go Book: $39.95
//   Laptop: $1299.00
//
// --- Functions as Values ---
//   Using mathFunc (multiply): 42
//   applyOperation(10, 3, subtract): 7
//   applyOperation(10, 3, multiplyNums): 30
//
// --- Anonymous Functions ---
//   Anonymous add: 15
//   double(21) = 42
//   double(50) = 100
//
// --- Closures ---
//   counter() = 1
//   counter() = 2
//   counter() = 3
//   anotherCounter() = 1
//   counter() still = 4
