package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// ========================================
// Week 13, Lesson 4 (Mini-Project): Container Isolation Demo
// ========================================
// This lesson demonstrates the CONCEPTS behind Linux container
// isolation (namespaces, cgroups, chroot). Containers like Docker
// use these kernel features to isolate processes.
//
// Full namespace isolation requires Linux. This program:
//   - On Linux: demonstrates real namespace isolation
//   - On macOS: explains the concepts and shows what's possible
//     with process isolation via exec.Command
//
// Usage:
//   go run main.go              # Show container concepts
//   go run main.go isolate      # Run isolated process demo
//   go run main.go filesystem   # Demonstrate filesystem isolation
//
// NOTE: Some features require root privileges on Linux.
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("Container Isolation Demo")
	fmt.Println("========================================")
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()

	if len(os.Args) < 2 {
		showContainerConcepts()
		demonstrateProcessIsolation()
		demonstrateFilesystemIsolation()
		demonstrateEnvironmentIsolation()
		showNamespaceReference()
		return
	}

	switch os.Args[1] {
	case "concepts":
		showContainerConcepts()
	case "isolate":
		demonstrateProcessIsolation()
	case "filesystem":
		demonstrateFilesystemIsolation()
	case "environment":
		demonstrateEnvironmentIsolation()
	case "namespaces":
		showNamespaceReference()
	case "child":
		// Internal: this is called by the isolation demo
		runAsIsolatedChild()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Commands: concepts, isolate, filesystem, environment, namespaces")
	}
}

// ========================================
// 1. Container Concepts
// ========================================

func showContainerConcepts() {
	fmt.Println("========================================")
	fmt.Println("1. What Makes a Container?")
	fmt.Println("========================================")

	fmt.Print(`
  A container is NOT a virtual machine. It's a regular process that
  runs with restricted views of the system, achieved through:

  1. NAMESPACES — What a process can SEE
     Namespaces partition kernel resources so that one set of processes
     sees one set of resources, and another set sees different resources.

  2. CGROUPS (Control Groups) — What a process can USE
     Cgroups limit, account for, and isolate resource usage (CPU, memory,
     disk I/O, network) of process groups.

  3. CHROOT / PIVOT_ROOT — Filesystem isolation
     Changes the apparent root directory for a process, restricting
     filesystem access to a subtree.

  4. CAPABILITIES — What a process can DO
     Linux capabilities break root privileges into smaller units that
     can be independently enabled/disabled.

  5. SECCOMP — System call filtering
     Restricts which system calls a process can make.

  Together, these create the illusion of a separate machine while
  sharing the host kernel.
`)

	fmt.Println("========================================")
	fmt.Println("Linux Namespace Types")
	fmt.Println("========================================")

	namespaces := []struct {
		name     string
		flag     string
		isolates string
	}{
		{"UTS", "CLONE_NEWUTS", "Hostname and domain name. Each container can have its own hostname."},
		{"PID", "CLONE_NEWPID", "Process IDs. Container's init process is PID 1 inside, different PID outside."},
		{"Mount", "CLONE_NEWNS", "Mount points. Container has its own filesystem mounts."},
		{"Network", "CLONE_NEWNET", "Network stack. Container gets its own interfaces, IPs, routing tables."},
		{"User", "CLONE_NEWUSER", "User/group IDs. Root inside container maps to non-root outside."},
		{"IPC", "CLONE_NEWIPC", "Inter-process communication. Separate message queues, semaphores."},
		{"Cgroup", "CLONE_NEWCGROUP", "Cgroup root directory. Hides host cgroup hierarchy."},
	}

	for _, ns := range namespaces {
		fmt.Printf("\n  %s Namespace (%s):\n", ns.name, ns.flag)
		fmt.Printf("    %s\n", ns.isolates)
	}
	fmt.Println()
}

// ========================================
// 2. Process Isolation
// ========================================

func demonstrateProcessIsolation() {
	fmt.Println("========================================")
	fmt.Println("2. Process Isolation Demo")
	fmt.Println("========================================")

	// We demonstrate isolation by running a child process with
	// restricted environment, modified attributes, and limited visibility.

	fmt.Println("\n--- Host Process Info ---")
	fmt.Printf("  PID:      %d\n", os.Getpid())
	fmt.Printf("  PPID:     %d\n", os.Getppid())

	hostname, _ := os.Hostname()
	fmt.Printf("  Hostname: %s\n", hostname)

	if runtime.GOOS != "windows" {
		fmt.Printf("  UID:      %d\n", os.Getuid())
		fmt.Printf("  GID:      %d\n", os.Getgid())
	}

	cwd, _ := os.Getwd()
	fmt.Printf("  CWD:      %s\n", cwd)

	// Run a child process with isolation characteristics
	fmt.Println("\n--- Spawning Isolated Child Process ---")
	fmt.Println("  (Simulating container isolation with exec.Command)")

	// Re-run ourselves with the "child" argument
	cmd := exec.Command(os.Args[0], "child")

	// Isolation technique 1: Custom environment
	// In a real container, the environment is completely separate
	cmd.Env = []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"HOME=/root",
		"USER=container",
		"CONTAINER_ID=demo-001",
		"HOSTNAME=container-host",
	}

	// Isolation technique 2: Different working directory
	tmpDir, err := os.MkdirTemp("", "container-root-*")
	if err != nil {
		fmt.Printf("  Error creating temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tmpDir)
	cmd.Dir = tmpDir

	// Isolation technique 3: Separate stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Isolation technique 4: Platform-specific process attributes
	configureSysProcAttr(cmd)

	err = cmd.Run()
	if err != nil {
		// The child may fail if it can't find itself — that's OK for the demo
		fmt.Printf("  Child process exited: %v\n", err)
		fmt.Println("  (This is expected — the child runs in a restricted environment)")
	}

	fmt.Println("\n--- Back in Host Process ---")
	fmt.Printf("  Host PID is still: %d\n", os.Getpid())
}

// configureSysProcAttr sets platform-specific process attributes.
// On macOS (and non-Linux platforms), Linux namespace flags like
// CLONE_NEWUTS are not available. On Linux, you would set:
//
//	cmd.SysProcAttr = &syscall.SysProcAttr{
//	    Cloneflags: syscall.CLONE_NEWUTS |   // New hostname namespace
//	                syscall.CLONE_NEWPID |   // New PID namespace
//	                syscall.CLONE_NEWNS |    // New mount namespace
//	                syscall.CLONE_NEWUSER,   // New user namespace
//	}
//
// These flags create new namespaces for the child process, providing
// real container-level isolation.
func configureSysProcAttr(cmd *exec.Cmd) {
	// SysProcAttr allows setting OS-specific process attributes.
	// On macOS, we have limited process isolation options.
	// On Linux, this is where you'd set namespace clone flags.

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group (available on both macOS and Linux)
	}

	switch runtime.GOOS {
	case "linux":
		fmt.Println("  [Linux] Process group isolation set.")
		fmt.Println("  For full namespace isolation (UTS, PID, mount, network),")
		fmt.Println("  set Cloneflags in SysProcAttr (requires root for most).")
	case "darwin":
		fmt.Println("  [macOS] Limited isolation (no namespace support)")
		fmt.Println("  macOS uses different isolation: sandbox-exec, App Sandbox")
	default:
		fmt.Printf("  [%s] Platform-specific isolation not configured\n", runtime.GOOS)
	}
}

// runAsIsolatedChild is called when the program re-execs itself as a child.
func runAsIsolatedChild() {
	fmt.Println("\n  --- Inside Isolated Child ---")
	fmt.Printf("  Child PID:      %d\n", os.Getpid())
	fmt.Printf("  Child PPID:     %d\n", os.Getppid())
	fmt.Printf("  Container ID:   %s\n", os.Getenv("CONTAINER_ID"))
	fmt.Printf("  Hostname env:   %s\n", os.Getenv("HOSTNAME"))
	fmt.Printf("  User env:       %s\n", os.Getenv("USER"))
	fmt.Printf("  HOME:           %s\n", os.Getenv("HOME"))

	cwd, _ := os.Getwd()
	fmt.Printf("  CWD:            %s\n", cwd)

	// Show that we have a restricted environment
	fmt.Printf("  Environment variables: %d\n", len(os.Environ()))
	for _, e := range os.Environ() {
		fmt.Printf("    %s\n", e)
	}

	// On Linux with UTS namespace, we could change hostname:
	if runtime.GOOS == "linux" {
		// syscall.Sethostname([]byte("container-host"))
		// hostname, _ := os.Hostname()
		fmt.Println("  (On Linux with CLONE_NEWUTS, hostname is isolated)")
	}

	fmt.Println("  --- End Isolated Child ---")
}

// ========================================
// 3. Filesystem Isolation
// ========================================

func demonstrateFilesystemIsolation() {
	fmt.Println("\n========================================")
	fmt.Println("3. Filesystem Isolation Demo")
	fmt.Println("========================================")

	// Create a "container rootfs" — a minimal filesystem
	rootfs, err := os.MkdirTemp("", "container-fs-*")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer os.RemoveAll(rootfs)

	fmt.Printf("\nContainer rootfs: %s\n", rootfs)

	// Build a minimal filesystem structure
	dirs := []string{
		"bin", "etc", "home", "tmp", "var/log", "proc",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(rootfs, d), 0755)
	}

	// Create some "container" files
	files := map[string]string{
		"etc/hostname":    "container-host\n",
		"etc/hosts":       "127.0.0.1 localhost\n127.0.0.1 container-host\n",
		"etc/resolv.conf": "nameserver 8.8.8.8\n",
		"home/readme.txt": "Welcome to the container!\n",
		"var/log/app.log": "Container started.\n",
	}
	for path, content := range files {
		os.WriteFile(filepath.Join(rootfs, path), []byte(content), 0644)
	}

	fmt.Println("\nContainer filesystem structure:")
	printDirTree(rootfs, "", true)

	// Show how chroot would work (requires root)
	fmt.Println("\nFilesystem isolation concepts:")
	fmt.Print(`
  chroot (change root):
    - Changes the apparent root directory to our rootfs
    - Process sees rootfs as "/" — can't access host files above it
    - Used by: chroot syscall, or syscall.Chroot() in Go
    - Limitation: can be escaped by a root process (see pivot_root)

  pivot_root (Linux only):
    - More secure than chroot — actually changes the root mount
    - Used by real container runtimes (Docker, containerd)
    - Old root can be unmounted completely

  Overlay filesystem:
    - Container images use layered filesystems
    - Read-only base layers + read-write top layer
    - Efficient sharing: multiple containers share base layers

  Example (requires root on Linux):
    syscall.Chroot(rootfs)
    os.Chdir("/")
    // Now "/" is the rootfs — host filesystem is invisible
`)

	// Demonstrate reading from our "container filesystem"
	fmt.Println("Reading from container filesystem:")
	for _, path := range []string{"etc/hostname", "etc/hosts", "home/readme.txt"} {
		content, err := os.ReadFile(filepath.Join(rootfs, path))
		if err != nil {
			continue
		}
		fmt.Printf("  /%s: %s", path, content)
	}
}

// ========================================
// 4. Environment Isolation
// ========================================

func demonstrateEnvironmentIsolation() {
	fmt.Println("\n========================================")
	fmt.Println("4. Environment Isolation Demo")
	fmt.Println("========================================")

	fmt.Println("\nDemonstrating how containers isolate the environment:")

	// Show host environment summary
	hostEnv := os.Environ()
	fmt.Printf("\n  Host environment: %d variables\n", len(hostEnv))

	// Create "container" environments
	containers := []struct {
		name string
		env  []string
		cmd  string
	}{
		{
			name: "web-server",
			env: []string{
				"PATH=/usr/local/bin:/usr/bin:/bin",
				"HOME=/app",
				"USER=www",
				"APP_PORT=8080",
				"APP_ENV=production",
				"DATABASE_URL=postgres://db:5432/myapp",
			},
			cmd: "echo 'Web server container: PORT=$APP_PORT ENV=$APP_ENV'",
		},
		{
			name: "worker",
			env: []string{
				"PATH=/usr/local/bin:/usr/bin:/bin",
				"HOME=/worker",
				"USER=worker",
				"QUEUE_URL=redis://redis:6379",
				"CONCURRENCY=4",
			},
			cmd: "echo 'Worker container: QUEUE=$QUEUE_URL CONCURRENCY=$CONCURRENCY'",
		},
		{
			name: "database",
			env: []string{
				"PATH=/usr/local/bin:/usr/bin:/bin",
				"HOME=/var/lib/db",
				"USER=postgres",
				"PGDATA=/var/lib/db/data",
				"POSTGRES_PASSWORD=secret123",
			},
			cmd: "echo 'Database container: USER=$USER PGDATA=$PGDATA'",
		},
	}

	for _, c := range containers {
		fmt.Printf("\n  --- Container: %s ---\n", c.name)
		fmt.Printf("  Environment (%d vars):\n", len(c.env))
		for _, e := range c.env {
			fmt.Printf("    %s\n", e)
		}

		// Run the command in the isolated environment
		cmd := exec.Command("sh", "-c", c.cmd)
		cmd.Env = c.env
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("  Error: %v\n", err)
		} else {
			fmt.Printf("  Output: %s", output)
		}
	}

	// Demonstrate that containers can't see each other's env
	fmt.Println("\n  Key point: Each container has its OWN environment.")
	fmt.Println("  Container A can't see Container B's DATABASE_URL.")
	fmt.Println("  Container C can't see Container A's APP_PORT.")
	fmt.Println("  This is environment isolation in action.")
}

// ========================================
// 5. Namespace Reference
// ========================================

func showNamespaceReference() {
	fmt.Println("\n========================================")
	fmt.Println("5. Linux Namespace Deep Dive")
	fmt.Println("========================================")

	fmt.Print(`
  === UTS Namespace (CLONE_NEWUTS) ===
  Isolates: hostname and NIS domain name
  Use: Each container has its own hostname
  Code:
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS,
    }
    // Inside child:
    syscall.Sethostname([]byte("my-container"))

  === PID Namespace (CLONE_NEWPID) ===
  Isolates: Process ID number space
  Use: Container sees its init as PID 1
  Effect:
    - Host sees container process as PID 12345
    - Container sees itself as PID 1
    - Container can't see or signal host processes
    - PID 1 in the container must reap zombie processes

  === Mount Namespace (CLONE_NEWNS) ===
  Isolates: Mount points
  Use: Container has its own filesystem view
  Effect:
    - Container can mount/unmount without affecting host
    - Combined with chroot/pivot_root for full FS isolation
    - Enables layered filesystems (overlayfs)

  === Network Namespace (CLONE_NEWNET) ===
  Isolates: Network stack (interfaces, IPs, routes, firewall rules)
  Use: Each container gets its own network interface
  Effect:
    - Container has its own lo (loopback) interface
    - Virtual ethernet pairs (veth) connect container to host
    - Container can bind to port 80 without conflicting with host
    - Docker bridge network connects containers together

  === User Namespace (CLONE_NEWUSER) ===
  Isolates: User and group IDs
  Use: Root inside container is non-root outside
  Effect:
    - Process can be UID 0 (root) inside the namespace
    - Maps to an unprivileged UID outside (e.g., UID 100000)
    - Enables rootless containers (no actual root needed)
    - Most secure namespace — doesn't require root to create

  === IPC Namespace (CLONE_NEWIPC) ===
  Isolates: System V IPC, POSIX message queues
  Use: Prevents containers from accessing each other's shared memory
  Effect:
    - Separate set of IPC identifiers
    - Container can't read another container's shared memory
    - Important for multi-tenant environments

  === Cgroup Namespace (CLONE_NEWCGROUP) ===
  Isolates: Cgroup root directory view
  Use: Container sees its cgroup as root of the hierarchy
  Effect:
    - Container can't see host cgroup structure
    - Prevents container from modifying its own resource limits
    - Added in Linux 4.6

  === Building a Container (Simplified) ===
  A container runtime essentially does this:

  1. Create namespaces (clone with flags)
  2. Set up cgroups (resource limits)
  3. Set up rootfs (overlay filesystem)
  4. pivot_root to new rootfs
  5. Mount /proc, /sys, /dev
  6. Set capabilities (drop unnecessary privileges)
  7. Apply seccomp filter (restrict syscalls)
  8. Execute the container's entrypoint
`)

	// Show container runtime architecture
	if runtime.GOOS == "darwin" {
		fmt.Println("  NOTE: macOS uses lightweight VMs (e.g., Docker Desktop)")
		fmt.Println("  to run a Linux kernel, which then provides namespaces.")
		fmt.Println("  macOS itself uses different isolation mechanisms:")
		fmt.Println("    - App Sandbox (entitlements)")
		fmt.Println("    - sandbox-exec (deprecated)")
		fmt.Println("    - System Integrity Protection (SIP)")
	}

	fmt.Println("\n========================================")
	fmt.Println("Container Isolation Demo complete!")
	fmt.Println("========================================")
}

// ========================================
// Utility Functions
// ========================================

// printDirTree prints a directory tree with indentation.
func printDirTree(root string, prefix string, isRoot bool) {
	if isRoot {
		fmt.Printf("  /\n")
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}

	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		fmt.Printf("  %s%s%s\n", prefix, connector, name)

		if entry.IsDir() {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			printDirTree(filepath.Join(root, entry.Name()), newPrefix, false)
		}
	}
}

// Ensure syscall is used (for namespace constants reference).
var _ = strings.Contains
var _ = syscall.SIGTERM
