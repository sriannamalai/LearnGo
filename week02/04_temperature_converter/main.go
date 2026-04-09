package main

import (
	"errors"
	"fmt"
	"strings"
)

// ========================================
// Mini-Project: Temperature Converter
// ========================================
// This project combines everything from Week 2:
//   - Functions with parameters and return values
//   - Multiple return values (result + error)
//   - Error handling with errors.New and fmt.Errorf
//   - Custom error types
//   - Variadic functions
//
// Convert between Celsius, Fahrenheit, and Kelvin.

// ========================================
// Custom error type for invalid temperatures
// ========================================
// Absolute zero is the lowest possible temperature.
// We use a custom error to provide helpful context.
type BelowAbsoluteZeroError struct {
	Value float64
	Scale string
	Limit float64
}

func (e *BelowAbsoluteZeroError) Error() string {
	return fmt.Sprintf("%.2f%s is below absolute zero (minimum: %.2f%s)",
		e.Value, e.Scale, e.Limit, e.Scale)
}

// Sentinel error for unsupported scale
var ErrUnknownScale = errors.New("unknown temperature scale")

// ========================================
// Core conversion functions
// ========================================
// Each function returns (result, error) following Go conventions.

// CelsiusToFahrenheit converts Celsius to Fahrenheit.
// Formula: F = C * 9/5 + 32
func CelsiusToFahrenheit(c float64) (float64, error) {
	if c < -273.15 {
		return 0, &BelowAbsoluteZeroError{Value: c, Scale: "C", Limit: -273.15}
	}
	return c*9.0/5.0 + 32, nil
}

// FahrenheitToCelsius converts Fahrenheit to Celsius.
// Formula: C = (F - 32) * 5/9
func FahrenheitToCelsius(f float64) (float64, error) {
	if f < -459.67 {
		return 0, &BelowAbsoluteZeroError{Value: f, Scale: "F", Limit: -459.67}
	}
	return (f - 32) * 5.0 / 9.0, nil
}

// CelsiusToKelvin converts Celsius to Kelvin.
// Formula: K = C + 273.15
func CelsiusToKelvin(c float64) (float64, error) {
	if c < -273.15 {
		return 0, &BelowAbsoluteZeroError{Value: c, Scale: "C", Limit: -273.15}
	}
	return c + 273.15, nil
}

// KelvinToCelsius converts Kelvin to Celsius.
// Formula: C = K - 273.15
func KelvinToCelsius(k float64) (float64, error) {
	if k < 0 {
		return 0, &BelowAbsoluteZeroError{Value: k, Scale: "K", Limit: 0}
	}
	return k - 273.15, nil
}

// FahrenheitToKelvin converts Fahrenheit to Kelvin.
// We chain two conversions: F -> C -> K
func FahrenheitToKelvin(f float64) (float64, error) {
	c, err := FahrenheitToCelsius(f)
	if err != nil {
		return 0, err
	}
	return CelsiusToKelvin(c)
}

// KelvinToFahrenheit converts Kelvin to Fahrenheit.
// We chain two conversions: K -> C -> F
func KelvinToFahrenheit(k float64) (float64, error) {
	c, err := KelvinToCelsius(k)
	if err != nil {
		return 0, err
	}
	return CelsiusToFahrenheit(c)
}

// ========================================
// Universal conversion function
// ========================================
// Convert takes a value and converts between any two scales.
// The scale names are normalized to uppercase single letters.
func Convert(value float64, from, to string) (float64, error) {
	// Normalize input to uppercase first letter
	from = strings.ToUpper(strings.TrimSpace(from))
	to = strings.ToUpper(strings.TrimSpace(to))

	// Build a key like "C->F" to select the right converter
	key := from + "->" + to

	switch key {
	case "C->F":
		return CelsiusToFahrenheit(value)
	case "F->C":
		return FahrenheitToCelsius(value)
	case "C->K":
		return CelsiusToKelvin(value)
	case "K->C":
		return KelvinToCelsius(value)
	case "F->K":
		return FahrenheitToKelvin(value)
	case "K->F":
		return KelvinToFahrenheit(value)
	case "C->C", "F->F", "K->K":
		// Same scale — return the value unchanged
		return value, nil
	default:
		return 0, fmt.Errorf("cannot convert from %q to %q: %w", from, to, ErrUnknownScale)
	}
}

// ========================================
// Batch conversion (variadic function)
// ========================================
// ConvertAll converts a list of temperatures from one scale to another.
// Returns a slice of results and the first error encountered.
func ConvertAll(from, to string, values ...float64) ([]float64, error) {
	results := make([]float64, 0, len(values))
	for _, v := range values {
		converted, err := Convert(v, from, to)
		if err != nil {
			return results, fmt.Errorf("converting %.2f: %w", v, err)
		}
		results = append(results, converted)
	}
	return results, nil
}

// ========================================
// Named return value function
// ========================================
// FormatTemperature returns a human-readable temperature string.
func FormatTemperature(value float64, scale string) (formatted string) {
	scale = strings.ToUpper(strings.TrimSpace(scale))
	switch scale {
	case "C":
		formatted = fmt.Sprintf("%.2f°C", value)
	case "F":
		formatted = fmt.Sprintf("%.2f°F", value)
	case "K":
		formatted = fmt.Sprintf("%.2f K", value) // Kelvin doesn't use ° symbol
	default:
		formatted = fmt.Sprintf("%.2f (unknown scale)", value)
	}
	return // naked return of named value
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Temperature Converter")
	fmt.Println("========================================")

	// ========================================
	// Interactive conversion from user input
	// ========================================
	fmt.Println("\n--- Interactive Mode ---")
	var value float64
	var from, to string

	fmt.Print("Enter temperature value: ")
	fmt.Scan(&value)
	fmt.Print("Enter source scale (C/F/K): ")
	fmt.Scan(&from)
	fmt.Print("Enter target scale (C/F/K): ")
	fmt.Scan(&to)

	result, err := Convert(value, from, to)
	if err != nil {
		var belowZero *BelowAbsoluteZeroError
		if errors.As(err, &belowZero) {
			fmt.Printf("\n  Physics error: %s\n", belowZero)
		} else if errors.Is(err, ErrUnknownScale) {
			fmt.Printf("\n  Input error: %s\n", err)
			fmt.Println("  Valid scales are: C (Celsius), F (Fahrenheit), K (Kelvin)")
		} else {
			fmt.Printf("\n  Error: %s\n", err)
		}
	} else {
		fmt.Printf("\n  %s = %s\n",
			FormatTemperature(value, from),
			FormatTemperature(result, to))
	}

	// ========================================
	// Demo: Common conversions
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("  Demo: Common Conversions")
	fmt.Println("========================================")

	// A helper to print conversions and handle errors cleanly
	printConversion := func(val float64, from, to string) {
		result, err := Convert(val, from, to)
		if err != nil {
			fmt.Printf("  %s -> Error: %s\n", FormatTemperature(val, from), err)
		} else {
			fmt.Printf("  %s = %s\n", FormatTemperature(val, from), FormatTemperature(result, to))
		}
	}

	// Water freezing/boiling points
	fmt.Println("\n  Water's key temperatures:")
	printConversion(0, "C", "F")    // Freezing
	printConversion(100, "C", "F")  // Boiling
	printConversion(0, "C", "K")    // Freezing in Kelvin
	printConversion(100, "C", "K")  // Boiling in Kelvin

	// Human body temperature
	fmt.Println("\n  Body temperature:")
	printConversion(98.6, "F", "C")
	printConversion(98.6, "F", "K")

	// Absolute zero
	fmt.Println("\n  Absolute zero:")
	printConversion(0, "K", "C")
	printConversion(0, "K", "F")

	// Room temperature
	fmt.Println("\n  Room temperature (72°F):")
	printConversion(72, "F", "C")
	printConversion(72, "F", "K")

	// ========================================
	// Demo: Error handling in action
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("  Demo: Error Handling")
	fmt.Println("========================================")

	// Below absolute zero
	fmt.Println("\n  Below absolute zero:")
	printConversion(-300, "C", "F")
	printConversion(-500, "F", "K")
	printConversion(-10, "K", "C")

	// Unknown scale
	fmt.Println("\n  Unknown scale:")
	printConversion(100, "X", "C")

	// ========================================
	// Demo: Batch conversion (variadic)
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("  Demo: Batch Conversion")
	fmt.Println("========================================")

	// Convert a week's high temperatures from F to C
	fmt.Println("\n  Weekly highs (°F -> °C):")
	weeklyHighsF := []float64{72, 75, 68, 80, 77, 85, 71}
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

	// Using the variadic ConvertAll function with slice expansion
	results, err := ConvertAll("F", "C", weeklyHighsF...)
	if err != nil {
		fmt.Println("  Error:", err)
	} else {
		for i, c := range results {
			fmt.Printf("    %s: %s = %s\n", days[i],
				FormatTemperature(weeklyHighsF[i], "F"),
				FormatTemperature(c, "C"))
		}
	}

	// ========================================
	// Demo: Batch with an error mid-stream
	// ========================================
	fmt.Println("\n  Batch with invalid temperature:")
	_, err = ConvertAll("K", "C", 300, 200, -5, 100)
	if err != nil {
		fmt.Println("  Stopped:", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("  Concepts Used:")
	fmt.Println("  - Functions with parameters & returns")
	fmt.Println("  - Multiple return values (value, error)")
	fmt.Println("  - Custom error types (BelowAbsoluteZeroError)")
	fmt.Println("  - Sentinel errors (ErrUnknownScale)")
	fmt.Println("  - errors.Is and errors.As")
	fmt.Println("  - Error wrapping with %w")
	fmt.Println("  - Variadic functions (ConvertAll)")
	fmt.Println("  - Named return values (FormatTemperature)")
	fmt.Println("  - Function values (printConversion)")
	fmt.Println("========================================")
}

// Sample output (with input: 100, C, F):
//
// ========================================
//   Temperature Converter
// ========================================
//
// --- Interactive Mode ---
// Enter temperature value: 100
// Enter source scale (C/F/K): C
// Enter target scale (C/F/K): F
//
//   100.00°C = 212.00°F
//
// ========================================
//   Demo: Common Conversions
// ========================================
//
//   Water's key temperatures:
//   0.00°C = 32.00°F
//   100.00°C = 212.00°F
//   0.00°C = 273.15 K
//   100.00°C = 373.15 K
//
//   Body temperature:
//   98.60°F = 37.00°C
//   98.60°F = 310.15 K
