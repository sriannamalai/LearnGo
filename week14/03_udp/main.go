package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// ========================================
// Week 14, Lesson 3: UDP Networking
// ========================================
// UDP (User Datagram Protocol) is a connectionless protocol that
// sends individual packets (datagrams) without establishing a
// connection. Unlike TCP, UDP does not guarantee delivery, ordering,
// or duplicate protection — but it's faster and lighter.
//
// When to use UDP vs TCP:
//   UDP: Real-time games, video streaming, DNS queries, IoT sensors
//   TCP: Web, email, file transfer, databases — anything needing reliability
//
// Usage:
//   go run main.go                # Run the full demo
//   go run main.go server         # Run just the UDP server
//   go run main.go client         # Run just the UDP client
// ========================================

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "server":
			runServer(":9001")
		case "client":
			runClient("localhost:9001")
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Usage: go run main.go [server|client]")
			return
		}
		return
	}

	// Run the full demo with concepts
	fmt.Println("========================================")
	fmt.Println("UDP Networking in Go")
	fmt.Println("========================================")

	// ========================================
	// 1. TCP vs UDP Comparison
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("1. TCP vs UDP Comparison")
	fmt.Println("========================================")

	fmt.Print(`
  TCP (Transmission Control Protocol):
    + Connection-oriented (3-way handshake)
    + Reliable delivery (acknowledgments, retransmission)
    + Ordered delivery (sequence numbers)
    + Flow control and congestion control
    + Stream-based (no message boundaries)
    - Higher overhead and latency
    - More complex state management

  UDP (User Datagram Protocol):
    + Connectionless (no handshake needed)
    + Low overhead (8-byte header vs 20+ for TCP)
    + Preserves message boundaries (datagram-based)
    + No head-of-line blocking
    + Supports multicast and broadcast
    - No delivery guarantee (packets can be lost)
    - No ordering guarantee (packets can arrive out of order)
    - No congestion control (can flood the network)
    - Maximum datagram size ~65,507 bytes (practical: ~1,472 for MTU)

  UDP Header (8 bytes):
    [Source Port (2)] [Dest Port (2)] [Length (2)] [Checksum (2)]

  TCP Header (20+ bytes):
    [Source Port] [Dest Port] [Sequence Number] [Ack Number]
    [Flags] [Window Size] [Checksum] [Urgent Pointer] [Options...]
`)

	// ========================================
	// 2. Run Server and Client Together
	// ========================================
	fmt.Println("========================================")
	fmt.Println("2. UDP Server + Client Demo")
	fmt.Println("========================================")

	port := ":9001"
	var wg sync.WaitGroup
	serverReady := make(chan struct{})

	// Start server in background
	wg.Add(1)
	go func() {
		defer wg.Done()
		runDemoServer(port, serverReady)
	}()

	// Wait for server to be ready
	<-serverReady

	// Run client
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(100 * time.Millisecond) // Brief pause
		runDemoClient("localhost" + port)
	}()

	wg.Wait()

	// ========================================
	// 3. UDP Concepts Deep Dive
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Key UDP Concepts")
	fmt.Println("========================================")

	fmt.Print(`
  Datagram-based communication:
    - Each Send/Receive is one complete message
    - Unlike TCP streams, message boundaries are preserved
    - A 100-byte send results in a 100-byte receive

  No connection state:
    - Server doesn't "accept" connections
    - Each packet is independent
    - Server processes packets from any source
    - Client doesn't need to "connect" (but can for convenience)

  net.ListenPacket vs net.ListenUDP:
    - ListenPacket("udp", addr) returns a net.PacketConn (interface)
    - ListenUDP("udp", addr) returns a *net.UDPConn (concrete type)
    - UDPConn has additional methods like ReadFromUDP, WriteToUDP

  net.DialUDP:
    - "Connects" to a specific remote address (sets default peer)
    - Allows using Read/Write instead of ReadFrom/WriteTo
    - Kernel filters incoming packets (only from connected peer)
    - Does NOT establish a real connection (no handshake)
`)

	// ========================================
	// 4. Practical UDP Patterns
	// ========================================
	fmt.Println("========================================")
	fmt.Println("4. Practical UDP Patterns")
	fmt.Println("========================================")

	fmt.Print(`
  Request/Response Pattern (like DNS):
    client: WriteTo(request, serverAddr)
    server: ReadFrom(buf) -> process -> WriteTo(response, clientAddr)
    client: ReadFrom(buf) -> handle response

  Fire-and-Forget Pattern (logging, metrics):
    client: WriteTo(data, serverAddr)  // don't wait for response
    server: ReadFrom(buf) -> store data

  Reliable UDP (custom protocol):
    - Add sequence numbers to messages
    - Implement acknowledgments
    - Retry on timeout
    - Used by: QUIC (HTTP/3), game networking protocols

  Multicast Pattern (one-to-many):
    server: WriteTo(data, multicastAddr)
    clients: Join multicast group, receive copies
    Used for: service discovery, live streaming

  Common UDP-based protocols:
    - DNS (port 53) — domain name resolution
    - DHCP (ports 67/68) — IP address assignment
    - NTP (port 123) — time synchronization
    - QUIC (port 443) — HTTP/3 transport
    - RTP — real-time audio/video
    - TFTP — trivial file transfer
`)

	// ========================================
	// 5. Multicast Demo
	// ========================================
	fmt.Println("========================================")
	fmt.Println("5. Connected vs Unconnected UDP")
	fmt.Println("========================================")

	demonstrateConnectedUDP()

	fmt.Println("\n========================================")
	fmt.Println("UDP Networking lesson complete!")
	fmt.Println("========================================")
}

// ========================================
// UDP Server
// ========================================

// runServer runs a standalone UDP server.
func runServer(addr string) {
	fmt.Printf("Starting UDP server on %s\n", addr)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	// net.ListenPacket creates a UDP socket for receiving datagrams.
	// Unlike TCP, there's no "accept" — we just read packets.
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Listening on %s\n\n", conn.LocalAddr())

	buf := make([]byte, 65535) // Maximum UDP datagram size

	for {
		// ReadFrom blocks until a datagram arrives.
		// It returns the number of bytes read and the sender's address.
		n, clientAddr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			continue
		}

		message := string(buf[:n])
		fmt.Printf("Received %d bytes from %s: %q\n", n, clientAddr, message)

		// Send a response back to the client
		response := fmt.Sprintf("Echo: %s (received %d bytes)", message, n)
		_, err = conn.WriteTo([]byte(response), clientAddr)
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
		}
	}
}

// runDemoServer runs a UDP server for the demo (auto-stops).
func runDemoServer(addr string, ready chan<- struct{}) {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Printf("Server error: %v\n", err)
		close(ready)
		return
	}
	defer conn.Close()

	fmt.Printf("\n  Server listening on %s\n", conn.LocalAddr())
	close(ready)

	// Set a deadline so the server eventually stops
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 4096)
	messagesHandled := 0

	for messagesHandled < 5 {
		n, clientAddr, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			break
		}

		message := string(buf[:n])
		fmt.Printf("  Server received from %s: %q\n", clientAddr, message)

		response := fmt.Sprintf("Echo: %s", message)
		conn.WriteTo([]byte(response), clientAddr)
		messagesHandled++
	}

	fmt.Printf("  Server handled %d messages, stopping.\n", messagesHandled)
}

// ========================================
// UDP Client
// ========================================

// runClient runs a standalone UDP client.
func runClient(addr string) {
	fmt.Printf("Sending UDP messages to %s\n\n", addr)

	// Resolve the UDP address
	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		return
	}

	// net.DialUDP creates a "connected" UDP socket.
	// This means we can use Read/Write instead of ReadFrom/WriteTo.
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Connected to %s (local: %s)\n\n", conn.RemoteAddr(), conn.LocalAddr())

	messages := []string{
		"Hello, UDP server!",
		"UDP is connectionless",
		"Messages are datagrams",
		"No guaranteed delivery",
		"But it's fast!",
	}

	buf := make([]byte, 4096)

	for _, msg := range messages {
		// Write sends the datagram
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Send error: %v\n", err)
			continue
		}
		fmt.Printf("Sent: %q\n", msg)

		// Read the response
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("  No response: %v\n", err)
		} else {
			fmt.Printf("  Response: %q\n", string(buf[:n]))
		}
	}
}

// runDemoClient runs a client for the integrated demo.
func runDemoClient(addr string) {
	serverAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		fmt.Printf("  Client error: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Printf("  Client error: %v\n", err)
		return
	}
	defer conn.Close()

	messages := []string{"Hello", "World", "UDP", "Is", "Fast"}
	buf := make([]byte, 4096)

	for _, msg := range messages {
		conn.Write([]byte(msg))
		fmt.Printf("  Client sent: %q\n", msg)

		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, err := conn.Read(buf)
		if err == nil {
			fmt.Printf("  Client received: %q\n", string(buf[:n]))
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("  Client done.")
}

// ========================================
// Connected vs Unconnected UDP
// ========================================

func demonstrateConnectedUDP() {
	fmt.Print(`
  Unconnected UDP (net.ListenPacket / net.ListenUDP):
    - Can send to ANY address with WriteTo
    - Can receive from ANY address with ReadFrom
    - Must specify destination for each send
    - Typical for servers

  Connected UDP (net.DialUDP):
    - Sets a default peer address
    - Can use Read/Write (simpler API)
    - Kernel filters: only receives from connected peer
    - Gets ICMP errors back (e.g., port unreachable)
    - Typical for clients

  Code comparison:

    // Unconnected (server-style):
    conn, _ := net.ListenPacket("udp", ":9001")
    buf := make([]byte, 1024)
    n, addr, _ := conn.ReadFrom(buf)        // Who sent this?
    conn.WriteTo(response, addr)              // Send back to them

    // Connected (client-style):
    conn, _ := net.DialUDP("udp", nil, serverAddr)
    conn.Write([]byte("hello"))               // Goes to serverAddr
    n, _ := conn.Read(buf)                    // Only from serverAddr
`)

	// Quick demonstration of both styles
	fmt.Println("  Demonstrating packet size limits:")
	fmt.Printf("  Maximum theoretical UDP payload: 65,507 bytes\n")
	fmt.Printf("  Typical safe size (avoid fragmentation): 1,472 bytes\n")
	fmt.Printf("  Common sizes: DNS=512, QUIC=1,280, custom=1,400\n")
}
