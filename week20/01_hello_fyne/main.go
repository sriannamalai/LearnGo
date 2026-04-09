package main

// ========================================
// Week 20 — Lesson 1: Hello Fyne
// ========================================
// This lesson covers:
//   - Installing and setting up Fyne
//   - Creating your first window
//   - Basic app lifecycle (Run, Quit)
//   - Your first widget (Label)
//   - Show vs ShowAndRun
//
// Prerequisites:
//   1. Install Fyne prerequisites for your OS:
//      - macOS: Xcode command line tools (xcode-select --install)
//      - Linux: sudo apt-get install gcc libgl1-mesa-dev xorg-dev
//      - Windows: MSYS2 with mingw-w64 GCC
//   2. go get fyne.io/fyne/v2@latest
//
// Run:
//   go run .

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 20 - Lesson 1: Hello Fyne")
	fmt.Println("========================================")
	fmt.Println()

	// ========================================
	// Creating a Fyne Application
	// ========================================
	// app.New() creates a new Fyne application instance.
	// This is the entry point for every Fyne program.
	// You can also use app.NewWithID("com.example.myapp")
	// to give your app a unique identifier (useful for
	// preferences storage and system integration).
	myApp := app.New()

	// ========================================
	// Creating a Window
	// ========================================
	// NewWindow creates a new window with a title.
	// The window is not yet visible — you must call Show()
	// or ShowAndRun() to display it.
	myWindow := myApp.NewWindow("Hello Fyne!")

	// ========================================
	// Setting Window Size
	// ========================================
	// Resize sets the initial window dimensions.
	// Fyne uses device-independent pixels (dp) so your
	// UI looks consistent across different screen densities.
	myWindow.Resize(fyne.NewSize(500, 400))

	// ========================================
	// Center the Window on Screen
	// ========================================
	// CenterOnScreen positions the window in the center
	// of the user's display. Without this, the window
	// position is determined by the OS window manager.
	myWindow.CenterOnScreen()

	// ========================================
	// Creating a Label Widget
	// ========================================
	// widget.NewLabel creates a simple text label.
	// Labels are the most basic widget — they display
	// read-only text to the user.
	helloLabel := widget.NewLabel("Hello, Fyne!")

	// You can style labels:
	// - Bold text with NewLabelWithStyle
	// - Text alignment (Leading, Center, Trailing)
	// - Text wrapping for long content
	styledLabel := widget.NewLabelWithStyle(
		"Welcome to Desktop App Development with Go!",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// ========================================
	// Rich Text Labels
	// ========================================
	// For more complex text, NewRichTextFromMarkdown
	// allows you to use Markdown formatting.
	richText := widget.NewRichTextFromMarkdown(
		"**Fyne** is a _cross-platform_ GUI toolkit for Go.\n\n" +
			"It supports:\n" +
			"- Windows, macOS, Linux\n" +
			"- iOS and Android\n" +
			"- Clean, modern design",
	)

	// ========================================
	// Building the Layout
	// ========================================
	// container.NewVBox arranges widgets vertically.
	// Each widget stacks below the previous one.
	// We'll explore layouts in depth in Lesson 3.
	content := container.NewVBox(
		styledLabel,
		widget.NewSeparator(),
		helloLabel,
		richText,
		widget.NewSeparator(),
		widget.NewLabel(fmt.Sprintf("Started at: %s", time.Now().Format("3:04 PM"))),
	)

	// ========================================
	// Adding a Simple Button
	// ========================================
	// Buttons take a label string and a callback function
	// that runs when the button is clicked.
	clickCount := 0
	counterLabel := widget.NewLabel("Clicks: 0")

	clickButton := widget.NewButton("Click Me!", func() {
		clickCount++
		counterLabel.SetText(fmt.Sprintf("Clicks: %d", clickCount))
		fmt.Printf("Button clicked! Total: %d\n", clickCount)
	})

	// Add button and counter to the layout
	content.Add(widget.NewSeparator())
	content.Add(clickButton)
	content.Add(counterLabel)

	// ========================================
	// Setting Window Content
	// ========================================
	// SetContent assigns a widget tree as the window's content.
	// The content fills the available window area.
	myWindow.SetContent(content)

	// ========================================
	// App Lifecycle
	// ========================================
	// ShowAndRun() does two things:
	//   1. Shows the window (makes it visible)
	//   2. Runs the app event loop (blocks until app quits)
	//
	// Alternative approach:
	//   myWindow.Show()   // Show the window
	//   myApp.Run()       // Start the event loop separately
	//
	// The event loop handles:
	//   - User input (mouse clicks, keyboard)
	//   - Widget rendering and updates
	//   - Window resize and move events
	//   - System events (dark mode changes, etc.)
	//
	// When the user closes the last window, the app
	// exits and Run()/ShowAndRun() returns.

	fmt.Println("Launching Fyne window...")
	fmt.Println("Close the window to exit the application.")
	myWindow.ShowAndRun()

	// Code after ShowAndRun executes when the app exits
	fmt.Println("Application closed. Goodbye!")
}
