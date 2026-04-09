package main

import (
	"fmt"
	"strings"

	"github.com/sri/learngo/week05/03_project/mathutil"
	"github.com/sri/learngo/week05/03_project/stringutil"
)

func main() {
	// ========================================
	// Multi-Package Project Demo
	// Demonstrates importing and using local packages
	// ========================================
	fmt.Println("========================================")
	fmt.Println("  Multi-Package Project")
	fmt.Println("  Using mathutil & stringutil packages")
	fmt.Println("========================================")

	// ========================================
	// mathutil package demos
	// ========================================
	fmt.Println("\n--- mathutil Package ---")

	// GCD and LCM
	fmt.Println("\n=== GCD and LCM ===")
	pairs := [][2]int{{12, 8}, {100, 75}, {17, 13}, {0, 5}}
	for _, pair := range pairs {
		a, b := pair[0], pair[1]
		fmt.Printf("  GCD(%d, %d) = %d\n", a, b, mathutil.GCD(a, b))
		fmt.Printf("  LCM(%d, %d) = %d\n", a, b, mathutil.LCM(a, b))
	}

	// Prime numbers
	fmt.Println("\n=== Prime Numbers ===")
	testNumbers := []int{1, 2, 7, 10, 13, 25, 97, 100}
	for _, n := range testNumbers {
		if mathutil.IsPrime(n) {
			fmt.Printf("  %3d is prime\n", n)
		} else {
			fmt.Printf("  %3d is NOT prime\n", n)
		}
	}

	primes := mathutil.PrimesUpTo(50)
	fmt.Printf("\n  Primes up to 50: %v\n", primes)

	// Factorial
	fmt.Println("\n=== Factorial ===")
	for _, n := range []int{0, 1, 5, 10, 12} {
		fmt.Printf("  %2d! = %d\n", n, mathutil.Factorial(n))
	}

	// Fibonacci
	fmt.Println("\n=== Fibonacci ===")
	fibs := mathutil.Fibonacci(15)
	fmt.Printf("  First 15 Fibonacci: %v\n", fibs)

	// Abs and Sqrt
	fmt.Println("\n=== Abs and Sqrt ===")
	fmt.Printf("  Abs(-42) = %d\n", mathutil.Abs(-42))
	fmt.Printf("  Abs(42)  = %d\n", mathutil.Abs(42))
	fmt.Printf("  Sqrt(144) = %d\n", mathutil.Sqrt(144))
	fmt.Printf("  Sqrt(50)  = %d\n", mathutil.Sqrt(50))

	// ========================================
	// stringutil package demos
	// ========================================
	fmt.Println("\n--- stringutil Package ---")

	// Reverse
	fmt.Println("\n=== Reverse ===")
	testStrings := []string{"Hello", "Go is fun", "racecar", "Hello, World!"}
	for _, s := range testStrings {
		fmt.Printf("  %-20s -> %s\n", fmt.Sprintf("%q", s), stringutil.Reverse(s))
	}

	// IsPalindrome
	fmt.Println("\n=== Palindrome Check ===")
	palindromeTests := []string{
		"racecar",
		"A man, a plan, a canal: Panama",
		"hello",
		"Was it a car or a cat I saw?",
		"Go",
		"Madam, I'm Adam",
	}
	for _, s := range palindromeTests {
		result := "no"
		if stringutil.IsPalindrome(s) {
			result = "YES"
		}
		fmt.Printf("  %-40s -> Palindrome? %s\n", fmt.Sprintf("%q", s), result)
	}

	// WordCount
	fmt.Println("\n=== Word Count ===")
	sentences := []string{
		"Hello World",
		"Go is a statically typed compiled language",
		"   spaces   everywhere   ",
		"",
		"one",
	}
	for _, s := range sentences {
		fmt.Printf("  %-50s -> %d words\n", fmt.Sprintf("%q", s), stringutil.WordCount(s))
	}

	// Capitalize
	fmt.Println("\n=== Capitalize ===")
	capitalizeTests := []string{
		"hello world",
		"go programming language",
		"the quick brown fox",
		"already Capitalized Words",
	}
	for _, s := range capitalizeTests {
		fmt.Printf("  %-35s -> %s\n", fmt.Sprintf("%q", s), stringutil.Capitalize(s))
	}

	// CountChar
	fmt.Println("\n=== Character Count ===")
	text := "mississippi"
	for _, ch := range []rune{'s', 'i', 'p', 'm'} {
		fmt.Printf("  '%c' in %q: %d times\n", ch, text, stringutil.CountChar(text, ch))
	}

	// Truncate
	fmt.Println("\n=== Truncate ===")
	longText := "The Go programming language is an open-source project"
	for _, maxLen := range []int{10, 20, 30, 100} {
		fmt.Printf("  Truncate(%d): %s\n", maxLen, stringutil.Truncate(longText, maxLen))
	}

	// IsAnagram
	fmt.Println("\n=== Anagram Check ===")
	anagramPairs := [][2]string{
		{"listen", "silent"},
		{"hello", "world"},
		{"Astronomer", "Moon starer"},
		{"Go Lang", "Log Nag"},
		{"abc", "def"},
	}
	for _, pair := range anagramPairs {
		result := "no"
		if stringutil.IsAnagram(pair[0], pair[1]) {
			result = "YES"
		}
		fmt.Printf("  %-15s & %-15s -> Anagram? %s\n",
			fmt.Sprintf("%q", pair[0]), fmt.Sprintf("%q", pair[1]), result)
	}

	// ========================================
	// Combining both packages
	// ========================================
	fmt.Println("\n--- Combining Both Packages ---")
	fmt.Println("\n=== Prime Word Game ===")
	sentence := "Go is a wonderful language for building fast and reliable software"
	words := strings.Fields(sentence)

	fmt.Printf("Sentence: %q\n", sentence)
	fmt.Printf("Words: %d\n\n", stringutil.WordCount(sentence))

	for _, word := range words {
		wordLen := len(word)
		isPrime := mathutil.IsPrime(wordLen)
		reversed := stringutil.Reverse(word)
		isPalin := stringutil.IsPalindrome(word)

		marker := " "
		if isPrime {
			marker = "*"
		}

		palindrome := ""
		if isPalin {
			palindrome = " [palindrome!]"
		}

		fmt.Printf("  %s %-12s len=%2d  reversed=%-12s%s\n",
			marker, word, wordLen, reversed, palindrome)
	}
	fmt.Println("\n  * = word length is a prime number")

	fmt.Println("\n========================================")
	fmt.Println("  Project complete!")
	fmt.Println("========================================")
}
