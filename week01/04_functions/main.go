package main

import (
	"fmt"
	"math"
)

// ========================================
// 1. Basic function — takes inputs, returns nothing
// ========================================
func greet(name string) {
	fmt.Printf("Hello, %s! Welcome to Go.\n", name)
}

// ========================================
// 2. Function with a return value
// ========================================
func add(a int, b int) int {
	return a + b
}

// When parameters share a type, you can shorten it:
func multiply(a, b int) int {
	return a * b
}

// ========================================
// 3. Multiple return values — a Go superpower!
//    This is how Go handles errors (no exceptions).
// ========================================
func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("cannot divide by zero")
	}
	return a / b, nil // nil means "no error"
}

// ========================================
// 4. Named return values
//    The return variables are declared in the signature.
//    A bare "return" sends them back automatically.
// ========================================
func rectangleInfo(width, height float64) (area, perimeter float64) {
	area = width * height
	perimeter = 2 * (width + height)
	return // "naked return" — sends area and perimeter
}

// ========================================
// 5. Variadic functions — accept any number of arguments
//    The ... means "zero or more". nums becomes a slice.
// ========================================
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// ========================================
// 6. Functions as first-class citizens
//    You can pass functions as arguments!
// ========================================
func applyToEach(nums []int, transform func(int) int) []int {
	result := make([]int, len(nums))
	for i, n := range nums {
		result[i] = transform(n)
	}
	return result
}

func main() {
	// --- 1. Basic function ---
	fmt.Println("=== Basic Function ===")
	greet("Sri")

	// --- 2. Return values ---
	fmt.Println("\n=== Return Values ===")
	result := add(10, 20)
	fmt.Println("10 + 20 =", result)
	fmt.Println("6 * 7 =", multiply(6, 7))

	// --- 3. Multiple returns + error handling ---
	fmt.Println("\n=== Multiple Returns & Errors ===")
	quotient, err := divide(10, 3)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("10 / 3 = %.2f\n", quotient)
	}

	// Try dividing by zero
	quotient, err = divide(10, 0)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("10 / 0 = %.2f\n", quotient)
	}

	// --- 4. Named returns ---
	fmt.Println("\n=== Named Return Values ===")
	area, perimeter := rectangleInfo(5, 3)
	fmt.Printf("Rectangle 5x3 → area=%.0f, perimeter=%.0f\n", area, perimeter)

	// --- 5. Variadic functions ---
	fmt.Println("\n=== Variadic Function ===")
	fmt.Println("sum(1, 2, 3) =", sum(1, 2, 3))
	fmt.Println("sum(10, 20, 30, 40) =", sum(10, 20, 30, 40))
	fmt.Println("sum() =", sum()) // zero args is fine!

	// Pass a slice to a variadic function using ...
	numbers := []int{5, 10, 15}
	fmt.Println("sum(slice...) =", sum(numbers...))

	// --- 6. Functions as values ---
	fmt.Println("\n=== Functions as Values ===")

	// Anonymous function (lambda) — a function with no name
	double := func(n int) int {
		return n * 2
	}
	fmt.Println("double(5) =", double(5))

	// Pass a function to another function
	nums := []int{1, 2, 3, 4, 5}
	doubled := applyToEach(nums, double)
	fmt.Println("Original:", nums)
	fmt.Println("Doubled: ", doubled)

	// Inline anonymous function
	squared := applyToEach(nums, func(n int) int {
		return n * n
	})
	fmt.Println("Squared: ", squared)

	// --- 7. Closures — functions that "remember" variables ---
	fmt.Println("\n=== Closures ===")
	counter := makeCounter()
	fmt.Println("counter():", counter()) // 1
	fmt.Println("counter():", counter()) // 2
	fmt.Println("counter():", counter()) // 3
	// Each call remembers the previous count!

	// --- 8. Using standard library functions ---
	fmt.Println("\n=== Standard Library (math) ===")
	fmt.Println("sqrt(144) =", math.Sqrt(144))
	fmt.Println("pow(2, 10) =", math.Pow(2, 10))
	fmt.Println("max(3, 7) =", max(3, 7)) // built-in since Go 1.21
}

// ========================================
// 7. Closures — a function that captures outside variables
//    This factory function returns a new counter each time.
// ========================================
func makeCounter() func() int {
	count := 0 // this variable is "closed over"
	return func() int {
		count++
		return count
	}
}
