package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/segmentio/kafka-go"
)

var writer *kafka.Writer

func initKafkaWriter(broker, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func sendMessageToKafka(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	message := r.URL.Query().Get("message")
	if message == "" {
		http.Error(w, "Message parameter is required", http.StatusBadRequest)
		return
	}

	err := writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Value: []byte(message),
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		log.Printf("Error writing message: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Message sent: %s", message)
}

func main() {
	broker := "kafka:9092"
	topic := "test-topic"

	writer = initKafkaWriter(broker, topic)
	defer writer.Close()

	http.HandleFunc("/send", sendMessageToKafka)

	fmt.Println("Server is running on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
