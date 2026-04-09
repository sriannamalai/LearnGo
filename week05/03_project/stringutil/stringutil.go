// Package stringutil provides string utility functions.
// This demonstrates creating a reusable Go package with exported functions.
package stringutil

import (
	"strings"
	"unicode"
)

// ========================================
// Reverse returns the input string reversed, handling Unicode correctly.
// ========================================
func Reverse(s string) string {
	// Convert to runes to handle multi-byte characters properly
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ========================================
// IsPalindrome checks if a string reads the same forwards and backwards.
// Ignores case and non-alphanumeric characters.
// ========================================
func IsPalindrome(s string) bool {
	// Clean the string: lowercase and keep only letters/digits
	var cleaned []rune
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cleaned = append(cleaned, r)
		}
	}

	// Compare from both ends toward the middle
	for i, j := 0, len(cleaned)-1; i < j; i, j = i+1, j-1 {
		if cleaned[i] != cleaned[j] {
			return false
		}
	}
	return true
}

// ========================================
// WordCount counts the number of words in a string.
// Words are separated by whitespace.
// ========================================
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// ========================================
// Capitalize capitalizes the first letter of each word in a string.
// ========================================
func Capitalize(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			runes := []rune(word)
			runes[0] = unicode.ToUpper(runes[0])
			words[i] = string(runes)
		}
	}
	return strings.Join(words, " ")
}

// ========================================
// CountChar counts how many times a character appears in a string.
// ========================================
func CountChar(s string, ch rune) int {
	count := 0
	for _, r := range s {
		if r == ch {
			count++
		}
	}
	return count
}

// ========================================
// Truncate shortens a string to maxLen characters, adding "..." if truncated.
// ========================================
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// ========================================
// IsAnagram checks if two strings are anagrams of each other.
// Ignores case and spaces.
// ========================================
func IsAnagram(a, b string) bool {
	// Count character frequencies for both strings
	countA := charFrequency(a)
	countB := charFrequency(b)

	// Compare frequency maps
	if len(countA) != len(countB) {
		return false
	}
	for ch, count := range countA {
		if countB[ch] != count {
			return false
		}
	}
	return true
}

// charFrequency returns a map of lowercase letter frequencies (unexported helper)
func charFrequency(s string) map[rune]int {
	freq := make(map[rune]int)
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) {
			freq[r]++
		}
	}
	return freq
}
