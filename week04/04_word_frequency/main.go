package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	// ========================================
	// Word Frequency Counter
	// Mini-project: slices + maps working together
	// ========================================
	fmt.Println("========================================")
	fmt.Println("  Word Frequency Counter")
	fmt.Println("========================================")

	text := `Go is an open source programming language that makes it simple to build
reliable and efficient software. Go was designed at Google by Robert Griesemer,
Rob Pike, and Ken Thompson. Go is syntactically similar to C, but with memory
safety, garbage collection, structural typing, and CSP-style concurrency.
Go is sometimes referred to as Golang because of its domain name, golang.org,
but the proper name is Go. Go has built-in support for concurrent programming
and a rich standard library. Go is used by many companies including Google,
Uber, Dropbox, and Docker. Go makes it easy to build simple, reliable, and
efficient software.`

	fmt.Println("\nInput text:")
	fmt.Println(text)
	fmt.Println()

	// ========================================
	// Step 1: Clean and split the text into words
	// ========================================
	// Convert to lowercase for case-insensitive counting
	lowered := strings.ToLower(text)

	// Replace punctuation with spaces so words are clean
	replacer := strings.NewReplacer(
		",", "",
		".", "",
		"!", "",
		"?", "",
		":", "",
		";", "",
		"(", "",
		")", "",
		"\n", " ",
	)
	cleaned := replacer.Replace(lowered)

	// Split into words (Fields handles multiple spaces)
	words := strings.Fields(cleaned)

	fmt.Printf("Total words: %d\n", len(words))

	// ========================================
	// Step 2: Count word frequencies using a map
	// ========================================
	frequency := make(map[string]int)
	for _, word := range words {
		frequency[word]++
	}

	fmt.Printf("Unique words: %d\n", len(frequency))

	// ========================================
	// Step 3: Sort by frequency (descending)
	// ========================================
	// We need a slice to sort — maps can't be sorted directly
	// Create a slice of word-count pairs
	type wordCount struct {
		word  string
		count int
	}

	pairs := make([]wordCount, 0, len(frequency))
	for word, count := range frequency {
		pairs = append(pairs, wordCount{word: word, count: count})
	}

	// Sort by count (descending), then alphabetically for ties
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].word < pairs[j].word
	})

	// ========================================
	// Step 4: Display results — top 15 words
	// ========================================
	fmt.Println("\n--- Top 15 Most Frequent Words ---")
	fmt.Printf("%-4s %-15s %s\n", "Rank", "Word", "Count")
	fmt.Println(strings.Repeat("-", 30))

	limit := 15
	if len(pairs) < limit {
		limit = len(pairs)
	}

	for i := 0; i < limit; i++ {
		// Create a simple bar chart
		bar := strings.Repeat("█", pairs[i].count)
		fmt.Printf("%-4d %-15s %d  %s\n", i+1, pairs[i].word, pairs[i].count, bar)
	}

	// ========================================
	// Step 5: Display all words alphabetically
	// ========================================
	fmt.Println("\n--- All Words (alphabetical) ---")

	// Get all unique words into a slice and sort
	allWords := make([]string, 0, len(frequency))
	for word := range frequency {
		allWords = append(allWords, word)
	}
	sort.Strings(allWords)

	for _, word := range allWords {
		fmt.Printf("  %-20s %d\n", word, frequency[word])
	}

	// ========================================
	// Step 6: Find words that appear exactly once
	// ========================================
	fmt.Println("\n--- Words appearing only once (hapax legomena) ---")

	var hapax []string
	for _, word := range allWords {
		if frequency[word] == 1 {
			hapax = append(hapax, word)
		}
	}
	fmt.Printf("Found %d unique words: %v\n", len(hapax), hapax)

	// ========================================
	// Step 7: Word length statistics
	// ========================================
	fmt.Println("\n--- Word Length Distribution ---")

	lengthDist := make(map[int]int)
	longestWord := ""
	for _, word := range allWords {
		lengthDist[len(word)]++
		if len(word) > len(longestWord) {
			longestWord = word
		}
	}

	// Get sorted lengths
	lengths := make([]int, 0, len(lengthDist))
	for l := range lengthDist {
		lengths = append(lengths, l)
	}
	sort.Ints(lengths)

	for _, l := range lengths {
		bar := strings.Repeat("▒", lengthDist[l])
		fmt.Printf("  %2d letters: %2d words %s\n", l, lengthDist[l], bar)
	}
	fmt.Printf("\nLongest word: %q (%d letters)\n", longestWord, len(longestWord))

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("  Summary")
	fmt.Println("========================================")
	fmt.Printf("  Total words:      %d\n", len(words))
	fmt.Printf("  Unique words:     %d\n", len(frequency))
	fmt.Printf("  Most common:      %q (%dx)\n", pairs[0].word, pairs[0].count)
	fmt.Printf("  Hapax legomena:   %d words\n", len(hapax))
	fmt.Printf("  Longest word:     %q\n", longestWord)
	fmt.Println("========================================")
}
