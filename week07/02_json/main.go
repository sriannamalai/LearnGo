package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ========================================
// Week 7, Lesson 2: JSON in Go
// ========================================
// Go's encoding/json package provides powerful tools for working
// with JSON data. This lesson covers marshaling (Go -> JSON),
// unmarshaling (JSON -> Go), struct tags, encoders/decoders,
// and advanced techniques like custom marshaling.
// ========================================

func main() {
	// ========================================
	// 1. json.Marshal — Go to JSON
	// ========================================
	fmt.Println("========================================")
	fmt.Println("1. json.Marshal — Go to JSON")
	fmt.Println("========================================")

	// json.Marshal converts a Go value to JSON bytes.
	// Only exported fields (capitalized) are included in JSON output.

	person := Person{
		FirstName: "Alice",
		LastName:  "Johnson",
		Age:       30,
		Email:     "alice@example.com",
	}

	jsonBytes, err := json.Marshal(person)
	if err != nil {
		fmt.Printf("Error marshaling: %v\n", err)
		return
	}
	fmt.Printf("  JSON: %s\n", string(jsonBytes))

	// Marshaling basic types
	fmt.Println("\n  Marshaling various types:")

	numJSON, _ := json.Marshal(42)
	fmt.Printf("  int:    %s\n", numJSON)

	strJSON, _ := json.Marshal("hello")
	fmt.Printf("  string: %s\n", strJSON)

	boolJSON, _ := json.Marshal(true)
	fmt.Printf("  bool:   %s\n", boolJSON)

	sliceJSON, _ := json.Marshal([]string{"Go", "Rust", "Python"})
	fmt.Printf("  slice:  %s\n", sliceJSON)

	mapJSON, _ := json.Marshal(map[string]int{"a": 1, "b": 2})
	fmt.Printf("  map:    %s\n", mapJSON)

	// ========================================
	// 2. Struct Tags
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("2. Struct Tags")
	fmt.Println("========================================")

	// Struct tags control how fields are marshaled/unmarshaled.
	// Format: `json:"fieldname,options"`

	user := User{
		ID:        1,
		Username:  "gopher",
		Email:     "gopher@go.dev",
		Password:  "secret123",   // Won't appear in JSON (json:"-")
		FirstName: "",            // Won't appear in JSON (omitempty)
		LastName:  "The Gopher",
		IsActive:  true,
	}

	userJSON, _ := json.Marshal(user)
	fmt.Printf("  User JSON: %s\n", string(userJSON))
	// Expected output:
	// Notice: no "password" field, no "first_name" (empty + omitempty),
	// field names use snake_case from tags

	// ========================================
	// 3. json.Unmarshal — JSON to Go
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("3. json.Unmarshal — JSON to Go")
	fmt.Println("========================================")

	// json.Unmarshal parses JSON bytes into a Go value.
	// You must pass a pointer so the function can modify the value.

	jsonStr := `{
		"first_name": "Bob",
		"last_name": "Smith",
		"age": 25,
		"email": "bob@example.com"
	}`

	var decoded Person
	err = json.Unmarshal([]byte(jsonStr), &decoded) // Pass pointer!
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("  Decoded: %+v\n", decoded)
	fmt.Printf("  Name: %s %s, Age: %d\n",
		decoded.FirstName, decoded.LastName, decoded.Age)

	// Unmarshaling into a map (when you don't know the structure)
	fmt.Println("\n  Unmarshaling into map[string]any:")
	var generic map[string]any
	json.Unmarshal([]byte(jsonStr), &generic)
	for key, val := range generic {
		fmt.Printf("    %s: %v (type: %T)\n", key, val, val)
	}
	// Note: JSON numbers become float64 by default in map[string]any

	// ========================================
	// 4. json.NewEncoder / json.NewDecoder
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("4. json.NewEncoder / json.NewDecoder")
	fmt.Println("========================================")

	// Encoders and Decoders work with io.Writer / io.Reader streams.
	// They're ideal for HTTP responses, files, or any stream.

	// Encode to a buffer (simulating a writer)
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.Encode(person) // Writes JSON + newline to buffer
	fmt.Printf("  Encoded: %s", buf.String())

	// Decode from a reader
	reader := strings.NewReader(`{"first_name":"Charlie","last_name":"Brown","age":8}`)
	decoder := json.NewDecoder(reader)

	var charlie Person
	err = decoder.Decode(&charlie)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("  Decoded: %s %s, Age: %d\n",
		charlie.FirstName, charlie.LastName, charlie.Age)

	// ========================================
	// 5. Pretty Printing
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("5. Pretty Printing")
	fmt.Println("========================================")

	// json.MarshalIndent adds indentation for readable output.

	team := Team{
		Name: "Go Team",
		Members: []Person{
			{FirstName: "Alice", LastName: "A", Age: 30, Email: "alice@go.dev"},
			{FirstName: "Bob", LastName: "B", Age: 25, Email: "bob@go.dev"},
		},
		Founded: 2009,
	}

	pretty, _ := json.MarshalIndent(team, "", "  ")
	fmt.Printf("  Pretty JSON:\n%s\n", string(pretty))

	// Encoder with indentation
	fmt.Println("\n  Encoder with indentation:")
	var prettyBuf bytes.Buffer
	enc := json.NewEncoder(&prettyBuf)
	enc.SetIndent("", "    ") // 4-space indent
	enc.Encode(person)
	fmt.Printf("%s", prettyBuf.String())

	// ========================================
	// 6. Nested JSON
	// ========================================
	fmt.Println("========================================")
	fmt.Println("6. Nested JSON")
	fmt.Println("========================================")

	// Go structs naturally map to nested JSON objects.

	order := Order{
		ID:     "ORD-001",
		Status: "shipped",
		Customer: Customer{
			Name:  "Alice Johnson",
			Email: "alice@example.com",
			Address: Address{
				Street: "123 Go Lane",
				City:   "Gopher City",
				State:  "CA",
				Zip:    "90210",
			},
		},
		Items: []OrderItem{
			{Product: "Go Book", Quantity: 1, Price: 29.99},
			{Product: "Go Stickers", Quantity: 5, Price: 2.99},
		},
		Total: 44.94,
	}

	orderJSON, _ := json.MarshalIndent(order, "", "  ")
	fmt.Printf("  Order JSON:\n%s\n", string(orderJSON))

	// Unmarshal nested JSON
	var decoded_order Order
	json.Unmarshal(orderJSON, &decoded_order)
	fmt.Printf("\n  Decoded order: %s\n", decoded_order.ID)
	fmt.Printf("  Customer: %s\n", decoded_order.Customer.Name)
	fmt.Printf("  City: %s\n", decoded_order.Customer.Address.City)
	fmt.Printf("  Items: %d\n", len(decoded_order.Items))

	// ========================================
	// 7. omitempty
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("7. omitempty")
	fmt.Println("========================================")

	// `omitempty` omits the field from JSON output if it has
	// its zero value (0 for numbers, "" for strings, nil for
	// pointers/slices/maps, false for bools).

	type Settings struct {
		Theme    string   `json:"theme,omitempty"`
		FontSize int      `json:"font_size,omitempty"`
		Debug    bool     `json:"debug,omitempty"`
		Tags     []string `json:"tags,omitempty"`
	}

	// With values set
	full := Settings{Theme: "dark", FontSize: 14, Debug: true, Tags: []string{"dev"}}
	fullJSON, _ := json.Marshal(full)
	fmt.Printf("  Full:    %s\n", fullJSON)

	// With zero values — omitted fields won't appear
	sparse := Settings{Theme: "light"} // Only theme set
	sparseJSON, _ := json.Marshal(sparse)
	fmt.Printf("  Sparse:  %s\n", sparseJSON)

	// Completely empty — all fields omitted
	empty := Settings{}
	emptyJSON, _ := json.Marshal(empty)
	fmt.Printf("  Empty:   %s\n", emptyJSON)

	// Using pointers to distinguish "not set" from "zero value"
	type NullableSettings struct {
		Theme    *string `json:"theme,omitempty"`
		FontSize *int    `json:"font_size,omitempty"`
	}

	// nil pointer is omitted; pointer to zero value is included
	zero := 0
	ns := NullableSettings{FontSize: &zero} // FontSize is 0 but present
	nsJSON, _ := json.Marshal(ns)
	fmt.Printf("  Nullable: %s\n", nsJSON)
	// Output: {"font_size":0} — FontSize appears because the pointer is non-nil

	// ========================================
	// 8. Custom Marshaling
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("8. Custom Marshaling")
	fmt.Println("========================================")

	// Implement json.Marshaler and json.Unmarshaler interfaces
	// for custom JSON representation.

	event := Event{
		Name: "Go Conference",
		Date: time.Date(2025, 6, 15, 9, 0, 0, 0, time.UTC),
	}

	eventJSON, _ := json.MarshalIndent(event, "", "  ")
	fmt.Printf("  Event JSON:\n%s\n", string(eventJSON))

	// Unmarshal the custom format back
	customJSON := `{"name":"Workshop","date":"2025-12-01"}`
	var workshop Event
	err = json.Unmarshal([]byte(customJSON), &workshop)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Unmarshaled: %s on %s\n",
			workshop.Name, workshop.Date.Format("January 2, 2006"))
	}

	// ========================================
	// 9. Working with Unknown/Dynamic JSON
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("9. Dynamic JSON with json.RawMessage")
	fmt.Println("========================================")

	// json.RawMessage lets you delay parsing part of the JSON.
	// Useful when the structure varies.

	messages := []string{
		`{"type":"text","payload":{"content":"Hello!"}}`,
		`{"type":"number","payload":{"value":42}}`,
	}

	for _, raw := range messages {
		var envelope Envelope
		json.Unmarshal([]byte(raw), &envelope)

		fmt.Printf("  Type: %s\n", envelope.Type)

		switch envelope.Type {
		case "text":
			var text TextPayload
			json.Unmarshal(envelope.Payload, &text)
			fmt.Printf("  Content: %s\n", text.Content)
		case "number":
			var num NumberPayload
			json.Unmarshal(envelope.Payload, &num)
			fmt.Printf("  Value: %d\n", num.Value)
		}
	}

	// ========================================
	// Summary
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println("- json.Marshal / MarshalIndent: Go struct -> JSON bytes")
	fmt.Println("- json.Unmarshal: JSON bytes -> Go struct (pass pointer!)")
	fmt.Println("- Struct tags: `json:\"name,omitempty\"` control field names")
	fmt.Println("- json:\"-\" excludes a field from JSON entirely")
	fmt.Println("- json.NewEncoder/Decoder: stream-based JSON I/O")
	fmt.Println("- Nested structs map naturally to nested JSON")
	fmt.Println("- Implement MarshalJSON/UnmarshalJSON for custom formats")
	fmt.Println("- json.RawMessage delays parsing for dynamic structures")
}

// ========================================
// Types
// ========================================

// Person demonstrates basic JSON marshaling.
type Person struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       int    `json:"age"`
	Email     string `json:"email"`
}

// User demonstrates struct tags with various options.
type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`                    // Never include in JSON
	FirstName string `json:"first_name,omitempty"` // Omit if empty
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
}

// Team demonstrates nested structures.
type Team struct {
	Name    string   `json:"name"`
	Members []Person `json:"members"`
	Founded int      `json:"founded"`
}

// Order and related types demonstrate deeply nested JSON.
type Order struct {
	ID       string      `json:"id"`
	Status   string      `json:"status"`
	Customer Customer    `json:"customer"`
	Items    []OrderItem `json:"items"`
	Total    float64     `json:"total"`
}

type Customer struct {
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Address Address `json:"address"`
}

type Address struct {
	Street string `json:"street"`
	City   string `json:"city"`
	State  string `json:"state"`
	Zip    string `json:"zip"`
}

type OrderItem struct {
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// Event demonstrates custom marshaling with a date field.
type Event struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

// MarshalJSON implements json.Marshaler for custom date format.
func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event // Prevent infinite recursion
	return json.Marshal(struct {
		Alias
		Date string `json:"date"`
	}{
		Alias: Alias(e),
		Date:  e.Date.Format("2006-01-02"), // YYYY-MM-DD format
	})
}

// UnmarshalJSON implements json.Unmarshaler for custom date format.
func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := struct {
		Alias
		Date string `json:"date"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.Name = aux.Name
	parsed, err := time.Parse("2006-01-02", aux.Date)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}
	e.Date = parsed
	return nil
}

// Envelope demonstrates json.RawMessage for dynamic payloads.
type Envelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type TextPayload struct {
	Content string `json:"content"`
}

type NumberPayload struct {
	Value int `json:"value"`
}
