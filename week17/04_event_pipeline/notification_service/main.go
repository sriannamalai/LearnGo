package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// ========================================
// Week 17 — Mini-Project: Notification Service
// ========================================
//
// This service subscribes to payment events and sends notifications.
// In a real system, this would send emails, SMS, push notifications, etc.
// Here we print to console for educational purposes.
//
// Pipeline position:
//   Payment Processor ──> [ORDERS.payment.*] ──> THIS SERVICE
//
// Prerequisites:
//   docker run -d --name nats -p 4222:4222 nats:latest -js
//
// To run:
//   go run main.go

// PaymentProcessedEvent is received from the Payment Processor.
type PaymentProcessedEvent struct {
	OrderID       string  `json:"order_id"`
	Customer      string  `json:"customer"`
	Email         string  `json:"email"`
	Amount        float64 `json:"amount"`
	PaymentID     string  `json:"payment_id"`
	Status        string  `json:"status"`
	FailureReason string  `json:"failure_reason,omitempty"`
	ProcessedAt   string  `json:"processed_at"`
}

func main() {
	fmt.Println("=== Week 17 Mini-Project: Notification Service ===")
	fmt.Println()

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("notification-service"))
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
	// Consumer for payment success notifications
	// ========================================
	successConsumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "notifier-success",
		FilterSubject: "ORDERS.payment.processed",
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create success consumer: %v", err)
	}

	successSub, err := successConsumer.Consume(func(msg jetstream.Msg) {
		var event PaymentProcessedEvent
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			log.Printf("Failed to parse event: %v", err)
			msg.Term()
			return
		}

		// ========================================
		// Send success notification (simulated)
		// ========================================
		fmt.Println("========================================")
		fmt.Println("  NOTIFICATION: Order Confirmed!")
		fmt.Println("========================================")
		fmt.Printf("  To:       %s <%s>\n", event.Customer, event.Email)
		fmt.Printf("  Subject:  Order %s Confirmed\n", event.OrderID)
		fmt.Printf("  Body:\n")
		fmt.Printf("    Dear %s,\n", event.Customer)
		fmt.Printf("    Your order %s has been confirmed!\n", event.OrderID)
		fmt.Printf("    Amount charged: $%.2f\n", event.Amount)
		fmt.Printf("    Payment ID: %s\n", event.PaymentID)
		fmt.Printf("    Thank you for your purchase!\n")
		fmt.Println("========================================")
		fmt.Println()

		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to start success consumer: %v", err)
	}
	defer successSub.Stop()

	// ========================================
	// Consumer for payment failure notifications
	// ========================================
	failureConsumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       "notifier-failure",
		FilterSubject: "ORDERS.payment.failed",
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		AckWait:       30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create failure consumer: %v", err)
	}

	failureSub, err := failureConsumer.Consume(func(msg jetstream.Msg) {
		var event PaymentProcessedEvent
		if err := json.Unmarshal(msg.Data(), &event); err != nil {
			log.Printf("Failed to parse event: %v", err)
			msg.Term()
			return
		}

		// ========================================
		// Send failure notification (simulated)
		// ========================================
		fmt.Println("========================================")
		fmt.Println("  NOTIFICATION: Payment Failed")
		fmt.Println("========================================")
		fmt.Printf("  To:       %s <%s>\n", event.Customer, event.Email)
		fmt.Printf("  Subject:  Payment Failed for Order %s\n", event.OrderID)
		fmt.Printf("  Body:\n")
		fmt.Printf("    Dear %s,\n", event.Customer)
		fmt.Printf("    Unfortunately, payment for order %s failed.\n", event.OrderID)
		fmt.Printf("    Reason: %s\n", event.FailureReason)
		fmt.Printf("    Amount: $%.2f\n", event.Amount)
		fmt.Printf("    Please update your payment method and try again.\n")
		fmt.Println("========================================")
		fmt.Println()

		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to start failure consumer: %v", err)
	}
	defer failureSub.Stop()

	fmt.Println("Notification Service listening for payment events...")
	fmt.Println("  Subscribed to: ORDERS.payment.processed")
	fmt.Println("  Subscribed to: ORDERS.payment.failed")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// ========================================
	// Graceful shutdown
	// ========================================
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nNotification Service shutting down...")
}
