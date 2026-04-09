package main

// ========================================
// Week 20 — Lesson 2: Widgets
// ========================================
// This lesson covers:
//   - Common Fyne widgets: Button, Entry, Label, Check,
//     Radio, Select, Slider, ProgressBar
//   - Event handling with OnChanged/OnTapped callbacks
//   - Data binding basics
//   - Building interactive forms
//
// Run:
//   go run .

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 20 - Lesson 2: Widgets")
	fmt.Println("========================================")
	fmt.Println()

	myApp := app.New()
	myWindow := myApp.NewWindow("Fyne Widgets Showcase")
	myWindow.Resize(fyne.NewSize(600, 700))
	myWindow.CenterOnScreen()

	// ========================================
	// Label Widgets
	// ========================================
	// Labels display read-only text. They are the simplest
	// widget and are used for headings, descriptions, and
	// displaying dynamic information.
	titleLabel := widget.NewLabelWithStyle(
		"Widget Showcase",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	statusLabel := widget.NewLabel("Status: Ready")

	// ========================================
	// Entry Widget (Text Input)
	// ========================================
	// Entry is a single-line text input field.
	// It supports placeholder text, validation, and
	// the OnChanged callback fires on every keystroke.
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter your name...")

	// OnChanged fires every time the text changes.
	// This is useful for live validation or filtering.
	nameEntry.OnChanged = func(text string) {
		fmt.Printf("Name field changed: %q\n", text)
	}

	// OnSubmitted fires when the user presses Enter.
	nameEntry.OnSubmitted = func(text string) {
		statusLabel.SetText(fmt.Sprintf("Status: Hello, %s!", text))
	}

	// PasswordEntry hides the typed characters.
	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Enter password...")

	// MultiLineEntry allows multiple lines of text.
	multiEntry := widget.NewMultiLineEntry()
	multiEntry.SetPlaceHolder("Enter a longer message here...")
	multiEntry.SetMinRowsVisible(3)

	// ========================================
	// Button Widget
	// ========================================
	// Buttons trigger actions when clicked.
	// They can have text labels and optional icons.
	submitButton := widget.NewButton("Submit", func() {
		name := nameEntry.Text
		message := multiEntry.Text
		if name == "" {
			statusLabel.SetText("Status: Please enter your name!")
			return
		}
		statusLabel.SetText(fmt.Sprintf("Status: Submitted by %s", name))
		fmt.Printf("Submitted — Name: %s, Message: %s\n", name, message)
	})
	submitButton.Importance = widget.HighImportance // Blue/primary button

	clearButton := widget.NewButton("Clear", func() {
		nameEntry.SetText("")
		passwordEntry.SetText("")
		multiEntry.SetText("")
		statusLabel.SetText("Status: Cleared")
	})

	// ========================================
	// Check Widget (Checkbox)
	// ========================================
	// Check is a boolean toggle with a label.
	// OnChanged fires when the state changes.
	agreeCheck := widget.NewCheck("I agree to the terms", func(checked bool) {
		if checked {
			submitButton.Enable()
			fmt.Println("Terms accepted")
		} else {
			submitButton.Disable()
			fmt.Println("Terms declined")
		}
	})
	agreeCheck.SetChecked(true)

	darkModeCheck := widget.NewCheck("Enable dark mode", func(checked bool) {
		mode := "light"
		if checked {
			mode = "dark"
		}
		statusLabel.SetText(fmt.Sprintf("Status: %s mode selected", mode))
	})

	// ========================================
	// Radio Widget (Radio Buttons)
	// ========================================
	// Radio presents a list of mutually exclusive options.
	// Only one can be selected at a time.
	priorityRadio := widget.NewRadioGroup(
		[]string{"Low", "Medium", "High", "Critical"},
		func(selected string) {
			fmt.Printf("Priority selected: %s\n", selected)
			statusLabel.SetText(fmt.Sprintf("Status: Priority set to %s", selected))
		},
	)
	priorityRadio.SetSelected("Medium")
	priorityRadio.Horizontal = true // Display options in a row

	// ========================================
	// Select Widget (Dropdown)
	// ========================================
	// Select provides a dropdown list of options.
	// It takes less space than Radio for many options.
	categorySelect := widget.NewSelect(
		[]string{"General", "Bug Report", "Feature Request", "Question", "Documentation"},
		func(selected string) {
			fmt.Printf("Category selected: %s\n", selected)
			statusLabel.SetText(fmt.Sprintf("Status: Category is %s", selected))
		},
	)
	categorySelect.SetSelected("General")
	categorySelect.PlaceHolder = "Choose a category..."

	// ========================================
	// Slider Widget
	// ========================================
	// Slider allows selecting a numeric value within a range.
	// OnChanged fires as the slider moves.
	sliderValueLabel := widget.NewLabel("Font size: 14")

	fontSlider := widget.NewSlider(8, 36)
	fontSlider.Value = 14
	fontSlider.Step = 1
	fontSlider.OnChanged = func(value float64) {
		sliderValueLabel.SetText(fmt.Sprintf("Font size: %d", int(value)))
	}

	// ========================================
	// ProgressBar Widget
	// ========================================
	// ProgressBar shows completion status (0.0 to 1.0).
	// ProgressBarInfinite shows an indeterminate animation.
	progress := widget.NewProgressBar()
	progress.SetValue(0.0)

	infiniteProgress := widget.NewProgressBarInfinite()
	infiniteProgress.Stop() // Start stopped

	progressButton := widget.NewButton("Simulate Progress", func() {
		infiniteProgress.Start()
		go func() {
			for i := 0; i <= 100; i += 10 {
				progress.SetValue(float64(i) / 100.0)
				// In a real app, this would track actual work.
				// time.Sleep(200 * time.Millisecond)
			}
			progress.SetValue(1.0)
			infiniteProgress.Stop()
			statusLabel.SetText("Status: Progress complete!")
		}()
	})

	// ========================================
	// Hyperlink Widget
	// ========================================
	// Hyperlink opens a URL in the system browser.
	// Useful for linking to documentation or websites.
	// link, _ := url.Parse("https://docs.fyne.io")
	// hyperlink := widget.NewHyperlink("Fyne Documentation", link)

	// ========================================
	// Data Binding Basics
	// ========================================
	// Data binding connects a data source to a widget.
	// When the data changes, the widget updates automatically,
	// and when the widget changes, the data updates too.
	//
	// This eliminates manual synchronization between
	// your data model and the UI.

	// Create a bound string
	boundName := binding.NewString()
	boundName.Set("World")

	// Create a widget bound to the data
	boundEntry := widget.NewEntryWithData(boundName)
	boundEntry.SetPlaceHolder("Type a name...")

	// This label automatically updates when boundName changes
	boundLabel := widget.NewLabelWithData(
		binding.NewSprintf("Hello, %s!", boundName),
	)

	// Create a bound float for a slider
	boundValue := binding.NewFloat()
	boundValue.Set(50.0)

	boundSlider := widget.NewSliderWithData(0, 100, boundValue)

	// Format the float value as a string for display
	boundValueStr := binding.FloatToStringWithFormat(boundValue, "Value: %.0f%%")
	boundValueLabel := widget.NewLabelWithData(boundValueStr)

	// Create a bound bool for a checkbox
	boundBool := binding.NewBool()
	boundBool.Set(false)

	boundCheck := widget.NewCheckWithData("Toggle me", boundBool)

	boundBoolStr := binding.BoolToString(boundBool)
	boundBoolLabel := widget.NewLabelWithData(
		binding.NewSprintf("Checked: %s", boundBoolStr),
	)

	// ========================================
	// External data binding demonstration
	// ========================================
	// You can listen for changes on bound values
	// to perform side effects.
	boundName.AddListener(binding.NewDataListener(func() {
		val, _ := boundName.Get()
		fmt.Printf("Bound name changed to: %s\n", val)
	}))

	// ========================================
	// List Widget
	// ========================================
	// List displays a scrollable list of items.
	// It uses a virtual rendering model — only visible
	// items are created, making it efficient for large lists.
	items := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}

	list := widget.NewList(
		// Length: how many items
		func() int { return len(items) },
		// CreateItem: creates a template widget
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		// UpdateItem: populates each item
		func(index widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(items[index])
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		statusLabel.SetText(fmt.Sprintf("Status: Selected %s", items[id]))
		fmt.Printf("List item selected: %s (index %d)\n", items[id], id)
	}

	// ========================================
	// Form Widget
	// ========================================
	// Form provides a structured layout for labeled inputs
	// with built-in Submit and Cancel buttons.
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("you@example.com")
	emailEntry.Validator = func(s string) error {
		if len(s) > 0 && !containsAt(s) {
			return fmt.Errorf("invalid email address")
		}
		return nil
	}

	ageEntry := widget.NewEntry()
	ageEntry.SetPlaceHolder("25")
	ageEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}
		age, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("age must be a number")
		}
		if age < 0 || age > 150 {
			return fmt.Errorf("age must be 0-150")
		}
		return nil
	}

	form := widget.NewForm(
		widget.NewFormItem("Email", emailEntry),
		widget.NewFormItem("Age", ageEntry),
	)
	form.OnSubmit = func() {
		statusLabel.SetText(fmt.Sprintf("Status: Form submitted — %s, age %s",
			emailEntry.Text, ageEntry.Text))
	}
	form.OnCancel = func() {
		emailEntry.SetText("")
		ageEntry.SetText("")
		statusLabel.SetText("Status: Form cancelled")
	}

	// ========================================
	// Assemble the UI with Tabs
	// ========================================
	// Tabs organize content into switchable panels.
	// AppTabs provides top-level navigation.
	tabs := container.NewAppTabs(
		container.NewTabItem("Input", container.NewVBox(
			widget.NewLabelWithStyle("Input Widgets", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			widget.NewLabel("Name:"),
			nameEntry,
			widget.NewLabel("Password:"),
			passwordEntry,
			widget.NewLabel("Message:"),
			multiEntry,
			container.NewHBox(submitButton, clearButton),
		)),
		container.NewTabItem("Selection", container.NewVBox(
			widget.NewLabelWithStyle("Selection Widgets", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			agreeCheck,
			darkModeCheck,
			widget.NewSeparator(),
			widget.NewLabel("Priority:"),
			priorityRadio,
			widget.NewSeparator(),
			widget.NewLabel("Category:"),
			categorySelect,
			widget.NewSeparator(),
			sliderValueLabel,
			fontSlider,
		)),
		container.NewTabItem("Progress", container.NewVBox(
			widget.NewLabelWithStyle("Progress Widgets", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			widget.NewLabel("Determinate:"),
			progress,
			widget.NewLabel("Indeterminate:"),
			infiniteProgress,
			progressButton,
		)),
		container.NewTabItem("Data Binding", container.NewVBox(
			widget.NewLabelWithStyle("Data Binding", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			widget.NewLabel("Bound Entry:"),
			boundEntry,
			boundLabel,
			widget.NewSeparator(),
			widget.NewLabel("Bound Slider:"),
			boundSlider,
			boundValueLabel,
			widget.NewSeparator(),
			boundCheck,
			boundBoolLabel,
		)),
		container.NewTabItem("List", container.NewVBox(
			widget.NewLabelWithStyle("List Widget", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			container.NewGridWrap(fyne.NewSize(400, 200), list),
		)),
		container.NewTabItem("Form", container.NewVBox(
			widget.NewLabelWithStyle("Form Widget", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewSeparator(),
			form,
		)),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// ========================================
	// Main Layout
	// ========================================
	mainContent := container.NewBorder(
		titleLabel,  // top
		statusLabel, // bottom
		nil,         // left
		nil,         // right
		tabs,        // center (fills remaining space)
	)

	myWindow.SetContent(mainContent)

	fmt.Println("Launching Widgets Showcase...")
	fmt.Println("Explore each tab to see different widget types.")
	myWindow.ShowAndRun()
	fmt.Println("Application closed.")
}

// containsAt is a simple helper to check for '@' in email addresses.
func containsAt(s string) bool {
	for _, c := range s {
		if c == '@' {
			return true
		}
	}
	return false
}
