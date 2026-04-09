package main

// ========================================
// Week 21 — Lesson 2: Advanced Bindings
// ========================================
// This lesson covers:
//   - Exposing Go structs to the frontend
//   - Complex types: structs, slices, maps
//   - Error handling in bound methods
//   - Events: Go → frontend (emit) and frontend → Go (on)
//   - Bidirectional event communication
//   - Wails runtime: dialogs, clipboard, window control
//
// In Wails, "bindings" are the bridge between Go and JavaScript.
// When you bind a struct, all its exported methods become
// callable from the frontend. Return types are automatically
// converted to JavaScript equivalents.
//
// Type mapping (Go → JavaScript):
//   string     → string
//   int/float  → number
//   bool       → boolean
//   struct     → object
//   []T        → Array
//   map[K]V    → object
//   error      → rejected Promise
//
// Run:
//   wails dev

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"
)

// ========================================
// Struct Types Exposed to Frontend
// ========================================
// When a Go struct is returned from a bound method, it
// becomes a plain JavaScript object. Field names are
// converted based on json tags (if present) or kept as-is.

// TodoItem represents a single task. The json tags control
// the field names in JavaScript.
type TodoItem struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
	Priority  string `json:"priority"` // "low", "medium", "high"
	CreatedAt string `json:"createdAt"`
}

// UserProfile demonstrates nested struct binding.
type UserProfile struct {
	Name    string            `json:"name"`
	Email   string            `json:"email"`
	Age     int               `json:"age"`
	Tags    []string          `json:"tags"`
	Settings map[string]string `json:"settings"`
}

// ApiResponse demonstrates generic response wrapping.
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ========================================
// TodoService — Struct with Bound Methods
// ========================================
// This struct manages todo items and demonstrates how
// to expose a service with full CRUD operations to
// the frontend. Each exported method becomes a JS function.

// TodoService manages a collection of todo items.
type TodoService struct {
	ctx   context.Context
	todos []TodoItem
	nextID int
}

// NewTodoService creates a new todo service.
func NewTodoService() *TodoService {
	return &TodoService{
		todos:  []TodoItem{},
		nextID: 1,
	}
}

// Startup is called by Wails when the app starts.
func (ts *TodoService) Startup(ctx context.Context) {
	ts.ctx = ctx
	fmt.Println("[TodoService] Started")

	// Add some sample items
	ts.AddTodo("Learn Wails bindings", "high")
	ts.AddTodo("Build a dashboard", "medium")
	ts.AddTodo("Deploy the app", "low")
}

// ========================================
// CRUD Methods — Available in Frontend
// ========================================

// GetAll returns all todo items as a slice.
// In JS: const todos = await TodoService.GetAll()
// Returns: Array of TodoItem objects
func (ts *TodoService) GetAll() []TodoItem {
	fmt.Printf("[TodoService] GetAll: returning %d items\n", len(ts.todos))
	return ts.todos
}

// GetByID returns a single todo item.
// Demonstrates error return — errors become rejected Promises.
// In JS: try { const todo = await TodoService.GetByID(1) } catch(err) { ... }
func (ts *TodoService) GetByID(id int) (TodoItem, error) {
	for _, todo := range ts.todos {
		if todo.ID == id {
			return todo, nil
		}
	}
	return TodoItem{}, fmt.Errorf("todo with ID %d not found", id)
}

// AddTodo creates a new todo item and returns it.
// In JS: const newTodo = await TodoService.AddTodo("title", "high")
func (ts *TodoService) AddTodo(title, priority string) TodoItem {
	if priority == "" {
		priority = "medium"
	}

	todo := TodoItem{
		ID:        ts.nextID,
		Title:     title,
		Completed: false,
		Priority:  priority,
		CreatedAt: time.Now().Format(time.RFC3339),
	}
	ts.todos = append(ts.todos, todo)
	ts.nextID++

	fmt.Printf("[TodoService] Added: %s (priority: %s)\n", title, priority)

	// ========================================
	// Emitting Events to Frontend
	// ========================================
	// Events allow Go to notify the frontend of changes.
	// The frontend listens with runtime.EventsOn().
	//
	// In a real Wails app, you would use:
	//   runtime.EventsEmit(ts.ctx, "todo:added", todo)
	//
	// The frontend would listen with:
	//   runtime.EventsOn("todo:added", (todo) => { ... })

	return todo
}

// ToggleTodo toggles the completed status of a todo.
// Returns the updated todo or an error.
func (ts *TodoService) ToggleTodo(id int) (TodoItem, error) {
	for i := range ts.todos {
		if ts.todos[i].ID == id {
			ts.todos[i].Completed = !ts.todos[i].Completed
			status := "incomplete"
			if ts.todos[i].Completed {
				status = "complete"
			}
			fmt.Printf("[TodoService] Toggled #%d to %s\n", id, status)

			// Emit event: runtime.EventsEmit(ts.ctx, "todo:updated", ts.todos[i])
			return ts.todos[i], nil
		}
	}
	return TodoItem{}, fmt.Errorf("todo with ID %d not found", id)
}

// DeleteTodo removes a todo item by ID.
// Returns an error if the ID is not found.
func (ts *TodoService) DeleteTodo(id int) error {
	for i, todo := range ts.todos {
		if todo.ID == id {
			ts.todos = append(ts.todos[:i], ts.todos[i+1:]...)
			fmt.Printf("[TodoService] Deleted #%d\n", id)

			// Emit event: runtime.EventsEmit(ts.ctx, "todo:deleted", id)
			return nil
		}
	}
	return fmt.Errorf("todo with ID %d not found", id)
}

// GetStats returns statistics about the todo list.
// Demonstrates returning a map (becomes a JS object).
func (ts *TodoService) GetStats() map[string]int {
	total := len(ts.todos)
	completed := 0
	highPriority := 0

	for _, todo := range ts.todos {
		if todo.Completed {
			completed++
		}
		if todo.Priority == "high" {
			highPriority++
		}
	}

	return map[string]int{
		"total":        total,
		"completed":    completed,
		"pending":      total - completed,
		"highPriority": highPriority,
	}
}

// ========================================
// Error Handling Patterns
// ========================================

// ValidatedAction demonstrates proper error handling.
// When a bound method returns an error, the JS Promise is
// rejected and the error message is available in the catch block.
func (ts *TodoService) ValidatedAction(action string) (ApiResponse, error) {
	if action == "" {
		return ApiResponse{}, errors.New("action cannot be empty")
	}

	validActions := []string{"archive", "export", "reset"}
	isValid := false
	for _, v := range validActions {
		if v == action {
			isValid = true
			break
		}
	}

	if !isValid {
		return ApiResponse{}, fmt.Errorf("invalid action: %q (valid: %s)",
			action, strings.Join(validActions, ", "))
	}

	return ApiResponse{
		Success: true,
		Message: fmt.Sprintf("Action '%s' completed", action),
		Data:    ts.GetStats(),
	}, nil
}

// ========================================
// Events Service — Bidirectional Events
// ========================================
// Events allow communication between Go and JS without
// direct method calls. This is useful for:
//   - Pushing real-time updates from Go to the frontend
//   - Background task notifications
//   - Global app state changes

// EventService demonstrates the Wails event system.
type EventService struct {
	ctx context.Context
}

// NewEventService creates a new event service.
func NewEventService() *EventService {
	return &EventService{}
}

// Startup registers event listeners in Go.
func (es *EventService) Startup(ctx context.Context) {
	es.ctx = ctx
	fmt.Println("[EventService] Started")

	// ========================================
	// Listening for Frontend Events in Go
	// ========================================
	// In a real Wails app:
	//
	//   runtime.EventsOn(es.ctx, "frontend:ping", func(data ...interface{}) {
	//       fmt.Println("[Go] Received ping from frontend:", data)
	//       // Respond with a pong event
	//       runtime.EventsEmit(es.ctx, "backend:pong", map[string]string{
	//           "message": "Pong!",
	//           "time":    time.Now().Format(time.RFC3339),
	//       })
	//   })
	//
	//   // Listen for user action events
	//   runtime.EventsOn(es.ctx, "user:action", func(data ...interface{}) {
	//       fmt.Printf("[Go] User action: %v\n", data)
	//   })
}

// SendNotification emits an event to the frontend.
// Demonstrates Go → Frontend communication.
func (es *EventService) SendNotification(title, message string) {
	fmt.Printf("[EventService] Sending notification: %s - %s\n", title, message)

	// In a real Wails app:
	// runtime.EventsEmit(es.ctx, "notification", map[string]string{
	//     "title":   title,
	//     "message": message,
	//     "time":    time.Now().Format("3:04 PM"),
	// })
}

// StartBackgroundTask simulates a long-running Go task
// that sends progress events to the frontend.
func (es *EventService) StartBackgroundTask(taskName string) {
	fmt.Printf("[EventService] Starting background task: %s\n", taskName)

	go func() {
		for i := 0; i <= 100; i += 10 {
			// In a real Wails app:
			// runtime.EventsEmit(es.ctx, "task:progress", map[string]interface{}{
			//     "task":     taskName,
			//     "progress": i,
			//     "status":   fmt.Sprintf("Processing step %d/10...", i/10),
			// })
			fmt.Printf("[Task] %s: %d%%\n", taskName, i)
			time.Sleep(500 * time.Millisecond)
		}

		// runtime.EventsEmit(es.ctx, "task:complete", map[string]string{
		//     "task":    taskName,
		//     "message": "Task completed successfully!",
		// })
		fmt.Printf("[Task] %s: Complete!\n", taskName)
	}()
}

// GetRandomQuote demonstrates returning varied data.
func (es *EventService) GetRandomQuote() map[string]string {
	quotes := []map[string]string{
		{"text": "Simplicity is the ultimate sophistication.", "author": "Leonardo da Vinci"},
		{"text": "Code is like humor. When you have to explain it, it's bad.", "author": "Cory House"},
		{"text": "First, solve the problem. Then, write the code.", "author": "John Johnson"},
		{"text": "Make it work, make it right, make it fast.", "author": "Kent Beck"},
	}
	return quotes[rand.Intn(len(quotes))]
}

// ========================================
// Window Control Methods
// ========================================
// Wails provides runtime functions to control the window
// from Go code. These are useful for custom title bars
// or programmatic window management.

// WindowOps demonstrates window control methods.
type WindowOps struct {
	ctx context.Context
}

// NewWindowOps creates a new WindowOps instance.
func NewWindowOps() *WindowOps {
	return &WindowOps{}
}

// Startup stores the context.
func (wo *WindowOps) Startup(ctx context.Context) {
	wo.ctx = ctx
}

// SetTitle changes the window title.
func (wo *WindowOps) SetTitle(title string) {
	// In a real Wails app:
	// runtime.WindowSetTitle(wo.ctx, title)
	fmt.Printf("[WindowOps] Title set to: %s\n", title)
}

// Minimize minimizes the window.
func (wo *WindowOps) Minimize() {
	// runtime.WindowMinimise(wo.ctx)
	fmt.Println("[WindowOps] Window minimized")
}

// ToggleFullscreen toggles fullscreen mode.
func (wo *WindowOps) ToggleFullscreen() {
	// runtime.WindowToggleMaximise(wo.ctx)
	fmt.Println("[WindowOps] Fullscreen toggled")
}

// ShowDialog opens a native dialog from Go.
func (wo *WindowOps) ShowDialog(title, message string) {
	// In a real Wails app:
	// runtime.MessageDialog(wo.ctx, runtime.MessageDialogOptions{
	//     Title:   title,
	//     Message: message,
	//     Type:    runtime.InfoDialog,
	// })
	fmt.Printf("[WindowOps] Dialog: %s — %s\n", title, message)
}

// ========================================
// Main
// ========================================

func main() {
	fmt.Println("========================================")
	fmt.Println("  Week 21 - Lesson 2: Advanced Bindings")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("In a real Wails app, these services would be")
	fmt.Println("bound with wails.Run() and called from JS.")
	fmt.Println()
	fmt.Println("Binding multiple structs:")
	fmt.Println("  wails.Run(&options.App{")
	fmt.Println("      Bind: []interface{}{")
	fmt.Println("          todoService,")
	fmt.Println("          eventService,")
	fmt.Println("          windowOps,")
	fmt.Println("      },")
	fmt.Println("  })")
	fmt.Println()

	// Simulate usage
	ctx := context.Background()

	todoService := NewTodoService()
	todoService.Startup(ctx)

	eventService := NewEventService()
	eventService.Startup(ctx)

	fmt.Println("\n--- Todo Service Demo ---")
	todos := todoService.GetAll()
	for _, t := range todos {
		fmt.Printf("  [%d] %s (priority: %s, done: %v)\n",
			t.ID, t.Title, t.Priority, t.Completed)
	}

	fmt.Println("\nToggling first todo...")
	toggled, _ := todoService.ToggleTodo(1)
	fmt.Printf("  Now: %s completed=%v\n", toggled.Title, toggled.Completed)

	fmt.Println("\nStats:", todoService.GetStats())

	fmt.Println("\n--- Error Handling Demo ---")
	result, err := todoService.ValidatedAction("archive")
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("  Success: %s\n", result.Message)
	}

	_, err = todoService.ValidatedAction("invalid")
	if err != nil {
		fmt.Printf("  Expected error: %v\n", err)
	}

	fmt.Println("\n--- Event Service Demo ---")
	quote := eventService.GetRandomQuote()
	fmt.Printf("  Quote: \"%s\" — %s\n", quote["text"], quote["author"])

	// Show system info
	fmt.Printf("\nRunning on: %s/%s with %d CPUs\n",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU())
}
