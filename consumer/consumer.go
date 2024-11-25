// consumer.go

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Kafka broker address (from docker-compose.yml)
	broker := "kafka:9092"

	// Create a Kafka reader (consumer)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{broker},
		Topic:     "test-topic",     // Kafka topic
		Partition: 0,                // Default partition
		GroupID:   "consumer-group", // Consumer group ID
	})

	defer reader.Close()

	for {
		// Read the next message from the topic
		message, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Fatalf("Error reading message: %v", err)
		}

		// Process the message
		fmt.Printf("Received: %s\n", string(message.Value))
	}
}
