package main

// ========================================
// Week 22 — Lesson 2: Go Mobile
// ========================================
// This lesson covers:
//   - golang.org/x/mobile overview and use cases
//   - gomobile bind: creating shared libraries
//   - Calling Go from native iOS (Swift/Objective-C)
//   - Calling Go from native Android (Java/Kotlin)
//   - Limitations and best practices
//   - When to use gomobile vs Fyne mobile
//
// What is gomobile?
//   gomobile is a tool that lets you:
//   1. Build pure Go mobile apps (gomobile build)
//   2. Create shared libraries callable from native code (gomobile bind)
//
//   Option 2 is far more common — it lets you write
//   business logic in Go and use native UI on each platform.
//
// Prerequisites:
//   1. Install gomobile:
//      go install golang.org/x/mobile/cmd/gomobile@latest
//      gomobile init
//
//   2. For iOS:
//      - macOS with Xcode
//      - gomobile bind creates an .xcframework
//
//   3. For Android:
//      - Android SDK and NDK
//      - Set ANDROID_HOME, ANDROID_NDK_HOME
//      - gomobile bind creates an .aar file
//
// Architecture when using gomobile bind:
//
//   ┌─────────────────────────────────────────────────┐
//   │                 Mobile App                       │
//   │                                                  │
//   │  ┌──────────────────┐  ┌─────────────────────┐  │
//   │  │  Native UI Layer │  │  Go Shared Library  │  │
//   │  │                  │  │                      │  │
//   │  │  iOS: SwiftUI    │──│  Business logic      │  │
//   │  │  Android: Jetpack│  │  Data processing     │  │
//   │  │                  │  │  Networking           │  │
//   │  │  Platform-native │  │  Cryptography         │  │
//   │  │  look and feel   │  │  Shared algorithms   │  │
//   │  └──────────────────┘  └─────────────────────┘  │
//   └─────────────────────────────────────────────────┘
//
// Build commands:
//   # Create iOS framework:
//   gomobile bind -target=ios -o Greeter.xcframework ./greeter
//
//   # Create Android AAR:
//   gomobile bind -target=android -o greeter.aar ./greeter
//
//   # Create both:
//   gomobile bind -target=ios,android ./greeter
//
// Run (desktop demo):
//   go run .

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// ========================================
// Package Structure for gomobile bind
// ========================================
// When using gomobile bind, your Go code lives in a
// separate package (not main). Only exported types and
// functions are accessible from native code.
//
// Example package structure:
//   myapp/
//     greeter/         <-- Go package for binding
//       greeter.go     <-- Exported functions
//     ios/             <-- Xcode project
//     android/         <-- Android Studio project
//
// Type restrictions for gomobile bind:
//   Supported parameter/return types:
//     - Signed integers, floats, booleans
//     - Strings and byte slices
//     - Functions (callbacks)
//     - Interfaces with exported methods
//     - Structs with exported fields
//
//   NOT supported:
//     - Maps (use JSON serialization instead)
//     - Channels
//     - Unsigned integers (except byte)
//     - Multiple return values (except T, error)
//     - Slices of slices

// ========================================
// Greeter — Simple Example
// ========================================
// This simulates what would be in a bindable package.
// In a real project, this would be in package greeter,
// not package main.

// Greeter provides greeting functionality.
// When bound, iOS/Android can create instances and call methods.
//
// Swift usage:
//   let greeter = GreeterNewGreeter("Alice")
//   let msg = greeter.greet("en")
//
// Kotlin usage:
//   val greeter = Greeter.newGreeter("Alice")
//   val msg = greeter.greet("en")
type Greeter struct {
	Name string
}

// NewGreeter creates a new Greeter. Factory functions are
// the standard pattern for gomobile — constructors must be
// package-level functions starting with "New".
func NewGreeter(name string) *Greeter {
	return &Greeter{Name: name}
}

// Greet returns a greeting in the specified language.
func (g *Greeter) Greet(language string) string {
	switch strings.ToLower(language) {
	case "es":
		return fmt.Sprintf("Hola, %s!", g.Name)
	case "fr":
		return fmt.Sprintf("Bonjour, %s!", g.Name)
	case "de":
		return fmt.Sprintf("Hallo, %s!", g.Name)
	case "ja":
		return fmt.Sprintf("Konnichiwa, %s!", g.Name)
	default:
		return fmt.Sprintf("Hello, %s!", g.Name)
	}
}

// ========================================
// CryptoUtils — Business Logic Example
// ========================================
// This demonstrates sharing non-trivial business logic
// between iOS and Android via Go.

// HashPassword creates a SHA-256 hash of the input.
// In production, use bcrypt or argon2 instead.
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ValidateEmail performs basic email validation.
func ValidateEmail(email string) bool {
	if len(email) < 5 {
		return false
	}
	atIndex := strings.Index(email, "@")
	if atIndex < 1 {
		return false
	}
	dotIndex := strings.LastIndex(email, ".")
	if dotIndex < atIndex+2 {
		return false
	}
	if dotIndex >= len(email)-1 {
		return false
	}
	return true
}

// GenerateToken creates a simple time-based token.
// In production, use proper JWT or similar.
func GenerateToken(userID string) string {
	data := fmt.Sprintf("%s:%d", userID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // Short token for demo
}

// ========================================
// DataProcessor — Complex Data Example
// ========================================
// When gomobile doesn't support a type (like maps),
// serialize to JSON and pass as a string.

// DataProcessor handles data transformations shared
// between iOS and Android.
type DataProcessor struct {
	data []float64
}

// NewDataProcessor creates a new processor.
func NewDataProcessor() *DataProcessor {
	return &DataProcessor{data: []float64{}}
}

// AddValue adds a data point.
func (dp *DataProcessor) AddValue(value float64) {
	dp.data = append(dp.data, value)
}

// Count returns the number of data points.
func (dp *DataProcessor) Count() int {
	return len(dp.data)
}

// Mean calculates the arithmetic mean.
func (dp *DataProcessor) Mean() float64 {
	if len(dp.data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range dp.data {
		sum += v
	}
	return sum / float64(len(dp.data))
}

// StdDev calculates the standard deviation.
func (dp *DataProcessor) StdDev() float64 {
	if len(dp.data) < 2 {
		return 0
	}
	mean := dp.Mean()
	sumSq := 0.0
	for _, v := range dp.data {
		diff := v - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(dp.data)-1))
}

// Median returns the median value.
func (dp *DataProcessor) Median() float64 {
	if len(dp.data) == 0 {
		return 0
	}

	sorted := make([]float64, len(dp.data))
	copy(sorted, dp.data)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// GetStatsJSON returns all statistics as a JSON string.
// This is the pattern for returning complex data to
// native code, since gomobile doesn't support maps.
//
// Swift: let json = processor.getStatsJSON()
// Kotlin: val json = processor.statsJSON
func (dp *DataProcessor) GetStatsJSON() string {
	stats := map[string]interface{}{
		"count":  dp.Count(),
		"mean":   dp.Mean(),
		"median": dp.Median(),
		"stdDev": dp.StdDev(),
		"min":    dp.min(),
		"max":    dp.max(),
	}

	data, err := json.Marshal(stats)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func (dp *DataProcessor) min() float64 {
	if len(dp.data) == 0 {
		return 0
	}
	min := dp.data[0]
	for _, v := range dp.data[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func (dp *DataProcessor) max() float64 {
	if len(dp.data) == 0 {
		return 0
	}
	max := dp.data[0]
	for _, v := range dp.data[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// ========================================
// Callback Interface
// ========================================
// gomobile supports interfaces for callbacks.
// Define an interface in Go, implement it in
// Swift/Kotlin, and pass it to Go functions.
//
// Go:
//   type ProgressCallback interface {
//       OnProgress(percent int)
//       OnComplete(result string)
//       OnError(message string)
//   }
//
// Swift:
//   class MyCallback: NSObject, GreeterProgressCallback {
//       func onProgress(_ percent: Int) { ... }
//       func onComplete(_ result: String?) { ... }
//       func onError(_ message: String?) { ... }
//   }
//
// Kotlin:
//   class MyCallback : Greeter.ProgressCallback {
//       override fun onProgress(percent: Long) { ... }
//       override fun onComplete(result: String?) { ... }
//       override fun onError(message: String?) { ... }
//   }

// ProgressCallback defines the interface for progress updates.
type ProgressCallback interface {
	OnProgress(percent int)
	OnComplete(result string)
	OnError(message string)
}

// ProcessWithCallback demonstrates async work with callbacks.
func ProcessWithCallback(input string, callback ProgressCallback) {
	go func() {
		for i := 0; i <= 100; i += 20 {
			callback.OnProgress(i)
			time.Sleep(200 * time.Millisecond)
		}

		result := fmt.Sprintf("Processed: %s (%d chars)", input, len(input))
		callback.OnComplete(result)
	}()
}

// ========================================
// Native Integration Examples (Comments)
// ========================================
//
// === iOS (Swift) ===
//
// 1. Build the framework:
//    gomobile bind -target=ios -o MyLib.xcframework ./mylib
//
// 2. Add to Xcode:
//    - Drag MyLib.xcframework into your project
//    - Ensure it's in "Frameworks, Libraries" under target
//
// 3. Use in Swift:
//    import MyLib
//
//    // Simple function call
//    let hash = MyLibHashPassword("secret123")
//
//    // Struct usage
//    let greeter = MyLibNewGreeter("Alice")
//    let msg = greeter!.greet("en")
//
//    // Data processing
//    let proc = MyLibNewDataProcessor()!
//    proc.addValue(10.5)
//    proc.addValue(20.3)
//    let stats = proc.getStatsJSON()
//
// === Android (Kotlin) ===
//
// 1. Build the AAR:
//    gomobile bind -target=android -o mylib.aar ./mylib
//
// 2. Add to Android Studio:
//    - Copy mylib.aar to app/libs/
//    - Add to build.gradle: implementation files('libs/mylib.aar')
//
// 3. Use in Kotlin:
//    import mylib.Mylib
//
//    // Simple function call
//    val hash = Mylib.hashPassword("secret123")
//
//    // Struct usage
//    val greeter = Mylib.newGreeter("Alice")
//    val msg = greeter.greet("en")
//
//    // Data processing
//    val proc = Mylib.newDataProcessor()
//    proc.addValue(10.5)
//    proc.addValue(20.3)
//    val stats = proc.statsJSON

// ========================================
// Main — Demo
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 22 - Lesson 2: Go Mobile")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("This file demonstrates Go code that can be")
	fmt.Println("shared with iOS and Android using gomobile bind.")
	fmt.Println()

	// ========================================
	// Demo: Greeter
	// ========================================
	fmt.Println("--- Greeter ---")
	g := NewGreeter("Gopher")
	languages := []string{"en", "es", "fr", "de", "ja"}
	for _, lang := range languages {
		fmt.Printf("  %s: %s\n", lang, g.Greet(lang))
	}

	// ========================================
	// Demo: Crypto Utils
	// ========================================
	fmt.Println("\n--- Crypto Utils ---")
	hash := HashPassword("mySecretPassword")
	fmt.Printf("  Hash: %s...\n", hash[:16])

	emails := []string{"valid@example.com", "invalid", "also@bad", "good@test.org"}
	for _, email := range emails {
		fmt.Printf("  %s: valid=%v\n", email, ValidateEmail(email))
	}

	token := GenerateToken("user_123")
	fmt.Printf("  Token: %s\n", token)

	// ========================================
	// Demo: Data Processor
	// ========================================
	fmt.Println("\n--- Data Processor ---")
	dp := NewDataProcessor()
	values := []float64{10.5, 20.3, 15.7, 30.1, 25.4, 18.9, 22.6}
	for _, v := range values {
		dp.AddValue(v)
	}

	fmt.Printf("  Count:  %d\n", dp.Count())
	fmt.Printf("  Mean:   %.2f\n", dp.Mean())
	fmt.Printf("  Median: %.2f\n", dp.Median())
	fmt.Printf("  StdDev: %.2f\n", dp.StdDev())
	fmt.Printf("  JSON:   %s\n", dp.GetStatsJSON())

	// ========================================
	// Demo: Callback
	// ========================================
	fmt.Println("\n--- Callback Pattern ---")

	// Simulate native callback implementation
	callback := &demoCallback{}
	ProcessWithCallback("Hello from Go!", callback)

	// Wait for async work to complete
	time.Sleep(2 * time.Second)

	fmt.Println("\n--- Build Commands ---")
	fmt.Println("  iOS:     gomobile bind -target=ios -o MyLib.xcframework ./mylib")
	fmt.Println("  Android: gomobile bind -target=android -o mylib.aar ./mylib")
	fmt.Println("  Both:    gomobile bind -target=ios,android ./mylib")
}

// demoCallback implements ProgressCallback for testing.
type demoCallback struct{}

func (d *demoCallback) OnProgress(percent int) {
	fmt.Printf("  Progress: %d%%\n", percent)
}

func (d *demoCallback) OnComplete(result string) {
	fmt.Printf("  Complete: %s\n", result)
}

func (d *demoCallback) OnError(message string) {
	fmt.Printf("  Error: %s\n", message)
}
