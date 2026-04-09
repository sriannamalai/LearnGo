package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ========================================
// Week 12, Lesson 4 (Mini-Project): Process Monitor
// ========================================
// A process monitor that lists running processes, watches for a
// specific process, and reports CPU/memory usage. Runs in a loop
// with configurable interval. Demonstrates exec.Command, signal
// handling, and graceful shutdown.
//
// Usage:
//   go run main.go                    # List all processes
//   go run main.go watch <name>       # Watch for a specific process
//   go run main.go top                # Show top processes by CPU
//   go run main.go top --interval 3   # Custom polling interval (seconds)
//
// Press Ctrl+C to stop monitoring.
// ========================================

// ProcessInfo holds parsed information about a running process.
type ProcessInfo struct {
	PID     int
	User    string
	CPU     float64
	Memory  float64
	VSZ     int64    // Virtual memory size in KB
	RSS     int64    // Resident set size in KB
	Command string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		fmt.Println("\nRunning default: listing all processes...")
		fmt.Println()
		listProcesses()
		return
	}

	switch os.Args[1] {
	case "list":
		listProcesses()
	case "watch":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go watch <process_name>")
			os.Exit(1)
		}
		interval := parseInterval(os.Args[3:])
		watchProcess(os.Args[2], interval)
	case "top":
		interval := parseInterval(os.Args[2:])
		topProcesses(interval)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// ========================================
// Commands
// ========================================

// printUsage prints the help message.
func printUsage() {
	fmt.Println("========================================")
	fmt.Println("Process Monitor - Week 12 Mini-Project")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run main.go list                  List all processes")
	fmt.Println("  go run main.go watch <name>          Watch for a specific process")
	fmt.Println("  go run main.go top                   Show top processes by CPU")
	fmt.Println("  go run main.go top --interval 3      Custom poll interval (seconds)")
	fmt.Println("  go run main.go help                  Show this help message")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop monitoring.")
}

// parseInterval extracts the --interval flag value from args.
func parseInterval(args []string) time.Duration {
	for i, arg := range args {
		if arg == "--interval" && i+1 < len(args) {
			sec, err := strconv.Atoi(args[i+1])
			if err == nil && sec > 0 {
				return time.Duration(sec) * time.Second
			}
		}
	}
	return 2 * time.Second // default interval
}

// listProcesses displays all running processes.
func listProcesses() {
	fmt.Println("========================================")
	fmt.Println("All Running Processes")
	fmt.Println("========================================")

	processes, err := getProcessList()
	if err != nil {
		fmt.Printf("Error getting process list: %v\n", err)
		os.Exit(1)
	}

	// Print header
	fmt.Printf("\n%-8s %-12s %6s %6s %10s %10s  %s\n",
		"PID", "USER", "%CPU", "%MEM", "VSZ(KB)", "RSS(KB)", "COMMAND")
	fmt.Println(strings.Repeat("-", 80))

	for _, p := range processes {
		cmd := p.Command
		if len(cmd) > 40 {
			cmd = cmd[:40] + "..."
		}
		fmt.Printf("%-8d %-12s %6.1f %6.1f %10d %10d  %s\n",
			p.PID, truncate(p.User, 12), p.CPU, p.Memory, p.VSZ, p.RSS, cmd)
	}

	fmt.Printf("\nTotal processes: %d\n", len(processes))
}

// watchProcess monitors for a specific process by name.
func watchProcess(name string, interval time.Duration) {
	fmt.Println("========================================")
	fmt.Printf("Watching for process: %q\n", name)
	fmt.Printf("Poll interval: %v\n", interval)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("========================================")

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately, then on interval
	checkAndReport(name)

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n\nStopping process watch. Goodbye!")
			return
		case <-ticker.C:
			checkAndReport(name)
		}
	}
}

// checkAndReport looks for a process by name and reports its status.
func checkAndReport(name string) {
	processes, err := getProcessList()
	if err != nil {
		fmt.Printf("[%s] Error: %v\n", timestamp(), err)
		return
	}

	found := []ProcessInfo{}
	for _, p := range processes {
		if strings.Contains(strings.ToLower(p.Command), strings.ToLower(name)) {
			found = append(found, p)
		}
	}

	fmt.Printf("\n[%s] Search for %q:\n", timestamp(), name)
	if len(found) == 0 {
		fmt.Printf("  Process %q is NOT running.\n", name)
	} else {
		fmt.Printf("  Found %d matching process(es):\n", len(found))
		for _, p := range found {
			fmt.Printf("    PID %-8d | CPU: %5.1f%% | MEM: %5.1f%% | RSS: %s | %s\n",
				p.PID, p.CPU, p.Memory, humanSize(p.RSS*1024), truncate(p.Command, 50))
		}

		// Aggregate stats
		var totalCPU, totalMem float64
		var totalRSS int64
		for _, p := range found {
			totalCPU += p.CPU
			totalMem += p.Memory
			totalRSS += p.RSS
		}
		fmt.Printf("  Totals: CPU: %.1f%% | MEM: %.1f%% | RSS: %s\n",
			totalCPU, totalMem, humanSize(totalRSS*1024))
	}
}

// topProcesses shows the top processes by CPU usage in a loop.
func topProcesses(interval time.Duration) {
	fmt.Println("========================================")
	fmt.Println("Top Processes by CPU Usage")
	fmt.Printf("Poll interval: %v\n", interval)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("========================================")

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately, then on interval
	showTop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n\nStopping top view. Goodbye!")
			return
		case <-ticker.C:
			showTop()
		}
	}
}

// showTop displays the top 15 processes by CPU usage.
func showTop() {
	processes, err := getProcessList()
	if err != nil {
		fmt.Printf("[%s] Error: %v\n", timestamp(), err)
		return
	}

	// Sort by CPU usage (descending)
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPU > processes[j].CPU
	})

	// Show top 15
	count := 15
	if len(processes) < count {
		count = len(processes)
	}

	// Clear-ish output (print separator)
	fmt.Printf("\n[%s] Top %d processes by CPU usage (of %d total):\n",
		timestamp(), count, len(processes))
	fmt.Printf("%-8s %-12s %6s %6s %10s  %s\n",
		"PID", "USER", "%CPU", "%MEM", "RSS", "COMMAND")
	fmt.Println(strings.Repeat("-", 70))

	for i := 0; i < count; i++ {
		p := processes[i]
		cmd := p.Command
		if len(cmd) > 35 {
			cmd = cmd[:35] + "..."
		}
		fmt.Printf("%-8d %-12s %6.1f %6.1f %10s  %s\n",
			p.PID, truncate(p.User, 12), p.CPU, p.Memory,
			humanSize(p.RSS*1024), cmd)
	}

	// Summary statistics
	var totalCPU, totalMem float64
	for _, p := range processes {
		totalCPU += p.CPU
		totalMem += p.Memory
	}
	fmt.Printf("\nSystem totals: CPU: %.1f%% | MEM: %.1f%% | Processes: %d\n",
		totalCPU, totalMem, len(processes))
}

// ========================================
// Process List Parsing
// ========================================

// getProcessList runs 'ps' and parses the output into ProcessInfo structs.
func getProcessList() ([]ProcessInfo, error) {
	var cmd *exec.Cmd

	// Use platform-appropriate ps command
	switch runtime.GOOS {
	case "darwin":
		// macOS ps syntax
		cmd = exec.Command("ps", "-axo", "pid,user,%cpu,%mem,vsz,rss,command")
	case "linux":
		// Linux ps syntax
		cmd = exec.Command("ps", "-eo", "pid,user,%cpu,%mem,vsz,rss,command")
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ps: %w", err)
	}

	return parseProcessOutput(string(output))
}

// parseProcessOutput parses ps output into ProcessInfo slices.
func parseProcessOutput(output string) ([]ProcessInfo, error) {
	var processes []ProcessInfo

	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip header line
	if scanner.Scan() {
		// Header line consumed
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		p, err := parseProcessLine(line)
		if err != nil {
			continue // Skip lines that don't parse
		}
		processes = append(processes, p)
	}

	return processes, scanner.Err()
}

// parseProcessLine parses a single line of ps output.
func parseProcessLine(line string) (ProcessInfo, error) {
	// Fields: PID USER %CPU %MEM VSZ RSS COMMAND
	// The command field may contain spaces, so we split carefully.
	fields := strings.Fields(line)
	if len(fields) < 7 {
		return ProcessInfo{}, fmt.Errorf("too few fields: %q", line)
	}

	pid, err := strconv.Atoi(fields[0])
	if err != nil {
		return ProcessInfo{}, fmt.Errorf("invalid PID: %q", fields[0])
	}

	cpu, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		cpu = 0
	}

	mem, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		mem = 0
	}

	vsz, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		vsz = 0
	}

	rss, err := strconv.ParseInt(fields[5], 10, 64)
	if err != nil {
		rss = 0
	}

	// Command is everything from field 6 onwards (may contain spaces)
	command := strings.Join(fields[6:], " ")

	return ProcessInfo{
		PID:     pid,
		User:    fields[1],
		CPU:     cpu,
		Memory:  mem,
		VSZ:     vsz,
		RSS:     rss,
		Command: command,
	}, nil
}

// ========================================
// Utility Functions
// ========================================

// timestamp returns the current time as a formatted string.
func timestamp() string {
	return time.Now().Format("15:04:05")
}

// humanSize converts bytes to a human-readable string.
func humanSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// truncate shortens a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
