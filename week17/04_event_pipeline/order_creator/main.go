package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ========================================
// Week 17 — Mini-Project: Order Creator
// ========================================
//
// This service creates orders and publishes OrderCreated events
// to the NATS JetStream pipeline.
//
// Pipeline architecture:
//   Order Creator ──publish──> ORDERS.created
//                                   |
//                          Payment Processor ──publish──> ORDERS.payment.processed
//                                                              |
//                                                   Notification Service
//                                                              |
//                                Monitor (observes all events)
//
// Prerequisites:
//   docker run -d --name nats -p 4222:4222 nats:latest -js
//
// Run order (in separate terminals):
//   1. cd ../monitor            && go run main.go
//   2. cd ../notification_service && go run main.go
//   3. cd ../payment_processor  && go run main.go
//   4. cd ../order_creator      && go run main.go  (this file — run last)

// OrderCreatedEvent is published when a new order is created.
type OrderCreatedEvent struct {
	OrderID   string      `json:"order_id"`
	Customer  string      `json:"customer"`
	Email     string      `json:"email"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	CreatedAt string      `json:"created_at"`
}

// OrderItem represents a line item in an order.
type OrderItem struct {
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func main() {
	fmt.Println("=== Week 17 Mini-Project: Order Creator ===")
	fmt.Println()

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("order-creator"))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Get JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	ctx := context.Background()

	// ========================================
	// Ensure the ORDERS stream exists
	// ========================================
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "ORDERS",
		Subjects:  []string{"ORDERS.>"},
		Storage:   jetstream.MemoryStorage,
		Retention: jetstream.LimitsPolicy,
		MaxMsgs:   10000,
		MaxAge:    1 * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}
	fmt.Println("ORDERS stream ready")
	fmt.Println()

	// ========================================
	// Create and publish sample orders
	// ========================================
	orders := []OrderCreatedEvent{
		{
			OrderID:  "ORD-001",
			Customer: "Alice Johnson",
			Email:    "alice@example.com",
			Items: []OrderItem{
				{Product: "Go Programming Book", Quantity: 1, Price: 49.99},
				{Product: "Stickers Pack", Quantity: 2, Price: 9.99},
			},
			Total: 69.97,
		},
		{
			OrderID:  "ORD-002",
			Customer: "Bob Smith",
			Email:    "bob@example.com",
			Items: []OrderItem{
				{Product: "Mechanical Keyboard", Quantity: 1, Price: 149.99},
			},
			Total: 149.99,
		},
		{
			OrderID:  "ORD-003",
			Customer: "Charlie Brown",
			Email:    "charlie@example.com",
			Items: []OrderItem{
				{Product: "Ultra-Wide Monitor", Quantity: 1, Price: 599.99},
				{Product: "Monitor Arm", Quantity: 1, Price: 79.99},
				{Product: "USB-C Cable", Quantity: 2, Price: 14.99},
			},
			Total: 709.96,
		},
		{
			OrderID:  "ORD-004",
			Customer: "Diana Prince",
			Email:    "diana@example.com",
			Items: []OrderItem{
				{Product: "Wireless Mouse", Quantity: 1, Price: 79.99},
			},
			Total: 79.99,
		},
		{
			OrderID:  "ORD-005",
			Customer: "Eve Wilson",
			Email:    "eve@example.com",
			Items: []OrderItem{
				{Product: "Standing Desk", Quantity: 1, Price: 449.99},
				{Product: "Anti-Fatigue Mat", Quantity: 1, Price: 49.99},
			},
			Total: 499.98,
		},
	}

	fmt.Println("Publishing OrderCreated events:")
	fmt.Println()

	for _, order := range orders {
		order.CreatedAt = time.Now().Format(time.RFC3339)
		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order %s: %v", order.OrderID, err)
			continue
		}

		// Publish to ORDERS.created subject
		ack, err := js.Publish(ctx, "ORDERS.created", data,
			// Message ID for deduplication — if we accidentally publish
			// the same event twice, JetStream will deduplicate it.
			jetstream.WithMsgID(order.OrderID),
		)
		if err != nil {
			log.Printf("Failed to publish order %s: %v", order.OrderID, err)
			continue
		}

		fmt.Printf("  [PUBLISHED] %s — %s, $%.2f (seq: %d)\n",
			order.OrderID, order.Customer, order.Total, ack.Sequence)

		// Small delay to simulate real-world order creation
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println()
	fmt.Println("All orders published!")
	fmt.Println("Check the other services to see the pipeline in action.")
}
