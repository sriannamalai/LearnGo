// Package basictest demonstrates basic Go testing patterns.
//
// To run the tests:
//   cd week09
//   go test -v ./01_basic_tests/
//
// To run a specific test:
//   go test -v -run TestAdd ./01_basic_tests/
package basictest

import (
	"errors"
	"math"
)

// ========================================
// Week 9, Lesson 1: Basic Tests
// ========================================
// These are simple math functions that we'll test.
// Notice: this is NOT package main — it's package basictest.
// Test files in the same directory share the same package
// and can access all exported AND unexported symbols.

// ========================================
// Basic Arithmetic Functions
// ========================================

// Add returns the sum of two integers.
func Add(a, b int) int {
	return a + b
}

// Subtract returns the difference of two integers.
func Subtract(a, b int) int {
	return a - b
}

// Multiply returns the product of two integers.
func Multiply(a, b int) int {
	return a * b
}

// Divide returns the quotient of two floats.
// Returns an error if the divisor is zero.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// ========================================
// Slightly more complex functions
// ========================================

// Abs returns the absolute value of an integer.
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Max returns the larger of two integers.
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min returns the smaller of two integers.
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// IsPrime returns true if n is a prime number.
func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n == 2 {
		return true
	}
	if n%2 == 0 {
		return false
	}
	for i := 3; i <= int(math.Sqrt(float64(n))); i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// Fibonacci returns the nth Fibonacci number (0-indexed).
// Fibonacci(0) = 0, Fibonacci(1) = 1, Fibonacci(2) = 1, etc.
func Fibonacci(n int) int {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}
