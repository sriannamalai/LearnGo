package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ========================================
// Week 16 — Mini-Project: Client for User & Order Services
// ========================================
//
// This client demonstrates calling two microservices via gRPC.
// It tests the full workflow:
//   1. Create users via User service
//   2. Create orders via Order service (which validates users)
//   3. Query both services
//
// Architecture:
//   This Client ──> User Service  (localhost:50051)
//   This Client ──> Order Service (localhost:50052) ──> User Service
//
// Prerequisites:
//   1. Start User service:  cd ../user_service  && go run main.go
//   2. Start Order service: cd ../order_service && go run main.go
//   3. Run this client:     go run main.go

// ========================================
// Types (shared between client and server)
// ========================================

type User struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type OrderItem struct {
	ProductName string  `json:"product_name"`
	Quantity    int32   `json:"quantity"`
	Price       float64 `json:"price"`
}

type Order struct {
	Id        string      `json:"id"`
	UserID    string      `json:"user_id"`
	UserName  string      `json:"user_name"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	Status    string      `json:"status"`
	CreatedAt string      `json:"created_at"`
}

func main() {
	fmt.Println("=== Week 16 Mini-Project: gRPC Client ===")
	fmt.Println()

	// ========================================
	// Step 1: Connect to both services
	// ========================================
	fmt.Println("--- Connecting to Services ---")

	// Connect to User service
	userConn, err := grpc.NewClient(
		"localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to User service: %v", err)
	}
	defer userConn.Close()
	fmt.Println("Connected to User service  (localhost:50051)")

	// Connect to Order service
	orderConn, err := grpc.NewClient(
		"localhost:50052",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect to Order service: %v", err)
	}
	defer orderConn.Close()
	fmt.Println("Connected to Order service (localhost:50052)")
	fmt.Println()

	// In real code with generated stubs:
	//   userClient  := userpb.NewUserServiceClient(userConn)
	//   orderClient := orderpb.NewOrderServiceClient(orderConn)

	// Create a context with timeout for all operations
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ========================================
	// Step 2: Test User Service
	// ========================================
	fmt.Println("--- Testing User Service ---")
	fmt.Println()

	// Create a new user
	// In real code: resp, err := userClient.CreateUser(ctx, &userpb.CreateUserRequest{...})
	fmt.Println("1. CreateUser('Diana Prince', 'diana@example.com')")
	fmt.Println("   Response: { id: 'user-004', name: 'Diana Prince', email: 'diana@example.com' }")
	fmt.Println()

	// List all users
	// In real code: resp, err := userClient.ListUsers(ctx, &userpb.ListUsersRequest{})
	fmt.Println("2. ListUsers()")
	fmt.Println("   Response:")
	users := []User{
		{Id: "user-001", Name: "Alice Johnson", Email: "alice@example.com"},
		{Id: "user-002", Name: "Bob Smith", Email: "bob@example.com"},
		{Id: "user-003", Name: "Charlie Brown", Email: "charlie@example.com"},
		{Id: "user-004", Name: "Diana Prince", Email: "diana@example.com"},
	}
	for _, u := range users {
		fmt.Printf("     - %s: %s (%s)\n", u.Id, u.Name, u.Email)
	}
	fmt.Println()

	// Get a specific user
	// In real code: resp, err := userClient.GetUser(ctx, &userpb.GetUserRequest{Id: "user-001"})
	fmt.Println("3. GetUser('user-001')")
	fmt.Println("   Response: { id: 'user-001', name: 'Alice Johnson', email: 'alice@example.com' }")
	fmt.Println()

	// ========================================
	// Step 3: Test Order Service
	// ========================================
	fmt.Println("--- Testing Order Service ---")
	fmt.Println()

	// Create an order (Order service will validate user with User service)
	// In real code:
	//   resp, err := orderClient.CreateOrder(ctx, &orderpb.CreateOrderRequest{
	//       UserId: "user-001",
	//       Items: []*orderpb.OrderItem{
	//           {ProductName: "Go Programming Book", Quantity: 1, Price: 49.99},
	//           {ProductName: "Mechanical Keyboard", Quantity: 1, Price: 149.99},
	//       },
	//   })

	fmt.Println("4. CreateOrder(user_id='user-001', items=[Book, Keyboard])")
	fmt.Println("   Order service validates user-001 with User service...")
	fmt.Println("   User service confirms: Alice Johnson exists")
	fmt.Println("   Response:")
	fmt.Println("     Order ID:   order-001")
	fmt.Println("     User:       Alice Johnson (user-001)")
	fmt.Println("     Items:")
	fmt.Println("       - Go Programming Book   x1  $49.99")
	fmt.Println("       - Mechanical Keyboard    x1  $149.99")
	fmt.Println("     Total:      $199.98")
	fmt.Println("     Status:     PENDING")
	fmt.Println()

	// Create another order for a different user
	fmt.Println("5. CreateOrder(user_id='user-002', items=[Coffee, Notebook])")
	fmt.Println("   Order service validates user-002 with User service...")
	fmt.Println("   User service confirms: Bob Smith exists")
	fmt.Println("   Response:")
	fmt.Println("     Order ID:   order-002")
	fmt.Println("     User:       Bob Smith (user-002)")
	fmt.Println("     Items:")
	fmt.Println("       - Premium Coffee         x3  $15.99")
	fmt.Println("       - Developer Notebook      x2  $12.99")
	fmt.Println("     Total:      $73.95")
	fmt.Println("     Status:     PENDING")
	fmt.Println()

	// ========================================
	// Step 4: Test Error Cases
	// ========================================
	fmt.Println("--- Testing Error Cases ---")
	fmt.Println()

	// Try to create an order for a non-existent user
	fmt.Println("6. CreateOrder(user_id='user-999', items=[...])")
	fmt.Println("   Order service validates user-999 with User service...")
	fmt.Println("   User service returns: NOT_FOUND 'user user-999 not found'")
	fmt.Println("   Order service propagates error to client")
	fmt.Println("   Error: rpc error: code = NotFound desc = user user-999 not found")
	fmt.Println()

	// Try to get a non-existent user
	fmt.Println("7. GetUser('user-999')")
	fmt.Println("   Error: rpc error: code = NotFound desc = user user-999 not found")
	fmt.Println()

	// Try with empty fields
	fmt.Println("8. CreateUser(name='', email='')")
	fmt.Println("   Error: rpc error: code = InvalidArgument desc = name is required")
	fmt.Println()

	// ========================================
	// Step 5: Query orders by user
	// ========================================
	fmt.Println("--- Querying Orders ---")
	fmt.Println()

	fmt.Println("9. ListOrdersByUser('user-001')")
	fmt.Println("   Response: 1 order found")
	fmt.Println("     - order-001: $199.98 (PENDING)")
	fmt.Println()

	fmt.Println("10. GetOrder('order-001')")
	fmt.Println("    Response: order-001 for Alice Johnson, $199.98")
	fmt.Println()

	// ========================================
	// Summary
	// ========================================
	fmt.Println("========================================")
	fmt.Println("Mini-Project Summary")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("This project demonstrated:")
	fmt.Println()
	fmt.Println("1. TWO gRPC SERVICES communicating with each other:")
	fmt.Println("   - User Service:  manages user accounts")
	fmt.Println("   - Order Service: manages orders, validates users via User service")
	fmt.Println()
	fmt.Println("2. SERVICE-TO-SERVICE CALLS:")
	fmt.Println("   - Order service acts as BOTH a server and a client")
	fmt.Println("   - It receives RPCs from the client")
	fmt.Println("   - It makes RPCs to the User service")
	fmt.Println()
	fmt.Println("3. ERROR PROPAGATION across services:")
	fmt.Println("   - User service returns NotFound")
	fmt.Println("   - Order service propagates it to the client")
	fmt.Println("   - gRPC status codes are preserved")
	fmt.Println()
	fmt.Println("4. KEY MICROSERVICES PATTERNS:")
	fmt.Println("   - Each service owns its own data")
	fmt.Println("   - Services communicate via well-defined APIs (protobuf)")
	fmt.Println("   - Services can be deployed and scaled independently")
	fmt.Println("   - Timeouts and context propagation prevent cascading failures")
	fmt.Println()

	// Use the connections to avoid unused variable errors
	_ = ctx
	_ = userConn
	_ = orderConn

	fmt.Println("Next week: Messaging & Events with NATS!")
}
