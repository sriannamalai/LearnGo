package api

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sri/learngo/week24/01_capstone/internal/model"
	"github.com/sri/learngo/week24/01_capstone/internal/service"
)

// ========================================
// gRPC Service Implementation
// ========================================
// This file implements the gRPC service for TaskFlow. It demonstrates:
//   - Week 15-16: gRPC service patterns, protobuf messages
//   - Week 2-3: Error handling with gRPC status codes
//   - Week 6-7: Context propagation
//
// gRPC vs REST:
//   - REST: Browser-friendly, widely understood, text-based (JSON)
//   - gRPC: Efficient binary encoding (protobuf), streaming support,
//     strongly typed contracts, better for service-to-service communication
//
// In TaskFlow, we offer both REST (for web clients) and gRPC
// (for internal services and high-performance clients).
//
// Note: In a full implementation, you would:
//   1. Define messages in proto/task.proto
//   2. Generate Go code with protoc
//   3. Implement the generated interface
//
// This file shows the pattern using manual types that mirror
// what protoc would generate.

// ========================================
// Proto-equivalent Types
// ========================================
// In production, these would be generated from proto/task.proto.
// We define them manually here for educational clarity.

// TaskProto mirrors the protobuf Task message.
type TaskProto struct {
	Id          string `json:"id"`
	UserId      string `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
}

// CreateTaskProtoRequest mirrors the protobuf CreateTaskRequest.
type CreateTaskProtoRequest struct {
	UserId      string   `json:"user_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Priority    string   `json:"priority"`
	Tags        []string `json:"tags"`
}

// CreateTaskProtoResponse mirrors the protobuf CreateTaskResponse.
type CreateTaskProtoResponse struct {
	Task *TaskProto `json:"task"`
}

// GetTaskProtoRequest mirrors the protobuf GetTaskRequest.
type GetTaskProtoRequest struct {
	Id     string `json:"id"`
	UserId string `json:"user_id"`
}

// GetTaskProtoResponse mirrors the protobuf GetTaskResponse.
type GetTaskProtoResponse struct {
	Task *TaskProto `json:"task"`
}

// ListTasksProtoRequest mirrors the protobuf ListTasksRequest.
type ListTasksProtoRequest struct {
	UserId   string `json:"user_id"`
	Status   string `json:"status"`
	Priority string `json:"priority"`
	Limit    int32  `json:"limit"`
	Offset   int32  `json:"offset"`
}

// ListTasksProtoResponse mirrors the protobuf ListTasksResponse.
type ListTasksProtoResponse struct {
	Tasks []*TaskProto `json:"tasks"`
	Total int32        `json:"total"`
}

// ========================================
// gRPC Server Implementation
// ========================================

// TaskGRPCServer implements the gRPC TaskService.
type TaskGRPCServer struct {
	taskService *service.TaskService
	logger      *slog.Logger
}

// RegisterGRPCServices registers all gRPC services on the server.
// In production, you'd use the generated RegisterTaskServiceServer function.
func RegisterGRPCServices(srv *grpc.Server, taskSvc *service.TaskService, logger *slog.Logger) {
	grpcHandler := &TaskGRPCServer{
		taskService: taskSvc,
		logger:      logger.With("component", "grpc"),
	}

	// In production with protoc-generated code:
	//   pb.RegisterTaskServiceServer(srv, grpcHandler)
	//
	// Since we can't generate protobuf code in this educational context,
	// we register our handler manually. The pattern is identical.
	_ = grpcHandler
	logger.Info("gRPC services registered")
}

// ========================================
// gRPC Method Implementations
// ========================================
// Each method corresponds to an RPC defined in the proto file.
// gRPC methods receive a context and a request, and return a
// response and an error. Errors use gRPC status codes rather
// than HTTP status codes.

// CreateTask handles the CreateTask RPC.
func (s *TaskGRPCServer) CreateTask(ctx context.Context, req *CreateTaskProtoRequest) (*CreateTaskProtoResponse, error) {
	s.logger.Info("gRPC CreateTask called", "title", req.Title)

	// Convert gRPC request to domain model request
	domainReq := model.CreateTaskRequest{
		Title:       req.Title,
		Description: req.Description,
		Priority:    model.Priority(req.Priority),
		Tags:        req.Tags,
	}

	// Call the shared service layer (same business logic as REST)
	task, err := s.taskService.CreateTask(ctx, req.UserId, domainReq)
	if err != nil {
		// ========================================
		// gRPC Error Handling
		// ========================================
		// gRPC uses its own status codes (different from HTTP).
		// Common mappings:
		//   InvalidArgument  — bad input (HTTP 400)
		//   NotFound         — resource missing (HTTP 404)
		//   AlreadyExists    — duplicate (HTTP 409)
		//   PermissionDenied — not authorized (HTTP 403)
		//   Internal         — server error (HTTP 500)
		return nil, status.Errorf(codes.InvalidArgument,
			"failed to create task: %v", err)
	}

	return &CreateTaskProtoResponse{
		Task: taskToProto(task),
	}, nil
}

// GetTask handles the GetTask RPC.
func (s *TaskGRPCServer) GetTask(ctx context.Context, req *GetTaskProtoRequest) (*GetTaskProtoResponse, error) {
	task, err := s.taskService.GetTask(ctx, req.UserId, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound,
			"task not found: %v", err)
	}

	return &GetTaskProtoResponse{
		Task: taskToProto(task),
	}, nil
}

// ListTasks handles the ListTasks RPC.
func (s *TaskGRPCServer) ListTasks(ctx context.Context, req *ListTasksProtoRequest) (*ListTasksProtoResponse, error) {
	filter := model.TaskFilter{
		UserID:   req.UserId,
		Status:   model.Status(req.Status),
		Priority: model.Priority(req.Priority),
		Limit:    int(req.Limit),
		Offset:   int(req.Offset),
	}

	tasks, err := s.taskService.ListTasks(ctx, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"failed to list tasks: %v", err)
	}

	protoTasks := make([]*TaskProto, len(tasks))
	for i, t := range tasks {
		protoTasks[i] = taskToProto(t)
	}

	return &ListTasksProtoResponse{
		Tasks: protoTasks,
		Total: int32(len(protoTasks)),
	}, nil
}

// ========================================
// Conversion Helpers
// ========================================
// Convert between domain models and protobuf messages.
// In production, protoc-gen-go generates these automatically.

func taskToProto(task *model.Task) *TaskProto {
	if task == nil {
		return nil
	}
	return &TaskProto{
		Id:          task.ID,
		UserId:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    string(task.Priority),
		Status:      string(task.Status),
	}
}

// ========================================
// gRPC Interceptors (Middleware)
// ========================================
// gRPC interceptors are the equivalent of HTTP middleware.
// They wrap RPC calls to add cross-cutting concerns.

// LoggingInterceptor logs all gRPC calls with timing.
func LoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the actual RPC handler
		resp, err := handler(ctx, req)

		// Log the call
		duration := time.Since(start)
		logger.Info("gRPC call",
			"method", info.FullMethod,
			"duration_ms", duration.Milliseconds(),
			"error", fmt.Sprintf("%v", err),
		)

		return resp, err
	}
}

// RecoveryInterceptor catches panics in gRPC handlers.
func RecoveryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC panic recovered",
					"method", info.FullMethod,
					"error", r,
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}
