package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// ========================================
// Week 7, Lesson 1: File I/O
// ========================================
// Go provides multiple ways to read and write files, from simple
// one-shot operations (os.ReadFile/WriteFile) to fine-grained
// buffered I/O (bufio). This lesson covers all the essentials.
// ========================================

func main() {
	// ========================================
	// 1. Writing Files with os.WriteFile
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. Writing Files with os.WriteFile")
	fmt.Println("========================================")

	// os.WriteFile is the simplest way to write a file.
	// It creates the file (or truncates if it exists), writes
	// the data, and closes it — all in one call.
	//
	// File permissions: 0644 means owner can read/write,
	// group and others can read only.

	content := []byte("Hello, Go File I/O!\nThis is line two.\nAnd line three.\n")
	err := os.WriteFile("sample.txt", content, 0644)
	if err != nil {
		fmt.Printf("Error writing file: %v\n", err)
		return
	}
	fmt.Println("  Wrote sample.txt successfully")

	// ========================================
	// 2. Reading Files with os.ReadFile
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Reading Files with os.ReadFile")
	fmt.Println("========================================")

	// os.ReadFile reads the entire file into memory.
	// Good for small to medium files. For large files,
	// use buffered reading (see section 5).

	data, err := os.ReadFile("sample.txt")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	fmt.Println("  Contents of sample.txt:")
	fmt.Printf("  %s", string(data)) // data is []byte, convert to string

	// File info
	fmt.Printf("  Bytes read: %d\n", len(data))

	// ========================================
	// 3. Creating Files with os.Create
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Creating Files with os.Create")
	fmt.Println("========================================")

	// os.Create creates a new file (or truncates an existing one).
	// It returns an *os.File handle for reading and writing.
	// ALWAYS defer file.Close() right after checking the error!

	file, err := os.Create("created.txt")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close() // IMPORTANT: always close files!

	// Write using the file handle
	bytesWritten, err := file.WriteString("Created with os.Create!\n")
	if err != nil {
		fmt.Printf("Error writing: %v\n", err)
		return
	}
	fmt.Printf("  Created created.txt, wrote %d bytes\n", bytesWritten)

	// Write raw bytes
	moreBytes, err := file.Write([]byte("Second line via Write.\n"))
	if err != nil {
		fmt.Printf("Error writing bytes: %v\n", err)
		return
	}
	fmt.Printf("  Wrote %d more bytes\n", moreBytes)

	// fmt.Fprintf can write to any io.Writer, including files
	n, err := fmt.Fprintf(file, "Third line via Fprintf: %d + %d = %d\n", 2, 3, 5)
	if err != nil {
		fmt.Printf("Error with Fprintf: %v\n", err)
		return
	}
	fmt.Printf("  Fprintf wrote %d bytes\n", n)

	// ========================================
	// 4. Opening Files with os.Open
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. Opening Files with os.Open")
	fmt.Println("========================================")

	// os.Open opens a file for READING ONLY.
	// For read+write, use os.OpenFile with flags.

	readFile, err := os.Open("created.txt")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer readFile.Close()

	// Read all content using io.ReadAll
	allContent, err := io.ReadAll(readFile)
	if err != nil {
		fmt.Printf("Error reading: %v\n", err)
		return
	}
	fmt.Println("  Contents of created.txt:")
	fmt.Printf("  %s", string(allContent))

	// ========================================
	// 5. Reading Line by Line with bufio.Scanner
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Reading Line by Line (bufio.Scanner)")
	fmt.Println("========================================")

	// bufio.Scanner is the idiomatic way to read a file line by line.
	// It's memory-efficient for large files.

	// First, create a multi-line file
	lines := []string{
		"Go was designed at Google",
		"It was first released in 2009",
		"Created by Robert Griesemer, Rob Pike, and Ken Thompson",
		"Go is statically typed and compiled",
		"It has built-in concurrency support",
	}
	os.WriteFile("facts.txt", []byte(strings.Join(lines, "\n")+"\n"), 0644)

	factsFile, err := os.Open("facts.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer factsFile.Close()

	scanner := bufio.NewScanner(factsFile)
	lineNum := 0
	fmt.Println("  Reading facts.txt line by line:")
	for scanner.Scan() { // Scan() advances to next line, returns false at EOF
		lineNum++
		line := scanner.Text() // Get the current line (without newline)
		fmt.Printf("  Line %d: %s\n", lineNum, line)
	}

	// Always check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("  Scanner error: %v\n", err)
	}
	fmt.Printf("  Total lines: %d\n", lineNum)

	// ========================================
	// 6. Writing Line by Line with bufio.Writer
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Writing Line by Line (bufio.Writer)")
	fmt.Println("========================================")

	// bufio.Writer buffers writes for efficiency.
	// This is faster than writing directly when making many
	// small writes, because it reduces system calls.

	outFile, err := os.Create("buffered_output.txt")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)

	items := []string{"apple", "banana", "cherry", "date", "elderberry"}
	for i, item := range items {
		// WriteString writes a string to the buffer
		fmt.Fprintf(writer, "%d. %s\n", i+1, item)
	}

	// IMPORTANT: Flush() writes any buffered data to the underlying writer.
	// Without Flush(), some data may not be written to the file!
	err = writer.Flush()
	if err != nil {
		fmt.Printf("Error flushing: %v\n", err)
		return
	}
	fmt.Println("  Wrote buffered_output.txt with bufio.Writer")

	// Verify
	verifyData, _ := os.ReadFile("buffered_output.txt")
	fmt.Printf("  Contents:\n%s", string(verifyData))

	// ========================================
	// 7. File Permissions
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. File Permissions")
	fmt.Println("========================================")

	// Go uses Unix-style file permissions (octal notation).
	//   Owner  Group  Others
	//   rwx    rwx    rwx
	//   421    421    421
	//
	// Common permissions:
	//   0644 — owner read/write, others read only (files)
	//   0755 — owner all, others read/execute (executables)
	//   0600 — owner read/write only (private files)
	//   0700 — owner all only (private directories)

	fmt.Println("  Common file permissions:")
	fmt.Println("  0644 (rw-r--r--) — Standard file")
	fmt.Println("  0755 (rwxr-xr-x) — Executable")
	fmt.Println("  0600 (rw-------) — Private file")
	fmt.Println("  0700 (rwx------) — Private directory")

	// Create a private file
	err = os.WriteFile("private.txt", []byte("secret data\n"), 0600)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Println("  Created private.txt with 0600 permissions")
	}

	// Get file permissions
	info, err := os.Stat("private.txt")
	if err == nil {
		fmt.Printf("  Permissions: %s\n", info.Mode().Perm())
	}

	// os.OpenFile gives full control over flags and permissions
	appendFile, err := os.OpenFile("append.txt",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, // Append, create if missing, write only
		0644,
	)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Fprintf(appendFile, "Appended line at run time\n")
	appendFile.Close()
	fmt.Println("  Created/appended to append.txt")

	// ========================================
	// 8. defer file.Close()
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. defer file.Close()")
	fmt.Println("========================================")

	// The `defer` keyword schedules a function call to run when
	// the enclosing function returns. This is the idiomatic way
	// to ensure files are closed, even if an error occurs.

	fmt.Println("  Pattern for safe file handling:")
	fmt.Println("  ")
	fmt.Println("  file, err := os.Open(\"file.txt\")")
	fmt.Println("  if err != nil {")
	fmt.Println("      return err")
	fmt.Println("  }")
	fmt.Println("  defer file.Close() // Runs when function returns")
	fmt.Println("  ")
	fmt.Println("  // ... work with file ...")
	fmt.Println("  // file.Close() is called automatically")

	// Demonstrating defer order (LIFO — last in, first out)
	fmt.Println("\n  Defer execution order (LIFO):")
	deferDemo()

	// ========================================
	// 9. Checking File Existence
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("9. Checking File Existence")
	fmt.Println("========================================")

	// Go doesn't have a direct FileExists function.
	// Use os.Stat and check the error.

	filesToCheck := []string{"sample.txt", "nonexistent.txt", "created.txt"}
	for _, fname := range filesToCheck {
		exists, info := fileExists(fname)
		if exists {
			fmt.Printf("  %-20s EXISTS (size: %d bytes)\n", fname, info.Size())
		} else {
			fmt.Printf("  %-20s DOES NOT EXIST\n", fname)
		}
	}

	// ========================================
	// 10. Putting It Together: Word Counter
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("10. Word Counter Example")
	fmt.Println("========================================")

	// Count words in facts.txt using line-by-line reading
	wordCount, lineCount, err := countWordsInFile("facts.txt")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  File: facts.txt\n")
		fmt.Printf("  Lines: %d\n", lineCount)
		fmt.Printf("  Words: %d\n", wordCount)
	}

	// ========================================
	// Cleanup: Remove temporary files
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Cleanup")
	fmt.Println("========================================")

	tempFiles := []string{"sample.txt", "created.txt", "facts.txt",
		"buffered_output.txt", "private.txt", "append.txt"}
	for _, f := range tempFiles {
		err := os.Remove(f)
		if err != nil {
			fmt.Printf("  Could not remove %s: %v\n", f, err)
		} else {
			fmt.Printf("  Removed %s\n", f)
		}
	}

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println("- os.WriteFile / os.ReadFile: simple one-shot operations")
	fmt.Println("- os.Create: create/truncate file, returns *os.File")
	fmt.Println("- os.Open: open for reading only")
	fmt.Println("- os.OpenFile: full control (append, create, permissions)")
	fmt.Println("- bufio.Scanner: read line by line efficiently")
	fmt.Println("- bufio.Writer: buffered writes (call Flush()!)")
	fmt.Println("- defer file.Close(): always close files with defer")
	fmt.Println("- os.Stat + errors.Is: check file existence")
}

// ========================================
// Helper Functions
// ========================================

// deferDemo shows the LIFO execution order of defer statements.
func deferDemo() {
	defer fmt.Println("    Third defer (runs first)")
	defer fmt.Println("    Second defer (runs second)")
	defer fmt.Println("    First defer (runs third)")
	fmt.Println("    Function body executes first")
}

// fileExists checks if a file exists and returns its FileInfo.
func fileExists(path string) (bool, os.FileInfo) {
	info, err := os.Stat(path)
	if err == nil {
		return true, info
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	// Some other error (e.g., permission denied)
	return false, nil
}

// countWordsInFile reads a file line by line and counts words and lines.
func countWordsInFile(filename string) (int, int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	totalWords := 0
	totalLines := 0

	for scanner.Scan() {
		totalLines++
		line := scanner.Text()
		words := strings.Fields(line) // Split by whitespace
		totalWords += len(words)
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	return totalWords, totalLines, nil
}
