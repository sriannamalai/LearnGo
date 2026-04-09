package main

// ========================================
// Week 21 — Lesson 3: Mini-Project
// Desktop Dashboard App (System Stats Viewer)
// ========================================
// This project combines everything from Week 21:
//   - Wails application setup with multiple bound structs
//   - Go backend for system information gathering
//   - Web frontend with charts and auto-refresh
//   - Event system for real-time updates
//
// Architecture:
//   Go Backend (main.go + app.go):
//     - Gathers CPU, memory, disk, and network info
//     - Uses runtime, os, and os/exec packages
//     - Exposes data via bound methods
//
//   Web Frontend (frontend/):
//     - Dashboard UI with cards and progress bars
//     - Auto-refreshes every few seconds
//     - Charts built with pure CSS (no external deps)
//
// Run:
//   wails dev
//
// Build:
//   wails build
//
// Note: Some system stats use platform-specific commands.
// The code includes fallbacks for cross-platform compatibility.

import (
	"context"
	"fmt"
	"log"
	"runtime"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 21 - Mini-Project: Dashboard")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("System Stats Dashboard")
	fmt.Println("This is a Wails desktop application.")
	fmt.Println()

	// ========================================
	// Wails Application Configuration
	// ========================================
	// In a real Wails app, you'd use:
	//
	//   import "github.com/wailsapp/wails/v2"
	//   import "github.com/wailsapp/wails/v2/pkg/options"
	//   import "github.com/wailsapp/wails/v2/pkg/options/assetserver"
	//
	//   //go:embed all:frontend
	//   var assets embed.FS
	//
	//   app := NewDashboardApp()
	//
	//   err := wails.Run(&options.App{
	//       Title:            "System Dashboard",
	//       Width:            1000,
	//       Height:           700,
	//       MinWidth:         800,
	//       MinHeight:        600,
	//       AssetServer:      &assetserver.Options{Assets: assets},
	//       BackgroundColour: &options.RGBA{R: 13, G: 17, B: 23, A: 1},
	//       OnStartup:        app.Startup,
	//       OnDomReady:       app.DomReady,
	//       OnShutdown:       app.Shutdown,
	//       Bind: []interface{}{
	//           app,
	//       },
	//   })
	//
	//   if err != nil {
	//       log.Fatal("Error:", err)
	//   }

	// For demonstration, run the Go backend directly
	app := NewDashboardApp()
	app.Startup(context.Background())

	fmt.Println("\n--- System Overview ---")
	overview, err := app.GetSystemOverview()
	if err != nil {
		log.Printf("Error getting overview: %v\n", err)
	} else {
		fmt.Printf("OS: %s (%s)\n", overview.OS, overview.Arch)
		fmt.Printf("Hostname: %s\n", overview.Hostname)
		fmt.Printf("Go Version: %s\n", overview.GoVersion)
		fmt.Printf("CPUs: %d\n", overview.NumCPU)
		fmt.Printf("Goroutines: %d\n", overview.NumGoroutine)
	}

	fmt.Println("\n--- Memory Stats ---")
	mem := app.GetMemoryStats()
	fmt.Printf("Allocated: %s\n", mem.Alloc)
	fmt.Printf("Total Allocated: %s\n", mem.TotalAlloc)
	fmt.Printf("System Memory: %s\n", mem.Sys)
	fmt.Printf("GC Cycles: %d\n", mem.NumGC)

	fmt.Println("\n--- Disk Usage ---")
	disks, err := app.GetDiskUsage()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		for _, d := range disks {
			fmt.Printf("  %s: %s used of %s (%s%%)\n",
				d.Mount, d.Used, d.Total, d.UsedPercent)
		}
	}

	fmt.Println("\n--- Process Info ---")
	procs := app.GetProcessInfo()
	fmt.Printf("Approximate running processes: %d\n", procs.Total)
	fmt.Printf("Current PID: %d\n", procs.CurrentPID)
	fmt.Printf("PPID: %d\n", procs.ParentPID)

	fmt.Println("\n--- Environment ---")
	env := app.GetEnvironmentInfo()
	fmt.Printf("User: %s\n", env.User)
	fmt.Printf("Home: %s\n", env.Home)
	fmt.Printf("Shell: %s\n", env.Shell)
	fmt.Printf("Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	app.Shutdown(context.Background())
	fmt.Println("\nDashboard demo complete.")
}
