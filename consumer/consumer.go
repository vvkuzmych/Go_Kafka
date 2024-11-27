package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Kafka broker address and topic
	broker := "kafka:9092"
	topic := "test-topic"
	groupID := "consumer-group"

	// Create a Kafka reader (consumer)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: groupID, // Dynamic partition assignment
	})

	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("Error closing reader: %v", err)
		}
	}()

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	go handleShutdown(cancel)

	fmt.Println("Consumer is running...")
	for {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			// Exit loop if the context is canceled
			if ctx.Err() != nil {
				fmt.Println("Shutting down consumer...")
				break
			}
			log.Printf("Error reading message: %v", err)
			continue
		}

		// Process the message
		fmt.Printf("Received: %s\n", string(message.Value))
	}
}

// handleShutdown listens for OS signals and cancels the context
func handleShutdown(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan
	cancel()
}
