package main

import "fmt"

// ========================================
// Helper types and functions for examples
// ========================================

// Point represents a 2D coordinate
type Point struct {
	X, Y float64
}

// Person represents a person with a name and age
type Person struct {
	Name string
	Age  int
}

func main() {
	// ========================================
	// What are pointers?
	// A pointer holds the MEMORY ADDRESS of a value
	// ========================================
	fmt.Println("=== Basics: & (address-of) and * (dereference) ===")

	x := 42
	p := &x // p is a pointer to x — it holds x's memory address

	fmt.Println("x  =", x)            // 42
	fmt.Println("&x =", &x)           // memory address like 0xc0000b2008
	fmt.Println("p  =", p)            // same address
	fmt.Println("*p =", *p)           // 42 — dereference: read the value at that address
	fmt.Printf("Type of p: %T\n", p)  // *int

	// Changing the value through the pointer changes the original
	*p = 100
	fmt.Println("\nAfter *p = 100:")
	fmt.Println("x  =", x)  // 100 — x changed!
	fmt.Println("*p =", *p) // 100

	// ========================================
	// Why pointers matter: pass-by-value in Go
	// ========================================
	fmt.Println("\n=== Pass-by-value (without pointers) ===")

	num := 10
	fmt.Println("Before doubleValue:", num)
	doubleValue(num)
	fmt.Println("After doubleValue: ", num) // still 10!

	fmt.Println("\n=== Pass-by-pointer ===")
	fmt.Println("Before doublePointer:", num)
	doublePointer(&num)
	fmt.Println("After doublePointer: ", num) // 20!

	// ========================================
	// Pointer to struct — most common use case
	// ========================================
	fmt.Println("\n=== Pointer to struct ===")

	pt := Point{X: 3.0, Y: 4.0}
	fmt.Println("Point:", pt)

	// Get a pointer to the struct
	pp := &pt
	fmt.Println("Pointer:", pp)

	// Go lets you access fields through a pointer WITHOUT explicit dereference
	// pp.X is shorthand for (*pp).X
	pp.X = 10.0
	pp.Y = 20.0
	fmt.Println("After modifying through pointer:", pt) // {10 20}

	// Create a struct and get its pointer in one step
	pp2 := &Point{X: 5.0, Y: 6.0}
	fmt.Println("Direct pointer:", pp2, "->", *pp2)

	// new() creates a zero-valued instance and returns a pointer
	pp3 := new(Point)
	fmt.Println("new(Point):", pp3, "->", *pp3) // {0 0}

	// ========================================
	// Nil pointers
	// ========================================
	fmt.Println("\n=== Nil pointers ===")

	var nilPtr *int
	fmt.Println("Nil pointer:", nilPtr)
	fmt.Println("Is nil?", nilPtr == nil) // true

	// Dereferencing a nil pointer PANICS!
	// fmt.Println(*nilPtr) // runtime error: invalid memory address

	// Always check for nil before dereferencing
	if nilPtr != nil {
		fmt.Println("Value:", *nilPtr)
	} else {
		fmt.Println("Pointer is nil — can't dereference")
	}

	// Safe pointer usage pattern
	safePtr := safeGet(nil)
	fmt.Println("Safe default:", safePtr)

	val := 42
	safePtr = safeGet(&val)
	fmt.Println("Safe with value:", safePtr)

	// ========================================
	// Passing pointers to functions
	// ========================================
	fmt.Println("\n=== Passing pointers to functions ===")

	alice := Person{Name: "Alice", Age: 30}
	fmt.Println("Before birthday:", alice)
	birthday(&alice)
	fmt.Println("After birthday: ", alice)

	// Swap using pointers
	a, b := 5, 10
	fmt.Printf("\nBefore swap: a=%d, b=%d\n", a, b)
	swap(&a, &b)
	fmt.Printf("After swap:  a=%d, b=%d\n", a, b)

	// ========================================
	// Pointer receivers vs value receivers
	// ========================================
	fmt.Println("\n=== Value receiver vs pointer receiver ===")

	p1 := Point{X: 1, Y: 2}

	// Value receiver — works on a copy
	fmt.Println("Original:", p1)
	scaled := p1.Scale(3) // returns a new Point
	fmt.Println("Scaled (new):", scaled)
	fmt.Println("Original unchanged:", p1)

	// Pointer receiver — modifies in place
	p1.Move(10, 20) // Go automatically takes &p1
	fmt.Println("After Move:", p1) // modified!

	// ========================================
	// When to use pointers — practical guidelines
	// ========================================
	fmt.Println("\n=== When to use pointers ===")

	fmt.Println(`
USE POINTERS WHEN:
  1. You need to modify the original value
  2. The struct is large (avoids expensive copies)
  3. You need to represent "no value" (nil)
  4. Implementing interfaces with pointer receivers
  5. Sharing data between goroutines (with synchronization)

USE VALUES WHEN:
  1. The data is small (int, bool, small structs)
  2. You want immutability (function can't change your data)
  3. The type is a map, slice, channel (already reference types)
  4. Concurrency safety (copies are inherently safe)`)

	// ========================================
	// Pointers and slices/maps (reference types)
	// ========================================
	fmt.Println("\n=== Slices and maps are already references ===")

	// You DON'T need a pointer to modify a slice's elements
	nums := []int{1, 2, 3}
	fmt.Println("Before:", nums)
	doubleSliceElements(nums)
	fmt.Println("After: ", nums) // modified!

	// But you DO need a pointer to change the slice header itself (append)
	fmt.Println("\nAppend without pointer:")
	appendWithoutPointer(nums)
	fmt.Println("Nums:", nums) // unchanged — append created a new slice header

	fmt.Println("\nAppend with pointer:")
	appendWithPointer(&nums)
	fmt.Println("Nums:", nums) // has 999!

	// ========================================
	// Pointer to pointer (rare but possible)
	// ========================================
	fmt.Println("\n=== Pointer to pointer ===")

	value := "hello"
	ptr := &value
	ptrPtr := &ptr

	fmt.Println("value: ", value)
	fmt.Println("*ptr:  ", *ptr)
	fmt.Println("**ptrPtr:", **ptrPtr)
	fmt.Printf("Types: value=%T, ptr=%T, ptrPtr=%T\n", value, ptr, ptrPtr)
}

// ========================================
// Functions demonstrating pointer concepts
// ========================================

// doubleValue receives a COPY — can't change the original
func doubleValue(n int) {
	n *= 2
	fmt.Println("  Inside doubleValue:", n) // 20 locally
}

// doublePointer receives a POINTER — can change the original
func doublePointer(n *int) {
	*n *= 2
	fmt.Println("  Inside doublePointer:", *n)
}

// safeGet returns the value at a pointer, or a default if nil
func safeGet(p *int) int {
	if p == nil {
		return -1 // default value
	}
	return *p
}

// birthday modifies a person's age through a pointer
func birthday(p *Person) {
	p.Age++ // Go auto-dereferences: same as (*p).Age++
	fmt.Printf("  Happy birthday, %s! You are now %d.\n", p.Name, p.Age)
}

// swap exchanges two values using pointers
func swap(a, b *int) {
	*a, *b = *b, *a
}

// ========================================
// Methods: value receiver vs pointer receiver
// ========================================

// Scale uses a VALUE receiver — returns a new Point, doesn't modify original
func (p Point) Scale(factor float64) Point {
	return Point{X: p.X * factor, Y: p.Y * factor}
}

// Move uses a POINTER receiver — modifies the Point in place
func (p *Point) Move(dx, dy float64) {
	p.X += dx
	p.Y += dy
}

// ========================================
// Functions showing reference type behavior
// ========================================

// doubleSliceElements modifies existing elements (slice is a reference)
func doubleSliceElements(s []int) {
	for i := range s {
		s[i] *= 2
	}
}

// appendWithoutPointer can't change the caller's slice header
func appendWithoutPointer(s []int) {
	s = append(s, 999)
	fmt.Println("  Inside function:", s)
}

// appendWithPointer CAN change the caller's slice
func appendWithPointer(s *[]int) {
	*s = append(*s, 999)
	fmt.Println("  Inside function:", *s)
}
