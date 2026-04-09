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
// Week 11, Lesson 4: Graph Operations
// ========================================
// ArangoDB's graph capabilities let you model relationships between
// entities. Graphs consist of:
//   - Vertex collections (nodes): contain documents representing entities
//   - Edge collections (edges): contain documents representing relationships
//   - Named graphs: combine vertex and edge collections into a queryable graph
//
// Prerequisites:
//   1. ArangoDB running
//   2. The "learngo" database exists
//   3. cd week11 && go mod tidy
//
// Run:
//   go run ./04_graph/

// ========================================
// Data Models
// ========================================

// City represents a vertex (node) in our graph.
type City struct {
	Key        string `json:"_key,omitempty"`
	Name       string `json:"name"`
	Country    string `json:"country"`
	Population int    `json:"population"`
}

// Road represents an edge (relationship) between two cities.
// Edges MUST have _from and _to fields pointing to vertex document IDs.
type Road struct {
	Key      string `json:"_key,omitempty"`
	From     string `json:"_from"`
	To       string `json:"_to"`
	Distance int    `json:"distance"` // kilometers
	Type     string `json:"type"`     // highway, local, etc.
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 11: Graph Operations")
	fmt.Println("========================================")

	ctx := context.Background()
	client := connectToArangoDB()

	db, err := client.Database(ctx, "learngo")
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}
	fmt.Println("Connected to 'learngo' database!\n")

	// ========================================
	// 1. Create Vertex and Edge Collections
	// ========================================
	fmt.Println("--- Create Collections ---")
	cities, roads := createCollections(ctx, db)

	// ========================================
	// 2. Create the Named Graph
	// ========================================
	fmt.Println("\n--- Create Graph ---")
	graph := createGraph(ctx, db)

	// ========================================
	// 3. Insert Vertices (Cities)
	// ========================================
	fmt.Println("\n--- Insert Vertices (Cities) ---")
	insertCities(ctx, cities)

	// ========================================
	// 4. Insert Edges (Roads)
	// ========================================
	fmt.Println("\n--- Insert Edges (Roads) ---")
	insertRoads(ctx, roads)

	// ========================================
	// 5. Query the Graph — Direct Neighbors
	// ========================================
	fmt.Println("\n--- 5. Direct Neighbors of Chennai ---")
	directNeighbors(ctx, db)

	// ========================================
	// 6. Graph Traversal — Multi-hop
	// ========================================
	fmt.Println("\n--- 6. Cities within 2 Hops of Chennai ---")
	multiHopTraversal(ctx, db)

	// ========================================
	// 7. Shortest Path
	// ========================================
	fmt.Println("\n--- 7. Shortest Path: Chennai to Delhi ---")
	shortestPath(ctx, db)

	// ========================================
	// 8. All Paths Between Cities
	// ========================================
	fmt.Println("\n--- 8. All Paths: Chennai to Delhi (max 4 hops) ---")
	allPaths(ctx, db)

	// ========================================
	// 9. Graph Queries with Filters
	// ========================================
	fmt.Println("\n--- 9. Highway-Only Connections from Mumbai ---")
	filteredTraversal(ctx, db)

	// ========================================
	// Cleanup
	// ========================================
	fmt.Println("\n--- Cleanup ---")
	_ = graph
	if err := db.RemoveGraph(ctx, "city_roads", nil); err != nil {
		log.Printf("Failed to remove graph: %v\n", err)
	} else {
		fmt.Println("Graph 'city_roads' removed.")
	}
	cities.Remove(ctx)
	roads.Remove(ctx)
	fmt.Println("Collections removed.")
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
// Create Collections
// ========================================

func createCollections(ctx context.Context, db arangodb.Database) (arangodb.Collection, arangodb.Collection) {
	// Clean up existing collections
	for _, name := range []string{"cities", "roads"} {
		if exists, _ := db.CollectionExists(ctx, name); exists {
			col, _ := db.Collection(ctx, name)
			col.Remove(ctx)
		}
	}

	// Create VERTEX collection (regular document collection)
	cities, err := db.CreateCollection(ctx, "cities", &arangodb.CreateCollectionProperties{
		Type: arangodb.CollectionTypeDocument,
	})
	if err != nil {
		log.Fatalf("Failed to create cities: %v\n", err)
	}
	fmt.Println("Vertex collection 'cities' created!")

	// Create EDGE collection — this is special!
	// Edge collections store documents with _from and _to fields
	// that reference documents in vertex collections.
	roads, err := db.CreateCollection(ctx, "roads", &arangodb.CreateCollectionProperties{
		Type: arangodb.CollectionTypeEdge,
	})
	if err != nil {
		log.Fatalf("Failed to create roads: %v\n", err)
	}
	fmt.Println("Edge collection 'roads' created!")

	return cities, roads
}

// ========================================
// Create Named Graph
// ========================================

func createGraph(ctx context.Context, db arangodb.Database) arangodb.Graph {
	graphName := "city_roads"

	// Remove existing graph (without dropping collections)
	if exists, _ := db.GraphExists(ctx, graphName); exists {
		db.RemoveGraph(ctx, graphName, nil)
	}

	// ========================================
	// Define the graph structure
	// ========================================
	// A graph has edge definitions that specify:
	//   - Which edge collection connects which vertex collections
	//   - The "from" vertex collections
	//   - The "to" vertex collections

	edgeDefinitions := []arangodb.EdgeDefinition{
		{
			Collection: "roads",
			From:       []string{"cities"},
			To:         []string{"cities"},
		},
	}

	graph, err := db.CreateGraph(ctx, graphName, &arangodb.CreateGraphOptions{
		EdgeDefinitions: edgeDefinitions,
	})
	if err != nil {
		log.Fatalf("Failed to create graph: %v\n", err)
	}

	fmt.Printf("Graph '%s' created!\n", graphName)
	return graph
}

// ========================================
// Insert Vertices
// ========================================

func insertCities(ctx context.Context, col arangodb.Collection) {
	cities := []City{
		{Key: "chennai", Name: "Chennai", Country: "India", Population: 10900000},
		{Key: "mumbai", Name: "Mumbai", Country: "India", Population: 20700000},
		{Key: "delhi", Name: "Delhi", Country: "India", Population: 31200000},
		{Key: "bangalore", Name: "Bangalore", Country: "India", Population: 12800000},
		{Key: "hyderabad", Name: "Hyderabad", Country: "India", Population: 10500000},
		{Key: "kolkata", Name: "Kolkata", Country: "India", Population: 14900000},
	}

	for _, city := range cities {
		meta, err := col.CreateDocument(ctx, city)
		if err != nil {
			log.Printf("Failed to insert %s: %v\n", city.Name, err)
			continue
		}
		fmt.Printf("  Vertex: %s (key: %s)\n", city.Name, meta.Key)
	}
}

// ========================================
// Insert Edges
// ========================================

func insertRoads(ctx context.Context, col arangodb.Collection) {
	// ========================================
	// Edge documents MUST have _from and _to
	// ========================================
	// Format: "collection_name/document_key"
	// e.g., "cities/chennai"

	roads := []Road{
		{From: "cities/chennai", To: "cities/bangalore", Distance: 350, Type: "highway"},
		{From: "cities/bangalore", To: "cities/chennai", Distance: 350, Type: "highway"},
		{From: "cities/chennai", To: "cities/hyderabad", Distance: 630, Type: "highway"},
		{From: "cities/hyderabad", To: "cities/chennai", Distance: 630, Type: "highway"},
		{From: "cities/bangalore", To: "cities/mumbai", Distance: 980, Type: "highway"},
		{From: "cities/mumbai", To: "cities/bangalore", Distance: 980, Type: "highway"},
		{From: "cities/mumbai", To: "cities/delhi", Distance: 1400, Type: "highway"},
		{From: "cities/delhi", To: "cities/mumbai", Distance: 1400, Type: "highway"},
		{From: "cities/hyderabad", To: "cities/mumbai", Distance: 710, Type: "highway"},
		{From: "cities/mumbai", To: "cities/hyderabad", Distance: 710, Type: "highway"},
		{From: "cities/delhi", To: "cities/kolkata", Distance: 1530, Type: "highway"},
		{From: "cities/kolkata", To: "cities/delhi", Distance: 1530, Type: "highway"},
		{From: "cities/chennai", To: "cities/kolkata", Distance: 1660, Type: "local"},
		{From: "cities/kolkata", To: "cities/chennai", Distance: 1660, Type: "local"},
	}

	for _, road := range roads {
		_, err := col.CreateDocument(ctx, road)
		if err != nil {
			log.Printf("Failed to insert road %s -> %s: %v\n", road.From, road.To, err)
			continue
		}
	}
	fmt.Printf("  Inserted %d road edges (bidirectional)\n", len(roads))
}

// ========================================
// 5. Direct Neighbors
// ========================================

func directNeighbors(ctx context.Context, db arangodb.Database) {
	// ========================================
	// Graph Traversal with FOR ... IN ... OUTBOUND
	// ========================================
	// OUTBOUND: follow edges in the _from -> _to direction
	// INBOUND: follow edges in the _to -> _from direction
	// ANY: follow edges in both directions
	// 1..1 means exactly 1 hop (direct neighbors)

	query := `
		FOR city, road IN 1..1 OUTBOUND 'cities/chennai' roads
			RETURN { city: city.name, distance: road.distance, type: road.type }
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  Chennai -> %s (%v km, %s)\n",
			result["city"], result["distance"], result["type"])
	}
}

// ========================================
// 6. Multi-Hop Traversal
// ========================================

func multiHopTraversal(ctx context.Context, db arangodb.Database) {
	// 1..2 means 1 to 2 hops away
	query := `
		FOR city, road, path IN 1..2 OUTBOUND 'cities/chennai' roads
			OPTIONS { uniqueVertices: 'path' }
			RETURN DISTINCT {
				city: city.name,
				hops: LENGTH(path.edges),
				total_distance: SUM(path.edges[*].distance)
			}
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  %s - %v hop(s), %v km total\n",
			result["city"], result["hops"], result["total_distance"])
	}
}

// ========================================
// 7. Shortest Path
// ========================================

func shortestPath(ctx context.Context, db arangodb.Database) {
	// ========================================
	// SHORTEST_PATH finds the path with fewest edges
	// ========================================
	query := `
		FOR v, e IN OUTBOUND SHORTEST_PATH 'cities/chennai' TO 'cities/delhi'
			GRAPH 'city_roads'
			RETURN { city: v.name, road_distance: e.distance }
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	totalDistance := 0.0
	step := 0
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		if result["road_distance"] != nil {
			dist, _ := result["road_distance"].(float64)
			totalDistance += dist
			fmt.Printf("  Step %d: -> %s (%v km)\n", step, result["city"], result["road_distance"])
		} else {
			fmt.Printf("  Start: %s\n", result["city"])
		}
		step++
	}
	fmt.Printf("  Total distance: %.0f km\n", totalDistance)
}

// ========================================
// 8. All Paths
// ========================================

func allPaths(ctx context.Context, db arangodb.Database) {
	query := `
		FOR path IN 1..4 OUTBOUND K_PATHS 'cities/chennai' TO 'cities/delhi'
			GRAPH 'city_roads'
			RETURN {
				cities: path.vertices[*].name,
				distances: path.edges[*].distance,
				total: SUM(path.edges[*].distance),
				hops: LENGTH(path.edges)
			}
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	pathNum := 1
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  Path %d: %v (total: %v km, %v hops)\n",
			pathNum, result["cities"], result["total"], result["hops"])
		pathNum++
	}
}

// ========================================
// 9. Filtered Traversal
// ========================================

func filteredTraversal(ctx context.Context, db arangodb.Database) {
	// Filter edges during traversal — only follow highway roads
	query := `
		FOR city, road IN 1..1 OUTBOUND 'cities/mumbai' roads
			FILTER road.type == "highway"
			SORT road.distance ASC
			RETURN { city: city.name, distance: road.distance }
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  Mumbai -> %s (%v km, highway)\n", result["city"], result["distance"])
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// Graph Components:
//   Vertex collections — entities (nodes)
//   Edge collections   — relationships (must have _from, _to)
//   Named graph        — combines vertex + edge definitions
//
// Graph Traversal AQL:
//   FOR v, e, p IN min..max DIRECTION start_vertex edge_collection
//   - v: visited vertex
//   - e: edge used to reach vertex
//   - p: full path (p.vertices, p.edges)
//   - DIRECTION: OUTBOUND, INBOUND, ANY
//   - min..max: traversal depth range
//
// Path Finding:
//   SHORTEST_PATH start TO target GRAPH 'name'  — fewest edges
//   K_PATHS start TO target GRAPH 'name'        — all paths (with depth limit)
//
// Edge Document Format:
//   { "_from": "collection/key", "_to": "collection/key", ...other fields }
//
// Traversal Options:
//   uniqueVertices: 'path'  — don't revisit vertices in same path
//   uniqueEdges: 'path'     — don't reuse edges in same path (default)
