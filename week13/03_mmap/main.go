package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	"unsafe"
)

// ========================================
// Week 13, Lesson 3: Memory-Mapped Files
// ========================================
// Memory-mapped files (mmap) allow you to map a file's contents
// directly into the process's virtual memory. Instead of explicit
// read/write system calls, you access file data as if it were
// a byte slice in memory. The OS handles paging data in and out.
//
// Benefits of mmap:
//   - Avoids copying data between kernel and user space
//   - Enables random access without seeking
//   - Multiple processes can share the same mapping
//   - Great for large files that don't fit in memory
//
// This lesson demonstrates mmap on macOS/Linux using syscall.
// On unsupported platforms, we show the concepts with a simulated
// approach.
//
// Usage: go run main.go
// ========================================

func main() {
	// ========================================
	// 1. Understanding Memory-Mapped Files
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. Understanding Memory-Mapped Files")
	fmt.Println("========================================")

	fmt.Print(`
  Traditional File I/O:
    Application -> read() syscall -> Kernel -> Disk -> Kernel buffer -> User buffer
    Two copies: disk->kernel, kernel->user

  Memory-Mapped I/O:
    Application -> Access memory address -> Page fault -> Kernel loads page from disk
    One copy: disk->shared memory page

  Key syscall: mmap(addr, length, prot, flags, fd, offset)
    - addr:   Desired memory address (usually 0, let OS choose)
    - length: How many bytes to map
    - prot:   Protection (PROT_READ, PROT_WRITE, PROT_EXEC)
    - flags:  MAP_SHARED (changes written to file) or MAP_PRIVATE (copy-on-write)
    - fd:     File descriptor
    - offset: Offset in the file to start mapping
`)
	fmt.Println()

	// Create a test workspace
	workspace, err := os.MkdirTemp("", "mmap-lesson-*")
	if err != nil {
		fmt.Printf("Error creating workspace: %v\n", err)
		return
	}
	defer os.RemoveAll(workspace)

	// ========================================
	// 2. Creating a Test File
	// ========================================
	fmt.Println("========================================")
	fmt.Println("2. Creating a Test File")
	fmt.Println("========================================")

	testFile := filepath.Join(workspace, "test_data.bin")
	createTestFile(testFile, 1024*1024) // 1 MB
	fmt.Printf("\nCreated test file: %s (1 MB)\n", testFile)

	// ========================================
	// 3. Memory-Mapped File Reading
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Memory-Mapped File Reading")
	fmt.Println("========================================")

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		fmt.Println("\nUsing real mmap on", runtime.GOOS)
		mmapReadDemo(testFile)
	} else {
		fmt.Printf("\nmmap via syscall not available on %s\n", runtime.GOOS)
		fmt.Println("Showing simulated approach instead.")
		simulatedMmapRead(testFile)
	}

	// ========================================
	// 4. Memory-Mapped File Writing
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Memory-Mapped File Writing")
	fmt.Println("========================================")

	writeFile := filepath.Join(workspace, "mmap_write.bin")
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		mmapWriteDemo(writeFile)
	} else {
		fmt.Println("mmap write demo requires macOS or Linux.")
		fmt.Println("Showing concept with standard I/O instead.")
		simulatedMmapWrite(writeFile)
	}

	// ========================================
	// 5. Performance Comparison
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Performance Comparison")
	fmt.Println("========================================")

	// Create a larger test file for benchmarking
	largeFile := filepath.Join(workspace, "large_file.bin")
	fileSize := 10 * 1024 * 1024 // 10 MB
	createTestFile(largeFile, fileSize)
	fmt.Printf("\nCreated large file: %d MB for benchmarking\n", fileSize/(1024*1024))

	// Standard read
	fmt.Println("\nMethod 1: Standard io.ReadAll")
	start := time.Now()
	standardReadAll(largeFile)
	fmt.Printf("  Time: %v\n", time.Since(start))

	// Buffered read
	fmt.Println("\nMethod 2: Buffered read (64KB chunks)")
	start = time.Now()
	bufferedRead(largeFile, 64*1024)
	fmt.Printf("  Time: %v\n", time.Since(start))

	// mmap read
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		fmt.Println("\nMethod 3: Memory-mapped read")
		start = time.Now()
		mmapReadBenchmark(largeFile)
		fmt.Printf("  Time: %v\n", time.Since(start))
	}

	fmt.Println("\nNote: Results vary by OS, disk type, and file caching.")
	fmt.Println("mmap shines for random access on large files.")

	// ========================================
	// 6. Random Access with mmap
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Random Access with mmap")
	fmt.Println("========================================")

	fmt.Println("\nmmap excels at random access patterns:")
	fmt.Println("  - Database index files")
	fmt.Println("  - Search engine inverted indexes")
	fmt.Println("  - Large binary file formats")
	fmt.Println("  - Shared memory between processes")

	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		mmapRandomAccessDemo(largeFile)
	}

	// ========================================
	// 7. mmap Caveats and Best Practices
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. mmap Caveats and Best Practices")
	fmt.Println("========================================")

	fmt.Print(`
  When to use mmap:
    + Reading large files (larger than available RAM)
    + Random access patterns (seeking to arbitrary positions)
    + Sharing data between processes (MAP_SHARED)
    + Read-only access to large datasets

  When NOT to use mmap:
    - Small files (overhead of page alignment not worth it)
    - Sequential reads (buffered I/O is simpler and competitive)
    - Files that grow dynamically (need to remap)
    - Network filesystems (unpredictable latency)

  Best practices:
    1. Always unmap (munmap) when done — avoid resource leaks
    2. Use MAP_PRIVATE for read-only access
    3. Handle SIGBUS — occurs if file is truncated while mapped
    4. Align offsets to page boundaries (usually 4KB)
    5. Don't mmap files larger than your address space
    6. Consider madvise hints for access patterns:
       - MADV_SEQUENTIAL: sequential access, prefetch ahead
       - MADV_RANDOM: random access, don't prefetch
       - MADV_WILLNEED: will need this soon, prefetch
       - MADV_DONTNEED: done with this, can reclaim
`)
	fmt.Println()

	// ========================================
	// 8. Cross-Platform Considerations
	// ========================================
	fmt.Println("========================================")
	fmt.Println("8. Cross-Platform Considerations")
	fmt.Println("========================================")

	fmt.Printf(`
  Current platform: %s/%s

  macOS and Linux:
    - Full mmap support via syscall.Mmap / syscall.Munmap
    - Page size typically 4KB (16KB on Apple Silicon)
    - syscall.Mmap(fd, offset, length, prot, flags)

  Windows:
    - Uses CreateFileMapping + MapViewOfFile
    - Go's syscall package supports this differently
    - Consider golang.org/x/sys/windows for production use

  Cross-platform libraries:
    - golang.org/x/exp/mmap (read-only, simple API)
    - github.com/edsrzf/mmap-go (full-featured)
    - These abstract away OS differences
`, runtime.GOOS, runtime.GOARCH)

	fmt.Println("\nPage size on this system:", os.Getpagesize(), "bytes")

	fmt.Println("\n========================================")
	fmt.Println("Memory-Mapped Files lesson complete!")
	fmt.Println("========================================")
}

// ========================================
// File Creation
// ========================================

// createTestFile creates a test file of the given size with patterned data.
func createTestFile(path string, size int) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	// Write patterned data so we can verify reads
	pattern := []byte("ABCDEFGHIJKLMNOP") // 16 bytes
	written := 0
	for written < size {
		remaining := size - written
		if remaining < len(pattern) {
			w.Write(pattern[:remaining])
			written += remaining
		} else {
			w.Write(pattern)
			written += len(pattern)
		}
	}
	w.Flush()
}

// ========================================
// mmap Implementations (macOS/Linux)
// ========================================

// mmapReadDemo demonstrates reading a file using mmap.
func mmapReadDemo(path string) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Printf("  Error opening file: %v\n", err)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		fmt.Printf("  Error stating file: %v\n", err)
		return
	}

	size := int(info.Size())
	fmt.Printf("  File size: %d bytes\n", size)

	// mmap the file as read-only
	data, err := syscall.Mmap(
		int(f.Fd()),           // file descriptor
		0,                     // offset (start of file)
		size,                  // length to map
		syscall.PROT_READ,     // read-only protection
		syscall.MAP_PRIVATE,   // private mapping (copy-on-write)
	)
	if err != nil {
		fmt.Printf("  Error mmap: %v\n", err)
		return
	}
	defer syscall.Munmap(data)

	fmt.Printf("  Successfully mapped %d bytes into memory\n", len(data))

	// Read data directly from the mapped memory
	fmt.Printf("  First 16 bytes: %q\n", string(data[:16]))
	fmt.Printf("  Bytes at offset 1000: %q\n", string(data[1000:1016]))
	fmt.Printf("  Last 16 bytes: %q\n", string(data[len(data)-16:]))

	// Count occurrences of a pattern
	pattern := byte('A')
	count := 0
	for _, b := range data {
		if b == pattern {
			count++
		}
	}
	fmt.Printf("  Occurrences of 'A': %d\n", count)
}

// mmapWriteDemo demonstrates writing to a file using mmap.
func mmapWriteDemo(path string) {
	// Create and size the file first
	size := 4096 // One page
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Printf("  Error creating file: %v\n", err)
		return
	}

	// File must be the right size before mapping
	if err := f.Truncate(int64(size)); err != nil {
		fmt.Printf("  Error sizing file: %v\n", err)
		f.Close()
		return
	}

	// mmap as read-write with shared mapping
	data, err := syscall.Mmap(
		int(f.Fd()),
		0,
		size,
		syscall.PROT_READ|syscall.PROT_WRITE, // read-write
		syscall.MAP_SHARED,                     // changes written to file
	)
	if err != nil {
		fmt.Printf("  Error mmap: %v\n", err)
		f.Close()
		return
	}

	// Write data through the memory map
	message := []byte("Hello from mmap! This was written directly to memory.")
	copy(data, message)

	// Write a second message at a different offset
	message2 := []byte("Second message at offset 100.")
	copy(data[100:], message2)

	// Sync changes to disk
	// On some systems, use syscall.Msync
	fmt.Println("\n  Written data through mmap:")
	fmt.Printf("  At offset 0: %q\n", string(data[:len(message)]))
	fmt.Printf("  At offset 100: %q\n", string(data[100:100+len(message2)]))

	// Unmap before closing
	syscall.Munmap(data)
	f.Close()

	// Verify by reading back with standard I/O
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  Error reading back: %v\n", err)
		return
	}
	fmt.Printf("  Verified (standard read, offset 0): %q\n",
		string(content[:len(message)]))
	fmt.Printf("  Verified (standard read, offset 100): %q\n",
		string(content[100:100+len(message2)]))
}

// mmapReadBenchmark reads an entire file using mmap (for benchmarking).
func mmapReadBenchmark(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	info, _ := f.Stat()
	size := int(info.Size())

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		fmt.Printf("  mmap error: %v\n", err)
		return
	}
	defer syscall.Munmap(data)

	// Touch every page to force loading (simulate reading all data)
	pageSize := os.Getpagesize()
	sum := 0
	for i := 0; i < len(data); i += pageSize {
		sum += int(data[i])
	}
	// Use sum to prevent compiler optimization
	_ = sum
	fmt.Printf("  Read %d bytes via mmap\n", size)
}

// mmapRandomAccessDemo shows random access performance with mmap.
func mmapRandomAccessDemo(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	info, _ := f.Stat()
	size := int(info.Size())

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_PRIVATE)
	if err != nil {
		fmt.Printf("  mmap error: %v\n", err)
		return
	}
	defer syscall.Munmap(data)

	fmt.Println("\nRandom access demo:")

	// Access specific offsets instantly — no seeking needed
	offsets := []int{0, size / 4, size / 2, 3 * size / 4, size - 16}
	for _, offset := range offsets {
		if offset+16 <= len(data) {
			fmt.Printf("  Offset %8d: %q\n", offset, string(data[offset:offset+16]))
		}
	}

	// Show the power: access the middle of a 10MB file instantly
	mid := size / 2
	fmt.Printf("\n  Middle of file (offset %d): %q\n", mid, string(data[mid:mid+16]))
	fmt.Println("  (No seeking, no buffering — just array indexing!)")
}

// ========================================
// Standard I/O Implementations (for comparison)
// ========================================

// standardReadAll reads an entire file using io.ReadAll.
func standardReadAll(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return
	}
	fmt.Printf("  Read %d bytes via ReadAll\n", len(data))
}

// bufferedRead reads a file in chunks using bufio.
func bufferedRead(path string, chunkSize int) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	buf := make([]byte, chunkSize)
	total := 0
	reader := bufio.NewReaderSize(f, chunkSize)
	for {
		n, err := reader.Read(buf)
		total += n
		if err != nil {
			break
		}
	}
	fmt.Printf("  Read %d bytes via buffered read (%d byte chunks)\n", total, chunkSize)
}

// simulatedMmapRead demonstrates the concept on unsupported platforms.
func simulatedMmapRead(path string) {
	fmt.Println("\n  [Simulated mmap read — using standard I/O underneath]")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Loaded %d bytes (would be mmap'd on macOS/Linux)\n", len(data))
	fmt.Printf("  First 16 bytes: %q\n", string(data[:16]))
	fmt.Printf("  Last 16 bytes: %q\n", string(data[len(data)-16:]))
}

// simulatedMmapWrite demonstrates the concept on unsupported platforms.
func simulatedMmapWrite(path string) {
	fmt.Println("\n  [Simulated mmap write — using standard I/O]")
	data := make([]byte, 4096)
	message := []byte("Hello from simulated mmap!")
	copy(data, message)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Wrote %d bytes (would be mmap'd on macOS/Linux)\n", len(data))
}

// Ensure unsafe is used (needed for potential pointer operations with mmap).
var _ = unsafe.Sizeof(0)
