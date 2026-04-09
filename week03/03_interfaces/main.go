package main

import (
	"fmt"
	"math"
	"strings"
)

// ========================================
// Lesson 3: Interfaces in Go
// ========================================
// Interfaces define BEHAVIOR — a set of method signatures.
// Any type that implements all the methods automatically
// satisfies the interface. No "implements" keyword needed!
//
// This is called "implicit implementation" and is one of
// Go's most powerful features.

// ========================================
// Defining an interface
// ========================================
// A Shape must be able to calculate its area and perimeter.
type Shape interface {
	Area() float64
	Perimeter() float64
}

// ========================================
// Types that implement Shape
// ========================================
// Neither Rectangle nor Circle explicitly says "implements Shape".
// They just happen to have Area() and Perimeter() methods.

type Rectangle struct {
	Width, Height float64
}

func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

type Triangle struct {
	A, B, C float64 // side lengths
}

func (t Triangle) Area() float64 {
	// Heron's formula
	s := (t.A + t.B + t.C) / 2
	return math.Sqrt(s * (s - t.A) * (s - t.B) * (s - t.C))
}

func (t Triangle) Perimeter() float64 {
	return t.A + t.B + t.C
}

// ========================================
// Using interfaces: write functions that accept any Shape
// ========================================
// This function works with ANY Shape — Rectangle, Circle, Triangle,
// or any future type that has Area() and Perimeter().
func printShapeInfo(s Shape) {
	fmt.Printf("    Area: %.2f\n", s.Area())
	fmt.Printf("    Perimeter: %.2f\n", s.Perimeter())
}

// Compare two shapes by area
func largerShape(a, b Shape) Shape {
	if a.Area() > b.Area() {
		return a
	}
	return b
}

// ========================================
// The Stringer interface (from fmt package)
// ========================================
// The fmt package defines:
//   type Stringer interface {
//       String() string
//   }
// If your type has a String() method, fmt.Println and friends
// will use it automatically.

func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle(%.1f x %.1f)", r.Width, r.Height)
}

func (c Circle) String() string {
	return fmt.Sprintf("Circle(r=%.1f)", c.Radius)
}

func (t Triangle) String() string {
	return fmt.Sprintf("Triangle(%.1f, %.1f, %.1f)", t.A, t.B, t.C)
}

// ========================================
// Multiple interfaces
// ========================================
// A type can satisfy multiple interfaces at once.

type Describable interface {
	Describe() string
}

type Animal struct {
	Name    string
	Species string
	Legs    int
}

func (a Animal) Describe() string {
	return fmt.Sprintf("%s is a %s with %d legs", a.Name, a.Species, a.Legs)
}

func (a Animal) String() string {
	return fmt.Sprintf("%s the %s", a.Name, a.Species)
}

// ========================================
// Interface composition
// ========================================
// Interfaces can embed other interfaces to build larger ones.
type Reader interface {
	Read() string
}

type Writer interface {
	Write(data string)
}

// ReadWriter requires both Read and Write
type ReadWriter interface {
	Reader
	Writer
}

// A type that satisfies ReadWriter
type Document struct {
	content string
}

func (d *Document) Read() string {
	return d.content
}

func (d *Document) Write(data string) {
	d.content += data
}

// ========================================
// The empty interface: any (interface{})
// ========================================
// The empty interface has zero methods, so EVERY type satisfies it.
// In Go 1.18+, 'any' is an alias for 'interface{}'.
// Use it when you need to accept any type (like generics-lite).

func printAnything(value any) {
	fmt.Printf("    Value: %v (Type: %T)\n", value, value)
}

// ========================================
// Type assertions
// ========================================
// Type assertions let you extract the concrete type from an interface.
// Syntax: value, ok := interfaceVar.(ConcreteType)

func describeShape(s Shape) string {
	// The "comma ok" pattern for safe type assertions
	if rect, ok := s.(Rectangle); ok {
		return fmt.Sprintf("A rectangle measuring %.1f x %.1f", rect.Width, rect.Height)
	}
	if circ, ok := s.(Circle); ok {
		return fmt.Sprintf("A circle with radius %.1f", circ.Radius)
	}
	if tri, ok := s.(Triangle); ok {
		return fmt.Sprintf("A triangle with sides %.1f, %.1f, %.1f", tri.A, tri.B, tri.C)
	}
	return "An unknown shape"
}

// ========================================
// Type switches
// ========================================
// A type switch is a cleaner way to handle multiple type assertions.

func classifyValue(v any) string {
	switch val := v.(type) {
	case int:
		if val > 0 {
			return fmt.Sprintf("positive integer: %d", val)
		}
		return fmt.Sprintf("non-positive integer: %d", val)
	case float64:
		return fmt.Sprintf("float: %.2f", val)
	case string:
		return fmt.Sprintf("string of length %d: %q", len(val), val)
	case bool:
		return fmt.Sprintf("boolean: %v", val)
	case []int:
		return fmt.Sprintf("int slice with %d elements", len(val))
	case Shape:
		return fmt.Sprintf("a Shape with area %.2f", val.Area())
	case nil:
		return "nil value"
	default:
		return fmt.Sprintf("unknown type: %T", val)
	}
}

// ========================================
// Practical interface: Sorter
// ========================================
type Sorter interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

type StringSlice []string

func (s StringSlice) Len() int           { return len(s) }
func (s StringSlice) Less(i, j int) bool { return strings.ToLower(s[i]) < strings.ToLower(s[j]) }
func (s StringSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// Simple bubble sort that works with ANY Sorter
func BubbleSort(s Sorter) {
	for i := 0; i < s.Len()-1; i++ {
		for j := 0; j < s.Len()-i-1; j++ {
			if s.Less(j+1, j) {
				s.Swap(j, j+1)
			}
		}
	}
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Lesson: Interfaces in Go")
	fmt.Println("========================================")

	// ========================================
	// Implicit interface implementation
	// ========================================
	fmt.Println("\n--- Implicit Implementation ---")

	// All three types implement Shape without explicitly saying so
	rect := Rectangle{Width: 10, Height: 5}
	circ := Circle{Radius: 7}
	tri := Triangle{A: 3, B: 4, C: 5}

	// We can store them all as Shape
	fmt.Printf("  %s:\n", rect)
	printShapeInfo(rect)

	fmt.Printf("  %s:\n", circ)
	printShapeInfo(circ)

	fmt.Printf("  %s:\n", tri)
	printShapeInfo(tri)

	// ========================================
	// Interface variables and polymorphism
	// ========================================
	fmt.Println("\n--- Polymorphism (Slice of Interfaces) ---")

	// A slice of Shape can hold any type that implements Shape
	shapes := []Shape{
		Rectangle{Width: 12, Height: 8},
		Circle{Radius: 5},
		Triangle{A: 5, B: 12, C: 13},
		Rectangle{Width: 3, Height: 3},
		Circle{Radius: 10},
	}

	// Calculate total area of all shapes — doesn't matter what they are
	totalArea := 0.0
	for _, s := range shapes {
		totalArea += s.Area()
	}
	fmt.Printf("  Total area of %d shapes: %.2f\n", len(shapes), totalArea)

	// Find the largest shape
	largest := shapes[0]
	for _, s := range shapes[1:] {
		largest = largerShape(largest, s)
	}
	fmt.Printf("  Largest shape: %v (area=%.2f)\n", largest, largest.Area())

	// ========================================
	// Stringer interface
	// ========================================
	fmt.Println("\n--- Stringer Interface ---")
	// fmt.Println automatically calls String() if available
	fmt.Println("  ", rect)
	fmt.Println("  ", circ)
	fmt.Println("  ", tri)

	animal := Animal{Name: "Max", Species: "Dog", Legs: 4}
	fmt.Println("  ", animal) // Uses String() method

	// ========================================
	// Type assertions
	// ========================================
	fmt.Println("\n--- Type Assertions ---")

	var s Shape = Circle{Radius: 5}

	// Safe type assertion with comma-ok pattern
	if c, ok := s.(Circle); ok {
		fmt.Printf("  It's a Circle! Radius=%.1f\n", c.Radius)
	}

	// Check for Rectangle (will be false since s is a Circle)
	if _, ok := s.(Rectangle); !ok {
		fmt.Println("  It's NOT a Rectangle")
	}

	// Describe shapes using type assertions
	fmt.Println("\n  Shape descriptions:")
	for _, shape := range shapes {
		fmt.Printf("    %s\n", describeShape(shape))
	}

	// ========================================
	// Type switches
	// ========================================
	fmt.Println("\n--- Type Switches ---")

	values := []any{
		42,
		3.14,
		"hello",
		true,
		[]int{1, 2, 3},
		Circle{Radius: 5},
		nil,
		-7,
	}

	for _, v := range values {
		fmt.Printf("    %s\n", classifyValue(v))
	}

	// ========================================
	// Empty interface (any)
	// ========================================
	fmt.Println("\n--- Empty Interface (any) ---")

	// any can hold any value
	printAnything(42)
	printAnything("Go is great")
	printAnything(3.14)
	printAnything(rect)
	printAnything([]int{1, 2, 3})

	// A map with string keys and any values (like JSON)
	config := map[string]any{
		"host":    "localhost",
		"port":    8080,
		"debug":   true,
		"timeout": 30.5,
	}
	fmt.Println("\n  Config map[string]any:")
	for key, val := range config {
		fmt.Printf("    %s = %v (%T)\n", key, val, val)
	}

	// ========================================
	// Interface composition
	// ========================================
	fmt.Println("\n--- Interface Composition ---")

	doc := &Document{}
	doc.Write("Hello, ")
	doc.Write("interfaces!")
	fmt.Printf("  Read: %q\n", doc.Read())

	// doc satisfies ReadWriter because it has both Read and Write
	var rw ReadWriter = doc
	rw.Write(" More text.")
	fmt.Printf("  After more writing: %q\n", rw.Read())

	// ========================================
	// Practical: sorting with an interface
	// ========================================
	fmt.Println("\n--- Sorting with Interfaces ---")

	names := StringSlice{"Charlie", "alice", "Bob", "diana"}
	fmt.Printf("  Before sort: %v\n", []string(names))
	BubbleSort(names)
	fmt.Printf("  After sort:  %v\n", []string(names))

	// ========================================
	// Interface nil gotcha
	// ========================================
	fmt.Println("\n--- Interface Nil Check ---")

	// An interface is nil only if BOTH its type and value are nil
	var shape Shape // nil interface — no type, no value
	fmt.Printf("  var shape Shape: shape == nil? %v\n", shape == nil)

	// Be aware: a nil pointer of a concrete type is NOT a nil interface
	var rectPtr *Rectangle // nil pointer
	shape = rectPtr        // shape now has a type (*Rectangle) but nil value
	fmt.Printf("  shape = (*Rectangle)(nil): shape == nil? %v\n", shape == nil)
	fmt.Println("  (has type info, so it's not nil — a common gotcha!)")

	fmt.Println("\n========================================")
	fmt.Println("  Key Takeaways:")
	fmt.Println("  - Interfaces define behavior (method sets)")
	fmt.Println("  - Implementation is implicit (no 'implements')")
	fmt.Println("  - Accept interfaces, return structs")
	fmt.Println("  - Type assertions: val, ok := i.(Type)")
	fmt.Println("  - Type switches for multiple type checks")
	fmt.Println("  - any (interface{}) accepts all types")
	fmt.Println("  - Stringer: implement String() for printing")
	fmt.Println("  - Compose interfaces by embedding")
	fmt.Println("========================================")
}
