// Package validator demonstrates table-driven tests — the most common
// testing pattern in Go.
//
// To run the tests:
//   cd week09
//   go test -v ./02_table_tests/
//
// To run a specific subtest:
//   go test -v -run "TestValidateEmail/valid_simple_email" ./02_table_tests/
package validator

import (
	"regexp"
	"strings"
	"unicode"
)

// ========================================
// Week 9, Lesson 2: Table-Driven Tests
// ========================================
// These are validation functions that benefit from testing
// with many different inputs — perfect for table-driven tests.

// ========================================
// Email Validation
// ========================================

// ValidateEmail checks if an email address is valid.
// Returns an error message if invalid, or empty string if valid.
func ValidateEmail(email string) string {
	if email == "" {
		return "email is required"
	}
	if len(email) > 254 {
		return "email is too long"
	}
	// Simple regex for email validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	if !matched {
		return "invalid email format"
	}
	return ""
}

// ========================================
// Password Validation
// ========================================

// ValidatePassword checks if a password meets security requirements.
// Requirements: at least 8 chars, one uppercase, one lowercase, one digit.
// Returns a slice of all validation errors (empty if valid).
func ValidatePassword(password string) []string {
	var errors []string

	if len(password) < 8 {
		errors = append(errors, "password must be at least 8 characters")
	}
	if len(password) > 128 {
		errors = append(errors, "password must be at most 128 characters")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}

	if !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	if !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	if !hasDigit {
		errors = append(errors, "password must contain at least one digit")
	}

	return errors
}

// ========================================
// Username Validation
// ========================================

// ValidateUsername checks if a username is valid.
// Rules: 3-20 chars, alphanumeric and underscores only, must start with a letter.
func ValidateUsername(username string) string {
	if username == "" {
		return "username is required"
	}
	if len(username) < 3 {
		return "username must be at least 3 characters"
	}
	if len(username) > 20 {
		return "username must be at most 20 characters"
	}
	if !unicode.IsLetter(rune(username[0])) {
		return "username must start with a letter"
	}
	for _, ch := range username {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return "username can only contain letters, digits, and underscores"
		}
	}
	return ""
}

// ========================================
// Age Validation
// ========================================

// ValidateAge checks if an age value is reasonable.
func ValidateAge(age int) string {
	if age < 0 {
		return "age cannot be negative"
	}
	if age < 13 {
		return "must be at least 13 years old"
	}
	if age > 150 {
		return "age is unrealistic"
	}
	return ""
}

// ========================================
// String Utilities (for testing)
// ========================================

// Slugify converts a string to a URL-friendly slug.
// "Hello World!" -> "hello-world"
func Slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)

	var result strings.Builder
	prevDash := false

	for _, ch := range s {
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			result.WriteRune(ch)
			prevDash = false
		} else if !prevDash && result.Len() > 0 {
			result.WriteRune('-')
			prevDash = true
		}
	}

	// Trim trailing dash
	out := result.String()
	return strings.TrimRight(out, "-")
}
