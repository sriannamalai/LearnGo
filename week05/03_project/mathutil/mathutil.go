// Package mathutil provides mathematical utility functions.
// This demonstrates creating a reusable Go package with exported functions.
package mathutil

import "math"

// ========================================
// GCD computes the Greatest Common Divisor of two integers
// using the Euclidean algorithm.
// ========================================
func GCD(a, b int) int {
	// Ensure positive values
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}

	// Euclidean algorithm: repeatedly replace the larger with the remainder
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// ========================================
// LCM computes the Least Common Multiple of two integers.
// Uses the relationship: LCM(a, b) = |a * b| / GCD(a, b)
// ========================================
func LCM(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}

	// Use absolute values
	result := a / GCD(a, b) * b // divide first to avoid overflow
	if result < 0 {
		return -result
	}
	return result
}

// ========================================
// IsPrime checks whether a number is prime.
// A prime number is greater than 1 and divisible only by 1 and itself.
// ========================================
func IsPrime(n int) bool {
	if n < 2 {
		return false
	}
	if n < 4 {
		return true // 2 and 3 are prime
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}

	// Check divisors up to sqrt(n) using 6k +/- 1 optimization
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// ========================================
// Factorial computes n! (n factorial).
// Returns -1 for negative inputs.
// ========================================
func Factorial(n int) int {
	if n < 0 {
		return -1 // error: factorial is not defined for negative numbers
	}
	if n <= 1 {
		return 1
	}

	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

// ========================================
// PrimesUpTo returns all prime numbers up to (and including) max.
// ========================================
func PrimesUpTo(max int) []int {
	var primes []int
	for i := 2; i <= max; i++ {
		if IsPrime(i) {
			primes = append(primes, i)
		}
	}
	return primes
}

// ========================================
// Fibonacci returns the first n Fibonacci numbers.
// ========================================
func Fibonacci(n int) []int {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []int{0}
	}

	fibs := make([]int, n)
	fibs[0] = 0
	fibs[1] = 1
	for i := 2; i < n; i++ {
		fibs[i] = fibs[i-1] + fibs[i-2]
	}
	return fibs
}

// ========================================
// Abs returns the absolute value of an integer.
// ========================================
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ========================================
// Sqrt returns the integer square root (floor of the square root).
// ========================================
func Sqrt(n int) int {
	if n < 0 {
		return -1
	}
	return int(math.Sqrt(float64(n)))
}
