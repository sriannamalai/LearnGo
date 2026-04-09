package main

// ========================================
// Week 21 — Lesson 1: Introduction to Wails
// ========================================
// This lesson covers:
//   - What is Wails and how it works
//   - Wails architecture: Go backend + web frontend
//   - Creating a Wails app
//   - Binding Go functions to the frontend
//   - App lifecycle and events
//
// Prerequisites:
//   1. Install Wails CLI:
//      go install github.com/wailsapp/wails/v2/cmd/wails@latest
//   2. Verify installation:
//      wails doctor
//   3. System requirements:
//      - macOS: Xcode command line tools
//      - Linux: gcc, libgtk-3-dev, libwebkit2gtk-4.0-dev
//      - Windows: WebView2 runtime (included in Win 11)
//
// Wails Architecture:
//   ┌────────────────────────────────────────────┐
//   │              Wails Application              │
//   │                                             │
//   │  ┌──────────────┐    ┌──────────────────┐  │
//   │  │  Go Backend   │<-->│  Web Frontend    │  │
//   │  │               │    │  (HTML/CSS/JS)   │  │
//   │  │  - Business   │    │                  │  │
//   │  │    logic      │    │  - UI rendering  │  │
//   │  │  - File I/O   │    │  - User events   │  │
//   │  │  - System     │    │  - DOM updates   │  │
//   │  │    access     │    │                  │  │
//   │  │  - Networking │    │  Uses system     │  │
//   │  │               │    │  WebView (not    │  │
//   │  │               │    │  bundled browser)│  │
//   │  └──────────────┘    └──────────────────┘  │
//   └────────────────────────────────────────────┘
//
// Key concepts:
//   - Go structs are "bound" to the frontend via wails.Bind()
//   - The frontend calls Go methods as if they were JS functions
//   - Go methods return values that become JS promises
//   - Events system allows bidirectional communication
//   - Uses the native OS WebView, keeping the binary small
//
// Run (development mode):
//   wails dev
//
// Build (production):
//   wails build
//
// Note: This file demonstrates the Go backend structure.
// The frontend/ directory contains the HTML that calls
// these Go functions.

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

// ========================================
// Application Struct
// ========================================
// In Wails, you define a struct whose methods become
// available to the frontend. This is the core of the
// "binding" system — public methods on bound structs
// are automatically callable from JavaScript.

// App is the main application struct. Its exported methods
// will be available in the frontend via window.go.MethodName().
type App struct {
	// ctx is the Wails application context.
	// It provides access to the Wails runtime for events,
	// dialogs, menus, and window operations.
	ctx context.Context
}

// NewApp creates a new App instance.
// This is called once when the application starts.
func NewApp() *App {
	return &App{}
}

// ========================================
// Lifecycle Methods
// ========================================
// Wails calls these methods at specific points in the
// application lifecycle. They're configured in the
// wails.Run() options.

// startup is called when the application starts.
// The context provides access to Wails runtime features.
// Use this for initialization: loading config, connecting
// to databases, setting up resources.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("[Go] Application started")
	fmt.Printf("[Go] Running on %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// domReady is called after the frontend DOM is fully loaded.
// This is safe to start interacting with the frontend from Go.
func (a *App) domReady(ctx context.Context) {
	fmt.Println("[Go] Frontend DOM is ready")
}

// beforeClose is called when the user tries to close the app.
// Return true to prevent closing (e.g., for unsaved changes).
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	fmt.Println("[Go] Application closing...")
	return false // Allow close
}

// shutdown is called when the application is terminating.
// Use this for cleanup: closing connections, saving state.
func (a *App) shutdown(ctx context.Context) {
	fmt.Println("[Go] Application shut down")
}

// ========================================
// Bound Methods — Called from Frontend
// ========================================
// Any exported (capitalized) method on a bound struct
// is automatically available in the frontend JavaScript.
// Parameters and return values are automatically marshaled
// between Go and JavaScript.

// Greet takes a name and returns a greeting message.
// In the frontend, call this as:
//   window.go.main.App.Greet("Alice")
// It returns a Promise that resolves to the greeting string.
func (a *App) Greet(name string) string {
	fmt.Printf("[Go] Greet called with: %s\n", name)

	if name == "" {
		return "Hello, World!"
	}
	return fmt.Sprintf("Hello, %s! Welcome to Wails!", name)
}

// GetTime returns the current server time as a formatted string.
// Demonstrates a simple no-argument bound method.
func (a *App) GetTime() string {
	now := time.Now()
	return now.Format("Monday, January 2, 2006 at 3:04:05 PM")
}

// GetSystemInfo returns basic system information.
// Demonstrates returning complex data that becomes
// a JavaScript object.
func (a *App) GetSystemInfo() map[string]string {
	return map[string]string{
		"os":       runtime.GOOS,
		"arch":     runtime.GOARCH,
		"goVer":    runtime.Version(),
		"cpus":     fmt.Sprintf("%d", runtime.NumCPU()),
		"compiler": runtime.Compiler,
	}
}

// Add performs a simple addition. Demonstrates multiple
// parameters and numeric return values.
func (a *App) Add(x, y int) int {
	fmt.Printf("[Go] Add called: %d + %d\n", x, y)
	return x + y
}

// ========================================
// Main Entry Point
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 21 - Lesson 1: Wails Intro")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("This is a Wails application.")
	fmt.Println("In a real Wails project, this main function")
	fmt.Println("would call wails.Run() to start the app.")
	fmt.Println()
	fmt.Println("To create a proper Wails project:")
	fmt.Println("  wails init -n myapp -t vanilla")
	fmt.Println()
	fmt.Println("The code below demonstrates the Go backend")
	fmt.Println("structure. See frontend/index.html for the UI.")
	fmt.Println()

	// ========================================
	// Wails Application Setup
	// ========================================
	// In a real Wails app, you would use:
	//
	//   import "github.com/wailsapp/wails/v2"
	//   import "github.com/wailsapp/wails/v2/pkg/options"
	//   import "github.com/wailsapp/wails/v2/pkg/options/assetserver"
	//
	//   app := NewApp()
	//
	//   err := wails.Run(&options.App{
	//       Title:  "Hello Wails",
	//       Width:  800,
	//       Height: 600,
	//       AssetServer: &assetserver.Options{
	//           Assets: assets,  // embed.FS with frontend files
	//       },
	//       OnStartup:     app.startup,
	//       OnDomReady:    app.domReady,
	//       OnBeforeClose: app.beforeClose,
	//       OnShutdown:    app.shutdown,
	//       Bind: []interface{}{
	//           app,  // Bind the App struct
	//       },
	//   })
	//
	//   if err != nil {
	//       log.Fatal(err)
	//   }

	// Demonstrate the Go functions directly
	app := NewApp()

	// Simulate lifecycle
	ctx := context.Background()
	app.startup(ctx)
	app.domReady(ctx)

	// Test bound methods
	fmt.Println("\n--- Testing Bound Methods ---")
	fmt.Println(app.Greet("Gopher"))
	fmt.Println(app.GetTime())
	fmt.Printf("System Info: %v\n", app.GetSystemInfo())
	fmt.Printf("3 + 5 = %d\n", app.Add(3, 5))

	app.shutdown(ctx)
}
