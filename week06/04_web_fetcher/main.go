package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ========================================
// Week 6, Lesson 4 (Mini-Project): Concurrent Web Fetcher
// ========================================
// This project brings together goroutines, channels, WaitGroups,
// and timeout handling to fetch multiple URLs concurrently.
//
// Features:
// - Concurrent HTTP GET requests using goroutines
// - Result collection via channels
// - Synchronization with sync.WaitGroup
// - Timeout handling with http.Client
// - Response time measurement
// - Summary statistics
// ========================================

// FetchResult holds the outcome of fetching a single URL.
type FetchResult struct {
	URL        string
	StatusCode int
	BytesRead  int
	Duration   time.Duration
	Error      error
	Preview    string // First N characters of the response body
}

func main() {
	fmt.Println("========================================")
	fmt.Println("Concurrent Web Fetcher")
	fmt.Println("========================================")

	// URLs to fetch concurrently.
	// These are public, reliable endpoints good for testing.
	urls := []string{
		"https://httpbin.org/get",
		"https://httpbin.org/ip",
		"https://jsonplaceholder.typicode.com/posts/1",
		"https://jsonplaceholder.typicode.com/users/1",
		"https://httpbin.org/headers",
		"https://httpbin.org/user-agent",
		"https://jsonplaceholder.typicode.com/todos/1",
		"https://httpbin.org/delay/1", // 1-second delay
	}

	fmt.Printf("\nFetching %d URLs concurrently...\n\n", len(urls))

	// ========================================
	// Fetch all URLs concurrently
	// ========================================
	startTime := time.Now()
	results := fetchAll(urls, 5*time.Second)
	totalDuration := time.Since(startTime)

	// ========================================
	// Display Results
	// ========================================
	fmt.Println("========================================")
	fmt.Println("Results")
	fmt.Println("========================================")

	var successCount, failCount int
	var totalBytes int
	var fastestDuration time.Duration
	var slowestDuration time.Duration
	fastestURL := ""
	slowestURL := ""

	for i, result := range results {
		fmt.Printf("\n--- URL %d ---\n", i+1)
		fmt.Printf("  URL:      %s\n", result.URL)

		if result.Error != nil {
			fmt.Printf("  Status:   ERROR\n")
			fmt.Printf("  Error:    %s\n", result.Error)
			fmt.Printf("  Duration: %s\n", result.Duration.Round(time.Millisecond))
			failCount++
			continue
		}

		fmt.Printf("  Status:   %d\n", result.StatusCode)
		fmt.Printf("  Bytes:    %d\n", result.BytesRead)
		fmt.Printf("  Duration: %s\n", result.Duration.Round(time.Millisecond))
		fmt.Printf("  Preview:  %s\n", result.Preview)

		successCount++
		totalBytes += result.BytesRead

		// Track fastest and slowest
		if fastestURL == "" || result.Duration < fastestDuration {
			fastestDuration = result.Duration
			fastestURL = result.URL
		}
		if result.Duration > slowestDuration {
			slowestDuration = result.Duration
			slowestURL = result.URL
		}
	}

	// ========================================
	// Summary Statistics
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary Statistics")
	fmt.Println("========================================")
	fmt.Printf("  Total URLs:      %d\n", len(urls))
	fmt.Printf("  Successful:      %d\n", successCount)
	fmt.Printf("  Failed:          %d\n", failCount)
	fmt.Printf("  Total bytes:     %d\n", totalBytes)
	fmt.Printf("  Total time:      %s\n", totalDuration.Round(time.Millisecond))

	if fastestURL != "" {
		fmt.Printf("  Fastest:         %s (%s)\n",
			truncateURL(fastestURL, 40), fastestDuration.Round(time.Millisecond))
	}
	if slowestURL != "" {
		fmt.Printf("  Slowest:         %s (%s)\n",
			truncateURL(slowestURL, 40), slowestDuration.Round(time.Millisecond))
	}

	// The key insight: total time should be MUCH less than the sum
	// of individual times, because the requests run concurrently!
	var sumDurations time.Duration
	for _, r := range results {
		sumDurations += r.Duration
	}
	fmt.Printf("\n  Sum of individual times: %s\n", sumDurations.Round(time.Millisecond))
	fmt.Printf("  Actual wall-clock time: %s\n", totalDuration.Round(time.Millisecond))
	fmt.Printf("  Speedup from concurrency: %.1fx\n",
		float64(sumDurations)/float64(totalDuration))

	fmt.Println("\n========================================")
	fmt.Println("Done!")
	fmt.Println("========================================")
}

// ========================================
// fetchAll fetches all URLs concurrently and returns results.
// ========================================
func fetchAll(urls []string, timeout time.Duration) []FetchResult {
	// Channel to collect results from goroutines
	resultsCh := make(chan FetchResult, len(urls))

	// WaitGroup to track when all goroutines finish
	var wg sync.WaitGroup

	// Create a shared HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Launch a goroutine for each URL
	for _, url := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			result := fetchURL(client, u)
			resultsCh <- result
		}(url)
	}

	// Close the results channel after all goroutines finish.
	// This runs in a separate goroutine so we don't block.
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect all results from the channel
	var results []FetchResult
	for result := range resultsCh {
		results = append(results, result)
	}

	return results
}

// ========================================
// fetchURL performs a single HTTP GET request.
// ========================================
func fetchURL(client *http.Client, url string) FetchResult {
	start := time.Now()

	// Make the HTTP GET request
	resp, err := client.Get(url)
	if err != nil {
		return FetchResult{
			URL:      url,
			Duration: time.Since(start),
			Error:    err,
		}
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FetchResult{
			URL:        url,
			StatusCode: resp.StatusCode,
			Duration:   time.Since(start),
			Error:      fmt.Errorf("reading body: %w", err),
		}
	}

	duration := time.Since(start)

	return FetchResult{
		URL:        url,
		StatusCode: resp.StatusCode,
		BytesRead:  len(body),
		Duration:   duration,
		Error:      nil,
		Preview:    makePreview(string(body), 80),
	}
}

// ========================================
// Utility Functions
// ========================================

// makePreview returns the first maxLen characters of s, cleaned up
// for display (no newlines, trimmed whitespace).
func makePreview(s string, maxLen int) string {
	// Replace newlines and extra spaces for a clean single-line preview
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.Join(strings.Fields(s), " ") // Collapse whitespace
	s = strings.TrimSpace(s)

	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// truncateURL shortens a URL for display.
func truncateURL(url string, maxLen int) string {
	if len(url) > maxLen {
		return url[:maxLen] + "..."
	}
	return url
}
