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
// Week 11, Lesson 1: Connecting to ArangoDB
// ========================================
// ArangoDB is a multi-model database: documents, graphs, and key-value
// all in one. This lesson covers connecting and creating databases.
//
// Prerequisites:
//   1. ArangoDB installed and running
//      - Docker: docker run -d --name arangodb -p 8529:8529 -e ARANGO_ROOT_PASSWORD=rootpassword arangodb:latest
//      - macOS: brew install arangodb
//   2. Web UI available at http://localhost:8529
//   3. cd week11 && go mod tidy
//
// Set connection details (optional — defaults work with Docker setup):
//   export ARANGO_URL="http://localhost:8529"
//   export ARANGO_PASSWORD="rootpassword"
//
// Run:
//   go run ./01_connect/

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 11: Connecting to ArangoDB")
	fmt.Println("========================================")

	ctx := context.Background()

	// ========================================
	// 1. Connection Configuration
	// ========================================
	// ArangoDB uses HTTP(S) for its API. The Go driver wraps this
	// into a convenient client library.

	arangoURL := os.Getenv("ARANGO_URL")
	if arangoURL == "" {
		arangoURL = "http://localhost:8529"
	}
	password := os.Getenv("ARANGO_PASSWORD")
	if password == "" {
		password = "rootpassword"
	}

	fmt.Printf("Connecting to ArangoDB at %s\n", arangoURL)

	// ========================================
	// 2. Create HTTP Connection
	// ========================================
	// The connection handles HTTP communication with ArangoDB.

	endpoint := connection.NewRoundRobinEndpoints([]string{arangoURL})
	conn := connection.NewHttpConnection(connection.HttpConfiguration{
		Endpoint:       endpoint,
		Authentication: connection.NewBasicAuth("root", password),
		ContentType:    connection.ApplicationJSON,
	})

	// ========================================
	// 3. Create the ArangoDB Client
	// ========================================
	// The client provides high-level methods for database operations.

	client := arangodb.NewClient(conn)

	fmt.Println("Client created successfully!")

	// ========================================
	// 4. Get Server Version (verify connection)
	// ========================================
	version, err := client.Version(ctx)
	if err != nil {
		log.Fatalf("Failed to get server version: %v\n", err)
	}

	fmt.Printf("\nServer Information:\n")
	fmt.Printf("  Version: %s\n", version.Version)
	fmt.Printf("  Server:  %s\n", version.Server)
	fmt.Printf("  License: %s\n", version.License)

	// ========================================
	// 5. List Existing Databases
	// ========================================
	fmt.Println("\n--- Existing Databases ---")

	databases, err := client.Databases(ctx)
	if err != nil {
		log.Fatalf("Failed to list databases: %v\n", err)
	}

	for _, db := range databases {
		info, _ := db.Info(ctx)
		fmt.Printf("  - %s (ID: %s)\n", info.Name, info.ID)
	}

	// ========================================
	// 6. Create a New Database
	// ========================================
	fmt.Println("\n--- Create Database ---")

	dbName := "learngo"

	// Check if database already exists
	exists, err := client.DatabaseExists(ctx, dbName)
	if err != nil {
		log.Fatalf("Failed to check database existence: %v\n", err)
	}

	var db arangodb.Database
	if exists {
		fmt.Printf("Database '%s' already exists, using it.\n", dbName)
		db, err = client.Database(ctx, dbName)
		if err != nil {
			log.Fatalf("Failed to open database: %v\n", err)
		}
	} else {
		fmt.Printf("Creating database '%s'...\n", dbName)
		db, err = client.CreateDatabase(ctx, dbName, nil)
		if err != nil {
			log.Fatalf("Failed to create database: %v\n", err)
		}
		fmt.Printf("Database '%s' created successfully!\n", dbName)
	}

	// ========================================
	// 7. Verify the database
	// ========================================
	dbInfo, err := db.Info(ctx)
	if err != nil {
		log.Fatalf("Failed to get database info: %v\n", err)
	}
	fmt.Printf("\nActive Database:\n")
	fmt.Printf("  Name: %s\n", dbInfo.Name)
	fmt.Printf("  ID:   %s\n", dbInfo.ID)

	// ========================================
	// 8. List collections in the database
	// ========================================
	fmt.Println("\n--- Collections ---")
	collections, err := db.Collections(ctx)
	if err != nil {
		log.Fatalf("Failed to list collections: %v\n", err)
	}

	if len(collections) == 0 {
		fmt.Println("  No collections yet (we'll create some in the next lessons)")
	} else {
		for _, col := range collections {
			fmt.Printf("  - %s\n", col.Name())
		}
	}

	fmt.Println("\n========================================")
	fmt.Println("  Connection successful!")
	fmt.Println("  ArangoDB Web UI: http://localhost:8529")
	fmt.Println("========================================")
}

// ========================================
// Key Concepts Recap
// ========================================
//
// ArangoDB Architecture:
//   - Multi-model: documents + graphs + key-value in one database
//   - Databases contain collections (like tables in SQL)
//   - Collections contain documents (JSON objects)
//   - Uses AQL (ArangoDB Query Language) for queries
//
// Go Driver Setup:
//   1. Create a connection (HTTP endpoint + authentication)
//   2. Create a client (wraps connection with high-level API)
//   3. Open or create a database
//   4. Work with collections and documents
//
// Connection Pattern:
//   endpoint := connection.NewRoundRobinEndpoints([]string{url})
//   conn := connection.NewHttpConnection(config)
//   client := arangodb.NewClient(conn)
//
// Database Operations:
//   client.Databases(ctx)           — list all databases
//   client.DatabaseExists(ctx, name) — check existence
//   client.Database(ctx, name)      — open existing database
//   client.CreateDatabase(ctx, name, opts) — create new database
