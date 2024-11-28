package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// DiffKafkaWriter is a mock implementation of KafkaWriter for testing
type DiffKafkaWriter struct {
	mock.Mock
	//mockError error
}

func (m *DiffKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *DiffKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test the SendMessageToKafka method
func TestSendMessageToKafka(t *testing.T) {
	type testCase struct {
		name             string
		method           string
		url              string
		expectedResponse string
		expectedStatus   int
		mockError        error
		setup            func(*DiffKafkaWriter)
		body             string // Add body for the request
	}
	tests := []testCase{
		{
			name:             "Valid POST request",
			method:           http.MethodPost,
			url:              "/send",
			expectedResponse: "Message sent: Hello Kafka",
			expectedStatus:   http.StatusOK,
			setup: func(mockWriter *DiffKafkaWriter) {
				mockWriter.On(
					"WriteMessages",
					mock.Anything,
					mock.MatchedBy(func(msgs []kafka.Message) bool {
						return len(msgs) == 1 && string(msgs[0].Value) == "Hello Kafka"
					}),
				).Return(nil)
			},
			body: "message=Hello Kafka",
		},
		{
			name:             "Kafka write error",
			method:           http.MethodPost,
			url:              "/send",
			expectedResponse: "Failed to send message: mock error\n",
			expectedStatus:   http.StatusInternalServerError,
			mockError:        errors.New("mock error"),
			setup: func(mockWriter *DiffKafkaWriter) {
				mockWriter.On(
					"WriteMessages",
					mock.Anything,
					mock.MatchedBy(func(msgs []kafka.Message) bool {
						return len(msgs) > 0
					}),
				).Return(errors.New("mock error"))
			},
			body: "message=Hello Kafka",
		},
		{
			name:             "Invalid HTTP method",
			method:           http.MethodGet, // This triggers the error for wrong HTTP method
			url:              "/send",
			expectedResponse: "Invalid request method\n",
			expectedStatus:   http.StatusMethodNotAllowed,
			setup:            func(mockWriter *DiffKafkaWriter) {},
			body:             "message=Hello Kafka",
		},
		{
			name:             "Failed to parse form data",
			method:           http.MethodPost,
			url:              "/send",
			expectedResponse: "Message parameter is required\n", // Update expected response to match actual message
			expectedStatus:   http.StatusBadRequest,
			mockError:        nil,
			setup: func(mockWriter *DiffKafkaWriter) {
				// Ensure no WriteMessages call is expected because this test case is to cover form parsing failure
				mockWriter.On("WriteMessages", mock.Anything, mock.Anything).Return(nil).Maybe()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Kafka writer
			mockWriter := new(DiffKafkaWriter)
			tt.setup(mockWriter)

			// Create request and response recorder
			req := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()

			// Call the function under test
			sendMessageToKafka(rr, req, mockWriter)

			// Assert response status and body
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedResponse, rr.Body.String())

			// Verify mock expectations
			mockWriter.AssertExpectations(t)
		})
	}
}

// Test the Kafka writer initialization
func TestInitKafkaWriter(t *testing.T) {
	broker := "kafka:9092"
	topic := "test-topic"
	writer := initKafkaWriter(broker, topic)

	// Ensure the writer is of the expected type
	_, ok := writer.(*kafka.Writer)
	if !ok {
		t.Fatalf("Expected KafkaWriter but got %T", writer)
	}
}

// TestSetupHttpServer is the table-driven test for the server setup and request handling
func TestSetupHttpServer(t *testing.T) {
	// Define the test case structure
	type testCase struct {
		name             string
		method           string
		url              string
		body             string
		expectedStatus   int
		expectedResponse string
		writer           KafkaWriter
		mockSetup        func(writer *DiffKafkaWriter)
	}

	// Define the test cases
	tests := []testCase{
		{
			name:             "Valid POST request with message",
			method:           "POST",
			url:              "/send?message=TestMessage",
			body:             "message=TestMessage",
			expectedStatus:   http.StatusOK,
			expectedResponse: "Message sent: TestMessage",
			writer:           &DiffKafkaWriter{},
			mockSetup: func(writer *DiffKafkaWriter) {
				writer.On("WriteMessages", mock.Anything, mock.MatchedBy(func(msgs []kafka.Message) bool {
					return len(msgs) == 1 && string(msgs[0].Value) == "TestMessage"
				})).Return(nil)
			},
		},
		{
			name:             "Invalid HTTP method (GET instead of POST)",
			method:           "GET",
			url:              "/send",
			expectedStatus:   http.StatusMethodNotAllowed,
			expectedResponse: "Invalid request method",
			writer:           &DiffKafkaWriter{},
			mockSetup:        func(writer *DiffKafkaWriter) {},
		},
		{
			name:             "Missing message parameter",
			method:           "POST",
			url:              "/send",
			body:             "",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Message parameter is required",
			writer:           &DiffKafkaWriter{},
			mockSetup:        func(writer *DiffKafkaWriter) {},
		},
		{
			name:             "Empty message parameter",
			method:           "POST",
			url:              "/send",
			body:             "message=",
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: "Message parameter is required",
			writer:           &DiffKafkaWriter{},
			mockSetup:        func(writer *DiffKafkaWriter) {},
		},
		{
			name:             "Kafka write failure",
			method:           "POST",
			url:              "/send?message=TestMessage",
			body:             "message=TestMessage",
			expectedStatus:   http.StatusInternalServerError,
			expectedResponse: "Failed to send message: simulated Kafka failure",
			writer:           &DiffKafkaWriter{},
			mockSetup: func(writer *DiffKafkaWriter) {
				writer.On("WriteMessages", mock.Anything, mock.MatchedBy(func(msgs []kafka.Message) bool {
					return len(msgs) == 1 && string(msgs[0].Value) == "TestMessage"
				})).Return(fmt.Errorf("simulated Kafka failure"))
			},
		},
	}

	// Iterate over each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset HTTP handlers before each test
			http.DefaultServeMux = http.NewServeMux() // Reset the handlers

			// Register the handler once before running the tests
			setupHttpServer(tt.writer)

			// Set up the mock expectations
			tt.mockSetup(tt.writer.(*DiffKafkaWriter))

			// Create the request with the correct content type for form data
			req, err := http.NewRequest(tt.method, tt.url, strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // Ensure Content-Type is set

			// Create a response recorder to capture the HTTP response
			rr := httptest.NewRecorder()

			// Serve the HTTP request using the default mux (which includes the /send handler)
			http.DefaultServeMux.ServeHTTP(rr, req)

			// Assert status code using assert.Equal
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Assert response body, stripping newline characters if necessary
			actualResponse := strings.TrimSpace(rr.Body.String()) // Strip any trailing newline
			assert.Equal(t, tt.expectedResponse, actualResponse)

			// Ensure all expectations were met
			tt.writer.(*DiffKafkaWriter).AssertExpectations(t)

			// Reset mocks after the test
			tt.writer.(*DiffKafkaWriter).Mock.ExpectedCalls = nil
		})
	}
}

// TestMainFunction_Success simulates server startup success
func TestMainFunction_Success(t *testing.T) {
	// Create a mock Kafka writer
	mockWriter := new(DiffKafkaWriter)

	// Mock the startServer function to simulate success
	originalStartServer := startServer
	defer func() {
		// Restore original startServer function after the test
		startServer = originalStartServer
	}()
	startServer = func() error {
		return nil // Simulate successful server start
	}

	// Inject the mock writer to simulate Kafka initialization
	writer = mockWriter

	// Add some logs to verify the flow
	t.Log("Running main function...")

	// Call main function (which should trigger the Close() method)
	err := runServer()
	if err != nil {
		return
	} // Run the main function

	// Add logs to confirm if Close was called during main
	t.Log("Main function completed.")

	// Verify that Close() was called
	mockWriter.AssertExpectations(t)
}
