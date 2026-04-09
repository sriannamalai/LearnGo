package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// ========================================
// Week 16 — Lesson 3: gRPC Client
// ========================================
//
// This lesson demonstrates how to build a gRPC client that connects
// to the server from Lesson 2 and calls its RPC methods.
//
// Prerequisites:
//   - The gRPC server from Lesson 2 must be running on localhost:50051
//
// To run:
//   1. Start the server: cd ../02_grpc_server && go run main.go
//   2. Run this client: go run main.go

// ========================================
// Message types (same as server, from proto generation)
// ========================================

type UserRole int32

const (
	UserRole_UNSPECIFIED UserRole = 0
	UserRole_ADMIN       UserRole = 1
	UserRole_MEMBER      UserRole = 2
	UserRole_GUEST       UserRole = 3
)

type User struct {
	Id    string   `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Age   int32    `json:"age"`
	Role  UserRole `json:"role"`
	Tags  []string `json:"tags"`
}

type CreateUserRequest struct {
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Age   int32    `json:"age"`
	Role  UserRole `json:"role"`
}

type CreateUserResponse struct {
	User *User `json:"user"`
}

type GetUserRequest struct {
	Id string `json:"id"`
}

type GetUserResponse struct {
	User *User `json:"user"`
}

type ListUsersRequest struct {
	PageSize int32 `json:"page_size"`
}

// ========================================
// Client stub (what protoc-gen-go-grpc generates)
// ========================================
//
// In real generated code, you get a client interface and constructor:
//
//   type UserServiceClient interface {
//       CreateUser(ctx, *CreateUserRequest, ...grpc.CallOption) (*CreateUserResponse, error)
//       GetUser(ctx, *GetUserRequest, ...grpc.CallOption) (*GetUserResponse, error)
//       ListUsers(ctx, *ListUsersRequest, ...grpc.CallOption) (UserService_ListUsersClient, error)
//   }
//
//   func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient
//
// The client stub handles:
//   - Serializing the request to protobuf
//   - Sending it over HTTP/2
//   - Deserializing the response
//   - Propagating errors as gRPC status codes

func main() {
	fmt.Println("=== Week 16, Lesson 3: gRPC Client ===")
	fmt.Println()

	// ========================================
	// Step 1: Establish a connection
	// ========================================
	fmt.Println("--- Connecting to gRPC Server ---")

	// grpc.Dial (now grpc.NewClient) creates a connection to the server.
	// The connection is lazy by default — it connects when you first call an RPC.
	//
	// Options:
	//   - insecure.NewCredentials() — no TLS (development only!)
	//   - grpc.WithBlock()          — wait for connection (not recommended)
	//   - grpc.WithTimeout()        — connection timeout

	serverAddr := "localhost:50051"
	fmt.Printf("Dialing %s...\n", serverAddr)

	conn, err := grpc.NewClient(
		serverAddr,
		// WARNING: insecure.NewCredentials() disables TLS.
		// In production, ALWAYS use TLS:
		//   grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	// Always close the connection when done
	defer conn.Close()

	fmt.Println("Connection established (lazy — actual connect on first RPC)")
	fmt.Println()

	// ========================================
	// Step 2: Create a client stub
	// ========================================
	// In real code with generated files:
	//   client := pb.NewUserServiceClient(conn)
	//
	// The client stub is a wrapper around the connection that knows
	// how to serialize/deserialize messages and call the right methods.

	fmt.Println("--- Client Stub ---")
	fmt.Println("In real code: client := pb.NewUserServiceClient(conn)")
	fmt.Println("The stub handles serialization and network calls.")
	fmt.Println()

	// ========================================
	// Step 3: Call Unary RPCs
	// ========================================
	fmt.Println("--- Calling Unary RPCs ---")
	fmt.Println()

	// Every RPC call takes a context. Use it for:
	//   - Timeouts: context.WithTimeout()
	//   - Cancellation: context.WithCancel()
	//   - Metadata: metadata.NewOutgoingContext()

	// CreateUser with a 5-second timeout
	fmt.Println("1. CreateUser RPC:")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	createReq := &CreateUserRequest{
		Name:  "Eve Wilson",
		Email: "eve@example.com",
		Age:   27,
		Role:  UserRole_MEMBER,
	}
	fmt.Printf("   Request: %+v\n", createReq)

	// In real code:
	//   resp, err := client.CreateUser(ctx, createReq)
	//
	// The stub serializes createReq to protobuf, sends it over HTTP/2,
	// waits for the response, deserializes it, and returns it.
	_ = ctx // used in real RPC call

	fmt.Println("   Response: (would contain the created user with generated ID)")
	fmt.Println()

	// GetUser
	fmt.Println("2. GetUser RPC:")
	getReq := &GetUserRequest{Id: "user-001"}
	fmt.Printf("   Request: %+v\n", getReq)
	fmt.Println("   Response: (would contain user-001 details)")
	fmt.Println()

	// ========================================
	// Step 4: Handle gRPC Errors
	// ========================================
	fmt.Println("--- Handling gRPC Errors ---")
	fmt.Println()

	// gRPC errors carry a status code and a message.
	// Use the status package to extract them.

	// Simulating a "not found" error from the server
	simulatedErr := status.Errorf(codes.NotFound, "user user-999 not found")

	// Check the error
	st, ok := status.FromError(simulatedErr)
	if ok {
		fmt.Printf("Error Code:    %s (%d)\n", st.Code(), st.Code())
		fmt.Printf("Error Message: %s\n", st.Message())
		fmt.Println()

		// You can switch on the code to handle different errors
		switch st.Code() {
		case codes.NotFound:
			fmt.Println("Action: Resource not found — could show 404 to user")
		case codes.InvalidArgument:
			fmt.Println("Action: Bad request — fix the input and retry")
		case codes.Unavailable:
			fmt.Println("Action: Server down — retry with backoff")
		case codes.PermissionDenied:
			fmt.Println("Action: Not authorized — check credentials")
		default:
			fmt.Println("Action: Unknown error — log and investigate")
		}
	}
	fmt.Println()

	// ========================================
	// Step 5: Server Streaming RPC
	// ========================================
	fmt.Println("--- Server Streaming RPC ---")
	fmt.Println()
	fmt.Println("ListUsers is a server-streaming RPC. The pattern is:")
	fmt.Println()
	fmt.Println("  // Create the stream")
	fmt.Println("  stream, err := client.ListUsers(ctx, &ListUsersRequest{PageSize: 10})")
	fmt.Println("  if err != nil { log.Fatal(err) }")
	fmt.Println()
	fmt.Println("  // Read responses from the stream in a loop")
	fmt.Println("  for {")
	fmt.Println("      resp, err := stream.Recv()")
	fmt.Println("      if err == io.EOF {")
	fmt.Println("          break  // Stream finished — server closed it")
	fmt.Println("      }")
	fmt.Println("      if err != nil {")
	fmt.Println("          log.Fatal(err)  // Real error")
	fmt.Println("      }")
	fmt.Println("      fmt.Printf(\"Received user: %s\\n\", resp.User.Name)")
	fmt.Println("  }")
	fmt.Println()
	fmt.Println("The stream is like a channel — you keep receiving until EOF.")
	fmt.Println()

	// ========================================
	// Connection State
	// ========================================
	fmt.Println("--- Connection State ---")
	fmt.Println()
	fmt.Printf("Current connection state: %s\n", conn.GetState().String())
	fmt.Println()
	fmt.Println("Connection states:")
	fmt.Println("  IDLE:              Not yet connected")
	fmt.Println("  CONNECTING:        Attempting to connect")
	fmt.Println("  READY:             Connected and ready for RPCs")
	fmt.Println("  TRANSIENT_FAILURE: Connection lost, retrying")
	fmt.Println("  SHUTDOWN:          Connection closed")
	fmt.Println()

	// ========================================
	// Best Practices
	// ========================================
	fmt.Println("--- gRPC Client Best Practices ---")
	fmt.Println()
	fmt.Println("1. Always use context with timeouts:")
	fmt.Println("   ctx, cancel := context.WithTimeout(ctx, 5*time.Second)")
	fmt.Println("   defer cancel()")
	fmt.Println()
	fmt.Println("2. Reuse connections — don't dial for every RPC:")
	fmt.Println("   conn is thread-safe, create once, share across goroutines")
	fmt.Println()
	fmt.Println("3. Handle errors with status codes:")
	fmt.Println("   st, ok := status.FromError(err)")
	fmt.Println()
	fmt.Println("4. Use TLS in production:")
	fmt.Println("   Never use insecure.NewCredentials() in production!")
	fmt.Println()
	fmt.Println("5. Implement retry logic for transient failures:")
	fmt.Println("   Codes Unavailable and ResourceExhausted are retryable")
	fmt.Println()
	fmt.Println("6. Close connections when shutting down:")
	fmt.Println("   defer conn.Close()")
}
