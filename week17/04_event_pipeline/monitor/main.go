package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ========================================
// Week 17 — Mini-Project: Event Monitor
// ========================================
//
// This service monitors ALL events in the pipeline and logs them.
// It subscribes to the entire ORDERS.> subject hierarchy.
//
// In a real system, this would:
//   - Store events in an analytics database
//   - Update dashboards in real-time
//   - Detect anomalies and trigger alerts
//
// Pipeline position:
//   ALL events ──> THIS SERVICE (observes everything)
//
// Prerequisites:
//   docker run -d --name nats -p 4222:4222 nats:latest -js
//
// To run (start this FIRST, before other services):
//   go run main.go

func main() {
	fmt.Println("=== Week 17 Mini-Project: Event Monitor ===")
	fmt.Println()

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("event-monitor"))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Get JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ========================================
	// Ensure the stream exists
	// ========================================
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "ORDERS",
		Subjects:  []string{"ORDERS.>"},
		Storage:   jetstream.MemoryStorage,
		Retention: jetstream.LimitsPolicy,
		MaxMsgs:   10000,
		MaxAge:    1 * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to ensure stream: %v", err)
	}

	// ========================================
	// Create a monitor consumer (no filter — see ALL events)
	// ========================================
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "event-monitor",
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		// No FilterSubject — we want ALL events in the ORDERS stream
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	// ========================================
	// Track statistics
	// ========================================
	var totalEvents atomic.Int64
	var ordersCreated atomic.Int64
	var paymentsSuccess atomic.Int64
	var paymentsFailed atomic.Int64

	// ========================================
	// Consume and log all events
	// ========================================
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		totalEvents.Add(1)
		count := totalEvents.Load()

		// Determine event type from subject
		subject := msg.Subject()
		eventType := categorizeEvent(subject)

		// Update counters
		switch subject {
		case "ORDERS.created":
			ordersCreated.Add(1)
		case "ORDERS.payment.processed":
			paymentsSuccess.Add(1)
		case "ORDERS.payment.failed":
			paymentsFailed.Add(1)
		}

		// Parse and log the event
		var raw map[string]interface{}
		json.Unmarshal(msg.Data(), &raw)

		// Format a nice log entry
		orderID := getString(raw, "order_id")
		customer := getString(raw, "customer")

		timestamp := time.Now().Format("15:04:05.000")

		fmt.Printf("[%s] #%d %s | %-20s | Order: %-8s | Customer: %s\n",
			timestamp, count, eventIcon(subject), eventType, orderID, customer)

		// Show extra details for payment events
		if subject == "ORDERS.payment.processed" {
			fmt.Printf("         Payment ID: %s | Amount: $%.2f\n",
				getString(raw, "payment_id"), getFloat(raw, "amount"))
		} else if subject == "ORDERS.payment.failed" {
			fmt.Printf("         Reason: %s | Amount: $%.2f\n",
				getString(raw, "failure_reason"), getFloat(raw, "amount"))
		}

		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
	defer cons.Stop()

	fmt.Println("Event Monitor watching ALL events on ORDERS.>")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()
	fmt.Println("---------------------------------------------------")
	fmt.Printf("%-14s %-5s %-8s | %-20s | %-14s | %s\n",
		"TIMESTAMP", "#", "", "EVENT TYPE", "ORDER", "CUSTOMER")
	fmt.Println("---------------------------------------------------")

	// ========================================
	// Print periodic stats
	// ========================================
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-ticker.C:
			fmt.Println()
			fmt.Println("--- PIPELINE STATS ---")
			fmt.Printf("  Total events:     %d\n", totalEvents.Load())
			fmt.Printf("  Orders created:   %d\n", ordersCreated.Load())
			fmt.Printf("  Payments success: %d\n", paymentsSuccess.Load())
			fmt.Printf("  Payments failed:  %d\n", paymentsFailed.Load())
			fmt.Println("----------------------")
			fmt.Println()
		case <-sigCh:
			fmt.Println()
			fmt.Println("=== FINAL STATS ===")
			fmt.Printf("Total events:     %d\n", totalEvents.Load())
			fmt.Printf("Orders created:   %d\n", ordersCreated.Load())
			fmt.Printf("Payments success: %d\n", paymentsSuccess.Load())
			fmt.Printf("Payments failed:  %d\n", paymentsFailed.Load())
			fmt.Println("Monitor shutting down...")
			return
		}
	}
}

// categorizeEvent returns a human-readable event type.
func categorizeEvent(subject string) string {
	switch subject {
	case "ORDERS.created":
		return "ORDER_CREATED"
	case "ORDERS.payment.processed":
		return "PAYMENT_SUCCESS"
	case "ORDERS.payment.failed":
		return "PAYMENT_FAILED"
	case "ORDERS.paid":
		return "ORDER_PAID"
	case "ORDERS.shipped":
		return "ORDER_SHIPPED"
	case "ORDERS.delivered":
		return "ORDER_DELIVERED"
	default:
		return subject
	}
}

// eventIcon returns a text indicator for the event type.
func eventIcon(subject string) string {
	switch subject {
	case "ORDERS.created":
		return "[NEW]"
	case "ORDERS.payment.processed":
		return "[PAY]"
	case "ORDERS.payment.failed":
		return "[ERR]"
	default:
		return "[---]"
	}
}

// getString safely extracts a string from a map.
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getFloat safely extracts a float64 from a map.
func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}
