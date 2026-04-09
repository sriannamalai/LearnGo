package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ========================================
// Week 12, Lesson 1: OS Package Basics
// ========================================
// The os package provides a platform-independent interface to
// operating system functionality. It covers environment variables,
// file information, permissions, working directories, temporary
// files, and command-line arguments.
// ========================================

func main() {
	// ========================================
	// 1. Command-Line Arguments (os.Args)
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. Command-Line Arguments (os.Args)")
	fmt.Println("========================================")

	// os.Args is a slice of strings containing the command and its arguments.
	// os.Args[0] is always the program name (path to the executable).
	fmt.Printf("\nProgram name: %s\n", os.Args[0])
	fmt.Printf("Number of arguments: %d\n", len(os.Args))

	if len(os.Args) > 1 {
		fmt.Println("Arguments provided:")
		for i, arg := range os.Args[1:] {
			fmt.Printf("  args[%d] = %q\n", i+1, arg)
		}
	} else {
		fmt.Println("No additional arguments provided.")
		fmt.Println("Try: go run main.go hello world --verbose")
	}

	// ========================================
	// 2. Environment Variables
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Environment Variables")
	fmt.Println("========================================")

	// os.Getenv reads an environment variable. Returns "" if not set.
	home := os.Getenv("HOME")
	fmt.Printf("\nHOME = %q\n", home)

	user := os.Getenv("USER")
	fmt.Printf("USER = %q\n", user)

	path := os.Getenv("PATH")
	fmt.Printf("PATH (first 80 chars) = %q...\n", path[:min(80, len(path))])

	// os.LookupEnv distinguishes between "not set" and "set to empty".
	// This is important when the absence of a variable has different
	// meaning than an empty value.
	val, exists := os.LookupEnv("HOME")
	if exists {
		fmt.Printf("\nHOME exists with value: %q\n", val)
	} else {
		fmt.Println("\nHOME is not set")
	}

	val, exists = os.LookupEnv("NONEXISTENT_VAR_12345")
	if exists {
		fmt.Printf("NONEXISTENT_VAR_12345 = %q\n", val)
	} else {
		fmt.Println("NONEXISTENT_VAR_12345 is not set (as expected)")
	}

	// os.Setenv sets an environment variable for the current process.
	// This does NOT affect the parent shell or other processes.
	fmt.Println("\nSetting MY_APP_MODE=production...")
	err := os.Setenv("MY_APP_MODE", "production")
	if err != nil {
		fmt.Printf("Error setting env var: %v\n", err)
	}
	fmt.Printf("MY_APP_MODE = %q\n", os.Getenv("MY_APP_MODE"))

	// os.Unsetenv removes an environment variable.
	os.Unsetenv("MY_APP_MODE")
	fmt.Printf("After Unsetenv, MY_APP_MODE = %q\n", os.Getenv("MY_APP_MODE"))

	// os.Environ returns all environment variables as a slice of "KEY=VALUE" strings.
	allEnv := os.Environ()
	fmt.Printf("\nTotal environment variables: %d\n", len(allEnv))
	fmt.Println("First 5 environment variables:")
	for i, env := range allEnv {
		if i >= 5 {
			break
		}
		// Truncate long values for display
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && len(parts[1]) > 50 {
			fmt.Printf("  %s=%s...\n", parts[0], parts[1][:50])
		} else {
			fmt.Printf("  %s\n", env)
		}
	}

	// os.ExpandEnv replaces ${var} or $var in a string.
	template := "Hello, $USER! Your home is $HOME."
	expanded := os.ExpandEnv(template)
	fmt.Printf("\nTemplate: %s\n", template)
	fmt.Printf("Expanded: %s\n", expanded)

	// ========================================
	// 3. Working Directory
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. Working Directory")
	fmt.Println("========================================")

	// os.Getwd returns the current working directory.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
	} else {
		fmt.Printf("\nCurrent working directory: %s\n", cwd)
	}

	// os.Chdir changes the working directory.
	// Be careful — this affects the entire process!
	tmpDir := os.TempDir()
	fmt.Printf("Changing to temp directory: %s\n", tmpDir)
	err = os.Chdir(tmpDir)
	if err != nil {
		fmt.Printf("Error changing directory: %v\n", err)
	} else {
		newCwd, _ := os.Getwd()
		fmt.Printf("Now in: %s\n", newCwd)
	}

	// Change back to the original directory
	os.Chdir(cwd)
	fmt.Printf("Changed back to: %s\n", cwd)

	// ========================================
	// 4. File Information (os.Stat)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. File Information (os.Stat)")
	fmt.Println("========================================")

	// os.Stat returns a FileInfo describing the named file.
	// It follows symlinks. Use os.Lstat for symlink info itself.
	info, err := os.Stat("main.go")
	if err != nil {
		fmt.Printf("\nError stating main.go: %v\n", err)
		fmt.Println("(Run this from the 01_os_basics directory)")
	} else {
		fmt.Println("\nFile info for main.go:")
		fmt.Printf("  Name:    %s\n", info.Name())
		fmt.Printf("  Size:    %d bytes\n", info.Size())
		fmt.Printf("  Mode:    %s\n", info.Mode())
		fmt.Printf("  ModTime: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		fmt.Printf("  IsDir:   %v\n", info.IsDir())
	}

	// Check if a file exists using os.Stat
	_, err = os.Stat("/nonexistent/file/path")
	if os.IsNotExist(err) {
		fmt.Println("\n/nonexistent/file/path does not exist (os.IsNotExist)")
	}

	// Check if a path is a directory
	info, err = os.Stat(".")
	if err == nil {
		fmt.Printf("\n'.' is a directory: %v\n", info.IsDir())
	}

	// ========================================
	// 5. File Permissions
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. File Permissions")
	fmt.Println("========================================")

	// File permissions in Go use os.FileMode, which maps to Unix permission bits.
	// Common permission values:
	//   0644 — owner read/write, group/others read only (typical for files)
	//   0755 — owner read/write/execute, group/others read/execute (typical for dirs)
	//   0600 — owner read/write only (private files, like SSH keys)
	//   0700 — owner read/write/execute only (private directories)

	fmt.Println("\nCommon permission modes:")
	modes := []os.FileMode{0644, 0755, 0600, 0700, 0666, 0777}
	for _, mode := range modes {
		fmt.Printf("  %04o = %s\n", mode, mode)
	}

	// Create a file with specific permissions
	testFile := filepath.Join(os.TempDir(), "go_os_test.txt")
	err = os.WriteFile(testFile, []byte("test content\n"), 0644)
	if err != nil {
		fmt.Printf("\nError creating test file: %v\n", err)
	} else {
		info, _ := os.Stat(testFile)
		fmt.Printf("\nCreated %s with permissions: %s\n", testFile, info.Mode())

		// os.Chmod changes file permissions
		os.Chmod(testFile, 0600)
		info, _ = os.Stat(testFile)
		fmt.Printf("After Chmod(0600): %s\n", info.Mode())

		// Clean up
		os.Remove(testFile)
	}

	// ========================================
	// 6. Temporary Files and Directories
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("6. Temporary Files and Directories")
	fmt.Println("========================================")

	// os.TempDir returns the default directory for temporary files.
	fmt.Printf("\nTemp directory: %s\n", os.TempDir())

	// os.CreateTemp creates a temporary file.
	// The pattern "*" is replaced with a random string.
	tmpFile, err := os.CreateTemp("", "myapp-*.txt")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
	} else {
		fmt.Printf("Created temp file: %s\n", tmpFile.Name())
		tmpFile.WriteString("temporary data\n")
		tmpFile.Close()
		// Always clean up temp files when done!
		os.Remove(tmpFile.Name())
		fmt.Println("Temp file removed after use.")
	}

	// os.MkdirTemp creates a temporary directory.
	tmpDirNew, err := os.MkdirTemp("", "myapp-*")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
	} else {
		fmt.Printf("Created temp dir: %s\n", tmpDirNew)
		// Create a file inside the temp directory
		testPath := filepath.Join(tmpDirNew, "data.txt")
		os.WriteFile(testPath, []byte("hello"), 0644)
		fmt.Printf("Created file inside temp dir: %s\n", testPath)
		// Clean up the entire temp directory
		os.RemoveAll(tmpDirNew)
		fmt.Println("Temp directory removed after use.")
	}

	// ========================================
	// 7. Platform Information
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. Platform Information")
	fmt.Println("========================================")

	fmt.Printf("\nOperating System: %s\n", runtime.GOOS)
	fmt.Printf("Architecture:     %s\n", runtime.GOARCH)
	fmt.Printf("Num CPUs:         %d\n", runtime.NumCPU())
	fmt.Printf("Go Version:       %s\n", runtime.Version())

	// os.Hostname returns the host name reported by the kernel.
	hostname, err := os.Hostname()
	if err == nil {
		fmt.Printf("Hostname:         %s\n", hostname)
	}

	// os.Getpid and os.Getppid return process IDs.
	fmt.Printf("Process ID (PID): %d\n", os.Getpid())
	fmt.Printf("Parent PID:       %d\n", os.Getppid())

	// os.Getuid, os.Getgid return user/group IDs (Unix only).
	if runtime.GOOS != "windows" {
		fmt.Printf("User ID (UID):    %d\n", os.Getuid())
		fmt.Printf("Group ID (GID):   %d\n", os.Getgid())
	}

	// os.UserHomeDir returns the current user's home directory.
	homeDir, err := os.UserHomeDir()
	if err == nil {
		fmt.Printf("Home directory:   %s\n", homeDir)
	}

	// os.UserCacheDir returns the default root directory for user-specific cached data.
	cacheDir, err := os.UserCacheDir()
	if err == nil {
		fmt.Printf("Cache directory:  %s\n", cacheDir)
	}

	// os.UserConfigDir returns the default root directory for user-specific config data.
	configDir, err := os.UserConfigDir()
	if err == nil {
		fmt.Printf("Config directory: %s\n", configDir)
	}

	// ========================================
	// 8. os.Exit and Error Handling
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. os.Exit and Error Handling")
	fmt.Println("========================================")

	// os.Exit terminates the program immediately with a status code.
	// Status 0 means success; non-zero means error.
	// WARNING: os.Exit does NOT run deferred functions!
	// It should only be used in main() or for fatal errors.

	fmt.Println("\nos.Exit(0) — success (we won't call it here)")
	fmt.Println("os.Exit(1) — general error")
	fmt.Println("os.Exit(2) — misuse of command")
	fmt.Println()
	fmt.Println("NOTE: os.Exit skips all deferred functions!")
	fmt.Println("Prefer returning errors from functions instead.")
	fmt.Println()
	fmt.Println("For graceful exits, use log.Fatal() which calls")
	fmt.Println("os.Exit(1) after logging, or simply return from main().")

	fmt.Println("\n========================================")
	fmt.Println("OS Basics lesson complete!")
	fmt.Println("========================================")
}
