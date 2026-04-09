package main

// ========================================
// Dashboard App — Go Backend
// ========================================
// This file contains the DashboardApp struct with all
// bound methods that the frontend calls to get system data.
//
// Methods gather information using:
//   - runtime package: CPU count, memory stats, goroutines
//   - os package: hostname, environment, file system
//   - os/exec package: system commands for disk, processes

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ========================================
// Data Types
// ========================================
// These structs are returned to the frontend and
// automatically become JavaScript objects.

// SystemOverview contains high-level system information.
type SystemOverview struct {
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	GoVersion    string `json:"goVersion"`
	NumCPU       int    `json:"numCPU"`
	NumGoroutine int    `json:"numGoroutine"`
	Uptime       string `json:"uptime"`
}

// MemoryStats contains memory usage data from the Go runtime.
type MemoryStats struct {
	Alloc      string `json:"alloc"`
	TotalAlloc string `json:"totalAlloc"`
	Sys        string `json:"sys"`
	NumGC      uint32 `json:"numGC"`
	AllocBytes uint64 `json:"allocBytes"`
	SysBytes   uint64 `json:"sysBytes"`
	UsedPct    int    `json:"usedPct"`
}

// DiskInfo contains information about a disk mount point.
type DiskInfo struct {
	FileSystem  string `json:"fileSystem"`
	Mount       string `json:"mount"`
	Total       string `json:"total"`
	Used        string `json:"used"`
	Available   string `json:"available"`
	UsedPercent string `json:"usedPercent"`
}

// ProcessInfo contains information about running processes.
type ProcessInfo struct {
	Total      int    `json:"total"`
	CurrentPID int    `json:"currentPID"`
	ParentPID  int    `json:"parentPID"`
	Executable string `json:"executable"`
}

// NetworkInfo contains basic network interface information.
type NetworkInfo struct {
	Hostname    string            `json:"hostname"`
	Interfaces  []string          `json:"interfaces"`
	ExternalIP  string            `json:"externalIP"`
	DNSServers  []string          `json:"dnsServers"`
	Connections map[string]string `json:"connections"`
}

// EnvironmentInfo contains environment variable information.
type EnvironmentInfo struct {
	User     string `json:"user"`
	Home     string `json:"home"`
	Shell    string `json:"shell"`
	Path     string `json:"path"`
	GoPath   string `json:"goPath"`
	GoRoot   string `json:"goRoot"`
	TempDir  string `json:"tempDir"`
	Hostname string `json:"hostname"`
}

// CPUSnapshot represents a CPU usage sample.
type CPUSnapshot struct {
	Time       string  `json:"time"`
	UsageEstimate float64 `json:"usageEstimate"`
	NumCPU     int     `json:"numCPU"`
	Goroutines int     `json:"goroutines"`
}

// ========================================
// DashboardApp
// ========================================

// DashboardApp is the main application struct bound to the
// Wails frontend. All exported methods are callable from JS.
type DashboardApp struct {
	ctx       context.Context
	startTime time.Time
}

// NewDashboardApp creates a new dashboard application.
func NewDashboardApp() *DashboardApp {
	return &DashboardApp{}
}

// ========================================
// Lifecycle Methods
// ========================================

// Startup is called when the Wails app starts.
func (d *DashboardApp) Startup(ctx context.Context) {
	d.ctx = ctx
	d.startTime = time.Now()
	fmt.Println("[Dashboard] Application started")
}

// DomReady is called when the frontend is loaded.
func (d *DashboardApp) DomReady(ctx context.Context) {
	fmt.Println("[Dashboard] Frontend ready")

	// In a real app, you could start a goroutine to push
	// periodic updates via events:
	//
	// go func() {
	//     ticker := time.NewTicker(2 * time.Second)
	//     defer ticker.Stop()
	//     for range ticker.C {
	//         stats := d.GetMemoryStats()
	//         runtime.EventsEmit(d.ctx, "stats:memory", stats)
	//         cpu := d.GetCPUSnapshot()
	//         runtime.EventsEmit(d.ctx, "stats:cpu", cpu)
	//     }
	// }()
}

// Shutdown is called when the application is closing.
func (d *DashboardApp) Shutdown(ctx context.Context) {
	fmt.Println("[Dashboard] Application shutting down")
}

// ========================================
// System Overview
// ========================================

// GetSystemOverview returns high-level system information.
func (d *DashboardApp) GetSystemOverview() (SystemOverview, error) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	uptime := time.Since(d.startTime).Round(time.Second).String()

	osName := runtime.GOOS
	switch osName {
	case "darwin":
		osName = "macOS"
	case "linux":
		osName = "Linux"
	case "windows":
		osName = "Windows"
	}

	return SystemOverview{
		Hostname:     hostname,
		OS:           osName,
		Arch:         runtime.GOARCH,
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		Uptime:       uptime,
	}, nil
}

// ========================================
// Memory Statistics
// ========================================

// GetMemoryStats returns current Go runtime memory statistics.
// Note: These are Go process stats, not system-wide stats.
// For system-wide stats, you'd use OS-specific commands
// or a library like github.com/shirou/gopsutil.
func (d *DashboardApp) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	usedPct := 0
	if m.Sys > 0 {
		usedPct = int(float64(m.Alloc) / float64(m.Sys) * 100)
	}

	return MemoryStats{
		Alloc:      formatBytes(m.Alloc),
		TotalAlloc: formatBytes(m.TotalAlloc),
		Sys:        formatBytes(m.Sys),
		NumGC:      m.NumGC,
		AllocBytes: m.Alloc,
		SysBytes:   m.Sys,
		UsedPct:    usedPct,
	}
}

// ========================================
// Disk Usage
// ========================================

// GetDiskUsage returns disk usage information.
// Uses the 'df' command on Unix systems.
func (d *DashboardApp) GetDiskUsage() ([]DiskInfo, error) {
	var disks []DiskInfo

	switch runtime.GOOS {
	case "darwin", "linux":
		// Use df -h for human-readable output
		out, err := exec.Command("df", "-h").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to run df: %w", err)
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines[1:] { // Skip header
			fields := strings.Fields(line)
			if len(fields) < 6 {
				continue
			}

			mount := fields[len(fields)-1]
			// Filter to interesting mount points
			if !strings.HasPrefix(mount, "/") ||
				strings.HasPrefix(mount, "/dev") ||
				strings.HasPrefix(mount, "/sys") ||
				strings.HasPrefix(mount, "/proc") ||
				strings.HasPrefix(mount, "/run") ||
				strings.HasPrefix(mount, "/snap") {
				if mount != "/" {
					continue
				}
			}

			disk := DiskInfo{
				FileSystem:  fields[0],
				Total:       fields[1],
				Used:        fields[2],
				Available:   fields[3],
				UsedPercent: strings.TrimSuffix(fields[4], "%"),
				Mount:       mount,
			}
			disks = append(disks, disk)
		}

	case "windows":
		// Use wmic on Windows
		out, err := exec.Command("wmic", "logicaldisk", "get",
			"caption,size,freespace").Output()
		if err != nil {
			return nil, fmt.Errorf("failed to run wmic: %w", err)
		}

		lines := strings.Split(string(out), "\n")
		for _, line := range lines[1:] {
			fields := strings.Fields(line)
			if len(fields) < 3 {
				continue
			}

			totalBytes, _ := strconv.ParseUint(fields[2], 10, 64)
			freeBytes, _ := strconv.ParseUint(fields[1], 10, 64)
			usedBytes := totalBytes - freeBytes
			usedPct := 0.0
			if totalBytes > 0 {
				usedPct = float64(usedBytes) / float64(totalBytes) * 100
			}

			disk := DiskInfo{
				FileSystem:  fields[0],
				Total:       formatBytes(totalBytes),
				Used:        formatBytes(usedBytes),
				Available:   formatBytes(freeBytes),
				UsedPercent: fmt.Sprintf("%.0f", usedPct),
				Mount:       fields[0],
			}
			disks = append(disks, disk)
		}
	}

	if len(disks) == 0 {
		// Provide a fallback entry
		disks = append(disks, DiskInfo{
			FileSystem:  "unknown",
			Total:       "N/A",
			Used:        "N/A",
			Available:   "N/A",
			UsedPercent: "0",
			Mount:       "/",
		})
	}

	return disks, nil
}

// ========================================
// Process Information
// ========================================

// GetProcessInfo returns information about running processes.
func (d *DashboardApp) GetProcessInfo() ProcessInfo {
	// Get approximate process count
	total := 0
	switch runtime.GOOS {
	case "darwin", "linux":
		out, err := exec.Command("sh", "-c", "ps aux | wc -l").Output()
		if err == nil {
			count, err := strconv.Atoi(strings.TrimSpace(string(out)))
			if err == nil {
				total = count - 1 // Subtract header line
			}
		}
	case "windows":
		out, err := exec.Command("tasklist").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			total = len(lines) - 3 // Subtract header lines
		}
	}

	execPath, _ := os.Executable()

	return ProcessInfo{
		Total:      total,
		CurrentPID: os.Getpid(),
		ParentPID:  os.Getppid(),
		Executable: execPath,
	}
}

// ========================================
// Network Information
// ========================================

// GetNetworkInfo returns basic network information.
func (d *DashboardApp) GetNetworkInfo() NetworkInfo {
	hostname, _ := os.Hostname()

	info := NetworkInfo{
		Hostname:    hostname,
		Interfaces:  []string{},
		ExternalIP:  "N/A (run curl ifconfig.me)",
		DNSServers:  []string{},
		Connections: map[string]string{},
	}

	// Get network interfaces
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("ifconfig", "-l").Output()
		if err == nil {
			info.Interfaces = strings.Fields(string(out))
		}

		// Get DNS servers
		dnsOut, err := exec.Command("sh", "-c",
			"scutil --dns | grep 'nameserver' | head -5 | awk '{print $3}'").Output()
		if err == nil {
			servers := strings.Fields(string(dnsOut))
			info.DNSServers = servers
		}

	case "linux":
		out, err := exec.Command("sh", "-c",
			"ip -o link show | awk '{print $2}' | tr -d ':'").Output()
		if err == nil {
			info.Interfaces = strings.Fields(string(out))
		}

		// Get DNS from resolv.conf
		dnsOut, err := exec.Command("sh", "-c",
			"grep 'nameserver' /etc/resolv.conf | awk '{print $2}'").Output()
		if err == nil {
			info.DNSServers = strings.Fields(string(dnsOut))
		}
	}

	// Count network connections
	switch runtime.GOOS {
	case "darwin", "linux":
		out, err := exec.Command("sh", "-c",
			"netstat -an 2>/dev/null | grep -c ESTABLISHED || echo 0").Output()
		if err == nil {
			info.Connections["established"] = strings.TrimSpace(string(out))
		}
	}

	return info
}

// ========================================
// Environment Information
// ========================================

// GetEnvironmentInfo returns environment variable info.
func (d *DashboardApp) GetEnvironmentInfo() EnvironmentInfo {
	currentUser, _ := user.Current()
	hostname, _ := os.Hostname()

	username := "unknown"
	homeDir := "unknown"
	if currentUser != nil {
		username = currentUser.Username
		homeDir = currentUser.HomeDir
	}

	return EnvironmentInfo{
		User:     username,
		Home:     homeDir,
		Shell:    os.Getenv("SHELL"),
		Path:     os.Getenv("PATH"),
		GoPath:   os.Getenv("GOPATH"),
		GoRoot:   runtime.GOROOT(),
		TempDir:  os.TempDir(),
		Hostname: hostname,
	}
}

// ========================================
// CPU Snapshot
// ========================================

// GetCPUSnapshot returns a point-in-time CPU snapshot.
// Note: Accurate CPU usage measurement requires sampling
// over time. For production use, consider gopsutil.
func (d *DashboardApp) GetCPUSnapshot() CPUSnapshot {
	snapshot := CPUSnapshot{
		Time:       time.Now().Format("15:04:05"),
		NumCPU:     runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
	}

	// Try to get load average on Unix systems
	switch runtime.GOOS {
	case "darwin", "linux":
		out, err := exec.Command("sh", "-c", "uptime | awk -F'load average:' '{print $2}' | awk -F, '{print $1}'").Output()
		if err == nil {
			loadStr := strings.TrimSpace(string(out))
			load, err := strconv.ParseFloat(loadStr, 64)
			if err == nil {
				// Convert load average to approximate CPU percentage
				// Load / NumCPU * 100 gives a rough estimate
				snapshot.UsageEstimate = (load / float64(runtime.NumCPU())) * 100
				if snapshot.UsageEstimate > 100 {
					snapshot.UsageEstimate = 100
				}
			}
		}
	}

	return snapshot
}

// ========================================
// History for Charts
// ========================================

// GetCPUHistory returns recent CPU snapshots for charting.
// In a real app, you'd store historical data. Here we
// generate sample data to demonstrate the frontend chart.
func (d *DashboardApp) GetCPUHistory() []CPUSnapshot {
	var history []CPUSnapshot
	now := time.Now()

	for i := 9; i >= 0; i-- {
		t := now.Add(-time.Duration(i) * 2 * time.Second)
		snapshot := CPUSnapshot{
			Time:       t.Format("15:04:05"),
			NumCPU:     runtime.NumCPU(),
			Goroutines: runtime.NumGoroutine(),
		}

		// Get current estimate for the latest point
		if i == 0 {
			current := d.GetCPUSnapshot()
			snapshot.UsageEstimate = current.UsageEstimate
		} else {
			// Historical values would come from stored data
			snapshot.UsageEstimate = 0
		}

		history = append(history, snapshot)
	}

	return history
}

// ========================================
// Helpers
// ========================================

// formatBytes converts bytes to a human-readable string.
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
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
