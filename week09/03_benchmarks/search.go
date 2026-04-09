// Package search demonstrates benchmarking in Go.
//
// To run the tests:
//   cd week09
//   go test -v ./03_benchmarks/
//
// To run benchmarks:
//   go test -bench=. ./03_benchmarks/
//   go test -bench=. -benchmem ./03_benchmarks/
//
// To run benchmarks with specific count:
//   go test -bench=. -count=5 ./03_benchmarks/
//
// To check test coverage:
//   go test -cover ./03_benchmarks/
//   go test -coverprofile=coverage.out ./03_benchmarks/
//   go tool cover -html=coverage.out
package search

import (
	"sort"
	"strings"
)

// ========================================
// Week 9, Lesson 3: Benchmarks
// ========================================
// These are search and string functions designed to demonstrate
// benchmarking. We provide multiple implementations of the same
// operation so you can compare their performance.

// ========================================
// Linear Search vs Binary Search
// ========================================

// LinearSearch searches for target in an unsorted slice.
// Returns the index if found, -1 if not found.
// Time complexity: O(n)
func LinearSearch(data []int, target int) int {
	for i, v := range data {
		if v == target {
			return i
		}
	}
	return -1
}

// BinarySearch searches for target in a SORTED slice.
// Returns the index if found, -1 if not found.
// Time complexity: O(log n)
func BinarySearch(data []int, target int) int {
	low, high := 0, len(data)-1

	for low <= high {
		mid := low + (high-low)/2 // Avoids integer overflow

		if data[mid] == target {
			return mid
		} else if data[mid] < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	return -1
}

// BinarySearchStdlib uses sort.SearchInts from the standard library.
// Returns the index if found, -1 if not found.
func BinarySearchStdlib(data []int, target int) int {
	i := sort.SearchInts(data, target)
	if i < len(data) && data[i] == target {
		return i
	}
	return -1
}

// ========================================
// String Concatenation Methods
// ========================================
// These demonstrate different ways to build strings in Go.
// Benchmarks reveal which approach is fastest.

// ConcatPlus builds a string using the + operator.
// This creates a new string on each iteration (slow for many items).
func ConcatPlus(items []string) string {
	result := ""
	for i, item := range items {
		if i > 0 {
			result += ", "
		}
		result += item
	}
	return result
}

// ConcatBuilder builds a string using strings.Builder.
// This is the most efficient way to build strings in Go.
func ConcatBuilder(items []string) string {
	var builder strings.Builder
	for i, item := range items {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(item)
	}
	return builder.String()
}

// ConcatJoin builds a string using strings.Join.
// Simple and efficient for joining slices.
func ConcatJoin(items []string) string {
	return strings.Join(items, ", ")
}

// ========================================
// Map vs Slice Lookup
// ========================================
// Demonstrates the performance difference between
// searching a slice vs looking up a map key.

// SliceContains checks if a string exists in a slice. O(n).
func SliceContains(data []string, target string) bool {
	for _, v := range data {
		if v == target {
			return true
		}
	}
	return false
}

// MapContains checks if a string exists as a map key. O(1) average.
func MapContains(data map[string]bool, target string) bool {
	return data[target]
}

// ========================================
// Helper: Generate test data
// ========================================

// GenerateSortedSlice creates a sorted slice of integers [0, 1, 2, ..., n-1]
func GenerateSortedSlice(n int) []int {
	data := make([]int, n)
	for i := range data {
		data[i] = i
	}
	return data
}

// GenerateStringSlice creates a slice of string items
func GenerateStringSlice(n int) []string {
	items := make([]string, n)
	for i := range items {
		items[i] = strings.Repeat("item", 1) + string(rune('A'+i%26))
	}
	return items
}

// SliceToMap converts a string slice to a map for lookup
func SliceToMap(data []string) map[string]bool {
	m := make(map[string]bool, len(data))
	for _, v := range data {
		m[v] = true
	}
	return m
}
