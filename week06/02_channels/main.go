package main

import (
	"fmt"
	"time"
)

// ========================================
// Week 6, Lesson 2: Channels
// ========================================
// Channels are Go's primary mechanism for communication between
// goroutines. They provide a way to send and receive values
// between concurrent goroutines, ensuring safe data transfer.
//
// Go's concurrency philosophy: "Don't communicate by sharing
// memory; share memory by communicating."
// ========================================

func main() {
	// ========================================
	// 1. Unbuffered Channels
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. Unbuffered Channels")
	fmt.Println("========================================")

	// An unbuffered channel has no capacity. A send blocks until
	// another goroutine receives, and vice versa. This creates
	// a synchronization point between goroutines.

	ch := make(chan string) // Create an unbuffered channel of strings

	go func() {
		ch <- "Hello from goroutine!" // Send blocks until someone receives
	}()

	msg := <-ch // Receive blocks until someone sends
	fmt.Printf("Received: %s\n", msg)

	// Unbuffered channels synchronize sender and receiver
	fmt.Println("\nSynchronization example:")
	done := make(chan bool)

	go func() {
		fmt.Println("  Working...")
		time.Sleep(50 * time.Millisecond)
		fmt.Println("  Done working!")
		done <- true // Signal completion
	}()

	<-done // Wait for the signal
	fmt.Println("  Worker finished, continuing main")

	// ========================================
	// 2. Buffered Channels
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Buffered Channels")
	fmt.Println("========================================")

	// Buffered channels have a capacity. Sends only block when
	// the buffer is full. Receives only block when the buffer
	// is empty.

	buffered := make(chan int, 3) // Buffer capacity of 3

	// We can send up to 3 values without a receiver
	buffered <- 10
	buffered <- 20
	buffered <- 30
	// buffered <- 40 // This would BLOCK (deadlock) — buffer is full!

	fmt.Printf("Buffer length: %d, capacity: %d\n", len(buffered), cap(buffered))

	// Receive values (FIFO order)
	fmt.Printf("Received: %d\n", <-buffered)
	fmt.Printf("Received: %d\n", <-buffered)
	fmt.Printf("Received: %d\n", <-buffered)
	fmt.Printf("Buffer length after receiving: %d\n", len(buffered))

	// Practical use: a job queue
	fmt.Println("\nBuffered channel as job queue:")
	jobs := make(chan string, 5)

	// Producer: add jobs
	jobs <- "job-1"
	jobs <- "job-2"
	jobs <- "job-3"

	// Consumer: process jobs
	for i := 0; i < 3; i++ {
		job := <-jobs
		fmt.Printf("  Processing: %s\n", job)
	}

	// ========================================
	// 3. Channel Direction (Send-Only, Receive-Only)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Channel Direction")
	fmt.Println("========================================")

	// You can restrict a channel to send-only or receive-only
	// in function signatures. This improves code safety.
	//   chan<- T  — send-only channel
	//   <-chan T  — receive-only channel

	dataCh := make(chan string, 1)

	// producer only sends, consumer only receives
	go producer(dataCh)
	go consumer(dataCh)

	time.Sleep(100 * time.Millisecond)

	// Another example with return channels
	resultCh := make(chan int, 1)
	go squareWorker(5, resultCh)
	result := <-resultCh
	fmt.Printf("Square of 5 = %d\n", result)

	// ========================================
	// 4. Select Statement
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Select Statement")
	fmt.Println("========================================")

	// `select` lets a goroutine wait on multiple channel operations.
	// It's like a switch statement but for channels.

	ch1 := make(chan string, 1)
	ch2 := make(chan string, 1)

	go func() {
		time.Sleep(20 * time.Millisecond)
		ch1 <- "one"
	}()
	go func() {
		time.Sleep(10 * time.Millisecond)
		ch2 <- "two"
	}()

	// Wait for whichever channel receives first
	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Printf("  Received from ch1: %s\n", msg1)
		case msg2 := <-ch2:
			fmt.Printf("  Received from ch2: %s\n", msg2)
		}
	}

	// Select with default (non-blocking)
	fmt.Println("\nNon-blocking select with default:")
	emptyCh := make(chan int, 1)

	select {
	case val := <-emptyCh:
		fmt.Printf("  Received: %d\n", val)
	default:
		fmt.Println("  No value ready — default case executed")
	}

	// ========================================
	// 5. Timeout Patterns
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Timeout Patterns")
	fmt.Println("========================================")

	// Using select with time.After for timeouts
	slowCh := make(chan string, 1)

	go func() {
		time.Sleep(200 * time.Millisecond) // Simulate slow work
		slowCh <- "slow result"
	}()

	select {
	case result := <-slowCh:
		fmt.Printf("  Got result: %s\n", result)
	case <-time.After(50 * time.Millisecond):
		fmt.Println("  Timeout! Operation took too long")
	}

	// Successful timeout pattern
	fastCh := make(chan string, 1)

	go func() {
		time.Sleep(10 * time.Millisecond) // Fast work
		fastCh <- "fast result"
	}()

	select {
	case result := <-fastCh:
		fmt.Printf("  Got result: %s\n", result)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("  Timeout!")
	}

	// ========================================
	// 6. Closing Channels
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Closing Channels")
	fmt.Println("========================================")

	// Closing a channel signals that no more values will be sent.
	// Receivers can detect a closed channel.

	numbers := make(chan int, 5)

	// Send some values and close
	go func() {
		for i := 1; i <= 5; i++ {
			numbers <- i
		}
		close(numbers) // Signal: no more values coming
	}()

	// Receiving from closed channel
	// The second return value (ok) is false when channel is closed and empty
	for {
		val, ok := <-numbers
		if !ok {
			fmt.Println("  Channel closed!")
			break
		}
		fmt.Printf("  Received: %d (ok=%v)\n", val, ok)
	}

	// Reading from a closed, empty channel returns the zero value immediately
	closedCh := make(chan int)
	close(closedCh)
	val := <-closedCh
	fmt.Printf("  Value from closed channel: %d (zero value)\n", val)

	// IMPORTANT: Only the sender should close a channel, never the receiver.
	// Sending to a closed channel causes a panic!

	// ========================================
	// 7. Range Over Channels
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. Range Over Channels")
	fmt.Println("========================================")

	// `range` on a channel receives values until the channel is closed.
	// This is the idiomatic way to consume all values from a channel.

	words := make(chan string)

	go func() {
		wordList := []string{"Go", "is", "awesome", "for", "concurrency"}
		for _, w := range wordList {
			words <- w
		}
		close(words) // Must close, or range will block forever
	}()

	fmt.Print("  Message: ")
	for word := range words {
		fmt.Printf("%s ", word)
	}
	fmt.Println()

	// Practical example: generating a sequence
	fmt.Println("\nFibonacci sequence via channel:")
	fibs := generateFibonacci(10)
	for val := range fibs {
		fmt.Printf("  %d", val)
	}
	fmt.Println()

	// ========================================
	// 8. Putting It All Together: Pipeline Pattern
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. Pipeline Pattern")
	fmt.Println("========================================")

	// A pipeline is a series of stages connected by channels.
	// Each stage is a goroutine that receives from one channel,
	// processes data, and sends to another channel.

	// Stage 1: Generate numbers
	nums := generateNumbers(1, 10)

	// Stage 2: Square each number
	squared := squareNumbers(nums)

	// Stage 3: Filter even results
	evens := filterEven(squared)

	// Consume the pipeline
	fmt.Println("Even squares of 1-10:")
	for val := range evens {
		fmt.Printf("  %d", val)
	}
	fmt.Println()

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println("- Unbuffered channels synchronize sender and receiver")
	fmt.Println("- Buffered channels allow sends without an immediate receiver")
	fmt.Println("- chan<- T (send-only), <-chan T (receive-only) for safety")
	fmt.Println("- select waits on multiple channels simultaneously")
	fmt.Println("- time.After creates timeout patterns with select")
	fmt.Println("- close(ch) signals no more values; range reads until closed")
	fmt.Println("- Pipelines chain stages via channels for data processing")
}

// ========================================
// Helper Functions
// ========================================

// producer sends data to a send-only channel.
func producer(ch chan<- string) {
	ch <- "data from producer"
	fmt.Println("  Producer sent data")
}

// consumer receives data from a receive-only channel.
func consumer(ch <-chan string) {
	msg := <-ch
	fmt.Printf("  Consumer received: %s\n", msg)
}

// squareWorker computes the square of n and sends the result.
func squareWorker(n int, result chan<- int) {
	result <- n * n
}

// generateFibonacci sends the first n Fibonacci numbers to a channel.
func generateFibonacci(n int) <-chan int {
	ch := make(chan int)
	go func() {
		a, b := 0, 1
		for i := 0; i < n; i++ {
			ch <- a
			a, b = b, a+b
		}
		close(ch)
	}()
	return ch
}

// Pipeline stage: generate numbers in a range.
func generateNumbers(start, end int) <-chan int {
	out := make(chan int)
	go func() {
		for i := start; i <= end; i++ {
			out <- i
		}
		close(out)
	}()
	return out
}

// Pipeline stage: square each number.
func squareNumbers(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

// Pipeline stage: keep only even numbers.
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
		close(out)
	}()
	return out
}

