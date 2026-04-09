package main

import "fmt"

func main() {
	// ========================================
	// Range over slices — index and value
	// ========================================
	fmt.Println("=== Range over slices ===")

	fruits := []string{"apple", "banana", "cherry", "date", "elderberry"}

	// Both index and value
	fmt.Println("Index and value:")
	for i, fruit := range fruits {
		fmt.Printf("  [%d] %s\n", i, fruit)
	}

	// ========================================
	// Using _ to skip index or value
	// ========================================
	fmt.Println("\n=== Skipping with _ ===")

	// Skip the index — just get values
	fmt.Println("Values only:")
	for _, fruit := range fruits {
		fmt.Println(" ", fruit)
	}

	// Skip the value — just get indices
	fmt.Println("Indices only:")
	for i := range fruits {
		fmt.Printf("  index %d\n", i)
	}

	// ========================================
	// Range over slices — modifying elements
	// ========================================
	fmt.Println("\n=== Range modifying elements ===")

	numbers := []int{1, 2, 3, 4, 5}
	fmt.Println("Before:", numbers)

	// The loop variable is a COPY — modifying it doesn't change the slice
	for _, num := range numbers {
		num *= 2 // This changes only the copy!
		_ = num
	}
	fmt.Println("After range (copy):", numbers) // unchanged!

	// To modify, use the index
	for i := range numbers {
		numbers[i] *= 2
	}
	fmt.Println("After range (index):", numbers) // doubled!

	// ========================================
	// Range over maps — key and value
	// ========================================
	fmt.Println("\n=== Range over maps ===")

	populations := map[string]int{
		"Tokyo":     14000000,
		"Delhi":     11000000,
		"Shanghai":  24000000,
		"Mumbai":    12000000,
		"São Paulo": 12300000,
	}

	// Key and value
	fmt.Println("City populations:")
	for city, pop := range populations {
		fmt.Printf("  %-12s %d\n", city, pop)
	}

	// Keys only
	fmt.Println("\nCities:")
	for city := range populations {
		fmt.Printf("  %s\n", city)
	}

	// Remember: map iteration order is randomized!
	fmt.Println("\n(Map order may differ each run)")

	// ========================================
	// Range over strings — rune iteration
	// ========================================
	fmt.Println("\n=== Range over strings (runes) ===")

	// Strings in Go are byte sequences, but range gives you runes (Unicode code points)
	greeting := "Hello, Go!"
	fmt.Printf("String: %q\n", greeting)
	fmt.Println("Byte-by-byte (using index):")
	for i := 0; i < len(greeting); i++ {
		fmt.Printf("  byte[%d] = %d (%c)\n", i, greeting[i], greeting[i])
	}

	fmt.Println("\nRune-by-rune (using range):")
	for i, ch := range greeting {
		fmt.Printf("  index %d: rune %U = '%c'\n", i, ch, ch)
	}

	// Unicode string — range handles multi-byte characters correctly
	fmt.Println("\n=== Unicode with range ===")
	emoji := "Go is fun! 🚀🎉"
	fmt.Printf("String: %s\n", emoji)
	fmt.Printf("Byte length: %d\n", len(emoji))

	runeCount := 0
	for i, ch := range emoji {
		fmt.Printf("  index %2d: '%c' (U+%04X)\n", i, ch, ch)
		runeCount++
	}
	fmt.Printf("Rune count: %d (vs byte length: %d)\n", runeCount, len(emoji))
	fmt.Println("Notice: multi-byte chars cause index gaps!")

	// Tamil text
	tamil := "வணக்கம்"
	fmt.Printf("\nTamil: %s\n", tamil)
	for i, ch := range tamil {
		fmt.Printf("  index %2d: '%c' (U+%04X)\n", i, ch, ch)
	}

	// ========================================
	// Range over integers (Go 1.22+)
	// ========================================
	fmt.Println("\n=== Range over integers (Go 1.22+) ===")

	fmt.Println("Count to 5:")
	for i := range 5 {
		fmt.Printf("  %d\n", i) // 0, 1, 2, 3, 4
	}

	// ========================================
	// Range with channels (preview)
	// ========================================
	fmt.Println("\n=== Range with channels (preview) ===")

	// Channels are covered in depth later — this is just a taste
	ch := make(chan string, 3)
	ch <- "first"
	ch <- "second"
	ch <- "third"
	close(ch) // Must close the channel for range to know when to stop

	// Range reads from a channel until it's closed
	for msg := range ch {
		fmt.Println(" ", msg)
	}
	fmt.Println("Channel drained!")

	// ========================================
	// Practical patterns with range
	// ========================================
	fmt.Println("\n=== Practical patterns ===")

	// Pattern 1: Filter elements
	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	var evens []int
	for _, n := range nums {
		if n%2 == 0 {
			evens = append(evens, n)
		}
	}
	fmt.Println("All:  ", nums)
	fmt.Println("Evens:", evens)

	// Pattern 2: Transform (map operation)
	words := []string{"hello", "world", "from", "go"}
	lengths := make([]int, len(words))
	for i, w := range words {
		lengths[i] = len(w)
	}
	fmt.Println("\nWords:  ", words)
	fmt.Println("Lengths:", lengths)

	// Pattern 3: Find element
	target := "world"
	found := false
	for i, w := range words {
		if w == target {
			fmt.Printf("\nFound %q at index %d\n", target, i)
			found = true
			break // stop searching once found
		}
	}
	if !found {
		fmt.Printf("\n%q not found\n", target)
	}

	// Pattern 4: Sum and average
	grades := []float64{85.5, 92.0, 78.5, 95.0, 88.5}
	sum := 0.0
	for _, g := range grades {
		sum += g
	}
	avg := sum / float64(len(grades))
	fmt.Printf("\nGrades: %v\n", grades)
	fmt.Printf("Sum: %.1f, Average: %.1f\n", sum, avg)
}
