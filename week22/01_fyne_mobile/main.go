package main

// ========================================
// Week 22 — Lesson 1: Fyne for Mobile
// ========================================
// This lesson covers:
//   - Adapting Fyne apps for mobile devices
//   - Responsive layouts for different screen sizes
//   - Touch-friendly widgets and interactions
//   - Mobile-specific considerations
//   - Building for iOS and Android
//
// Key differences between desktop and mobile:
//   - Smaller screens require simpler layouts
//   - Touch targets should be at least 44x44 dp
//   - No hover states — design for tap, swipe, long-press
//   - Virtual keyboard overlays content
//   - Screen orientation can change
//   - Limited multitasking — apps may be suspended
//   - Battery and data usage matter
//
// Prerequisites:
//   1. Install Fyne CLI:
//      go install fyne.io/fyne/v2/cmd/fyne@latest
//
//   2. For iOS builds:
//      - macOS with Xcode installed
//      - Apple Developer account (for device deployment)
//      - fyne package -os ios -appID com.example.myapp
//
//   3. For Android builds:
//      - Android SDK and NDK installed
//      - Set ANDROID_HOME and ANDROID_NDK_HOME
//      - fyne package -os android -appID com.example.myapp
//
// Build commands:
//   Desktop:  go run .
//   iOS:      fyne package -os ios -appID com.example.fynemobile
//   Android:  fyne package -os android -appID com.example.fynemobile
//   Web:      fyne package -os web

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 22 - Lesson 1: Fyne Mobile")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("This app demonstrates mobile-friendly")
	fmt.Println("Fyne patterns. On desktop, resize the")
	fmt.Println("window to simulate different screen sizes.")
	fmt.Println()

	myApp := app.NewWithID("com.example.fynemobile")
	myWindow := myApp.NewWindow("Fyne Mobile Demo")

	// ========================================
	// Responsive Layout Strategy
	// ========================================
	// Fyne uses device-independent pixels (dp) which
	// automatically scale for screen density. However,
	// you still need to design for different screen SIZES.
	//
	// Mobile strategy:
	//   - Use VBox as the primary layout (vertical scrolling)
	//   - Avoid HBox with many items (horizontal space is limited)
	//   - Use container.NewScroll for long content
	//   - Use AppTabs for navigation (becomes bottom tabs on mobile)
	//   - Full-width buttons and inputs

	// ========================================
	// Touch-Friendly Widgets
	// ========================================
	// On mobile, buttons should be large enough to tap easily.
	// Fyne handles this, but custom widgets should ensure
	// a minimum tap target of 44x44 dp.

	// Large, touch-friendly buttons
	primaryButton := widget.NewButton("Primary Action", func() {
		fmt.Println("Primary button tapped!")
	})
	primaryButton.Importance = widget.HighImportance

	secondaryButton := widget.NewButton("Secondary Action", func() {
		fmt.Println("Secondary button tapped!")
	})

	// ========================================
	// Mobile-Optimized Input
	// ========================================
	// On mobile, Entry widgets trigger the virtual keyboard.
	// Consider:
	//   - Using placeholders to save label space
	//   - Minimizing the number of input fields visible at once
	//   - Using Select dropdowns instead of free-text where possible

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Your name")

	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email address")

	// Select is more mobile-friendly than free text for
	// constrained choices
	categorySelect := widget.NewSelect(
		[]string{"Personal", "Work", "Shopping", "Health", "Travel"},
		func(selected string) {
			fmt.Printf("Selected: %s\n", selected)
		},
	)
	categorySelect.PlaceHolder = "Select category"

	// ========================================
	// Card-Style Layout
	// ========================================
	// Cards provide visual grouping that works well on mobile.
	// They give content clear boundaries and padding.

	infoCard := widget.NewCard(
		"Welcome",
		"This app adapts to mobile screens",
		container.NewVBox(
			widget.NewLabel("Fyne apps run on iOS, Android, and desktop."),
			widget.NewLabel("The same Go code powers all platforms."),
		),
	)

	inputCard := widget.NewCard(
		"Quick Form",
		"Mobile-optimized inputs",
		container.NewVBox(
			nameEntry,
			emailEntry,
			categorySelect,
			primaryButton,
		),
	)

	// ========================================
	// List-Based Navigation
	// ========================================
	// On mobile, lists are the primary navigation pattern.
	// They provide large tap targets and clear hierarchy.

	menuItems := []struct {
		title    string
		subtitle string
	}{
		{"Profile", "View and edit your profile"},
		{"Settings", "App preferences and configuration"},
		{"Notifications", "Manage notification settings"},
		{"Privacy", "Privacy and security options"},
		{"Help & Support", "FAQ and contact support"},
		{"About", "App version and credits"},
	}

	menuList := widget.NewList(
		func() int { return len(menuItems) },
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabelWithStyle("Title", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Subtitle"),
			)
		},
		func(index widget.ListItemID, item fyne.CanvasObject) {
			box := item.(*fyne.Container)
			titleWidget := box.Objects[0].(*widget.RichText)
			subtitleWidget := box.Objects[1].(*widget.Label)
			titleWidget.ParseMarkdown("**" + menuItems[index].title + "**")
			subtitleWidget.SetText(menuItems[index].subtitle)
		},
	)
	menuList.OnSelected = func(id widget.ListItemID) {
		fmt.Printf("Selected: %s\n", menuItems[id].title)
	}

	settingsCard := widget.NewCard(
		"Menu",
		"Tap any item to navigate",
		container.NewGridWrap(fyne.NewSize(400, 250), menuList),
	)

	// ========================================
	// Orientation Awareness
	// ========================================
	// Fyne handles orientation changes automatically.
	// Your layouts should be flexible enough to work in
	// both portrait and landscape. Tips:
	//   - Prefer VBox with Scroll for portrait
	//   - Use percentage-based widths (Grid) not fixed
	//   - Test by resizing the desktop window

	orientationLabel := widget.NewLabel("Resize window to test responsive layout")

	// ========================================
	// Mobile-Style Action Bar
	// ========================================
	// Instead of traditional menus, mobile apps use
	// action bars or bottom navigation.

	actionBar := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.HomeIcon(), func() {
			fmt.Println("Home tapped")
		}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
			fmt.Println("Search tapped")
		}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
			fmt.Println("Add tapped")
		}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.AccountIcon(), func() {
			fmt.Println("Profile tapped")
		}),
		layout.NewSpacer(),
	)

	// Visual separator for the action bar
	actionBarBg := canvas.NewRectangle(theme.Color(theme.ColorNameOverlayBackground))
	actionBarBg.SetMinSize(fyne.NewSize(0, 50))

	bottomBar := container.NewStack(
		actionBarBg,
		container.NewCenter(actionBar),
	)

	// ========================================
	// Tab Navigation
	// ========================================
	// AppTabs on mobile automatically become bottom tabs.
	// Keep tab count to 3-5 for mobile usability.

	homeTab := container.NewTabItemWithIcon("Home", theme.HomeIcon(),
		container.NewScroll(
			container.NewVBox(
				infoCard,
				inputCard,
				orientationLabel,
			),
		),
	)

	menuTab := container.NewTabItemWithIcon("Menu", theme.ListIcon(),
		container.NewScroll(
			container.NewVBox(
				settingsCard,
			),
		),
	)

	// ========================================
	// Pull-to-Refresh Pattern
	// ========================================
	// While Fyne doesn't have a built-in pull-to-refresh,
	// you can implement refresh with a button.
	refreshLabel := widget.NewLabel("Last refreshed: never")
	refreshButton := widget.NewButtonWithIcon("Refresh Data", theme.ViewRefreshIcon(), func() {
		refreshLabel.SetText(fmt.Sprintf("Last refreshed: %s", time.Now().Format("3:04:05 PM")))
	})

	actionsTab := container.NewTabItemWithIcon("Actions", theme.SettingsIcon(),
		container.NewScroll(
			container.NewVBox(
				widget.NewCard("Refresh", "Simulate pull-to-refresh", container.NewVBox(
					refreshLabel,
					refreshButton,
				)),
				widget.NewCard("Buttons", "Touch-friendly actions", container.NewVBox(
					primaryButton,
					secondaryButton,
					widget.NewButton("Danger Action", func() {}),
				)),
			),
		),
	)

	tabs := container.NewAppTabs(homeTab, menuTab, actionsTab)
	tabs.SetTabLocation(container.TabLocationBottom) // Bottom tabs for mobile

	// ========================================
	// Final Layout
	// ========================================
	// Border layout with bottom action bar
	mainLayout := container.NewBorder(
		nil,       // top
		bottomBar, // bottom — persistent navigation
		nil, nil,
		tabs, // center — tab content
	)

	myWindow.SetContent(mainLayout)

	// Set a mobile-like default size
	// On actual mobile devices, this is ignored (fullscreen)
	myWindow.Resize(fyne.NewSize(375, 667)) // iPhone-like dimensions
	myWindow.CenterOnScreen()

	fmt.Println("Launching mobile demo...")
	fmt.Println("Try resizing the window to simulate different devices.")
	myWindow.ShowAndRun()
	fmt.Println("Application closed.")
}
