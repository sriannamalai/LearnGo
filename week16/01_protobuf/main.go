package main

import (
	"fmt"
)

// ========================================
// Week 16 — Lesson 1: Protocol Buffers (Protobuf)
// ========================================
//
// Protocol Buffers is a language-neutral, platform-neutral mechanism
// for serializing structured data. It's the foundation of gRPC.
//
// Key concepts:
//   1. You define data structures in .proto files
//   2. The protoc compiler generates Go code from .proto files
//   3. Generated code gives you structs + serialization methods
//   4. Messages are encoded in a compact binary format
//
// See proto/user.proto for the protobuf definitions.

// ========================================
// What protoc generates — manual equivalents
// ========================================
//
// When you run protoc on user.proto, it generates Go structs
// that look roughly like this. We define them manually here
// for educational purposes so you can understand what's happening.

// UserRole corresponds to the enum in the .proto file.
// In generated code, this is an int32 type with named constants.
type UserRole int32

const (
	UserRole_UNSPECIFIED UserRole = 0
	UserRole_ADMIN       UserRole = 1
	UserRole_MEMBER      UserRole = 2
	UserRole_GUEST       UserRole = 3
)

// String returns the human-readable name of the role.
// Generated code includes this method automatically.
func (r UserRole) String() string {
	switch r {
	case UserRole_ADMIN:
		return "ADMIN"
	case UserRole_MEMBER:
		return "MEMBER"
	case UserRole_GUEST:
		return "GUEST"
	default:
		return "UNSPECIFIED"
	}
}

// User represents the User message from user.proto.
// In real generated code, this struct would have additional
// unexported fields for protobuf internal state.
type User struct {
	Id    string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name  string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Email string   `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	Age   int32    `protobuf:"varint,4,opt,name=age,proto3" json:"age,omitempty"`
	Role  UserRole `protobuf:"varint,5,opt,name=role,proto3,enum=user.UserRole" json:"role,omitempty"`
	Tags  []string `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
}

// CreateUserRequest is what the client sends to create a user.
type CreateUserRequest struct {
	Name  string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Email string   `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"`
	Age   int32    `protobuf:"varint,3,opt,name=age,proto3" json:"age,omitempty"`
	Role  UserRole `protobuf:"varint,4,opt,name=role,proto3,enum=user.UserRole" json:"role,omitempty"`
}

// CreateUserResponse is what the server sends back.
type CreateUserResponse struct {
	User *User `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
}

// GetUserRequest asks for a user by ID.
type GetUserRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

// GetUserResponse contains the requested user.
type GetUserResponse struct {
	User *User `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
}

func main() {
	fmt.Println("=== Week 16, Lesson 1: Protocol Buffers (Protobuf) ===")
	fmt.Println()

	// ========================================
	// What are Protocol Buffers?
	// ========================================
	fmt.Println("--- What are Protocol Buffers? ---")
	fmt.Println("Protocol Buffers (protobuf) is Google's data serialization format.")
	fmt.Println("It's like JSON but:")
	fmt.Println("  - Smaller: binary encoding is much more compact")
	fmt.Println("  - Faster:  binary parsing is much faster than text parsing")
	fmt.Println("  - Typed:   schema is defined in .proto files")
	fmt.Println("  - Versioned: field numbers enable backward compatibility")
	fmt.Println()

	// ========================================
	// The Protobuf Workflow
	// ========================================
	fmt.Println("--- The Protobuf Workflow ---")
	fmt.Println("1. Define your data in a .proto file (see proto/user.proto)")
	fmt.Println("2. Run the protoc compiler to generate Go code:")
	fmt.Println("   protoc --go_out=. --go-grpc_out=. proto/user.proto")
	fmt.Println("3. Use the generated structs in your Go code")
	fmt.Println("4. protobuf handles serialization/deserialization")
	fmt.Println()

	// ========================================
	// Demonstrating the struct equivalents
	// ========================================
	fmt.Println("--- Using Protobuf-style Structs ---")

	// Create a user (like what protoc-generated code would look like)
	user := &User{
		Id:    "user-001",
		Name:  "Alice Johnson",
		Email: "alice@example.com",
		Age:   30,
		Role:  UserRole_ADMIN,
		Tags:  []string{"engineer", "team-lead"},
	}

	fmt.Printf("User: %+v\n", user)
	fmt.Printf("  ID:    %s\n", user.Id)
	fmt.Printf("  Name:  %s\n", user.Name)
	fmt.Printf("  Email: %s\n", user.Email)
	fmt.Printf("  Age:   %d\n", user.Age)
	fmt.Printf("  Role:  %s (%d)\n", user.Role, user.Role)
	fmt.Printf("  Tags:  %v\n", user.Tags)
	fmt.Println()

	// Create a request message
	req := &CreateUserRequest{
		Name:  "Bob Smith",
		Email: "bob@example.com",
		Age:   25,
		Role:  UserRole_MEMBER,
	}
	fmt.Printf("CreateUserRequest: %+v\n", req)
	fmt.Println()

	// Simulating what the server would return
	resp := &CreateUserResponse{
		User: &User{
			Id:    "user-002",
			Name:  req.Name,
			Email: req.Email,
			Age:   req.Age,
			Role:  req.Role,
		},
	}
	fmt.Printf("CreateUserResponse: %+v\n", resp)
	fmt.Printf("  Created user ID: %s\n", resp.User.Id)
	fmt.Println()

	// ========================================
	// Key Protobuf Concepts
	// ========================================
	fmt.Println("--- Key Protobuf Concepts ---")
	fmt.Println()

	fmt.Println("1. Field Numbers (the = 1, = 2, etc.):")
	fmt.Println("   - Used in binary encoding, NOT the field name")
	fmt.Println("   - NEVER change or reuse field numbers!")
	fmt.Println("   - Numbers 1-15 use 1 byte (use for frequent fields)")
	fmt.Println("   - Numbers 16-2047 use 2 bytes")
	fmt.Println()

	fmt.Println("2. Default Values (proto3):")
	fmt.Println("   - string: empty string \"\"")
	fmt.Println("   - int32/int64: 0")
	fmt.Println("   - bool: false")
	fmt.Println("   - enum: first value (must be 0)")
	fmt.Println("   - message: nil")
	fmt.Println("   - repeated: empty slice")
	fmt.Println()

	// Demonstrate default values
	emptyUser := &User{}
	fmt.Printf("Empty user defaults: Id=%q, Name=%q, Age=%d, Role=%s, Tags=%v\n",
		emptyUser.Id, emptyUser.Name, emptyUser.Age, emptyUser.Role, emptyUser.Tags)
	fmt.Println()

	fmt.Println("3. Wire Types (how data is encoded):")
	fmt.Println("   - Varint: int32, int64, bool, enum")
	fmt.Println("   - 64-bit: fixed64, double")
	fmt.Println("   - Length-delimited: string, bytes, messages, repeated")
	fmt.Println("   - 32-bit: fixed32, float")
	fmt.Println()

	fmt.Println("4. Backward Compatibility Rules:")
	fmt.Println("   - You CAN add new fields (old code ignores them)")
	fmt.Println("   - You CAN remove fields (mark as reserved)")
	fmt.Println("   - You CANNOT change field numbers")
	fmt.Println("   - You CANNOT change field types (mostly)")
	fmt.Println()

	// ========================================
	// Proto vs JSON comparison
	// ========================================
	fmt.Println("--- Protobuf vs JSON Size Comparison ---")
	fmt.Println()

	// JSON representation of our user
	jsonStr := `{"id":"user-001","name":"Alice Johnson","email":"alice@example.com","age":30,"role":1,"tags":["engineer","team-lead"]}`
	fmt.Printf("JSON:     %d bytes\n", len(jsonStr))
	fmt.Println("  ", jsonStr)
	fmt.Println()

	// Protobuf would encode this much more compactly
	// (field numbers instead of names, varint encoding for numbers)
	fmt.Println("Protobuf: ~60-70 bytes (estimated)")
	fmt.Println("  Binary format — not human-readable but much smaller")
	fmt.Println("  Field names are replaced by 1-2 byte field numbers")
	fmt.Println("  Numbers use variable-length encoding (varint)")
	fmt.Println()

	// ========================================
	// Service Definition Concepts
	// ========================================
	fmt.Println("--- gRPC Service Patterns ---")
	fmt.Println()
	fmt.Println("Protobuf services define RPC methods. Four patterns:")
	fmt.Println()
	fmt.Println("1. Unary RPC (most common):")
	fmt.Println("   rpc GetUser(GetUserRequest) returns (GetUserResponse)")
	fmt.Println("   Client sends one request, server sends one response.")
	fmt.Println()
	fmt.Println("2. Server Streaming RPC:")
	fmt.Println("   rpc ListUsers(ListUsersRequest) returns (stream GetUserResponse)")
	fmt.Println("   Client sends one request, server sends a stream of responses.")
	fmt.Println()
	fmt.Println("3. Client Streaming RPC:")
	fmt.Println("   rpc UploadUsers(stream CreateUserRequest) returns (UploadResponse)")
	fmt.Println("   Client sends a stream of requests, server sends one response.")
	fmt.Println()
	fmt.Println("4. Bidirectional Streaming RPC:")
	fmt.Println("   rpc Chat(stream ChatMessage) returns (stream ChatMessage)")
	fmt.Println("   Both client and server send streams of messages.")
	fmt.Println()

	fmt.Println("--- Next Steps ---")
	fmt.Println("In the next lesson, we'll build a gRPC server that")
	fmt.Println("implements the UserService defined in our .proto file.")
}
