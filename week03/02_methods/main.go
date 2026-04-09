package main

import (
	"fmt"
	"math"
)

// ========================================
// Lesson 2: Methods in Go
// ========================================
// Methods are functions attached to a type. Instead of calling
// doSomething(myStruct), you call myStruct.DoSomething().
//
// Syntax: func (receiver Type) MethodName(params) returnType { body }
//
// The "receiver" is the struct instance the method operates on.

// ========================================
// A simple struct with methods
// ========================================
type Rectangle struct {
	Width  float64
	Height float64
}

// Value receiver: gets a COPY of the Rectangle.
// Use value receivers when the method only READS data.
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Methods can return multiple values too
func (r Rectangle) Dimensions() (float64, float64) {
	return r.Width, r.Height
}

// Methods can take parameters just like regular functions
func (r Rectangle) IsLargerThan(other Rectangle) bool {
	return r.Area() > other.Area()
}

// ========================================
// Value receiver vs Pointer receiver
// ========================================
// Key decision: use a value receiver (r Rectangle) or
// a pointer receiver (r *Rectangle)?

// Pointer receiver: gets a POINTER to the original.
// Use pointer receivers when the method MODIFIES the struct.
func (r *Rectangle) Scale(factor float64) {
	// This modifies the ORIGINAL Rectangle
	r.Width *= factor
	r.Height *= factor
}

// Another pointer receiver method
func (r *Rectangle) SetDimensions(width, height float64) {
	r.Width = width
	r.Height = height
}

// ========================================
// Why pointer receivers? A demonstration
// ========================================
// This value receiver CANNOT modify the original (it gets a copy)
func (r Rectangle) ScaleBroken(factor float64) {
	// This modifies a COPY — the original is unchanged!
	r.Width *= factor
	r.Height *= factor
	// The copy is discarded when this method returns
}

// ========================================
// Circle type with methods
// ========================================
type Circle struct {
	Radius float64
}

func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Circumference() float64 {
	return 2 * math.Pi * c.Radius
}

func (c Circle) Diameter() float64 {
	return 2 * c.Radius
}

func (c *Circle) Grow(amount float64) {
	c.Radius += amount
}

func (c *Circle) Shrink(amount float64) {
	if c.Radius-amount > 0 {
		c.Radius -= amount
	}
}

// ========================================
// BankAccount: practical pointer receiver example
// ========================================
type BankAccount struct {
	Owner   string
	Balance float64
}

// NewBankAccount is a "constructor function" — a common Go pattern.
// Go doesn't have constructors, so we use regular functions that
// return a pointer to a new struct.
func NewBankAccount(owner string, initialBalance float64) *BankAccount {
	return &BankAccount{
		Owner:   owner,
		Balance: initialBalance,
	}
}

func (a *BankAccount) Deposit(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("deposit amount must be positive, got %.2f", amount)
	}
	a.Balance += amount
	return nil
}

func (a *BankAccount) Withdraw(amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("withdrawal amount must be positive, got %.2f", amount)
	}
	if amount > a.Balance {
		return fmt.Errorf("insufficient funds: balance=%.2f, requested=%.2f",
			a.Balance, amount)
	}
	a.Balance -= amount
	return nil
}

// String method — we'll learn about the Stringer interface in the
// next lesson, but for now just know that fmt.Println will call
// this automatically!
func (a BankAccount) String() string {
	return fmt.Sprintf("Account(%s: $%.2f)", a.Owner, a.Balance)
}

// ========================================
// Methods on any named type (not just structs!)
// ========================================
// You can define methods on any type you create with 'type'.

type Celsius float64
type Fahrenheit float64

func (c Celsius) ToFahrenheit() Fahrenheit {
	return Fahrenheit(c*9.0/5.0 + 32)
}

func (f Fahrenheit) ToCelsius() Celsius {
	return Celsius((f - 32) * 5.0 / 9.0)
}

func (c Celsius) String() string {
	return fmt.Sprintf("%.1f°C", float64(c))
}

func (f Fahrenheit) String() string {
	return fmt.Sprintf("%.1f°F", float64(f))
}

// ========================================
// Method chaining with pointer receivers
// ========================================
type Builder struct {
	content string
}

// Each method returns *Builder so calls can be chained
func (b *Builder) Add(text string) *Builder {
	b.content += text
	return b // return the pointer so the next method can be called
}

func (b *Builder) AddLine(text string) *Builder {
	b.content += text + "\n"
	return b
}

func (b *Builder) Build() string {
	return b.content
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Lesson: Methods in Go")
	fmt.Println("========================================")

	// ========================================
	// Calling methods on a struct
	// ========================================
	fmt.Println("\n--- Rectangle Methods (Value Receivers) ---")

	rect := Rectangle{Width: 10, Height: 5}
	fmt.Printf("  Rectangle: %.0f x %.0f\n", rect.Width, rect.Height)
	fmt.Printf("  Area: %.1f\n", rect.Area())
	fmt.Printf("  Perimeter: %.1f\n", rect.Perimeter())

	w, h := rect.Dimensions()
	fmt.Printf("  Dimensions: %.0f x %.0f\n", w, h)

	// ========================================
	// Comparing with methods
	// ========================================
	fmt.Println("\n--- Method with Parameters ---")
	small := Rectangle{Width: 3, Height: 4}
	big := Rectangle{Width: 10, Height: 20}
	fmt.Printf("  Small area: %.0f\n", small.Area())
	fmt.Printf("  Big area: %.0f\n", big.Area())
	fmt.Printf("  Is big larger than small? %v\n", big.IsLargerThan(small))
	fmt.Printf("  Is small larger than big? %v\n", small.IsLargerThan(big))

	// ========================================
	// Pointer receiver: modifying the original
	// ========================================
	fmt.Println("\n--- Pointer Receivers (Modify Original) ---")

	rect2 := Rectangle{Width: 10, Height: 5}
	fmt.Printf("  Before Scale: %.0f x %.0f (area=%.0f)\n",
		rect2.Width, rect2.Height, rect2.Area())

	rect2.Scale(2) // Doubles both dimensions
	fmt.Printf("  After Scale(2): %.0f x %.0f (area=%.0f)\n",
		rect2.Width, rect2.Height, rect2.Area())

	rect2.SetDimensions(7, 3)
	fmt.Printf("  After SetDimensions(7, 3): %.0f x %.0f\n", rect2.Width, rect2.Height)

	// ========================================
	// Value receiver pitfall: modifications are lost!
	// ========================================
	fmt.Println("\n--- Value Receiver Pitfall ---")

	rect3 := Rectangle{Width: 10, Height: 5}
	fmt.Printf("  Before ScaleBroken: %.0f x %.0f\n", rect3.Width, rect3.Height)

	rect3.ScaleBroken(2) // This does NOTHING to rect3!
	fmt.Printf("  After ScaleBroken(2): %.0f x %.0f  <-- unchanged!\n",
		rect3.Width, rect3.Height)
	fmt.Println("  (Value receivers get a copy — changes are lost)")

	// ========================================
	// Go automatically handles &/pointer for method calls
	// ========================================
	fmt.Println("\n--- Automatic Pointer Dereferencing ---")

	// Even though Scale has a pointer receiver (*Rectangle),
	// Go lets you call it on a value — it automatically takes &rect4
	rect4 := Rectangle{Width: 5, Height: 5}
	rect4.Scale(3) // Go does (&rect4).Scale(3) under the hood
	fmt.Printf("  Works on values too: %.0f x %.0f\n", rect4.Width, rect4.Height)

	// And pointer variables can call value receiver methods
	rect5 := &Rectangle{Width: 8, Height: 4}
	fmt.Printf("  Pointer calling Area(): %.1f\n", rect5.Area()) // (*rect5).Area()

	// ========================================
	// Circle methods
	// ========================================
	fmt.Println("\n--- Circle Methods ---")

	c := Circle{Radius: 5}
	fmt.Printf("  Circle (r=%.1f):\n", c.Radius)
	fmt.Printf("    Diameter: %.2f\n", c.Diameter())
	fmt.Printf("    Area: %.2f\n", c.Area())
	fmt.Printf("    Circumference: %.2f\n", c.Circumference())

	c.Grow(2.5)
	fmt.Printf("  After Grow(2.5): r=%.1f, area=%.2f\n", c.Radius, c.Area())

	c.Shrink(1)
	fmt.Printf("  After Shrink(1): r=%.1f\n", c.Radius)

	// ========================================
	// BankAccount: practical example
	// ========================================
	fmt.Println("\n--- BankAccount (Constructor + Methods) ---")

	// Use the constructor function
	account := NewBankAccount("Sri", 1000)
	fmt.Printf("  %s\n", account) // String() method is called automatically

	err := account.Deposit(500)
	if err != nil {
		fmt.Println("  Error:", err)
	}
	fmt.Printf("  After deposit $500: %s\n", account)

	err = account.Withdraw(200)
	if err != nil {
		fmt.Println("  Error:", err)
	}
	fmt.Printf("  After withdraw $200: %s\n", account)

	// Try to withdraw too much
	err = account.Withdraw(5000)
	if err != nil {
		fmt.Println("  Withdraw $5000:", err)
	}

	// Try invalid deposit
	err = account.Deposit(-50)
	if err != nil {
		fmt.Println("  Deposit -$50:", err)
	}

	// ========================================
	// Methods on non-struct types
	// ========================================
	fmt.Println("\n--- Methods on Custom Types ---")

	temp := Celsius(100)
	fmt.Printf("  %s = %s\n", temp, temp.ToFahrenheit())

	bodyTemp := Fahrenheit(98.6)
	fmt.Printf("  %s = %s\n", bodyTemp, bodyTemp.ToCelsius())

	// ========================================
	// Method chaining
	// ========================================
	fmt.Println("\n--- Method Chaining ---")

	message := (&Builder{}).
		AddLine("Dear Go Learner,").
		AddLine("").
		AddLine("Methods are functions attached to types.").
		AddLine("Value receivers read, pointer receivers write.").
		AddLine("").
		Add("Happy coding!").
		Build()

	fmt.Println(message)

	// ========================================
	// When to use value vs pointer receivers
	// ========================================
	fmt.Println("========================================")
	fmt.Println("  When to Use Each Receiver Type:")
	fmt.Println("========================================")
	fmt.Println("  Value receiver (r Rectangle):")
	fmt.Println("    - Method only READS the struct")
	fmt.Println("    - Struct is small (a few fields)")
	fmt.Println("    - You want immutability")
	fmt.Println()
	fmt.Println("  Pointer receiver (r *Rectangle):")
	fmt.Println("    - Method MODIFIES the struct")
	fmt.Println("    - Struct is large (avoid copying)")
	fmt.Println("    - Consistency (if any method needs a")
	fmt.Println("      pointer, use pointer for all)")
	fmt.Println("========================================")
}
