package main

import (
	"fmt"
	"sync"
	"time"
)

// ========================================
// Week 6, Lesson 1: Goroutines
// ========================================
// Goroutines are lightweight threads managed by the Go runtime.
// They are one of Go's most powerful features, enabling concurrent
// execution with minimal overhead. A goroutine costs only ~2KB of
// stack space (compared to ~1MB for an OS thread).
// ========================================

func main() {
	// ========================================
	// 1. What Are Goroutines?
	// ========================================
	// A goroutine is a function that runs concurrently with other
	// goroutines in the same address space. You launch one with
	// the `go` keyword.

	fmt.Println("========================================")
	fmt.Println("1. Basic Goroutines")
	fmt.Println("========================================")

	// Without goroutines — sequential execution
	fmt.Println("\nSequential execution:")
	sayHello("Alice")
	sayHello("Bob")

	// With goroutines — concurrent execution
	fmt.Println("\nConcurrent execution (with goroutines):")
	go sayHello("Charlie")
	go sayHello("Diana")

	// IMPORTANT: Without this sleep, main() would exit before
	// the goroutines finish. We'll learn better ways (WaitGroup) soon.
	time.Sleep(100 * time.Millisecond)
	fmt.Println("(Main goroutine continues while others run)")

	// ========================================
	// 2. The go Keyword
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. The go Keyword")
	fmt.Println("========================================")

	// Any function call can be prefixed with `go` to run it
	// as a goroutine. The function starts executing concurrently
	// and the calling goroutine continues immediately.
	fmt.Println("\nLaunching 5 goroutines:")
	for i := 1; i <= 5; i++ {
		go func(id int) {
			fmt.Printf("  Goroutine %d is running\n", id)
		}(i) // Pass i as argument — see closure note below
	}
	time.Sleep(100 * time.Millisecond)

	// ========================================
	// 3. sync.WaitGroup — The Right Way to Wait
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. sync.WaitGroup")
	fmt.Println("========================================")

	// WaitGroup is the proper way to wait for goroutines to finish.
	// It's a counter: Add(n) increments, Done() decrements, Wait() blocks
	// until the counter reaches zero.

	var wg sync.WaitGroup

	fmt.Println("\nUsing WaitGroup to wait for goroutines:")
	for i := 1; i <= 5; i++ {
		wg.Add(1) // Increment counter BEFORE launching goroutine
		go func(id int) {
			defer wg.Done() // Decrement counter when goroutine finishes
			time.Sleep(time.Duration(id*10) * time.Millisecond)
			fmt.Printf("  Worker %d finished\n", id)
		}(i)
	}

	wg.Wait() // Block until all goroutines call Done()
	fmt.Println("All workers completed!")

	// ========================================
	// 4. Anonymous Goroutine Functions
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Anonymous Goroutine Functions")
	fmt.Println("========================================")

	// You can launch goroutines with anonymous (inline) functions.
	// This is very common in Go.

	var wg2 sync.WaitGroup

	// Anonymous function with no parameters
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		fmt.Println("  Anonymous goroutine with no params")
	}()

	// Anonymous function with parameters
	wg2.Add(1)
	go func(msg string) {
		defer wg2.Done()
		fmt.Printf("  Anonymous goroutine says: %s\n", msg)
	}("Hello from anonymous!")

	// Anonymous function capturing a value
	name := "Go Developer"
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		// This captures `name` from the enclosing scope (closure)
		fmt.Printf("  Greeting: Welcome, %s!\n", name)
	}()

	wg2.Wait()

	// ========================================
	// 5. Demonstrating Concurrent Execution
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Demonstrating Concurrent Execution")
	fmt.Println("========================================")

	// This example clearly shows goroutines running concurrently.
	// Each goroutine prints its progress, and you'll see interleaved output.

	var wg3 sync.WaitGroup

	tasks := []string{"Download", "Process", "Upload"}
	for _, task := range tasks {
		wg3.Add(1)
		go func(taskName string) {
			defer wg3.Done()
			for step := 1; step <= 3; step++ {
				fmt.Printf("  [%s] Step %d/3\n", taskName, step)
				time.Sleep(20 * time.Millisecond)
			}
			fmt.Printf("  [%s] Complete!\n", taskName)
		}(task)
	}

	wg3.Wait()
	fmt.Println("All tasks finished!")
	// Expected output: Steps from different tasks will be interleaved,
	// proving they run concurrently, not sequentially.

	// ========================================
	// 6. Race Condition Awareness
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Race Condition Awareness")
	fmt.Println("========================================")

	// A race condition occurs when multiple goroutines access shared
	// data concurrently and at least one of them writes to it.

	// UNSAFE: Multiple goroutines incrementing a shared counter
	fmt.Println("\nUNSAFE counter (race condition):")
	unsafeCounter := 0
	var wg4 sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg4.Add(1)
		go func() {
			defer wg4.Done()
			unsafeCounter++ // DATA RACE! Multiple goroutines read-modify-write
		}()
	}
	wg4.Wait()
	fmt.Printf("  Expected: 1000, Got: %d (may be less due to race!)\n", unsafeCounter)

	// SAFE: Using a mutex to protect shared state
	fmt.Println("\nSAFE counter (with mutex):")
	safeCounter := 0
	var mu sync.Mutex
	var wg5 sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg5.Add(1)
		go func() {
			defer wg5.Done()
			mu.Lock()         // Acquire the lock
			safeCounter++     // Only one goroutine can access this at a time
			mu.Unlock()       // Release the lock
		}()
	}
	wg5.Wait()
	fmt.Printf("  Expected: 1000, Got: %d (always correct!)\n", safeCounter)

	// TIP: Run your programs with `go run -race main.go` to detect
	// race conditions. The race detector will report any data races
	// it finds at runtime.

	// ========================================
	// 7. Common Goroutine Pitfall: Closure Variables
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. Closure Variable Pitfall")
	fmt.Println("========================================")

	// WRONG: Capturing the loop variable directly
	fmt.Println("\nBUGGY (captures loop variable by reference):")
	var wg6 sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg6.Add(1)
		go func() {
			defer wg6.Done()
			// By the time this runs, the loop may have finished,
			// so `i` might be 5 for all goroutines.
			// NOTE: As of Go 1.22+, the loop variable is per-iteration,
			// so this is now safe. But passing as a parameter is still
			// a good habit for clarity.
			fmt.Printf("  Captured i = %d\n", i)
		}()
	}
	wg6.Wait()

	// CORRECT: Pass the variable as a function argument
	fmt.Println("\nCORRECT (pass as argument):")
	var wg7 sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg7.Add(1)
		go func(val int) {
			defer wg7.Done()
			fmt.Printf("  Passed val = %d\n", val)
		}(i) // i is copied here, each goroutine gets its own value
	}
	wg7.Wait()

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println("- Use `go funcName()` to launch a goroutine")
	fmt.Println("- Use sync.WaitGroup to wait for goroutines to finish")
	fmt.Println("- Pass loop variables as arguments to avoid closure bugs")
	fmt.Println("- Use mutexes or channels (next lesson) to protect shared state")
	fmt.Println("- Run with `go run -race main.go` to detect race conditions")
}

// sayHello prints a greeting with a small delay to simulate work.
func sayHello(name string) {
	fmt.Printf("  Hello, %s!\n", name)
}
