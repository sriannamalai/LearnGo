package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// ========================================
// Week 17 — Lesson 2: NATS Request/Reply
// ========================================
//
// Request/Reply is NATS's pattern for synchronous communication.
// Unlike pub/sub (fire-and-forget), request/reply lets you:
//   - Send a request and WAIT for a response
//   - Set timeouts for responses
//   - Build service-like APIs over messaging
//
// How it works under the hood:
//   1. Client creates a unique "inbox" subject (e.g., _INBOX.abc123)
//   2. Client subscribes to the inbox
//   3. Client publishes the request with Reply set to the inbox
//   4. Service receives the request, processes it, publishes to the Reply subject
//   5. Client receives the response on the inbox
//
// This is how you build microservices with NATS!
//
// Prerequisites:
//   docker run -d --name nats -p 4222:4222 nats:latest
//
// To run:
//   go run main.go

// ========================================
// Data types for our services
// ========================================

// PriceRequest asks for a product price.
type PriceRequest struct {
	ProductID string `json:"product_id"`
}

// PriceResponse contains the price lookup result.
type PriceResponse struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	InStock   bool    `json:"in_stock"`
	Error     string  `json:"error,omitempty"`
}

// HealthResponse from a service health check.
type HealthResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Uptime    string `json:"uptime"`
	RequestID string `json:"request_id"`
}

func main() {
	fmt.Println("=== Week 17, Lesson 2: NATS Request/Reply ===")
	fmt.Println()

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, nats.Name("request-reply-demo"))
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	fmt.Printf("Connected to NATS at %s\n", nc.ConnectedUrl())
	fmt.Println()

	// ========================================
	// Step 1: Simple Request/Reply
	// ========================================
	fmt.Println("--- Simple Request/Reply ---")
	fmt.Println()

	// Set up a "service" that responds to requests.
	// This is the RESPONDER — it subscribes and replies.
	nc.Subscribe("echo", func(msg *nats.Msg) {
		fmt.Printf("  [Service] Received request: %s\n", string(msg.Data))
		fmt.Printf("  [Service] Reply subject: %s\n", msg.Reply)

		// msg.Respond() sends a reply back to the requester.
		// It publishes to the msg.Reply subject (the inbox).
		msg.Respond([]byte("Echo: " + string(msg.Data)))
	})

	// This is the REQUESTER — it sends a request and waits.
	// Request() publishes a message and waits for a single reply.
	// The timeout prevents waiting forever if the service is down.
	fmt.Println("Sending request to 'echo' service...")
	resp, err := nc.Request("echo", []byte("Hello, NATS!"), 2*time.Second)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	fmt.Printf("  [Client] Response: %s\n", string(resp.Data))
	fmt.Println()

	// ========================================
	// Step 2: JSON Request/Reply (Price Service)
	// ========================================
	fmt.Println("--- JSON Request/Reply: Price Service ---")
	fmt.Println()

	// Product catalog (simulating a database)
	catalog := map[string]PriceResponse{
		"PROD-001": {ProductID: "PROD-001", Name: "Go Programming Book", Price: 49.99, InStock: true},
		"PROD-002": {ProductID: "PROD-002", Name: "Mechanical Keyboard", Price: 149.99, InStock: true},
		"PROD-003": {ProductID: "PROD-003", Name: "Ultra-Wide Monitor", Price: 599.99, InStock: false},
	}

	// Price service — processes JSON requests
	nc.Subscribe("price.lookup", func(msg *nats.Msg) {
		// Deserialize the request
		var req PriceRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			errResp, _ := json.Marshal(PriceResponse{Error: "invalid request format"})
			msg.Respond(errResp)
			return
		}

		// Look up the product
		product, exists := catalog[req.ProductID]
		if !exists {
			errResp, _ := json.Marshal(PriceResponse{
				ProductID: req.ProductID,
				Error:     "product not found",
			})
			msg.Respond(errResp)
			return
		}

		// Send the response
		respData, _ := json.Marshal(product)
		msg.Respond(respData)
	})

	// Client: look up prices for multiple products
	productIDs := []string{"PROD-001", "PROD-002", "PROD-003", "PROD-999"}

	for _, id := range productIDs {
		req := PriceRequest{ProductID: id}
		reqData, _ := json.Marshal(req)

		resp, err := nc.Request("price.lookup", reqData, 2*time.Second)
		if err != nil {
			fmt.Printf("  Request for %s failed: %v\n", id, err)
			continue
		}

		var priceResp PriceResponse
		json.Unmarshal(resp.Data, &priceResp)

		if priceResp.Error != "" {
			fmt.Printf("  %s: ERROR — %s\n", id, priceResp.Error)
		} else {
			stock := "In Stock"
			if !priceResp.InStock {
				stock = "Out of Stock"
			}
			fmt.Printf("  %s: %s — $%.2f (%s)\n", id, priceResp.Name, priceResp.Price, stock)
		}
	}
	fmt.Println()

	// ========================================
	// Step 3: Request/Reply with Queue Groups
	// ========================================
	fmt.Println("--- Request/Reply with Queue Groups ---")
	fmt.Println()
	fmt.Println("Queue groups work with request/reply too!")
	fmt.Println("Multiple service instances, but only one handles each request.")
	fmt.Println()

	// Start 3 instances of a "worker" service in the same queue group
	for i := 1; i <= 3; i++ {
		workerName := fmt.Sprintf("worker-%d", i)
		nc.QueueSubscribe("compute", "compute-workers", func(msg *nats.Msg) {
			// Simulate some processing
			result := fmt.Sprintf("Computed by %s: %s -> result_%d",
				workerName, string(msg.Data), rand.Intn(1000))
			msg.Respond([]byte(result))
		})
	}

	// Send 5 requests — each handled by one worker
	for i := 1; i <= 5; i++ {
		resp, err := nc.Request("compute", []byte(fmt.Sprintf("job-%d", i)), 2*time.Second)
		if err != nil {
			fmt.Printf("  Request %d failed: %v\n", i, err)
			continue
		}
		fmt.Printf("  Job %d: %s\n", i, string(resp.Data))
	}
	fmt.Println()

	// ========================================
	// Step 4: Timeout Handling
	// ========================================
	fmt.Println("--- Timeout Handling ---")
	fmt.Println()

	// Service with random delays
	nc.Subscribe("slow.service", func(msg *nats.Msg) {
		delay := time.Duration(rand.Intn(3000)) * time.Millisecond
		time.Sleep(delay)
		msg.Respond([]byte(fmt.Sprintf("Processed after %v", delay)))
	})

	// Try requests with a short timeout
	fmt.Println("Sending requests to 'slow.service' with 1s timeout:")
	for i := 1; i <= 5; i++ {
		start := time.Now()
		resp, err := nc.Request("slow.service", []byte("ping"), 1*time.Second)
		elapsed := time.Since(start)

		if err == nats.ErrTimeout {
			fmt.Printf("  Request %d: TIMEOUT after %v\n", i, elapsed.Round(time.Millisecond))
		} else if err != nil {
			fmt.Printf("  Request %d: Error — %v\n", i, err)
		} else {
			fmt.Printf("  Request %d: %s (%v)\n", i, string(resp.Data), elapsed.Round(time.Millisecond))
		}
	}
	fmt.Println()

	// ========================================
	// Step 5: Scatter/Gather Pattern
	// ========================================
	fmt.Println("--- Scatter/Gather Pattern ---")
	fmt.Println()
	fmt.Println("Sometimes you want ALL services to respond, not just one.")
	fmt.Println("Use a regular subscribe (not queue group) and collect replies.")
	fmt.Println()

	// Multiple health check responders (each is a different service)
	services := []string{"user-service", "order-service", "payment-service"}
	for _, svcName := range services {
		name := svcName // capture for closure
		nc.Subscribe("health.check", func(msg *nats.Msg) {
			resp := HealthResponse{
				Service:   name,
				Status:    "healthy",
				Uptime:    "2h30m",
				RequestID: string(msg.Data),
			}
			data, _ := json.Marshal(resp)
			msg.Respond(data)
		})
	}

	// Collect responses from all services
	// Use NewInbox() to create a unique reply subject
	inbox := nats.NewInbox()
	var responses []HealthResponse
	var mu sync.Mutex
	var responseWg sync.WaitGroup
	responseWg.Add(len(services))

	nc.Subscribe(inbox, func(msg *nats.Msg) {
		var hr HealthResponse
		json.Unmarshal(msg.Data, &hr)
		mu.Lock()
		responses = append(responses, hr)
		mu.Unlock()
		responseWg.Done()
	})

	// Publish with the inbox as the reply subject
	nc.PublishRequest("health.check", inbox, []byte("req-001"))
	nc.Flush()

	// Wait for all responses (with timeout)
	done := make(chan struct{})
	go func() {
		responseWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("All services responded:")
		for _, hr := range responses {
			fmt.Printf("  %s: %s (uptime: %s)\n", hr.Service, hr.Status, hr.Uptime)
		}
	case <-time.After(2 * time.Second):
		fmt.Println("Timeout — not all services responded")
		mu.Lock()
		for _, hr := range responses {
			fmt.Printf("  %s: %s\n", hr.Service, hr.Status)
		}
		mu.Unlock()
	}
	fmt.Println()

	// ========================================
	// Key Concepts Summary
	// ========================================
	fmt.Println("--- Key Concepts ---")
	fmt.Println()
	fmt.Println("1. REQUEST/REPLY is synchronous messaging:")
	fmt.Println("   Client sends request, waits for response")
	fmt.Println()
	fmt.Println("2. Under the hood: uses temporary INBOX subjects")
	fmt.Println("   NATS creates _INBOX.xxx for each request")
	fmt.Println()
	fmt.Println("3. ALWAYS set TIMEOUTS on requests:")
	fmt.Println("   nc.Request(subject, data, 5*time.Second)")
	fmt.Println()
	fmt.Println("4. QUEUE GROUPS work with request/reply:")
	fmt.Println("   Distribute requests across service instances")
	fmt.Println()
	fmt.Println("5. SCATTER/GATHER collects from all responders:")
	fmt.Println("   Use PublishRequest with a custom inbox")
	fmt.Println()
	fmt.Println("Next: JetStream for persistent, durable messaging")
}
