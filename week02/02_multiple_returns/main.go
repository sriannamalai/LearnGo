package main

import (
	"fmt"
	"math"
	"strings"
)

// ========================================
// Lesson 2: Multiple Return Values
// ========================================
// Go functions can return multiple values. This is one of Go's
// most distinctive features and is used everywhere — especially
// for returning results alongside errors.

// ========================================
// Basic multiple return values
// ========================================
// A function can return two (or more) values by listing
// the types in parentheses after the parameter list.
func divide(a, b float64) (float64, string) {
	if b == 0 {
		return 0, "error: division by zero"
	}
	return a / b, ""
}

// ========================================
// Swap — a classic example
// ========================================
// Swapping two values is trivial with multiple returns.
// No need for a temporary variable!
func swap(a, b string) (string, string) {
	return b, a
}

// ========================================
// Returning a result and a boolean (the "comma ok" pattern)
// ========================================
// This is a very common Go pattern. The second return value
// indicates whether the operation succeeded.
func findElement(slice []string, target string) (int, bool) {
	for i, v := range slice {
		if v == target {
			return i, true // found it!
		}
	}
	return -1, false // not found
}

// ========================================
// Named return values
// ========================================
// You can name your return values. They act as variables
// declared at the top of the function. A bare "return"
// statement returns their current values (called a "naked return").
func rectangleInfo(width, height float64) (area float64, perimeter float64) {
	area = width * height
	perimeter = 2 * (width + height)
	return // naked return — returns area and perimeter
}

// Named returns also serve as documentation — the names tell
// the caller what each value means.
func circleInfo(radius float64) (area float64, circumference float64) {
	area = math.Pi * radius * radius
	circumference = 2 * math.Pi * radius
	return
}

// ========================================
// Variadic functions: variable number of arguments
// ========================================
// The ... (ellipsis) before the type means "zero or more of this type".
// Inside the function, nums is a []int (a slice of ints).
func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// Variadic functions can have regular parameters too,
// but the variadic parameter must be last.
func joinWithSeparator(sep string, words ...string) string {
	return strings.Join(words, sep)
}

// ========================================
// Combining multiple returns with variadic parameters
// ========================================
func minMax(nums ...int) (min, max int) {
	if len(nums) == 0 {
		return 0, 0
	}

	min = nums[0]
	max = nums[0]

	for _, n := range nums[1:] {
		if n < min {
			min = n
		}
		if n > max {
			max = n
		}
	}
	return // naked return of named values
}

// ========================================
// Stats example: multiple calculations at once
// ========================================
func stats(numbers ...float64) (count int, total float64, average float64) {
	count = len(numbers)
	if count == 0 {
		return // all zero values
	}

	for _, n := range numbers {
		total += n
	}
	average = total / float64(count)
	return
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Lesson: Multiple Returns & Variadic")
	fmt.Println("========================================")

	// ========================================
	// Using multiple return values
	// ========================================
	fmt.Println("\n--- Multiple Return Values ---")

	result, errMsg := divide(10, 3)
	if errMsg != "" {
		fmt.Println("  Error:", errMsg)
	} else {
		fmt.Printf("  10 / 3 = %.4f\n", result)
	}

	result, errMsg = divide(10, 0)
	if errMsg != "" {
		fmt.Println("  10 / 0 =", errMsg)
	}

	// ========================================
	// Swap example
	// ========================================
	fmt.Println("\n--- Swap ---")
	first, second := swap("hello", "world")
	fmt.Printf("  swap(\"hello\", \"world\") = %q, %q\n", first, second)

	// ========================================
	// The "comma ok" pattern
	// ========================================
	fmt.Println("\n--- Comma-Ok Pattern ---")
	fruits := []string{"apple", "banana", "cherry", "date"}

	index, found := findElement(fruits, "cherry")
	if found {
		fmt.Printf("  Found \"cherry\" at index %d\n", index)
	}

	index, found = findElement(fruits, "grape")
	if !found {
		fmt.Printf("  \"grape\" not found (index=%d, found=%v)\n", index, found)
	}

	// ========================================
	// Discarding return values with _
	// ========================================
	fmt.Println("\n--- Discarding Values with _ ---")
	// If you only need one of the return values, use _ for the rest.
	// Go requires you to use all declared variables, so _ is the way
	// to say "I don't need this one."
	_, found2 := findElement(fruits, "banana")
	fmt.Printf("  Is banana in the list? %v\n", found2)

	// ========================================
	// Named return values
	// ========================================
	fmt.Println("\n--- Named Return Values ---")

	area, perimeter := rectangleInfo(5, 3)
	fmt.Printf("  Rectangle 5x3: area=%.1f, perimeter=%.1f\n", area, perimeter)

	cArea, circumference := circleInfo(10)
	fmt.Printf("  Circle r=10: area=%.2f, circumference=%.2f\n", cArea, circumference)

	// ========================================
	// Variadic functions
	// ========================================
	fmt.Println("\n--- Variadic Functions ---")

	// Call with any number of arguments
	fmt.Printf("  sum() = %d\n", sum())
	fmt.Printf("  sum(1) = %d\n", sum(1))
	fmt.Printf("  sum(1, 2, 3) = %d\n", sum(1, 2, 3))
	fmt.Printf("  sum(1, 2, 3, 4, 5) = %d\n", sum(1, 2, 3, 4, 5))

	// Pass a slice to a variadic function using ...
	numbers := []int{10, 20, 30, 40, 50}
	fmt.Printf("  sum(slice...) = %d\n", sum(numbers...))

	// ========================================
	// Variadic with regular parameters
	// ========================================
	fmt.Println("\n--- Mixed Parameters ---")
	fmt.Printf("  %s\n", joinWithSeparator(", ", "Go", "is", "awesome"))
	fmt.Printf("  %s\n", joinWithSeparator(" -> ", "learn", "practice", "build"))
	fmt.Printf("  %s\n", joinWithSeparator(" | ", "Mon", "Tue", "Wed", "Thu", "Fri"))

	// ========================================
	// Combining concepts: minMax with variadic
	// ========================================
	fmt.Println("\n--- MinMax (variadic + multiple returns) ---")
	min, max := minMax(38, 12, 95, 7, 63, 41)
	fmt.Printf("  min=%d, max=%d\n", min, max)

	min, max = minMax(5, 5, 5)
	fmt.Printf("  min=%d, max=%d (all same)\n", min, max)

	// ========================================
	// Stats example
	// ========================================
	fmt.Println("\n--- Stats (three return values) ---")
	count, total, avg := stats(85.5, 92.0, 78.5, 95.0, 88.0)
	fmt.Printf("  Scores: count=%d, total=%.1f, average=%.1f\n", count, total, avg)

	count, total, avg = stats()
	fmt.Printf("  Empty: count=%d, total=%.1f, average=%.1f\n", count, total, avg)

	fmt.Println("\n========================================")
	fmt.Println("  Key Takeaways:")
	fmt.Println("  - Return multiple values: (type1, type2)")
	fmt.Println("  - Named returns document what you return")
	fmt.Println("  - Use _ to discard unneeded return values")
	fmt.Println("  - Variadic: func f(args ...T) — args is []T")
	fmt.Println("  - Pass slices with slice... syntax")
	fmt.Println("========================================")
}
