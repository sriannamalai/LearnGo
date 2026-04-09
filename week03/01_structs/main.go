package main

import "fmt"

// ========================================
// Lesson 1: Structs in Go
// ========================================
// Structs are Go's way of creating custom data types.
// They group related data together — like a blueprint for an object.
// If you're coming from other languages, think of them as
// classes (but without inheritance).

// ========================================
// Defining a struct
// ========================================
// Use the 'type' keyword followed by the struct name and 'struct'.
// Fields are listed with their names and types.
type Person struct {
	FirstName string
	LastName  string
	Age       int
	Email     string
}

// ========================================
// Nested structs
// ========================================
// Structs can contain other structs as fields.
type Address struct {
	Street  string
	City    string
	State   string
	ZipCode string
	Country string
}

type Employee struct {
	Person      Person  // Nested struct (named field)
	Department  string
	Salary      float64
	HomeAddress Address // Another nested struct
}

// ========================================
// Embedded structs (anonymous fields)
// ========================================
// When you embed a struct without a field name, its fields
// are "promoted" — you can access them directly.
type Student struct {
	Person // Embedded (anonymous) — fields are promoted
	GPA    float64
	Major  string
	Year   int
}

// ========================================
// Struct with a slice field
// ========================================
type Playlist struct {
	Name   string
	Songs  []string
	Likes  int
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Lesson: Structs in Go")
	fmt.Println("========================================")

	// ========================================
	// Creating struct instances
	// ========================================
	fmt.Println("\n--- Creating Structs ---")

	// Method 1: Named fields (recommended — order doesn't matter)
	p1 := Person{
		FirstName: "Sri",
		LastName:  "Annamalai",
		Age:       30,
		Email:     "sri@example.com",
	}
	fmt.Printf("  p1: %+v\n", p1) // %+v shows field names

	// Method 2: Positional (must match field order — less readable)
	p2 := Person{"Jane", "Doe", 28, "jane@example.com"}
	fmt.Printf("  p2: %+v\n", p2)

	// Method 3: Zero value struct (all fields get zero values)
	var p3 Person
	fmt.Printf("  p3 (zero): %+v\n", p3) // All empty/zero

	// Method 4: Partial initialization (unset fields get zero values)
	p4 := Person{
		FirstName: "Alice",
		Age:       25,
		// LastName and Email will be "" (zero value for string)
	}
	fmt.Printf("  p4 (partial): %+v\n", p4)

	// ========================================
	// Accessing and modifying fields
	// ========================================
	fmt.Println("\n--- Accessing Fields ---")

	fmt.Printf("  Name: %s %s\n", p1.FirstName, p1.LastName)
	fmt.Printf("  Age: %d\n", p1.Age)
	fmt.Printf("  Email: %s\n", p1.Email)

	// Modify a field directly
	p1.Age = 31
	p1.Email = "sri.new@example.com"
	fmt.Printf("  Updated: %s is now %d, email: %s\n",
		p1.FirstName, p1.Age, p1.Email)

	// ========================================
	// Pointers to structs
	// ========================================
	fmt.Println("\n--- Pointers to Structs ---")

	// Creating a pointer to a struct with &
	p5 := &Person{
		FirstName: "Bob",
		LastName:  "Smith",
		Age:       35,
	}
	fmt.Printf("  p5 (pointer): %+v\n", *p5)

	// Go lets you access fields through a pointer without dereferencing!
	// p5.FirstName is automatically (*p5).FirstName
	fmt.Printf("  p5.FirstName: %s\n", p5.FirstName)

	// When you pass a pointer, changes affect the original
	birthday(p5)
	fmt.Printf("  After birthday: %s is now %d\n", p5.FirstName, p5.Age)

	// ========================================
	// Nested structs
	// ========================================
	fmt.Println("\n--- Nested Structs ---")

	emp := Employee{
		Person: Person{
			FirstName: "Maya",
			LastName:  "Patel",
			Age:       29,
			Email:     "maya@company.com",
		},
		Department: "Engineering",
		Salary:     95000,
		HomeAddress: Address{
			Street:  "123 Main St",
			City:    "San Francisco",
			State:   "CA",
			ZipCode: "94102",
			Country: "USA",
		},
	}

	fmt.Printf("  Employee: %s %s\n", emp.Person.FirstName, emp.Person.LastName)
	fmt.Printf("  Department: %s\n", emp.Department)
	fmt.Printf("  City: %s, %s\n", emp.HomeAddress.City, emp.HomeAddress.State)

	// ========================================
	// Embedded (anonymous) structs
	// ========================================
	fmt.Println("\n--- Embedded Structs ---")

	student := Student{
		Person: Person{
			FirstName: "Alex",
			LastName:  "Kim",
			Age:       20,
			Email:     "alex@university.edu",
		},
		GPA:   3.8,
		Major: "Computer Science",
		Year:  3,
	}

	// Because Person is embedded, its fields are promoted.
	// You can access them directly on Student!
	fmt.Printf("  Name: %s %s\n", student.FirstName, student.LastName) // promoted!
	fmt.Printf("  Age: %d\n", student.Age)                             // promoted!
	fmt.Printf("  GPA: %.1f\n", student.GPA)
	fmt.Printf("  Major: %s (Year %d)\n", student.Major, student.Year)

	// You can still access via the embedded type name if needed
	fmt.Printf("  Via Person: %s\n", student.Person.Email)

	// ========================================
	// Anonymous structs (inline, one-off structs)
	// ========================================
	fmt.Println("\n--- Anonymous Structs ---")

	// Sometimes you need a quick struct without defining a type.
	// This is common for test data, configs, or one-time groupings.
	config := struct {
		Host    string
		Port    int
		Debug   bool
	}{
		Host:  "localhost",
		Port:  8080,
		Debug: true,
	}
	fmt.Printf("  Server: %s:%d (debug=%v)\n", config.Host, config.Port, config.Debug)

	// ========================================
	// Structs are value types (copies!)
	// ========================================
	fmt.Println("\n--- Structs Are Value Types ---")

	original := Person{FirstName: "Original", Age: 25}
	copied := original // This creates a COPY, not a reference

	copied.FirstName = "Copied"
	copied.Age = 99

	fmt.Printf("  original: %s, age %d\n", original.FirstName, original.Age) // unchanged!
	fmt.Printf("  copied:   %s, age %d\n", copied.FirstName, copied.Age)
	fmt.Println("  (Modifying the copy did NOT change the original)")

	// ========================================
	// Comparing structs
	// ========================================
	fmt.Println("\n--- Comparing Structs ---")

	a := Person{FirstName: "Go", LastName: "Lang", Age: 15}
	b := Person{FirstName: "Go", LastName: "Lang", Age: 15}
	c := Person{FirstName: "Go", LastName: "Lang", Age: 16}

	fmt.Printf("  a == b: %v\n", a == b) // true — all fields match
	fmt.Printf("  a == c: %v\n", a == c) // false — Age differs

	// ========================================
	// Struct with slice field
	// ========================================
	fmt.Println("\n--- Structs with Slices ---")

	playlist := Playlist{
		Name:  "Coding Vibes",
		Songs: []string{"Lofi Beat 1", "Synthwave Dream", "Chillhop Flow"},
		Likes: 42,
	}

	fmt.Printf("  Playlist: %s (%d likes)\n", playlist.Name, playlist.Likes)
	fmt.Println("  Songs:")
	for i, song := range playlist.Songs {
		fmt.Printf("    %d. %s\n", i+1, song)
	}

	// Add a song to the playlist
	playlist.Songs = append(playlist.Songs, "Deep Focus")
	fmt.Printf("  After adding: %d songs\n", len(playlist.Songs))

	// ========================================
	// Slice of structs
	// ========================================
	fmt.Println("\n--- Slice of Structs ---")

	team := []Person{
		{FirstName: "Alice", LastName: "A", Age: 30},
		{FirstName: "Bob", LastName: "B", Age: 25},
		{FirstName: "Charlie", LastName: "C", Age: 35},
	}

	fmt.Println("  Team members:")
	for _, member := range team {
		fmt.Printf("    %s %s (age %d)\n", member.FirstName, member.LastName, member.Age)
	}

	fmt.Println("\n========================================")
	fmt.Println("  Key Takeaways:")
	fmt.Println("  - Structs group related data together")
	fmt.Println("  - Use named fields for clarity")
	fmt.Println("  - Nested structs compose complex types")
	fmt.Println("  - Embedded structs promote fields")
	fmt.Println("  - Structs are value types (copying!)")
	fmt.Println("  - Use pointers (&) to modify originals")
	fmt.Println("  - Anonymous structs for quick one-offs")
	fmt.Println("========================================")
}

// birthday increments a person's age.
// Takes a pointer so it modifies the original Person.
func birthday(p *Person) {
	p.Age++
}

// Sample output:
//
// ========================================
//   Lesson: Structs in Go
// ========================================
//
// --- Creating Structs ---
//   p1: {FirstName:Sri LastName:Annamalai Age:30 Email:sri@example.com}
//   p2: {FirstName:Jane LastName:Doe Age:28 Email:jane@example.com}
//   p3 (zero): {FirstName: LastName: Age:0 Email:}
//   p4 (partial): {FirstName:Alice LastName: Age:25 Email:}
