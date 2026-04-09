package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
)

// ========================================
// Week 11, Lesson 2: Document CRUD Operations
// ========================================
// ArangoDB stores data as JSON documents in collections.
// This lesson covers creating collections and performing
// CRUD (Create, Read, Update, Delete) on documents.
//
// Prerequisites:
//   1. ArangoDB running (see lesson 01 for setup)
//   2. The "learngo" database exists (created in lesson 01)
//   3. cd week11 && go mod tidy
//
// Run:
//   go run ./02_documents/

// ========================================
// Data Model
// ========================================
// In ArangoDB, every document has a _key (unique within collection),
// _id (collection_name/_key), and _rev (revision for optimistic locking).

// Product represents a document in our "products" collection.
type Product struct {
	Key         string  `json:"_key,omitempty"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	InStock     bool    `json:"in_stock"`
	Tags        []string `json:"tags,omitempty"`
	Description string  `json:"description,omitempty"`
}

// DocumentMeta holds ArangoDB metadata returned after operations.
type DocumentMeta struct {
	Key string `json:"_key"`
	ID  string `json:"_id"`
	Rev string `json:"_rev"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 11: Document CRUD Operations")
	fmt.Println("========================================")

	ctx := context.Background()

	// ========================================
	// Connect to ArangoDB
	// ========================================
	client := connectToArangoDB()
	db, err := client.Database(ctx, "learngo")
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}
	fmt.Println("Connected to 'learngo' database!\n")

	// ========================================
	// 1. Create a Collection
	// ========================================
	fmt.Println("--- Create Collection ---")
	col := createCollection(ctx, db)

	// ========================================
	// 2. INSERT Documents (Create)
	// ========================================
	fmt.Println("\n--- Insert Documents ---")
	insertDocuments(ctx, col)

	// ========================================
	// 3. READ Documents (by key)
	// ========================================
	fmt.Println("\n--- Read Documents ---")
	readDocuments(ctx, col)

	// ========================================
	// 4. UPDATE Documents
	// ========================================
	fmt.Println("\n--- Update Documents ---")
	updateDocuments(ctx, col)

	// ========================================
	// 5. REPLACE Documents
	// ========================================
	fmt.Println("\n--- Replace Document ---")
	replaceDocument(ctx, col)

	// ========================================
	// 6. DELETE Documents
	// ========================================
	fmt.Println("\n--- Delete Document ---")
	deleteDocument(ctx, col)

	// ========================================
	// 7. Batch Operations
	// ========================================
	fmt.Println("\n--- Batch Insert ---")
	batchInsert(ctx, col)

	// ========================================
	// 8. Show final collection count
	// ========================================
	fmt.Println("\n--- Collection Status ---")
	count, err := col.Count(ctx)
	if err != nil {
		log.Printf("Failed to count: %v\n", err)
	} else {
		fmt.Printf("Total documents in 'products': %d\n", count)
	}

	// ========================================
	// Cleanup
	// ========================================
	fmt.Println("\n--- Cleanup ---")
	if err := col.Remove(ctx); err != nil {
		log.Printf("Failed to remove collection: %v\n", err)
	} else {
		fmt.Println("Collection 'products' removed.")
	}
}

// ========================================
// Connect Helper
// ========================================

func connectToArangoDB() arangodb.Client {
	arangoURL := os.Getenv("ARANGO_URL")
	if arangoURL == "" {
		arangoURL = "http://localhost:8529"
	}
	password := os.Getenv("ARANGO_PASSWORD")
	if password == "" {
		password = "rootpassword"
	}

	endpoint := connection.NewRoundRobinEndpoints([]string{arangoURL})
	conn := connection.NewHttpConnection(connection.HttpConfiguration{
		Endpoint:       endpoint,
		Authentication: connection.NewBasicAuth("root", password),
		ContentType:    connection.ApplicationJSON,
	})

	return arangodb.NewClient(conn)
}

// ========================================
// Create Collection
// ========================================

func createCollection(ctx context.Context, db arangodb.Database) arangodb.Collection {
	colName := "products"

	// Check if collection exists
	exists, err := db.CollectionExists(ctx, colName)
	if err != nil {
		log.Fatalf("Failed to check collection: %v\n", err)
	}

	if exists {
		// Drop and recreate for a clean demo
		col, _ := db.Collection(ctx, colName)
		col.Remove(ctx)
		fmt.Printf("Dropped existing '%s' collection.\n", colName)
	}

	// Create a new document collection
	// CollectionTypeDocument is for regular documents (vs. CollectionTypeEdge for graphs)
	col, err := db.CreateCollection(ctx, colName, &arangodb.CreateCollectionProperties{
		Type: arangodb.CollectionTypeDocument,
	})
	if err != nil {
		log.Fatalf("Failed to create collection: %v\n", err)
	}

	fmt.Printf("Collection '%s' created!\n", colName)
	return col
}

// ========================================
// Insert Documents
// ========================================

func insertDocuments(ctx context.Context, col arangodb.Collection) {
	// ========================================
	// Insert a single document
	// ========================================
	// CreateDocument inserts one document and returns its metadata.

	keyboard := Product{
		Key:      "keyboard-1",
		Name:     "Mechanical Keyboard",
		Category: "peripherals",
		Price:    89.99,
		InStock:  true,
		Tags:     []string{"mechanical", "rgb", "wireless"},
	}

	meta, err := col.CreateDocument(ctx, keyboard)
	if err != nil {
		log.Fatalf("Failed to insert document: %v\n", err)
	}
	fmt.Printf("Inserted: key=%s, id=%s, rev=%s\n", meta.Key, meta.ID.String(), meta.Rev)

	// Insert more documents
	products := []Product{
		{
			Key:      "mouse-1",
			Name:     "Ergonomic Mouse",
			Category: "peripherals",
			Price:    49.99,
			InStock:  true,
			Tags:     []string{"ergonomic", "wireless"},
		},
		{
			Key:      "monitor-1",
			Name:     "4K Monitor",
			Category: "displays",
			Price:    399.99,
			InStock:  true,
			Tags:     []string{"4k", "ips", "27-inch"},
		},
		{
			Key:         "webcam-1",
			Name:        "HD Webcam",
			Category:    "peripherals",
			Price:       79.99,
			InStock:     false,
			Description: "1080p webcam with built-in microphone",
		},
	}

	for _, p := range products {
		meta, err := col.CreateDocument(ctx, p)
		if err != nil {
			log.Printf("Failed to insert %s: %v\n", p.Name, err)
			continue
		}
		fmt.Printf("Inserted: %s (key: %s)\n", p.Name, meta.Key)
	}
}

// ========================================
// Read Documents
// ========================================

func readDocuments(ctx context.Context, col arangodb.Collection) {
	// ========================================
	// Read by key
	// ========================================
	// ReadDocument fetches a document by its _key.

	var product Product
	meta, err := col.ReadDocument(ctx, "keyboard-1", &product)
	if err != nil {
		log.Printf("Failed to read document: %v\n", err)
		return
	}

	fmt.Printf("Read document (key=%s, rev=%s):\n", meta.Key, meta.Rev)
	fmt.Printf("  Name:     %s\n", product.Name)
	fmt.Printf("  Category: %s\n", product.Category)
	fmt.Printf("  Price:    $%.2f\n", product.Price)
	fmt.Printf("  In Stock: %v\n", product.InStock)
	fmt.Printf("  Tags:     %v\n", product.Tags)

	// ========================================
	// Check if a document exists
	// ========================================
	exists, err := col.DocumentExists(ctx, "keyboard-1")
	if err != nil {
		log.Printf("Failed to check existence: %v\n", err)
	} else {
		fmt.Printf("\nDocument 'keyboard-1' exists: %v\n", exists)
	}

	exists, err = col.DocumentExists(ctx, "nonexistent-key")
	if err != nil {
		log.Printf("Failed to check existence: %v\n", err)
	} else {
		fmt.Printf("Document 'nonexistent-key' exists: %v\n", exists)
	}
}

// ========================================
// Update Documents (partial update)
// ========================================

func updateDocuments(ctx context.Context, col arangodb.Collection) {
	// ========================================
	// Partial Update
	// ========================================
	// UpdateDocument merges the provided fields with the existing document.
	// Fields not included in the update are left unchanged.

	// Only update the price and stock status
	update := map[string]any{
		"price":    74.99,
		"in_stock": true,
	}

	meta, err := col.UpdateDocument(ctx, "webcam-1", update)
	if err != nil {
		log.Printf("Failed to update document: %v\n", err)
		return
	}
	fmt.Printf("Updated 'webcam-1' (new rev: %s)\n", meta.Rev)

	// Verify the update
	var product Product
	_, err = col.ReadDocument(ctx, "webcam-1", &product)
	if err != nil {
		log.Printf("Failed to read updated doc: %v\n", err)
		return
	}
	fmt.Printf("  Price is now: $%.2f\n", product.Price)
	fmt.Printf("  In stock: %v\n", product.InStock)
	fmt.Printf("  Name unchanged: %s\n", product.Name)
}

// ========================================
// Replace Document (full replacement)
// ========================================

func replaceDocument(ctx context.Context, col arangodb.Collection) {
	// ========================================
	// Full Replace
	// ========================================
	// ReplaceDocument replaces the ENTIRE document (except _key, _id).
	// Any fields not in the replacement are REMOVED.

	replacement := Product{
		Name:     "Premium 4K Monitor",
		Category: "displays",
		Price:    449.99,
		InStock:  true,
		Tags:     []string{"4k", "ips", "32-inch", "hdr"},
	}

	meta, err := col.ReplaceDocument(ctx, "monitor-1", replacement)
	if err != nil {
		log.Printf("Failed to replace document: %v\n", err)
		return
	}
	fmt.Printf("Replaced 'monitor-1' (new rev: %s)\n", meta.Rev)

	// Verify
	var product Product
	_, err = col.ReadDocument(ctx, "monitor-1", &product)
	if err != nil {
		log.Printf("Failed to read: %v\n", err)
		return
	}
	fmt.Printf("  Name: %s\n", product.Name)
	fmt.Printf("  Price: $%.2f\n", product.Price)
	fmt.Printf("  Tags: %v\n", product.Tags)
}

// ========================================
// Delete Document
// ========================================

func deleteDocument(ctx context.Context, col arangodb.Collection) {
	// ========================================
	// Delete by key
	// ========================================

	_, err := col.DeleteDocument(ctx, "mouse-1")
	if err != nil {
		log.Printf("Failed to delete: %v\n", err)
		return
	}
	fmt.Println("Deleted document 'mouse-1'")

	// Verify it's gone
	exists, _ := col.DocumentExists(ctx, "mouse-1")
	fmt.Printf("Document 'mouse-1' exists after delete: %v\n", exists)
}

// ========================================
// Batch Operations
// ========================================

func batchInsert(ctx context.Context, col arangodb.Collection) {
	// ========================================
	// Insert multiple documents at once
	// ========================================
	// CreateDocuments (plural) inserts a batch of documents efficiently.

	accessories := []Product{
		{Key: "cable-1", Name: "USB-C Cable", Category: "accessories", Price: 9.99, InStock: true},
		{Key: "stand-1", Name: "Laptop Stand", Category: "accessories", Price: 34.99, InStock: true},
		{Key: "pad-1", Name: "Mouse Pad XL", Category: "accessories", Price: 19.99, InStock: true},
	}

	metas, errs, err := col.CreateDocuments(ctx, accessories)
	if err != nil {
		log.Printf("Batch insert failed: %v\n", err)
		return
	}

	fmt.Printf("Batch inserted %d documents:\n", len(metas))
	for i, meta := range metas {
		if errs[i] != nil {
			fmt.Printf("  Error for doc %d: %v\n", i, errs[i])
		} else {
			fmt.Printf("  [%d] key=%s\n", i, meta.Key)
		}
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Document Structure:
//   Every document has system fields:
//   _key — unique identifier within a collection (string)
//   _id  — globally unique: "collection_name/_key"
//   _rev — revision string for optimistic locking
//
// CRUD Operations:
//   col.CreateDocument(ctx, doc)      — Insert one document
//   col.ReadDocument(ctx, key, &doc)  — Read by key
//   col.UpdateDocument(ctx, key, partial) — Partial update (merge)
//   col.ReplaceDocument(ctx, key, doc)    — Full replace
//   col.DeleteDocument(ctx, key)      — Delete by key
//
// Batch Operations:
//   col.CreateDocuments(ctx, docs)    — Insert multiple
//   col.UpdateDocuments(ctx, docs)    — Update multiple
//   col.DeleteDocuments(ctx, keys)    — Delete multiple
//
// Update vs Replace:
//   Update: merges fields (unmentioned fields stay)
//   Replace: replaces entire document (unmentioned fields are removed)
//
// Collection Types:
//   CollectionTypeDocument (default) — for regular documents
//   CollectionTypeEdge              — for graph edges
