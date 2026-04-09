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
// Week 14, Lesson 5 (Mini-Project): TCP Chat Server
// ========================================
// A multi-client TCP chat server where messages are broadcast to
// all connected clients. Features:
//   - Multiple simultaneous clients (one goroutine each)
//   - Message broadcasting to all clients
//   - Join/leave notifications
//   - Commands: /nick, /quit, /list, /help, /msg
//   - Client and server modes in one binary
//
// Usage:
//   go run main.go server              # Start the chat server
//   go run main.go server 8888         # Custom port
//   go run main.go client              # Connect as client
//   go run main.go client localhost:8888
//
// Quick test with multiple terminals:
//   Terminal 1: go run main.go server
//   Terminal 2: go run main.go client
//   Terminal 3: go run main.go client
// ========================================

// ========================================
// Chat Server Types
// ========================================

// Client represents a connected chat client.
type Client struct {
	conn     net.Conn
	nickname string
	outgoing chan string
	quit     chan struct{}
}

// ChatServer manages all connected clients and message routing.
type ChatServer struct {
	mu         sync.RWMutex
	clients    map[*Client]bool
	broadcast  chan string
	register   chan *Client
	unregister chan *Client
	quit       chan struct{}
}

// NewChatServer creates a new chat server.
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan string, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		quit:       make(chan struct{}),
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "server":
		port := "9000"
		if len(os.Args) > 2 {
			port = os.Args[2]
		}
		runServer(port)

	case "client":
		addr := "localhost:9000"
		if len(os.Args) > 2 {
			addr = os.Args[2]
		}
		runClient(addr)

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("========================================")
	fmt.Println("TCP Chat Server & Client")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go server [port]        Start chat server (default: 9000)")
	fmt.Println("  go run main.go client [addr:port]   Connect as client (default: localhost:9000)")
	fmt.Println()
	fmt.Println("Chat commands:")
	fmt.Println("  /nick <name>   Change your nickname")
	fmt.Println("  /list          List connected users")
	fmt.Println("  /msg <user> <text>  Send private message")
	fmt.Println("  /help          Show help")
	fmt.Println("  /quit          Disconnect")
}

// ========================================
// Server Implementation
// ========================================

func runServer(port string) {
	server := NewChatServer()

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}

	fmt.Println("========================================")
	fmt.Println("Chat Server Started")
	fmt.Println("========================================")
	fmt.Printf("Listening on port %s\n", port)
	fmt.Println("Waiting for clients...")
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	// Start the message router
	go server.run()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down server...")
		close(server.quit)
		listener.Close()
	}()

	// Accept loop
	clientNum := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-server.quit:
				server.shutdown()
				return
			default:
				log.Printf("Accept error: %v\n", err)
				continue
			}
		}

		clientNum++
		client := &Client{
			conn:     conn,
			nickname: fmt.Sprintf("user%d", clientNum),
			outgoing: make(chan string, 64),
			quit:     make(chan struct{}),
		}

		server.register <- client
		go server.handleClient(client)
	}
}

// run is the main server event loop that routes messages.
func (s *ChatServer) run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			count := len(s.clients)
			s.mu.Unlock()
			log.Printf("Client connected: %s (%s) [%d online]\n",
				client.nickname, client.conn.RemoteAddr(), count)
			s.broadcast <- fmt.Sprintf("*** %s has joined the chat (%d online) ***",
				client.nickname, count)

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.outgoing)
			}
			count := len(s.clients)
			s.mu.Unlock()
			log.Printf("Client disconnected: %s [%d online]\n", client.nickname, count)
			s.broadcast <- fmt.Sprintf("*** %s has left the chat (%d online) ***",
				client.nickname, count)

		case message := <-s.broadcast:
			s.mu.RLock()
			for client := range s.clients {
				select {
				case client.outgoing <- message:
				default:
					// Client's buffer is full — skip this message
					log.Printf("Dropped message for slow client: %s\n", client.nickname)
				}
			}
			s.mu.RUnlock()

		case <-s.quit:
			return
		}
	}
}

// handleClient manages a single client connection.
func (s *ChatServer) handleClient(client *Client) {
	defer func() {
		s.unregister <- client
		client.conn.Close()
	}()

	// Send welcome message
	client.outgoing <- fmt.Sprintf("Welcome to the chat, %s!", client.nickname)
	client.outgoing <- "Type /help for available commands."
	client.outgoing <- fmt.Sprintf("There are %d user(s) online.", s.clientCount())

	// Start writer goroutine
	go s.clientWriter(client)

	// Reader loop
	reader := bufio.NewReader(client.conn)
	for {
		client.conn.SetReadDeadline(time.Now().Add(10 * time.Minute))
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					client.outgoing <- "Connection timed out."
				}
			}
			return
		}

		message := strings.TrimSpace(line)
		if message == "" {
			continue
		}

		// Handle commands
		if strings.HasPrefix(message, "/") {
			if s.handleCommand(client, message) {
				return // Client wants to quit
			}
			continue
		}

		// Broadcast the message
		s.broadcast <- fmt.Sprintf("[%s] %s", client.nickname, message)
	}
}

// clientWriter sends messages from the outgoing channel to the client.
func (s *ChatServer) clientWriter(client *Client) {
	for message := range client.outgoing {
		fmt.Fprintf(client.conn, "%s\n", message)
	}
}

// handleCommand processes a chat command. Returns true if client should disconnect.
func (s *ChatServer) handleCommand(client *Client, cmd string) bool {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "/quit":
		client.outgoing <- "Goodbye!"
		return true

	case "/nick":
		if len(parts) < 2 {
			client.outgoing <- "Usage: /nick <new_name>"
			return false
		}
		newNick := parts[1]

		// Validate nickname
		if len(newNick) > 20 {
			client.outgoing <- "Nickname too long (max 20 characters)."
			return false
		}
		if strings.ContainsAny(newNick, " \t\n/") {
			client.outgoing <- "Nickname cannot contain spaces or /."
			return false
		}

		// Check for duplicates
		s.mu.RLock()
		for c := range s.clients {
			if c.nickname == newNick && c != client {
				client.outgoing <- fmt.Sprintf("Nickname %q is already taken.", newNick)
				s.mu.RUnlock()
				return false
			}
		}
		s.mu.RUnlock()

		oldNick := client.nickname
		client.nickname = newNick
		s.broadcast <- fmt.Sprintf("*** %s is now known as %s ***", oldNick, newNick)

	case "/list":
		s.mu.RLock()
		users := make([]string, 0, len(s.clients))
		for c := range s.clients {
			marker := ""
			if c == client {
				marker = " (you)"
			}
			users = append(users, c.nickname+marker)
		}
		s.mu.RUnlock()

		client.outgoing <- fmt.Sprintf("Online users (%d):", len(users))
		for _, u := range users {
			client.outgoing <- fmt.Sprintf("  - %s", u)
		}

	case "/msg":
		if len(parts) < 3 {
			client.outgoing <- "Usage: /msg <nickname> <message>"
			return false
		}
		targetNick := parts[1]
		privateMsg := strings.Join(parts[2:], " ")

		s.mu.RLock()
		var target *Client
		for c := range s.clients {
			if c.nickname == targetNick {
				target = c
				break
			}
		}
		s.mu.RUnlock()

		if target == nil {
			client.outgoing <- fmt.Sprintf("User %q not found.", targetNick)
		} else {
			target.outgoing <- fmt.Sprintf("[PM from %s] %s", client.nickname, privateMsg)
			client.outgoing <- fmt.Sprintf("[PM to %s] %s", targetNick, privateMsg)
		}

	case "/help":
		client.outgoing <- "Available commands:"
		client.outgoing <- "  /nick <name>         Change your nickname"
		client.outgoing <- "  /list                List online users"
		client.outgoing <- "  /msg <user> <text>   Send private message"
		client.outgoing <- "  /help                Show this help"
		client.outgoing <- "  /quit                Disconnect"

	default:
		client.outgoing <- fmt.Sprintf("Unknown command: %s (type /help)", parts[0])
	}

	return false
}

// clientCount returns the number of connected clients.
func (s *ChatServer) clientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// shutdown disconnects all clients.
func (s *ChatServer) shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for client := range s.clients {
		fmt.Fprintf(client.conn, "Server is shutting down. Goodbye!\n")
		client.conn.Close()
	}
	fmt.Printf("Disconnected %d client(s).\n", len(s.clients))
}

// ========================================
// Client Implementation
// ========================================

func runClient(addr string) {
	fmt.Println("========================================")
	fmt.Println("Chat Client")
	fmt.Println("========================================")
	fmt.Printf("Connecting to %s...\n\n", addr)

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		log.Fatalf("Connection failed: %v\n", err)
	}
	defer conn.Close()

	fmt.Println("Connected! Type messages and press Enter.")
	fmt.Println("Commands: /nick <name>, /list, /msg <user> <text>, /quit, /help")
	fmt.Println()

	done := make(chan struct{})

	// Read from server and print to stdout
	go func() {
		defer close(done)
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Println("\nDisconnected from server.")
				} else {
					fmt.Printf("\nConnection error: %v\n", err)
				}
				return
			}
			// Clear current input line, print server message, then prompt
			fmt.Printf("\r%s> ", strings.TrimSpace(line))
		}
	}()

	// Read from stdin and send to server
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					fmt.Fprintf(conn, "/quit\n")
				}
				return
			}

			message := strings.TrimSpace(line)
			if message == "" {
				continue
			}

			_, err = fmt.Fprintf(conn, "%s\n", message)
			if err != nil {
				fmt.Printf("Send error: %v\n", err)
				return
			}

			if message == "/quit" {
				return
			}
		}
	}()

	// Wait for server disconnect
	<-done
}
