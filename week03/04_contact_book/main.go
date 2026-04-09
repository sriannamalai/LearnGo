package main

import (
	"fmt"
	"strings"
)

// ========================================
// Mini-Project: In-Memory Contact Book
// ========================================
// This project combines everything from Week 3:
//   - Structs (Contact, InMemoryStore)
//   - Methods (value and pointer receivers)
//   - Interfaces (ContactStorage for abstraction)
//   - Stringer interface for pretty printing
//   - Error handling patterns from Week 2

// ========================================
// Contact struct
// ========================================
// Each contact has a name, email, phone, and optional notes.
type Contact struct {
	Name  string
	Email string
	Phone string
	Notes string
}

// Implement the Stringer interface for pretty printing
func (c Contact) String() string {
	s := fmt.Sprintf("  Name:  %s\n  Email: %s\n  Phone: %s",
		c.Name, c.Email, c.Phone)
	if c.Notes != "" {
		s += fmt.Sprintf("\n  Notes: %s", c.Notes)
	}
	return s
}

// HasInfo checks if a contact matches a search query
// (case-insensitive search across all fields).
func (c Contact) HasInfo(query string) bool {
	query = strings.ToLower(query)
	return strings.Contains(strings.ToLower(c.Name), query) ||
		strings.Contains(strings.ToLower(c.Email), query) ||
		strings.Contains(strings.ToLower(c.Phone), query) ||
		strings.Contains(strings.ToLower(c.Notes), query)
}

// ========================================
// ContactStorage interface
// ========================================
// This interface defines what any contact storage must support.
// We could swap InMemoryStore for a database-backed store later
// without changing the code that uses ContactStorage.
type ContactStorage interface {
	Add(contact Contact) error
	Search(query string) []Contact
	Delete(name string) (Contact, error)
	List() []Contact
	Count() int
}

// ========================================
// InMemoryStore struct
// ========================================
// Implements ContactStorage using a simple slice.
type InMemoryStore struct {
	contacts []Contact
}

// NewInMemoryStore creates a new empty contact store.
// This is the constructor pattern — common in Go.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		contacts: make([]Contact, 0),
	}
}

// Add inserts a new contact into the store.
// Uses a pointer receiver because it modifies the store.
func (s *InMemoryStore) Add(contact Contact) error {
	// Validate the contact
	if strings.TrimSpace(contact.Name) == "" {
		return fmt.Errorf("contact name cannot be empty")
	}
	if strings.TrimSpace(contact.Email) == "" && strings.TrimSpace(contact.Phone) == "" {
		return fmt.Errorf("contact must have at least an email or phone number")
	}

	// Check for duplicate names (case-insensitive)
	for _, existing := range s.contacts {
		if strings.EqualFold(existing.Name, contact.Name) {
			return fmt.Errorf("contact %q already exists", contact.Name)
		}
	}

	s.contacts = append(s.contacts, contact)
	return nil
}

// Search finds contacts matching the query string.
// Uses a value receiver since it only reads the data.
// Returns a new slice (doesn't modify the store).
func (s *InMemoryStore) Search(query string) []Contact {
	if strings.TrimSpace(query) == "" {
		return nil
	}

	var results []Contact
	for _, c := range s.contacts {
		if c.HasInfo(query) {
			results = append(results, c)
		}
	}
	return results
}

// Delete removes a contact by name and returns the deleted contact.
// Uses a pointer receiver because it modifies the store.
func (s *InMemoryStore) Delete(name string) (Contact, error) {
	for i, c := range s.contacts {
		if strings.EqualFold(c.Name, name) {
			// Remove by swapping with last element and truncating
			// (order doesn't matter for a contact book)
			deleted := s.contacts[i]
			s.contacts[i] = s.contacts[len(s.contacts)-1]
			s.contacts = s.contacts[:len(s.contacts)-1]
			return deleted, nil
		}
	}
	return Contact{}, fmt.Errorf("contact %q not found", name)
}

// List returns all contacts in the store.
func (s *InMemoryStore) List() []Contact {
	// Return a copy to prevent external modification
	result := make([]Contact, len(s.contacts))
	copy(result, s.contacts)
	return result
}

// Count returns the number of contacts.
func (s *InMemoryStore) Count() int {
	return len(s.contacts)
}

// ========================================
// ContactBook: high-level wrapper
// ========================================
// ContactBook uses the ContactStorage interface, so it
// doesn't care HOW contacts are stored — just that the
// storage can Add, Search, Delete, and List.
type ContactBook struct {
	storage ContactStorage // interface — could be any implementation
	name    string
}

// NewContactBook creates a contact book with the given storage backend.
func NewContactBook(name string, storage ContactStorage) *ContactBook {
	return &ContactBook{
		storage: storage,
		name:    name,
	}
}

func (cb *ContactBook) String() string {
	return fmt.Sprintf("ContactBook(%q, %d contacts)", cb.name, cb.storage.Count())
}

// AddContact adds a contact and prints the result.
func (cb *ContactBook) AddContact(name, email, phone, notes string) {
	contact := Contact{
		Name:  name,
		Email: email,
		Phone: phone,
		Notes: notes,
	}

	err := cb.storage.Add(contact)
	if err != nil {
		fmt.Printf("  [ERROR] Could not add contact: %s\n", err)
	} else {
		fmt.Printf("  [OK] Added: %s\n", name)
	}
}

// SearchContacts searches and displays results.
func (cb *ContactBook) SearchContacts(query string) {
	results := cb.storage.Search(query)
	if len(results) == 0 {
		fmt.Printf("  No contacts found matching %q\n", query)
		return
	}

	fmt.Printf("  Found %d contact(s) matching %q:\n", len(results), query)
	for i, c := range results {
		fmt.Printf("\n  --- Result %d ---\n", i+1)
		fmt.Println(c)
	}
}

// DeleteContact removes a contact and confirms.
func (cb *ContactBook) DeleteContact(name string) {
	deleted, err := cb.storage.Delete(name)
	if err != nil {
		fmt.Printf("  [ERROR] %s\n", err)
	} else {
		fmt.Printf("  [OK] Deleted: %s\n", deleted.Name)
	}
}

// ListAll displays all contacts.
func (cb *ContactBook) ListAll() {
	contacts := cb.storage.List()
	if len(contacts) == 0 {
		fmt.Println("  Contact book is empty.")
		return
	}

	fmt.Printf("  All contacts (%d):\n", len(contacts))
	for i, c := range contacts {
		fmt.Printf("\n  --- Contact %d ---\n", i+1)
		fmt.Println(c)
	}
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Contact Book")
	fmt.Println("========================================")

	// ========================================
	// Create a contact book with in-memory storage
	// ========================================
	// The ContactBook accepts any ContactStorage — right now
	// we use InMemoryStore, but we could swap in a database
	// implementation without changing ContactBook at all.
	store := NewInMemoryStore()
	book := NewContactBook("My Contacts", store)
	fmt.Printf("  Created: %s\n", book)

	// ========================================
	// Add contacts
	// ========================================
	fmt.Println("\n--- Adding Contacts ---")

	book.AddContact("Sri Annamalai", "sri@example.com", "555-0101", "Go learner")
	book.AddContact("Jane Smith", "jane@example.com", "555-0102", "Colleague")
	book.AddContact("Bob Johnson", "bob@work.com", "555-0103", "")
	book.AddContact("Alice Chen", "alice@university.edu", "555-0104", "Study partner")
	book.AddContact("Maya Patel", "maya@startup.io", "555-0105", "Met at Go meetup")

	fmt.Printf("\n  %s\n", book)

	// ========================================
	// Try to add a duplicate
	// ========================================
	fmt.Println("\n--- Duplicate Detection ---")
	book.AddContact("sri annamalai", "other@email.com", "555-9999", "Duplicate!")

	// ========================================
	// Try to add invalid contacts
	// ========================================
	fmt.Println("\n--- Validation ---")
	book.AddContact("", "no@name.com", "555-0000", "")           // Empty name
	book.AddContact("No Contact Info", "", "", "Missing details") // No email or phone

	// ========================================
	// List all contacts
	// ========================================
	fmt.Println("\n--- List All Contacts ---")
	book.ListAll()

	// ========================================
	// Search for contacts
	// ========================================
	fmt.Println("\n--- Search by Name ---")
	book.SearchContacts("alice")

	fmt.Println("\n--- Search by Email Domain ---")
	book.SearchContacts("example.com")

	fmt.Println("\n--- Search by Notes ---")
	book.SearchContacts("meetup")

	fmt.Println("\n--- Search with No Results ---")
	book.SearchContacts("xyz123")

	// ========================================
	// Delete a contact
	// ========================================
	fmt.Println("\n--- Delete a Contact ---")
	book.DeleteContact("Bob Johnson")
	fmt.Printf("  Contacts remaining: %d\n", store.Count())

	// Try to delete someone who doesn't exist
	book.DeleteContact("Nobody Here")

	// ========================================
	// List again after deletion
	// ========================================
	fmt.Println("\n--- List After Deletion ---")
	book.ListAll()

	// ========================================
	// Demonstrate interface polymorphism
	// ========================================
	fmt.Println("\n--- Interface in Action ---")
	fmt.Println("  The ContactBook works with any ContactStorage.")
	fmt.Println("  We used InMemoryStore, but could swap in:")
	fmt.Println("    - FileStore (save to JSON file)")
	fmt.Println("    - SQLStore (use a database)")
	fmt.Println("    - APIStore (call a REST API)")
	fmt.Println("  ...without changing ContactBook at all!")

	// ========================================
	// Interactive mode
	// ========================================
	fmt.Println("\n========================================")
	fmt.Println("  Interactive Mode")
	fmt.Println("========================================")
	fmt.Println("  Commands: add, search, delete, list, quit")

	for {
		fmt.Print("\n> Enter command: ")
		var cmd string
		fmt.Scan(&cmd)

		switch strings.ToLower(cmd) {
		case "add":
			var name, email, phone string
			fmt.Print("  Name: ")
			fmt.Scan(&name)
			fmt.Print("  Email: ")
			fmt.Scan(&email)
			fmt.Print("  Phone: ")
			fmt.Scan(&phone)
			book.AddContact(name, email, phone, "")

		case "search":
			var query string
			fmt.Print("  Search query: ")
			fmt.Scan(&query)
			book.SearchContacts(query)

		case "delete":
			var name string
			fmt.Print("  Name to delete: ")
			fmt.Scan(&name)
			book.DeleteContact(name)

		case "list":
			book.ListAll()

		case "quit", "exit", "q":
			fmt.Println("\n  Goodbye! Your contacts were not saved (in-memory only).")
			fmt.Println("\n========================================")
			fmt.Println("  Concepts Used in This Project:")
			fmt.Println("  - Structs: Contact, InMemoryStore, ContactBook")
			fmt.Println("  - Methods: value and pointer receivers")
			fmt.Println("  - Interfaces: ContactStorage abstraction")
			fmt.Println("  - Stringer: String() for pretty printing")
			fmt.Println("  - Error handling: validation, not-found")
			fmt.Println("  - Constructor pattern: NewInMemoryStore()")
			fmt.Println("  - Composition: ContactBook has-a ContactStorage")
			fmt.Println("========================================")
			return

		default:
			fmt.Printf("  Unknown command %q. Try: add, search, delete, list, quit\n", cmd)
		}
	}
}

// Sample output:
//
// ========================================
//   Contact Book
// ========================================
//   Created: ContactBook("My Contacts", 0 contacts)
//
// --- Adding Contacts ---
//   [OK] Added: Sri Annamalai
//   [OK] Added: Jane Smith
//   [OK] Added: Bob Johnson
//   [OK] Added: Alice Chen
//   [OK] Added: Maya Patel
//
//   ContactBook("My Contacts", 5 contacts)
//
// --- Duplicate Detection ---
//   [ERROR] Could not add contact: contact "sri annamalai" already exists
//
// --- Search by Name ---
//   Found 1 contact(s) matching "alice":
//
//   --- Result 1 ---
//   Name:  Alice Chen
//   Email: alice@university.edu
//   Phone: 555-0104
//   Notes: Study partner
