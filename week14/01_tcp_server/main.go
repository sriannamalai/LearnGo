package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ========================================
// Week 14, Lesson 1: TCP Server
// ========================================
// TCP (Transmission Control Protocol) provides reliable, ordered,
// and error-checked delivery of data between applications. In Go,
// the net package provides everything you need to build TCP servers.
//
// This lesson covers:
//   - net.Listen: binding to a port
//   - Accepting connections in a loop
//   - Handling each connection in its own goroutine
//   - Reading/writing with bufio
//   - Building a simple echo server
//   - Graceful shutdown
//
// Usage:
//   go run main.go               # Run the echo server on port 9000
//   go run main.go 8080          # Run on custom port
//
// Test with:
//   nc localhost 9000            # or: telnet localhost 9000
//   echo "hello" | nc localhost 9000
// ========================================

func main() {
	port := "9000"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	fmt.Println("========================================")
	fmt.Println("TCP Echo Server")
	fmt.Println("========================================")
	fmt.Printf("Starting server on port %s...\n\n", port)

	// ========================================
	// 1. Creating a TCP Listener
	// ========================================
	// net.Listen creates a TCP listener bound to an address.
	// The address format is "host:port". Use ":port" to listen
	// on all interfaces, or "localhost:port" for local only.
	//
	// net.Listen returns a net.Listener interface with:
	//   Accept() — blocks until a new connection arrives
	//   Close()  — stops listening
	//   Addr()   — returns the listener's address

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to start listener: %v\n", err)
	}
	defer listener.Close()

	fmt.Printf("Listening on %s\n", listener.Addr())
	fmt.Println("Waiting for connections...")
	fmt.Println("Test with: nc localhost " + port)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	// ========================================
	// 2. Graceful Shutdown Setup
	// ========================================
	// We use signal handling to shut down gracefully,
	// allowing in-progress connections to finish.

	var wg sync.WaitGroup
	quit := make(chan struct{})

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down server...")
		close(quit) // Signal all goroutines to stop
		listener.Close() // Unblock Accept()
	}()

	// ========================================
	// 3. Accept Loop
	// ========================================
	// The accept loop runs forever, accepting new connections.
	// Each connection gets its own goroutine for handling.

	connID := 0
	for {
		// Accept blocks until a new connection arrives.
		// It returns a net.Conn (the connection) and an error.
		conn, err := listener.Accept()
		if err != nil {
			// Check if we're shutting down
			select {
			case <-quit:
				fmt.Println("Accept loop stopped.")
				goto shutdown
			default:
				log.Printf("Accept error: %v\n", err)
				continue
			}
		}

		connID++
		wg.Add(1)

		// ========================================
		// 4. Handle Each Connection in a Goroutine
		// ========================================
		// This is the standard pattern: one goroutine per connection.
		// Goroutines are cheap (~2KB stack), so thousands of
		// concurrent connections are feasible.

		go func(id int) {
			defer wg.Done()
			handleConnection(conn, id, quit)
		}(connID)
	}

shutdown:
	// Wait for all active connections to finish
	fmt.Println("Waiting for active connections to close...")
	wg.Wait()
	fmt.Println("Server stopped gracefully.")
}

// ========================================
// 5. Connection Handler
// ========================================

// handleConnection processes a single TCP connection.
// This is an echo server: it reads lines and sends them back.
func handleConnection(conn net.Conn, id int, quit <-chan struct{}) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	log.Printf("[Conn %d] New connection from %s\n", id, remoteAddr)

	// Send welcome message
	fmt.Fprintf(conn, "Welcome to the Go TCP Echo Server! (Connection #%d)\n", id)
	fmt.Fprintf(conn, "Type messages and press Enter. Type 'quit' to disconnect.\n")
	fmt.Fprintf(conn, "Commands: /time, /info, /quit\n\n")

	// ========================================
	// 6. Reading with bufio
	// ========================================
	// bufio.NewReader wraps the connection with buffered reading.
	// This is more efficient than reading one byte at a time
	// and provides convenient methods like ReadString.

	reader := bufio.NewReader(conn)

	// Set initial read deadline — connections shouldn't hang forever
	conn.SetDeadline(time.Now().Add(5 * time.Minute))

	for {
		// Check if server is shutting down
		select {
		case <-quit:
			fmt.Fprintf(conn, "Server is shutting down. Goodbye!\n")
			log.Printf("[Conn %d] Disconnected (server shutdown)\n", id)
			return
		default:
		}

		// Read a line from the client
		// ReadString reads until the delimiter (including it)
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("[Conn %d] Client disconnected (EOF)\n", id)
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Fprintf(conn, "Connection timed out. Goodbye!\n")
				log.Printf("[Conn %d] Connection timed out\n", id)
			} else {
				log.Printf("[Conn %d] Read error: %v\n", id, err)
			}
			return
		}

		// Reset deadline on activity
		conn.SetDeadline(time.Now().Add(5 * time.Minute))

		// Trim whitespace
		message := strings.TrimSpace(line)
		if message == "" {
			continue
		}

		log.Printf("[Conn %d] Received: %q\n", id, message)

		// ========================================
		// 7. Processing Commands
		// ========================================
		switch {
		case message == "/quit" || message == "quit":
			fmt.Fprintf(conn, "Goodbye!\n")
			log.Printf("[Conn %d] Client quit\n", id)
			return

		case message == "/time":
			fmt.Fprintf(conn, "Server time: %s\n", time.Now().Format(time.RFC3339))

		case message == "/info":
			fmt.Fprintf(conn, "Connection ID: %d\n", id)
			fmt.Fprintf(conn, "Remote address: %s\n", remoteAddr)
			fmt.Fprintf(conn, "Local address: %s\n", conn.LocalAddr().String())

		default:
			// ========================================
			// 8. Echo Response
			// ========================================
			// Write the message back to the client.
			// fmt.Fprintf writes formatted data to any io.Writer,
			// and net.Conn implements io.Writer.

			response := fmt.Sprintf("Echo: %s\n", message)
			_, err := conn.Write([]byte(response))
			if err != nil {
				log.Printf("[Conn %d] Write error: %v\n", id, err)
				return
			}
		}
	}
}
