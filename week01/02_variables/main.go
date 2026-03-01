package main

import "fmt"

func main() {
	// ========================================
	// Way 1: var keyword with explicit type
	// ========================================
	var name string = "Sri"
	var age int = 30
	var height float64 = 5.9
	var isLearning bool = true

	fmt.Println("=== Explicit type declaration ===")
	fmt.Println("Name:", name)
	fmt.Println("Age:", age)
	fmt.Println("Height:", height)
	fmt.Println("Learning Go?", isLearning)

	// ========================================
	// Way 2: var with type inference (Go figures out the type)
	// ========================================
	var language = "Go" // Go infers this is a string
	var version = 1.26  // Go infers this is float64

	fmt.Println("\n=== Type inference ===")
	fmt.Println("Language:", language)
	fmt.Println("Version:", version)

	// ========================================
	// Way 3: Short declaration with := (most common!)
	// ========================================
	city := "Chennai"   // string
	year := 2026        // int
	pi := 3.14159       // float64
	active := true      // bool

	fmt.Println("\n=== Short declaration := ===")
	fmt.Println("City:", city)
	fmt.Println("Year:", year)
	fmt.Println("Pi:", pi)
	fmt.Println("Active:", active)

	// ========================================
	// Constants: values that never change
	// ========================================
	const maxRetries = 3
	const appName = "LearnGo"

	fmt.Println("\n=== Constants ===")
	fmt.Println("App:", appName)
	fmt.Println("Max retries:", maxRetries)

	// ========================================
	// Zero values: uninitialized variables get defaults
	// ========================================
	var emptyString string  // ""
	var zeroInt int         // 0
	var zeroFloat float64   // 0.0
	var falseBool bool      // false

	fmt.Println("\n=== Zero values (defaults) ===")
	fmt.Printf("string: %q\n", emptyString)  // %q shows quotes
	fmt.Println("int:", zeroInt)
	fmt.Println("float64:", zeroFloat)
	fmt.Println("bool:", falseBool)
}
