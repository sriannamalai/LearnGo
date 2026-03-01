package main

import "fmt"

func main() {
	// ========================================
	// IF / ELSE — no parentheses needed!
	// ========================================
	score := 85

	fmt.Println("=== If/Else ===")
	if score >= 90 {
		fmt.Println("Grade: A")
	} else if score >= 80 {
		fmt.Println("Grade: B")
	} else if score >= 70 {
		fmt.Println("Grade: C")
	} else {
		fmt.Println("Grade: F")
	}

	// If with a short statement (declare + check in one line)
	if num := 42; num%2 == 0 {
		fmt.Println(num, "is even")
	} else {
		fmt.Println(num, "is odd")
	}
	// Note: 'num' is NOT accessible here — it's scoped to the if block

	// ========================================
	// FOR LOOP — Go's only loop keyword!
	// ========================================
	fmt.Println("\n=== For Loop (classic) ===")
	for i := 1; i <= 5; i++ {
		fmt.Printf("  Count: %d\n", i)
	}

	// While-style loop (for with just a condition)
	fmt.Println("\n=== While-style loop ===")
	countdown := 3
	for countdown > 0 {
		fmt.Printf("  %d...\n", countdown)
		countdown--
	}
	fmt.Println("  Go!")

	// Looping over a range of values
	fmt.Println("\n=== Range loop ===")
	fruits := []string{"apple", "banana", "cherry"}
	for index, fruit := range fruits {
		fmt.Printf("  %d: %s\n", index, fruit)
	}

	// Skip the index with _
	fmt.Println("\n=== Range (skip index) ===")
	for _, fruit := range fruits {
		fmt.Println(" ", fruit)
	}

	// ========================================
	// SWITCH — cleaner than long if/else chains
	// ========================================
	fmt.Println("\n=== Switch ===")
	day := "Wednesday"

	switch day {
	case "Monday", "Tuesday", "Wednesday", "Thursday", "Friday":
		fmt.Println(day, "is a weekday")
	case "Saturday", "Sunday":
		fmt.Println(day, "is a weekend")
	default:
		fmt.Println("Unknown day")
	}

	// Switch without a value (acts like if/else)
	temperature := 28
	fmt.Println("\n=== Switch without value ===")
	switch {
	case temperature > 35:
		fmt.Println("It's very hot!")
	case temperature > 25:
		fmt.Println("It's warm and pleasant")
	case temperature > 15:
		fmt.Println("It's cool")
	default:
		fmt.Println("It's cold!")
	}
}
