package search

import (
	"testing"
)

// ========================================
// Week 9, Lesson 3: Benchmark Tests
// ========================================
// Benchmark functions start with Benchmark (not Test).
// They take *testing.B (not *testing.T).
// The key: your code must run b.N times — Go adjusts b.N
// to get a stable measurement.
//
// Run benchmarks:
//   go test -bench=. ./03_benchmarks/
//   go test -bench=. -benchmem ./03_benchmarks/
//   go test -bench=BenchmarkSearch -benchmem ./03_benchmarks/

// ========================================
// 1. Regular Tests (for correctness)
// ========================================
// Always test correctness FIRST, then benchmark for performance.

func TestLinearSearch(t *testing.T) {
	data := []int{5, 3, 8, 1, 9, 2, 7}

	tests := []struct {
		name   string
		target int
		want   int
	}{
		{name: "find first", target: 5, want: 0},
		{name: "find middle", target: 1, want: 3},
		{name: "find last", target: 7, want: 6},
		{name: "not found", target: 42, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LinearSearch(data, tt.target)
			if got != tt.want {
				t.Errorf("LinearSearch(data, %d) = %d; want %d", tt.target, got, tt.want)
			}
		})
	}
}

func TestBinarySearch(t *testing.T) {
	data := []int{1, 3, 5, 7, 9, 11, 13}

	tests := []struct {
		name   string
		target int
		want   int
	}{
		{name: "find first", target: 1, want: 0},
		{name: "find middle", target: 7, want: 3},
		{name: "find last", target: 13, want: 6},
		{name: "not found", target: 42, want: -1},
		{name: "not found between", target: 4, want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BinarySearch(data, tt.target)
			if got != tt.want {
				t.Errorf("BinarySearch(data, %d) = %d; want %d", tt.target, got, tt.want)
			}
		})
	}
}

func TestConcatMethods(t *testing.T) {
	items := []string{"Go", "is", "awesome"}
	want := "Go, is, awesome"

	if got := ConcatPlus(items); got != want {
		t.Errorf("ConcatPlus = %q; want %q", got, want)
	}
	if got := ConcatBuilder(items); got != want {
		t.Errorf("ConcatBuilder = %q; want %q", got, want)
	}
	if got := ConcatJoin(items); got != want {
		t.Errorf("ConcatJoin = %q; want %q", got, want)
	}
}

// ========================================
// 2. Basic Benchmark
// ========================================
// A benchmark function:
//   - Starts with Benchmark
//   - Takes *testing.B
//   - Loops b.N times (Go chooses b.N automatically)

func BenchmarkLinearSearch(b *testing.B) {
	data := GenerateSortedSlice(10000)
	target := 9999 // Worst case: last element

	// b.ResetTimer() resets the benchmark timer.
	// Use it after expensive setup that shouldn't be measured.
	b.ResetTimer()

	for b.Loop() {
		LinearSearch(data, target)
	}
}

func BenchmarkBinarySearch(b *testing.B) {
	data := GenerateSortedSlice(10000)
	target := 9999

	b.ResetTimer()

	for b.Loop() {
		BinarySearch(data, target)
	}
}

func BenchmarkBinarySearchStdlib(b *testing.B) {
	data := GenerateSortedSlice(10000)
	target := 9999

	b.ResetTimer()

	for b.Loop() {
		BinarySearchStdlib(data, target)
	}
}

// ========================================
// 3. Benchmarks with Sub-benchmarks
// ========================================
// Just like t.Run for subtests, you can use b.Run for sub-benchmarks.
// This lets you test different input sizes.

func BenchmarkSearchComparison(b *testing.B) {
	sizes := []int{100, 1000, 10000, 100000}

	for _, size := range sizes {
		data := GenerateSortedSlice(size)
		target := size - 1 // Search for last element (worst case for linear)

		b.Run("Linear/"+itoa(size), func(b *testing.B) {
			for b.Loop() {
				LinearSearch(data, target)
			}
		})

		b.Run("Binary/"+itoa(size), func(b *testing.B) {
			for b.Loop() {
				BinarySearch(data, target)
			}
		})
	}
}

// ========================================
// 4. Benchmarks with Memory Allocation Reporting
// ========================================
// Use b.ReportAllocs() to include allocation stats in output.
// This shows allocs/op and bytes/op.

func BenchmarkConcatPlus(b *testing.B) {
	items := GenerateStringSlice(100)
	b.ReportAllocs() // Report memory allocations
	b.ResetTimer()

	for b.Loop() {
		ConcatPlus(items)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	items := GenerateStringSlice(100)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		ConcatBuilder(items)
	}
}

func BenchmarkConcatJoin(b *testing.B) {
	items := GenerateStringSlice(100)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		ConcatJoin(items)
	}
}

// ========================================
// 5. Benchmark: Map vs Slice Lookup
// ========================================

func BenchmarkSliceContains(b *testing.B) {
	data := GenerateStringSlice(1000)
	target := data[len(data)-1] // Last element

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		SliceContains(data, target)
	}
}

func BenchmarkMapContains(b *testing.B) {
	data := GenerateStringSlice(1000)
	m := SliceToMap(data)
	target := data[len(data)-1]

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		MapContains(m, target)
	}
}

// ========================================
// Helper
// ========================================

// Simple int-to-string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Benchmark function signature:
//   func BenchmarkXxx(b *testing.B)
//
// The b.N loop (or b.Loop()):
//   for b.Loop() {     // Go 1.24+: preferred, cleaner
//       MyFunction()
//   }
//   // Or the classic way:
//   // for i := 0; i < b.N; i++ {
//   //     MyFunction()
//   // }
//
// Useful b methods:
//   b.ResetTimer()    — reset after expensive setup
//   b.ReportAllocs()  — include alloc stats in output
//   b.Run(name, fn)   — create sub-benchmarks
//   b.SetBytes(n)     — report throughput (bytes/sec)
//
// Running benchmarks:
//   go test -bench=.                    — run all benchmarks
//   go test -bench=BenchmarkSearch      — run matching benchmarks
//   go test -bench=. -benchmem          — include memory stats
//   go test -bench=. -count=5           — run 5 times for stability
//   go test -bench=. -benchtime=5s      — run each for 5 seconds
//
// Test coverage:
//   go test -cover                      — show coverage percentage
//   go test -coverprofile=coverage.out  — generate coverage data
//   go tool cover -html=coverage.out    — view coverage in browser
//   go tool cover -func=coverage.out    — show coverage by function
//
// Sample benchmark output:
//   BenchmarkLinearSearch-8     150234    7892 ns/op    0 B/op    0 allocs/op
//   BenchmarkBinarySearch-8   25641025    46.7 ns/op    0 B/op    0 allocs/op
//
//   Columns: name, iterations, time/op, bytes/op, allocs/op
