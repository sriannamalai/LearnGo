package main

// ========================================
// Week 20 — Lesson 3: Layouts
// ========================================
// This lesson covers:
//   - Fyne layout system: how containers arrange widgets
//   - VBox and HBox — vertical and horizontal stacking
//   - GridWithColumns — grid layout
//   - Border — header/footer/sidebar layout
//   - Center, Max, Padded — utility layouts
//   - Combining layouts for complex UIs
//   - Theming basics
//
// Run:
//   go run .

import (
	"fmt"
	"image/color"

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
	fmt.Println("  Week 20 - Lesson 3: Layouts")
	fmt.Println("========================================")
	fmt.Println()

	myApp := app.New()
	myWindow := myApp.NewWindow("Fyne Layouts Showcase")
	myWindow.Resize(fyne.NewSize(700, 600))
	myWindow.CenterOnScreen()

	// ========================================
	// VBox Layout — Vertical Stacking
	// ========================================
	// VBox arranges children vertically, one below another.
	// Each child takes its minimum height; the last item
	// can expand if you use layout.NewSpacer().
	vboxDemo := container.NewVBox(
		widget.NewLabelWithStyle("VBox Layout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		makeColoredBox("Item 1", color.NRGBA{R: 200, G: 100, B: 100, A: 255}),
		makeColoredBox("Item 2", color.NRGBA{R: 100, G: 200, B: 100, A: 255}),
		makeColoredBox("Item 3", color.NRGBA{R: 100, G: 100, B: 200, A: 255}),
		layout.NewSpacer(), // Pushes items above to the top
		widget.NewLabel("(Spacer above pushes items to top)"),
	)

	// ========================================
	// HBox Layout — Horizontal Stacking
	// ========================================
	// HBox arranges children horizontally, side by side.
	// Each child takes its minimum width.
	hboxDemo := container.NewVBox(
		widget.NewLabelWithStyle("HBox Layout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewHBox(
			makeColoredBox("Left", color.NRGBA{R: 200, G: 100, B: 100, A: 255}),
			makeColoredBox("Center", color.NRGBA{R: 100, G: 200, B: 100, A: 255}),
			makeColoredBox("Right", color.NRGBA{R: 100, G: 100, B: 200, A: 255}),
		),
		widget.NewSeparator(),
		widget.NewLabel("With spacer (items spread out):"),
		container.NewHBox(
			makeColoredBox("Left", color.NRGBA{R: 200, G: 150, B: 100, A: 255}),
			layout.NewSpacer(), // Spacer pushes items apart
			makeColoredBox("Right", color.NRGBA{R: 100, G: 150, B: 200, A: 255}),
		),
		widget.NewSeparator(),
		widget.NewLabel("Centered with spacers on both sides:"),
		container.NewHBox(
			layout.NewSpacer(),
			makeColoredBox("Centered", color.NRGBA{R: 150, G: 200, B: 150, A: 255}),
			layout.NewSpacer(),
		),
	)

	// ========================================
	// Grid Layout — Columns and Rows
	// ========================================
	// NewGridWithColumns arranges items in a grid with
	// a fixed number of columns. Rows are added as needed.
	// All cells are the same size.
	gridDemo := container.NewVBox(
		widget.NewLabelWithStyle("Grid Layout (3 columns)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWithColumns(3,
			makeColoredBox("1", color.NRGBA{R: 200, G: 100, B: 100, A: 255}),
			makeColoredBox("2", color.NRGBA{R: 100, G: 200, B: 100, A: 255}),
			makeColoredBox("3", color.NRGBA{R: 100, G: 100, B: 200, A: 255}),
			makeColoredBox("4", color.NRGBA{R: 200, G: 200, B: 100, A: 255}),
			makeColoredBox("5", color.NRGBA{R: 200, G: 100, B: 200, A: 255}),
			makeColoredBox("6", color.NRGBA{R: 100, G: 200, B: 200, A: 255}),
		),
		widget.NewSeparator(),
		widget.NewLabel("Grid with rows (2 rows):"),
		container.NewGridWithRows(2,
			makeColoredBox("Row1-A", color.NRGBA{R: 180, G: 120, B: 100, A: 255}),
			makeColoredBox("Row1-B", color.NRGBA{R: 120, G: 180, B: 100, A: 255}),
			makeColoredBox("Row2-A", color.NRGBA{R: 100, G: 120, B: 180, A: 255}),
			makeColoredBox("Row2-B", color.NRGBA{R: 180, G: 180, B: 100, A: 255}),
		),
	)

	// ========================================
	// Border Layout — Header/Footer/Sidebar
	// ========================================
	// NewBorder places up to 4 items at the edges
	// (top, bottom, left, right) and one item in the center.
	// The center item fills all remaining space.
	// Pass nil for any edge you don't need.
	borderHeader := widget.NewLabelWithStyle(
		"Header (Top)",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)
	borderFooter := widget.NewLabel("Footer (Bottom)")
	borderLeft := container.NewVBox(
		widget.NewLabel("Sidebar"),
		widget.NewButton("Nav 1", func() {}),
		widget.NewButton("Nav 2", func() {}),
	)
	borderCenter := container.NewCenter(
		widget.NewLabel("Center content fills remaining space"),
	)

	borderDemo := container.NewVBox(
		widget.NewLabelWithStyle("Border Layout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(600, 200),
			container.NewBorder(
				borderHeader, // top
				borderFooter, // bottom
				borderLeft,   // left
				nil,          // right (unused)
				borderCenter, // center
			),
		),
	)

	// ========================================
	// Center Layout
	// ========================================
	// NewCenter places its child in the center of the
	// available space, both horizontally and vertically.
	centerDemo := container.NewVBox(
		widget.NewLabelWithStyle("Center Layout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(400, 150),
			container.NewCenter(
				widget.NewButton("I'm centered!", func() {}),
			),
		),
	)

	// ========================================
	// Max Layout
	// ========================================
	// NewMax (also known as Stack) layers children on top
	// of each other, each filling the full container size.
	// Useful for overlays, backgrounds, or card-like UIs.
	background := canvas.NewRectangle(color.NRGBA{R: 50, G: 50, B: 80, A: 255})
	foreground := container.NewCenter(
		widget.NewLabelWithStyle(
			"Overlaid on background",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true, Italic: true},
		),
	)

	maxDemo := container.NewVBox(
		widget.NewLabelWithStyle("Max (Stack) Layout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(400, 100),
			container.NewStack(background, foreground),
		),
	)

	// ========================================
	// Padded Layout
	// ========================================
	// NewPadded adds consistent padding around its child.
	// This is useful for adding breathing room around content.
	paddedInner := widget.NewLabel("This content has padding around it")
	paddedBg := canvas.NewRectangle(color.NRGBA{R: 70, G: 130, B: 180, A: 255})

	paddedDemo := container.NewVBox(
		widget.NewLabelWithStyle("Padded Layout", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(400, 80),
			container.NewStack(
				paddedBg,
				container.NewPadded(paddedInner),
			),
		),
	)

	// ========================================
	// Complex Layout Example
	// ========================================
	// Real applications combine multiple layouts.
	// Here we build a typical app structure:
	// - Top toolbar with buttons
	// - Left sidebar with navigation
	// - Center content area
	// - Bottom status bar
	toolbar := container.NewHBox(
		widget.NewButtonWithIcon("New", theme.DocumentCreateIcon(), func() {}),
		widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), func() {}),
		widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), func() {}),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {}),
	)

	sidebar := container.NewVBox(
		widget.NewLabelWithStyle("Navigation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewButton("Dashboard", func() {}),
		widget.NewButton("Documents", func() {}),
		widget.NewButton("Settings", func() {}),
		layout.NewSpacer(),
		widget.NewLabel("v1.0.0"),
	)

	mainArea := container.NewCenter(
		container.NewVBox(
			widget.NewLabelWithStyle("Main Content Area", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("This is where primary content goes"),
		),
	)

	statusBar := container.NewHBox(
		widget.NewLabel("Ready"),
		layout.NewSpacer(),
		widget.NewLabel("Ln 1, Col 1"),
		widget.NewSeparator(),
		widget.NewLabel("UTF-8"),
	)

	complexDemo := container.NewVBox(
		widget.NewLabelWithStyle("Complex Layout (App Shell)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewGridWrap(fyne.NewSize(600, 250),
			container.NewBorder(
				toolbar,   // top
				statusBar, // bottom
				sidebar,   // left
				nil,       // right
				mainArea,  // center
			),
		),
	)

	// ========================================
	// Theming Basics
	// ========================================
	// Fyne apps automatically support light and dark themes.
	// You can set the theme programmatically:
	//
	//   myApp.Settings().SetTheme(theme.DarkTheme())
	//   myApp.Settings().SetTheme(theme.LightTheme())
	//
	// Custom themes implement the fyne.Theme interface:
	//
	//   type MyTheme struct{}
	//   func (m MyTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color { ... }
	//   func (m MyTheme) Font(style fyne.TextStyle) fyne.Resource { ... }
	//   func (m MyTheme) Icon(name fyne.ThemeIconName) fyne.Resource { ... }
	//   func (m MyTheme) Size(name fyne.ThemeSizeName) float32 { ... }
	//
	// Theme colors include:
	//   theme.ColorNamePrimary — accent color
	//   theme.ColorNameBackground — window background
	//   theme.ColorNameButton — button background
	//   theme.ColorNameForeground — text color

	currentTheme := "System Default"
	themeLabel := widget.NewLabel(fmt.Sprintf("Current theme: %s", currentTheme))

	themingDemo := container.NewVBox(
		widget.NewLabelWithStyle("Theming", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		themeLabel,
		container.NewHBox(
			widget.NewButton("Dark Theme", func() {
				myApp.Settings().SetTheme(theme.DarkTheme())
				themeLabel.SetText("Current theme: Dark")
			}),
			widget.NewButton("Light Theme", func() {
				myApp.Settings().SetTheme(theme.LightTheme())
				themeLabel.SetText("Current theme: Light")
			}),
		),
		widget.NewSeparator(),
		widget.NewLabel("Theme icons from Fyne:"),
		container.NewHBox(
			widget.NewIcon(theme.HomeIcon()),
			widget.NewIcon(theme.SettingsIcon()),
			widget.NewIcon(theme.SearchIcon()),
			widget.NewIcon(theme.InfoIcon()),
			widget.NewIcon(theme.WarningIcon()),
			widget.NewIcon(theme.ErrorIcon()),
			widget.NewIcon(theme.DeleteIcon()),
			widget.NewIcon(theme.MailComposeIcon()),
		),
	)

	// ========================================
	// Assemble into Tabs
	// ========================================
	tabs := container.NewAppTabs(
		container.NewTabItem("VBox", container.NewScroll(vboxDemo)),
		container.NewTabItem("HBox", container.NewScroll(hboxDemo)),
		container.NewTabItem("Grid", container.NewScroll(gridDemo)),
		container.NewTabItem("Border", container.NewScroll(borderDemo)),
		container.NewTabItem("Center/Max/Pad", container.NewScroll(
			container.NewVBox(centerDemo, maxDemo, paddedDemo),
		)),
		container.NewTabItem("Complex", container.NewScroll(complexDemo)),
		container.NewTabItem("Theming", container.NewScroll(themingDemo)),
	)

	myWindow.SetContent(tabs)

	fmt.Println("Launching Layouts Showcase...")
	fmt.Println("Explore each tab to see different layout strategies.")
	myWindow.ShowAndRun()
	fmt.Println("Application closed.")
}

// ========================================
// Helper: Create a colored box with a label
// ========================================
// makeColoredBox creates a small colored rectangle with
// centered text, useful for visualizing layout behavior.
func makeColoredBox(label string, c color.Color) fyne.CanvasObject {
	bg := canvas.NewRectangle(c)
	bg.SetMinSize(fyne.NewSize(80, 40))

	text := canvas.NewText(label, color.White)
	text.Alignment = fyne.TextAlignCenter
	text.TextStyle = fyne.TextStyle{Bold: true}

	return container.NewStack(bg, container.NewCenter(text))
}
