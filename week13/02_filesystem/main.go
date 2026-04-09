package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ========================================
// Week 13, Lesson 2: Advanced Filesystem Operations
// ========================================
// Go's path/filepath, os, and io/fs packages provide powerful
// tools for working with the filesystem. This lesson covers
// directory walking, file watching (polling), symlinks, file
// locking, and recursive directory operations.
// ========================================

func main() {
	// Create a temporary workspace for our experiments
	workspace, err := os.MkdirTemp("", "fs-lesson-*")
	if err != nil {
		fmt.Printf("Error creating workspace: %v\n", err)
		return
	}
	defer os.RemoveAll(workspace) // Clean up when done
	fmt.Printf("Workspace: %s\n\n", workspace)

	// Set up test directory structure
	setupTestDirs(workspace)

	// ========================================
	// 1. filepath.Walk and filepath.WalkDir
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. filepath.Walk and filepath.WalkDir")
	fmt.Println("========================================")

	// filepath.WalkDir is the newer, more efficient version of filepath.Walk.
	// It uses fs.DirEntry instead of os.FileInfo, avoiding extra stat calls.

	fmt.Println("\nWalking directory tree with WalkDir:")
	err = filepath.WalkDir(workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("  Error accessing %s: %v\n", path, err)
			return nil // Continue walking despite errors
		}

		// Calculate relative path for cleaner output
		relPath, _ := filepath.Rel(workspace, path)
		if relPath == "." {
			relPath = "(root)"
		}

		// Show indentation based on depth
		depth := strings.Count(relPath, string(filepath.Separator))
		indent := strings.Repeat("  ", depth)

		if d.IsDir() {
			fmt.Printf("  %s[DIR]  %s/\n", indent, d.Name())
		} else {
			info, _ := d.Info()
			size := int64(0)
			if info != nil {
				size = info.Size()
			}
			fmt.Printf("  %s[FILE] %s (%d bytes)\n", indent, d.Name(), size)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Walk error: %v\n", err)
	}

	// Walking with filtering — find only .txt files
	fmt.Println("\nFinding all .txt files:")
	var txtFiles []string
	filepath.WalkDir(workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && filepath.Ext(path) == ".txt" {
			relPath, _ := filepath.Rel(workspace, path)
			txtFiles = append(txtFiles, relPath)
		}
		return nil
	})
	for _, f := range txtFiles {
		fmt.Printf("  %s\n", f)
	}

	// Skipping directories
	fmt.Println("\nWalking but skipping 'logs' directory:")
	filepath.WalkDir(workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "logs" {
			fmt.Printf("  Skipping directory: %s\n", d.Name())
			return filepath.SkipDir // Skip this directory entirely
		}
		if !d.IsDir() {
			relPath, _ := filepath.Rel(workspace, path)
			fmt.Printf("  %s\n", relPath)
		}
		return nil
	})

	// ========================================
	// 2. File Watching (Polling Approach)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. File Watching (Polling Approach)")
	fmt.Println("========================================")

	// Go's standard library doesn't include a file watcher.
	// The polling approach checks file modification times at intervals.
	// For production, consider the fsnotify package (third-party).

	fmt.Println("\nDemonstrating file change detection:")
	watchDir := filepath.Join(workspace, "watched")
	os.MkdirAll(watchDir, 0755)

	// Take initial snapshot
	snapshot1 := takeSnapshot(watchDir)
	fmt.Printf("  Initial snapshot: %d files\n", len(snapshot1))

	// Make some changes
	os.WriteFile(filepath.Join(watchDir, "new_file.txt"), []byte("created"), 0644)
	time.Sleep(10 * time.Millisecond) // Ensure timestamp differs

	// Take new snapshot and compare
	snapshot2 := takeSnapshot(watchDir)
	changes := detectChanges(snapshot1, snapshot2)
	fmt.Printf("  After creating new_file.txt:\n")
	for _, change := range changes {
		fmt.Printf("    %s: %s\n", change.Type, change.Path)
	}

	// Modify a file
	os.WriteFile(filepath.Join(watchDir, "new_file.txt"), []byte("modified content"), 0644)
	time.Sleep(10 * time.Millisecond)

	snapshot3 := takeSnapshot(watchDir)
	changes = detectChanges(snapshot2, snapshot3)
	fmt.Printf("  After modifying new_file.txt:\n")
	for _, change := range changes {
		fmt.Printf("    %s: %s\n", change.Type, change.Path)
	}

	// Delete a file
	os.Remove(filepath.Join(watchDir, "new_file.txt"))
	snapshot4 := takeSnapshot(watchDir)
	changes = detectChanges(snapshot3, snapshot4)
	fmt.Printf("  After deleting new_file.txt:\n")
	for _, change := range changes {
		fmt.Printf("    %s: %s\n", change.Type, change.Path)
	}

	// ========================================
	// 3. Symlinks
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Symbolic Links")
	fmt.Println("========================================")

	// Symbolic links (symlinks) are pointers to other files/directories.
	// os.Symlink creates them, os.Readlink reads the target.

	srcFile := filepath.Join(workspace, "src", "main.go")
	linkPath := filepath.Join(workspace, "main_link.go")

	// Create a symlink
	err = os.Symlink(srcFile, linkPath)
	if err != nil {
		fmt.Printf("\nError creating symlink: %v\n", err)
	} else {
		fmt.Printf("\nCreated symlink: %s -> %s\n", filepath.Base(linkPath), filepath.Base(srcFile))
	}

	// Read symlink target
	target, err := os.Readlink(linkPath)
	if err != nil {
		fmt.Printf("Error reading symlink: %v\n", err)
	} else {
		fmt.Printf("Symlink target: %s\n", target)
	}

	// os.Stat follows symlinks (gets info about the TARGET)
	info, err := os.Stat(linkPath)
	if err != nil {
		fmt.Printf("Stat error: %v\n", err)
	} else {
		fmt.Printf("os.Stat (follows link): name=%s, size=%d\n", info.Name(), info.Size())
	}

	// os.Lstat does NOT follow symlinks (gets info about the LINK itself)
	linfo, err := os.Lstat(linkPath)
	if err != nil {
		fmt.Printf("Lstat error: %v\n", err)
	} else {
		fmt.Printf("os.Lstat (link itself): name=%s, mode=%s\n", linfo.Name(), linfo.Mode())
		isSymlink := linfo.Mode()&os.ModeSymlink != 0
		fmt.Printf("Is symlink: %v\n", isSymlink)
	}

	// Walking with symlink detection
	fmt.Println("\nDetecting symlinks during walk:")
	filepath.WalkDir(workspace, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			target, _ := os.Readlink(path)
			relPath, _ := filepath.Rel(workspace, path)
			fmt.Printf("  Symlink: %s -> %s\n", relPath, target)
		}
		return nil
	})

	// ========================================
	// 4. File Locking (Simple Approach)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. File Locking (Simple Approach)")
	fmt.Println("========================================")

	// Go doesn't have built-in file locking in the standard library.
	// Common approaches:
	// 1. Lock files (create a .lock file)
	// 2. Advisory locking with syscall.Flock (Unix)
	// We'll demonstrate the lock file approach (cross-platform).

	fmt.Println("\nDemonstrating lock file approach:")
	lockPath := filepath.Join(workspace, "app.lock")

	// Acquire lock
	acquired, err := acquireLock(lockPath)
	if err != nil {
		fmt.Printf("  Error acquiring lock: %v\n", err)
	} else if acquired {
		fmt.Println("  Lock acquired successfully.")

		// Try to acquire again (should fail)
		acquired2, _ := acquireLock(lockPath)
		fmt.Printf("  Second lock attempt: acquired=%v (expected false)\n", acquired2)

		// Release lock
		releaseLock(lockPath)
		fmt.Println("  Lock released.")

		// Now it should succeed again
		acquired3, _ := acquireLock(lockPath)
		fmt.Printf("  Third lock attempt after release: acquired=%v\n", acquired3)
		releaseLock(lockPath)
	}

	// ========================================
	// 5. Recursive Directory Operations
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Recursive Directory Operations")
	fmt.Println("========================================")

	// Recursive copy
	fmt.Println("\nRecursive directory copy:")
	srcDir := filepath.Join(workspace, "src")
	dstDir := filepath.Join(workspace, "src_backup")

	err = copyDirRecursive(srcDir, dstDir)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Println("  Copied src/ to src_backup/")
		// Verify
		var count int
		filepath.WalkDir(dstDir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				count++
			}
			return nil
		})
		fmt.Printf("  Files in backup: %d\n", count)
	}

	// Recursive file search with content matching
	fmt.Println("\nRecursive content search (grep-like):")
	results := searchFiles(workspace, "important")
	for _, r := range results {
		relPath, _ := filepath.Rel(workspace, r.Path)
		fmt.Printf("  %s (line %d): %s\n", relPath, r.Line, r.Content)
	}

	// Calculate directory size
	fmt.Println("\nRecursive directory size calculation:")
	size := dirSize(workspace)
	fmt.Printf("  Total workspace size: %s\n", humanSize(size))

	// Directory tree display
	fmt.Println("\nDirectory tree:")
	printTree(workspace, "", true)

	// ========================================
	// 6. Filepath Utilities
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Filepath Utilities")
	fmt.Println("========================================")

	testPath := "/home/user/documents/report.pdf"
	fmt.Printf("\nPath: %s\n", testPath)
	fmt.Printf("  Dir:     %s\n", filepath.Dir(testPath))
	fmt.Printf("  Base:    %s\n", filepath.Base(testPath))
	fmt.Printf("  Ext:     %s\n", filepath.Ext(testPath))
	fmt.Printf("  Clean:   %s\n", filepath.Clean("/home/user/../user/./docs/"))
	fmt.Printf("  Join:    %s\n", filepath.Join("/home", "user", "docs", "file.txt"))
	fmt.Printf("  IsAbs:   %v\n", filepath.IsAbs(testPath))
	fmt.Printf("  IsAbs:   %v (relative path)\n", filepath.IsAbs("docs/file.txt"))

	// filepath.Match for glob pattern matching
	fmt.Println("\nGlob pattern matching:")
	testMatches := []struct {
		pattern, name string
	}{
		{"*.go", "main.go"},
		{"*.go", "main.txt"},
		{"test_*.go", "test_util.go"},
		{"doc[0-9]*", "doc5.txt"},
		{"doc[0-9]*", "docs.txt"},
	}
	for _, tm := range testMatches {
		matched, _ := filepath.Match(tm.pattern, tm.name)
		fmt.Printf("  Match(%q, %q) = %v\n", tm.pattern, tm.name, matched)
	}

	// filepath.Glob finds files matching a pattern
	fmt.Println("\nGlob search for *.txt files in workspace:")
	matches, _ := filepath.Glob(filepath.Join(workspace, "**", "*.txt"))
	// Note: filepath.Glob doesn't support ** (double star).
	// Use WalkDir for recursive glob-like behavior.
	matches, _ = filepath.Glob(filepath.Join(workspace, "*.txt"))
	for _, m := range matches {
		relPath, _ := filepath.Rel(workspace, m)
		fmt.Printf("  %s\n", relPath)
	}

	// ========================================
	// 7. File Checksums
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. File Checksums")
	fmt.Println("========================================")

	checksumFile := filepath.Join(workspace, "src", "main.go")
	hash, err := fileChecksum(checksumFile)
	if err != nil {
		fmt.Printf("\nError: %v\n", err)
	} else {
		relPath, _ := filepath.Rel(workspace, checksumFile)
		fmt.Printf("\nSHA-256 of %s:\n  %s\n", relPath, hash)
	}

	fmt.Println("\n========================================")
	fmt.Println("Advanced Filesystem lesson complete!")
	fmt.Println("========================================")
}

// ========================================
// Setup and Utility Types
// ========================================

// setupTestDirs creates a test directory structure for demonstrations.
func setupTestDirs(base string) {
	// Create directory structure
	dirs := []string{
		"src",
		"src/utils",
		"docs",
		"logs",
		"config",
		"watched",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(base, d), 0755)
	}

	// Create test files
	files := map[string]string{
		"src/main.go":      "package main\n\nfunc main() {\n\t// important entry point\n}\n",
		"src/utils/util.go": "package utils\n\n// Helper is an important utility\nfunc Helper() {}\n",
		"docs/readme.txt":  "This is the readme file.\nIt contains important information.\n",
		"docs/notes.txt":   "Development notes go here.\n",
		"logs/app.log":     "2024-01-01 INFO Starting application\n2024-01-01 ERROR Something failed\n",
		"config/app.yaml":  "port: 8080\nhost: localhost\n",
	}
	for path, content := range files {
		os.WriteFile(filepath.Join(base, path), []byte(content), 0644)
	}
}

// ========================================
// File Watching Types and Functions
// ========================================

// FileSnapshot represents a point-in-time view of a directory's contents.
type FileSnapshot map[string]FileState

// FileState holds metadata about a single file.
type FileState struct {
	ModTime time.Time
	Size    int64
	IsDir   bool
}

// FileChange represents a detected change.
type FileChange struct {
	Type string // "created", "modified", "deleted"
	Path string
}

// takeSnapshot records the current state of all files in a directory.
func takeSnapshot(dir string) FileSnapshot {
	snapshot := make(FileSnapshot)
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == dir {
			return nil
		}
		relPath, _ := filepath.Rel(dir, path)
		info, err := d.Info()
		if err != nil {
			return nil
		}
		snapshot[relPath] = FileState{
			ModTime: info.ModTime(),
			Size:    info.Size(),
			IsDir:   d.IsDir(),
		}
		return nil
	})
	return snapshot
}

// detectChanges compares two snapshots and returns detected changes.
func detectChanges(old, new FileSnapshot) []FileChange {
	var changes []FileChange

	// Check for created and modified files
	for path, newState := range new {
		oldState, existed := old[path]
		if !existed {
			changes = append(changes, FileChange{Type: "created", Path: path})
		} else if newState.ModTime != oldState.ModTime || newState.Size != oldState.Size {
			changes = append(changes, FileChange{Type: "modified", Path: path})
		}
	}

	// Check for deleted files
	for path := range old {
		if _, exists := new[path]; !exists {
			changes = append(changes, FileChange{Type: "deleted", Path: path})
		}
	}

	return changes
}

// ========================================
// File Locking
// ========================================

// acquireLock attempts to create a lock file. Returns true if acquired.
func acquireLock(lockPath string) (bool, error) {
	// O_CREATE|O_EXCL ensures the file is created atomically.
	// If the file already exists, OpenFile returns an error.
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return false, nil // Lock already held
		}
		return false, err
	}
	// Write PID to lock file for debugging
	fmt.Fprintf(f, "%d\n", os.Getpid())
	f.Close()
	return true, nil
}

// releaseLock removes the lock file.
func releaseLock(lockPath string) error {
	return os.Remove(lockPath)
}

// ========================================
// Recursive Operations
// ========================================

// copyDirRecursive copies a directory tree from src to dst.
func copyDirRecursive(src, dst string) error {
	// Get source info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	// Create destination directory
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	// Walk the source tree
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// SearchResult holds a file search match.
type SearchResult struct {
	Path    string
	Line    int
	Content string
}

// searchFiles searches for a string in all files under root.
func searchFiles(root, query string) []SearchResult {
	var results []SearchResult

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Only search text files (simple heuristic: skip binary-looking files)
		ext := filepath.Ext(path)
		textExts := map[string]bool{
			".go": true, ".txt": true, ".log": true,
			".yaml": true, ".json": true, ".md": true,
		}
		if !textExts[ext] {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
				results = append(results, SearchResult{
					Path:    path,
					Line:    lineNum,
					Content: strings.TrimSpace(line),
				})
			}
		}
		return nil
	})

	return results
}

// dirSize calculates the total size of all files in a directory tree.
func dirSize(path string) int64 {
	var total int64
	filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}

// printTree displays a directory tree with ASCII art.
func printTree(root string, prefix string, isLast bool) {
	info, err := os.Stat(root)
	if err != nil {
		return
	}

	// Print current entry
	name := filepath.Base(root)
	if prefix == "" {
		fmt.Printf("  %s/\n", name)
	} else {
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		if info.IsDir() {
			fmt.Printf("  %s%s%s/\n", prefix, connector, name)
		} else {
			fmt.Printf("  %s%s%s (%d bytes)\n", prefix, connector, name, info.Size())
		}
	}

	if !info.IsDir() {
		return
	}

	// Read directory entries
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	// Recurse into children
	for i, entry := range entries {
		childPath := filepath.Join(root, entry.Name())
		newPrefix := prefix
		if prefix != "" {
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
		}
		printTree(childPath, newPrefix, i == len(entries)-1)
	}
}

// fileChecksum computes SHA-256 hash of a file.
func fileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// humanSize converts bytes to human-readable format.
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

