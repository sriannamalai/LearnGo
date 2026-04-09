package validator

import (
	"testing"
)

// ========================================
// Week 9, Lesson 2: Table-Driven Tests
// ========================================
// Table-driven tests are THE standard Go testing pattern.
// Instead of writing a separate test function for each case,
// you define a slice of test cases and loop over them.

// ========================================
// 1. Basic Table-Driven Test
// ========================================
// The pattern:
//   1. Define a slice of structs (the "table") with inputs and expected outputs
//   2. Loop over them
//   3. Use t.Run for subtests (each case gets its own name)

func TestValidateEmail(t *testing.T) {
	// The test table: a slice of anonymous structs
	tests := []struct {
		name  string // descriptive name for this test case
		email string // input
		want  string // expected output ("" means valid)
	}{
		// ---- Valid emails ----
		{name: "valid simple email", email: "user@example.com", want: ""},
		{name: "valid with dots", email: "first.last@example.com", want: ""},
		{name: "valid with plus", email: "user+tag@example.com", want: ""},
		{name: "valid with subdomain", email: "user@mail.example.com", want: ""},

		// ---- Invalid emails ----
		{name: "empty email", email: "", want: "email is required"},
		{name: "missing @", email: "userexample.com", want: "invalid email format"},
		{name: "missing domain", email: "user@", want: "invalid email format"},
		{name: "missing local part", email: "@example.com", want: "invalid email format"},
		{name: "double @", email: "user@@example.com", want: "invalid email format"},
		{name: "missing TLD", email: "user@example", want: "invalid email format"},
		{name: "spaces in email", email: "user @example.com", want: "invalid email format"},
	}

	// Loop over the table and run each case as a subtest
	for _, tt := range tests {
		// t.Run creates a subtest with the given name.
		// Each subtest runs independently and shows up in test output.
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got != tt.want {
				t.Errorf("ValidateEmail(%q) = %q; want %q", tt.email, got, tt.want)
			}
		})
	}
}

// ========================================
// 2. Table test with multiple return values
// ========================================

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantCount int // how many errors we expect
	}{
		{name: "valid password", password: "SecurePass1", wantCount: 0},
		{name: "valid complex", password: "MyP@ssw0rd!", wantCount: 0},
		{name: "too short", password: "Ab1", wantCount: 1},          // only "too short"
		{name: "no uppercase", password: "lowercase1!", wantCount: 1}, // missing uppercase
		{name: "no lowercase", password: "UPPERCASE1!", wantCount: 1}, // missing lowercase
		{name: "no digit", password: "NoDigitsHere!", wantCount: 1},   // missing digit
		{name: "empty password", password: "", wantCount: 4},          // too short + no upper + no lower + no digit
		{name: "only numbers", password: "12345678", wantCount: 2},    // no upper + no lower
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidatePassword(tt.password)
			if len(errors) != tt.wantCount {
				t.Errorf("ValidatePassword(%q) returned %d errors; want %d\n  errors: %v",
					tt.password, len(errors), tt.wantCount, errors)
			}
		})
	}
}

// ========================================
// 3. Table test with a test helper
// ========================================
// t.Helper() marks a function as a test helper. When it reports errors,
// Go shows the line number of the CALLER, not the helper itself.

// assertValidationResult is a test helper that checks a validation result.
func assertValidationResult(t *testing.T, funcName, input, got, want string) {
	t.Helper() // This line is crucial! Without it, error locations point here instead of the caller.

	if got != want {
		t.Errorf("%s(%q) = %q; want %q", funcName, input, got, want)
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     string
	}{
		{name: "valid simple", username: "alice", want: ""},
		{name: "valid with numbers", username: "alice123", want: ""},
		{name: "valid with underscore", username: "alice_bob", want: ""},
		{name: "valid minimum length", username: "abc", want: ""},
		{name: "empty username", username: "", want: "username is required"},
		{name: "too short", username: "ab", want: "username must be at least 3 characters"},
		{name: "too long", username: "abcdefghijklmnopqrstu", want: "username must be at most 20 characters"},
		{name: "starts with number", username: "1alice", want: "username must start with a letter"},
		{name: "starts with underscore", username: "_alice", want: "username must start with a letter"},
		{name: "contains space", username: "alice bob", want: "username can only contain letters, digits, and underscores"},
		{name: "contains special char", username: "alice@bob", want: "username can only contain letters, digits, and underscores"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateUsername(tt.username)
			// Using our helper function
			assertValidationResult(t, "ValidateUsername", tt.username, got, tt.want)
		})
	}
}

// ========================================
// 4. Table test for numeric validation
// ========================================

func TestValidateAge(t *testing.T) {
	tests := []struct {
		name string
		age  int
		want string
	}{
		{name: "valid age 25", age: 25, want: ""},
		{name: "valid age 13", age: 13, want: ""},
		{name: "valid age 100", age: 100, want: ""},
		{name: "valid age 150", age: 150, want: ""},
		{name: "negative age", age: -1, want: "age cannot be negative"},
		{name: "too young", age: 12, want: "must be at least 13 years old"},
		{name: "baby", age: 0, want: "must be at least 13 years old"},
		{name: "unrealistic", age: 151, want: "age is unrealistic"},
		{name: "way too old", age: 999, want: "age is unrealistic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateAge(tt.age)
			if got != tt.want {
				t.Errorf("ValidateAge(%d) = %q; want %q", tt.age, got, tt.want)
			}
		})
	}
}

// ========================================
// 5. Table test for string transformations
// ========================================

func TestSlugify(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple words", input: "Hello World", want: "hello-world"},
		{name: "already lowercase", input: "hello world", want: "hello-world"},
		{name: "with punctuation", input: "Hello, World!", want: "hello-world"},
		{name: "multiple spaces", input: "hello   world", want: "hello-world"},
		{name: "leading/trailing spaces", input: "  hello world  ", want: "hello-world"},
		{name: "single word", input: "Hello", want: "hello"},
		{name: "with numbers", input: "Go 1.22 Release", want: "go-1-22-release"},
		{name: "special characters", input: "What's New?", want: "what-s-new"},
		{name: "empty string", input: "", want: ""},
		{name: "only spaces", input: "   ", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q; want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Table-driven test pattern:
//   tests := []struct {
//       name   string
//       input  InputType
//       want   OutputType
//   }{
//       {name: "case 1", input: ..., want: ...},
//       {name: "case 2", input: ..., want: ...},
//   }
//   for _, tt := range tests {
//       t.Run(tt.name, func(t *testing.T) {
//           got := FunctionUnderTest(tt.input)
//           if got != tt.want {
//               t.Errorf(...)
//           }
//       })
//   }
//
// t.Run("name", func(t *testing.T) {...})
//   - Creates a subtest with the given name
//   - Each subtest can be run independently: go test -run "TestFunc/subtest_name"
//   - Shows up clearly in verbose test output
//
// t.Helper()
//   - Marks a function as a test helper
//   - Error reports show the caller's line number, not the helper's
//   - Always call t.Helper() at the start of helper functions
//
// Running subtests:
//   go test -v -run "TestValidateEmail/valid_simple_email" ./02_table_tests/
//   go test -v -run "TestValidateEmail/empty" ./02_table_tests/
