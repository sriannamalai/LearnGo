package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// ========================================
// Week 8, Lesson 1: Basic HTTP Server
// ========================================
// Learn how to create a web server in Go using the standard library.
// Go's net/http package is powerful enough for production use — no
// framework required!
//
// Run this program:
//   go run .
//
// Then visit these URLs in your browser or use curl:
//   http://localhost:8080/
//   http://localhost:8080/hello
//   http://localhost:8080/time
//   http://localhost:8080/headers
//   http://localhost:8080/greet?name=Sri

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 8: Basic HTTP Server")
	fmt.Println("========================================")

	// ========================================
	// 1. The simplest handler: http.HandleFunc
	// ========================================
	// http.HandleFunc registers a function to handle requests for a path.
	// The function receives:
	//   - w http.ResponseWriter: used to write the response back to the client
	//   - r *http.Request: contains all information about the incoming request

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/time", timeHandler)
	http.HandleFunc("/headers", headersHandler)
	http.HandleFunc("/greet", greetHandler)

	// ========================================
	// 2. Inline handler using anonymous function
	// ========================================
	// For simple handlers, you can use an anonymous function directly.
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Set a custom header
		w.Header().Set("Content-Type", "application/json")
		// Write JSON response
		fmt.Fprintf(w, `{"status": "ok", "uptime": "running"}`)
	})

	// ========================================
	// 3. Serving static content (HTML)
	// ========================================
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		// Set content type to HTML so the browser renders it properly
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `
		<!DOCTYPE html>
		<html>
		<head><title>About</title></head>
		<body>
			<h1>About This Server</h1>
			<p>This is a simple HTTP server built with Go's standard library.</p>
			<p>No frameworks needed — just <code>net/http</code>!</p>
			<ul>
				<li><a href="/">Home</a></li>
				<li><a href="/hello">Hello</a></li>
				<li><a href="/time">Current Time</a></li>
				<li><a href="/greet?name=World">Greet</a></li>
			</ul>
		</body>
		</html>`
		fmt.Fprint(w, html)
	})

	// ========================================
	// 4. Start the server with http.ListenAndServe
	// ========================================
	// ListenAndServe starts an HTTP server on the given address.
	// It blocks forever (or until an error occurs).
	// The second argument (nil) means "use the default ServeMux" —
	// that's where our HandleFunc registrations went.

	addr := ":8080"
	fmt.Printf("\nServer starting on http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop the server.")
	fmt.Println()
	fmt.Println("Try these endpoints:")
	fmt.Println("  http://localhost:8080/")
	fmt.Println("  http://localhost:8080/hello")
	fmt.Println("  http://localhost:8080/time")
	fmt.Println("  http://localhost:8080/headers")
	fmt.Println("  http://localhost:8080/greet?name=Sri")
	fmt.Println("  http://localhost:8080/health")
	fmt.Println("  http://localhost:8080/about")

	// log.Fatal will print the error and exit if ListenAndServe fails
	// (e.g., if the port is already in use)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// ========================================
// Handler Functions
// ========================================

// homeHandler handles the root path "/"
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Log the request to the server console
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)

	// fmt.Fprintf writes to any io.Writer — including http.ResponseWriter!
	fmt.Fprintf(w, "Welcome to the Go HTTP Server!\n")
	fmt.Fprintf(w, "Method: %s\n", r.Method)
	fmt.Fprintf(w, "Path: %s\n", r.URL.Path)
	fmt.Fprintf(w, "Remote Address: %s\n", r.RemoteAddr)
}

// helloHandler demonstrates a simple text response
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)

	// w.Write takes a byte slice — another way to write responses
	w.Write([]byte("Hello from Go! This is your HTTP server speaking.\n"))
}

// timeHandler returns the current time
func timeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)

	now := time.Now()
	fmt.Fprintf(w, "Current Time: %s\n", now.Format(time.RFC1123))
	fmt.Fprintf(w, "Unix Timestamp: %d\n", now.Unix())
	fmt.Fprintf(w, "Date: %s\n", now.Format("2006-01-02"))
}

// headersHandler shows all request headers — useful for debugging
func headersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)

	fmt.Fprintf(w, "Request Headers:\n")
	fmt.Fprintf(w, "================\n")
	for name, values := range r.Header {
		for _, value := range values {
			fmt.Fprintf(w, "%s: %s\n", name, value)
		}
	}
}

// greetHandler demonstrates reading query parameters
func greetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[%s] %s %s\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path)

	// ========================================
	// Reading query parameters from the URL
	// ========================================
	// For a URL like /greet?name=Sri&lang=en
	// r.URL.Query() returns a map of all query parameters

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "World" // default value if no name is provided
	}

	lang := r.URL.Query().Get("lang")

	switch lang {
	case "es":
		fmt.Fprintf(w, "Hola, %s!\n", name)
	case "fr":
		fmt.Fprintf(w, "Bonjour, %s!\n", name)
	case "ta":
		fmt.Fprintf(w, "Vanakkam, %s!\n", name)
	default:
		fmt.Fprintf(w, "Hello, %s!\n", name)
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// 1. http.HandleFunc(pattern, handlerFunc)
//    - Registers a function to handle requests matching the pattern
//    - The handler function signature: func(http.ResponseWriter, *http.Request)
//
// 2. http.ResponseWriter (w)
//    - Write response body: fmt.Fprintf(w, ...) or w.Write([]byte(...))
//    - Set headers: w.Header().Set("Key", "Value")
//    - Set status code: w.WriteHeader(http.StatusNotFound)
//
// 3. *http.Request (r)
//    - r.Method: GET, POST, PUT, DELETE, etc.
//    - r.URL.Path: the request path
//    - r.URL.Query(): query parameters
//    - r.Header: request headers
//    - r.RemoteAddr: client's address
//
// 4. http.ListenAndServe(addr, handler)
//    - Starts the server on the given address
//    - Pass nil for handler to use the DefaultServeMux
//    - Blocks until the server stops
