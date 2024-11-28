package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/segmentio/kafka-go"
)

// KafkaWriter is an interface that wraps around the necessary methods of kafka.Writer
type KafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

var writer KafkaWriter

// initKafkaWriter creates a new KafkaWriter for real Kafka interactions
func initKafkaWriter(broker, topic string) KafkaWriter {
	return &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

// sendMessageToKafka now accepts a KafkaWriter for better testability
//func sendMessageToKafka(w http.ResponseWriter, r *http.Request, writer KafkaWriter) {
//	if r.Method != http.MethodPost {
//		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
//		return
//	}
//
//	message := r.URL.Query().Get("message")
//	if message == "" {
//		http.Error(w, "Message parameter is required", http.StatusBadRequest)
//		return
//	}
//
//	err := writer.WriteMessages(
//		context.Background(),
//		kafka.Message{
//			Value: []byte(message),
//		},
//	)
//	if err != nil {
//		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
//		log.Printf("Error writing message: %v", err)
//		return
//	}
//
//	w.WriteHeader(http.StatusOK)
//	fmt.Fprintf(w, "Message sent: %s", message)
//}

func sendMessageToKafka(w http.ResponseWriter, r *http.Request, writer KafkaWriter) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the form data for POST requests
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")
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

// initializeKafkaWriter sets up the Kafka writer and returns it
func initializeKafkaWriter() KafkaWriter {
	broker := "kafka:9092"
	topic := "test-topic"
	return initKafkaWriter(broker, topic)
}

// setupHttpServer sets up the HTTP handler
func setupHttpServer(writer KafkaWriter) {
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		sendMessageToKafka(w, r, writer)
	})
}

// startServer starts the HTTP server
var startServer = func() error {
	fmt.Println("Server is running on :8080...")
	return http.ListenAndServe(":8080", nil)
}

// runServer encapsulates the main logic of running the server and handles errors
func runServer() error {
	// Initialize Kafka writer
	writer := initializeKafkaWriter()
	defer writer.Close()

	// Setup HTTP server
	setupHttpServer(writer)

	// Start the server
	if err := startServer(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
		return err
	}
	return nil
}

func main() {
	if err := runServer(); err != nil {
		log.Fatal(err)
	}
}
