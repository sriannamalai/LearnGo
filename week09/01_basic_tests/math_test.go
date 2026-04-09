package basictest

import (
	"testing"
)

// ========================================
// Week 9, Lesson 1: Basic Test Functions
// ========================================
// Go test rules:
//   1. Test files end with _test.go
//   2. Test functions start with Test (capital T)
//   3. Test functions take exactly one parameter: *testing.T
//   4. Run with: go test -v ./01_basic_tests/

// ========================================
// 1. Simple test with t.Error
// ========================================
// t.Error reports a failure but continues running the test.
// Use this when you want to check multiple things in one test.

func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Add(2, 3) = %d; want 5", result)
	}

	// You can have multiple checks in one test
	result = Add(-1, 1)
	if result != 0 {
		t.Errorf("Add(-1, 1) = %d; want 0", result)
	}

	result = Add(0, 0)
	if result != 0 {
		t.Errorf("Add(0, 0) = %d; want 0", result)
	}
}

// ========================================
// 2. Test with t.Fatal
// ========================================
// t.Fatal reports a failure AND stops the test immediately.
// Use this when subsequent checks depend on the current one passing.

func TestDivide(t *testing.T) {
	// Test normal division
	result, err := Divide(10, 2)
	if err != nil {
		// If we can't even divide, no point checking the result
		t.Fatalf("Divide(10, 2) returned unexpected error: %v", err)
	}
	if result != 5.0 {
		t.Errorf("Divide(10, 2) = %f; want 5.0", result)
	}

	// Test division by zero
	_, err = Divide(10, 0)
	if err == nil {
		t.Error("Divide(10, 0) expected an error, got nil")
	}
}

// ========================================
// 3. Test with t.Errorf (formatted error messages)
// ========================================
// t.Errorf is like t.Error but with Printf-style formatting.
// Always include: what you called, what you got, what you wanted.

func TestSubtract(t *testing.T) {
	got := Subtract(10, 3)
	want := 7
	if got != want {
		// Good error message format: "FuncName(args) = got; want expected"
		t.Errorf("Subtract(10, 3) = %d; want %d", got, want)
	}
}

func TestMultiply(t *testing.T) {
	got := Multiply(4, 5)
	want := 20
	if got != want {
		t.Errorf("Multiply(4, 5) = %d; want %d", got, want)
	}

	// Test with zero
	got = Multiply(100, 0)
	want = 0
	if got != want {
		t.Errorf("Multiply(100, 0) = %d; want %d", got, want)
	}

	// Test with negative numbers
	got = Multiply(-3, 4)
	want = -12
	if got != want {
		t.Errorf("Multiply(-3, 4) = %d; want %d", got, want)
	}
}

// ========================================
// 4. Testing boolean returns
// ========================================

func TestIsPrime(t *testing.T) {
	// Test known primes
	if !IsPrime(2) {
		t.Error("IsPrime(2) = false; want true")
	}
	if !IsPrime(7) {
		t.Error("IsPrime(7) = false; want true")
	}
	if !IsPrime(13) {
		t.Error("IsPrime(13) = false; want true")
	}

	// Test known non-primes
	if IsPrime(0) {
		t.Error("IsPrime(0) = true; want false")
	}
	if IsPrime(1) {
		t.Error("IsPrime(1) = true; want false")
	}
	if IsPrime(4) {
		t.Error("IsPrime(4) = true; want false")
	}
	if IsPrime(9) {
		t.Error("IsPrime(9) = true; want false")
	}
}

// ========================================
// 5. Testing a sequence of values
// ========================================

func TestFibonacci(t *testing.T) {
	// Test the first several Fibonacci numbers
	expected := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}

	for i, want := range expected {
		got := Fibonacci(i)
		if got != want {
			t.Errorf("Fibonacci(%d) = %d; want %d", i, got, want)
		}
	}
}

// ========================================
// 6. Testing edge cases
// ========================================

func TestAbs(t *testing.T) {
	// Positive number stays positive
	if got := Abs(5); got != 5 {
		t.Errorf("Abs(5) = %d; want 5", got)
	}

	// Negative number becomes positive
	if got := Abs(-5); got != 5 {
		t.Errorf("Abs(-5) = %d; want 5", got)
	}

	// Zero stays zero
	if got := Abs(0); got != 0 {
		t.Errorf("Abs(0) = %d; want 0", got)
	}
}

func TestMax(t *testing.T) {
	if got := Max(3, 5); got != 5 {
		t.Errorf("Max(3, 5) = %d; want 5", got)
	}
	if got := Max(5, 3); got != 5 {
		t.Errorf("Max(5, 3) = %d; want 5", got)
	}
	if got := Max(3, 3); got != 3 {
		t.Errorf("Max(3, 3) = %d; want 3", got)
	}
}

func TestMin(t *testing.T) {
	if got := Min(3, 5); got != 3 {
		t.Errorf("Min(3, 5) = %d; want 3", got)
	}
	if got := Min(5, 3); got != 3 {
		t.Errorf("Min(5, 3) = %d; want 3", got)
	}
	if got := Min(3, 3); got != 3 {
		t.Errorf("Min(3, 3) = %d; want 3", got)
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Test file naming:   xxx_test.go (Go only compiles these during testing)
// Test function name: TestXxx(t *testing.T) (must start with Test + capital letter)
//
// Reporting failures:
//   t.Error(args...)    — report failure, continue test
//   t.Errorf(fmt, ...)  — report failure with formatting, continue test
//   t.Fatal(args...)    — report failure, STOP this test immediately
//   t.Fatalf(fmt, ...)  — report failure with formatting, STOP test
//   t.Log(args...)      — log info (only shown with -v flag)
//   t.Logf(fmt, ...)    — log info with formatting
//
// Running tests:
//   go test              — run tests in current directory
//   go test -v           — verbose output (show all test names + t.Log)
//   go test ./...        — run tests in all subdirectories
//   go test -run TestAdd — run only tests matching "TestAdd"
//
// Error message best practice:
//   t.Errorf("FunctionName(input) = %v; want %v", got, want)
