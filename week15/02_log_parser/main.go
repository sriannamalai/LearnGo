package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ========================================
// Week 15, Lesson 2: Log Parser
// ========================================
// A log file parser that reads structured log formats, filters
// by severity/date/pattern, aggregates statistics, and displays
// a summary. Supports Apache/nginx combined log format and
// JSON structured logs.
//
// Usage:
//   go run main.go                       # Parse demo logs
//   go run main.go /path/to/access.log   # Parse a real log file
//   go run main.go --demo                # Generate and parse demo data
//
// The parser auto-detects the log format (Apache, JSON, or plain text).
// ========================================

// ========================================
// Types
// ========================================

// LogEntry represents a parsed log line.
type LogEntry struct {
	Timestamp time.Time
	Level     string // INFO, WARN, ERROR, DEBUG, etc.
	Message   string
	Source    string // IP address or source identifier
	Method   string // HTTP method (for access logs)
	Path     string // URL path (for access logs)
	Status   int    // HTTP status code
	Size     int64  // Response size in bytes
	RawLine  string
	LineNum  int
	Format   string // "apache", "json", "plain"
}

// LogStats holds aggregated statistics.
type LogStats struct {
	TotalLines    int
	ParsedLines   int
	FailedLines   int
	LevelCounts   map[string]int
	StatusCounts  map[int]int
	SourceCounts  map[string]int
	PathCounts    map[string]int
	MethodCounts  map[string]int
	HourlyCounts  map[int]int
	ErrorMessages []LogEntry
	EarliestTime  time.Time
	LatestTime    time.Time
}

func main() {
	fmt.Println("========================================")
	fmt.Println("Log Parser")
	fmt.Println("========================================")

	if len(os.Args) > 1 && os.Args[1] != "--demo" {
		// Parse a real log file
		parseFile(os.Args[1])
		return
	}

	// Generate and parse demo data
	runDemo()
}

// ========================================
// Demo Mode
// ========================================

func runDemo() {
	// Create temp workspace
	workspace, err := os.MkdirTemp("", "logparser-*")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer os.RemoveAll(workspace)

	// ========================================
	// 1. Apache/Nginx Combined Log Format
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("1. Apache/Nginx Combined Log Format")
	fmt.Println("========================================")

	apacheLog := filepath.Join(workspace, "access.log")
	generateApacheLog(apacheLog)
	fmt.Printf("\nGenerated: %s\n", apacheLog)
	parseFile(apacheLog)

	// ========================================
	// 2. JSON Structured Logs
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. JSON Structured Logs")
	fmt.Println("========================================")

	jsonLog := filepath.Join(workspace, "app.json.log")
	generateJSONLog(jsonLog)
	fmt.Printf("\nGenerated: %s\n", jsonLog)
	parseFile(jsonLog)

	// ========================================
	// 3. Plain Text Application Logs
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Plain Text Application Logs")
	fmt.Println("========================================")

	plainLog := filepath.Join(workspace, "app.log")
	generatePlainLog(plainLog)
	fmt.Printf("\nGenerated: %s\n", plainLog)
	parseFile(plainLog)

	// ========================================
	// 4. Filtering Examples
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Filtering Examples")
	fmt.Println("========================================")

	entries := parseFileToEntries(plainLog)
	fmt.Printf("\nTotal entries: %d\n", len(entries))

	// Filter by level
	errors := filterByLevel(entries, "ERROR")
	fmt.Printf("\nERROR entries (%d):\n", len(errors))
	for _, e := range errors {
		fmt.Printf("  [%s] %s\n", e.Timestamp.Format("15:04:05"), e.Message)
	}

	warnings := filterByLevel(entries, "WARN")
	fmt.Printf("\nWARN entries (%d):\n", len(warnings))
	for _, e := range warnings {
		fmt.Printf("  [%s] %s\n", e.Timestamp.Format("15:04:05"), e.Message)
	}

	// Filter by pattern
	pattern := "database"
	matched := filterByPattern(entries, pattern)
	fmt.Printf("\nEntries matching %q (%d):\n", pattern, len(matched))
	for _, e := range matched {
		fmt.Printf("  [%s] [%s] %s\n", e.Timestamp.Format("15:04:05"), e.Level, e.Message)
	}

	fmt.Println("\n========================================")
	fmt.Println("Log Parser lesson complete!")
	fmt.Println("========================================")
}

// ========================================
// Parsing
// ========================================

// parseFile reads and analyzes a log file.
func parseFile(path string) {
	entries := parseFileToEntries(path)
	if len(entries) == 0 {
		fmt.Println("  No entries parsed.")
		return
	}

	stats := computeStats(entries)
	printStats(stats, path)
}

// parseFileToEntries reads a log file and returns parsed entries.
func parseFileToEntries(path string) []LogEntry {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer f.Close()

	var entries []LogEntry
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry := parseLine(line, lineNum)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	return entries
}

// parseLine attempts to parse a log line in various formats.
func parseLine(line string, lineNum int) *LogEntry {
	// Try JSON format first
	if strings.HasPrefix(strings.TrimSpace(line), "{") {
		if entry := parseJSONLine(line, lineNum); entry != nil {
			return entry
		}
	}

	// Try Apache combined log format
	if entry := parseApacheLine(line, lineNum); entry != nil {
		return entry
	}

	// Try plain text log format
	if entry := parsePlainLine(line, lineNum); entry != nil {
		return entry
	}

	return nil
}

// Apache combined log format regex:
// 192.168.1.1 - - [10/Oct/2024:13:55:36 -0700] "GET /api/users HTTP/1.1" 200 2326
var apacheRegex = regexp.MustCompile(
	`^(\S+)\s+\S+\s+\S+\s+\[([^\]]+)\]\s+"(\S+)\s+(\S+)\s+\S+"\s+(\d+)\s+(\d+)`)

func parseApacheLine(line string, lineNum int) *LogEntry {
	matches := apacheRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	// Parse timestamp: 10/Oct/2024:13:55:36 -0700
	t, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[2])
	if err != nil {
		return nil
	}

	status := 0
	fmt.Sscanf(matches[5], "%d", &status)
	size := int64(0)
	fmt.Sscanf(matches[6], "%d", &size)

	level := "INFO"
	if status >= 500 {
		level = "ERROR"
	} else if status >= 400 {
		level = "WARN"
	}

	return &LogEntry{
		Timestamp: t,
		Level:     level,
		Source:    matches[1],
		Method:   matches[3],
		Path:     matches[4],
		Status:   status,
		Size:     size,
		RawLine:  line,
		LineNum:  lineNum,
		Format:   "apache",
		Message:  fmt.Sprintf("%s %s -> %d", matches[3], matches[4], status),
	}
}

// JSONLogEntry represents a JSON-formatted log line.
type JSONLogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Service   string `json:"service"`
	TraceID   string `json:"trace_id"`
}

func parseJSONLine(line string, lineNum int) *LogEntry {
	var jsonEntry JSONLogEntry
	if err := json.Unmarshal([]byte(line), &jsonEntry); err != nil {
		return nil
	}

	if jsonEntry.Level == "" || jsonEntry.Message == "" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, jsonEntry.Timestamp)
	if err != nil {
		t = time.Now()
	}

	return &LogEntry{
		Timestamp: t,
		Level:     strings.ToUpper(jsonEntry.Level),
		Message:   jsonEntry.Message,
		Source:    jsonEntry.Service,
		RawLine:   line,
		LineNum:   lineNum,
		Format:    "json",
	}
}

// Plain text log format: 2024-01-15 10:30:45 [INFO] Application started
var plainRegex = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+(.+)$`)

func parsePlainLine(line string, lineNum int) *LogEntry {
	matches := plainRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	t, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		return nil
	}

	return &LogEntry{
		Timestamp: t,
		Level:     strings.ToUpper(matches[2]),
		Message:   matches[3],
		RawLine:   line,
		LineNum:   lineNum,
		Format:    "plain",
	}
}

// ========================================
// Statistics
// ========================================

// computeStats aggregates statistics from log entries.
func computeStats(entries []LogEntry) LogStats {
	stats := LogStats{
		TotalLines:   len(entries),
		ParsedLines:  len(entries),
		LevelCounts:  make(map[string]int),
		StatusCounts: make(map[int]int),
		SourceCounts: make(map[string]int),
		PathCounts:   make(map[string]int),
		MethodCounts: make(map[string]int),
		HourlyCounts: make(map[int]int),
	}

	for _, e := range entries {
		stats.LevelCounts[e.Level]++
		stats.HourlyCounts[e.Timestamp.Hour()]++

		if e.Status > 0 {
			stats.StatusCounts[e.Status]++
		}
		if e.Source != "" {
			stats.SourceCounts[e.Source]++
		}
		if e.Path != "" {
			stats.PathCounts[e.Path]++
		}
		if e.Method != "" {
			stats.MethodCounts[e.Method]++
		}

		if e.Level == "ERROR" {
			stats.ErrorMessages = append(stats.ErrorMessages, e)
		}

		if stats.EarliestTime.IsZero() || e.Timestamp.Before(stats.EarliestTime) {
			stats.EarliestTime = e.Timestamp
		}
		if e.Timestamp.After(stats.LatestTime) {
			stats.LatestTime = e.Timestamp
		}
	}

	return stats
}

// printStats displays log statistics.
func printStats(stats LogStats, path string) {
	fmt.Printf("\n--- Log Analysis: %s ---\n", filepath.Base(path))

	// Time range
	if !stats.EarliestTime.IsZero() {
		fmt.Printf("  Time range: %s to %s\n",
			stats.EarliestTime.Format("2006-01-02 15:04:05"),
			stats.LatestTime.Format("2006-01-02 15:04:05"))
		duration := stats.LatestTime.Sub(stats.EarliestTime)
		fmt.Printf("  Duration:   %v\n", duration.Round(time.Second))
	}

	fmt.Printf("  Total entries: %d\n", stats.ParsedLines)

	// Level distribution
	fmt.Println("\n  Log Levels:")
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	for _, level := range levels {
		count := stats.LevelCounts[level]
		if count > 0 {
			bar := strings.Repeat("#", min(count, 40))
			pct := float64(count) / float64(stats.ParsedLines) * 100
			fmt.Printf("    %-6s %4d (%5.1f%%) %s\n", level, count, pct, bar)
		}
	}

	// HTTP status codes (if present)
	if len(stats.StatusCounts) > 0 {
		fmt.Println("\n  HTTP Status Codes:")
		statusCodes := sortedKeys(stats.StatusCounts)
		for _, code := range statusCodes {
			count := stats.StatusCounts[code]
			bar := strings.Repeat("#", min(count, 30))
			fmt.Printf("    %d  %4d  %s\n", code, count, bar)
		}
	}

	// Top sources (if present)
	if len(stats.SourceCounts) > 0 {
		fmt.Println("\n  Top Sources (top 5):")
		topSources := topN(stats.SourceCounts, 5)
		for _, item := range topSources {
			fmt.Printf("    %-20s %4d requests\n", item.Key, item.Count)
		}
	}

	// Top paths (if present)
	if len(stats.PathCounts) > 0 {
		fmt.Println("\n  Top Paths (top 5):")
		topPaths := topN(stats.PathCounts, 5)
		for _, item := range topPaths {
			fmt.Printf("    %-30s %4d hits\n", item.Key, item.Count)
		}
	}

	// HTTP methods (if present)
	if len(stats.MethodCounts) > 0 {
		fmt.Println("\n  HTTP Methods:")
		for method, count := range stats.MethodCounts {
			fmt.Printf("    %-6s %4d\n", method, count)
		}
	}

	// Recent errors
	if len(stats.ErrorMessages) > 0 {
		fmt.Printf("\n  Recent Errors (last %d):\n", min(len(stats.ErrorMessages), 5))
		start := len(stats.ErrorMessages) - 5
		if start < 0 {
			start = 0
		}
		for _, e := range stats.ErrorMessages[start:] {
			msg := e.Message
			if len(msg) > 60 {
				msg = msg[:60] + "..."
			}
			fmt.Printf("    [%s] %s\n", e.Timestamp.Format("15:04:05"), msg)
		}
	}

	// Hourly distribution
	if len(stats.HourlyCounts) > 1 {
		fmt.Println("\n  Hourly Distribution:")
		maxCount := 0
		for _, count := range stats.HourlyCounts {
			if count > maxCount {
				maxCount = count
			}
		}
		for hour := 0; hour < 24; hour++ {
			count := stats.HourlyCounts[hour]
			if count > 0 {
				barLen := count * 30 / maxCount
				if barLen < 1 {
					barLen = 1
				}
				bar := strings.Repeat("#", barLen)
				fmt.Printf("    %02d:00  %4d  %s\n", hour, count, bar)
			}
		}
	}
}

// ========================================
// Filtering
// ========================================

// filterByLevel returns entries matching the given level.
func filterByLevel(entries []LogEntry, level string) []LogEntry {
	var result []LogEntry
	for _, e := range entries {
		if e.Level == level {
			result = append(result, e)
		}
	}
	return result
}

// filterByPattern returns entries whose message matches a pattern.
func filterByPattern(entries []LogEntry, pattern string) []LogEntry {
	var result []LogEntry
	lowerPattern := strings.ToLower(pattern)
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Message), lowerPattern) {
			result = append(result, e)
		}
	}
	return result
}

// ========================================
// Demo Log Generators
// ========================================

func generateApacheLog(path string) {
	f, _ := os.Create(path)
	defer f.Close()

	ips := []string{"192.168.1.1", "10.0.0.5", "172.16.0.100", "192.168.1.50", "10.0.0.22"}
	paths := []string{"/", "/api/users", "/api/products", "/login", "/static/style.css",
		"/api/orders", "/health", "/favicon.ico", "/api/search?q=test"}
	methods := []string{"GET", "GET", "GET", "POST", "GET", "GET", "GET", "GET", "GET"}
	statuses := []int{200, 200, 200, 201, 200, 200, 200, 404, 500, 200, 304, 200, 403}

	baseTime := time.Date(2024, 10, 15, 8, 0, 0, 0, time.FixedZone("PDT", -7*3600))
	for i := 0; i < 100; i++ {
		t := baseTime.Add(time.Duration(i) * 3 * time.Minute)
		ip := ips[i%len(ips)]
		p := paths[i%len(paths)]
		m := methods[i%len(methods)]
		s := statuses[i%len(statuses)]
		size := 500 + (i * 47 % 5000)

		fmt.Fprintf(f, "%s - - [%s] \"%s %s HTTP/1.1\" %d %d\n",
			ip, t.Format("02/Jan/2006:15:04:05 -0700"), m, p, s, size)
	}
}

func generateJSONLog(path string) {
	f, _ := os.Create(path)
	defer f.Close()

	levels := []string{"info", "info", "info", "warn", "error", "debug", "info"}
	messages := []string{
		"Request processed successfully",
		"User authenticated",
		"Cache hit for key users:123",
		"Slow query detected (>100ms)",
		"Database connection timeout",
		"Processing request for /api/data",
		"Response sent in 45ms",
		"Failed to parse JSON body",
		"Rate limit exceeded for client",
		"Health check passed",
	}
	services := []string{"api-server", "auth-service", "cache", "db-proxy", "api-server"}

	baseTime := time.Date(2024, 10, 15, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 50; i++ {
		t := baseTime.Add(time.Duration(i) * 2 * time.Minute)
		entry := JSONLogEntry{
			Timestamp: t.Format(time.RFC3339),
			Level:     levels[i%len(levels)],
			Message:   messages[i%len(messages)],
			Service:   services[i%len(services)],
			TraceID:   fmt.Sprintf("trace-%04d", i),
		}
		data, _ := json.Marshal(entry)
		fmt.Fprintln(f, string(data))
	}
}

func generatePlainLog(path string) {
	f, _ := os.Create(path)
	defer f.Close()

	entries := []struct {
		level   string
		message string
	}{
		{"INFO", "Application started on port 8080"},
		{"INFO", "Loading configuration from /etc/app/config.yaml"},
		{"DEBUG", "Database connection pool initialized (max=20)"},
		{"INFO", "Connected to database at localhost:5432"},
		{"INFO", "Cache service connected at redis:6379"},
		{"WARN", "Configuration value 'timeout' not set, using default 30s"},
		{"INFO", "HTTP server listening on :8080"},
		{"INFO", "Processing incoming request from 192.168.1.5"},
		{"DEBUG", "Query executed in 12ms: SELECT * FROM users"},
		{"INFO", "Request completed: 200 OK (45ms)"},
		{"WARN", "Slow database query detected: 150ms for users table"},
		{"INFO", "User 'alice' logged in successfully"},
		{"ERROR", "Failed to connect to external API: connection refused"},
		{"INFO", "Retry attempt 1/3 for external API"},
		{"INFO", "External API connection restored"},
		{"ERROR", "Database connection lost: timeout after 5s"},
		{"WARN", "Reconnecting to database..."},
		{"INFO", "Database connection re-established"},
		{"INFO", "Processing batch job: 500 records"},
		{"INFO", "Batch job completed: 498 success, 2 failures"},
		{"ERROR", "Disk space warning: /var/log is 92% full"},
		{"WARN", "Memory usage at 85%: consider scaling"},
		{"INFO", "Graceful shutdown initiated"},
		{"INFO", "Waiting for 3 active connections to close"},
		{"INFO", "All connections closed. Server stopped."},
	}

	baseTime := time.Date(2024, 10, 15, 9, 0, 0, 0, time.UTC)
	for i, e := range entries {
		t := baseTime.Add(time.Duration(i) * 5 * time.Minute)
		fmt.Fprintf(f, "%s [%s] %s\n", t.Format("2006-01-02 15:04:05"), e.level, e.message)
	}
}

// ========================================
// Utility Functions
// ========================================

// CountItem holds a key-count pair for sorting.
type CountItem struct {
	Key   string
	Count int
}

// topN returns the top N items by count from a string->int map.
func topN(m map[string]int, n int) []CountItem {
	items := make([]CountItem, 0, len(m))
	for k, v := range m {
		items = append(items, CountItem{k, v})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})
	if len(items) > n {
		items = items[:n]
	}
	return items
}

// sortedKeys returns sorted int keys from a map.
func sortedKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
