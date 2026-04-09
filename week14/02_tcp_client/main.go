package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// ========================================
// Week 14, Lesson 2: TCP Client
// ========================================
// A TCP client connects to a server, sends messages, and reads
// responses. This lesson covers:
//   - net.Dial: connecting to a server
//   - Sending messages and reading responses
//   - Connection timeouts
//   - Reconnection logic
//   - Error handling for network operations
//
// Usage:
//   go run main.go                     # Connect to localhost:9000
//   go run main.go localhost:8080      # Connect to custom address
//   go run main.go demo                # Run automated demo
//
// First start the TCP server from Lesson 1:
//   cd ../01_tcp_server && go run main.go
// ========================================

func main() {
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		runDemo()
		return
	}

	addr := "localhost:9000"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}

	fmt.Println("========================================")
	fmt.Println("TCP Client")
	fmt.Println("========================================")
	fmt.Printf("Connecting to %s...\n\n", addr)

	// ========================================
	// 1. Basic Connection with net.Dial
	// ========================================
	// net.Dial connects to the specified network address.
	// It returns a net.Conn and an error.
	// The connection implements io.Reader and io.Writer.

	conn, err := connectWithRetry(addr, 3, 2*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect: %v\n", err)
	}
	defer conn.Close()

	fmt.Printf("Connected to %s\n", conn.RemoteAddr())
	fmt.Printf("Local address: %s\n\n", conn.LocalAddr())

	// ========================================
	// 2. Reading the Server's Welcome Message
	// ========================================
	serverReader := bufio.NewReader(conn)

	// Read welcome messages from server
	fmt.Println("--- Server says ---")
	for {
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		line, err := serverReader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break // No more welcome messages
			}
			break
		}
		fmt.Print(line)
	}
	fmt.Println("---")

	// Reset deadline for interactive use
	conn.SetReadDeadline(time.Time{})

	// ========================================
	// 3. Interactive Mode
	// ========================================
	fmt.Println("\nType messages to send (or 'quit' to exit):")

	stdinReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("\nEOF received, disconnecting.")
			}
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "quit" || input == "/quit" {
			// Send quit command to server
			fmt.Fprintf(conn, "%s\n", input)
			// Read final response
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			response, err := serverReader.ReadString('\n')
			if err == nil {
				fmt.Printf("Server: %s", response)
			}
			break
		}

		// ========================================
		// 4. Sending Messages
		// ========================================
		// Send the message to the server.
		// We add a newline because the server reads line by line.
		_, err = fmt.Fprintf(conn, "%s\n", input)
		if err != nil {
			log.Printf("Send error: %v\n", err)
			log.Println("Connection may be lost. Attempting reconnect...")
			break
		}

		// ========================================
		// 5. Reading the Response
		// ========================================
		// Set a read timeout so we don't block forever
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		response, err := serverReader.ReadString('\n')
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("(No response from server — timeout)")
			} else {
				log.Printf("Read error: %v\n", err)
				break
			}
		} else {
			fmt.Printf("Server: %s", response)
		}
	}

	fmt.Println("Disconnected.")
}

// ========================================
// 6. Connection with Timeout
// ========================================

// connectWithTimeout connects to an address with a timeout.
func connectWithTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	// net.DialTimeout is like net.Dial but with a connection timeout.
	// Without a timeout, Dial can block for minutes on unreachable hosts.
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return conn, nil
}

// ========================================
// 7. Reconnection Logic
// ========================================

// connectWithRetry attempts to connect with exponential backoff.
func connectWithRetry(addr string, maxRetries int, initialDelay time.Duration) (net.Conn, error) {
	var lastErr error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("Connection attempt %d/%d...\n", attempt, maxRetries)

		conn, err := connectWithTimeout(addr, 5*time.Second)
		if err == nil {
			if attempt > 1 {
				fmt.Println("Reconnected successfully!")
			}
			return conn, nil
		}

		lastErr = err
		fmt.Printf("  Failed: %v\n", err)

		if attempt < maxRetries {
			fmt.Printf("  Retrying in %v...\n", delay)
			time.Sleep(delay)
			// Exponential backoff: double the delay each time
			delay *= 2
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// ========================================
// 8. Automated Demo
// ========================================

// runDemo shows all TCP client concepts without requiring a running server.
func runDemo() {
	fmt.Println("========================================")
	fmt.Println("TCP Client — Automated Demo")
	fmt.Println("========================================")
	fmt.Println()

	// --- Concept 1: Basic net.Dial ---
	demonstrateBasicDial()

	// --- Concept 2: Timeout ---
	demonstrateTimeout()

	// --- Concept 3: Connection info ---
	demonstrateConnInfo()

	// --- Concept 4: Reconnection ---
	demonstrateReconnection()

	// --- Concept 5: Multiple connections ---
	demonstrateMultipleConns()

	fmt.Println("\n========================================")
	fmt.Println("TCP Client lesson complete!")
	fmt.Println("========================================")
}

func demonstrateBasicDial() {
	fmt.Println("========================================")
	fmt.Println("1. Basic net.Dial")
	fmt.Println("========================================")

	fmt.Print(`
  // Connect to a TCP server
  conn, err := net.Dial("tcp", "localhost:9000")
  if err != nil {
      log.Fatal(err)
  }
  defer conn.Close()

  // Send a message (conn implements io.Writer)
  fmt.Fprintf(conn, "Hello, server!\n")

  // Read response (conn implements io.Reader)
  reader := bufio.NewReader(conn)
  response, _ := reader.ReadString('\n')
  fmt.Println("Response:", response)
`)

	// Actually try connecting to a known service
	fmt.Println("Trying to connect to a well-known service...")
	conn, err := net.DialTimeout("tcp", "example.com:80", 3*time.Second)
	if err != nil {
		fmt.Printf("  Could not connect: %v\n", err)
		fmt.Println("  (This is normal if behind a firewall)")
	} else {
		fmt.Printf("  Connected to: %s\n", conn.RemoteAddr())
		fmt.Printf("  Local address: %s\n", conn.LocalAddr())

		// Send a simple HTTP request
		fmt.Fprintf(conn, "HEAD / HTTP/1.0\r\nHost: example.com\r\n\r\n")
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		if err == nil {
			fmt.Printf("  Response: %s", line)
		}
		conn.Close()
	}
}

func demonstrateTimeout() {
	fmt.Println("\n========================================")
	fmt.Println("2. Connection Timeouts")
	fmt.Println("========================================")

	fmt.Print(`
  Three types of timeouts:

  1. Connection timeout (how long to wait for connection):
     conn, err := net.DialTimeout("tcp", addr, 5*time.Second)

  2. Read deadline (how long to wait for data):
     conn.SetReadDeadline(time.Now().Add(10*time.Second))
     data, err := reader.ReadString('\n')
     // err will be a net.Error with Timeout() == true

  3. Write deadline (how long to wait for send buffer):
     conn.SetWriteDeadline(time.Now().Add(5*time.Second))
     _, err := conn.Write(data)

  4. Combined deadline (both read and write):
     conn.SetDeadline(time.Now().Add(30*time.Second))
`)

	// Demonstrate timeout to unreachable address
	fmt.Println("Demonstrating connection timeout (1s to unreachable address):")
	start := time.Now()
	_, err := net.DialTimeout("tcp", "192.0.2.1:9999", 1*time.Second)
	elapsed := time.Since(start)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Printf("  Timed out after %v (as expected)\n", elapsed)
		} else {
			fmt.Printf("  Connection failed: %v (%v)\n", err, elapsed)
		}
	}
}

func demonstrateConnInfo() {
	fmt.Println("\n========================================")
	fmt.Println("3. Connection Information")
	fmt.Println("========================================")

	fmt.Print(`
  net.Conn provides:
    conn.RemoteAddr()  — server's address (IP:port)
    conn.LocalAddr()   — client's address (IP:port)
    conn.Close()       — close the connection

  net.TCPConn (type assertion) adds:
    tcpConn.SetKeepAlive(true)
    tcpConn.SetKeepAlivePeriod(30*time.Second)
    tcpConn.SetNoDelay(true)  // disable Nagle's algorithm
    tcpConn.SetLinger(0)      // close immediately, don't wait
`)

	// Show network interfaces
	fmt.Println("Network interfaces on this machine:")
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			addrs, _ := iface.Addrs()
			if len(addrs) > 0 {
				fmt.Printf("  %s: ", iface.Name)
				addrStrs := []string{}
				for _, addr := range addrs {
					addrStrs = append(addrStrs, addr.String())
				}
				fmt.Println(strings.Join(addrStrs, ", "))
			}
		}
	}
}

func demonstrateReconnection() {
	fmt.Println("\n========================================")
	fmt.Println("4. Reconnection Logic")
	fmt.Println("========================================")

	os.Stdout.WriteString(`
  Production clients should handle disconnections gracefully:

  func connectWithRetry(addr string) (net.Conn, error) {
      delay := 1 * time.Second
      maxDelay := 30 * time.Second

      for {
          conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
          if err == nil {
              return conn, nil
          }

          log.Printf("Connection failed: %v, retrying in %v", err, delay)
          time.Sleep(delay)

          // Exponential backoff with max
          delay *= 2
          if delay > maxDelay {
              delay = maxDelay
          }
      }
  }

  Key patterns:
    - Exponential backoff (1s, 2s, 4s, 8s, ...)
    - Maximum retry delay (cap at 30s)
    - Jitter (add random delay to prevent thundering herd)
    - Maximum retry count (give up eventually)
    - Context cancellation (stop retrying on shutdown)
`)
}

func demonstrateMultipleConns() {
	fmt.Println("========================================")
	fmt.Println("5. Multiple Connections")
	fmt.Println("========================================")

	fmt.Print(`
  A client can maintain multiple connections:

  // Connection pool pattern
  type ConnPool struct {
      mu    sync.Mutex
      conns []net.Conn
      addr  string
      max   int
  }

  func (p *ConnPool) Get() (net.Conn, error) {
      p.mu.Lock()
      defer p.mu.Unlock()
      if len(p.conns) > 0 {
          conn := p.conns[len(p.conns)-1]
          p.conns = p.conns[:len(p.conns)-1]
          return conn, nil
      }
      return net.Dial("tcp", p.addr)
  }

  func (p *ConnPool) Put(conn net.Conn) {
      p.mu.Lock()
      defer p.mu.Unlock()
      if len(p.conns) < p.max {
          p.conns = append(p.conns, conn)
      } else {
          conn.Close()
      }
  }

  Connection pooling benefits:
    - Avoids TCP handshake overhead for each request
    - Limits total number of connections
    - Reuses established connections
    - Used by http.Client, database drivers, etc.
`)
}
