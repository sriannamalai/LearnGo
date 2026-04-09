package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// ========================================
// Week 12, Lesson 2: Process Management
// ========================================
// The os/exec package runs external commands. It provides fine-grained
// control over stdin, stdout, stderr, environment variables, working
// directory, and command timeouts. This is Go's way of interacting
// with the operating system's process model.
// ========================================

func main() {
	// ========================================
	// 1. Basic Command Execution
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. Basic Command Execution")
	fmt.Println("========================================")

	// exec.Command creates a Cmd struct but doesn't run it yet.
	// The first argument is the command; the rest are its arguments.
	cmd := exec.Command("echo", "Hello from Go!")

	// Run() starts the command and waits for it to complete.
	// It returns an error if the command fails to start or exits non-zero.
	fmt.Println("\nRunning 'echo Hello from Go!':")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	// Note: Run() doesn't capture output — it goes to os.Stdout by default
	// only if Stdout is not set. Actually, by default stdout goes nowhere!
	// Let's fix that.

	// ========================================
	// 2. Capturing Output
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Capturing Output")
	fmt.Println("========================================")

	// CombinedOutput() runs the command and returns combined stdout+stderr.
	cmd = exec.Command("echo", "Hello, captured output!")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("\nCombinedOutput: %s", output)

	// Output() returns just stdout (stderr goes to os.Stderr).
	cmd = exec.Command("date")
	output, err = cmd.Output()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("date output: %s", output)

	// Using separate buffers for stdout and stderr
	fmt.Println("\nCapturing stdout and stderr separately:")
	var stdout, stderr bytes.Buffer

	// 'ls' on a valid path (stdout) vs invalid path (stderr)
	cmd = exec.Command("ls", "/tmp")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Stderr: %s\n", stderr.String())
	} else {
		lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
		fmt.Printf("ls /tmp returned %d entries (showing first 3):\n", len(lines))
		for i, line := range lines {
			if i >= 3 {
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  %s\n", line)
		}
	}

	// Demonstrate stderr capture
	stdout.Reset()
	stderr.Reset()
	cmd = exec.Command("ls", "/nonexistent_path_12345")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("\nCommand failed (expected): %v\n", err)
		fmt.Printf("Stderr: %s", stderr.String())
	}

	// ========================================
	// 3. Passing Arguments
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Passing Arguments")
	fmt.Println("========================================")

	// Arguments are passed as separate strings — NOT as a single
	// shell command string. This is safer (no shell injection).
	fmt.Println("\nUsing 'ls' with flags:")
	cmd = exec.Command("ls", "-la", "-h", "/tmp")
	output, err = cmd.Output()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		fmt.Printf("ls -la -h /tmp (%d lines, showing first 5):\n", len(lines))
		for i, line := range lines {
			if i >= 5 {
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  %s\n", line)
		}
	}

	// Using printf-style command with multiple arguments
	fmt.Println("\nUsing 'printf' with format string:")
	cmd = exec.Command("printf", "Name: %s, Age: %d\\n", "Alice", "30")
	output, _ = cmd.Output()
	fmt.Printf("printf output: %s\n", output)

	// IMPORTANT: Do NOT do this — it won't work as expected:
	//   exec.Command("ls -la")  // WRONG — treats "ls -la" as the command name
	// Instead:
	//   exec.Command("ls", "-la")  // CORRECT — "ls" is command, "-la" is argument

	// ========================================
	// 4. Setting Environment Variables
	// ========================================
	fmt.Println("========================================")
	fmt.Println("4. Setting Environment Variables")
	fmt.Println("========================================")

	// By default, child processes inherit the parent's environment.
	// You can override the entire environment by setting cmd.Env.
	cmd = exec.Command("env")
	// Set a custom environment — the child process gets ONLY these vars
	cmd.Env = []string{
		"MY_APP=learngo",
		"MY_VERSION=1.0",
		"PATH=/usr/bin:/bin",
	}
	output, _ = cmd.Output()
	envLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Println("\nCustom environment (child only sees these):")
	for _, line := range envLines {
		fmt.Printf("  %s\n", line)
	}

	// To ADD to the existing environment rather than replace it:
	cmd = exec.Command("sh", "-c", "echo MY_CUSTOM_VAR=$MY_CUSTOM_VAR")
	cmd.Env = append(os.Environ(), "MY_CUSTOM_VAR=hello_from_go")
	output, _ = cmd.Output()
	fmt.Printf("\nAdded to existing environment: %s", output)

	// ========================================
	// 5. Setting Working Directory
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Setting Working Directory")
	fmt.Println("========================================")

	// cmd.Dir sets the working directory for the child process.
	cmd = exec.Command("pwd")
	cmd.Dir = "/tmp"
	output, _ = cmd.Output()
	fmt.Printf("\nChild process working directory: %s", output)

	cmd = exec.Command("pwd")
	cmd.Dir = os.Getenv("HOME")
	output, _ = cmd.Output()
	fmt.Printf("Child process in HOME: %s", output)

	// ========================================
	// 6. Command with Shell
	// ========================================
	fmt.Println("========================================")
	fmt.Println("6. Running Shell Commands")
	fmt.Println("========================================")

	// Sometimes you need shell features like pipes, redirects, or glob expansion.
	// Use "sh -c" (or the appropriate shell) to interpret a command string.
	// WARNING: Be careful with user input — this CAN be a shell injection vector!
	fmt.Println("\nUsing shell for pipes:")
	shellCmd := "echo 'apple\nbanana\ncherry\ndate\nelderberry' | sort | head -3"
	cmd = exec.Command("sh", "-c", shellCmd)
	output, err = cmd.Output()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Shell pipe result:\n%s", output)
	}

	// ========================================
	// 7. Checking if a Command Exists
	// ========================================
	fmt.Println("========================================")
	fmt.Println("7. Checking if a Command Exists")
	fmt.Println("========================================")

	// exec.LookPath searches for a command in PATH.
	commands := []string{"go", "git", "python3", "nonexistent_cmd"}
	for _, name := range commands {
		path, err := exec.LookPath(name)
		if err != nil {
			fmt.Printf("\n  %s: NOT found in PATH\n", name)
		} else {
			fmt.Printf("\n  %s: found at %s\n", name, path)
		}
	}

	// ========================================
	// 8. Command Timeout with Context
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. Command Timeout with Context")
	fmt.Println("========================================")

	// exec.CommandContext lets you cancel a command with a context.
	// This is essential for preventing runaway processes.

	// Command that completes within timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, "sleep", "0.1")
	fmt.Println("\nRunning 'sleep 0.1' with 5s timeout...")
	start := time.Now()
	err = cmd.Run()
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("Error: %v (took %v)\n", err, elapsed)
	} else {
		fmt.Printf("Completed successfully in %v\n", elapsed)
	}

	// Command that will be killed by timeout
	ctx, cancel = context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	cmd = exec.CommandContext(ctx, "sleep", "10")
	fmt.Println("\nRunning 'sleep 10' with 500ms timeout...")
	start = time.Now()
	err = cmd.Run()
	elapsed = time.Since(start)
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Printf("Command timed out after %v (as expected)\n", elapsed)
	} else if err != nil {
		fmt.Printf("Error: %v (took %v)\n", err, elapsed)
	}

	// ========================================
	// 9. Start and Wait (Non-blocking)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("9. Start and Wait (Non-blocking)")
	fmt.Println("========================================")

	// cmd.Start() starts the command without waiting.
	// cmd.Wait() then blocks until it finishes.
	// This lets you do work while the command runs.

	cmd = exec.Command("sleep", "0.5")
	fmt.Println("\nStarting 'sleep 0.5' in background...")
	start = time.Now()
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to start: %v\n", err)
	} else {
		fmt.Printf("Command started (PID: %d)\n", cmd.Process.Pid)
		fmt.Println("Doing other work while command runs...")
		time.Sleep(100 * time.Millisecond)
		fmt.Println("...still working...")

		// Now wait for it to finish
		err = cmd.Wait()
		elapsed = time.Since(start)
		if err != nil {
			fmt.Printf("Command failed: %v\n", err)
		} else {
			fmt.Printf("Command completed in %v\n", elapsed)
		}
	}

	// ========================================
	// 10. Exit Codes
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("10. Exit Codes")
	fmt.Println("========================================")

	// When a command exits with non-zero status, Run/Wait returns
	// an *exec.ExitError. You can extract the exit code from it.
	cmd = exec.Command("sh", "-c", "exit 42")
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("\nCommand exited with code: %d\n", exitErr.ExitCode())
		} else {
			fmt.Printf("\nFailed to run command: %v\n", err)
		}
	}

	// Successful exit
	cmd = exec.Command("true")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Unexpected error: %v\n", err)
	} else {
		fmt.Println("'true' command succeeded (exit code 0)")
	}

	// ========================================
	// 11. Platform-Specific Examples
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("11. Platform-Specific Commands")
	fmt.Println("========================================")

	switch runtime.GOOS {
	case "darwin":
		fmt.Println("\nRunning macOS-specific commands:")

		// Get macOS version
		cmd = exec.Command("sw_vers", "-productVersion")
		output, err = cmd.Output()
		if err == nil {
			fmt.Printf("  macOS version: %s", output)
		}

		// Get system uptime
		cmd = exec.Command("uptime")
		output, err = cmd.Output()
		if err == nil {
			fmt.Printf("  Uptime: %s", output)
		}

	case "linux":
		fmt.Println("\nRunning Linux-specific commands:")
		cmd = exec.Command("uname", "-r")
		output, err = cmd.Output()
		if err == nil {
			fmt.Printf("  Kernel: %s", output)
		}

	default:
		fmt.Printf("\nRunning on %s — adjust commands as needed.\n", runtime.GOOS)
	}

	fmt.Println("\n========================================")
	fmt.Println("Process Management lesson complete!")
	fmt.Println("========================================")
}
