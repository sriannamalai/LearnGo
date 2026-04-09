package main

import "fmt"

func main() {
	// ========================================
	// Mini-Project: Simple Calculator
	// ========================================
	// This project combines everything from Week 1:
	//   - Variables (storing numbers and operators)
	//   - Control flow (switch for operations, if for error handling)
	//   - fmt package (Printf for formatted output, Scan for input)

	fmt.Println("========================================")
	fmt.Println("       Simple Go Calculator")
	fmt.Println("========================================")

	// ========================================
	// Get input from the user
	// ========================================
	// fmt.Scan reads from standard input (the terminal).
	// We pass pointers (&variable) so Scan can modify the variables.
	var num1, num2 float64
	var operator string

	fmt.Print("Enter first number: ")
	fmt.Scan(&num1)

	fmt.Print("Enter operator (+, -, *, /): ")
	fmt.Scan(&operator)

	fmt.Print("Enter second number: ")
	fmt.Scan(&num2)

	// ========================================
	// Perform the calculation using switch
	// ========================================
	// switch is perfect here — cleaner than a chain of if/else.
	// Each case handles one arithmetic operation.
	var result float64
	var valid bool = true

	switch operator {
	case "+":
		result = num1 + num2
	case "-":
		result = num1 - num2
	case "*":
		result = num1 * num2
	case "/":
		// Always check for division by zero!
		// Dividing by zero in math is undefined, and we should
		// catch this before it causes unexpected behavior.
		if num2 == 0 {
			fmt.Println("\nError: Division by zero is not allowed!")
			valid = false
		} else {
			result = num1 / num2
		}
	default:
		// default catches any operator we don't recognize
		fmt.Printf("\nError: Unknown operator %q\n", operator)
		valid = false
	}

	// ========================================
	// Display the result
	// ========================================
	// Only print the result if the calculation was valid.
	// %.2f formats a float to 2 decimal places.
	if valid {
		fmt.Printf("\n  %.2f %s %.2f = %.2f\n", num1, operator, num2, result)
	}

	// ========================================
	// Bonus: Demo mode with hardcoded examples
	// ========================================
	// Let's also show some calculations without needing input,
	// so you can see the calculator logic in action.
	fmt.Println("\n========================================")
	fmt.Println("  Demo: Hardcoded Examples")
	fmt.Println("========================================")

	// A simple helper approach using just variables and a loop
	// (We'll learn functions properly in Week 2!)
	operations := []struct {
		a  float64
		op string
		b  float64
	}{
		{10, "+", 5},
		{20, "-", 8},
		{7, "*", 6},
		{100, "/", 3},
		{42, "/", 0}, // This should trigger the division by zero error
	}

	for _, calc := range operations {
		switch calc.op {
		case "+":
			fmt.Printf("  %.0f + %.0f = %.2f\n", calc.a, calc.b, calc.a+calc.b)
		case "-":
			fmt.Printf("  %.0f - %.0f = %.2f\n", calc.a, calc.b, calc.a-calc.b)
		case "*":
			fmt.Printf("  %.0f * %.0f = %.2f\n", calc.a, calc.b, calc.a*calc.b)
		case "/":
			if calc.b == 0 {
				fmt.Printf("  %.0f / %.0f = Error: division by zero!\n", calc.a, calc.b)
			} else {
				fmt.Printf("  %.0f / %.0f = %.2f\n", calc.a, calc.b, calc.a/calc.b)
			}
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("  Thanks for using Go Calculator!")
	fmt.Println("========================================")
}

// Sample output (with input: 15, +, 10):
//
// ========================================
//        Simple Go Calculator
// ========================================
// Enter first number: 15
// Enter operator (+, -, *, /): +
// Enter second number: 10
//
//   15.00 + 10.00 = 25.00
//
// ========================================
//   Demo: Hardcoded Examples
// ========================================
//   10 + 5 = 15.00
//   20 - 8 = 12.00
//   7 * 6 = 42.00
//   100 / 3 = 33.33
//   42 / 0 = Error: division by zero!
//
// ========================================
//   Thanks for using Go Calculator!
// ========================================
