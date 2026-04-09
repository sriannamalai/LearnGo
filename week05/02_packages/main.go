package main

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	// ========================================
	// Package Concepts in Go
	// ========================================
	fmt.Println("========================================")
	fmt.Println("  Packages in Go")
	fmt.Println("========================================")

	// ========================================
	// Exported vs Unexported identifiers
	// ========================================
	fmt.Println("\n=== Exported vs Unexported ===")

	fmt.Println(`
  In Go, visibility is determined by the first letter:
  - Uppercase = Exported (public)   -> fmt.Println, math.Pi
  - Lowercase = Unexported (private) -> only within same package

  Examples:
    fmt.Println   -> "Println" starts with uppercase = exported
    math.Pi       -> "Pi" starts with uppercase = exported
    math.pi       -> would be unexported (can't access from outside)

  This applies to:
    - Functions
    - Types (structs, interfaces)
    - Struct fields
    - Variables and constants
    - Methods`)

	// Demonstrate with a local type
	p := ExportedPerson{
		Name: "Alice", // exported field — accessible
		Age:  30,      // exported field — accessible
		// email is unexported — set via constructor
	}
	fmt.Printf("\n  Person: %+v\n", p)
	fmt.Println("  (The 'email' field is unexported — can't set directly from outside)")

	// ========================================
	// The init() function
	// ========================================
	fmt.Println("\n=== The init() function ===")

	fmt.Println(`
  Every package can have an init() function:
  - Runs automatically BEFORE main()
  - Used for initialization tasks
  - Can have multiple init() functions per file
  - Cannot be called manually
  - Order: imported packages' init() -> this package's init() -> main()`)

	// The init() function at the bottom of this file already ran
	fmt.Println("  The init() function at the bottom already ran before main()!")

	// ========================================
	// strings package — string manipulation
	// ========================================
	fmt.Println("\n=== strings package ===")

	s := "Hello, Go World!"

	fmt.Println("Original:    ", s)
	fmt.Println("ToUpper:     ", strings.ToUpper(s))
	fmt.Println("ToLower:     ", strings.ToLower(s))
	fmt.Println("Contains Go?:", strings.Contains(s, "Go"))
	fmt.Println("HasPrefix:   ", strings.HasPrefix(s, "Hello"))
	fmt.Println("HasSuffix:   ", strings.HasSuffix(s, "World!"))
	fmt.Println("Count 'o':   ", strings.Count(s, "o"))
	fmt.Println("Index of Go: ", strings.Index(s, "Go"))
	fmt.Println("Replace:     ", strings.Replace(s, "Go", "Golang", 1))
	fmt.Println("ReplaceAll:  ", strings.ReplaceAll(s, "o", "0"))
	fmt.Println("Trim:        ", strings.TrimSpace("  spaces  "))
	fmt.Println("TrimLeft:    ", strings.TrimLeft("xxhello", "x"))

	// Split and Join
	csv := "apple,banana,cherry,date"
	parts := strings.Split(csv, ",")
	fmt.Println("\nSplit CSV:   ", parts)
	fmt.Println("Join with |: ", strings.Join(parts, " | "))

	// Fields splits on whitespace
	sentence := "  hello   world   from   go  "
	words := strings.Fields(sentence)
	fmt.Println("Fields:      ", words) // clean split

	// Builder for efficient string concatenation
	var builder strings.Builder
	for i := 0; i < 5; i++ {
		builder.WriteString(fmt.Sprintf("item%d ", i))
	}
	fmt.Println("Builder:     ", builder.String())

	// Repeat
	fmt.Println("Repeat:      ", strings.Repeat("Go! ", 3))

	// Map function — transform each rune
	shouted := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return unicode.ToUpper(r)
		}
		return r
	}, "hello, world!")
	fmt.Println("Map:         ", shouted)

	// ========================================
	// strconv package — type conversions
	// ========================================
	fmt.Println("\n=== strconv package ===")

	// Int to string
	num := 42
	str := strconv.Itoa(num)
	fmt.Printf("Itoa: %d -> %q\n", num, str)

	// String to int (returns error!)
	val, err := strconv.Atoi("123")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Atoi: %q -> %d\n", "123", val)
	}

	// Invalid conversion
	_, err = strconv.Atoi("not_a_number")
	fmt.Println("Atoi error: ", err)

	// Float conversions
	pi := 3.14159
	piStr := strconv.FormatFloat(pi, 'f', 2, 64) // format, precision, bitSize
	fmt.Printf("FormatFloat: %f -> %q\n", pi, piStr)

	parsedPi, _ := strconv.ParseFloat("3.14159", 64)
	fmt.Printf("ParseFloat: %q -> %f\n", "3.14159", parsedPi)

	// Bool conversions
	boolStr := strconv.FormatBool(true)
	fmt.Printf("FormatBool: %t -> %q\n", true, boolStr)

	parsedBool, _ := strconv.ParseBool("true")
	fmt.Printf("ParseBool: %q -> %t\n", "true", parsedBool)

	// ========================================
	// math package — mathematical functions
	// ========================================
	fmt.Println("\n=== math package ===")

	fmt.Printf("Pi:    %.10f\n", math.Pi)
	fmt.Printf("E:     %.10f\n", math.E)
	fmt.Printf("Sqrt(144): %.0f\n", math.Sqrt(144))
	fmt.Printf("Pow(2,10): %.0f\n", math.Pow(2, 10))
	fmt.Printf("Abs(-42):  %.0f\n", math.Abs(-42))
	fmt.Printf("Ceil(3.2): %.0f\n", math.Ceil(3.2))
	fmt.Printf("Floor(3.8):%.0f\n", math.Floor(3.8))
	fmt.Printf("Round(3.5):%.0f\n", math.Round(3.5))
	fmt.Printf("Max(10,20):%.0f\n", math.Max(10, 20))
	fmt.Printf("Min(10,20):%.0f\n", math.Min(10, 20))
	fmt.Printf("Log(100):  %.4f\n", math.Log(100))   // natural log
	fmt.Printf("Log10(100):%.4f\n", math.Log10(100))  // base-10 log

	// Special values
	fmt.Printf("MaxInt:    %d\n", math.MaxInt)
	fmt.Printf("MaxFloat64:%.2e\n", math.MaxFloat64)
	fmt.Printf("+Inf:      %f\n", math.Inf(1))
	fmt.Println("IsNaN:    ", math.IsNaN(math.NaN()))

	// Trig functions (radians)
	angle := math.Pi / 4 // 45 degrees
	fmt.Printf("\n45 degrees = %.4f radians\n", angle)
	fmt.Printf("Sin: %.4f\n", math.Sin(angle))
	fmt.Printf("Cos: %.4f\n", math.Cos(angle))

	// ========================================
	// sort package — sorting slices
	// ========================================
	fmt.Println("\n=== sort package ===")

	// Sort ints
	numbers := []int{5, 3, 8, 1, 9, 2, 7}
	fmt.Println("Before sort:", numbers)
	sort.Ints(numbers)
	fmt.Println("After sort: ", numbers)

	// Sort strings
	names := []string{"Charlie", "Alice", "Bob", "Diana"}
	fmt.Println("\nBefore sort:", names)
	sort.Strings(names)
	fmt.Println("After sort: ", names)

	// Sort float64s
	floats := []float64{3.14, 1.41, 2.72, 0.58}
	sort.Float64s(floats)
	fmt.Println("\nSorted floats:", floats)

	// Reverse sort (using sort.Slice)
	sort.Slice(numbers, func(i, j int) bool {
		return numbers[i] > numbers[j]
	})
	fmt.Println("\nReverse sorted:", numbers)

	// Custom sort — sort strings by length
	words2 := []string{"go", "python", "c", "javascript", "rust"}
	sort.Slice(words2, func(i, j int) bool {
		return len(words2[i]) < len(words2[j])
	})
	fmt.Println("\nSorted by length:", words2)

	// Check if sorted
	fmt.Println("Is sorted?", sort.IntsAreSorted(numbers)) // false (reverse order)
	sort.Ints(numbers)
	fmt.Println("After re-sort, is sorted?", sort.IntsAreSorted(numbers)) // true

	// Binary search (slice must be sorted)
	idx := sort.SearchInts(numbers, 5)
	fmt.Printf("Binary search for 5: found at index %d\n", idx)

	// ========================================
	// math/rand — random numbers
	// ========================================
	fmt.Println("\n=== math/rand package ===")

	fmt.Println("Random int:       ", rand.Intn(100))    // 0 to 99
	fmt.Println("Random float:     ", rand.Float64())      // 0.0 to 1.0
	fmt.Printf("Random range [5,15): %d\n", rand.Intn(10)+5)

	// Shuffle a slice
	deck := []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	fmt.Println("Shuffled deck:", deck)

	// ========================================
	// Package naming conventions
	// ========================================
	fmt.Println("\n=== Package naming conventions ===")

	fmt.Println(`
  Go package naming best practices:
  - Short, concise, lowercase:    "fmt", "http", "json"
  - No underscores or mixedCaps:  "strconv" not "str_conv"
  - Singular, not plural:         "model" not "models"
  - Avoid "util", "common", "misc" (too vague)
  - Package name should describe what it provides
  - Avoid stuttering: http.Server not http.HTTPServer

  Standard library examples:
    fmt      -> formatted I/O
    os       -> operating system functions
    io       -> I/O primitives
    net      -> networking
    sync     -> synchronization primitives
    time     -> time operations
    encoding -> encoding/decoding (json, xml, csv)
    crypto   -> cryptographic functions
    testing  -> test framework`)
}

// ========================================
// Exported vs Unexported example types
// ========================================

// ExportedPerson is visible outside this package (starts with uppercase)
type ExportedPerson struct {
	Name  string // exported field
	Age   int    // exported field
	email string // unexported field — only accessible within this package
}

// unexportedHelper wouldn't be accessible from outside this package
func unexportedHelper() string {
	return "I'm private to this package"
}

// ========================================
// init() function — runs before main()
// ========================================
func init() {
	fmt.Println("[init] Package initialized! This runs before main().")
}
