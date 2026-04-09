package main

import "fmt"

func main() {
	// ========================================
	// Creating maps — key-value pairs
	// ========================================
	fmt.Println("=== Creating maps ===")

	// Map literal
	ages := map[string]int{
		"Alice": 30,
		"Bob":   25,
		"Carol": 35,
	}
	fmt.Println("Ages:", ages)

	// Empty map with make()
	scores := make(map[string]int)
	fmt.Println("Empty scores map:", scores)

	// A nil map — reads return zero values, but writing panics!
	var nilMap map[string]int
	fmt.Printf("Nil map: %v, is nil? %t\n", nilMap, nilMap == nil)
	// nilMap["key"] = 1 // This would PANIC! Always use make() before writing.

	// ========================================
	// Adding and reading keys
	// ========================================
	fmt.Println("\n=== Adding and reading ===")

	scores["math"] = 95
	scores["science"] = 88
	scores["english"] = 92
	fmt.Println("Scores:", scores)

	// Reading a key
	mathScore := scores["math"]
	fmt.Println("Math score:", mathScore)

	// Reading a missing key returns the zero value (no error!)
	historyScore := scores["history"]
	fmt.Println("History score:", historyScore) // 0 — but is it really 0 or missing?

	// ========================================
	// Comma-ok idiom — checking if a key exists
	// ========================================
	fmt.Println("\n=== Comma-ok idiom ===")

	// The second return value tells you if the key was found
	val, ok := scores["math"]
	fmt.Printf("math: value=%d, exists=%t\n", val, ok)

	val, ok = scores["history"]
	fmt.Printf("history: value=%d, exists=%t\n", val, ok)

	// Common pattern: check before using
	if score, exists := scores["science"]; exists {
		fmt.Println("Science score found:", score)
	} else {
		fmt.Println("Science score not found")
	}

	if _, exists := scores["art"]; !exists {
		fmt.Println("Art score does not exist")
	}

	// ========================================
	// Updating and deleting keys
	// ========================================
	fmt.Println("\n=== Updating and deleting ===")

	fmt.Println("Before:", scores)

	// Update: just assign to the same key
	scores["math"] = 98
	fmt.Println("After update math:", scores)

	// Delete: use the built-in delete() function
	delete(scores, "english")
	fmt.Println("After delete english:", scores)

	// Deleting a non-existent key is safe — no panic
	delete(scores, "nonexistent")
	fmt.Println("After deleting nonexistent key:", scores)

	// ========================================
	// Iterating over maps
	// ========================================
	fmt.Println("\n=== Iterating maps ===")

	capitals := map[string]string{
		"India":   "New Delhi",
		"Japan":   "Tokyo",
		"France":  "Paris",
		"Germany": "Berlin",
		"Brazil":  "Brasilia",
	}

	// Iterate over key-value pairs
	fmt.Println("Countries and capitals:")
	for country, capital := range capitals {
		fmt.Printf("  %s -> %s\n", country, capital)
	}

	// Iterate over keys only
	fmt.Println("\nJust the countries:")
	for country := range capitals {
		fmt.Printf("  %s\n", country)
	}

	// NOTE: Map iteration order is NOT guaranteed!
	// Running this multiple times may produce different orders.

	// ========================================
	// Maps of slices — one key, multiple values
	// ========================================
	fmt.Println("\n=== Maps of slices ===")

	// Track hobbies per person
	hobbies := map[string][]string{
		"Alice": {"reading", "hiking", "coding"},
		"Bob":   {"gaming", "cooking"},
	}

	// Add a hobby for Bob
	hobbies["Bob"] = append(hobbies["Bob"], "cycling")

	// Add a new person
	hobbies["Carol"] = []string{"painting", "yoga"}

	for person, hobbyList := range hobbies {
		fmt.Printf("  %s: %v\n", person, hobbyList)
	}

	// ========================================
	// Nested maps — maps of maps
	// ========================================
	fmt.Println("\n=== Nested maps ===")

	// Student grades by subject
	studentGrades := map[string]map[string]int{
		"Alice": {
			"math":    95,
			"science": 88,
			"english": 92,
		},
		"Bob": {
			"math":    78,
			"science": 85,
			"english": 90,
		},
	}

	// Access nested values
	fmt.Println("Alice's math grade:", studentGrades["Alice"]["math"])

	// Add a new student (must initialize the inner map!)
	studentGrades["Carol"] = map[string]int{
		"math":    91,
		"science": 96,
	}

	// Add a subject to an existing student
	studentGrades["Carol"]["english"] = 89

	// Print all grades
	for student, grades := range studentGrades {
		fmt.Printf("  %s: ", student)
		for subject, grade := range grades {
			fmt.Printf("%s=%d ", subject, grade)
		}
		fmt.Println()
	}

	// ========================================
	// Counting with maps — a very common pattern
	// ========================================
	fmt.Println("\n=== Counting with maps ===")

	// Count character frequencies
	text := "hello world"
	charCount := make(map[rune]int)
	for _, ch := range text {
		charCount[ch]++
	}

	fmt.Printf("Character frequencies in %q:\n", text)
	for ch, count := range charCount {
		fmt.Printf("  '%c': %d\n", ch, count)
	}

	// ========================================
	// Map as a set — using map[T]bool
	// ========================================
	fmt.Println("\n=== Map as a set ===")

	// Use a map to track unique items
	seen := make(map[string]bool)
	words := []string{"go", "is", "fun", "go", "is", "awesome", "fun"}

	var unique []string
	for _, word := range words {
		if !seen[word] {
			seen[word] = true
			unique = append(unique, word)
		}
	}
	fmt.Println("All words:", words)
	fmt.Println("Unique words:", unique)

	// Alternative: map[T]struct{} uses zero memory per value
	visited := make(map[string]struct{})
	visited["page1"] = struct{}{}
	visited["page2"] = struct{}{}
	if _, ok := visited["page1"]; ok {
		fmt.Println("page1 has been visited")
	}

	// ========================================
	// Maps are reference types
	// ========================================
	fmt.Println("\n=== Maps are reference types ===")

	original := map[string]int{"a": 1, "b": 2}
	alias := original // alias points to the SAME map
	alias["c"] = 3

	fmt.Println("Original:", original) // has "c" too!
	fmt.Println("Alias:   ", alias)

	// To make a true copy, you must copy each element
	copyMap := make(map[string]int)
	for k, v := range original {
		copyMap[k] = v
	}
	copyMap["d"] = 4
	fmt.Println("Original after copy modified:", original) // no "d"
	fmt.Println("Copy:", copyMap)
}
