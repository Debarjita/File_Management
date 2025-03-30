# Dockerfile
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o fileserver ./cmd/server

# Create final image
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/fileserver .

# Create directory for local storage
RUN mkdir -p /app/storage

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./fileserver"]