package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ========================================
// Week 12, Lesson 3: Signal Handling
// ========================================
// Signals are asynchronous notifications sent to a process by the
// operating system. In Go, the os/signal package lets you catch
// and handle signals for graceful shutdown, cleanup, and more.
//
// Common signals:
//   SIGINT  (2)  — Ctrl+C, interrupt
//   SIGTERM (15) — Termination request (e.g., from kill command)
//   SIGHUP  (1)  — Terminal hangup, often used for config reload
//   SIGUSR1 (10) — User-defined signal 1
//   SIGUSR2 (12) — User-defined signal 2
//   SIGKILL (9)  — Cannot be caught — immediate termination
//   SIGSTOP       — Cannot be caught — pause process
//
// Usage: go run main.go
// Then press Ctrl+C to see signal handling in action.
// ========================================

func main() {
	// ========================================
	// 1. Basic Signal Handling
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. Basic Signal Handling")
	fmt.Println("========================================")

	// signal.Notify sends incoming signals to a channel.
	// You create a buffered channel and register it for specific signals.
	// The channel MUST be buffered — otherwise signals can be missed.

	fmt.Println("\nDemonstrating basic signal setup...")
	fmt.Println("(We'll show more advanced patterns below)")

	// Create a buffered channel for signals
	sigChan := make(chan os.Signal, 1)

	// Register for SIGINT (Ctrl+C) and SIGTERM
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Stop receiving signals on this channel (cleanup for this demo)
	signal.Stop(sigChan)
	fmt.Println("Signal channel created and configured.")
	fmt.Println("In production, you'd block on <-sigChan to wait for signals.")

	// ========================================
	// 2. Graceful Shutdown Pattern
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Graceful Shutdown Pattern")
	fmt.Println("========================================")

	// The graceful shutdown pattern is one of the most important
	// patterns in Go server programming. It allows your application
	// to finish in-progress work before exiting.

	fmt.Println("\nStarting a simulated server with graceful shutdown...")
	fmt.Println("The server will run for 3 seconds, then auto-stop.")
	fmt.Println("(In production, it would wait for Ctrl+C)")

	runGracefulShutdownDemo()

	// ========================================
	// 3. signal.NotifyContext
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. signal.NotifyContext")
	fmt.Println("========================================")

	// signal.NotifyContext returns a context that is canceled when
	// one of the specified signals is received. This is the modern,
	// idiomatic way to handle signals in Go.

	fmt.Println("\nDemonstrating signal.NotifyContext...")
	fmt.Println("This creates a context that cancels on SIGINT/SIGTERM.")

	// Create a context that cancels on signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop() // Stop signal delivery when done

	// Simulate work with the context
	// In production, you'd pass this ctx to your server, database connections, etc.
	runContextDemo(ctx)

	// ========================================
	// 4. Cleanup on Exit
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Cleanup on Exit")
	fmt.Println("========================================")

	fmt.Println("\nDemonstrating cleanup pattern...")
	runCleanupDemo()

	// ========================================
	// 5. Multiple Signal Handlers
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Multiple Signal Handlers")
	fmt.Println("========================================")

	fmt.Println("\nDemonstrating different actions for different signals:")
	fmt.Println("  SIGINT  (Ctrl+C)   -> Graceful shutdown")
	fmt.Println("  SIGTERM (kill)     -> Immediate shutdown")
	fmt.Println("  SIGHUP  (hangup)  -> Reload configuration")
	fmt.Println("  SIGUSR1            -> Print status")

	runMultiSignalDemo()

	// ========================================
	// 6. Double-Signal Force Quit Pattern
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Double-Signal Force Quit Pattern")
	fmt.Println("========================================")

	// A common pattern: first Ctrl+C initiates graceful shutdown,
	// second Ctrl+C forces immediate exit.

	fmt.Println("\nThis pattern is used in production servers:")
	fmt.Println("  First  Ctrl+C -> 'Shutting down gracefully...'")
	fmt.Println("  Second Ctrl+C -> 'Forced shutdown!'")
	fmt.Println()
	fmt.Println("Example code (not running live):")
	fmt.Print(`
  sigChan := make(chan os.Signal, 1)
  signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

  // First signal: start graceful shutdown
  <-sigChan
  fmt.Println("Shutting down gracefully... (Ctrl+C again to force)")

  // Start graceful shutdown in a goroutine
  go func() {
      // ... cleanup work ...
      os.Exit(0)
  }()

  // Second signal: force quit
  <-sigChan
  fmt.Println("Forced shutdown!")
  os.Exit(1)
`)
	fmt.Println()

	// ========================================
	// 7. signal.Ignore and signal.Reset
	// ========================================
	fmt.Println("========================================")
	fmt.Println("7. signal.Ignore and signal.Reset")
	fmt.Println("========================================")

	// signal.Ignore causes the specified signals to be ignored.
	// The process won't be interrupted by these signals.
	fmt.Println("\nsignal.Ignore(syscall.SIGUSR1)")
	fmt.Println("  -> SIGUSR1 is now ignored by this process")
	signal.Ignore(syscall.SIGUSR1)

	// signal.Ignored checks if a signal is being ignored.
	fmt.Printf("SIGUSR1 is ignored: %v\n", signal.Ignored(syscall.SIGUSR1))
	fmt.Printf("SIGINT is ignored: %v\n", signal.Ignored(syscall.SIGINT))

	// signal.Reset restores the default behavior for signals.
	signal.Reset(syscall.SIGUSR1)
	fmt.Println("\nsignal.Reset(syscall.SIGUSR1)")
	fmt.Println("  -> SIGUSR1 restored to default behavior")

	// ========================================
	// 8. Real-World: HTTP Server Graceful Shutdown
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. Real-World: HTTP Server Pattern")
	fmt.Println("========================================")

	fmt.Println("\nThe standard pattern for a production HTTP server:")
	os.Stdout.WriteString(`
  func main() {
      // Create server
      srv := &http.Server{Addr: ":8080", Handler: mux}

      // Start server in a goroutine
      go func() {
          if err := srv.ListenAndServe(); err != http.ErrServerClosed {
              log.Fatalf("HTTP server error: %v", err)
          }
      }()

      // Wait for interrupt signal
      ctx, stop := signal.NotifyContext(context.Background(),
          syscall.SIGINT, syscall.SIGTERM)
      defer stop()
      <-ctx.Done()

      // Graceful shutdown with 10s timeout
      shutdownCtx, cancel := context.WithTimeout(
          context.Background(), 10*time.Second)
      defer cancel()

      if err := srv.Shutdown(shutdownCtx); err != nil {
          log.Fatalf("Shutdown error: %v", err)
      }
      fmt.Println("Server stopped gracefully")
  }
`)
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("Signal Handling lesson complete!")
	fmt.Println("========================================")
}

// ========================================
// Helper Functions
// ========================================

// runGracefulShutdownDemo simulates a server that shuts down gracefully.
func runGracefulShutdownDemo() {
	// Create a signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Simulate some "connections" being handled
	var wg sync.WaitGroup
	done := make(chan struct{})

	// Simulate 3 in-progress tasks
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			duration := time.Duration(id) * 500 * time.Millisecond
			fmt.Printf("  Task %d: working for %v\n", id, duration)
			select {
			case <-time.After(duration):
				fmt.Printf("  Task %d: completed\n", id)
			case <-done:
				fmt.Printf("  Task %d: interrupted, cleaning up\n", id)
			}
		}(i)
	}

	// Use a timer to simulate receiving a signal after 1 second
	// In production, you'd wait on <-sigChan instead
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("\n  [Simulated shutdown signal received]")
		close(done) // Signal all tasks to stop
	}()

	// Wait for all tasks to finish
	wg.Wait()
	fmt.Println("  All tasks completed. Server shut down gracefully.")
}

// runContextDemo demonstrates using signal.NotifyContext.
func runContextDemo(ctx context.Context) {
	// Create a child context with a timeout for this demo
	demoCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Simulate work that respects context cancellation
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-demoCtx.Done():
			fmt.Printf("  Context done: %v (after %d ticks)\n", demoCtx.Err(), count)
			return
		case <-ticker.C:
			count++
			fmt.Printf("  Tick %d: still working...\n", count)
		}
	}
}

// runCleanupDemo demonstrates cleanup on exit.
func runCleanupDemo() {
	// Create a temporary file to demonstrate cleanup
	tmpFile, err := os.CreateTemp("", "signal-demo-*.txt")
	if err != nil {
		fmt.Printf("  Error creating temp file: %v\n", err)
		return
	}
	tmpPath := tmpFile.Name()
	tmpFile.WriteString("important data\n")
	tmpFile.Close()
	fmt.Printf("  Created temp file: %s\n", tmpPath)

	// Register cleanup function
	// In production, this would be in a signal handler
	cleanup := func() {
		fmt.Println("  Running cleanup...")

		// Close database connections (simulated)
		fmt.Println("    Closing database connections...")
		time.Sleep(50 * time.Millisecond)

		// Flush buffers (simulated)
		fmt.Println("    Flushing buffers...")
		time.Sleep(50 * time.Millisecond)

		// Remove temp files
		if err := os.Remove(tmpPath); err != nil {
			fmt.Printf("    Error removing temp file: %v\n", err)
		} else {
			fmt.Println("    Removed temp file.")
		}

		fmt.Println("  Cleanup complete!")
	}

	// Simulate receiving shutdown signal
	fmt.Println("  [Simulated shutdown signal]")
	cleanup()
}

// runMultiSignalDemo shows handling different signals differently.
func runMultiSignalDemo() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGUSR1)
	defer signal.Stop(sigChan)

	// Use a timeout for the demo
	timeout := time.After(1 * time.Second)

	fmt.Println("\n  Listening for signals for 1 second...")
	fmt.Printf("  (Send signals with: kill -SIGUSR1 %d)\n", os.Getpid())

	for {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGINT:
				fmt.Println("  Received SIGINT -> Starting graceful shutdown")
			case syscall.SIGTERM:
				fmt.Println("  Received SIGTERM -> Immediate shutdown")
			case syscall.SIGHUP:
				fmt.Println("  Received SIGHUP -> Reloading configuration")
			case syscall.SIGUSR1:
				fmt.Println("  Received SIGUSR1 -> Printing status")
			}
		case <-timeout:
			fmt.Println("  Demo timeout reached. No external signals received.")
			return
		}
	}
}
