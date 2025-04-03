FROM golang:1.22.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o quiz-app ./cmd/api

# Use a minimal alpine image for the final container
FROM alpine:latest

# Install necessary certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/quiz-app .

# Copy configuration files if needed
COPY config.yaml ./

# Expose the application port
EXPOSE 8080

# Command to run the application
CMD ["./quiz-app"]