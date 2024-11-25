# Consumer Dockerfile

# Start from the official Golang image
FROM golang:1.23-alpine

# Set the working directory
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod tidy

# Copy the rest of the application files
COPY . ./

# Build the Go application
RUN go build -o consumer .

# Command to run the consumer
CMD ["./consumer"]
