package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// ========================================
// Week 13, Lesson 1: Pipes and IPC
// ========================================
// Pipes are a fundamental inter-process communication (IPC)
// mechanism. They provide a unidirectional data channel: one end
// writes, the other reads. Go supports both OS-level pipes
// (os.Pipe) and in-process pipes (io.Pipe), plus connecting
// stdin/stdout between external commands.
// ========================================

func main() {
	// ========================================
	// 1. os.Pipe — OS-Level Pipes
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. os.Pipe — OS-Level Pipes")
	fmt.Println("========================================")

	// os.Pipe creates a connected pair of *os.File.
	// Data written to the write end can be read from the read end.
	// This uses the operating system's pipe mechanism.

	reader, writer, err := os.Pipe()
	if err != nil {
		fmt.Printf("Error creating pipe: %v\n", err)
		return
	}

	// Write to the pipe in a goroutine
	go func() {
		defer writer.Close()
		messages := []string{
			"Hello from the writer!",
			"Pipes are awesome!",
			"This is the last message.",
		}
		for _, msg := range messages {
			fmt.Fprintf(writer, "%s\n", msg)
		}
	}()

	// Read from the pipe
	fmt.Println("\nReading from os.Pipe:")
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Printf("  Received: %s\n", scanner.Text())
	}
	reader.Close()

	if err := scanner.Err(); err != nil {
		fmt.Printf("Read error: %v\n", err)
	}

	// ========================================
	// 2. Connecting Command Stdin/Stdout
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Connecting Command Stdin/Stdout")
	fmt.Println("========================================")

	// You can feed data into a command's stdin using cmd.Stdin.
	// This is like piping data to a command in the shell:
	//   echo "data" | sort

	fmt.Println("\nFeeding data to 'sort' command via stdin:")
	cmd := exec.Command("sort")
	cmd.Stdin = strings.NewReader("banana\napple\ncherry\ndate\nelderberry\n")

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Sorted output:\n%s", out.String())
	}

	// Writing to stdin using StdinPipe
	fmt.Println("\nUsing StdinPipe for interactive input:")
	cmd = exec.Command("cat")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("Error getting stdin pipe: %v\n", err)
		return
	}

	var catOutput bytes.Buffer
	cmd.Stdout = &catOutput

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		return
	}

	// Write data to the command's stdin
	lines := []string{"Line 1: Hello", "Line 2: World", "Line 3: Done"}
	for _, line := range lines {
		fmt.Fprintf(stdin, "%s\n", line)
	}
	stdin.Close() // Close stdin to signal EOF to the command

	err = cmd.Wait()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("cat echoed back:\n%s", catOutput.String())
	}

	// ========================================
	// 3. Piping Between Commands
	// ========================================
	fmt.Println("========================================")
	fmt.Println("3. Piping Between Commands")
	fmt.Println("========================================")

	// This is the equivalent of shell pipes:
	//   echo "hello\nworld" | grep "world" | tr 'a-z' 'A-Z'
	// We connect one command's stdout to the next command's stdin.

	fmt.Println("\nPiping: echo | grep | tr (shell equivalent)")
	pipeCommands()

	// A simpler approach: pipe two commands together
	fmt.Println("\nSimpler example: ls | wc -l (count files):")
	pipeTwoCommands()

	// ========================================
	// 4. io.Pipe — In-Process Pipes
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. io.Pipe — In-Process Pipes")
	fmt.Println("========================================")

	// io.Pipe creates a synchronous, in-memory pipe.
	// Unlike os.Pipe, this doesn't use file descriptors.
	// Writes block until reads consume the data (no buffering).
	// This is useful for connecting io.Writer producers to
	// io.Reader consumers within the same process.

	fmt.Println("\nUsing io.Pipe for in-process communication:")
	ioPipeDemo()

	// ========================================
	// 5. io.Pipe with Commands
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. io.Pipe with Commands")
	fmt.Println("========================================")

	// io.Pipe can bridge in-process data with external commands.
	// Here we generate data in Go, pipe it through an external
	// command (sort), and read the result back in Go.

	fmt.Println("\nGenerating data in Go -> piping to 'sort -r' -> reading result:")
	ioPipeWithCommand()

	// ========================================
	// 6. Bidirectional Communication
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Bidirectional Communication")
	fmt.Println("========================================")

	// For bidirectional communication with a process, use both
	// StdinPipe and StdoutPipe. This is like having a conversation
	// with the process.

	fmt.Println("\nBidirectional communication with 'cat' (echo server):")
	bidirectionalDemo()

	// ========================================
	// 7. Multi-Stage Pipeline
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. Multi-Stage Pipeline")
	fmt.Println("========================================")

	// Build a pipeline with multiple stages, all running concurrently.
	// This demonstrates Go's strength in orchestrating processes.

	fmt.Println("\nMulti-stage pipeline: generate -> sort -> unique -> count:")
	multiStagePipeline()

	// ========================================
	// 8. os.Pipe vs io.Pipe Comparison
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. os.Pipe vs io.Pipe Comparison")
	fmt.Println("========================================")

	fmt.Print(`
  os.Pipe():
    - Creates OS-level file descriptors
    - Has kernel-level buffering (~64KB on Linux/macOS)
    - Can be passed to child processes (cmd.ExtraFiles)
    - Returns *os.File (can be used with any file operation)
    - Slightly higher overhead (system calls)

  io.Pipe():
    - Pure Go, in-memory, synchronous
    - No buffering — writes block until read
    - Cannot be passed to child processes
    - Returns io.PipeReader/io.PipeWriter
    - Lower overhead, good for in-process streaming
    - Supports CloseWithError for error propagation

  When to use which:
    - os.Pipe: connecting to external processes, need buffering
    - io.Pipe: in-process streaming, connecting Go components
`)
	fmt.Println()

	fmt.Println("========================================")
	fmt.Println("Pipes and IPC lesson complete!")
	fmt.Println("========================================")
}

// ========================================
// Helper Functions
// ========================================

// pipeCommands demonstrates piping between multiple commands.
func pipeCommands() {
	// Command 1: echo some data
	cmd1 := exec.Command("echo", "hello\nworld\nhello\ngo\nworld")

	// Command 2: grep for "world"
	cmd2 := exec.Command("grep", "world")

	// Command 3: convert to uppercase
	cmd3 := exec.Command("tr", "a-z", "A-Z")

	// Connect cmd1's stdout to cmd2's stdin
	cmd2Stdin, err := cmd2.StdinPipe()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	// Connect cmd2's stdout to cmd3's stdin
	cmd3Stdin, err := cmd3.StdinPipe()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	// Capture cmd1 output
	var cmd1Out bytes.Buffer
	cmd1.Stdout = &cmd1Out

	// Pipe cmd2 output
	var cmd2Out bytes.Buffer
	cmd2.Stdout = &cmd2Out

	// Capture final output from cmd3
	var finalOutput bytes.Buffer
	cmd3.Stdout = &finalOutput

	// Start all commands
	cmd1.Run()

	// Feed cmd1 output to cmd2
	cmd2.Start()
	cmd2Stdin.Write(cmd1Out.Bytes())
	cmd2Stdin.Close()
	cmd2.Wait()

	// Feed cmd2 output to cmd3
	cmd3.Start()
	cmd3Stdin.Write(cmd2Out.Bytes())
	cmd3Stdin.Close()
	cmd3.Wait()

	fmt.Printf("  Result: %s", finalOutput.String())
}

// pipeTwoCommands pipes stdout of one command to stdin of another.
func pipeTwoCommands() {
	// Using shell for simplicity (equivalent of: ls /tmp | wc -l)
	cmd := exec.Command("sh", "-c", "ls /tmp | wc -l")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}
	fmt.Printf("  Files in /tmp: %s", output)

	// The pure Go approach using StdoutPipe:
	cmd1 := exec.Command("ls", "/tmp")
	cmd2 := exec.Command("wc", "-l")

	// Get cmd1's stdout as a pipe
	pipe, err := cmd1.StdoutPipe()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	// Set cmd2's stdin to cmd1's stdout pipe
	cmd2.Stdin = pipe

	// Capture cmd2's output
	var out bytes.Buffer
	cmd2.Stdout = &out

	// Start cmd1, then cmd2, then wait for both
	cmd1.Start()
	cmd2.Start()
	cmd1.Wait()
	cmd2.Wait()

	fmt.Printf("  Files in /tmp (pure Go pipe): %s", out.String())
}

// ioPipeDemo demonstrates in-process pipe communication.
func ioPipeDemo() {
	pr, pw := io.Pipe()

	// Writer goroutine: produces data
	go func() {
		defer pw.Close()
		for i := 1; i <= 5; i++ {
			msg := fmt.Sprintf("Message %d: %s\n", i, strings.Repeat("*", i))
			_, err := pw.Write([]byte(msg))
			if err != nil {
				fmt.Printf("  Write error: %v\n", err)
				return
			}
		}
	}()

	// Reader: consumes data
	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		fmt.Printf("  io.Pipe received: %s\n", scanner.Text())
	}
	pr.Close()
}

// ioPipeWithCommand uses io.Pipe to bridge Go code and an external command.
func ioPipeWithCommand() {
	pr, pw := io.Pipe()

	// Generate data in a goroutine
	go func() {
		defer pw.Close()
		fruits := []string{"banana", "apple", "cherry", "date", "elderberry", "fig", "grape"}
		for _, f := range fruits {
			fmt.Fprintln(pw, f)
		}
	}()

	// Sort in reverse order using the external 'sort' command
	cmd := exec.Command("sort", "-r")
	cmd.Stdin = pr

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	fmt.Printf("  Sorted (reverse):\n")
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fmt.Printf("    %s\n", line)
	}
}

// bidirectionalDemo shows two-way communication with a process.
func bidirectionalDemo() {
	// 'cat' echoes whatever it receives — a simple echo server
	cmd := exec.Command("cat")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	err = cmd.Start()
	if err != nil {
		fmt.Printf("  Error starting: %v\n", err)
		return
	}

	// Read responses in a goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Printf("  <- Received: %s\n", scanner.Text())
		}
	}()

	// Send messages
	messages := []string{"ping", "hello", "goodbye"}
	for _, msg := range messages {
		fmt.Printf("  -> Sending: %s\n", msg)
		fmt.Fprintln(stdin, msg)
	}

	// Close stdin to signal EOF
	stdin.Close()

	// Wait for all output to be read
	<-done
	cmd.Wait()
}

// multiStagePipeline builds a multi-stage data processing pipeline.
func multiStagePipeline() {
	// Stage 1: Generate data (in Go)
	// Stage 2: Sort (external command)
	// Stage 3: Unique (external command)
	// Stage 4: Count lines (external command)

	// Generate data with duplicates
	data := "banana\napple\ncherry\napple\nbanana\ndate\ncherry\napple\nfig\ndate\n"

	fmt.Printf("  Input data (%d lines, with duplicates):\n", strings.Count(data, "\n"))
	for _, line := range strings.Split(strings.TrimSpace(data), "\n") {
		fmt.Printf("    %s\n", line)
	}

	// Build pipeline using shell (practical approach)
	cmd := exec.Command("sh", "-c", "sort | uniq -c | sort -rn")
	cmd.Stdin = strings.NewReader(data)

	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	fmt.Printf("\n  Pipeline result (count + unique, sorted by frequency):\n")
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fmt.Printf("    %s\n", strings.TrimSpace(line))
	}
}
