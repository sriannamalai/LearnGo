package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/connection"
)

// ========================================
// Week 11, Mini-Project: Social Network Graph
// ========================================
// A social network modeled as a graph in ArangoDB.
// Users are vertices, friendships are edges.
// Demonstrates graph traversal for finding friends-of-friends
// and recommending new connections.
//
// Prerequisites:
//   1. ArangoDB running
//   2. The "learngo" database exists
//   3. cd week11 && go mod tidy
//
// Run:
//   go run ./05_social_network/
//
// Commands:
//   users             — List all users
//   friends <name>    — Show user's friends
//   fof <name>        — Friends-of-friends
//   recommend <name>  — Recommend new friends
//   mutual <a> <b>    — Mutual friends between two users
//   path <a> <b>      — Shortest connection path
//   popular           — Most connected users
//   add <name> <age> <city>  — Add a new user
//   connect <a> <b>   — Create a friendship
//   quit              — Exit

// ========================================
// Data Models
// ========================================

// User represents a person in the social network.
type User struct {
	Key       string   `json:"_key,omitempty"`
	Name      string   `json:"name"`
	Age       int      `json:"age"`
	City      string   `json:"city"`
	Interests []string `json:"interests,omitempty"`
}

// Friendship represents a relationship between two users.
type Friendship struct {
	From  string `json:"_from"`
	To    string `json:"_to"`
	Since string `json:"since,omitempty"`
}

// ========================================
// SocialNetwork manages the graph database
// ========================================

type SocialNetwork struct {
	db      arangodb.Database
	users   arangodb.Collection
	friends arangodb.Collection
}

func NewSocialNetwork(db arangodb.Database) (*SocialNetwork, error) {
	ctx := context.Background()
	sn := &SocialNetwork{db: db}

	// ========================================
	// Setup Collections
	// ========================================

	// Users vertex collection
	if exists, _ := db.CollectionExists(ctx, "social_users"); exists {
		col, _ := db.Collection(ctx, "social_users")
		col.Remove(ctx)
	}
	users, err := db.CreateCollection(ctx, "social_users", &arangodb.CreateCollectionProperties{
		Type: arangodb.CollectionTypeDocument,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create users collection: %w", err)
	}
	sn.users = users

	// Friendships edge collection
	if exists, _ := db.CollectionExists(ctx, "social_friends"); exists {
		col, _ := db.Collection(ctx, "social_friends")
		col.Remove(ctx)
	}
	friends, err := db.CreateCollection(ctx, "social_friends", &arangodb.CreateCollectionProperties{
		Type: arangodb.CollectionTypeEdge,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create friends collection: %w", err)
	}
	sn.friends = friends

	// ========================================
	// Create Named Graph
	// ========================================
	graphName := "social_network"
	if exists, _ := db.GraphExists(ctx, graphName); exists {
		db.RemoveGraph(ctx, graphName, nil)
	}

	_, err = db.CreateGraph(ctx, graphName, &arangodb.CreateGraphOptions{
		EdgeDefinitions: []arangodb.EdgeDefinition{
			{
				Collection: "social_friends",
				From:       []string{"social_users"},
				To:         []string{"social_users"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create graph: %w", err)
	}

	return sn, nil
}

// ========================================
// Add User
// ========================================

func (sn *SocialNetwork) AddUser(ctx context.Context, user User) error {
	// Use name as key (lowercased, no spaces)
	user.Key = strings.ToLower(strings.ReplaceAll(user.Name, " ", "_"))
	_, err := sn.users.CreateDocument(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to add user: %w", err)
	}
	return nil
}

// ========================================
// Create Friendship (bidirectional)
// ========================================

func (sn *SocialNetwork) AddFriendship(ctx context.Context, user1Key, user2Key, since string) error {
	// Create edges in both directions for bidirectional friendship
	edge1 := Friendship{
		From:  "social_users/" + user1Key,
		To:    "social_users/" + user2Key,
		Since: since,
	}
	edge2 := Friendship{
		From:  "social_users/" + user2Key,
		To:    "social_users/" + user1Key,
		Since: since,
	}

	if _, err := sn.friends.CreateDocument(ctx, edge1); err != nil {
		return fmt.Errorf("failed to create friendship: %w", err)
	}
	if _, err := sn.friends.CreateDocument(ctx, edge2); err != nil {
		return fmt.Errorf("failed to create reverse friendship: %w", err)
	}
	return nil
}

// ========================================
// List All Users
// ========================================

func (sn *SocialNetwork) ListUsers(ctx context.Context) {
	query := `
		FOR u IN social_users
			LET friendCount = LENGTH(
				FOR f IN social_friends
					FILTER f._from == CONCAT('social_users/', u._key)
					RETURN 1
			)
			SORT u.name
			RETURN { name: u.name, age: u.age, city: u.city, friends: friendCount }
	`

	cursor, err := sn.db.Query(ctx, query, nil)
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	defer cursor.Close()

	fmt.Printf("  %-15s %4s %-15s %s\n", "Name", "Age", "City", "Friends")
	fmt.Println("  --------------- ---- --------------- -------")

	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  %-15s %4.0f %-15s %.0f\n",
			result["name"], result["age"], result["city"], result["friends"])
	}
}

// ========================================
// Get Friends of a User
// ========================================

func (sn *SocialNetwork) GetFriends(ctx context.Context, userKey string) {
	query := `
		FOR friend, edge IN 1..1 OUTBOUND CONCAT('social_users/', @userKey) social_friends
			RETURN { name: friend.name, city: friend.city, since: edge.since }
	`

	cursor, err := sn.db.Query(ctx, query, &arangodb.QueryOptions{
		BindVars: map[string]interface{}{"userKey": userKey},
	})
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	defer cursor.Close()

	count := 0
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  %s (%s) - friends since %s\n",
			result["name"], result["city"], result["since"])
		count++
	}
	if count == 0 {
		fmt.Println("  No friends found.")
	}
}

// ========================================
// Friends of Friends (2 hops away)
// ========================================

func (sn *SocialNetwork) FriendsOfFriends(ctx context.Context, userKey string) {
	// ========================================
	// Graph Traversal: exactly 2 hops away
	// ========================================
	// Find people who are friends with your friends but NOT your direct friends
	// and NOT yourself.

	query := `
		LET directFriends = (
			FOR f IN 1..1 OUTBOUND CONCAT('social_users/', @userKey) social_friends
				RETURN f._key
		)
		FOR fof IN 2..2 OUTBOUND CONCAT('social_users/', @userKey) social_friends
			OPTIONS { uniqueVertices: 'path' }
			FILTER fof._key != @userKey
			FILTER fof._key NOT IN directFriends
			COLLECT name = fof.name, city = fof.city WITH COUNT INTO count
			SORT count DESC
			RETURN { name: name, city: city, mutual_connections: count }
	`

	cursor, err := sn.db.Query(ctx, query, &arangodb.QueryOptions{
		BindVars: map[string]interface{}{"userKey": userKey},
	})
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	defer cursor.Close()

	count := 0
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  %s (%s) - %v mutual connection(s)\n",
			result["name"], result["city"], result["mutual_connections"])
		count++
	}
	if count == 0 {
		fmt.Println("  No friends-of-friends found.")
	}
}

// ========================================
// Friend Recommendations
// ========================================

func (sn *SocialNetwork) RecommendFriends(ctx context.Context, userKey string) {
	// ========================================
	// Smart Recommendations
	// ========================================
	// Recommend users based on:
	//   1. Number of mutual friends (most important)
	//   2. Same city (bonus)
	// Exclude: self, existing friends

	query := `
		LET user = DOCUMENT(CONCAT('social_users/', @userKey))
		LET directFriends = (
			FOR f IN 1..1 OUTBOUND CONCAT('social_users/', @userKey) social_friends
				RETURN f._key
		)
		FOR candidate IN social_users
			FILTER candidate._key != @userKey
			FILTER candidate._key NOT IN directFriends
			LET mutualFriends = (
				FOR f IN 1..1 OUTBOUND CONCAT('social_users/', candidate._key) social_friends
					FILTER f._key IN directFriends
					RETURN f.name
			)
			LET sameCity = candidate.city == user.city ? 1 : 0
			LET score = LENGTH(mutualFriends) * 10 + sameCity * 5
			FILTER score > 0
			SORT score DESC
			RETURN {
				name: candidate.name,
				city: candidate.city,
				mutual_friends: mutualFriends,
				mutual_count: LENGTH(mutualFriends),
				same_city: sameCity == 1,
				score: score
			}
	`

	cursor, err := sn.db.Query(ctx, query, &arangodb.QueryOptions{
		BindVars: map[string]interface{}{"userKey": userKey},
	})
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	defer cursor.Close()

	count := 0
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		cityNote := ""
		if sameCity, ok := result["same_city"].(bool); ok && sameCity {
			cityNote = " [same city]"
		}
		fmt.Printf("  %s (%s) - %v mutual friend(s): %v%s\n",
			result["name"], result["city"],
			result["mutual_count"], result["mutual_friends"], cityNote)
		count++
	}
	if count == 0 {
		fmt.Println("  No recommendations available.")
	}
}

// ========================================
// Mutual Friends
// ========================================

func (sn *SocialNetwork) MutualFriends(ctx context.Context, user1Key, user2Key string) {
	query := `
		LET friends1 = (
			FOR f IN 1..1 OUTBOUND CONCAT('social_users/', @user1) social_friends
				RETURN f
		)
		LET friends2 = (
			FOR f IN 1..1 OUTBOUND CONCAT('social_users/', @user2) social_friends
				RETURN f._key
		)
		FOR f IN friends1
			FILTER f._key IN friends2
			RETURN { name: f.name, city: f.city }
	`

	cursor, err := sn.db.Query(ctx, query, &arangodb.QueryOptions{
		BindVars: map[string]interface{}{
			"user1": user1Key,
			"user2": user2Key,
		},
	})
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	defer cursor.Close()

	count := 0
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  %s (%s)\n", result["name"], result["city"])
		count++
	}
	if count == 0 {
		fmt.Println("  No mutual friends.")
	} else {
		fmt.Printf("  --- %d mutual friend(s) ---\n", count)
	}
}

// ========================================
// Shortest Path Between Users
// ========================================

func (sn *SocialNetwork) ShortestPath(ctx context.Context, fromKey, toKey string) {
	query := `
		FOR v IN OUTBOUND SHORTEST_PATH
			CONCAT('social_users/', @from) TO CONCAT('social_users/', @to)
			GRAPH 'social_network'
			RETURN v.name
	`

	cursor, err := sn.db.Query(ctx, query, &arangodb.QueryOptions{
		BindVars: map[string]interface{}{
			"from": fromKey,
			"to":   toKey,
		},
	})
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	defer cursor.Close()

	var path []string
	for cursor.HasMore() {
		var name string
		_, err := cursor.ReadDocument(ctx, &name)
		if err != nil {
			break
		}
		path = append(path, name)
	}

	if len(path) == 0 {
		fmt.Println("  No connection found.")
	} else {
		fmt.Printf("  %s\n", strings.Join(path, " -> "))
		fmt.Printf("  (%d degrees of separation)\n", len(path)-1)
	}
}

// ========================================
// Most Popular Users
// ========================================

func (sn *SocialNetwork) MostPopular(ctx context.Context) {
	query := `
		FOR u IN social_users
			LET friendCount = LENGTH(
				FOR f IN social_friends
					FILTER f._from == CONCAT('social_users/', u._key)
					RETURN 1
			)
			SORT friendCount DESC
			RETURN { name: u.name, city: u.city, friends: friendCount }
	`

	cursor, err := sn.db.Query(ctx, query, nil)
	if err != nil {
		log.Printf("Query failed: %v\n", err)
		return
	}
	defer cursor.Close()

	rank := 1
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  #%d: %s (%s) - %.0f friends\n",
			rank, result["name"], result["city"], result["friends"])
		rank++
	}
}

// ========================================
// Cleanup
// ========================================

func (sn *SocialNetwork) Cleanup(ctx context.Context) {
	sn.db.RemoveGraph(ctx, "social_network", nil)
	sn.users.Remove(ctx)
	sn.friends.Remove(ctx)
}

// ========================================
// Main
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 11 Project: Social Network Graph")
	fmt.Println("========================================")

	ctx := context.Background()
	client := connectToArangoDB()

	db, err := client.Database(ctx, "learngo")
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}

	// Initialize the social network
	sn, err := NewSocialNetwork(db)
	if err != nil {
		log.Fatalf("Failed to initialize social network: %v\n", err)
	}

	fmt.Println("Social network initialized!\n")

	// ========================================
	// Seed sample data
	// ========================================
	fmt.Println("--- Seeding Sample Data ---")
	seedSampleData(ctx, sn)

	fmt.Println()
	printHelp()

	// ========================================
	// Interactive Command Loop
	// ========================================
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nsocial> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "users":
			sn.ListUsers(ctx)

		case "friends":
			if len(parts) < 2 {
				fmt.Println("Usage: friends <name>")
				continue
			}
			key := strings.ToLower(strings.Join(parts[1:], "_"))
			fmt.Printf("Friends of %s:\n", parts[1])
			sn.GetFriends(ctx, key)

		case "fof":
			if len(parts) < 2 {
				fmt.Println("Usage: fof <name>")
				continue
			}
			key := strings.ToLower(strings.Join(parts[1:], "_"))
			fmt.Printf("Friends-of-friends for %s:\n", parts[1])
			sn.FriendsOfFriends(ctx, key)

		case "recommend":
			if len(parts) < 2 {
				fmt.Println("Usage: recommend <name>")
				continue
			}
			key := strings.ToLower(strings.Join(parts[1:], "_"))
			fmt.Printf("Friend recommendations for %s:\n", parts[1])
			sn.RecommendFriends(ctx, key)

		case "mutual":
			if len(parts) < 3 {
				fmt.Println("Usage: mutual <name1> <name2>")
				continue
			}
			key1 := strings.ToLower(parts[1])
			key2 := strings.ToLower(parts[2])
			fmt.Printf("Mutual friends of %s and %s:\n", parts[1], parts[2])
			sn.MutualFriends(ctx, key1, key2)

		case "path":
			if len(parts) < 3 {
				fmt.Println("Usage: path <name1> <name2>")
				continue
			}
			key1 := strings.ToLower(parts[1])
			key2 := strings.ToLower(parts[2])
			fmt.Printf("Path from %s to %s:\n", parts[1], parts[2])
			sn.ShortestPath(ctx, key1, key2)

		case "popular":
			fmt.Println("Most connected users:")
			sn.MostPopular(ctx)

		case "add":
			if len(parts) < 4 {
				fmt.Println("Usage: add <name> <age> <city>")
				continue
			}
			name := parts[1]
			age := 0
			fmt.Sscanf(parts[2], "%d", &age)
			city := parts[3]
			err := sn.AddUser(ctx, User{Name: name, Age: age, City: city})
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Added user: %s\n", name)
			}

		case "connect":
			if len(parts) < 3 {
				fmt.Println("Usage: connect <name1> <name2>")
				continue
			}
			key1 := strings.ToLower(parts[1])
			key2 := strings.ToLower(parts[2])
			err := sn.AddFriendship(ctx, key1, key2, "2024")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Connected %s and %s!\n", parts[1], parts[2])
			}

		case "help":
			printHelp()

		case "quit", "exit", "q":
			fmt.Println("\nCleaning up...")
			sn.Cleanup(ctx)
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmd)
		}
	}

	// Cleanup on EOF
	sn.Cleanup(ctx)
}

// ========================================
// Seed Sample Data
// ========================================

func seedSampleData(ctx context.Context, sn *SocialNetwork) {
	// Add users
	users := []User{
		{Name: "Alice", Age: 30, City: "Chennai", Interests: []string{"Go", "databases", "music"}},
		{Name: "Bob", Age: 28, City: "Chennai", Interests: []string{"Python", "ML", "gaming"}},
		{Name: "Charlie", Age: 35, City: "Mumbai", Interests: []string{"Go", "DevOps", "hiking"}},
		{Name: "Diana", Age: 32, City: "Bangalore", Interests: []string{"Rust", "systems", "cooking"}},
		{Name: "Eve", Age: 27, City: "Chennai", Interests: []string{"JS", "React", "travel"}},
		{Name: "Frank", Age: 40, City: "Delhi", Interests: []string{"Java", "architecture", "chess"}},
		{Name: "Grace", Age: 29, City: "Mumbai", Interests: []string{"Go", "cloud", "photography"}},
		{Name: "Hank", Age: 33, City: "Bangalore", Interests: []string{"Python", "AI", "running"}},
	}

	for _, u := range users {
		if err := sn.AddUser(ctx, u); err != nil {
			log.Printf("Failed to add %s: %v\n", u.Name, err)
		} else {
			fmt.Printf("  Added: %s (%s)\n", u.Name, u.City)
		}
	}

	// Create friendships
	friendships := []struct {
		a, b, since string
	}{
		{"alice", "bob", "2020"},
		{"alice", "charlie", "2021"},
		{"alice", "eve", "2019"},
		{"bob", "diana", "2022"},
		{"bob", "eve", "2020"},
		{"charlie", "grace", "2021"},
		{"charlie", "frank", "2023"},
		{"diana", "hank", "2022"},
		{"eve", "grace", "2023"},
		{"frank", "hank", "2021"},
	}

	for _, f := range friendships {
		if err := sn.AddFriendship(ctx, f.a, f.b, f.since); err != nil {
			log.Printf("Failed to connect %s and %s: %v\n", f.a, f.b, err)
		}
	}
	fmt.Printf("  Created %d friendships.\n", len(friendships))
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

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  users             - List all users")
	fmt.Println("  friends <name>    - Show user's friends")
	fmt.Println("  fof <name>        - Friends-of-friends")
	fmt.Println("  recommend <name>  - Recommend new friends")
	fmt.Println("  mutual <a> <b>    - Mutual friends")
	fmt.Println("  path <a> <b>      - Shortest connection path")
	fmt.Println("  popular           - Most connected users")
	fmt.Println("  add <name> <age> <city> - Add user")
	fmt.Println("  connect <a> <b>   - Create friendship")
	fmt.Println("  help              - Show this help")
	fmt.Println("  quit              - Exit")
}

// ========================================
// Sample Session
// ========================================
//
// social> users
//   Name            Age  City            Friends
//   --------------- ---- --------------- -------
//   Alice            30  Chennai         3
//   Bob              28  Chennai         3
//   Charlie          35  Mumbai          3
//   Diana            32  Bangalore       2
//   Eve              27  Chennai         3
//   Frank            40  Delhi           2
//   Grace            29  Mumbai          2
//   Hank             33  Bangalore       2
//
// social> friends alice
// Friends of alice:
//   Bob (Chennai) - friends since 2020
//   Charlie (Mumbai) - friends since 2021
//   Eve (Chennai) - friends since 2019
//
// social> fof alice
// Friends-of-friends for alice:
//   Diana (Bangalore) - 1 mutual connection(s)
//   Grace (Mumbai) - 2 mutual connection(s)
//   Frank (Delhi) - 1 mutual connection(s)
//
// social> recommend alice
// Friend recommendations for alice:
//   Grace (Mumbai) - 2 mutual friend(s): ["Charlie", "Eve"]
//   Diana (Bangalore) - 1 mutual friend(s): ["Bob"]
//   Frank (Delhi) - 1 mutual friend(s): ["Charlie"]
//
// social> mutual bob eve
// Mutual friends of bob and eve:
//   Alice (Chennai)
//   --- 1 mutual friend(s) ---
//
// social> path alice hank
// Path from alice to hank:
//   Alice -> Bob -> Diana -> Hank
//   (3 degrees of separation)
//
// social> popular
// Most connected users:
//   #1: Alice (Chennai) - 3 friends
//   #2: Bob (Chennai) - 3 friends
//   #3: Charlie (Mumbai) - 3 friends
//   #4: Eve (Chennai) - 3 friends
//   #5: Diana (Bangalore) - 2 friends
//
// social> quit
// Cleaning up...
// Goodbye!
