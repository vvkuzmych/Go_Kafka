// producer.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Kafka broker address (from docker-compose.yml)
	broker := "kafka:9092"

	// Create a Kafka writer (producer)
	writer := kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    "test-topic",        // Kafka topic
		Balancer: &kafka.LeastBytes{}, // Balance messages across partitions
	}

	defer func(writer *kafka.Writer) {
		err := writer.Close()
		if err != nil {
			return
		}
	}(&writer)

	for {
		// Create a message to send
		message := "Hello from Go Producer"

		// Send the message
		err := writer.WriteMessages(
			context.Background(),
			kafka.Message{
				Value: []byte(message),
			},
		)
		if err != nil {
			log.Printf("Error writing message: %v", err)
		} else {
			fmt.Printf("Sent: %s\n", message)
		}

		// Sleep for 1 second before sending the next message
		time.Sleep(1 * time.Second)
	}
}
