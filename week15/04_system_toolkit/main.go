package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ========================================
// Week 15, Lesson 4 (Mini-Project): System Toolkit
// ========================================
// A multi-tool system utility combining file watcher, log tail,
// and disk monitor into one CLI. Uses os.Args to select the mode.
// Graceful shutdown on SIGINT (Ctrl+C).
//
// Usage:
//   go run main.go watch <dir> [interval_seconds]
//       Monitor directory for file changes (polling).
//
//   go run main.go tail <file> [lines]
//       Tail a log file with follow mode (like tail -f).
//
//   go run main.go disk <dir> [interval_seconds]
//       Monitor disk usage of a directory over time.
//
//   go run main.go help
//       Show this help message.
//
// All modes support Ctrl+C for graceful shutdown.
// ========================================

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "watch":
		dir := "."
		interval := 2
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		if len(os.Args) > 3 {
			fmt.Sscanf(os.Args[3], "%d", &interval)
		}
		runWatch(dir, time.Duration(interval)*time.Second)

	case "tail":
		if len(os.Args) < 3 {
			fmt.Println("Usage: go run main.go tail <file> [lines]")
			os.Exit(1)
		}
		file := os.Args[2]
		lines := 10
		if len(os.Args) > 3 {
			fmt.Sscanf(os.Args[3], "%d", &lines)
		}
		runTail(file, lines)

	case "disk":
		dir := "."
		interval := 5
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		if len(os.Args) > 3 {
			fmt.Sscanf(os.Args[3], "%d", &interval)
		}
		runDisk(dir, time.Duration(interval)*time.Second)

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("========================================")
	fmt.Println("System Toolkit - Week 15 Mini-Project")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("A multi-tool system utility combining file watcher,")
	fmt.Println("log tail, and disk monitor into one CLI.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  watch <dir> [interval]     Monitor directory for changes")
	fmt.Println("  tail  <file> [lines]       Tail a file (like tail -f)")
	fmt.Println("  disk  <dir> [interval]     Monitor disk usage over time")
	fmt.Println("  help                       Show this help")
	fmt.Println()
	fmt.Println("All commands support Ctrl+C for graceful shutdown.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go watch /var/log 3")
	fmt.Println("  go run main.go tail /var/log/system.log 20")
	fmt.Println("  go run main.go disk ~/Projects 10")
}

// ========================================
// Tool 1: File Watcher
// ========================================

// FileSnapshot holds a point-in-time view of files in a directory.
type FileSnapshot map[string]FileState

// FileState holds metadata about a file.
type FileState struct {
	Size    int64
	ModTime time.Time
	IsDir   bool
}

func runWatch(dir string, interval time.Duration) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: %s is not a valid directory\n", absDir)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("File Watcher")
	fmt.Println("========================================")
	fmt.Printf("Directory: %s\n", absDir)
	fmt.Printf("Interval:  %v\n", interval)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("========================================")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initial scan
	snapshot := scanDir(absDir)
	fmt.Printf("[%s] Initial scan: %d items found\n\n", timestamp(), len(snapshot))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	totalCreated, totalModified, totalDeleted := 0, 0, 0

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n[%s] Watcher stopped.\n", timestamp())
			fmt.Printf("  Created: %d, Modified: %d, Deleted: %d\n",
				totalCreated, totalModified, totalDeleted)
			return
		case <-ticker.C:
			newSnapshot := scanDir(absDir)

			// Detect changes
			for path, newState := range newSnapshot {
				oldState, existed := snapshot[path]
				if !existed {
					totalCreated++
					fmt.Printf("[%s] + CREATED  %s (%s)\n",
						timestamp(), path, humanSize(newState.Size))
				} else if newState.ModTime != oldState.ModTime || newState.Size != oldState.Size {
					totalModified++
					diff := newState.Size - oldState.Size
					sign := "+"
					if diff < 0 {
						sign = ""
					}
					fmt.Printf("[%s] ~ MODIFIED %s (%s%s)\n",
						timestamp(), path, sign, humanSize(diff))
				}
			}

			for path := range snapshot {
				if _, exists := newSnapshot[path]; !exists {
					totalDeleted++
					fmt.Printf("[%s] - DELETED  %s\n", timestamp(), path)
				}
			}

			snapshot = newSnapshot
		}
	}
}

// scanDir takes a snapshot of all files in a directory.
func scanDir(dir string) FileSnapshot {
	snapshot := make(FileSnapshot)
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == dir {
			return nil
		}
		// Skip hidden
		if strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		info, err := d.Info()
		if err != nil {
			return nil
		}

		snapshot[relPath] = FileState{
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   d.IsDir(),
		}
		return nil
	})
	return snapshot
}

// ========================================
// Tool 2: Log Tail
// ========================================

func runTail(file string, initialLines int) {
	absFile, err := filepath.Abs(file)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if info.IsDir() {
		fmt.Printf("Error: %s is a directory, not a file\n", absFile)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("Log Tail (tail -f)")
	fmt.Println("========================================")
	fmt.Printf("File:     %s\n", absFile)
	fmt.Printf("Size:     %s\n", humanSize(info.Size()))
	fmt.Printf("Initial:  last %d lines\n", initialLines)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("========================================")
	fmt.Println()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Read and display the last N lines
	lastLines, err := tailFile(absFile, initialLines)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	for _, line := range lastLines {
		fmt.Println(line)
	}

	if len(lastLines) > 0 {
		fmt.Println("--- tail follow mode ---")
	}

	// Follow mode: watch for new content
	followFile(ctx, absFile)

	fmt.Printf("\n[%s] Tail stopped.\n", timestamp())
}

// tailFile reads the last N lines from a file.
func tailFile(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Read all lines (for simplicity; for very large files,
	// you'd seek from the end)
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}

	return lines, nil
}

// followFile watches a file for new content (like tail -f).
func followFile(ctx context.Context, path string) {
	// Get initial file size
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	lastSize := info.Size()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			currentSize := info.Size()
			if currentSize > lastSize {
				// File has grown — read the new content
				f, err := os.Open(path)
				if err != nil {
					continue
				}

				// Seek to where we left off
				_, err = f.Seek(lastSize, io.SeekStart)
				if err != nil {
					f.Close()
					continue
				}

				scanner := bufio.NewScanner(f)
				for scanner.Scan() {
					fmt.Println(scanner.Text())
				}
				f.Close()

				lastSize = currentSize
			} else if currentSize < lastSize {
				// File was truncated (common with log rotation)
				fmt.Printf("[%s] File truncated (log rotation?). Re-reading from start.\n",
					timestamp())
				lastSize = 0
			}
		}
	}
}

// ========================================
// Tool 3: Disk Monitor
// ========================================

// DiskSnapshot holds disk usage data at a point in time.
type DiskSnapshot struct {
	Time      time.Time
	TotalSize int64
	FileCount int
	DirCount  int
}

func runDisk(dir string, interval time.Duration) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: %s is not a valid directory\n", absDir)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("Disk Usage Monitor")
	fmt.Println("========================================")
	fmt.Printf("Directory: %s\n", absDir)
	fmt.Printf("Interval:  %v\n", interval)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("========================================")
	fmt.Println()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var history []DiskSnapshot

	// Initial scan
	snap := takeDiskSnapshot(absDir)
	history = append(history, snap)
	printDiskSnapshot(snap, nil)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n[%s] Disk monitor stopped.\n", timestamp())
			printDiskSummary(history)
			return
		case <-ticker.C:
			snap := takeDiskSnapshot(absDir)
			var prev *DiskSnapshot
			if len(history) > 0 {
				prev = &history[len(history)-1]
			}
			history = append(history, snap)
			printDiskSnapshot(snap, prev)
		}
	}
}

// takeDiskSnapshot scans a directory and records usage data.
func takeDiskSnapshot(dir string) DiskSnapshot {
	snap := DiskSnapshot{
		Time: time.Now(),
	}

	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Skip hidden
		if strings.HasPrefix(d.Name(), ".") && path != dir {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			snap.DirCount++
		} else {
			snap.FileCount++
			info, err := d.Info()
			if err == nil {
				snap.TotalSize += info.Size()
			}
		}
		return nil
	})

	return snap
}

// printDiskSnapshot displays a single disk snapshot with comparison.
func printDiskSnapshot(current DiskSnapshot, prev *DiskSnapshot) {
	ts := current.Time.Format("15:04:05")

	if prev == nil {
		fmt.Printf("[%s] Size: %10s | Files: %6d | Dirs: %5d\n",
			ts, humanSize(current.TotalSize), current.FileCount, current.DirCount)
		return
	}

	sizeDiff := current.TotalSize - prev.TotalSize
	fileDiff := current.FileCount - prev.FileCount
	dirDiff := current.DirCount - prev.DirCount

	sizeStr := humanSize(current.TotalSize)
	diffStr := ""

	if sizeDiff != 0 {
		sign := "+"
		if sizeDiff < 0 {
			sign = ""
		}
		diffStr = fmt.Sprintf(" (%s%s)", sign, humanSize(sizeDiff))
	}

	fileDiffStr := ""
	if fileDiff != 0 {
		fileDiffStr = fmt.Sprintf(" (%+d)", fileDiff)
	}

	dirDiffStr := ""
	if dirDiff != 0 {
		dirDiffStr = fmt.Sprintf(" (%+d)", dirDiff)
	}

	fmt.Printf("[%s] Size: %10s%-12s | Files: %6d%-6s | Dirs: %5d%-5s\n",
		ts, sizeStr, diffStr, current.FileCount, fileDiffStr, current.DirCount, dirDiffStr)
}

// printDiskSummary shows a summary of disk usage changes over time.
func printDiskSummary(history []DiskSnapshot) {
	if len(history) < 2 {
		return
	}

	first := history[0]
	last := history[len(history)-1]
	duration := last.Time.Sub(first.Time)

	fmt.Println("\n--- Disk Monitor Summary ---")
	fmt.Printf("  Duration:      %v\n", duration.Round(time.Second))
	fmt.Printf("  Snapshots:     %d\n", len(history))
	fmt.Printf("  Initial size:  %s (%d files)\n", humanSize(first.TotalSize), first.FileCount)
	fmt.Printf("  Final size:    %s (%d files)\n", humanSize(last.TotalSize), last.FileCount)

	sizeDiff := last.TotalSize - first.TotalSize
	if sizeDiff > 0 {
		fmt.Printf("  Size change:   +%s\n", humanSize(sizeDiff))
	} else if sizeDiff < 0 {
		fmt.Printf("  Size change:   -%s\n", humanSize(-sizeDiff))
	} else {
		fmt.Println("  Size change:   (no change)")
	}

	fileDiff := last.FileCount - first.FileCount
	if fileDiff != 0 {
		fmt.Printf("  File change:   %+d files\n", fileDiff)
	}

	// Find peak size
	var peakSize int64
	var peakTime time.Time
	for _, snap := range history {
		if snap.TotalSize > peakSize {
			peakSize = snap.TotalSize
			peakTime = snap.Time
		}
	}
	if peakSize > first.TotalSize {
		fmt.Printf("  Peak size:     %s at %s\n", humanSize(peakSize), peakTime.Format("15:04:05"))
	}
}

// ========================================
// Utility Functions
// ========================================

// timestamp returns the current time formatted for log output.
func timestamp() string {
	return time.Now().Format("15:04:05")
}

// humanSize converts bytes to a human-readable string.
func humanSize(bytes int64) string {
	negative := false
	if bytes < 0 {
		negative = true
		bytes = -bytes
	}

	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	var result string
	switch {
	case bytes >= GB:
		result = fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		result = fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		result = fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		result = fmt.Sprintf("%d B", bytes)
	}

	if negative {
		result = "-" + result
	}
	return result
}
