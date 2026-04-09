package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ========================================
// Week 15, Lesson 1: File Watcher
// ========================================
// A file watcher monitors a directory for changes by polling file
// modification times at regular intervals. It reports created,
// modified, and deleted files. This is a cross-platform approach
// that works on macOS, Linux, and Windows.
//
// For production use, consider the fsnotify package which uses
// OS-level file system event APIs (inotify on Linux, FSEvents on
// macOS, ReadDirectoryChangesW on Windows).
//
// Usage:
//   go run main.go                    # Watch current directory
//   go run main.go /path/to/dir       # Watch specific directory
//   go run main.go /path 2            # Watch with 2-second interval
//   go run main.go /path 1 --verbose  # Verbose mode
//
// Press Ctrl+C to stop watching.
// ========================================

// ========================================
// Types
// ========================================

// FileState holds metadata about a file at a point in time.
type FileState struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Mode    fs.FileMode
}

// FileEvent represents a detected file change.
type FileEvent struct {
	Type     string // "created", "modified", "deleted"
	Path     string
	RelPath  string
	OldState *FileState
	NewState *FileState
}

// Watcher monitors a directory for file changes.
type Watcher struct {
	dir      string
	interval time.Duration
	verbose  bool
	snapshot map[string]FileState
	events   chan FileEvent
	stats    WatcherStats
}

// WatcherStats tracks overall watcher statistics.
type WatcherStats struct {
	Created   int
	Modified  int
	Deleted   int
	Polls     int
	StartTime time.Time
}

func main() {
	// Parse arguments
	dir := "."
	interval := 2 * time.Second
	verbose := false

	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	if len(os.Args) > 2 {
		if sec, err := time.ParseDuration(os.Args[2] + "s"); err == nil {
			interval = sec
		}
	}
	for _, arg := range os.Args {
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		}
	}

	// Resolve to absolute path
	absDir, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	// Verify directory exists
	info, err := os.Stat(absDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", absDir)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("File Watcher")
	fmt.Println("========================================")
	fmt.Printf("Watching:  %s\n", absDir)
	fmt.Printf("Interval:  %v\n", interval)
	fmt.Printf("Verbose:   %v\n", verbose)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println("========================================")

	// Create and start watcher
	watcher := NewWatcher(absDir, interval, verbose)

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Run the watcher
	watcher.Watch(ctx)

	// Print summary
	watcher.PrintSummary()
}

// ========================================
// Watcher Implementation
// ========================================

// NewWatcher creates a new file watcher.
func NewWatcher(dir string, interval time.Duration, verbose bool) *Watcher {
	return &Watcher{
		dir:      dir,
		interval: interval,
		verbose:  verbose,
		snapshot: make(map[string]FileState),
		events:   make(chan FileEvent, 100),
		stats: WatcherStats{
			StartTime: time.Now(),
		},
	}
}

// Watch starts the file watching loop.
func (w *Watcher) Watch(ctx context.Context) {
	// Take initial snapshot
	w.snapshot = w.takeSnapshot()
	fmt.Printf("\nInitial scan: %d files found.\n\n", len(w.snapshot))

	if w.verbose {
		for relPath, state := range w.snapshot {
			fmt.Printf("  [INIT] %s (%s, %s)\n",
				relPath, humanSize(state.Size), state.ModTime.Format("15:04:05"))
		}
		fmt.Println()
	}

	// Start event processor
	go w.processEvents()

	// Polling loop
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n\nStopping file watcher...")
			return
		case <-ticker.C:
			w.poll()
		}
	}
}

// poll takes a new snapshot and compares it with the previous one.
func (w *Watcher) poll() {
	w.stats.Polls++
	newSnapshot := w.takeSnapshot()

	// Detect created and modified files
	for path, newState := range newSnapshot {
		oldState, existed := w.snapshot[path]
		if !existed {
			w.events <- FileEvent{
				Type:     "created",
				Path:     filepath.Join(w.dir, path),
				RelPath:  path,
				NewState: &newState,
			}
		} else if newState.ModTime != oldState.ModTime || newState.Size != oldState.Size {
			oldCopy := oldState // copy for pointer
			w.events <- FileEvent{
				Type:     "modified",
				Path:     filepath.Join(w.dir, path),
				RelPath:  path,
				OldState: &oldCopy,
				NewState: &newState,
			}
		}
	}

	// Detect deleted files
	for path, oldState := range w.snapshot {
		if _, exists := newSnapshot[path]; !exists {
			oldCopy := oldState
			w.events <- FileEvent{
				Type:     "deleted",
				Path:     filepath.Join(w.dir, path),
				RelPath:  path,
				OldState: &oldCopy,
			}
		}
	}

	// Update snapshot
	w.snapshot = newSnapshot
}

// processEvents handles detected file events.
func (w *Watcher) processEvents() {
	for event := range w.events {
		timestamp := time.Now().Format("15:04:05")

		switch event.Type {
		case "created":
			w.stats.Created++
			icon := "+"
			typeStr := "FILE"
			if event.NewState != nil && event.NewState.IsDir {
				typeStr = "DIR "
			}
			fmt.Printf("[%s] %s CREATED  [%s] %s", timestamp, icon, typeStr, event.RelPath)
			if event.NewState != nil {
				fmt.Printf(" (%s)", humanSize(event.NewState.Size))
			}
			fmt.Println()

		case "modified":
			w.stats.Modified++
			icon := "~"
			fmt.Printf("[%s] %s MODIFIED       %s", timestamp, icon, event.RelPath)
			if event.OldState != nil && event.NewState != nil {
				sizeDiff := event.NewState.Size - event.OldState.Size
				if sizeDiff > 0 {
					fmt.Printf(" (+%s)", humanSize(sizeDiff))
				} else if sizeDiff < 0 {
					fmt.Printf(" (-%s)", humanSize(-sizeDiff))
				}
			}
			fmt.Println()

		case "deleted":
			w.stats.Deleted++
			icon := "-"
			fmt.Printf("[%s] %s DELETED        %s\n", timestamp, icon, event.RelPath)
		}
	}
}

// takeSnapshot scans the directory and records file metadata.
func (w *Watcher) takeSnapshot() map[string]FileState {
	snapshot := make(map[string]FileState)

	filepath.WalkDir(w.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden files and directories
		name := d.Name()
		if strings.HasPrefix(name, ".") && path != w.dir {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip the root directory itself
		if path == w.dir {
			return nil
		}

		relPath, err := filepath.Rel(w.dir, path)
		if err != nil {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		snapshot[relPath] = FileState{
			Path:    path,
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   d.IsDir(),
			Mode:    info.Mode(),
		}

		return nil
	})

	return snapshot
}

// PrintSummary displays watcher statistics.
func (w *Watcher) PrintSummary() {
	elapsed := time.Since(w.stats.StartTime)

	fmt.Println("\n========================================")
	fmt.Println("File Watcher Summary")
	fmt.Println("========================================")
	fmt.Printf("  Duration:   %v\n", elapsed.Round(time.Second))
	fmt.Printf("  Polls:      %d\n", w.stats.Polls)
	fmt.Printf("  Created:    %d\n", w.stats.Created)
	fmt.Printf("  Modified:   %d\n", w.stats.Modified)
	fmt.Printf("  Deleted:    %d\n", w.stats.Deleted)
	fmt.Printf("  Total events: %d\n", w.stats.Created+w.stats.Modified+w.stats.Deleted)
	fmt.Printf("  Final file count: %d\n", len(w.snapshot))
	fmt.Println("========================================")
}

// ========================================
// Utility Functions
// ========================================

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
