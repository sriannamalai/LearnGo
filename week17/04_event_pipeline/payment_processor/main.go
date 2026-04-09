package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ========================================
// Week 17 — Mini-Project: Payment Processor
// ========================================
//
// This service subscribes to ORDERS.created events, processes payments,
// and publishes ORDERS.payment.processed or ORDERS.payment.failed events.
//
// Pipeline position:
//   Order Creator ──> [ORDERS.created] ──> THIS SERVICE ──> [ORDERS.payment.*]
//
// Prerequisites:
//   docker run -d --name nats -p 4222:4222 nats:latest -js
//
// To run:
//   go run main.go

// OrderCreatedEvent is received from the Order Creator.
type OrderCreatedEvent struct {
	OrderID   string      `json:"order_id"`
	Customer  string      `json:"customer"`
	Email     string      `json:"email"`
	Items     []OrderItem `json:"items"`
	Total     float64     `json:"total"`
	CreatedAt string      `json:"created_at"`
}

type OrderItem struct {
	Product  string  `json:"product"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// PaymentProcessedEvent is published after payment processing.
type PaymentProcessedEvent struct {
	OrderID       string  `json:"order_id"`
	Customer      string  `json:"customer"`
	Email         string  `json:"email"`
	Amount        float64 `json:"amount"`
	PaymentID     string  `json:"payment_id"`
	Status        string  `json:"status"` // "success" or "failed"
	FailureReason string  `json:"failure_reason,omitempty"`
	ProcessedAt   string  `json:"processed_at"`
}

func main() {
	fmt.Println("=== Week 17 Mini-Project: Payment Processor ===")
	fmt.Println()

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("payment-processor"))
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
	// Create a durable consumer for order processing
	// ========================================
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "payment-processor",
		FilterSubject: "ORDERS.created",
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
		MaxDeliver:    3, // Retry up to 3 times on failure
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	fmt.Println("Payment Processor listening for ORDERS.created events...")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// ========================================
	// Process messages continuously
	// ========================================
	// Consume creates a push-based consumer that calls the handler
	// for each message. This is the simplest consumption pattern.
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		// Parse the order event
		var order OrderCreatedEvent
		if err := json.Unmarshal(msg.Data(), &order); err != nil {
			log.Printf("Failed to parse order event: %v", err)
			msg.Term() // Bad message — don't redeliver
			return
		}

		fmt.Printf("[RECEIVED] Order %s from %s — $%.2f\n",
			order.OrderID, order.Customer, order.Total)

		// ========================================
		// Simulate payment processing
		// ========================================
		// In real life: call Stripe, PayPal, etc.
		time.Sleep(200 * time.Millisecond) // Simulate processing time

		paymentEvent := PaymentProcessedEvent{
			OrderID:     order.OrderID,
			Customer:    order.Customer,
			Email:       order.Email,
			Amount:      order.Total,
			PaymentID:   fmt.Sprintf("PAY-%s", order.OrderID[4:]),
			ProcessedAt: time.Now().Format(time.RFC3339),
		}

		// Simulate occasional payment failures (20% failure rate)
		if rand.Float64() < 0.2 {
			paymentEvent.Status = "failed"
			paymentEvent.FailureReason = "insufficient funds"
			fmt.Printf("  [PAYMENT FAILED] %s: %s\n", order.OrderID, paymentEvent.FailureReason)

			// Publish failure event
			data, _ := json.Marshal(paymentEvent)
			js.Publish(ctx, "ORDERS.payment.failed", data)
		} else {
			paymentEvent.Status = "success"
			fmt.Printf("  [PAYMENT SUCCESS] %s: Payment ID %s\n",
				order.OrderID, paymentEvent.PaymentID)

			// Publish success event
			data, _ := json.Marshal(paymentEvent)
			js.Publish(ctx, "ORDERS.payment.processed", data)
		}

		// Acknowledge the message — it has been processed
		msg.Ack()
		fmt.Println()
	})
	if err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
	defer cons.Stop()

	// ========================================
	// Graceful shutdown on Ctrl+C
	// ========================================
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nPayment Processor shutting down...")
}
