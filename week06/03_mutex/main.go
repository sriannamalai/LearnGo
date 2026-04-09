package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ========================================
// Week 6, Lesson 3: Mutex & Concurrency Patterns
// ========================================
// When multiple goroutines need to access shared state, we need
// synchronization primitives. This lesson covers sync.Mutex,
// sync.RWMutex, sync/atomic operations, and common concurrency
// patterns like worker pools, fan-in, and fan-out.
// ========================================

func main() {
	// ========================================
	// 1. sync.Mutex — Mutual Exclusion Lock
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. sync.Mutex")
	fmt.Println("========================================")

	// A Mutex ensures that only one goroutine can access a
	// critical section at a time. Lock() acquires, Unlock() releases.

	fmt.Println("\nSafe counter with Mutex:")
	counter := &SafeCounter{value: 0}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Increment()
		}()
	}
	wg.Wait()
	fmt.Printf("  Final count: %d (expected: 1000)\n", counter.Value())

	// Using mutex to protect a map
	fmt.Println("\nSafe map with Mutex:")
	safeMap := &SafeMap{data: make(map[string]int)}

	var wg2 sync.WaitGroup
	languages := []string{"Go", "Python", "Rust", "Java", "Go", "Rust", "Go"}
	for _, lang := range languages {
		wg2.Add(1)
		go func(l string) {
			defer wg2.Done()
			safeMap.Increment(l)
		}(lang)
	}
	wg2.Wait()

	fmt.Println("  Language counts:")
	for k, v := range safeMap.GetAll() {
		fmt.Printf("    %s: %d\n", k, v)
	}

	// ========================================
	// 2. sync.RWMutex — Read-Write Mutex
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. sync.RWMutex")
	fmt.Println("========================================")

	// RWMutex allows multiple concurrent readers OR a single writer.
	// This is more efficient when reads vastly outnumber writes.
	//   RLock/RUnlock — for readers (multiple allowed)
	//   Lock/Unlock   — for writers (exclusive access)

	cache := &Cache{
		data: map[string]string{
			"name":  "Go Developer",
			"level": "intermediate",
		},
	}

	var wg3 sync.WaitGroup

	// Launch many readers
	for i := 0; i < 5; i++ {
		wg3.Add(1)
		go func(id int) {
			defer wg3.Done()
			val := cache.Get("name")
			fmt.Printf("  Reader %d: name = %s\n", id, val)
		}(i)
	}

	// Launch a writer
	wg3.Add(1)
	go func() {
		defer wg3.Done()
		cache.Set("level", "advanced")
		fmt.Println("  Writer: updated level to advanced")
	}()

	wg3.Wait()
	fmt.Printf("  Final level: %s\n", cache.Get("level"))

	// ========================================
	// 3. Atomic Operations (sync/atomic)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Atomic Operations")
	fmt.Println("========================================")

	// For simple operations (counters, flags), atomic operations
	// are faster than mutexes because they use hardware-level
	// atomic instructions — no locking overhead.

	var atomicCounter atomic.Int64

	var wg4 sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg4.Add(1)
		go func() {
			defer wg4.Done()
			atomicCounter.Add(1) // Atomic increment
		}()
	}
	wg4.Wait()
	fmt.Printf("  Atomic counter: %d (expected: 1000)\n", atomicCounter.Load())

	// atomic.Bool for flags
	var running atomic.Bool
	running.Store(true)
	fmt.Printf("  Running: %v\n", running.Load())
	running.Store(false)
	fmt.Printf("  Running: %v\n", running.Load())

	// atomic.Value for any type (useful for config hot-reloading)
	var config atomic.Value
	config.Store(AppConfig{MaxConns: 100, Debug: false})
	cfg := config.Load().(AppConfig)
	fmt.Printf("  Config: MaxConns=%d, Debug=%v\n", cfg.MaxConns, cfg.Debug)

	// Swap: store new value and get old value
	config.Store(AppConfig{MaxConns: 200, Debug: true})
	newCfg := config.Load().(AppConfig)
	fmt.Printf("  Updated Config: MaxConns=%d, Debug=%v\n", newCfg.MaxConns, newCfg.Debug)

	// CompareAndSwap on Int64
	var cas atomic.Int64
	cas.Store(42)
	swapped := cas.CompareAndSwap(42, 100) // Only swap if current value is 42
	fmt.Printf("  CAS: swapped=%v, value=%d\n", swapped, cas.Load())
	swapped = cas.CompareAndSwap(42, 200) // Won't swap — value is now 100
	fmt.Printf("  CAS: swapped=%v, value=%d\n", swapped, cas.Load())

	// ========================================
	// 4. Protecting Shared State
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Protecting Shared State")
	fmt.Println("========================================")

	// Guidelines for choosing synchronization:
	// - Simple counters/flags → sync/atomic
	// - Protecting a data structure → sync.Mutex
	// - Read-heavy workloads → sync.RWMutex
	// - Communicating between goroutines → channels

	fmt.Println("  sync/atomic:  Best for simple counters and flags")
	fmt.Println("  sync.Mutex:   Best for protecting data structures")
	fmt.Println("  sync.RWMutex: Best for read-heavy, write-light workloads")
	fmt.Println("  channels:     Best for goroutine communication")

	// Example: thread-safe bank account
	fmt.Println("\nThread-safe bank account:")
	account := &BankAccount{balance: 1000}

	var wg5 sync.WaitGroup
	// 10 deposits and 10 withdrawals concurrently
	for i := 0; i < 10; i++ {
		wg5.Add(2)
		go func() {
			defer wg5.Done()
			account.Deposit(100)
		}()
		go func() {
			defer wg5.Done()
			account.Withdraw(50)
		}()
	}
	wg5.Wait()
	// 1000 + (10 * 100) - (10 * 50) = 1500
	fmt.Printf("  Final balance: $%.2f (expected: $1500.00)\n", account.Balance())

	// ========================================
	// 5. Worker Pool Pattern
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Worker Pool Pattern")
	fmt.Println("========================================")

	// A worker pool limits the number of goroutines processing
	// tasks concurrently. This prevents overwhelming resources.

	const numWorkers = 3
	const numJobs = 9

	jobs := make(chan Job, numJobs)
	results := make(chan Result, numJobs)

	// Start workers
	var wg6 sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg6.Add(1)
		go worker(w, jobs, results, &wg6)
	}

	// Send jobs
	for j := 1; j <= numJobs; j++ {
		jobs <- Job{ID: j, Data: j * 10}
	}
	close(jobs) // No more jobs

	// Wait for workers to finish, then close results
	go func() {
		wg6.Wait()
		close(results)
	}()

	// Collect results
	fmt.Println("  Results:")
	for r := range results {
		fmt.Printf("    Job %d: input=%d, result=%d (worker %d)\n",
			r.JobID, r.Input, r.Output, r.WorkerID)
	}

	// ========================================
	// 6. Fan-Out Pattern
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Fan-Out Pattern")
	fmt.Println("========================================")

	// Fan-out: one channel distributes work to multiple goroutines.
	// Each value is processed by exactly one goroutine.

	source := make(chan int, 10)
	go func() {
		for i := 1; i <= 10; i++ {
			source <- i
		}
		close(source)
	}()

	// Fan out to 3 workers
	var fanOutWg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		fanOutWg.Add(1)
		go func(workerID int) {
			defer fanOutWg.Done()
			for val := range source {
				fmt.Printf("  Worker %d processed: %d\n", workerID, val)
				time.Sleep(10 * time.Millisecond)
			}
		}(w)
	}
	fanOutWg.Wait()

	// ========================================
	// 7. Fan-In Pattern
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. Fan-In Pattern")
	fmt.Println("========================================")

	// Fan-in: multiple channels merge into one channel.
	// Useful for combining results from multiple sources.

	ch1 := produceNumbers("A", 1, 3)
	ch2 := produceNumbers("B", 4, 6)
	ch3 := produceNumbers("C", 7, 9)

	// Merge all channels into one
	merged := fanIn(ch1, ch2, ch3)

	fmt.Println("  Merged results:")
	for val := range merged {
		fmt.Printf("    %s\n", val)
	}

	// ========================================
	// 8. sync.Once — Run Something Exactly Once
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. sync.Once")
	fmt.Println("========================================")

	// sync.Once ensures a function is executed exactly once,
	// even when called from multiple goroutines. Useful for
	// one-time initialization.

	var once sync.Once
	var wg7 sync.WaitGroup

	initFunc := func() {
		fmt.Println("  Initialization (runs only once)")
	}

	for i := 0; i < 5; i++ {
		wg7.Add(1)
		go func(id int) {
			defer wg7.Done()
			once.Do(initFunc) // Only the first call executes initFunc
			fmt.Printf("  Goroutine %d: after init\n", id)
		}(i)
	}
	wg7.Wait()

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println("- sync.Mutex: exclusive lock for one goroutine at a time")
	fmt.Println("- sync.RWMutex: multiple readers OR one writer")
	fmt.Println("- sync/atomic: lock-free operations for simple types")
	fmt.Println("- Worker Pool: fixed goroutines processing a job channel")
	fmt.Println("- Fan-Out: one source, many consumers")
	fmt.Println("- Fan-In: many sources, one consumer")
	fmt.Println("- sync.Once: thread-safe one-time initialization")
}

// ========================================
// Types and Helper Functions
// ========================================

// SafeCounter is a thread-safe counter using Mutex.
type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// SafeMap is a thread-safe map using Mutex.
type SafeMap struct {
	mu   sync.Mutex
	data map[string]int
}

func (m *SafeMap) Increment(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key]++
}

func (m *SafeMap) GetAll() map[string]int {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Return a copy to avoid data races on the returned map
	copy := make(map[string]int, len(m.data))
	for k, v := range m.data {
		copy[k] = v
	}
	return copy
}

// Cache is a thread-safe key-value store using RWMutex.
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func (c *Cache) Get(key string) string {
	c.mu.RLock() // Multiple readers can hold this lock
	defer c.mu.RUnlock()
	return c.data[key]
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock() // Exclusive access for writing
	defer c.mu.Unlock()
	c.data[key] = value
}

// AppConfig is used to demonstrate atomic.Value.
type AppConfig struct {
	MaxConns int
	Debug    bool
}

// BankAccount is a thread-safe bank account.
type BankAccount struct {
	mu      sync.Mutex
	balance float64
}

func (a *BankAccount) Deposit(amount float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance += amount
}

func (a *BankAccount) Withdraw(amount float64) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.balance < amount {
		return false
	}
	a.balance -= amount
	return true
}

func (a *BankAccount) Balance() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// Job and Result types for the worker pool.
type Job struct {
	ID   int
	Data int
}

type Result struct {
	JobID    int
	Input    int
	Output   int
	WorkerID int
}

// worker processes jobs from the jobs channel and sends results.
func worker(id int, jobs <-chan Job, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		results <- Result{
			JobID:    job.ID,
			Input:    job.Data,
			Output:   job.Data * 2, // Double the input
			WorkerID: id,
		}
	}
}

// produceNumbers generates string-labeled numbers on a channel.
func produceNumbers(label string, start, end int) <-chan string {
	ch := make(chan string)
	go func() {
		for i := start; i <= end; i++ {
			ch <- fmt.Sprintf("[%s] %d", label, i)
			time.Sleep(10 * time.Millisecond)
		}
		close(ch)
	}()
	return ch
}

// fanIn merges multiple channels into one.
func fanIn(channels ...<-chan string) <-chan string {
	merged := make(chan string)
	var wg sync.WaitGroup

	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan string) {
			defer wg.Done()
			for val := range c {
				merged <- val
			}
		}(ch)
	}

	// Close merged channel when all input channels are done
	go func() {
		wg.Wait()
		close(merged)
	}()

	return merged
}
