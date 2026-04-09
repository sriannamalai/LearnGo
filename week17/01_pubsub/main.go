package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// ========================================
// Week 17 — Lesson 1: NATS Pub/Sub Basics
// ========================================
//
// NATS is a lightweight, high-performance messaging system for microservices.
// It supports pub/sub, request/reply, and queue groups.
//
// Key concepts:
//   - Subjects:     Named channels for messages (e.g., "orders.created")
//   - Publishers:   Send messages to a subject
//   - Subscribers:  Receive messages from a subject
//   - Queue Groups: Load-balance messages across subscribers
//
// Prerequisites:
//   # Start NATS server using Docker:
//   docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:latest
//
//   # Or install natively:
//   brew install nats-server
//   nats-server
//
//   # NATS listens on port 4222 by default
//   # Monitor at http://localhost:8222
//
// To run:
//   go run main.go

func main() {
	fmt.Println("=== Week 17, Lesson 1: NATS Pub/Sub Basics ===")
	fmt.Println()

	// ========================================
	// Step 1: Connect to NATS
	// ========================================
	fmt.Println("--- Connecting to NATS ---")

	// nats.Connect establishes a connection to the NATS server.
	// nats.DefaultURL is "nats://127.0.0.1:4222"
	nc, err := nats.Connect(
		nats.DefaultURL,
		// Connection options:
		nats.Name("pubsub-demo"),         // Client name (shows in monitoring)
		nats.ReconnectWait(2*time.Second), // Wait between reconnect attempts
		nats.MaxReconnects(10),            // Max reconnection attempts

		// Event handlers for connection lifecycle:
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Printf("Disconnected from NATS: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Printf("Reconnected to NATS at %s", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("NATS connection closed")
		}),
	)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	fmt.Printf("Connected to NATS at %s\n", nc.ConnectedUrl())
	fmt.Printf("Client ID: %s\n", nc.Opts.Name)
	fmt.Println()

	// ========================================
	// Step 2: Simple Subscribe
	// ========================================
	fmt.Println("--- Simple Subscribe ---")
	fmt.Println("Subscribing to 'greetings' subject...")

	var wg sync.WaitGroup

	// Subscribe to a subject. The callback runs in a separate goroutine
	// managed by NATS whenever a message arrives.
	sub, err := nc.Subscribe("greetings", func(msg *nats.Msg) {
		fmt.Printf("  [Subscriber] Received on '%s': %s\n", msg.Subject, string(msg.Data))
		wg.Done()
	})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	_ = sub // We'll use this later to unsubscribe

	// Publish messages
	fmt.Println("Publishing messages to 'greetings'...")
	messages := []string{"Hello, NATS!", "How are you?", "Goodbye!"}
	wg.Add(len(messages))

	for _, msg := range messages {
		err := nc.Publish("greetings", []byte(msg))
		if err != nil {
			log.Printf("Failed to publish: %v", err)
		}
		fmt.Printf("  [Publisher] Sent: %s\n", msg)
	}

	// Flush ensures all published messages are sent to the server.
	// NATS buffers messages for efficiency.
	nc.Flush()

	// Wait for all messages to be received
	wg.Wait()
	fmt.Println()

	// ========================================
	// Step 3: Wildcard Subscriptions
	// ========================================
	fmt.Println("--- Wildcard Subscriptions ---")
	fmt.Println()
	fmt.Println("NATS supports two wildcard tokens:")
	fmt.Println("  *  matches a single token:  'orders.*'    matches 'orders.created', 'orders.shipped'")
	fmt.Println("  >  matches one or more:     'orders.>'    matches 'orders.us.created', 'orders.eu.shipped'")
	fmt.Println()

	// Single-level wildcard: * matches exactly one token
	wg.Add(3)
	nc.Subscribe("events.*", func(msg *nats.Msg) {
		fmt.Printf("  [events.*] Received on '%s': %s\n", msg.Subject, string(msg.Data))
		wg.Done()
	})

	// Multi-level wildcard: > matches one or more tokens
	wg.Add(3)
	nc.Subscribe("events.>", func(msg *nats.Msg) {
		fmt.Printf("  [events.>] Received on '%s': %s\n", msg.Subject, string(msg.Data))
		wg.Done()
	})

	fmt.Println("Publishing to various event subjects:")

	// This matches BOTH events.* and events.>
	nc.Publish("events.created", []byte("order-001 created"))
	fmt.Println("  Published to 'events.created'")

	// This matches ONLY events.> (events.* needs exactly one token)
	nc.Publish("events.us.created", []byte("US order created"))
	fmt.Println("  Published to 'events.us.created'")

	// This matches BOTH
	nc.Publish("events.shipped", []byte("order-001 shipped"))
	fmt.Println("  Published to 'events.shipped'")

	nc.Flush()
	wg.Wait()
	fmt.Println()

	// ========================================
	// Step 4: Queue Groups (Load Balancing)
	// ========================================
	fmt.Println("--- Queue Groups (Load Balancing) ---")
	fmt.Println()
	fmt.Println("Queue groups distribute messages across subscribers.")
	fmt.Println("Only ONE subscriber in the group receives each message.")
	fmt.Println("This is how you scale consumers!")
	fmt.Println()

	// Create 3 subscribers in the same queue group "workers"
	receivedBy := make(map[string]int)
	var mu sync.Mutex

	for i := 1; i <= 3; i++ {
		workerName := fmt.Sprintf("worker-%d", i)
		// QueueSubscribe adds this subscriber to a load-balanced group.
		// Messages are distributed round-robin among group members.
		nc.QueueSubscribe("tasks", "workers", func(msg *nats.Msg) {
			mu.Lock()
			receivedBy[workerName]++
			mu.Unlock()
			wg.Done()
		})
	}

	// Publish 9 tasks — they'll be distributed among 3 workers
	taskCount := 9
	wg.Add(taskCount)

	fmt.Printf("Publishing %d tasks to queue group 'workers' (3 workers)...\n", taskCount)
	for i := 1; i <= taskCount; i++ {
		nc.Publish("tasks", []byte(fmt.Sprintf("task-%d", i)))
	}
	nc.Flush()
	wg.Wait()

	fmt.Println("Distribution:")
	mu.Lock()
	for worker, count := range receivedBy {
		fmt.Printf("  %s received %d messages\n", worker, count)
	}
	mu.Unlock()
	fmt.Println("Notice: messages are roughly evenly distributed!")
	fmt.Println()

	// ========================================
	// Step 5: Synchronous Subscribe
	// ========================================
	fmt.Println("--- Synchronous Subscribe ---")
	fmt.Println()

	// SubscribeSync returns a subscription you can pull from manually.
	// Useful when you want to control the receive loop yourself.
	syncSub, err := nc.SubscribeSync("sync.demo")
	if err != nil {
		log.Fatalf("Failed to sync subscribe: %v", err)
	}

	// Publish a message
	nc.Publish("sync.demo", []byte("sync message 1"))
	nc.Publish("sync.demo", []byte("sync message 2"))
	nc.Flush()

	// Pull messages with a timeout
	for i := 0; i < 2; i++ {
		msg, err := syncSub.NextMsg(1 * time.Second)
		if err != nil {
			log.Printf("Timeout waiting for message: %v", err)
			break
		}
		fmt.Printf("  Sync received: %s\n", string(msg.Data))
	}
	fmt.Println()

	// ========================================
	// Step 6: Unsubscribe and Auto-Unsubscribe
	// ========================================
	fmt.Println("--- Unsubscribe ---")

	// Unsubscribe from the greetings subject
	sub.Unsubscribe()
	fmt.Println("Unsubscribed from 'greetings'")

	// AutoUnsubscribe: automatically unsubscribe after N messages
	autoSub, _ := nc.Subscribe("auto.demo", func(msg *nats.Msg) {
		fmt.Printf("  Auto-sub received: %s\n", string(msg.Data))
	})
	autoSub.AutoUnsubscribe(2) // Only receive 2 messages, then auto-unsub

	nc.Publish("auto.demo", []byte("message 1"))
	nc.Publish("auto.demo", []byte("message 2"))
	nc.Publish("auto.demo", []byte("message 3 — won't be received"))
	nc.Flush()
	time.Sleep(100 * time.Millisecond) // Brief wait for delivery
	fmt.Println("  (message 3 was not received — auto-unsubscribed after 2)")
	fmt.Println()

	// ========================================
	// Key Concepts Summary
	// ========================================
	fmt.Println("--- Key Concepts ---")
	fmt.Println()
	fmt.Println("1. SUBJECTS are like topics/channels for messages")
	fmt.Println("   Use dots for hierarchy: 'orders.us.created'")
	fmt.Println()
	fmt.Println("2. PUB/SUB is fire-and-forget:")
	fmt.Println("   Publisher doesn't know who's listening")
	fmt.Println("   Messages are lost if no subscribers exist")
	fmt.Println()
	fmt.Println("3. WILDCARDS for flexible subscriptions:")
	fmt.Println("   * = one token, > = one or more tokens")
	fmt.Println()
	fmt.Println("4. QUEUE GROUPS for load balancing:")
	fmt.Println("   Multiple subscribers, each message delivered once")
	fmt.Println()
	fmt.Println("5. FLUSH ensures messages are sent:")
	fmt.Println("   NATS batches messages for performance")
	fmt.Println()
	fmt.Println("Next: Request/Reply pattern for synchronous communication")
}
