package main

import "fmt"

func main() {
	// ========================================
	// ARRAYS — Fixed size, rarely used directly
	// ========================================
	fmt.Println("=== Arrays (fixed size) ===")

	// Declare an array of 5 integers
	var numbers [5]int
	numbers[0] = 10
	numbers[1] = 20
	numbers[2] = 30
	numbers[3] = 40
	numbers[4] = 50
	fmt.Println("Array:", numbers)
	fmt.Println("Length:", len(numbers))

	// Array literal — size is part of the type!
	colors := [3]string{"red", "green", "blue"}
	fmt.Println("Colors:", colors)

	// Let Go count the elements with [...]
	primes := [...]int{2, 3, 5, 7, 11}
	fmt.Println("Primes:", primes)

	// Important: arrays are VALUE types (copies on assignment)
	original := [3]int{1, 2, 3}
	copied := original
	copied[0] = 999
	fmt.Println("\nOriginal:", original) // [1 2 3] — unchanged!
	fmt.Println("Copied:  ", copied)    // [999 2 3]

	// ========================================
	// SLICES — Dynamic, the workhorse collection
	// ========================================
	fmt.Println("\n=== Slices (dynamic) ===")

	// Slice literal (no size specified — that's what makes it a slice)
	fruits := []string{"apple", "banana", "cherry"}
	fmt.Println("Fruits:", fruits)
	fmt.Printf("Length: %d, Capacity: %d\n", len(fruits), cap(fruits))

	// ========================================
	// make() — Create slices with specific length and capacity
	// ========================================
	fmt.Println("\n=== make() ===")

	// make([]Type, length, capacity)
	scores := make([]int, 3, 10)
	fmt.Println("Scores:", scores)                                     // [0 0 0]
	fmt.Printf("Length: %d, Capacity: %d\n", len(scores), cap(scores)) // 3, 10

	// make with just length (capacity = length)
	names := make([]string, 5)
	fmt.Printf("Names length: %d, capacity: %d\n", len(names), cap(names)) // 5, 5

	// ========================================
	// append() — Add elements to a slice
	// ========================================
	fmt.Println("\n=== append() ===")

	cities := []string{"Chennai", "Mumbai"}
	fmt.Printf("Before append: %v (len=%d, cap=%d)\n", cities, len(cities), cap(cities))

	// Append returns a NEW slice (may have a new underlying array)
	cities = append(cities, "Delhi")
	cities = append(cities, "Bangalore", "Kolkata") // append multiple
	fmt.Printf("After append:  %v (len=%d, cap=%d)\n", cities, len(cities), cap(cities))

	// Append one slice to another using ...
	moreCities := []string{"Hyderabad", "Pune"}
	cities = append(cities, moreCities...)
	fmt.Println("All cities:", cities)

	// ========================================
	// Slicing syntax — creating sub-slices
	// ========================================
	fmt.Println("\n=== Slicing syntax [low:high] ===")

	nums := []int{10, 20, 30, 40, 50, 60, 70}
	fmt.Println("Original:", nums)
	fmt.Println("nums[1:4]:", nums[1:4]) // [20 30 40] — index 1, 2, 3
	fmt.Println("nums[:3]: ", nums[:3])  // [10 20 30] — from start
	fmt.Println("nums[4:]: ", nums[4:])  // [50 60 70] — to end
	fmt.Println("nums[:]:  ", nums[:])   // all elements

	// IMPORTANT: slices share the same underlying array!
	subSlice := nums[2:5]
	fmt.Println("\nSub-slice:", subSlice) // [30 40 50]
	subSlice[0] = 999
	fmt.Println("After modifying sub-slice:")
	fmt.Println("  Sub-slice:", subSlice) // [999 40 50]
	fmt.Println("  Original: ", nums)     // [10 20 999 40 50 60 70] — also changed!

	// Full slice expression [low:high:max] controls capacity
	safe := nums[1:3:3] // length=2, capacity=2
	fmt.Printf("\nFull slice expr nums[1:3:3]: %v (len=%d, cap=%d)\n",
		safe, len(safe), cap(safe))

	// ========================================
	// copy() — Create an independent copy
	// ========================================
	fmt.Println("\n=== copy() ===")

	source := []int{1, 2, 3, 4, 5}
	dest := make([]int, len(source))
	copied2 := copy(dest, source) // returns number of elements copied
	fmt.Printf("Copied %d elements\n", copied2)
	fmt.Println("Source:", source)
	fmt.Println("Dest:  ", dest)

	// Modifying dest does NOT affect source
	dest[0] = 999
	fmt.Println("After modifying dest:")
	fmt.Println("  Source:", source) // unchanged
	fmt.Println("  Dest:  ", dest)

	// ========================================
	// Nil slices vs empty slices
	// ========================================
	fmt.Println("\n=== Nil slices vs empty slices ===")

	var nilSlice []int          // nil slice — no underlying array
	emptySlice := []int{}       // empty slice — has an array, but zero length
	madeSlice := make([]int, 0) // also empty

	fmt.Printf("nil slice:   %v, is nil? %t, len=%d\n", nilSlice, nilSlice == nil, len(nilSlice))
	fmt.Printf("empty slice: %v, is nil? %t, len=%d\n", emptySlice, emptySlice == nil, len(emptySlice))
	fmt.Printf("made slice:  %v, is nil? %t, len=%d\n", madeSlice, madeSlice == nil, len(madeSlice))

	// Both nil and empty slices work with append!
	nilSlice = append(nilSlice, 42)
	fmt.Println("After appending to nil slice:", nilSlice)

	// ========================================
	// Multi-dimensional slices (slice of slices)
	// ========================================
	fmt.Println("\n=== Multi-dimensional slices ===")

	// Create a 3x4 grid
	rows := 3
	cols := 4
	grid := make([][]int, rows)
	for i := range grid {
		grid[i] = make([]int, cols)
		for j := range grid[i] {
			grid[i][j] = i*cols + j + 1 // fill with sequential numbers
		}
	}

	fmt.Println("3x4 Grid:")
	for i, row := range grid {
		fmt.Printf("  Row %d: %v\n", i, row)
	}

	// Jagged slices — inner slices can have different lengths
	fmt.Println("\nJagged slice (triangle):")
	triangle := make([][]int, 4)
	for i := range triangle {
		triangle[i] = make([]int, i+1)
		for j := range triangle[i] {
			triangle[i][j] = j + 1
		}
	}
	for _, row := range triangle {
		fmt.Println(" ", row)
	}

	// ========================================
	// Common patterns: removing elements
	// ========================================
	fmt.Println("\n=== Removing elements from a slice ===")

	letters := []string{"a", "b", "c", "d", "e"}
	fmt.Println("Original:", letters)

	// Remove element at index 2 ("c") — order preserved
	indexToRemove := 2
	letters = append(letters[:indexToRemove], letters[indexToRemove+1:]...)
	fmt.Println("After removing index 2:", letters)

	// Remove without preserving order (swap with last, then truncate) — faster
	items := []string{"x", "y", "z", "w"}
	fmt.Println("\nItems:", items)
	removeIdx := 1 // remove "y"
	items[removeIdx] = items[len(items)-1]
	items = items[:len(items)-1]
	fmt.Println("After fast remove index 1:", items) // order changed

	// ========================================
	// Len vs Cap — understanding capacity growth
	// ========================================
	fmt.Println("\n=== Length vs Capacity growth ===")

	var growing []int
	fmt.Printf("Start: len=%d, cap=%d\n", len(growing), cap(growing))
	for i := 1; i <= 10; i++ {
		growing = append(growing, i)
		fmt.Printf("  After append %d: len=%d, cap=%d\n", i, len(growing), cap(growing))
	}
	fmt.Println("Notice: capacity roughly doubles when exceeded!")
}
