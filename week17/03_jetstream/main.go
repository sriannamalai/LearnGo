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
// Week 17 — Lesson 3: NATS JetStream
// ========================================
//
// JetStream is NATS's built-in persistence layer. While core NATS
// is fire-and-forget (messages are lost if no subscriber is listening),
// JetStream adds:
//
//   - Message persistence:   Messages are stored on disk/memory
//   - Delivery guarantees:   At-least-once delivery
//   - Consumer groups:       Durable subscriptions that survive restarts
//   - Replay:                Replay messages from any point in time
//   - Acknowledgments:       Consumers must ACK messages
//
// Think of it as NATS's answer to Kafka, RabbitMQ, or Redis Streams.
//
// Prerequisites:
//   # Start NATS with JetStream enabled:
//   docker run -d --name nats -p 4222:4222 nats:latest -js
//
//   # The -js flag enables JetStream
//   # You can also use: nats-server -js
//
// To run:
//   go run main.go

// Order represents an order event stored in JetStream.
type Order struct {
	ID        string  `json:"id"`
	Customer  string  `json:"customer"`
	Product   string  `json:"product"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

func main() {
	fmt.Println("=== Week 17, Lesson 3: NATS JetStream ===")
	fmt.Println()

	// ========================================
	// Step 1: Connect and get JetStream context
	// ========================================
	fmt.Println("--- Connecting to NATS JetStream ---")

	nc, err := nats.Connect(nats.DefaultURL, nats.Name("jetstream-demo"))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer nc.Close()

	// Create a JetStream context from the NATS connection.
	// This is the entry point for all JetStream operations.
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	fmt.Println("Connected to NATS with JetStream enabled")
	fmt.Println()

	ctx := context.Background()

	// ========================================
	// Step 2: Create a Stream
	// ========================================
	fmt.Println("--- Creating a Stream ---")
	fmt.Println()
	fmt.Println("A STREAM is like a log — it captures messages matching")
	fmt.Println("certain subjects and stores them for later consumption.")
	fmt.Println()

	// Create or update the ORDERS stream.
	// The stream captures all messages published to subjects matching "ORDERS.>".
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: "ORDERS",
		// Subjects this stream captures. Supports wildcards.
		Subjects: []string{"ORDERS.>"},

		// Storage backend: FileStorage (disk) or MemoryStorage (RAM)
		Storage: jetstream.MemoryStorage,

		// Retention policy:
		//   LimitsPolicy:   Keep messages until limits are hit (default)
		//   InterestPolicy: Keep only while consumers exist
		//   WorkQueuePolicy: Each message consumed exactly once
		Retention: jetstream.LimitsPolicy,

		// Limits — prevent unbounded growth
		MaxMsgs:    1000,                // Max messages in the stream
		MaxBytes:   1024 * 1024,         // Max total size (1 MB)
		MaxAge:     24 * time.Hour,      // Max message age
		MaxMsgSize: 1024 * 64,           // Max single message size (64 KB)
		Duplicates: 5 * time.Minute,     // Dedup window

		// Discard policy: what happens when limits are hit?
		//   DiscardOld: Remove oldest messages (default)
		//   DiscardNew: Reject new messages
		Discard: jetstream.DiscardOld,
	})
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	fmt.Printf("Stream '%s' created/updated\n", stream.CachedInfo().Config.Name)
	fmt.Printf("  Subjects: %v\n", stream.CachedInfo().Config.Subjects)
	fmt.Printf("  Storage:  Memory\n")
	fmt.Printf("  MaxMsgs:  %d\n", stream.CachedInfo().Config.MaxMsgs)
	fmt.Println()

	// ========================================
	// Step 3: Publish Messages to the Stream
	// ========================================
	fmt.Println("--- Publishing Messages ---")
	fmt.Println()

	orders := []Order{
		{ID: "ORD-001", Customer: "Alice", Product: "Laptop", Amount: 999.99, Status: "created"},
		{ID: "ORD-002", Customer: "Bob", Product: "Keyboard", Amount: 149.99, Status: "created"},
		{ID: "ORD-003", Customer: "Charlie", Product: "Monitor", Amount: 499.99, Status: "created"},
		{ID: "ORD-004", Customer: "Diana", Product: "Mouse", Amount: 79.99, Status: "created"},
		{ID: "ORD-005", Customer: "Eve", Product: "Headphones", Amount: 199.99, Status: "created"},
	}

	for _, order := range orders {
		order.CreatedAt = time.Now().Format(time.RFC3339)
		data, _ := json.Marshal(order)

		// Publish to a subject captured by the ORDERS stream.
		// js.Publish returns an *PubAck with the stream sequence number.
		ack, err := js.Publish(ctx, "ORDERS.created", data)
		if err != nil {
			log.Printf("Failed to publish order %s: %v", order.ID, err)
			continue
		}

		fmt.Printf("  Published %s: seq=%d, stream=%s\n",
			order.ID, ack.Sequence, ack.Stream)
	}
	fmt.Println()

	// ========================================
	// Step 4: Create a Consumer
	// ========================================
	fmt.Println("--- Creating Consumers ---")
	fmt.Println()
	fmt.Println("A CONSUMER is a view into a stream. It tracks which")
	fmt.Println("messages have been delivered and acknowledged.")
	fmt.Println()
	fmt.Println("Consumer types:")
	fmt.Println("  - Durable:   Named, persists across disconnects")
	fmt.Println("  - Ephemeral: Unnamed, deleted when idle")
	fmt.Println()

	// Create a durable consumer
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		// Durable name — the consumer persists across restarts.
		// Without this, the consumer is ephemeral.
		Durable: "order-processor",

		// Deliver policy: where to start consuming
		//   DeliverAllPolicy:       From the beginning
		//   DeliverLastPolicy:      Only the latest message
		//   DeliverNewPolicy:       Only new messages (after subscribe)
		//   DeliverByStartSequence: From a specific sequence number
		//   DeliverByStartTime:     From a specific timestamp
		DeliverPolicy: jetstream.DeliverAllPolicy,

		// Ack policy: how the server tracks consumption
		//   AckExplicitPolicy: Consumer must explicitly ACK (recommended)
		//   AckNonePolicy:     No ACKs needed (fire-and-forget)
		//   AckAllPolicy:      ACKing one ACKs all prior messages
		AckPolicy: jetstream.AckExplicitPolicy,

		// Ack wait: how long server waits for ACK before redelivering
		AckWait: 30 * time.Second,

		// Max deliver: max redelivery attempts before giving up
		MaxDeliver: 3,

		// Filter subject: only receive messages matching this pattern
		FilterSubject: "ORDERS.created",
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	fmt.Printf("Consumer '%s' created\n", consumer.CachedInfo().Name)
	fmt.Printf("  Deliver Policy: All\n")
	fmt.Printf("  Ack Policy:     Explicit\n")
	fmt.Printf("  Filter:         ORDERS.created\n")
	fmt.Println()

	// ========================================
	// Step 5: Consume Messages (Pull-based)
	// ========================================
	fmt.Println("--- Consuming Messages (Pull-based) ---")
	fmt.Println()
	fmt.Println("Pull consumers request messages in batches.")
	fmt.Println("This gives the consumer control over the rate.")
	fmt.Println()

	// Fetch a batch of messages
	batch, err := consumer.Fetch(5, jetstream.FetchMaxWait(5*time.Second))
	if err != nil {
		log.Fatalf("Failed to fetch: %v", err)
	}

	fmt.Println("Processing batch:")
	for msg := range batch.Messages() {
		var order Order
		json.Unmarshal(msg.Data(), &order)

		fmt.Printf("  Processing %s: %s bought %s ($%.2f)\n",
			order.ID, order.Customer, order.Product, order.Amount)

		// ========================================
		// Acknowledgment — CRITICAL for reliability
		// ========================================
		// You MUST acknowledge messages to prevent redelivery.
		//
		// Options:
		//   msg.Ack()         — processed successfully
		//   msg.Nak()         — failed, redeliver immediately
		//   msg.NakWithDelay() — failed, redeliver after delay
		//   msg.Term()        — failed permanently, don't redeliver
		//   msg.InProgress()  — still working, extend ack deadline

		if order.Amount > 500 {
			fmt.Printf("    High-value order! Extra processing...\n")
			// msg.InProgress() // Extend the ack deadline while processing
		}

		msg.Ack() // Mark as processed
		fmt.Printf("    Acknowledged (seq: %d)\n", msg.Headers().Get("Nats-Sequence"))
	}

	if batch.Error() != nil {
		fmt.Printf("Batch error: %v\n", batch.Error())
	}
	fmt.Println()

	// ========================================
	// Step 6: Publish more with different subjects
	// ========================================
	fmt.Println("--- Subject Hierarchy in Streams ---")
	fmt.Println()
	fmt.Println("The ORDERS stream captures 'ORDERS.>' which includes:")
	fmt.Println("  ORDERS.created, ORDERS.paid, ORDERS.shipped, etc.")
	fmt.Println()

	// Publish events with different subjects (all captured by ORDERS stream)
	js.Publish(ctx, "ORDERS.paid", []byte(`{"id":"ORD-001","status":"paid"}`))
	js.Publish(ctx, "ORDERS.shipped", []byte(`{"id":"ORD-001","status":"shipped"}`))
	js.Publish(ctx, "ORDERS.delivered", []byte(`{"id":"ORD-001","status":"delivered"}`))

	fmt.Println("Published order lifecycle events:")
	fmt.Println("  ORDERS.paid      — ORD-001")
	fmt.Println("  ORDERS.shipped   — ORD-001")
	fmt.Println("  ORDERS.delivered — ORD-001")
	fmt.Println()

	// Get stream info to see message counts
	info, _ := stream.Info(ctx)
	fmt.Printf("Stream stats: %d messages, %d bytes\n",
		info.State.Msgs, info.State.Bytes)
	fmt.Println()

	// ========================================
	// Step 7: Replay — Read from the beginning
	// ========================================
	fmt.Println("--- Replay: Reading Historical Messages ---")
	fmt.Println()
	fmt.Println("One of JetStream's superpowers: you can replay messages!")
	fmt.Println("New consumers can start from the beginning, a sequence, or a time.")
	fmt.Println()

	// Create a new consumer that reads ALL messages from the start
	allEventsConsumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "all-events-reader",
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		// No filter — get ALL subjects in the stream
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// Fetch all messages
	allBatch, err := allEventsConsumer.Fetch(20, jetstream.FetchMaxWait(2*time.Second))
	if err != nil {
		log.Fatalf("Failed to fetch: %v", err)
	}

	fmt.Println("All messages in ORDERS stream (replayed from beginning):")
	msgCount := 0
	for msg := range allBatch.Messages() {
		msgCount++
		fmt.Printf("  [seq %s] Subject: %-20s Data: %s\n",
			msg.Headers().Get("Nats-Sequence"),
			msg.Subject(),
			truncate(string(msg.Data()), 60))
		msg.Ack()
	}
	fmt.Printf("Total messages replayed: %d\n", msgCount)
	fmt.Println()

	// ========================================
	// Key Concepts Summary
	// ========================================
	fmt.Println("--- JetStream Key Concepts ---")
	fmt.Println()
	fmt.Println("1. STREAMS capture and persist messages:")
	fmt.Println("   - Define subjects to capture (supports wildcards)")
	fmt.Println("   - Configure retention, limits, storage type")
	fmt.Println()
	fmt.Println("2. CONSUMERS read from streams:")
	fmt.Println("   - Durable: persist across disconnects")
	fmt.Println("   - Track delivery state per consumer")
	fmt.Println("   - Multiple consumers can read the same stream independently")
	fmt.Println()
	fmt.Println("3. ACKNOWLEDGMENTS ensure reliability:")
	fmt.Println("   - Ack:  message processed successfully")
	fmt.Println("   - Nak:  failed, redeliver")
	fmt.Println("   - Term: failed permanently, skip")
	fmt.Println()
	fmt.Println("4. REPLAY enables powerful patterns:")
	fmt.Println("   - New services can catch up on history")
	fmt.Println("   - Event sourcing: rebuild state from events")
	fmt.Println("   - Debugging: replay production events locally")
	fmt.Println()
	fmt.Println("5. JetStream vs Core NATS:")
	fmt.Println("   - Core NATS: fast, simple, fire-and-forget")
	fmt.Println("   - JetStream: persistent, reliable, at-least-once delivery")
	fmt.Println("   - Use Core NATS for transient data (metrics, logs)")
	fmt.Println("   - Use JetStream for important events (orders, payments)")
	fmt.Println()
	fmt.Println("Next: Mini-project — Event-driven order processing pipeline!")
}

// truncate shortens a string to the given max length.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
