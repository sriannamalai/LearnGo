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
// Week 11, Lesson 3: AQL Queries from Go
// ========================================
// AQL (ArangoDB Query Language) is ArangoDB's query language.
// It's similar to SQL but designed for documents and graphs.
//
// Prerequisites:
//   1. ArangoDB running
//   2. The "learngo" database exists
//   3. cd week11 && go mod tidy
//
// Run:
//   go run ./03_aql/

// ========================================
// Data Models
// ========================================

// Employee represents an employee document.
type Employee struct {
	Key        string  `json:"_key,omitempty"`
	Name       string  `json:"name"`
	Department string  `json:"department"`
	Salary     float64 `json:"salary"`
	Age        int     `json:"age"`
	Skills     []string `json:"skills,omitempty"`
	Active     bool    `json:"active"`
}

// DeptStats represents aggregated department statistics.
type DeptStats struct {
	Department string  `json:"department"`
	Count      int     `json:"count"`
	AvgSalary  float64 `json:"avg_salary"`
	MinSalary  float64 `json:"min_salary"`
	MaxSalary  float64 `json:"max_salary"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 11: AQL Queries from Go")
	fmt.Println("========================================")

	ctx := context.Background()
	client := connectToArangoDB()

	db, err := client.Database(ctx, "learngo")
	if err != nil {
		log.Fatalf("Failed to open database: %v\n", err)
	}
	fmt.Println("Connected to 'learngo' database!\n")

	// ========================================
	// Setup: Create collection and seed data
	// ========================================
	fmt.Println("--- Setup ---")
	col := setupCollection(ctx, db)
	seedData(ctx, col)

	// ========================================
	// 1. Simple Query — Get All Documents
	// ========================================
	fmt.Println("\n--- 1. Simple Query: All Employees ---")
	simpleQuery(ctx, db)

	// ========================================
	// 2. Filtered Query — WHERE clause
	// ========================================
	fmt.Println("\n--- 2. Filtered Query: Engineering Department ---")
	filteredQuery(ctx, db)

	// ========================================
	// 3. Parameterized Query (SAFE from injection)
	// ========================================
	fmt.Println("\n--- 3. Parameterized Query ---")
	parameterizedQuery(ctx, db, "Engineering", 80000)

	// ========================================
	// 4. Sorting and Limiting
	// ========================================
	fmt.Println("\n--- 4. Sort and Limit: Top 3 by Salary ---")
	sortAndLimitQuery(ctx, db)

	// ========================================
	// 5. Aggregation Queries
	// ========================================
	fmt.Println("\n--- 5. Aggregation: Department Statistics ---")
	aggregationQuery(ctx, db)

	// ========================================
	// 6. Update via AQL
	// ========================================
	fmt.Println("\n--- 6. Update via AQL: 10% Raise for Engineering ---")
	updateViaAQL(ctx, db)

	// ========================================
	// 7. Complex Query with Multiple Conditions
	// ========================================
	fmt.Println("\n--- 7. Complex Query: Active, Skilled Employees ---")
	complexQuery(ctx, db)

	// ========================================
	// 8. Subquery / LET
	// ========================================
	fmt.Println("\n--- 8. LET / Subquery: Enriched Results ---")
	subquery(ctx, db)

	// ========================================
	// Cleanup
	// ========================================
	fmt.Println("\n--- Cleanup ---")
	col.Remove(ctx)
	fmt.Println("Collection 'employees' removed.")
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
// Setup
// ========================================

func setupCollection(ctx context.Context, db arangodb.Database) arangodb.Collection {
	exists, _ := db.CollectionExists(ctx, "employees")
	if exists {
		col, _ := db.Collection(ctx, "employees")
		col.Remove(ctx)
	}

	col, err := db.CreateCollection(ctx, "employees", &arangodb.CreateCollectionProperties{
		Type: arangodb.CollectionTypeDocument,
	})
	if err != nil {
		log.Fatalf("Failed to create collection: %v\n", err)
	}
	fmt.Println("Collection 'employees' created!")
	return col
}

func seedData(ctx context.Context, col arangodb.Collection) {
	employees := []Employee{
		{Key: "alice", Name: "Alice Johnson", Department: "Engineering", Salary: 95000, Age: 30, Skills: []string{"Go", "Python", "Docker"}, Active: true},
		{Key: "bob", Name: "Bob Smith", Department: "Engineering", Salary: 88000, Age: 28, Skills: []string{"Go", "Kubernetes"}, Active: true},
		{Key: "charlie", Name: "Charlie Brown", Department: "Marketing", Salary: 72000, Age: 35, Skills: []string{"SEO", "Analytics"}, Active: true},
		{Key: "diana", Name: "Diana Prince", Department: "Engineering", Salary: 105000, Age: 32, Skills: []string{"Go", "Rust", "C++"}, Active: true},
		{Key: "eve", Name: "Eve Wilson", Department: "Marketing", Salary: 68000, Age: 27, Skills: []string{"Content", "Social Media"}, Active: true},
		{Key: "frank", Name: "Frank Miller", Department: "Sales", Salary: 75000, Age: 40, Skills: []string{"Negotiation", "CRM"}, Active: false},
		{Key: "grace", Name: "Grace Lee", Department: "Engineering", Salary: 92000, Age: 29, Skills: []string{"Go", "React", "TypeScript"}, Active: true},
		{Key: "hank", Name: "Hank Davis", Department: "Sales", Salary: 82000, Age: 38, Skills: []string{"Leadership", "CRM"}, Active: true},
	}

	for _, emp := range employees {
		_, err := col.CreateDocument(ctx, emp)
		if err != nil {
			log.Printf("Failed to insert %s: %v\n", emp.Name, err)
		}
	}
	fmt.Printf("Seeded %d employees.\n", len(employees))
}

// ========================================
// 1. Simple Query
// ========================================

func simpleQuery(ctx context.Context, db arangodb.Database) {
	// ========================================
	// Basic AQL: FOR ... RETURN
	// ========================================
	// AQL uses FOR to iterate over a collection.
	// RETURN specifies what to output.

	query := `FOR e IN employees RETURN e`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	// Iterate over results
	for cursor.HasMore() {
		var emp Employee
		_, err := cursor.ReadDocument(ctx, &emp)
		if err != nil {
			log.Printf("Read error: %v\n", err)
			break
		}
		fmt.Printf("  %s - %s ($%.0f)\n", emp.Name, emp.Department, emp.Salary)
	}
}

// ========================================
// 2. Filtered Query
// ========================================

func filteredQuery(ctx context.Context, db arangodb.Database) {
	// ========================================
	// AQL FILTER — like SQL WHERE
	// ========================================
	query := `
		FOR e IN employees
			FILTER e.department == "Engineering"
			SORT e.salary DESC
			RETURN e
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	for cursor.HasMore() {
		var emp Employee
		_, err := cursor.ReadDocument(ctx, &emp)
		if err != nil {
			break
		}
		fmt.Printf("  %s - $%.0f - Skills: %v\n", emp.Name, emp.Salary, emp.Skills)
	}
}

// ========================================
// 3. Parameterized Query (SAFE!)
// ========================================

func parameterizedQuery(ctx context.Context, db arangodb.Database, dept string, minSalary float64) {
	// ========================================
	// ALWAYS use bind variables for user input!
	// ========================================
	// Like SQL parameter binding, this prevents AQL injection.
	// Use @varName in the query and pass values in the bindVars map.

	query := `
		FOR e IN employees
			FILTER e.department == @dept
			FILTER e.salary >= @minSalary
			FILTER e.active == true
			SORT e.salary DESC
			RETURN { name: e.name, salary: e.salary, skills: e.skills }
	`

	// Bind variables — the safe way to pass parameters
	bindVars := map[string]interface{}{
		"dept":      dept,
		"minSalary": minSalary,
	}

	cursor, err := db.Query(ctx, query, &arangodb.QueryOptions{
		BindVars: bindVars,
	})
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	fmt.Printf("Active %s employees earning >= $%.0f:\n", dept, minSalary)
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  %s - $%.0f\n", result["name"], result["salary"])
	}
}

// ========================================
// 4. Sort and Limit
// ========================================

func sortAndLimitQuery(ctx context.Context, db arangodb.Database) {
	query := `
		FOR e IN employees
			SORT e.salary DESC
			LIMIT 3
			RETURN { name: e.name, department: e.department, salary: e.salary }
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	rank := 1
	for cursor.HasMore() {
		var result map[string]any
		_, err := cursor.ReadDocument(ctx, &result)
		if err != nil {
			break
		}
		fmt.Printf("  #%d: %s (%s) - $%.0f\n", rank, result["name"], result["department"], result["salary"])
		rank++
	}
}

// ========================================
// 5. Aggregation
// ========================================

func aggregationQuery(ctx context.Context, db arangodb.Database) {
	// ========================================
	// COLLECT — AQL's GROUP BY equivalent
	// ========================================
	query := `
		FOR e IN employees
			COLLECT dept = e.department
			AGGREGATE count = LENGTH(1),
			          avgSalary = AVG(e.salary),
			          minSalary = MIN(e.salary),
			          maxSalary = MAX(e.salary)
			SORT avgSalary DESC
			RETURN {
				department: dept,
				count: count,
				avg_salary: avgSalary,
				min_salary: minSalary,
				max_salary: maxSalary
			}
	`

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		log.Fatalf("Query failed: %v\n", err)
	}
	defer cursor.Close()

	fmt.Printf("%-15s %5s %10s %10s %10s\n", "Department", "Count", "Avg", "Min", "Max")
	fmt.Println("--------------- ----- ---------- ---------- ----------")

	for cursor.HasMore() {
		var stats DeptStats
		_, err := cursor.ReadDocument(ctx, &stats)
		if err != nil {
			break
		}
		fmt.Printf("%-15s %5d $%9.0f $%9.0f $%9.0f\n",
			stats.Department, stats.Count, stats.AvgSalary, stats.MinSalary, stats.MaxSalary)
	}
}

// ========================================
// 6. Update via AQL
// ========================================

func updateViaAQL(ctx context.Context, db arangodb.Database) {
	// ========================================
	// UPDATE in AQL
	// ========================================
	// You can update documents directly in AQL queries.

	query := `
		FOR e IN employees
			FILTER e.department == "Engineering"
			UPDATE e WITH { salary: e.salary * 1.10 } IN employees
			RETURN { name: e.name, old_salary: e.salary, new_salary: e.salary * 1.10 }
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
		fmt.Printf("  %s: $%.0f -> $%.0f\n",
			result["name"], result["old_salary"], result["new_salary"])
	}
}

// ========================================
// 7. Complex Query
// ========================================

func complexQuery(ctx context.Context, db arangodb.Database) {
	// Multiple filters, array operations, and computed fields
	query := `
		FOR e IN employees
			FILTER e.active == true
			FILTER LENGTH(e.skills) >= 2
			LET skillCount = LENGTH(e.skills)
			SORT skillCount DESC, e.salary DESC
			RETURN {
				name: e.name,
				department: e.department,
				salary: e.salary,
				skill_count: skillCount,
				skills: e.skills
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
		fmt.Printf("  %s (%s) - %v skills: %v\n",
			result["name"], result["department"], result["skill_count"], result["skills"])
	}
}

// ========================================
// 8. Subquery with LET
// ========================================

func subquery(ctx context.Context, db arangodb.Database) {
	// LET assigns computed values or subquery results
	query := `
		LET totalEmployees = LENGTH(employees)
		LET totalSalary = SUM(FOR e IN employees RETURN e.salary)

		FOR e IN employees
			FILTER e.active == true
			LET salaryPercent = (e.salary / totalSalary) * 100
			SORT salaryPercent DESC
			LIMIT 5
			RETURN {
				name: e.name,
				salary: e.salary,
				salary_percent: ROUND(salaryPercent * 100) / 100,
				total_employees: totalEmployees
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
		fmt.Printf("  %s - $%.0f (%.1f%% of total payroll)\n",
			result["name"], result["salary"], result["salary_percent"])
	}
}

// ========================================
// Key Concepts Recap
// ========================================
//
// AQL Basics:
//   FOR doc IN collection    — iterate over documents (like SQL FROM)
//   FILTER condition          — filter results (like SQL WHERE)
//   SORT field ASC/DESC      — sort results (like SQL ORDER BY)
//   LIMIT count              — limit results (like SQL LIMIT)
//   RETURN expression        — what to output (like SQL SELECT)
//
// Aggregation:
//   COLLECT key = field      — group by (like SQL GROUP BY)
//   AGGREGATE fn(field)      — aggregate functions (AVG, SUM, MIN, MAX, LENGTH)
//
// Parameterized Queries (ALWAYS use for user input):
//   query := "FOR e IN emps FILTER e.name == @name RETURN e"
//   bindVars := map[string]interface{}{"name": "Alice"}
//   db.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
//
// Modification Queries:
//   UPDATE doc WITH {fields} IN collection
//   REMOVE doc IN collection
//   INSERT {fields} INTO collection
//
// Go Driver Pattern:
//   cursor, err := db.Query(ctx, aql, opts)
//   defer cursor.Close()
//   for cursor.HasMore() {
//       var result Type
//       _, err := cursor.ReadDocument(ctx, &result)
//   }
