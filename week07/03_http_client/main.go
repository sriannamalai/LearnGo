package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ========================================
// Week 7, Lesson 3: HTTP Client
// ========================================
// Go's net/http package provides a powerful HTTP client for making
// requests. This lesson covers GET, POST, custom headers, timeouts,
// and parsing JSON responses from APIs.
//
// We use https://jsonplaceholder.typicode.com — a free, public
// REST API for testing and prototyping.
// ========================================

func main() {
	// ========================================
	// 1. Simple HTTP GET
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. Simple HTTP GET")
	fmt.Println("========================================")

	// http.Get is the simplest way to make a GET request.
	// Always close resp.Body when done!

	fmt.Println("\nFetching a post from JSONPlaceholder...")
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp.Body.Close() // ALWAYS close the body

	// Check the status code
	fmt.Printf("  Status: %s\n", resp.Status)
	fmt.Printf("  Status Code: %d\n", resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("  Error reading body: %v\n", err)
		return
	}

	fmt.Printf("  Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("  Body length: %d bytes\n", len(body))
	fmt.Printf("  Body preview: %.100s...\n", string(body))

	// ========================================
	// 2. Parsing JSON Responses
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Parsing JSON Responses")
	fmt.Println("========================================")

	// Fetch and parse a post into a struct
	fmt.Println("\nFetching and parsing a post:")
	post, err := fetchPost(1)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Post ID:     %d\n", post.ID)
		fmt.Printf("  User ID:     %d\n", post.UserID)
		fmt.Printf("  Title:       %s\n", post.Title)
		fmt.Printf("  Body (50ch): %.50s...\n", post.Body)
	}

	// Fetch and parse a user
	fmt.Println("\nFetching and parsing a user:")
	user, err := fetchUser(1)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Name:    %s\n", user.Name)
		fmt.Printf("  Email:   %s\n", user.Email)
		fmt.Printf("  Phone:   %s\n", user.Phone)
		fmt.Printf("  Website: %s\n", user.Website)
		fmt.Printf("  Company: %s\n", user.Company.Name)
		fmt.Printf("  City:    %s\n", user.Address.City)
	}

	// ========================================
	// 3. HTTP POST with JSON Body
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. HTTP POST with JSON Body")
	fmt.Println("========================================")

	// Create a new post by sending JSON in the request body.
	// JSONPlaceholder accepts POSTs but doesn't actually create
	// resources — it returns what would have been created.

	newPost := Post{
		UserID: 1,
		Title:  "Learning Go HTTP Client",
		Body:   "Go makes HTTP requests simple and efficient with net/http.",
	}

	createdPost, err := createPost(newPost)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Created post with ID: %d\n", createdPost.ID)
		fmt.Printf("  Title: %s\n", createdPost.Title)
	}

	// ========================================
	// 4. Reading Response Body
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Reading Response Body")
	fmt.Println("========================================")

	// There are two main ways to read a response body:

	// Method 1: io.ReadAll (read all at once, good for small responses)
	fmt.Println("\n  Method 1: io.ReadAll — reads entire body into []byte")
	fmt.Println("    body, err := io.ReadAll(resp.Body)")

	// Method 2: json.NewDecoder (stream directly, more efficient)
	fmt.Println("  Method 2: json.NewDecoder — streams from reader")
	fmt.Println("    json.NewDecoder(resp.Body).Decode(&result)")

	// Demonstrating json.NewDecoder (preferred for JSON responses)
	fmt.Println("\n  Using json.NewDecoder:")
	resp2, err := http.Get("https://jsonplaceholder.typicode.com/todos/1")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp2.Body.Close()

	var todo Todo
	err = json.NewDecoder(resp2.Body).Decode(&todo)
	if err != nil {
		fmt.Printf("  Error decoding: %v\n", err)
		return
	}
	fmt.Printf("  Todo: %s (completed: %v)\n", todo.Title, todo.Completed)

	// ========================================
	// 5. Setting Headers
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Setting Headers")
	fmt.Println("========================================")

	// To set custom headers, use http.NewRequest + client.Do
	// instead of http.Get.

	req, err := http.NewRequest("GET", "https://jsonplaceholder.typicode.com/posts/2", nil)
	if err != nil {
		fmt.Printf("  Error creating request: %v\n", err)
		return
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "LearnGo-HTTPClient/1.0")
	req.Header.Set("X-Custom-Header", "learning-go")

	fmt.Println("  Request headers:")
	for key, values := range req.Header {
		fmt.Printf("    %s: %s\n", key, values[0])
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp3, err := client.Do(req)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp3.Body.Close()

	fmt.Printf("  Response status: %s\n", resp3.Status)

	// Display some response headers
	fmt.Println("  Response headers (selected):")
	fmt.Printf("    Content-Type: %s\n", resp3.Header.Get("Content-Type"))
	fmt.Printf("    Cache-Control: %s\n", resp3.Header.Get("Cache-Control"))

	// ========================================
	// 6. Custom http.Client with Timeout
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Custom http.Client with Timeout")
	fmt.Println("========================================")

	// The default http.Client has NO timeout — it will wait forever!
	// Always create a custom client with a timeout in production.

	// Good practice: create a client with timeout
	customClient := &http.Client{
		Timeout: 5 * time.Second, // Total request timeout
	}

	fmt.Println("  Custom client with 5s timeout")

	start := time.Now()
	resp4, err := customClient.Get("https://jsonplaceholder.typicode.com/posts/3")
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("  Error (possibly timeout): %v\n", err)
	} else {
		defer resp4.Body.Close()
		fmt.Printf("  Status: %s (took %s)\n", resp4.Status, duration.Round(time.Millisecond))
	}

	// Demonstrate timeout with a deliberately slow endpoint
	fmt.Println("\n  Testing timeout with 1-second limit on slow endpoint:")
	shortClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	start = time.Now()
	_, err = shortClient.Get("https://httpbin.org/delay/5") // 5-second delay
	duration = time.Since(start)
	if err != nil {
		fmt.Printf("  Timed out after %s: %v\n", duration.Round(time.Millisecond), err)
	}

	// ========================================
	// 7. Fetching Multiple Items (List)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. Fetching a List of Items")
	fmt.Println("========================================")

	fmt.Println("\nFetching all todos for user 1:")
	resp5, err := customClient.Get("https://jsonplaceholder.typicode.com/users/1/todos")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	defer resp5.Body.Close()

	var todos []Todo
	err = json.NewDecoder(resp5.Body).Decode(&todos)
	if err != nil {
		fmt.Printf("  Error decoding: %v\n", err)
		return
	}

	completed := 0
	for _, t := range todos {
		if t.Completed {
			completed++
		}
	}
	fmt.Printf("  Total todos: %d\n", len(todos))
	fmt.Printf("  Completed:   %d\n", completed)
	fmt.Printf("  Pending:     %d\n", len(todos)-completed)

	// Show first 5
	fmt.Println("\n  First 5 todos:")
	for i, t := range todos {
		if i >= 5 {
			break
		}
		status := "[ ]"
		if t.Completed {
			status = "[x]"
		}
		fmt.Printf("    %s %s\n", status, t.Title)
	}

	// ========================================
	// 8. Error Handling Best Practices
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. Error Handling Best Practices")
	fmt.Println("========================================")

	fmt.Println("  1. Always check the error from http.Get/Do")
	fmt.Println("  2. Always defer resp.Body.Close()")
	fmt.Println("  3. Check resp.StatusCode — a non-nil error means")
	fmt.Println("     network failure, not HTTP errors (4xx, 5xx)")
	fmt.Println("  4. Always set a timeout on http.Client")
	fmt.Println("  5. Read the entire body even if you don't need it,")
	fmt.Println("     to allow connection reuse (or close the body)")

	// Demonstrating status code checking
	fmt.Println("\n  Checking for HTTP errors:")
	resp6, err := customClient.Get("https://jsonplaceholder.typicode.com/posts/9999")
	if err != nil {
		fmt.Printf("  Network error: %v\n", err)
		return
	}
	defer resp6.Body.Close()

	// http.Get returns a nil error even for 404!
	// You must check StatusCode yourself.
	if resp6.StatusCode != http.StatusOK {
		fmt.Printf("  HTTP error: %d %s\n", resp6.StatusCode, resp6.Status)
	} else {
		fmt.Printf("  Success: %s\n", resp6.Status)
	}

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println("- http.Get: simple GET request")
	fmt.Println("- http.Post: simple POST with body")
	fmt.Println("- http.NewRequest + client.Do: full control")
	fmt.Println("- io.ReadAll: read entire response body")
	fmt.Println("- json.NewDecoder: stream JSON from response body")
	fmt.Println("- Always defer resp.Body.Close()")
	fmt.Println("- Always use http.Client{Timeout: ...} in production")
	fmt.Println("- Check StatusCode — errors are not network errors")
}

// ========================================
// Types
// ========================================

// Post represents a JSONPlaceholder post.
type Post struct {
	UserID int    `json:"userId"`
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Body   string `json:"body"`
}

// User represents a JSONPlaceholder user.
type User struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Phone    string  `json:"phone"`
	Website  string  `json:"website"`
	Address  UAddr   `json:"address"`
	Company  Company `json:"company"`
}

type UAddr struct {
	Street  string `json:"street"`
	Suite   string `json:"suite"`
	City    string `json:"city"`
	Zipcode string `json:"zipcode"`
}

type Company struct {
	Name        string `json:"name"`
	CatchPhrase string `json:"catchPhrase"`
	Bs          string `json:"bs"`
}

// Todo represents a JSONPlaceholder todo item.
type Todo struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

// ========================================
// Helper Functions
// ========================================

// fetchPost retrieves a single post by ID.
func fetchPost(id int) (Post, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d", id)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return Post{}, fmt.Errorf("GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Post{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var post Post
	err = json.NewDecoder(resp.Body).Decode(&post)
	if err != nil {
		return Post{}, fmt.Errorf("decode failed: %w", err)
	}

	return post, nil
}

// fetchUser retrieves a single user by ID.
func fetchUser(id int) (User, error) {
	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", id)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return User{}, fmt.Errorf("GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return User{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var user User
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return User{}, fmt.Errorf("decode failed: %w", err)
	}

	return user, nil
}

// createPost sends a POST request to create a new post.
func createPost(post Post) (Post, error) {
	// Marshal the post to JSON
	jsonBody, err := json.Marshal(post)
	if err != nil {
		return Post{}, fmt.Errorf("marshal failed: %w", err)
	}

	// Make the POST request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(
		"https://jsonplaceholder.typicode.com/posts",
		"application/json",
		bytes.NewReader(jsonBody),
	)
	if err != nil {
		return Post{}, fmt.Errorf("POST failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return Post{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	// Parse the response
	var created Post
	err = json.NewDecoder(resp.Body).Decode(&created)
	if err != nil {
		return Post{}, fmt.Errorf("decode failed: %w", err)
	}

	return created, nil
}
