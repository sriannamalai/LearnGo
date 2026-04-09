package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// ========================================
// Week 19 — Lesson 1: Resilience Patterns
// ========================================
//
// In a distributed system, failures are inevitable. Resilience patterns
// help your services survive and recover from failures gracefully.
//
// This lesson implements four key patterns FROM SCRATCH:
//   1. Circuit Breaker:      Prevents cascading failures
//   2. Retry with Backoff:   Handles transient failures
//   3. Timeout:              Prevents indefinite blocking
//   4. Bulkhead:             Limits concurrent access
//
// No external libraries — we build each pattern to understand the concepts.
//
// To run:
//   go run main.go

// ========================================
// 1. CIRCUIT BREAKER
// ========================================
//
// The circuit breaker prevents your service from repeatedly calling
// a failing dependency. It has three states:
//
//   CLOSED  ──(failures reach threshold)──>  OPEN
//     ^                                         |
//     |                                    (timeout expires)
//     |                                         |
//     └───(successful probe)───  HALF-OPEN  <───┘
//
//   CLOSED:    Normal operation. Requests pass through.
//              Failures are counted. If threshold is reached, switch to OPEN.
//
//   OPEN:      Requests fail immediately (no call to the dependency).
//              After a timeout, switch to HALF-OPEN.
//
//   HALF-OPEN: Allow ONE probe request through.
//              If it succeeds: switch to CLOSED (dependency recovered).
//              If it fails: switch back to OPEN (dependency still down).

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed   CircuitState = iota // Normal operation
	StateOpen                         // Blocking requests
	StateHalfOpen                     // Testing recovery
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu sync.Mutex

	// Configuration
	name             string
	failureThreshold int           // Failures to trigger OPEN
	resetTimeout     time.Duration // How long OPEN lasts before HALF-OPEN
	halfOpenMax      int           // Max concurrent requests in HALF-OPEN

	// State
	state        CircuitState
	failures     int       // Consecutive failure count
	successes    int       // Consecutive success count (in HALF-OPEN)
	lastFailure  time.Time // When the last failure happened
	totalCalls   int
	totalSuccess int
	totalFailure int
}

// NewCircuitBreaker creates a circuit breaker with the given settings.
func NewCircuitBreaker(name string, threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: threshold,
		resetTimeout:     timeout,
		halfOpenMax:      1,
		state:            StateClosed,
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Call executes the given function through the circuit breaker.
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check if we should transition from OPEN to HALF-OPEN
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.successes = 0
			fmt.Printf("    [CB %s] State: OPEN -> HALF-OPEN (timeout expired, probing...)\n", cb.name)
		} else {
			cb.totalCalls++
			cb.totalFailure++
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}

	cb.totalCalls++
	cb.mu.Unlock()

	// Execute the function
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		// Failure
		cb.failures++
		cb.totalFailure++
		cb.lastFailure = time.Now()

		if cb.state == StateHalfOpen {
			// Probe failed — back to OPEN
			cb.state = StateOpen
			fmt.Printf("    [CB %s] State: HALF-OPEN -> OPEN (probe failed)\n", cb.name)
		} else if cb.failures >= cb.failureThreshold {
			// Too many failures — switch to OPEN
			cb.state = StateOpen
			fmt.Printf("    [CB %s] State: CLOSED -> OPEN (threshold reached: %d failures)\n",
				cb.name, cb.failures)
		}
		return err
	}

	// Success
	cb.totalSuccess++

	if cb.state == StateHalfOpen {
		// Probe succeeded — back to CLOSED
		cb.state = StateClosed
		cb.failures = 0
		fmt.Printf("    [CB %s] State: HALF-OPEN -> CLOSED (probe succeeded, recovered!)\n", cb.name)
	} else {
		cb.failures = 0 // Reset failure count on success
	}
	return nil
}

// Stats returns circuit breaker statistics.
func (cb *CircuitBreaker) Stats() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return fmt.Sprintf("State: %s, Calls: %d, Success: %d, Failures: %d",
		cb.state, cb.totalCalls, cb.totalSuccess, cb.totalFailure)
}

// ========================================
// 2. RETRY WITH EXPONENTIAL BACKOFF
// ========================================
//
// When a call fails due to a transient error, retry with increasing
// delays between attempts. This prevents overwhelming a recovering service.
//
// Delay pattern: baseDelay * 2^attempt + random jitter
//   Attempt 0: ~100ms
//   Attempt 1: ~200ms
//   Attempt 2: ~400ms
//   Attempt 3: ~800ms
//   (capped at maxDelay)

// RetryConfig configures the retry behavior.
type RetryConfig struct {
	MaxAttempts int           // Maximum number of attempts (including first)
	BaseDelay   time.Duration // Initial delay between retries
	MaxDelay    time.Duration // Maximum delay between retries
	Jitter      bool          // Add randomness to prevent thundering herd
}

// DefaultRetryConfig provides sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 4,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Jitter:      true,
	}
}

// Retry executes the function with exponential backoff.
func Retry(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		}

		lastErr = fn()
		if lastErr == nil {
			if attempt > 0 {
				fmt.Printf("    [Retry] Succeeded on attempt %d\n", attempt+1)
			}
			return nil // Success!
		}

		if attempt < cfg.MaxAttempts-1 {
			// Calculate delay with exponential backoff
			delay := cfg.BaseDelay * time.Duration(math.Pow(2, float64(attempt)))
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}

			// Add jitter (±25%) to prevent thundering herd
			if cfg.Jitter {
				jitter := time.Duration(rand.Int63n(int64(delay) / 2))
				delay = delay + jitter - time.Duration(int64(delay)/4)
			}

			fmt.Printf("    [Retry] Attempt %d failed: %v. Retrying in %v...\n",
				attempt+1, lastErr, delay.Round(time.Millisecond))

			// Wait with context awareness
			select {
			case <-time.After(delay):
				// Continue to next attempt
			case <-ctx.Done():
				return fmt.Errorf("retry cancelled during wait: %w", ctx.Err())
			}
		}
	}

	return fmt.Errorf("all %d attempts failed, last error: %w", cfg.MaxAttempts, lastErr)
}

// ========================================
// 3. TIMEOUT PATTERN
// ========================================
//
// Wraps a function with a deadline. If the function doesn't complete
// within the timeout, it returns an error. This prevents indefinite
// blocking when a dependency hangs.

// ErrTimeout is returned when an operation exceeds its deadline.
var ErrTimeout = errors.New("operation timed out")

// WithTimeout executes fn with a timeout. If fn takes longer than
// the timeout, it returns ErrTimeout. Note: this doesn't cancel fn!
// For true cancellation, fn should accept and respect a context.
func WithTimeout(timeout time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	type result struct {
		value interface{}
		err   error
	}

	ch := make(chan result, 1)

	go func() {
		v, err := fn()
		ch <- result{v, err}
	}()

	select {
	case r := <-ch:
		return r.value, r.err
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

// ========================================
// 4. BULKHEAD PATTERN
// ========================================
//
// The bulkhead pattern limits the number of concurrent calls to
// a dependency. Like bulkheads in a ship — if one compartment floods,
// the others stay dry.
//
// This prevents one slow dependency from consuming all resources
// (goroutines, connections, memory) and taking down the whole service.

// Bulkhead limits concurrent access to a resource.
type Bulkhead struct {
	name    string
	sem     chan struct{}  // Semaphore (buffered channel)
	timeout time.Duration // Max wait time to acquire a slot
}

// NewBulkhead creates a bulkhead with the given concurrency limit.
func NewBulkhead(name string, maxConcurrent int, timeout time.Duration) *Bulkhead {
	return &Bulkhead{
		name:    name,
		sem:     make(chan struct{}, maxConcurrent),
		timeout: timeout,
	}
}

// ErrBulkheadFull is returned when all slots are occupied.
var ErrBulkheadFull = errors.New("bulkhead is full")

// Execute runs fn within the bulkhead's concurrency limit.
func (b *Bulkhead) Execute(fn func() error) error {
	// Try to acquire a slot
	select {
	case b.sem <- struct{}{}:
		// Slot acquired — execute the function
		defer func() { <-b.sem }() // Release the slot when done
		return fn()
	case <-time.After(b.timeout):
		// Couldn't acquire a slot in time
		return fmt.Errorf("%w: %s (timeout: %s)", ErrBulkheadFull, b.name, b.timeout)
	}
}

// ========================================
// Simulated external service (for testing)
// ========================================

// externalService simulates a flaky external service.
type externalService struct {
	mu           sync.Mutex
	failureRate  float64 // 0.0 to 1.0
	minLatency   time.Duration
	maxLatency   time.Duration
	callCount    int
}

func newExternalService(failureRate float64, minLatency, maxLatency time.Duration) *externalService {
	return &externalService{
		failureRate: failureRate,
		minLatency:  minLatency,
		maxLatency:  maxLatency,
	}
}

func (s *externalService) Call() error {
	s.mu.Lock()
	s.callCount++
	count := s.callCount
	s.mu.Unlock()

	// Simulate latency
	latency := s.minLatency + time.Duration(rand.Int63n(int64(s.maxLatency-s.minLatency)))
	time.Sleep(latency)

	// Simulate failures
	if rand.Float64() < s.failureRate {
		return fmt.Errorf("service error on call #%d (latency: %v)", count, latency)
	}
	return nil
}

func (s *externalService) SetFailureRate(rate float64) {
	s.mu.Lock()
	s.failureRate = rate
	s.mu.Unlock()
}

func main() {
	fmt.Println("=== Week 19, Lesson 1: Resilience Patterns ===")
	fmt.Println()

	// ========================================
	// Demo 1: Circuit Breaker
	// ========================================
	fmt.Println("--- 1. Circuit Breaker Pattern ---")
	fmt.Println()
	fmt.Println("Circuit breaker prevents cascading failures by stopping")
	fmt.Println("calls to a failing service.")
	fmt.Println()

	// Create a flaky service that fails 80% of the time
	flakyService := newExternalService(0.8, 10*time.Millisecond, 50*time.Millisecond)

	// Create a circuit breaker: opens after 3 failures, resets after 1 second
	cb := NewCircuitBreaker("payment-api", 3, 1*time.Second)

	fmt.Println("Phase 1: Service is failing (80% failure rate)")
	for i := 1; i <= 8; i++ {
		err := cb.Call(func() error {
			return flakyService.Call()
		})
		status := "OK"
		if err != nil {
			status = err.Error()
		}
		fmt.Printf("  Call %d: %s\n", i, status)
	}
	fmt.Printf("  Stats: %s\n", cb.Stats())
	fmt.Println()

	fmt.Println("Phase 2: Service recovers, waiting for circuit to half-open...")
	flakyService.SetFailureRate(0.0) // Service recovers
	time.Sleep(1100 * time.Millisecond)

	for i := 9; i <= 12; i++ {
		err := cb.Call(func() error {
			return flakyService.Call()
		})
		status := "OK"
		if err != nil {
			status = err.Error()
		}
		fmt.Printf("  Call %d: %s\n", i, status)
	}
	fmt.Printf("  Stats: %s\n", cb.Stats())
	fmt.Println()

	// ========================================
	// Demo 2: Retry with Exponential Backoff
	// ========================================
	fmt.Println("--- 2. Retry with Exponential Backoff ---")
	fmt.Println()
	fmt.Println("Retry handles transient failures by trying again")
	fmt.Println("with increasing delays between attempts.")
	fmt.Println()

	attempt := 0
	ctx := context.Background()

	// Function that fails twice, then succeeds
	fmt.Println("Scenario: fails twice, then succeeds")
	err := Retry(ctx, DefaultRetryConfig(), func() error {
		attempt++
		if attempt <= 2 {
			return fmt.Errorf("transient error (attempt %d)", attempt)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("  Final result: FAILED — %v\n", err)
	} else {
		fmt.Printf("  Final result: SUCCESS on attempt %d\n", attempt)
	}
	fmt.Println()

	// Function that always fails
	fmt.Println("Scenario: always fails")
	err = Retry(ctx, RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   50 * time.Millisecond,
		MaxDelay:    200 * time.Millisecond,
		Jitter:      true,
	}, func() error {
		return errors.New("permanent failure")
	})
	fmt.Printf("  Final result: %v\n", err)
	fmt.Println()

	// ========================================
	// Demo 3: Timeout Pattern
	// ========================================
	fmt.Println("--- 3. Timeout Pattern ---")
	fmt.Println()
	fmt.Println("Timeouts prevent indefinite blocking when a service hangs.")
	fmt.Println()

	// Fast operation (completes within timeout)
	fmt.Println("Scenario: fast operation (50ms) with 200ms timeout")
	result, err := WithTimeout(200*time.Millisecond, func() (interface{}, error) {
		time.Sleep(50 * time.Millisecond)
		return "fast result", nil
	})
	if err != nil {
		fmt.Printf("  Result: ERROR — %v\n", err)
	} else {
		fmt.Printf("  Result: %v\n", result)
	}
	fmt.Println()

	// Slow operation (exceeds timeout)
	fmt.Println("Scenario: slow operation (500ms) with 100ms timeout")
	result, err = WithTimeout(100*time.Millisecond, func() (interface{}, error) {
		time.Sleep(500 * time.Millisecond)
		return "slow result", nil
	})
	if err != nil {
		fmt.Printf("  Result: ERROR — %v\n", err)
	} else {
		fmt.Printf("  Result: %v\n", result)
	}
	fmt.Println()

	// ========================================
	// Demo 4: Bulkhead Pattern
	// ========================================
	fmt.Println("--- 4. Bulkhead Pattern ---")
	fmt.Println()
	fmt.Println("Bulkhead limits concurrent calls to a dependency,")
	fmt.Println("preventing one slow service from consuming all resources.")
	fmt.Println()

	// Create a bulkhead allowing max 3 concurrent calls
	bh := NewBulkhead("database", 3, 500*time.Millisecond)

	fmt.Println("Launching 6 concurrent requests with bulkhead (max 3 concurrent):")

	var wg sync.WaitGroup
	for i := 1; i <= 6; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			start := time.Now()
			err := bh.Execute(func() error {
				// Simulate a 200ms database query
				time.Sleep(200 * time.Millisecond)
				return nil
			})
			elapsed := time.Since(start).Round(time.Millisecond)
			if err != nil {
				fmt.Printf("  Request %d: REJECTED after %v (%v)\n", id, elapsed, err)
			} else {
				fmt.Printf("  Request %d: OK in %v\n", id, elapsed)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println()

	// ========================================
	// Combining Patterns
	// ========================================
	fmt.Println("--- Combining Resilience Patterns ---")
	fmt.Println()
	fmt.Println("In production, you combine patterns:")
	fmt.Println()
	fmt.Println("  request")
	fmt.Println("    |")
	fmt.Println("    v")
	fmt.Println("  [Bulkhead] — limit concurrency")
	fmt.Println("    |")
	fmt.Println("    v")
	fmt.Println("  [Circuit Breaker] — fail fast if service is down")
	fmt.Println("    |")
	fmt.Println("    v")
	fmt.Println("  [Retry + Backoff] — handle transient failures")
	fmt.Println("    |")
	fmt.Println("    v")
	fmt.Println("  [Timeout] — prevent indefinite blocking")
	fmt.Println("    |")
	fmt.Println("    v")
	fmt.Println("  external service call")
	fmt.Println()

	// Example: combined resilient call
	fmt.Println("Combined resilient call example:")
	cb2 := NewCircuitBreaker("api", 3, 2*time.Second)
	bh2 := NewBulkhead("api-pool", 5, 1*time.Second)
	svc := newExternalService(0.3, 20*time.Millisecond, 80*time.Millisecond)

	resilientCall := func() error {
		// Layer 1: Bulkhead
		return bh2.Execute(func() error {
			// Layer 2: Circuit Breaker
			return cb2.Call(func() error {
				// Layer 3: Retry
				return Retry(context.Background(), RetryConfig{
					MaxAttempts: 2,
					BaseDelay:   50 * time.Millisecond,
					MaxDelay:    200 * time.Millisecond,
					Jitter:      true,
				}, func() error {
					// Layer 4: Timeout
					_, err := WithTimeout(100*time.Millisecond, func() (interface{}, error) {
						return nil, svc.Call()
					})
					return err
				})
			})
		})
	}

	for i := 1; i <= 5; i++ {
		err := resilientCall()
		status := "OK"
		if err != nil {
			status = err.Error()
		}
		fmt.Printf("  Resilient call %d: %s\n", i, status)
	}
	fmt.Println()

	fmt.Println("--- Summary ---")
	fmt.Println()
	fmt.Println("1. CIRCUIT BREAKER: Stops calling a failing service")
	fmt.Println("   Use when: dependency has extended outages")
	fmt.Println()
	fmt.Println("2. RETRY: Handles transient, short-lived failures")
	fmt.Println("   Use when: failures are temporary (network blips)")
	fmt.Println()
	fmt.Println("3. TIMEOUT: Prevents indefinite blocking")
	fmt.Println("   Use when: dependencies might hang")
	fmt.Println()
	fmt.Println("4. BULKHEAD: Limits concurrent access")
	fmt.Println("   Use when: you need to protect shared resources")
}
