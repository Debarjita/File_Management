# Use the official Golang image
FROM golang:1.21

# Set environment variables
ENV GO111MODULE=on

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go app
RUN go build -o main .

# Expose the port your app runs on (adjust if needed)
EXPOSE 8080

# Run the app
CMD ["./main"]
