package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ========================================
// Week 15, Lesson 3: Disk Usage Analyzer
// ========================================
// A disk usage analyzer that walks a directory tree, calculates
// sizes, displays the top-N largest files and directories, and
// presents a tree-like output with human-readable sizes.
//
// Similar to the Unix 'du' and 'ncdu' commands.
//
// Usage:
//   go run main.go                   # Analyze current directory
//   go run main.go /path/to/dir      # Analyze specific directory
//   go run main.go /path -n 20       # Show top 20 largest
//   go run main.go /path --tree      # Show tree view
//   go run main.go /path --all       # Include hidden files
// ========================================

// ========================================
// Types
// ========================================

// DirEntry holds size information for a file or directory.
type DirEntry struct {
	Path     string
	RelPath  string
	Size     int64
	IsDir    bool
	Children int    // Number of files in directory
	Depth    int
}

// Config holds command-line options.
type Config struct {
	RootDir    string
	TopN       int
	ShowTree   bool
	ShowHidden bool
	MaxDepth   int
}

func main() {
	config := parseArgs()

	// Resolve and validate path
	absPath, err := filepath.Abs(config.RootDir)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: %s is not a directory\n", absPath)
		os.Exit(1)
	}

	fmt.Println("========================================")
	fmt.Println("Disk Usage Analyzer")
	fmt.Println("========================================")
	fmt.Printf("Scanning: %s\n", absPath)
	fmt.Println()

	// Scan the directory tree
	files, dirs, totalSize, err := scanDirectory(absPath, config)
	if err != nil {
		fmt.Printf("Error scanning: %v\n", err)
		os.Exit(1)
	}

	// Display results
	printSummary(absPath, files, dirs, totalSize)
	printTopFiles(files, config.TopN)
	printTopDirs(dirs, config.TopN)
	printSizeDistribution(files)
	printExtensionBreakdown(files, 10)

	if config.ShowTree {
		fmt.Println("\n========================================")
		fmt.Println("Directory Tree")
		fmt.Println("========================================")
		printDirTree(absPath, config.MaxDepth, config.ShowHidden)
	}

	fmt.Println("\n========================================")
	fmt.Println("Analysis complete!")
	fmt.Println("========================================")
}

// ========================================
// Argument Parsing
// ========================================

func parseArgs() Config {
	config := Config{
		RootDir:    ".",
		TopN:       10,
		ShowTree:   false,
		ShowHidden: false,
		MaxDepth:   3,
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-n", "--top":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.TopN)
				i++
			}
		case "--tree":
			config.ShowTree = true
		case "--all", "-a":
			config.ShowHidden = true
		case "--depth", "-d":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &config.MaxDepth)
				i++
			}
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		default:
			if !strings.HasPrefix(args[i], "-") {
				config.RootDir = args[i]
			}
		}
	}

	return config
}

func printHelp() {
	fmt.Println("Disk Usage Analyzer")
	fmt.Println()
	fmt.Println("Usage: go run main.go [directory] [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -n, --top N     Show top N largest items (default: 10)")
	fmt.Println("  --tree          Show directory tree view")
	fmt.Println("  -a, --all       Include hidden files/directories")
	fmt.Println("  -d, --depth N   Tree depth limit (default: 3)")
	fmt.Println("  -h, --help      Show this help")
}

// ========================================
// Directory Scanning
// ========================================

// scanDirectory walks the directory tree and collects size data.
func scanDirectory(root string, config Config) ([]DirEntry, []DirEntry, int64, error) {
	var files []DirEntry
	dirSizes := make(map[string]*DirEntry)
	dirChildren := make(map[string]int)
	var totalSize int64

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		name := d.Name()

		// Skip hidden files unless --all
		if !config.ShowHidden && strings.HasPrefix(name, ".") && path != root {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "." {
			relPath = filepath.Base(root)
		}

		depth := strings.Count(relPath, string(filepath.Separator))

		if d.IsDir() {
			dirSizes[path] = &DirEntry{
				Path:    path,
				RelPath: relPath,
				IsDir:   true,
				Depth:   depth,
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		size := info.Size()
		totalSize += size

		files = append(files, DirEntry{
			Path:    path,
			RelPath: relPath,
			Size:    size,
			IsDir:   false,
			Depth:   depth,
		})

		// Add size to all parent directories
		dir := filepath.Dir(path)
		for dir != filepath.Dir(root) && dir != root+"/.." {
			if entry, ok := dirSizes[dir]; ok {
				entry.Size += size
				dirChildren[dir]++
			}
			parentDir := filepath.Dir(dir)
			if parentDir == dir {
				break // Reached filesystem root
			}
			dir = parentDir
		}
		// Add to root directory
		if entry, ok := dirSizes[root]; ok {
			entry.Size += size
			dirChildren[root]++
		}

		return nil
	})

	// Convert dir map to slice and set children counts
	dirs := make([]DirEntry, 0, len(dirSizes))
	for path, entry := range dirSizes {
		entry.Children = dirChildren[path]
		dirs = append(dirs, *entry)
	}

	return files, dirs, totalSize, err
}

// ========================================
// Display Functions
// ========================================

// printSummary shows overall disk usage statistics.
func printSummary(root string, files, dirs []DirEntry, totalSize int64) {
	fmt.Println("--- Summary ---")
	fmt.Printf("  Directory:     %s\n", root)
	fmt.Printf("  Total size:    %s\n", humanSize(totalSize))
	fmt.Printf("  Files:         %d\n", len(files))
	fmt.Printf("  Directories:   %d\n", len(dirs))

	if len(files) > 0 {
		avgSize := totalSize / int64(len(files))
		fmt.Printf("  Average file:  %s\n", humanSize(avgSize))
	}
}

// printTopFiles displays the largest files.
func printTopFiles(files []DirEntry, n int) {
	if len(files) == 0 {
		return
	}

	// Sort by size descending
	sorted := make([]DirEntry, len(files))
	copy(sorted, files)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Size > sorted[j].Size
	})

	if n > len(sorted) {
		n = len(sorted)
	}

	fmt.Printf("\n--- Top %d Largest Files ---\n", n)
	maxSize := sorted[0].Size

	for i := 0; i < n; i++ {
		f := sorted[i]
		bar := sizeBar(f.Size, maxSize, 20)
		fmt.Printf("  %10s  %s  %s\n", humanSize(f.Size), bar, f.RelPath)
	}
}

// printTopDirs displays the largest directories.
func printTopDirs(dirs []DirEntry, n int) {
	if len(dirs) == 0 {
		return
	}

	// Sort by size descending
	sorted := make([]DirEntry, len(dirs))
	copy(sorted, dirs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Size > sorted[j].Size
	})

	if n > len(sorted) {
		n = len(sorted)
	}

	fmt.Printf("\n--- Top %d Largest Directories ---\n", n)
	maxSize := sorted[0].Size

	for i := 0; i < n; i++ {
		d := sorted[i]
		bar := sizeBar(d.Size, maxSize, 20)
		fmt.Printf("  %10s  %s  %s/ (%d files)\n",
			humanSize(d.Size), bar, d.RelPath, d.Children)
	}
}

// printSizeDistribution shows the distribution of file sizes.
func printSizeDistribution(files []DirEntry) {
	if len(files) == 0 {
		return
	}

	fmt.Println("\n--- Size Distribution ---")

	buckets := []struct {
		label string
		min   int64
		max   int64
		count int
	}{
		{"    < 1 KB", 0, 1024, 0},
		{" 1-10 KB", 1024, 10 * 1024, 0},
		{"10-100 KB", 10 * 1024, 100 * 1024, 0},
		{"100 KB-1 MB", 100 * 1024, 1024 * 1024, 0},
		{"  1-10 MB", 1024 * 1024, 10 * 1024 * 1024, 0},
		{" 10-100 MB", 10 * 1024 * 1024, 100 * 1024 * 1024, 0},
		{"  > 100 MB", 100 * 1024 * 1024, 1<<62, 0},
	}

	for _, f := range files {
		for i := range buckets {
			if f.Size >= buckets[i].min && f.Size < buckets[i].max {
				buckets[i].count++
				break
			}
		}
	}

	maxCount := 0
	for _, b := range buckets {
		if b.count > maxCount {
			maxCount = b.count
		}
	}

	for _, b := range buckets {
		if b.count > 0 {
			barLen := 1
			if maxCount > 0 {
				barLen = b.count * 30 / maxCount
				if barLen < 1 {
					barLen = 1
				}
			}
			bar := strings.Repeat("#", barLen)
			fmt.Printf("  %11s  %5d files  %s\n", b.label, b.count, bar)
		}
	}
}

// printExtensionBreakdown shows size by file extension.
func printExtensionBreakdown(files []DirEntry, n int) {
	if len(files) == 0 {
		return
	}

	extSizes := make(map[string]int64)
	extCounts := make(map[string]int)

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.RelPath))
		if ext == "" {
			ext = "(none)"
		}
		extSizes[ext] += f.Size
		extCounts[ext]++
	}

	// Sort by size
	type extInfo struct {
		ext   string
		size  int64
		count int
	}
	exts := make([]extInfo, 0, len(extSizes))
	for ext, size := range extSizes {
		exts = append(exts, extInfo{ext, size, extCounts[ext]})
	}
	sort.Slice(exts, func(i, j int) bool {
		return exts[i].size > exts[j].size
	})

	if n > len(exts) {
		n = len(exts)
	}

	fmt.Printf("\n--- Top %d File Types by Size ---\n", n)
	for i := 0; i < n; i++ {
		e := exts[i]
		fmt.Printf("  %-10s  %10s  %5d files\n", e.ext, humanSize(e.size), e.count)
	}
}

// printDirTree prints a visual directory tree with sizes.
func printDirTree(root string, maxDepth int, showHidden bool) {
	printTreeEntry(root, "", true, 0, maxDepth, showHidden)
}

func printTreeEntry(path string, prefix string, isRoot bool, depth int, maxDepth int, showHidden bool) {
	if depth > maxDepth {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		return
	}

	name := filepath.Base(path)

	if isRoot {
		size := calcDirSize(path, showHidden)
		fmt.Printf("  %s (%s)\n", name, humanSize(size))
	}

	if !info.IsDir() {
		return
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	// Filter entries
	var filtered []os.DirEntry
	for _, e := range entries {
		if !showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		filtered = append(filtered, e)
	}

	// Sort: directories first, then by name
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].IsDir() != filtered[j].IsDir() {
			return filtered[i].IsDir()
		}
		return filtered[i].Name() < filtered[j].Name()
	})

	for i, entry := range filtered {
		isLast := i == len(filtered)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		childPath := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			size := calcDirSize(childPath, showHidden)
			fmt.Printf("  %s%s%s/ (%s)\n", prefix, connector, entry.Name(), humanSize(size))

			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			printTreeEntry(childPath, newPrefix, false, depth+1, maxDepth, showHidden)
		} else {
			finfo, err := entry.Info()
			if err != nil {
				continue
			}
			fmt.Printf("  %s%s%s (%s)\n", prefix, connector, entry.Name(), humanSize(finfo.Size()))
		}
	}
}

// calcDirSize calculates the total size of a directory recursively.
func calcDirSize(dir string, showHidden bool) int64 {
	var total int64
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !showHidden && strings.HasPrefix(d.Name(), ".") && path != dir {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

// ========================================
// Utility Functions
// ========================================

// humanSize converts bytes to human-readable format.
func humanSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TB", float64(bytes)/float64(TB))
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

// sizeBar creates a proportional bar for visualization.
func sizeBar(size, maxSize int64, maxLen int) string {
	if maxSize == 0 {
		return ""
	}
	barLen := int(float64(size) / float64(maxSize) * float64(maxLen))
	if barLen < 1 && size > 0 {
		barLen = 1
	}
	return "[" + strings.Repeat("#", barLen) + strings.Repeat(" ", maxLen-barLen) + "]"
}
